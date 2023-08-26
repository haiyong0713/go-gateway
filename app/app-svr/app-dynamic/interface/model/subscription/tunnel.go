package subscription

import (
	livexroomfeed "git.bilibili.co/bapis/bapis-go/live/xroom-feed"
)

const (
	TunnelTypeDraw = "image"
	TunnelTypeLive = "live_room"
)

type Live struct {
	*livexroomfeed.HistoryCardInfo
}
