package extractor

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type TarExtractor struct{}

func (e *TarExtractor) Extract(archivePath, destDir string) (string, error) {
	file, err := os.Open(archivePath)
	if err != nil {
		return "", fmt.Errorf("failed to open archive: %w", err)
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return "", fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create destination directory: %w", err)
	}

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("failed to read tar entry: %w", err)
		}

		cleanName := filepath.Clean(header.Name)
		if strings.Contains(cleanName, "..") {
			continue
		}

		target := filepath.Join(destDir, cleanName)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return "", fmt.Errorf("failed to create directory: %w", err)
			}

		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return "", fmt.Errorf("failed to create parent directory: %w", err)
			}

			outFile, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return "", fmt.Errorf("failed to create file: %w", err)
			}

			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return "", fmt.Errorf("failed to write file: %w", err)
			}
			outFile.Close()

		case tar.TypeSymlink:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return "", fmt.Errorf("failed to create parent directory for symlink: %w", err)
			}
			os.Remove(target)
			if err := os.Symlink(header.Linkname, target); err != nil {
				return "", fmt.Errorf("failed to create symlink: %w", err)
			}

		case tar.TypeLink:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return "", fmt.Errorf("failed to create parent directory for link: %w", err)
			}
			linkTarget := filepath.Join(destDir, header.Linkname)
			os.Remove(target)
			if err := os.Link(linkTarget, target); err != nil {
				return "", fmt.Errorf("failed to create hard link: %w", err)
			}
		}
	}

	return findJavaHome(destDir)
}
