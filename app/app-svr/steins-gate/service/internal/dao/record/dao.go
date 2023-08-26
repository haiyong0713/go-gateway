package record

import (
	"time"

	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	"go-common/library/sync/pipeline/fanout"

	"go-gateway/app/app-svr/steins-gate/service/conf"
)

// Dao dao.
type Dao struct {
	c            *conf.Config
	db           *sql.DB
	rds          *redis.Pool
	cache        *fanout.Fanout
	recordExpire int32
}

// New new a dao and return.
func New(c *conf.Config) (dao *Dao) {
	dao = &Dao{
		c:            c,
		db:           sql.NewMySQL(c.MySQL.Steinsgate),
		rds:          redis.NewPool(c.Redis.Graph),
		cache:        fanout.New("graph_mc", fanout.Worker(1), fanout.Buffer(10240)),
		recordExpire: int32(time.Duration(c.Redis.RecordExpiration) / time.Second),
	}
	return

}
