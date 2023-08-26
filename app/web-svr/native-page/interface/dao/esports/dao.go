package esports

import (
	"context"

	esportsgrpc "git.bilibili.co/bapis/bapis-go/operational/esportsservice"

	"go-gateway/app/web-svr/native-page/interface/conf"
)

type Dao struct {
	client esportsgrpc.EsportsServiceClient
}

func NewDao(cfg *conf.Config) *Dao {
	esports, err := esportsgrpc.NewClient(cfg.EsportsGRPC)
	if err != nil {
		panic(err)
	}
	return &Dao{client: esports}
}

func (d *Dao) GetSportsSeasonMedalTable(c context.Context, seasonId int64, typ esportsgrpc.MedalTableTypeEnum) (*esportsgrpc.GetSportsSeasonMedalTableResponse, error) {
	req := &esportsgrpc.GetSportsSeasonMedalTableReq{
		SeasonId: seasonId,
		Type:     typ,
	}
	rly, err := d.client.GetSportsSeasonMedalTable(c, req)
	if err != nil {
		return nil, err
	}
	return rly, nil
}

func (d *Dao) GetSportsEventMatches(c context.Context, ids []int64) (*esportsgrpc.GetSportsEventMatchesResponse, error) {
	req := &esportsgrpc.GetSportsEventMatchesReq{Ids: ids}
	rly, err := d.client.GetSportsEventMatches(c, req)
	if err != nil {
		return nil, err
	}
	return rly, nil
}
