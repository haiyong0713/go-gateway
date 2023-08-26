package recommend

import (
	"context"

	"go-common/library/cache/memcache"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"

	"github.com/pkg/errors"
)

const (
	_prefixRcmd = "rc3"
)

func keyRcmd() string {
	return _prefixRcmd
}

// RcmdCache get ai cache data from cache
func (d *Dao) RcmdCache(c context.Context) (is []*ai.Item, err error) {
	var r *memcache.Item
	conn := d.mc.Get(c)
	key := keyRcmd()
	defer conn.Close()
	if r, err = conn.Get(key); err != nil {
		if err == memcache.ErrNotFound {
			err = nil
			return
		}
		err = errors.Wrap(err, key)
		return
	}
	if err = conn.Scan(r, &is); err != nil {
		err = errors.Wrapf(err, "%s", r.Value)
	}
	return
}
