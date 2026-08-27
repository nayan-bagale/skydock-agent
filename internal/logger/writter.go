package logger

import (
	"io"

	"gopkg.in/natefinch/lumberjack.v2"
)

func NewFileWriter(path string) io.Writer {
	return &lumberjack.Logger{
		Filename:   path,
		MaxSize:    50,
		MaxBackups: 10,
		MaxAge:     30,
		Compress:   true,
	}
}
