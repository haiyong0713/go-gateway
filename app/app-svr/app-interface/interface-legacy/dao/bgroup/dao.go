package bgroup

import (
	"context"

	bgroupgrpc "git.bilibili.co/bapis/bapis-go/platform/service/bgroup/v2"

	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"
)

type Dao struct {
	client bgroupgrpc.BGroupServiceClient
}

func NewDao(cfg *conf.Config) *Dao {
	client, err := bgroupgrpc.NewClient(cfg.BgroupGRPC)
	if err != nil {
		panic(err)
	}
	return &Dao{client: client}
}

func (d *Dao) MemberIn(ctx context.Context, req *bgroupgrpc.MemberInReq) (map[string]bool, error) {
	rly, err := d.client.MemberIn(ctx, req)
	if err != nil {
		return nil, err
	}
	inMap := make(map[string]bool, len(rly.Results))
	for _, v := range rly.Results {
		if v == nil {
			continue
		}
		inMap[v.Name] = v.In
	}
	return inMap, nil
}
