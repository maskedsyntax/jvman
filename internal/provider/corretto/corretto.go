package corretto

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"time"

	"github.com/maskedsyntax/jvman/internal/provider"
)

const (
	githubAPI    = "https://api.github.com/repos/corretto"
	downloadBase = "https://corretto.aws/downloads/resources"
	vendorName   = "corretto"
)

var supportedVersions = []int{23, 22, 21, 17, 11, 8}

type Corretto struct {
	client *http.Client
}

func New() *Corretto {
	return &Corretto{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Corretto) Name() string {
	return vendorName
}

func mapOS() string {
	switch runtime.GOOS {
	case "darwin":
		return "macosx"
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
	default:
		return "x64"
	}
}

func normalizeArch(arch string) string {
	switch arch {
	case "amd64", "x86_64", "x64":
		return "x64"
	case "arm64", "aarch64":
		return "aarch64"
	default:
		return arch
	}
}

type githubRelease struct {
	TagName string `json:"tag_name"`
}

func (c *Corretto) ListAvailableVersions() ([]provider.Release, error) {
	var releases []provider.Release
	for _, v := range supportedVersions {
		releases = append(releases, provider.Release{
			Version: strconv.Itoa(v),
			Vendor:  vendorName,
		})
	}
	return releases, nil
}

func (c *Corretto) GetRelease(version string, opts *provider.Options) (*provider.Release, error) {
	os := mapOS()
	arch := mapArch()
	if opts != nil && opts.Arch != "" {
		arch = normalizeArch(opts.Arch)
	}

	repoName := fmt.Sprintf("corretto-%s", version)
	apiURL := fmt.Sprintf("%s/%s/releases/latest", githubAPI, repoName)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch release info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("version %s not found", version)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	tag := release.TagName
	downloadURL, fileName := buildDownloadURL(tag, os, arch)

	return &provider.Release{
		Version:      version,
		FullVersion:  tag,
		Vendor:       vendorName,
		DownloadURL:  downloadURL,
		Checksum:     "",
		ChecksumType: "",
		FileName:     fileName,
		OS:           os,
		Arch:         arch,
	}, nil
}

func buildDownloadURL(tag, os, arch string) (string, string) {
	var fileName string

	if os == "windows" {
		fileName = fmt.Sprintf("amazon-corretto-%s-windows-%s-jdk.zip", tag, arch)
	} else {
		fileName = fmt.Sprintf("amazon-corretto-%s-%s-%s.tar.gz", tag, os, arch)
	}

	downloadURL := fmt.Sprintf("%s/%s/%s", downloadBase, tag, fileName)
	return downloadURL, fileName
}

func VersionName(version string) string {
	return fmt.Sprintf("%s-%s", vendorName, version)
}
