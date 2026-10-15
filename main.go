//go:build !cli

package main

import (
	"embed"
	"fmt"
	"log"
	"os"
	"strings"

	"app/core"
	"app/internal/cli"
	"app/internal/gui"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

//go:embed all:frontend/dist
var assets embed.FS

var title = "PhotiOS (Google Photos for iOS) — Unlimited Storage"

func main() {
	// Run the CLI when the arguments name one of its commands; otherwise start the GUI.
	if cli.IsCLIInvocation(os.Args[1:]) {
		os.Exit(cli.Run(os.Args[1:], cli.Info{
			ExecutableName: "PhotiOS",
			HasGUI:         true,
			Version:        getAppVersion(),
		}))
	}

	runGUI()
}

func runGUI() {
	normalizeFrontendDevServerURL()

	configManager := &core.ConfigManager{}

	// Create iPhone service with event emitter (will be set after app is created)
	var iphoneService *core.IPhoneService

	wailsApp := application.New(application.Options{
		Name:        "com.photios.app",
		Description: "PhotiOS (Google Photos for iOS) - Unlimited Storage",
		Services: []application.Service{
			application.NewService(configManager),
		},
		Assets: application.AssetOptions{
			Handler: application.BundledAssetFileServer(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	// Initialize iPhone service with event emitter
	iphoneService = core.NewIPhoneService(configManager, wailsApp.Logger, func(name string, data interface{}) {
		wailsApp.Event.Emit(name, data)
	})

	// Register iPhone service with Wails
	wailsApp.RegisterService(application.NewService(iphoneService))

	window := wailsApp.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:               title,
		Frameless:           false,
		Width:               480,
		Height:              700,
		EnableFileDrop:      true,
		DisableResize:       false,
		MaximiseButtonState: application.ButtonDisabled,
		BackgroundType:      application.BackgroundTypeTranslucent,
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 0,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		URL: "/",
	})

	uploadManager := core.NewUploadManager(gui.NewWailsReporter(wailsApp), wailsApp.Logger)

	// Start iPhone detection
	iphoneService.StartDetection()

	// Listen for upload cancel event
	wailsApp.Event.On("uploadCancel", func(e *application.CustomEvent) {
		uploadManager.Cancel()
	})

	// Listen for iPhone scan request
	wailsApp.Event.On("iphone-scan-request", func(e *application.CustomEvent) {
		wailsApp.Logger.Info("iPhone scan requested")
		if err := iphoneService.ScanPhotos(); err != nil {
			wailsApp.Logger.Error("Failed to start iPhone scan", "error", err)
		}
	})

	// Listen for iPhone scan cancel
	wailsApp.Event.On("iphone-scan-cancel", func(e *application.CustomEvent) {
		iphoneService.CancelScan()
	})

	// Listen for iPhone upload request
	wailsApp.Event.On("startIPhoneUpload", func(e *application.CustomEvent) {
		var paths []string
		if data, ok := e.Data.(core.StartUploadEvent); ok && len(data.Files) > 0 {
			paths = iphoneService.GetSelectedPhotoPaths(data.Files)
		} else {
			paths = iphoneService.GetUnsyncedPhotoPaths()
		}
		wailsApp.Logger.Info("Starting iPhone upload", "fileCount", len(paths))
		if len(paths) > 0 {
			uploadManager.Upload(paths, configManager.SessionUploadOptions())
		}
	})

	window.OnWindowEvent(events.Common.WindowFilesDropped, func(event *application.WindowEvent) {
		paths := event.Context().DroppedFiles()
		dropTarget := event.Context().DropTargetDetails()

		var dropZone string
		if dropTarget != nil {
			dropZone = dropTarget.Attributes["data-drop-zone"]
			wailsApp.Logger.Info("Drop target detected",
				"dropZone", dropZone,
				"elementID", dropTarget.ElementID)
		}

		// Emit event to frontend with drop details
		wailsApp.Event.Emit("files-dropped", core.FilesDroppedEvent{
			Files:    paths,
			DropZone: dropZone,
		})
	})

	// Listen for upload request from frontend (after drop zone is determined)
	wailsApp.Event.On("startUpload", func(e *application.CustomEvent) {
		if data, ok := e.Data.(core.StartUploadEvent); ok {
			wailsApp.Logger.Info("Starting upload", "fileCount", len(data.Files))
			uploadManager.Upload(data.Files, configManager.SessionUploadOptions())
		} else {
			wailsApp.Logger.Error("startUpload: unexpected data type", "type", fmt.Sprintf("%T", e.Data))
		}
	})

	err := wailsApp.Run()
	if err != nil {
		iphoneService.StopDetection()
		log.Fatal(err)
	}
	iphoneService.StopDetection()
}

func normalizeFrontendDevServerURL() {
	const envName = "FRONTEND_DEVSERVER_URL"

	value := os.Getenv(envName)
	value = strings.Replace(value, "http://localhost:", "http://127.0.0.1:", 1)
	value = strings.Replace(value, "https://localhost:", "https://127.0.0.1:", 1)
	if value != os.Getenv(envName) {
		_ = os.Setenv(envName, value)
	}
}
