package fawkes

import (
	"context"
	"encoding/json"
	"fmt"

	"go-common/library/cache/redis"

	"go-gateway/app/app-svr/fawkes/service/model/app"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"

	"github.com/pkg/errors"
)

var (
	LASER_BUSINESS_KEY_LASER = "LOG"
	LASER_BUSINESS_KEY_CMD   = "CMD"
)

// Generate Redis key
func getRedisKey(mid int64, buvid, mobiApp, business string) (key string, err error) {

	if buvid != "" {
		key = fmt.Sprintf("FAWKES:LASER:%v:%v:%v", business, mobiApp, buvid)
	} else if mid != 0 {
		key = fmt.Sprintf("FAWKES:LASER:%v:%v:%v", business, mobiApp, mid)
	}
	if key == "" {
		err = errors.Wrap(err, "mid, buvid 不可同时为空")
	}
	return key, err
}

// SetCacheLaserMessage
func (d *Dao) SetCacheLaserMessage(ctx context.Context, laser *app.Laser) (key string, err error) {
	var (
		bs []byte
	)
	if key, err = getRedisKey(laser.MID, laser.Buvid, laser.MobiApp, LASER_BUSINESS_KEY_LASER); err != nil {
		log.Error("getRedisKey error(%v)", err)
		return
	}
	if bs, err = json.Marshal(laser); err != nil {
		log.Error("json.Marshal error(%v)", err)
		return
	}
	if _, err = d.redis.Do(ctx, "HSET", key, laser.ID, string(bs)); err != nil {
		log.Error("d.redis.Do error(%v)", err)
		return
	}
	log.Warn("d.redis.Do success:%v", laser.ID)
	return
}

// GetCacheLaserMessage
func (d *Dao) GetCacheLaserMessage(ctx context.Context, mid int64, buvid, mobiApp string) (res []*app.Laser, err error) {
	var (
		key string
	)
	if key, err = getRedisKey(mid, buvid, mobiApp, LASER_BUSINESS_KEY_LASER); err != nil {
		log.Error("getRedisKey error(%v)", err)
		return
	}
	var (
		bss [][]byte
	)
	if bss, err = redis.ByteSlices(d.redis.Do(ctx, "HGETALL", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Error("StatCache conn.Do(HGETALL,%s) error(%v)", key, err)
		}
		return
	}
	for i := 1; i <= len(bss); i += 2 {
		stat := new(app.Laser)
		if err = json.Unmarshal(bss[i], stat); err != nil {
			log.Error("json.Unmarshal(%s) error(%v)", string(bss[i]), err)
			continue
		}
		res = append(res, stat)
	}
	return
}

// DeleteCacheLaserMessage
func (d *Dao) DeleteCacheLaserMessage(ctx context.Context, mid int64, buvid, mobiApp string, laserID int64) (err error) {
	var (
		key string
	)
	if key, err = getRedisKey(mid, buvid, mobiApp, LASER_BUSINESS_KEY_LASER); err != nil {
		log.Error("getRedisKey error(%v)", err)
		return
	}
	if _, err = d.redis.Do(ctx, "HDEL", key, laserID); err != nil {
		log.Error("d.redis.Do error(%v)", err)
		return
	}
	return
}

// Cache Laser Command Message
func (d *Dao) SetCacheLaserCommandMessage(ctx context.Context, cmd *app.LaserCmd) (key string, err error) {
	var (
		bs []byte
	)
	if key, err = getRedisKey(cmd.MID, cmd.Buvid, cmd.MobiApp, LASER_BUSINESS_KEY_CMD); err != nil {
		log.Error("getRedisKey error(%v)", err)
		return
	}
	if bs, err = json.Marshal(cmd); err != nil {
		log.Error("json.Marshal error(%v)", err)
		return
	}
	if _, err = d.redis.Do(ctx, "HSET", key, cmd.ID, string(bs)); err != nil {
		log.Error("d.redis.Do error(%v)", err)
		return
	}
	log.Error("d.redis.Do success(%v-%v)", cmd.ID, string(bs))
	return
}

// GetCacheLaserCommandMessage
func (d *Dao) GetCacheLaserCommandMessage(ctx context.Context, mid int64, buvid, mobiApp string) (res []*app.LaserCmd, err error) {
	var (
		key string
	)
	if key, err = getRedisKey(mid, buvid, mobiApp, LASER_BUSINESS_KEY_CMD); err != nil {
		log.Error("getRedisKey error(%v)", err)
		return
	}
	var (
		bss [][]byte
	)
	if bss, err = redis.ByteSlices(d.redis.Do(ctx, "HGETALL", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Error("StatCache conn.Do(HGETALL,%s) error(%v)", key, err)
		}
		return
	}
	for i := 1; i <= len(bss); i += 2 {
		stat := new(app.LaserCmd)
		if err = json.Unmarshal(bss[i], stat); err != nil {
			log.Error("json.Unmarshal(%s) error(%v)", string(bss[i]), err)
			continue
		}
		res = append(res, stat)
	}
	return
}

// DeleteCacheLaserCommandMessage
func (d *Dao) DeleteCacheLaserCommandMessage(ctx context.Context, mid int64, buvid, mobiApp string, id int64) (err error) {
	var (
		key string
	)
	if key, err = getRedisKey(mid, buvid, mobiApp, LASER_BUSINESS_KEY_CMD); err != nil {
		log.Error("getRedisKey error(%v)", err)
		return
	}
	if _, err = d.redis.Do(ctx, "HDEL", key, id); err != nil {
		log.Error("d.redis.Do error(%v)", err)
	}
	return
}
