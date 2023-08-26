package vcloud

import (
	"context"
	"fmt"

	"go-gateway/app/app-svr/app-view/interface/conf"

	vcloud "git.bilibili.co/bapis/bapis-go/video/vod/playurlstory"
)

type Dao struct {
	playurlClient vcloud.PlayurlServiceClient
}

func New(c *conf.Config) *Dao {
	client, err := vcloud.NewClient(c.VCloudClient)
	if err != nil {
		panic(fmt.Sprintf("vcloud NewClient error(%v)", err))
	}
	return &Dao{
		playurlClient: client,
	}
}

func (d *Dao) ShortFormVideoInfo(ctx context.Context, param *vcloud.RequestMsg, cid int64) (*vcloud.ResponseItem, error) {
	reply, err := d.playurlClient.Playurl(ctx, param)
	if err != nil {
		return nil, err
	}
	return reply.Data[uint64(cid)], nil
}
