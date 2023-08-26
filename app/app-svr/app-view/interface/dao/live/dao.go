package live

import (
	"context"
	"fmt"

	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/app-view/interface/conf"
	"go-gateway/app/app-svr/app-view/interface/model/live"

	livexroom "git.bilibili.co/bapis/bapis-go/live/xroom"
)

const (
	_entryFrom        = "ugc_video_detail"
	_endPageEntryFrom = "ugc_video_close"
	_reqBiz           = "/bilibili.app.view.v1.View/View"
)

// Dao is space dao
type Dao struct {
	roomRPCClient livexroom.RoomClient
}

// New initial space dao
func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.roomRPCClient, err = livexroom.NewClient(c.LiveClient); err != nil {
		panic(fmt.Sprintf("livexroom newLiveRoomClient error (%+v)", err))
	}
	return
}

// LivingRoom is get living rooms from api
func (d *Dao) LivingRoom(c context.Context, uid int64, platform, brand, net string, build int, mid int64) (*live.Live, error) {
	req := &livexroom.EntryRoomInfoReq{
		EntryFrom:     []string{_entryFrom, _endPageEntryFrom},
		Uids:          []int64{uid}, //up主
		Uid:           mid,          //用户
		Uipstr:        metadata.String(c, metadata.RemoteIP),
		Platform:      platform,
		Build:         int64(build),
		DeviceName:    brand,
		Network:       net,
		FilterOffline: 1, //只获取在播的
		ReqBiz:        _reqBiz,
	}
	resp, err := d.roomRPCClient.EntryRoomInfo(c, req)
	if err != nil {
		log.Error("d.roomRPCClient.EntryRoomInfo uid(%d) error(%v)", uid, err)
		return nil, err
	}
	if resp == nil || len(resp.List) == 0 {
		return nil, fmt.Errorf("LivingRoom uid(%d) no response", uid)
	}
	v, ok := resp.List[uid]
	if !ok || v == nil {
		return nil, fmt.Errorf("LivingRoom uid(%d) no response", uid)
	}
	if v.LiveStatus != 1 {
		return nil, fmt.Errorf("uid(%d) is not living", uid)
	}
	l := &live.Live{
		Mid:        v.Uid,
		RoomID:     v.RoomId,
		URI:        v.JumpUrl[_entryFrom],
		EndPageUri: v.JumpUrl[_endPageEntryFrom],
	}
	return l, nil
}
