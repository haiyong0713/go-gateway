package fit

import (
	"context"
	"fmt"
	"go-gateway/app/web-svr/activity/job/model/fit"
	"strings"

	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	xsql "go-common/library/database/sql"
	"go-gateway/app/web-svr/activity/job/conf"
	favgrpc "go-main/app/community/favorite/service/api"
)

// 生成 mock 代码
//go:generate mockgen -source dao.go  -destination dao.mock.go -package college

const (
	prefix    = "fit_activity"
	separator = ":"
)

// Dao dao interface
type Dao interface {
	Close()
	GetPlanList(ctx context.Context, offset, limit int) (res []*fit.PlanRecordRes, err error)
	UpdateOnePlanById(c context.Context, planId int64, views int32, danmaku int32) (affected int64, err error)
	Folders(c context.Context, folderIDs []int64, typ int32) (*favgrpc.FoldersReply, error)
	FavoritesAll(c context.Context, tp int32, mid, uid, fid int64, pn, ps int32) (fav *favgrpc.FavoritesReply, err error)
	Ping(c context.Context) error
}

// Dao dao.
type dao struct {
	c         *conf.Config
	redis     *redis.Pool
	db        *xsql.DB
	FavClient favgrpc.FavoriteClient
}

// New init
func newDao(c *conf.Config) (newdao Dao) {
	fclient, err := favgrpc.New(c.FavoriteClient)
	if err != nil {
		panic(err)
	}
	newdao = &dao{
		c:         c,
		db:        sql.NewMySQL(c.MySQL.Like),
		redis:     redis.NewPool(c.Redis.Config),
		FavClient: fclient,
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

// Ping ping
func (d *dao) Ping(c context.Context) error {
	return d.db.Ping(c)
}

// buildKey build key
func buildKey(args ...interface{}) string {
	strArgs := make([]string, len(args), len(args))
	for i, val := range args {
		strArgs[i] = fmt.Sprint(val)
	}
	return prefix + separator + strings.Join(strArgs, separator)
}
