package dao

import (
	"context"

	actplatv2grpc "git.bilibili.co/bapis/bapis-go/platform/interface/act-plat-v2"
)

type actplatv2Dao struct {
	client actplatv2grpc.ActPlatClient
}

func (d *actplatv2Dao) GetHistory(c context.Context, req *actplatv2grpc.GetHistoryReq) (*actplatv2grpc.GetHistoryResp, error) {
	rly, err := d.client.GetHistory(c, req)
	if err != nil {
		return nil, err
	}
	return rly, nil
}

func (d *actplatv2Dao) GetCounterRes(c context.Context, req *actplatv2grpc.GetCounterResReq) (*actplatv2grpc.GetCounterResResp, error) {
	rly, err := d.client.GetCounterRes(c, req)
	if err != nil {
		return nil, err
	}
	return rly, nil
}

func (d *actplatv2Dao) GetTotalRes(c context.Context, req *actplatv2grpc.GetTotalResReq) (*actplatv2grpc.GetTotalResResp, error) {
	rly, err := d.client.GetTotalRes(c, req)
	if err != nil {
		return nil, err
	}
	return rly, nil
}
