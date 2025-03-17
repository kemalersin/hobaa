package app

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/kemalersin/hobaa/pkg/config"
	"github.com/kemalersin/hobaa/pkg/dpi"
	"github.com/kemalersin/hobaa/pkg/resources"
	"github.com/kemalersin/hobaa/pkg/utils"
	"github.com/kemalersin/hobaa/pkg/webview"
	"github.com/kemalersin/hobaa/pkg/winapi"
)

// App represents the main application
type App struct {
	webView     *webview.WebView
	appDataDir  string
	iconsDir    string
	webViewDir  string
	execName    string
	execDir     string
	execPath    string
	siteConfig  *config.SiteConfig
	currentSite *config.Site
	forceMode   bool
	iconChanged bool
	hwnd        syscall.Handle
	changeIcon  bool
	targetExe   string
	iconPath    string
}

// New creates a new application instance
func New() *App {
	// Enable high DPI support
	dpi.SetProcessDpiAwareness()

	// Create app instance
	app := &App{}

	// Parse command line flags
	app.parseFlags()

	// Initialize app data directories
	app.initAppData()

	// Extract resources
	app.extractResources()

	// If in change-icon mode, change the icon and exit
	if app.changeIcon {
		app.changeIconAndExit()
		return app
	}

	// Load site configuration
	app.loadSiteConfig()

	// Handle site configuration
	app.handleSiteConfig()

	return app
}

// parseFlags parses command line flags
func (a *App) parseFlags() {
	// Define flags
	forceFlag := flag.Bool("force", false, "Force application to start")
	changeIconFlag := flag.Bool("change-icon", false, "Change icon of target executable")
	targetExeFlag := flag.String("target-exe", "", "Target executable to change icon")
	iconPathFlag := flag.String("icon-path", "", "Path to icon file")

	// Parse flags
	flag.Parse()

	// Set force mode
	a.forceMode = *forceFlag
	a.changeIcon = *changeIconFlag
	a.targetExe = *targetExeFlag
	a.iconPath = *iconPathFlag
}

// initAppData initializes the application data directories
func (a *App) initAppData() {
	// Get executable path and name
	var err error
	a.execPath, err = os.Executable()
	if err != nil {
		a.execPath = os.Args[0]
	}

	a.execDir = filepath.Dir(a.execPath)
	a.execName = filepath.Base(a.execPath)
	if ext := filepath.Ext(a.execName); ext != "" {
		a.execName = a.execName[:len(a.execName)-len(ext)]
	}

	// Get AppData directory
	appData := os.Getenv("APPDATA")
	if appData == "" {
		appData = os.Getenv("LOCALAPPDATA")
	}
	if appData == "" {
		appData = filepath.Join(a.execDir, "AppData")
	}

	// Create application directory in AppData
	a.appDataDir = filepath.Join(appData, "Hobaa")
	os.MkdirAll(a.appDataDir, 0755)

	// Create icons directory
	a.iconsDir = filepath.Join(a.appDataDir, "icons")
	os.MkdirAll(a.iconsDir, 0755)

	// Set WebView directory to AppData directory
	a.webViewDir = a.appDataDir
}

// extractResources extracts resources to the AppData directory
func (a *App) extractResources() {
	// Copy rcedit.exe if it doesn't exist
	rceditPath := filepath.Join(a.appDataDir, "rcedit.exe")
	if _, err := os.Stat(rceditPath); os.IsNotExist(err) {
		resources.CopyRcedit(a.execDir, rceditPath)
	}

	// Copy default icon if it doesn't exist
	iconPath := filepath.Join(a.iconsDir, "hobaa.ico")
	if _, err := os.Stat(iconPath); os.IsNotExist(err) {
		resources.CopyDefaultIcon(a.execDir, iconPath)
	}

	// Copy sites.json if it doesn't exist
	appDataSitesPath := config.GetAppDataSitesPath(a.appDataDir)
	if _, err := os.Stat(appDataSitesPath); os.IsNotExist(err) {
		resources.CopySitesJson(appDataSitesPath)
	}

	// Copy all icons from resources to AppData icons directory
	resources.CopyAllIcons(a.iconsDir)
}

// changeIconAndExit changes the icon of the target executable and exits
func (a *App) changeIconAndExit() {
	// Check if target executable and icon path are provided
	if a.targetExe == "" || a.iconPath == "" {
		fmt.Println("Target executable or icon path not provided")
		os.Exit(1)
	}

	// Get rcedit path
	rceditPath := filepath.Join(a.appDataDir, "rcedit.exe")

	// Wait a bit to ensure the original process has exited
	winapi.Sleep(1000)

	// Change the icon
	err := utils.SetExecutableIcon(a.targetExe, a.iconPath, rceditPath)
	if err != nil {
		fmt.Printf("Failed to set icon: %v\n", err)
		os.Exit(1)
	}

	// Clear icon cache
	winapi.ClearIconCache()

	// Start the original executable with --force parameter
	cmd := exec.Command(a.targetExe, "--force")
	cmd.Start()

	// Exit
	os.Exit(0)
}

// loadSiteConfig loads the site configuration
func (a *App) loadSiteConfig() {
	// Create site configuration
	a.siteConfig = config.NewSiteConfig(a.appDataDir)

	// Get sites.json paths
	appDataSitesPath := config.GetAppDataSitesPath(a.appDataDir)

	// Load sites from AppData
	a.siteConfig.LoadFromFile(appDataSitesPath)

	// Check if current EXE filename exists in sites.json
	a.currentSite = a.siteConfig.GetSiteByName(a.execName)

	// If this is the main hobaa.exe and the site doesn't exist, create a default site
	if a.execName == "hobaa" && a.currentSite == nil {
		// Create default site for hobaa
		site := config.CreateSiteFromURL(a.execName, "https://www.google.com")
		site.IsActive = true
		a.siteConfig.AddSite(site)
		a.siteConfig.SaveToFile(appDataSitesPath)
		a.currentSite = a.siteConfig.GetSiteByName(a.execName)
	}
}

// handleSiteConfig handles site configuration based on EXE filename
func (a *App) handleSiteConfig() {
	// Get sites.json paths
	appDataSitesPath := config.GetAppDataSitesPath(a.appDataDir)
	workingDirSitesPath := config.GetWorkingDirSitesPath(a.execDir)

	// If force mode is enabled, set current site as active
	if a.forceMode && a.currentSite != nil {
		// Set current site as active and all others as inactive
		a.siteConfig.SetActiveSite(a.execName)

		// Save to AppData
		a.siteConfig.SaveToFile(appDataSitesPath)

		// Clear Windows application cache
		a.clearWindowsCache()

		// Launch application
		return
	}

	// If current site exists and is active, launch application directly
	if a.currentSite != nil && a.currentSite.IsActive {
		return
	}

	// Check if icon exists for current EXE name
	iconPath := filepath.Join(a.iconsDir, a.execName+".ico")
	iconExists := false
	if _, err := os.Stat(iconPath); !os.IsNotExist(err) {
		iconExists = true
	}

	// If site exists but is not active, and icon exists, just set it active and use existing icon
	if a.currentSite != nil && !a.currentSite.IsActive && iconExists {
		// Set current site as active
		a.siteConfig.SetActiveSite(a.execName)

		// Save to AppData
		a.siteConfig.SaveToFile(appDataSitesPath)

		// Set icon change flag and launch icon changer if needed
		a.iconChanged = true
		if !a.forceMode {
			a.launchIconChanger(iconPath)
		}
		return
	}

	// If site is not found or not active, check other sources
	if a.currentSite == nil || !a.currentSite.IsActive {
		// Check working directory sites.json
		workingDirConfig := config.NewSiteConfig(a.execDir)
		if err := workingDirConfig.LoadFromFile(workingDirSitesPath); err == nil {
			if site := workingDirConfig.GetSiteByName(a.execName); site != nil {
				a.updateSiteFromSource(site)
				return
			}
		}

		// Check GitHub sites.json
		githubConfig := config.NewSiteConfig("")
		if err := githubConfig.LoadFromGitHub(config.DefaultGitHubSitesURL); err == nil {
			if site := githubConfig.GetSiteByName(a.execName); site != nil {
				a.updateSiteFromSource(site)
				return
			}
		}

		// If still not found, check if EXE name is a URL
		if utils.IsValidURL(a.execName) || utils.IsValidURL("https://"+a.execName) {
			// Create URL if needed
			url := a.execName
			if !strings.HasPrefix(url, "http") {
				url = "https://" + url
			}

			// Get favicon URL
			faviconURL, err := utils.GetFaviconURL(url)

			// Create site from URL
			site := config.CreateSiteFromURL(a.execName, url)

			// Set icon URL if available
			if err == nil && faviconURL != "" {
				site.Icon = faviconURL
			} else {
				// If favicon URL cannot be retrieved, use default icon path
				defaultIconPath := filepath.Join(a.iconsDir, "hobaa.ico")
				if _, err := os.Stat(defaultIconPath); !os.IsNotExist(err) {
					// Use a placeholder URL to indicate default icon
					site.Icon = "default://hobaa.ico"
				}
			}

			// Try to download favicon only if icon doesn't exist
			if !iconExists {
				a.downloadFavicon(url, a.execName)
			} else {
				// If icon exists, set icon change flag and launch icon changer if needed
				a.iconChanged = true
				if !a.forceMode {
					a.launchIconChanger(iconPath)
				}
			}

			// Add site to config and set as active
			a.siteConfig.AddSite(site)
			a.siteConfig.SetActiveSite(a.execName)
			a.siteConfig.SaveToFile(appDataSitesPath)
			a.currentSite = a.siteConfig.GetSiteByName(a.execName)
		} else {
			// EXE name is not a URL and not found in any config, create default site with Google
			// Use default icon
			defaultIconPath := filepath.Join(a.iconsDir, "hobaa.ico")
			iconPath := filepath.Join(a.iconsDir, a.execName+".ico")

			// Copy default icon to site-specific icon if it doesn't exist
			if !iconExists {
				// First check if the icon exists in resources
				err := resources.EnsureIconExists(a.execName+".ico", a.iconsDir)
				if err != nil && defaultIconPath != "" {
					// If not in resources, copy the default icon
					if _, err := os.Stat(defaultIconPath); !os.IsNotExist(err) {
						resources.CopyFile(defaultIconPath, iconPath)
						iconExists = true
					}
				} else {
					iconExists = true
				}
			}

			// Create default site with Google URL
			site := config.CreateSiteFromURL(a.execName, "https://www.google.com")
			site.Icon = "default://hobaa.ico"
			a.siteConfig.AddSite(site)
			a.siteConfig.SetActiveSite(a.execName)
			a.siteConfig.SaveToFile(appDataSitesPath)
			a.currentSite = a.siteConfig.GetSiteByName(a.execName)

			// Set icon change flag and launch icon changer if needed
			if iconExists && !a.forceMode {
				a.iconChanged = true
				a.launchIconChanger(iconPath)
			}
		}
	} else if iconExists && !a.forceMode {
		// If icon exists but we're not in force mode, trigger icon change
		a.iconChanged = true
		a.launchIconChanger(iconPath)
	}

	// If icon was changed, restart application
	if a.iconChanged && !a.forceMode {
		a.restartWithForce()
	}
}

// updateSiteFromSource updates a site from a source configuration
func (a *App) updateSiteFromSource(site *config.Site) {
	// Get sites.json path
	appDataSitesPath := config.GetAppDataSitesPath(a.appDataDir)

	// Check if site already exists in config
	existingSite := a.siteConfig.GetSiteByName(site.Name)

	// If site exists, preserve width and height
	if existingSite != nil {
		// Preserve width and height if they exist
		if existingSite.Width > 0 {
			site.Width = existingSite.Width
		}
		if existingSite.Height > 0 {
			site.Height = existingSite.Height
		}
	}

	// Set site as active
	site.IsActive = true

	// Add site to config
	a.siteConfig.AddSite(*site)
	a.siteConfig.SetActiveSite(site.Name)

	// Check if icon URL is specified
	if site.Icon != "" {
		// Get icon filename from URL
		iconName := a.execName + ".ico"

		// If icon URL contains a filename, use that instead
		if strings.Contains(site.Icon, "/") {
			parts := strings.Split(site.Icon, "/")
			lastPart := parts[len(parts)-1]
			if strings.HasSuffix(lastPart, ".ico") {
				iconName = lastPart
			}
		}

		// Check if icon exists in AppData
		iconPath := filepath.Join(a.iconsDir, iconName)
		iconExists := false
		if _, err := os.Stat(iconPath); !os.IsNotExist(err) {
			iconExists = true
		}

		// If icon doesn't exist, try to download it or copy from resources
		if !iconExists {
			// First try to ensure the icon exists in the resources
			err := resources.EnsureIconExists(iconName, a.iconsDir)
			if err != nil {
				// If not in resources, try to download it
				a.downloadIcon(site.Icon, a.execName)
			}
		}

		// Set icon change flag and launch icon changer if needed
		if !a.forceMode {
			a.iconChanged = true
			a.launchIconChanger(iconPath)
		}
	}

	// Save to AppData
	a.siteConfig.SaveToFile(appDataSitesPath)
	a.currentSite = a.siteConfig.GetSiteByName(site.Name)
}

// downloadIcon downloads an icon from a URL
func (a *App) downloadIcon(iconURL, name string) {
	// Create icon path
	iconPath := filepath.Join(a.iconsDir, name+".ico")

	// Check if icon already exists
	iconExists := false
	if _, err := os.Stat(iconPath); !os.IsNotExist(err) {
		iconExists = true
	}

	// If icon doesn't exist, download and convert it
	if !iconExists {
		// Download icon
		tempPath := filepath.Join(a.iconsDir, "temp_"+name)
		if err := utils.DownloadFile(iconURL, tempPath); err != nil {
			return
		}

		// Check if downloaded file is an ICO file
		if !utils.IsICOFile(tempPath) {
			// Convert to ICO
			rceditPath := filepath.Join(a.appDataDir, "rcedit.exe")
			if err := utils.ConvertToICO(tempPath, rceditPath, iconPath); err != nil {
				// If conversion fails, use default icon
				defaultIconPath := filepath.Join(a.iconsDir, "hobaa.ico")
				if _, err := os.Stat(defaultIconPath); !os.IsNotExist(err) {
					resources.CopyFile(defaultIconPath, iconPath)
				}
			}
			// Remove temporary file
			os.Remove(tempPath)
		} else {
			// Rename file
			os.Rename(tempPath, iconPath)
		}
	}

	// Set icon change flag
	a.iconChanged = true

	// If not in force mode, launch icon changer
	if !a.forceMode {
		a.launchIconChanger(iconPath)
	}
}

// launchIconChanger launches a copy of the application to change the icon
func (a *App) launchIconChanger(iconPath string) {
	// Create a copy of the executable in AppData
	appDataExePath := filepath.Join(a.appDataDir, "hobaa_icon_changer.exe")

	// Check if the icon changer already exists
	if _, err := os.Stat(appDataExePath); os.IsNotExist(err) {
		// Copy the current executable to AppData only if it doesn't exist
		if err := copyFile(a.execPath, appDataExePath); err != nil {
			fmt.Printf("Failed to copy executable: %v\n", err)
			return
		}
	}

	// Launch the copy with change-icon parameter
	cmd := exec.Command(appDataExePath,
		"--change-icon",
		"--target-exe", a.execPath,
		"--icon-path", iconPath)

	if err := cmd.Start(); err != nil {
		fmt.Printf("Failed to launch icon changer: %v\n", err)
		return
	}

	// Exit the current process
	os.Exit(0)
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

// downloadFavicon downloads a favicon from a website
func (a *App) downloadFavicon(websiteURL, name string) {
	// Get favicon URL
	faviconURL, err := utils.GetFaviconURL(websiteURL)
	if err != nil {
		// If favicon URL cannot be retrieved, first check if the icon exists in resources
		iconPath := filepath.Join(a.iconsDir, name+".ico")

		// Try to ensure the icon exists in resources
		err := resources.EnsureIconExists(name+".ico", a.iconsDir)
		if err != nil {
			// If not in resources, use default icon
			defaultIconPath := filepath.Join(a.iconsDir, "hobaa.ico")

			// Copy default icon to site-specific icon
			if _, err := os.Stat(defaultIconPath); !os.IsNotExist(err) {
				resources.CopyFile(defaultIconPath, iconPath)
			}
		}

		// Set icon change flag
		a.iconChanged = true

		// If not in force mode, launch icon changer
		if !a.forceMode {
			a.launchIconChanger(iconPath)
		}
		return
	}

	// Download icon
	a.downloadIcon(faviconURL, name)

	// Update site configuration with icon URL
	if a.currentSite != nil {
		a.currentSite.Icon = faviconURL

		// Save to AppData
		appDataSitesPath := config.GetAppDataSitesPath(a.appDataDir)
		a.siteConfig.SaveToFile(appDataSitesPath)
	}
}

// clearWindowsCache clears Windows application cache without restarting Explorer
func (a *App) clearWindowsCache() {
	// Use Shell API to clear icon cache
	winapi.ClearIconCache()
}

// restartWithForce restarts the application with the --force parameter
func (a *App) restartWithForce() {
	utils.RestartApplication()
}

// SaveWindowSizeToConfig saves the window size to the site configuration
func (a *App) SaveWindowSizeToConfig(width, height int) error {
	// Update the current site with the new dimensions
	if a.currentSite != nil {
		a.currentSite.Width = width
		a.currentSite.Height = height

		// Save to AppData
		appDataSitesPath := config.GetAppDataSitesPath(a.appDataDir)
		return a.siteConfig.SaveToFile(appDataSitesPath)
	}

	return nil
}

// Run starts the application
func (a *App) Run() {
	// If in change-icon mode, don't run the application
	if a.changeIcon {
		return
	}

	// Check if site exists and is active, or if force mode is enabled
	if (a.currentSite != nil && a.currentSite.IsActive) || a.forceMode {
		// Set default title, URL, and dimensions
		title := "Hobaa"
		url := "https://www.google.com"
		width := 1920
		height := 1080

		// Use site configuration if available
		if a.currentSite != nil {
			if a.currentSite.Title != "" {
				title = a.currentSite.Title
			}
			if a.currentSite.URL != "" {
				url = a.currentSite.URL
			}
			if a.currentSite.Width > 0 {
				width = a.currentSite.Width
			}
			if a.currentSite.Height > 0 {
				height = a.currentSite.Height
			}
		}

		// Validate URL
		if !utils.IsValidURL(url) {
			// If URL is not valid, use Google
			url = "https://www.google.com"
		}

		// Capitalize first letter of title
		if len(title) > 0 {
			title = strings.ToUpper(title[:1]) + title[1:]
		}

		// Get icon path for the window title
		iconPath := filepath.Join(a.iconsDir, a.execName+".ico")
		if _, err := os.Stat(iconPath); os.IsNotExist(err) {
			// Use default icon if specific icon doesn't exist
			iconPath = filepath.Join(a.iconsDir, "hobaa.ico")
		}

		// Create webview
		a.webView = webview.New(webview.WindowOptions{
			Title:   title,
			URL:     url,
			Width:   width,
			Height:  height,
			Debug:   true,
			Icon:    iconPath,
			DataDir: a.webViewDir, // Set WebView data directory
		})
		defer a.webView.Destroy()

		// Get window handle
		a.hwnd = syscall.Handle(uintptr(a.webView.Window()))

		// Start monitoring window size in a goroutine
		go winapi.MonitorWindowSize(a.hwnd, func(width, height int) {
			a.SaveWindowSizeToConfig(width, height)
		})

		// Run webview
		a.webView.Run()
	}
}
