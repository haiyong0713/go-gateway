package native

import (
	"go-common/library/cache/credis"
	"time"

	"go-common/library/database/sql"
	xsql "go-common/library/database/sql"
	"go-common/library/queue/databus"
	"go-common/library/sync/pipeline/fanout"

	"go-gateway/app/web-svr/native-page/interface/conf"
)

// Dao dao.
type Dao struct {
	db                *xsql.DB
	redis             credis.Redis
	mcRegularExpire   int32
	nativePub         *databus.Databus
	cache             *fanout.Fanout
	openDynamic       bool
	wlByMidExpire     int64
	wlByMidNullExpire int64
}

// New dao new.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		cache:             fanout.New("cache"),
		db:                sql.NewMySQL(c.MySQL.Like),
		redis:             credis.NewRedis(c.Redis.Config),
		nativePub:         databus.New(c.DataBus.NativePub),
		openDynamic:       c.Rule.OpenDynamic,
		mcRegularExpire:   int32(time.Duration(c.Rule.RegularExpire) / time.Second),
		wlByMidExpire:     c.NativePage.WhiteListByMidExpire,
		wlByMidNullExpire: c.NativePage.WhiteListByMidNullExpire,
	}
	return
}

// Close Dao
func (d *Dao) Close() {
	if d.db != nil {
		d.db.Close()
	}
}
