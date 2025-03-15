# Development Guide

This document provides information for developers who want to contribute to or modify the Hobaa application.

## Project Structure

```
hobaa/
├── .cursor/           # Cursor IDE configuration
├── pkg/               # Go packages
│   ├── app/           # Main application logic
│   ├── dpi/           # DPI awareness functionality
│   ├── webview/       # WebView wrapper
│   ├── config/        # Configuration handling
│   └── utils/         # Utility functions
├── resources/         # Application resources
│   ├── icons/         # Icon files
│   ├── default.ico    # Default application icon
│   └── rcedit.exe     # Resource editor tool
├── winres/            # Windows resource files
│   └── winres.json    # Resource configuration
├── main.go            # Application entry point
├── go.mod             # Go module definition
├── go.sum             # Go module checksums
├── build.bat          # Build script
├── sites.json         # Website configurations
├── LICENSE            # MIT license
└── README.md          # Project documentation
```

## Build Process

The build process is managed by the `build.bat` script, which performs the following steps:

1. Sets environment variables for optimized builds
2. Generates Windows resources using go-winres
3. Cleans up old builds
4. Updates Go dependencies
5. Builds the executable with optimizations
6. Compresses the executable with UPX (if available)
7. Optionally creates test copies with different names

### Resource Generation

Windows resources (icons, version info, manifest) are defined in `winres/winres.json` and processed by go-winres. The configuration includes:

- Application icon
- Version information
- Application manifest with DPI awareness settings
- Application description and copyright information

### Optimization Flags

The build uses several optimization flags:

- `-ldflags="-s -w -H=windowsgui"`: Strips debug information and sets the application to run as a GUI application without a console window
- `-trimpath`: Removes file path information from the binary
- `-buildvcs=false`: Disables embedding of version control information

## Module Structure

### app

The `app` package contains the main application logic, including:

- Application initialization
- WebView creation and configuration
- Application lifecycle management

### dpi

The `dpi` package handles DPI awareness for high-resolution displays:

- Sets process DPI awareness using Windows API
- Provides fallback methods for different Windows versions

### webview

The `webview` package wraps the WebView2 functionality:

- Creates and manages the WebView window
- Handles navigation
- Provides utility functions for working with the WebView

### config

The `config` package (planned) will handle configuration management:

- Loading site configurations from sites.json
- Managing user preferences
- Handling application settings

### utils

The `utils` package (planned) will provide utility functions:

- File operations
- Icon handling
- Path management

## Adding New Features

When adding new features, follow these guidelines:

1. Maintain modular structure by placing code in appropriate packages
2. Keep package line count under 500 lines
3. Write reusable code and avoid repetition
4. Update documentation to reflect changes
5. Follow Go best practices and coding conventions

## Testing

To test the application with different website configurations:

1. Build the application with the test parameter: `build.bat test`
2. This creates test copies (github.exe, twitter.exe)
3. Run these executables to verify that they load the correct websites

## Debugging

For debugging:

1. Set `Debug: true` in the WebView options
2. Build without the windowsgui flag to see console output
3. Use Go's built-in debugging tools or Delve debugger

## Release Process

To create a release:

1. Update version information in winres.json
2. Run the build script
3. Test the application thoroughly
4. Create a GitHub release with the compiled executable 