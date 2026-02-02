package registry

import (
	"fmt"
	"os"

	"github.com/maskedsyntax/jvman/internal/config"
	"github.com/maskedsyntax/jvman/internal/paths"
)

type Registry struct {
	cfg *config.Config
}

func New(cfg *config.Config) *Registry {
	return &Registry{cfg: cfg}
}

func (r *Registry) Add(name, path, vendor string) error {
	r.cfg.Installed[name] = config.InstalledJVM{
		Path:   path,
		Vendor: vendor,
	}
	return config.Save(r.cfg)
}

func (r *Registry) Remove(name string) error {
	jvm, exists := r.cfg.Installed[name]
	if !exists {
		return fmt.Errorf("JVM %s is not installed", name)
	}

	if err := os.RemoveAll(jvm.Path); err != nil {
		return fmt.Errorf("failed to remove JVM directory: %w", err)
	}

	delete(r.cfg.Installed, name)

	if r.cfg.Global == name {
		r.cfg.Global = ""
	}

	for dir, v := range r.cfg.LocalOverrides {
		if v == name {
			delete(r.cfg.LocalOverrides, dir)
		}
	}

	return config.Save(r.cfg)
}

func (r *Registry) Get(name string) (*config.InstalledJVM, error) {
	jvm, exists := r.cfg.Installed[name]
	if !exists {
		return nil, fmt.Errorf("JVM %s is not installed", name)
	}

	if _, err := os.Stat(jvm.Path); os.IsNotExist(err) {
		return nil, fmt.Errorf("JVM %s installation path does not exist", name)
	}

	return &jvm, nil
}

func (r *Registry) List() map[string]config.InstalledJVM {
	return r.cfg.Installed
}

func (r *Registry) IsInstalled(name string) bool {
	_, exists := r.cfg.Installed[name]
	return exists
}

func (r *Registry) SetGlobal(name string) error {
	if _, err := r.Get(name); err != nil {
		return err
	}

	r.cfg.Global = name
	return config.Save(r.cfg)
}

func (r *Registry) GetGlobal() string {
	return r.cfg.Global
}

func (r *Registry) SetLocalOverride(dir, name string) error {
	if _, err := r.Get(name); err != nil {
		return err
	}

	r.cfg.LocalOverrides[dir] = name
	return config.Save(r.cfg)
}

func (r *Registry) FindByVersion(version, vendor string) string {
	expectedName := fmt.Sprintf("%s-%s", vendor, version)
	if _, exists := r.cfg.Installed[expectedName]; exists {
		return expectedName
	}

	for name, jvm := range r.cfg.Installed {
		if jvm.Vendor == vendor {
			binPath := paths.JavaBinaryPath(jvm.Path)
			if _, err := os.Stat(binPath); err == nil {
				if name == expectedName {
					return name
				}
			}
		}
	}

	return ""
}
