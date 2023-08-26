package s10

import (
	"context"
	"encoding/json"
	"fmt"
	"go-common/library/cache/redis"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/component"
	"go-gateway/app/web-svr/activity/interface/model/s10"
)

const lottery = "s10:lt:%d"

func lotteryKey(mid int64) string {
	return fmt.Sprintf(lottery, mid)
}

func (d *Dao) DelLotteryFieldCache(ctx context.Context, mid int64, robin int32) (err error) {
	conn := component.S10PointShopRedis.Get(ctx)
	defer conn.Close()
	ltkey := lotteryKey(mid)
	if _, err = conn.Do("HDEL", ltkey, robin); err != nil {
		log.Errorc(ctx, "s10 conn.Do() key:%s error(%v)", ltkey, err)
	}
	return
}

func (d *Dao) AddLotteryCache(ctx context.Context, mid int64, robin map[int32]*s10.MatchUser) error {
	conn := component.S10PointShopRedis.Get(ctx)
	defer conn.Close()
	ltkey := lotteryKey(mid)
	for index, v := range robin {
		bytes, err := json.Marshal(v)
		if err != nil {
			log.Errorc(ctx, "s10 json.Marshal() ltkey:%d, error(%v)", ltkey, err)
			return err
		}
		if err = conn.Send("HSET", ltkey, index, bytes); err != nil {
			log.Errorc(ctx, "s10 conn.Send(HSET ltkey:%d) error(%v)", ltkey, err)
			return err
		}
	}

	err := conn.Send("EXPIRE", ltkey, d.lotteryExpire)
	if err != nil {
		log.Errorc(ctx, "s10 conn.Send(EXPIRE ltkey:%d) error(%v)", ltkey, err)
		return err
	}
	if err = conn.Flush(); err != nil {
		log.Errorc(ctx, "s10 conn.Flush(ltkey:%d) error(%v)", ltkey, err)
		return err
	}
	for i := 0; i < len(robin)+1; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Errorc(ctx, "s10 conn.Receive(ltkey:%d) error(%v)", ltkey, err)
			return err
		}
	}
	return nil
}

func (d *Dao) LotteryFieldCache(ctx context.Context, mid int64, robin int32) (*s10.MatchUser, error) {
	ltkey := lotteryKey(mid)
	conn := component.S10PointShopRedis.Get(ctx)
	defer conn.Close()
	value, err := redis.Bytes(conn.Do("HGET", ltkey, robin))
	if err != nil {
		if err == redis.ErrNil {
			return nil, nil
		}
		log.Errorc(ctx, "s10 conn.Do(key:%s) error(%v)", ltkey, err)
		return nil, err
	}
	var res *s10.MatchUser
	if err = json.Unmarshal(value, &res); err != nil {
		log.Errorc(ctx, "s10 json.Unmarshal() key:%s error(%v)", ltkey, err)
	}
	return res, err
}

func (d *Dao) LotteryCache(ctx context.Context, mid int64) (map[int32]*s10.MatchUser, error) {
	ltkey := lotteryKey(mid)
	conn := component.S10PointShopRedis.Get(ctx)
	defer conn.Close()
	values, err := redis.Values(conn.Do("HGETALL", ltkey))
	if err != nil {
		if err == redis.ErrNil {
			return nil, nil
		}
		log.Errorc(ctx, "s10 conn.Do(key:%s) error(%v)", ltkey, err)
		return nil, err
	}
	if len(values) == 0 {
		return nil, nil
	}
	if len(values)%2 != 0 {
		return nil, ecode.New(123)
	}
	res := make(map[int32]*s10.MatchUser, 5)
	for i := 0; i < len(values); i += 2 {
		key, err1 := redis.Int64(values[i], nil)
		value, err2 := redis.Bytes(values[i+1], nil)
		if err1 != nil || err2 != nil {
			err = ecode.New(324243)
			return nil, err
		}
		var tmp *s10.MatchUser
		if err = json.Unmarshal(value, &tmp); err != nil {
			log.Errorc(ctx, "s10 json.Unmarshal() key:%s error(%v)", ltkey, err)
			fmt.Println(err)
			return nil, err
		}
		res[int32(key)] = tmp
	}
	return res, nil
}
