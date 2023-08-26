package audit

import (
	"go-common/library/database/sql"
	"go-common/library/sync/pipeline/fanout"

	filtergrpc "git.bilibili.co/bapis/bapis-go/filter/service"

	"go-gateway/app/app-svr/steins-gate/service/conf"
)

// Dao dao.
type Dao struct {
	c            *conf.Config
	db           *sql.DB
	cache        *fanout.Fanout
	filterClient filtergrpc.FilterClient
}

// New new a dao and return.
func New(c *conf.Config) (dao *Dao) {
	dao = &Dao{
		c:     c,
		cache: fanout.New("cache"),
		db:    sql.NewMySQL(c.MySQL.Steinsgate),
	}
	var err error
	if dao.filterClient, err = filtergrpc.NewClient(c.Filter); err != nil {
		panic(err)
	}
	return

}
