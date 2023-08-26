package mobile

import (
	"context"
	"fmt"

	"go-common/library/cache/memcache"
	"go-gateway/app/app-svr/app-wall/interface/model/mobile"
)

const (
	_prefix = "mobile_users_%v"
)

func keyMobile(usermob string) string {
	return fmt.Sprintf(_prefix, usermob)
}

func (d *Dao) AddMobileCache(ctx context.Context, usermob string, m []*mobile.Mobile) error {
	key := keyMobile(usermob)
	conn := d.mc.Conn(ctx)
	defer conn.Close()
	if err := conn.Set(&memcache.Item{Key: key, Object: m, Flags: memcache.FlagJSON, Expiration: d.expire}); err != nil {
		return err
	}
	return nil
}

func (d *Dao) MobileCache(ctx context.Context, usermob string) ([]*mobile.Mobile, error) {
	key := keyMobile(usermob)
	conn := d.mc.Conn(ctx)
	defer conn.Close()
	r, err := conn.Get(key)
	if err != nil {
		if err == memcache.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	var res []*mobile.Mobile
	if err = conn.Scan(r, &res); err != nil {
		return nil, err
	}
	return res, nil
}
