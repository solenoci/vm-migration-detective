package vddk

import (
	"os"
	"sync"
)

var (
	mu     sync.RWMutex
	libDir string
)

// SetLibDir sets the VDDK library directory path for internal use
func SetLibDir(dir string) {
	mu.Lock()
	defer mu.Unlock()
	libDir = dir
}

// GetLibDir returns the configured VDDK library directory
// If not set, searches in common locations
func GetLibDir() string {
	mu.RLock()
	dir := libDir
	mu.RUnlock()

	if dir != "" {
		return dir
	}

	// If not set, search in common locations
	return FindLibDir()
}

// FindLibDir searches for VDDK library in common locations
// Returns empty string if not found
func FindLibDir() string {
	// Default locations to check
	paths := []string{
		"/opt/vmware-vix-disklib",
		"/usr/lib64/vmware-vix-disklib",
		"/usr/local/vmware-vix-disklib",
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

// GetLibPath returns the full library path (lib64 subdirectory)
// Used for LD_LIBRARY_PATH filtering
func GetLibPath() string {
	dir := GetLibDir()
	if dir == "" {
		return ""
	}
	return dir + "/lib64"
}
