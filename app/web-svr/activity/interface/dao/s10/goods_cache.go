package s10

import (
	"context"
	"fmt"
	xtime "go-common/library/time"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/component"
	"math"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/log"
)

const (
	restCount      = "s10:rc:%d"
	roundRestCount = "s10:rrc:%d:%d:%d"
	INCRLua        = `
	if redis.call("EXISTS",KEYS[1]) == 1 then
     	return redis.call("INCR",KEYS[1])
    else
     	return nil
    end
	`
	DECRLua = `
	if redis.call("EXISTS",KEYS[1]) == 1 then
     	return redis.call("DECR",KEYS[1])
    else
     	return nil
    end
	`
)

func restCountKey(gid int32) string {
	return fmt.Sprintf(restCount, gid)
}

func roundRestCountKey(gid int32, currTime time.Time) string {
	_, month, day := currTime.Date()
	return fmt.Sprintf(roundRestCount, gid, int64(month), day)
}

func (d *Dao) AddRestCountByGoodsCache(ctx context.Context, gid int32, rest int64) error {
	conn := component.S10PointShopRedis.Get(ctx)
	defer conn.Close()
	rckey := restCountKey(gid)
	reply, err := conn.Do("SET", rckey, math.MaxInt64-rest, "EX", d.restCountGoodsExpire, "NX")
	if err != nil {
		log.Errorc(ctx, "s10 conn.Do(key:%s) error(%v)", rckey, err)
		return err
	}
	if reply == nil {
		return ecode.ActivityKeyExists
	}
	return nil
}

func (d *Dao) DelRestCountByGoodsCache(ctx context.Context, gid int32) error {
	conn := component.S10PointShopRedis.Get(ctx)
	defer conn.Close()
	rckey := restCountKey(gid)
	_, err := conn.Do("DEL", rckey)
	if err != nil {
		log.Errorc(ctx, "s10 conn.Do(key:%s) error(%v)", rckey, err)
	}
	return err
}

// return: total,cost,err
func (d *Dao) RestCountByGoodsCache(ctx context.Context, gid int32) (int64, bool, error) {
	rckey := restCountKey(gid)
	conn := component.S10PointShopRedis.Get(ctx)
	defer conn.Close()
	res, err := redis.Int64(conn.Do("GET", rckey))
	if err != nil {
		if err == redis.ErrNil {
			return 0, false, nil
		}
		log.Errorc(ctx, "s10 d.dao.RestCountByGoodsCache(key:%s) error(%v)", rckey, err)
	}
	return math.MaxInt64 - res, true, nil
}

func (d *Dao) IncrRestCountByGoodsCache(ctx context.Context, gid int32) (exist bool, err error) {
	conn := component.S10PointShopRedis.Get(ctx)
	defer conn.Close()
	rckey := restCountKey(gid)
	if _, err = redis.Int64(conn.Do("EVAL", INCRLua, 1, rckey)); err != nil {
		if err == redis.ErrNil {
			return false, nil
		}
		log.Errorc(ctx, "s10 conn.Do(key:%s) error(%v)", rckey, err)
		return false, err
	}
	return true, nil
}

func (d *Dao) DecrRestCountByGoodsCache(ctx context.Context, gid int32) (exist bool, err error) {
	conn := component.S10PointShopRedis.Get(ctx)
	defer conn.Close()
	rckey := restCountKey(gid)
	if _, err = redis.Int64(conn.Do("EVAL", DECRLua, 1, rckey)); err != nil {
		if err == redis.ErrNil {
			return false, nil
		}
		log.Errorc(ctx, "s10 conn.Do(key:%s) error(%v)", rckey, err)
		return false, err
	}
	return true, nil
}

func (d *Dao) AddRoundRestCountByGoodsCache(ctx context.Context, gid int32, rest int64, currTime xtime.Time) (err error) {
	conn := component.S10PointShopRedis.Get(ctx)
	defer conn.Close()
	rckey := roundRestCountKey(gid, currTime.Time())
	reply, err := conn.Do("SET", rckey, math.MaxInt64-rest, "EX", d.roundRestCountGoodsExpire, "NX")
	if err != nil {
		log.Errorc(ctx, "s10 conn.Do(key:%s) error(%v)", rckey, err)
		return err
	}
	if reply == nil {
		return ecode.ActivityKeyExists
	}
	return nil
}

// return: total,cost,err
func (d *Dao) RoundRestCountByGoodsCache(ctx context.Context, gid int32) (int64, bool, error) {
	currTime := time.Now()
	rckey := roundRestCountKey(gid, currTime)
	conn := component.S10PointShopRedis.Get(ctx)
	defer conn.Close()
	res, err := redis.Int64(conn.Do("GET", rckey))
	if err != nil {
		if err == redis.ErrNil {
			return 0, false, nil
		}
		log.Errorc(ctx, "s10 d.dao.RestCountByGoodsCache(key:%s) error(%v)", rckey, err)
		return 0, false, err
	}
	return math.MaxInt64 - res, true, nil
}

func (d *Dao) DelRoundRestCountByGoodsCache(ctx context.Context, gid int32, currDate xtime.Time) error {
	conn := component.S10PointShopRedis.Get(ctx)
	defer conn.Close()
	rckey := roundRestCountKey(gid, currDate.Time())
	_, err := conn.Do("DEL", rckey)
	if err != nil {
		log.Errorc(ctx, "s10 conn.Do(key:%s) error(%v)", rckey, err)
	}
	return err
}

func (d *Dao) IncrRoundRestCountByGoodsCache(ctx context.Context, gid int32, currTime xtime.Time) (exist bool, err error) {
	conn := component.S10PointShopRedis.Get(ctx)
	defer conn.Close()
	rckey := roundRestCountKey(gid, currTime.Time())
	if _, err = redis.Int64(conn.Do("EVAL", INCRLua, 1, rckey)); err != nil {
		if err == redis.ErrNil {
			return false, nil
		}
		log.Errorc(ctx, "s10 conn.Do(key:%s) error(%v)", rckey, err)
		return false, err
	}
	return true, nil
}

func (d *Dao) DecrRoundRestCountByGoodsCache(ctx context.Context, gid int32, currTime xtime.Time) (exist bool, err error) {
	conn := component.S10PointShopRedis.Get(ctx)
	defer conn.Close()
	rckey := roundRestCountKey(gid, currTime.Time())
	if _, err = redis.Int64(conn.Do("EVAL", DECRLua, 1, rckey)); err != nil {
		if err == redis.ErrNil {
			return false, nil
		}
		log.Errorc(ctx, "s10 conn.Do(key:%s) error(%v)", rckey, err)
		return false, err
	}
	return true, nil
}
