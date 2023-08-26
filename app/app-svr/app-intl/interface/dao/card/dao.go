package card

import (
	"context"

	"go-common/library/database/sql"
	"go-common/library/log"

	"go-gateway/app/app-svr/app-card/interface/model/card/operate"
	"go-gateway/app/app-svr/app-intl/interface/conf"
)

const (
	_followSQL = "SELECT `id`,`type`,`long_title`,`content` FROM `card_follow` WHERE `deleted`=0"
)

// Dao is dao
type Dao struct {
	db        *sql.DB
	followGet *sql.Stmt
}

// New new dao
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		db: sql.NewMySQL(c.MySQL.Show),
	}
	// prepare
	d.followGet = d.db.Prepared(_followSQL)
	return
}

// Follow is.
func (d *Dao) Follow(c context.Context) (cm map[int64]*operate.Follow, err error) {
	var rows *sql.Rows
	if rows, err = d.followGet.Query(c); err != nil {
		return
	}
	defer rows.Close()
	cm = make(map[int64]*operate.Follow)
	for rows.Next() {
		c := &operate.Follow{}
		if err = rows.Scan(&c.ID, &c.Type, &c.Title, &c.Content); err != nil {
			return
		}
		c.Change()
		cm[c.ID] = c
	}
	if err = rows.Err(); err != nil {
		log.Error("rows.Err() error=%+v", err)
		return nil, err
	}
	return
}
