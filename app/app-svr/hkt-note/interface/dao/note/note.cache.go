package note

import (
	"context"
	"fmt"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/app-svr/hkt-note/interface/model/note"
	notegrpc "go-gateway/app/app-svr/hkt-note/service/api"

	"github.com/pkg/errors"
)

const (
	_keyAid  = "_note"
	_keyList = "_note_list"
)

// RsSetNX NXset get
func (d *Dao) NoteAidSetNX(c context.Context, mid int64, req *note.NoteAddReq) (bool, error) {
	var (
		rkey = d.AidKey(req, mid)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	reply, err := redis.String(conn.Do("SET", rkey, req.NoteId, "EX", d.noteExpire, "NX"))
	if err != nil {
		if err == redis.ErrNil { // 锁已存在，未拿到
			return false, nil
		}
		return false, err
	}
	if reply != "OK" {
		return false, nil
	}
	log.Warn("noteInfo NoteAidSetNX rkey(%s) mid(%d) req(%+v) get the lock", rkey, mid, req)
	return true, nil
}

func (d *Dao) RemCacheNoteList(c context.Context, val map[int64]*notegrpc.SimpleNoteCard) error {
	conn := d.redis.Get(c)
	defer conn.Close()
	for _, v := range val {
		if err := conn.Send("ZREM", listKey(v.Mid), listVal(v.NoteId, v.Oid)); err != nil {
			return errors.Wrapf(err, "RemCacheNoteList dtl(%+v)", v)
		}
	}
	if err := conn.Flush(); err != nil {
		return errors.Wrapf(err, "RemCacheNoteList val(%+v)", val)
	}
	return nil
}

func (d *Dao) DelKey(c context.Context, key string) error {
	conn := d.redis.Get(c)
	defer conn.Close()
	if _, err := conn.Do("DEL", key); err != nil {
		log.Error("NoteError DelKey(%s) err(%+v)", key, err)
		return err
	}
	return nil
}

func (d *Dao) AidKey(req *note.NoteAddReq, mid int64) string {
	if req.OidType == note.OidTypeCheese {
		return fmt.Sprintf("%d_%d_%d%s", req.Oid, mid, note.OidTypeCheese, _keyAid)
	}
	return fmt.Sprintf("%d_%d%s", req.Oid, mid, _keyAid)
}

func listKey(mid int64) string {
	return fmt.Sprintf("%d%s", mid, _keyList)
}

func listVal(noteId, oid int64) string {
	return fmt.Sprintf("%d-%d", noteId, oid)
}
