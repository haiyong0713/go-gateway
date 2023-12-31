// Code generated by kratos tool mcgen. DO NOT EDIT.

/*
  Package page is a generated mc cache package.
  It is generated from:
  type _mc interface {
		// mc: -key=pageKey -struct_name=Dao
		CacheGetPageByID(c context.Context, id int64) (*model.ActPage, error)
		// mc: -key=pageKey -expire=d.pageExpire -encode=json -struct_name=Dao -check_null_code=$==nil||$.ID==0
		AddCacheGetPageByID(c context.Context, id int64, value *model.ActPage) error
	}
*/

package page

import (
	"context"
	"fmt"

	"go-common/library/cache/memcache"
	"go-common/library/log"
	model "go-gateway/app/web-svr/activity/interface/model/page"
)

var _ _mc

// CacheGetPageByID get data from mc
func (d *Dao) CacheGetPageByID(c context.Context, id int64) (res *model.ActPage, err error) {
	key := pageKey(id)
	res = &model.ActPage{}
	if err = d.mc.Get(c, key).Scan(res); err != nil {
		res = nil
		if err == memcache.ErrNotFound {
			err = nil
		}
	}
	if err != nil {
		log.Errorv(c, log.KV("CacheGetPageByID", fmt.Sprintf("%+v", err)), log.KV("key", key))
		return
	}
	return
}

// AddCacheGetPageByID Set data to mc
func (d *Dao) AddCacheGetPageByID(c context.Context, id int64, val *model.ActPage) (err error) {
	if val == nil {
		return
	}
	key := pageKey(id)
	item := &memcache.Item{Key: key, Object: val, Expiration: d.pageExpire, Flags: memcache.FlagJSON}
	if val == nil || val.ID == 0 {
		item.Expiration = 300
	}
	if err = d.mc.Set(c, item); err != nil {
		log.Errorv(c, log.KV("AddCacheGetPageByID", fmt.Sprintf("%+v", err)), log.KV("key", key))
		return
	}
	return
}
