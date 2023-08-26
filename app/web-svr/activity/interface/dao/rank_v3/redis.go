package rank

import (
	"context"
	"encoding/json"
	"go-common/library/cache/redis"
	"go-common/library/log"

	rankmdl "go-gateway/app/web-svr/activity/interface/model/rank_v3"
)

// AddRankRule add cache lottery
func (d *Dao) AddRankRule(c context.Context, ruleID int64, val *rankmdl.Rule) (err error) {
	var (
		key = buildKey(ruleKey, ruleID)
		bs  []byte
	)
	if bs, err = json.Marshal(val); err != nil {
		log.Errorc(c, "json.Marshal(%v) error (%v)", val, err)
		return
	}
	if _, err = d.redisCache.Do(c, "SETEX", key, d.OneDayExpire, bs); err != nil {
		log.Errorc(c, "AddRankRule conn.Send(SETEX, %s, %v, %s) error(%v)", key, d.OneDayExpire, string(bs), err)
		return
	}
	return
}

// GetRankRule 用户集卡情况
func (d *Dao) GetRankRule(c context.Context, ruleID int64) (res *rankmdl.Rule, err error) {
	var (
		key = buildKey(ruleKey, ruleID)
		bs  []byte
	)
	if bs, err = redis.Bytes(d.redisCache.Do(c, "GET", key)); err != nil {
		log.Errorc(c, "GetRankRule conn.Do(GET key(%v)) error(%v)", key, err)
		return
	}
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Errorc(c, "json.Unmarshal(%v) error(%v)", bs, err)
	}
	return
}

// DeleteRankRule ...
func (d *Dao) DeleteRankRule(c context.Context, ruleID int64) (err error) {
	var (
		key = buildKey(ruleKey, ruleID)
	)
	if _, err = d.redisCache.Do(c, "DEL", key); err != nil {
		log.Errorc(c, "DeleteRankRule conn.Do(DEL, %s) error(%v)", key, err)
	}
	return
}

// AddRankBase add cache lottery
func (d *Dao) AddRankBase(c context.Context, baseID int64, val *rankmdl.Base) (err error) {
	var (
		key = buildKey(baseKey, baseID)
		bs  []byte
	)
	if bs, err = json.Marshal(val); err != nil {
		log.Errorc(c, "json.Marshal(%v) error (%v)", val, err)
		return
	}
	if _, err = d.redisCache.Do(c, "SETEX", key, d.OneDayExpire, bs); err != nil {
		log.Errorc(c, "AddRankBase conn.Send(SETEX, %s, %v, %s) error(%v)", key, d.OneDayExpire, string(bs), err)
		return
	}
	return
}

// GetRankBase 用户集卡情况
func (d *Dao) GetRankBase(c context.Context, baseID int64) (res *rankmdl.Base, err error) {
	var (
		key = buildKey(baseKey, baseID)
		bs  []byte
	)
	if bs, err = redis.Bytes(d.redisCache.Do(c, "GET", key)); err != nil {
		log.Errorc(c, "GetRankBase conn.Do(GET key(%v)) error(%v)", key, err)
		return
	}
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Errorc(c, "json.Unmarshal(%v) error(%v)", bs, err)
	}
	return
}

// DeleteRankBase ...
func (d *Dao) DeleteRankBase(c context.Context, baseID int64) (err error) {
	var (
		key = buildKey(baseKey, baseID)
	)
	if _, err = d.redisCache.Do(c, "DEL", key); err != nil {
		log.Errorc(c, "DeleteRankBase conn.Do(DEL, %s) error(%v)", key, err)
	}
	return
}
