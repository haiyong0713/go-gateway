package dao

import (
	"context"
	"encoding/json"
	"time"

	"go-common/library/cache"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/stat/metric"
	"go-gateway/app/app-svr/up-archive/service/api"
	"go-gateway/app/app-svr/up-archive/service/internal/model"
)

var (
	_metricTotal = metric.NewBusinessMetricCount("es_query_total", "method", "caller")
)

func (d *dao) ArcSearch(ctx context.Context, mid int64, tid int64, keyword string, kwFields []string, highlight bool, pn int, ps int, order api.SearchOrder, without []api.Without, sort api.Sort) (*model.ArcPassedSearchReply, error) {
	const _prefix = "degrade_search"
	key := cacheSFSearch(_prefix, mid, tid, keyword, kwFields, highlight, pn, ps, order, without, sort)
	f := func(ctx context.Context) (interface{}, error) {
		return d.arcPassedSearch(ctx, mid, tid, keyword, kwFields, highlight, pn, ps, order, without, sort)
	}
	bf := func(data json.RawMessage) (interface{}, error) {
		var reply *model.ArcPassedSearchReply
		err := json.Unmarshal(data, &reply)
		return reply, err
	}
	reply, err := d.ArcSearchDegrade(ctx, key, f, bf)
	if err != nil {
		return nil, err
	}
	return reply.(*model.ArcPassedSearchReply), nil
}

func (d *dao) ArcSearchTag(ctx context.Context, mid int64, keyword string, kwFields []string, without []api.Without) (map[int64]int64, error) {
	const _prefix = "degrade_search_tag"
	key := cacheSFSearch(_prefix, mid, keyword, kwFields, without)
	f := func(ctx context.Context) (interface{}, error) {
		return d.arcPassedSearchTag(ctx, mid, keyword, kwFields, without)
	}
	bf := func(data json.RawMessage) (interface{}, error) {
		var reply map[int64]int64
		err := json.Unmarshal(data, &reply)
		return reply, err
	}
	reply, err := d.ArcSearchDegrade(ctx, key, f, bf)
	if err != nil {
		return nil, err
	}
	return reply.(map[int64]int64), nil
}

func (d *dao) ArcSearchCursor(ctx context.Context, mid, score int64, containScore bool, ps int, without []api.Without, sort api.Sort) (*model.ArcPassedSearchReply, error) {
	const _prefix = "degrade_search_cursor"
	key := cacheSFSearch(_prefix, mid, score, containScore, ps, without, sort)
	f := func(ctx context.Context) (interface{}, error) {
		return d.arcPassedSearchCursor(ctx, mid, score, containScore, ps, without, sort)
	}
	bf := func(data json.RawMessage) (interface{}, error) {
		var reply *model.ArcPassedSearchReply
		err := json.Unmarshal(data, &reply)
		return reply, err
	}
	reply, err := d.ArcSearchDegrade(ctx, key, f, bf)
	if err != nil {
		return nil, err
	}
	return reply.(*model.ArcPassedSearchReply), nil
}

func (d *dao) ArcSearchCursorAid(ctx context.Context, mid, score int64, equalScore bool, aid int64, tid int64, ps int, order api.SearchOrder, without []api.Without, sort api.Sort) (*model.ArcCursorAidSearchReply, error) {
	const _prefix = "degrade_search_cursor_aid"
	key := cacheSFSearch(_prefix, mid, score, equalScore, aid, tid, ps, order, without, sort)
	f := func(ctx context.Context) (interface{}, error) {
		return d.arcPassedSearchCursorAid(ctx, mid, score, equalScore, aid, tid, ps, order, without, sort)
	}
	bf := func(data json.RawMessage) (interface{}, error) {
		var reply *model.ArcCursorAidSearchReply
		err := json.Unmarshal(data, &reply)
		return reply, err
	}
	reply, err := d.ArcSearchDegrade(ctx, key, f, bf)
	if err != nil {
		return nil, err
	}
	return reply.(*model.ArcCursorAidSearchReply), nil
}

func (d *dao) ArcSearchScore(ctx context.Context, mid, aid, tid int64, order api.SearchOrder, without []api.Without) (*model.ArcScoreResult, error) {
	const _prefix = "degrade_search_score"
	key := cacheSFSearch(_prefix, mid, aid, tid, order, without)
	f := func(ctx context.Context) (interface{}, error) {
		return d.arcPassedSearchScore(ctx, mid, aid, tid, order, without)
	}
	bf := func(data json.RawMessage) (interface{}, error) {
		var reply *model.ArcScoreResult
		err := json.Unmarshal(data, &reply)
		return reply, err
	}
	reply, err := d.ArcSearchDegrade(ctx, key, f, bf)
	if err != nil {
		return nil, err
	}
	return reply.(*model.ArcScoreResult), nil
}

func (d *dao) ArcsSearchSort(ctx context.Context, mids []int64, tid int64, ps int, order api.SearchOrder, sort api.Sort) (map[int64][]int64, error) {
	const _prefix = "degrade_search_sort"
	key := cacheSFSearch(_prefix, mids, tid, ps, order, sort)
	f := func(ctx context.Context) (interface{}, error) {
		return d.arcsPassedSearchSort(ctx, mids, tid, ps, order, sort)
	}
	bf := func(data json.RawMessage) (interface{}, error) {
		var reply map[int64][]int64
		err := json.Unmarshal(data, &reply)
		return reply, err
	}
	reply, err := d.ArcSearchDegrade(ctx, key, f, bf)
	if err != nil {
		return nil, err
	}
	return reply.(map[int64][]int64), nil
}

func (d *dao) ArcSearchDegrade(ctx context.Context, key string, f func(ctx context.Context) (interface{}, error), bf func(data json.RawMessage) (interface{}, error)) (interface{}, error) {
	var (
		backupReply interface{}
		legal       bool
	)
	addCache := true
	func() {
		searchCache, err := d.CacheArcSearch(ctx, key)
		if err != nil {
			log.Error("%+v", err)
			addCache = false
			return
		}
		if searchCache == nil {
			return
		}
		legal = true
		cache.MetricHits.Inc("bts:ArcSearch")
		if addCache {
			duration, err := d.ac.Get("DegradeDuration").Duration()
			if err != nil {
				log.Error("%+v", err)
				duration = time.Hour
			}
			addCache = time.Now().Sub(searchCache.Ctime) > duration
		}
		if searchCache.Reply != nil {
			if backupReply, err = bf(searchCache.Reply); err != nil {
				log.Error("%+v", err)
				legal = false
				addCache = true
			}
		}
	}()
	if legal && !addCache {
		return backupReply, nil
	}
	method := metadata.String(ctx, metadata.FullMethod)
	caller := metadata.String(ctx, metadata.Caller)
	if caller == "" {
		caller = "no_user"
	}
	_metricTotal.Inc(method, caller)
	reply, err := f(ctx)
	if err != nil {
		log.Error("%+v", err)
		if legal {
			return backupReply, nil
		}
		return nil, err
	}
	if !addCache {
		return reply, nil
	}
	cache.MetricMisses.Inc("bts:ArcSearch")
	miss := reply
	d.cache.Do(ctx, func(ctx context.Context) {
		if err := d.AddCacheArcSearch(ctx, key, miss); err != nil {
			log.Error("%+v", err)
		}
	})
	return reply, nil
}
