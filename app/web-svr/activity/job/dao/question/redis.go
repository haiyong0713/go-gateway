package question

import (
	"context"
	"encoding/json"
	"fmt"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/job/component"
	quesmdl "go-gateway/app/web-svr/activity/job/model/question"
)

const (
	_hourTop    = "hour_top"
	_hourPeople = "hour_people"
	_roundTop   = "round_top"
)

func poolKey(baseID, poolID int64) string {
	return fmt.Sprintf("pool_%d_%d", baseID, poolID)
}

func latestPoolID(baseID int64) string {
	return fmt.Sprintf("pool_latest_%d", baseID)
}

func poolAllKey(baseID, poolID int64) string {
	return fmt.Sprintf("pool_all_%d_%d", baseID, poolID)
}

func (d *Dao) userInfoKey(mid int64, strWeek string) string {
	return fmt.Sprintf("answer_user_%d_%s", mid, strWeek)
}

// SetPoolIDsCache set pool ids cache.
func (d *Dao) SetPoolIDsCache(c context.Context, baseID, poolID int64, ids []int64, expire int32) (err error) {
	if len(ids) == 0 {
		return
	}
	key := poolKey(baseID, poolID)
	conn := component.GlobalRedis.Conn(c)
	defer conn.Close()
	if err = conn.Send("DEL", key); err != nil {
		log.Errorc(c, "SetPoolIDsCache conn.Do(DEL %s) error(%v)", key, err)
		return
	}
	args := redis.Args{}.Add(key)
	for i, v := range ids {
		args = args.Add(i + 1).Add(v)
	}
	if err = conn.Send("HMSET", args...); err != nil {
		log.Errorc(c, "SetPoolIDsCache conn.Do(HMSET %s) error(%v)", key, err)
		return
	}
	if err = conn.Send("EXPIRE", key, expire); err != nil {
		log.Errorc(c, "SetPoolIDsCache conn.Send(Expire, %s, %d) error(%v)", key, expire, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Errorc(c, "SetPoolIDsCache conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < 3; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Errorc(c, "SetPoolIDsCache conn.Receive() error(%v)", err)
			return
		}
	}
	if _, err = conn.Do("SETEX", latestPoolID(baseID), 86400, poolID); err != nil {
		log.Errorc(c, "SetPoolIDsCache SETEX error(%v)", err)
	}
	return
}

// SetAllDetails
func (d *Dao) SetAllDetails(c context.Context, baseID, poolID int64, ids string, expire int32) (err error) {
	key := poolAllKey(baseID, poolID)
	if _, err = component.GlobalRedis.Do(c, "SETEX", key, expire, ids); err != nil {
		log.Errorc(c, "SetAllDetails conn.Do(SET) key(%s) error(%v)", key, err)
	}
	return
}

// CacheUserInfo .
func (d *Dao) CacheUserInfo(ctx context.Context, mid int64, strWeek string) (res *quesmdl.AnswerUserInfo, err error) {
	var (
		key = d.userInfoKey(mid, strWeek)
		bs  []byte
	)
	if bs, err = redis.Bytes(component.GlobalRedis.Do(ctx, "GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			log.Warn("CacheUserInfo(%s) return nil", key)
		} else {
			log.Errorc(ctx, "CacheUserInfo conn.Do(GET key(%v)) error(%v)", key, err)
		}
		return
	}
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Errorc(ctx, "CacheUserInfo json.Unmarshal(%v) error(%v)", bs, err)
	}
	return
}

// AddCacheUserInfo
func (d *Dao) AddCacheUserInfo(ctx context.Context, mid int64, strWeek string, data *quesmdl.AnswerUserInfo) (err error) {
	var (
		key = d.userInfoKey(mid, strWeek)
		bs  []byte
	)
	if bs, err = json.Marshal(data); err != nil {
		log.Errorc(ctx, "AddCacheUserInfo json.Marshal(%v) error (%+v)", data, err)
		return
	}
	if _, err = component.GlobalRedis.Do(ctx, "SETEX", key, 5000000, bs); err != nil {
		log.Errorc(ctx, "AddCacheUserInfo conn.Send(SETEX, %s, %v, %s) error(%v)", key, 5000000, string(bs), err)
	}
	return
}

// AddCacheUserTop
func (d *Dao) AddCacheUserTop(ctx context.Context, data []*quesmdl.UserRank) (err error) {
	var (
		bs []byte
	)
	if bs, err = json.Marshal(data); err != nil {
		log.Errorc(ctx, "AddCacheUserTop json.Marshal(%v) error (%+v)", data, err)
		return
	}
	if _, err = component.GlobalRedis.Do(ctx, "SETEX", _hourTop, 5000000, bs); err != nil {
		log.Errorc(ctx, "AddCacheUserTop conn.Send(SETEX, %s, %v, %s) error(%v)", _hourTop, 5000000, string(bs), err)
	}
	return
}

// AddCacheHourPeople
func (d *Dao) AddCacheHourPeople(ctx context.Context, data *quesmdl.HourPeople) (err error) {
	var (
		bs []byte
	)
	if bs, err = json.Marshal(data); err != nil {
		log.Errorc(ctx, "AddCacheHourPeople json.Marshal(%v) error (%+v)", data, err)
		return
	}
	if _, err = component.GlobalRedis.Do(ctx, "SETEX", _hourPeople, 86400, bs); err != nil {
		log.Errorc(ctx, "AddCacheHourPeople conn.Send(SETEX, %s, %v, %s) error(%v)", _hourPeople, 86400, string(bs), err)
	}
	return
}

// AddCacheRoundTop
func (d *Dao) AddCacheRoundTop(ctx context.Context, mids string) (err error) {
	if _, err = component.GlobalRedis.Do(ctx, "SETEX", _roundTop, 5000000, mids); err != nil {
		log.Errorc(ctx, "AddCacheRoundTop conn.Do(SET) key(%s) error(%v)", _roundTop, err)
	}
	return
}
