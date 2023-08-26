package dao

import (
	"context"

	chargrpc "git.bilibili.co/bapis/bapis-go/pgc/service/media"
)

type characterDao struct {
	client chargrpc.CharacterClient
}

func (d *characterDao) RelInfos(c context.Context, req *chargrpc.CharacterIdsOidsReq) (*chargrpc.CharacterRelInfosReply, error) {
	rly, err := d.client.RelInfos(c, req)
	if err != nil {
		return nil, err
	}
	return rly, nil
}
