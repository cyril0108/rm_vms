package transmux

import (
	"context"
	"net/http"

	"github.com/asticode/go-astits"

	"nvr_core/logger"
	"nvr_core/stream"
)

var LOG = logger.NewLogger("[nvr_core]","[transmux]")


const (
	VideoPID   uint16 = 256
	AudioPID   uint16 = 257
	VideoPESID uint8  = 224 // 0xE0
	AudioPESID uint8  = 192 // 0xC0
)

// =====================================================================
//  MUXER STATE MACHINE: Handles Dynamic PMT, PIDs, and Packetization
// =====================================================================

type TSMuxSession struct {
	muxer           *astits.Muxer
	firstPTS        int64 // Track the baseline timestamp for this session
	hasFirstPTS     bool  // Flag to know if we've locked the baseline
	pmtWritten      bool
	audioRegistered bool
	audioCodec      uint32
	flusher         http.Flusher
}

func NewTSMuxSession(ctx context.Context, w http.ResponseWriter) *TSMuxSession {
	flusher, _ := w.(http.Flusher)
	return &TSMuxSession{
		muxer:     astits.NewMuxer(ctx, w),
		flusher:   flusher,
	}
}

// ProcessPacket manages the state of the stream (Dynamic PMT Binding)
func (s *TSMuxSession) ProcessPacket(packet stream.StreamPacket) error {

	// log.Printf("[TSMuxSession] ProcessPacket: %d", len(packet.Payload))

	// --- LATE AUDIO BINDING LOGIC ---
	if packet.MediaType == stream.MediaTypeAudio {
		if !s.audioRegistered {
			s.audioCodec = packet.CodecID
			if s.pmtWritten {
				// Audio arrived AFTER video keyframe. Inject dynamically.
				s.muxer.AddElementaryStream(astits.PMTElementaryStream{
					ElementaryPID: AudioPID,
					StreamType:    astits.StreamType(stream.GetTSStreamType(s.audioCodec)),
				})
				s.audioRegistered = true
				s.muxer.WriteTables()
			} else {
				return nil // PMT not written yet, drop audio packet but keep codec saved
			}
		}
	}

	// log.Printf("[TSMuxSession] audio passed")

	// --- VIDEO KEYFRAME / INITIAL BINDING LOGIC ---
	if packet.MediaType == stream.MediaTypeVideo && packet.IsKeyFrame {
		if !s.pmtWritten {

			LOG.Info("[ProcessPacket] ", "VideoCodec", stream.GetTSStreamType(packet.CodecID))
			//  Bind Video
			s.muxer.AddElementaryStream(astits.PMTElementaryStream{
				ElementaryPID: VideoPID,
				StreamType:    astits.StreamType(stream.GetTSStreamType(packet.CodecID)),
			})
			s.muxer.SetPCRPID(VideoPID)

			//  Bind Audio (If it arrived before this keyframe)
			if s.audioCodec != 0 && !s.audioRegistered {
				s.muxer.AddElementaryStream(astits.PMTElementaryStream{
					ElementaryPID: AudioPID,
					StreamType:    astits.StreamType(stream.GetTSStreamType(s.audioCodec)),
				})
				s.audioRegistered = true
			}
			s.pmtWritten = true
		}

		// Force table rewrite on every keyframe for VLC stability
		s.muxer.WriteTables()

		// log.Printf("[TSMuxSession] video")

	} else if !s.pmtWritten {

		// log.Printf("[TSMuxSession] no pmt")

		// Drop all packets until the foundational PMT is established
		return nil
	}

	// --- DISPATCH TO PACKETIZER ---
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
		streamID = AudioPESID
		targetPID = AudioPID
	} else {
		return nil // Ignore unknown media types
	}

	// --- TIMESTAMP NORMALIZATION FIX ---
	if !s.hasFirstPTS {
		s.firstPTS = packet.PTS
		s.hasFirstPTS = true
	}

	// Subtract the baseline so this specific HTTP connection starts near 0.
	// We add 90000 (exactly 1 second) as a safety buffer. If the stream contains B-frames, 
	// DTS can sometimes be lower than PTS. Starting at 1 second prevents DTS from dipping into negative numbers.
	normalizedPTS := (packet.PTS - s.firstPTS)
	normalizedDTS := (packet.DTS - s.firstPTS)

	// Failsafe: If your C++ worker isn't explicitly calculating DTS yet, 
	// H.264/MPEG-TS requires it to be present. For Baseline profiles, DTS == PTS.
	if packet.DTS == 0 && packet.PTS != 0 {
		normalizedDTS = normalizedPTS
	}

	// Package the raw frame into a PES
	pes := &astits.PESData{
		Header: &astits.PESHeader{
			OptionalHeader: &astits.PESOptionalHeader{
				PTS: &astits.ClockReference{Base: normalizedPTS},
				DTS: &astits.ClockReference{Base: normalizedDTS},
			},
			StreamID: streamID,
		},
		Data: packet.Payload,
	}

	muxerData := &astits.MuxerData{
		PES: pes,
		PID: targetPID,
	}

	// Inject the Master Clock (PCR) on video frames
	if isVideo {
		muxerData.AdaptationField = &astits.PacketAdaptationField{
			HasPCR: true,
			PCR:    &astits.ClockReference{Base: normalizedPTS},
		}
	}

	// Write the data to the underlying io.Writer (HTTP Response)
	if _, err := s.muxer.WriteData(muxerData); err != nil {
		return err
	}

	if isVideo && s.flusher != nil {
		s.flusher.Flush()
	}

	return nil
}