package ugctab

import (
	"go-common/library/database/sql"
	"go-gateway/app/app-svr/resource/service/conf"
	"go-gateway/app/app-svr/resource/service/model"
)

type Dao struct {
	db       *sql.DB
	c        *conf.Config
	tabCache []*model.UgcTabItem
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:  c,
		db: sql.NewMySQL(c.DB.Show),
	}
	return
}

func (d *Dao) Close() {
	if d.db != nil {
		d.db.Close()
	}
}
