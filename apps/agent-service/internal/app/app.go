package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	constants "github.com/nayan-bagale/skydock-agent/internal"
	"github.com/nayan-bagale/skydock-agent/internal/database"
	filepkg "github.com/nayan-bagale/skydock-agent/internal/file"
	"github.com/nayan-bagale/skydock-agent/internal/logger"
	"github.com/nayan-bagale/skydock-agent/internal/reconciler"
	"github.com/nayan-bagale/skydock-agent/internal/repository"
	"github.com/nayan-bagale/skydock-agent/internal/watcher"
)

const version = "1.0.0"

func Run() error {
	log := initLogger()
	// Use one context for all long-running services so shutdown can be coordinated.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

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

	// Reconcile before starting the watcher so the database has a complete
	// baseline and events are only responsible for changes after startup.
	reconcilerService := reconciler.New(constants.Directories, fileRepo, log)
	if err := reconcilerService.Reconcile(); err != nil {
		return fmt.Errorf("initial filesystem reconciliation: %w", err)
	}

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
	go reconcilerService.Run(ctx, reconciler.DefaultInterval)

	log.Info("SkyDock agent started")

	<-ctx.Done()
	return nil
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
