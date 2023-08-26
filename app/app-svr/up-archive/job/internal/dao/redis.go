package dao

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/conf/paladin.v2"
	"go-common/library/log"
	"go-gateway/app/app-svr/up-archive/job/internal/model"
	"go-gateway/app/app-svr/up-archive/service/api"

	"github.com/pkg/errors"
)

func NewRedis() (r *redis.Redis, cf func(), err error) {
	var (
		cfg redis.Config
		ct  paladin.Map
	)
	if err = paladin.Get("redis.toml").Unmarshal(&ct); err != nil {
		return
	}
	if err = ct.Get("Client").UnmarshalTOML(&cfg); err != nil {
		return
	}
	r = redis.NewRedis(&cfg)
	cf = func() { r.Close() }
	return
}

func (d *dao) PingRedis(ctx context.Context) (err error) {
	if _, err = d.redis.Do(ctx, "SET", "ping", "pong"); err != nil {
		log.Error("conn.Set(PING) error(%v)", err)
	}
	return
}

func arcPassedKey(mid int64, without api.Without) string {
	switch without {
	case api.Without_none:
		return fmt.Sprintf("%d_arc_passed", mid)
	case api.Without_staff:
		return fmt.Sprintf("%d_arc_simple", mid)
	default:
		return fmt.Sprintf("%d_arc_%s", mid, without.String())
	}
}

func isEmptyCache(arcs []*model.UpArc) bool {
	if len(arcs) != 1 {
		return false
	}
	for _, v := range arcs {
		if v == nil {
			return false
		}
		if v.Aid == -1 && v.Score == 0 {
			return true
		}
	}
	return false
}

func (d *dao) emptyExpire() int32 {
	rand.Seed(time.Now().UnixNano())
	return d.emptyCacheExpire + rand.Int31n(d.emptyCacheRand)
}

func (d *dao) AddCacheArcPassed(ctx context.Context, mid int64, arcs []*model.UpArc, without api.Without) error {
	if len(arcs) == 0 {
		return nil
	}
	key := arcPassedKey(mid, without)
	if key == "" {
		return nil
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
	conn := d.redis.Conn(ctx)
	defer conn.Close()
	if err := conn.Send("ZADD", passedArg...); err != nil {
		return errors.Wrapf(err, "AddCacheArcPassed passedArg ZADD mid:%d key:%s", mid, key)
	}
	if isEmptyCache(arcs) {
		if err := conn.Send("EXPIRE", key, d.emptyExpire()); err != nil {
			return errors.Wrapf(err, "AddCacheArcPassed EXPIRE mid:%d key:%s", mid, key)
		}
	}
	if err := conn.Flush(); err != nil {
		return errors.Wrapf(err, "AddCacheArcPassed mid:%d conn.Flush", mid)
	}
	return nil
}

func (d *dao) AppendCacheArcPassed(ctx context.Context, mid int64, arcs []*model.UpArc, without api.Without) error {
	if len(arcs) == 0 {
		return nil
	}
	key := arcPassedKey(mid, without)
	if key == "" {
		return nil
	}
	conn := d.redis.Conn(ctx)
	defer conn.Close()
	// 取缓存里面数据，确认长度和内容是不是空缓存
	aids, err := redis.Int64s(conn.Do("ZREVRANGE", key, 0, 2))
	if err != nil {
		return errors.Wrapf(err, "AppendCacheArcPassed mid:%d key:%s redis.Int64s error:%v", mid, key, err)
	}
	if len(aids) == 0 {
		return nil
	}
	if len(aids) == 1 && aids[0] == -1 {
		if _, err = conn.Do("DEL", key); err != nil {
			return errors.Wrapf(err, "AppendCacheArcPassed mid:%d key:%s DEL error:%v", mid, key, err)
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
		return errors.Wrapf(err, "AppendCacheArcPassed passedArg ZADD mid:%d key:%s error:%v", mid, key, err)
	}
	if err = conn.Flush(); err != nil {
		return errors.Wrapf(err, "AppendCacheArcPassed mid:%d conn.Flush error:%v", mid, err)
	}
	return nil
}

func (d *dao) DelCacheArcPassed(ctx context.Context, mid int64, arcs []*model.UpArc, without api.Without) error {
	key := arcPassedKey(mid, without)
	if key == "" {
		return nil
	}
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
		return errors.Wrapf(err, "DelCacheArcPassed passedArg ZREM mid:%d key:%s error:%v", mid, key, err)
	}
	if err := conn.Flush(); err != nil {
		return errors.Wrapf(err, "DelCacheArcPassed mid:%d conn.Flush error:%v", mid, err)
	}
	return nil
}

func (d *dao) BuildArcPassedLock(ctx context.Context, mid int64) (bool, error) {
	conn := d.redis.Conn(ctx)
	defer conn.Close()
	key := fmt.Sprintf("%d_passed_lock", mid)
	reply, err := conn.Do("SET", key, "LOCK", "EX", d.lockExpire, "NX")
	if err != nil {
		if err == redis.ErrNil {
			return false, nil
		}
		return false, errors.Wrapf(err, "BuildArcPassedLock mid:%d", mid)
	}
	return reply == "OK", nil
}

func (d *dao) DelCacheAllArcPassed(ctx context.Context, mid int64) error {
	key := arcPassedKey(mid, api.Without_none)
	withoutStaffKey := arcPassedKey(mid, api.Without_staff)
	withoutNoSpaceKey := arcPassedKey(mid, api.Without_no_space)
	storyKey := arcStoryPassedKey(mid)
	conn := d.redis.Conn(ctx)
	defer conn.Close()
	if err := conn.Send("DEL", key); err != nil {
		return errors.Wrapf(err, "DelCacheAllArcPassed mid:%d DEL key:%s", mid, key)
	}
	if err := conn.Send("DEL", withoutStaffKey); err != nil {
		return errors.Wrapf(err, "DelCacheAllArcPassed mid:%d DEL key:%s", mid, withoutStaffKey)
	}
	if err := conn.Send("DEL", withoutNoSpaceKey); err != nil {
		return errors.Wrapf(err, "DelCacheAllArcPassed mid:%d DEL key:%s", mid, withoutNoSpaceKey)
	}
	if err := conn.Send("DEL", storyKey); err != nil {
		return errors.Wrapf(err, "DelCacheAllArcPassed mid:%d DEL key:%s", mid, storyKey)
	}
	if err := conn.Flush(); err != nil {
		return errors.Wrapf(err, "DelCacheAllArcPassed mid:%d flush", mid)
	}
	return nil
}

func (d *dao) DelCacheArcNoSpace(ctx context.Context, mid int64) error {
	withoutNoSpaceKey := arcPassedKey(mid, api.Without_no_space)
	if _, err := d.redis.Do(ctx, "DEL", withoutNoSpaceKey); err != nil {
		return errors.Wrapf(err, "DelCacheAllArcPassed mid:%d DEL key:%s", mid, withoutNoSpaceKey)
	}
	return nil
}

func (d *dao) CacheArcPassedExists(ctx context.Context, mid int64, without api.Without) (bool, error) {
	conn := d.redis.Conn(ctx)
	defer conn.Close()
	key := arcPassedKey(mid, without)
	if key == "" {
		return false, nil
	}
	exist, err := redis.Bool(conn.Do("EXISTS", key))
	if err != nil {
		return false, errors.Wrapf(err, "CacheArcPassedExists key:%s", key)
	}
	return exist, nil
}
