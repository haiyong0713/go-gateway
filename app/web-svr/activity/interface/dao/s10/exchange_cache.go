package s10

import (
	"context"
	"fmt"
	"go-gateway/app/web-svr/activity/interface/component"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/ecode"
	"go-common/library/log"
	xtime "go-common/library/time"
)

const (
	exchangeUser      = "s10:eu:%d"
	exchangeUserRound = "s10:eur:%d:%d:%d"
)

func exchangeUserkey(mid int64) string {
	return fmt.Sprintf(exchangeUser, mid)
}
func exchangeUserRoundkey(mid int64, currtime time.Time) string {
	_, month, day := currtime.Date()
	return fmt.Sprintf(exchangeUserRound, mid, int(month), day)
}

func (d *Dao) AddExchangeStaticCache(ctx context.Context, mid int64, exStatic map[int32]int32) (err error) {
	conn := component.S10PointShopRedis.Get(ctx)
	defer conn.Close()
	eukey := exchangeUserkey(mid)
	for k, v := range exStatic {
		if err = conn.Send("HSET", eukey, k, v); err != nil {
			log.Errorc(ctx, "s10 conn.Send(key:%s) error(%v)", eukey, err)
			return
		}
	}
	if err = conn.Send("EXPIRE", eukey, d.exchangeExpire); err != nil {
		log.Errorc(ctx, "s10 conn.Send(key:%s) error(%v)", eukey, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Errorc(ctx, "s10 conn.Flush(key:%s) error(%v)", eukey, err)
		return
	}
	for i := 0; i < len(exStatic)+1; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Errorc(ctx, "s10 conn.Receive(key:%s) error(%v)", eukey, err)
			return
		}
	}
	return
}

func (d *Dao) DelExchangeStaticCache(ctx context.Context, mid int64) error {
	eukey := exchangeUserkey(mid)
	conn := component.S10PointShopRedis.Get(ctx)
	defer conn.Close()
	_, err := conn.Do("DEL", eukey)
	if err != nil {
		log.Errorc(ctx, "s10 conn.Do(key:%s) error(%v)", eukey, err)
	}
	return err
}

func (d *Dao) ExchangeStaticCache(ctx context.Context, mid int64) (map[int32]int32, error) {
	eukey := exchangeUserkey(mid)
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
	res := make(map[int32]int32, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key, err1 := redis.Int64(values[i], nil)
		value, err2 := redis.Int64(values[i+1], nil)
		if err1 != nil || err2 != nil {
			err = ecode.New(324243)
			return nil, err
		}
		res[int32(key)] = int32(value)
	}
	return res, nil
}

func (d *Dao) ExchangeFieldStaticCache(ctx context.Context, mid int64, gid int32) (int32, bool, error) {
	eukey := exchangeUserkey(mid)
	conn := component.S10PointShopRedis.Get(ctx)
	defer conn.Close()
	res, err := redis.Int64(conn.Do("HGET", eukey, gid))
	if err != nil {
		if err == redis.ErrNil {
			return 0, false, nil
		}
		log.Errorc(ctx, "s10 conn.Do(key:%s,gid:%d) error(%v)", eukey, gid, err)
		return 0, false, err
	}
	return int32(res), true, nil
}

func (d *Dao) AddExchangeRoundStaticCache(ctx context.Context, mid int64, currtime xtime.Time, exStatic map[int32]int32) (err error) {
	conn := component.S10PointShopRedis.Get(ctx)
	defer conn.Close()
	eukey := exchangeUserRoundkey(mid, currtime.Time())
	for k, v := range exStatic {
		if err = conn.Send("HSET", eukey, k, v); err != nil {
			log.Errorc(ctx, "s10 conn.Send(key:%s) error(%v)", eukey, err)
			return
		}
	}
	if err = conn.Send("EXPIRE", eukey, d.roundExchangeExipre); err != nil {
		log.Errorc(ctx, "s10 conn.Send(key:%s) error(%v)", eukey, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Errorc(ctx, "s10 conn.Flush(key:%s) error(%v)", eukey, err)
		return
	}
	for i := 0; i < len(exStatic)+1; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Errorc(ctx, "s10 conn.Receive(key:%s) error(%v)", eukey, err)
			return
		}
	}
	return
}

func (d *Dao) ExchangeRoundStaticCache(ctx context.Context, mid int64, currtime xtime.Time) (map[int32]int32, error) {
	eukey := exchangeUserRoundkey(mid, currtime.Time())
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
	res := make(map[int32]int32, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key, err1 := redis.Int64(values[i], nil)
		value, err2 := redis.Int64(values[i+1], nil)
		if err1 != nil || err2 != nil {
			err = ecode.New(324243)
			return nil, err
		}
		res[int32(key)] = int32(value)
	}
	return res, nil
}

func (d *Dao) ExchangeFieldRoundStaticCache(ctx context.Context, mid int64, gid int32, currtime xtime.Time) (int32, bool, error) {
	eukey := exchangeUserRoundkey(mid, currtime.Time())
	conn := component.S10PointShopRedis.Get(ctx)
	defer conn.Close()
	res, err := redis.Int64(conn.Do("HGET", eukey, gid))
	if err != nil {
		if err == redis.ErrNil {
			return 0, false, nil
		}
		log.Errorc(ctx, "s10 conn.Do(key:%s,gid:%d) error(%v)", eukey, gid, err)
		return 0, false, err
	}
	return int32(res), true, nil
}

func (d *Dao) DelRoundExchangeStaticCache(ctx context.Context, mid int64, currtime xtime.Time) error {
	eukey := exchangeUserRoundkey(mid, currtime.Time())
	conn := component.S10PointShopRedis.Get(ctx)
	defer conn.Close()
	_, err := conn.Do("DEL", eukey)
	if err != nil {
		log.Errorc(ctx, "s10 conn.Do(key:%s) error(%v)", eukey, err)
	}
	return err
}
