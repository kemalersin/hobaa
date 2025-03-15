package dpi

import (
	"syscall"
)

// Required definitions for Windows API functions
var (
	user32                 = syscall.NewLazyDLL("user32.dll")
	setProcessDPIAware     = user32.NewProc("SetProcessDPIAware")
	shcore                 = syscall.NewLazyDLL("shcore.dll")
	setProcessDpiAwareness = shcore.NewProc("SetProcessDpiAwareness")
)

// DPI awareness levels
const (
	PROCESS_DPI_UNAWARE           = 0
	PROCESS_SYSTEM_DPI_AWARE      = 1
	PROCESS_PER_MONITOR_DPI_AWARE = 2
)

// SetProcessDpiAwareness enables high DPI support for the application
func SetProcessDpiAwareness() {
	// For Windows 8.1
	_, _, _ = setProcessDpiAwareness.Call(PROCESS_SYSTEM_DPI_AWARE) // Use system DPI
	
	// Fallback method for Windows Vista and later
	_, _, _ = setProcessDPIAware.Call()
} 