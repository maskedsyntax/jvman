//go:build !windows

package shim

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/maskedsyntax/jvman/internal/paths"
)

type unixManager struct{}

func newPlatformManager() Manager {
	return &unixManager{}
}

func (m *unixManager) CreateShims() error {
	binDir, err := getBinDir()
	if err != nil {
		return fmt.Errorf("failed to get bin directory: %w", err)
	}

	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("failed to create bin directory: %w", err)
	}

	baseDir, err := paths.BaseDir()
	if err != nil {
		return fmt.Errorf("failed to get base directory: %w", err)
	}

	for _, binary := range shimBinaries {
		shimPath := filepath.Join(binDir, binary)
		if err := createUnixShim(shimPath, binary, baseDir); err != nil {
			return fmt.Errorf("failed to create shim for %s: %w", binary, err)
		}
	}

	return nil
}

func (m *unixManager) RemoveShims() error {
	binDir, err := getBinDir()
	if err != nil {
		return fmt.Errorf("failed to get bin directory: %w", err)
	}

	for _, binary := range shimBinaries {
		shimPath := filepath.Join(binDir, binary)
		os.Remove(shimPath)
	}

	return nil
}

func createUnixShim(shimPath, binary, baseDir string) error {
	script := fmt.Sprintf(`#!/bin/sh
set -e

resolve_version() {
    dir="$(pwd)"
    while [ "$dir" != "/" ]; do
        if [ -f "$dir/.jvman" ]; then
            cat "$dir/.jvman"
            return
        fi
        dir="$(dirname "$dir")"
    done

    if [ -f "%s/config.json" ]; then
        global=$(grep -o '"global"[[:space:]]*:[[:space:]]*"[^"]*"' "%s/config.json" 2>/dev/null | head -1 | sed 's/.*"global"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/')
        if [ -n "$global" ]; then
            echo "$global"
            return
        fi
    fi
}

version=$(resolve_version)
if [ -z "$version" ]; then
    echo "jvman: no Java version configured. Run 'jvman global <version>' or create a .jvman file." >&2
    exit 1
fi

java_home="%s/jvms/$version"
if [ ! -d "$java_home" ]; then
    echo "jvman: Java version '$version' is not installed. Run 'jvman install $version'." >&2
    exit 1
fi

exec "$java_home/bin/%s" "$@"
`, baseDir, baseDir, baseDir, binary)

	if err := os.WriteFile(shimPath, []byte(script), 0755); err != nil {
		return err
	}

	return nil
}
