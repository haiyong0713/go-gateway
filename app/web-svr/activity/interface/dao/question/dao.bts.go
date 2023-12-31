// Code generated by kratos tool btsgen. DO NOT EDIT.

/*
  Package question is a generated cache proxy package.
  It is generated from:
  type _bts interface {
		// get question detail info
		// bts: -struct_name=Dao
		Detail(c context.Context, id int64) (*question.Detail, error)
		// get question details info
		// bts: -struct_name=Dao
		Details(c context.Context, ids []int64) (map[int64]*question.Detail, error)
		// get user last answer log
		// bts: -struct_name=Dao
		LastQuesLog(c context.Context, mid int64, baseID int64) (*question.UserAnswerLog, error)
	}
*/

package question

import (
	"context"

	"go-common/library/cache"
	"go-gateway/app/web-svr/activity/interface/model/question"
)

var _ _bts

// Detail get question detail info
func (d *Dao) Detail(c context.Context, id int64) (res *question.Detail, err error) {
	addCache := true
	res, err = d.CacheDetail(c, id)
	if err != nil {
		addCache = false
		err = nil
	}
	if res != nil {
		cache.MetricHits.Inc("bts:Detail")
		return
	}
	cache.MetricMisses.Inc("bts:Detail")
	res, err = d.RawDetail(c, id)
	if err != nil {
		return
	}
	miss := res
	if !addCache {
		return
	}
	d.cache.Do(c, func(c context.Context) {
		d.AddCacheDetail(c, id, miss)
	})
	return
}

// Details get question details info
func (d *Dao) Details(c context.Context, ids []int64) (res map[int64]*question.Detail, err error) {
	if len(ids) == 0 {
		return
	}
	addCache := true
	if res, err = d.CacheDetails(c, ids); err != nil {
		addCache = false
		res = nil
		err = nil
	}
	var miss []int64
	for _, key := range ids {
		if (res == nil) || (res[key] == nil) {
			miss = append(miss, key)
		}
	}
	cache.MetricHits.Add(float64(len(ids)-len(miss)), "bts:Details")
	missLen := len(miss)
	if missLen == 0 {
		return
	}
	var missData map[int64]*question.Detail
	cache.MetricMisses.Add(float64(len(miss)), "bts:Details")
	missData, err = d.RawDetails(c, miss)
	if res == nil {
		res = make(map[int64]*question.Detail, len(ids))
	}
	for k, v := range missData {
		res[k] = v
	}
	if err != nil {
		return
	}
	if !addCache {
		return
	}
	d.cache.Do(c, func(c context.Context) {
		d.AddCacheDetails(c, missData)
	})
	return
}

// LastQuesLog get user last answer log
func (d *Dao) LastQuesLog(c context.Context, mid int64, baseID int64) (res *question.UserAnswerLog, err error) {
	addCache := true
	res, err = d.CacheLastQuesLog(c, mid, baseID)
	if err != nil {
		addCache = false
		err = nil
	}
	if res != nil {
		cache.MetricHits.Inc("bts:LastQuesLog")
		return
	}
	cache.MetricMisses.Inc("bts:LastQuesLog")
	res, err = d.RawLastQuesLog(c, mid, baseID)
	if err != nil {
		return
	}
	miss := res
	if !addCache {
		return
	}
	d.cache.Do(c, func(c context.Context) {
		d.AddCacheLastQuesLog(c, mid, miss, baseID)
	})
	return
}
