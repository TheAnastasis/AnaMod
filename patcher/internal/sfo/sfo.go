package sfo

import (
	"bytes"
	"encoding/binary"
	"errors"
)

const (
	headerSize = 0x14
	entrySize  = 16
)

var (
	ErrNotSFO    = errors.New("PARAM.SFO is damaged or is not the real file.")
	ErrNoTitleID = errors.New("PARAM.SFO has no title ID????")
)

func TitleID(data []byte) (string, error) {
	if len(data) < headerSize || !bytes.Equal(data[:4], []byte("\x00PSF")) {
		return "", ErrNotSFO
	}
	keyStart := int(binary.LittleEndian.Uint32(data[0x08:]))
	dataStart := int(binary.LittleEndian.Uint32(data[0x0C:]))
	count := int(binary.LittleEndian.Uint32(data[0x10:]))
	for i := range count {
		entry, ok := slice(data, headerSize+i*entrySize, entrySize)
		if !ok {
			return "", ErrNotSFO
		}
		key, ok := zeroEnded(data, keyStart+int(binary.LittleEndian.Uint16(entry[0:])))
		if !ok {
			return "", ErrNotSFO
		}
		if key != "TITLE_ID" {
			continue
		}
		length := int(binary.LittleEndian.Uint32(entry[4:]))
		value, ok := slice(data, dataStart+int(binary.LittleEndian.Uint32(entry[12:])), length)
		if !ok {
			return "", ErrNotSFO
		}
		return string(bytes.TrimRight(value, "\x00")), nil
	}
	return "", ErrNoTitleID
}

func slice(data []byte, start, length int) ([]byte, bool) {
	if start < 0 || length < 0 || start > len(data) || length > len(data)-start {
		return nil, false
	}
	return data[start : start+length], true
}

func zeroEnded(data []byte, start int) (string, bool) {
	if start < 0 || start >= len(data) {
		return "", false
	}
	end := bytes.IndexByte(data[start:], 0)
	if end < 0 {
		return "", false
	}
	return string(data[start : start+end]), true
}
