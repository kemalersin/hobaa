@echo off
echo Building Hobaa application...

REM Set environment variables for optimized build
set CGO_ENABLED=0
set GOOS=windows
set GOARCH=amd64

REM Generate resources using go-winres
echo Generating resources with go-winres...
go-winres make --out rsrc.syso --no-suffix

REM Check if resource generation was successful
if %ERRORLEVEL% NEQ 0 (
    echo Resource generation failed with error code %ERRORLEVEL%
    exit /b %ERRORLEVEL%
)

echo Cleaning up old builds...
if exist hobaa.exe del hobaa.exe

echo Running go mod tidy...
go mod tidy

echo Building executable with optimizations...
go build -ldflags="-s -w -H=windowsgui" -trimpath -buildvcs=false

REM Check if build was successful
if %ERRORLEVEL% NEQ 0 (
    echo Build failed with error code %ERRORLEVEL%
    exit /b %ERRORLEVEL%
)

echo Checking UPX...
where upx >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo UPX not found in PATH. Skipping compression...
) else (
    echo Compressing with UPX...
    upx --best --no-backup --no-progress hobaa.exe
)

echo Build completed successfully!

REM Optional: Create test copies if "test" parameter is provided
if "%1"=="test" (
    echo Creating test copies...
    if exist github.exe del github.exe
    if exist twitter.exe del twitter.exe
    copy hobaa.exe github.exe
    copy hobaa.exe twitter.exe
    echo Test copies created!
) else (
    echo Test copies not created. Use "build.bat test" to create test copies.
)

echo Build completed!
echo Final file size:
dir hobaa.exe | find "hobaa.exe"