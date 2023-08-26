package article

import (
	"context"
	"encoding/json"
	"fmt"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/app-svr/hkt-note/service/model/article"

	"github.com/pkg/errors"
)

const (
	_keyArtDetailByNoteId = "_article_detail_byNoteId" // key:noteId维度
	_keyArtDetailByCvid   = "_article_detail"          // key:cvid维度
	_keyArtContent        = "_article_content"
	_keyArtListInUser     = "_user_article_list"
	_keyArtListInArc      = "_arc_article_list"
	_keyArtCountInArc     = "_art_count"   // 稿件下客态笔记数
	_keyAutoPullAid       = "auto_pull_%d" //aid下是否有直接拉起笔记
)

func (d *Dao) cacheArtDetails(c context.Context, ids []int64, tp string) (cached map[int64]*article.ArtDtlCache, missed []int64, err error) {
	conn := d.redis.Get(c)
	defer conn.Close()
	var (
		args    = redis.Args{}
		keysMap = make(map[int64]struct{})
	)
	for _, id := range ids {
		if _, ok := keysMap[id]; ok {
			continue
		}
		args = args.Add(artDetailKey(id, tp))
		keysMap[id] = struct{}{}
	}
	var items [][]byte
	if items, err = redis.ByteSlices(conn.Do("MGET", args...)); err != nil {
		err = errors.Wrapf(err, "cacheArtDetails args(%+v)", args)
		return
	}
	cached = make(map[int64]*article.ArtDtlCache)
	for _, bs := range items {
		if bs == nil {
			continue
		}
		con := &article.ArtDtlCache{}
		if e := json.Unmarshal(bs, con); e != nil {
			log.Warn("noteWarn cacheArtDetails Unmarshal bs(%s) error(%v)", bs, e)
			continue
		}
		switch tp {
		case article.TpArtDetailNoteId:
			cached[con.NoteId] = con
			delete(keysMap, con.NoteId)
		case article.TpArtDetailCvid:
			cached[con.Cvid] = con
			delete(keysMap, con.Cvid)
		default:
		}
	}
	for id := range keysMap {
		missed = append(missed, id)
	}
	return
}

func (d *Dao) cacheArtListCount(c context.Context, key string) (int, error) {
	conn := d.redis.Get(c)
	defer conn.Close()
	count, err := redis.Int(conn.Do("ZCARD", key))
	if err != nil {
		return 0, errors.Wrapf(err, "ArtListCount key(%s)", key)
	}
	return count, nil
}

func (d *Dao) addCacheArtList(c context.Context, key string, val []*article.ArtList) error {
	conn := d.redis.Get(c)
	defer conn.Close()
	args := redis.Args{}.Add(key)
	for _, v := range val {
		score := v.Pubtime
		if score <= 0 { // 先发后审且人工未过审前，用修改时间替代
			score = v.Mtime
		}
		args = args.Add(score).Add(article.ToArtListVal(v.Cvid, v.NoteId))
	}
	err := conn.Send("ZADD", args...)
	if err != nil {
		return errors.Wrapf(err, "AddCacheArtList key(%s)", key)
	}
	if err = conn.Flush(); err != nil {
		return errors.Wrapf(err, "AddCacheArtList key(%s)", key)

	}
	for i := 0; i < len(val); i++ {
		if _, err = conn.Receive(); err != nil {
			return errors.Wrapf(err, "AddCacheArtList key(%s)", key)
		}
	}
	return nil
}

func (d *Dao) cacheArtList(c context.Context, key string, min, max int64) ([]string, error) {
	conn := d.redis.Get(c)
	defer conn.Close()
	res, err := redis.Strings(conn.Do("ZREVRANGE", key, min, max))
	if err != nil {
		return nil, errors.Wrapf(err, "cacheArtList key(%s) min(%d) max(%d)", key, min, max)
	}
	return res, nil
}

func (d *Dao) addCacheArtDetail(c context.Context, id int64, tp string, val *article.ArtDtlCache) error {
	conn := d.redis.Get(c)
	defer conn.Close()
	bs, err := json.Marshal(val)
	if err != nil {
		return errors.Wrapf(err, "AddCacheArtDetail id(%d) tp(%s) val(%+v)", id, tp, val)
	}
	if _, err = conn.Do("SETEX", artDetailKey(id, tp), d.artExpire, bs); err != nil {
		return errors.Wrapf(err, "AddCacheArtDetail id(%d) tp(%s) val(%+v)", id, tp, val)
	}
	return nil
}

func (d *Dao) cacheArtDetail(c context.Context, id int64, tp string) (*article.ArtDtlCache, error) {
	conn := d.redis.Get(c)
	defer conn.Close()
	item, err := redis.Bytes(conn.Do("GET", artDetailKey(id, tp)))
	if err != nil {
		return nil, err
	}
	cache := &article.ArtDtlCache{}
	if err = json.Unmarshal(item, &cache); err != nil {
		err = errors.Wrapf(err, "cacheArtDetail key(%s) item(%s)", artDetailKey(id, tp), item)
		return nil, err
	}
	return cache, nil
}

func (d *Dao) addCacheArtContent(c context.Context, cvid int64, val *article.ArtContCache) error {
	conn := d.redis.Get(c)
	defer conn.Close()
	bs, err := json.Marshal(val)
	if err != nil {
		return errors.Wrapf(err, "AddCacheArtContent cvid(%d) val(%+v)", cvid, val)
	}
	if _, err = conn.Do("SETEX", artContentKey(cvid), d.artExpire, bs); err != nil {
		return errors.Wrapf(err, "AddCacheArtContent cvid(%d) val(%+v)", cvid, val)
	}
	return nil
}

func (d *Dao) cacheArtContent(c context.Context, cvid int64) (*article.ArtContCache, error) {
	conn := d.redis.Get(c)
	defer conn.Close()
	item, err := redis.Bytes(conn.Do("GET", artContentKey(cvid)))
	if err != nil {
		return nil, err
	}
	cache := &article.ArtContCache{}
	if err = json.Unmarshal(item, &cache); err != nil {
		err = errors.Wrapf(err, "cacheArtContent key(%s) item(%s)", artContentKey(cvid), item)
		return nil, err
	}
	return cache, nil
}

func (d *Dao) addCacheArtCntInArc(c context.Context, oid, oidType, total int64) error {
	conn := d.redis.Get(c)
	defer conn.Close()
	if _, err := conn.Do("SETEX", artCntInArcKey(oid, oidType), d.artExpire, total); err != nil {
		return errors.Wrapf(err, "addCacheArtCntInArc oid(%d) oidType(%d) total(%d)", oid, oidType, total)
	}
	return nil
}

func (d *Dao) cacheArtCntInArc(c context.Context, oid, oidType int64) (int64, error) {
	conn := d.redis.Get(c)
	defer conn.Close()
	item, err := redis.Int64(conn.Do("GET", artCntInArcKey(oid, oidType)))
	if err != nil {
		// 可能是err可能是ErrNil，不存在的时候也返回0，回源DB
		return 0, err
	}
	return item, nil
}

func artDetailKey(id int64, tp string) string {
	switch tp {
	case article.TpArtDetailNoteId:
		return fmt.Sprintf("%d%s", id, _keyArtDetailByNoteId)
	case article.TpArtDetailCvid:
		return fmt.Sprintf("%d%s", id, _keyArtDetailByCvid)
	}
	return ""
}

func artContentKey(cvid int64) string {
	return fmt.Sprintf("%d%s", cvid, _keyArtContent)
}

func (d *Dao) arcListKey(oid, oidType int64) string {
	return fmt.Sprintf("%d_%d%s", oid, oidType, _keyArtListInArc)
}

func (d *Dao) userListKey(mid int64) string {
	return fmt.Sprintf("%d%s", mid, _keyArtListInUser)
}

func artCntInArcKey(oid, oidType int64) string {
	return fmt.Sprintf("%d_%d%s", oid, oidType, _keyArtCountInArc)
}

func (d *Dao) GetAutoPullCvid(ctx context.Context, aid int64) (cvid int64, err error) {
	key := fmt.Sprintf(_keyAutoPullAid, aid)
	conn := d.redis.Get(ctx)
	defer conn.Close()
	cvid, err = redis.Int64(conn.Do("GET", key))
	if err != nil && err != redis.ErrNil {
		log.Errorc(ctx, "GetAutoPullCvid err %v ,key %v", err, key)
		return 0, errors.Wrapf(err, "GetAutoPullCvid key(%s)", key)
	}
	return cvid, nil
}

func (d *Dao) SetAutoPullCvid(ctx context.Context, aid, cvid int64) (err error) {
	key := fmt.Sprintf(_keyAutoPullAid, aid)
	conn := d.redis.Get(ctx)
	defer conn.Close()
	_, err = conn.Do("SET", key, cvid)
	if err != nil {
		log.Errorc(ctx, "SetAutoPullCvid err %v ,key %v", err, key)
		return errors.Wrapf(err, "SetAutoPullCvid key(%s)", key)
	}
	return nil
}
