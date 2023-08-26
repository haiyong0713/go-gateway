package vote

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"go-gateway/app/web-svr/activity/interface/api"
	"io"

	"go-common/library/cache/redis"
	"go-common/library/log"
)

// CacheActivity get data from redis
func (d *Dao) CacheActivity(c context.Context, id int64) (res *api.VoteActivity, err error) {
	key := redisActivityConfigCacheKey(id)
	reply, err1 := redis.Bytes(d.redis.Do(c, "GET", key))
	if err1 != nil {
		if err1 == redis.ErrNil {
			return
		}
		err = err1
		log.Errorc(c, "d.CacheActivity(get key: %v) err: %+v", key, err)
		return
	}
	res = &api.VoteActivity{}
	gzipReader, err1 := gzip.NewReader(bytes.NewReader(reply))
	if err1 != nil {
		err = err1
		log.Errorc(c, "d.CacheActivity(get key: %v) err: %+v", key, err)
		return
	}
	defer gzipReader.Close()
	var buf bytes.Buffer
	io.Copy(&buf, gzipReader)
	err = json.Unmarshal(buf.Bytes(), res)
	if err != nil {
		log.Errorc(c, "d.CacheActivity(get key: %v) err: %+v", key, err)
		return
	}
	return
}

// AddCacheActivity Set data to redis
func (d *Dao) AddCacheActivity(c context.Context, id int64, val *api.VoteActivity) (err error) {
	if val == nil {
		return
	}
	key := redisActivityConfigCacheKey(id)
	var bs []byte
	bs, err = json.Marshal(val)
	if err != nil {
		log.Errorc(c, "d.AddCacheActivity(get key: %v) err: %+v", key, err)
		return
	}
	var wriBuf bytes.Buffer
	gw := gzip.NewWriter(&wriBuf)
	_, err = gw.Write(bs)
	if err != nil {
		log.Errorc(c, "d.AddCacheActivity(get key: %v) err: %+v", key, err)
		return
	}
	gw.Close()
	bs = wriBuf.Bytes()
	expire := d.activityCacheExpire
	if _, err = d.redis.Do(c, "set", key, bs, "EX", expire); err != nil {
		log.Errorc(c, "d.AddCacheActivity(get key: %v) err: %+v", key, err)
		return
	}
	return
}

// CacheDataSourceGroup get data from redis
func (d *Dao) CacheDataSourceGroup(c context.Context, id int64) (res *api.VoteDataSourceGroupItem, err error) {
	key := redisActivityDSGCacheKey(id)
	reply, err1 := redis.Bytes(d.redis.Do(c, "GET", key))
	if err1 != nil {
		if err1 == redis.ErrNil {
			return
		}
		err = err1
		log.Errorc(c, "d.CacheActivity(get key: %v) err: %+v", key, err)
		return
	}
	res = &api.VoteDataSourceGroupItem{}
	gzipReader, err1 := gzip.NewReader(bytes.NewReader(reply))
	if err1 != nil {
		err = err1
		log.Errorc(c, "d.CacheActivity(get key: %v) err: %+v", key, err)
		return
	}
	defer gzipReader.Close()
	var buf bytes.Buffer
	io.Copy(&buf, gzipReader)
	err = json.Unmarshal(buf.Bytes(), res)
	if err != nil {
		log.Errorc(c, "d.CacheActivity(get key: %v) err: %+v", key, err)
		return
	}
	return
}

// AddCacheDataSourceGroup Set data to redis
func (d *Dao) AddCacheDataSourceGroup(c context.Context, id int64, val *api.VoteDataSourceGroupItem) (err error) {
	if val == nil {
		return
	}
	key := redisActivityDSGCacheKey(id)
	var bs []byte
	bs, err = json.Marshal(val)
	if err != nil {
		log.Errorc(c, "d.AddCacheActivity(get key: %v) err: %+v", key, err)
		return
	}
	var wriBuf bytes.Buffer
	gw := gzip.NewWriter(&wriBuf)
	_, err = gw.Write(bs)
	if err != nil {
		log.Errorc(c, "d.AddCacheActivity(get key: %v) err: %+v", key, err)
		return
	}
	gw.Close()
	bs = wriBuf.Bytes()
	expire := d.activityCacheExpire
	if _, err = d.redis.Do(c, "set", key, bs, "EX", expire); err != nil {
		log.Errorc(c, "d.AddCacheActivity(get key: %v) err: %+v", key, err)
		return
	}
	return
}
