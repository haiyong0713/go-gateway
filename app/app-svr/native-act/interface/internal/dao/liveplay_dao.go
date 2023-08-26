package dao

import (
	"context"

	liveplaygrpc "git.bilibili.co/bapis/bapis-go/live/live-play/v1"
)

type livePlayDao struct {
	client liveplaygrpc.TopicClient
}

func (d *livePlayDao) GetListByRoomId(c context.Context, roomIDs []int64, isLive int64) (map[int64]*liveplaygrpc.RoomList, error) {
	req := &liveplaygrpc.GetListByRoomIdReq{
		RoomIds: roomIDs,
		Filter:  &liveplaygrpc.Filter{IsLive: isLive},
	}
	rly, err := d.client.GetListByRoomId(c, req)
	if err != nil {
		return nil, err
	}
	list := make(map[int64]*liveplaygrpc.RoomList, len(rly.List))
	for _, item := range rly.List {
		list[item.RoomId] = item
	}
	return list, nil
}

func (d *livePlayDao) GetListByActId(c context.Context, req *liveplaygrpc.GetListByActIdReq) (*liveplaygrpc.GetListByActIdResp, error) {
	rly, err := d.client.GetListByActId(c, req)
	if err != nil {
		return nil, err
	}
	return rly, nil
}
