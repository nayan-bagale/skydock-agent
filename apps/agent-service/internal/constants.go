package constants

import "time"

const (
	DownloadsDirectory = "/Users/nayanbagale/Developer/skydock-unified/Watching"
	// DocumentsDirectory   = "/Users/nayanbagale/Documents"
	TemporaryFileSuffix  = "~"
	DS_StoreFileName     = ".DS_Store"
	DebounceInterval     = 500 * time.Microsecond
	RenameSettleInterval = 400 * time.Millisecond
)

var Directories = []string{DownloadsDirectory}
