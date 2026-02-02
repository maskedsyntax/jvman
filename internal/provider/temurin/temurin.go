package temurin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/maskedsyntax/jvman/internal/provider"
)

const (
	baseURL    = "https://api.adoptium.net/v3"
	vendorName = "temurin"
)

type Temurin struct {
	client *http.Client
}

func New() *Temurin {
	return &Temurin{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (t *Temurin) Name() string {
	return vendorName
}

func mapOS() string {
	switch runtime.GOOS {
	case "darwin":
		return "mac"
	case "windows":
		return "windows"
	default:
		return "linux"
	}
}

func mapArch() string {
	switch runtime.GOARCH {
	case "amd64":
		return "x64"
	case "arm64":
		return "aarch64"
	case "386":
		return "x32"
	default:
		return "x64"
	}
}

type availableRelease struct {
	Versions []int `json:"available_releases"`
	LTS      []int `json:"available_lts_releases"`
}

type packageInfo struct {
	Checksum     string `json:"checksum"`
	ChecksumLink string `json:"checksum_link"`
	DownloadURL  string `json:"link"`
	Name         string `json:"name"`
	Size         int64  `json:"size"`
}

type binaryInfo struct {
	Architecture string      `json:"architecture"`
	OS           string      `json:"os"`
	ImageType    string      `json:"image_type"`
	Package      packageInfo `json:"package"`
}

type releaseInfo struct {
	Binary      binaryInfo `json:"binary"`
	ReleaseName string     `json:"release_name"`
	Version     struct {
		Major    int    `json:"major"`
		Minor    int    `json:"minor"`
		Security int    `json:"security"`
		Semver   string `json:"semver"`
	} `json:"version"`
}

func (t *Temurin) ListAvailableVersions() ([]provider.Release, error) {
	url := fmt.Sprintf("%s/info/available_releases", baseURL)
	resp, err := t.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch available releases: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var available availableRelease
	if err := json.NewDecoder(resp.Body).Decode(&available); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	sort.Sort(sort.Reverse(sort.IntSlice(available.Versions)))

	var releases []provider.Release
	for _, v := range available.Versions {
		releases = append(releases, provider.Release{
			Version: strconv.Itoa(v),
			Vendor:  vendorName,
		})
	}

	return releases, nil
}

func (t *Temurin) GetRelease(version string) (*provider.Release, error) {
	os := mapOS()
	arch := mapArch()

	url := fmt.Sprintf(
		"%s/assets/latest/%s/hotspot?architecture=%s&image_type=jdk&os=%s&vendor=eclipse",
		baseURL, version, arch, os,
	)

	resp, err := t.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch release info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("version %s not found for %s/%s", version, os, arch)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var releases []releaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(releases) == 0 {
		return nil, fmt.Errorf("no releases found for version %s", version)
	}

	release := releases[0]

	return &provider.Release{
		Version:      version,
		FullVersion:  release.ReleaseName,
		Vendor:       vendorName,
		DownloadURL:  release.Binary.Package.DownloadURL,
		Checksum:     release.Binary.Package.Checksum,
		ChecksumType: "sha256",
		FileName:     release.Binary.Package.Name,
		OS:           release.Binary.OS,
		Arch:         release.Binary.Architecture,
	}, nil
}

func VersionName(version string) string {
	return fmt.Sprintf("%s-%s", vendorName, strings.TrimPrefix(version, "jdk-"))
}
