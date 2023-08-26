package dao

import (
	"context"

	xecode "go-common/library/ecode"
	"go-common/library/log"

	liveRoom "git.bilibili.co/bapis/bapis-go/live/xroom"
)

func (d *dao) LiveRoomInfo(ctx context.Context, roomIds []int64) (roomInfos map[int64]*liveRoom.Infos, err error) {
	roomInfos = make(map[int64]*liveRoom.Infos)
	req := new(liveRoom.RoomIDsReq)
	{
		req.RoomIds = roomIds
		req.Attrs = []string{"show", "status"}
	}

	res, err := d.liveRoomClient.GetMultiple(ctx, req)
	if err != nil {
		log.Errorc(ctx, "[Dao][LiveRoom][GetMultiple][Error], err:%+v", err)
		return
	}
	if res == nil || res.List == nil || len(res.List) == 0 {
		log.Warnc(ctx, "[Dao][LiveRoom][GetMultiple][Res][Empty], res:%+v, err:%+v", res, err)
		err = xecode.Errorf(xecode.RequestErr, "获取房间信息失败")
		return
	}
	roomInfos = res.List
	return
}
