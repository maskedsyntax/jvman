package paths

import (
	"os"
	"path/filepath"
	"runtime"
)

const (
	baseDirName   = ".jvman"
	jvmsDirName   = "jvms"
	binDirName    = "bin"
	configName    = "config.json"
	localFileName = ".jvman"
)

func homeDir() (string, error) {
	if runtime.GOOS == "windows" {
		home := os.Getenv("USERPROFILE")
		if home != "" {
			return home, nil
		}
	}
	return os.UserHomeDir()
}

func BaseDir() (string, error) {
	home, err := homeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, baseDirName), nil
}

func JvmsDir() (string, error) {
	base, err := BaseDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, jvmsDirName), nil
}

func BinDir() (string, error) {
	base, err := BaseDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, binDirName), nil
}

func ConfigPath() (string, error) {
	base, err := BaseDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, configName), nil
}

func JvmPath(name string) (string, error) {
	jvms, err := JvmsDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(jvms, name), nil
}

func LocalVersionFile() string {
	return localFileName
}

func EnsureDirectories() error {
	dirs := []func() (string, error){BaseDir, JvmsDir, BinDir}
	for _, dirFn := range dirs {
		dir, err := dirFn()
		if err != nil {
			return err
		}
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return nil
}

func JavaBinaryPath(jvmPath string) string {
	if runtime.GOOS == "windows" {
		return filepath.Join(jvmPath, "bin", "java.exe")
	}
	return filepath.Join(jvmPath, "bin", "java")
}

func JvmBinDir(jvmPath string) string {
	return filepath.Join(jvmPath, "bin")
}
