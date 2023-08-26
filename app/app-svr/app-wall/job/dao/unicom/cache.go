package unicom

import (
	"context"
	"encoding/json"
	"fmt"

	unicomdl "go-gateway/app/app-svr/app-wall/interface/model/unicom"

	"go-common/library/cache/memcache"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-wall/job/model/unicom"

	"github.com/pkg/errors"
)

const (
	_prefix         = "unicoms_user_%v"
	_userbindkey    = "unicoms_user_bind_%d"
	_userpackkey    = "unicom_user_pack_%d"
	_mobilePrefix   = "mobile_users_%v"
	_usermobInfoKey = "unicom_user_mob_%s_%d"
	_couponInfoKey  = "coupon_info_%d_%s"
)

const (
	// 固定两天过期用来资产侧异步对账用的
	_couponV2Expired = 172800
)

func keyUserBind(mid int64) string {
	return fmt.Sprintf(_userbindkey, mid)
}

func keyUnicom(usermob string) string {
	return fmt.Sprintf(_prefix, usermob)
}

func keyUserPack(id int64) string {
	return fmt.Sprintf(_userpackkey, id)
}

func keyMobile(usermob string) string {
	return fmt.Sprintf(_mobilePrefix, usermob)
}

func keyUsermobInfo(fakeID string, period int64) string {
	return fmt.Sprintf(_usermobInfoKey, fakeID, period)
}

func keyCouponInfo(mid int64, bizId string) string {
	return fmt.Sprintf(_couponInfoKey, mid, bizId)
}

// UserBindCache user bind cache
func (d *Dao) UserBindCache(c context.Context, mid int64) (*unicom.UserBind, error) {
	key := keyUserBind(mid)
	conn := d.mc.Get(c)
	defer conn.Close()
	r, err := conn.Get(key)
	if err != nil {
		if err == memcache.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	var res *unicom.UserBind
	if err = conn.Scan(r, &res); err != nil {
		return nil, err
	}
	return res, nil
}

func (d *Dao) AddUserBindCache(ctx context.Context, mid int64, ub *unicom.UserBind) error {
	key := keyUserBind(mid)
	conn := d.mc.Get(ctx)
	defer conn.Close()
	if err := conn.Set(&memcache.Item{Key: key, Object: ub, Flags: memcache.FlagJSON, Expiration: 0}); err != nil {
		return err
	}
	return nil
}

func (d *Dao) DeleteUserBindCache(ctx context.Context, mid int64) error {
	key := keyUserBind(mid)
	conn := d.mc.Get(ctx)
	defer conn.Close()
	if err := conn.Delete(key); err != nil {
		if err == memcache.ErrNotFound {
			return nil
		}
		return err
	}
	return nil
}

func (d *Dao) UnicomCache(c context.Context, usermob string) ([]*unicom.Unicom, error) {
	key := keyUnicom(usermob)
	conn := d.mc.Get(c)
	defer conn.Close()
	r, err := conn.Get(key)
	if err != nil {
		if err == memcache.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	var res []*unicom.Unicom
	if err = conn.Scan(r, &res); err != nil {
		return nil, err
	}
	return res, nil
}

func (d *Dao) DeleteUnicomCache(ctx context.Context, usermob string) error {
	key := keyUnicom(usermob)
	conn := d.mc.Get(ctx)
	defer conn.Close()
	if err := conn.Delete(key); err != nil {
		if err == memcache.ErrNotFound {
			return nil
		}
		return err
	}
	return nil
}

// UserPackCache user packs
func (d *Dao) UserPackCache(c context.Context, id int64) (res *unicom.UserPack, err error) {
	var (
		key  = keyUserPack(id)
		conn = d.mc.Get(c)
		r    *memcache.Item
	)
	defer conn.Close()
	if r, err = conn.Get(key); err != nil {
		log.Error("UserBindCache MemchDB.Get(%s) error(%v)", key, err)
		return
	}
	if err = conn.Scan(r, &res); err != nil {
		log.Error("r.Scan(%s) error(%v)", r.Value, err)
	}
	return
}

// AddUserPackCache add user pack cache
func (d *Dao) AddUserPackCache(c context.Context, id int64, u *unicom.UserPack) (err error) {
	var (
		key  = keyUserPack(id)
		conn = d.mc.Get(c)
	)
	if err = conn.Set(&memcache.Item{Key: key, Object: u, Flags: memcache.FlagJSON, Expiration: d.flowKeyExpired}); err != nil {
		log.Error("AddUserPackCache d.mc.Set(%s,%v) error(%v)", key, u, err)
	}
	conn.Close()
	return
}

// DeleteUserPackCache delete user pack cache
func (d *Dao) DeleteUserPackCache(c context.Context, id int64) (err error) {
	var (
		key  = keyUserPack(id)
		conn = d.mc.Get(c)
	)
	defer conn.Close()
	if err = conn.Delete(key); err != nil {
		if err == memcache.ErrNotFound {
			err = nil
			return
		}
		log.Error("DeleteUserPackCache MemchDB.Delete(%s) error(%v)", key, err)
		return
	}
	return
}

func (d *Dao) AddUnicomCache(ctx context.Context, usermob string, u []*unicom.Unicom) error {
	key := keyUnicom(usermob)
	conn := d.mc.Get(ctx)
	defer conn.Close()
	if err := conn.Set(&memcache.Item{Key: key, Object: u, Flags: memcache.FlagJSON, Expiration: d.expire}); err != nil {
		return err
	}
	return nil
}

func (d *Dao) DeleteMobileCache(ctx context.Context, usermob string) error {
	key := keyMobile(usermob)
	conn := d.mc.Get(ctx)
	defer conn.Close()
	if err := conn.Delete(key); err != nil {
		if err == memcache.ErrNotFound {
			return nil
		}
		return err
	}
	return nil
}

func (d *Dao) DelUsermobInfoCache(c context.Context, fakeID string, period int64) error {
	key := keyUsermobInfo(fakeID, period)
	conn := d.mc.Get(c)
	defer conn.Close()
	if err := conn.Delete(key); err != nil {
		if err == memcache.ErrNotFound {
			return nil
		}
		log.Error("[dao.DelUsermobInfoCache] d.mc.Delete(%s) error(%v)", key, err)
		return err
	}
	return nil
}

// 缓存优惠券请求参数用于验证合法性
func (d *Dao) AddCouponV2ReqCache(c context.Context, info *unicomdl.CouponParam) error {
	key := keyCouponInfo(info.AssetRequest.Mid, info.AssetRequest.SourceBizId)
	data, err := json.Marshal(info)
	if err != nil {
		return errors.WithStack(err)
	}
	if _, err = d.redis.Do(c, "SETEX", key, _couponV2Expired, data); err != nil {
		return errors.WithStack(err)
	}
	return nil
}
