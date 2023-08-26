package season

import (
	"context"
	"fmt"

	"go-common/library/log"

	"go-gateway/app/web-svr/web-show/interface/conf"

	seasongrpc "git.bilibili.co/bapis/bapis-go/cheese/service/season/season"
)

type Dao struct {
	c         *conf.Config
	rpcClient seasongrpc.SeasonClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c: c,
	}
	var err error
	if d.rpcClient, err = seasongrpc.NewClient(c.SeasonGRPC); err != nil {
		panic(fmt.Sprintf("rpcClient NewClientt error (%+v)", err))
	}
	return
}

func (d *Dao) CardsByEpIds(c context.Context, epids []int32) (res *seasongrpc.CardsByEpIdsReply, err error) {
	args := &seasongrpc.CardsByEpIdsReq{EpIds: epids}
	if res, err = d.rpcClient.CardsByEpIds(c, args); err != nil {
		log.Error("%v", err)
	}
	return
}
