// Package utils provides utility functions for the application
package utils

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// IsValidURL checks if a string is a valid URL
func IsValidURL(str string) bool {
	// Check for common TLDs to avoid treating random EXE names as URLs
	commonTLDs := []string{".com", ".net", ".org", ".io", ".co", ".edu", ".gov", ".info", ".biz", ".app"}
	
	hasTLD := false
	for _, tld := range commonTLDs {
		if strings.HasSuffix(str, tld) || strings.Contains(str, tld+"/") {
			hasTLD = true
			break
		}
	}
	
	// If no common TLD found, it's probably not a URL
	if !hasTLD && !strings.Contains(str, ".") {
		return false
	}
	
	// Try to parse the URL
	u, err := url.Parse(str)
	
	// Check if it's a valid URL with scheme and host
	if err == nil && u.Scheme != "" && u.Host != "" {
		return true
	}
	
	// If no scheme, try adding https:// and check again
	if err != nil || u.Scheme == "" {
		u, err = url.Parse("https://" + str)
		return err == nil && u.Host != ""
	}
	
	return false
}

// DownloadFile downloads a file from a URL to a local path
func DownloadFile(url, targetPath string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(targetPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Send GET request
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Create output file
	out, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Copy response body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

// DownloadJSON downloads a JSON file from a URL and returns its contents
func DownloadJSON(url string) ([]byte, error) {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Send GET request
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %s", resp.Status)
	}

	// Read response body
	return io.ReadAll(resp.Body)
}

// GetFaviconURL returns the favicon URL for a website
func GetFaviconURL(websiteURL string) (string, error) {
	// Parse URL
	u, err := url.Parse(websiteURL)
	if err != nil {
		return "", err
	}

	// Create favicon URL
	faviconURL := fmt.Sprintf("%s://%s/favicon.ico", u.Scheme, u.Host)
	
	// Check if favicon exists
	resp, err := http.Head(faviconURL)
	if err != nil || resp.StatusCode != http.StatusOK {
		// Try alternative favicon URL
		faviconURL = fmt.Sprintf("%s://%s/favicon.png", u.Scheme, u.Host)
		resp, err = http.Head(faviconURL)
		if err != nil || resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("favicon not found")
		}
	}
	
	return faviconURL, nil
}

// IsICOFile checks if a file is an ICO file
func IsICOFile(filePath string) bool {
	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		return false
	}
	defer file.Close()

	// Read first 4 bytes
	header := make([]byte, 4)
	if _, err := file.Read(header); err != nil {
		return false
	}

	// Check ICO header (0x00 0x00 0x01 0x00)
	return header[0] == 0x00 && header[1] == 0x00 && header[2] == 0x01 && header[3] == 0x00
}

// GetFileExtension returns the file extension from a URL
func GetFileExtension(urlStr string) string {
	u, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}
	
	path := u.Path
	ext := filepath.Ext(path)
	
	return strings.ToLower(ext)
} 