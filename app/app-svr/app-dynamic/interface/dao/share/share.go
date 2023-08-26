package share

import (
	"context"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-dynamic/interface/conf"

	shareApi "git.bilibili.co/bapis/bapis-go/community/interface/share"
)

type Dao struct {
	c         *conf.Config
	shareGRPC shareApi.ShareClient
}

func New(c *conf.Config) *Dao {
	d := &Dao{
		c: c,
	}
	var err error
	if d.shareGRPC, err = shareApi.NewClient(c.ShareGRPC); err != nil {
		panic(err)
	}
	return d
}

func (d *Dao) BusinessChannels(ctx context.Context, req *shareApi.BusinessChannelsReq) ([]*shareApi.ShareChannel, error) {
	reply, err := d.shareGRPC.BusinessChannels(ctx, req)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	return reply.GetShareChannels(), nil
}
