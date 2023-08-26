package invite

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go-common/library/cache/redis"
	xsql "go-common/library/database/sql"
	"go-gateway/app/web-svr/activity/interface/component"
	"go-gateway/app/web-svr/activity/interface/conf"
	mdl "go-gateway/app/web-svr/activity/interface/model/invite"
)

const (
	inviterKey = "inviter"
)

// 生成 mock 代码
//go:generate mockgen -source dao.go  -destination dao.mock.go -package invite

// Dao dao interface
type Dao interface {
	AddCacheToken(ctx context.Context, token string, data *mdl.FiToken) (err error)
	ClearUserShareLogCache(c context.Context, mid int64, activityUID string) (err error)
	UserShareLogCache(c context.Context, mid int64, activityUID string) (res *mdl.UserShareLog, err error)
	AddUserShareLogCache(ctx context.Context, mid int64, activityUID string, data *mdl.UserShareLog) (err error)
	AddOldUserCache(ctx context.Context, telHash string) (err error)
	GetMidBindInviter(c context.Context, telHash string, activityUID string) (res int64, err error)
	SetMidBindInviter(c context.Context, telHash string, inviter int64, activityUID string) (err error)
	GetInviteByTelHash(ctx context.Context, mid int64, activityUID string, telHash string) (res *mdl.InviteRelation, err error)
	CacheGetMidToken(c context.Context, mid, tp int64, activityUID string, source int64) (res string, err error)
	CacheMidToken(c context.Context, mid, tp int64, activityUID, token string, source int64) (err error)

	AddToken(ctx context.Context, mid, tp, expire int64, activityUID, token string, source int64) (int64, error)
	updateFirstShareTime(ctx context.Context, mid int64, activityUID string, now time.Time) (int64, error)
	updateLastShareTime(ctx context.Context, mid int64, activityUID string, now time.Time) (int64, error)
	AddUserShareLog(ctx context.Context, mid int64, activityUID string, now time.Time) error
	updateFirstShareExpire(ctx context.Context, mid int64, activityUID string, now time.Time) (int64, error)
	AddAllInviteLog(ctx context.Context, param *mdl.AllInviteLog) error
	UpdateIsShareExpire(ctx context.Context, mid int64, activityUID string) (int64, error)
	UserShareLog(ctx context.Context, mid int64, activityUID string) (*mdl.UserShareLog, error)
	CacheToken(ctx context.Context, token string) (res *mdl.FiToken, err error)
	GetInviteMidByTel(ctx context.Context, tel string) (res *mdl.InviteRelation, err error)
	SelToken(ctx context.Context, token string) (res *mdl.FiToken, err error)
	AddInviteRelationLog(ctx context.Context, inviteRel *mdl.InviteRelation) error
	AddInviteRelation(ctx context.Context, ir *mdl.InviteRelation) error
	SetInviteRelation(ctx context.Context, ir *mdl.InviteRelation) (int64, error)

	Close()
}

// Dao dao.
type dao struct {
	c                *conf.Config
	redis            *redis.Pool
	db               *xsql.DB
	firstShareExpire int64
	actExpire        int32
	bindExpire       int64
	tokenExpire      int64
}

// New init
func newDao(c *conf.Config) (newdao Dao) {
	newdao = &dao{
		c:                c,
		redis:            redis.NewPool(c.Redis.Config),
		db:               component.GlobalDB,
		actExpire:        int32(time.Duration(c.Redis.ActExpire) / time.Second),
		tokenExpire:      int64(time.Duration(c.Redis.InviteTokenExpire) / time.Second),
		firstShareExpire: int64(time.Duration(c.Redis.FirstShareExpire) / time.Second),
		bindExpire:       int64(time.Duration(c.Invite.BindExpire) / time.Second),
	}
	return
}

// New new a dao and return.
func New(c *conf.Config) (d Dao) {
	return newDao(c)
}

// Close Dao
func (d *dao) Close() {
	if d.redis != nil {
		d.redis.Close()
	}
}

// buildKey build key
func buildKey(args ...interface{}) string {
	strArgs := make([]string, len(args), len(args))
	for i, val := range args {
		strArgs[i] = fmt.Sprint(val)
	}
	return prefix + separator + strings.Join(strArgs, separator)
}
