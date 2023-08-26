package resource

import (
	"context"

	xsql "go-common/library/database/sql"
	"go-common/library/log"
	"go-common/library/stat/prom"

	resourcegrpc "go-gateway/app/app-svr/resource/service/api/v1"
	"go-gateway/app/web-svr/web-show/interface/conf"

	resv2grpc "git.bilibili.co/bapis/bapis-go/resource/service/v2"
)

// Dao struct
type Dao struct {
	db      *xsql.DB
	videodb *xsql.DB
	// ad_active
	selAllVdoActStmt   *xsql.Stmt
	selVdoActMTCntStmt *xsql.Stmt
	delAllVdoActStmt   *xsql.Stmt
	// ad
	selAdVdoActStmt   *xsql.Stmt
	selAdMtCntVdoStmt *xsql.Stmt
	// res
	selAllResStmt    *xsql.Stmt
	selAllAssignStmt *xsql.Stmt
	selDefBannerStmt *xsql.Stmt
	// resource clien
	ResourceClient resourcegrpc.ResourceClient
	resv2Client    resv2grpc.ResourceClient

	BanResGRPCToken string
}

// New init mysql db
func New(c *conf.Config) (dao *Dao) {
	dao = &Dao{
		db:      xsql.NewMySQL(c.MySQL.Res),
		videodb: xsql.NewMySQL(c.MySQL.Ads),
	}
	var err error
	if dao.ResourceClient, err = resourcegrpc.NewClient(c.ResourceGRPC); err != nil {
		panic(err)
	}
	if dao.resv2Client, err = resv2grpc.NewClient(c.Resourcev2GRPC); err != nil {
		panic(err)
	}
	dao.BanResGRPCToken = c.BanResGRPCToken
	dao.initActive()
	dao.initRes()
	dao.initAd()
	return
}

// Close close the resource.
func (dao *Dao) Close() {
	dao.db.Close()
}

// PromError err
func PromError(name string, format string, args ...interface{}) {
	prom.BusinessErrCount.Incr(name)
	log.Error(format, args...)
}

// Ping Dao
func (dao *Dao) Ping(c context.Context) (err error) {
	if err = dao.db.Ping(c); err != nil {
		log.Error("dao.db.Ping error(%v)", err)
		return
	}
	if err = dao.videodb.Ping(c); err != nil {
		log.Error("dao.videodb.Ping error(%v)", err)
	}
	return
}

// BeginTran Dao
func (dao *Dao) BeginTran(c context.Context) (tx *xsql.Tx, err error) {
	tx, err = dao.videodb.Begin(c)
	return
}
