package cache

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"anamod/internal/align"
	"anamod/internal/atomicfile"
)

const (
	headerSize = 12
	recordSize = 256
	nameSize   = 244
	sizeOffset = 252
	blockSize  = 0x800
	indexPath  = "DATA/D1848E4C.CHK"
)

func Refresh(devHDD0, titleID, relative, recordName string, content []byte) (bool, error) {
	usrdir := filepath.Join(devHDD0, "game", titleID, "USRDIR")
	cached := filepath.Join(usrdir, filepath.FromSlash(relative))
	info, err := os.Stat(cached)
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	notAsExpected := fmt.Errorf("RPCS3's hard drive cache (%s) isn't laid out as expected?",
		cached)
	index, err := os.ReadFile(filepath.Join(usrdir, filepath.FromSlash(indexPath)))
	if err != nil {
		return false, notAsExpected
	}
	start := findRecord(index, recordName)
	slot := int(info.Size())
	if start < 0 || int(binary.BigEndian.Uint32(index[start+sizeOffset:])) != slot {
		return false, notAsExpected
	}
	size := max(slot, align.Up(len(content), blockSize))
	padded := make([]byte, size)
	copy(padded, content)
	if err := atomicfile.Write(cached, padded); err != nil {
		return false, fmt.Errorf("could not write %s: %w", cached, err)
	}
	if size != slot {
		binary.BigEndian.PutUint32(index[start+sizeOffset:], uint32(size))
		if err := atomicfile.Write(filepath.Join(usrdir, filepath.FromSlash(indexPath)),
			index); err != nil {
			return false, fmt.Errorf("could not write the cache index: %w", err)
		}
	}
	return true, nil
}

func findRecord(index []byte, name string) int {
	if len(index) < headerSize {
		return -1
	}
	count := int(binary.BigEndian.Uint32(index[4:]))
	for number := range count {
		start := headerSize + number*recordSize
		if start+recordSize > len(index) {
			return -1
		}
		text := index[start : start+nameSize]
		if end := bytes.IndexByte(text, 0); end >= 0 {
			text = text[:end]
		}
		if string(text) == name {
			return start
		}
	}
	return -1
}
