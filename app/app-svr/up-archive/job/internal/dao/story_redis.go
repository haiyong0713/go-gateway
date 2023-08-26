package dao

import (
	"context"
	"fmt"

	"go-common/library/cache/redis"
	"go-gateway/app/app-svr/up-archive/job/internal/model"

	"github.com/pkg/errors"
)

func arcStoryPassedKey(mid int64) string {
	return fmt.Sprintf("%d_arc_story", mid)
}

func (d *dao) AddCacheArcStoryPassed(ctx context.Context, mid int64, arcs []*model.UpArc) error {
	if len(arcs) == 0 {
		return nil
	}
	key := arcStoryPassedKey(mid)
	passedArg := redis.Args{}.Add(key)
	var passedAdd bool
	for _, v := range arcs {
		if v == nil {
			continue
		}
		passedAdd = true
		passedArg = passedArg.Add(v.Score).Add(v.Aid)
	}
	if !passedAdd {
		return nil
	}
	conn := d.redis.Conn(ctx)
	defer conn.Close()
	if err := conn.Send("ZADD", passedArg...); err != nil {
		return errors.Wrapf(err, "AddCacheArcStoryPassed passedArg ZADD mid:%d key:%s", mid, key)
	}
	if isEmptyCache(arcs) {
		if err := conn.Send("EXPIRE", key, d.emptyExpire()); err != nil {
			return errors.Wrapf(err, "AddCacheArcStoryPassed EXPIRE mid:%d key:%s", mid, key)
		}
	}
	if err := conn.Flush(); err != nil {
		return errors.Wrapf(err, "AddCacheArcStoryPassed mid:%d conn.Flush", mid)
	}
	return nil
}

func (d *dao) AppendCacheArcStoryPassed(ctx context.Context, mid int64, arcs []*model.UpArc) error {
	if len(arcs) == 0 {
		return nil
	}
	key := arcStoryPassedKey(mid)
	conn := d.redis.Conn(ctx)
	defer conn.Close()
	// 取缓存里面数据，确认长度和内容是不是空缓存
	aids, err := redis.Int64s(conn.Do("ZREVRANGE", key, 0, 2))
	if err != nil {
		return errors.Wrapf(err, "AppendCacheArcStoryPassed mid:%d key:%s redis.Int64s error:%v", mid, key, err)
	}
	if len(aids) == 0 {
		return nil
	}
	if len(aids) == 1 && aids[0] == -1 {
		if _, err = conn.Do("DEL", key); err != nil {
			return errors.Wrapf(err, "AppendCacheArcStoryPassed mid:%d key:%s DEL error:%v", mid, key, err)
		}
	}
	passedArg := redis.Args{}.Add(key)
	var passedAdd bool
	for _, v := range arcs {
		if v == nil {
			continue
		}
		passedAdd = true
		passedArg = passedArg.Add(v.Score).Add(v.Aid)
	}
	if !passedAdd {
		return nil
	}
	if err = conn.Send("ZADD", passedArg...); err != nil {
		return errors.Wrapf(err, "AppendCacheArcStoryPassed passedArg ZADD mid:%d key:%s error:%v", mid, key, err)
	}
	if err = conn.Flush(); err != nil {
		return errors.Wrapf(err, "AppendCacheArcStoryPassed mid:%d conn.Flush error:%v", mid, err)
	}
	return nil
}

func (d *dao) DelCacheArcStoryPassed(ctx context.Context, mid int64, arcs []*model.UpArc) error {
	key := arcStoryPassedKey(mid)
	passedArg := redis.Args{}.Add(key)
	var passedDel bool
	for _, v := range arcs {
		if v == nil {
			continue
		}
		passedDel = true
		passedArg = passedArg.Add(v.Aid)
	}
	if !passedDel {
		return nil
	}
	conn := d.redis.Conn(ctx)
	defer conn.Close()
	if err := conn.Send("ZREM", passedArg...); err != nil {
		return errors.Wrapf(err, "DelCacheArcStoryPassed passedArg ZREM mid:%d key:%s error:%v", mid, key, err)
	}
	if err := conn.Flush(); err != nil {
		return errors.Wrapf(err, "DelCacheArcStoryPassed mid:%d conn.Flush error:%v", mid, err)
	}
	return nil
}
