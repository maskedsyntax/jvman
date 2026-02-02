package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/maskedsyntax/jvman/internal/config"
	"github.com/maskedsyntax/jvman/internal/downloader"
	"github.com/maskedsyntax/jvman/internal/extractor"
	"github.com/maskedsyntax/jvman/internal/paths"
	"github.com/maskedsyntax/jvman/internal/provider"
	"github.com/maskedsyntax/jvman/internal/provider/corretto"
	"github.com/maskedsyntax/jvman/internal/provider/temurin"
	"github.com/maskedsyntax/jvman/internal/provider/zulu"
	"github.com/maskedsyntax/jvman/internal/registry"
	"github.com/maskedsyntax/jvman/internal/resolver"
	"github.com/maskedsyntax/jvman/internal/shim"
)

var (
	version = "dev"
)

var vendors = map[string]func() provider.Vendor{
	"temurin":  func() provider.Vendor { return temurin.New() },
	"corretto": func() provider.Vendor { return corretto.New() },
	"zulu":     func() provider.Vendor { return zulu.New() },
}

var versionNameFuncs = map[string]func(string) string{
	"temurin":  temurin.VersionName,
	"corretto": corretto.VersionName,
	"zulu":     zulu.VersionName,
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "jvman",
	Short: "Cross-platform Java Version Manager",
	Long:  "jvman is a tool for installing and switching between multiple Java versions.",
}

var (
	installVendor string
	listVendor    string
)

func init() {
	installCmd.Flags().StringVarP(&installVendor, "vendor", "v", "temurin", "JDK vendor (temurin, corretto, zulu)")
	listCmd.Flags().StringVarP(&listVendor, "vendor", "v", "", "Filter by vendor (temurin, corretto, zulu)")

	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(globalCmd)
	rootCmd.AddCommand(useCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(whichCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(initCmd)
}

var installCmd = &cobra.Command{
	Use:   "install <version>",
	Short: "Install a Java version",
	Long:  "Download and install a specific Java version.\n\nExamples:\n  jvman install 21\n  jvman install 17 --vendor=corretto\n  jvman install 11 -v zulu",
	Args:  cobra.ExactArgs(1),
	RunE:  runInstall,
}

func runInstall(cmd *cobra.Command, args []string) error {
	version := args[0]

	vendorFactory, ok := vendors[installVendor]
	if !ok {
		return fmt.Errorf("unknown vendor: %s (available: temurin, corretto, zulu)", installVendor)
	}

	versionNameFunc := versionNameFuncs[installVendor]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	reg := registry.New(cfg)
	vendor := vendorFactory()

	installName := versionNameFunc(version)
	if reg.IsInstalled(installName) {
		fmt.Printf("Java %s (%s) is already installed\n", version, installName)
		return nil
	}

	fmt.Printf("Fetching release info for Java %s from %s...\n", version, installVendor)
	release, err := vendor.GetRelease(version)
	if err != nil {
		return fmt.Errorf("failed to get release info: %w", err)
	}

	fmt.Printf("Found: %s\n", release.FullVersion)
	fmt.Printf("Downloading from %s...\n", release.DownloadURL)

	dl := downloader.New()
	jvmsDir, err := paths.JvmsDir()
	if err != nil {
		return fmt.Errorf("failed to get jvms directory: %w", err)
	}

	tmpDir := filepath.Join(jvmsDir, ".tmp")
	result, err := dl.Download(release.DownloadURL, tmpDir, release.FileName, release.Checksum)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	fmt.Println("Extracting...")
	ext := extractor.ForFile(release.FileName)
	extractDir := filepath.Join(tmpDir, "extract")
	javaHome, err := ext.Extract(result.FilePath, extractDir)
	if err != nil {
		os.RemoveAll(tmpDir)
		return fmt.Errorf("extraction failed: %w", err)
	}

	installPath, err := paths.JvmPath(installName)
	if err != nil {
		os.RemoveAll(tmpDir)
		return fmt.Errorf("failed to get install path: %w", err)
	}

	os.RemoveAll(installPath)
	if err := os.Rename(javaHome, installPath); err != nil {
		os.RemoveAll(tmpDir)
		return fmt.Errorf("failed to move JDK to install path: %w", err)
	}

	os.RemoveAll(tmpDir)

	if err := reg.Add(installName, installPath, vendor.Name()); err != nil {
		return fmt.Errorf("failed to register installation: %w", err)
	}

	shimMgr := shim.New()
	if err := shimMgr.CreateShims(); err != nil {
		fmt.Printf("Warning: failed to create shims: %v\n", err)
	}

	fmt.Printf("Successfully installed Java %s as %s\n", version, installName)

	if cfg.Global == "" {
		if err := reg.SetGlobal(installName); err != nil {
			return fmt.Errorf("failed to set global version: %w", err)
		}
		fmt.Printf("Set %s as global default\n", installName)
	}

	return nil
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed and available Java versions",
	RunE:  runList,
}

func runList(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	reg := registry.New(cfg)
	installed := reg.List()

	fmt.Println("Installed versions:")
	if len(installed) == 0 {
		fmt.Println("  (none)")
	} else {
		for name, jvm := range installed {
			marker := "  "
			if name == cfg.Global {
				marker = "* "
			}
			fmt.Printf("%s%s (%s)\n", marker, name, jvm.Vendor)
		}
	}

	vendorsToList := []string{"temurin", "corretto", "zulu"}
	if listVendor != "" {
		if _, ok := vendors[listVendor]; !ok {
			return fmt.Errorf("unknown vendor: %s", listVendor)
		}
		vendorsToList = []string{listVendor}
	}

	for _, vendorName := range vendorsToList {
		fmt.Println()
		fmt.Printf("Available versions (%s):\n", vendorName)

		vendorFactory := vendors[vendorName]
		versionNameFunc := versionNameFuncs[vendorName]
		vendor := vendorFactory()

		available, err := vendor.ListAvailableVersions()
		if err != nil {
			fmt.Printf("  Error fetching: %v\n", err)
			continue
		}

		for _, rel := range available {
			installedName := versionNameFunc(rel.Version)
			status := ""
			if reg.IsInstalled(installedName) {
				status = " [installed]"
			}
			fmt.Printf("  %s%s\n", rel.Version, status)
		}
	}

	return nil
}

var globalCmd = &cobra.Command{
	Use:   "global <version>",
	Short: "Set the global default Java version",
	Args:  cobra.ExactArgs(1),
	RunE:  runGlobal,
}

func runGlobal(cmd *cobra.Command, args []string) error {
	version := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	reg := registry.New(cfg)

	name := resolveInstalledName(reg, version)
	if name == "" {
		return fmt.Errorf("Java version %s is not installed. Run 'jvman install %s' first", version, version)
	}

	if err := reg.SetGlobal(name); err != nil {
		return fmt.Errorf("failed to set global version: %w", err)
	}

	shimMgr := shim.New()
	if err := shimMgr.CreateShims(); err != nil {
		fmt.Printf("Warning: failed to update shims: %v\n", err)
	}

	fmt.Printf("Global Java version set to %s\n", name)
	return nil
}

var useCmd = &cobra.Command{
	Use:   "use <version>",
	Short: "Set the Java version for the current directory",
	Long:  "Creates a .jvman file in the current directory to specify the Java version",
	Args:  cobra.ExactArgs(1),
	RunE:  runUse,
}

func runUse(cmd *cobra.Command, args []string) error {
	version := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	reg := registry.New(cfg)

	name := resolveInstalledName(reg, version)
	if name == "" {
		return fmt.Errorf("Java version %s is not installed. Run 'jvman install %s' first", version, version)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	localFile := filepath.Join(cwd, paths.LocalVersionFile())
	if err := os.WriteFile(localFile, []byte(name+"\n"), 0644); err != nil {
		return fmt.Errorf("failed to create .jvman file: %w", err)
	}

	fmt.Printf("Created .jvman file with version %s\n", name)
	return nil
}

var removeCmd = &cobra.Command{
	Use:   "remove <version>",
	Short: "Remove an installed Java version",
	Args:  cobra.ExactArgs(1),
	RunE:  runRemove,
}

func runRemove(cmd *cobra.Command, args []string) error {
	version := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	reg := registry.New(cfg)

	name := resolveInstalledName(reg, version)
	if name == "" {
		return fmt.Errorf("Java version %s is not installed", version)
	}

	if err := reg.Remove(name); err != nil {
		return fmt.Errorf("failed to remove: %w", err)
	}

	fmt.Printf("Removed %s\n", name)
	return nil
}

var whichCmd = &cobra.Command{
	Use:   "which",
	Short: "Show the currently active Java installation",
	RunE:  runWhich,
}

func runWhich(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	res := resolver.New(cfg)
	resolution, err := res.Resolve()
	if err != nil {
		return fmt.Errorf("failed to resolve version: %w", err)
	}

	if resolution == nil {
		fmt.Println("No Java version is currently configured")
		return nil
	}

	fmt.Printf("Version: %s\n", resolution.Version)
	fmt.Printf("Path: %s\n", resolution.Path)
	fmt.Printf("Source: %s\n", resolution.Source)

	javaBin := paths.JavaBinaryPath(resolution.Path)
	fmt.Printf("Java binary: %s\n", javaBin)

	return nil
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show jvman version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("jvman version %s\n", version)
	},
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize jvman (create directories and shims)",
	RunE:  runInit,
}

func runInit(cmd *cobra.Command, args []string) error {
	if err := paths.EnsureDirectories(); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	shimMgr := shim.New()
	if err := shimMgr.CreateShims(); err != nil {
		return fmt.Errorf("failed to create shims: %w", err)
	}

	binDir, _ := paths.BinDir()
	fmt.Println("jvman initialized successfully!")
	fmt.Println()
	fmt.Println("Add the following to your shell profile:")
	fmt.Printf("  export PATH=\"%s:$PATH\"\n", binDir)

	return nil
}

func resolveInstalledName(reg *registry.Registry, version string) string {
	if reg.IsInstalled(version) {
		return version
	}

	for _, versionNameFunc := range versionNameFuncs {
		name := versionNameFunc(version)
		if reg.IsInstalled(name) {
			return name
		}
	}

	installed := reg.List()
	for name := range installed {
		if strings.Contains(name, version) {
			return name
		}
	}

	return ""
}
