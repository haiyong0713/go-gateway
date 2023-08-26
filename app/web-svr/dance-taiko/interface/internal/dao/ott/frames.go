package ott

import (
	"context"
	"encoding/json"
	"fmt"
	"go-common/library/log"

	"go-common/library/cache"
	"go-common/library/cache/redis"
	"go-gateway/app/web-svr/dance-taiko/interface/internal/model"

	"github.com/pkg/errors"
)

const (
	_selectKeyFrames = "SELECT `aid`, `cid`, `key_frames`, `url` FROM `dance_key_frames` WHERE `cid`=? AND `is_deleted`=0"

	cacheKeyFramesKey = "key_frames_%d"
)

// LoadFrames get data from cache if miss will call source method, then add to cache.
func (d *dao) LoadFrames(c context.Context, cid int64) (res *model.FramesCache, err error) {
	addCache := true
	res, err = d.cacheKeyFrames(c, cid)
	if err != nil {
		addCache = false
		err = nil
	}
	defer func() {
		if res != nil && res.Aid == -1 {
			res = nil
		}
	}()
	if res != nil {
		cache.MetricHits.Inc("bts:LoadFrames")
		return
	}
	cache.MetricMisses.Inc("bts:LoadFrames")
	res, err = d.rawKeyFrames(c, cid)
	if err != nil {
		return
	}
	miss := res
	if miss == nil {
		miss = &model.FramesCache{Aid: -1}
	}
	if !addCache {
		return
	}
	d.cache.Do(c, func(c context.Context) {
		if err := d.addCacheKeyFrames(c, cid, miss); err != nil {
			log.Error("LoadFrames frames(%v) err(%v)", miss, err)
		}
	})
	return
}

func (d *dao) cacheKeyFrames(c context.Context, cid int64) (*model.FramesCache, error) {
	var (
		conn = d.redis.Get(c)
		key  = fmt.Sprintf(cacheKeyFramesKey, cid)
	)
	defer conn.Close()
	reply, err := redis.Bytes(conn.Do("GET", key))
	if err != nil {
		if err == redis.ErrNil {
			return nil, nil
		}
		return nil, errors.Wrapf(err, "CacheKeyFrames cid(%d)", cid)
	}
	res := new(model.FramesCache)
	if err = json.Unmarshal(reply, &res); err != nil {
		return nil, errors.Wrapf(err, "CacheKeyFrames cid(%d)", cid)
	}
	return res, nil
}

func (d *dao) addCacheKeyFrames(c context.Context, cid int64, frames *model.FramesCache) error {
	var (
		conn = d.redis.Get(c)
		key  = fmt.Sprintf(cacheKeyFramesKey, cid)
	)
	defer conn.Close()
	data, err := json.Marshal(frames)
	if err != nil {
		return errors.Wrapf(err, "AddCacheKeyFrames cid(%d) data(%v)", cid, frames)
	}
	if _, err := conn.Do("SET", key, data, "EX", d.framesExipre); err != nil {
		return errors.Wrapf(err, "AddCacheKeyFrames cid(%d) data(%v)", cid, data)
	}
	return nil
}

func (d *dao) rawKeyFrames(c context.Context, cid int64) (*model.FramesCache, error) {
	row := d.db.QueryRow(c, _selectKeyFrames, cid)
	var res = new(model.FramesCache)
	if err := row.Scan(&res.Aid, &res.Cid, &res.KeyFrames, &res.Url); err != nil {
		return nil, errors.Wrapf(err, "RawKeyFrames cid(%d)", cid)
	}
	return res, nil
}
