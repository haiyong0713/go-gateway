package note

import (
	"github.com/go-resty/resty/v2"
	"go-common/library/cache/redis"
	xsql "go-common/library/database/sql"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/sync/pipeline/fanout"
	xtime "go-common/library/time"
	"go-gateway/app/app-svr/hkt-note/job/conf"
	"time"
)

// Dao is archive dao.
type Dao struct {
	c           *conf.Config
	db          *xsql.DB
	grpc        *grpc
	redis       *redis.Pool
	noteExpire  int
	cache       *fanout.Fanout
	client      *bm.Client
	restyClient *resty.Client
}

// New new a archive dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:           c,
		grpc:        NewGrpc(),
		client:      bm.NewClient(c.HTTPClient),
		db:          xsql.NewMySQL(c.DB.Note),
		redis:       redis.NewPool(c.Redis.Config),
		noteExpire:  int(time.Duration(c.Redis.NoteExpire) / time.Second),
		cache:       fanout.New("cache", fanout.Worker(10), fanout.Buffer(10240)),
		restyClient: resty.New(),
	}
	if c.Redis.BotPushExpire == 0 {
		c.Redis.BotPushExpire = xtime.Duration(10 * 24 * time.Hour)
	}
	return
}
