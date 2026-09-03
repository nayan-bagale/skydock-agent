package database

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

const fileName = "skydock.sqlite3"

func DefaultPath() (string, error) {
	var baseDir string

	switch runtime.GOOS {
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}

		baseDir = filepath.Join(
			home,
			"Library",
			"Application Support",
			"SkyDock",
			"database",
		)
	case "windows":
		baseDir = filepath.Join(
			os.Getenv("LOCALAPPDATA"),
			"SkyDock",
			"database",
		)
	case "linux":
		dataDir, err := os.UserConfigDir()
		if err != nil {
			return "", err
		}

		baseDir = filepath.Join(
			dataDir,
			"SkyDock",
			"database",
		)
	default:
		return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return "", fmt.Errorf("create database directory: %w", err)
	}

	return filepath.Join(baseDir, fileName), nil
}
