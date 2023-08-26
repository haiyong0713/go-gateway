package http

import (
	"encoding/json"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	pb "go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/interface/model/like"
	"go-gateway/app/web-svr/activity/interface/service"
)

func syncOrders(ctx *bm.Context) {
	p := new(like.ParamMsg)
	if err := ctx.Bind(p); err != nil {
		return
	}

	var err error
	syncStuct := pb.StockServerSyncStruct{}
	if err = json.Unmarshal([]byte(p.Msg), &syncStuct); err != nil {
		log.Errorc(ctx, "syncOrders json.Unmarshal msg(%s) error(%v)", p.Msg, err)
		return
	}
	log.Infoc(ctx, "syncOrders syncStuct :%v", syncStuct)
	ctx.JSON(service.StockSvr.AckStockOrders(ctx, &pb.FeedBackStocksReq{
		StockId:  syncStuct.StockId,
		StockNos: syncStuct.StockOrders,
		Ts:       syncStuct.Ts,
	}))
}
