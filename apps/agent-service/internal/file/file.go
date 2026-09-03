package file

import (
	"fmt"
	"os"
	"strings"
	"syscall"

	internal "github.com/nayan-bagale/skydock-agent/internal"
)

func GetFileMetadata(filePath string) (*FileMeta, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("stat file: %w", err)
	}

	checksum, err := Checksum(filePath)

	if err != nil {
		return nil, err
	}

	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return nil, fmt.Errorf("unexpected file stat type for %s", filePath)
	}

	inode := stat.Ino
	device := stat.Dev

	return &FileMeta{
		Path:        filePath,
		Name:        info.Name(),
		Size:        info.Size(),
		ModifiedAt:  info.ModTime(),
		IsDirectory: info.IsDir(),
		Inode:       inode,
		Device:      uint64(device),
		Checksum:    checksum,
		RemoteID:    "",
		SyncStatus:  "",
	}, nil

}

func IsValidFile(name string) bool {
	if strings.HasSuffix(name, internal.DS_StoreFileName) {
		return false
	}
	return true
}
