package esports

import (
	"context"
	"encoding/json"
	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/model/esports"
	"time"
)

func (d *Dao) CacheEsportsArenaFav(c context.Context, actId string, fav *esports.EsportsActFav) (err error) {
	if fav == nil {
		return
	}

	var (
		key = buildKey(actId, fav.Mid)
		bs  []byte
	)
	if bs, err = json.Marshal(fav); err != nil {
		log.Errorc(c, "json.Marshal(%v) error (%v)", fav, err)
		return
	}
	if _, err = d.redisStore.Do(c, "SETEX", key, d.SevenDayExpire, bs); err != nil {
		log.Errorc(c, "AddEsportsArenaFav conn.Send(SETEX, %s, %v, %s) error(%v)", key, d.SevenDayExpire, string(bs), err)
		return
	}
	return
}

func (d *Dao) GetEsportsArenaFav(c context.Context, actId string, mid int64) (fav *esports.EsportsActFav, err error) {
	var (
		key = buildKey(actId, mid)
		bs  []byte
	)
	if bs, err = redis.Bytes(d.redisStore.Do(c, "GET", key)); err != nil {
		log.Errorc(c, "GetEsportsArenaFav conn.Do(GET key(%v)) error(%v)", key, err)
		return
	}
	if bs != nil && len(bs) > 0 {
		if err = json.Unmarshal(bs, &fav); err != nil {
			log.Errorc(c, "json.Unmarshal(%v) error(%v)", bs, err)
		}
	}
	return
}

func (d *Dao) SetNxOrder(c context.Context, mid int64, orderNo string) (ok bool, err error) {
	key := buildKey("order", mid, orderNo)
	return redis.Bool(d.redisStore.Do(c, "SETNX", key, time.Now().Unix()))
}

func (d *Dao) IncrLotteryTimes(c context.Context, date string, mid int64) (times int64, err error) {
	key := buildKey("times", date, mid)
	return redis.Int64(d.redisStore.Do(c, "INCR", key))
}
