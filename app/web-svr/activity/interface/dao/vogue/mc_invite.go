package dao

import (
	"context"
	"fmt"

	"go-common/library/cache/memcache"
	"go-common/library/log"
	model "go-gateway/app/web-svr/activity/interface/model/vogue"
)

const (
	_inviteList = "invite_list_%d_%d"
)

func keyInviteList(uid int64, id int64) string {
	return fmt.Sprintf(_inviteList, uid, id)
}

func (d *Dao) CacheInviteList(c context.Context, uid int64, id int64) (res []*model.Invite, err error) {
	key := keyInviteList(uid, id)
	if err := d.mc.Get(c, key).Scan(&res); err != nil {
		if err == memcache.ErrNotFound {
			err = nil
			res = nil
		} else {
			log.Error("conn.Get error(%v)", err)
		}
	}
	return
}

func (d *Dao) AddCacheInviteList(c context.Context, uid int64, value []*model.Invite, id int64) (err error) {
	key := keyInviteList(uid, id)
	if err = d.mc.Set(c, &memcache.Item{
		Key:        key,
		Object:     value,
		Expiration: d.confExpire,
		Flags:      memcache.FlagJSON,
	}); err != nil {
		log.Error("AddCacheGoodsList(%d) value(%v) error(%v)", key, value, err)
		return
	}
	return
}

func (d *Dao) DelCacheInviteList(c context.Context, uid int64, id int64) (err error) {
	key := keyInviteList(uid, id)
	if err = d.mc.Delete(c, key); err != nil {
		if err == memcache.ErrNotFound {
			err = nil
		} else {
			log.Error("DelCacheInviteList(%d) value(%v) error(%v)", key, err)
		}
	}
	return
}
