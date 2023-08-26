package dao

import (
	"context"
	"encoding/json"
	"fmt"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-common/library/xstr"
	arcmdl "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/web/interface/model"

	"github.com/pkg/errors"
)

const (
	_keyRkFmt             = "r_v2_%d_%d_%d_%d"
	_keyRkRegionFmt       = "cache_rank_region_%d_%d_%d"
	_keyRkTagFmt          = "cache_rank_tag_%d_%d"
	_webTopCacheKey       = "cache_new_web_top"
	_keyBakPrefix         = "b_"
	_webTopBakCacheKey    = _keyBakPrefix + _webTopCacheKey
	_webTopHotBakCacheKey = "cache_web_top_hot"
)

func keyRkList(rid int16, rankType, day, arcType int) string {
	return fmt.Sprintf(_keyRkFmt, rid, rankType, day, arcType)
}

func keyRkListBak(rid int16, rankType, day, arcType int) string {
	return _keyBakPrefix + keyRkList(rid, rankType, day, arcType)
}

func keyRkIndex(day int) string {
	return fmt.Sprintf("cache_rank_index_%d", day)
}

func keyRkIndexBak(day int) string {
	return _keyBakPrefix + keyRkIndex(day)
}

func keyRkRegionList(rid int64, day, original int) string {
	return fmt.Sprintf(_keyRkRegionFmt, rid, day, original)
}

func keyRkRegionListBak(rid int64, day, original int) string {
	return _keyBakPrefix + keyRkRegionList(rid, day, original)
}

func keyRkRecommendList(rid int64) string {
	return fmt.Sprintf("cache_rank_rcmd_%d", rid)
}

func keyLpRkRecommendList(business string) string {
	return fmt.Sprintf("cache_lp_rank_rcmd_%s", business)
}

func keyRkRecommendListBak(rid int64) string {
	return _keyBakPrefix + keyRkRecommendList(rid)
}

func keyLpRkRecommendListBak(business string) string {
	return _keyBakPrefix + keyLpRkRecommendList(business)
}

func keyRkTagList(rid int16, tagID int64) string {
	return fmt.Sprintf(_keyRkTagFmt, rid, tagID)
}

func keyRkTagListBak(rid int16, tagID int64) string {
	return _keyBakPrefix + keyRkTagList(rid, tagID)
}

// RankingCache get rank list from cache.
func (d *Dao) RankingCache(c context.Context, rid int16, rankType, day, arcType int) (data *model.RankData, err error) {
	key := keyRkList(rid, rankType, day, arcType)
	conn := d.redis.Get(c)
	defer conn.Close()
	data, err = d.rankingCache(conn, key)
	return
}

// RankingBakCache get rank list from bak cache.
func (d *Dao) RankingBakCache(c context.Context, rid int16, rankType, day, arcType int) (data *model.RankData, err error) {
	d.cacheProm.Incr("ranking_remote_cache")
	key := keyRkListBak(rid, rankType, day, arcType)
	conn := d.redisBak.Get(c)
	defer conn.Close()
	data, err = d.rankingCache(conn, key)
	if data == nil || len(data.List) == 0 {
		log.Error("RankingBakCache(%s) is nil", key)
	}
	return
}

func keyRkV2(typ, rid int64) string {
	return fmt.Sprintf("cache_rank_list_%d_%d", typ, rid)
}

func (d *Dao) RankingV2Cache(ctx context.Context, typ, rid int64) (*model.RankV2Cache, error) {
	key := keyRkV2(typ, rid)
	conn := d.redisBak.Get(ctx)
	defer conn.Close()
	reply, err := redis.Bytes(conn.Do("GET", key))
	if err != nil {
		return nil, err
	}
	res := new(model.RankV2Cache)
	if err = json.Unmarshal(reply, &res); err != nil {
		return nil, err
	}
	return res, nil
}

func keyRkV2Bak(typ, rid int64) string {
	return _keyBakPrefix + keyRkV2(typ, rid)
}

func (d *Dao) RankingV2BakCache(ctx context.Context, typ, rid int64) (*model.RankV2, error) {
	key := keyRkV2Bak(typ, rid)
	conn := d.redisBak.Get(ctx)
	defer conn.Close()
	reply, err := redis.Bytes(conn.Do("GET", key))
	if err != nil {
		return nil, err
	}
	res := new(model.RankV2)
	if err = json.Unmarshal(reply, &res); err != nil {
		return nil, err
	}
	return res, nil
}

func (d *Dao) SetRankingV2BakCache(ctx context.Context, typ, rid int64, data *model.RankV2) error {
	key := keyRkV2Bak(typ, rid)
	conn := d.redisBak.Get(ctx)
	defer conn.Close()
	bs, err := json.Marshal(data)
	if err != nil {
		return err
	}
	if _, err = conn.Do("SET", key, bs); err != nil {
		return err
	}
	return nil
}

// RankingIndexCache get rank index from cache.
func (d *Dao) RankingIndexCache(ctx context.Context, day int) ([]int64, error) {
	key := keyRkIndex(day)
	aids, err := d.rankingAidsCache(ctx, key)
	if err != nil {
		return nil, err
	}
	return aids, nil
}

func (d *Dao) rankingAidsCache(ctx context.Context, key string) ([]int64, error) {
	conn := d.redisBak.Get(ctx)
	defer conn.Close()
	reply, err := redis.String(conn.Do("GET", key))
	if err != nil {
		return nil, err
	}
	return xstr.SplitInts(reply)
}

func (d *Dao) rankingNewArcsCache(ctx context.Context, key string) ([]*model.NewArchive, error) {
	conn := d.redisBak.Get(ctx)
	defer conn.Close()
	reply, err := redis.Bytes(conn.Do("GET", key))
	if err != nil {
		return nil, err
	}
	var res []*model.NewArchive
	if err := json.Unmarshal(reply, &res); err != nil {
		return nil, err
	}
	return res, nil
}

// RankingIndexBakCache get rank index from bak cache.
func (d *Dao) RankingIndexBakCache(c context.Context, day int) (arcs []*model.IndexArchive, err error) {
	d.cacheProm.Incr("ranking_index_remote_cache")
	key := keyRkIndexBak(day)
	conn := d.redisBak.Get(c)
	defer conn.Close()
	arcs, err = d.rankingIndexCache(conn, key)
	if len(arcs) == 0 {
		log.Error("RankingIndexBakCache(%s) is nil", key)
	}
	return
}

// RankingRegionCache get rank cate list from cache.
func (d *Dao) RankingRegionCache(ctx context.Context, rid int64, day, original int) (arcs []*model.NewArchive, err error) {
	arcs, err = d.rankingNewArcsCache(ctx, keyRkRegionList(rid, day, original))
	return
}

// RankingRegionBakCache get rank cate list from bak cache.
func (d *Dao) RankingRegionBakCache(c context.Context, rid int64, day, original int) (arcs []*model.RegionArchive, err error) {
	d.cacheProm.Incr("ranking_region_remote_cache")
	key := keyRkRegionListBak(rid, day, original)
	conn := d.redisBak.Get(c)
	defer conn.Close()
	arcs, err = d.rankingRegionCache(conn, key)
	if len(arcs) == 0 {
		log.Error("RankingRegionBakCache(%s) is nil", key)
	}
	return
}

// RankingRecommendCache get rank recommend list from cache.
func (d *Dao) RankingRecommendCache(ctx context.Context, rid int64) ([]int64, error) {
	key := keyRkRecommendList(rid)
	aids, err := d.rankingAidsCache(ctx, key)
	if err != nil {
		return nil, err
	}
	return aids, nil
}

func (d *Dao) LpRankingRecommendCache(ctx context.Context, business string) ([]int64, error) {
	key := keyLpRkRecommendList(business)
	aids, err := d.rankingAidsCache(ctx, key)
	if err != nil {
		return nil, err
	}
	return aids, nil
}

// RankingRecommendBakCache get rank recommend list from bak cache.
func (d *Dao) RankingRecommendBakCache(c context.Context, rid int64) (arcs []*model.IndexArchive, err error) {
	d.cacheProm.Incr("ranking_rec_remote_cache")
	key := keyRkRecommendListBak(rid)
	conn := d.redisBak.Get(c)
	defer conn.Close()
	arcs, err = d.rankingIndexCache(conn, key)
	if len(arcs) == 0 {
		log.Error("RankingRecommendBakCache(%s) is nil", key)
	}
	return
}

func (d *Dao) LpRankingRecommendBakCache(c context.Context, business string) (arcs []*model.IndexArchive, err error) {
	d.cacheProm.Incr("lp_ranking_rec_remote_cache")
	key := keyLpRkRecommendListBak(business)
	conn := d.redisBak.Get(c)
	defer conn.Close()
	arcs, err = d.rankingIndexCache(conn, key)
	if len(arcs) == 0 {
		log.Error("LpRankingRecommendBakCache(%s) is nil", key)
	}
	return
}

// RankingTagCache get ranking tag from cache.
func (d *Dao) RankingTagCache(ctx context.Context, rid int16, tagID int64) ([]*model.NewArchive, error) {
	return d.rankingNewArcsCache(ctx, keyRkTagList(rid, tagID))
}

// RankingTagBakCache get ranking tag from bak cache.
func (d *Dao) RankingTagBakCache(c context.Context, rid int16, tagID int64) (arcs []*model.TagArchive, err error) {
	d.cacheProm.Incr("ranking_tag_remote_cache")
	key := keyRkTagListBak(rid, tagID)
	conn := d.redisBak.Get(c)
	defer conn.Close()
	arcs, err = d.rankingTagCache(conn, key)
	if len(arcs) == 0 {
		log.Error("RankingTagBakCache(%s) is nil", key)
	}
	return
}

func (d *Dao) rankingCache(conn redis.Conn, key string) (arcs *model.RankData, err error) {
	var value []byte
	if value, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Error("conn.Do(GET, %s) error(%v)", key, err)
		}
		return
	}
	arcs = new(model.RankData)
	if err = json.Unmarshal(value, &arcs); err != nil {
		log.Error("json.Unmarshal(%v) error(%v)", value, err)
	}
	return
}

func (d *Dao) rankingIndexCache(conn redis.Conn, key string) (arcs []*model.IndexArchive, err error) {
	var value []byte
	if value, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Error("conn.Do(GET, %s) error(%v)", key, err)
		}
		return
	}
	arcs = []*model.IndexArchive{}
	if err = json.Unmarshal(value, &arcs); err != nil {
		log.Error("json.Unmarshal(%v) error(%v)", value, err)
	}
	return
}

func (d *Dao) rankingRegionCache(conn redis.Conn, key string) (arcs []*model.RegionArchive, err error) {
	var value []byte
	if value, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Error("conn.Do(GET, %s) error(%v)", key, err)
		}
		return
	}
	arcs = []*model.RegionArchive{}
	if err = json.Unmarshal(value, &arcs); err != nil {
		log.Error("json.Unmarshal(%v) error(%v)", value, err)
	}
	return
}

func (d *Dao) rankingTagCache(conn redis.Conn, key string) (arcs []*model.TagArchive, err error) {
	var value []byte
	if value, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Error("conn.Do(GET, %s) error(%v)", key, err)
		}
		return
	}
	arcs = []*model.TagArchive{}
	if err = json.Unmarshal(value, &arcs); err != nil {
		log.Error("json.Unmarshal(%v) error(%v)", value, err)
	}
	return
}

// SetRankingCache set ranking data to cache
func (d *Dao) SetRankingCache(c context.Context, rid int16, rankType, day, arcType int, data *model.RankData) (err error) {
	key := keyRkList(rid, rankType, day, arcType)
	conn := d.redis.Get(c)
	defer conn.Close()
	if err = d.setRkCache(conn, key, d.redisRkExpire, data); err != nil {
		return
	}
	key = keyRkListBak(rid, rankType, day, arcType)
	connBak := d.redisBak.Get(c)
	err = d.setRkCache(connBak, key, d.redisRkBakExpire, data)
	connBak.Close()
	return
}

// SetRankingIndexCache set ranking index data to cache
func (d *Dao) SetRankingIndexCache(c context.Context, day int, arcs []*model.IndexArchive) (err error) {
	key := keyRkIndexBak(day)
	connBak := d.redisBak.Get(c)
	err = d.setRkIndexCache(connBak, key, d.redisRkBakExpire, arcs)
	connBak.Close()
	return
}

// SetRankingRegionCache set ranking data to cache
func (d *Dao) SetRankingRegionCache(c context.Context, rid int64, day, original int, arcs []*model.RegionArchive) (err error) {
	key := keyRkRegionListBak(rid, day, original)
	connBak := d.redisBak.Get(c)
	err = d.setRkRegionCache(connBak, key, d.redisRkBakExpire, arcs)
	connBak.Close()
	return
}

// SetRankingRecommendCache set ranking data to bak cache
func (d *Dao) SetRankingRecommendCache(c context.Context, rid int64, arcs []*model.IndexArchive) (err error) {
	key := keyRkRecommendListBak(rid)
	connBak := d.redisBak.Get(c)
	err = d.setRkIndexCache(connBak, key, d.redisRkBakExpire, arcs)
	connBak.Close()
	return
}

func (d *Dao) SetLpRankingRecommendCache(c context.Context, business string, arcs []*model.IndexArchive) (err error) {
	key := keyLpRkRecommendListBak(business)
	connBak := d.redisBak.Get(c)
	err = d.setRkIndexCache(connBak, key, d.redisRkBakExpire, arcs)
	connBak.Close()
	return
}

// SetRankingTagCache set ranking tag data to cache
func (d *Dao) SetRankingTagCache(c context.Context, rid int16, tagID int64, arcs []*model.TagArchive) (err error) {
	key := keyRkTagListBak(rid, tagID)
	connBak := d.redisBak.Get(c)
	err = d.setRkTagCache(connBak, key, d.redisRkBakExpire, arcs)
	connBak.Close()
	return
}

// SetWebTopCache .
func (d *Dao) SetWebTopCache(c context.Context, arcs []*model.BvArc) (err error) {
	key := _webTopBakCacheKey
	connBak := d.redisBak.Get(c)
	err = d.setWebTopCache(connBak, key, d.redisRkBakExpire, arcs)
	connBak.Close()
	return
}

func (d *Dao) setWebTopCache(conn redis.Conn, key string, expire int32, arcs []*model.BvArc) (err error) {
	var bs []byte
	if bs, err = json.Marshal(arcs); err != nil {
		log.Error("json.Marshal(%v) error (%v)", arcs, err)
		return
	}
	if err = conn.Send("SET", key, bs); err != nil {
		log.Error("conn.Send(SET, %s, %s) error(%v)", key, string(bs), err)
		return
	}
	if err = conn.Send("EXPIRE", key, expire); err != nil {
		log.Error("conn.Send(Expire, %s, %d) error(%v)", key, expire, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("conn.Receive() error(%v)", err)
			return
		}
	}
	return
}

func (d *Dao) setRkCache(conn redis.Conn, key string, expire int32, arcs *model.RankData) (err error) {
	var bs []byte
	if bs, err = json.Marshal(arcs); err != nil {
		log.Error("json.Marshal(%v) error (%v)", arcs, err)
		return
	}
	if err = conn.Send("SET", key, bs); err != nil {
		log.Error("conn.Send(SET, %s, %s) error(%v)", key, string(bs), err)
		return
	}
	if err = conn.Send("EXPIRE", key, expire); err != nil {
		log.Error("conn.Send(Expire, %s, %d) error(%v)", key, expire, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("conn.Receive() error(%v)", err)
			return
		}
	}
	return
}

func (d *Dao) setRkIndexCache(conn redis.Conn, key string, expire int32, arcs []*model.IndexArchive) (err error) {
	var bs []byte
	if bs, err = json.Marshal(arcs); err != nil {
		log.Error("json.Marshal(%v) error (%v)", arcs, err)
		return
	}
	if err = conn.Send("SET", key, bs); err != nil {
		log.Error("conn.Send(SET, %s, %s) error(%v)", key, string(bs), err)
		return
	}
	if err = conn.Send("EXPIRE", key, expire); err != nil {
		log.Error("conn.Send(Expire, %s, %d) error(%v)", key, expire, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("conn.Receive() error(%v)", err)
			return
		}
	}
	return
}

func (d *Dao) setRkRegionCache(conn redis.Conn, key string, expire int32, arcs []*model.RegionArchive) (err error) {
	var bs []byte
	if bs, err = json.Marshal(arcs); err != nil {
		log.Error("json.Marshal(%v) error (%v)", arcs, err)
		return
	}
	if err = conn.Send("SET", key, bs); err != nil {
		log.Error("conn.Send(SET, %s, %s) error(%v)", key, string(bs), err)
		return
	}
	if err = conn.Send("EXPIRE", key, expire); err != nil {
		log.Error("conn.Send(Expire, %s, %d) error(%v)", key, expire, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("conn.Receive() error(%v)", err)
			return
		}
	}
	return
}

func (d *Dao) setRkTagCache(conn redis.Conn, key string, expire int32, arcs []*model.TagArchive) (err error) {
	var bs []byte
	if bs, err = json.Marshal(arcs); err != nil {
		log.Error("json.Marshal(%v) error (%v)", arcs, err)
		return
	}
	if err = conn.Send("SET", key, bs); err != nil {
		log.Error("conn.Send(SET, %s, %s) error(%v)", key, string(bs), err)
		return
	}
	if err = conn.Send("EXPIRE", key, expire); err != nil {
		log.Error("conn.Send(Expire, %s, %d) error(%v)", key, expire, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("conn.Receive() error(%v)", err)
			return
		}
	}
	return
}

// WebTopCache get wx hot cache.
func (d *Dao) WebTopCache(c context.Context) ([]int64, error) {
	key := _webTopCacheKey
	conn := d.redisBak.Get(c)
	defer conn.Close()
	reply, err := redis.String(conn.Do("GET", key))
	if err != nil {
		return nil, err
	}
	return xstr.SplitInts(reply)
}

// WebTopBakCache get wx hot bak cache.
func (d *Dao) WebTopBakCache(c context.Context) (arcs []*model.BvArc, err error) {
	key := _webTopBakCacheKey
	conn := d.redisBak.Get(c)
	defer conn.Close()
	arcs, err = webTopCache(conn, key)
	return
}

func webTopCache(conn redis.Conn, key string) (res []*model.BvArc, err error) {
	var value []byte
	if value, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Error("conn.Do(GET, %s) error(%v)", key, err)
		}
		return
	}
	res = []*model.BvArc{}
	if err = json.Unmarshal(value, &res); err != nil {
		log.Error("json.Unmarshal(%v) error(%v)", value, err)
	}
	return
}

func (d *Dao) SetWebTopHotCache(ctx context.Context, arcs []*arcmdl.Arc) error {
	key := _webTopHotBakCacheKey
	conn := d.redisBak.Get(ctx)
	defer conn.Close()
	bs, err := json.Marshal(arcs)
	if err != nil {
		return errors.Wrapf(err, "%+v", arcs)
	}
	if _, err := conn.Do("SETEX", key, d.redisRkBakExpire, bs); err != nil {
		return errors.Wrapf(err, "SETEX %v,%v,%s", key, d.redisRkBakExpire, bs)
	}
	return nil
}

func (d *Dao) WebTopHotBakCache(ctx context.Context) ([]*arcmdl.Arc, error) {
	key := _webTopHotBakCacheKey
	conn := d.redisBak.Get(ctx)
	defer conn.Close()
	value, err := redis.Bytes(conn.Do("GET", key))
	if err != nil {
		if err == redis.ErrNil {
			return nil, nil
		}
		return nil, errors.Wrap(err, key)
	}
	var arcs []*arcmdl.Arc
	if err := json.Unmarshal(value, &arcs); err != nil {
		return nil, errors.Wrapf(err, "%s", value)
	}
	return arcs, nil
}
