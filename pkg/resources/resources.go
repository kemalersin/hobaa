// Package resources provides access to application resources
package resources

import (
	"embed"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// embeddedFiles holds the embedded resources
var embeddedFiles embed.FS

// SetEmbeddedFiles sets the embedded files from main package
func SetEmbeddedFiles(files embed.FS) {
	embeddedFiles = files
}

// CopyFile copies a file from source to destination
func CopyFile(src, dst string) error {
	// Open source file
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// Create destination directory if it doesn't exist
	dstDir := filepath.Dir(dst)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return err
	}

	// Create destination file
	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	// Copy the file
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	// Sync to ensure the file is written
	err = destFile.Sync()
	if err != nil {
		return err
	}

	// If the source file is executable, make the destination executable too
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	// Set permissions
	return os.Chmod(dst, srcInfo.Mode())
}

// CopyEmbeddedFile copies an embedded file to the specified path
func CopyEmbeddedFile(embeddedPath, targetPath string) error {
	// Read embedded file
	data, err := embeddedFiles.ReadFile(embeddedPath)
	if err != nil {
		return err
	}

	// Create destination directory if it doesn't exist
	dstDir := filepath.Dir(targetPath)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return err
	}

	// Create destination file
	destFile, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	// Write data to file
	_, err = destFile.Write(data)
	if err != nil {
		return err
	}

	// Sync to ensure the file is written
	err = destFile.Sync()
	if err != nil {
		return err
	}

	// Set executable permission for rcedit.exe
	if filepath.Base(embeddedPath) == "rcedit.exe" {
		return os.Chmod(targetPath, 0755) // rwxr-xr-x
	}

	return nil
}

// CopyDefaultIcon copies the default icon to the specified path
func CopyDefaultIcon(execDir, targetPath string) error {
	return CopyEmbeddedFile("resources/default.ico", targetPath)
}

// CopyRcedit copies the rcedit.exe to the specified path
func CopyRcedit(execDir, targetPath string) error {
	return CopyEmbeddedFile("resources/rcedit.exe", targetPath)
}

// CopySitesJson copies the sites.json file to the specified path
func CopySitesJson(targetPath string) error {
	return CopyEmbeddedFile("resources/sites.json", targetPath)
}

// CopyAllIcons copies all icon files from the embedded resources to the target directory
func CopyAllIcons(targetDir string) error {
	// Create target directory if it doesn't exist
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return err
	}

	// Walk through the embedded icons directory
	return fs.WalkDir(embeddedFiles, "resources/icons/ico", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Get the icon filename
		iconName := filepath.Base(path)

		// Create the target path
		targetPath := filepath.Join(targetDir, iconName)

		// Check if the icon already exists
		if _, err := os.Stat(targetPath); os.IsNotExist(err) {
			// Copy the icon file
			return CopyEmbeddedFile(path, targetPath)
		}

		return nil
	})
}

// EnsureIconExists ensures that a specific icon exists in the target directory
func EnsureIconExists(iconName string, targetDir string) error {
	// Create the full target path
	targetPath := filepath.Join(targetDir, iconName)

	// Check if the icon already exists
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		// If icon path starts with ico/, try to download from GitHub
		if strings.HasPrefix(iconName, "ico/") {
			// Remove ico/ prefix
			iconName = strings.TrimPrefix(iconName, "ico/")
			// Create GitHub URL
			githubURL := "https://raw.githubusercontent.com/kemalersin/hobaa/refs/heads/main/resources/icons/ico/" + iconName

			// Download the icon
			resp, err := http.Get(githubURL)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			// Create destination file
			destFile, err := os.Create(targetPath)
			if err != nil {
				return err
			}
			defer destFile.Close()

			// Copy the file
			_, err = io.Copy(destFile, resp.Body)
			if err != nil {
				return err
			}

			return destFile.Sync()
		}

		// Try to find the icon in the embedded resources
		embeddedPath := filepath.Join("resources/icons/ico", iconName)

		// Copy the icon file
		return CopyEmbeddedFile(embeddedPath, targetPath)
	}

	return nil
}
