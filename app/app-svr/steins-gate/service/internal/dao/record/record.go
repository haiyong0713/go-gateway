package record

import (
	"context"

	"go-common/library/stat/prom"

	"go-gateway/app/app-svr/steins-gate/service/api"
	"go-gateway/app/app-svr/steins-gate/service/internal/model"
)

// Record get data from cache if miss will call source method, then add to cache.
func (d *Dao) Record(c context.Context, mid, graphID int64, buvid string) (res *api.GameRecords, err error) {
	addCache := true
	res, err = d.CacheRecord(c, mid, graphID, buvid)
	if err != nil {
		addCache = false
		err = nil
	}
	if res != nil {
		prom.CacheHit.Incr("Records")
		return
	}
	prom.CacheMiss.Incr("Records")
	res, err = d.RawRecord(c, mid, graphID, false)
	if err != nil || res == nil {
		return
	}
	miss := res
	miss.Buvid = buvid // 回填缓存增加buvid，否则会报错
	if !addCache {
		return
	}
	d.cache.Do(c, func(c context.Context) {
		//nolint:errcheck
		d.AddCacheRecord(c, miss, &model.NodeInfoParam{
			MobiApp: "bts_mobiApp",
		})
	})
	return
}

// Records def
func (d *Dao) Records(c context.Context, req *model.RecordReq) (res map[int64]*api.GameRecords, missAIDs []int64, err error) {
	if len(req.GraphWithAID) == 0 {
		return
	}
	addCache := true
	if res, err = d.CacheRecords(c, req.MID, req.PickGraphIDs(), req.Buvid); err != nil {
		addCache = false
		res = nil
		err = nil
	}
	var missGIDs []int64
	for graphID := range req.GraphWithAID {
		if (res == nil) || (res[graphID] == nil) {
			missGIDs = append(missGIDs, graphID)
		}
	}
	prom.CacheHit.Add("Record", int64(len(req.GraphWithAID)-len(missGIDs)))
	missLen := len(missGIDs)
	if missLen == 0 {
		return
	}
	prom.CacheMiss.Add("Record", int64(missLen))
	missData, err := d.RawRecords(c, req.MID, missGIDs, req.Buvid)
	if err != nil {
		return
	}
	if res == nil {
		res = make(map[int64]*api.GameRecords, len(req.GraphWithAID))
	}
	for _, graphID := range missGIDs { // 这里注意，有点怪的之前cache和db都是用gid查，gid查不到要返回对应的missAID，用于再查老版本的存档
		if rec, ok := missData[graphID]; !ok {
			missAIDs = append(missAIDs, req.GraphWithAID[graphID])
		} else {
			res[graphID] = rec
		}
	}
	if !addCache {
		return
	}
	d.cache.Do(c, func(c context.Context) {
		//nolint:errcheck
		d.AddCacheRecords(c, missData)
	})
	return

}
