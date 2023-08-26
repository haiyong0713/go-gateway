// Code generated by kratos tool mcgen. DO NOT EDIT.

/*
  Package prediction is a generated mc cache package.
  It is generated from:
  type _mc interface {
		// mc: -key=predictionKey
		CachePredictions(c context.Context, ids []int64) (res map[int64]*premdl.Prediction, err error)
		// mc: -key=predictionKey -expire=d.mcPerpetualExpire -encode=pb
		AddCachePredictions(c context.Context, val map[int64]*premdl.Prediction) error
		// mc: -key=predictionKey
		DelCachePredictions(c context.Context, ids []int64) error
		// mc: -key=predItemKey
		CachePredItems(c context.Context, ids []int64) (res map[int64]*premdl.PredictionItem, err error)
		// mc: -key=predItemKey -expire=d.mcPerpetualExpire -encode=pb
		AddCachePredItems(c context.Context, val map[int64]*premdl.PredictionItem) error
		// mc: -key=predItemKey
		DelCachePredItems(c context.Context, ids []int64) error
	}
*/

package prediction

import (
	"context"
	"fmt"

	"go-common/library/cache/memcache"
	"go-common/library/log"
	premdl "go-gateway/app/web-svr/activity/interface/model/prediction"
)

var _ _mc

// CachePredictions get data from mc
func (d *Dao) CachePredictions(c context.Context, ids []int64) (res map[int64]*premdl.Prediction, err error) {
	l := len(ids)
	if l == 0 {
		return
	}
	keysMap := make(map[string]int64, l)
	keys := make([]string, 0, l)
	for _, id := range ids {
		key := predictionKey(id)
		keysMap[key] = id
		keys = append(keys, key)
	}
	replies, err := d.mc.GetMulti(c, keys)
	if err != nil {
		log.Errorv(c, log.KV("CachePredictions", fmt.Sprintf("%+v", err)), log.KV("keys", keys))
		return
	}
	for _, key := range replies.Keys() {
		v := &premdl.Prediction{}
		err = replies.Scan(key, v)
		if err != nil {
			log.Errorv(c, log.KV("CachePredictions", fmt.Sprintf("%+v", err)), log.KV("key", key))
			return
		}
		if res == nil {
			res = make(map[int64]*premdl.Prediction, len(keys))
		}
		res[keysMap[key]] = v
	}
	return
}

// AddCachePredictions Set data to mc
func (d *Dao) AddCachePredictions(c context.Context, values map[int64]*premdl.Prediction) (err error) {
	if len(values) == 0 {
		return
	}
	for id, val := range values {
		key := predictionKey(id)
		item := &memcache.Item{Key: key, Object: val, Expiration: d.mcPerpetualExpire, Flags: memcache.FlagProtobuf}
		if err = d.mc.Set(c, item); err != nil {
			log.Errorv(c, log.KV("AddCachePredictions", fmt.Sprintf("%+v", err)), log.KV("key", key))
			return
		}
	}
	return
}

// DelCachePredictions delete data from mc
func (d *Dao) DelCachePredictions(c context.Context, ids []int64) (err error) {
	if len(ids) == 0 {
		return
	}
	for _, id := range ids {
		key := predictionKey(id)
		if err = d.mc.Delete(c, key); err != nil {
			if err == memcache.ErrNotFound {
				err = nil
				continue
			}
			log.Errorv(c, log.KV("DelCachePredictions", fmt.Sprintf("%+v", err)), log.KV("key", key))
			return
		}
	}
	return
}

// CachePredItems get data from mc
func (d *Dao) CachePredItems(c context.Context, ids []int64) (res map[int64]*premdl.PredictionItem, err error) {
	l := len(ids)
	if l == 0 {
		return
	}
	keysMap := make(map[string]int64, l)
	keys := make([]string, 0, l)
	for _, id := range ids {
		key := predItemKey(id)
		keysMap[key] = id
		keys = append(keys, key)
	}
	replies, err := d.mc.GetMulti(c, keys)
	if err != nil {
		log.Errorv(c, log.KV("CachePredItems", fmt.Sprintf("%+v", err)), log.KV("keys", keys))
		return
	}
	for _, key := range replies.Keys() {
		v := &premdl.PredictionItem{}
		err = replies.Scan(key, v)
		if err != nil {
			log.Errorv(c, log.KV("CachePredItems", fmt.Sprintf("%+v", err)), log.KV("key", key))
			return
		}
		if res == nil {
			res = make(map[int64]*premdl.PredictionItem, len(keys))
		}
		res[keysMap[key]] = v
	}
	return
}

// AddCachePredItems Set data to mc
func (d *Dao) AddCachePredItems(c context.Context, values map[int64]*premdl.PredictionItem) (err error) {
	if len(values) == 0 {
		return
	}
	for id, val := range values {
		key := predItemKey(id)
		item := &memcache.Item{Key: key, Object: val, Expiration: d.mcPerpetualExpire, Flags: memcache.FlagProtobuf}
		if err = d.mc.Set(c, item); err != nil {
			log.Errorv(c, log.KV("AddCachePredItems", fmt.Sprintf("%+v", err)), log.KV("key", key))
			return
		}
	}
	return
}

// DelCachePredItems delete data from mc
func (d *Dao) DelCachePredItems(c context.Context, ids []int64) (err error) {
	if len(ids) == 0 {
		return
	}
	for _, id := range ids {
		key := predItemKey(id)
		if err = d.mc.Delete(c, key); err != nil {
			if err == memcache.ErrNotFound {
				err = nil
				continue
			}
			log.Errorv(c, log.KV("DelCachePredItems", fmt.Sprintf("%+v", err)), log.KV("key", key))
			return
		}
	}
	return
}