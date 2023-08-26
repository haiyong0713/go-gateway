package dao

import (
	"context"
	"encoding/json"
	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/web-svr/web/interface/model"
)

const (
	_onlineListKey    = "cache_online_aids"
	_onlineListBakKey = _keyBakPrefix + _onlineListKey
)

func (d *Dao) OnlineListCache(ctx context.Context) ([]*model.OnlineAid, error) {
	key := _onlineListKey
	conn := d.redisBak.Get(ctx)
	defer conn.Close()
	reply, err := redis.Bytes(conn.Do("GET", key))
	if err != nil {
		return nil, err
	}
	var res []*model.OnlineAid
	if err := json.Unmarshal(reply, &res); err != nil {
		return nil, err
	}
	return res, nil
}

// OnlineListBakCache get online list bak cache.
func (d *Dao) OnlineListBakCache(c context.Context) (rs []*model.OnlineArc, err error) {
	conn := d.redisBak.Get(c)
	defer conn.Close()
	var values []byte
	if values, err = redis.Bytes(conn.Do("GET", _onlineListBakKey)); err != nil {
		if err == redis.ErrNil {
			err = nil
			log.Warn("OnlineListBakCache redis (%s) return nil ", _onlineListKey)
		} else {
			log.Error("OnlineListBakCache conn.Do(GET,%s) error(%v)", _onlineListKey, err)
		}
		return
	}
	if err = json.Unmarshal(values, &rs); err != nil {
		log.Error("OnlineListBakCache json.Unmarshal(%v) error(%v)", values, err)
	}
	return
}

// SetOnlineListBakCache set online list bak cache.
func (d *Dao) SetOnlineListBakCache(c context.Context, data []*model.OnlineArc) (err error) {
	conn := d.redisBak.Get(c)
	defer conn.Close()
	var bs []byte
	if bs, err = json.Marshal(data); err != nil {
		log.Error("SetOnlineListBakCache json.Marshal(%v) error(%v)", data, err)
		return
	}
	if err = conn.Send("SET", _onlineListBakKey, bs); err != nil {
		log.Error("SetOnlineListBakCache conn.Send(SET,%s,%s) error(%v)", _onlineListBakKey, string(bs), err)
		return
	}
	if err = conn.Send("EXPIRE", _onlineListBakKey, d.redisOlListBakExpire); err != nil {
		log.Error("SetOnlineListBakCache conn.Send(EXPIRE,%s,%d) error(%v)", _onlineListBakKey, d.redisOlListBakExpire, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("SetOnlineListBakCache conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("SetOnlineListBakCache conn.Recevie(%d) error(%v0", i, err)
		}
	}
	return
}
