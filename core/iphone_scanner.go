//go:build !cli

package core

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	// Register decoders for thumbnail generation
	_ "image/gif"
	_ "image/png"
)

// IPhonePhoto represents a photo found on the connected iPhone.
type IPhonePhoto struct {
	Path      string `json:"path"`
	FileName  string `json:"fileName"`
	Size      int64  `json:"size"`
	Extension string `json:"extension"`
	// Thumbnail is a base64-encoded JPEG data URI (populated lazily or for previews)
	Thumbnail string `json:"thumbnail,omitempty"`
	// IsSynced indicates whether this photo already exists in Google Photos
	IsSynced bool `json:"isSynced"`
}

// IPhoneScanProgress represents scan progress for the frontend.
type IPhoneScanProgress struct {
	Phase         string `json:"phase"` // "discovering", "checking", "complete"
	Total         int    `json:"total"`
	Scanned       int    `json:"scanned"`
	UnsyncedCount int    `json:"unsyncedCount"`
}

// IPhoneScanResult is the final result of a scan.
type IPhoneScanResult struct {
	Photos        []IPhonePhoto `json:"photos"`
	TotalOnPhone  int           `json:"totalOnPhone"`
	UnsyncedCount int           `json:"unsyncedCount"`
	SyncedCount   int           `json:"syncedCount"`
}

// IPhoneScanner scans an iPhone's DCIM directory and compares against Google Photos.
type IPhoneScanner struct {
	mu         sync.Mutex
	logger     *slog.Logger
	cancel     context.CancelFunc
	onProgress func(IPhoneScanProgress)
	onComplete func(IPhoneScanResult)
}

// NewIPhoneScanner creates a scanner that reports progress via callbacks.
func NewIPhoneScanner(logger *slog.Logger, onProgress func(IPhoneScanProgress), onComplete func(IPhoneScanResult)) *IPhoneScanner {
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}
	return &IPhoneScanner{
		logger:     logger,
		onProgress: onProgress,
		onComplete: onComplete,
	}
}

// Cancel stops an in-progress scan.
func (s *IPhoneScanner) Cancel() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
	}
}

// ScanIPhonePhotos scans the iPhone DCIM directory and checks each photo
// against Google Photos to find unsynced photos.
func (s *IPhoneScanner) ScanIPhonePhotos(devicePath string, apiOpts ApiOptions) {
	ctx, cancel := context.WithCancel(context.Background())
	s.mu.Lock()
	s.cancel = cancel
	s.mu.Unlock()

	go func() {
		defer cancel()
		s.doScan(ctx, devicePath, apiOpts)
	}()
}

func (s *IPhoneScanner) doScan(ctx context.Context, devicePath string, apiOpts ApiOptions) {
	// Phase 1: Discover all photos on the iPhone
	s.reportProgress(IPhoneScanProgress{Phase: "discovering", Total: 0, Scanned: 0})

	allPhotos, err := s.discoverPhotos(ctx, devicePath)
	if err != nil {
		s.logger.Error("Failed to discover photos", "error", err)
		s.reportComplete(IPhoneScanResult{})
		return
	}

	if len(allPhotos) == 0 {
		s.reportComplete(IPhoneScanResult{TotalOnPhone: 0})
		return
	}

	s.logger.Info("Discovered photos on iPhone", "count", len(allPhotos))
	s.reportProgress(IPhoneScanProgress{Phase: "checking", Total: len(allPhotos), Scanned: 0})

	// Phase 2: Create API client and check hashes against Google Photos
	api, err := NewApi(apiOpts)
	if err != nil {
		s.logger.Error("Failed to create API for hash check", "error", err)
		result := IPhoneScanResult{
			Photos:        allPhotos,
			TotalOnPhone:  len(allPhotos),
			UnsyncedCount: len(allPhotos),
			SyncedCount:   0,
		}
		s.reportComplete(result)
		return
	}

	// Load local sync cache to make repeated checks fast
	syncCache := loadSyncCache()

	unsyncedPhotos := make([]IPhonePhoto, 0, len(allPhotos))
	syncedCount := 0

	for i, photo := range allPhotos {
		select {
		case <-ctx.Done():
			s.reportComplete(IPhoneScanResult{})
			return
		default:
		}

		if !IsValidMediaFile(photo.Path) {
			s.logger.Debug("Skipping invalid or corrupt media file", "path", photo.Path)
			_ = os.Remove(photo.Path)
			continue
		}

		hash, err := CalculateSHA1(ctx, photo.Path)
		if err != nil {
			s.logger.Debug("Hash calculation failed", "path", photo.Path, "error", err)
			photo.IsSynced = false
			unsyncedPhotos = append(unsyncedPhotos, photo)
		} else {
			hashKey := fmt.Sprintf("%x", hash)
			isSynced := false

			if syncCache[hashKey] {
				isSynced = true
			} else {
				mediaKey, err := api.FindRemoteMediaByHash(hash)
				if err == nil && len(mediaKey) > 0 {
					isSynced = true
					syncCache[hashKey] = true
				}
			}

			if isSynced {
				syncedCount++
				photo.IsSynced = true
			} else {
				photo.IsSynced = false
				if len(unsyncedPhotos) < 150 {
					if thumb, err := GenerateThumbnail(photo.Path, 320); err == nil && len(thumb) > 0 {
						photo.Thumbnail = "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(thumb)
					}
				}
				unsyncedPhotos = append(unsyncedPhotos, photo)
			}
		}

		s.reportProgress(IPhoneScanProgress{
			Phase:         "checking",
			Total:         len(allPhotos),
			Scanned:       i + 1,
			UnsyncedCount: len(unsyncedPhotos),
		})
	}

	// Save sync cache
	saveSyncCache(syncCache)

	result := IPhoneScanResult{
		Photos:        unsyncedPhotos,
		TotalOnPhone:  len(allPhotos),
		UnsyncedCount: len(unsyncedPhotos),
		SyncedCount:   syncedCount,
	}

	s.logger.Info("Scan complete", "total", len(allPhotos), "synced", syncedCount, "unsynced", len(unsyncedPhotos))
	s.reportComplete(result)
}

// discoverPhotos finds all supported photo/video files on the iPhone.
func (s *IPhoneScanner) discoverPhotos(ctx context.Context, devicePath string) ([]IPhonePhoto, error) {
	// If path is a real mounted filesystem path and exists, walk it
	if devicePath != "" && !strings.HasPrefix(devicePath, "WPD:") {
		if fi, err := os.Stat(devicePath); err == nil && fi.IsDir() {
			return s.discoverPhotosDirect(ctx, devicePath)
		}
	}

	// Otherwise, discover and stage photos via Windows Shell COM
	return s.discoverPhotosViaShell(ctx)
}

// discoverPhotosDirect recursively walks a local drive folder.
func (s *IPhoneScanner) discoverPhotosDirect(ctx context.Context, dcimPath string) ([]IPhonePhoto, error) {
	var photos []IPhonePhoto
	seenBases := make(map[string]int)

	err := filepath.Walk(dcimPath, func(path string, info os.FileInfo, err error) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err != nil {
			return nil
		}

		if info.IsDir() {
			if strings.HasPrefix(info.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}

		if info.Size() < 1024 || !isSupportedByGooglePhotos(path) || !IsValidMediaFile(path) {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		name := info.Name()
		base := strings.ToLower(strings.TrimSuffix(name, filepath.Ext(name)))

		if existingIdx, exists := seenBases[base]; exists {
			if ext == ".heic" && photos[existingIdx].Extension != ".heic" {
				photos[existingIdx] = IPhonePhoto{
					Path:      path,
					FileName:  name,
					Size:      info.Size(),
					Extension: ext,
				}
			}
			return nil
		}

		seenBases[base] = len(photos)
		photos = append(photos, IPhonePhoto{
			Path:      path,
			FileName:  name,
			Size:      info.Size(),
			Extension: ext,
		})

		if len(photos)%50 == 0 {
			s.reportProgress(IPhoneScanProgress{
				Phase: "discovering",
				Total: len(photos),
			})
		}

		return nil
	})

	return photos, err
}

// discoverPhotosViaShell stages iPhone media from WPD into local app data cache
// and returns the local file paths.
func (s *IPhoneScanner) discoverPhotosViaShell(ctx context.Context) ([]IPhonePhoto, error) {
	cacheBase := filepath.Join(os.Getenv("LOCALAPPDATA"), "gotohp", "iphone_cache")
	if err := os.MkdirAll(cacheBase, 0o755); err != nil {
		return nil, fmt.Errorf("create cache dir: %w", err)
	}

	psScript := `$cacheBase = Join-Path $env:LOCALAPPDATA 'gotohp\iphone_cache'
if (-not (Test-Path $cacheBase)) { New-Item -ItemType Directory -Path $cacheBase -Force | Out-Null }
# Clean up any leftover 0-byte or corrupted stubs from previous interrupted runs
Get-ChildItem $cacheBase -Recurse -File -ErrorAction SilentlyContinue | Where-Object { $_.Length -lt 1024 } | Remove-Item -Force -ErrorAction SilentlyContinue

$shell = New-Object -ComObject Shell.Application
$dev = $shell.NameSpace(17).Items() | Where-Object { $_.Name -like '*iPhone*' -or $_.Name -like '*Apple*' } | Select-Object -First 1
if ($dev) {
    $storage = $dev.GetFolder
    foreach ($sub in $storage.Items()) {
        if ($sub.Name -eq 'Internal Storage') {
            $internal = $sub.GetFolder
            foreach ($folderItem in $internal.Items()) {
                if ($folderItem.IsFolder) {
                    $folder = $folderItem.GetFolder
                    $photoFolders = @()
                    $hasSub = $false
                    foreach ($child in $folder.Items()) {
                        if ($child.IsFolder) { $photoFolders += $child; $hasSub = $true }
                    }
                    if (-not $hasSub) { $photoFolders += $folderItem }
                    foreach ($pf in $photoFolders) {
                        $pfFolder = $pf.GetFolder
                        $folderDir = Join-Path $cacheBase $pf.Name
                        if (-not (Test-Path $folderDir)) { New-Item -ItemType Directory -Path $folderDir -Force | Out-Null }
                        $destFolder = $shell.NameSpace($folderDir)

                        # Deduplicate: if an iPhone folder contains both IMG_1234.HEIC and IMG_1234.JPG,
                        # prioritize native HEIC and discard the virtual transcode JPG
                        $rawItems = @($pfFolder.Items())
                        $itemMap = @{}
                        foreach ($f in $rawItems) {
                            $ext = [System.IO.Path]::GetExtension($f.Name).ToLower()
                            if ($ext -match '^\.(jpg|jpeg|png|heic|mov|mp4|gif|webp|dng)$') {
                                $base = [System.IO.Path]::GetFileNameWithoutExtension($f.Name)
                                if ($ext -eq '.jpg' -or $ext -eq '.jpeg') {
                                    if (-not $itemMap.ContainsKey($base + '.heic')) {
                                        $itemMap[$base + $ext] = $f
                                    }
                                } elseif ($ext -eq '.heic') {
                                    $itemMap.Remove($base + '.jpg')
                                    $itemMap.Remove($base + '.jpeg')
                                    $itemMap[$base + $ext] = $f
                                } else {
                                    $itemMap[$base + $ext] = $f
                                }
                            }
                        }

                        foreach ($f in $itemMap.Values) {
                            $targetPath = Join-Path $folderDir $f.Name
                            if (Test-Path $targetPath) {
                                $existing = Get-Item $targetPath -ErrorAction SilentlyContinue
                                if ($existing -and $existing.Length -lt 1024) {
                                    Remove-Item $targetPath -Force -ErrorAction SilentlyContinue
                                }
                            }
                            if (-not (Test-Path $targetPath)) {
                                $destFolder.CopyHere($f, 16)
                                $waited = 0
                                while ($waited -lt 30) {
                                    Start-Sleep -Milliseconds 150
                                    $waited++
                                    if (Test-Path $targetPath) {
                                        $cur = Get-Item $targetPath -ErrorAction SilentlyContinue
                                        if ($cur -and $cur.Length -gt 1024) {
                                            $s1 = $cur.Length
                                            Start-Sleep -Milliseconds 100
                                            $cur2 = Get-Item $targetPath -ErrorAction SilentlyContinue
                                            if ($cur2 -and $cur2.Length -eq $s1) { break }
                                        }
                                    }
                                }
                            }
                            if (Test-Path $targetPath) {
                                $item = Get-Item $targetPath -ErrorAction SilentlyContinue
                                if ($item -and $item.Length -gt 1024) {
                                    Write-Output ($item.FullName + '|' + $item.Name + '|' + $item.Length)
                                } else {
                                    Remove-Item $targetPath -Force -ErrorAction SilentlyContinue
                                }
                            }
                        }
                    }
                }
            }
        }
    }
}`

	cmd := exec.CommandContext(ctx, "powershell", "-NoProfile", "-NonInteractive", "-Command", psScript)
	hideWindow(cmd)
	output, err := cmd.Output()
	if err != nil {
		s.logger.Error("PowerShell discovery error", "error", err)
		return nil, err
	}

	var photos []IPhonePhoto
	seenBases := make(map[string]int) // maps lowerBase -> index in photos

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) >= 2 {
			path := parts[0]
			name := parts[1]
			var size int64
			if len(parts) >= 3 {
				fmt.Sscanf(parts[2], "%d", &size)
			}
			if size >= 1024 && isSupportedByGooglePhotos(name) && IsValidMediaFile(path) {
				ext := strings.ToLower(filepath.Ext(name))
				base := strings.ToLower(strings.TrimSuffix(name, filepath.Ext(name)))

				// If we already added a .jpg for this base and now have .heic, replace it
				if existingIdx, exists := seenBases[base]; exists {
					if ext == ".heic" && photos[existingIdx].Extension != ".heic" {
						photos[existingIdx] = IPhonePhoto{
							Path:      path,
							FileName:  name,
							Size:      size,
							Extension: ext,
						}
					}
					// Otherwise, ignore duplicate .jpg
					continue
				}

				seenBases[base] = len(photos)
				photos = append(photos, IPhonePhoto{
					Path:      path,
					FileName:  name,
					Size:      size,
					Extension: ext,
				})
			}
		}
	}

	return photos, nil
}

// scaleImage resamples an image to target dimensions using pure Go.
func scaleImage(src image.Image, targetWidth, targetHeight int) *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))
	bounds := src.Bounds()
	srcW := bounds.Dx()
	srcH := bounds.Dy()
	if srcW == 0 || srcH == 0 {
		return dst
	}

	for y := 0; y < targetHeight; y++ {
		sy := bounds.Min.Y + (y*srcH)/targetHeight
		for x := 0; x < targetWidth; x++ {
			sx := bounds.Min.X + (x*srcW)/targetWidth
			dst.Set(x, y, src.At(sx, sy))
		}
	}
	return dst
}

// GenerateThumbnail creates a small JPEG thumbnail from a photo file.
// It checks for a cached .thumb_<maxSize>.jpg on disk, tries OS platform extraction
// (supporting HEIC, MOV, MP4, etc.), and falls back to Go standard image decoders.
func GenerateThumbnail(filePath string, maxSize int) ([]byte, error) {
	if maxSize <= 0 {
		maxSize = 160
	}

	// 1. Check disk cache if available
	thumbDir := filepath.Join(os.Getenv("LOCALAPPDATA"), "gotohp", "iphone_cache", ".thumbs")
	_ = os.MkdirAll(thumbDir, 0o755)
	hash := sha1.Sum([]byte(fmt.Sprintf("%s_%d", filePath, maxSize)))
	cacheFile := filepath.Join(thumbDir, fmt.Sprintf("%x.jpg", hash))

	if data, err := os.ReadFile(cacheFile); err == nil && len(data) > 0 {
		return data, nil
	}

	// 2. Try platform-native extractor (Windows Shell IShellItemImageFactory for HEIC/MOV/etc)
	if thumb, err := getPlatformThumbnail(filePath, maxSize); err == nil && len(thumb) > 0 {
		_ = os.WriteFile(cacheFile, thumb, 0o644)
		return thumb, nil
	}

	// 3. Fallback to standard Go image decoders
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}
	defer f.Close()

	src, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}

	bounds := src.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("invalid dimensions")
	}

	var newWidth, newHeight int
	if width > height {
		newWidth = maxSize
		newHeight = int(float64(height) * float64(maxSize) / float64(width))
	} else {
		newHeight = maxSize
		newWidth = int(float64(width) * float64(maxSize) / float64(height))
	}

	if newWidth < 1 {
		newWidth = 1
	}
	if newHeight < 1 {
		newHeight = 1
	}

	dst := scaleImage(src, newWidth, newHeight)

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, dst, &jpeg.Options{Quality: 70}); err != nil {
		return nil, fmt.Errorf("encode: %w", err)
	}

	data := buf.Bytes()
	_ = os.WriteFile(cacheFile, data, 0o644)
	return data, nil
}

func (s *IPhoneScanner) reportProgress(p IPhoneScanProgress) {
	if s.onProgress != nil {
		s.onProgress(p)
	}
}

func (s *IPhoneScanner) reportComplete(r IPhoneScanResult) {
	if s.onComplete != nil {
		s.onComplete(r)
	}
}
