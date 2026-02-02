package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/maskedsyntax/jvman/internal/paths"
	"github.com/maskedsyntax/jvman/internal/provider"
)

const (
	cacheFileName = "cache.json"
	defaultTTL    = 1 * time.Hour
)

type VendorCache struct {
	Versions  []provider.Release `json:"versions"`
	UpdatedAt time.Time          `json:"updated_at"`
}

type CacheData struct {
	Vendors map[string]VendorCache `json:"vendors"`
}

type Cache struct {
	data     CacheData
	filePath string
	ttl      time.Duration
}

func New() (*Cache, error) {
	baseDir, err := paths.BaseDir()
	if err != nil {
		return nil, err
	}

	filePath := filepath.Join(baseDir, cacheFileName)

	c := &Cache{
		data: CacheData{
			Vendors: make(map[string]VendorCache),
		},
		filePath: filePath,
		ttl:      defaultTTL,
	}

	c.load()
	return c, nil
}

func (c *Cache) load() {
	data, err := os.ReadFile(c.filePath)
	if err != nil {
		return
	}

	json.Unmarshal(data, &c.data)
	if c.data.Vendors == nil {
		c.data.Vendors = make(map[string]VendorCache)
	}
}

func (c *Cache) save() error {
	if err := paths.EnsureDirectories(); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c.data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(c.filePath, data, 0644)
}

func (c *Cache) GetVersions(vendorName string) ([]provider.Release, bool) {
	vc, exists := c.data.Vendors[vendorName]
	if !exists {
		return nil, false
	}

	if time.Since(vc.UpdatedAt) > c.ttl {
		return nil, false
	}

	return vc.Versions, true
}

func (c *Cache) SetVersions(vendorName string, versions []provider.Release) error {
	c.data.Vendors[vendorName] = VendorCache{
		Versions:  versions,
		UpdatedAt: time.Now(),
	}
	return c.save()
}

func (c *Cache) Clear() error {
	c.data.Vendors = make(map[string]VendorCache)
	return c.save()
}

func (c *Cache) ClearVendor(vendorName string) error {
	delete(c.data.Vendors, vendorName)
	return c.save()
}
