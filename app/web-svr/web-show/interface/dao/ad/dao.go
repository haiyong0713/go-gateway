package ad

import (
	"context"

	xsql "go-common/library/database/sql"
	"go-common/library/log"
	"go-common/library/naming/discovery"
	httpx "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/resolver"
	"go-common/library/stat/prom"
	"go-gateway/app/web-svr/web-show/interface/conf"
)

// Dao define db struct
type Dao struct {
	cpt *xsql.DB
	// sql
	selAdsStmt *xsql.Stmt
	// cpt
	httpClient *httpx.Client
	cpmURL     string
}

const (
	_cpmURL = "/api/bce/pc"
)

// PromError err
func PromError(name string, format string, args ...interface{}) {
	prom.BusinessErrCount.Incr(name)
	log.Error(format, args...)
}

// New init mysql db
func New(c *conf.Config) (dao *Dao) {
	dao = &Dao{
		cpt:        xsql.NewMySQL(c.MySQL.Cpt),
		httpClient: httpx.NewClient(c.HTTPClient, httpx.SetResolver(resolver.New(nil, discovery.Builder()))),
		cpmURL:     c.Host.AdDiscovery + _cpmURL,
	}
	dao.selAdsStmt = dao.cpt.Prepared(_selAds)
	return
}

// Close close the resource.
func (dao *Dao) Close() {
	dao.cpt.Close()
}

// Ping ping mysql
func (dao *Dao) Ping(c context.Context) error {
	return dao.cpt.Ping(c)
}
