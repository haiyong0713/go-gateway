package lottery

import (
	"context"
	"encoding/json"
	"math/rand"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/log"
	lottery "go-gateway/app/web-svr/activity/interface/model/lottery_v2"

	"github.com/pkg/errors"
)

// CacheLottery cache lottery
func (d *dao) CacheLottery(c context.Context, sid string) (res *lottery.Lottery, err error) {
	var (
		key = buildKey(sid)
		bs  []byte
	)
	if bs, err = redis.Bytes(d.redisNew.Do(c, "GET", key)); err != nil {
		log.Errorc(c, "CacheLottery conn.Do(GET key(%v)) error(%v)", key, err)
		return
	}
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Errorc(c, "json.Unmarshal(%v) error(%v)", bs, err)
	}
	return
}

// AddCacheLottery add cache lottery
func (d *dao) AddCacheLottery(c context.Context, sid string, val *lottery.Lottery) (err error) {
	var (
		key = buildKey(sid)
		bs  []byte
	)
	if bs, err = json.Marshal(val); err != nil {
		log.Errorc(c, "json.Marshal(%v) error (%v)", val, err)
		return
	}
	if _, err = d.redisNew.Do(c, "SETEX", key, d.lotteryExpire, bs); err != nil {
		log.Errorc(c, "AddCacheLottery conn.Send(SETEX, %s, %v, %s) error(%v)", key, d.lotteryExpire, string(bs), err)
		return
	}
	return
}

// DeleteLottery ...
func (d *dao) DeleteLottery(c context.Context, sid string) (err error) {
	var (
		key = buildKey(sid)
	)
	if _, err = d.redisNew.Do(c, "DEL", key); err != nil {
		log.Errorc(c, "DeleteLottery conn.Do(DEL, %s) error(%v)", key, err)
	}
	return
}

// CacheLotteryInfo cache lottery info
func (d *dao) CacheLotteryInfo(c context.Context, sid string) (res *lottery.Info, err error) {
	var (
		key = buildKey(infoKey, sid)
		bs  []byte
	)
	if bs, err = redis.Bytes(d.redisNew.Do(c, "GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			log.Warnc(c, "CacheLotteryInfo(%s) return nil", key)
		} else {
			log.Errorc(c, "CacheLotteryInfo conn.Do(GET key(%v)) error(%v)", key, err)
		}
		return
	}
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Errorc(c, "json.Unmarshal(%v) error(%v)", bs, err)
	}
	return
}

// AddCacheLotteryInfo add cache lottery info
func (d *dao) AddCacheLotteryInfo(c context.Context, sid string, val *lottery.Info) (err error) {
	var (
		key = buildKey(infoKey, sid)
		bs  []byte
	)
	if bs, err = json.Marshal(val); err != nil {
		log.Errorc(c, "json.Marshal(%v) error (%v)", val, err)
		return
	}
	if _, err = d.redisNew.Do(c, "SETEX", key, d.lotteryExpire, bs); err != nil {
		log.Errorc(c, "AddCacheLotteryInfo conn.Send(SETEX, %s, %v, %s) error(%v)", key, d.lotteryExpire, string(bs), err)
	}
	return
}

// DeleteLotteryInfo ...
func (d *dao) DeleteLotteryInfo(c context.Context, sid string) (err error) {
	var (
		key = buildKey(infoKey, sid)
	)
	if _, err = d.redisNew.Do(c, "DEL", key); err != nil {
		log.Errorc(c, "DeleteLotteryInfo conn.Do(DEL, %s) error(%v)", key, err)
	}
	return
}

// CacheLotteryTimesConfig cache lottery times config
func (d *dao) CacheLotteryTimesConfig(c context.Context, sid string) (res []*lottery.TimesConfig, err error) {
	var (
		key = buildKey(timesConfKey, sid)
		bs  []byte
	)
	if bs, err = redis.Bytes(d.redisNew.Do(c, "GET", key)); err != nil {
		log.Errorc(c, "CacheLotteryTimesConfig conn.Do(GET key(%v)) error(%v)", key, err)
		return
	}
	res = []*lottery.TimesConfig{}
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Errorc(c, "json.Unmarshal(%v) error(%v)", bs, err)
	}
	return
}

// AddCacheLotteryTimesConfig add cache lottery times config
func (d *dao) AddCacheLotteryTimesConfig(c context.Context, sid string, list []*lottery.TimesConfig) (err error) {
	var (
		key = buildKey(timesConfKey, sid)
		bs  []byte
	)
	if bs, err = json.Marshal(list); err != nil {
		log.Error("json.Marshal(%v) error (%v)", list, err)
		return
	}
	if _, err = d.redisNew.Do(c, "SETEX", key, d.lotteryExpire, bs); err != nil {
		log.Errorc(c, "conn.Send(SETEX, %s, %v, %s) error(%v)", key, d.lotteryExpire, string(bs), err)
	}
	return
}

// DeleteLotteryTimesConfig ...
func (d *dao) DeleteLotteryTimesConfig(c context.Context, sid string) (err error) {
	var (
		key = buildKey(timesConfKey, sid)
	)
	if _, err = d.redisNew.Do(c, "DEL", key); err != nil {
		log.Errorc(c, "DeleteLotteryTimesConfig conn.Do(DEL, %s) error(%v)", key, err)
	}
	return
}

// CacheLotteryGift cache lottery gift
func (d *dao) CacheLotteryGift(c context.Context, sid string) (res []*lottery.Gift, err error) {
	var (
		key = buildKey(giftKey, sid)
		bs  []byte
	)
	if bs, err = redis.Bytes(d.redisNew.Do(c, "GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			log.Warn("CacheLotteryGift(%s) return nil", key)
		} else {
			log.Errorc(c, "CacheLotteryGift conn.Do(GET key(%v)) error(%v)", key, err)
		}
		return
	}
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Errorc(c, "json.Unmarshal(%s) error(%v)", string(bs), err)
	}
	return
}

// AddCacheLotteryGift ...
func (d *dao) AddCacheLotteryGift(c context.Context, sid string, list []*lottery.Gift) (err error) {
	var (
		key = buildKey(giftKey, sid)
		bs  []byte
	)
	if bs, err = json.Marshal(list); err != nil {
		log.Errorc(c, "json.Marshal(%v) error (%v)", list, err)
		return
	}
	if _, err = d.redisNew.Do(c, "SETEX", key, d.lotteryExpire, bs); err != nil {
		log.Errorc(c, "AddCacheLotteryGift conn.Send(SETEX, %s, %v, %s) error(%v)", key, d.lotteryExpire, string(bs), err)
	}
	return
}

// DeleteLotteryGift ...
func (d *dao) DeleteLotteryGift(c context.Context, sid string) (err error) {
	var (
		key = buildKey(giftKey, sid)
	)
	if _, err = d.redisNew.Do(c, "DEL", key); err != nil {
		log.Errorc(c, "DeleteLotteryGift conn.Do(DEL, %s) error(%v)", key, err)
	}
	return
}

// CacheLotteryAddrCheck ...
func (d *dao) CacheLotteryAddrCheck(c context.Context, id, mid int64) (res int64, err error) {
	var (
		key = buildKey(addressKey, id, mid)
	)
	if res, err = redis.Int64(d.redisNew.Do(c, "GET", key)); err != nil {
		if err != redis.ErrNil {
			log.Errorc(c, "CacheLotteryAddrCheck conn.Do(GET key(%v)) error(%v)", key, err)
		}
		return res, err
	}
	return

}

// AddCacheLotteryAddrCheck ...
func (d *dao) AddCacheLotteryAddrCheck(c context.Context, id, mid int64, val int64) (err error) {
	var (
		key = buildKey(addressKey, id, mid)
	)
	if _, err = d.redisNew.Do(c, "SETEX", key, d.lotteryExpire, val); err != nil {
		log.Errorc(c, "AddCacheLotteryAddrCheck conn.Send(SETEX, %s, %v, %d) error(%v)", key, d.lotteryExpire, val, err)
	}
	return
}

// CacheIPRequestCheck ...
func (d *dao) CacheIPRequestCheck(c context.Context, ip string) (res int, err error) {
	var (
		key = buildKey(ipKey, ip)
	)
	if res, err = redis.Int(d.redisNew.Do(c, "GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			log.Warnc(c, "CacheIPRequestCheck(%s) return nil", key)
		} else {
			log.Errorc(c, "CacheIPRequestCheck conn.Do(GET key(%v)) error(%v)", key, err)
		}
	}
	return
}

// AddCacheIPRequestCheck ...
func (d *dao) AddCacheIPRequestCheck(c context.Context, ip string, val int) (err error) {
	var (
		key = buildKey(ipKey, ip)
	)
	if _, err = d.redisNew.Do(c, "SETEX", key, d.lotteryIPExpire, val); err != nil {
		log.Errorc(c, "AddCacheIPRequestCheck conn.Send(SETEX, %s, %v, %d) error(%v)", key, d.lotteryIPExpire, val, err)
	}
	return
}

// CacheLotteryTimes ...
func (d *dao) CacheLotteryTimes(c context.Context, sid int64, mid int64, remark string) (list map[string]int, err error) {

	list = make(map[string]int)
	key := buildKey(timesKey, sid, remark, mid)
	args := redis.Args{}.Add(key)
	// for _, v := range recordBatch {
	// 	args = args.Add(v)
	// }
	var tmp map[string]int
	if tmp, err = redis.IntMap(d.redisNew.Do(c, "HGETALL", args...)); err != nil {
		log.Errorc(c, "CacheLotteryTimes redis.Ints(MGET) args(%v) error(%v)", args, err)
		return
	}
	if len(tmp) == 0 {
		return nil, redis.ErrNil
	}
	return tmp, nil
}

// AddCacheLotteryTimes ...
func (d *dao) AddCacheLotteryTimes(c context.Context, sid int64, mid int64, remark string, list map[string]int) (err error) {
	if len(list) == 0 {
		return
	}
	var (
		conn = d.redisNew.Conn(c)
		key  = buildKey(timesKey, sid, remark, mid)
		args = redis.Args{}.Add(key)
	)
	for k, v := range list {
		args = args.Add(k).Add(v)
	}
	defer conn.Close()
	if err = conn.Send("DEL", key); err != nil {
		log.Errorc(c, "AddCacheLotteryTimes conn.Send(DEL, %s) error(%v)", key, err)
		return
	}
	if err = conn.Send("HMSET", args...); err != nil {
		log.Errorc(c, "AddCacheLotteryTimes conn.Send(HMSET) error(%v)", err)
		return
	}
	if err = conn.Send("EXPIRE", key, d.lotteryTimesExpire); err != nil {
		log.Errorc(c, "conn.Send(EXPIRE, %s) error(%v)", key, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Errorc(c, "conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < 3; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Errorc(c, "conn.Receive() error(%v)", err)
			return
		}
	}
	return
}

// IncrTimes /  IncrUsedTimes.
func (d *dao) IncrTimes(c context.Context, sid int64, mid int64, list map[string]int, status string) (err error) {
	if len(list) == 0 {
		return
	}
	var (
		key = buildKey(timesKey, sid, status, mid)
	)
	for k, v := range list {
		if _, err = redis.Int(d.redisNew.Do(c, "HINCRBY", key, k, v)); err != nil {
			err = errors.Wrap(err, "IncrTimes redis.Do(HINCRBY)")
			return
		}
	}
	return
}

// CacheLotteryWinList ...
func (d *dao) CacheLotteryWinList(c context.Context, sid int64) (res []*lottery.GiftMid, err error) {
	var (
		key = buildKey(winKey, sid)
		bs  []byte
	)
	if bs, err = redis.Bytes(d.redisNew.Do(c, "GET", key)); err != nil {
		log.Errorc(c, "CacheLotteryWinList conn.Do(GET key(%v)) error(%v)", key, err)
		return
	}
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Errorc(c, "json.Unmarshal(%v) error(%v)", bs, err)
	}
	return
}
func (d *dao) getExpireBeforeDawn() int64 {

	timeStr := time.Now().Format("2006-01-02")
	//使用Parse 默认获取为UTC时区 需要获取本地时区 所以使用ParseInLocation
	rand.Seed(time.Now().Unix())

	t, _ := time.ParseInLocation("2006-01-02 15:04:05", timeStr+" 23:59:59", time.Local)
	expireTime := t.Unix() - time.Now().Unix() + rand.Int63n(1000)
	return expireTime
}

// AddCacheLotteryWinList ...
func (d *dao) AddCacheLotteryWinList(c context.Context, sid int64, list []*lottery.GiftMid) (err error) {
	var (
		key = buildKey(winKey, sid)
		bs  []byte
	)
	if bs, err = json.Marshal(list); err != nil {
		log.Errorc(c, "json.Marshal(%v) error (%v)", list, err)
		return
	}
	if _, err = d.redisNew.Do(c, "SETEX", key, d.getExpireBeforeDawn(), bs); err != nil {
		log.Errorc(c, "AddCacheLotteryWinList conn.Send(SETEX, %s, %v, %s) error(%v)", key, d.lotteryWinListExpire, string(bs), err)
	}
	return
}

// get lottery log from redis
func (d *dao) CacheLotteryActionLog(c context.Context, sid int64, mid int64, start, end int64) (res []*lottery.RecordDetail, err error) {
	var (
		key = buildKey(actionKey, sid, mid)
	)
	values, err := redis.Values(d.redisNew.Do(c, "ZREVRANGE", key, start, end))
	if err != nil {
		log.Errorc(c, "CacheLotteryActionLog conn.Do(ZREVRANGE, %s) error(%v)", key, err)
		return
	}
	if len(values) == 0 {
		return
	}
	for len(values) > 0 {
		bs := []byte{}
		if values, err = redis.Scan(values, &bs); err != nil {
			log.Errorc(c, "redis.Scan(%v) error(%v)", values, err)
			return
		}
		list := &lottery.RecordDetail{}
		if err = json.Unmarshal(bs, list); err != nil {
			log.Errorc(c, "json.Unmarshal(%v) error(%v)", bs, err)
			return
		}
		res = append(res, list)
	}
	return
}

// CacheLotteryWinLog get lottery log from redis
func (d *dao) CacheLotteryWinLog(c context.Context, sid, mid, start, end int64) (res []*lottery.MidWinList, err error) {
	var (
		key = buildKey(realyWinKey, sid, mid)
	)
	values, err := redis.Values(d.redisNew.Do(c, "ZREVRANGE", key, start, end))
	if err != nil {
		log.Errorc(c, "CacheLotteryActionLog conn.Do(ZREVRANGE, %s) error(%v)", key, err)
		return
	}
	if len(values) == 0 {
		return
	}
	for len(values) > 0 {
		bs := []byte{}
		if values, err = redis.Scan(values, &bs); err != nil {
			log.Errorc(c, "redis.Scan(%v) error(%v)", values, err)
			return
		}
		list := &lottery.MidWinList{}
		if err = json.Unmarshal(bs, list); err != nil {
			log.Errorc(c, "json.Unmarshal(%v) error(%v)", bs, err)
			return
		}
		res = append(res, list)
	}
	return
}

// DeleteLotteryActionLog ...
func (d *dao) DeleteLotteryWinLog(c context.Context, sid, mid int64) (err error) {
	var (
		key = buildKey(realyWinKey, sid, mid)
	)

	if _, err = d.redisNew.Do(c, "DEL", key); err != nil {
		log.Errorc(c, "DeleteLotteryWinLog conn.Do(DEL, %s) error(%v)", key, err)
	}
	return
}

// set lottery log into redis
func (d *dao) AddCacheLotteryWinLog(c context.Context, sid, mid int64, list []*lottery.MidWinList) (err error) {
	if len(list) == 0 {
		return
	}
	var (
		key  = buildKey(realyWinKey, sid, mid)
		conn = d.redisNew.Conn(c)
		bs   []byte
	)
	defer conn.Close()
	if err = conn.Send("DEL", key); err != nil {
		log.Errorc(c, "AddCacheLotteryWinLog conn.Send(DEL, %s) error(%v)", key, err)
		return
	}
	args := redis.Args{}.Add(key)
	for _, v := range list {
		if bs, err = json.Marshal(v); err != nil {
			log.Errorc(c, "AddCacheLotteryWinLog json.Marshal() error(%v)", err)
			return
		}
		args = args.Add(v.Mtime).Add(bs)
	}
	if err = conn.Send("ZADD", args...); err != nil {
		log.Errorc(c, "AddCacheLotteryWinLog conn.Send(ZADD, %s) error(%v)", key, err)
		return
	}
	if err = conn.Send("EXPIRE", key, d.lotteryWinListExpire); err != nil {
		log.Errorc(c, "AddCacheLotteryWinLog conn.Send(EXPIRE,%s) error(%v)", key, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Errorc(c, "conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < 3; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Errorc(c, "conn.Receive() error(%v)", err)
			return
		}
	}
	return
}

// set lottery log into redis
func (d *dao) AddCacheLotteryActionLog(c context.Context, sid int64, mid int64, list []*lottery.RecordDetail) (err error) {
	if len(list) == 0 {
		return
	}
	var (
		key  = buildKey(actionKey, sid, mid)
		conn = d.redisNew.Conn(c)
		bs   []byte
	)
	defer conn.Close()
	if err = conn.Send("DEL", key); err != nil {
		log.Errorc(c, "AddCacheLotteryActionLog conn.Send(DEL, %s) error(%v)", key, err)
		return
	}
	args := redis.Args{}.Add(key)
	for _, v := range list {
		if bs, err = json.Marshal(v); err != nil {
			log.Errorc(c, "AddCacheLotteryActionLog json.Marshal() error(%v)", err)
			return
		}
		args = args.Add(v.Ctime).Add(bs)
	}
	if err = conn.Send("ZADD", args...); err != nil {
		log.Errorc(c, "AddCacheLotteryActionLog conn.Send(ZADD, %s) error(%v)", key, err)
		return
	}
	if err = conn.Send("EXPIRE", key, d.lotteryExpire); err != nil {
		log.Errorc(c, "AddCacheLotteryActionLog conn.Send(EXPIRE,%s) error(%v)", key, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Errorc(c, "conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < 3; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Errorc(c, "conn.Receive() error(%v)", err)
			return
		}
	}
	return
}

// AddLotteryActionLog add lottery log into redis
func (d *dao) AddLotteryActionLog(c context.Context, sid int64, mid int64, list []*lottery.RecordDetail) (err error) {
	if len(list) == 0 {
		return
	}
	var (
		key  = buildKey(actionKey, sid, mid)
		conn = d.redisNew.Conn(c)
		bs   []byte
	)
	defer conn.Close()
	args := redis.Args{}.Add(key)
	for _, v := range list {
		lr := &lottery.RecordDetail{ID: v.ID, Mid: v.Mid, Num: v.Num, GiftID: v.GiftID, Type: v.Type, Ctime: v.Ctime, CID: v.CID}
		if bs, err = json.Marshal(lr); err != nil {
			log.Error("AddLotteryActionLog json.Marshal() error(%v)", err)
			return
		}
		args = args.Add(lr.Ctime).Add(bs)
	}
	if err = conn.Send("ZADD", args...); err != nil {
		log.Errorc(c, "AddLotteryActionLog conn.Send(ZADD, %s) error(%v)", key, err)
		return
	}
	if err = conn.Send("EXPIRE", key, d.lotteryExpire); err != nil {
		log.Errorc(c, "AddLotteryActionLog conn.Send(EXPIRE,%s) error(%v)", key, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Errorc(c, "conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Errorc(c, "conn.Receive() error(%v)", err)
			return
		}
	}
	return
}

// DeleteLotteryActionLog ...
func (d *dao) DeleteLotteryActionLog(c context.Context, sid int64, mid int64) (err error) {
	var (
		key = buildKey(actionKey, sid, mid)
	)
	if _, err = d.redisNew.Do(c, "DEL", key); err != nil {
		log.Errorc(c, "DeleteMemberGroup conn.Do(DEL, %s) error(%v)", key, err)
	}
	return
}

// CacheSendGiftNum 已发送商品数记录
func (d *dao) CacheSendGiftNum(c context.Context, sid int64, giftIds []int64) (num map[int64]int64, err error) {
	num = make(map[int64]int64)
	if len(giftIds) == 0 {
		return num, nil
	}
	var (
		key  = buildKey(giftNumKey, sid)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	args := redis.Args{}.Add(key)
	list := make([]int64, 0, len(giftIds))
	for _, k := range giftIds {
		args = args.Add(k)
		list = append(list, k)
	}
	values, err := redis.Int64s(conn.Do("HMGET", args...))
	if err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		log.Errorc(c, "CacheSendGiftNum conn.Do(HMGET %v) error(%v)", args, err)
		return
	}
	for k, v := range values {
		num[list[k]] = v
	}
	return
}

// IncrGiftSendNum 新增发送的商品数
func (d *dao) IncrGiftSendNum(c context.Context, sid int64, giftIDNum map[int64]int) (resMap map[int64]int64, err error) {
	resMap = make(map[int64]int64)
	if len(giftIDNum) == 0 {
		return resMap, nil
	}
	var (
		key  = buildKey(giftNumKey, sid)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	defer func() {
		if err = conn.Send("EXPIRE", key, d.lotteryExpire); err != nil {
			log.Error("EXPIRE conn.Send() error(%v)", err)
			return
		}
	}()
	for id, num := range giftIDNum {
		res, err := redis.Int64(conn.Do("HINCRBY", key, id, num))
		if err != nil {
			err = errors.Wrap(err, "IncrTimes redis.Do(HINCRBY)")
			log.Errorc(c, "IncrGiftSendNum HINCRBY conn.Send(key:%s,field:%s,value:%d) error(%v)", key, giftKey, id, err)
			return resMap, err
		}
		resMap[id] = res
	}

	return
}

// CacheSendDayGiftNum 已发送每日商品数记录
func (d *dao) CacheSendDayGiftNum(c context.Context, sid int64, day string, giftKeys []string) (num map[string]int64, err error) {
	num = make(map[string]int64)
	if len(giftKeys) == 0 {
		return
	}
	var (
		key  = buildKey(giftDayNumKey, sid, day)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	args := redis.Args{}.Add(key)
	list := make([]string, 0, len(giftKeys))
	for _, k := range giftKeys {
		args = args.Add(k)
		list = append(list, k)
	}
	values, err := redis.Int64s(conn.Do("HMGET", args...))
	if err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		log.Errorc(c, "CacheSendDayGiftNum conn.Do(HMGET %v) error(%v)", args, err)
		return
	}
	for k, v := range values {
		num[list[k]] = v
	}
	return
}

// IncrGiftSendDayNum 新增发送日商品数
func (d *dao) IncrGiftSendDayNum(c context.Context, sid int64, day string, giftKeysNum map[string]int, exipreTime int64) (resGiftKeysNum map[string]int64, err error) {
	resGiftKeysNum = make(map[string]int64)
	if giftKeysNum == nil || len(giftKeysNum) == 0 {
		return resGiftKeysNum, nil
	}
	var (
		key  = buildKey(giftDayNumKey, sid, day)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	defer func() {
		if err = conn.Send("EXPIRE", key, exipreTime); err != nil {
			log.Errorc(c, "EXPIRE conn.Send() error(%v)", err)
			return
		}
	}()
	for giftKey, v := range giftKeysNum {
		res, err := redis.Int64(conn.Do("HINCRBY", key, giftKey, v))
		if err != nil {
			err = errors.Wrap(err, "IncrTimes redis.Do(HINCRBY)")
			log.Errorc(c, "IncrGiftSendDayNum HINCRBY conn.Send(key:%s,field:%s,value:%d) error(%v)", key, giftKey, v, err)
			return resGiftKeysNum, err
		}
		resGiftKeysNum[giftKey] = res

	}

	return
}

// AddCacheMemberGroup ...
func (d *dao) AddCacheMemberGroup(c context.Context, sid string, list []*lottery.MemberGroup) (err error) {
	var (
		key = buildKey(memberGroupKey, sid)
		bs  []byte
	)
	if bs, err = json.Marshal(list); err != nil {
		log.Error("json.Marshal(%v) error (%v)", list, err)
		return
	}
	if _, err = d.redisNew.Do(c, "SETEX", key, d.lotteryExpire, bs); err != nil {
		log.Errorc(c, "AddCacheMemberGroup conn.Send(SETEX, %s, %v, %s) error(%v)", key, d.lotteryExpire, string(bs), err)
	}
	return
}

// DeleteMemberGroup ...
func (d *dao) DeleteMemberGroup(c context.Context, sid string) (err error) {
	var (
		key = buildKey(memberGroupKey, sid)
	)
	if _, err = d.redisNew.Do(c, "DEL", key); err != nil {
		log.Error("DeleteMemberGroup conn.Do(DEL, %s) error(%v)", key, err)
	}
	return
}

// CacheMemberGroup cache lottery gift
func (d *dao) CacheMemberGroup(c context.Context, sid string) (res []*lottery.MemberGroup, err error) {
	var (
		key = buildKey(memberGroupKey, sid)
		bs  []byte
	)
	if bs, err = redis.Bytes(d.redisNew.Do(c, "GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			log.Warnc(c, "CacheMemberGroup(%s) return nil", key)
		} else {
			log.Errorc(c, "CacheMemberGroup conn.Do(GET key(%v)) error(%v)", key, err)
		}
		return
	}
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Errorc(c, "json.Unmarshal(%s) error(%v)", string(bs), err)
	}
	return
}

// CacheQPSLimit cache qps limit
func (d *dao) CacheQPSLimit(c context.Context, mid int64) (num int64, err error) {
	timestamp := time.Now().Unix()
	var (
		key = buildKey(qpsLimitKey, mid, timestamp)
		res interface{}
	)

	if res, err = d.redisNew.Do(c, "INCR", key); err != nil {
		log.Errorc(c, "CacheQpsLimit conn.Send(INCR, %s) error(%v)", key, err)
		return
	}
	num = res.(int64)
	if num == int64(1) {
		if res, err = d.redisNew.Do(c, "EXPIRE", key, d.qpsLimitExpire); err != nil {
			log.Errorc(c, "CacheQpsLimit conn.Send(INCEXPIRER, %s) error(%v)", key, err)
		}
	}
	return num, nil
}

// GiftOtherSendNumIncr Dao
func (d *dao) GiftOtherSendNumIncr(c context.Context, key string, num int) (res bool, err error) {
	var (
		rkey = buildKey(key)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if err = conn.Send("INCRBY", rkey, num); err != nil {
		log.Error("IncrWithExpire conn.Do(INCR key(%s)) error(%v)", rkey, err)
		return
	}
	if err = conn.Send("EXPIRE", rkey, d.lotteryExpire*2); err != nil {
		log.Error("IncrWithExpire conn.Do(expire key(%s)) error(%v)", rkey, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("IncrWithExpire conn.Flush() error(%v)", err)
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("IncrWithExpire Receive error(%v)", err)
			return
		}
	}
	return
}
