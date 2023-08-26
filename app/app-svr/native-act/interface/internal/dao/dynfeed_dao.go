package dao

import (
	"context"

	dynfeedgrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/feed"
)

type dynfeedDao struct {
	client dynfeedgrpc.FeedClient
}

func (d *dynfeedDao) FetchDynIdByRevs(c context.Context, req *dynfeedgrpc.FetchDynIdByRevsReq) (*dynfeedgrpc.FetchDynIdByRevsRsp, error) {
	rly, err := d.client.FetchDynIdByRevs(c, req)
	if err != nil {
		return nil, err
	}
	return rly, nil
}
