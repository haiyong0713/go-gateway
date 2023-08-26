package pedia

import (
	"context"
	"fmt"

	"go-gateway/app/app-svr/app-channel/interface/conf"

	baikegrpc "git.bilibili.co/bapis/bapis-go/community/interface/baike"
)

type Dao struct {
	// grpc
	baikegrpc baikegrpc.BaikeClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.baikegrpc, err = baikegrpc.NewClientBaike(c.BaikeGRPC); err != nil {
		panic(fmt.Sprintf("baikegrpc NewClient error (%+v)", err))
	}
	return
}

func (d *Dao) BaikeDetails(ctx context.Context, req *baikegrpc.BaikeDetailReq) (*baikegrpc.BaikeDetailRsp, error) {
	return d.baikegrpc.BaikeDetail(ctx, req)
}

func (d *Dao) BaikeFeed(ctx context.Context, req *baikegrpc.BaikeFeedReq) (*baikegrpc.BaikeFeedRsp, error) {
	return d.baikegrpc.BaikeFeed(ctx, req)
}
