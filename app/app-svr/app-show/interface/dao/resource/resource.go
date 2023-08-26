package resource

import (
	"context"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-show/interface/conf"
	resource "go-gateway/app/app-svr/resource/service/model"
	resrpc "go-gateway/app/app-svr/resource/service/rpc/client"
)

type Dao struct {
	c *conf.Config
	// rpc
	resRpc *resrpc.Service
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c: c,
		// rpc
		resRpc: resrpc.New(c.ResourceRPC),
	}
	return
}

func (d *Dao) ResBanner(ctx context.Context, plat int8, build int, mid int64, resIDStr, channel, ip, buvid, network, mobiApp, device, adExtra string, isAd bool) (res map[int][]*resource.Banner, err error) {
	arg := &resource.ArgBanner{
		Plat:    plat,
		ResIDs:  resIDStr,
		Build:   build,
		MID:     mid,
		Channel: channel,
		IP:      ip,
		Buvid:   buvid,
		Network: network,
		MobiApp: mobiApp,
		Device:  device,
		IsAd:    isAd,
		AdExtra: adExtra,
	}
	var bs *resource.Banners
	if bs, err = d.resRpc.Banners(ctx, arg); err != nil || bs == nil {
		log.Error("d.resRpc.Banners(%v) error(%v)", arg, err)
		return
	}
	if len(bs.Banner) > 0 {
		res = bs.Banner
	}
	return
}
