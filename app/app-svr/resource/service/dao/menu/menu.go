package menu

import (
	"context"

	"go-common/library/log"

	api "go-gateway/app/app-svr/resource/service/api/v1"
	"go-gateway/app/app-svr/resource/service/model"
)

const (
	_getAllMenuSQL   = "SELECT id,plat,name,ctype,cvalue,plat_ver,stime,etime,status,color,badge FROM app_menus ORDER BY `order` ASC"
	_getAllActiveSQL = "SELECT id,parent_id,name,background,type,content FROM app_active"
)

func (d *Dao) Menus(c context.Context) (res []*api.Menu, err error) {
	rows, err := d.db.Query(c, _getAllMenuSQL)
	if err != nil {
		log.Error("%+v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		m := &api.Menu{}
		if err = rows.Scan(&m.TabId, &m.Plat, &m.Name, &m.CType, &m.CValue, &m.PlatVersion, &m.STime, &m.ETime, &m.Status, &m.Color, &m.Badge); err != nil {
			log.Error("%+v", err)
			return
		}
		res = append(res, m)
	}
	err = rows.Err()
	return
}

func (d *Dao) Actives(c context.Context) (res []*api.Active, err error) {
	rows, err := d.db.Query(c, _getAllActiveSQL)
	if err != nil {
		log.Error("%+v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		ac := &api.Active{}
		if err = rows.Scan(&ac.Id, &ac.ParentID, &ac.Name, &ac.Background, &ac.Type, &ac.Content); err != nil {
			log.Error("%+v", err)
			return
		}
		res = append(res, ac)
	}
	err = rows.Err()
	return
}

func (d *Dao) AllMenus(c context.Context) (res []*model.Menu, err error) {
	rows, err := d.db.Query(c, _getAllMenuSQL)
	if err != nil {
		log.Error("%+v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		m := &model.Menu{}
		if err = rows.Scan(&m.TabID, &m.Plat, &m.Name, &m.CType, &m.CValue, &m.PlatVersion, &m.STime, &m.ETime, &m.Status, &m.Color, &m.Badge); err != nil {
			log.Error("%+v", err)
			return
		}
		if m.CValue == "" {
			continue
		}
		if !m.Change() {
			continue
		}
		res = append(res, m)
	}
	err = rows.Err()
	return
}
