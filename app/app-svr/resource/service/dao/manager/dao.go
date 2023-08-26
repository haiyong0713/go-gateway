package manager

import (
	"context"
	"go-common/library/cache/credis"

	"go-common/library/database/sql"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/resource/service/conf"

	opIcon "git.bilibili.co/bapis/bapis-go/manager/operation/icon"
)

// Dao manager dao
type Dao struct {
	db           *sql.DB
	c            *conf.Config
	httpClient   *bm.Client
	opIconClient opIcon.OperationItemIconV1Client
	redis        credis.Redis
}

// New new manager dao
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:          c,
		db:         sql.NewMySQL(c.DB.Manager),
		httpClient: bm.NewClient(c.HTTPClient),
		redis:      credis.NewRedis(c.Redis.Comm),
	}
	var err error
	if d.opIconClient, err = opIcon.NewClientOperationItemIconV1(c.OpIconGRPC); err != nil {
		panic(err)
	}
	return
}

// Close close db resource.
func (d *Dao) Close() {
	if d.db != nil {
		d.db.Close()
		d.redis.Close()
	}
}

func (d *Dao) OpIconList(ctx context.Context) (*opIcon.ListResp, error) {
	return d.opIconClient.List(ctx, &opIcon.ListReq{})
}
