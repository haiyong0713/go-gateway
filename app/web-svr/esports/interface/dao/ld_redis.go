package dao

import (
	"context"
	"encoding/json"
	"fmt"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/web-svr/esports/interface/model"
)

const (
	_keyLolGame  = "l_g_%d"
	_keyDotaGame = "d_g_%d"
	_keyOwGame   = "o_g_%d"
)

func keyLolGame(id int64) string {
	return fmt.Sprintf(_keyLolGame, id)
}

func keyDotaGame(id int64) string {
	return fmt.Sprintf(_keyDotaGame, id)
}

func keyOwGame(id int64) string {
	return fmt.Sprintf(_keyOwGame, id)
}

// CacheLolGames.
func (d *Dao) CacheLolGames(c context.Context, matchID int64) (res []*model.LolGame, err error) {
	key := keyLolGame(matchID)
	res, err = d.commonGames(c, key)
	return
}

// CacheDotaGames.
func (d *Dao) CacheDotaGames(c context.Context, matchID int64) (res []*model.LolGame, err error) {
	key := keyDotaGame(matchID)
	res, err = d.commonGames(c, key)
	return
}

func (d *Dao) commonGames(c context.Context, key string) (res []*model.LolGame, err error) {
	var (
		bs   []byte
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if bs, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Error("conn.Do(GET,%s) error(%v)", key, err)
		}
		return
	}
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Error("json.Unmarshal(%s) error(%v)", string(bs), err)
	}
	return
}

// CacheOwGames.
func (d *Dao) CacheOwGames(c context.Context, matchID int64) (res []*model.OwGame, err error) {
	var (
		bs   []byte
		key  = keyOwGame(matchID)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if bs, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Error("conn.Do(GET,%s) error(%v)", key, err)
		}
		return
	}
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Error("json.Unmarshal(%s) error(%v)", string(bs), err)
	}
	return
}

// AddCacheLolGames.
func (d *Dao) AddCacheLolGames(c context.Context, matchID int64, data []*model.LolGame) (err error) {
	key := keyLolGame(matchID)
	err = d.addCommonGames(c, key, data)
	return
}

// AddCacheDotaGames.
func (d *Dao) AddCacheDotaGames(c context.Context, matchID int64, data []*model.LolGame) (err error) {
	key := keyDotaGame(matchID)
	err = d.addCommonGames(c, key, data)
	return
}

func (d *Dao) addCommonGames(c context.Context, key string, data []*model.LolGame) (err error) {
	var (
		bs   []byte
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if bs, err = json.Marshal(data); err != nil {
		log.Error("json.Marshal() error(%v)", err)
		return
	}
	if _, err = conn.Do("SETEX", key, d.listExpire, bs); err != nil {
		log.Error("conn.Do(SETEX, %s, %d, %d)", key, d.listExpire, bs)
	}
	return
}

// AddCacheOwGames.
func (d *Dao) AddCacheOwGames(c context.Context, matchID int64, data []*model.OwGame) (err error) {
	var (
		bs   []byte
		key  = keyOwGame(matchID)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if bs, err = json.Marshal(data); err != nil {
		log.Error("json.Marshal() error(%v)", err)
		return
	}
	if _, err = conn.Do("SETEX", key, d.listExpire, bs); err != nil {
		log.Error("conn.Do(SETEX, %s, %d, %d)", key, d.listExpire, bs)
	}
	return
}
