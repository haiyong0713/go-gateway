package dynamic

import (
	"context"
	"encoding/json"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-job/job/conf"
	"go-gateway/app/app-svr/app-job/job/model/recommend"
)

const (
	_schoolKey = "dc_key"
)

type Dao struct {
	// redis
	redis *redis.Pool
}

func New(c *conf.Config) *Dao {
	d := &Dao{
		redis: redis.NewPool(c.Redis.DynamicSchool.Config),
	}
	return d
}

func (d *Dao) AddSchoolCache(c context.Context, list []*recommend.Item) error {
	var count int
	conn := d.redis.Get(c)
	defer conn.Close()
	now := time.Now().Unix()
	for i, l := range list {
		bs, err := json.Marshal(l)
		if err != nil {
			log.Error("json.Marshal error(%v)", err)
			return err
		}
		score := now + int64(len(list)-i)
		if err := conn.Send("ZADD", _schoolKey, score, string(bs)); err != nil {
			log.Error("conn.Send(ZADD, %s) error(%v)", _schoolKey, err)
			return err
		}
		count++
	}
	if err := conn.Send("ZREMRANGEBYSCORE", _schoolKey, "-inf", now); err != nil {
		log.Error("conn.Send(ZREMRANGEBYSCORE) error(%v)", err)
		return err
	}
	count++
	if err := conn.Flush(); err != nil {
		log.Error("conn.Flush error(%v)", err)
		return err
	}
	for i := 0; i < count; i++ {
		if _, err := conn.Receive(); err != nil {
			log.Error("conn.Receive() error(%v)", err)
			return err
		}
	}
	return nil
}
