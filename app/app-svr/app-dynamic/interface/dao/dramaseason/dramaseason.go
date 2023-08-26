package dramaseason

import (
	"context"

	"go-gateway/app/app-svr/app-dynamic/interface/conf"

	dramaseasongrpc "git.bilibili.co/bapis/bapis-go/maoer/drama/dramaseason"
)

type Dao struct {
	c                 *conf.Config
	dramaseasonClient dramaseasongrpc.DramaSeasonClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c: c,
	}
	var err error
	if d.dramaseasonClient, err = dramaseasongrpc.NewClient(c.DramaseasonGRPC); err != nil {
		panic(err)
	}
	return
}

func (d *Dao) FeedCardDrama(ctx context.Context, dramaIDs []int64) (map[int64]*dramaseasongrpc.FeedCardDramaInfo, error) {
	arg := &dramaseasongrpc.GetFeedCardDramaReq{
		DramaIds: dramaIDs,
	}
	reply, err := d.dramaseasonClient.GetFeedCardDrama(ctx, arg)
	if err != nil {
		return nil, err
	}
	return reply.GetItems(), nil
}
