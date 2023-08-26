package tab

import (
	"context"
	"time"

	"go-common/library/database/sql"
	"go-gateway/app/app-svr/app-channel/interface/conf"
	"go-gateway/app/app-svr/app-channel/interface/model/tab"
)

const (
	_getAllMenuSQL = `SELECT ct.id,ct.tag_id,ct.tab_id,ct.title,ct.priority,a.name from channel_tab AS ct,app_active AS a 
WHERE ct.stime<? AND ct.etime>? AND ct.is_delete=0 AND a.id=ct.tab_id AND ct.is_new=0 ORDER BY ct.priority ASC`
	_getAllMenuNewSQL = `SELECT ct.id,ct.tag_id,ct.tab_id,ct.title,ct.priority,a.name from channel_tab AS ct,app_active AS a 
WHERE ct.stime<? AND ct.etime>? AND ct.is_delete=0 AND a.id=ct.tab_id AND ct.is_new=1 ORDER BY ct.priority ASC`
)

type Dao struct {
	db *sql.DB
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		db: sql.NewMySQL(c.MySQL.Show),
	}
	return
}

// Menus menus tab
func (d *Dao) Menus(c context.Context, now time.Time) (menus map[int64][]*tab.Menu, err error) {
	var (
		rows *sql.Rows
	)
	if rows, err = d.db.Query(c, _getAllMenuSQL, now.Unix(), now.Unix()); err != nil {
		return
	}
	defer rows.Close()
	menus = map[int64][]*tab.Menu{}
	for rows.Next() {
		m := &tab.Menu{}
		if err = rows.Scan(&m.ID, &m.TagID, &m.TabID, &m.Name, &m.Priority, &m.Title); err != nil {
			return
		}
		menus[m.TagID] = append(menus[m.TagID], m)
	}
	err = rows.Err()
	return
}

// Menus menus tab
func (d *Dao) MenusNew(c context.Context, now time.Time) (menus map[int64][]*tab.Menu, err error) {
	var (
		rows *sql.Rows
	)
	if rows, err = d.db.Query(c, _getAllMenuNewSQL, now.Unix(), now.Unix()); err != nil {
		return
	}
	defer rows.Close()
	menus = map[int64][]*tab.Menu{}
	for rows.Next() {
		m := &tab.Menu{}
		if err = rows.Scan(&m.ID, &m.TagID, &m.TabID, &m.Name, &m.Priority, &m.Title); err != nil {
			return
		}
		menus[m.TagID] = append(menus[m.TagID], m)
	}
	err = rows.Err()
	return
}
