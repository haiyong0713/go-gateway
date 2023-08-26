package entry

import (
	"go-common/library/database/sql"
	"go-gateway/app/app-svr/resource/service/conf"
)

// Dao struct user of color entry Dao.
type Dao struct {
	db *sql.DB
	c  *conf.Config
}

// New create a instance of color entry Dao and return.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:  c,
		db: sql.NewMySQL(c.DB.Show),
	}
	return
}

// Close close db resource.
func (d *Dao) Close() {
	if d.db != nil {
		d.db.Close()
	}
}
