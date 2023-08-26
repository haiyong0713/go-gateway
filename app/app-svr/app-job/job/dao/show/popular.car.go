package show

import (
	"context"
	"encoding/json"
	"fmt"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-job/job/model/show"
)

const (
	_carHotTenprefix = "%d_hchashmap_car"
)

func getCarHotKey(i int) string {
	return fmt.Sprintf(_carHotTenprefix, i)
}

func (d *Dao) AddCarPopularCardTenCache(c context.Context, i int, cards []*show.PopularCardAI) (err error) {
	if len(cards) == 0 {
		return
	}
	var (
		key  = getCarHotKey(i)
		conn = d.redis.Get(c)
		item []byte
	)
	defer conn.Close()
	args := redis.Args{}.Add(key)
	for index, card := range cards {
		if item, err = json.Marshal(card); err != nil {
			log.Error("Marshal error(%v) card(%+v) index(%d)", err, card, index)
			return
		}
		args = args.Add(index).Add(item)
	}
	if _, err = conn.Do("HSET", args...); err != nil {
		log.Error("conn.Send(HSET,%v) error(%v)", args, err)
	}
	return
}

func (d *Dao) TotalCarPopularCardTenCache(c context.Context, i int) (count int, err error) {
	conn := d.redis.Get(c)
	defer conn.Close()
	if count, err = redis.Int(conn.Do("HLEN", getCarHotKey(i))); err != nil {
		if err != redis.ErrNil {
			log.Error("conn.Do(HLEN, %s) error(%v)", getCarHotKey(i), err)
			return
		}
		err = nil
	}
	return
}

func (d *Dao) DelCarPopularCardTenCache(c context.Context, i, start, end int) (err error) {
	var (
		key  = getCarHotKey(i)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	args := redis.Args{}.Add(key)
	for start < end {
		args = args.Add(start)
		start++
	}
	if _, err = conn.Do("HDEL", args...); err != nil {
		log.Error("conn.Do(HDEL,%v) error(%v)", args, err)
	}
	return
}
