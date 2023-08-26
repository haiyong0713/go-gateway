package guess

import (
	"context"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	"go-common/library/sync/pipeline/fanout"
	"go-gateway/app/web-svr/activity/interface/component"
	"go-gateway/app/web-svr/activity/interface/conf"
)

// Dao dao struct.
type Dao struct {
	db          *sql.DB
	cache       *fanout.Fanout
	redis       *redis.Pool
	guessExpire int32
	percent     float32
	maxOdds     float32
}

// New init
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		db:          component.GlobalDB,
		cache:       fanout.New("cache"),
		redis:       redis.NewPool(c.Redis.Config),
		percent:     c.Rule.GuessPercent,
		maxOdds:     c.Rule.GuessMaxOdds,
		guessExpire: int32(time.Duration(c.Redis.GuessExpire) / time.Second),
	}
	return
}

// Close Dao
func (d *Dao) Close() {
	if d.redis != nil {
		d.redis.Close()
	}
}

// Ping Dao
func (d *Dao) Ping(c context.Context) error {
	return d.db.Ping(c)
}
