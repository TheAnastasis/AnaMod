package install

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"anamod/internal/atomicfile"
)

const BackupDir = "ANAMOD_BACKUP"

const RecordName = "record.json"

type Record struct {
	AnamodVersion    string            `json:"anamod_version"`
	Options          []string          `json:"options"`
	Written          map[string]string `json:"written"`
	Added            []string          `json:"added"`
	Backups          map[string]string `json:"backups"`
	RPCS3Dir         string            `json:"rpcs3_dir"`
	OfficialFPSWasOn bool              `json:"official_fps_was_on"`
}

func recordPath(root string) string {
	return filepath.Join(root, BackupDir, RecordName)
}

func ReadRecord(root string) (*Record, error) {
	data, err := os.ReadFile(recordPath(root))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var record Record
	if err := json.Unmarshal(data, &record); err != nil {
		return nil, fmt.Errorf("the install record %s is damaged: %w", recordPath(root), err)
	}
	return &record, nil
}

func WriteRecord(root string, record *Record) error {
	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return err
	}
	return atomicfile.Write(recordPath(root), data)
}

func fileSHA256(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func bytesSHA256(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func gamePath(root, relative string) string {
	return filepath.Join(root, filepath.FromSlash(relative))
}

func backupPath(root, relative string) string {
	return filepath.Join(root, BackupDir, "files", filepath.FromSlash(relative))
}
