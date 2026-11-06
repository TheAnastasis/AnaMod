package xenfile

import (
	"bytes"
	"compress/flate"
	"fmt"
	"io"
)

func Inflate(data []byte) ([]byte, error) {
	reader := flate.NewReader(bytes.NewReader(data))
	defer reader.Close()
	inflated, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("this file is not compressed the typical way the 360 disc compresses files: %w", err)
	}
	return inflated, nil
}

func Deflate(data []byte) ([]byte, error) {
	var out bytes.Buffer
	writer, err := flate.NewWriter(&out, flate.BestCompression)
	if err != nil {
		return nil, err
	}
	if _, err := writer.Write(data); err != nil {
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
