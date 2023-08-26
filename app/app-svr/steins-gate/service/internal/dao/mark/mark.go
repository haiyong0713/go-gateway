package mark

import (
	"context"
	"sync"

	"go-common/library/cache"
	"go-common/library/log"
	"go-common/library/stat/prom"
	"go-common/library/sync/errgroup.v2"
)

// Mark get data from cache if miss will call source method, then add to cache.
func (d *Dao) Mark(c context.Context, aid int64, mid int64) (res int64, err error) {
	addCache := true
	res, err = d.CacheMark(c, aid, mid)
	if err != nil {
		addCache = false
		err = nil
	}
	defer func() {
		if res == -1 {
			res = 0
		}
	}()
	if res != 0 {
		cache.MetricHits.Inc("bts:Mark")
		return
	}
	cache.MetricMisses.Inc("bts:Mark")
	res, err = d.rawMark(c, aid, mid)
	if err != nil {
		return
	}
	miss := res
	if miss == 0 {
		miss = -1
	}
	if !addCache {
		return
	}
	//nolint:errcheck
	d.AddCacheMark(c, aid, mid, miss)
	return
}

func (d *Dao) AddMark(c context.Context, aid, mid, mark int64) (err error) {
	if err = d.addMark(c, aid, mid, mark); err != nil {
		log.Error("d.AddMark(aid:%v,mid:%v,mark:%v) error(%v)", aid, mid, mark, err)
		return
	}
	if markCacheErr := d.fanout.Do(c, func(ctx context.Context) {
		//nolint:errcheck
		d.AddCacheMark(ctx, aid, mid, mark)
	}); markCacheErr != nil {
		log.Error("d.AddCacheMark(aid:%v,mid:%v,mark:%v) error(%v)", aid, mid, mark, err)
	}
	return
}

// Evaluation get data from cache if miss will call source method, then add to cache.
func (d *Dao) Evaluation(c context.Context, aid int64) (res int64, err error) {
	addCache := true
	res, err = d.CacheEvaluation(c, aid)
	if err != nil {
		addCache = false
		err = nil
	}
	defer func() {
		if res == -1 {
			res = 0
		}
	}()
	if res != 0 {
		cache.MetricHits.Inc("bts:Evaluation")
		return
	}
	cache.MetricMisses.Inc("bts:Evaluation")
	res, err = d.rawEvaluation(c, aid)
	if err != nil {
		return
	}
	miss := res
	if miss == 0 {
		miss = -1
	}
	if !addCache {
		return
	}
	//nolint:errcheck
	d.AddCacheEvaluation(c, aid, miss)
	return
}

func (d *Dao) Evaluations(c context.Context, keys []int64) (res map[int64]int64, err error) {
	if len(keys) == 0 {
		return
	}
	addCache := true
	if res, err = d.CacheEvaluations(c, keys); err != nil {
		addCache = false
		res = nil
		err = nil
	}
	var miss []int64
	for _, key := range keys {
		if (res == nil) || (res[key] == 0) {
			miss = append(miss, key)
		}
		if (res != nil) && res[key] == -1 {
			res[key] = 0
		}
	}
	prom.CacheHit.Add("Evaluations", int64(len(keys)-len(miss)))
	missLen := len(miss)
	if missLen == 0 {
		return
	}
	missData := make(map[int64]int64, missLen)
	prom.CacheMiss.Add("Evaluations", int64(missLen))
	var mutex sync.Mutex
	group := errgroup.WithContext(c)
	//nolint:gomnd
	if missLen > 10 {
		group.GOMAXPROCS(10)
	}
	var run = func(ms []int64) {
		group.Go(func(ctx context.Context) (err error) {
			data, err := d.rawEvaluations(ctx, ms)
			mutex.Lock()
			for k, v := range data {
				missData[k] = v
			}
			mutex.Unlock()
			return
		})
	}
	var (
		i int
		n = missLen / 50
	)
	for i = 0; i < n; i++ {
		run(miss[i*50 : (i+1)*50])
	}
	if len(miss[i*50:]) > 0 {
		run(miss[i*50:])
	}
	err = group.Wait()
	if res == nil {
		res = make(map[int64]int64, len(keys))
	}
	for k, v := range missData {
		if v == 0 {
			v = -1
		}
		res[k] = v
	}
	if err != nil {
		return
	}
	if !addCache {
		return
	}
	//nolint:errcheck
	d.fanout.Do(c, func(c context.Context) {
		//nolint:errcheck
		d.AddCacheEvaluations(c, missData)
	})
	return
}

func (d *Dao) Marks(c context.Context, keys []int64, mid int64) (res map[int64]int64, err error) {
	if len(keys) == 0 {
		return
	}
	addCache := true
	if res, err = d.CacheMarks(c, keys, mid); err != nil {
		addCache = false
		res = nil
		err = nil
	}
	var miss []int64
	for _, key := range keys {
		if (res == nil) || (res[key] == 0) {
			miss = append(miss, key)
		}
		if (res != nil) && res[key] == -1 {
			res[key] = 0
		}
	}
	prom.CacheHit.Add("Marks", int64(len(keys)-len(miss)))
	missLen := len(miss)
	if missLen == 0 {
		return
	}
	missData := make(map[int64]int64, missLen)
	prom.CacheMiss.Add("Marks", int64(missLen))
	var mutex sync.Mutex
	group := errgroup.WithContext(c)
	//nolint:gomnd
	if missLen > 10 {
		group.GOMAXPROCS(10)
	}
	var run = func(ms []int64) {
		group.Go(func(ctx context.Context) (err error) {
			data, err := d.rawMarksM(ctx, ms, mid)
			mutex.Lock()
			for k, v := range data {
				missData[k] = v
			}
			mutex.Unlock()
			return
		})
	}
	var (
		i int
		n = missLen / 50
	)
	for i = 0; i < n; i++ {
		run(miss[i*50 : (i+1)*50])
	}
	if len(miss[i*50:]) > 0 {
		run(miss[i*50:])
	}
	err = group.Wait()
	if res == nil {
		res = make(map[int64]int64, len(keys))
	}
	for k, v := range missData {
		if v == 0 {
			v = -1
		}
		res[k] = v
	}
	if err != nil {
		return
	}
	if !addCache {
		return
	}
	//nolint:errcheck
	d.fanout.Do(c, func(c context.Context) {
		//nolint:errcheck
		d.AddCacheMarks(c, missData, mid)
	})
	return

}
