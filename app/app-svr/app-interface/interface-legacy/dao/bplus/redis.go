package bplus

import (
	"context"
	"strconv"

	"go-common/library/cache/redis"
	"go-common/library/log"
	xtime "go-common/library/time"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model/space"

	"github.com/pkg/errors"
)

const (
	_prefixContributeAttr            = "cba_"
	_prefixContribute                = "cb_"
	_prefixContributeAttrCooperation = "cbacoop_"
	_prefixContributeCooperation     = "cbcoop_v2_"
	_prefixContributeAttrComic       = "cbacomic_v2_"
	_prefixContributeComic           = "cbcomic_v2_"
)

func keyContributeAttr(vmid int64) string {
	return _prefixContributeAttr + strconv.FormatInt(vmid, 10)
}

func keyContribute(vmid int64) string {
	return _prefixContribute + strconv.FormatInt(vmid, 10)
}

func keyContributeAttrCooperation(vmid int64) string {
	return _prefixContributeAttrCooperation + strconv.FormatInt(vmid, 10)
}

func keyContributeCooperation(vmid int64) string {
	return _prefixContributeCooperation + strconv.FormatInt(vmid, 10)
}

func keyContributeAttrComic(vmid int64) string {
	return _prefixContributeAttrComic + strconv.FormatInt(vmid, 10)
}

func keyContributeComic(vmid int64) string {
	return _prefixContributeComic + strconv.FormatInt(vmid, 10)
}

// AddContributeCache .
func (d *Dao) AddContributeCache(c context.Context, vmid int64, attrs *space.Attrs, items []*space.Item, isCooperation, isComic bool) (err error) {
	var (
		attr int32
		key  string
	)
	conn := d.redis.Get(c)
	defer conn.Close()
	// comic > cooperation > other
	if isComic {
		key = keyContributeComic(vmid)
	} else if isCooperation {
		key = keyContributeCooperation(vmid)
	} else {
		key = keyContribute(vmid)
	}
	for _, item := range items {
		score := item.CTime.Time().Unix()
		item.FormatKey()
		if err = conn.Send("ZADD", key, score, item.Member); err != nil {
			err = errors.Wrapf(err, "conn.Send(ZADD,%s,%d,%d)", key, score, item.Member)
			return
		}
	}
	if err = conn.Send("EXPIRE", key, d.contributeExpire); err != nil {
		err = errors.Wrapf(err, "conn.Send(Expire,%s,%d)", key, d.contributeExpire)
		return
	}
	var keyAttr string
	// comic > cooperation > other
	if isComic {
		keyAttr = keyContributeAttrComic(vmid)
	} else if isCooperation {
		keyAttr = keyContributeAttrCooperation(vmid)
	} else {
		keyAttr = keyContributeAttr(vmid)
	}
	if attrs != nil {
		if attrs.Archive {
			attr = model.AttrSet(attr, model.AttrYes, model.AttrBitArchive)
		}
		if attrs.Article {
			attr = model.AttrSet(attr, model.AttrYes, model.AttrBitArticle)
		}
		if attrs.Clip {
			attr = model.AttrSet(attr, model.AttrYes, model.AttrBitClip)
		}
		if attrs.Album {
			attr = model.AttrSet(attr, model.AttrYes, model.AttrBitAlbum)
		}
		if attrs.Audio {
			attr = model.AttrSet(attr, model.AttrYes, model.AttrBitAudio)
		}
		if attrs.Comic {
			attr = model.AttrSet(attr, model.AttrYes, model.AttrBitComic)
		}
	}
	if err = conn.Send("SET", keyAttr, attr); err != nil {
		err = errors.Wrapf(err, "conn.Send(SET,%s,%d)", keyAttr, attr)
		return
	}
	if err = conn.Send("EXPIRE", keyAttr, d.contributeExpire); err != nil {
		err = errors.Wrapf(err, "conn.Send(Expire,%s,%d)", key, d.contributeExpire)
		return
	}
	if err = conn.Flush(); err != nil {
		return
	}
	for i := 0; i < len(items)+3; i++ {
		if _, err = conn.Receive(); err != nil {
			return
		}
	}
	return
}

// RangeContributeCache .
func (d *Dao) RangeContributeCache(c context.Context, vmid int64, pn, ps int, isCooperation, isComic bool) (items []*space.Item, err error) {
	var (
		key string
		vs  []interface{}
	)
	conn := d.redis.Get(c)
	defer conn.Close()
	// comic > cooperation > other
	if isComic {
		key = keyContributeComic(vmid)
	} else if isCooperation {
		key = keyContributeCooperation(vmid)
	} else {
		key = keyContribute(vmid)
	}
	start := (pn - 1) * ps
	stop := pn*ps - 1
	if err = conn.Send("ZREVRANGE", key, start, stop, "WITHSCORES"); err != nil {
		err = errors.Wrapf(err, "conn.Send(ZREVRANGE,%s,%d,%d)", key, start, stop)
		return
	}
	if err = conn.Send("EXPIRE", key, d.contributeExpire); err != nil {
		err = errors.Wrapf(err, "conn.Send(Expire,%s,%d)", key, d.contributeExpire)
		return
	}
	if err = conn.Flush(); err != nil {
		return
	}
	if vs, err = redis.Values(conn.Receive()); err != nil {
		return
	}
	if _, err = conn.Receive(); err != nil {
		return
	}
	if len(vs) == 0 {
		return
	}
	items = make([]*space.Item, 0, ps)
	for len(vs) > 0 {
		var (
			member int64
			score  int64
		)
		if vs, err = redis.Scan(vs, &member, &score); err != nil {
			log.Error("redis.Scan(%v) error(%v)", vs, err)
			err = nil
			continue
		}
		if member != 0 && score != 0 {
			item := &space.Item{Member: member, CTime: xtime.Time(score)}
			item.ParseKey()
			if item.Goto != "" {
				items = append(items, item)
			}
		}
	}
	return
}

func (d *Dao) RangeContributionCache(c context.Context, vmid int64, cursor *model.Cursor) (items []*space.Item, err error) {
	var (
		vs          []interface{}
		rank        int64
		start, stop int64
	)
	key := keyContribute(vmid)
	conn := d.redis.Get(c)
	defer conn.Close()
	if cursor.MoveUpward() || cursor.MoveDownward() {
		if rank, err = redis.Int64(conn.Do("ZREVRANK", key, cursor.Current)); err != nil {
			if err == redis.ErrNil {
				err = nil
				return
			}
			err = errors.Wrapf(err, "conn.Do(ZREVRANK,%s,%d)", key, cursor.Current)
			return
		}
	}
	if cursor.Latest() {
		start = 0
		stop = rank + int64(cursor.Size) - 1
	} else if cursor.MoveUpward() {
		if rank == 0 {
			return
		}
		if start = rank - int64(cursor.Size); start < 0 {
			start = 0
		}
		stop = rank - 1
	} else if cursor.MoveDownward() {
		start = rank + 1
		stop = rank + int64(cursor.Size)
	}
	if err = conn.Send("ZREVRANGE", key, start, stop, "WITHSCORES"); err != nil {
		err = errors.Wrapf(err, "conn.Send(ZREVRANGE,%s,%d,%d)", key, start, stop)
		return
	}
	if err = conn.Send("EXPIRE", key, d.contributeExpire); err != nil {
		err = errors.Wrapf(err, "conn.Send(Expire,%s,%d)", key, d.contributeExpire)
		return
	}
	if err = conn.Flush(); err != nil {
		return
	}
	if vs, err = redis.Values(conn.Receive()); err != nil {
		return
	}
	if _, err = conn.Receive(); err != nil {
		return
	}
	if len(vs) == 0 {
		return
	}
	items = make([]*space.Item, 0, len(vs))
	for len(vs) > 0 {
		var (
			member int64
			score  int64
		)
		if vs, err = redis.Scan(vs, &member, &score); err != nil {
			log.Error("redis.Scan(%v) error(%v)", vs, err)
			err = nil
			continue
		}
		if member != 0 && score != 0 {
			item := &space.Item{Member: member, CTime: xtime.Time(score)}
			item.ParseKey()
			if item.Goto != "" {
				items = append(items, item)
			}
		}
	}
	return
}

// AttrCache .
func (d *Dao) AttrCache(c context.Context, vmid int64, isCooperation, isComic bool) (attrs *space.Attrs, err error) {
	var (
		attr int64
		key  string
	)
	conn := d.redis.Get(c)
	defer conn.Close()
	// comic > cooperation > other
	if isComic {
		key = keyContributeAttrComic(vmid)
	} else if isCooperation {
		key = keyContributeAttrCooperation(vmid)
	} else {
		key = keyContributeAttr(vmid)
	}
	if attr, err = redis.Int64(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		err = errors.Wrapf(err, "conn.Do(GET,%s)", key)
		return
	}
	attrs = &space.Attrs{}
	if model.AttrVal(int32(attr), model.AttrBitArchive) == model.AttrYes {
		attrs.Archive = true
	}
	if model.AttrVal(int32(attr), model.AttrBitArticle) == model.AttrYes {
		attrs.Article = true
	}
	if model.AttrVal(int32(attr), model.AttrBitClip) == model.AttrYes {
		attrs.Clip = true
	}
	if model.AttrVal(int32(attr), model.AttrBitAlbum) == model.AttrYes {
		attrs.Album = true
	}
	if model.AttrVal(int32(attr), model.AttrBitAudio) == model.AttrYes {
		attrs.Audio = true
	}
	return
}
