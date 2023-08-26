package dao

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"

	"go-common/library/cache/redis"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"

	"go-gateway/app/web-svr/space/interface/model"
)

const (
	_creativeViewDataURI = "/x/internal/creative/data/up/view/stats"
)

func (d *Dao) CreativeViewDataCache(c context.Context, mid int64) (*model.CreativeView, error) {
	var (
		key  = keyCreativeViewData(mid)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	value, err := redis.Bytes(conn.Do("GET", key))
	if err != nil {
		if err == redis.ErrNil {
			return nil, nil
		}
		log.Error("conn.Do(GET, %s) error(%v)", key, err)
		return nil, err
	}
	data := &model.CreativeView{}
	if err = json.Unmarshal(value, &data); err != nil {
		log.Error("json.Unmarshal(%v) error(%v)", value, err)
		return nil, err
	}
	return data, nil
}

func (d *Dao) SetCreativeViewDataCache(c context.Context, mid int64, data *model.CreativeView) error {
	var (
		bs   []byte
		key  = keyCreativeViewData(mid)
		conn = d.redis.Get(c)
		err  error
	)
	defer conn.Close()
	if bs, err = json.Marshal(data); err != nil {
		log.Error("json.Marshal(%v) error (%v)", data, err)
		return err
	}
	return setKvCache(conn, key, bs, d.getCacheExpire(d.redisMinExpire, d.redisMaxExpire))
}

func (d *Dao) CreativeViewData(c context.Context, mid int64) (*model.CreativeView, error) {
	var (
		params = url.Values{}
		ip     = metadata.String(c, metadata.RemoteIP)
	)
	params.Set("mid", strconv.FormatInt(mid, 10))
	var res struct {
		Code int                 `json:"code"`
		Data *model.CreativeView `json:"data"`
	}
	if err := d.httpR.Get(c, d.creativeViewDataURL, ip, params, &res); err != nil {
		log.Error("d.httpR.Get(%s) error(%v)", d.creativeViewDataURL, err)
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		log.Error("d.httpR.Get(%s) code(%d) error", d.creativeViewDataURL, res.Code)
		return nil, ecode.Int(res.Code)
	}
	if res.Data == nil {
		log.Warn("d.CreativeViewData empty url:%s, mid:%d", d.creativeViewDataURL, mid)
	}
	return res.Data, nil
}
