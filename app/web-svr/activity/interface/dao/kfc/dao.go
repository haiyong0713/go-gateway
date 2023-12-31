package kfc

import (
	"time"

	"go-common/library/cache/memcache"
	xsql "go-common/library/database/sql"
	"go-common/library/log"
	httpx "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/interface/conf"
)

const (
	_kfcWinnerURI = "/gift/v4/Smalltv/getKfcWinnerById"
)

// PromError stat and log.
func PromError(name string, format string, args ...interface{}) {
	log.Error(format, args...)
}

// Dao dao.
type Dao struct {
	mc              *memcache.Memcache
	db              *xsql.DB
	client          *httpx.Client
	mcKfcExpire     int32
	mcKfcCodeExpire int32
	kfcWinnerURL    string
}

// New dao new.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		mc:              memcache.New(c.Memcache.Like),
		db:              xsql.NewMySQL(c.MySQL.Like),
		client:          httpx.NewClient(c.HTTPClientKfc),
		mcKfcExpire:     int32(time.Duration(c.Memcache.KfcExpire) / time.Second),
		mcKfcCodeExpire: int32(time.Duration(c.Memcache.KfcCodeExpire) / time.Second),
		kfcWinnerURL:    c.Host.LiveCo + _kfcWinnerURI,
	}
	return
}

// Close Dao
func (d *Dao) Close() {
	if d.db != nil {
		d.db.Close()
	}
	if d.mc != nil {
		d.mc.Close()
	}
}
