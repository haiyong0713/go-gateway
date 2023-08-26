package rank

import (
	"context"
	"encoding/json"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/job/model/rank"

	"github.com/pkg/errors"
)

const (
	rankExpire = 2592000
)

// SetRank 设置排名
func (d *dao) SetRank(c context.Context, rankNameKey string, rankBatch []*rank.Redis) (err error) {
	if len(rankBatch) == 0 {
		return
	}
	var (
		bs   []byte
		key  = buildKey(rankNameKey, rankKey)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if bs, err = json.Marshal(rankBatch); err != nil {
		log.Errorc(c, "json.Marshal() error(%v)", err)
		return
	}
	if _, err = conn.Do("SETEX", key, rankExpire, bs); err != nil {
		log.Errorc(c, "SetRank conn.Do(SET) key(%s) error(%v)", key, err)
	}
	return
}

// GetRank 获取排名
func (d *dao) GetRank(c context.Context, rankNameKey string) (res []*rank.Redis, err error) {
	var (
		bs   []byte
		key  = buildKey(rankNameKey, rankKey)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if bs, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Errorc(c, "conn.Do(GET,%s) error(%v)", key, err)
		}
		return
	}
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Errorc(c, "json.Unmarshal(%s) error(%v)", string(bs), err)
	}
	return
}

// SetMidRank 设置用户维度的排行榜结果
func (d *dao) SetMidRank(c context.Context, rankNameKey string, midRank []*rank.Redis) (err error) {
	if len(midRank) == 0 {
		return
	}
	conn := d.redis.Get(c)
	defer conn.Close()
	args := redis.Args{}
	keys := make([]string, 0)
	for _, v := range midRank {
		var bs []byte
		if bs, err = json.Marshal(v); err != nil {
			log.Error("SetMidRank json.Marshal() error(%v)", err)
			return
		}
		args = args.Add(buildKey(rankNameKey, midKey, v.Mid)).Add(bs)
		keys = append(keys, buildKey(rankNameKey, midKey, v.Mid))
	}
	if err := conn.Send("MSET", args...); err != nil {
		err = errors.Wrap(err, "SetMidRank conn.Do(MSET)")
	}
	count := 1
	for _, v := range keys {
		if err := conn.Send("EXPIRE", v, d.dataExpire); err != nil {
			log.Errorc(c, "SetMidRank conn.Send(Expire, %s, %d) error(%v)", v, d.dataExpire, err)
			return err
		}
		count++
	}
	if err1 := conn.Flush(); err1 != nil {
		log.Errorc(c, "SetMidRank Flush error(%v)", err1)
		return err
	}
	for i := 0; i < count; i++ {
		if _, err2 := conn.Receive(); err2 != nil {
			log.Errorc(c, "SetMidRank conn.Receive() error(%v)", err2)
			return err2
		}
	}
	return

}
