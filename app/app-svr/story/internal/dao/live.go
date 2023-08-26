package dao

import (
	"context"

	liverankgrpc "git.bilibili.co/bapis/bapis-go/live/rankdb/v1"
	livegrpc "git.bilibili.co/bapis/bapis-go/live/xroom-gate"
)

func (d *dao) LiveRoomInfos(ctx context.Context, req *livegrpc.EntryRoomInfoReq) (map[int64]*livegrpc.EntryRoomInfoResp_EntryList, error) {
	reply, err := d.liveClient.EntryRoomInfo(ctx, req)
	if err != nil {
		return nil, err
	}
	return reply.List, nil
}

func (d *dao) LiveHotRank(ctx context.Context, ids []int64) (map[int64]*liverankgrpc.IsInHotRankResp_HotRankData, error) {
	reply, err := d.liveRankClient.IsInHotRank(ctx, &liverankgrpc.IsInHotRankReq{
		RoomIds: ids,
	})
	if err != nil {
		return nil, err
	}
	return reply.GetList(), nil
}
