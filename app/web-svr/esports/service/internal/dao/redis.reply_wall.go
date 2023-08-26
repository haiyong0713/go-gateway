package dao

import (
	"context"
	"encoding/json"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/web-svr/esports/service/internal/model"
)

const (
	_replyWallList = "reply_wall_0701"
)

// ReplyWallList get data from cache if miss will call source method, then add to cache.
func (d *dao) ReplyWallList(ctx context.Context) (res []*model.ReplyWallModel, err error) {
	addCache := true
	res, err = d.CacheReplyWallList(ctx)
	if err != nil {
		addCache = false
		err = nil
	}
	defer func() {
		if len(res) != 0 && res[0].ID == -1 {
			res = nil
		}
	}()
	if len(res) != 0 {
		return
	}
	res, err = d.RawReplyWall()
	if err != nil {
		return
	}
	var miss = res
	if len(miss) == 0 {
		miss = []*model.ReplyWallModel{{ID: -1}}
	}
	if !addCache {
		return
	}
	d.cache.Do(ctx, func(c context.Context) {
		d.AddCacheReplyWallList(c, miss)
	})
	return
}

func (d *dao) CacheReplyWallList(ctx context.Context) (res []*model.ReplyWallModel, err error) {
	conn := d.redis.Conn(ctx)
	defer d.connClose(ctx, conn)
	key := _replyWallList
	var data string
	if data, err = redis.String(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		log.Errorc(ctx, "CacheReplyWallList() Do(GET) key:%v error:%v", key, err)
		return
	}
	if err = json.Unmarshal([]byte(data), &res); err != nil {
		log.Errorc(ctx, "CacheReplyWallList() json.Unmarshal() key:%v error:%v", key, err)
		return nil, err
	}
	return
}

func (d *dao) AddCacheReplyWallList(ctx context.Context, value []*model.ReplyWallModel) (err error) {
	conn := d.redis.Conn(ctx)
	defer d.connClose(ctx, conn)
	key := _replyWallList
	var data []byte
	if data, err = json.Marshal(value); err != nil {
		log.Errorc(ctx, "AddReplyWallList(%+v) error(%v)", key, err)
		return err
	}
	if _, err = conn.Do("SET", key, string(data), "EX", 300); err != nil {
		log.Error("AddReplyWallList(%+v) error(%v)", key, err)
	}
	return
}

func (d *dao) DelCacheReplyWallList(ctx context.Context) (err error) {
	conn := d.redis.Conn(ctx)
	defer d.connClose(ctx, conn)
	key := _replyWallList
	if _, err = conn.Do("DEL", key); err != nil {
		log.Error("DelCacheReplyWallList(%v) error(%v)", key, err)
		return
	}
	return
}
