package note

import (
	"context"
	"encoding/json"
	"fmt"
	"go-common/library/log"

	"go-common/library/cache/redis"
	ntmdl "go-gateway/app/app-svr/hkt-note/job/model/note"

	"github.com/pkg/errors"
)

const (
	_keyDetail  = "_note_detail"
	_keyContent = "_note_content"
	KeyUser     = "_note_user"
	_keyList    = "_note_list"
	_keyAid     = "_note"
)

func (d *Dao) AddCacheNoteAid(c context.Context, mid, aid, noteId int64, oidType int) error {
	conn := d.redis.Get(c)
	defer conn.Close()
	if _, err := conn.Do("SETEX", d.AidKey(mid, aid, oidType), d.noteExpire, noteId); err != nil {
		return errors.Wrapf(err, "AddCacheNoteAid val(%d)", noteId)
	}
	return nil
}

func (d *Dao) AddCacheNoteUser(c context.Context, mid int64, val *ntmdl.UserCache) error {
	conn := d.redis.Get(c)
	defer conn.Close()
	bs, err := json.Marshal(val)
	if err != nil {
		return errors.Wrapf(err, "AddCacheNoteUser val(%+v)", val)
	}
	if _, err = conn.Do("SETEX", d.UserKey(mid), d.noteExpire, bs); err != nil {
		return errors.Wrapf(err, "AddCacheNoteUser val(%+v)", val)
	}
	return nil
}

func (d *Dao) AddCacheNoteDetail(c context.Context, noteId int64, val *ntmdl.DtlCache) error {
	conn := d.redis.Get(c)
	defer conn.Close()
	bs, err := json.Marshal(val)
	if err != nil {
		return errors.Wrapf(err, "AddCacheNoteDetail val(%+v)", val)
	}
	if _, err = conn.Do("SETEX", d.DetailKey(noteId), d.noteExpire, bs); err != nil {
		return errors.Wrapf(err, "AddCacheNoteDetail val(%+v)", val)
	}
	return nil
}

func (d *Dao) AddCacheNoteContent(c context.Context, noteId int64, val *ntmdl.ContCache) error {
	conn := d.redis.Get(c)
	defer conn.Close()
	bs, err := json.Marshal(val)
	if err != nil {
		return errors.Wrapf(err, "AddCacheNoteContent val(%+v)", val)
	}
	if _, err = conn.Do("SETEX", d.ContentKey(noteId), d.noteExpire, bs); err != nil {
		return errors.Wrapf(err, "AddCacheNoteContent val(%+v)", val)
	}
	return nil
}

func (d *Dao) CacheNoteDetail(c context.Context, noteId int64) (*ntmdl.DtlCache, error) {
	conn := d.redis.Get(c)
	defer conn.Close()
	item, err := redis.Bytes(conn.Do("GET", d.DetailKey(noteId)))
	if err != nil {
		err = errors.Wrapf(err, "CacheNoteDetail key(%s)", d.DetailKey(noteId))
		return nil, err
	}
	cache := &ntmdl.DtlCache{}
	if err = json.Unmarshal(item, &cache); err != nil {
		err = errors.Wrapf(err, "CacheNoteDetail key(%s) item(%s)", d.DetailKey(noteId), item)
		return nil, err
	}
	return cache, nil
}

func (d *Dao) CacheNoteContent(c context.Context, noteId int64) (*ntmdl.ContCache, error) {
	conn := d.redis.Get(c)
	defer conn.Close()
	item, err := redis.Bytes(conn.Do("GET", d.ContentKey(noteId)))
	if err != nil {
		err = errors.Wrapf(err, "CacheNoteContent key(%s)", d.ContentKey(noteId))
		return nil, err
	}
	cache := &ntmdl.ContCache{}
	if err = json.Unmarshal(item, &cache); err != nil {
		err = errors.Wrapf(err, "CacheNoteContent key(%s) item(%s)", d.ContentKey(noteId), item)
		return nil, err
	}
	return cache, nil
}

func (d *Dao) AddCacheNoteList(c context.Context, mid int64, val string, score int64) error {
	conn := d.redis.Get(c)
	defer conn.Close()
	if _, err := conn.Do("ZADD", listKey(mid), score, val); err != nil {
		return errors.Wrapf(err, "AddCacheNoteList key(%s) val(%s) score(%d)", listKey(mid), val, score)
	}
	return nil
}

func (d *Dao) RemCacheNoteList(c context.Context, mid int64, val string) error {
	conn := d.redis.Get(c)
	defer conn.Close()
	if _, err := conn.Do("ZREM", listKey(mid), val); err != nil {
		return errors.Wrapf(err, "RemCacheNoteList key(%s) val(%s)", listKey(mid), val)
	}
	return nil
}

func (d *Dao) DelKey(c context.Context, key string) error {
	conn := d.redis.Get(c)
	defer conn.Close()
	if _, err := conn.Do("DEL", key); err != nil {
		return errors.Wrapf(err, "DelKey key(%s)", key)
	}
	return nil
}

func (d *Dao) DetailKey(noteId int64) string {
	return fmt.Sprintf("%d%s", noteId, _keyDetail)
}

func (d *Dao) ContentKey(noteId int64) string {
	return fmt.Sprintf("%d%s", noteId, _keyContent)
}

func (d *Dao) UserKey(mid int64) string {
	return fmt.Sprintf("%d%s", mid, KeyUser)
}

func listKey(mid int64) string {
	return fmt.Sprintf("%d%s", mid, _keyList)
}

func (d *Dao) AidKey(mid, oid int64, oidType int) string {
	if oidType == ntmdl.OidTypeCheese {
		return fmt.Sprintf("%d_%d_%d%s", oid, mid, ntmdl.OidTypeCheese, _keyAid)
	}
	return fmt.Sprintf("%d_%d%s", oid, mid, _keyAid)
}

func botPushRecordKey(aid int64) string {

	return fmt.Sprintf("bot_push_%d", aid)

}

// flag
func (d *Dao) SetBotPushRecord(ctx context.Context, aid, date int64) error {
	conn := d.redis.Get(ctx)
	defer conn.Close()
	if _, err := conn.Do("SETEX", botPushRecordKey(aid), d.c.Redis.BotPushExpire, date); err != nil {
		log.Errorc(ctx, "d.SetBotPushRecord val(%d)", aid)
		return err
	}
	return nil
}

func (d *Dao) GetBotPushRecord(ctx context.Context, aid int64) (int64, error) {
	conn := d.redis.Get(ctx)
	defer conn.Close()
	date, err := redis.Int64(conn.Do("GET", botPushRecordKey(aid)))
	if err != nil {
		if err == redis.ErrNil {
			return 0, nil
		}
		log.Errorc(ctx, "d.GetBotPushRecord val(%d)", aid)
		return 0, err
	}
	return date, nil
}
