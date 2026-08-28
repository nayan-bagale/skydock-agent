package main

import (
	"os"

	internal "github.com/nayan-bagale/skydock-agent/internal"
	"github.com/nayan-bagale/skydock-agent/internal/logger"
	"github.com/nayan-bagale/skydock-agent/internal/watcher"
)

const (
	version = "1.0.0"
)

func main() {
	log := initLogger()

	defer func() {
		if err := log.Close(); err != nil {
			os.Stderr.WriteString(
				"failed to close logger: " + err.Error() + "\n",
			)
		}
	}()

	log.Info(
		"SkyDock agent starting",
		"version", version,
		"pid", os.Getpid(),
	)

	// ------------------------------------
	// File watcher
	// ------------------------------------

	watch, err := initWatcher(log)

	if err != nil {
		panic("failed to initialize watcher: " + err.Error())
	}

	defer func() {
		if err := watch.Close(); err != nil {
			log.Error(
				"failed to close watcher",
				"error", err,
			)
		}
	}()

	attachDirectoriesToWatcher(watch, log)

	go watch.StartWatcher()

	log.Info("SkyDock agent started")

	// Keep agent alive.
	select {}
}

func initLogger() *logger.Logger {
	logCfg := logger.DefaultConfig()

	logPath, err := logger.DefaultLogPath()
	if err != nil {
		panic(err)
	}
	logCfg.FilePath = logPath

	log, err := logger.New(logCfg)
	if err != nil {
		panic(err)
	}

	return log
}

func initWatcher(log *logger.Logger) (*watcher.Watcher, error) {
	watch, err := watcher.New(log)
	if err != nil {
		log.Error(
			"failed to create watcher",
			"error", err,
		)
		return nil, err
	}

	return watch, nil
}

func attachDirectoriesToWatcher(watch *watcher.Watcher, log *logger.Logger) bool {
	dirs, err := watcher.GetAllDirs(internal.Directories)
	if err != nil {
		log.Error(
			"failed to get directories",
			"error", err,
		)
		return false
	}

	for _, dir := range dirs {
		if err := watch.Watch(dir); err != nil {
			log.Error(
				"failed to watch directory",
				"path", dir,
				"error", err,
			)
			return false
		}
	}

	log.Info(
		"directories attached to watcher",
		"count", len(dirs),
	)

	return true
}
