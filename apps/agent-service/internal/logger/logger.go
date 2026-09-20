package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
)

type Logger struct {
	*slog.Logger
	file io.Closer
}

func New(cfg Config) (*Logger, error) {
	var writer io.Writer

	if cfg.Development {
		writer = os.Stdout
	} else {
		path := cfg.FilePath

		if path == "" {
			var err error

			path, err = DefaultLogPath()
			if err != nil {
				return nil, fmt.Errorf("resolve log path: %w", err)
			}
		}

		fileWriter := NewFileWriter(path)
		fileCloser, ok := fileWriter.(io.Closer)
		if !ok {
			return nil, fmt.Errorf("file writer does not implement io.Closer")
		}

		writer = io.MultiWriter(os.Stdout, fileWriter)

		logger := slog.New(
			slog.NewJSONHandler(
				writer,
				&slog.HandlerOptions{
					Level: cfg.Level,
				},
			),
		)

		return &Logger{
			Logger: logger,
			file:   fileCloser,
		}, nil
	}

	logger := slog.New(
		slog.NewTextHandler(
			writer,
			&slog.HandlerOptions{
				Level: cfg.Level,
			},
		),
	)

	return &Logger{
		Logger: logger,
	}, nil
}

func (l *Logger) enabled() bool {
	return l != nil && l.Logger != nil
}

func (l *Logger) Info(msg string, args ...any) {
	if !l.enabled() {
		return
	}
	l.Logger.Info(msg, args...)
}

func (l *Logger) Error(msg string, args ...any) {
	if !l.enabled() {
		return
	}
	l.Logger.Error(msg, args...)
}

func (l *Logger) Warn(msg string, args ...any) {
	if !l.enabled() {
		return
	}
	l.Logger.Warn(msg, args...)
}

func (l *Logger) Debug(msg string, args ...any) {
	if !l.enabled() {
		return
	}
	l.Logger.Debug(msg, args...)
}

func (l *Logger) Close() error {
	if l == nil || l.file == nil {
		return nil
	}

	return l.file.Close()
}
