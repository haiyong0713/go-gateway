package dao

import (
	"context"
	"fmt"

	"go-gateway/app/app-svr/app-listener/interface/internal/model"

	listenerSvc "git.bilibili.co/bapis/bapis-go/dynamic/service/listener"
)

type PickCardsOpt struct {
	Mid, Offset int64
	Buvid       string
}

// 获取播单组
func (d *dao) PickCards(ctx context.Context, opt PickCardsOpt) ([]model.SinglePick, int64, error) {
	req := &listenerSvc.GetCollectionGroupsReq{
		Offset: opt.Offset, Mid: opt.Mid, Buvid: opt.Buvid,
	}
	resp, err := d.listenerGRPC.GetCollectionGroups(ctx, req)
	if err != nil {
		return nil, 0, wrapDaoError(err, "listenerGRPC.GetCollectionGroups", req)
	}
	ret := make([]model.SinglePick, 0, len(resp.Data))
	for i := range resp.GetData() {
		ret = append(ret, model.SinglePick{Pick: resp.Data[i]})
	}
	return ret, resp.Offset, nil
}

type CardDetailsOpt struct {
	CardId int64
	PickId int64
}

// 获取单个播单详情
func (d *dao) CardDetail(ctx context.Context, opt CardDetailsOpt) (ret model.SingleCollection, err error) {
	req := &listenerSvc.GetCollectionReq{CollectionId: opt.CardId, PickId: opt.PickId}
	resp, err := d.listenerGRPC.GetCollection(ctx, req)
	if err != nil {
		err = wrapDaoError(err, "listenerGRPC.GetCollection", req)
		return
	}
	if resp == nil {
		err = fmt.Errorf("dao.listenerGRPC.GetCollection: unexpected nil Collection while GetCollection(%d)", opt.CardId)
		return
	}
	ret.PickId = opt.PickId
	ret.Collection = resp
	return
}
