package provider

type Release struct {
	Version      string
	FullVersion  string
	Vendor       string
	DownloadURL  string
	Checksum     string
	ChecksumType string
	FileName     string
	OS           string
	Arch         string
}

type Options struct {
	Arch string
}

type Vendor interface {
	Name() string
	ListAvailableVersions() ([]Release, error)
	GetRelease(version string, opts *Options) (*Release, error)
}
