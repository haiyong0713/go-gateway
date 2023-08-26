package dao

import (
	"context"

	xroomfeedgrpc "git.bilibili.co/bapis/bapis-go/live/xroom-feed"
	"go-common/library/net/metadata"
)

type liveXRoomFeedDao struct {
	client xroomfeedgrpc.DynamicClient
}

// GetCardInfo .
func (d *liveXRoomFeedDao) GetCardInfo(c context.Context, RoomIds []int64, mid, build int64, platform, device string, isHttps bool) (map[uint64]*xroomfeedgrpc.LiveCardInfo, error) {
	ip := metadata.String(c, metadata.RemoteIP)
	rly, err := d.client.GetCardInfo(c, &xroomfeedgrpc.GetCardInfoReq{RoomIds: RoomIds, Uid: mid, Build: build, Platform: platform, DeviceName: device, IsHttps: isHttps, Ip: ip})
	if err != nil {
		return nil, err
	}
	if rly != nil {
		return rly.LivePlayInfo, nil
	}
	return nil, nil
}
