package stockserver

import (
	"context"
	"github.com/pkg/errors"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/ecode"
	pb "go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/interface/model/stock"
	"go-gateway/app/web-svr/activity/interface/tool"
	"strings"
)

func (s *Service) CreateStockRecord(ctx context.Context, req *pb.CreateStockRecordReq) (replay *pb.CreateStockRecordResp, err error) {
	replay = new(pb.CreateStockRecordResp)
	var confDB *stock.ConfItemDB
	if confDB, err = buildConfItemDB(req); err != nil {
		return
	}
	replay.StockId, err = s.dao.AddStockServerConf(ctx, confDB)
	if err != nil && strings.Contains(err.Error(), "Duplicate entry") {
		var listConf []*stock.ConfItemDB
		if listConf, err = s.dao.RawGetConfListByRid(ctx, req.ResourceId, req.ForeignActId); err != nil {
			return
		}
		if len(listConf) <= 0 {
			err = ecode.ActivityNotExist
			return
		}
		replay.StockId = listConf[0].ID
	}
	return
}

func (s *Service) UpdateStockServerConf(ctx context.Context, req *pb.CreateStockRecordReq) (replay *pb.UpdateStockRecordResp, err error) {
	replay = new(pb.UpdateStockRecordResp)
	if req.StockId <= 0 {
		err = ecode.SystemActivityParamsErr
		return
	}
	var confDB *stock.ConfItemDB
	if confDB, err = buildConfItemDB(req); err != nil {
		return
	}
	if replay.EffectRows, err = s.dao.UpdateStockServerConf(ctx, confDB); err == nil && replay.EffectRows > 0 {
		// 后台更新成功，立即更新缓存
		s.dao.CacheStockConfRecord(ctx, map[int64]*stock.ConfItemDB{
			confDB.ID: confDB,
		})
	}
	return
}

func (s *Service) QueryStockRecord(ctx context.Context, stockId int64, skipCache bool) (replay *pb.CreateStockRecordReq, err error) {
	replay = new(pb.CreateStockRecordReq)
	var confList []*stock.ConfItemDB
	if confList, err = s.GetConfListByIDs(ctx, []int64{stockId}, skipCache); err != nil {
		return
	}
	if len(confList) > 0 {
		replay, err = convert2CreateStockRecordReq(confList[0])
	}
	return
}

func (s *Service) BatchQueryStockRecord(ctx context.Context, stockIds []int64, skipCache bool) (replay []*pb.CreateStockRecordReq, err error) {
	var confs []*stock.ConfItemDB
	if confs, err = s.GetConfListByIDs(ctx, stockIds, skipCache); err != nil {
		return
	}
	for _, v := range confs {
		var tmp *pb.CreateStockRecordReq
		if tmp, err = convert2CreateStockRecordReq(v); err != nil {
			return
		}
		replay = append(replay, tmp)
	}
	return
}

func (s *Service) GetStocksByIds(ctx context.Context, req *pb.GetStocksReq) (replay *pb.GetStocksResp, err error) {
	replay = new(pb.GetStocksResp)
	var confList []*pb.CreateStockRecordReq
	if confList, err = s.BatchQueryStockRecord(ctx, req.StockIds, false); err != nil {
		return
	}
	replay.StockMap, err = s.GetStocksByConfs(ctx, confList, req.Mid)
	return
}

func (s *Service) GetStocksByConfs(ctx context.Context, confList []*pb.CreateStockRecordReq, mid int64) (stockMap map[int64]*pb.StocksItemList, err error) {
	var cycleLimitStock map[int64][]*pb.StocksItem
	if cycleLimitStock, err = s.getCycleLimitStock(ctx, confList, mid); err != nil {
		return
	}
	stockMap = map[int64]*pb.StocksItemList{}
	for k, v := range cycleLimitStock {
		if sl, ok := stockMap[k]; ok && sl != nil {
			sl.List = append(sl.List, v...)
			continue
		}
		stockMap[k] = &pb.StocksItemList{
			List: v,
		}
	}
	return
}

// ConsumerSingleStockById 扣减一件库存
func (s *Service) ConsumerSingleStockById(ctx context.Context, req *pb.ConsumerSingleStockReq) (stockNo string, err error) {
	var stockNoSet []string
	if stockNoSet, err = s.ConsumerStockById(ctx, &pb.ConsumerStockReq{
		StockId: req.StockId,
		RetryId: req.RetryId,
		Num:     1,
		Ts:      req.Ts,
		Mid:     req.Mid,
	}); err != nil {
		return
	}
	if len(stockNoSet) <= 0 {
		err = ecode.StockServerConsumerFailedError
		return
	}
	return stockNoSet[0], err
}

func (s *Service) GetRetryResult(ctx context.Context, req *pb.ConsumerStockReq) (stockNoSet []string, err error) {
	return s.dao.GetRetryResult(ctx, req.StockId, req.RetryId)
}

// ComsumerStockById 通过stock_id进行库存扣减 ， 返回对应数量(req.Num)的库存号码(理论上不重复)
func (s *Service) ConsumerStockById(ctx context.Context, req *pb.ConsumerStockReq) (stockNoSet []string, err error) {
	var (
		confRecord     *pb.CreateStockRecordReq
		consumerParams *stock.ConsumerStockReq
	)
	if confRecord, err = s.QueryStockRecord(ctx, req.StockId, false); err != nil || confRecord == nil {
		log.Errorc(ctx, "ConsumerStockById QueryStockRecord stock_id:%v , err:%+v", req.StockId, err)
		return nil, errors.Wrapf(err, "can not find conf of stock_id:%v", req.StockId)
	}
	// 1、检查库存领取时间限制
	if err = checkCycleTimeLimit(ctx, confRecord, req.Ts); err != nil {
		log.Errorc(ctx, "ConsumerStockById checkCycleTimeLimit stock_id:%d , err:%+v", req.StockId, err)
		return
	}
	// 2、检查库存配置参数
	if consumerParams, err = buildConsumerParams(confRecord, req.Ts, int(req.Num)); err != nil {
		log.Errorc(ctx, "ConsumerStockById buildConsumerParams stock_id:%d , err:%+v", req.StockId, err)
		return
	}
	// 3、检查当前库存情况
	if err = s.checkStoreLimit(ctx, confRecord, req.Num, req.Mid); err != nil {
		log.Errorc(ctx, "ConsumerStockById checkStoreLimit stock_id:%d , err:%+v", req.StockId, err)
		return
	}

	if consumerParams.TotalStore > 0 || consumerParams.CycleStore > 0 {
		if stockNoSet, err = s.dao.ConsumerStock(ctx, consumerParams, req.RetryId); err != nil {
			return
		}
	}

	if consumerParams.UserStore > 0 {
		// 用户的领取记录
		return s.dao.UpdateUserStockLimitAndRetryResult(ctx, consumerParams, stockNoSet, req.Mid, req.RetryId)
	}

	//  请求快照
	ok, err1 := s.dao.StoreRetryResult(ctx, req.StockId, req.RetryId, stockNoSet)
	log.Infoc(ctx, "StoreRetryResult , ok:%v , err:%+v", ok, err1)
	return
}

func (s *Service) GetConfListByIDs(ctx context.Context, stockIds []int64, skipCache bool) (confList []*stock.ConfItemDB, err error) {
	var newStockIds []int64

	if !skipCache {
		var recordMap map[int64]*stock.ConfItemDB
		if recordMap, err = s.dao.GetStockConfRecordFromCache(ctx, stockIds); err == nil && len(recordMap) > 0 {
			newStockIds = make([]int64, 0, len(recordMap))
			for k, v := range recordMap {
				if v != nil && v.ID > 0 {
					newStockIds = append(newStockIds, k)
					confList = append(confList, v)
				}
			}
		}
	}

	var missIds []int64
	for _, v := range stockIds {
		if !tool.InInt64Slice(v, newStockIds) {
			missIds = append(missIds, v)
		}
	}
	if len(missIds) <= 0 {
		return
	}

	var (
		dbRecords []*stock.ConfItemDB
		confMap   map[int64]*stock.ConfItemDB
	)
	if dbRecords, err = s.dao.RawGetConfListByIDs(ctx, missIds); err != nil {
		return
	}
	if len(dbRecords) > 0 {
		confMap = make(map[int64]*stock.ConfItemDB)
		for _, confItem := range dbRecords {
			confList = append(confList, confItem)
			confMap[confItem.ID] = confItem
		}
	}

	s.cache.SyncDo(ctx, func(ctx context.Context) {
		if skipCache || len(confMap) <= 0 {
			return
		}
		err1 := s.dao.CacheStockConfRecord(ctx, confMap)
		log.Infoc(ctx, "GetConfListByIDs CacheStockConfRecord err:%v", err1)
	})

	return
}
