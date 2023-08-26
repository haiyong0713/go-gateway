package dao

import (
	"context"
	"fmt"

	"go-common/library/cache/redis"

	"github.com/pkg/errors"
)

func arcStoryPassedKey(mid int64) string {
	return fmt.Sprintf("%d_arc_story", mid)
}

func (d *dao) CacheArcPassedStory(ctx context.Context, mid, start, end int64, isAsc bool) ([]int64, error) {
	conn := d.redis.Conn(ctx)
	defer conn.Close()
	key := arcStoryPassedKey(mid)
	cmd := "ZREVRANGE"
	if isAsc {
		cmd = "ZRANGE"
	}
	aids, err := redis.Int64s(conn.Do(cmd, key, start, end))
	if err != nil {
		return nil, errors.Wrapf(err, "CacheArcPassedStory key:%s start:%d end:%d", key, start, end)
	}
	return aids, nil
}

func (d *dao) CacheArcPassedStoryTotal(ctx context.Context, mid int64) (int64, error) {
	conn := d.redis.Conn(ctx)
	defer conn.Close()
	key := arcStoryPassedKey(mid)
	total, err := redis.Int64(conn.Do("ZCARD", key))
	if err != nil {
		return 0, errors.Wrapf(err, "CacheArcPassedStoryTotal key:%s", key)
	}
	return total, nil
}

func (d *dao) ExpireEmptyArcPassedStory(ctx context.Context, mid int64) error {
	conn := d.redis.Conn(ctx)
	defer conn.Close()
	key := arcStoryPassedKey(mid)
	ttl, err := redis.Int64(conn.Do("TTL", key))
	if err != nil {
		return errors.Wrapf(err, "ExpireEmptyArcPassedStory TTL key:%s", key)
	}
	// only expire if not has expire
	if ttl != -1 {
		return nil
	}
	if _, err = conn.Do("EXPIRE", key, d.emptyExpire()); err != nil {
		return errors.Wrapf(err, "ExpireEmptyArcPassedStory EXPIRE key:%s", key)
	}
	return nil
}

func (d *dao) CacheArcPassedStoryExists(ctx context.Context, mid int64) (bool, error) {
	conn := d.redis.Conn(ctx)
	defer conn.Close()
	key := arcStoryPassedKey(mid)
	exist, err := redis.Bool(conn.Do("EXISTS", key))
	if err != nil {
		return false, errors.Wrapf(err, "CacheArcPassedStoryExists key:%s", key)
	}
	return exist, nil
}

func (d *dao) CacheArcPassedStoryAidRank(ctx context.Context, mid, aid int64, isAsc bool) (int64, error) {
	conn := d.redis.Conn(ctx)
	defer conn.Close()
	key := arcStoryPassedKey(mid)
	cmd := "ZREVRANK"
	if isAsc {
		cmd = "ZRANK"
	}
	rank, err := redis.Int64(conn.Do(cmd, key, aid))
	if err != nil {
		if err == redis.ErrNil {
			return 0, nil
		}
		return 0, errors.Wrapf(err, "CacheArcPassedStoryAidRank conn.Do(%s) key:%s", cmd, key)
	}
	return rank + 1, nil
}
