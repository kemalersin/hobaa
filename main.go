package main

import (
	"embed"

	"github.com/kemalersin/hobaa/pkg/app"
	"github.com/kemalersin/hobaa/pkg/resources"
)

//go:embed resources/default.ico resources/rcedit.exe resources/icons/ico/* sites.json
var embeddedFiles embed.FS

func main() {
	// Set embedded files
	resources.SetEmbeddedFiles(embeddedFiles)

	// Create and run application
	application := app.New()
	application.Run()
}
