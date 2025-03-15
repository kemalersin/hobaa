package webview

import (
	"github.com/jchv/go-webview2"
	"github.com/kemalersin/hobaa/pkg/winapi"
	"os"
	"path/filepath"
	"unsafe"
)

// WebView represents a webview window
type WebView struct {
	window webview2.WebView
}

// WindowOptions contains options for creating a webview window
type WindowOptions struct {
	Title    string
	URL      string
	Width    int
	Height   int
	Debug    bool
	Icon     string // Path to icon file
	DataDir  string // Path to WebView data directory
}

// New creates a new webview with the given options
func New(options WindowOptions) *WebView {
	// Set default values if not provided
	if options.Width <= 0 {
		options.Width = 1920
	}
	if options.Height <= 0 {
		options.Height = 1080
	}

	// Create webview
	w := webview2.NewWithOptions(webview2.WebViewOptions{
		Debug:     options.Debug,
		AutoFocus: true,
		DataPath:  options.DataDir, // Set WebView data directory
		WindowOptions: webview2.WindowOptions{
			Title:  options.Title,
			Width:  uint(options.Width),
			Height: uint(options.Height),
			Center: true,
		},
	})

	// Create WebView instance
	webView := &WebView{
		window: w,
	}

	// Set icon if provided
	if options.Icon != "" {
		webView.SetIcon(options.Icon)
	}
	
	// Inject back button script
	injectBackButtonScript(w)

	// Navigate to URL if provided
	if options.URL != "" {
		webView.Navigate(options.URL)
	}

	return webView
}

// Navigate navigates to the specified URL
func (w *WebView) Navigate(url string) {
	w.window.Navigate(url)
}

// Run starts the webview main loop
func (w *WebView) Run() {
	w.window.Run()
}

// Destroy destroys the webview
func (w *WebView) Destroy() {
	w.window.Destroy()
}

// Window returns the native window handle
func (w *WebView) Window() unsafe.Pointer {
	return w.window.Window()
}

// SetIcon sets the window icon
func (w *WebView) SetIcon(iconPath string) {
	// Check if icon exists
	if _, err := os.Stat(iconPath); os.IsNotExist(err) {
		return
	}
	
	// Set window icon using Windows API
	hwnd := w.Window()
	if hwnd != nil {
		// Convert to uintptr
		handle := uintptr(hwnd)
		// Set icon
		winapi.SetWindowIcon(handle, iconPath)
	}
}

// Init initializes the webview with JavaScript
func (w *WebView) Init(js string) {
	w.window.Init(js)
}

// injectBackButtonScript injects JavaScript to add a back button to the WebView
func injectBackButtonScript(w webview2.WebView) {
	// Wait for the DOM to be loaded
	w.Init(`
		// Create and inject CSS for back button
		function createBackButtonStyles() {
			const style = document.createElement('style');
			style.textContent = ` + "`" + `
				#hobaa-back-button {
					position: absolute;
					top: 20px;
					left: 20px;
					width: 40px;
					height: 40px;
					border-radius: 50%;
					background-color: rgba(0, 0, 0, 0.6);
					color: white;
					display: flex;
					align-items: center;
					justify-content: center;
					cursor: pointer;
					z-index: 9999;
					opacity: 0;
					transition: opacity 0.3s ease;
					font-size: 24px;
					border: none;
					outline: none;
					box-shadow: 0 2px 5px rgba(0, 0, 0, 0.2);
					line-height: 1; /* Fix line height for vertical alignment */
					padding-bottom: 3px; /* Fine-tune vertical centering of the arrow */
				}
				
				#hobaa-back-button:hover {
					background-color: rgba(0, 0, 0, 0.8);
				}
				
				#hobaa-back-button-container {
					position: absolute;
					top: 0;
					left: 0;
					width: 80px;
					height: 80px;
					z-index: 9998;
					display: none; /* Hidden by default */
				}
				
				#hobaa-back-button-container:hover #hobaa-back-button {
					opacity: 1;
				}
			` + "`" + `;
			document.head.appendChild(style);
		}
		
		// Create back button element
		function createBackButton() {
			// Create container for hover detection
			const container = document.createElement('div');
			container.id = 'hobaa-back-button-container';
			
			// Create the button
			const button = document.createElement('button');
			button.id = 'hobaa-back-button';
			button.innerHTML = '&#8592;'; // Left arrow
			button.addEventListener('click', () => {
				history.back();
			});
			
			// Add button to container
			container.appendChild(button);
			
			// Add container to body
			document.body.appendChild(container);
			
			// Update back button visibility
			updateBackButtonVisibility();
		}
		
		// Update back button visibility
		function updateBackButtonVisibility() {
			const container = document.getElementById('hobaa-back-button-container');
			if (!container) return;
			
			// Check if there is a page to go back to
			const canGoBack = window.history.length > 1 && document.referrer !== '';
			container.style.display = canGoBack ? 'block' : 'none';
		}
		
		// Initialize back button when DOM is loaded
		function initBackButton() {
			createBackButtonStyles();
			createBackButton();
			updateBackButtonVisibility();
		}
		
		// Check if document is already loaded
		if (document.readyState === 'complete' || document.readyState === 'interactive') {
			initBackButton();
		} else {
			document.addEventListener('DOMContentLoaded', initBackButton);
		}
		
		// Reinitialize back button when page changes
		window.addEventListener('popstate', () => {
			// Remove existing button if any
			const existingContainer = document.getElementById('hobaa-back-button-container');
			if (existingContainer) {
				existingContainer.remove();
			}
			
			// Create new button
			setTimeout(() => {
				initBackButton();
			}, 300);
		});
		
		// Also handle page loads
		window.addEventListener('load', () => {
			// Remove existing button if any
			const existingContainer = document.getElementById('hobaa-back-button-container');
			if (existingContainer) {
				existingContainer.remove();
			}
			
			// Create new button
			setTimeout(() => {
				initBackButton();
			}, 300);
		});
		
		// Monitor in-page navigation
		const originalPushState = history.pushState;
		history.pushState = function() {
			originalPushState.apply(this, arguments);
			setTimeout(updateBackButtonVisibility, 100);
		};
		
		const originalReplaceState = history.replaceState;
		history.replaceState = function() {
			originalReplaceState.apply(this, arguments);
			setTimeout(updateBackButtonVisibility, 100);
		};
	`);
}

// GetExecutablePath returns the path of the current executable
func GetExecutablePath() string {
	// Get the path of the current executable
	exePath, err := os.Executable()
	if err != nil {
		return ""
	}
	return exePath
}

// GetExecutableDir returns the directory of the current executable
func GetExecutableDir() string {
	// Get the path of the current executable
	exePath := GetExecutablePath()
	if exePath == "" {
		return ""
	}
	return filepath.Dir(exePath)
} 