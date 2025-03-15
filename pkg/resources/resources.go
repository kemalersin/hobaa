// Package resources provides access to application resources
package resources

import (
	"embed"
	"io"
	"os"
	"path/filepath"
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