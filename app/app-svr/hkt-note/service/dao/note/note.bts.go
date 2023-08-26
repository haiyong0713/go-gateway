package note

import (
	"context"

	"go-common/library/cache"
	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	"go-common/library/log"
	notegrpc "go-gateway/app/app-svr/hkt-note/service/api"
	"go-gateway/app/app-svr/hkt-note/service/model/note"

	"github.com/pkg/errors"
)

const (
	// -99在NoteAid缓存中表示该oid+mid下没有笔记id
	NoteAidCacheZeroValue = -99
)

// NoteDetail get data from cache if miss will call source method, then add to cache.
func (d *Dao) NoteAid(c context.Context, req *notegrpc.NoteListInArcReq) ([]int64, error) {
	addCache := true
	res, err := d.cacheNoteAid(c, req)
	if err != nil {
		if err != redis.ErrNil {
			addCache = false
		}
	}
	if len(res) != 0 && res[0] != -1 {
		cache.MetricHits.Inc("bts:NoteAid")
		if res[0] == NoteAidCacheZeroValue {
			res = make([]int64, 0)
			return res, nil
		}
		return res, nil
	}
	cache.MetricMisses.Inc("bts:NoteAid")
	res, err = d.rawNoteAid(c, req)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	cacheValue := res
	if !addCache {
		return res, nil
	}
	if len(cacheValue) == 0 {
		//本身oid下就没有笔记id,set一个特殊标识值
		/*	d.cache.Do(c, func(ctx context.Context) {
				if err := d.delKey(ctx, d.aidKey(req)); err != nil {
					log.Warn("noteWarn NoteAid err(%+v)", err)
				}
			})
			return miss, nil*/
		cacheValue = make([]int64, 1)
		cacheValue[0] = NoteAidCacheZeroValue
	}
	d.cache.Do(c, func(ctx context.Context) {
		if err := d.addCacheNoteAid(ctx, req, cacheValue); err != nil {
			log.Warn("noteWarn NoteAid err(%+v)", err)
		}
	})
	return res, nil
}

// NoteDetail get data from cache if miss will call source method, then add to cache.
func (d *Dao) NoteDetail(c context.Context, key int64, mid int64) (*note.DtlCache, error) {
	addCache := true
	res, err := d.cacheNoteDetail(c, key)
	if err != nil {
		if err != redis.ErrNil {
			addCache = false
		}
	}
	if res != nil {
		cache.MetricHits.Inc("bts:NoteDetail")
		return res, nil
	}
	cache.MetricMisses.Inc("bts:NoteDetail")
	res, err = d.rawNoteDetail(c, key, mid)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	miss := res
	if miss == nil {
		miss = &note.DtlCache{NoteId: -1}
	}
	if !addCache {
		return miss, nil
	}
	d.cache.Do(c, func(ctx context.Context) {
		if err := d.AddCacheNoteDetail(ctx, key, miss); err != nil {
			log.Warn("noteWarn NoteDetail err(%+v)", err)
		}
	})
	return miss, nil
}

// NoteContent get data from cache if miss will call source method, then add to cache.
func (d *Dao) NoteContent(c context.Context, key int64) (*note.ContCache, error) {
	addCache := true
	res, err := d.cacheNoteContent(c, key)
	if err != nil {
		if err != redis.ErrNil {
			addCache = false
		}
	}
	if res != nil {
		cache.MetricHits.Inc("bts:NoteContent")
		return res, nil
	}
	cache.MetricMisses.Inc("bts:NoteContent")
	res, err = d.rawNoteContent(c, key)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	miss := res
	if miss == nil {
		miss = &note.ContCache{NoteId: -1}
	}
	if !addCache {
		return miss, nil
	}
	d.cache.Do(c, func(c context.Context) {
		if err := d.AddCacheNoteContent(c, key, miss); err != nil {
			log.Warn("noteWarn NoteContent err(%+v)", err)
		}
	})
	return miss, nil
}

// NoteUser get data from cache if miss will call source method, then add to cache.
func (d *Dao) NoteUser(c context.Context, key int64) (*note.UserCache, error) {
	addCache := true
	res, err := d.cacheNoteUser(c, key)
	if err != nil {
		if err != redis.ErrNil {
			addCache = false
		}
	}
	if res != nil {
		cache.MetricHits.Inc("bts:NoteUser")
		return res, nil
	}
	cache.MetricMisses.Inc("bts:NoteUser")
	res, err = d.rawNoteUser(c, key)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	miss := res
	if miss == nil {
		miss = &note.UserCache{Mid: -1}
	}
	if !addCache {
		return miss, nil
	}
	d.cache.Do(c, func(c context.Context) {
		if err := d.addCacheNoteUser(c, key, miss); err != nil {
			log.Warn("noteWarn NoteUser err(%+v)", err)
		}
	})
	return miss, nil
}

// NoteDetails get data from cache if miss will call source method, then add to cache.
func (d *Dao) NoteDetails(c context.Context, keys []int64, mid int64) (map[int64]*note.DtlCache, error) {
	if len(keys) == 0 {
		return nil, nil
	}
	addCache := true
	var miss []int64
	res, miss, err := d.cacheNoteDetails(c, keys, mid)
	if err != nil {
		addCache = false
		res = nil
	}
	for _, key := range keys {
		if (res == nil) || (res[key] == nil) {
			miss = append(miss, key)
		}
	}
	for k, v := range res {
		if v != nil && v.NoteId == -1 {
			delete(res, k)
		}
	}
	missLen := len(miss)
	if missLen == 0 {
		return res, nil
	}
	var missData map[int64]*note.DtlCache
	missData, err = d.rawNoteDetails(c, miss, mid)
	if err != nil {
		log.Error("noteWarn NoteDetails miss(%v) err(%+v)", miss, err)
		if res == nil {
			return nil, err
		}
	}
	if res == nil {
		res = make(map[int64]*note.DtlCache, len(keys))
	}
	for k, v := range missData {
		res[k] = v
	}
	if missData == nil {
		missData = map[int64]*note.DtlCache{}
	}
	for _, key := range miss {
		if res[key] == nil {
			missData[key] = &note.DtlCache{NoteId: -1}
		}
	}
	if !addCache {
		return res, nil
	}
	for _, m := range missData {
		curM := m
		d.cache.Do(c, func(ctx context.Context) {
			if err := d.AddCacheNoteDetail(ctx, curM.NoteId, curM); err != nil {
				log.Warn("noteWarn NoteDetails err(%+v)", err)
			}
		})
	}
	return res, nil
}

// NoteDetail get data from cache if miss will call source method, then add to cache.
func (d *Dao) NoteList(c context.Context, mid, min, max, total int64) ([]string, error) {
	addCache := true
	res, err := d.cacheNoteList(c, mid, min, max, total)
	if err != nil && err != redis.ErrNil {
		log.Warn("noteWarn err(%+v)", err)
		addCache = false
	}
	if len(res) > 0 {
		cache.MetricHits.Inc("bts:NoteList")
		return res, nil
	}
	cache.MetricMisses.Inc("bts:NoteList")
	var noteList []*note.NtList
	noteList, res, err = d.rawNoteList(c, mid, min, max)
	if err != nil && err != sql.ErrNoRows {
		return nil, errors.Wrapf(err, "NoteList mid(%d)", mid)
	}
	if !addCache {
		return res, nil
	}
	d.cache.Do(c, func(ctx context.Context) {
		if err := d.AddCacheAllNoteList(ctx, mid, noteList); err != nil {
			log.Warn("noteWarn NoteDetail err(%+v)", err)
		}
	})
	return res, nil
}
