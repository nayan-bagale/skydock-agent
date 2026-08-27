package constants

import "time"

const (
	DownloadsDirectory   = "/Users/nayanbagale/Downloads"
	DocumentsDirectory   = "/Users/nayanbagale/Documents"
	TemporaryFileSuffix  = "~"
	DebounceInterval     = 500 * time.Microsecond
)

var Directories = []string{DownloadsDirectory, DocumentsDirectory}
