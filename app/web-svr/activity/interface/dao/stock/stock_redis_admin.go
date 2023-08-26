package stock

import (
	"context"
	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/model/stock"
)

// IncrStock 支持:并发、幂等
func (d *Dao) IncrStock(ctx context.Context, stockId, giftVer int64, incr int) (replay int64, err error) {
	var (
		stockKey = getStockKey(&stock.StockReq{StockBaseReq: stock.StockBaseReq{StockId: stockId}})
	)
	if replay, err = redis.Int64(d.redis.Do(ctx, "EVAL", IncrStockLua, 2, getLockKey(&stock.StockBaseReq{StockId: stockId}, giftVer), stockKey, incr)); err != nil {
		log.Errorc(ctx, "YyincrStock conn.Do(key:%s) error(%v)", stockKey, err)
		return
	}
	log.Infoc(ctx, "YyincrStock begin , stockKey:%v , replay:%v", stockKey, replay)
	return
}

// DecrStock 支持:并发、幂等
func (d *Dao) DecrStock(ctx context.Context, stockId, giftVer int64, decr int) (decrNum int64, err error) {
	var (
		stockKey = getStockKey(&stock.StockReq{StockBaseReq: stock.StockBaseReq{StockId: stockId}})
	)
	log.Infoc(ctx, "YyDecrStock begin , stockKey:%v", stockKey)
	if decrNum, err = redis.Int64(d.redis.Do(ctx, "EVAL", DecrStockLua, 2, getLockKey(&stock.StockBaseReq{StockId: stockId}, giftVer), stockKey, decr)); err != nil {
		log.Errorc(ctx, "YyDecrStock conn.Do(key:%s) error(%v)", stockKey, err)
		return 0, err
	}
	return decrNum, nil
}
