package show

import (
	"context"
	"encoding/json"
	"fmt"

	"go-common/library/cache/redis"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-job/job/model/show"

	"github.com/pkg/errors"
)

const (
	_hotTenprefix        = "%d_hchashmap"
	_hotHeTongtabcardURL = "/data/rank/reco-app-remen-card-%d.json"
)

func getHotKey(i int) string {
	return fmt.Sprintf(_hotTenprefix, i)
}

func (d *Dao) AddPopularCardTenCache(c context.Context, i int, cards []*show.PopularCardAI) (err error) {
	if len(cards) == 0 {
		return
	}
	var (
		key  = getHotKey(i)
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

func (d *Dao) TotalPopularCardTenCache(c context.Context, i int) (count int, err error) {
	conn := d.redis.Get(c)
	defer conn.Close()
	if count, err = redis.Int(conn.Do("HLEN", getHotKey(i))); err != nil {
		if err != redis.ErrNil {
			log.Error("conn.Do(HLEN, %s) error(%v)", getHotKey(i), err)
			return
		}
		err = nil
	}
	return
}

func (d *Dao) DelPopularCardTenCache(c context.Context, i, start, end int) (err error) {
	var (
		key  = getHotKey(i)
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

func (d *Dao) HotHeTongTabCard(c context.Context, i int) (list []*show.CardListAI, err error) {
	var res struct {
		Code int                `json:"code"`
		List []*show.CardListAI `json:"list"`
	}
	if err = d.client.Get(c, fmt.Sprintf(d.hotHeTongtabcardURL, i), "", nil, &res); err != nil {
		err = errors.Wrap(err, fmt.Sprintf(d.hotHeTongtabcardURL, i))
		return
	}
	if res.Code != 0 {
		err = errors.Wrap(ecode.Int(res.Code), fmt.Sprintf("code(%d)", res.Code))
		return
	}
	list = res.List
	return
}
