package shm

import (
	"encoding/binary"
	"errors"
)

type FrameMetadata struct {
	Magic      uint32
	FrameSize  uint32
	EpochMs    uint64
	PTS        int64
	DTS        int64
	CodecID    uint32
	IsKeyFrame uint8
	MediaType  uint8
}

func (fm *FrameMetadata) GetMagic(metaBytes []byte) (uint32) {
	return binary.LittleEndian.Uint32(metaBytes[0:4])
}

func (fm *FrameMetadata) LoadFrom(metaBytes []byte) (error) {

	if len(metaBytes) < MetadataSize {
		return errors.New("incomplete metadata header")
	}

	fm.Magic     = binary.LittleEndian.Uint32(metaBytes[0:4])
	fm.FrameSize = binary.LittleEndian.Uint32(metaBytes[4:8])
	fm.EpochMs   = binary.LittleEndian.Uint64(metaBytes[8:16])

	// Cast uint64 to int64 for the timestamps
	fm.PTS       = int64(binary.LittleEndian.Uint64(metaBytes[16:24]))
	fm.DTS       = int64(binary.LittleEndian.Uint64(metaBytes[24:32]))

	fm.CodecID =    binary.LittleEndian.Uint32(metaBytes[32:36])
	fm.IsKeyFrame = metaBytes[36]
	fm.MediaType =  metaBytes[37]

	return nil

}
