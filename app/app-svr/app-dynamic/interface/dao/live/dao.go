package live

import (
	bm "go-common/library/net/http/blademaster"

	"go-gateway/app/app-svr/app-dynamic/interface/conf"

	livexroom "git.bilibili.co/bapis/bapis-go/live/xroom"
	livexroomfeed "git.bilibili.co/bapis/bapis-go/live/xroom-feed"
	livexroomgate "git.bilibili.co/bapis/bapis-go/live/xroom-gate"
)

type Dao struct {
	c                 *conf.Config
	livexroomgrpc     livexroom.RoomClient
	livexroomfeedgrpc livexroomfeed.DynamicClient
	livexroomgategrpc livexroomgate.XroomgateClient
	// http client
	client *bm.Client
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:      c,
		client: bm.NewClient(c.HTTPClient),
	}
	var err error
	if d.livexroomgrpc, err = livexroom.NewClient(c.LivexRoomGRPC); err != nil {
		panic(err)
	}
	if d.livexroomfeedgrpc, err = livexroomfeed.NewClient(c.LivexRoomFeedGRPC); err != nil {
		panic(err)
	}
	if d.livexroomgategrpc, err = livexroomgate.NewClientXroomgate(c.LivexRoomGateGRPC); err != nil {
		panic(err)
	}
	return
}
