package lol

import (
	"context"
	"fmt"
	"strconv"

	"go-common/library/cache/redis"
	"go-common/library/log"
	lolmdl "go-gateway/app/web-svr/activity/interface/model/lol"
)

const (
	_multi       = 100
	_userKey     = "uc_%d"
	_userListKey = "ucl_%d"
)

func userCoinKey(mid int64) string {
	return fmt.Sprintf(_userKey, mid)
}

func userListKey(mid int64) string {
	return fmt.Sprintf(_userListKey, mid)
}

func (d *Dao) ExpireUserCache(c context.Context, mid int64) (ok bool, err error) {
	key := userCoinKey(mid)
	conn := d.redis.Get(c)
	defer conn.Close()
	if ok, err = redis.Bool(conn.Do("EXPIRE", key, d.listExpire)); err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		log.Error("ExpireUserCache conn.Do(EXPIRE, %d) error(%+v)", mid, err)
	}
	return
}

func (d *Dao) UserCoinCache(c context.Context, mid int64) (res float64, err error) {
	key := userCoinKey(mid)
	conn := d.redis.Get(c)
	defer conn.Close()
	var coins int64
	if coins, err = redis.Int64(conn.Do("GET", key)); err != nil {
		log.Error("UserCoinCache conn.Do(GET, %d) error(%+v)", mid, err)
	}
	res, err = strconv.ParseFloat(fmt.Sprintf("%.1f", float64(coins)/_multi), 64)
	return
}

func (d *Dao) ExpireUserListCache(c context.Context, mid int64) (ok bool, err error) {
	key := userListKey(mid)
	conn := d.redis.Get(c)
	defer conn.Close()
	if ok, err = redis.Bool(conn.Do("EXPIRE", key, d.listExpire)); err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		log.Error("ExpireUserListCache conn.Do(EXPIRE, %d) error(%+v)", mid, err)
	}
	return
}

func (d *Dao) UserListCache(c context.Context, mid int64) (res []*lolmdl.ContestDetail, err error) {
	key := userListKey(mid)
	conn := d.redis.Get(c)
	defer conn.Close()
	var values []int64
	if values, err = redis.Int64s(conn.Do("HGETALL", key)); err != nil {
		log.Error("UserListCache conn.Do(HGETALL, %d) error(%+v)", mid, err)
		return
	}
	res = make([]*lolmdl.ContestDetail, 0, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		coinsAfterConvert, _ := strconv.ParseFloat(fmt.Sprintf("%.1f", float64(values[i+1])/_multi), 64)
		res = append(res, &lolmdl.ContestDetail{
			ContestID: values[i],
			Coins:     coinsAfterConvert,
		})
	}
	return
}

func (d *Dao) SetUserListCache(c context.Context, mid int64, lists []*lolmdl.ContestDetail) (err error) {
	key := userListKey(mid)
	conn := d.redis.Get(c)
	defer conn.Close()
	args := []interface{}{key}
	for _, list := range lists {
		args = append(args, list.ContestID)
		args = append(args, list.Coins)
	}
	if _, err = conn.Do("HMSET", args...); err != nil {
		log.Error("SetUserListCache conn.Do(HMSET, %d, %+v) error(%+v)", mid, lists, err)
	}
	return
}

func (d *Dao) ExistUserListCache(c context.Context, mid int64, contestIDs []int64) (res map[int64]bool, err error) {
	key := userListKey(mid)
	conn := d.redis.Get(c)
	defer conn.Close()
	for _, id := range contestIDs {
		if err = conn.Send("HEXISTS", key, id); err != nil {
			log.Error("conn.Send(HEXISTS %s) error(%+v)", key, err)
			return
		}
	}
	if err = conn.Flush(); err != nil {
		log.Error("conn.Flush() error(%+v)", err)
		return
	}
	res = make(map[int64]bool)
	for i := 0; i < len(contestIDs); i++ {
		var ok bool
		if ok, err = redis.Bool(conn.Receive()); err != nil {
			log.Error("conn.Receive() error(%+v)", err)
			return
		}
		res[contestIDs[i]] = ok
	}
	return
}
