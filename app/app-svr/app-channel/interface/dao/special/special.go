package special

import (
	"context"

	"go-common/library/database/sql"
	"go-gateway/app/app-svr/app-card/interface/model/card/operate"
	"go-gateway/app/app-svr/app-channel/interface/conf"
)

const (
	_getSQL = "SELECT `id`,`title`,`desc`,`cover`,`scover`,`re_type`,`re_value`,`corner`,`size`,`card` FROM special_card"
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

func (d *Dao) Card(c context.Context) (scm map[int64]*operate.Special, err error) {
	rows, err := d.db.Query(c, _getSQL)
	if err != nil {
		return
	}
	defer rows.Close()
	scm = map[int64]*operate.Special{}
	for rows.Next() {
		sc := &operate.Special{}
		var cardType int
		if err = rows.Scan(&sc.ID, &sc.Title, &sc.Desc, &sc.Cover, &sc.SingleCover, &sc.ReType, &sc.ReValue, &sc.Badge, &sc.Size, &cardType); err != nil {
			return
		}
		sc.Change()
		// nolint:gomnd
		switch cardType {
		case 3:
			if sc.Desc == "" {
				sc.Desc = "立即查看"
			}
		}
		scm[sc.ID] = sc
	}
	err = rows.Err()
	return scm, err
}

// Close close memcache resource.
func (d *Dao) Close() {
	if d.db != nil {
		d.db.Close()
	}
}
