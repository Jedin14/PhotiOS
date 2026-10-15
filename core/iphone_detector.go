//go:build !cli

package core

import (
	"log/slog"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// IPhoneStatus represents the current state of an iPhone connection.
type IPhoneStatus struct {
	Connected  bool   `json:"connected"`
	DeviceName string `json:"deviceName"`
	DevicePath string `json:"devicePath"`
}

// IPhoneDetector monitors for iPhone USB connections on Windows.
type IPhoneDetector struct {
	mu          sync.Mutex
	status      IPhoneStatus
	onConnect   func(IPhoneStatus)
	onDisconnect func()
	logger      *slog.Logger
	stopCh      chan struct{}
	stopped     bool
}

// NewIPhoneDetector creates a detector that polls for iPhone connections.
func NewIPhoneDetector(logger *slog.Logger, onConnect func(IPhoneStatus), onDisconnect func()) *IPhoneDetector {
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}
	return &IPhoneDetector{
		onConnect:    onConnect,
		onDisconnect: onDisconnect,
		logger:       logger,
		stopCh:       make(chan struct{}),
	}
}

// Start begins polling for iPhone connections.
func (d *IPhoneDetector) Start() {
	go d.pollLoop()
}

// Stop terminates the polling loop.
func (d *IPhoneDetector) Stop() {
	d.mu.Lock()
	defer d.mu.Unlock()
	if !d.stopped {
		close(d.stopCh)
		d.stopped = true
	}
}

// Status returns the current iPhone connection status.
func (d *IPhoneDetector) Status() IPhoneStatus {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.status
}

func (d *IPhoneDetector) pollLoop() {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	// Check immediately on start
	d.checkConnection()

	for {
		select {
		case <-d.stopCh:
			return
		case <-ticker.C:
			d.checkConnection()
		}
	}
}

func (d *IPhoneDetector) checkConnection() {
	status := detectIPhone(d.logger)

	d.mu.Lock()
	wasConnected := d.status.Connected
	d.status = status
	d.mu.Unlock()

	if status.Connected && !wasConnected {
		d.logger.Info("iPhone connected", "device", status.DeviceName, "path", status.DevicePath)
		if d.onConnect != nil {
			d.onConnect(status)
		}
	} else if !status.Connected && wasConnected {
		d.logger.Info("iPhone disconnected")
		if d.onDisconnect != nil {
			d.onDisconnect()
		}
	}
}

// detectIPhone uses PowerShell and Windows Shell COM to check for connected Apple devices silently.
func detectIPhone(logger *slog.Logger) IPhoneStatus {
	psScript := `$shell = New-Object -ComObject Shell.Application
$devices = $shell.NameSpace(17)
if ($devices) {
    foreach ($item in $devices.Items()) {
        if ($item.Name -like '*iPhone*' -or $item.Name -like '*Apple*') {
            Write-Output ('WPD|' + $item.Name)
            exit
        }
    }
}
Get-PSDrive -PSProvider FileSystem | ForEach-Object {
    $dcim = Join-Path $_.Root 'DCIM'
    if (Test-Path $dcim) {
        Write-Output ($_.Root + '|' + $_.Name + ' (iPhone)')
        exit
    }
}`

	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", psScript)
	hideWindow(cmd)
	output, err := cmd.Output()
	if err != nil {
		return IPhoneStatus{}
	}

	result := strings.TrimSpace(string(output))
	if result == "" {
		return IPhoneStatus{}
	}

	parts := strings.SplitN(result, "|", 2)
	if len(parts) == 2 {
		prefix := parts[0]
		name := parts[1]
		if prefix == "WPD" {
			return IPhoneStatus{
				Connected:  true,
				DeviceName: name,
				DevicePath: "WPD:" + name,
			}
		}
		return IPhoneStatus{
			Connected:  true,
			DeviceName: name,
			DevicePath: filepath.Join(prefix, "DCIM"),
		}
	}

	return IPhoneStatus{
		Connected:  true,
		DeviceName: result,
		DevicePath: "WPD:" + result,
	}
}
