package like

import (
	"context"
	"encoding/json"
	"fmt"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-common/library/xstr"
	"go-gateway/app/web-svr/activity/interface/component"
	"go-gateway/app/web-svr/activity/interface/model/like"
)

const (
	_hourTop    = "hour_top"
	_hourPeople = "hour_people"
	_roundTop   = "round_top"
)

func poolAllKey(baseID, poolID int64) string {
	return fmt.Sprintf("pool_all_%d_%d", baseID, poolID)
}

func userInfoKey(mid int64, strWeek string) string {
	return fmt.Sprintf("answer_user_%d_%s", mid, strWeek)
}

func pendantRuleKey(mid int64) string {
	return fmt.Sprintf("pendan_rule_%d", mid)
}

func knowRuleKey(mid int64) string {
	return fmt.Sprintf("know_rule_%d", mid)
}

func userHpKey(mid, poolID int64) string {
	return fmt.Sprintf("user_hp_%d_%d", mid, poolID)
}

// PoolQuestionPage pool question page.
func (d *Dao) PoolQuestionPage(ctx context.Context, baseID, poolID int64) (ids []int64, err error) {
	key := poolAllKey(baseID, poolID)
	data, err := redis.String(component.GlobalRedis.Do(ctx, "GET", key))
	if err != nil {
		log.Errorc(ctx, "PoolQuestionPage GET error:%v", err)
		return nil, err
	}
	if data == "" {
		return
	}
	if ids, err = xstr.SplitInts(data); err != nil {
		log.Errorc(ctx, "PoolQuestionPage SplitInts data:%s error:%v", data, err)
	}
	return
}

// CacheUserInfo .
func (d *Dao) CacheUserInfo(ctx context.Context, mid int64, strWeek string) (res *like.AnswerUserInfo, err error) {
	var (
		key = userInfoKey(mid, strWeek)
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
func (d *Dao) AddCacheUserInfo(ctx context.Context, mid int64, strWeek string, data *like.AnswerUserInfo) (err error) {
	var (
		key = userInfoKey(mid, strWeek)
		bs  []byte
	)
	if bs, err = json.Marshal(data); err != nil {
		log.Error("AddCacheUserInfo json.Marshal(%v) error (%+v)", data, err)
		return
	}
	if _, err = component.GlobalRedis.Do(ctx, "SETEX", key, 5000000, bs); err != nil {
		log.Errorc(ctx, "AddCacheUserInfo conn.Send(SETEX, %s, %v, %s) error(%v)", key, 5000000, string(bs), err)
	}
	return
}

// AddCacheUserHp
func (d *Dao) AddCacheUserHp(ctx context.Context, mid, currentRound int64, data *like.AnswerHp) (err error) {
	var (
		key = userHpKey(mid, currentRound)
		bs  []byte
	)
	if bs, err = json.Marshal(data); err != nil {
		log.Error("AddCacheUserHp json.Marshal(%v) error (%+v)", data, err)
		return
	}
	if _, err = component.GlobalRedis.Do(ctx, "SETEX", key, 86400, bs); err != nil {
		log.Errorc(ctx, "AddCacheUserHp conn.Send(SETEX, %s, %v, %s) error(%v)", key, 86400, string(bs), err)
	}
	return
}

// CacheUserHp .
func (d *Dao) CacheUserHp(ctx context.Context, mid, currentRound int64) (res *like.AnswerHp, err error) {
	var (
		key = userHpKey(mid, currentRound)
		bs  []byte
	)
	if bs, err = redis.Bytes(component.GlobalRedis.Do(ctx, "GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			log.Warn("CacheUserHp(%s) return nil", key)
		} else {
			log.Errorc(ctx, "CacheUserHp conn.Do(GET key(%v)) error(%v)", key, err)
		}
		return
	}
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Errorc(ctx, "CacheUserHp json.Unmarshal(%v) error(%v)", bs, err)
	}
	return
}

// AddCacheKnowRule
func (d *Dao) AddCacheKnowRule(ctx context.Context, mid, val int64) (err error) {
	var key = knowRuleKey(mid)
	err = d.addCacheIn64(ctx, key, val)
	return
}

// CacheKnowRule .
func (d *Dao) CacheKnowRule(ctx context.Context, mid int64) (r int64, err error) {
	var key = knowRuleKey(mid)
	r, err = d.cacheInt64(ctx, key)
	return
}

func (d *Dao) addCacheIn64(ctx context.Context, key string, val int64) (err error) {
	if _, err = component.GlobalRedis.Do(ctx, "SETEX", key, d.answerExpire, val); err != nil { //57天
		log.Errorc(ctx, "AddCacheKnowRule conn.Send(SETEX, %s, %v, %s) error(%v)", key, d.answerExpire, 1, err)
	}
	return
}

func (d *Dao) cacheInt64(ctx context.Context, key string) (r int64, err error) {
	if r, err = redis.Int64(component.GlobalRedis.Do(ctx, "GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			log.Warn("cacheKnowRule(%s) return nil", key)
		} else {
			log.Errorc(ctx, "cacheKnowRule conn.Do(GET key(%v)) error(%v)", key, err)
		}
	}
	return
}

// CacheUserTop
func (d *Dao) CacheUserTop(ctx context.Context) (res []*like.UserRank, err error) {
	var (
		bs []byte
	)
	if bs, err = redis.Bytes(component.GlobalRedis.Do(ctx, "GET", _hourTop)); err != nil {
		if err == redis.ErrNil {
			err = nil
			log.Warn("CacheUserTop(%s) return nil", _hourTop)
		} else {
			log.Errorc(ctx, "CacheUserTop conn.Do(GET key(%v)) error(%v)", _hourTop, err)
		}
		return
	}
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Errorc(ctx, "CacheUserTop json.Unmarshal(%v) error(%v)", bs, err)
	}
	return
}

// CacheHourPeople
func (d *Dao) CacheHourPeople(ctx context.Context) (res *like.HourPeople, err error) {
	var (
		bs []byte
	)
	if bs, err = redis.Bytes(component.GlobalRedis.Do(ctx, "GET", _hourPeople)); err != nil {
		if err == redis.ErrNil {
			err = nil
			log.Warn("CacheHourPeople(%s) return nil", _hourPeople)
		} else {
			log.Errorc(ctx, "CacheHourPeople conn.Do(GET key(%v)) error(%v)", _hourPeople, err)
		}
		return
	}
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Errorc(ctx, "CacheHourPeople json.Unmarshal(%v) error(%v)", bs, err)
	}
	return
}

// CachePendantRule
func (d *Dao) CachePendantRule(ctx context.Context, mid int64) (res *like.PendantRule, err error) {
	var (
		key = pendantRuleKey(mid)
		bs  []byte
	)
	if bs, err = redis.Bytes(component.GlobalRedis.Do(ctx, "GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			log.Warn("CachePendantRule(%s) return nil", key)
		} else {
			log.Errorc(ctx, "CachePendantRule conn.Do(GET key(%v)) error(%v)", key, err)
		}
		return
	}
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Errorc(ctx, "CachePendantRule json.Unmarshal(%v) error(%v)", bs, err)
	}
	return
}

// AddCachePendantRule
func (d *Dao) AddCachePendantRule(ctx context.Context, mid int64, data *like.PendantRule) (err error) {
	var (
		key = pendantRuleKey(mid)
		bs  []byte
	)
	if bs, err = json.Marshal(data); err != nil {
		log.Error("AddCachePendantRule json.Marshal(%v) error (%+v)", data, err)
		return
	}
	if _, err = component.GlobalRedis.Do(ctx, "SETEX", key, d.answerExpire, bs); err != nil {
		log.Errorc(ctx, "AddCachePendantRule conn.Send(SETEX, %s, %v, %s) error(%v)", key, d.answerExpire, string(bs), err)
	}
	return
}

// CacheWeekTop
func (d *Dao) CacheWeekTop(ctx context.Context) (mids []int64, err error) {
	data, err := redis.String(component.GlobalRedis.Do(ctx, "GET", _roundTop))
	if err != nil {
		if err == redis.ErrNil {
			err = nil
			mids = make([]int64, 0)
			return
		}
		log.Errorc(ctx, "CacheWeekTop GET error:%v", err)
		return nil, err
	}
	if data == "" {
		return
	}
	if mids, err = xstr.SplitInts(data); err != nil {
		log.Errorc(ctx, "CacheWeekTop SplitInts data:%s error:%v", data, err)
	}
	return
}
