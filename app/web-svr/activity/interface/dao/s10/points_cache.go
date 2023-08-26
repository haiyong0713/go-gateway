package s10

import (
	"context"
	"fmt"
	"go-gateway/app/web-svr/activity/interface/component"

	"go-common/library/cache/redis"
	"go-common/library/ecode"
	"go-common/library/log"
)

const userPoint = "s10:up:%d"

func userPointKey(mid int64) string {
	return fmt.Sprintf(userPoint, mid)
}

// typ:0-total;1-cost
func (d *Dao) AddPointsCache(ctx context.Context, mid int64, typ, total int32) (err error) {
	conn := component.S10PointShopRedis.Get(ctx)
	defer conn.Close()
	eukey := userPointKey(mid)
	if err = conn.Send("HSET", eukey, typ, total); err != nil {
		log.Errorc(ctx, "s10 conn.Send(key:%s) error(%v)", eukey, err)
		return
	}

	if err = conn.Send("EXPIRE", eukey, d.restPointExpire); err != nil {
		log.Errorc(ctx, "s10 conn.Send(key:%s) error(%v)", eukey, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Errorc(ctx, "s10 conn.Flush(key:%s) error(%v)", eukey, err)
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Errorc(ctx, "s10 conn.Receive(key:%s) error(%v)", eukey, err)
			return
		}
	}
	return
}

// typ:0-total;1-cost
func (d *Dao) AddAllPointsCache(ctx context.Context, mid int64, res ...int32) (err error) {
	conn := component.S10PointShopRedis.Get(ctx)
	defer conn.Close()
	eukey := userPointKey(mid)
	for i, v := range res {
		if err = conn.Send("HSET", eukey, i, v); err != nil {
			log.Errorc(ctx, "s10 conn.Send(key:%s) error(%v)", eukey, err)
			return
		}
	}
	if err = conn.Send("EXPIRE", eukey, d.restPointExpire); err != nil {
		log.Errorc(ctx, "s10 conn.Send(key:%s) error(%v)", eukey, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Errorc(ctx, "s10 conn.Flush(key:%s) error(%v)", eukey, err)
		return
	}
	for i := 0; i < len(res)+1; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Errorc(ctx, "s10 conn.Receive(key:%s) error(%v)", eukey, err)
			return
		}
	}
	return
}

func (d *Dao) PointsCache(ctx context.Context, mid int64) (map[int32]int32, error) {
	eukey := userPointKey(mid)
	conn := component.S10PointShopRedis.Get(ctx)
	defer conn.Close()
	values, err := redis.Values(conn.Do("HGETALL", eukey))
	if err != nil {
		if err == redis.ErrNil {
			return nil, nil
		}
		log.Errorc(ctx, "s10 conn.Do(HGETALL key:%d) error(%v)", eukey, err)
		return nil, err
	}
	if len(values) == 0 {
		return nil, nil
	}
	if len(values)%2 != 0 {
		return nil, ecode.New(123)
	}
	points := make(map[int32]int32, 2)
	for i := 0; i < len(values); i += 2 {
		key, err1 := redis.Int64(values[i], nil)
		value, err2 := redis.Int64(values[i+1], nil)
		if err1 != nil || err2 != nil {
			return nil, ecode.New(324243)
		}
		points[int32(key)] = int32(value)
	}
	return points, nil
}

func (d *Dao) DelPointsFieldCache(ctx context.Context, mid int64, typ int32) (err error) {
	conn := component.S10PointShopRedis.Get(ctx)
	defer conn.Close()
	if _, err = conn.Do("HDEL", userPointKey(mid), typ); err != nil {
		log.Error("DelRestPointCache conn.Do(DEL, %d) error(%v)", mid, err)
	}
	return
}
