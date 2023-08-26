package article

import (
	"context"
	"encoding/json"
	"fmt"

	"go-common/library/ecode"
	"go-gateway/app/app-svr/hkt-note/interface/model/article"
	notegrpc "go-gateway/app/app-svr/hkt-note/service/api"

	"github.com/pkg/errors"
)

const (
	_keyArtDetailByNoteId = "_article_detail_byNoteId" // key:noteId维度
	_keyArtDetailByCvid   = "_article_detail"          // key:cvid维度
	_keyArtListInUser     = "_user_article_list"
)

func artDetailKey(id int64, tp string) string {
	switch tp {
	case article.TpArtDetailNoteId:
		return fmt.Sprintf("%d%s", id, _keyArtDetailByNoteId)
	case article.TpArtDetailCvid:
		return fmt.Sprintf("%d%s", id, _keyArtDetailByCvid)
	}
	return ""
}

func (d *Dao) AddCacheArtDetail(c context.Context, id int64, tp string, val *article.ArtDtlCache) error {
	conn := d.redis.Get(c)
	defer conn.Close()
	key := artDetailKey(id, tp)
	if key == "" {
		return errors.Wrapf(ecode.NothingFound, "AddCacheArtDetail val(%+v) tp(%s) invalid", val, tp)
	}
	bs, err := json.Marshal(val)
	if err != nil {
		return errors.Wrapf(err, "AddCacheArtDetail val(%+v)", val)
	}
	if _, err = conn.Do("SETEX", key, d.artExpire, bs); err != nil {
		return errors.Wrapf(err, "AddCacheArtDetail val(%+v)", val)
	}
	return nil
}

func (d *Dao) RemCachesArtListUser(c context.Context, mid int64, arts map[int64]*notegrpc.SimpleArticleCard) error {
	conn := d.redis.Get(c)
	defer conn.Close()
	key := userListKey(mid)
	for _, a := range arts {
		if err := conn.Send("ZREM", key, toArtListVal(a.Cvid, a.NoteId)); err != nil {
			return errors.Wrapf(err, "RemCachesArtListUser art(%+v)", a)
		}
	}
	if err := conn.Flush(); err != nil {
		return errors.Wrapf(err, "RemCachesArtListUser arts(%+v)", arts)
	}
	return nil
}

func userListKey(mid int64) string {
	return fmt.Sprintf("%d%s", mid, _keyArtListInUser)
}

func toArtListVal(cvid, noteId int64) string {
	return fmt.Sprintf("%d-%d", cvid, noteId)
}
