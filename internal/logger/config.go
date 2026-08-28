package logger

import "log/slog"

type Config struct {
	Level       slog.Level
	Development bool
	FilePath    string
}

func DefaultConfig() Config {
	return Config{
		Level:       slog.LevelInfo,
		Development: true,
	}
}
