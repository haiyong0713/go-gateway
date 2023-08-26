package model

import v1 "go-gateway/app/app-svr/app-listener/interface/api/v1"

type Playlist struct {
	Items   []*v1.PlayItem
	From    v1.PlaylistSource
	Batch   string
	TrackID string
}
