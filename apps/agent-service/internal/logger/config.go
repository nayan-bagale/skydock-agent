package logger

import (
	"log/slog"

	constants "github.com/nayan-bagale/skydock-agent/internal"
)

type Config struct {
	Level       slog.Level
	Development bool
	FilePath    string
	MaxSizeMB   int
	MaxBackups  int
	MaxAgeDays  int
	Compress    bool
}

func DefaultConfig() Config {
	return Config{
		Level:       constants.LogLevel,
		Development: constants.LogDevelopment,
		MaxSizeMB:   constants.LogMaxSizeMB,
		MaxBackups:  constants.LogMaxBackups,
		MaxAgeDays:  constants.LogMaxAgeDays,
		Compress:    constants.LogCompress,
	}
}
