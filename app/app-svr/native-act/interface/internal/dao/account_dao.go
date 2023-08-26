package dao

import (
	"context"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	"go-common/library/net/metadata"
)

type accountDao struct {
	client accountgrpc.AccountClient
}

func (d *accountDao) Infos3(c context.Context, mids []int64) (map[int64]*accountgrpc.Info, error) {
	req := &accountgrpc.MidsReq{Mids: mids, RealIp: metadata.String(c, metadata.RemoteIP)}
	rly, err := d.client.Infos3(c, req)
	if err != nil {
		return nil, err
	}
	return rly.Infos, nil
}

func (d *accountDao) Cards3(c context.Context, req *accountgrpc.MidsReq) (map[int64]*accountgrpc.Card, error) {
	rly, err := d.client.Cards3(c, req)
	if err != nil {
		return nil, err
	}
	return rly.Cards, nil
}
