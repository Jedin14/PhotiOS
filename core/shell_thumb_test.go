package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestExtractWindowsShellThumbnail(t *testing.T) {
	cacheBase := filepath.Join(os.Getenv("LOCALAPPDATA"), "gotohp", "iphone_cache")
	var heicFile, movFile string
	filepath.Walk(cacheBase, func(p string, fi os.FileInfo, err error) error {
		if err == nil && !fi.IsDir() && !strings.Contains(p, ".thumb_") {
			ext := strings.ToUpper(filepath.Ext(p))
			if ext == ".HEIC" && heicFile == "" {
				heicFile = p
			} else if ext == ".MOV" && movFile == "" {
				movFile = p
			}
		}
		return nil
	})

	if heicFile != "" {
		t.Logf("Testing HEIC thumbnail generation on: %s", heicFile)
		start := time.Now()
		thumb, err := GenerateThumbnail(heicFile, 160)
		elapsed := time.Since(start)
		if err != nil {
			t.Errorf("HEIC failed: %v", err)
		} else {
			t.Logf("HEIC thumbnail success in %v! Size: %d bytes", elapsed, len(thumb))
		}

		// Second call should hit disk cache
		start = time.Now()
		thumbCached, err := GenerateThumbnail(heicFile, 160)
		elapsedCached := time.Since(start)
		if err != nil || len(thumbCached) != len(thumb) {
			t.Errorf("Cached call failed: %v", err)
		} else {
			t.Logf("Cached HEIC thumbnail returned in %v!", elapsedCached)
		}
	}

	if movFile != "" {
		t.Logf("Testing MOV thumbnail generation on: %s", movFile)
		thumb, err := GenerateThumbnail(movFile, 160)
		if err != nil {
			t.Errorf("MOV failed: %v", err)
		} else {
			t.Logf("MOV thumbnail success! Size: %d bytes", len(thumb))
		}
	}
}
