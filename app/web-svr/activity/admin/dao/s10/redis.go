package s10

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go-gateway/app/web-svr/activity/admin/component"
	"go-gateway/app/web-svr/activity/admin/model/s10"

	"go-common/library/cache/redis"
	"go-common/library/log"
	xtime "go-common/library/time"
)

const lottery = "s10:lt:%d"

func lotteryKey(mid int64) string {
	return fmt.Sprintf(lottery, mid)
}
func (d *Dao) DelLotteryCache(ctx context.Context, mid int64, robin int32) error {
	conn := component.S10PointShopRedis.Get(ctx)
	defer conn.Close()
	ltkey := lotteryKey(mid)
	_, err := redis.Int64(conn.Do("HDEL", ltkey, robin))
	if err != nil {
		log.Errorc(ctx, "s10 conn.Do() key:%s,error:%v", ltkey, err)
	}
	return err
}

func (d *Dao) AddLotteryFieldCache(ctx context.Context, mid int64, robin int32, gift *s10.MatchUser) error {
	conn := component.S10PointShopRedis.Get(ctx)
	defer conn.Close()
	ltkey := lotteryKey(mid)
	bytes, err := json.Marshal(gift)
	if err != nil {
		log.Errorc(ctx, "s10 json.Marshal() ltkey:%d, error(%v)", ltkey, err)
		return err
	}
	if err = conn.Send("HSET", ltkey, robin, bytes); err != nil {
		log.Errorc(ctx, "s10 conn.Send(HSET ltkey:%d) error(%v)", ltkey, err)
		return err
	}
	err = conn.Send("EXPIRE", ltkey, d.lotteryExpire)
	if err != nil {
		log.Errorc(ctx, "s10 conn.Send(EXPIRE ltkey:%d) error(%v)", ltkey, err)
		return err
	}
	if err = conn.Flush(); err != nil {
		log.Errorc(ctx, "s10 conn.Flush(ltkey:%d) error(%v)", ltkey, err)
		return err
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Errorc(ctx, "s10 conn.Receive(ltkey:%d) error(%v)", ltkey, err)
			return err
		}
	}
	return nil
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
			log.Errorc(ctx, "conn.Send(key:%s) error(%v)", eukey, err)
			return
		}
	}
	if err = conn.Send("EXPIRE", eukey, d.exchangeExpire); err != nil {
		log.Errorc(ctx, "conn.Send(key:%s) error(%v)", eukey, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Errorc(ctx, "conn.Flush(key:%s) error(%v)", eukey, err)
		return
	}
	for i := 0; i < len(exStatic)+1; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Errorc(ctx, "conn.Receive(key:%s) error(%v)", eukey, err)
			return
		}
	}
	return
}

func (d *Dao) DelExchangeStaticCache(ctx context.Context, mid int64) error {
	conn := component.S10PointShopRedis.Get(ctx)
	defer conn.Close()
	eukey := exchangeUserkey(mid)
	_, err := redis.Int64(conn.Do("DEL", eukey))
	if err != nil {
		log.Errorc(ctx, "s10 conn.Do() key:%s,error:%v", eukey, err)
	}
	return err
}

func (d *Dao) DelExchangeRoundStaticCache(ctx context.Context, mid int64, currtime xtime.Time) error {
	conn := component.S10PointShopRedis.Get(ctx)
	defer conn.Close()
	eukey := exchangeUserRoundkey(mid, currtime.Time())
	_, err := redis.Int64(conn.Do("DEL", eukey))
	if err != nil {
		log.Errorc(ctx, "s10 conn.Do() key:%s,error:%v", eukey, err)
	}
	return err
}

func (d *Dao) AddExchangeRoundStaticCache(ctx context.Context, mid int64, currtime xtime.Time, exStatic map[int32]int32) (err error) {
	conn := component.S10PointShopRedis.Get(ctx)
	defer conn.Close()
	eukey := exchangeUserRoundkey(mid, currtime.Time())
	for k, v := range exStatic {
		if err = conn.Send("HSET", eukey, k, v); err != nil {
			log.Errorc(ctx, "conn.Send(key:%s) error(%v)", eukey, err)
			return
		}
	}
	if err = conn.Send("EXPIRE", eukey, d.roundExchangeExpire); err != nil {
		log.Errorc(ctx, "conn.Send(key:%s) error(%v)", eukey, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Errorc(ctx, "conn.Flush(key:%s) error(%v)", eukey, err)
		return
	}
	for i := 0; i < len(exStatic)+1; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Errorc(ctx, "conn.Receive(key:%s) error(%v)", eukey, err)
			return
		}
	}
	return
}

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
		log.Errorc(ctx, "conn.Send(key:%s) error(%v)", eukey, err)
		return
	}

	if err = conn.Send("EXPIRE", eukey, d.restPointExpire); err != nil {
		log.Errorc(ctx, "conn.Send(key:%s) error(%v)", eukey, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Errorc(ctx, "conn.Flush(key:%s) error(%v)", eukey, err)
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Errorc(ctx, "conn.Receive(key:%s) error(%v)", eukey, err)
			return
		}
	}
	return
}

const restCount = "s10:rc:%d"

func restCountKey(gid int32) string {
	return fmt.Sprintf(restCount, gid)
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
