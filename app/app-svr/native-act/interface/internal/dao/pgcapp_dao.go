package dao

import (
	"context"

	pgcappgrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card/app"
)

type pgcappDao struct {
	client pgcappgrpc.AppCardClient
}

func (d *pgcappDao) QueryWid(c context.Context, req *pgcappgrpc.QueryWidReq) (*pgcappgrpc.QueryWidReply, error) {
	rly, err := d.client.QueryWid(c, req)
	if err != nil {
		return nil, err
	}
	return rly, nil
}

func (d *pgcappDao) SeasonBySeasonId(c context.Context, req *pgcappgrpc.SeasonBySeasonIdReq) (*pgcappgrpc.SeasonBySeasonIdReply, error) {
	rly, err := d.client.SeasonBySeasonId(c, req)
	if err != nil {
		return nil, err
	}
	return rly, nil
}

func (d *pgcappDao) SeasonByPlayId(c context.Context, req *pgcappgrpc.SeasonByPlayIdReq) (*pgcappgrpc.SeasonByPlayIdReply, error) {
	rly, err := d.client.SeasonByPlayId(c, req)
	if err != nil {
		return nil, err
	}
	return rly, nil
}
