package model

import watchedmdl "git.bilibili.co/bapis/bapis-go/live/xroom-gate/common"

// Live .
type Live struct {
	RoomStatus    int64  `json:"roomStatus"`
	LiveStatus    int64  `json:"liveStatus"`
	URL           string `json:"url"`
	Title         string `json:"title"`
	Cover         string `json:"cover"`
	RoomID        int64  `json:"roomid"`
	RoundStatus   int64  `json:"roundStatus"`
	BroadcastType int64  `json:"broadcast_type"`
	// 看过开过
	WatchedShow *watchedmdl.WatchedShow `json:"watched_show"`
}
