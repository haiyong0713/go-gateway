package stockserver

import (
	"context"
	"go-common/library/log"
	pb "go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/interface/model/stock"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func (s *Service) getLimitKeyByStockId(ctx context.Context, stockId int64, ts int64) (limitKey string, err error) {
	var stockRecord *pb.CreateStockRecordReq
	// 跳开缓存，直接查库（FeedBackStock失败，要求主动回退库存的case应该不多）
	if stockRecord, err = s.QueryStockRecord(ctx, stockId, false); err != nil || stockRecord == nil {
		return
	}

	if ts <= 0 {
		ts = time.Now().Unix()
	}

	for _, v := range stockRecord.CycleLimitObj {
		if limitKey == "" {
			limitKey, _ = getLimitCacheKey(v.CycleType, ts)
			break
		}
	}
	return
}

// FeedBackStock 下游业务处理异常，要求主动回退库存
func (s *Service) FeedBackStock(ctx context.Context, req *pb.FeedBackStocksReq) (effectsRows int32, err error) {
	var limitKey string
	if limitKey, err = s.getLimitKeyByStockId(ctx, req.StockId, req.Ts); err != nil {
		return
	}

	for _, orderno := range req.StockNos {
		var rows int
		rows, err = s.dao.SmoveStockNo(ctx, req.StockId, limitKey, orderno)
		if err != nil {
			return
		}
		effectsRows += int32(rows)
	}
	return
}

func (s *Service) AckStockOrders(ctx context.Context, req *pb.FeedBackStocksReq) (effectsRows int32, err error) {
	var limitKey string
	if limitKey, err = s.getLimitKeyByStockId(ctx, req.StockId, req.Ts); err != nil {
		log.Warnc(ctx, "getLimitKeyByStockId StockId: %v , err:%+v", req.StockId, err)
		return
	}
	var replay int
	replay, err = s.dao.AckStockOrderNos(ctx, req.StockId, limitKey, req.StockNos)
	effectsRows = int32(replay)
	if err != nil || effectsRows <= 0 {
		log.Warnc(ctx, "AckStockOrders StockId: %v , orders:%v , replay:%v , %+v", req.StockId, req.StockNos, replay, err)
	}
	return
}

func (s *Service) EffectiveStockList(ctx context.Context, req *pb.EffectiveStockListReq) (list []*pb.CreateStockRecordReq, err error) {
	var offset int32
	offset = (req.PageNumber - 1) * req.PageSize
	var confs []*stock.ConfItemDB
	if confs, err = s.dao.RawGetConfListByTime(ctx, req.BeginTime, req.EndTime, offset, req.PageSize); err != nil {
		return
	}
	for _, v := range confs {
		var tmp *pb.CreateStockRecordReq
		if tmp, err = convert2CreateStockRecordReq(v); err != nil {
			return
		}
		list = append(list, tmp)
	}
	return
}

func (s *Service) GetStockOrderById(ctx context.Context, req *pb.GetStockOrderByIdReq) (replay []*pb.GetStockOrderByIdItem, err error) {
	var limitKey string
	if limitKey, err = s.getLimitKeyByStockId(ctx, req.StockId, 0); err != nil {
		return
	}
	var stockOrders []string
	stockOrders, err = s.dao.RandGetStockNos(ctx, req.StockId, limitKey, req.SyncNum)
	for _, v := range stockOrders {
		var item *pb.GetStockOrderByIdItem
		if item, err = buildGetStockOrderByIdItem(v); err != nil {
			log.Errorc(ctx, "GetStockOrderById buildGetStockOrderByIdItem err:%+v", err)
		}
		replay = append(replay, item)
	}
	return
}

func buildGetStockOrderByIdItem(stockOrder string) (item *pb.GetStockOrderByIdItem, err error) {
	item = &pb.GetStockOrderByIdItem{
		StockNo: stockOrder,
	}
	orderNoSlice := strings.Split(stockOrder, ":")
	if len(orderNoSlice) >= 5 {
		if item.UniqueId, err = url.QueryUnescape(orderNoSlice[0]); err != nil {
			return
		}
		if item.CreateTime, err = strconv.ParseInt(orderNoSlice[1], 10, 64); err != nil {
			return
		}
	}
	return
}

/*
库存模块：对账，同步功能
*/
