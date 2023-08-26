package unicom

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"go-common/library/cache/memcache"
	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-wall/interface/model/unicom"
)

const (
	_prefix         = "unicoms_user_%v"
	_userbindkey    = "unicoms_user_bind_%d"
	_userpackkey    = "unicom_user_pack_%d"
	_usermobInfoKey = "unicom_user_mob_%s_%d"
	_couponInfoKey  = "coupon_info_%d_%s"
)

const (
	// 固定两天过期用来资产侧异步对账用的
	_couponV2Expired = 172800
)

func keyUnicom(usermob string) string {
	return fmt.Sprintf(_prefix, usermob)
}

func keyUserBind(mid int64) string {
	return fmt.Sprintf(_userbindkey, mid)
}

func keyUserPack(id int64) string {
	return fmt.Sprintf(_userpackkey, id)
}

func keyUsermobInfo(fakeID string, period int64) string {
	return fmt.Sprintf(_usermobInfoKey, fakeID, period)
}

func keyCouponInfo(mid int64, bizId string) string {
	return fmt.Sprintf(_couponInfoKey, mid, bizId)
}

// AddUnicomCache
func (d *Dao) AddUnicomCache(c context.Context, usermob string, u []*unicom.Unicom) (err error) {
	var (
		key  = keyUnicom(usermob)
		conn = d.mc.Conn(c)
	)
	if err = conn.Set(&memcache.Item{Key: key, Object: u, Flags: memcache.FlagJSON, Expiration: d.expire}); err != nil {
		log.Error("addUnicomCache d.mc.Set(%s,%v) error(%v)", key, u, err)
	}
	conn.Close()
	return
}

// UnicomCache
func (d *Dao) UnicomCache(ctx context.Context, usermob string) ([]*unicom.Unicom, error) {
	key := keyUnicom(usermob)
	conn := d.mc.Conn(ctx)
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

// UserBindCache user bind cache
func (d *Dao) UserBindCache(c context.Context, mid int64) (ub *unicom.UserBind, err error) {
	var (
		key  = keyUserBind(mid)
		conn = d.mc.Conn(c)
		r    *memcache.Item
	)
	defer conn.Close()
	if r, err = conn.Get(key); err != nil {
		log.Error("UserBindCache MemchDB.Get(%s) error(%v)", key, err)
		return
	}
	if err = conn.Scan(r, &ub); err != nil {
		log.Error("r.Scan(%s) error(%v)", r.Value, err)
	}
	return
}

// AddUserBindCache add user user bind cache
func (d *Dao) AddUserBindCache(c context.Context, mid int64, ub *unicom.UserBind) (err error) {
	var (
		key  = keyUserBind(mid)
		conn = d.mc.Conn(c)
	)
	if err = conn.Set(&memcache.Item{Key: key, Object: ub, Flags: memcache.FlagJSON, Expiration: 0}); err != nil {
		log.Error("AddUserBindCache d.mc.Set(%s,%v) error(%v)", key, ub, err)
	}
	conn.Close()
	return
}

// DeleteUserBindCache delete user bind cache
func (d *Dao) DeleteUserBindCache(c context.Context, mid int64) (err error) {
	var (
		key  = keyUserBind(mid)
		conn = d.mc.Conn(c)
	)
	defer conn.Close()
	if err = conn.Delete(key); err != nil {
		if err == memcache.ErrNotFound {
			err = nil
			return
		}
		log.Error("DeleteUserBindCache MemchDB.Delete(%s) error(%v)", key, err)
		return
	}
	return
}

// UserPackCache user packs
func (d *Dao) UserPackCache(c context.Context, id int64) (res *unicom.UserPack, err error) {
	var (
		key  = keyUserPack(id)
		conn = d.mc.Conn(c)
		r    *memcache.Item
	)
	defer conn.Close()
	if r, err = conn.Get(key); err != nil {
		log.Error("UserPackCache MemchDB.Get(%s) error(%v)", key, err)
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
		conn = d.mc.Conn(c)
	)
	if err = conn.Set(&memcache.Item{Key: key, Object: u, Flags: memcache.FlagJSON, Expiration: d.flowKeyExpired}); err != nil {
		log.Error("AddUserPackCache d.mc.Set(%s,%v) error(%v)", key, u, err)
	}
	conn.Close()
	return
}

// UsersBindCache user bind cache
func (d *Dao) UsersBindCache(c context.Context, mids []int64) (ubs map[int64]*unicom.UserBind, err error) {
	var (
		keys []string
		rs   map[string]*memcache.Item
		conn = d.mc.Conn(c)
	)
	for _, mid := range mids {
		key := keyUserBind(mid)
		keys = append(keys, key)
	}
	defer conn.Close()
	if rs, err = conn.GetMulti(keys); err != nil {
		log.Error("UsersBindCache MemchDB.GetMulti(%s) error(%v)", keys, err)
		return
	}
	ubs = map[int64]*unicom.UserBind{}
	for _, r := range rs {
		ub := &unicom.UserBind{}
		if err = conn.Scan(r, &ub); err != nil {
			log.Error("r.Scan(%s) error(%v)", r.Value, err)
			return
		}
		ubs[ub.Mid] = ub
	}
	return
}

func (d *Dao) AddUsermobInfoCache(c context.Context, fakeID string, period int64, info *unicom.UserMobInfo) error {
	key := keyUsermobInfo(fakeID, period)
	conn := d.mc.Conn(c)
	defer conn.Close()
	item := &memcache.Item{Key: key, Object: info, Flags: memcache.FlagJSON, Expiration: d.usermobExpire}
	// 空缓存
	if info.Usermob == "" {
		item.Expiration = d.emptyExpire
	}
	if err := conn.Set(item); err != nil {
		log.Error("[dao.AddUsermobInfoCache] d.mc.Set(%s,%v) error(%v)", key, info, err)
		return err
	}
	return nil
}

func (d *Dao) GetUsermobInfoCache(c context.Context, fakeID string, period int64) (*unicom.UserMobInfo, error) {
	key := keyUsermobInfo(fakeID, period)
	conn := d.mc.Conn(c)
	defer conn.Close()
	result, err := conn.Get(key)
	if err != nil {
		log.Error("[dao.GetUsermobInfoCache] d.mc.Get(%s) error(%v)", key, err)
		return nil, err
	}
	var info *unicom.UserMobInfo
	if err = conn.Scan(result, &info); err != nil {
		log.Error("[dao.GetUsermobInfoCache] r.Scan(%s) error(%v)", result.Value, err)
		return nil, err
	}
	return info, nil
}

// 缓存优惠券请求参数用于验证合法性
func (d *Dao) AddCouponV2ReqCache(c context.Context, info *unicom.CouponParam) error {
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

func (d *Dao) GetCouponV2ReqCache(c context.Context, mid int64, bizId string) (*unicom.CouponParam, error) {
	key := keyCouponInfo(mid, bizId)
	data := &unicom.CouponParam{}
	reply, err := redis.Bytes(d.redis.Do(c, "GET", key))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if err = json.Unmarshal(reply, data); err != nil {
		return nil, errors.WithStack(err)
	}
	return data, nil
}
