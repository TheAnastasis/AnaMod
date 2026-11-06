package logfile

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"anamod/internal/atomicfile"
)

const FileName = "patcher.log"

const MaxSize = 1 << 20

type Log struct {
	mu      sync.Mutex
	file    *os.File
	thisRun strings.Builder
	pending string
}

func Open(dir string) (*Log, error) {
	trim(filepath.Join(dir, FileName))
	file, err := os.OpenFile(filepath.Join(dir, FileName), os.O_CREATE|os.O_APPEND|os.O_WRONLY,
		0o644)
	if err != nil {
		return nil, err
	}
	return &Log{file: file}, nil
}

func trim(path string) {
	info, err := os.Stat(path)
	if err != nil || info.Size() <= MaxSize {
		return
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	data = data[len(data)-MaxSize/2:]
	if newline := bytes.IndexByte(data, '\n'); newline >= 0 {
		data = data[newline+1:]
	}
	atomicfile.Write(path, data)
}

func Memory() *Log {
	return &Log{}
}

func (l *Log) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	lines := strings.Split(l.pending+string(p), "\n")
	l.pending = lines[len(lines)-1]
	var err error
	for _, line := range lines[:len(lines)-1] {
		stamped := time.Now().Format("2006-01-02 15:04:05") + " " + line + "\n"
		l.thisRun.WriteString(stamped)
		if l.file == nil {
			continue
		}
		if _, writeErr := l.file.WriteString(stamped); writeErr != nil && err == nil {
			err = writeErr
		}
	}
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

func (l *Log) Text() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.thisRun.String()
}

func (l *Log) Close() error {
	l.mu.Lock()
	pending := l.pending
	l.mu.Unlock()
	if pending != "" {
		l.Write([]byte("\n"))
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file == nil {
		return nil
	}
	return l.file.Close()
}
