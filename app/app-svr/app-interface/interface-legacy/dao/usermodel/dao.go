package usermodel

import (
	"context"
	"time"

	"go-common/library/cache/redis"
	xsql "go-common/library/database/sql"
	"go-common/library/log"
	"go-common/library/net/netutil"
	"go-common/library/retry"
	"go-common/library/sync/pipeline/fanout"

	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"
	antiadmdl "go-gateway/app/app-svr/app-interface/interface-legacy/model/anti_addiction"
	familymdl "go-gateway/app/app-svr/app-interface/interface-legacy/model/family"
	usermodel "go-gateway/app/app-svr/app-interface/interface-legacy/model/usermodel"
)

// Dao is audit dao.
type dao struct {
	db                  *xsql.DB
	redis               *redis.Redis
	emptyCacheExpire    int64
	emptyCacheRand      int64
	cache               *fanout.Fanout
	antiAddictionExpire int64
	familyRelsExpire    int64
	sleepRemindExpire   int64
	unlockErrorExpire   int64
}

//go:generate kratos tool btsgen
type Dao interface {
	Close()
	// bts: -nullcache=[]*usermodel.User{{ID:-1}} -check_null_code=len($)==1&&$[0]!=nil&&$[0].ID==-1 -singleflight=true
	UserModels(ctx context.Context, mid int64, mobiApp string, deviceToken string) ([]*usermodel.User, error)
	AddUserModel(ctx context.Context, user *usermodel.User, sync bool) (userID int64, devID int64, err error)
	AddSpecialModeLog(ctx context.Context, log *usermodel.SpecialModeLog) error
	SetCacheAntiAddictionTime(ctx context.Context, deviceToken string, mid, day, useTime int64) error
	GetCacheAntiAddictionTime(ctx context.Context, deviceToken string, mid, day int64) (int64, error)
	GetCacheAntiAddictionMID(ctx context.Context, deviceToken string, day int64) (int64, error)
	AddManualForceLog(ctx context.Context, log *usermodel.ManualForceLog) error
	UpdateOperation(ctx context.Context, id, mid int64, op int) error
	UpdateManualForceAndCache(ctx context.Context, user *usermodel.User) error
	// bts: -nullcache=[]*familymdl.FamilyRelation{{ID:-1}} -check_null_code=len($)==1&&$[0]!=nil&&$[0].ID==-1
	FamilyRelsOfParent(ctx context.Context, parentMid int64) ([]*familymdl.FamilyRelation, error)
	// bts: -nullcache=&familymdl.FamilyRelation{ID:-1} -check_null_code=$!=nil&&$.ID==-1
	FamilyRelsOfChild(ctx context.Context, childMid int64) (*familymdl.FamilyRelation, error)
	UnbindFamily(ctx context.Context, rel *familymdl.FamilyRelation) error
	BindFamily(ctx context.Context, pmid, cmid, duration int64) error
	LatestFamilyRel(ctx context.Context, pmid int64, cmid int64) (*familymdl.FamilyRelation, error)
	AddFamilyLogs(ctx context.Context, items []*familymdl.FamilyLog) error
	UpdateTimelock(ctx context.Context, rel *familymdl.FamilyRelation) error
	// bts: -nullcache=&antiadmdl.SleepRemind{ID:-1} -check_null_code=$!=nil&&$.ID==-1
	SleepRemind(ctx context.Context, mid int64) (*antiadmdl.SleepRemind, error)
	AddSleepRemind(ctx context.Context, val *antiadmdl.SleepRemind) error
	UpdateSleepRemind(ctx context.Context, val *antiadmdl.SleepRemind) error
	GetTeenagerModelPWD(ctx context.Context, wsxcde string) (string, error)
	SetCacheTeenagersUnlockErrorNum(ctx context.Context, deviceToken string, mid, day int64, pwdFrom int32) (int64, error)
	GetCacheTeenagersUnlockErrorNum(ctx context.Context, deviceToken string, mid, day int64, pwdFrom int32) (int64, error)
}

// New new a audit dao.
func New(c *conf.Config) Dao {
	return &dao{
		db:                  xsql.NewMySQL(c.MySQL.Show),
		redis:               redis.NewRedis(c.Redis.Interface.Config),
		emptyCacheExpire:    int64(time.Duration(c.Redis.Interface.EmptyCacheExpire) / time.Second),
		emptyCacheRand:      int64(time.Duration(c.Redis.Interface.EmptyCacheRand) / time.Second),
		cache:               fanout.New("cache"),
		antiAddictionExpire: int64(time.Duration(c.Redis.Interface.AntiAddictionExpire) / time.Second),
		familyRelsExpire:    int64(time.Duration(c.Redis.Interface.FamilyRelsExpire) / time.Second),
		sleepRemindExpire:   int64(time.Duration(c.Redis.Interface.SleepRemindExpire) / time.Second),
		unlockErrorExpire:   int64(time.Duration(c.Redis.Interface.UnlockErrorExpire) / time.Second),
	}
}

func (d *dao) AddUserModel(ctx context.Context, user *usermodel.User, sync bool) (userID int64, devID int64, err error) {
	if sync {
		userID, devID, err = d.addSyncUserModel(ctx, user)
		if err != nil {
			return 0, 0, err
		}
		if err := d.delCacheSyncUserModel(ctx, user.Mid, user.MobiApp, user.DeviceToken); err != nil {
			log.Error("日志告警 删除青少年缓存错误:%+v", err)
			d.cache.Do(ctx, func(ctx context.Context) {
				if err := retry.WithAttempts(ctx, "del_cache_user_state", 5, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
					return d.delCacheSyncUserModel(ctx, user.Mid, user.MobiApp, user.DeviceToken)
				}); err != nil {
					log.Error("日志告警 删除青少年缓存错误:%+v", err)
				}
			})
		}
		return userID, devID, nil
	}
	userID, devID, err = d.addUserModel(ctx, user)
	if err != nil {
		return 0, 0, err
	}
	d.attemptDelCacheUserModel(ctx, user)
	return userID, devID, nil
}

func (d *dao) attemptDelCacheUserModel(ctx context.Context, user *usermodel.User) {
	err := d.delCacheUserModel(ctx, user.Mid, user.MobiApp, user.DeviceToken)
	if err == nil {
		return
	}
	log.Error("日志告警 删除青少年缓存错误:%+v", err)
	_ = d.cache.Do(ctx, func(ctx context.Context) {
		if err := retry.WithAttempts(ctx, "del_cache_user_state", 5, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
			return d.delCacheUserModel(ctx, user.Mid, user.MobiApp, user.DeviceToken)
		}); err != nil {
			log.Error("日志告警 删除青少年缓存错误:%+v", err)
		}
	})
}

func (d *dao) UpdateManualForceAndCache(ctx context.Context, user *usermodel.User) error {
	if err := d.updateManualForce(ctx, user.ID, user.ManualForce, user.MfTime, user.MfOperator); err != nil {
		return err
	}
	d.attemptDelCacheUserModel(ctx, &usermodel.User{Mid: user.Mid})
	return nil
}

func (d *dao) UpdateOperation(ctx context.Context, id, mid int64, op int) error {
	if err := d.updateOperationDB(ctx, id, op); err != nil {
		return err
	}
	d.attemptDelCacheUserModel(ctx, &usermodel.User{Mid: mid})
	return nil
}

func (d *dao) UnbindFamily(ctx context.Context, rel *familymdl.FamilyRelation) error {
	if rel == nil {
		return nil
	}
	if err := d.unbindFamily(ctx, rel.ID); err != nil {
		return err
	}
	_ = d.cache.Do(ctx, func(ctx context.Context) {
		if err := retry.WithAttempts(ctx, "del_parent_cache_from_family_unbind", 5, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
			return d.DelCacheFamilyRelsOfParent(ctx, rel.ParentMid)
		}); err != nil {
			log.Error("日志告警 删除parent亲子关系缓存失败, relation=%+v error=%+v", rel, err)
		}
		if err := retry.WithAttempts(ctx, "del_child_cache_from_family_unbind", 5, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
			return d.DelCacheFamilyRelsOfChild(ctx, rel.ChildMid)
		}); err != nil {
			log.Error("日志告警 删除child亲子关系缓存失败, relation=%+v error=%+v", rel, err)
		}
	})
	return nil
}

func (d *dao) BindFamily(ctx context.Context, pmid, cmid, duration int64) error {
	if err := d.bindFamily(ctx, pmid, cmid, duration); err != nil {
		return err
	}
	_ = d.cache.Do(ctx, func(ctx context.Context) {
		if err := retry.WithAttempts(ctx, "del_parent_cache_from_family_bind", 5, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
			return d.DelCacheFamilyRelsOfParent(ctx, pmid)
		}); err != nil {
			log.Error("日志告警 删除parent亲子关系缓存失败, pmid=%+v error=%+v", pmid, err)
		}
		if err := retry.WithAttempts(ctx, "del_child_cache_from_family_bind", 5, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
			return d.DelCacheFamilyRelsOfChild(ctx, cmid)
		}); err != nil {
			log.Error("日志告警 删除child亲子关系缓存失败, cmid=%+v error=%+v", cmid, err)
		}
	})
	return nil
}

func (d *dao) AddSleepRemind(ctx context.Context, val *antiadmdl.SleepRemind) error {
	if err := d.addSleepRemind(ctx, val); err != nil {
		return err
	}
	_ = d.cache.Do(ctx, func(ctx context.Context) {
		if err := retry.WithAttempts(ctx, "del_sleep_remind_cache_of_add", 5, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
			return d.DelCacheSleepRemind(ctx, val.Mid)
		}); err != nil {
			log.Error("日志告警 删除睡眠提醒缓存失败, mid=%+v error=%+v", val.Mid, err)
		}
	})
	return nil
}

func (d *dao) UpdateSleepRemind(ctx context.Context, val *antiadmdl.SleepRemind) error {
	if err := d.updateSleepRemind(ctx, val); err != nil {
		return err
	}
	_ = d.cache.Do(ctx, func(ctx context.Context) {
		if err := retry.WithAttempts(ctx, "del_sleep_remind_cache_of_update", 5, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
			return d.DelCacheSleepRemind(ctx, val.Mid)
		}); err != nil {
			log.Error("日志告警 删除睡眠提醒缓存失败, mid=%+v error=%+v", val.Mid, err)
		}
	})
	return nil
}

func (d *dao) UpdateTimelock(ctx context.Context, rel *familymdl.FamilyRelation) error {
	if rel == nil {
		return nil
	}
	if err := d.updateTimelock(ctx, rel.ID, rel.TimelockState, rel.DailyDuration); err != nil {
		return err
	}
	_ = d.cache.Do(ctx, func(ctx context.Context) {
		if err := retry.WithAttempts(ctx, "del_parent_cache_from_update_timelock", 5, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
			return d.DelCacheFamilyRelsOfParent(ctx, rel.ParentMid)
		}); err != nil {
			log.Error("日志告警 删除parent亲子关系缓存失败, relation=%+v error=%+v", rel, err)
		}
		if err := retry.WithAttempts(ctx, "del_child_cache_from_update_timelock", 5, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
			return d.DelCacheFamilyRelsOfChild(ctx, rel.ChildMid)
		}); err != nil {
			log.Error("日志告警 删除child亲子关系缓存失败, relation=%+v error=%+v", rel, err)
		}
	})
	return nil
}

func (d *dao) Close() {
	if d.cache != nil {
		d.cache.Close()
	}
	if d.db != nil {
		d.db.Close()
	}
}
