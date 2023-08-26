package like

import (
	"context"
	"encoding/json"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/model/like"
)

const (
	_fateSwitchKey = "fate_switch_total"
	_fateConfKey   = "fate_conf"
)

// FateInfoCache .
func (d *Dao) FateInfoCache(c context.Context, key string) (data int64, err error) {
	var conn = d.redis.Get(c)
	defer conn.Close()
	if data, err = redis.Int64(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Error("FateInfoCache conn.Do(GET, key(%s)) error(%v)", key, err)
		}
	}
	return
}

// FateSwitchCache .
func (d *Dao) FateSwitchCache(c context.Context) (data *like.FateSwitch, err error) {
	var (
		conn = d.redis.Get(c)
		bs   []byte
	)
	defer conn.Close()
	if bs, err = redis.Bytes(conn.Do("GET", _fateSwitchKey)); err != nil {
		if err == redis.ErrNil {
			data = nil
			err = nil
		} else {
			log.Error("FateSwitchCache conn.Do(GET key(%v)) error(%v)", _fateSwitchKey, err)
		}
		return
	}
	data = new(like.FateSwitch)
	if err = json.Unmarshal(bs, &data); err != nil {
		log.Error("FateSwitchCache json.Unmarshal(%s) error(%v)", string(bs), err)
	}
	return
}

// FateConfCache .
func (d *Dao) FateConfCache(c context.Context) (data *like.FateConfData, err error) {
	var (
		conn = d.redis.Get(c)
		bs   []byte
	)
	defer conn.Close()
	if bs, err = redis.Bytes(conn.Do("GET", _fateConfKey)); err != nil {
		if err == redis.ErrNil {
			data = nil
			err = nil
		} else {
			log.Error("FateConfCache conn.Do(GET key(%v)) error(%v)", _fateConfKey, err)
		}
		return
	}
	data = new(like.FateConfData)
	if err = json.Unmarshal(bs, &data); err != nil {
		log.Error("FateConfCache json.Unmarshal(%s) error(%v)", string(bs), err)
	}
	return
}
