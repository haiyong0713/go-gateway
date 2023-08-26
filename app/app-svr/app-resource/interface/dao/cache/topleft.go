package cache

import (
	"context"
	"fmt"
	"go-common/library/cache/credis"
	"go-common/library/cache/redis"
	"go-common/library/log"

	"go-gateway/app/app-svr/app-resource/interface/conf"
)

const (
	blackListScene = "main.homepage.avatar.0.click"
)

type Dao struct {
	c *conf.Config
	// redis
	Redis credis.Redis
}

// New new a archive dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:     c,
		Redis: credis.NewRedis(c.Redis.TopLeft),
	}
	return
}

func buildTopLeftBlackListKey(mid int64) string {
	return fmt.Sprintf("%d-%s-v2", mid, blackListScene)
}

func (d *Dao) HitTopLeftBlackList(ctx context.Context, mid int64) bool {
	conn := d.Redis.Conn(ctx)
	defer conn.Close()
	reply, err := redis.Bytes(conn.Do("GET", buildTopLeftBlackListKey(mid)))
	if err != nil {
		if err != redis.ErrNil {
			log.Error("HitTopLeftBlackList redis error(%+v), mid(%d)", err, mid)
		}
		return false
	}
	if len(reply) == 0 {
		return false
	}
	return true
}
