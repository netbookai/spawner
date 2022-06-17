package common

import (
	"fmt"
	"time"
)

// contains the set of function which helps with generating resource names, display names

//now return the current time in yyyyMMddhhmmss format
func now() string {
	return time.Now().Format("20060102150405")
}

//VolumeName return the display name for volume or snapshot as per the kind
// kind is either
func VolumeName(size int64) string {
	return fmt.Sprintf("vol-%d-%s", size, now())
}

//SnapshotDisplayName
func SnapshotDisplayName(volumeId string) string {
	return fmt.Sprintf("snapof-%s-%s", volumeId, now())
}

//CopySnapshotName
func CopySnapshotName(snapshotId string) string {
	return fmt.Sprintf("copy-%s-%s", snapshotId, now())
}
