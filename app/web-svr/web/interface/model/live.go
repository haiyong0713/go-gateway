package model

import (
	livegrpcmdl "git.bilibili.co/bapis/bapis-go/live/xroom"
	livecommon "git.bilibili.co/bapis/bapis-go/live/xroom-gate/common"
)

type LiveRoomInfo struct {
	// room_id 房间长号
	RoomId int64 `json:"room_id"`
	// uid 主播uid
	Uid int64 `json:"uid"`
	// 房间状态相关 开播状态： 1: 直播中  其他：非直播中, 99: 因加密等原因处于特殊状态等价于非直播中
	LiveStatus int64 `json:"live_status,omitempty"`
	// 展示相关
	Show *livegrpcmdl.RoomShowInfo `json:"show"`
	// 分区相关
	Area *livegrpcmdl.RoomAreaInfo `json:"area"`
	// 看过
	WatchedShow *livecommon.WatchedShow `json:"watched_show"`
}
