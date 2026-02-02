package zulu

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"runtime"
	"sort"
	"strconv"
	"time"

	"github.com/maskedsyntax/jvman/internal/provider"
)

const (
	baseURL    = "https://api.azul.com/metadata/v1/zulu/packages/"
	vendorName = "zulu"
)

type Zulu struct {
	client *http.Client
}

func New() *Zulu {
	return &Zulu{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (z *Zulu) Name() string {
	return vendorName
}

func mapOS() string {
	switch runtime.GOOS {
	case "darwin":
		return "macos"
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
		return "i686"
	default:
		return "x64"
	}
}

func archiveType() string {
	if runtime.GOOS == "windows" {
		return "zip"
	}
	return "tar.gz"
}

type zuluPackage struct {
	DownloadURL    string `json:"download_url"`
	Name           string `json:"name"`
	Sha256Hash     string `json:"sha256_hash"`
	JavaVersion    []int  `json:"java_version"`
	ZuluVersion    []int  `json:"zulu_version"`
	LatestInChain  bool   `json:"latest"`
}

func (z *Zulu) ListAvailableVersions() ([]provider.Release, error) {
	params := url.Values{}
	params.Set("os", mapOS())
	params.Set("arch", mapArch())
	params.Set("archive_type", archiveType())
	params.Set("java_package_type", "jdk")
	params.Set("javafx_bundled", "false")
	params.Set("release_status", "ga")
	params.Set("availability_types", "CA")
	params.Set("page_size", "100")

	reqURL := baseURL + "?" + params.Encode()

	resp, err := z.client.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch available releases: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var packages []zuluPackage
	if err := json.NewDecoder(resp.Body).Decode(&packages); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	versionSet := make(map[int]bool)
	for _, pkg := range packages {
		if len(pkg.JavaVersion) > 0 {
			versionSet[pkg.JavaVersion[0]] = true
		}
	}

	var versions []int
	for v := range versionSet {
		versions = append(versions, v)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(versions)))

	var releases []provider.Release
	for _, v := range versions {
		releases = append(releases, provider.Release{
			Version: strconv.Itoa(v),
			Vendor:  vendorName,
		})
	}

	return releases, nil
}

func (z *Zulu) GetRelease(version string) (*provider.Release, error) {
	params := url.Values{}
	params.Set("os", mapOS())
	params.Set("arch", mapArch())
	params.Set("archive_type", archiveType())
	params.Set("java_package_type", "jdk")
	params.Set("javafx_bundled", "false")
	params.Set("release_status", "ga")
	params.Set("availability_types", "CA")
	params.Set("java_version", version)
	params.Set("latest", "true")
	params.Set("page_size", "1")

	reqURL := baseURL + "?" + params.Encode()

	resp, err := z.client.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch release info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var packages []zuluPackage
	if err := json.NewDecoder(resp.Body).Decode(&packages); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(packages) == 0 {
		return nil, fmt.Errorf("no releases found for version %s", version)
	}

	pkg := packages[0]

	fullVersion := fmt.Sprintf("zulu%d.%d.%d-ca-jdk%d.%d.%d",
		safeIndex(pkg.ZuluVersion, 0),
		safeIndex(pkg.ZuluVersion, 1),
		safeIndex(pkg.ZuluVersion, 2),
		safeIndex(pkg.JavaVersion, 0),
		safeIndex(pkg.JavaVersion, 1),
		safeIndex(pkg.JavaVersion, 2),
	)

	return &provider.Release{
		Version:      version,
		FullVersion:  fullVersion,
		Vendor:       vendorName,
		DownloadURL:  pkg.DownloadURL,
		Checksum:     pkg.Sha256Hash,
		ChecksumType: "sha256",
		FileName:     pkg.Name,
		OS:           mapOS(),
		Arch:         mapArch(),
	}, nil
}

func safeIndex(slice []int, index int) int {
	if index < len(slice) {
		return slice[index]
	}
	return 0
}

func VersionName(version string) string {
	return fmt.Sprintf("%s-%s", vendorName, version)
}
