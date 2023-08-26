package like

import (
	"context"
	"encoding/json"
	"go-gateway/app/web-svr/activity/interface/component"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/model/like"
)

const _steinListKey = "stein_l"

// SteinList .
func (d *Dao) SteinList(c context.Context) (data *like.SteinList, err error) {
	var (
		bs []byte
	)
	if bs, err = redis.Bytes(component.GlobalRedisStore.Do(c, "GET", _steinListKey)); err != nil {
		if err == redis.ErrNil {
			data = nil
			err = nil
		} else {
			log.Error("SteinList conn.Do(GET key(%v)) error(%v)", _steinListKey, err)
		}
		return
	}
	data = new(like.SteinList)
	if err = json.Unmarshal(bs, data); err != nil {
		log.Error("SteinList json.Unmarshal(%s) error(%v)", string(bs), err)
	}
	return
}
