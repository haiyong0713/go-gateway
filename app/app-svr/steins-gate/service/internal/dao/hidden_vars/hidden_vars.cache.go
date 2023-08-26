package hidden_vars

import (
	"context"
	"encoding/json"
	"fmt"

	"go-gateway/app/app-svr/steins-gate/service/internal/model"

	"go-common/library/cache/redis"
	"go-common/library/log"
)

const (
	_hiddenVarRecordHasCursor = "%d_%d_%d_%d_%s_hv"
)

func (d *Dao) hvarKey(mid, gid, nid, cursor int64, buvid string) string {
	return fmt.Sprintf(_hiddenVarRecordHasCursor, mid, gid, nid, cursor, buvid)
}

func (d *Dao) setHvarCache(c context.Context, mid, gid, nid, cursor int64, buvid string, a *model.HiddenVarsRecord) (err error) {
	if a == nil {
		return
	}
	var item []byte
	if item, err = json.Marshal(a); err != nil { // json序列化
		log.Error("hvar.Marshal error(%v)", err)
		return
	}
	key := d.hvarKey(mid, gid, nid, cursor, buvid)
	conn := d.rds.Get(c)
	defer conn.Close()
	if _, err = conn.Do("SETEX", key, d.getCacheExpire(d.hvarExpirationMinH, d.hvarExpirationMaxH), item); err != nil {
		log.Error("conn.Do(SET, %s, %+v) error(%v)", key, a, err)
	}
	return
}

func (d *Dao) hvarCache(c context.Context, mid, gid, nid, cursor int64, buvid string) (a *model.HiddenVarsRecord, err error) {
	var (
		key  = d.hvarKey(mid, gid, nid, cursor, buvid)
		conn = d.rds.Get(c)
		item []byte
	)
	defer conn.Close()
	if item, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		log.Error("conn.Do(GET, %s) error(%v)", key, err)
		return
	}
	a = new(model.HiddenVarsRecord)
	if err = json.Unmarshal(item, &a); err != nil { // json反序列化
		log.Error("hvar Unmarshal error(%v)", err)
		return
	}
	return
}

func (d *Dao) getCacheExpire(min, max int) (res int) {
	return model.RandInt(d.rand, min, max)

}
