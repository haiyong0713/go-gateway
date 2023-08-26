package rank

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-job/job/model/rank"
)

const (
	_rank = "rank_key_v2_%s_%d"
)

func keyRank(key string, rid int) string {
	return fmt.Sprintf(_rank, key, rid)
}

func (d *Dao) AddRankCache(c context.Context, order string, rid int, list []*rank.List) error {
	var count int
	conn := d.redis.Get(c)
	defer conn.Close()
	key := keyRank(order, rid)
	now := time.Now().Unix()
	for i, l := range list {
		bs, err := json.Marshal(l)
		if err != nil {
			log.Error("json.Marshal error(%v)", err)
			return err
		}
		score := now + int64(len(list)-i)
		if err := conn.Send("ZADD", key, score, string(bs)); err != nil {
			log.Error("conn.Send(ZADD, %s) error(%v)", key, err)
			return err
		}
		count++
	}
	if err := conn.Send("ZREMRANGEBYSCORE", key, "-inf", now); err != nil {
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
