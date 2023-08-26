package stock

import (
	"context"
	"encoding/json"
	"github.com/pkg/errors"
	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/model/stock"
)

const (
	stockConfRecordPre = "stock:conf:record:pre"
)

func (d *Dao) CacheStockConfRecord(ctx context.Context, records map[int64]*stock.ConfItemDB) (err error) {
	if len(records) > 0 {
		for stockId, v := range records {
			var data []byte
			if data, err = json.Marshal(v); err != nil {
				return
			}
			if _, err = redis.String(d.redis.Do(ctx, "SETEX", buildKey(stockConfRecordPre, stockId), d.confExpire, data)); err != nil {
				return errors.Wrap(err, "CacheStockConfRecord conn.Do(MSET)")
			}
		}
	}
	return
}

func (d *Dao) GetStockConfRecordFromCache(ctx context.Context, stockIds []int64) (records map[int64]*stock.ConfItemDB, err error) {
	stockKeys := redis.Args{}
	for _, v := range stockIds {
		stockKeys = stockKeys.Add(buildKey(stockConfRecordPre, v))
	}

	log.Infoc(ctx, "GetStockConfRecordFromCache stockKeys:%v", stockKeys)
	if len(stockKeys) > 0 {
		var stockArr [][]byte
		if stockArr, err = redis.ByteSlices(d.redis.Do(ctx, "MGET", stockKeys...)); err != nil {
			log.Errorc(ctx, "GetStockConfRecordFromCache  keys:%v , err :%v", stockKeys, err)
			if err == redis.ErrNil {
				err = nil
			}
			return
		}
		// log.Infoc(ctx, "GetStockConfRecordFromCache mget , value:%v", stockArr)
		records = make(map[int64]*stock.ConfItemDB)
		for k, v := range stockArr {
			if k >= 0 && k < len(stockIds) && v != nil {
				var temp = &stock.ConfItemDB{}
				if err1 := json.Unmarshal(v, temp); err1 == nil {
					records[stockIds[k]] = temp
				}
			}
		}
	}
	return
}
