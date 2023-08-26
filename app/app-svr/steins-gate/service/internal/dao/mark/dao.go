package mark

import (
	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	"go-common/library/sync/pipeline/fanout"

	"go-gateway/app/app-svr/steins-gate/service/conf"
)

// Dao dao.
type Dao struct {
	c      *conf.Config
	rds    *redis.Pool
	db     *sql.DB
	fanout *fanout.Fanout
}

// New new a dao and return.
func New(c *conf.Config) (dao *Dao) {
	dao = &Dao{
		c:      c,
		rds:    redis.NewPool(c.Redis.Graph),
		db:     sql.NewMySQL(c.MySQL.Steinsgate),
		fanout: fanout.New("cache"),
	}
	return

}
