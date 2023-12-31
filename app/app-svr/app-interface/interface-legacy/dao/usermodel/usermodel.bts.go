// Code generated by kratos tool btsgen. DO NOT EDIT.

/*
  Package usermodel is a generated cache proxy package.
  It is generated from:
  type Dao interface {
		Close()
		// bts: -nullcache=[]*usermodel.User{{ID:-1}} -check_null_code=len($)==1&&$[0]!=nil&&$[0].ID==-1 -singleflight=true
		UserModels(ctx context.Context, mid int64, mobiApp string, deviceToken string) ([]*usermodel.User, error)
		AddUserModel(ctx context.Context, user *usermodel.User, sync bool) (userID int64, devID int64, err error)
		AddSpecialModeLog(ctx context.Context, log *usermodel.SpecialModeLog) error
		SetCacheAntiAddictionTime(ctx context.Context, deviceToken string, mid, day, useTime int64) error
		GetCacheAntiAddictionTime(ctx context.Context, deviceToken string, mid, day int64) (int64, error)
		GetCacheAntiAddictionMID(ctx context.Context, deviceToken string, day int64) (int64, error)
		AddManualForceLog(ctx context.Context, log *usermodel.ManualForceLog) error
		UpdateManualForceAndCache(ctx context.Context, user *usermodel.User) error
		// bts: -nullcache=[]*familymdl.FamilyRelation{{ID:-1}} -check_null_code=len($)==1&&$[0]!=nil&&$[0].ID==-1
		FamilyRelsOfParent(ctx context.Context, parentMid int64) ([]*familymdl.FamilyRelation, error)
		// bts: -nullcache=&familymdl.FamilyRelation{ID:-1} -check_null_code=$!=nil&&$.ID==-1
		FamilyRelsOfChild(ctx context.Context, childMid int64) (*familymdl.FamilyRelation, error)
		UnbindFamily(ctx context.Context, rel *familymdl.FamilyRelation) error
		BindFamily(ctx context.Context, pmid, cmid, duration int64) error
		LatestFamilyRel(ctx context.Context, pmid int64, cmid int64) (*familymdl.FamilyRelation, error)
		AddFamilyLogs(ctx context.Context, items []*familymdl.FamilyLog) error
		UpdateTimelock(ctx context.Context, rel *familymdl.FamilyRelation) error
		// bts: -nullcache=&antiadmdl.SleepRemind{ID:-1} -check_null_code=$!=nil&&$.ID==-1
		SleepRemind(ctx context.Context, mid int64) (*antiadmdl.SleepRemind, error)
		AddSleepRemind(ctx context.Context, val *antiadmdl.SleepRemind) error
		UpdateSleepRemind(ctx context.Context, val *antiadmdl.SleepRemind) error
	}
*/

package usermodel

import (
	"context"

	"go-common/library/cache"
	antiadmdl "go-gateway/app/app-svr/app-interface/interface-legacy/model/anti_addiction"
	familymdl "go-gateway/app/app-svr/app-interface/interface-legacy/model/family"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model/usermodel"

	"golang.org/x/sync/singleflight"
)

var cacheSingleFlights = [1]*singleflight.Group{{}}

// UserModels get data from cache if miss will call source method, then add to cache.
func (d *dao) UserModels(c context.Context, mid int64, mobiApp string, deviceToken string) (res []*usermodel.User, err error) {
	addCache := true
	res, err = d.CacheUserModels(c, mid, mobiApp, deviceToken)
	if err != nil {
		addCache = false
		err = nil
	}
	defer func() {
		if len(res) == 1 && res[0] != nil && res[0].ID == -1 {
			res = nil
		}
	}()
	if len(res) != 0 {
		cache.MetricHits.Inc("bts:UserModels")
		return
	}
	var rr interface{}
	sf := d.cacheSFUserModels(mid, mobiApp, deviceToken)
	rr, err, _ = cacheSingleFlights[0].Do(sf, func() (r interface{}, e error) {
		cache.MetricMisses.Inc("bts:UserModels")
		r, e = d.RawUserModels(c, mid, mobiApp, deviceToken)
		return
	})
	res = rr.([]*usermodel.User)
	if err != nil {
		return
	}
	miss := res
	if len(miss) == 0 {
		miss = []*usermodel.User{{ID: -1}}
	}
	if !addCache {
		return
	}
	d.cache.Do(c, func(c context.Context) {
		d.AddCacheUserModels(c, mid, miss, mobiApp, deviceToken)
	})
	return
}

// FamilyRelsOfParent get data from cache if miss will call source method, then add to cache.
func (d *dao) FamilyRelsOfParent(c context.Context, parentMid int64) (res []*familymdl.FamilyRelation, err error) {
	addCache := true
	res, err = d.CacheFamilyRelsOfParent(c, parentMid)
	if err != nil {
		addCache = false
		err = nil
	}
	defer func() {
		if len(res) == 1 && res[0] != nil && res[0].ID == -1 {
			res = nil
		}
	}()
	if len(res) != 0 {
		cache.MetricHits.Inc("bts:FamilyRelsOfParent")
		return
	}
	cache.MetricMisses.Inc("bts:FamilyRelsOfParent")
	res, err = d.RawFamilyRelsOfParent(c, parentMid)
	if err != nil {
		return
	}
	miss := res
	if len(miss) == 0 {
		miss = []*familymdl.FamilyRelation{{ID: -1}}
	}
	if !addCache {
		return
	}
	d.cache.Do(c, func(c context.Context) {
		d.AddCacheFamilyRelsOfParent(c, parentMid, miss)
	})
	return
}

// FamilyRelsOfChild get data from cache if miss will call source method, then add to cache.
func (d *dao) FamilyRelsOfChild(c context.Context, childMid int64) (res *familymdl.FamilyRelation, err error) {
	addCache := true
	res, err = d.CacheFamilyRelsOfChild(c, childMid)
	if err != nil {
		addCache = false
		err = nil
	}
	defer func() {
		if res != nil && res.ID == -1 {
			res = nil
		}
	}()
	if res != nil {
		cache.MetricHits.Inc("bts:FamilyRelsOfChild")
		return
	}
	cache.MetricMisses.Inc("bts:FamilyRelsOfChild")
	res, err = d.RawFamilyRelsOfChild(c, childMid)
	if err != nil {
		return
	}
	miss := res
	if miss == nil {
		miss = &familymdl.FamilyRelation{ID: -1}
	}
	if !addCache {
		return
	}
	d.cache.Do(c, func(c context.Context) {
		d.AddCacheFamilyRelsOfChild(c, childMid, miss)
	})
	return
}

// SleepRemind get data from cache if miss will call source method, then add to cache.
func (d *dao) SleepRemind(c context.Context, mid int64) (res *antiadmdl.SleepRemind, err error) {
	addCache := true
	res, err = d.CacheSleepRemind(c, mid)
	if err != nil {
		addCache = false
		err = nil
	}
	defer func() {
		if res != nil && res.ID == -1 {
			res = nil
		}
	}()
	if res != nil {
		cache.MetricHits.Inc("bts:SleepRemind")
		return
	}
	cache.MetricMisses.Inc("bts:SleepRemind")
	res, err = d.RawSleepRemind(c, mid)
	if err != nil {
		return
	}
	miss := res
	if miss == nil {
		miss = &antiadmdl.SleepRemind{ID: -1}
	}
	if !addCache {
		return
	}
	d.cache.Do(c, func(c context.Context) {
		d.AddCacheSleepRemind(c, mid, miss)
	})
	return
}
