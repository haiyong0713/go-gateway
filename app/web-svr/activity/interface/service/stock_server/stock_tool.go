package stockserver

import (
	"context"
	"encoding/json"
	"fmt"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/ecode"
	pb "go-gateway/app/web-svr/activity/interface/api"
	daoStock "go-gateway/app/web-svr/activity/interface/dao/stock"
	"go-gateway/app/web-svr/activity/interface/model/stock"
	"time"
)

const (
	_timePeriodTemplate       = "20060102"
	_DailyPeriodTemplate      = "20060102 15:04:05"
	_MonthlyPeriodTemplate    = "200601"
	_DailyPeriodExtraTemplate = "15:04:05"
)

func getLimitCacheKey(cycleType int32, ts int64) (string, error) {
	switch cycleType {
	case int32(pb.StockServerCycleType_DayCycle):
		return time.Unix(ts, 0).Format(_timePeriodTemplate), nil
	case int32(pb.StockServerCycleType_ActCycle):
		return pb.StockServerCycleType_name[cycleType], nil
	}
	return "", ecode.SystemActivityParamsErr
}

func buildConfItemDB(req *pb.CreateStockRecordReq) (*stock.ConfItemDB, error) {
	var (
		NewRuleList []pb.CycleLimitStruct
		RuleList    = make([]pb.CycleLimitStruct, 0)
		err         error
	)
	if err = json.Unmarshal([]byte(req.CycleLimit), &RuleList); err != nil {
		return nil, err
	}
	for _, v := range RuleList {
		// check cycle type
		if _, ok := pb.StockServerCycleType_name[v.CycleType]; !ok ||
			v.CycleType == int32(pb.StockServerCycleType_CycleTypeInvailed) {
			continue
		}
		// check limit type
		if _, ok := pb.StockServerLimitType_name[v.LimitType]; !ok ||
			v.LimitType == int32(pb.StockServerLimitType_LimitTypeInvailed) {
			continue
		}
		NewRuleList = append(NewRuleList, v)
	}
	RuleList = NewRuleList
	if len(RuleList) == 0 {
		return nil, ecode.SystemActivityParamsErr
	}
	var ruleInfo []byte
	if ruleInfo, err = json.Marshal(RuleList); err != nil {
		return nil, err
	}

	res := &stock.ConfItemDB{
		ID:             req.StockId,
		ResourceId:     req.ResourceId,
		ResourceVer:    req.ResourceVer,
		ForeignActId:   req.ForeignActId,
		DescribeInfo:   req.DescInfo,
		RulesInfo:      string(ruleInfo),
		StockStartTime: req.StockStartTime,
		StockEndTime:   req.StockEndTime,
	}
	if newReq, err1 := convert2CreateStockRecordReq(res); err1 == nil {
		if _, err = buildConsumerParams(newReq, time.Now().Unix(), 1); err != nil {
			return nil, err
		}
	}
	return res, nil
}

func buildConsumerParams(confRecord *pb.CreateStockRecordReq, ts int64, num int) (consumerParams *stock.ConsumerStockReq, err error) {
	consumerParams = &stock.ConsumerStockReq{
		StockId:       confRecord.StockId,
		StoreVer:      confRecord.ResourceVer,
		ConsumerStock: num,
	}

	for _, v := range confRecord.CycleLimitObj {
		// 无库存限制
		var storeKey string
		if storeKey, err = getLimitCacheKey(v.CycleType, ts); err != nil {
			return
		}
		consumerParams.CycleStoreKeyPre = storeKey
		if v.LimitType == int32(pb.StockServerLimitType_StoreUpperLimit) && v.Store > 0 {
			if v.CycleType == int32(pb.StockServerCycleType_ActCycle) {
				consumerParams.CycleStore = v.Store
				consumerParams.TotalStore = v.Store
			}
			if v.CycleType == int32(pb.StockServerCycleType_DayCycle) {
				consumerParams.CycleStore = v.Store
			}
		}
		if v.UserNum > 0 {
			consumerParams.UserStore = v.UserNum
		}
	}
	if consumerParams.TotalStore <= 0 && consumerParams.CycleStore <= 0 && consumerParams.UserStore <= 0 {
		err = ecode.StockServerConfInvalidError
	}
	return
}

func (s *Service) getCycleLimitStock(ctx context.Context, confList []*pb.CreateStockRecordReq, mid int64) (stockMap map[int64][]*pb.StocksItem, err error) {
	stockMap = make(map[int64][]*pb.StocksItem)
	// 取每个周期内的库存
	var stockReqs []*daoStock.GetGiftStockReq
	for _, conf := range confList {
		if len(conf.CycleLimitObj) == 0 {
			continue
		}

		for _, item := range conf.CycleLimitObj {
			var limitKey string
			limitKey, err = getLimitCacheKey(item.CycleType, time.Now().Unix())
			if err == nil && limitKey != "" {
				stockReqs = append(stockReqs, &daoStock.GetGiftStockReq{
					StockId:      conf.StockId,
					LimitKey:     limitKey,
					StockType:    item.CycleType,
					LimitNum:     item.Store,
					UserLimitNum: item.UserNum,
					Mid:          mid,
					CycleLimit:   item,
				})
			}
		}
	}

	if len(stockReqs) > 0 {
		var leftStockMap map[*daoStock.GetGiftStockReq]*daoStock.GetGiftStockResp
		if leftStockMap, err = s.dao.BatchGetGiftStock(ctx, stockReqs); err != nil {
			return
		}
		for k, v := range leftStockMap {
			stockMap[k.StockId] = append(stockMap[k.StockId], &pb.StocksItem{
				StockType:     k.StockType,
				LimitNum:      k.LimitNum,
				StockNum:      k.LimitNum - v.LimitStock,
				UserLimitNum:  k.UserLimitNum,
				UserStockNum:  k.UserLimitNum - v.UserLimitStock,
				CycleLimitObj: k.CycleLimit,
			})
		}
	}
	return
}

func (s *Service) checkStoreLimit(ctx context.Context, confRecord *pb.CreateStockRecordReq, expectNum int32, mid int64) (err error) {
	var stockMap map[int64]*pb.StocksItemList
	if stockMap, err = s.GetStocksByConfs(ctx, []*pb.CreateStockRecordReq{confRecord}, mid); err != nil {
		return
	}
	if limitList, ok := stockMap[confRecord.StockId]; ok && limitList != nil {
		for _, v := range limitList.List {
			log.Infoc(ctx, "stock over limit , type:%v , limit:%v , StockNum:%v , UserLimitNum:%v , UserStockNum:%v",
				v.StockType, v.LimitNum, v.StockNum, v.UserLimitNum, v.UserStockNum)

			if v.LimitNum > 0 && v.StockNum < expectNum {
				return ecode.StockServerNoStockInCycleError
			}
			if v.UserLimitNum > 0 && v.UserStockNum < expectNum {
				return ecode.StockServerUserStockUsedUpError
			}
		}
	}
	return
}

func checkCycleTimeLimit(ctx context.Context, confRecord *pb.CreateStockRecordReq, ts int64) (err error) {
	// check stock time
	if confRecord.StockStartTime > 0 && confRecord.StockEndTime > 0 && confRecord.StockStartTime < confRecord.StockEndTime {
		if ts < confRecord.StockStartTime.Time().Unix() || ts > confRecord.StockEndTime.Time().Unix() {
			log.Errorc(ctx, "checkCycleTimeLimit confRecord StockStartTime:%v , StockEndTime:%v , ts:%v", confRecord.StockStartTime, confRecord.StockEndTime, ts)
			return ecode.StockServerInvalidStockTimeError
		}
	}

	// check cycle stock limit time
	for _, v := range confRecord.CycleLimitObj {
		if v.CycleType == int32(pb.StockServerCycleType_DayCycle) {
			period := time.Unix(ts, 0).Format(_timePeriodTemplate)
			if v.CycleStartTime != "" {
				beginTimestamp, _ := time.ParseInLocation(_DailyPeriodTemplate, fmt.Sprintf("%s %s", period, v.CycleStartTime), time.Local)
				if ts < beginTimestamp.Unix() {
					log.Errorc(ctx, "checkCycleTimeLimit CycleLimitObj beginTimestamp:%v , ts:%v", beginTimestamp, ts)
					return ecode.StockServerInvalidStockTimeError
				}
			}

			if v.CycleEndTime != "" {
				endTimestamp, _ := time.ParseInLocation(_DailyPeriodTemplate, fmt.Sprintf("%s %s", period, v.CycleEndTime), time.Local)
				if ts > endTimestamp.Unix() {
					log.Errorc(ctx, "checkCycleTimeLimit CycleLimitObj endTimestamp:%v , ts:%v", endTimestamp, ts)
					return ecode.StockServerInvalidStockTimeError
				}
			}
		}
	}
	return
}

func convert2CreateStockRecordReq(confItem *stock.ConfItemDB) (replay *pb.CreateStockRecordReq, err error) {
	if confItem == nil {
		return
	}
	replay = &pb.CreateStockRecordReq{
		StockId:        confItem.ID,
		ResourceId:     confItem.ResourceId,
		ResourceVer:    confItem.ResourceVer,
		ForeignActId:   confItem.ForeignActId,
		CycleLimit:     confItem.RulesInfo,
		DescInfo:       confItem.DescribeInfo,
		StockStartTime: confItem.StockStartTime,
		StockEndTime:   confItem.StockEndTime,
	}
	if err = json.Unmarshal([]byte(confItem.RulesInfo), &replay.CycleLimitObj); err != nil {
		return
	}
	return
}
