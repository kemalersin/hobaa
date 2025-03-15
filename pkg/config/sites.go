// Package config provides configuration management for the application
package config

import (
	"encoding/json"
	"github.com/kemalersin/hobaa/pkg/utils"
	"os"
	"path/filepath"
)

// Site represents a website configuration
type Site struct {
	Name     string `json:"name"`
	Title    string `json:"title"`
	URL      string `json:"url"`
	Width    int    `json:"width,omitempty"`
	Height   int    `json:"height,omitempty"`
	Icon     string `json:"icon,omitempty"`
	IsActive bool   `json:"is_active,omitempty"`
}

// SiteConfig represents the configuration for all sites
type SiteConfig struct {
	Sites     []Site
	ConfigDir string
}

// Default GitHub URL for sites.json
const DefaultGitHubSitesURL = "https://raw.githubusercontent.com/kemalersin/hobaa/refs/heads/main/sites.json"

// NewSiteConfig creates a new site configuration
func NewSiteConfig(configDir string) *SiteConfig {
	return &SiteConfig{
		Sites:     []Site{},
		ConfigDir: configDir,
	}
}

// LoadFromFile loads site configuration from a file
func (c *SiteConfig) LoadFromFile(filePath string) error {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil // File doesn't exist, return empty config
	}

	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	// Parse JSON
	return json.Unmarshal(data, &c.Sites)
}

// LoadFromGitHub loads site configuration from GitHub
func (c *SiteConfig) LoadFromGitHub(url string) error {
	// Download JSON from GitHub
	data, err := utils.DownloadJSON(url)
	if err != nil {
		return err
	}

	// Parse JSON
	var sites []Site
	if err := json.Unmarshal(data, &sites); err != nil {
		return err
	}

	// Add sites to config without overwriting existing sites
	for _, site := range sites {
		// Check if site already exists
		if existing := c.GetSiteByName(site.Name); existing == nil {
			// Only add if it doesn't exist
			c.Sites = append(c.Sites, site)
		}
	}

	return nil
}

// SaveToFile saves site configuration to a file
func (c *SiteConfig) SaveToFile(filePath string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Marshal JSON
	data, err := json.MarshalIndent(c.Sites, "", "  ")
	if err != nil {
		return err
	}

	// Write file
	return os.WriteFile(filePath, data, 0644)
}

// GetSiteByName returns a site by name
func (c *SiteConfig) GetSiteByName(name string) *Site {
	for i := range c.Sites {
		if c.Sites[i].Name == name {
			return &c.Sites[i]
		}
	}
	return nil
}

// AddSite adds a site to the configuration
func (c *SiteConfig) AddSite(site Site) {
	// Check if site already exists
	if existing := c.GetSiteByName(site.Name); existing != nil {
		// Update existing site
		*existing = site
		return
	}

	// Add new site
	c.Sites = append(c.Sites, site)
}

// SetActiveSite sets a site as active and all others as inactive
func (c *SiteConfig) SetActiveSite(name string) {
	for i := range c.Sites {
		c.Sites[i].IsActive = (c.Sites[i].Name == name)
	}
}

// CreateSiteFromURL creates a site configuration from a URL
func CreateSiteFromURL(name, url string) Site {
	return Site{
		Name:     name,
		Title:    name,
		URL:      url,
		IsActive: true,
	}
}

// GetAppDataSitesPath returns the path to the sites.json file in AppData
func GetAppDataSitesPath(appDataDir string) string {
	return filepath.Join(appDataDir, "sites.json")
}

// GetWorkingDirSitesPath returns the path to the sites.json file in the working directory
func GetWorkingDirSitesPath(execDir string) string {
	return filepath.Join(execDir, "sites.json")
} 