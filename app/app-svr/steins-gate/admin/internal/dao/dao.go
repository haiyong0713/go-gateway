package dao

import (
	"go-common/library/database/sql"
	bm "go-common/library/net/http/blademaster"

	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/app-svr/steins-gate/admin/conf"
)

// Dao dao.
type Dao struct {
	c               *conf.Config
	db              *sql.DB
	videoupURL      string
	client          *bm.Client
	httpVideoClient *bm.Client
	bvcDimensionURL string
	arcClient       arcgrpc.ArchiveClient
}

// New new a dao and return.
func New(c *conf.Config) (dao *Dao) {
	dao = &Dao{
		c: c,
		// mysql
		db: sql.NewMySQL(c.MySQL.Steinsgate),
		// http client
		client: bm.NewClient(c.HTTPClient),
		// video_up url
		videoupURL: c.Host.Videoup + _videoUpViewURI,

		httpVideoClient: bm.NewClient(c.VideoClient),
		bvcDimensionURL: c.Host.Bvc + _dimensionURI,
	}
	var err error
	if dao.arcClient, err = arcgrpc.NewClient(c.Archive); err != nil {
		panic(err)
	}
	return
}

// Close close the resource.
func (d *Dao) Close() {

}
