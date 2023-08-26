package system

import (
	xsql "go-common/library/database/sql"
	"go-gateway/app/web-svr/activity/interface/component"
	"go-gateway/app/web-svr/activity/interface/conf"
	model "go-gateway/app/web-svr/activity/interface/model/system"
)

type Dao struct {
	c              *conf.Config
	db             *xsql.DB
	MapKeyWorkCode map[string]*model.User
	MapKeyOAID     map[int64]*model.User
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:              c,
		db:             component.GlobalDB,
		MapKeyWorkCode: make(map[string]*model.User),
		MapKeyOAID:     make(map[int64]*model.User),
	}
	return
}
