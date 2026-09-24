package logger

import (
	"io"

	"gopkg.in/natefinch/lumberjack.v2"
)

func NewFileWriter(cfg Config) io.Writer {
	return &lumberjack.Logger{
		Filename:   cfg.FilePath,
		MaxSize:    cfg.MaxSizeMB,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAgeDays,
		Compress:   cfg.Compress,
	}
}
