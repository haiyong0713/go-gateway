package stock

import (
	"context"
	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/ecode"
	pb "go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/interface/model/stock"
	"time"
)

func (d *Dao) ConsumerStock(ctx context.Context, param *stock.ConsumerStockReq, retryId string) (uniqIds []string, err error) {
	var sbr = &stock.StockBaseReq{
		StockId:  param.StockId,
		LimitKey: param.CycleStoreKeyPre,
	}
	var (
		stockKey = getStockKey(&stock.StockReq{
			StockBaseReq: stock.StockBaseReq{StockId: param.StockId},
		})
		cycleLimitKey = getStockKey(&stock.StockReq{
			StockBaseReq: *sbr,
		})
		uniqSetKey   = getUniqSetKey(sbr)
		backUpSetKey = getBackUpSetKey(sbr)
		reply        []string
		cnt          int
		num          = param.ConsumerStock
	)
	log.Infoc(ctx, "user consumer stock begin , stockKey: %v, cycleLimitKey: %v,  uniqSetKey: %v, backUpSetKey: %v",
		stockKey, cycleLimitKey, uniqSetKey, backUpSetKey)

	// 1、首先尝试从备份库取数据
	if cnt, err = d.GetBackUpStock(ctx, sbr.StockId, sbr.LimitKey); err == nil && cnt > 0 && num > 0 {
		log.Infoc(ctx, "Consumer backUp stock backUpSetKey:%v SCARD:%v , num:%v", backUpSetKey, cnt, num)
		if reply, err = redis.Strings(d.redis.Do(ctx, "EVAL", ConsumerBackUpStockLua, 2, backUpSetKey, uniqSetKey, num)); err == nil {
			uniqIds = append(uniqIds, reply...)
		}
		num = num - len(uniqIds)
	}
	// 2、再次尝试从正式库取数据
	if num > 0 {
		log.Infoc(ctx, "Consumer real stock stockKey: %v , cycleLimitKey: %v , TotalStore: %v , CycleStore:%v , num:%v", stockKey, cycleLimitKey, param.TotalStore, param.CycleStore, num)
		nowTime := time.Now().Unix()
		if reply, err = redis.Strings(d.redis.Do(ctx, "EVAL", ConsumerStockLua,
			3, stockKey, cycleLimitKey, uniqSetKey, param.TotalStore, param.CycleStore, num, param.StockId, param.StoreVer, retryId, nowTime)); err == nil {
			uniqIds = append(uniqIds, reply...)
		}
	}

	if err == nil && len(uniqIds) <= 0 {
		err = ecode.StockServerNoStockError
	}

	if err != nil {
		log.Errorc(ctx, "user consumer stock err , stock_nos:%v  , err:%v", uniqIds, err)
	}
	return
}

func (d *Dao) UpdateUserStockLimitAndRetryResult(ctx context.Context, param *stock.ConsumerStockReq, stockNoSet []string, mid int64, retryId string) (reply []string, err error) {
	var sbr = &stock.StockBaseReq{
		StockId:  param.StockId,
		LimitKey: param.CycleStoreKeyPre,
	}
	var (
		uniqSetKey    = getUniqSetKey(sbr)
		backUpSetKey  = getBackUpSetKey(sbr)
		userLimitKey  = getUserStockLimitKey(sbr, mid)
		retryStockKey = getRetryRequestKey(&stock.StockBaseReq{StockId: param.StockId}, retryId)
		num           = param.ConsumerStock
	)

	luaArgs := redis.Args{UserStockIncrLua, 4, userLimitKey, backUpSetKey, uniqSetKey, retryStockKey, param.UserStore, num, len(stockNoSet)}
	for _, v := range stockNoSet {
		luaArgs = luaArgs.Add(v)
	}
	if len(stockNoSet) <= 0 {
		luaArgs = luaArgs.Add(param.StockId, param.CycleStoreKeyPre, mid)
	}
	reply, err = redis.Strings(d.redis.Do(ctx, "EVAL", luaArgs...))
	if err == nil && len(reply) <= 0 {
		err = ecode.StockServerConsumerFailedError
	}
	if err != nil {
		log.Errorc(ctx, "Update UserStockLimit And RetryResult , stockNoSet:%v , luaArgs: %v , newStockNoSet:%v , err:%+v", stockNoSet, luaArgs, reply, err)
	}
	return
}

// GetGiftStocks 批量获取奖品库存
func (d *Dao) GetGiftStocks(ctx context.Context, sr []*stock.StockReq) (stocks map[int64]int, err error) {
	stockKeys := redis.Args{}
	for _, v := range sr {
		stockKeys = stockKeys.Add(getStockKey(v))
	}

	log.Infoc(ctx, "GetGiftStocks stockKeys:%v", stockKeys)
	if len(stockKeys) > 0 {
		var stockArr []int
		if stockArr, err = redis.Ints(d.redis.Do(ctx, "MGET", stockKeys...)); err != nil {
			log.Errorc(ctx, "GetGiftStocks  keys:%v , err :%v", stockKeys, err)
			if err == redis.ErrNil {
				err = nil
			}
			return
		}
		log.Infoc(ctx, "GetGiftStocks mget , value:%v", stockArr)
		stocks = make(map[int64]int)
		for k, v := range stockArr {
			if k >= 0 && k < len(sr) {
				stocks[sr[k].StockId] = v
			}
		}
	}
	return
}

// SmoveStockNo 将超时丢失的订单号，重新找回来
func (d *Dao) SmoveStockNo(ctx context.Context, stockId int64, limitKey string, stockNo string) (replay int, err error) {
	var sbr = &stock.StockBaseReq{stockId, limitKey}
	var (
		uniqSetKey   = getUniqSetKey(sbr)
		backUpSetKey = getBackUpSetKey(sbr)
	)
	if replay, err = redis.Int(d.redis.Do(ctx, "SMOVE", uniqSetKey, backUpSetKey, stockNo)); err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
	}
	log.Infoc(ctx, "SmoveStockNo stockNo:%s , source:%s , destination:%s , replay:%d ,  error(%+v)", stockNo, uniqSetKey, backUpSetKey, replay, err)
	return
}

// RandGetStockNo 随机获取已经发放的库存ID
func (d *Dao) RandGetStockNos(ctx context.Context, stockId int64, limitKey string, num int32) (stocks []string, err error) {
	var uniqSetKey = getUniqSetKey(&stock.StockBaseReq{stockId, limitKey})

	if stocks, err = redis.Strings(d.redis.Do(ctx, "SRANDMEMBER", uniqSetKey, num)); err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		log.Errorc(ctx, "RandGetStockNo SRANDMEMBER:%s error(%+v)", uniqSetKey, err)
	}
	return
}

// AckStockNo 确认相应的库存订单号已经成功消费
func (d *Dao) AckStockOrderNos(ctx context.Context, stockId int64, limitKey string, stockNos []string) (replay int, err error) {
	var sbr = &stock.StockBaseReq{stockId, limitKey}
	var (
		uniqSetKey   = getUniqSetKey(sbr)
		backUpSetKey = getBackUpSetKey(sbr)
	)
	ackRedisKeys := []string{uniqSetKey, backUpSetKey}
	for _, rediskey := range ackRedisKeys {
		args := redis.Args{}
		args = append(args, rediskey)
		for _, v := range stockNos {
			args = append(args, v)
		}
		var replay2 int
		if replay2, err = redis.Int(d.redis.Do(ctx, "SREM", args...)); err != nil {
			return
		}
		replay += replay2
	}
	return
}

// GetBackUpStock 获取备份库存数量
func (d *Dao) GetBackUpStock(ctx context.Context, stockId int64, limitKey string) (stockNum int, err error) {
	var backUpSetKey = getBackUpSetKey(&stock.StockBaseReq{stockId, limitKey})
	if stockNum, err = redis.Int(d.redis.Do(ctx, "SCARD", backUpSetKey)); err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
	}
	return
}

func (d *Dao) GetGiftStock(ctx context.Context, sr *stock.StockReq) (stockNum int, err error) {
	return redis.Int(d.redis.Do(ctx, "GET", getStockKey(sr)))
}

type GetGiftStockReq struct {
	StockId      int64
	LimitKey     string
	StockType    int32
	LimitNum     int32
	UserLimitNum int32
	Mid          int64
	CycleLimit   *pb.CycleLimitStruct
}

type GetGiftStockResp struct {
	LimitStock     int32
	UserLimitStock int32
}

func (d *Dao) BatchGetGiftStock(ctx context.Context, req []*GetGiftStockReq) (stockNumMap map[*GetGiftStockReq]*GetGiftStockResp, err error) {
	stockKeys := redis.Args{}
	for _, v := range req {
		// 用户-周期级别的库存使用量
		userLimitStockKey := getUserStockLimitKey(&stock.StockBaseReq{StockId: v.StockId, LimitKey: v.LimitKey}, v.Mid)
		// 周期级别的库存使用量
		limitStockKey := getStockKey(&stock.StockReq{
			StockBaseReq: stock.StockBaseReq{
				StockId:  v.StockId,
				LimitKey: v.LimitKey,
			},
		})
		stockKeys = stockKeys.Add(limitStockKey, userLimitStockKey)
	}
	var stocks []int
	if stocks, err = redis.Ints(d.redis.Do(ctx, "MGET", stockKeys...)); err != nil {
		log.Errorc(ctx, "BatchGetUserStockNum  stockKeys:%v , err :%+v", stockKeys, err)
	}
	log.Infoc(ctx, "BatchGetGiftStock , BatchGetUserStockNum , stockKeys:%v , stocks:%v", stockKeys, stocks)
	stockNumMap = make(map[*GetGiftStockReq]*GetGiftStockResp)
	for index, stockReq := range req {
		if index*2+1 < len(stocks) {
			stockNumMap[stockReq] = &GetGiftStockResp{
				LimitStock:     int32(stocks[index*2]),
				UserLimitStock: int32(stocks[index*2+1]),
			}
		}
	}
	return
}
