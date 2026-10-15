//go:build windows

package core

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"syscall"
	"unsafe"
)

func getPlatformThumbnail(filePath string, maxSize int) ([]byte, error) {
	return extractWindowsShellThumbnail(filePath, maxSize)
}

var (
	modShell32 = syscall.NewLazyDLL("shell32.dll")
	modGdi32   = syscall.NewLazyDLL("gdi32.dll")
	modOle32   = syscall.NewLazyDLL("ole32.dll")

	procSHCreateItemFromParsingName = modShell32.NewProc("SHCreateItemFromParsingName")
	procCreateCompatibleDC          = modGdi32.NewProc("CreateCompatibleDC")
	procDeleteDC                    = modGdi32.NewProc("DeleteDC")
	procDeleteObject                = modGdi32.NewProc("DeleteObject")
	procGetObject                   = modGdi32.NewProc("GetObjectW")
	procGetDIBits                   = modGdi32.NewProc("GetDIBits")
	procCoInitializeEx              = modOle32.NewProc("CoInitializeEx")
	procCoUninitialize              = modOle32.NewProc("CoUninitialize")
)

type winGUID struct {
	Data1 uint32
	Data2 uint16
	Data3 uint16
	Data4 [8]byte
}

// IID_IShellItemImageFactory: {bcc18b79-ba16-442f-80c4-8a59c30c463b}
var iidIShellItemImageFactory = winGUID{
	Data1: 0xbcc18b79,
	Data2: 0xba16,
	Data3: 0x442f,
	Data4: [8]byte{0x80, 0xc4, 0x8a, 0x59, 0xc3, 0x0c, 0x46, 0x3b},
}

type winSIZE struct {
	cx int32
	cy int32
}

type winBITMAP struct {
	bmType       int32
	bmWidth      int32
	bmHeight     int32
	bmWidthBytes int32
	bmPlanes     uint16
	bmBitsPixel  uint16
	bmBits       uintptr
}

type winBITMAPINFOHEADER struct {
	biSize          uint32
	biWidth         int32
	biHeight        int32
	biPlanes        uint16
	biBitCount      uint16
	biCompression   uint32
	biSizeImage     uint32
	biXPelsPerMeter int32
	biYPelsPerMeter int32
	biClrUsed       uint32
	biClrImportant  uint32
}

type winBITMAPINFO struct {
	bmiHeader winBITMAPINFOHEADER
	bmiColors [1]uint32
}

// extractWindowsShellThumbnail extracts a thumbnail using Windows Shell IShellItemImageFactory.
// Supports HEIC, MOV, MP4, JPEG, PNG, etc. natively without third-party dependencies.
func extractWindowsShellThumbnail(filePath string, maxSize int) ([]byte, error) {
	// Initialize COM for this thread if needed
	procCoInitializeEx.Call(0, 2) // COINIT_APARTMENTTHREADED = 0x2
	defer procCoUninitialize.Call()

	pathPtr, err := syscall.UTF16PtrFromString(filePath)
	if err != nil {
		return nil, fmt.Errorf("utf16 path: %w", err)
	}

	var factory uintptr
	r1, _, _ := procSHCreateItemFromParsingName.Call(
		uintptr(unsafe.Pointer(pathPtr)),
		0,
		uintptr(unsafe.Pointer(&iidIShellItemImageFactory)),
		uintptr(unsafe.Pointer(&factory)),
	)
	if int32(r1) < 0 || factory == 0 {
		return nil, fmt.Errorf("SHCreateItemFromParsingName failed: hr=0x%x", r1)
	}

	// IShellItemImageFactory vtable:
	// 0: QueryInterface
	// 1: AddRef
	// 2: Release
	// 3: GetImage(SIZE size, SIIGBF flags, HBITMAP *phbm)
	vtable := *(**[4]uintptr)(unsafe.Pointer(factory))
	release := vtable[2]
	getImage := vtable[3]

	defer syscall.SyscallN(release, factory)

	size := winSIZE{cx: int32(maxSize), cy: int32(maxSize)}
	// SIIGBF_THUMBNAILONLY = 0x2, SIIGBF_SCALEUP = 0x1, SIIGBF_RESIZETOFIT = 0x0
	// 0x0 allows scaling/fallback to icon or thumbnail
	var hBitmap uintptr
	r1, _, _ = syscall.SyscallN(
		getImage,
		factory,
		*(*uintptr)(unsafe.Pointer(&size)),
		0, // SIIGBF_RESIZETOFIT
		uintptr(unsafe.Pointer(&hBitmap)),
	)
	if int32(r1) < 0 || hBitmap == 0 {
		return nil, fmt.Errorf("GetImage failed: hr=0x%x", r1)
	}
	defer procDeleteObject.Call(hBitmap)

	// Convert HBITMAP to RGBA image
	var bm winBITMAP
	procGetObject.Call(hBitmap, uintptr(unsafe.Sizeof(bm)), uintptr(unsafe.Pointer(&bm)))
	if bm.bmWidth <= 0 || bm.bmHeight <= 0 {
		return nil, fmt.Errorf("invalid bitmap dimensions: %dx%d", bm.bmWidth, bm.bmHeight)
	}

	width := bm.bmWidth
	height := bm.bmHeight

	hdc, _, _ := procCreateCompatibleDC.Call(0)
	if hdc == 0 {
		return nil, fmt.Errorf("CreateCompatibleDC failed")
	}
	defer procDeleteDC.Call(hdc)

	var bi winBITMAPINFO
	bi.bmiHeader.biSize = uint32(unsafe.Sizeof(bi.bmiHeader))
	bi.bmiHeader.biWidth = width
	bi.bmiHeader.biHeight = -height // top-down
	bi.bmiHeader.biPlanes = 1
	bi.bmiHeader.biBitCount = 32
	bi.bmiHeader.biCompression = 0 // BI_RGB

	rawSize := int(width * height * 4)
	pixels := make([]byte, rawSize)

	lines, _, _ := procGetDIBits.Call(
		hdc,
		hBitmap,
		0,
		uintptr(height),
		uintptr(unsafe.Pointer(&pixels[0])),
		uintptr(unsafe.Pointer(&bi)),
		0, // DIB_RGB_COLORS
	)
	if lines == 0 {
		return nil, fmt.Errorf("GetDIBits failed")
	}

	// Convert BGRA -> RGBA in place
	for i := 0; i < len(pixels); i += 4 {
		pixels[i], pixels[i+2] = pixels[i+2], pixels[i] // swap B and R
		pixels[i+3] = 255                               // ensure full opacity
	}

	img := &image.RGBA{
		Pix:    pixels,
		Stride: int(width) * 4,
		Rect:   image.Rect(0, 0, int(width), int(height)),
	}

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 75}); err != nil {
		return nil, fmt.Errorf("jpeg encode: %w", err)
	}

	return buf.Bytes(), nil
}
