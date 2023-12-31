// Code generated by kratos tool btsgen. DO NOT EDIT.

/*
  Package dao is a generated cache proxy package.
  It is generated from:
  type _bts interface {
		// bts: -nullcache=&model.TopPhotoArc{Aid:-1} -check_null_code=$==nil||$.Aid==-1 -singleflight=true
		TopPhotoArc(c context.Context, mid int64) (*model.TopPhotoArc, error)
	}
*/

package dao

import (
	"context"

	"go-common/library/cache"
	"go-gateway/app/web-svr/space/interface/model"

	"golang.org/x/sync/singleflight"
)

var _ _bts
var cacheSingleFlights = [1]*singleflight.Group{{}}

// TopPhotoArc get data from cache if miss will call source method, then add to cache.
func (d *dao) TopPhotoArc(c context.Context, mid int64) (res *model.TopPhotoArc, err error) {
	addCache := true
	res, err = d.CacheTopPhotoArc(c, mid)
	if err != nil {
		addCache = false
		err = nil
	}
	defer func() {
		if res == nil || res.Aid == -1 {
			res = nil
		}
	}()
	if res != nil {
		cache.MetricHits.Inc("bts:TopPhotoArc")
		return
	}
	var rr interface{}
	sf := d.cacheSFTopPhotoArc(mid)
	rr, err, _ = cacheSingleFlights[0].Do(sf, func() (r interface{}, e error) {
		cache.MetricMisses.Inc("bts:TopPhotoArc")
		r, e = d.RawTopPhotoArc(c, mid)
		return
	})
	res = rr.(*model.TopPhotoArc)
	if err != nil {
		return
	}
	miss := res
	if miss == nil {
		miss = &model.TopPhotoArc{Aid: -1}
	}
	if !addCache {
		return
	}
	d.cache.Do(c, func(c context.Context) {
		d.AddCacheTopPhotoArc(c, mid, miss)
	})
	return
}
