//go:build !windows

package core

import "fmt"

func getPlatformThumbnail(filePath string, maxSize int) ([]byte, error) {
	return nil, fmt.Errorf("platform thumbnail not supported on non-windows")
}
