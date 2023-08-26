package garb

import (
	"context"

	"go-common/library/log"

	"go-gateway/app/app-svr/app-dynamic/interface/conf"

	garbmdl "git.bilibili.co/bapis/bapis-go/garb/model"
	garbgrpc "git.bilibili.co/bapis/bapis-go/garb/service"
)

type Dao struct {
	c          *conf.Config
	grpcClient garbgrpc.GarbClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c: c,
	}
	var err error
	if d.grpcClient, err = garbgrpc.NewClient(c.GarbGRPC); err != nil {
		panic(err)
	}
	return
}

func (d *Dao) Decorations(c context.Context, mid int64, ids []int64) (map[int64]*garbmdl.DynamicGarbInfo, error) {
	resTmp, err := d.grpcClient.DynamicGarbInfo(c, &garbgrpc.DynamicGarbInfoReq{Mid: mid, ItemIDs: ids})
	if err != nil {
		log.Error("Decorations error %v", err)
		return nil, err
	}
	return resTmp.GetDynamicGarbInfo(), nil
}
