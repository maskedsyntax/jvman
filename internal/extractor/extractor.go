package extractor

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type Extractor interface {
	Extract(archivePath, destDir string) (string, error)
}

func New() Extractor {
	if runtime.GOOS == "windows" {
		return &ZipExtractor{}
	}
	return &TarExtractor{}
}

func ForFile(filename string) Extractor {
	lower := strings.ToLower(filename)
	if strings.HasSuffix(lower, ".zip") {
		return &ZipExtractor{}
	}
	return &TarExtractor{}
}

func findJavaHome(extractedDir string) (string, error) {
	entries, err := os.ReadDir(extractedDir)
	if err != nil {
		return "", err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		candidatePath := filepath.Join(extractedDir, entry.Name())

		if hasJavaBinary(candidatePath) {
			return candidatePath, nil
		}

		contentsHome := filepath.Join(candidatePath, "Contents", "Home")
		if hasJavaBinary(contentsHome) {
			return contentsHome, nil
		}
	}

	if hasJavaBinary(extractedDir) {
		return extractedDir, nil
	}

	return "", fmt.Errorf("could not find java binary in extracted contents")
}

func hasJavaBinary(dir string) bool {
	javaBin := filepath.Join(dir, "bin", "java")
	if runtime.GOOS == "windows" {
		javaBin = filepath.Join(dir, "bin", "java.exe")
	}

	info, err := os.Stat(javaBin)
	return err == nil && !info.IsDir()
}
