package dao

import (
	"context"

	populargrpc "git.bilibili.co/bapis/bapis-go/manager/service/popular"
)

type popularDao struct {
	client populargrpc.PopularClient
}

func (d *popularDao) PageArcs(c context.Context, req *populargrpc.PageArcsReq) (*populargrpc.PageArcsResp, error) {
	rly, err := d.client.PageArcs(c, req)
	if err != nil {
		return nil, err
	}
	return rly, nil
}

func (d *popularDao) TimeLine(c context.Context, req *populargrpc.TimeLineRequest) (*populargrpc.TimeLineReply, error) {
	rly, err := d.client.TimeLine(c, req)
	if err != nil {
		return nil, err
	}
	return rly, nil
}
