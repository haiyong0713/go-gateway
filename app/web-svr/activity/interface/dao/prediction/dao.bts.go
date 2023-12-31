// Code generated by kratos tool btsgen. DO NOT EDIT.

/*
  Package prediction is a generated cache proxy package.
  It is generated from:
  type _bts interface {
		// bts:-sync=true
		Predictions(c context.Context, ids []int64) (map[int64]*premdl.Prediction, error)
		// bts:-sync=true
		PredItems(c context.Context, ids []int64) (map[int64]*premdl.PredictionItem, error)
	}
*/

package prediction

import (
	"context"

	"go-common/library/cache"
	premdl "go-gateway/app/web-svr/activity/interface/model/prediction"
)

var _ _bts

// Predictions get data from cache if miss will call source method, then add to cache.
func (d *Dao) Predictions(c context.Context, ids []int64) (res map[int64]*premdl.Prediction, err error) {
	if len(ids) == 0 {
		return
	}
	addCache := true
	if res, err = d.CachePredictions(c, ids); err != nil {
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
	cache.MetricHits.Add(float64(len(ids)-len(miss)), "bts:Predictions")
	missLen := len(miss)
	if missLen == 0 {
		return
	}
	var missData map[int64]*premdl.Prediction
	cache.MetricMisses.Add(float64(len(miss)), "bts:Predictions")
	missData, err = d.RawPredictions(c, miss)
	if res == nil {
		res = make(map[int64]*premdl.Prediction, len(ids))
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
	d.AddCachePredictions(c, missData)
	return
}

// PredItems get data from cache if miss will call source method, then add to cache.
func (d *Dao) PredItems(c context.Context, ids []int64) (res map[int64]*premdl.PredictionItem, err error) {
	if len(ids) == 0 {
		return
	}
	addCache := true
	if res, err = d.CachePredItems(c, ids); err != nil {
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
	cache.MetricHits.Add(float64(len(ids)-len(miss)), "bts:PredItems")
	missLen := len(miss)
	if missLen == 0 {
		return
	}
	var missData map[int64]*premdl.PredictionItem
	cache.MetricMisses.Add(float64(len(miss)), "bts:PredItems")
	missData, err = d.RawPredItems(c, miss)
	if res == nil {
		res = make(map[int64]*premdl.PredictionItem, len(ids))
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
	d.AddCachePredItems(c, missData)
	return
}
