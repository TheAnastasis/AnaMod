package update

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const APIURLTemplate = "https://api.github.com/repos/%s/releases/latest"

const DataAsset = "anamod-data.zip"

const maxDownload = 200 << 20

type Release struct {
	Version string
	PageURL string
	DataURL string
	Body    string
}

func Latest(client *http.Client, apiURL string) (Release, error) {
	body, err := get(client, apiURL, "application/vnd.github+json")
	if err != nil {
		return Release{}, err
	}
	var answer struct {
		TagName string `json:"tag_name"`
		HTMLURL string `json:"html_url"`
		Body    string `json:"body"`
		Assets  []struct {
			Name string `json:"name"`
			URL  string `json:"browser_download_url"`
		} `json:"assets"`
	}
	if err := json.Unmarshal(body, &answer); err != nil {
		return Release{}, fmt.Errorf("an issue occurred downloading from GitHub: %w", err)
	}
	release := Release{
		Version: strings.TrimPrefix(answer.TagName, "v"),
		PageURL: answer.HTMLURL,
		Body:    answer.Body,
	}
	for _, asset := range answer.Assets {
		if asset.Name == DataAsset {
			release.DataURL = asset.URL
		}
	}
	if release.Version == "" || release.DataURL == "" {
		return Release{}, errors.New("the latest release has no " + DataAsset)
	}
	return release, nil
}

func Download(client *http.Client, url string) ([]byte, error) {
	return get(client, url, "application/octet-stream")
}

func get(client *http.Client, url, accept string) ([]byte, error) {
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", accept)
	request.Header.Set("User-Agent", "AnaModPatcher")
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s answered %s", url, response.Status)
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, maxDownload+1))
	if err != nil {
		return nil, err
	}
	if len(body) > maxDownload {
		return nil, fmt.Errorf("%s is larger than expected", url)
	}
	return body, nil
}
