package rank

import (
	"context"
	"encoding/json"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/model/rank"
)

// GetRank 设置排名
func (d *dao) GetRank(c context.Context, rankActivityKey string) (res []*rank.Redis, err error) {
	var (
		bs   []byte
		key  = buildKey(rankActivityKey, rankKey)
		conn = d.redis.Get(c)
	)
	res = make([]*rank.Redis, 0)
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

// GetMidRank 获取用户维度排行
func (d *dao) GetMidRank(c context.Context, rankActivityKey string, mid int64) (res *rank.Redis, err error) {
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
