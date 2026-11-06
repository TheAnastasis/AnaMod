package xex

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)

const (
	headerCountOffset = 0x14
	headerListOffset  = 0x18
	executionInfoKey  = 0x00040006
	titleIDOffset     = 0x0C
)

var ErrNotXex = errors.New("this is not a 360 executable")

func word(data []byte, offset uint32) (uint32, error) {
	if uint64(offset)+4 > uint64(len(data)) {
		return 0, fmt.Errorf("the executable is cut short at 0x%X", offset)
	}
	return binary.BigEndian.Uint32(data[offset:]), nil
}

func headers(data []byte) (map[uint32]uint32, error) {
	if len(data) < headerListOffset || !bytes.Equal(data[:4], []byte("XEX2")) {
		return nil, ErrNotXex
	}
	count := binary.BigEndian.Uint32(data[headerCountOffset:])
	found := map[uint32]uint32{}
	for i := range count {
		key, err := word(data, headerListOffset+i*8)
		if err != nil {
			return nil, err
		}
		value, err := word(data, headerListOffset+i*8+4)
		if err != nil {
			return nil, err
		}
		found[key] = value
	}
	return found, nil
}

func TitleID(data []byte) (string, error) {
	found, err := headers(data)
	if err != nil {
		return "", err
	}
	info, ok := found[executionInfoKey]
	if !ok {
		return "", errors.New("this Xbox 360 executable has no title ID")
	}
	id, err := word(data, info+titleIDOffset)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%08X", id), nil
}
