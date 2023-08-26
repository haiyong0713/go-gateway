package stock

import (
	"context"
	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/model/stock"
	"strings"
)

// InitGiftStock 初始化奖品库存
func (d *Dao) InitGiftStock(ctx context.Context, stockId int64, limitKey string, stocks int) (ok bool, err error) {
	stockKey := getStockKey(&stock.StockReq{StockBaseReq: stock.StockBaseReq{StockId: stockId, LimitKey: limitKey}})
	log.Infoc(ctx, "InitGiftStock stockKey:%v", stockKey)
	if ok, err = redis.Bool(d.redis.Do(ctx, "SETNX", stockKey, stocks)); err != nil {
		log.Errorc(ctx, "InitGiftStock  stockKey:%v , err :%v", stockKey, err)
	}
	return
}

const stockSpittag = "$"

func (d *Dao) StoreRetryResult(ctx context.Context, stockId int64, retryId string, stockNos []string) (ok bool, err error) {
	stockKey := getRetryRequestKey(&stock.StockBaseReq{StockId: stockId}, retryId)
	log.Infoc(ctx, "CacheRetryResult stockKey:%v", stockKey)
	value := strings.Join(stockNos, stockSpittag)
	if ok, err = redis.Bool(d.redis.Do(ctx, "SET", stockKey, value, "EX", 86400*30, "NX")); err != nil {
		log.Errorc(ctx, "CacheRetryResult  stockKey:%v , err :%v", stockKey, err)
	}
	return
}

func (d *Dao) GetRetryResult(ctx context.Context, stockId int64, retryId string) (stockNos []string, err error) {
	stockKey := getRetryRequestKey(&stock.StockBaseReq{StockId: stockId}, retryId)
	log.Infoc(ctx, "GetRetryResult stockKey:%v", stockKey)
	var data string
	if data, err = redis.String(d.redis.Do(ctx, "GET", stockKey)); err != nil {
		log.Errorc(ctx, "GetRetryResult  stockKey:%v , err :%+v", stockKey, err)
		if err == redis.ErrNil {
			return
		}
	}
	data = strings.Trim(data, stockSpittag)
	stockNos = strings.Split(data, stockSpittag)
	return
}

func (d *Dao) IncrUserStockNum(ctx context.Context, req *stock.UserStockCache, incr int) (res int, err error) {
	redisKey := getUserStockLimitKey(&stock.StockBaseReq{StockId: req.StockId, LimitKey: req.LimitKey}, req.Mid)
	log.Infoc(ctx, "IncrUserStockLimit redisKey:%v , incr:%v", redisKey, incr)
	if res, err = redis.Int(d.redis.Do(ctx, "INCRBY", redisKey, incr)); err != nil {
		log.Errorc(ctx, "IncrUserStockLimit  stockKey:%v , err :%+v", redisKey, err)
	}
	return
}

func (d *Dao) GetUserStockNum(ctx context.Context, req *stock.UserStockCache) (stockNum int, err error) {
	redisKey := getUserStockLimitKey(&stock.StockBaseReq{StockId: req.StockId, LimitKey: req.LimitKey}, req.Mid)
	log.Infoc(ctx, "GetUserStockNum redisKey:%v", redisKey)
	if stockNum, err = redis.Int(d.redis.Do(ctx, "GET", redisKey)); err != nil {
		log.Errorc(ctx, "GetUserStockNum  stockKey:%v , err :%+v", redisKey, err)
	}
	return
}

func (d *Dao) BatchGetUserStockNum(ctx context.Context, req []*stock.UserStockCache) (stockMap map[*stock.UserStockCache]int, err error) {
	stockKeys := redis.Args{}
	for _, v := range req {
		stockKeys = stockKeys.Add(getUserStockLimitKey(&stock.StockBaseReq{StockId: v.StockId, LimitKey: v.LimitKey}, v.Mid))
	}
	var stocks []int
	if stocks, err = redis.Ints(d.redis.Do(ctx, "MGET", stockKeys...)); err != nil {
		log.Errorc(ctx, "BatchGetUserStockNum  stockKeys:%v , err :%+v", stockKeys, err)
	}
	stockMap = make(map[*stock.UserStockCache]int)
	for index, num := range stocks {
		if index > 0 && index < len(req) {
			stockMap[req[index]] = num
		}
	}
	return
}
