package converge

import (
	"context"

	"go-common/library/database/sql"
	"go-gateway/app/app-svr/app-card/interface/model/card/operate"
	"go-gateway/app/app-svr/app-channel/interface/conf"
)

const (
	_getSQL = "SELECT id,re_type,re_value,title,cover,content FROM content_card"
)

type Dao struct {
	db *sql.DB
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		db: sql.NewMySQL(c.MySQL.Manager),
	}
	return
}

func (d *Dao) Cards(c context.Context) (cm map[int64]*operate.Converge, err error) {
	rows, err := d.db.Query(c, _getSQL)
	if err != nil {
		return
	}
	defer rows.Close()
	cm = map[int64]*operate.Converge{}
	for rows.Next() {
		c := &operate.Converge{}
		if err = rows.Scan(&c.ID, &c.ReType, &c.ReValue, &c.Title, &c.Cover, &c.Content); err != nil {
			return
		}
		if c.Title == "" {
			continue
		}
		c.Change()
		cm[c.ID] = c
	}
	err = rows.Err()
	return
}

// Close close memcache resource.
func (dao *Dao) Close() {
	if dao.db != nil {
		dao.db.Close()
	}
}
