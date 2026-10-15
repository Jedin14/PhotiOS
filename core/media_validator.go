package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

var syncCacheMu sync.Mutex

// IsValidMediaFile checks whether a file on disk exists, is non-empty (>1024 bytes),
// and begins with recognized image/video container/codec magic bytes.
// This filters out 0-byte ghost copies and Apple iCloud/AAE metadata plists/stubs.
func IsValidMediaFile(filePath string) bool {
	info, err := os.Stat(filePath)
	if err != nil || info.Size() < 1024 {
		return false
	}

	f, err := os.Open(filePath)
	if err != nil {
		return false
	}
	defer f.Close()

	header := make([]byte, 16)
	n, err := f.Read(header)
	if err != nil || n < 8 {
		return false
	}

	// 1. JPEG: 0xFF, 0xD8, 0xFF
	if header[0] == 0xFF && header[1] == 0xD8 && header[2] == 0xFF {
		return true
	}

	// 2. PNG: 0x89, 'P', 'N', 'G', '\r', '\n', 0x1A, '\n'
	if bytes.HasPrefix(header, []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}) {
		return true
	}

	// 3. GIF: 'GIF87a' or 'GIF89a'
	if bytes.HasPrefix(header, []byte("GIF8")) {
		return true
	}

	// 4. WebP: 'RIFF....WEBP'
	if bytes.HasPrefix(header, []byte("RIFF")) && n >= 12 && string(header[8:12]) == "WEBP" {
		return true
	}

	// 5. TIFF / DNG: 'II*\x00' or 'MM\x00*'
	if bytes.HasPrefix(header, []byte{0x49, 0x49, 0x2A, 0x00}) || bytes.HasPrefix(header, []byte{0x4D, 0x4D, 0x00, 0x2A}) {
		return true
	}

	// 6. ISO Base Media File Format (HEIC, HEIF, AVIF, MOV, MP4):
	// Box format: 4-byte box length + 4-byte box type
	boxType := string(header[4:8])
	switch boxType {
	case "ftyp", "moov", "wide", "mdat", "free", "skip":
		return true
	}

	return false
}

// AddHashToSyncCache persists a known synced SHA-1 hash to the local sync cache.
func AddHashToSyncCache(sha1Hash []byte) {
	if len(sha1Hash) == 0 {
		return
	}
	syncCacheMu.Lock()
	defer syncCacheMu.Unlock()

	cache := loadSyncCache()
	hashKey := fmt.Sprintf("%x", sha1Hash)
	cache[hashKey] = true
	saveSyncCache(cache)
}

// loadSyncCache loads known synced SHA1 hashes.
func loadSyncCache() map[string]bool {
	cache := make(map[string]bool)
	cacheFile := getSyncCacheFilePath()
	data, err := os.ReadFile(cacheFile)
	if err == nil {
		_ = json.Unmarshal(data, &cache)
	}
	return cache
}

// saveSyncCache persists known synced SHA1 hashes.
func saveSyncCache(cache map[string]bool) {
	cacheFile := getSyncCacheFilePath()
	data, err := json.Marshal(cache)
	if err == nil {
		_ = os.WriteFile(cacheFile, data, 0o644)
	}
}

func getSyncCacheFilePath() string {
	dir := os.Getenv("LOCALAPPDATA")
	if dir == "" {
		dir = os.TempDir()
	}
	return filepath.Join(dir, "gotohp", "iphone_synced.json")
}
