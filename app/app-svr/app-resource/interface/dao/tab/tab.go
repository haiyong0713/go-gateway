package tab

import (
	"context"
	"time"

	"go-common/library/database/sql"
	"go-gateway/app/app-svr/app-resource/interface/component"
	"go-gateway/app/app-svr/app-resource/interface/conf"
	"go-gateway/app/app-svr/app-resource/interface/model/tab"
)

const (
	_getAllMenuSQL = "SELECT id,plat,name,ctype,cvalue,plat_ver,status,color,badge,attribute,area,show_purposed,area_policy FROM app_menus WHERE stime<? AND etime>? AND status=1 ORDER BY `order` ASC"
)

type Dao struct {
	db *sql.DB
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		db: component.GlobalDB,
	}
	return
}

// Menus menus tab
func (d *Dao) Menus(c context.Context, now time.Time) (menus []*tab.Menu, err error) {
	var rows *sql.Rows
	if rows, err = d.db.Query(c, _getAllMenuSQL, now, now); err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		m := &tab.Menu{}
		if err = rows.Scan(&m.TabID, &m.Plat, &m.Name, &m.CType, &m.CValue, &m.PlatVersion, &m.Status, &m.Color, &m.Badge, &m.Attribute, &m.Area, &m.ShowPurposed, &m.AreaPolicy); err != nil {
			return
		}
		if m.CValue != "" {
			m.Change()
			menus = append(menus, m)
		}
	}
	err = rows.Err()
	return
}
