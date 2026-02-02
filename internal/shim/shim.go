package shim

import "github.com/maskedsyntax/jvman/internal/paths"

var shimBinaries = []string{
	"java",
	"javac",
	"jar",
	"jshell",
	"javadoc",
	"jarsigner",
	"keytool",
	"jlink",
	"jpackage",
}

type Manager interface {
	CreateShims() error
	RemoveShims() error
}

func New() Manager {
	return newPlatformManager()
}

func GetShimBinaries() []string {
	return shimBinaries
}

func getBinDir() (string, error) {
	return paths.BinDir()
}
