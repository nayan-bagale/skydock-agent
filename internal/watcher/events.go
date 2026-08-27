package watcher

import (
	"github.com/fsnotify/fsnotify"
)

type Event struct {
	Path string
	Op   fsnotify.Op
}
