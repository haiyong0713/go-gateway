package timemachine

import (
	"context"
	"fmt"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/model/timemachine"
)

func timemachineKey(mid int64) string {
	return fmt.Sprintf("tm_%d", mid)
}

func (d *Dao) AddCacheTimemachine(c context.Context, mid int64, data *timemachine.Item) (err error) {
	if data == nil {
		return
	}
	var item []byte
	if item, err = data.Marshal(); err != nil {
		log.Error("AddCacheTimemachine data.Marshal error(%v)", err)
		return
	}
	key := timemachineKey(mid)
	if _, err = d.redis.Do(c, "SET", key, item); err != nil {
		log.Error("conn.Do(SET, %s, %+v) error(%v)", key, data, err)
	}
	return
}

func (d *Dao) CacheTimemachine(c context.Context, mid int64) (data *timemachine.Item, err error) {
	var (
		key  = timemachineKey(mid)
		item []byte
	)
	if item, err = redis.Bytes(d.redis.Do(c, "GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		log.Error("conn.Do(GET, %s) error(%v)", key, err)
		return
	}
	data = new(timemachine.Item)
	if err = data.Unmarshal(item); err != nil {
		log.Error("CacheTimemachine data.Unmarshal error(%v)", err)
		data = nil
		return
	}
	return
}
