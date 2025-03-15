# Hobaa!

Hobaa is a tool that allows you to use websites as desktop applications. The application can open different websites based on the name of the EXE file.

## Features

- Access multiple websites with a single application
- Open different websites based on the EXE file name
- Automatic icon download and application
- Customizable window sizes
- Back button support

## Technology

- Go programming language
- WebView2 (github.com/jchv/go-webview2)
- Windows API (golang.org/x/sys)

## Build Instructions

### Prerequisites

- Go 1.16 or higher
- go-winres (`go install github.com/tc-hib/go-winres@latest`)
- UPX (optional, for executable compression)

### Building from Source

1. Clone the repository:
   ```
   git clone https://github.com/kemalersin/hobaa.git
   cd hobaa
   ```

2. Run the build script:
   ```
   build.bat
   ```

   This will:
   - Generate resources using go-winres
   - Clean up old builds
   - Run go mod tidy to manage dependencies
   - Build the executable with optimizations
   - Compress the executable with UPX (if available)

3. For testing with multiple website configurations:
   ```
   build.bat test
   ```
   
   This will create additional test executables (github.exe, twitter.exe) that can be used to test the application's ability to load different websites based on the executable name.

### Build Options

The build script uses the following optimizations:

- CGO_ENABLED=0: Disables CGO for a more portable build
- GOOS=windows and GOARCH=amd64: Targets 64-bit Windows
- -ldflags="-s -w -H=windowsgui": Strips debug information and hides the console window
- -trimpath: Removes file path information from the binary
- UPX compression: Reduces the executable size (when available)

## Usage

1. Download the application
2. Rename the EXE file according to the website you want to open (e.g., youtube.exe, twitter.exe)
3. Run the application

## Supported Sites

The application supports all sites defined in the sites.json file. By default, the following sites are supported:

- Google
- X (Twitter)
- Instagram
- LinkedIn
- Facebook
- YouTube
- Reddit
- GitHub
- Gmail
- Google Maps
- Google Drive
- Google Docs
- Dropbox
- ChatGPT
- Ekşisözlük

## License

This project is licensed under the MIT License. 