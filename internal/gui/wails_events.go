//go:build !cli

package gui

import (
	"app/core"
	"github.com/wailsapp/wails/v3/pkg/application"
)

func init() {
	application.RegisterEvent[core.AlbumStatus]("albumProgress")
	application.RegisterEvent[core.AlbumStatus]("albumComplete")
	application.RegisterEvent[core.AlbumError]("albumError")
	application.RegisterEvent[core.UploadBatchStart]("uploadStart")
	application.RegisterEvent[application.Void]("uploadStop")
	application.RegisterEvent[core.FileUploadResult]("FileStatus")
	application.RegisterEvent[core.ThreadStatus]("ThreadStatus")
	application.RegisterEvent[core.PreflightWarning]("uploadWarning")
	application.RegisterEvent[application.Void]("uploadCancel")
	application.RegisterEvent[int64]("uploadTotalBytes")
	application.RegisterEvent[int64]("uploadTotalBytesDelta")
	application.RegisterEvent[core.FilesDroppedEvent]("files-dropped")
	application.RegisterEvent[core.StartUploadEvent]("startUpload")

	// iPhone events
	application.RegisterEvent[core.IPhoneStatus]("iphone-connected")
	application.RegisterEvent[application.Void]("iphone-disconnected")
	application.RegisterEvent[core.IPhoneScanProgress]("iphone-scan-progress")
	application.RegisterEvent[core.IPhoneScanResult]("iphone-scan-complete")
	application.RegisterEvent[application.Void]("iphone-scan-request")
	application.RegisterEvent[application.Void]("iphone-scan-cancel")
	application.RegisterEvent[core.StartUploadEvent]("startIPhoneUpload")
}

