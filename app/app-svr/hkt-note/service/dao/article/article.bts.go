package article

import (
	"context"

	"go-common/library/cache"
	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	"go-common/library/log"
	xecode "go-gateway/app/app-svr/hkt-note/ecode"
	"go-gateway/app/app-svr/hkt-note/service/model/article"

	"github.com/pkg/errors"
)

func (d *Dao) ArtDetails(c context.Context, keys []int64, tp string) (map[int64]*article.ArtDtlCache, error) {
	if len(keys) == 0 {
		return make(map[int64]*article.ArtDtlCache), nil
	}
	addCache := true
	var miss []int64
	res, miss, err := d.cacheArtDetails(c, keys, tp)
	if err != nil {
		addCache = false
		res = nil
	}
	for _, key := range keys {
		if res == nil || res[key] == nil {
			miss = append(miss, key)
		}
	}
	for k, v := range res {
		if v != nil && v.Cvid == -1 {
			delete(res, k)
		}
	}
	if len(miss) == 0 {
		if res == nil {
			return nil, errors.Wrapf(xecode.ArtDetailNotFound, "ArtDetails keys(%v) tp(%s)", keys, tp)
		}
		return res, nil
	}
	var (
		missData    map[int64]*article.ArtDtlCache
		pubStatusEq int
	)
	if tp == article.TpArtDetailCvid {
		pubStatusEq = article.PubStatusPassed
	}
	missData, err = d.rawArtDetails(c, miss, tp, pubStatusEq, 0)
	if err != nil {
		log.Error("artWarn ArtDetails miss(%v) err(%+v)", miss, err)
		if res == nil {
			return nil, err
		}
	}
	if res == nil {
		res = make(map[int64]*article.ArtDtlCache, len(keys))
	}
	for k, v := range missData {
		res[k] = v
	}
	if missData == nil {
		missData = map[int64]*article.ArtDtlCache{}
	}
	for _, key := range miss {
		if res[key] == nil {
			missData[key] = &article.ArtDtlCache{Cvid: -1}
		}
	}
	if !addCache {
		return res, nil
	}
	for k, v := range missData {
		var (
			curId  = k
			curVal = v
		)
		d.cache.Do(c, func(ctx context.Context) {
			if e := d.addCacheArtDetail(ctx, curId, tp, curVal); e != nil {
				log.Warn("artWarn ArtDetails err(%+v)", e)
			}
		})
	}
	return res, nil
}

func (d *Dao) ArtCountInUser(c context.Context, mid int64) (int, error) {
	addCache := true
	count, err := d.cacheArtListCount(c, d.userListKey(mid))
	if err != nil && err != redis.ErrNil {
		log.Warn("artWarn err(%+v)", err)
		addCache = false
	}
	if count > 0 {
		cache.MetricHits.Inc("bts:ArtCountInUser")
		return count, nil
	}
	cache.MetricMisses.Inc("bts:ArtCountInUser")
	var artList []*article.ArtList
	artList, _, err = d.rawArtListInUser(c, 0, -1, mid)
	if err != nil && err != sql.ErrNoRows {
		return 0, errors.Wrapf(err, "ArtCountInUser mid(%d)", mid)
	}
	if !addCache {
		return len(artList), nil
	}
	d.cache.Do(c, func(ctx context.Context) {
		if e := d.addCacheArtList(ctx, d.userListKey(mid), artList); e != nil {
			log.Warn("artWarn ArtCountInUser err(%+v)", e)
		}
	})
	return len(artList), nil
}

func (d *Dao) ArtCountInArc(c context.Context, oid, oidType int64) (int64, error) {
	addCache := true
	count, err := d.cacheArtCntInArc(c, oid, oidType)
	if err != nil && err != redis.ErrNil {
		log.Warn("artWarn err(%+v)", err)
		addCache = false
	}
	if count > 0 || count == -1 {
		cache.MetricHits.Inc("bts:ArtCountInArc")
		if count == -1 {
			return 0, nil
		}
		return count, nil
	}
	cache.MetricMisses.Inc("bts:ArtCountInArc")
	count, err = d.rawArtCountInArc(c, oid, oidType)
	if err != nil {
		return 0, errors.Wrapf(err, "ArtCountInArc oid(%d) oidType(%d)", oid, oidType)
	}
	if !addCache {
		return count, nil
	}
	cacheCount := count
	if cacheCount == 0 {
		// 如果从db中捞出来是0，缓存中放入-1表示该oid下没有笔记
		cacheCount = -1
	}
	d.cache.Do(c, func(ctx context.Context) {
		if e := d.addCacheArtCntInArc(ctx, oid, oidType, cacheCount); e != nil {
			log.Warn("artWarn ArtCountInArc err(%+v)", e)
		}
	})
	return count, nil
}

func (d *Dao) ArtListInUser(c context.Context, min, max, mid int64) ([]string, error) {
	addCache := true
	res, err := d.cacheArtList(c, d.userListKey(mid), min, max)
	if err != nil && err != redis.ErrNil {
		log.Warn("artWarn err(%+v)", err)
		addCache = false
	}
	if len(res) > 0 {
		cache.MetricHits.Inc("bts:ArtListInUser")
		return res, nil
	}
	cache.MetricMisses.Inc("bts:ArtListInUser")
	var artList []*article.ArtList
	artList, res, err = d.rawArtListInUser(c, min, max, mid)
	if err != nil && err != sql.ErrNoRows {
		return nil, errors.Wrapf(err, "ArtListInUser mid(%d)", mid)
	}
	if !addCache {
		return res, nil
	}
	d.cache.Do(c, func(ctx context.Context) {
		if e := d.addCacheArtList(ctx, d.userListKey(mid), artList); e != nil {
			log.Warn("artWarn ArtListInUser err(%+v)", e)
		}
	})
	return res, nil
}

func (d *Dao) ArtListInArc(c context.Context, min, max, oid, oidType int64) ([]string, error) {
	addCache := true
	res, err := d.cacheArtList(c, d.arcListKey(oid, oidType), min, max)
	if err != nil && err != redis.ErrNil {
		log.Warn("artWarn err(%+v)", err)
		addCache = false
	}
	if len(res) > 0 {
		cache.MetricHits.Inc("bts:ArtListInArc")
		return res, nil
	}
	cache.MetricMisses.Inc("bts:ArtListInArc")
	var artList []*article.ArtList
	artList, res, err = d.rawArtListInArc(c, min, max, oid, oidType)
	if err != nil && err != sql.ErrNoRows {
		return nil, errors.Wrapf(err, "ArtListInArc oid(%d) oidType(%d)", oid, oidType)
	}
	if !addCache {
		return res, nil
	}
	d.cache.Do(c, func(ctx context.Context) {
		if e := d.addCacheArtList(ctx, d.arcListKey(oid, oidType), artList); e != nil {
			log.Warn("artWarn ArtListInArc err(%+v)", e)
		}
	})
	return res, nil
}

func (d *Dao) ArtDetail(c context.Context, key int64, tp string) (*article.ArtDtlCache, error) {
	addCache := true
	res, err := d.cacheArtDetail(c, key, tp)
	if err != nil && err != redis.ErrNil {
		addCache = false
	}
	if res != nil {
		cache.MetricHits.Inc("bts:ArtDetail")
		return res, nil
	}
	cache.MetricMisses.Inc("bts:ArtDetail")
	var pubStatusEq int
	if tp == article.TpArtDetailCvid {
		pubStatusEq = article.PubStatusPassed
	}
	res, err = d.RawArtDetail(c, key, tp, pubStatusEq, 0)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	miss := res
	if miss == nil {
		miss = &article.ArtDtlCache{Cvid: -1}
	}
	if !addCache {
		return miss, nil
	}
	d.cache.Do(c, func(ctx context.Context) {
		if e := d.addCacheArtDetail(ctx, key, tp, miss); e != nil {
			log.Warn("artWarn ArtDetail err(%+v)", e)
		}
	})
	return miss, nil
}

func (d *Dao) ArtContent(c context.Context, key, ver int64) (*article.ArtContCache, error) {
	addCache := true
	res, err := d.cacheArtContent(c, key)
	if err != nil && err != redis.ErrNil {
		addCache = false
	}
	if res != nil && res.PubVersion == ver {
		cache.MetricHits.Inc("bts:ArtContent")
		return res, nil
	}
	cache.MetricMisses.Inc("bts:ArtContent")
	res, err = d.rawArtContent(c, key, ver)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	miss := res
	if miss == nil {
		miss = &article.ArtContCache{Cvid: -1}
	}
	if !addCache {
		return miss, nil
	}
	d.cache.Do(c, func(c context.Context) {
		if e := d.addCacheArtContent(c, key, miss); e != nil {
			log.Warn("artWarn ArtContent err(%+v)", e)
		}
	})
	return miss, nil
}
