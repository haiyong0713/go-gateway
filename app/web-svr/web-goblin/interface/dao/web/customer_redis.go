package web

import (
	"context"
	"encoding/json"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/web-svr/web-goblin/interface/model/web"
)

const _customerKey = "c_b"

// CustomerCache get customer cache.
func (d *Dao) CustomerCache(c context.Context) (rs map[string]*web.CustomerCenter, err error) {
	conn := d.redis.Get(c)
	defer conn.Close()
	var values []byte
	if values, err = redis.Bytes(conn.Do("GET", _customerKey)); err != nil {
		if err == redis.ErrNil {
			err = nil
			log.Warn("CustomerCache redis (%s) return nil ", _customerKey)
		} else {
			log.Error("CustomerCache conn.Do(GET,%s) error(%v)", _customerKey, err)
		}
		return
	}
	if err = json.Unmarshal(values, &rs); err != nil {
		log.Error("CustomerCache json.Unmarshal(%v) error(%v)", values, err)
	}
	return
}

// SetCustomerCache set online list bak cache.
func (d *Dao) SetCustomerCache(c context.Context, data map[string]*web.CustomerCenter) (err error) {
	conn := d.redis.Get(c)
	defer conn.Close()
	var bs []byte
	if bs, err = json.Marshal(data); err != nil {
		log.Error("SetCustomerCache json.Marshal(%v) error(%v)", data, err)
		return
	}
	if _, err = conn.Do("SETEX", _customerKey, d.cusExpire, bs); err != nil {
		log.Error("conn.Do(SETEX, %s, %d, %d)", _customerKey, d.cusExpire, bs)
		return
	}
	return
}
