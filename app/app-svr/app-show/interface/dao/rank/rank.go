package rank

import (
	"context"
	"encoding/json"
	"fmt"
	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-show/interface/conf"
	"go-gateway/app/app-svr/app-show/interface/model/rank"
)

const (
	_rank = "rank_key_v2_%s_%d"
)

func keyRank(key string, rid int) string {
	return fmt.Sprintf(_rank, key, rid)
}

// Dao is rank dao.
type Dao struct {
	// redis
	redis *redis.Pool
}

// New rank dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		// redis
		redis: redis.NewPool(c.Redis.Recommend.Config),
	}
	return
}

// RankCache is
func (d *Dao) RankCache(c context.Context, order string, rid, start, end int) (list []*rank.List, err error) {
	conn := d.redis.Get(c)
	defer conn.Close()
	key := keyRank(order, rid)
	values, err := redis.Values(conn.Do("ZREVRANGE", key, start, end))
	if err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		log.Error("redis (ZREVRANGE,%s,%d,%d) error(%v)", key, start, end, err)
		return
	}
	if len(values) == 0 {
		return
	}
	list = make([]*rank.List, 0, len(values))
	for len(values) > 0 {
		bs := []byte{}
		if values, err = redis.Scan(values, &bs); err != nil {
			log.Error("redis.Scan vs(%v) error(%v)", values, err)
			return
		}
		l := &rank.List{}
		if err = json.Unmarshal([]byte(bs), &l); err != nil {
			log.Error("%+v", err)
			return
		}
		list = append(list, l)
	}
	return
}
