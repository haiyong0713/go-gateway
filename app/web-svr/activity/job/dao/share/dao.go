package share

import (
	"context"

	"go-common/library/cache/redis"
	xsql "go-common/library/database/sql"
	"go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/job/conf"
	"go-gateway/app/web-svr/activity/job/model/share"
)

// 生成 mock 代码
//go:generate mockgen -source dao.go  -destination dao.mock.go -package share

// Dao dao interface
type Dao interface {
	Close()
	ShareURL(ctx context.Context, business string, token string, addLinks []string) (*share.Share, error)
	ShareRemoveURL(ctx context.Context, business string, token string, removeLinks []string) (*share.Share, error)
	Ping(c context.Context) error
}

// Dao dao.
type dao struct {
	c        *conf.Config
	redis    *redis.Pool
	db       *xsql.DB
	shareURL string
	client   *blademaster.Client
}

// New init
func newDao(c *conf.Config) (newdao Dao) {
	newdao = &dao{
		c:        c,
		shareURL: c.Share.ShareURL + shareURLURI,
		client:   blademaster.NewClient(c.HTTPClient),
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
