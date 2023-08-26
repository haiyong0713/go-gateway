package bws

import (
	"context"
	"fmt"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/ecode"
	"go-common/library/log"
	bwsmdl "go-gateway/app/web-svr/activity/interface/model/bws"
)

const (
	_bwsCatchUp = "bws_c_up_%d_%s"
)

func bwsCatchUpKey(bid int64, key string) string {
	return fmt.Sprintf(_bwsCatchUp, bid, key)
}

func (d *Dao) CatchUps(c context.Context, bid int64, userkey string, start, end int) ([]*bwsmdl.CatchUser, int, error) {
	res, err := d.catchUpsCache(c, bid, userkey, 0, -1)
	if err != nil {
		log.Error("d.catchUpCache error(%v)", err)
	}
	if len(res) > 0 {
		return res, len(res), nil
	}
	if len(res) == 0 {
		// 回源该用户所有捕获到的数据
		res, err = d.CatchUser(c, bid, userkey)
		if err != nil {
			log.Error("d.CatchUser error(%v)", err)
			return nil, 0, err
		}
		if err = d.addCatchUpsCache(c, bid, userkey, res); err != nil {
			log.Error("d.addCatchUpsCache error(%v)", err)
			return nil, 0, err
		}
	}
	count := len(res)
	if end < 0 {
		return res, count, nil
	}
	if start > len(res) {
		return nil, 0, ecode.RequestErr
	}
	if end < len(res) {
		return res[start:end], count, nil
	}
	return res[start:], count, nil
}

func (d *Dao) AddCatchUps(c context.Context, bid int64, userkey string, bups []*bwsmdl.CatchUser) error {
	if err := d.addCatchUpsCache(c, bid, userkey, bups); err != nil {
		log.Error("d.addCatchUpsCache error(%v)", err)
		return err
	}
	for _, m := range bups {
		tmp := &bwsmdl.CatchUser{}
		*tmp = *m
		d.addCache(func() {
			d.InCatchUser(context.TODO(), bid, tmp.Mid, tmp.Key)
		})
	}
	return nil
}

func (d *Dao) addCatchUpsCache(c context.Context, bid int64, userkey string, miss []*bwsmdl.CatchUser) error {
	if len(miss) == 0 {
		return ecode.RequestErr
	}
	var count int
	cacheKey := bwsCatchUpKey(bid, userkey)
	conn := d.redis.Get(c)
	defer conn.Close()
	now := time.Now().Unix()
	args := redis.Args{}.Add(cacheKey)
	for i, v := range miss {
		if v == nil {
			continue
		}
		score := now + int64(len(miss)-i)
		item, err := v.Marshal()
		if err != nil {
			log.Error("up.Marshal error(%v)", err)
			return err
		}
		args = args.Add(score).Add(item)
	}
	if err := conn.Send("ZADD", args...); err != nil {
		log.Error("redis.Bool error(%v)", err)
		return err
	}
	count++
	if err := conn.Send("EXPIRE", cacheKey, d.bluetoothExpire); err != nil {
		log.Error("AppendUserPointsCache conn.Send(Expire, %s) error(%v)", cacheKey, err)
		return err
	}
	count++
	if err := conn.Flush(); err != nil {
		log.Error("conn.Flush error(%v)", err)
		return err
	}
	for i := 0; i < count; i++ {
		if _, err := conn.Receive(); err != nil {
			log.Error("conn.Receive() error(%v)", err)
			return err
		}
	}
	return nil
}

// catchUpCache .
func (d *Dao) catchUpsCache(c context.Context, bid int64, userkey string, pn, ps int) ([]*bwsmdl.CatchUser, error) {
	conn := d.redis.Get(c)
	defer conn.Close()
	cacheKey := bwsCatchUpKey(bid, userkey)
	bss, err := redis.ByteSlices(conn.Do("ZREVRANGE", cacheKey, pn, ps))
	if err != nil {
		log.Error("conn.Do key(%s) error(%v)", cacheKey, err)
		return nil, err
	}
	var res []*bwsmdl.CatchUser
	for _, bs := range bss {
		if bss == nil {
			continue
		}
		b := &bwsmdl.CatchUser{}
		if err = b.Unmarshal(bs); err != nil {
			log.Error("upBluetoothCache Unmarshal error(%v)", err)
			continue
		}
		res = append(res, b)
	}
	return res, nil
}
