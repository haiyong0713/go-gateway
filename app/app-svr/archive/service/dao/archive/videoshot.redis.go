package archive

import (
	"context"
	"encoding/json"

	"go-gateway/app/app-svr/archive/service/model"
	"go-gateway/app/app-svr/archive/service/model/videoshot"

	"go-common/library/cache/redis"
)

// NewVideoShotCache is contains HD vs
func (d *Dao) NewVideoShotCache(c context.Context, cid int64) (*videoshot.Videoshot, error) {
	key := model.NewVideoShotKey(cid)
	conn := d.sArcRds.Get(c)
	defer conn.Close()
	vsbs, err := redis.Bytes(conn.Do("GET", key))
	if err != nil {
		if err == redis.ErrNil {
			d.missProm.Incr("VideoShotCache")
		}
		return nil, err
	}
	var vs = &videoshot.Videoshot{}
	if err := json.Unmarshal(vsbs, vs); err != nil {
		return nil, err
	}
	d.hitProm.Incr("VideoShotCache")
	return vs, nil
}

// AddNewVideoShotCache is
func (d *Dao) AddNewVideoShotCache(c context.Context, cid int64, vs *videoshot.Videoshot) error {
	key := model.NewVideoShotKey(cid)
	vsbs, err := json.Marshal(vs)
	if err != nil {
		return err
	}
	conn := d.sArcRds.Get(c)
	defer conn.Close()
	_, err = conn.Do("SET", key, vsbs)
	return err
}
