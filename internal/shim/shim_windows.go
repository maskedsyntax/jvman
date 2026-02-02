//go:build windows

package shim

import (
	"fmt"
	"os"
	"path/filepath"
)

type windowsManager struct{}

func newPlatformManager() Manager {
	return &windowsManager{}
}

func (m *windowsManager) CreateShims() error {
	binDir, err := getBinDir()
	if err != nil {
		return fmt.Errorf("failed to get bin directory: %w", err)
	}

	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("failed to create bin directory: %w", err)
	}

	for _, binary := range shimBinaries {
		shimPath := filepath.Join(binDir, binary+".cmd")
		if err := createWindowsShim(shimPath, binary); err != nil {
			return fmt.Errorf("failed to create shim for %s: %w", binary, err)
		}
	}

	return nil
}

func (m *windowsManager) RemoveShims() error {
	binDir, err := getBinDir()
	if err != nil {
		return fmt.Errorf("failed to get bin directory: %w", err)
	}

	for _, binary := range shimBinaries {
		shimPath := filepath.Join(binDir, binary+".cmd")
		os.Remove(shimPath)
		shimPath = filepath.Join(binDir, binary+".exe")
		os.Remove(shimPath)
	}

	return nil
}

func createWindowsShim(shimPath, binary string) error {
	script := fmt.Sprintf(`@echo off
setlocal enabledelayedexpansion

set "JVMAN_HOME=%%USERPROFILE%%\.jvman"
set "VERSION="

rem Check for .jvman file in current and parent directories
set "DIR=%%CD%%"
:findversion
if exist "%%DIR%%\.jvman" (
    set /p VERSION=<"%%DIR%%\.jvman"
    goto :found
)
for %%%%i in ("%%DIR%%\..") do set "PARENT=%%%%~fi"
if "%%PARENT%%"=="%%DIR%%" goto :checkglobal
set "DIR=%%PARENT%%"
goto :findversion

:checkglobal
rem Read global version from config.json
if exist "%%JVMAN_HOME%%\config.json" (
    for /f "tokens=2 delims=:," %%%%a in ('findstr /c:"\"global\"" "%%JVMAN_HOME%%\config.json"') do (
        set "VERSION=%%%%~a"
        set "VERSION=!VERSION:"=!"
        set "VERSION=!VERSION: =!"
    )
)

:found
if "%%VERSION%%"=="" (
    echo jvman: no Java version configured. Run 'jvman global ^<version^>' or create a .jvman file. >&2
    exit /b 1
)

set "JAVA_HOME=%%JVMAN_HOME%%\jvms\%%VERSION%%"
if not exist "%%JAVA_HOME%%" (
    echo jvman: Java version '%%VERSION%%' is not installed. Run 'jvman install %%VERSION%%'. >&2
    exit /b 1
)

"%%JAVA_HOME%%\bin\%s.exe" %%*
`, binary)

	return os.WriteFile(shimPath, []byte(script), 0755)
}
