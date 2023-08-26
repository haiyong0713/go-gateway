package lottery

import (
	"context"
	"fmt"
	"strconv"

	"go-common/library/cache/memcache"
	"go-common/library/log"
	"go-common/library/stat/prom"
)

// lotteryMcNumKey
func lotteryMcNumKey(sid int64, high, mc int) string {
	return fmt.Sprintf("lottery_mc_%d_%d_%d", sid, mc, high)
}

// CacheLotteryMcNum get data from mc
func (d *dao) CacheLotteryMcNum(c context.Context, sid int64, high, mc int) (res int64, err error) {
	key := lotteryMcNumKey(sid, high, mc)
	var v string
	err = d.mc.Get(c, key).Scan(&v)
	if err != nil {
		if err == memcache.ErrNotFound {
			err = nil
			return
		}
		prom.BusinessErrCount.Incr("mc:CacheLotteryMcNum")
		log.Errorv(c, log.KV("CacheLotteryMcNum", fmt.Sprintf("%+v", err)), log.KV("key", key))
		return
	}
	r, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		prom.BusinessErrCount.Incr("mc:CacheLotteryMcNum")
		log.Errorv(c, log.KV("CacheLotteryMcNum", fmt.Sprintf("%+v", err)), log.KV("key", key))
		return
	}
	res = r
	return
}

// AddCacheLotteryMcNum Set data to mc
func (d *dao) AddCacheLotteryMcNum(c context.Context, sid int64, high, mc int, val int64) (err error) {
	key := lotteryMcNumKey(sid, high, mc)
	bs := []byte(strconv.FormatInt(val, 10))
	item := &memcache.Item{Key: key, Value: bs, Expiration: d.mcLotteryExpire, Flags: memcache.FlagRAW}
	if err = d.mc.Set(c, item); err != nil {
		prom.BusinessErrCount.Incr("mc:AddCacheLotteryMcNum")
		log.Errorv(c, log.KV("AddCacheLotteryMcNum", fmt.Sprintf("%+v", err)), log.KV("key", key))
		return
	}
	return
}
