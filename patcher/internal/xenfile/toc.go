package xenfile

import (
	"bytes"
	"cmp"
	"encoding/binary"
	"errors"
	"fmt"
	"slices"

	"anamod/internal/align"
)

const (
	tocMagic      = 0x544F4331
	tocHeaderSize = 0x20
	tocRowSize    = 16
	tocSector     = 0x800
)

type TocRow struct {
	Key      uint32
	Inflated uint32
	Stored   uint32
	Extra    uint32
}

func ReadTOC(table []byte) ([]TocRow, error) {
	if len(table) < tocHeaderSize || binary.BigEndian.Uint32(table) != tocMagic {
		return nil, errors.New("this is not the 360's table of compressed files...")
	}
	count := int(binary.BigEndian.Uint32(table[4:]))
	if size := binary.BigEndian.Uint32(table[8:]); size != tocHeaderSize {
		return nil, fmt.Errorf("the table's header is 0x%X bytes, not 0x20", size)
	}
	if tocHeaderSize+count*tocRowSize > len(table) {
		return nil, fmt.Errorf("the table says it has %d rows, but the file is too short? send logs", count)
	}
	rows := make([]TocRow, count)
	for i := range rows {
		words := table[tocHeaderSize+i*tocRowSize:]
		rows[i] = TocRow{
			Key:      binary.BigEndian.Uint32(words),
			Inflated: binary.BigEndian.Uint32(words[4:]),
			Stored:   binary.BigEndian.Uint32(words[8:]),
			Extra:    binary.BigEndian.Uint32(words[12:]),
		}
	}
	return rows, nil
}

func UpdateTOC(table []byte, changes []TocRow) ([]byte, error) {
	rows, err := ReadTOC(table)
	if err != nil {
		return nil, err
	}
	for _, change := range changes {
		index := slices.IndexFunc(rows, func(row TocRow) bool { return row.Key == change.Key })
		if index < 0 {
			rows = append(rows, TocRow{Key: change.Key, Inflated: change.Inflated,
				Stored: change.Stored})
			continue
		}
		rows[index].Inflated, rows[index].Stored = change.Inflated, change.Stored
	}
	slices.SortFunc(rows, func(a, b TocRow) int { return cmp.Compare(a.Key, b.Key) })
	var out bytes.Buffer
	out.Write(table[:tocHeaderSize])
	binary.BigEndian.PutUint32(out.Bytes()[4:], uint32(len(rows)))
	for _, row := range rows {
		binary.Write(&out, binary.BigEndian, row)
	}
	out.Write(make([]byte, align.Up(out.Len(), tocSector)-out.Len()))
	return out.Bytes(), nil
}
