package dao

import (
	"context"

	appshowgrpc "git.bilibili.co/bapis/bapis-go/app/show/v1"
)

type appshowDao struct {
	client appshowgrpc.AppShowClient
}

func (d *appshowDao) SelectedSerie(c context.Context, req *appshowgrpc.SelectedSerieReq) (*appshowgrpc.SelectedSerieRly, error) {
	rly, err := d.client.SelectedSerie(c, req)
	if err != nil {
		return nil, err
	}
	return rly, nil
}

func (d *appshowDao) BatchSerie(c context.Context, req *appshowgrpc.BatchSerieReq) (map[int64]*appshowgrpc.SerieConfig, error) {
	rly, err := d.client.BatchSerie(c, req)
	if err != nil {
		return nil, err
	}
	if rly != nil {
		return rly.List, nil
	}
	return nil, nil
}
