package ott

import (
	"context"
	"fmt"

	"go-common/library/cache/redis"
	"go-gateway/app/web-svr/dance-taiko/interface/api"

	"github.com/pkg/errors"
)

func playerStatKey(gameId, mid int64) string {
	return fmt.Sprintf("ott_stat_%d_%d", gameId, mid)
}

func (d *dao) AddCachePlayerStat(c context.Context, gameId, mid int64, stats []*api.StatAcc) error {
	if len(stats) == 0 {
		return nil
	}
	conn := d.redis.Get(c)
	defer conn.Close()

	args := redis.Args{}.Add(playerStatKey(gameId, mid))
	for _, v := range stats {
		args = args.Add(v.Ts).Add(v.Acc)
	}
	if _, err := conn.Do("ZADD", args...); err != nil {
		return errors.Wrapf(err, "AddCachePlayerStat gameId(%d) mid(%d)", gameId, mid)
	}
	return nil
}
