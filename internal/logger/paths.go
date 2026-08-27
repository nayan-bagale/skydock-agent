package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

func DefaultLogPath() (string, error) {
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
			"logs",
		)

	case "windows":
		baseDir = filepath.Join(
			os.Getenv("LOCALAPPDATA"),
			"SkyDock",
			"logs",
		)

	case "linux":
		dataDir, err := os.UserConfigDir()
		if err != nil {
			return "", err
		}

		baseDir = filepath.Join(
			dataDir,
			"SkyDock",
			"logs",
		)

	default:
		return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return "", fmt.Errorf("create log directory: %w", err)
	}

	return filepath.Join(baseDir, "agent.log"), nil
}
