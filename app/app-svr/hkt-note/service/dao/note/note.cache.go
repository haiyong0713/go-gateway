package note

import (
	"context"
	"encoding/json"
	"fmt"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-common/library/xstr"
	notegrpc "go-gateway/app/app-svr/hkt-note/service/api"
	"go-gateway/app/app-svr/hkt-note/service/model/note"

	"github.com/pkg/errors"
)

const (
	_keyDetail  = "_note_detail"
	_keyContent = "_note_content"
	_keyUser    = "_note_user"
	_keyList    = "_note_list"
	_keyAid     = "_note"
)

func (d *Dao) addCacheNoteAid(c context.Context, req *notegrpc.NoteListInArcReq, noteIds []int64) error {
	conn := d.redis.Get(c)
	defer conn.Close()
	idsStr := xstr.JoinInts(noteIds)
	if _, err := conn.Do("SETEX", d.aidKey(req), d.aidNoteExpire, idsStr); err != nil {
		return errors.Wrapf(err, "AddCacheNoteAid key(%s) val(%s)", d.aidKey(req), idsStr)
	}
	return nil
}

func (d *Dao) cacheNoteAid(c context.Context, req *notegrpc.NoteListInArcReq) ([]int64, error) {
	conn := d.redis.Get(c)
	defer conn.Close()
	idStr, err := redis.String(conn.Do("GET", d.aidKey(req)))
	if err != nil {
		return nil, err
	}
	ids, err := xstr.SplitInts(idStr)
	if err != nil {
		return nil, err
	}
	return ids, nil
}

func (d *Dao) AddCacheNoteDetail(c context.Context, noteId int64, val *note.DtlCache) error {
	conn := d.redis.Get(c)
	defer conn.Close()
	bs, err := json.Marshal(val)
	if err != nil {
		return errors.Wrapf(err, "AddCacheNoteDetail val(%+v)", val)
	}
	if _, err = conn.Do("SETEX", detailKey(noteId), d.noteExpire, bs); err != nil {
		return errors.Wrapf(err, "AddCacheNoteDetail val(%+v)", val)
	}
	return nil
}

func (d *Dao) AddCacheNoteContent(c context.Context, noteId int64, val *note.ContCache) error {
	conn := d.redis.Get(c)
	defer conn.Close()
	bs, err := json.Marshal(val)
	if err != nil {
		return errors.Wrapf(err, "AddCacheNoteContent val(%+v)", val)
	}
	if _, err = conn.Do("SETEX", contentKey(noteId), d.noteExpire, bs); err != nil {
		return errors.Wrapf(err, "AddCacheNoteContent val(%+v)", val)
	}
	return nil
}

func (d *Dao) cacheNoteDetail(c context.Context, noteId int64) (*note.DtlCache, error) {
	conn := d.redis.Get(c)
	defer conn.Close()
	item, err := redis.Bytes(conn.Do("GET", detailKey(noteId)))
	if err != nil {
		return nil, err
	}
	cache := &note.DtlCache{}
	if err = json.Unmarshal(item, &cache); err != nil {
		err = errors.Wrapf(err, "CacheNoteDetail key(%s) item(%s)", detailKey(noteId), item)
		return nil, err
	}
	return cache, nil
}

func (d *Dao) cacheNoteDetails(c context.Context, noteIds []int64, mid int64) (cached map[int64]*note.DtlCache, missed []int64, err error) {
	conn := d.redis.Get(c)
	defer conn.Close()
	var (
		args    = redis.Args{}
		keysMap = make(map[int64]struct{})
	)
	for _, id := range noteIds {
		if _, ok := keysMap[id]; ok {
			continue
		}
		args = args.Add(detailKey(id))
		keysMap[id] = struct{}{}
	}
	var items [][]byte
	if items, err = redis.ByteSlices(conn.Do("MGET", args...)); err != nil {
		err = errors.Wrapf(err, "cachesNoteDetail args(%+v)", args)
		return
	}
	cached = make(map[int64]*note.DtlCache)
	for _, bs := range items {
		if bs == nil {
			continue
		}
		con := &note.DtlCache{}
		if err = json.Unmarshal(bs, con); err != nil {
			log.Warn("noteWarn cachesNoteDetail Unmarshal bs(%s) error(%v)", bs, err)
			continue
		}
		// mid不匹配的不返回也不回源
		if con.Mid == mid {
			cached[con.NoteId] = con
		}
		delete(keysMap, con.NoteId)
	}
	for aid := range keysMap {
		missed = append(missed, aid)
	}
	return
}

func (d *Dao) cacheNoteContent(c context.Context, noteId int64) (*note.ContCache, error) {
	conn := d.redis.Get(c)
	defer conn.Close()
	item, err := redis.Bytes(conn.Do("GET", contentKey(noteId)))
	if err != nil {
		return nil, err
	}
	cache := &note.ContCache{}
	if err = json.Unmarshal(item, &cache); err != nil {
		err = errors.Wrapf(err, "CacheNoteContent key(%s) item(%s)", contentKey(noteId), item)
		return nil, err
	}
	return cache, nil
}

func (d *Dao) cacheNoteUser(c context.Context, mid int64) (*note.UserCache, error) {
	conn := d.redis.Get(c)
	defer conn.Close()
	item, err := redis.Bytes(conn.Do("GET", userKey(mid)))
	if err != nil {
		return nil, err
	}
	cache := &note.UserCache{}
	if err = json.Unmarshal(item, &cache); err != nil {
		err = errors.Wrapf(err, "cacheNoteUser key(%s) item(%s)", userKey(mid), item)
		return nil, err
	}
	return cache, nil
}

func (d *Dao) addCacheNoteUser(c context.Context, mid int64, val *note.UserCache) error {
	conn := d.redis.Get(c)
	defer conn.Close()
	bs, err := json.Marshal(val)
	if err != nil {
		return errors.Wrapf(err, "addCacheNoteUser val(%+v)", val)
	}
	if _, err = conn.Do("SETEX", userKey(mid), d.noteExpire, bs); err != nil {
		return errors.Wrapf(err, "addCacheNoteUser val(%+v)", val)
	}
	return nil
}

func (d *Dao) cacheNoteList(c context.Context, mid, min, max, total int64) (keys []string, err error) {
	var (
		conn = d.redis.Get(c)
		key  = listKey(mid)
	)
	defer conn.Close()
	// 分页数据
	if max > 0 {
		if keys, err = redis.Strings(conn.Do("ZREVRANGE", key, min, max)); err != nil {
			return nil, errors.Wrapf(err, "cacheNoteList mid(%d) min(%d) max(%d) total(%d)", mid, min, max, total)
		}
		return keys, nil
	}
	// 全量数据，分片获取
	round := int(total) / 100 // nolint:gomnd
	if int(total)%100 > 0 {
		round = round + 1
	}
	for i := 0; i < round; i++ {
		if err = conn.Send("ZREVRANGE", key, i*100, (i+1)*100-1); err != nil {
			return nil, errors.Wrapf(err, "cacheNoteList mid(%d) min(%d) max(%d) total(%d) i(%d)", mid, min, max, total, i)
		}
	}
	if err = conn.Flush(); err != nil {
		return nil, errors.Wrapf(err, "cacheNoteList mid(%d) min(%d) max(%d) total(%d)", mid, min, max, total)
	}
	for i := 0; i < round; i++ {
		var tmp []string
		if tmp, err = redis.Strings(conn.Receive()); err != nil {
			return nil, errors.Wrapf(err, "cacheNoteList mid(%d) min(%d) max(%d) total(%d) i(%d)", mid, min, max, total, i)
		}
		keys = append(keys, tmp...)
	}
	return keys, nil
}

func (d *Dao) AddCacheAllNoteList(c context.Context, mid int64, val []*note.NtList) error {
	var (
		key = listKey(mid)
	)
	conn := d.redis.Get(c)
	defer conn.Close()
	args := redis.Args{}.Add(key)
	for _, v := range val {
		args = args.Add(v.Mtime).Add(fmt.Sprintf("%d-%d", v.NoteId, v.Oid))
	}
	err := conn.Send("ZADD", args...)
	if err != nil {
		return errors.Wrapf(err, "AddCacheAllNoteList mid(%d)", mid)
	}
	if err = conn.Flush(); err != nil {
		return errors.Wrapf(err, "AddCacheAllNoteList mid(%d)", mid)

	}
	if _, err = conn.Receive(); err != nil {
		return errors.Wrapf(err, "AddCacheAllNoteList mid(%d)", mid)
	}
	return nil
}

/*func (d *Dao) delKey(c context.Context, key string) error {
	conn := d.redis.Get(c)
	defer conn.Close()
	if _, err := conn.Do("DEL", key); err != nil {
		return errors.Wrapf(err, "DelKey key(%s)", key)
	}
	return nil
}*/

func detailKey(noteId int64) string {
	return fmt.Sprintf("%d%s", noteId, _keyDetail)
}

func contentKey(noteId int64) string {
	return fmt.Sprintf("%d%s", noteId, _keyContent)
}

func userKey(mid int64) string {
	return fmt.Sprintf("%d%s", mid, _keyUser)
}

func listKey(mid int64) string {
	return fmt.Sprintf("%d%s", mid, _keyList)
}

func (d *Dao) aidKey(req *notegrpc.NoteListInArcReq) string {
	if req.OidType == note.OidTypeCheese {
		return fmt.Sprintf("%d_%d_%d%s", req.Oid, req.Mid, note.OidTypeCheese, _keyAid)
	}
	return fmt.Sprintf("%d_%d%s", req.Oid, req.Mid, _keyAid)
}
