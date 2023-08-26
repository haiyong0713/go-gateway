package native

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	xsql "go-common/library/database/sql"
	"go-common/library/log"
	"go-common/library/xstr"

	v1 "go-gateway/app/web-svr/native-page/interface/api"
	"go-gateway/app/web-svr/native-page/interface/model/dynamic"

	"github.com/pkg/errors"
)

var (
	_tsModuleSQL    = "SELECT `id`,`category`,`ts_id`,`ctime`,`mtime`,`state`,`p_type`,`rank`,`meta`,`remark`,`width`,`length`,`ukey`,`num`,`attribute` FROM `native_ts_module` WHERE id in (%s)"
	_tsIDsSQL       = "SELECT `id`,`rank` FROM `native_ts_module` WHERE ts_id =? AND `state`=1"
	_tsAddModuleSQL = "INSERT INTO `native_ts_module` (`category`,`ts_id`,`state`,`p_type`,`rank`,`meta`,`remark`,`width`,`length`,`ukey`,`num`,`attribute`) VALUES %s"
	_tsAddModResSQL = "INSERT INTO `native_ts_module_resource` (`module_id`,`resource_id`,`resource_type`,`rank`,`resource_from`,`state`,`ext`) VALUES %s"
	_tsModResIDsSQL = "SELECT `module_id`,`resource_id`,`resource_type`,`rank`,`resource_from`,`ext` FROM `native_ts_module_resource` where `module_id` in (%s) AND `state`=1"
	_tsMInvalidSQL  = "UPDATE `native_ts_module` set `state`=0 where `id` in (%s)"
)

// RawNtTsModulesExt .
func (d *Dao) RawNtTsModulesExt(c context.Context, moduleIDs []int64) (list map[int64]*dynamic.NativeTsModuleExt, err error) {
	if len(moduleIDs) == 0 {
		return
	}
	rows, err := d.db.Query(c, fmt.Sprintf(_tsModuleSQL, xstr.JoinInts(moduleIDs)))
	if err != nil {
		if err == xsql.ErrNoRows {
			err = nil
		}
		return
	}
	defer rows.Close()
	list = make(map[int64]*dynamic.NativeTsModuleExt)
	for rows.Next() {
		t := &dynamic.NativeTsModuleExt{}
		if err = rows.Scan(&t.Id, &t.Category, &t.TsID, &t.Ctime, &t.Mtime, &t.State, &t.PType, &t.Rank, &t.Meta, &t.Remark, &t.Width, &t.Length, &t.Ukey, &t.Num, &t.Attribute); err != nil {
			err = errors.Wrap(err, "rows.Scan")
			return
		}
		list[t.Id] = t
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "rows.Err")
		return
	}
	resources, err := d.NtTsModResources(c, moduleIDs)
	if err != nil {
		return
	}
	for _, v := range list {
		if _, ok := resources[v.Id]; !ok {
			continue
		}
		v.Resources = resources[v.Id]
	}
	return
}

func (d *Dao) NtTsModResources(c context.Context, moduleIDs []int64) (map[int64][]*v1.NativeTsModuleResource, error) {
	if len(moduleIDs) == 0 {
		return map[int64][]*v1.NativeTsModuleResource{}, nil
	}
	querySql := fmt.Sprintf(_tsModResIDsSQL, xstr.JoinInts(moduleIDs))
	rows, err := d.db.Query(c, querySql)
	if err != nil {
		if err == sql.ErrNoRows {
			return map[int64][]*v1.NativeTsModuleResource{}, nil
		}
		log.Errorc(c, "Fail to query tsModResIDsSQL, sql=%s error=%+v", querySql, err)
		return nil, err
	}
	defer rows.Close()
	list := make(map[int64][]*v1.NativeTsModuleResource, len(moduleIDs))
	for rows.Next() {
		m := &v1.NativeTsModuleResource{}
		err = rows.Scan(&m.ModuleID, &m.ResourceID, &m.ResourceType, &m.Rank, &m.ResourceFrom, &m.Ext)
		if err != nil {
			log.Errorc(c, "Fail to scan NativeTsModuleResource row, error=%+v", err)
			continue
		}
		if _, ok := list[m.ModuleID]; !ok {
			list[m.ModuleID] = make([]*v1.NativeTsModuleResource, 0, 10)
		}
		list[m.ModuleID] = append(list[m.ModuleID], m)
	}
	err = rows.Err()
	if err != nil {
		log.Errorc(c, "Fail to get NativeTsModuleResource rows, error=%+v", err)
		return nil, err
	}
	return list, nil
}

// NtTsModuleIDSearch .
func (d *Dao) NtTsModuleIDSearch(c context.Context, tsID int64) (list []*v1.NativeTsModule, err error) {
	var rows *xsql.Rows
	if rows, err = d.db.Query(c, _tsIDsSQL, tsID); err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		t := &v1.NativeTsModule{}
		if err = rows.Scan(&t.Id, &t.Rank); err != nil {
			err = errors.Wrap(err, "rows.Scan")
			return
		}
		list = append(list, t)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "rows.Err")
	}
	return
}

// TsPageSave .
func (d *Dao) TsModuleSave(c context.Context, vals []*dynamic.NativeTsModuleExt, tsID int64) error {
	if len(vals) == 0 {
		return nil
	}
	var (
		rows    []interface{}
		rowsTmp []string
	)
	moduleExts := make([]*dynamic.NativeTsModuleExt, 0, len(vals))
	for _, v := range vals {
		if len(v.Resources) != 0 {
			moduleExts = append(moduleExts, v)
			continue
		}
		rowsTmp = append(rowsTmp, "(?,?,?,?,?,?,?,?,?,?,?,?)")
		rows = append(rows, v.Category, tsID, v.State, v.PType, v.Rank, v.Meta, v.Remark, v.Width, v.Length, v.Ukey, v.Num, v.Attribute)
	}
	if len(rows) > 0 && len(rowsTmp) > 0 {
		sqlStr := fmt.Sprintf(_tsAddModuleSQL, strings.Join(rowsTmp, ","))
		_, err := d.db.Exec(c, sqlStr, rows...)
		if err != nil {
			log.Error("TsModuleSave arg:%v error(%v)", vals, err)
			return err
		}
	}
	return d.TsModuleExtSave(c, moduleExts, tsID)
}

func (d *Dao) TsModuleExtSave(c context.Context, moduleExts []*dynamic.NativeTsModuleExt, tsID int64) error {
	if len(moduleExts) == 0 {
		return nil
	}
	var err error
	tx, err := d.db.Begin(c)
	if err != nil {
		log.Errorc(c, "Fail to begin transaction, error=%+v", err)
		return err
	}
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			log.Errorc(c, "%+v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Errorc(c, "Fail to Rollback, error=%+v", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Errorc(c, "Fail to Commit, error=%+v", err)
		}
	}()
	for _, v := range moduleExts {
		sqlStr := fmt.Sprintf(_tsAddModuleSQL, "(?,?,?,?,?,?,?,?,?,?,?,?)")
		res, err := tx.Exec(sqlStr, v.Category, tsID, v.State, v.PType, v.Rank, v.Meta, v.Remark, v.Width, v.Length, v.Ukey, v.Num, v.Attribute)
		if err != nil {
			log.Errorc(c, "Fail to create native_ts_module, sqlStr=%+v error=%+v", sqlStr, err)
			return err
		}
		modID, err := res.LastInsertId()
		if err != nil {
			log.Errorc(c, "Fail to get native_ts_module id, sqlStr=%+v error=%+v", sqlStr, err)
			return err
		}
		err = d.TsModuleResourceSave(c, tx, v.Resources, modID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *Dao) TsModuleResourceSave(c context.Context, tx *xsql.Tx, tsResources []*v1.NativeTsModuleResource, modID int64) error {
	if len(tsResources) == 0 {
		return nil
	}
	rows := make([]interface{}, 0, len(tsResources))
	sqlParts := make([]string, 0, len(tsResources))
	for rank, v := range tsResources {
		sqlParts = append(sqlParts, "(?,?,?,?,?,?,?)")
		rows = append(rows, modID, v.ResourceID, v.ResourceType, rank+1, v.ResourceFrom, 1, v.Ext)
	}
	sqlStr := fmt.Sprintf(_tsAddModResSQL, strings.Join(sqlParts, ","))
	_, err := tx.Exec(sqlStr, rows...)
	if err != nil {
		log.Error("Fail to save native_ts_module_resource, sqlStr=%+v error=%+v", sqlStr, err)
		return err
	}
	return nil
}

func (d *Dao) DeleteTsModule(c context.Context, modIDs []int64) error {
	if len(modIDs) == 0 {
		return nil
	}
	if _, err := d.db.Exec(c, fmt.Sprintf(_tsMInvalidSQL, xstr.JoinInts(modIDs))); err != nil {
		log.Error("Fail to delete native_ts_module, moduleIDs=%+v error=%+v", modIDs, err)
		return err
	}
	return nil
}
