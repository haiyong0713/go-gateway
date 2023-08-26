package sidebar

import (
	"context"
	"fmt"
	"time"

	xsql "go-common/library/database/sql"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-resource/interface/component"
	"go-gateway/app/app-svr/app-resource/interface/conf"
	"go-gateway/app/app-svr/app-resource/interface/model/sidebar"
)

const (
	_initSidebarKey = "sidebar_%d_%d"
	_selSideSQL     = `SELECT s.id,s.plat,s.module,s.name,s.logo,s.logo_white,s.param,s.rank,l.build,l.conditions,s.tip FROM 
		sidebar AS s,sidebar_limit AS l WHERE s.state=1 AND s.id=l.s_id AND s.online_time<? ORDER BY s.rank DESC,l.id ASC`
)

type Dao struct {
	db  *xsql.DB
	get *xsql.Stmt
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		db: component.GlobalDB,
	}
	d.get = d.db.Prepared(_selSideSQL)
	return
}

// SideBar
func (d *Dao) SideBar(ctx context.Context, now time.Time) (ss map[string][]*sidebar.SideBar, limits map[int64][]*sidebar.Limit, err error) {
	ss = map[string][]*sidebar.SideBar{}
	limits = map[int64][]*sidebar.Limit{}
	rows, err := d.get.Query(ctx, now)
	if err != nil {
		log.Error("mysqlDB.Query error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		s := &sidebar.SideBar{}
		if err = rows.Scan(&s.ID, &s.Plat, &s.Module, &s.Name, &s.Logo, &s.LogoWhite, &s.Param, &s.Rank, &s.Build, &s.Conditions, &s.Tip); err != nil {
			log.Error("row.Scan error(%v)", err)
			return
		}
		key := fmt.Sprintf(_initSidebarKey, s.Plat, s.Module)
		if _, ok := limits[s.ID]; !ok {
			ss[key] = append(ss[key], s)
		}
		limit := &sidebar.Limit{
			ID:        s.ID,
			Build:     s.Build,
			Condition: s.Conditions,
		}
		limits[s.ID] = append(limits[s.ID], limit)
	}
	if err = rows.Err(); err != nil {
		log.Error("rows.Err() error=%+v", err)
		return nil, nil, err
	}
	return
}

func (d *Dao) Close() {
	if d.db != nil {
		d.db.Close()
	}
}
