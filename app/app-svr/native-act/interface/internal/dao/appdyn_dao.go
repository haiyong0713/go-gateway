package dao

import (
	"context"

	appdyngrpc "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
)

type appdynDao struct {
	client appdyngrpc.DynamicClient
}

func (d *appdynDao) DynServerDetails(c context.Context, req *appdyngrpc.DynServerDetailsReq) (map[int64]*appdyngrpc.DynamicItem, error) {
	rly, err := d.client.DynServerDetails(c, req)
	if err != nil {
		return nil, err
	}
	return rly.Items, nil
}
