package space

import (
	"context"

	"go-gateway/app/app-svr/app-job/job/model/space"

	"github.com/pkg/errors"
)

func (d *Dao) DelContributeIDCache(c context.Context, vmid, id int64, gt string) (err error) {
	conn := d.redis.Get(c)
	key := keyContribute(vmid)
	member := space.FormatKey(id, gt)
	if _, err = conn.Do("ZREM", key, member); err != nil {
		err = errors.Wrapf(err, "conn.Do(ZREM,%s,%d)", key, id)
	}
	conn.Close()
	if d.c.Contribute.Cluster {
		err = d.DelContrbIDCache(c, vmid, id, gt)
	}
	return
}

func (d *Dao) DelContributeCache(c context.Context, vmid int64, isCooperation, isComic bool) (err error) {
	var key string
	conn := d.redis.Get(c)
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
	if d.c.Contribute.Cluster {
		err = d.DelContrbCache(c, vmid, isCooperation, isComic)
	}
	return
}

func (d *Dao) AddContributeList(c context.Context, vmid int64, items []*space.Item, isCooperation, isComic bool) (leftItems []*space.Item, err error) {
	if len(items) == 0 {
		return
	}
	contrbItems := items
	var key string
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
	if d.c.Contribute.Cluster {
		leftItems, err = d.AddContrbList(c, vmid, contrbItems, isCooperation, isComic)
	}
	return
}

func (d *Dao) AddContributeAttr(c context.Context, vmid int64, attrs *space.Attrs, isCooperation, isComic bool) (err error) {
	var (
		key     string
		keyAttr string
	)
	conn := d.redis.Get(c)
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
	if d.c.Contribute.Cluster {
		err = d.AddContrbAttr(c, vmid, attrs, isCooperation, isComic)
	}
	return
}
