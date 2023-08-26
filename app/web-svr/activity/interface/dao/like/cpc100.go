package like

import (
	"context"
	"fmt"
	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/component"
)

const cpc100PVCacheKey = "cpc100_pv"
const cpc100TVCacheKey = "cpc100_tv"

func (d *Dao) Cpc100Unlock(c context.Context, mid int64, key string) (err error) {
	rdsKey := fmt.Sprintf("cpc100_unlock_%d", mid)
	_, err = component.GlobalRedis.Do(c, "HSET", rdsKey, key, 1)
	if err != nil {
		log.Errorc(c, "Cpc100Unlock HSET %s %s err[%+v]", rdsKey, key, err)
	}
	component.GlobalRedis.Do(c, "EXPIRE", rdsKey, 5184000) //60天过期
	return
}

func (d *Dao) Cpc100UnlockInfo(c context.Context, mid int64) (info map[string]int, err error) {
	rdsKey := fmt.Sprintf("cpc100_unlock_%d", mid)
	info, err = redis.IntMap(component.GlobalRedis.Do(c, "HGETALL", rdsKey))
	if err != nil {
		log.Errorc(c, "Cpc100UnlockInfo HGETALL %s err[%+v]", rdsKey, err)
	}
	return
}

func (d *Dao) CpcGetPV(ctx context.Context) (pv int64, err error) {
	conn := d.redis.Get(ctx)
	defer func() {
		_ = conn.Close()
	}()
	pv, err = redis.Int64(conn.Do("GET", cpc100PVCacheKey))
	if err == redis.ErrNil {
		err = nil
	}
	return
}

func (d *Dao) CpcGetTopicView(ctx context.Context) (topicView int64, err error) {
	conn := d.redis.Get(ctx)
	defer func() {
		_ = conn.Close()
	}()
	topicView, err = redis.Int64(conn.Do("GET", cpc100TVCacheKey))
	if err == redis.ErrNil {
		err = nil
	}
	return
}
