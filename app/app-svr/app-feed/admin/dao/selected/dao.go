package selected

import (
	"go-common/library/cache/redis"
	"go-common/library/database/orm"
	"go-common/library/net/http/blademaster"
	"go-common/library/queue/databus"
	"go-gateway/app/app-svr/app-feed/admin/conf"

	showGRPC "go-gateway/app/app-svr/app-show/interface/api"

	"github.com/jinzhu/gorm"
)

// Dao struct user of color egg Dao.
type Dao struct {
	// db
	DB              *gorm.DB
	TagDB           *gorm.DB
	HttpClient      *blademaster.Client
	Config          *conf.Config
	archiveHonorPub *databus.Databus
	ottSeriesPub    *databus.Databus
	selRedis        *redis.Pool
	showGrpc        showGRPC.AppShowClient // showGrpc 上海机房缓存
	showGrpcSH004   showGRPC.AppShowClient // showGrpcSH004 嘉定机房缓存
}

// New create a instance of color egg Dao and return.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		// db
		DB:              orm.NewMySQL(c.ORM),
		TagDB:           orm.NewMySQL(c.ORMTag),
		HttpClient:      blademaster.NewClient(c.HTTPClient.Read),
		Config:          c,
		archiveHonorPub: databus.New(c.ArchiveHonorDatabus),
		ottSeriesPub:    databus.New(c.OTTSeriesDatabus),
		selRedis:        redis.NewPool(c.SelectedRedis.Config),
	}
	var err error
	if d.showGrpc, err = showGRPC.NewClient(nil); err != nil {
		panic(err)
	}
	if d.showGrpcSH004, err = showGRPC.NewClient(c.ShowGrpcSH004); err != nil {
		panic(err)
	}
	d.initORM()
	return
}

func (d *Dao) initORM() {
	d.DB.LogMode(true)
	d.TagDB.LogMode(true)
}
