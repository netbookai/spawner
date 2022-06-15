package helper

import (
	"fmt"
	"time"
)

// contains the set of function which helps with generating resource names, display names

type Kind string

const (
	VolumeKind   Kind = "vol"
	SnapshotKind Kind = "snap"
)

func (k Kind) String() string {
	return string(k)
}

//now return the current time in yyyyMMddhhmmss format
func now() string {
	return time.Now().Format("20060102150405")
}

//DisplayName return the display name for volume or snapshot as per the kind
// kind is either
func DisplayName(kind Kind, size int64) string {
	return fmt.Sprintf("%s-%d-%s", kind, size, now())
}

//SnapshotDisplayName
func SnapshotDisplayName(volumeId string) string {
	return fmt.Sprintf("snapof-%s-%s", volumeId, now())
}

//CopySnapName
func CopySnapName(snapshotId string) string {
	return fmt.Sprintf("copy-%s-%s", snapshotId, now())
}
