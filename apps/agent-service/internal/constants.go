package constants

import (
	"log/slog"
	"strings"
	"time"
)

const (
	LogLevel             = slog.LevelInfo
	LogDevelopment       = true
	LogMaxSizeMB         = 50
	LogMaxBackups        = 10
	LogMaxAgeDays        = 30
	LogCompress          = true
	DebounceInterval     = 500 * time.Microsecond
	RenameSettleInterval = 400 * time.Millisecond
	ReconcileInterval    = 30 * time.Second
	TemporaryFileSuffix  = "~"
	DS_StoreFileName     = ".DS_Store"
)

// IgnoredName reports whether path ends with a name the watcher should skip.
func IgnoredName(path string) bool {
	return DS_StoreFileName != "" && strings.HasSuffix(path, DS_StoreFileName)
}
