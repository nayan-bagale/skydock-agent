package main

import (
	"fmt"
	"os"

	internal "github.com/nayan-bagale/skydock-agent/internal"
	"github.com/nayan-bagale/skydock-agent/internal/logger"
	"github.com/nayan-bagale/skydock-agent/internal/watcher"
)

const (
	version = "1.0.0"
)

func main() {
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
		"log_path", logPath,
	)

	fmt.Println("log file:", logPath)

	// ------------------------------------
	// File watcher
	// ------------------------------------

	watch, err := watcher.New(log)
	if err != nil {
		log.Error(
			"failed to create watcher",
			"error", err,
		)
		return
	}

	defer func() {
		if err := watch.Close(); err != nil {
			log.Error(
				"failed to close watcher",
				"error", err,
			)
		}
	}()

	dirs, err := watcher.GetAllDirs(internal.Directories)
	if err != nil {
		log.Error(
			"failed to get directories",
			"error", err,
		)
		return
	}

	log.Info(
		"directories discovered",
		"count", len(dirs),
	)

	for _, dir := range dirs {
		log.Debug(
			"directory discovered",
			"path", dir,
		)
	}

	// ------------------------------------
	// Start watcher
	// ------------------------------------

	go watch.StartWatcher()

	for _, dir := range dirs {
		if err := watch.Watch(dir); err != nil {
			log.Error(
				"failed to watch directory",
				"path", dir,
				"error", err,
			)
			return
		}

		log.Info(
			"watching directory",
			"path", dir,
		)
	}

	log.Info("SkyDock agent started")

	// Keep agent alive.
	select {}
}
