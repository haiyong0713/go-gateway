package entrance

import (
	"context"
	"fmt"

	"go-common/library/cache/credis"
	"go-common/library/cache/redis"
	resApi "go-gateway/app/app-svr/app-resource/interface/api/v1"
	"go-gateway/app/app-svr/app-resource/interface/conf"
	model "go-gateway/app/app-svr/app-resource/interface/model/entrance"
)

const (
	_prefixEntranceInfocKey = "entrance_infoc_%s_%d_%d"
)

type Dao struct {
	conf  *conf.Config
	redis credis.Redis
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		conf:  c,
		redis: credis.NewRedis(c.Redis.Resource.Config),
	}
	return
}

func (d *Dao) AddBusinessInfocCache(c context.Context, req *model.BusinessInfocReq) error {
	conn := d.redis.Conn(c)
	defer conn.Close()
	key := fmt.Sprintf(_prefixEntranceInfocKey, req.Business, req.Mid, req.UpMid)
	if _, err := conn.Do("SETEX", key, d.conf.EntranceKeyExpire.LiveReserveBlock, 1); err != nil {
		return err
	}
	return nil
}

func (d *Dao) BusinessInfocKeyExists(c context.Context, req *resApi.CheckEntranceInfocRequest) (bool, error) {
	conn := d.redis.Conn(c)
	defer conn.Close()
	key := fmt.Sprintf(_prefixEntranceInfocKey, req.Business, req.Mid, req.UpMid)
	exist, err := redis.Bool(conn.Do("EXISTS", key))
	if err != nil {
		return false, err
	}
	return exist, nil
}

// Close Dao
func (d *Dao) Close() {
	if d.redis != nil {
		d.redis.Close()
	}
}
