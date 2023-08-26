package like

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/job/model/like"

	"github.com/pkg/errors"
)

func imageUserKey(sid int64, day string, typ int) string {
	return fmt.Sprintf("img_rk_%d_%s_%d", sid, day, typ)
}

func (d *Dao) SetImageUpCache(c context.Context, sid int64, day string, typ int, list []*like.ImageUp) (err error) {
	if len(list) == 0 {
		return
	}
	key := imageUserKey(sid, day, typ)
	conn := d.redis.Get(c)
	defer conn.Close()
	args := redis.Args{}.Add(key)
	for _, v := range list {
		args = args.Add(v.Score).Add(v.Mid)
	}
	if err = conn.Send("DEL", key); err != nil {
		log.Error("SetImageUpCache conn.Do(DEL %s) error(%v)", key, err)
		return
	}
	if err = conn.Send("ZADD", args...); err != nil {
		log.Error("SetImageUpCache conn.Do(ZADD %s) error(%v)", key, err)
		return
	}
	if err = conn.Send("EXPIRE", key, d.imageUpExpire); err != nil {
		log.Error("SetImageUpCache conn.Send(Expire, %s, %d) error(%v)", key, d.imageUpExpire, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("SetImageUpCache conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < 3; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("SetImageUpCache conn.Receive() error(%v)", err)
			return
		}
	}
	return
}

func (d *Dao) ImageUpCache(c context.Context, sid int64, day string, typ int) (list map[int64]float64, err error) {
	conn := d.redis.Get(c)
	defer conn.Close()
	key := imageUserKey(sid, day, typ)
	resMap, err := redis.StringMap(conn.Do("ZRANGE", key, 0, -1, "WITHSCORES"))
	if err != nil {
		err = errors.Wrapf(err, "ZRANGE key:%s", key)
		return
	}
	if len(resMap) == 0 {
		return
	}
	list = make(map[int64]float64, len(resMap))
	for k, v := range resMap {
		mid, _ := strconv.ParseInt(k, 10, 64)
		if mid <= 0 {
			continue
		}
		score, _ := strconv.ParseFloat(v, 64)
		list[mid] = score
	}
	return
}

func (d *Dao) AddCacheStupidTotal(ctx context.Context, sid, total int64) error {
	conn := d.redis.Get(ctx)
	defer conn.Close()
	if _, err := conn.Do("SETEX", fmt.Sprintf("stupid:vv:%d:%s", sid, stupidKey()), d.stupidExpire, total); err != nil {
		return err
	}
	return nil
}

func stupidKey() string {
	return time.Now().Format("2006010215")
}

func (d *Dao) AddCacheStupidList(c context.Context, sid int64, arcs map[int64][]*like.StupidVv) error {
	if len(arcs) == 0 {
		return errors.New("arcs nil")
	}
	var (
		key  string
		keys []string
		args = redis.Args{}
	)
	conn := d.redis.Get(c)
	defer conn.Close()
	for k, v := range arcs {
		bs, err := json.Marshal(v)
		if err != nil {
			log.Error("json.Marshal err(%v)", err)
			continue
		}
		key = fmt.Sprintf("stupid:arc:%d:%d:%s", sid, k, stupidKey())
		keys = append(keys, key)
		args = args.Add(key).Add(string(bs))
	}
	if err := conn.Send("MSET", args...); err != nil {
		return err
	}
	count := 1
	for _, v := range keys {
		count++
		if err := conn.Send("EXPIRE", v, d.stupidExpire); err != nil {
			return err
		}
	}
	if err := conn.Flush(); err != nil {
		return err
	}
	for i := 0; i < count; i++ {
		if _, err := conn.Receive(); err != nil {
			return err
		}
	}
	return nil
}
