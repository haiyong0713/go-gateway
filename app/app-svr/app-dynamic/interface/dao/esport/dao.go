package esport

import (
	"context"

	"go-gateway/app/app-svr/app-dynamic/interface/conf"

	esportGrpc "git.bilibili.co/bapis/bapis-go/operational/esportsservice"
)

type Dao struct {
	c          *conf.Config
	esportgrpc esportGrpc.EsportsServiceClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c: c,
	}
	var err error
	if d.esportgrpc, err = esportGrpc.NewClient(c.EsportsGRPC); err != nil {
		panic(err)
	}
	return
}

func (d *Dao) AdditionalEsport(c context.Context, mid int64, esportIDs []int64) (map[int64]*esportGrpc.ContestDetail, error) {
	resTmp, err := d.esportgrpc.GetContests(c, &esportGrpc.GetContestsRequest{Mid: mid, Cids: esportIDs})
	if err != nil {
		return nil, err
	}
	var res = make(map[int64]*esportGrpc.ContestDetail)
	for _, contest := range resTmp.Contests {
		res[contest.ID] = contest
	}
	return res, nil
}
