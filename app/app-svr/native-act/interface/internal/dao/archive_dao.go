package dao

import (
	"context"

	arcgrpc "go-gateway/app/app-svr/archive/service/api"
)

type archiveDao struct {
	client arcgrpc.ArchiveClient
}

func (d *archiveDao) Arcs(c context.Context, req *arcgrpc.ArcsRequest) (map[int64]*arcgrpc.Arc, error) {
	rly, err := d.client.Arcs(c, req)
	if err != nil {
		return nil, err
	}
	if rly == nil {
		return map[int64]*arcgrpc.Arc{}, nil
	}
	return rly.Arcs, nil
}

func (d *archiveDao) ArcsPlayer(c context.Context, req *arcgrpc.ArcsPlayerRequest) (map[int64]*arcgrpc.ArcPlayer, error) {
	rly, err := d.client.ArcsPlayer(c, req)
	if err != nil {
		return nil, err
	}
	if rly == nil {
		return map[int64]*arcgrpc.ArcPlayer{}, nil
	}
	return rly.GetArcsPlayer(), nil
}
