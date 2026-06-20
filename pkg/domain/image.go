package domain

import "time"

// ImageInfo is the summary view of an image as rendered by the GUI list
// panel. Equivalent to docker's image.Summary, podman's
// ImageSummary.
type ImageInfo struct {
	ID          string
	ParentID    string
	RepoTags    []string
	RepoDigests []string
	Created     time.Time
	Size        int64
	SharedSize  int64
	VirtualSize int64
	Labels      map[string]string
	Containers  int64 // number of containers using this image, -1 if unknown
}

// ImageHistoryItem is one entry of an image's build history.
type ImageHistoryItem struct {
	ID        string
	Created   time.Time
	CreatedBy string
	Size      int64
	Comment   string
	Tags      []string
}
