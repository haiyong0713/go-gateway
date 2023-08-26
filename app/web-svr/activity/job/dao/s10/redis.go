package s10

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go-gateway/app/web-svr/activity/job/dao"
	"go-gateway/app/web-svr/activity/job/model/s10"

	"go-common/library/log"
	xtime "go-common/library/time"
)

const lottery = "s10:lt:%d"

func lotteryKey(mid int64) string {
	return fmt.Sprintf(lottery, mid)
}
func (d *Dao) AddLotteryCache(ctx context.Context, mid int64, robin int32, gift *s10.MatchUser) error {
	conn := dao.S10PointShopRedis.Get(ctx)
	defer conn.Close()
	ltkey := lotteryKey(mid)
	bytes, err := json.Marshal(gift)
	if err != nil {
		log.Errorc(ctx, "json.Marshal() ltkey:%d, error(%v)", ltkey, err)
		return err
	}

	if err = conn.Send("HSET", ltkey, robin, bytes); err != nil {
		log.Errorc(ctx, "conn.Send(HSET ltkey:%d) error(%v)", ltkey, err)
		return err
	}

	if err = conn.Send("EXPIRE", ltkey, d.lotteryExpire); err != nil {
		log.Errorc(ctx, "conn.Send(EXPIRE ltkey:%d) error(%v)", ltkey, err)
		return err
	}
	if err = conn.Flush(); err != nil {
		log.Errorc(ctx, "conn.Flush(ltkey:%d) error(%v)", ltkey, err)
		return err
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Errorc(ctx, "conn.Receive(ltkey:%d) error(%v)", ltkey, err)
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
	conn := dao.S10PointShopRedis.Get(ctx)
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

func (d *Dao) AddExchangeRoundStaticCache(ctx context.Context, mid int64, currtime xtime.Time, exStatic map[int32]int32) (err error) {
	conn := dao.S10PointShopRedis.Get(ctx)
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
	conn := dao.S10PointShopRedis.Get(ctx)
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

const userFlow = "s10:sf:%d"

func userFlowKey(mid int64) string {
	return fmt.Sprintf(userFlow, mid)
}

// value:1- 哨兵；2-联通；3-移动
func (d *Dao) AddUserFlowCache(ctx context.Context, mid, value int64) error {
	conn := dao.S10PointShopRedis.Get(ctx)
	defer conn.Close()
	ufkey := userFlowKey(mid)

	err := conn.Send("SET", ufkey, value)
	if err != nil {
		log.Errorc(ctx, "s10 conn.Send(key:%s) error:%v", ufkey, err)
		return err
	}
	err = conn.Send("EXPIRE", ufkey, d.userFlowExpire)
	if err != nil {
		log.Errorc(ctx, "s10 conn.Send(key:%s) error:%v", ufkey, err)
		return err
	}
	err = conn.Flush()
	if err != nil {
		log.Errorc(ctx, "s10 conn.Flush(key:%s) error:%v", ufkey, err)
		return err
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Errorc(ctx, "s10 conn.Receive(key:%s) error:%v", ufkey, err)
			return err
		}
	}
	return nil
}
