package lottery

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/component"
	l "go-gateway/app/web-svr/activity/interface/model/lottery"

	"github.com/pkg/errors"
)

// actLotteryKey act_lottery table line cache .
func actLotteryKey(sid string) string {
	return fmt.Sprintf("act_lottery_%s", sid)
}

// actLotteryInfo act_lottery_info table line cache
func actLotteryInfoKey(sid string) string {
	return fmt.Sprintf("act_lottery_info_%s", sid)
}

// actLotteryInfo act_lottery_info table line cache
func actLotteryTimesConfKey(sid string) string {
	return fmt.Sprintf("act_lottery_times_conf_%s", sid)
}

// actLotteryGift act_lottery_gift table line cache
func actLotteryGiftKey(sid string) string {
	return fmt.Sprintf("act_lottery_gift_%s", sid)
}

// lotteryAddrCheck.
func lotteryAddrCheckKey(id, mid int64) string {
	return fmt.Sprintf("act_lottery_address_%d_%d", id, mid)
}

// ipReqkey.
func ipReqKey(ip string) string {
	return fmt.Sprintf("act_lottery_ip_%s", ip)
}

// lotteryTimesKey
func lotteryTimesField(ltc *l.LotteryTimesConfig) string {
	if ltc.AddType == 1 {
		nowTs := time.Now().Format("2006-01-02")
		return fmt.Sprintf("%d_%d_%s", ltc.Type, ltc.ID, nowTs)
	}
	return fmt.Sprintf("%d_%d_%s", ltc.Type, ltc.ID, "0")
}

func lotteryWinListKey(sid int64) string {
	return fmt.Sprintf("act_lottery_win_%d", sid)
}

func lotteryActionKey(sid int64, mid int64) string {
	return fmt.Sprintf("act_lottery_action_%d_%d", sid, mid)
}
func lotteryWinLogKey(sid int64, mid int64) string {
	return fmt.Sprintf("lottery_new:realy_win:%d:%d", sid, mid)

}

func lotteryGiftNumKey(id int64) string {
	return fmt.Sprintf("lottery_gift_num_%d", id)
}

func lotteryTimesMapKey(ltc *l.LotteryTimesConfig) string {
	return strconv.Itoa(ltc.Type) + strconv.FormatInt(ltc.ID, 10)
}

func lotteryTimesKey(sid, mid int64, remark string) string {
	return fmt.Sprintf("lottery_times_%s_%d_%d", remark, sid, mid)
}

func lotteryPlayWindowMidKey(mid int64) string {
	return fmt.Sprintf("lottery_play_mid_%d_window", mid)
}

func lotteryPlayWindowBuvidKey(buvid string) string {
	return fmt.Sprintf("lottery_play_buvid_%s_window", buvid)
}

func (dao *Dao) CacheLottery(c context.Context, sid string) (res *l.Lottery, err error) {
	var (
		key  = actLotteryKey(sid)
		conn = dao.redis.Get(c)
		bs   []byte
	)
	defer conn.Close()
	if bs, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			log.Warn("CacheLottery(%s) return nil", key)
		} else {
			log.Error("CacheLottery conn.Do(GET key(%v)) error(%v)", key, err)
		}
		return
	}
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Error("json.Unmarshal(%v) error(%v)", bs, err)
	}
	return
}

func (dao *Dao) getExpireBeforeDawn() int64 {

	timeStr := time.Now().Format("2006-01-02")
	//使用Parse 默认获取为UTC时区 需要获取本地时区 所以使用ParseInLocation
	rand.Seed(time.Now().Unix())

	t, _ := time.ParseInLocation("2006-01-02 15:04:05", timeStr+" 23:59:59", time.Local)
	expireTime := t.Unix() - time.Now().Unix() + rand.Int63n(1000)
	return expireTime
}

func (dao *Dao) AddCacheLottery(c context.Context, sid string, val *l.Lottery) (err error) {
	var (
		key  = actLotteryKey(sid)
		conn = dao.redis.Get(c)
		bs   []byte
	)
	defer conn.Close()
	if bs, err = json.Marshal(val); err != nil {
		log.Error("json.Marshal(%v) error (%v)", val, err)
		return
	}
	if err = conn.Send("SETEX", key, dao.lotteryExpire, bs); err != nil {
		log.Error("AddCacheLottery conn.Send(SETEX, %s, %v, %s) error(%v)", key, dao.lotteryExpire, string(bs), err)
		return
	}
	return
}

// DeleteLottery ...
func (dao *Dao) DeleteLottery(c context.Context, sid string) (err error) {
	var (
		key  = actLotteryKey(sid)
		conn = dao.redis.Get(c)
	)
	defer conn.Close()
	if _, err = conn.Do("DEL", key); err != nil {
		log.Error("DeleteLottery conn.Do(DEL, %s) error(%v)", key, err)
	}
	return
}

func (dao *Dao) CacheLotteryInfo(c context.Context, sid string) (res *l.LotteryInfo, err error) {
	var (
		key  = actLotteryInfoKey(sid)
		conn = dao.redis.Get(c)
		bs   []byte
	)
	defer conn.Close()
	if bs, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			log.Warn("CacheLotteryInfo(%s) return nil", key)
		} else {
			log.Error("CacheLotteryInfo conn.Do(GET key(%v)) error(%v)", key, err)
		}
		return
	}
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Error("json.Unmarshal(%v) error(%v)", bs, err)
	}
	return
}

func (dao *Dao) AddCacheLotteryInfo(c context.Context, sid string, val *l.LotteryInfo) (err error) {
	var (
		key  = actLotteryInfoKey(sid)
		conn = dao.redis.Get(c)
		bs   []byte
	)
	defer conn.Close()
	if bs, err = json.Marshal(val); err != nil {
		log.Error("json.Marshal(%v) error (%v)", val, err)
		return
	}
	if err = conn.Send("SETEX", key, dao.lotteryExpire, bs); err != nil {
		log.Error("AddCacheLotteryInfo conn.Send(SETEX, %s, %v, %s) error(%v)", key, dao.lotteryExpire, string(bs), err)
	}
	return
}

// DeleteLotteryInfo ...
func (dao *Dao) DeleteLotteryInfo(c context.Context, sid string) (err error) {
	var (
		key  = actLotteryInfoKey(sid)
		conn = dao.redis.Get(c)
	)
	defer conn.Close()
	if _, err = conn.Do("DEL", key); err != nil {
		log.Error("DeleteLotteryInfo conn.Do(DEL, %s) error(%v)", key, err)
	}
	return
}

func (dao *Dao) CacheLotteryTimesConfig(c context.Context, sid string) (res []*l.LotteryTimesConfig, err error) {
	var (
		key  = actLotteryTimesConfKey(sid)
		conn = dao.redis.Get(c)
		bs   []byte
	)
	defer conn.Close()
	if bs, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			log.Warn("CacheLotteryTimesConfig(%s) return nil", key)
		} else {
			log.Error("CacheLotteryTimesConfig conn.Do(GET key(%v)) error(%v)", key, err)
		}
		return
	}
	res = []*l.LotteryTimesConfig{}
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Error("json.Unmarshal(%v) error(%v)", bs, err)
	}
	return
}

func (dao *Dao) AddCacheLotteryTimesConfig(c context.Context, sid string, list []*l.LotteryTimesConfig) (err error) {
	var (
		key  = actLotteryTimesConfKey(sid)
		conn = dao.redis.Get(c)
		bs   []byte
	)
	defer conn.Close()
	if bs, err = json.Marshal(list); err != nil {
		log.Error("json.Marshal(%v) error (%v)", list, err)
		return
	}
	if err = conn.Send("SETEX", key, dao.lotteryExpire, bs); err != nil {
		log.Error("conn.Send(SETEX, %s, %v, %s) error(%v)", key, dao.lotteryExpire, string(bs), err)
	}
	return
}

// DeleteLotteryTimesConfig ...
func (dao *Dao) DeleteLotteryTimesConfig(c context.Context, sid string) (err error) {
	var (
		key  = actLotteryTimesConfKey(sid)
		conn = dao.redis.Get(c)
	)
	defer conn.Close()
	if _, err = conn.Do("DEL", key); err != nil {
		log.Error("DeleteLotteryTimesConfig conn.Do(DEL, %s) error(%v)", key, err)
	}
	return
}

func (dao *Dao) CacheLotteryGift(c context.Context, sid string) (res []*l.LotteryGift, err error) {
	var (
		key = actLotteryGiftKey(sid)
		bs  []byte
	)
	if bs, err = redis.Bytes(component.GlobalRedis.Do(c, "GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			log.Warn("CacheLotteryGift(%s) return nil", key)
		} else {
			log.Error("CacheLotteryGift conn.Do(GET key(%v)) error(%v)", key, err)
		}
		return
	}
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Error("json.Unmarshal(%s) error(%v)", string(bs), err)
	}
	return
}

func (dao *Dao) AddCacheLotteryGift(c context.Context, sid string, list []*l.LotteryGift) (err error) {
	var (
		key = actLotteryGiftKey(sid)
		bs  []byte
	)
	if bs, err = json.Marshal(list); err != nil {
		log.Error("json.Marshal(%v) error (%v)", list, err)
		return
	}
	if _, err = component.GlobalRedis.Do(c, "SETEX", key, dao.lotteryExpire, bs); err != nil {
		log.Error("AddCacheLotteryGift conn.Send(SETEX, %s, %v, %s) error(%v)", key, dao.lotteryExpire, string(bs), err)
	}
	return
}

// DeleteLotteryGift ...
func (dao *Dao) DeleteLotteryGift(c context.Context, sid string) (err error) {
	var (
		key = actLotteryGiftKey(sid)
	)
	if _, err = component.GlobalRedis.Do(c, "DEL", key); err != nil {
		log.Error("DeleteLotteryGift conn.Do(DEL, %s) error(%v)", key, err)
	}
	return
}

func (dao *Dao) CacheLotteryAddrCheck(c context.Context, id, mid int64) (res int64, err error) {
	var (
		key  = lotteryAddrCheckKey(id, mid)
		conn = dao.redis.Get(c)
	)
	defer conn.Close()
	if res, err = redis.Int64(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			log.Warn("CacheLotteryAddrCheck(%s) return nil", key)
		} else {
			log.Error("CacheLotteryAddrCheck conn.Do(GET key(%v)) error(%v)", key, err)
		}
	}
	return
}

func (dao *Dao) AddCacheLotteryAddrCheck(c context.Context, id, mid int64, val int64) (err error) {
	var (
		key  = lotteryAddrCheckKey(id, mid)
		conn = dao.redis.Get(c)
	)
	defer conn.Close()
	if err = conn.Send("SETEX", key, dao.lotteryExpire, val); err != nil {
		log.Error("AddCacheLotteryAddrCheck conn.Send(SETEX, %s, %v, %d) error(%v)", key, dao.lotteryExpire, val, err)
	}
	return
}

func (dao *Dao) CacheIPRequestCheck(c context.Context, ip string) (res int, err error) {
	var (
		key  = ipReqKey(ip)
		conn = dao.redis.Get(c)
	)
	defer conn.Close()
	if res, err = redis.Int(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			log.Warn("CacheIPRequestCheck(%s) return nil", key)
		} else {
			log.Error("CacheIPRequestCheck conn.Do(GET key(%v)) error(%v)", key, err)
		}
	}
	return
}

func (dao *Dao) AddCacheIPRequestCheck(c context.Context, ip string, val int) (err error) {
	var (
		key  = ipReqKey(ip)
		conn = dao.redis.Get(c)
	)
	defer conn.Close()
	if err = conn.Send("SETEX", key, dao.lotteryIPExpire, val); err != nil {
		log.Error("AddCacheIPRequestCheck conn.Send(SETEX, %s, %v, %d) error(%v)", key, dao.lotteryIPExpire, val, err)
	}
	return
}

func (dao *Dao) CacheLotteryTimes(c context.Context, sid int64, mid int64, lotteryTimesConfig []*l.LotteryTimesConfig, remark string) (list map[string]int, err error) {
	if len(lotteryTimesConfig) == 0 {
		return
	}
	conn := dao.redis.Get(c)
	defer conn.Close()
	key := lotteryTimesKey(sid, mid, remark)
	args := redis.Args{}.Add(key)
	for _, v := range lotteryTimesConfig {
		args = args.Add(lotteryTimesField(v))
	}
	var tmp []int
	if tmp, err = redis.Ints(conn.Do("HMGET", args...)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Error("CacheLotteryTimes redis.Ints(MGET) args(%v) error(%v)", args, err)
			return
		}
	}
	list = make(map[string]int, len(tmp))
	for i, value := range lotteryTimesConfig {
		list[lotteryTimesMapKey(value)] = tmp[i]
	}
	return
}

func (dao *Dao) AddCacheLotteryTimes(c context.Context, sid int64, mid int64, lotteryTimesConfig []*l.LotteryTimesConfig, remark string, list map[string]int) (err error) {
	if len(list) == 0 {
		return
	}
	var (
		keyID  string
		keyIDs []string
		conn   = dao.redis.Get(c)
		key    = lotteryTimesKey(sid, mid, remark)
		args   = redis.Args{}.Add(key)
	)
	defer conn.Close()
	for _, v := range lotteryTimesConfig {
		keyID = lotteryTimesField(v)
		keyIDs = append(keyIDs, keyID)
		args = args.Add(keyID).Add(list[lotteryTimesMapKey(v)])
	}
	if err = conn.Send("DEL", key); err != nil {
		log.Error("AddCacheLotteryTimes conn.Send(DEL, %s) error(%v)", key, err)
		return
	}
	if err = conn.Send("HMSET", args...); err != nil {
		log.Error("AddCacheLotteryTimes conn.Send(HMSET) error(%v)", err)
		return
	}
	if err = conn.Send("EXPIRE", key, dao.lotteryTimesExpire); err != nil {
		log.Error("conn.Send(EXPIRE, %s) error(%v)", key, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < 3; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("conn.Receive() error(%v)", err)
			return
		}
	}
	return
}

// IncrAddTimes /  IncrUsedTimes.
func (dao *Dao) IncrTimes(c context.Context, sid int64, mid int64, ltc *l.LotteryTimesConfig, val int, status string) (res int, err error) {
	var (
		key   = lotteryTimesKey(sid, mid, status)
		field = lotteryTimesField(ltc)
		conn  = dao.redis.Get(c)
	)
	defer conn.Close()
	if res, err = redis.Int(conn.Do("HINCRBY", key, field, val)); err != nil {
		err = errors.Wrap(err, "IncrTimes redis.Do(HINCRBY)")
		return
	}
	return
}

// getLotteryWinList
func (dao *Dao) CacheLotteryWinList(c context.Context, sid int64) (res []*l.GiftList, err error) {
	var (
		key  = lotteryWinListKey(sid)
		conn = dao.redis.Get(c)
		bs   []byte
	)
	defer conn.Close()
	if bs, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			log.Warn("CacheLotteryWinList(%s) return nil", key)
		} else {
			log.Error("CacheLotteryWinList conn.Do(GET key(%v)) error(%v)", key, err)
		}
		return
	}
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Error("json.Unmarshal(%v) error(%v)", bs, err)
	}
	return
}

// setLotteryWinList
func (dao *Dao) AddCacheLotteryWinList(c context.Context, sid int64, list []*l.GiftList) (err error) {
	var (
		key  = lotteryWinListKey(sid)
		conn = dao.redis.Get(c)
		bs   []byte
	)
	defer conn.Close()
	if bs, err = json.Marshal(list); err != nil {
		log.Error("json.Marshal(%v) error (%v)", list, err)
		return
	}
	if err = conn.Send("SETEX", key, dao.getExpireBeforeDawn(), bs); err != nil {
		log.Error("AddCacheLotteryWinList conn.Send(SETEX, %s, %v, %s) error(%v)", key, dao.lotteryWinListExpire, string(bs), err)
	}
	return
}

// get lottery log from redis
func (dao *Dao) CacheLotteryActionLog(c context.Context, sid int64, mid int64, start, end int64) (res []*l.LotteryRecordDetail, err error) {
	var (
		key  = lotteryActionKey(sid, mid)
		conn = dao.redis.Get(c)
	)
	defer conn.Close()
	values, err := redis.Values(conn.Do("ZREVRANGE", key, start, end))
	if err != nil {
		log.Error("CacheLotteryActionLog conn.Do(ZREVRANGE, %s) error(%v)", key, err)
		return
	}
	if len(values) == 0 {
		return
	}
	for len(values) > 0 {
		bs := []byte{}
		if values, err = redis.Scan(values, &bs); err != nil {
			log.Error("redis.Scan(%v) error(%v)", values, err)
			return
		}
		list := &l.LotteryRecordDetail{}
		if err = json.Unmarshal(bs, list); err != nil {
			log.Error("json.Unmarshal(%v) error(%v)", bs, err)
			return
		}
		res = append(res, list)
	}
	return
}

// DeleteLotteryWinLog ...
func (dao *Dao) DeleteLotteryWinLog(c context.Context, sid, mid int64) (err error) {
	var (
		key  = lotteryWinLogKey(sid, mid)
		conn = dao.redis.Get(c)
	)
	defer conn.Close()

	if _, err = conn.Do("DEL", key); err != nil {
		log.Errorc(c, "DeleteLotteryWinLog conn.Do(DEL, %s) error(%v)", key, err)
	}
	return
}

// set lottery log into redis
func (dao *Dao) AddCacheLotteryActionLog(c context.Context, sid int64, mid int64, list []*l.LotteryRecordDetail) (err error) {
	if len(list) == 0 {
		return
	}
	var (
		key  = lotteryActionKey(sid, mid)
		conn = dao.redis.Get(c)
		bs   []byte
	)
	defer conn.Close()
	if err = conn.Send("DEL", key); err != nil {
		log.Error("AddCacheLotteryActionLog conn.Send(DEL, %s) error(%v)", key, err)
		return
	}
	args := redis.Args{}.Add(key)
	for _, v := range list {
		if bs, err = json.Marshal(v); err != nil {
			log.Error("AddCacheLotteryActionLog json.Marshal() error(%v)", err)
			return
		}
		args = args.Add(v.Ctime).Add(bs)
	}
	if err = conn.Send("ZADD", args...); err != nil {
		log.Error("AddCacheLotteryActionLog conn.Send(ZADD, %s) error(%v)", key, err)
		return
	}
	if err = conn.Send("EXPIRE", key, dao.lotteryExpire); err != nil {
		log.Error("AddCacheLotteryActionLog conn.Send(EXPIRE,%s) error(%v)", key, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < 3; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("conn.Receive() error(%v)", err)
			return
		}
	}
	return
}

// add lottery log into redis
func (dao *Dao) AddLotteryActionLog(c context.Context, sid int64, mid int64, list []*l.LotteryRecordDetail) (err error) {
	if len(list) == 0 {
		return
	}
	var (
		key  = lotteryActionKey(sid, mid)
		conn = dao.redis.Get(c)
		bs   []byte
	)
	defer conn.Close()
	args := redis.Args{}.Add(key)
	for _, v := range list {
		lr := &l.LotteryRecordDetail{ID: v.ID, Mid: v.Mid, Num: v.Num, GiftID: v.GiftID, Type: v.Type, Ctime: v.Ctime, CID: v.CID}
		if bs, err = json.Marshal(lr); err != nil {
			log.Error("AddLotteryActionLog json.Marshal() error(%v)", err)
			return
		}
		args = args.Add(lr.Ctime).Add(bs)
	}
	if err = conn.Send("ZADD", args...); err != nil {
		log.Error("AddLotteryActionLog conn.Send(ZADD, %s) error(%v)", key, err)
		return
	}
	if err = conn.Send("EXPIRE", key, dao.lotteryExpire); err != nil {
		log.Error("AddLotteryActionLog conn.Send(EXPIRE,%s) error(%v)", key, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("conn.Receive() error(%v)", err)
			return
		}
	}
	return
}

func (dao *Dao) CacheGiftNum(c context.Context, sid int64, giftMap map[int64]*l.LotteryGift) (num map[int64]int64, err error) {
	var (
		key  = lotteryGiftNumKey(sid)
		conn = dao.redis.Get(c)
	)
	defer conn.Close()
	args := redis.Args{}.Add(key)
	list := make([]int64, 0, len(giftMap))
	for k := range giftMap {
		args = args.Add(k)
		list = append(list, k)
	}
	values, err := redis.Int64s(conn.Do("HMGET", args...))
	if err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		log.Error("CacheGiftNum conn.Do(HMGET %v) error(%v)", args, err)
		return
	}
	num = make(map[int64]int64, len(giftMap))
	for k, v := range values {
		num[list[k]] = v
	}
	return
}

func (dao *Dao) IncrGiftNum(c context.Context, sid int64, giftID int64) (err error) {
	var (
		key  = lotteryGiftNumKey(sid)
		conn = dao.redis.Get(c)
	)
	defer conn.Close()
	if err = conn.Send("HINCRBY", key, giftID, 1); err != nil {
		log.Error("HINCRBY conn.Send(key:%s,field:%d,value:%d) error(%v)", key, giftID, 1, err)
		return
	}
	if err = conn.Send("EXPIRE", key, dao.lotteryExpire); err != nil {
		log.Error("EXPIRE conn.Send() error(%v)", err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("conn.Receive() error(%v)", err)
			return
		}
	}
	return
}

// GetLotteryPlayWindowBuvid .
func (dao *Dao) GetLotteryPlayWindowBuvid(c context.Context, buvid string) (res bool, err error) {
	var (
		key  = lotteryPlayWindowBuvidKey(buvid)
		conn = dao.redis.Get(c)
	)
	defer conn.Close()
	reply, err := conn.Do("GET", key)
	if err != nil {
		if err == redis.ErrNil {
			return false, nil
		}
		log.Errorc(c, "GetLotteryPlayWindowBuvid cache error %+v", err)
		return false, err
	}
	return reply != nil, nil
}

// AddLotteryPlayWindowBuvid .
func (dao *Dao) AddLotteryPlayWindowBuvid(c context.Context, buvid string, expire int64) error {
	var (
		key  = lotteryPlayWindowBuvidKey(buvid)
		conn = dao.redis.Get(c)
	)
	defer conn.Close()
	_, err := conn.Do("SET", key, 1, "EX", expire, "NX")
	if err != nil {
		log.Errorc(c, "AddLotteryPlayWindowBuvid cache error %+v", err)
	}
	return err
}

// GetLotteryPlayWindowMid .
func (dao *Dao) GetLotteryPlayWindowMid(c context.Context, mid int64) (res bool, err error) {
	var (
		key  = lotteryPlayWindowMidKey(mid)
		conn = dao.redis.Get(c)
	)
	defer conn.Close()
	reply, err := conn.Do("GET", key)
	if err != nil {
		if err == redis.ErrNil {
			return false, nil
		}
		log.Errorc(c, "GetLotteryPlayWindowMid cache error %+v", err)
		return false, err
	}
	return reply != nil, nil
}

// AddLotteryPlayWindowMid .
func (dao *Dao) AddLotteryPlayWindowMid(c context.Context, mid, expire int64) error {
	var (
		key  = lotteryPlayWindowMidKey(mid)
		conn = dao.redis.Get(c)
	)
	defer conn.Close()
	_, err := conn.Do("SET", key, 1, "EX", expire, "NX")
	if err != nil {
		log.Errorc(c, "AddLotteryPlayWindowMid cache error %+v", err)
	}
	return err
}
