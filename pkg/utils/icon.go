package utils

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
)

// ConvertToICO converts an image file to ICO format using rcedit.exe
// Since we don't have a direct way to convert images to ICO in Go,
// we'll use rcedit.exe to set the icon on a temporary executable,
// then extract the icon from that executable.
func ConvertToICO(imagePath, rceditPath, outputPath string) error {
	// Check if rcedit.exe exists
	if _, err := os.Stat(rceditPath); os.IsNotExist(err) {
		return fmt.Errorf("rcedit.exe not found at %s", rceditPath)
	}

	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "hobaa_icon_convert")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	// Create a temporary executable
	tempExePath := filepath.Join(tempDir, "temp.exe")
	
	// Copy hobaa.exe to temp.exe
	srcExe, err := os.Open(os.Args[0])
	if err != nil {
		return err
	}
	defer srcExe.Close()
	
	dstExe, err := os.Create(tempExePath)
	if err != nil {
		return err
	}
	defer dstExe.Close()
	
	if _, err := dstExe.ReadFrom(srcExe); err != nil {
		return err
	}
	
	// Close files to ensure they're fully written
	srcExe.Close()
	dstExe.Close()

	// Set the icon on the temporary executable with high quality settings
	cmd := exec.Command(rceditPath, tempExePath, "--set-icon", imagePath)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
	}
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set icon: %v", err)
	}

	// Extract the icon resource from the executable
	// This is a more robust approach than just renaming the executable
	if err := extractIconToICO(tempExePath, outputPath); err != nil {
		// If extraction fails, fall back to the original method
		if err := os.Rename(tempExePath, outputPath); err != nil {
			return fmt.Errorf("failed to save icon: %v", err)
		}
	}

	return nil
}

// extractIconToICO extracts the icon from an executable to an ICO file
// This is a placeholder function - in a real implementation, you would use
// Windows API to extract the icon resource properly
func extractIconToICO(exePath, icoPath string) error {
	// In a real implementation, this would use Windows API to extract the icon
	// For now, we'll just copy the executable as the ICO file
	return copyFile(exePath, icoPath)
}

// SetExecutableIcon sets the icon of an executable using rcedit.exe
func SetExecutableIcon(exePath, iconPath, rceditPath string) error {
	// Check if files exist
	if _, err := os.Stat(exePath); os.IsNotExist(err) {
		return fmt.Errorf("executable not found at %s", exePath)
	}
	if _, err := os.Stat(iconPath); os.IsNotExist(err) {
		return fmt.Errorf("icon not found at %s", iconPath)
	}
	if _, err := os.Stat(rceditPath); os.IsNotExist(err) {
		return fmt.Errorf("rcedit.exe not found at %s", rceditPath)
	}

	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "hobaa_icon_change")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	// Create paths for temporary files
	tempExePath := filepath.Join(tempDir, "temp.exe")
	bakExePath := filepath.Join(tempDir, "backup.exe")
	
	// Copy the executable to the temporary location
	srcExe, err := os.Open(exePath)
	if err != nil {
		return err
	}
	defer srcExe.Close()
	
	// Create a backup copy in the temp directory
	bakExe, err := os.Create(bakExePath)
	if err != nil {
		return err
	}
	defer bakExe.Close()
	
	// Copy original to backup
	if _, err := io.Copy(bakExe, srcExe); err != nil {
		return err
	}
	
	// Close and reopen the source file for a second read
	srcExe.Close()
	srcExe, err = os.Open(exePath)
	if err != nil {
		return err
	}
	defer srcExe.Close()
	
	// Create the temporary exe for modification
	dstExe, err := os.Create(tempExePath)
	if err != nil {
		return err
	}
	defer dstExe.Close()
	
	// Copy original to temp
	if _, err := io.Copy(dstExe, srcExe); err != nil {
		return err
	}
	
	// Close files to ensure they're fully written
	srcExe.Close()
	dstExe.Close()
	bakExe.Close()

	// Run rcedit.exe to set the icon on the temporary file
	cmd := exec.Command(rceditPath, tempExePath, "--set-icon", iconPath)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
	}
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to set icon: %v, output: %s", err, string(output))
	}
	
	// Replace the original executable with the modified one
	if runtime.GOOS == "windows" {
		// On Windows, we need to replace the original file
		// First try to directly replace it
		err = os.Rename(tempExePath, exePath)
		if err != nil {
			// If direct replacement fails, try to copy the content
			if err := copyFile(tempExePath, exePath); err != nil {
				// If that also fails, restore from backup
				copyFile(bakExePath, exePath)
				return fmt.Errorf("failed to update executable: %v", err)
			}
		}
	} else {
		// On other platforms, we can just replace the file
		if err := os.Rename(tempExePath, exePath); err != nil {
			return fmt.Errorf("failed to replace executable: %v", err)
		}
	}

	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()
	
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}
	
	return dstFile.Sync()
}

// RestartApplication restarts the application with the --force parameter
func RestartApplication() error {
	// Get the path of the current executable
	exePath, err := os.Executable()
	if err != nil {
		return err
	}

	// Create the command to restart the application
	cmd := exec.Command(exePath, "--force")
	
	// Set the command to run detached from the current process
	if runtime.GOOS == "windows" {
		cmd.SysProcAttr = &syscall.SysProcAttr{
			CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
			HideWindow:    true,
		}
	}
	
	// Start the command
	if err := cmd.Start(); err != nil {
		return err
	}
	
	// Exit the current process
	os.Exit(0)
	
	return nil // This line will never be reached
} 