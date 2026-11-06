package xenfile

import (
	"errors"
	"fmt"

	"anamod/internal/align"
)

const (
	bufferStep       = 0x800
	inPlaceTolerance = 277
	endOfBlock       = 256
	firstLength      = 257
	maxCodeLength    = 15
)

var errBadStream = errors.New("not a complete stream")

var (
	lengthBase  = []int{3, 4, 5, 6, 7, 8, 9, 10, 11, 13, 15, 17, 19, 23, 27, 31, 35, 43, 51, 59, 67, 83, 99, 115, 131, 163, 195, 227, 258}
	lengthExtra = []uint{0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 2, 2, 2, 2, 3, 3, 3, 3, 4, 4, 4, 4, 5, 5, 5, 5, 0}
	spanExtra   = []uint{0, 0, 0, 0, 1, 1, 2, 2, 3, 3, 4, 4, 5, 5, 6, 6, 7, 7, 8, 8, 9, 9, 10, 10, 11, 11, 12, 12, 13, 13}
	lengthOrder = []int{16, 17, 18, 0, 8, 7, 9, 6, 10, 5, 11, 4, 12, 3, 13, 2, 14, 1, 15}
)

func CheckInPlace(packed []byte) error {
	written, furthest, err := walk(packed)
	if err != nil {
		return err
	}
	buffer := bufferSize(written)
	if len(packed) > buffer {
		return fmt.Errorf("doesn't fit 360 buffer: %d compressed bytes for %d "+
			"bytes of room", len(packed), buffer)
	}
	if overrun := furthest - (buffer - len(packed)); overrun > inPlaceTolerance {
		return fmt.Errorf("unpacking it on the 360 would overwrite %d bytes it hasn't read "+
			"yet (the disc's own files go up to %d)", overrun, inPlaceTolerance)
	}
	return nil
}

func bufferSize(written int) int {
	return align.Up(written, bufferStep)
}

type bitReader struct {
	data  []byte
	read  int
	bits  uint32
	count uint
}

func (r *bitReader) take(count uint) (int, error) {
	for r.count < count {
		if r.read >= len(r.data) {
			return 0, errBadStream
		}
		r.bits |= uint32(r.data[r.read]) << r.count
		r.read++
		r.count += 8
	}
	value := r.bits & (1<<count - 1)
	r.bits >>= count
	r.count -= count
	return int(value), nil
}

type codes struct {
	lengths [maxCodeLength + 1]int
	symbols []int
}

func newCodes(lengths []int) codes {
	var table codes
	for _, length := range lengths {
		table.lengths[length]++
	}
	table.lengths[0] = 0
	var offsets [maxCodeLength + 2]int
	for length := 1; length <= maxCodeLength; length++ {
		offsets[length+1] = offsets[length] + table.lengths[length]
	}
	table.symbols = make([]int, len(lengths))
	for symbol, length := range lengths {
		if length != 0 {
			table.symbols[offsets[length]] = symbol
			offsets[length]++
		}
	}
	return table
}

func (c codes) decode(r *bitReader) (int, error) {
	code, first, index := 0, 0, 0
	for length := 1; length <= maxCodeLength; length++ {
		bit, err := r.take(1)
		if err != nil {
			return 0, err
		}
		code |= bit
		count := c.lengths[length]
		if code-count < first {
			return c.symbols[index+code-first], nil
		}
		index += count
		first = (first + count) << 1
		code <<= 1
	}
	return 0, errBadStream
}

func fixedCodes() (codes, codes) {
	literals := make([]int, 288)
	for symbol := range literals {
		switch {
		case symbol < 144:
			literals[symbol] = 8
		case symbol < 256:
			literals[symbol] = 9
		case symbol < 280:
			literals[symbol] = 7
		default:
			literals[symbol] = 8
		}
	}
	spans := make([]int, 30)
	for symbol := range spans {
		spans[symbol] = 5
	}
	return newCodes(literals), newCodes(spans)
}

func readCodes(r *bitReader) (codes, codes, error) {
	var sizes [3]int
	for i, count := range []uint{5, 5, 4} {
		size, err := r.take(count)
		if err != nil {
			return codes{}, codes{}, err
		}
		sizes[i] = size
	}
	literalCount, spanCount, lengthCount := sizes[0]+257, sizes[1]+1, sizes[2]+4
	if lengthCount > len(lengthOrder) {
		return codes{}, codes{}, errBadStream
	}
	lengthLengths := make([]int, len(lengthOrder))
	for _, position := range lengthOrder[:lengthCount] {
		length, err := r.take(3)
		if err != nil {
			return codes{}, codes{}, err
		}
		lengthLengths[position] = length
	}
	lengthCodes := newCodes(lengthLengths)
	lengths := make([]int, literalCount+spanCount)
	for filled := 0; filled < len(lengths); {
		symbol, err := lengthCodes.decode(r)
		if err != nil {
			return codes{}, codes{}, err
		}
		if symbol < 16 {
			lengths[filled] = symbol
			filled++
			continue
		}
		value, count, base := 0, uint(3), 11
		switch symbol {
		case 16:
			if filled == 0 {
				return codes{}, codes{}, errBadStream
			}
			value, count, base = lengths[filled-1], 2, 3
		case 17:
			base = 3
		default:
			count = 7
		}
		extra, err := r.take(count)
		if err != nil {
			return codes{}, codes{}, err
		}
		repeat := base + extra
		if filled+repeat > len(lengths) {
			return codes{}, codes{}, errBadStream
		}
		for range repeat {
			lengths[filled] = value
			filled++
		}
	}
	return newCodes(lengths[:literalCount]), newCodes(lengths[literalCount:]), nil
}

type walker struct {
	bitReader
	written  int
	furthest int
}

func (w *walker) note() {
	if w.read < len(w.data) {
		w.furthest = max(w.furthest, w.written-w.read)
	}
}

func (w *walker) storedBlock() error {
	w.bits, w.count = 0, 0
	if w.read+4 > len(w.data) {
		return errBadStream
	}
	size := int(w.data[w.read]) | int(w.data[w.read+1])<<8
	w.read += 4
	if w.read+size > len(w.data) {
		return errBadStream
	}
	w.note()
	w.read += size
	w.written += size
	w.note()
	return nil
}

func (w *walker) codedBlock(literals, spans codes) error {
	for {
		symbol, err := literals.decode(&w.bitReader)
		if err != nil {
			return err
		}
		if symbol == endOfBlock {
			return nil
		}
		if symbol < endOfBlock {
			w.written++
			w.note()
			continue
		}
		symbol -= firstLength
		if symbol >= len(lengthBase) {
			return errBadStream
		}
		extra, err := w.take(lengthExtra[symbol])
		if err != nil {
			return err
		}
		span, err := spans.decode(&w.bitReader)
		if err != nil {
			return err
		}
		if span >= len(spanExtra) {
			return errBadStream
		}
		if _, err := w.take(spanExtra[span]); err != nil {
			return err
		}
		w.written += lengthBase[symbol] + extra
		w.note()
	}
}

func walk(packed []byte) (written, furthest int, err error) {
	w := &walker{bitReader: bitReader{data: packed}, furthest: -len(packed)}
	for {
		final, err := w.take(1)
		if err != nil {
			return 0, 0, err
		}
		kind, err := w.take(2)
		if err != nil {
			return 0, 0, err
		}
		switch kind {
		case 0:
			err = w.storedBlock()
		case 1:
			literals, spans := fixedCodes()
			err = w.codedBlock(literals, spans)
		case 2:
			var literals, spans codes
			if literals, spans, err = readCodes(&w.bitReader); err == nil {
				err = w.codedBlock(literals, spans)
			}
		default:
			err = errBadStream
		}
		if err != nil {
			return 0, 0, err
		}
		if final == 1 {
			return w.written, w.furthest, nil
		}
	}
}
