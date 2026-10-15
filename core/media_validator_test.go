package core

import (
	"crypto/sha1"
	"os"
	"path/filepath"
	"testing"
)

func TestIsValidMediaFile(t *testing.T) {
	tempDir := t.TempDir()

	// 1. 0-byte file
	zeroFile := filepath.Join(tempDir, "zero.jpg")
	if err := os.WriteFile(zeroFile, []byte{}, 0o644); err != nil {
		t.Fatal(err)
	}
	if IsValidMediaFile(zeroFile) {
		t.Errorf("expected 0-byte file to be invalid")
	}

	// 2. <1024 byte stub file (e.g. Apple metadata stub)
	stubFile := filepath.Join(tempDir, "stub.jpg")
	stubContent := []byte("CFN^bplist00?_Adjustments.plist")
	if err := os.WriteFile(stubFile, stubContent, 0o644); err != nil {
		t.Fatal(err)
	}
	if IsValidMediaFile(stubFile) {
		t.Errorf("expected <1024 byte stub file to be invalid")
	}

	// 3. >1024 byte text file
	textFile := filepath.Join(tempDir, "not_image.jpg")
	textContent := make([]byte, 2048)
	copy(textContent, []byte("This is plain text with no valid image magic bytes!"))
	if err := os.WriteFile(textFile, textContent, 0o644); err != nil {
		t.Fatal(err)
	}
	if IsValidMediaFile(textFile) {
		t.Errorf("expected text file with .jpg extension to be invalid")
	}

	// 4. Valid JPEG
	jpegFile := filepath.Join(tempDir, "valid.jpg")
	jpegContent := make([]byte, 2048)
	jpegContent[0] = 0xFF
	jpegContent[1] = 0xD8
	jpegContent[2] = 0xFF
	jpegContent[3] = 0xE0
	if err := os.WriteFile(jpegFile, jpegContent, 0o644); err != nil {
		t.Fatal(err)
	}
	if !IsValidMediaFile(jpegFile) {
		t.Errorf("expected valid JPEG to be valid")
	}

	// 5. Valid HEIC (ftyp box)
	heicFile := filepath.Join(tempDir, "valid.heic")
	heicContent := make([]byte, 2048)
	// Box size: 0x00, 0x00, 0x00, 0x18
	heicContent[3] = 0x18
	// Box type: 'ftyp'
	copy(heicContent[4:8], []byte("ftyp"))
	// Major brand: 'heic'
	copy(heicContent[8:12], []byte("heic"))
	if err := os.WriteFile(heicFile, heicContent, 0o644); err != nil {
		t.Fatal(err)
	}
	if !IsValidMediaFile(heicFile) {
		t.Errorf("expected valid HEIC to be valid")
	}
}

func TestSyncCache(t *testing.T) {
	testHash := sha1.Sum([]byte("test image data"))
	AddHashToSyncCache(testHash[:])

	cache := loadSyncCache()
	hashKey := filepath.Clean(string(testHash[:]))
	_ = hashKey
	hashHex := ""
	for _, b := range testHash {
		hashHex += string("0123456789abcdef"[b>>4]) + string("0123456789abcdef"[b&0xf])
	}
	if !cache[hashHex] {
		t.Errorf("expected hash %s to be present in sync cache", hashHex)
	}
}
