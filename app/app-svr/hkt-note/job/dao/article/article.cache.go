package article

import (
	"context"
	"encoding/json"
	"fmt"

	"go-common/library/ecode"
	"go-gateway/app/app-svr/hkt-note/job/model/article"

	"github.com/pkg/errors"
)

const (
	_keyArtDetailByNoteId = "_article_detail_byNoteId" // key:noteId维度
	_keyArtDetailByCvid   = "_article_detail"          // key:cvid维度
	_keyArtContent        = "_article_content"
	_keyArtListInUser     = "_user_article_list"
	_keyArtListInArc      = "_arc_article_list"
	_keyArtCountInArc     = "_art_count" // 稿件下客态笔记数
)

func (d *Dao) AddCacheArtCntInArc(c context.Context, oid int64, oidType, total int) error {
	conn := d.redis.Get(c)
	defer conn.Close()
	key := artCntInArcKey(oid, oidType)
	if _, err := conn.Do("SETEX", key, d.ArtExpire, total); err != nil {
		return errors.Wrapf(err, "AddCacheArtCntInArc oid(%d) oidType(%d) val(%d)", oid, oidType, total)
	}
	return nil
}

func (d *Dao) AddCacheArtDetail(c context.Context, id int64, val *article.ArtDtlCache, tp string, expire int) error {
	conn := d.redis.Get(c)
	defer conn.Close()
	bs, err := json.Marshal(val)
	if err != nil {
		return errors.Wrapf(err, "AddCacheArtDetail val(%+v) tp(%s)", val, tp)
	}
	key := artDetailKey(id, tp)
	if key == "" {
		return errors.Wrapf(ecode.NothingFound, "AddCacheArtDetail val(%+v) tp(%s)", val, tp)
	}
	if _, err = conn.Do("SETEX", key, expire, bs); err != nil {
		return errors.Wrapf(err, "AddCacheArtDetail val(%+v) tp(%s)", val, tp)
	}
	return nil
}

func (d *Dao) AddCacheArtContent(c context.Context, cvid int64, val *article.ArtContCache) error {
	conn := d.redis.Get(c)
	defer conn.Close()
	bs, err := json.Marshal(val)
	if err != nil {
		return errors.Wrapf(err, "AddCacheArtContent val(%+v)", val)
	}
	if _, err = conn.Do("SETEX", artContentKey(cvid), d.ArtExpire, bs); err != nil {
		return errors.Wrapf(err, "AddCacheArtContent val(%+v)", val)
	}
	return nil
}

func (d *Dao) AddCacheArtList(c context.Context, key, val string, score int64) error {
	conn := d.redis.Get(c)
	defer conn.Close()
	if _, err := conn.Do("ZADD", key, score, val); err != nil {
		return errors.Wrapf(err, "AddCacheArtList key(%s) val(%s) score(%d)", key, val, score)
	}
	return nil
}

func (d *Dao) RemCacheArtList(c context.Context, key, val string) error {
	conn := d.redis.Get(c)
	defer conn.Close()
	if _, err := conn.Do("ZREM", key, val); err != nil {
		return errors.Wrapf(err, "RemCacheArtList key(%s) val(%s)", key, val)
	}
	return nil
}

func artDetailKey(id int64, tp string) string {
	switch tp {
	case article.TpArtDetailNoteId:
		return fmt.Sprintf("%d%s", id, _keyArtDetailByNoteId)
	case article.TpArtDetailCvid:
		return fmt.Sprintf("%d%s", id, _keyArtDetailByCvid)
	default:
		return ""
	}
}

func artContentKey(cvid int64) string {
	return fmt.Sprintf("%d%s", cvid, _keyArtContent)
}

func (d *Dao) ArcListKey(oid int64, oidType int) string {
	return fmt.Sprintf("%d_%d%s", oid, oidType, _keyArtListInArc)
}

func (d *Dao) UserListKey(mid int64) string {
	return fmt.Sprintf("%d%s", mid, _keyArtListInUser)
}

func artCntInArcKey(oid int64, oidType int) string {
	return fmt.Sprintf("%d_%d%s", oid, oidType, _keyArtCountInArc)
}
