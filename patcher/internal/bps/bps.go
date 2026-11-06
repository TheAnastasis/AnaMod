package bps

import (
	"bytes"
	"encoding/binary"
	"errors"
	"hash/crc32"
)

const (
	sourceRead = iota
	targetRead
	sourceCopy
	targetCopy
)

const footerSize = 12

var (
	ErrWrongSource = errors.New("the patch was made from a different source file")
	ErrDamaged     = errors.New("AnaMod's patch data is damaged. Download the patcher again")
)

type reader struct {
	patch []byte
	pos   int
	end   int
}

func (r *reader) number() (uint64, error) {
	value, shift := uint64(0), uint64(1)
	for {
		if r.pos >= r.end || shift > 1<<56 {
			return 0, ErrDamaged
		}
		x := r.patch[r.pos]
		r.pos++
		value += uint64(x&0x7f) * shift
		if x&0x80 != 0 {
			return value, nil
		}
		shift <<= 7
		value += shift
	}
}

func (r *reader) offset() (int, error) {
	n, err := r.number()
	if err != nil {
		return 0, err
	}
	if n>>1 > 1<<40 {
		return 0, ErrDamaged
	}
	if n&1 == 1 {
		return -int(n >> 1), nil
	}
	return int(n >> 1), nil
}

func Apply(source, patch []byte) ([]byte, error) {
	if len(patch) < 4+3+footerSize || !bytes.Equal(patch[:4], []byte("BPS1")) {
		return nil, ErrDamaged
	}
	footer := patch[len(patch)-footerSize:]
	sourceCRC := binary.LittleEndian.Uint32(footer[0:])
	targetCRC := binary.LittleEndian.Uint32(footer[4:])
	patchCRC := binary.LittleEndian.Uint32(footer[8:])
	if crc32.ChecksumIEEE(patch[:len(patch)-4]) != patchCRC {
		return nil, ErrDamaged
	}
	r := &reader{patch: patch, pos: 4, end: len(patch) - footerSize}
	sourceSize, err := r.number()
	if err != nil {
		return nil, err
	}
	targetSize, err := r.number()
	if err != nil {
		return nil, err
	}
	metadataSize, err := r.number()
	if err != nil {
		return nil, err
	}
	if metadataSize > uint64(r.end-r.pos) {
		return nil, ErrDamaged
	}
	r.pos += int(metadataSize)
	if sourceSize != uint64(len(source)) || crc32.ChecksumIEEE(source) != sourceCRC {
		return nil, ErrWrongSource
	}
	if targetSize > 1<<32 {
		return nil, ErrDamaged
	}
	size := int(targetSize)
	out := make([]byte, 0, size)
	sourcePos, targetPos := 0, 0
	for r.pos < r.end {
		action, err := r.number()
		if err != nil {
			return nil, err
		}
		mode := int(action & 3)
		if action>>2 >= uint64(size) {
			return nil, ErrDamaged
		}
		length := int(action>>2) + 1
		if len(out)+length > size {
			return nil, ErrDamaged
		}
		switch mode {
		case sourceRead:
			if len(out)+length > len(source) {
				return nil, ErrDamaged
			}
			out = append(out, source[len(out):len(out)+length]...)
		case targetRead:
			if length > r.end-r.pos {
				return nil, ErrDamaged
			}
			out = append(out, patch[r.pos:r.pos+length]...)
			r.pos += length
		case sourceCopy:
			move, err := r.offset()
			if err != nil {
				return nil, err
			}
			sourcePos += move
			if sourcePos < 0 || sourcePos+length > len(source) {
				return nil, ErrDamaged
			}
			out = append(out, source[sourcePos:sourcePos+length]...)
			sourcePos += length
		case targetCopy:
			move, err := r.offset()
			if err != nil {
				return nil, err
			}
			targetPos += move
			if targetPos < 0 || targetPos >= len(out) {
				return nil, ErrDamaged
			}
			for range length {
				out = append(out, out[targetPos])
				targetPos++
			}
		}
	}
	if len(out) != size || crc32.ChecksumIEEE(out) != targetCRC {
		return nil, ErrDamaged
	}
	return out, nil
}
