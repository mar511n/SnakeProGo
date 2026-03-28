package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	GitHubRepoOwner = "mar511n" // Change to your GitHub username
	GitHubRepoName  = "SnakeProGo"
	GitHubReleases  = "https://github.com/" + GitHubRepoOwner + "/" + GitHubRepoName + "/releases/download"
)

// DirectoryStructure defines the folders needed for SnakeProGo
var DirectoryStructure = []string{
	"config/userconfig",
}

// Installer handles the installation of SnakeProGo
type Installer struct {
	installDir string
	version    string
	os         string
	arch       string
}

// NewInstaller creates a new installer instance
func NewInstaller(installDir, version string) *Installer {
	return &Installer{
		installDir: installDir,
		version:    version,
		os:         runtime.GOOS,
		arch:       runtime.GOARCH,
	}
}

// GetBinaryURL returns the appropriate GitHub release URL for the binary
func (i *Installer) GetBinaryURL() string {
	filename := fmt.Sprintf("SnakeProGo-%s-%s.zip", i.os, i.arch)
	return fmt.Sprintf("%s/v%s/%s", GitHubReleases, i.version, filename)
}

// GetAssetsURL returns the URL to download game assets
func (i *Installer) GetAssetsURL() string {
	return fmt.Sprintf("%s/v%s/assets.zip", GitHubReleases, i.version)
}

// CreateDirectories creates the required directory structure
func (i *Installer) CreateDirectories() error {
	fmt.Println("📁 Creating directory structure...")

	for _, dir := range DirectoryStructure {
		fullPath := filepath.Join(i.installDir, dir)
		if err := os.MkdirAll(fullPath, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", fullPath, err)
		}
		fmt.Printf("  ✓ Created %s\n", dir)
	}

	return nil
}

// DownloadFile downloads a file from URL to destination
func (i *Installer) DownloadFile(url, destPath string) error {
	fmt.Printf("⬇️  Downloading from: %s\n", url)

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	file, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, resp.Body); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Printf("  ✓ Downloaded to %s\n", destPath)
	return nil
}

// UnzipFile extracts a zip file to destination
func (i *Installer) UnzipFile(zipPath, destDir string) error {
	fmt.Printf("📦 Extracting %s...\n", filepath.Base(zipPath))

	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("failed to open zip: %w", err)
	}
	defer reader.Close()

	for _, file := range reader.File {
		fpath := filepath.Join(destDir, file.Name)

		// Security check: prevent directory traversal
		if !strings.HasPrefix(fpath, filepath.Clean(destDir)+string(os.PathSeparator)) && fpath != filepath.Clean(destDir) {
			return fmt.Errorf("invalid file path in zip: %s", fpath)
		}

		if file.FileInfo().IsDir() {
			os.MkdirAll(fpath, file.Mode())
			continue
		}

		os.MkdirAll(filepath.Dir(fpath), 0755)

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}

		rc, err := file.Open()
		if err != nil {
			outFile.Close()
			return fmt.Errorf("failed to open zip entry: %w", err)
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return fmt.Errorf("failed to extract file: %w", err)
		}

		fmt.Printf("  ✓ Extracted %s\n", file.Name)
	}

	return nil
}

// InstallBinary downloads and installs the game binary
func (i *Installer) InstallBinary() error {
	fmt.Println("\n🎮 Installing binary...")

	binaryURL := i.GetBinaryURL()
	tempZip := filepath.Join(i.installDir, "binary.zip")
	defer os.Remove(tempZip)

	if err := i.DownloadFile(binaryURL, tempZip); err != nil {
		return fmt.Errorf("failed to download binary: %w", err)
	}

	if err := i.UnzipFile(tempZip, i.installDir); err != nil {
		return fmt.Errorf("failed to extract binary: %w", err)
	}

	return nil
}

// InstallAssets downloads and installs game assets
func (i *Installer) InstallAssets() error {
	fmt.Println("\n🎨 Installing assets...")

	assetsURL := i.GetAssetsURL()
	tempZip := filepath.Join(i.installDir, "assets.zip")
	defer os.Remove(tempZip)

	if err := i.DownloadFile(assetsURL, tempZip); err != nil {
		return fmt.Errorf("failed to download assets: %w", err)
	}

	if err := i.UnzipFile(tempZip, filepath.Join(i.installDir, "res")); err != nil {
		return fmt.Errorf("failed to extract assets: %w", err)
	}

	return nil
}

// Install runs the complete installation process
func (i *Installer) Install() error {
	fmt.Printf("🚀 Installing SnakeProGo v%s\n", i.version)
	fmt.Printf("📍 Installation directory: %s\n", i.installDir)
	fmt.Printf("🖥️  Platform: %s/%s\n\n", i.os, i.arch)

	if err := i.CreateDirectories(); err != nil {
		return err
	}

	if err := i.InstallBinary(); err != nil {
		return err
	}

	if err := i.InstallAssets(); err != nil {
		return err
	}

	fmt.Println("\n✅ Installation complete!")
	fmt.Printf("\n📝 Next steps:\n")
	fmt.Printf("   1. Navigate to: %s\n", i.installDir)
	binary := "SnakeProGo"
	if i.os == "windows" {
		binary += ".exe"
	}
	fmt.Printf("   2. Run: ./%s\n", binary)
	return nil
}

func main() {
	userHome, err := os.UserHomeDir()
	if err != nil {
		userHome = "." // Fallback to current directory if home cannot be determined
	}
	installDir := filepath.Join(userHome, "snakeprogo")
	version := flag.String("version", "latest", "Version to install (default: latest)")
	flag.Parse()

	// Create installDir if it doesn't exist
	if err := os.MkdirAll(installDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to create directory: %v\n", err)
		os.Exit(1)
	}

	installer := NewInstaller(installDir, *version)

	if err := installer.Install(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
