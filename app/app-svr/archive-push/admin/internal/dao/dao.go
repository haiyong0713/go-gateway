package dao

import (
	"context"
	"net/http"

	accountGRPC "git.bilibili.co/bapis/bapis-go/account/service"
	activityGRPC "git.bilibili.co/bapis/bapis-go/activity/service"
	archiveGRPC "git.bilibili.co/bapis/bapis-go/archive/service"
	tagGRPC "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	"github.com/google/wire"
	"go-common/library/cache/redis"
	"go-common/library/conf/paladin"
	"go-common/library/database/gorm"
	"go-common/library/database/sql"
	bm "go-common/library/net/http/blademaster"

	"go-gateway/app/app-svr/archive-push/admin/internal/model"
)

var Provider = wire.NewSet(New, NewDB, NewRedis, NewORM, NewBMClient, NewHTTPClient, NewArchiveGRPC, NewTagGRPC, NewAccountGRPC, NewActivityGRPC)

// dao dao.
type Dao struct {
	DB                 *sql.DB
	ORM                *gorm.DB
	redis              *redis.Redis
	bmClient           *bm.Client
	httpClient         *http.Client
	hosts              *model.Hosts
	archiveGRPCClient  archiveGRPC.ArchiveClient
	tagGRPCClient      tagGRPC.TagRPCClient
	accountGRPCClient  accountGRPC.AccountClient
	activityGRPCClient activityGRPC.ActivityClient
}

// New new a dao and return.
func New(r *redis.Redis, db *sql.DB, orm *gorm.DB, bmClient *bm.Client, httpClient *http.Client, archiveGRPCClient archiveGRPC.ArchiveClient, tagGRPCClient tagGRPC.TagRPCClient, accountGRPCClient accountGRPC.AccountClient, activityGRPCClient activityGRPC.ActivityClient) (d *Dao, cf func(), err error) {
	cf = func() { db.Close() }
	d = &Dao{
		DB:                 db,
		ORM:                orm,
		redis:              r,
		bmClient:           bmClient,
		httpClient:         httpClient,
		hosts:              &model.Hosts{},
		archiveGRPCClient:  archiveGRPCClient,
		tagGRPCClient:      tagGRPCClient,
		accountGRPCClient:  accountGRPCClient,
		activityGRPCClient: activityGRPCClient,
	}
	var ct paladin.TOML
	if err = paladin.Get("http.toml").Unmarshal(&ct); err != nil {
		panic(err)
	}
	if err = ct.Get("Hosts").UnmarshalTOML(d.hosts); err != nil {
		panic(err)
	}
	cf = d.Close
	return
}

// Close close the resource.
func (d *Dao) Close() {
	d.DB.Close()
}

// Ping ping the resource.
func (d *Dao) Ping(ctx context.Context) (err error) {
	return d.DB.Ping(ctx)
}
