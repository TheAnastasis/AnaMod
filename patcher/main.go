package main

import (
	_ "embed"
	"errors"
	"log"

	"anamod/ui"
	"fyne.io/fyne/v2/app"
	"github.com/BurntSushi/toml"
)

const UpdateRepo = "TheAnastasis/AnaMod"

//go:embed data/anamod-data.zip
var dataPack []byte

//go:embed FyneApp.toml
var appMetadata []byte

func patcherVersion(metadata []byte) (string, error) {
	var parsed struct {
		Details struct {
			Version string
		}
	}
	if err := toml.Unmarshal(metadata, &parsed); err != nil {
		return "", err
	}
	if parsed.Details.Version == "" {
		return "", errors.New("FyneApp.toml has no Version")
	}
	return parsed.Details.Version, nil
}

func main() {
	version, err := patcherVersion(appMetadata)
	if err != nil {
		log.Fatal(err)
	}
	ui.Run(app.NewWithID("net.anamod.patcher"), dataPack, version, UpdateRepo)
}
