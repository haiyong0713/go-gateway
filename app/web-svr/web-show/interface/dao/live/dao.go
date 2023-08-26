package live

import (
	"context"
	"fmt"
	"go-gateway/app/web-svr/web-show/interface/model/resource"

	"go-gateway/app/web-svr/web-show/interface/conf"

	"go-common/library/log"

	livegrpcmdl "git.bilibili.co/bapis/bapis-go/live/xroom"
	livegrpc "git.bilibili.co/bapis/bapis-go/live/xroom-gate"
)

type Dao struct {
	c         *conf.Config
	rpcClient livegrpc.XroomgateClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c: c,
	}
	var err error
	if d.rpcClient, err = livegrpc.NewClientXroomgate(c.LiveGRPC); err != nil {
		panic(fmt.Sprintf("rpcClient NewClientt error (%+v)", err))
	}
	return
}

func (d *Dao) RoomByIds(c context.Context, roomIds []int64) (res map[int64]*resource.LiveRoomInfo, err error) {
	args := &livegrpc.EntryRoomInfoReq{
		RoomIds:    roomIds,
		EntryFrom:  []string{"NONE"},
		NotPlayurl: 1,
		ReqBiz:     "/x/web-show/res/locs",
	}
	var reply *livegrpc.EntryRoomInfoResp
	if reply, err = d.rpcClient.EntryRoomInfo(c, args); err != nil {
		log.Error("%v", err)
		return
	}
	res = make(map[int64]*resource.LiveRoomInfo, len(roomIds))
	for _, id := range roomIds {
		res[id] = nil
	}
	for _, value := range reply.List {
		if _, ok := res[value.RoomId]; ok {
			res[value.RoomId] = packingInfo(value)
		}
		if value.ShortId != 0 {
			if _, ok := res[value.ShortId]; ok {
				res[value.ShortId] = packingInfo(value)
			}
		}
	}
	for id, value := range res {
		if value == nil {
			delete(res, id)
		}
	}
	return
}

func packingInfo(list *livegrpc.EntryRoomInfoResp_EntryList) *resource.LiveRoomInfo {
	if list == nil {
		return nil
	}
	info := &resource.LiveRoomInfo{
		RoomId: list.RoomId,
		Uid:    list.Uid,
		Show: &livegrpcmdl.RoomShowInfo{
			ShortId:         list.ShortId,
			Title:           list.Title,
			Cover:           list.Cover,
			Keyframe:        list.Keyframe,
			PopularityCount: list.PopularityCount,
		},
		Area: &livegrpcmdl.RoomAreaInfo{
			AreaName: list.AreaName,
		},
		WatchedShow: list.WatchedShow,
	}
	return info
}
