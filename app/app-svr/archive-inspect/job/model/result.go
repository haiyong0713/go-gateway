package model

import (
	"go-gateway/app/app-svr/archive/service/api"
	"time"
)

type ArchiveInfo struct {
	Arc *api.Arc
	Ip  string
}

type ArcExpand struct {
	Aid          int64     `json:"aid"`
	Mid          int64     `json:"mid"`
	ArcType      int64     `json:"arc_type"`
	RoomId       int64     `json:"room_id"`
	PremiereTime time.Time `json:"premiere_time"`
}

type SeasonEpisode struct {
	SeasonId  int64 `json:"season_id"`
	SectionId int64 `json:"section_id"`
	EpisodeId int64 `json:"episode_id"`
	Aid       int64 `json:"aid"`
	Attribute int64 `json:"attribute"`
}

func (sep *SeasonEpisode) AttrVal(bit uint) int32 {
	return int32((sep.Attribute >> bit) & int64(1))
}
