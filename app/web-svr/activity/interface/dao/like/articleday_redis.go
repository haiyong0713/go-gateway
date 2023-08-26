package like

import (
	"context"
	"encoding/json"
	"fmt"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/model/like"
)

const (
	_articleRightKey = "a_day2_right"
	_articleMidKey   = "a_day2_%d"
)

func articleDayKey(mid int64) string {
	return fmt.Sprintf(_articleMidKey, mid)
}

func (d *Dao) CacheArticleDay(ctx context.Context, mid int64) (res *like.ArticleDay, err error) {
	var (
		key  = articleDayKey(mid)
		conn = d.redis.Get(ctx)
		bs   []byte
	)
	defer conn.Close()
	if bs, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			log.Warn("CacheArticleDay(%s) return nil", key)
		} else {
			log.Error("CacheArticleDay conn.Do(GET key(%v)) error(%v)", key, err)
		}
		return
	}
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Error("CacheArticleDay json.Unmarshal(%v) error(%v)", bs, err)
	}
	return
}

func (d *Dao) AddCacheArticleDay(c context.Context, mid int64, data *like.ArticleDay) (err error) {
	var (
		key  = articleDayKey(mid)
		conn = d.redis.Get(c)
		bs   []byte
	)
	defer conn.Close()
	if bs, err = json.Marshal(data); err != nil {
		log.Error("CacheArticleDay json.Marshal(%v) error (%+v)", data, err)
		return
	}
	if err = conn.Send("SETEX", key, d.matchExpire, bs); err != nil {
		log.Error("AddCacheArticleDay conn.Send(SETEX, %s, %v, %s) error(%v)", key, d.matchExpire, string(bs), err)
	}
	return
}

// ArticleRightInfo .
func (d *Dao) ArticleRightInfo(c context.Context) (data *like.RightInfo, err error) {
	var (
		conn = d.redis.Get(c)
		bs   []byte
	)
	defer conn.Close()
	if bs, err = redis.Bytes(conn.Do("GET", _articleRightKey)); err != nil {
		if err == redis.ErrNil {
			data = nil
			err = nil
		} else {
			log.Error("ArticleRightInfo conn.Do(GET key(%v)) error(%v)", _articleRightKey, err)
		}
		return
	}
	data = new(like.RightInfo)
	if err = json.Unmarshal(bs, data); err != nil {
		log.Error("ArticleRightInfo json.Unmarshal(%s) error(%v)", string(bs), err)
	}
	return
}
