package dao

import (
	"context"
	"go-common/library/log"
	"go-gateway/app/web-svr/web/interface/model"

	vtradegrpc "git.bilibili.co/bapis/bapis-go/vas/trans/trade/service"
)

const (
	_platformPC = "web"
)

// pay season trade create
func (d *Dao) TradeCreate(ctx context.Context, req *model.TradeCreateReq) (*vtradegrpc.TradeCreateReply, error) {
	params := &vtradegrpc.TradeCreateReq{
		Platform:  _platformPC,
		Mid:       req.Mid,
		From:      req.SpmID,
		ProductId: req.ProductId,
	}
	res, err := d.tradeGRPC.TradeCreate(ctx, params)
	if err != nil {
		log.Error("【TradeCreate】Fail to create trade, params is %v", params)
		return nil, err
	}
	return res, nil
}
