package grpc

import (
	"context"
	"encoding/json"
	"go-common/library/log"
	pb "go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/interface/service"
	"net/url"
	"time"
)

// 创建库存记录
func (s *activityService) CreateStockRecord(ctx context.Context, req *pb.CreateStockRecordReq) (*pb.CreateStockRecordResp, error) {
	data, _ := json.Marshal(req)
	log.Infoc(ctx, "CreateStockRecord req:%s", data)
	return service.StockSvr.CreateStockRecord(ctx, req)
}

// 更新库存配置
func (s *activityService) UpdateStockRecord(ctx context.Context, req *pb.CreateStockRecordReq) (*pb.UpdateStockRecordResp, error) {
	data, _ := json.Marshal(req)
	log.Infoc(ctx, "UpdateStockRecord req:%s", data)
	return service.StockSvr.UpdateStockServerConf(ctx, req)
}

func (c *activityService) BatchQueryStockRecord(ctx context.Context, req *pb.GetStocksReq) (replay *pb.BatchStockRecord, err error) {
	replay = new(pb.BatchStockRecord)
	replay.List, err = service.StockSvr.BatchQueryStockRecord(ctx, req.StockIds, req.SkipCache)
	return
}

func (s *activityService) ConsumerStockById(ctx context.Context, req *pb.ConsumerStockReq) (replay *pb.ConsumerStockResp, err error) {
	replay = new(pb.ConsumerStockResp)
	if req.Ts <= 0 {
		req.Ts = time.Now().Unix()
	}
	req.RetryId = url.QueryEscape(req.RetryId)
	if replay.StockNo, err = service.StockSvr.GetRetryResult(ctx, req); err == nil && len(replay.StockNo) > 0 {
		return
	}
	replay.StockNo, err = service.StockSvr.ConsumerStockById(ctx, req)
	return
}

func (s *activityService) GetStocksByIds(ctx context.Context, req *pb.GetStocksReq) (*pb.GetStocksResp, error) {
	return service.StockSvr.GetStocksByIds(ctx, req)
}

func (s *activityService) FeedBackStocks(ctx context.Context, req *pb.FeedBackStocksReq) (replay *pb.FeedBackStocksResp, err error) {
	replay = new(pb.FeedBackStocksResp)
	replay.EffectRows, err = service.StockSvr.FeedBackStock(ctx, req)
	return
}

func (s *activityService) AckStockOrders(ctx context.Context, req *pb.FeedBackStocksReq) (replay *pb.FeedBackStocksResp, err error) {
	replay = new(pb.FeedBackStocksResp)
	replay.EffectRows, err = service.StockSvr.AckStockOrders(ctx, req)
	return
}

func (s *activityService) EffectiveStockList(ctx context.Context, req *pb.EffectiveStockListReq) (replay *pb.EffectiveStockListResp, err error) {
	replay = new(pb.EffectiveStockListResp)
	replay.List, err = service.StockSvr.EffectiveStockList(ctx, req)
	return
}

func (c *activityService) GetStockOrderById(ctx context.Context, req *pb.GetStockOrderByIdReq) (replay *pb.GetStockOrderByIdResp, err error) {
	replay = new(pb.GetStockOrderByIdResp)
	replay.List, err = service.StockSvr.GetStockOrderById(ctx, req)
	return
}
