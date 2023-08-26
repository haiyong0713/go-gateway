package cache

//处理play-service一些需要jd和ylf缓存共用一份的逻辑

import (
	"context"
	"encoding/json"
	"fmt"

	"go-common/library/cache/credis"
	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/app-svr/playurl/service/conf"
	"go-gateway/app/app-svr/playurl/service/model"

	"github.com/pkg/errors"
)

// Dao is archive dao.
type Dao struct {
	c *conf.Config
	// redis
	Redis credis.Redis
}

// New new a archive dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:     c,
		Redis: credis.NewRedis(c.Redis.MixRedis),
	}
	return
}

func confABTestKey(buvid string) string {
	return fmt.Sprintf("nu_%s", buvid)
}

// SetConfABTest .
func (d *Dao) SetNXConfABTest(c context.Context, buvid string, hasChanged int64) (bool, error) {
	conn := d.Redis.Conn(c)
	defer conn.Close()
	ok, err := redis.String(conn.Do("SET", confABTestKey(buvid), hasChanged, "EX", d.c.Custom.NewDeviceTime, "NX"))
	if err != nil {
		return false, errors.Wrapf(err, "SetNXConfABTest error buvid(%s)", buvid)
	}
	if ok != "OK" {
		return false, nil
	}
	return true, nil
}

func (d *Dao) GetConfABTest(c context.Context, buvid string) (bool, error) {
	conn := d.Redis.Conn(c)
	defer conn.Close()
	res, err := redis.Bool(conn.Do("GET", confABTestKey(buvid)))
	if err != nil {
		return false, err
	}
	return res, nil
}

func (d *Dao) SetConfABTest(c context.Context, buvid string, hasChanged int64) error {
	conn := d.Redis.Conn(c)
	defer conn.Close()
	_, err := conn.Do("SET", confABTestKey(buvid), hasChanged, "EX", d.c.Custom.NewDeviceTime)
	return err
}

func (d *Dao) FetchOnlineInfo(c context.Context, aid int64) (map[int64]int64, int64, bool) {
	conn := d.Redis.Conn(c)
	defer conn.Close()
	res, err := redis.Bytes(conn.Do("GET", model.OnlineKey(aid)))
	if err != nil {
		if err != redis.ErrNil {
			log.Error("FetchOnlineInfo error(%+v)", err)
			return nil, 0, false
		}
		return nil, 0, false
	}
	var onlineInfo = &model.OnlineInfo{}
	if err := json.Unmarshal(res, onlineInfo); err != nil {
		log.Error("Unmarshal error(%+v) aid(%d)", err, aid)
		return nil, 0, false
	}
	return onlineInfo.AppCidCount, onlineInfo.Time, true
}
