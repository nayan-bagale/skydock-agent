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

func (l *Logger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}

	return nil
}
