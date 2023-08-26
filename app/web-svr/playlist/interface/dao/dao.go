package dao

import (
	"context"
	"fmt"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/queue/databus"
	"go-gateway/app/web-svr/playlist/interface/conf"
)

const _replyURL = "/x/internal/v2/reply/subject/regist"

// Dao dao struct.
type Dao struct {
	// config
	c *conf.Config
	// db
	db *sql.DB
	// redis
	redis      *redis.Pool
	statExpire int32
	plExpire   int32
	// http client
	http *bm.Client
	// stmt
	videosStmt map[string]*sql.Stmt
	// databus
	viewDbus  *databus.Databus
	shareDbus *databus.Databus
	// search video URL
	searchURL string
	replyURL  string
}

// New new dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		// config
		c:          c,
		db:         sql.NewMySQL(c.Mysql),
		redis:      redis.NewPool(c.Redis.Config),
		statExpire: int32(time.Duration(c.Redis.StatExpire) / time.Second),
		plExpire:   int32(time.Duration(c.Redis.PlExpire) / time.Second),
		http:       bm.NewClient(c.HTTPClient),
		viewDbus:   databus.New(c.ViewDatabus),
		shareDbus:  databus.New(c.ShareDatabus),
		searchURL:  c.Host.Search + _searchURL,
		replyURL:   c.Host.ReplyURL + _replyURL,
	}
	d.videosStmt = make(map[string]*sql.Stmt, _plArcSub)
	for i := 0; i < _plArcSub; i++ {
		key := fmt.Sprintf("%02d", i)
		d.videosStmt[key] = d.db.Prepared(fmt.Sprintf(_plArcsSQL, key))
	}
	return
}

// Ping ping dao
func (d *Dao) Ping(c context.Context) (err error) {
	if err = d.db.Ping(c); err != nil {
		return
	}
	err = d.pingRedis(c)
	return
}

func (d *Dao) pingRedis(c context.Context) (err error) {
	conn := d.redis.Get(c)
	_, err = conn.Do("SET", "PING", "PONG")
	conn.Close()
	return
}
