package live

import (
	"context"

	"go-common/library/ecode"
	"go-common/library/net/metadata"

	"go-gateway/app/web-svr/native-page/interface/conf"

	playgrpc "git.bilibili.co/bapis/bapis-go/live/live-play/v1"
	livegrpc "git.bilibili.co/bapis/bapis-go/live/xroom-feed"
	roomgategrpc "git.bilibili.co/bapis/bapis-go/live/xroom-gate"
)

type Dao struct {
	liveClient     livegrpc.DynamicClient
	playClient     playgrpc.TopicClient
	roomGateClient roomgategrpc.XroomgateClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.liveClient, err = livegrpc.NewClient(c.LiveClient); err != nil {
		panic(err)
	}
	if d.playClient, err = playgrpc.NewClient(c.LiveClient); err != nil {
		panic(err)
	}
	if d.roomGateClient, err = roomgategrpc.NewClientXroomgate(c.RoomGateClient); err != nil {
		panic(err)
	}
	return
}

// GetCardInfo .
func (d *Dao) GetCardInfo(c context.Context, roomIDs []int64, mid int64, isHttps bool) (map[uint64]*livegrpc.LiveCardInfo, error) {
	ip := metadata.String(c, metadata.RemoteIP)
	rly, err := d.liveClient.GetCardInfo(c, &livegrpc.GetCardInfoReq{RoomIds: roomIDs, Uid: mid, Ip: ip, Platform: "web", IsHttps: isHttps})
	if err != nil {
		return nil, err
	}
	if rly != nil {
		return rly.LivePlayInfo, nil
	}
	return nil, nil
}

func (d *Dao) GetListByRoomId(c context.Context, roomIDs []int64, isLive int64) (map[int64]*playgrpc.RoomList, error) {
	rly, err := d.playClient.GetListByRoomId(c, &playgrpc.GetListByRoomIdReq{RoomIds: roomIDs, Filter: &playgrpc.Filter{IsLive: isLive}})
	if err != nil {
		return nil, err
	}
	if rly == nil {
		return nil, ecode.NothingFound
	}
	res := make(map[int64]*playgrpc.RoomList)
	for _, v := range rly.List {
		if v == nil {
			continue
		}
		res[v.RoomId] = v
	}
	return res, nil
}

func (d *Dao) GetListByActId(c context.Context, actID, tabID, isLive, ps, offset int64) (*playgrpc.GetListByActIdResp, error) {
	return d.playClient.GetListByActId(c, &playgrpc.GetListByActIdReq{ActId: actID, TabId: tabID, PageSize: ps, Offset: offset, Filter: &playgrpc.Filter{IsLive: isLive}})
}

func (d *Dao) SessionInfoBatch(c context.Context, uidsMap map[int64][]string, entryFrom []string) (map[int64]*roomgategrpc.SessionInfos, error) {
	reqUids := make(map[int64]*roomgategrpc.LiveIds)
	for k, liveIDs := range uidsMap {
		reqUids[k] = &roomgategrpc.LiveIds{LiveIds: liveIDs}
	}
	req := &roomgategrpc.SessionInfoBatchReq{
		UidLiveIds: reqUids,
		EntryFrom:  entryFrom,
	}
	seRly, e := d.roomGateClient.SessionInfoBatch(c, req)
	if e != nil { //错误降级处理
		return nil, e
	}
	if seRly == nil {
		return make(map[int64]*roomgategrpc.SessionInfos), nil
	}
	return seRly.List, nil
}
