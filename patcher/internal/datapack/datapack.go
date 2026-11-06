package datapack

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

const SupportedFormat = 1

const ManifestName = "manifest.json"

var ErrNeedsNewerPatcher = errors.New("this data pack needs a newer patcher. Download the " +
	"latest AnaMod Patcher")

type Manifest struct {
	Format        int    `json:"format"`
	AnamodVersion string `json:"anamod_version"`
	MinPatcher    string `json:"min_patcher"`
	Games         []Game `json:"games"`
}

type Game struct {
	ID           string        `json:"id"`
	Platform     string        `json:"platform"`
	TitleID      string        `json:"title_id"`
	Name         string        `json:"name"`
	Eboot        EbootInfo     `json:"eboot"`
	Changed      []ChangedFile `json:"changed"`
	Added        []AddedFile   `json:"added"`
	RPCS3        RPCS3Info     `json:"rpcs3"`
	InstallNotes []string      `json:"install_notes"`
}

func (g Game) OptionIDs() []string {
	var ids []string
	for _, option := range g.RPCS3.Options {
		ids = append(ids, option.ID)
	}
	return ids
}

type EbootInfo struct {
	Path   string   `json:"path"`
	SHA256 []string `json:"sha256"`
}

type ChangedFile struct {
	Path          string `json:"path"`
	VanillaSHA256 string `json:"vanilla_sha256"`
	ResultSHA256  string `json:"result_sha256"`
	Diff          string `json:"diff"`
	CacheRecord   string `json:"cache_record"`
	Compressed    bool   `json:"compressed"`
	TocKey        uint32 `json:"toc_key"`
	Table         bool   `json:"table"`
}

type AddedFile struct {
	Path       string `json:"path"`
	SHA256     string `json:"sha256"`
	File       string `json:"file"`
	Compressed bool   `json:"compressed"`
	TocKey     uint32 `json:"toc_key"`
}

type RPCS3Info struct {
	PPUHash    string   `json:"ppu_hash"`
	PatchFile  string   `json:"patch_file"`
	Title      string   `json:"title"`
	Serial     string   `json:"serial"`
	AppVersion string   `json:"app_version"`
	Options    []Option `json:"options"`
}

type Option struct {
	ID            string `json:"id"`
	Label         string `json:"label"`
	Patch         string `json:"patch"`
	FallbackPatch string `json:"fallback_patch"`
	Help          string `json:"help"`
}

type Pack struct {
	Manifest Manifest
	files    *zip.Reader
}

func Open(zipData []byte) (*Pack, error) {
	files, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return nil, fmt.Errorf("the data pack is probably damaged: %w", err)
	}
	data, err := readMember(files, ManifestName)
	if err != nil {
		return nil, fmt.Errorf("the data pack has no manifest?: %w", err)
	}
	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("the data pack's manifest was modified: %w", err)
	}
	if manifest.Format > SupportedFormat {
		return nil, ErrNeedsNewerPatcher
	}
	return &Pack{Manifest: manifest, files: files}, nil
}

func (p *Pack) ReadFile(name string) ([]byte, error) {
	data, err := readMember(p.files, name)
	if err != nil {
		return nil, fmt.Errorf("the data pack has no %s: %w", name, err)
	}
	return data, nil
}

func readMember(files *zip.Reader, name string) ([]byte, error) {
	file, err := files.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return io.ReadAll(file)
}

func CompareVersions(a, b string) int {
	aParts := strings.Split(a, ".")
	bParts := strings.Split(b, ".")
	for i := range max(len(aParts), len(bParts)) {
		aNumber, bNumber := versionPart(aParts, i), versionPart(bParts, i)
		if aNumber < bNumber {
			return -1
		}
		if aNumber > bNumber {
			return 1
		}
	}
	return 0
}

func versionPart(parts []string, i int) int {
	if i >= len(parts) {
		return 0
	}
	number, _ := strconv.Atoi(parts[i])
	return number
}
