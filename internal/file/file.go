package file

import (
	"os"
	"syscall"
)

func GetFileMetadata(filePath string) (*FileMeta, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		panic(err)
	}

	checksum, err := Checksum(filePath)

	if err != nil {
		return nil, err
	}

	stat := info.Sys().(*syscall.Stat_t)

	inode := stat.Ino
	device := stat.Dev

	return &FileMeta{
		filePath,
		info.Name(),
		info.Size(),
		info.ModTime(),
		info.IsDir(),
		inode,          // Inode (not available in this example)
		uint64(device), // Device (not available in this example)
		checksum,
		"", // RemoteID (not available in this example)
		"", // SyncStatus (not available in this example)
	}, nil

}
