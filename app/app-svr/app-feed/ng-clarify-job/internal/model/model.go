package model

import (
	"go-gateway/app/app-svr/app-feed/ng-clarify-job/api/session"
)

type IndexSession = session.IndexSession

type PresignedURLReply struct {
	URL string `json:"url"`
}

type ScanArchiveIndexReply struct {
	Index   []*ArchiveIndex `json:"index"`
	NextKey string          `json:"next_key"`
	HasNext bool            `json:"has_next"`
}

type ArchiveIndex struct {
	Path        string `json:"name"`
	URL         string `json:"url"`
	CreatedAt   int64  `json:"created_at"`
	CreatedTime string `json:"created_time"`
	RawSize     int64  `json:"raw_size"`
	GzipedSize  int64  `json:"gziped_size"`
}
