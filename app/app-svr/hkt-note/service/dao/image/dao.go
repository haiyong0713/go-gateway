package image

import (
	"time"

	"go-common/library/cache/redis"
	xsql "go-common/library/database/sql"
	"go-common/library/sync/pipeline/fanout"
	"go-gateway/app/app-svr/hkt-note/service/conf"
)

type Dao struct {
	c         *conf.Config
	dbw       *xsql.DB
	dbr       *xsql.DB
	redis     *redis.Pool
	cache     *fanout.Fanout
	imgExpire int
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:         c,
		dbw:       xsql.NewMySQL(c.DB.NoteWrite),
		dbr:       xsql.NewMySQL(c.DB.NoteRead),
		redis:     redis.NewPool(c.Redis.Config),
		cache:     fanout.New("cache", fanout.Worker(10), fanout.Buffer(10240)),
		imgExpire: int(time.Duration(c.Redis.ImgExpire) / time.Second),
	}
	return
}
