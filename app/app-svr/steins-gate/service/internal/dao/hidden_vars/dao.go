package hidden_vars

import (
	"math/rand"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	"go-common/library/sync/pipeline/fanout"

	"go-gateway/app/app-svr/steins-gate/service/conf"
)

// Dao dao.
type Dao struct {
	c                  *conf.Config
	db                 *sql.DB
	rds                *redis.Pool
	cache              *fanout.Fanout
	hvarExpirationMinH int
	hvarExpirationMaxH int
	rand               *rand.Rand
}

// New new a dao and return.
func New(c *conf.Config) (dao *Dao) {
	dao = &Dao{
		c:  c,
		db: sql.NewMySQL(c.MySQL.Steinsgate),
		// redis
		rds:   redis.NewPool(c.Redis.Graph),
		cache: fanout.New("hidden_vars_cache"),
		// expire
		hvarExpirationMinH: int(time.Duration(c.Redis.HvarExpirationMinH) / time.Second),
		hvarExpirationMaxH: int(time.Duration(c.Redis.HvarExpirationMaxH) / time.Second),
		rand:               rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	return

}
