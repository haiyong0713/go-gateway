package trade

import (
	"context"
	"fmt"

	tradegrpc "git.bilibili.co/bapis/bapis-go/vas/trans/trade/service"
	"go-gateway/app/app-svr/app-view/interface/conf"
)

type Dao struct {
	tradeClient tradegrpc.VasTransTradeClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.tradeClient, err = tradegrpc.NewClientVasTransTrade(c.TradeClient); err != nil {
		panic(fmt.Sprintf("trade NewClient not found err(%v)", err))
	}
	return
}

func (d *Dao) TradeOrderStateInfo(ctx context.Context, mid int64, orderID string) (*tradegrpc.TradeOrderStateInfoReply, error) {
	req := &tradegrpc.TradeOrderStateInfoReq{
		Mid:     mid,
		OrderId: orderID,
	}
	return d.tradeClient.TradeOrderStateInfo(ctx, req)
}

func (d *Dao) TradeOrderCreate(ctx context.Context, mid, build int64, productID, platform, from string) (*tradegrpc.TradeCreateReply, error) {
	req := &tradegrpc.TradeCreateReq{
		ProductId: productID,
		Platform:  platform,
		Build:     build,
		Mid:       mid,
		From:      from,
	}
	return d.tradeClient.TradeCreate(ctx, req)
}
