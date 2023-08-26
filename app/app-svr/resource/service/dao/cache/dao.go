package cache

import (
	"context"
	"fmt"

	"go-common/library/cache/credis"
	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/app-svr/resource/service/conf"
)

// Dao is resource dao.
type Dao struct {
	c *conf.Config
	// redis
	redis         credis.Redis
	redisEntrance credis.Redis
}

// New init redis
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:             c,
		redis:         credis.NewRedis(c.Redis.Comm),
		redisEntrance: credis.NewRedis(c.Redis.Entrance),
	}
	return
}

// Close close the resource.
func (d *Dao) Close() {
	d.redis.Close()
}

func (d *Dao) KeyMenuVer(id int64, buvid, ver string) string {
	return fmt.Sprintf("imv_%s_%d_%s", buvid, id, ver)
}

// CacheMenuVer .
func (d *Dao) CacheMenuVer(c context.Context, id int64, buvid, ver string) (int, error) {
	conn := d.redis.Conn(c)
	defer conn.Close()
	key := d.KeyMenuVer(id, buvid, ver)
	count, err := redis.Int(conn.Do("GET", key))
	if err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Error("CacheMenuVer get %s error(%v)", key, err)
		}
	}
	return count, err
}

// AddMenuVer .
func (d *Dao) AddMenuVer(c context.Context, id int64, buvid, ver string) error {
	conn := d.redis.Conn(c)
	defer conn.Close()
	key := d.KeyMenuVer(id, buvid, ver)
	if _, err := conn.Do("SET", key, 1); err != nil {
		log.Error("AddMenuVer set %s error(%v)", key, err)
		return err
	}
	return nil
}

func (d *Dao) MenuExtVers(c context.Context, keys []string) (clicks map[string]int, err error) {
	len := len(keys)
	if len <= 0 {
		return
	}
	conn := d.redis.Conn(c)
	defer conn.Close()
	params := make([]interface{}, len)
	for index, key := range keys {
		params[index] = key
	}
	clickCache, err := redis.Ints(conn.Do("MGET", params...))
	if err != nil {
		if err != redis.ErrNil {
			log.Error("Dao:MenuExtVers() mget %s error(%v)", keys, err)
		}
		return
	}

	clicks = make(map[string]int)
	for index, clicked := range clickCache {
		key := params[index].(string)
		clicks[key] = clicked
	}

	return
}
