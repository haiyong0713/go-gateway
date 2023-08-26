package recommend

import (
	"context"

	"go-common/library/cache/memcache"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"

	"github.com/pkg/errors"
)

const (
	_prefixRcmd           = "rc3"
	_prefixFollowModeList = "fml"
)

func keyRcmd() string {
	return _prefixRcmd
}

func keyFollowModeList() string {
	return _prefixFollowModeList
}

// RcmdCache get ai cache data from cache
func (d *Dao) RcmdCache(c context.Context) (is []*ai.Item, err error) {
	key := keyRcmd()
	if err = d.mc.Get(c, key).Scan(&is); err != nil {
		if err == memcache.ErrNotFound {
			err = nil
			return
		}
		err = errors.Wrap(err, key)
	}
	return
}

// AddFollowModeListCache is.
func (d *Dao) AddFollowModeListCache(c context.Context, list map[int64]struct{}) (err error) {
	key := keyFollowModeList()
	item := &memcache.Item{Key: key, Object: list, Flags: memcache.FlagJSON, Expiration: d.expireMc}
	if err = d.mc.Set(c, item); err != nil {
		err = errors.Wrap(err, key)
	}
	return
}

// FollowModeListCache is.
func (d *Dao) FollowModeListCache(c context.Context) (list map[int64]struct{}, err error) {
	key := keyFollowModeList()
	if err = d.mc.Get(c, key).Scan(&list); err != nil {
		if err == memcache.ErrNotFound {
			err = nil
			return
		}
		err = errors.Wrap(err, key)
	}
	return
}
