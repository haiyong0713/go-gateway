package rank

import (
	"context"
	"encoding/json"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/model/rank"
)

// GetRank 设置排名
func (d *dao) GetRank(c context.Context, rankActivityKey string) (res []*rank.MidRank, err error) {
	var (
		bs   []byte
		key  = buildKey(rankActivityKey, rankKey)
		conn = d.redis.Get(c)
	)
	res = make([]*rank.MidRank, 0)
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

// SetRankConfig ...
func (d *dao) SetRankConfig(c context.Context, sid int64, sidSource int, rankConfig *rank.Rank) (err error) {
	var (
		key  = buildKey(configKey, sid, sidSource)
		conn = d.redis.Get(c)
		bs   []byte
	)
	defer conn.Close()
	if bs, err = json.Marshal(rankConfig); err != nil {
		log.Error("json.Marshal(%v) error (%v)", rankConfig, err)
		return
	}
	if err = conn.Send("SETEX", key, d.rankConfigExpire, bs); err != nil {
		log.Error("conn.Send(SETEX, %s, %v, %s) error(%v)", key, d.rankConfigExpire, string(bs), err)
	}
	return
}

// GetRankConfig ...
func (d *dao) GetRankConfig(c context.Context, sid int64, sidSource int) (rankConfig *rank.Rank, err error) {
	var (
		bs   []byte
		key  = buildKey(configKey, sid, sidSource)
		conn = d.redis.Get(c)
	)
	rankConfig = &rank.Rank{}
	defer conn.Close()
	if bs, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Error("conn.Do(GET,%s) error(%v)", key, err)
		}
		return
	}
	if err = json.Unmarshal(bs, &rankConfig); err != nil {
		log.Error("json.Unmarshal(%s) error(%v)", string(bs), err)
	}
	return
}

// GetMidRank 获取用户维度排行
func (d *dao) GetMidRank(c context.Context, rankActivityKey string, mid int64) (res *rank.MidRank, err error) {
	var (
		bs   []byte
		key  = buildKey(rankActivityKey, midKey, mid)
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
