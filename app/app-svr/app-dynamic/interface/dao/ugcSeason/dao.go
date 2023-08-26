package ugcSeason

import (
	"context"

	"go-common/library/log"

	"go-gateway/app/app-svr/app-dynamic/interface/conf"
	ugcseasongrpc "go-gateway/app/app-svr/ugc-season/service/api"
)

type Dao struct {
	c          *conf.Config
	grpcClient ugcseasongrpc.UGCSeasonClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c: c,
	}
	var err error
	if d.grpcClient, err = ugcseasongrpc.NewClient(c.UGCSeasonGRPC); err != nil {
		panic(err)
	}
	return
}

func (d *Dao) Seasons(c context.Context, ids []int64) (map[int64]*ugcseasongrpc.Season, error) {
	res, err := d.grpcClient.Seasons(c, &ugcseasongrpc.SeasonsRequest{SeasonIds: ids})
	if err != nil {
		log.Error("Seasons %v", err)
		return nil, err
	}
	return res.GetSeasons(), nil
}
