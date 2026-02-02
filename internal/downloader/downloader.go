package downloader

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/schollz/progressbar/v3"
)

type Downloader struct {
	client *retryablehttp.Client
}

func New() *Downloader {
	client := retryablehttp.NewClient()
	client.RetryMax = 3
	client.RetryWaitMin = 1 * time.Second
	client.RetryWaitMax = 5 * time.Second
	client.Logger = nil

	return &Downloader{
		client: client,
	}
}

type DownloadResult struct {
	FilePath string
	Checksum string
}

func (d *Downloader) Download(url, destDir, filename string, expectedChecksum string) (*DownloadResult, error) {
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create destination directory: %w", err)
	}

	destPath := filepath.Join(destDir, filename)

	resp, err := d.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	file, err := os.Create(destPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	bar := progressbar.NewOptions64(
		resp.ContentLength,
		progressbar.OptionSetDescription("Downloading"),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(40),
		progressbar.OptionThrottle(65*time.Millisecond),
		progressbar.OptionShowCount(),
		progressbar.OptionOnCompletion(func() {
			fmt.Fprint(os.Stderr, "\n")
		}),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetRenderBlankState(true),
	)

	hash := sha256.New()
	writer := io.MultiWriter(file, hash, bar)

	if _, err := io.Copy(writer, resp.Body); err != nil {
		os.Remove(destPath)
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	checksum := hex.EncodeToString(hash.Sum(nil))

	if expectedChecksum != "" && checksum != expectedChecksum {
		os.Remove(destPath)
		return nil, fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, checksum)
	}

	return &DownloadResult{
		FilePath: destPath,
		Checksum: checksum,
	}, nil
}
