package bml

import (
	"context"
	"encoding/json"
	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/model/bml"
	"time"
)

const (
	_userGuessOrderCachePre = "user:guess:order"
	_jokerKeyCache          = "joker:key:cache"
	_frequencyControl       = "user:frequency:control"
	_jokerKeyGuessTime      = "user:joker:key:guess:time"
	_userAnswerRecord       = "user:answer:record"
)

func (d *Dao) AddGuessOrderCache(ctx context.Context, mid int64, orderList []*bml.GuessOrderRecord, expireTime int32) (err error) {
	cacheKey := buildKey(_userGuessOrderCachePre, mid)
	var (
		data []byte
	)
	if data, err = json.Marshal(orderList); err != nil {
		return err
	}
	if expireTime <= 0 {
		expireTime = d.userExpire
	}
	if _, err = d.redis.Do(ctx, "SETEX", cacheKey, expireTime, data); err != nil {
		log.Errorc(ctx, "AddGuessOrderCache conn.Do(SETEX) key(%s) error(%v)", cacheKey, err)
	}
	return
}

func (d *Dao) CacheJokerKey(ctx context.Context, jokerKey string, expireTime int32) (err error) {
	cacheKey := buildKey(_jokerKeyCache, jokerKey)
	if expireTime <= 0 {
		expireTime = d.dataExpire * 10
	}
	nowTime := time.Now().Unix()
	var reply interface{}
	if reply, err = d.redis.Do(ctx, "SET", cacheKey, nowTime, "EX", expireTime, "NX"); err != nil {
		log.Errorc(ctx, "CacheJokerKey conn.Do(SETEX) key(%s) error(%v) reply:%v", cacheKey, err, reply)
	}
	return
}

func (d *Dao) GetJokerKeyCache(ctx context.Context, jokerKey string) (ts int64, err error) {
	cacheKey := buildKey(_jokerKeyCache, jokerKey)
	return redis.Int64(d.redis.Do(ctx, "GET", cacheKey))
}

func (d *Dao) FrequencyControl(ctx context.Context, mid int64, guessType int, expireTime int64) (err error) {
	cacheKey := buildKey(_frequencyControl, mid, guessType)
	if expireTime <= 0 {
		expireTime = 2
	}
	nowTime := time.Now().Unix()
	var reply interface{}
	reply, err = d.redis.Do(ctx, "SET", cacheKey, nowTime, "EX", expireTime, "NX")
	log.Infoc(ctx, "CacheJokerKey FrequencyControl key(%s) error(%v) reply:%v", cacheKey, err, reply)
	if err != nil {
		return
	}
	if reply == nil {
		err = ecode.SpringFestivalTooFastErr
	}
	return
}

func (d *Dao) IncrJokerKeyGuessTime(ctx context.Context, mid int64) (times int64, err error) {
	cacheKey := buildKey(_jokerKeyGuessTime, time.Now().Format("20060102"), mid)
	times, err = redis.Int64(d.redis.Do(ctx, "INCR", cacheKey))
	_ = d.cache.SyncDo(ctx, func(ctx context.Context) {
		if times == 1 && err == nil {
			replay, err2 := d.redis.Do(ctx, "EXPIRE", cacheKey, 86400*7)
			log.Infoc(ctx, "IncrJokerKeyGuessTime , redis key:%v , EXPIRE:%v , err:%v", cacheKey, replay, err2)
		}
	})
	return
}

func (d *Dao) CacheUserAnswerRecord(ctx context.Context, item *bml.GuessRecordItem, mid int64, expireTime int64) (err error) {
	cacheKey := buildKey(_userAnswerRecord, item.GuessType, mid)
	data, _ := json.Marshal(item)
	var reply interface{}
	reply, err = d.redis.Do(ctx, "SET", cacheKey, data, "EX", expireTime, "NX")
	log.Infoc(ctx, "CacheUserAnswerRecord  key(%s) error(%v) reply:%v", cacheKey, err, reply)
	return
}

func (d *Dao) GetUserAnswerRecord(ctx context.Context, mid int64, guessType int) (item *bml.GuessRecordItem, err error) {
	cacheKey := buildKey(_userAnswerRecord, guessType, mid)
	var data []byte
	if data, err = redis.Bytes(d.redis.Do(ctx, "GET", cacheKey)); err != nil {
		return
	}
	item = &bml.GuessRecordItem{}
	err = json.Unmarshal(data, item)
	return
}
