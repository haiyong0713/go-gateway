package dao

import (
	"context"

	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
)

type relationDao struct {
	relation relationgrpc.RelationClient
}

func (d *relationDao) StatsGRPC(ctx context.Context, mids []int64) (map[int64]*relationgrpc.StatReply, error) {
	req := &relationgrpc.MidsReq{Mids: mids}
	statsReply, err := d.relation.Stats(ctx, req)
	if err != nil {
		return nil, err
	}
	return statsReply.StatReplyMap, nil
}
