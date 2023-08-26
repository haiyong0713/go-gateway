package question

import (
	"time"

	"go-common/library/cache/memcache"
	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	"go-common/library/sync/pipeline/fanout"
	"go-gateway/app/web-svr/activity/interface/component"
	"go-gateway/app/web-svr/activity/interface/conf"
)

// Dao dao struct.
type Dao struct {
	db             *sql.DB
	mc             *memcache.Memcache
	questionExpire int32
	lastLogExpire  int32
	redis          *redis.Pool
	answerExpire   int32
	cache          *fanout.Fanout
}

// New new dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		db:    component.GlobalDB,
		mc:    memcache.New(c.Memcache.Like),
		redis: redis.NewPool(c.Redis.Config),
		cache: fanout.New("cache"),
	}
	d.questionExpire = int32(time.Duration(c.Memcache.QuestionExpire) / time.Second)
	d.lastLogExpire = int32(time.Duration(c.Memcache.LastLogExpire) / time.Second)
	d.answerExpire = int32(time.Duration(c.Redis.AnswerExpire) / time.Second)
	return d
}

// Close Dao
func (d *Dao) Close() {
	if d.redis != nil {
		d.redis.Close()
	}
	if d.mc != nil {
		d.mc.Close()
	}
}
