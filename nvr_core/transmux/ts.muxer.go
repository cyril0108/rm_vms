package transmux

import (
	"context"
	"fmt"
	"net/http"

	"github.com/asticode/go-astits"

	"nvr_core/logger"
	"nvr_core/stream"
)

var LOG = logger.NewLogger("[nvr_core][transmux]")


const (
	VideoPID   uint16 = 256
	AudioPID   uint16 = 257
	VideoPESID uint8  = 224 // 0xE0
	AudioPESID uint8  = 192 // 0xC0 (For AAC/MP3)
	PrivatePESID uint8  = 189 // 0xBD (For G.711/PCM)
)


// =====================================================================
//  MUXER STATE MACHINE: Handles Dynamic PMT, PIDs, and Packetization
// =====================================================================

type TSMuxSession struct {
	muxer           *astits.Muxer

	clock           *StreamClock

	// firstPTS        int64 // Track the baseline timestamp for this session
	// lastDTS         int64 // Track the baseline timestamp for this session
	// hasFirstPTS     bool  // Flag to know if we've locked the baseline
	// audioRegistered bool
	pmtWritten      bool
	videoCodec      uint32
	audioCodec      uint32
	flusher         http.Flusher

	// hasSentFirstIFrame bool
	debugCount      int
	// warmupCount     int
}

func NewTSMuxSession(ctx context.Context, w http.ResponseWriter, vCodec uint32, aCodec uint32) *TSMuxSession {
	flusher, _ := w.(http.Flusher)
	return &TSMuxSession{
		muxer:      astits.NewMuxer(ctx, w),
		clock: NewStreamClock(),
		flusher:    flusher,
		videoCodec: vCodec,
		audioCodec: aCodec,
	}
}

// ProcessPacket manages the state of the stream (Zero-Delay Predefined PMT)
func (s *TSMuxSession) ProcessPacket(packet stream.StreamPacket) error {

	s.FactCheckNullUnitHeaders(&packet)
	s.FactCheckDiagnostic(&packet)
	s.FactCheckADTS(&packet)

	// --- INSTANT PMT GENERATION ---
	// Because we already know the codecs, we just wait for the first Video Keyframe 
	// to ensure the stream starts cleanly, then write the PMT immediately.
	if !s.pmtWritten {
		if packet.MediaType == stream.MediaTypeVideo && packet.IsKeyFrame {

			LOG.Info("[ProcessPacket] Writing Predefined PMT", 
				"VideoCodec", s.videoCodec, 
				"AudioCodec", s.audioCodec)

			// Bind Video
			s.muxer.AddElementaryStream(astits.PMTElementaryStream{
				ElementaryPID: VideoPID,
				StreamType:    stream.GetTSStreamType(s.videoCodec),
			})
			s.muxer.SetPCRPID(VideoPID)

			// Bind Audio WITH the Registration Descriptors
			if s.audioCodec != 0 {
				s.muxer.AddElementaryStream(astits.PMTElementaryStream{
					ElementaryPID:               AudioPID,
					StreamType:                  stream.GetTSStreamType(s.audioCodec), 
					ElementaryStreamDescriptors: s.PrepareAudioDescriptors(),
				})
			}

			s.pmtWritten = true
			s.muxer.WriteTables() // Written exactly once!
			
		} else {
			// Drop all P-Frames or Audio packets until that first Video Keyframe arrives
			return nil 
		}
	}

	// --- PAYLOAD DISPATCH ---
	return s.writePayload(packet)
}

// writePayload wraps the raw data into PES packets and writes to HTTP stream
func (s *TSMuxSession) writePayload(packet stream.StreamPacket) error {
	var streamID uint8
	var targetPID uint16
	isVideo := packet.MediaType == stream.MediaTypeVideo

	if isVideo {
		streamID = VideoPESID
		targetPID = VideoPID
	} else if packet.MediaType == stream.MediaTypeAudio {
		if s.AudioIsPCM() {
			streamID = PrivatePESID
		} else {
			streamID = AudioPESID 
		}
		targetPID = AudioPID
	} else {
		return nil 
	}

	// --- DELEGATE TO STREAM CLOCK ---
	var normalizedPTS, normalizedDTS, pcrBase int64

	if isVideo {
		s.InjectAUD(&packet)
		normalizedPTS, normalizedDTS, pcrBase = s.clock.NormalizeVideo(packet.PTS, packet.DTS)
	} else {
		normalizedPTS, normalizedDTS = s.clock.NormalizeAudio(packet.PTS, packet.DTS)
	}


	// --- PACKAGE PES ---
	optHeader := &astits.PESOptionalHeader{
	    PTS: &astits.ClockReference{Base: normalizedPTS},
	}

	// THE FIX: Only inject DTS if it differs from PTS (Mandatory for Video B-frames, forbidden for Audio)
	if isVideo && (normalizedDTS != normalizedPTS) {
	    optHeader.DTS = &astits.ClockReference{Base: normalizedDTS}
	}

	pes := &astits.PESData{
	    Header: &astits.PESHeader{
	        OptionalHeader: optHeader,
	        StreamID:       streamID,
	    },
	    Data: packet.Payload,
	}

	muxerData := &astits.MuxerData{
		PES: pes,
		PID: targetPID,
	}

	// Inject the Master Clock (PCR) on video frames using the "Right Now" time
	if isVideo {
		pcr := &astits.ClockReference{Base: pcrBase}
		muxerData.AdaptationField = &astits.PacketAdaptationField{
			HasPCR: true,
			PCR:    pcr,
		}
		s.FactCheckPCR(pcr)
	}

	if _, err := s.muxer.WriteData(muxerData); err != nil {
		return err
	}

	if isVideo && s.flusher != nil {
		s.flusher.Flush()
	}

	return nil
}

func (s *TSMuxSession) InjectAUD(packet *stream.StreamPacket) {
	if len(packet.Payload) > 4 {
	    // If the 5th byte is not 0x09 (AUD), inject it.
	    if packet.Payload[4] != 0x09 {
	        aud := []byte{0x00, 0x00, 0x00, 0x01, 0x09, 0xF0}
	        packet.Payload = append(aud, packet.Payload...)
	    }
	}
}


func (s *TSMuxSession) AudioIsPCM() bool {
	return s.audioCodec == stream.FFMpegCodecULaw || s.audioCodec == stream.FFMpegCodecALaw
}

func (s *TSMuxSession) PrepareAudioDescriptors() []*astits.Descriptor {
	var audioDescriptors []*astits.Descriptor

	if s.audioCodec == stream.FFMpegCodecULaw { // G.711 PCMU (u-law)
	    audioDescriptors = append(audioDescriptors, &astits.Descriptor{
	        Tag: astits.DescriptorTagRegistration, // 0x05
	        Registration: &astits.DescriptorRegistration{
	            FormatIdentifier: 0x756c6177, // "ulaw"
	            // FIX: Explicitly initialize empty slice to bypass go-astits CRC corruption
	            AdditionalIdentificationInfo: []byte{}, 
	        },
	    })
	} else if s.audioCodec == stream.FFMpegCodecALaw { // G.711 PCMA (a-law)
	    audioDescriptors = append(audioDescriptors, &astits.Descriptor{
	        Tag: astits.DescriptorTagRegistration,
	        Registration: &astits.DescriptorRegistration{
	            FormatIdentifier: 0x616c6177, // "alaw"
	            // FIX: Explicitly initialize empty slice
	            AdditionalIdentificationInfo: []byte{}, 
	        },
	    })
	}

	return audioDescriptors
}

// =============================================
// ====== FACT CHECK DIAGNOSTIC FUNCTIONS ======
// =============================================
func (s *TSMuxSession) FactCheckDiagnostic(packet *stream.StreamPacket) {
	if s.debugCount < 15 {
		LOG.Info("RAW DATA", 
			"media", packet.MediaType, 
			"key", packet.IsKeyFrame, 
			"pts", packet.PTS, 
			"dts", packet.DTS)
		s.debugCount++
	}
}

func (s *TSMuxSession) FactCheckPCR(pcr *astits.ClockReference) {
	if s.debugCount < 15 {
		LOG.Info("PCR", 
			"base", pcr.Base, 
			"time", pcr.Time(), 
			"duration", pcr.Duration())
		s.debugCount++
	}
}

func (s *TSMuxSession) FactCheckNullUnitHeaders(packet *stream.StreamPacket) {

	// --- FACT CHECK DIAGNOSTIC: NAL UNIT HEADERS ---
	if packet.MediaType == stream.MediaTypeVideo && s.debugCount < 10 {
		length := len(packet.Payload)
		if length > 20 {
			length = 20
		}
		// Print the first 20 bytes in Hexadecimal
		LOG.Info("VIDEO PAYLOAD HEAD", 
			"isKey", packet.IsKeyFrame, 
			"hex", fmt.Sprintf("%X", packet.Payload[:length]))
		s.debugCount++
	}
	// -----------------------------------------------

}


func (s *TSMuxSession) FactCheckADTS(packet *stream.StreamPacket) {

	// --- FACT CHECK DIAGNOSTIC: ADTS HEADER ---
	if packet.MediaType == stream.MediaTypeAudio && s.debugCount < 10 {
	    length := len(packet.Payload)
	    if length > 15 { length = 15 }
	    LOG.Info("AUDIO PAYLOAD HEAD", "hex", fmt.Sprintf("%X", packet.Payload[:length]))
	    s.debugCount++
	}
	// ------------------------------------------

}

// ===END=== FACT CHECK DIAGNOSTIC FUNCTIONS ======


// ========= Legacy/Preserve FUNCTIONS ======

// func (s *TSMuxSession) ProcessPacketNormal(packet stream.StreamPacket) error {

// 	s.FactCheckDiagnostic(&packet)

// 	// --- LATE AUDIO BINDING LOGIC ---
// 	if packet.MediaType == stream.MediaTypeAudio {
// 		if !s.audioRegistered {
// 			s.audioCodec = packet.CodecID
// 			if s.pmtWritten {
// 				// Audio arrived AFTER video keyframe. Inject dynamically.
// 				descriptors := s.PrepareAudioDescriptors()
// 				s.muxer.AddElementaryStream(astits.PMTElementaryStream{
// 					ElementaryPID: AudioPID,
// 					StreamType:    stream.GetTSStreamType(s.audioCodec),
// 					ElementaryStreamDescriptors: descriptors,
// 				})
// 				s.audioRegistered = true
// 				s.muxer.WriteTables()

// 				LOG.Info("[ProcessPacket] Audio registered",
// 					"pktcodec", packet.CodecID,
// 					"0x90", 0x90,
// 					"StreamType", stream.GetTSStreamType(packet.CodecID),
// 					"astits", astits.StreamType(0x90),
// 				)

// 			} else {
// 				return nil // PMT not written yet, drop audio packet but keep codec saved
// 			}
// 		}
// 	}

// 	// log.Printf("[TSMuxSession] audio passed")

// 	// --- VIDEO KEYFRAME / INITIAL BINDING LOGIC ---
// 	if packet.MediaType == stream.MediaTypeVideo && packet.IsKeyFrame {
// 		if !s.pmtWritten {

// 			LOG.Info("[ProcessPacket] ", "VideoCodec", stream.GetTSStreamType(packet.CodecID))

// 			//  Bind Video
// 			s.muxer.AddElementaryStream(astits.PMTElementaryStream{
// 				ElementaryPID: VideoPID,
// 				StreamType:    stream.GetTSStreamType(packet.CodecID),
// 			})
// 			s.muxer.SetPCRPID(VideoPID)

// 			//  Bind Audio (If it arrived before this keyframe)
// 			if s.audioCodec != 0 && !s.audioRegistered {

// 				descriptors := s.PrepareAudioDescriptors()
// 				s.muxer.AddElementaryStream(astits.PMTElementaryStream{
// 					ElementaryPID: AudioPID,
// 					StreamType:    stream.GetTSStreamType(s.audioCodec),
// 					ElementaryStreamDescriptors: descriptors,
// 				})
// 				s.audioRegistered = true
// 				LOG.Info("[ProcessPacket] Audio registered",
// 					"pktcodec", packet.CodecID,
// 					"StreamType", stream.GetTSStreamType(packet.CodecID),
// 					"astits", astits.StreamType(stream.GetTSStreamType(s.audioCodec)),
// 				)

// 			}
// 			s.pmtWritten = true

// 			// Write the tables ONLY ONCE when the PMT is successfully built!
// 			s.muxer.WriteTables()
// 		}

// 	} else if !s.pmtWritten {
// 		// Drop all packets until the foundational PMT is established
// 		return nil
// 	}

// 	// --- DISPATCH TO PACKETIZER ---
// 	return s.writePayload(packet)
// }



// // ProcessPacket manages the state of the stream (Synchronous PMT Binding)
// func (s *TSMuxSession) ProcessPacketX(packet stream.StreamPacket) error {

// 	s.FactCheckDiagnostic(&packet)

// 	// --- WARMUP & SNIFFING PHASE ---
// 	if !s.pmtWritten {
// 		s.warmupCount++

// 		if packet.MediaType == stream.MediaTypeVideo && packet.IsKeyFrame {
// 			s.videoCodec = packet.CodecID
// 		} else if packet.MediaType == stream.MediaTypeAudio {
// 			s.audioCodec = packet.CodecID
// 		}

// 		// We only finalize PMT if we have the video codec AND (we found audio OR timed out)
// 		if s.videoCodec != 0 && (s.audioCodec != 0 || s.warmupCount > 60) {
			
// 			LOG.Info("[ProcessPacket] Writing Unified PMT", "VideoCodec", s.videoCodec, "AudioCodec", s.audioCodec)

// 			s.muxer.AddElementaryStream(astits.PMTElementaryStream{
// 				ElementaryPID: VideoPID,
// 				StreamType:    astits.StreamType(stream.GetTSStreamType(s.videoCodec)),
// 			})
// 			s.muxer.SetPCRPID(VideoPID)

// 			if s.audioCodec != 0 {
// 				s.muxer.AddElementaryStream(astits.PMTElementaryStream{
// 					ElementaryPID: AudioPID,
// 					StreamType:    astits.StreamType(stream.GetTSStreamType(s.audioCodec)),
// 				})
// 			}

// 			s.pmtWritten = true
// 			s.muxer.WriteTables()
// 		}
// 		// ALWAYS drop packets during the warmup phase
// 		return nil
// 	}

// 	// --- THE ALIGNMENT GATEKEEPER ---
// 	// Even though the PMT is written, we cannot start sending payloads until the camera 
// 	// generates its NEXT Video Keyframe. If we let Audio or P-frames through first, VLC deadlocks.
// 	// if !s.hasSentFirstIFrame {
// 	// 	if packet.MediaType == stream.MediaTypeVideo && packet.IsKeyFrame {
// 	// 		s.hasSentFirstIFrame = true
// 	// 		LOG.Info("[ProcessPacket] First Keyframe caught, opening floodgates.")
// 	// 	} else {
// 	// 		return nil // Drop EVERYTHING until the Keyframe arrives
// 	// 	}
// 	// }

// 	// --- DISPATCH TO PACKETIZER ---
// 	return s.writePayload(packet)
// }
