package extractor

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type ZipExtractor struct{}

func (e *ZipExtractor) Extract(archivePath, destDir string) (string, error) {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return "", fmt.Errorf("failed to open zip archive: %w", err)
	}
	defer reader.Close()

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create destination directory: %w", err)
	}

	for _, file := range reader.File {
		cleanName := filepath.Clean(file.Name)
		if strings.Contains(cleanName, "..") {
			continue
		}

		target := filepath.Join(destDir, cleanName)

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(target, file.Mode()); err != nil {
				return "", fmt.Errorf("failed to create directory: %w", err)
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return "", fmt.Errorf("failed to create parent directory: %w", err)
		}

		if err := extractZipFile(file, target); err != nil {
			return "", err
		}
	}

	return findJavaHome(destDir)
}

func extractZipFile(file *zip.File, target string) error {
	reader, err := file.Open()
	if err != nil {
		return fmt.Errorf("failed to open zip entry: %w", err)
	}
	defer reader.Close()

	writer, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR|os.O_TRUNC, file.Mode())
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer writer.Close()

	if _, err := io.Copy(writer, reader); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
