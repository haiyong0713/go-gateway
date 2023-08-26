package model

import (
	v1 "git.bilibili.co/bapis/bapis-go/pgc/service/season/season"
)

// FollowType .
const (
	FollowTypeAnime    = 1
	FollowTypeCinema   = 2
	FollowStatusWant   = 1
	FollowStatusIng    = 2
	FollowStatusFinish = 3
)

// Bangumi bangumi struct.
type Bangumi struct {
	SeasonID      string `json:"season_id"`
	ShareURL      string `json:"share_url"`
	Title         string `json:"title"`
	IsFinish      string `json:"is_finish"`
	Favorites     string `json:"favorites"`
	NewestEpIndex string `json:"newest_ep_index"`
	LastEpIndex   string `json:"last_ep_index"`
	TotalCount    string `json:"total_count"`
	Cover         string `json:"cover"`
	Evaluate      string `json:"evaluate"`
	Brief         string `json:"brief"`
}

// FollowCard follow card.
type FollowCard struct {
	*v1.CardInfoProto
	FollowStatus int32  `json:"follow_status"`
	IsNew        int32  `json:"is_new"`
	Progress     string `json:"progress"`
	BothFollow   bool   `json:"both_follow"`
}
