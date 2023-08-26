package dynamic

import (
	"context"
	"fmt"

	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"

	dyngrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/feed"
	dynsharegrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/publish"
)

type Dao struct {
	dynGrpc      dyngrpc.FeedClient
	dynShareGrpc dynsharegrpc.PublishClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.dynGrpc, err = dyngrpc.NewClient(c.DynGRPC); err != nil {
		panic(fmt.Sprintf("dynGrpc NewClient error(%v)", err))
	}
	if d.dynShareGrpc, err = dynsharegrpc.NewClient(c.DynShareGRPC); err != nil {
		panic(fmt.Sprintf("dynShareGrpc NewClient error(%v)", err))
	}
	return
}

func (d *Dao) DynSimpleInfos(ctx context.Context, args *dyngrpc.DynSimpleInfosReq) (*dyngrpc.DynSimpleInfosRsp, error) {
	return d.dynGrpc.DynSimpleInfos(ctx, args)
}

func (d *Dao) GetReserveDynShareContent(ctx context.Context, args *dynsharegrpc.GetReserveDynShareContentReq) (*dynsharegrpc.GetReserveDynShareContentRsp, error) {
	return d.dynShareGrpc.GetReserveDynShareContent(ctx, args)
}
