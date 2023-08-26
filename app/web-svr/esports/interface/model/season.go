package model

import (
	v1 "go-gateway/app/web-svr/esports/interface/api/v1"
)

type TeamInSeason struct {
	TeamId    int64  `json:"team_id"`
	TeamTitle string `json:"team_title"`
	RegionId  int64  `json:"region_id"`
	SeasonId  int64  `json:"season_id"`
	Rank      int64  `json:"rank"`
	Logo      string `json:"logo"`
	Uid       int64  `json:"uid"`
	LeidaID   int64  `json:"leida_id"`
}

type SeasonTeam struct {
	*TeamInSeason
	IsSub bool `json:"is_sub"`
}

type MatchSeason struct {
	SeasonID    int64  `json:"season_id"`
	SeasonTitle string `json:"season_title"`
	Logo        string `json:"logo"`
	MatchID     int64  `json:"match_id"`
	Stime       int64  `json:"stime"`
	Etime       int64  `json:"etime"`
	IsSub       bool   `json:"is_sub"`
	StartSeason int64  `json:"start_season"`
}

type SeriesPointMatchMore struct {
	*v1.SeriesPointMatchInfo
	Season *ComponentSeason `json:"season"`
}
