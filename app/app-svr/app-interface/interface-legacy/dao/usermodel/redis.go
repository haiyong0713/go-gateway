package usermodel

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/log"

	antiadmdl "go-gateway/app/app-svr/app-interface/interface-legacy/model/anti_addiction"
	familymdl "go-gateway/app/app-svr/app-interface/interface-legacy/model/family"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model/usermodel"

	"github.com/pkg/errors"
)

const (
	_prefixModelUser           = "model_user_%d"
	_prefixDeviceModelUser     = "device_model_user_%s_%s"
	_teenagersAntiAddiction    = "teenagers_anti_addiction_%s_%d_%d"
	_teenagersAntiAddictionMID = "teenagers_anti_addiction_%s_%d"
	// 失错限制
	_teenagersUnlockErrorNumByDevice = "teenagers_unlock_%s_%d_%d"
	_teenagersUnlockErrorNumByMID    = "teenagers_unlock_%d_%d_%d"
)

func (d *dao) emptyExpire() int64 {
	rand.Seed(time.Now().UnixNano())
	return d.emptyCacheExpire + rand.Int63n(d.emptyCacheRand)
}

func (d *dao) cacheSFUserModels(mid int64, mobiApp, deviceToken string) string {
	if mid != 0 {
		return fmt.Sprintf(_prefixModelUser, mid)
	}
	return fmt.Sprintf(_prefixDeviceModelUser, mobiApp, deviceToken)
}

func keyAntiAddiction(deviceToken string, mid, day int64) string {
	return fmt.Sprintf(_teenagersAntiAddiction, deviceToken, mid, day)
}

func keyAntiAddictionMID(deviceToken string, day int64) string {
	return fmt.Sprintf(_teenagersAntiAddictionMID, deviceToken, day)
}

func familyRelsOfParentKey(mid int64) string {
	return fmt.Sprintf("fyrel_parent_%d", mid)
}

func familyRelsOfChildKey(mid int64) string {
	return fmt.Sprintf("fyrel_child_%d", mid)
}

func sleepRemindKey(mid int64) string {
	return fmt.Sprintf("sleep_remind_%d", mid)
}

func keyTeenagersUnlockErrorNumByDevice(deviceToken string, pwdFrom int32, day int64) string {
	return fmt.Sprintf(_teenagersUnlockErrorNumByDevice, deviceToken, pwdFrom, day)
}

func keyTeenagersUnlockErrorNumByMid(mid int64, pwdFrom int32, day int64) string {
	return fmt.Sprintf(_teenagersUnlockErrorNumByMID, mid, pwdFrom, day)
}

func (d *dao) AddCacheUserModels(ctx context.Context, mid int64, miss []*usermodel.User, mobiApp, deviceToken string) error {
	key := d.cacheSFUserModels(mid, mobiApp, deviceToken)
	bs, err := json.Marshal(miss)
	if err != nil {
		return err
	}
	if miss[0].ID == -1 {
		_, err := d.redis.Do(ctx, "SETEX", key, d.emptyExpire(), bs)
		return err
	}
	_, err = d.redis.Do(ctx, "SET", key, bs)
	return err
}

func (d *dao) delCacheUserModel(ctx context.Context, mid int64, mobiApp, deviceToken string) error {
	key := d.cacheSFUserModels(mid, mobiApp, deviceToken)
	if _, err := d.redis.Do(ctx, "DEL", key); err != nil {
		return errors.Wrapf(err, "conn.Do(DEL,%s)", key)
	}
	return nil
}

func (d *dao) delCacheSyncUserModel(ctx context.Context, mid int64, mobiApp, deviceToken string) error {
	key1 := d.cacheSFUserModels(mid, mobiApp, deviceToken)
	key2 := d.cacheSFUserModels(0, mobiApp, deviceToken)
	if _, err := d.redis.Do(ctx, "DEL", key1, key2); err != nil {
		return errors.Wrapf(err, "redis.Do(DEL,%s,%s)", key1, key2)
	}
	return nil
}

func (d *dao) CacheUserModels(ctx context.Context, mid int64, mobiApp, deviceToken string) ([]*usermodel.User, error) {
	key := d.cacheSFUserModels(mid, mobiApp, deviceToken)
	bs, err := redis.Bytes(d.redis.Do(ctx, "GET", key))
	if err != nil {
		if err == redis.ErrNil {
			return nil, nil
		}
		return nil, err
	}
	var res []*usermodel.User
	if err := json.Unmarshal(bs, &res); err != nil {
		return nil, err
	}
	return res, nil
}

// 设置防沉迷时间
func (d *dao) SetCacheAntiAddictionTime(ctx context.Context, deviceToken string, mid, day, useTime int64) error {
	key := keyAntiAddiction(deviceToken, mid, day)
	keyMID := keyAntiAddictionMID(deviceToken, day)
	p := d.redis.Pipeline()
	p.Send("SETEX", key, d.antiAddictionExpire, useTime)
	p.Send("SETEX", keyMID, d.antiAddictionExpire, mid)
	if _, err := p.Exec(ctx); err != nil {
		log.Error("dao.SetAntiAddictionTime redis set error,err:(%v),key:(%s)", err, key)
		return err
	}
	return nil
}

// 获取防沉迷时间
func (d *dao) GetCacheAntiAddictionTime(ctx context.Context, deviceToken string, mid, day int64) (int64, error) {
	key := keyAntiAddiction(deviceToken, mid, day)
	res, err := redis.Int64(d.redis.Do(ctx, "GET", key))
	if err != nil {
		if err == redis.ErrNil {
			return 0, nil
		}
		log.Error("dao.GetAntiAddictionTime redis get error,err:(%v),key:(%s)", err, key)
		return 0, err
	}
	return res, nil
}

// 获取今天最新登陆的MID
func (d *dao) GetCacheAntiAddictionMID(ctx context.Context, deviceToken string, day int64) (int64, error) {
	key := keyAntiAddictionMID(deviceToken, day)
	res, err := redis.Int64(d.redis.Do(ctx, "GET", key))
	if err != nil {
		if err == redis.ErrNil {
			return 0, nil
		}
		log.Error("dao.GetCacheAntiAddictionMID redis get error,err:(%v),key:(%s)", err, key)
		return 0, err
	}
	return res, nil
}

// 设置当天青少年模式解锁错误次数
func (d *dao) SetCacheTeenagersUnlockErrorNum(ctx context.Context, deviceToken string, mid, day int64, pwdFrom int32) (int64, error) {
	key := keyTeenagersUnlockErrorNumByDevice(deviceToken, pwdFrom, day)
	if mid > 0 {
		key = keyTeenagersUnlockErrorNumByMid(mid, pwdFrom, day)
	}
	p := d.redis.Pipeline()
	p.Send("INCR", key)
	p.Send("EXPIRE", key, d.unlockErrorExpire)
	replies, err := p.Exec(ctx)
	if err != nil {
		log.Error("d.SetCacheTeenagersUnlockErrorNum err:(%v), key:(%s)", err, key)
		return 0, err
	}
	res, err := redis.Int64(replies.Scan())
	if err != nil {
		log.Error("d.SetCacheTeenagersUnlockErrorNum err:(%v), key:(%s)", err, key)
		return 0, err
	}
	return res, nil
}

// 获取当天青少年模式解锁错误次数
func (d *dao) GetCacheTeenagersUnlockErrorNum(ctx context.Context, deviceToken string, mid, day int64, pwdFrom int32) (int64, error) {
	key := keyTeenagersUnlockErrorNumByDevice(deviceToken, pwdFrom, day)
	if mid > 0 {
		key = keyTeenagersUnlockErrorNumByMid(mid, pwdFrom, day)
	}
	res, err := redis.Int64(d.redis.Do(ctx, "GET", key))
	if err != nil {
		if err == redis.ErrNil {
			return 0, nil
		}
		log.Error("d.GetCacheTeenagersUnlockErrorNum err:(%v),key:(%s)", err, key)
		return 0, err
	}
	return res, nil
}

//go:generate kratos tool redisgen
type _redis interface {
	// redis: -struct_name=dao -key=familyRelsOfParentKey
	CacheFamilyRelsOfParent(ctx context.Context, id int64) ([]*familymdl.FamilyRelation, error)
	// redis: -struct_name=dao -key=familyRelsOfParentKey -expire=d.familyRelsExpire
	AddCacheFamilyRelsOfParent(ctx context.Context, id int64, val []*familymdl.FamilyRelation) error
	// redis: -struct_name=dao -key=familyRelsOfParentKey
	DelCacheFamilyRelsOfParent(ctx context.Context, id int64) error
	// redis: -struct_name=dao -key=familyRelsOfChildKey
	CacheFamilyRelsOfChild(ctx context.Context, id int64) (*familymdl.FamilyRelation, error)
	// redis: -struct_name=dao -key=familyRelsOfChildKey -expire=d.familyRelsExpire
	AddCacheFamilyRelsOfChild(ctx context.Context, id int64, val *familymdl.FamilyRelation) error
	// redis: -struct_name=dao -key=familyRelsOfChildKey
	DelCacheFamilyRelsOfChild(ctx context.Context, id int64) error
	// redis: -struct_name=dao -key=sleepRemindKey
	CacheSleepRemind(ctx context.Context, mid int64) (*antiadmdl.SleepRemind, error)
	// redis: -struct_name=dao -key=sleepRemindKey -expire=d.sleepRemindExpire
	AddCacheSleepRemind(ctx context.Context, mid int64, val *antiadmdl.SleepRemind) error
	// redis: -struct_name=dao -key=sleepRemindKey
	DelCacheSleepRemind(ctx context.Context, mid int64) error
}
