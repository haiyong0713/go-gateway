package dao

import (
	"context"

	roomgategrpc "git.bilibili.co/bapis/bapis-go/live/xroom-gate"
)

type roomGateDao struct {
	client roomgategrpc.XroomgateClient
}

func (d *roomGateDao) SessionInfoBatch(ctx context.Context, req *roomgategrpc.SessionInfoBatchReq) (*roomgategrpc.SessionInfoBatchResp, error) {
	rly, err := d.client.SessionInfoBatch(ctx, req)
	if err != nil {
		return nil, err
	}
	return rly, nil
}
