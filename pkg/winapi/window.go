// Package winapi provides Windows API functions for the application
package winapi

import (
	"syscall"
	"time"
	"unsafe"
)

// RECT represents a Windows RECT structure
type RECT struct {
	Left   int32
	Top    int32
	Right  int32
	Bottom int32
}

// Shell API constants
const (
	SHCNE_ASSOCCHANGED = 0x08000000
	SHCNF_IDLIST       = 0x0000
	
	// Icon constants
	ICON_SMALL  = 0
	ICON_BIG    = 1
	ICON_SMALL2 = 2
	
	// Window messages
	WM_SETICON = 0x0080
	
	// Image loading constants
	LR_LOADFROMFILE = 0x0010
	IMAGE_ICON      = 1
)

var (
	user32              = syscall.NewLazyDLL("user32.dll")
	procGetWindowRect   = user32.NewProc("GetWindowRect")
	procGetClientRect   = user32.NewProc("GetClientRect")
	procGetWindowLongW  = user32.NewProc("GetWindowLongW")
	procSetWindowLongW  = user32.NewProc("SetWindowLongW")
	procSetWindowPos    = user32.NewProc("SetWindowPos")
	procGetWindowTextW  = user32.NewProc("GetWindowTextW")
	procSetWindowTextW  = user32.NewProc("SetWindowTextW")
	procLoadImageW      = user32.NewProc("LoadImageW")
	procSendMessageW    = user32.NewProc("SendMessageW")
	
	shell32             = syscall.NewLazyDLL("shell32.dll")
	shChangeNotify      = shell32.NewProc("SHChangeNotify")
)

// GetWindowSize gets the window size using Windows API
func GetWindowSize(hwnd syscall.Handle) (int, int, error) {
	var rect RECT
	ret, _, err := procGetWindowRect.Call(
		uintptr(hwnd),
		uintptr(unsafe.Pointer(&rect)),
	)
	
	if ret == 0 {
		return 0, 0, err
	}
	
	width := int(rect.Right - rect.Left)
	height := int(rect.Bottom - rect.Top)
	
	return width, height, nil
}

// MonitorWindowSize periodically checks the window size and calls the callback when changed
func MonitorWindowSize(hwnd syscall.Handle, callback func(width, height int)) {
	// Wait a bit for the window to fully initialize
	time.Sleep(1 * time.Second)
	
	lastWidth, lastHeight := 0, 0
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	
	for range ticker.C {
		// Check if the window still exists
		if hwnd == 0 {
			return
		}
		
		// Get the current window size
		width, height, err := GetWindowSize(hwnd)
		if err != nil {
			continue
		}
		
		// If the size has changed, call the callback
		if width != lastWidth || height != lastHeight {
			lastWidth, lastHeight = width, height
			callback(width, height)
		}
	}
}

// ClearIconCache clears the Windows icon cache using Shell API
func ClearIconCache() {
	shChangeNotify.Call(
		uintptr(SHCNE_ASSOCCHANGED),
		uintptr(SHCNF_IDLIST),
		0,
		0,
	)
}

// Sleep waits for the specified milliseconds
func Sleep(milliseconds int) {
	time.Sleep(time.Duration(milliseconds) * time.Millisecond)
}

// SetWindowIcon sets the icon for a window
func SetWindowIcon(hwnd uintptr, iconPath string) {
	// Convert icon path to UTF16
	iconPathW, _ := syscall.UTF16PtrFromString(iconPath)
	
	// Load small icon (16x16 for taskbar)
	hIconSmall, _, _ := procLoadImageW.Call(
		0,
		uintptr(unsafe.Pointer(iconPathW)),
		IMAGE_ICON,
		16, // Small icon size
		16,
		LR_LOADFROMFILE,
	)
	
	// Load medium icon (32x32 for alt+tab)
	hIconMedium, _, _ := procLoadImageW.Call(
		0,
		uintptr(unsafe.Pointer(iconPathW)),
		IMAGE_ICON,
		32, // Medium icon size
		32,
		LR_LOADFROMFILE,
	)
	
	// Load large icon (48x48 for high DPI displays)
	hIconLarge, _, _ := procLoadImageW.Call(
		0,
		uintptr(unsafe.Pointer(iconPathW)),
		IMAGE_ICON,
		48, // Large icon size
		48,
		LR_LOADFROMFILE,
	)
	
	// Load extra large icon (256x256 for Windows 10/11)
	hIconExtraLarge, _, _ := procLoadImageW.Call(
		0,
		uintptr(unsafe.Pointer(iconPathW)),
		IMAGE_ICON,
		256, // Extra large icon size
		256,
		LR_LOADFROMFILE,
	)
	
	// Set small icon (for window caption)
	if hIconSmall != 0 {
		procSendMessageW.Call(
			hwnd,
			WM_SETICON,
			ICON_SMALL,
			hIconSmall,
		)
	}
	
	// Set small icon 2 (for taskbar)
	if hIconMedium != 0 {
		procSendMessageW.Call(
			hwnd,
			WM_SETICON,
			ICON_SMALL2,
			hIconMedium,
		)
	}
	
	// Set big icon (for alt+tab)
	if hIconLarge != 0 {
		procSendMessageW.Call(
			hwnd,
			WM_SETICON,
			ICON_BIG,
			hIconLarge,
		)
	} else if hIconExtraLarge != 0 {
		// If large icon failed, try extra large
		procSendMessageW.Call(
			hwnd,
			WM_SETICON,
			ICON_BIG,
			hIconExtraLarge,
		)
	}
} 