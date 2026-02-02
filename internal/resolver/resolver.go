package resolver

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/maskedsyntax/jvman/internal/config"
	"github.com/maskedsyntax/jvman/internal/paths"
)

type Resolver struct {
	cfg *config.Config
}

func New(cfg *config.Config) *Resolver {
	return &Resolver{cfg: cfg}
}

type Resolution struct {
	Version string
	Path    string
	Source  string
}

func (r *Resolver) Resolve() (*Resolution, error) {
	if res := r.resolveFromLocalFile(); res != nil {
		return res, nil
	}

	if res := r.resolveFromLocalOverride(); res != nil {
		return res, nil
	}

	if res := r.resolveFromGlobal(); res != nil {
		return res, nil
	}

	return nil, nil
}

func (r *Resolver) resolveFromLocalFile() *Resolution {
	cwd, err := os.Getwd()
	if err != nil {
		return nil
	}

	dir := cwd
	for {
		localFile := filepath.Join(dir, paths.LocalVersionFile())
		data, err := os.ReadFile(localFile)
		if err == nil {
			version := strings.TrimSpace(string(data))
			if jvm, exists := r.cfg.Installed[version]; exists {
				return &Resolution{
					Version: version,
					Path:    jvm.Path,
					Source:  "local file: " + localFile,
				}
			}
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return nil
}

func (r *Resolver) resolveFromLocalOverride() *Resolution {
	cwd, err := os.Getwd()
	if err != nil {
		return nil
	}

	if version, exists := r.cfg.LocalOverrides[cwd]; exists {
		if jvm, exists := r.cfg.Installed[version]; exists {
			return &Resolution{
				Version: version,
				Path:    jvm.Path,
				Source:  "local override",
			}
		}
	}

	return nil
}

func (r *Resolver) resolveFromGlobal() *Resolution {
	if r.cfg.Global == "" {
		return nil
	}

	if jvm, exists := r.cfg.Installed[r.cfg.Global]; exists {
		return &Resolution{
			Version: r.cfg.Global,
			Path:    jvm.Path,
			Source:  "global",
		}
	}

	return nil
}

func (r *Resolver) ResolveJavaBinary() (string, error) {
	res, err := r.Resolve()
	if err != nil {
		return "", err
	}

	if res == nil {
		return "", nil
	}

	return paths.JavaBinaryPath(res.Path), nil
}

func (r *Resolver) ResolveBinary(name string) (string, error) {
	res, err := r.Resolve()
	if err != nil {
		return "", err
	}

	if res == nil {
		return "", nil
	}

	binDir := paths.JvmBinDir(res.Path)
	return filepath.Join(binDir, name), nil
}
