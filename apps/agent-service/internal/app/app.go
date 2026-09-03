package app

import (
	"fmt"
	"os"

	constants "github.com/nayan-bagale/skydock-agent/internal"
	"github.com/nayan-bagale/skydock-agent/internal/database"
	filepkg "github.com/nayan-bagale/skydock-agent/internal/file"
	"github.com/nayan-bagale/skydock-agent/internal/logger"
	"github.com/nayan-bagale/skydock-agent/internal/repository"
	"github.com/nayan-bagale/skydock-agent/internal/watcher"
)

const version = "1.0.0"

func Run() error {
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

	dbPath, err := database.DefaultPath()
	if err != nil {
		return fmt.Errorf("resolve database path: %w", err)
	}

	db, err := database.Open(dbPath)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer func() {
		if err := database.Close(db); err != nil {
			log.Error("failed to close database", "error", err)
		}
	}()

	if err := database.Migrate(db); err != nil {
		return fmt.Errorf("migrate database: %w", err)
	}

	log.Info("database initialized", "path", dbPath)

	fileRepo := repository.NewFileRepository(db)

	validateFiles(log, fileRepo)

	watch, err := initWatcher(log, fileRepo)
	if err != nil {
		return fmt.Errorf("initialize watcher: %w", err)
	}

	defer func() {
		if err := watch.Close(); err != nil {
			log.Error("failed to close watcher", "error", err)
		}
	}()

	if !attachDirectoriesToWatcher(watch, log) {
		return fmt.Errorf("attach directories to watcher")
	}

	go watch.StartWatcher()

	log.Info("SkyDock agent started")

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

func validateFiles(log *logger.Logger, fileRepo *repository.FileRepository) {
	rawFiles, err := filepkg.GetAllFiles(constants.Directories)

	var files []string

	for _, f := range rawFiles {
		if filepkg.IsValidFile(f) {
			files = append(files, f)
		}
	}

	if err != nil {
		log.Error("failed to get files", "error", err)
		return
	}

	for _, f := range files {
		meta, err := filepkg.GetFileMetadata(f)
		if err != nil {
			log.Error("failed to get file metadata", "path", f, "error", err)
			continue
		}

		if err := fileRepo.Upsert(meta); err != nil {
			log.Error("failed to persist file metadata", "path", meta.Path, "error", err)
		}

		log.Info(
			"file metadata",
			"path", meta.Path,
			"name", meta.Name,
			"size", meta.Size,
			"modified", meta.ModifiedAt,
			"isDir", meta.IsDirectory,
			"inode", meta.Inode,
			"device", meta.Device,
			"checksum", meta.Checksum,
			"remoteID", meta.RemoteID,
			"syncStatus", meta.SyncStatus,
		)
	}
}

func initWatcher(log *logger.Logger, fileRepo *repository.FileRepository) (*watcher.Watcher, error) {
	watch, err := watcher.New(log, fileRepo)
	if err != nil {
		log.Error("failed to create watcher", "error", err)
		return nil, err
	}

	return watch, nil
}

func attachDirectoriesToWatcher(watch *watcher.Watcher, log *logger.Logger) bool {
	dirs, err := filepkg.GetAllDirs(constants.Directories)
	if err != nil {
		log.Error("failed to get directories", "error", err)
		return false
	}

	for _, dir := range dirs {
		if err := watch.Watch(dir); err != nil {
			log.Error("failed to watch directory", "path", dir, "error", err)
			return false
		}
	}

	log.Info("directories attached to watcher", "count", len(dirs))

	return true
}
