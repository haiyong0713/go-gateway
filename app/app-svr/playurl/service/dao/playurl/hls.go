package playurl

import (
	"context"

	"go-common/library/ecode"
	"go-common/library/log"
	v2 "go-gateway/app/app-svr/playurl/service/api/v2"
	"go-gateway/app/app-svr/playurl/service/model"

	hlsgrpc "git.bilibili.co/bapis/bapis-go/video/vod/playurlhls"
)

// HlsScheduler 获取hls播放列表.
func (d *Dao) HlsScheduler(c context.Context, arg *hlsgrpc.M3U8RequestMsg) (*v2.HlsResponseMsg, error) {
	rly, err := d.playurlhlsGRPC.HlsScheduler(c, arg)
	if err != nil {
		return nil, err
	}
	if rly == nil {
		return nil, ecode.NothingFound
	}
	if rly.Code != uint32(ecode.OK.Code()) {
		log.Error("HlsScheduler cid(%d) code(%d) arg(%+v)", arg.Cid, rly.Code, arg)
		return nil, ecode.NothingFound
	}
	res := new(v2.HlsResponseMsg)
	res.FromPlayurlHls(rly)
	return res, nil
}

// MasterScheduler .
func (d *Dao) MasterScheduler(c context.Context, arg *hlsgrpc.M3U8RequestMsg, dc *model.DolbyConf) (*v2.MasterScheduler, error) {
	rly, err := d.playurlhlsGRPC.MasterScheduler(c, arg)
	if err != nil {
		return nil, err
	}
	if rly == nil {
		return nil, ecode.NothingFound
	}
	if rly.Code != uint32(ecode.OK.Code()) {
		log.Error("MasterScheduler cid(%d) code(%d) arg(%+v)", arg.Cid, rly.Code, arg)
		return nil, ecode.NothingFound
	}
	res := new(v2.MasterScheduler)
	supportDolby := dc.SupportDolby()
	res.FromPlayurlMaster(rly, supportDolby)
	return res, nil
}

// M3U8Scheduler
func (d *Dao) M3U8Scheduler(c context.Context, arg *hlsgrpc.M3U8RequestMsg) (*v2.M3U8ResponseMsg, error) {
	rly, err := d.playurlhlsGRPC.M3U8Scheduler(c, arg)
	if err != nil {
		return nil, err
	}
	if rly == nil {
		return nil, ecode.NothingFound
	}
	if rly.Code != uint32(ecode.OK.Code()) {
		log.Error("M3U8Scheduler cid(%d) code(%d) arg(%+v)", arg.Cid, rly.Code, arg)
		return nil, ecode.NothingFound
	}
	res := new(v2.M3U8ResponseMsg)
	res.FromPlayurlM3U8(rly)
	return res, nil
}
