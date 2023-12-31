// Code generated by kratos tool btsgen. DO NOT EDIT.

/*
  Package kfc is a generated cache proxy package.
  It is generated from:
  type _bts interface {
		// bts: -sync=true
		KfcCoupon(c context.Context, id int64) (*kfc.BnjKfcCoupon, error)
	}
*/

package kfc

import (
	"context"

	"go-common/library/cache"
	"go-gateway/app/web-svr/activity/interface/model/kfc"
)

var _ _bts

// KfcCoupon get data from cache if miss will call source method, then add to cache.
func (d *Dao) KfcCoupon(c context.Context, id int64) (res *kfc.BnjKfcCoupon, err error) {
	addCache := true
	res, err = d.CacheKfcCoupon(c, id)
	if err != nil {
		addCache = false
		err = nil
	}
	if res != nil {
		cache.MetricHits.Inc("bts:KfcCoupon")
		return
	}
	cache.MetricMisses.Inc("bts:KfcCoupon")
	res, err = d.RawKfcCoupon(c, id)
	if err != nil {
		return
	}
	miss := res
	if !addCache {
		return
	}
	d.AddCacheKfcCoupon(c, id, miss)
	return
}
