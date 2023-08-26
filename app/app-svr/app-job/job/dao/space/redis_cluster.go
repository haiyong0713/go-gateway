package space

import (
	"context"
	"strconv"

	"go-gateway/app/app-svr/app-job/job/model/space"

	"github.com/pkg/errors"
)

const (
	_prefixContributeAttr            = "cba_"
	_prefixContribute                = "cb_"
	_prefixContributeAttrCooperation = "cbacoop_"
	_prefixContributeCooperation     = "cbcoop_"
	_prefixContributeAttrComic       = "cbacomic_"
	_prefixContributeComic           = "cbcomic_"
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

func (d *Dao) DelContrbIDCache(c context.Context, vmid, id int64, gt string) (err error) {
	conn := d.interRds.Get(c)
	key := keyContribute(vmid)
	member := space.FormatKey(id, gt)
	if _, err = conn.Do("ZREM", key, member); err != nil {
		err = errors.Wrapf(err, "conn.Do(ZREM,%s,%d)", key, id)
	}
	conn.Close()
	return
}

func (d *Dao) DelContrbCache(c context.Context, vmid int64, isCooperation, isComic bool) (err error) {
	var key string
	conn := d.interRds.Get(c)
	// comic > cooperation > other
	if isComic {
		key = keyContributeComic(vmid)
	} else if isCooperation {
		key = keyContributeCooperation(vmid)
	} else {
		key = keyContribute(vmid)
	}
	if _, err = conn.Do("DEL", key); err != nil {
		err = errors.Wrapf(err, "conn.Do(DEL,%s)", key)
	}
	conn.Close()
	return
}

// AddContributeList .
func (d *Dao) AddContrbList(c context.Context, vmid int64, items []*space.Item, isCooperation, isComic bool) (leftItems []*space.Item, err error) {
	if len(items) == 0 {
		return
	}
	var key string
	conn := d.interRds.Get(c)
	defer conn.Close()
	// comic > cooperation > other
	if isComic {
		key = keyContributeComic(vmid)
	} else if isCooperation {
		key = keyContributeCooperation(vmid)
	} else {
		key = keyContribute(vmid)
	}
	for len(items) > 0 {
		n := 128
		if l := len(items); n > l {
			n = l
		}
		for _, item := range items[:n] {
			score := item.CTime.Time().Unix()
			item.FormatKey()
			if err = conn.Send("ZADD", key, score, item.Member); err != nil {
				leftItems = items
				err = errors.Wrapf(err, "conn.Send(ZADD,%s,%d,%d)", key, score, item.Member)
				return
			}
		}
		if err = conn.Flush(); err != nil {
			leftItems = items
			return
		}
		for i := 0; i < n; i++ {
			if _, err = conn.Receive(); err != nil {
				leftItems = items
				return
			}
		}
		items = items[n:]
	}
	return
}

func (d *Dao) AddContrbAttr(c context.Context, vmid int64, attrs *space.Attrs, isCooperation, isComic bool) (err error) {
	var (
		key     string
		keyAttr string
	)
	conn := d.interRds.Get(c)
	defer conn.Close()
	// comic > cooperation > other
	if isComic {
		key = keyContributeComic(vmid)
		keyAttr = keyContributeAttrComic(vmid)
	} else if isCooperation {
		key = keyContributeCooperation(vmid)
		keyAttr = keyContributeAttrCooperation(vmid)
	} else {
		key = keyContribute(vmid)
		keyAttr = keyContributeAttr(vmid)
	}
	if err = conn.Send("EXPIRE", key, d.expireContribute); err != nil {
		err = errors.Wrapf(err, "conn.Send(EXPIRE,%s,%d)", key, d.expireContribute)
		return
	}
	attr := attrs.Attr()
	if err = conn.Send("SET", keyAttr, attr); err != nil {
		err = errors.Wrapf(err, "conn.Send(SET,%s,%d)", keyAttr, attr)
		return
	}
	if err = conn.Send("EXPIRE", keyAttr, d.expireContribute); err != nil {
		err = errors.Wrapf(err, "conn.Send(EXPIRE,%s,%d)", keyAttr, d.expireContribute)
		return
	}
	if err = conn.Flush(); err != nil {
		return
	}
	for i := 0; i < 3; i++ {
		if _, err = conn.Receive(); err != nil {
			return
		}
	}
	return
}
