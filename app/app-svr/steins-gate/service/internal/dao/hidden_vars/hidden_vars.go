package hidden_vars

import (
	"context"
	"encoding/json"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/stat/prom"

	"go-gateway/app/app-svr/steins-gate/service/api"
	"go-gateway/app/app-svr/steins-gate/service/internal/model"
)

// HiddenVars bts: -batch=50 -max_group=10 -batch_err=continue
func (d *Dao) HiddenVars(c context.Context, mid int64, graphInfo *api.GraphInfo, req *model.HvarReq, buvid string) (res *model.HiddenVarsRecord, err error) {
	if graphInfo == nil {
		log.Error("hvarsInfo Mid %d Buvid %s GraphInfo is Nil", mid, buvid)
		err = ecode.RequestErr
		return
	}
	res, err = d.hvarCache(c, mid, graphInfo.Id, req.CurrentID, req.CurrentCursorID, buvid)
	if err != nil {
		err = nil
	}
	if res != nil {
		prom.CacheHit.Incr("HiddenVars")
		return
	}
	prom.CacheMiss.Incr("HiddenVars")
	if res, err = d.rawHiddenVarsRec(c, mid, graphInfo.Id, req.CurrentID, req.CurrentCursorID); err == nil && res != nil {
		d.cache.Do(c, func(c context.Context) { // 需要异步去更新缓存
			//nolint:errcheck
			d.setHvarCache(c, mid, graphInfo.Id, req.CurrentID, req.CurrentCursorID, buvid, res)
		})
		return
	} // 如果数据库没有，需要根据record去重新计算
	res, err = d.hvarsInfo(c, mid, buvid, graphInfo, req)
	return
}

func (d *Dao) addHiddenVarsRecDBCache(c context.Context, recs map[int64]*model.HiddenVarsRecord, mid, gid int64, buvid string, cursorMap map[int64]int64) {
	for toN, rec := range recs { // 沿路设置隐藏变量存档
		var (
			copyRecord = new(model.HiddenVarsRecord)
			copyToN    = toN
		)
		copyRecord.DeepCopy(rec)
		d.cache.Do(c, func(ctx context.Context) { // 异步去补偿数据库和缓存
			//nolint:errcheck
			d.AddHiddenVarRecDBCache(ctx, mid, gid, copyToN, cursorMap[copyToN], buvid, copyRecord, false)
		})
	}
	//nolint:gosimple
	return
}

func (d *Dao) AddHiddenVarRecDBCache(c context.Context, mid, gid, toN, cursor int64, buvid string, rec *model.HiddenVarsRecord, cacheOnly bool) (err error) {
	var (
		vars  []*model.HiddenVar
		value []byte
	)
	if rec == nil {
		log.Error("AddHiddenVarRecDBCache Mid %d Gid %d Nid %d, Record is Null!", mid, gid, toN)
		return
	}
	for _, item := range rec.Vars {
		vars = append(vars, item)
	}
	if value, err = json.Marshal(vars); err != nil {
		return
	}
	if !cacheOnly {
		if err = d.addHiddenVarsRec(c, &model.HiddenVarRec{ // 先更新数据库再缓存
			MID:       mid,
			GraphID:   gid,
			CurrentID: toN,
			CursorID:  cursor,
			Value:     string(value),
		}); err != nil {
			log.Error("AddHiddenVarRecDBCache addHiddenVarsRec mid %d gid %d err(%+v) ", mid, gid, err)
			return
		}
	}
	var copyRecord = new(model.HiddenVarsRecord)
	copyRecord.DeepCopy(rec)
	// copy the record to save in cache
	if err = d.setHvarCache(c, mid, gid, toN, cursor, buvid, copyRecord); err != nil {
		log.Error("AddHiddenVarRecDBCache setHvarCache mid %d gid %d err(%+v) ", mid, gid, err)
	}
	return

}
