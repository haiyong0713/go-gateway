package service

import (
	"context"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/ecode"
	pb "go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/job/model/stock"
	"time"
)

func (s *Service) StockServerSyncJob() {
	var (
		ctx             = context.Background()
		nowTime         = time.Now().Unix()
		stockActListReq = &pb.EffectiveStockListReq{
			BeginTime: nowTime,
			EndTime:   nowTime,
			PageSize:  s.c.StockServerJobConf.SyncPageSize,
		}
		GetStockOrderByIdReq = &pb.GetStockOrderByIdReq{
			SyncNum: s.c.StockServerJobConf.SyncNum,
		}
		stockActListResp *pb.EffectiveStockListResp
		err              error
	)
	for i := 1; i < s.c.StockServerJobConf.SyncMaxLoop; i++ {
		stockActListReq.PageNumber = int32(i)
		if stockActListResp, err = s.actGRPC.EffectiveStockList(ctx, stockActListReq); err != nil || stockActListResp == nil {
			log.Errorc(ctx, "StockServerSyncJob EffectiveStockList err:%+v , stockActListResp:%v", err, stockActListResp)
			break
		}

		for _, v := range stockActListResp.List {
			url, ok := s.c.StockServerJobConf.AckUrlMap[v.ResourceId]
			if !ok || url == "" {
				log.Infoc(ctx, "StockServerSyncJob can not  find StockServerJobConf:%v", v.ResourceId)
				continue
			}
			GetStockOrderByIdReq.StockId = v.StockId
			var orders *pb.GetStockOrderByIdResp
			if orders, err = s.actGRPC.GetStockOrderById(ctx, GetStockOrderByIdReq); err != nil || orders == nil {
				continue
			}

			for _, item := range orders.List {
				if err = s.syncSingleRecord(ctx, nowTime, url, v.StockId, item); err != nil {
					log.Errorc(ctx, "StockServerSyncJob syncSingleRecord err:%v", err)
				}
			}
		}

		// 没有数据了，主动退出
		if len(stockActListResp.List) <= int(stockActListReq.PageSize) {
			log.Infoc(ctx, "StockServerSyncJob end, loop:%v , list_size:%v", i, len(stockActListResp.List))
			break
		}
	}
	return
}

func (s *Service) syncSingleRecord(ctx context.Context, nowTime int64, url string, stockId int64, item *pb.GetStockOrderByIdItem) (err error) {
	if item.CreateTime+s.c.StockServerJobConf.SyncTimeGap > nowTime {
		log.Infoc(ctx, "StockServerSyncJob new order: %v  , now time:%v", item.StockNo, nowTime)
		return ecode.ActivityTaskNotStart
	}
	var ok bool
	ok, err = s.stockDao.SyncStockWithOtherPlatform(ctx, url, &stock.SyncParamStruct{
		StockNo: item.StockNo,
		RetryId: item.UniqueId,
	})

	if err != nil {
		log.Warnc(ctx, "StockServerSyncJob SyncStockWithOtherPlatform stock_id :%v , stock_no:%v , err:%v", stockId, item.StockNo, err)
		return
	}

	if ok {
		var replay interface{}
		replay, err = s.actGRPC.AckStockOrders(ctx, &pb.FeedBackStocksReq{
			StockId:  stockId,
			StockNos: []string{item.StockNo},
			Ts:       item.CreateTime,
		})
		log.Infoc(ctx, "StockServerSyncJob AckStockOrders stock_id:%v , stock_no:%v , replay:%v , err:%v", stockId, item.StockNo, replay, err)
		return
	}

	/*replay, err := s.actGRPC.FeedBackStocks(ctx, &pb.FeedBackStocksReq{
		StockId: v.StockId,
		StockNo: []string{item.StockNo},
		Ts:      item.CreateTime,
	})*/
	log.Infoc(ctx, "StockServerSyncJob FeedBackStocks stock_id:%v , stock_no:%v", stockId, item.StockNo)
	return
}
