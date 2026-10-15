//go:build !cli

package core

import (
	"encoding/base64"
	"log/slog"
	"sync"
)

// IPhoneService is a Wails-bound service that exposes iPhone functionality
// to the frontend. It coordinates detection, scanning, and upload of
// iPhone photos.
type IPhoneService struct {
	mu            sync.Mutex
	detector      *IPhoneDetector
	scanner       *IPhoneScanner
	configManager *ConfigManager
	logger        *slog.Logger
	emitEvent     func(name string, data interface{})

	// Cached scan results for the frontend
	lastScanResult *IPhoneScanResult
}

// NewIPhoneService creates an iPhone service that uses the given event emitter.
func NewIPhoneService(configManager *ConfigManager, logger *slog.Logger, emitEvent func(name string, data interface{})) *IPhoneService {
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}
	svc := &IPhoneService{
		configManager: configManager,
		logger:        logger,
		emitEvent:     emitEvent,
	}

	// Set up iPhone detection
	svc.detector = NewIPhoneDetector(logger,
		func(status IPhoneStatus) {
			svc.emitEvent("iphone-connected", status)
		},
		func() {
			svc.emitEvent("iphone-disconnected", nil)
		},
	)

	return svc
}

// StartDetection begins polling for iPhone connections. Called once at app startup.
func (s *IPhoneService) StartDetection() {
	s.detector.Start()
}

// StopDetection stops the detection polling.
func (s *IPhoneService) StopDetection() {
	s.detector.Stop()
}

// GetIPhoneStatus returns the current iPhone connection status.
func (s *IPhoneService) GetIPhoneStatus() IPhoneStatus {
	return s.detector.Status()
}

// ScanPhotos initiates a scan of the connected iPhone's photos.
// Progress and results are emitted as events.
func (s *IPhoneService) ScanPhotos() error {
	status := s.detector.Status()
	if !status.Connected {
		return nil
	}

	// Cancel any existing scan
	s.mu.Lock()
	if s.scanner != nil {
		s.scanner.Cancel()
	}
	s.mu.Unlock()

	// Get API options from current config
	apiOpts := s.configManager.SessionUploadOptions().Api

	scanner := NewIPhoneScanner(s.logger,
		func(progress IPhoneScanProgress) {
			s.emitEvent("iphone-scan-progress", progress)
		},
		func(result IPhoneScanResult) {
			s.mu.Lock()
			s.lastScanResult = &result
			s.mu.Unlock()
			s.emitEvent("iphone-scan-complete", result)
		},
	)

	s.mu.Lock()
	s.scanner = scanner
	s.mu.Unlock()

	scanner.ScanIPhonePhotos(status.DevicePath, apiOpts)
	return nil
}

// CancelScan cancels an in-progress scan.
func (s *IPhoneService) CancelScan() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.scanner != nil {
		s.scanner.Cancel()
	}
}

// GetThumbnail returns a base64-encoded JPEG thumbnail for a photo.
func (s *IPhoneService) GetThumbnail(path string) string {
	data, err := GenerateThumbnail(path, 320)
	if err != nil {
		s.logger.Debug("Thumbnail generation failed", "path", path, "error", err)
		return ""
	}
	return "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(data)
}

// GetPreview returns a higher-resolution base64-encoded JPEG preview (720px) for lightbox/modal preview.
func (s *IPhoneService) GetPreview(path string) string {
	data, err := GenerateThumbnail(path, 720)
	if err != nil {
		s.logger.Debug("Preview generation failed", "path", path, "error", err)
		return ""
	}
	return "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(data)
}

// GetUnsyncedPhotoPaths returns the file paths from the last scan result.
// Used to feed into the existing upload pipeline.
func (s *IPhoneService) GetUnsyncedPhotoPaths() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.lastScanResult == nil {
		return nil
	}
	paths := make([]string, len(s.lastScanResult.Photos))
	for i, photo := range s.lastScanResult.Photos {
		paths[i] = photo.Path
	}
	return paths
}

// GetSelectedPhotoPaths filters scan results to the given paths.
func (s *IPhoneService) GetSelectedPhotoPaths(selectedPaths []string) []string {
	// Validate paths exist in scan results
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.lastScanResult == nil {
		return nil
	}

	validPaths := make(map[string]bool)
	for _, photo := range s.lastScanResult.Photos {
		validPaths[photo.Path] = true
	}

	result := make([]string, 0, len(selectedPaths))
	for _, path := range selectedPaths {
		if validPaths[path] {
			result = append(result, path)
		}
	}
	return result
}
