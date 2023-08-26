// Code generated by kratos tool btsgen. DO NOT EDIT.

/*
  Package page is a generated cache proxy package.
  It is generated from:
  type _bts interface {
		// bts: -struct_name=Dao -nullcache=&model.ActPage{ID:0} -check_null_code=$==nil||$.ID==0 -sync=true
		GetPageByID(c context.Context, id int64) (*model.ActPage, error)
	}
*/

package page

import (
	"context"

	"go-common/library/cache"
	model "go-gateway/app/web-svr/activity/interface/model/page"
)

var _ _bts

// GetPageByID get data from cache if miss will call source method, then add to cache.
func (d *Dao) GetPageByID(c context.Context, id int64) (res *model.ActPage, err error) {
	addCache := true
	res, err = d.CacheGetPageByID(c, id)
	if err != nil {
		addCache = false
		err = nil
	}
	defer func() {
		if res == nil || res.ID == 0 {
			res = nil
		}
	}()
	if res != nil {
		cache.MetricHits.Inc("bts:GetPageByID")
		return
	}
	cache.MetricMisses.Inc("bts:GetPageByID")
	res, err = d.RawGetPageByID(c, id)
	if err != nil {
		return
	}
	miss := res
	if miss == nil {
		miss = &model.ActPage{ID: 0}
	}
	if !addCache {
		return
	}
	d.AddCacheGetPageByID(c, id, miss)
	return
}