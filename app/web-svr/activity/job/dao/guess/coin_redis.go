package guess

import (
	"context"
	"fmt"

	"go-common/library/cache/redis"
	"go-common/library/log"
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

func (d *Dao) IncrUserCoinCache(c context.Context, mid int64, coins float64) (err error) {
	key := userCoinKey(mid)
	conn := d.guRedis.Get(c)
	defer conn.Close()
	if err = conn.Send("INCRBY", key, int64(coins*_multi)); err != nil {
		log.Error("IncrUserCoinCache conn.Do(INCRBY, %d, %v) error(%+v)", mid, coins, err)
		return
	}
	if err = conn.Send("EXPIRE", key, d.userExpire); err != nil {
		log.Error("IncrUserCoinCache conn.Do(EXPIRE, %d, %v) error(%+v)", mid, coins, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("conn.Flush() error(%+v)", err)
		return
	}
	if _, err = conn.Receive(); err != nil {
		log.Error("conn.Receive() error(%+v)", err)
		return
	}
	if _, err = conn.Receive(); err != nil {
		log.Error("conn.Receive() error(%+v)", err)
		return
	}
	return
}

func (d *Dao) ExpireUserListCache(c context.Context, mid int64) (ok bool, err error) {
	key := userListKey(mid)
	conn := d.guRedis.Get(c)
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

func (d *Dao) AddUserListCache(c context.Context, mid, contestID int64, count float64) (err error) {
	key := userListKey(mid)
	expireTime := 86400 * 30 * 6
	conn := d.guRedis.Get(c)
	defer conn.Close()
	if _, err = conn.Do("HSET", key, contestID, int64(count*_multi)); err != nil {
		log.Error("AddUserListCache conn.Do(HSET, %d, %v, %d) error(%+v)", mid, contestID, count, err)
	}
	if _, err = conn.Do("EXPIRE", key, expireTime); err != nil {
		log.Error("AddUserListCache conn.Do(EXPIRE, %s, %d) error(%+v)", key, expireTime, err)
		return
	}
	return
}

func (d *Dao) BatchAddUserListCache(c context.Context, mid int64, m map[int64]float64) (err error) {
	if len(m) == 0 {
		return nil
	}

	key := userListKey(mid)
	args := redis.Args{}.Add(key)
	conn := d.guRedis.Get(c)

	for k, v := range m {
		args = args.Add(k).Add(int64(v * _multi))
	}

	defer func() {
		_ = conn.Close()
	}()

	_, err = conn.Do("HMSET", args...)

	return
}

func (d *Dao) ResetUserCoinCache(c context.Context, mid int64, coins float64) (err error) {
	if coins == 0 {
		return nil
	}

	key := userCoinKey(mid)
	conn := d.guRedis.Get(c)
	defer conn.Close()
	if err = conn.Send("SET", key, int64(coins*_multi)); err != nil {
		log.Error("IncrUserCoinCache conn.Do(INCRBY, %d, %v) error(%+v)", mid, coins, err)
		return
	}
	if err = conn.Send("EXPIRE", key, d.userExpire); err != nil {
		log.Error("IncrUserCoinCache conn.Do(EXPIRE, %d, %v) error(%+v)", mid, coins, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("conn.Flush() error(%+v)", err)
		return
	}
	if _, err = conn.Receive(); err != nil {
		log.Error("conn.Receive() error(%+v)", err)
		return
	}
	if _, err = conn.Receive(); err != nil {
		log.Error("conn.Receive() error(%+v)", err)
		return
	}
	return
}
