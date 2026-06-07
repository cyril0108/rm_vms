package stream

import "github.com/asticode/go-astits"

const FFMpegCodecULaw = 65542;
const FFMpegCodecALaw = 65543;

const StreamTypeHEVC = 0x24;

// Map FFmpeg AVCodecID to MPEG-TS Stream Types
func GetTSStreamType(ffmpegCodecID uint32) astits.StreamType {
	switch ffmpegCodecID {
	case 27: // AV_CODEC_ID_H264
		return astits.StreamTypeH264Video // 0x1b
	case 173: // AV_CODEC_ID_HEVC (H.265)
		return StreamTypeHEVC // astits doesn't have a constant for H.265, but 0x24 is the ISO standard
	case 86018: // AV_CODEC_ID_AAC
		return astits.StreamTypeAACAudio // 0x0f

	// --- G.711 PCM Audio Codecs ---
	case FFMpegCodecULaw: // AV_CODEC_ID_PCM_MULAW (G.711 µ-law)
		return astits.StreamTypePrivateData // Commonly used private stream type for PCM audio in TS
	case FFMpegCodecALaw: // AV_CODEC_ID_PCM_ALAW (G.711 A-law)
		return astits.StreamTypePrivateData

	default:
		return 0 // Unknown
	}
}