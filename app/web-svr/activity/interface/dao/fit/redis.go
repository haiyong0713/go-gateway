package fit

import (
	"context"
	"encoding/json"
	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/model/fit"
)

// CacheGetPlanList get all plans from cache.
func (d *dao) CacheGetPlanList(ctx context.Context) (res []*fit.PlanRecordRes, err error) {
	var (
		key = buildKey(planListKey)
		bs  []byte
	)
	if bs, err = redis.Bytes(d.redis.Do(ctx, "GET", key)); err != nil {
		log.Errorc(ctx, "CacheGetPlanList conn.Do(GET key(%v)) error(%v)", key, err)
		return
	}
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Errorc(ctx, "CacheGetPlanList json.Unmarshal(%s) error(%v)", string(bs), err)
	}
	return

}

// CacheSetPlanList set all plans into cache.
func (d *dao) CacheSetPlanList(ctx context.Context, data []*fit.PlanRecordRes) (err error) {
	var (
		key = buildKey(planListKey)
		bs  []byte
	)
	if bs, err = json.Marshal(data); err != nil {
		log.Errorc(ctx, "CacheSetPlanList json.Marshal(%v) error (%v)", data, err)
		return
	}
	if _, err = d.redis.Do(ctx, "SETEX", key, d.fitPlanListExpire, bs); err != nil {
		log.Errorc(ctx, "CacheSetPlanList conn.Send(SETEX, %s, %v, %s) error(%v)", key, d.fitPlanListExpire, string(bs), err)
	}
	return
}

// CacheGetPlanDeatailById get plan detail from cache.
func (d *dao) CacheGetPlanDeatailById(ctx context.Context, planId int64) (res *fit.PlanWeekBodanList, err error) {
	var (
		key = buildKey(planIdKey, planId)
		bs  []byte
	)
	if bs, err = redis.Bytes(d.redis.Do(ctx, "GET", key)); err != nil {
		log.Errorc(ctx, "CacheGetPlanDeatailById conn.Do(GET key(%v)) error(%v)", key, err)
		return
	}
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Errorc(ctx, "CacheGetPlanDeatailById json.Unmarshal(%s) error(%v)", string(bs), err)
	}
	return

}

// CacheSetPlanDeatailById set plan detail into cache.
func (d *dao) CacheSetPlanDeatailById(ctx context.Context, planId int64, data *fit.PlanWeekBodanList) (err error) {
	var (
		key = buildKey(planIdKey, planId)
		bs  []byte
	)
	if bs, err = json.Marshal(data); err != nil {
		log.Errorc(ctx, "CacheSetPlanDeatailById json.Marshal(%v) error (%v)", data, err)
		return
	}
	if _, err = d.redis.Do(ctx, "SETEX", key, d.fitPlanDetailByIdExpire, bs); err != nil {
		log.Errorc(ctx, "CacheSetPlanDeatailById conn.Send(SETEX, %s, %v, %s) error(%v)", key, d.fitPlanDetailByIdExpire, string(bs), err)
	}
	return
}
