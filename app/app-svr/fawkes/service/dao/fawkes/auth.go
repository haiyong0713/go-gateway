package fawkes

import (
	"context"
	"fmt"
	"strings"

	"go-common/library/database/sql"
	"go-common/library/database/xsql"

	"go-gateway/app/app-svr/fawkes/service/api/app/auth"
	authmdl "go-gateway/app/app-svr/fawkes/service/model/auth"
)

const (
	// auth_group
	_selectAllAuthGroup = `SELECT id,name,operator,ctime,mtime FROM auth_group`
	_selectAuthGroup    = `SELECT id,name,operator,ctime,mtime FROM auth_group WHERE name=?`
	_addAuthGroup       = `INSERT INTO auth_group (name, operator) VALUES (?,?)`
	_updateAuthGroup    = `UPDATE auth_group SET name=?,operator=? WHERE id=?`

	// auth_item
	_selectItem           = `SELECT id,auth_group_id,name,fe_key,be_url,url_param,operator,ctime,mtime,is_active FROM auth_item WHERE name=? AND be_url=? AND url_param=?`
	_selectItemByUrl      = `SELECT id,auth_group_id,name,fe_key,be_url,url_param,operator,ctime,mtime,is_active FROM auth_item WHERE be_url=?`
	_selectItemByID       = `SELECT id,auth_group_id,name,fe_key,be_url,url_param,operator,ctime,mtime,is_active FROM auth_item WHERE id=?`
	_selectAllItem        = `SELECT id,auth_group_id,name,fe_key,be_url,url_param,operator,ctime,mtime,is_active FROM auth_item`
	_addAuthItem          = `INSERT INTO auth_item(auth_group_id,name,fe_key,be_url,url_param,operator)VALUES (?,?,?,?,?,?)`
	_updateAuthItem       = `UPDATE auth_item set name=?,fe_key=?,be_url=?,url_param=?,operator=? WHERE id=?`
	_updateAuthItemActive = `UPDATE auth_item set is_active=?,operator=? WHERE id=?`
	_deleteAuthItem       = `DELETE from auth_item WHERE id=?`

	// auth_item_role_relation
	_selectAllRelation           = `SELECT id,auth_item_id,auth_role_value,operator,ctime,mtime FROM auth_item_role_relation`
	_selectAuthRelation          = `SELECT id,auth_item_id,auth_role_value,operator,ctime,mtime FROM auth_item_role_relation where auth_item_id = ?`
	_batchAddItemRoleRelation    = `INSERT INTO auth_item_role_relation (auth_item_id,auth_role_value,operator) VALUES %s ON DUPLICATE KEY UPDATE auth_item_id=auth_item_id, auth_role_value=auth_role_value, operator=operator`
	_batchDeleteItemRoleRelation = `DELETE FROM auth_item_role_relation WHERE (auth_item_id,auth_role_value) IN (%s)`
)

func (d *Dao) SelectAuthGroup(ctx context.Context, name string) (group *authmdl.Group, err error) {
	var rows *sql.Rows
	if rows, err = d.db.Query(ctx, _selectAuthGroup, name); err != nil {
		return
	}
	if err = rows.Err(); err != nil {
		return
	}
	groups := make([]*authmdl.Group, 0, 1)
	if err = xsql.ScanSlice(rows, &groups); err != nil {
		return
	}
	if len(groups) == 0 {
		return
	}
	group = groups[0]
	return
}

func (d *Dao) SelectAuthItem(ctx context.Context, name, beUrl, urlParam string) (item *authmdl.Item, err error) {
	var rows *sql.Rows
	if rows, err = d.db.Query(ctx, _selectItem, name, beUrl, urlParam); err != nil {
		return
	}
	if err = rows.Err(); err != nil {
		return
	}
	items := make([]*authmdl.Item, 0, 1)
	if err = xsql.ScanSlice(rows, &items); err != nil {
		return
	}
	if len(items) == 0 {
		return
	}
	item = items[0]
	return
}

func (d *Dao) SelectAuthItemByUrl(ctx context.Context, url string) (items []*authmdl.Item, err error) {
	var rows *sql.Rows
	if rows, err = d.db.Query(ctx, _selectItemByUrl, url); err != nil {
		return
	}
	if err = rows.Err(); err != nil {
		return
	}
	if err = xsql.ScanSlice(rows, &items); err != nil {
		return
	}
	return
}

func (d *Dao) SelectAuthItemById(ctx context.Context, id int64) (item *authmdl.Item, err error) {
	var rows *sql.Rows
	if rows, err = d.db.Query(ctx, _selectItemByID, id); err != nil {
		return
	}
	if err = rows.Err(); err != nil {
		return
	}
	items := make([]*authmdl.Item, 0, 1)
	if err = xsql.ScanSlice(rows, &items); err != nil {
		return
	}
	if len(items) == 0 {
		return
	}
	item = items[0]
	return
}

func (d *Dao) SelectAllAuthGroup(ctx context.Context) (groups []*authmdl.Group, err error) {
	var rows *sql.Rows
	if rows, err = d.db.Query(ctx, _selectAllAuthGroup); err != nil {
		return
	}
	if err = rows.Err(); err != nil {
		return
	}
	if err = xsql.ScanSlice(rows, &groups); err != nil {
		return
	}
	return
}

func (d *Dao) SelectAllIAuthItem(ctx context.Context) (items []*authmdl.Item, err error) {
	var rows *sql.Rows
	if rows, err = d.db.Query(ctx, _selectAllItem); err != nil {
		return
	}
	if err = rows.Err(); err != nil {
		return
	}
	if err = xsql.ScanSlice(rows, &items); err != nil {
		return
	}
	return
}

func (d *Dao) SelectAllAuthRelation(ctx context.Context) (relations []*authmdl.ItemRoleRelation, err error) {
	var rows *sql.Rows
	if rows, err = d.db.Query(ctx, _selectAllRelation); err != nil {
		return
	}
	if err = rows.Err(); err != nil {
		return
	}
	if err = xsql.ScanSlice(rows, &relations); err != nil {
		return
	}
	return
}

func (d *Dao) SelectAuthRelation(ctx context.Context, authItemId int64) (relations []*authmdl.ItemRoleRelation, err error) {
	var rows *sql.Rows
	if rows, err = d.db.Query(ctx, _selectAuthRelation, authItemId); err != nil {
		return
	}
	if err = rows.Err(); err != nil {
		return
	}
	if err = xsql.ScanSlice(rows, &relations); err != nil {
		return
	}
	return
}

func (d *Dao) AddAuthGroup(ctx context.Context, name string, operator string) (id int64, err error) {
	row, err := d.db.Exec(ctx, _addAuthGroup, name, operator)
	if err != nil {
		return
	}
	return row.LastInsertId()
}

func (d *Dao) UpdateAuthGroup(ctx context.Context, groupId int64, name string, operator string) (id int64, err error) {
	row, err := d.db.Exec(ctx, _updateAuthGroup, name, operator, groupId)
	if err != nil {
		return
	}
	return row.LastInsertId()
}

func (d *Dao) AddAuthItem(ctx context.Context, groupId int64, name string, feKey string, beUrl string, param string, operator string) (id int64, err error) {
	row, err := d.db.Exec(ctx, _addAuthItem, groupId, name, feKey, beUrl, param, operator)
	if err != nil {
		return
	}
	return row.LastInsertId()
}

func (d *Dao) UpdateAuthItem(ctx context.Context, itemId int64, name string, feKey string, beUrl string, param string, operator string) (id int64, err error) {
	row, err := d.db.Exec(ctx, _updateAuthItem, name, feKey, beUrl, param, operator, itemId)
	if err != nil {
		return
	}
	return row.LastInsertId()
}

func (d *Dao) UpdateAuthItemActive(ctx context.Context, itemId int64, isActive bool, operator string) (id int64, err error) {
	row, err := d.db.Exec(ctx, _updateAuthItemActive, isActive, operator, itemId)
	if err != nil {
		return
	}
	return row.LastInsertId()
}

func (d *Dao) DeleteAuthItem(ctx context.Context, itemId int64) (id int64, err error) {
	result, err := d.db.Exec(ctx, _deleteAuthItem, itemId)
	if err != nil {
		return
	}
	return result.LastInsertId()
}

func (d *Dao) AddAuthItemRoleRelation(ctx context.Context, item []*auth.Grant, operator string) (id int64, err error) {
	if len(item) == 0 {
		return
	}
	var (
		sqls = make([]string, 0, len(item))
		args = make([]interface{}, 0)
	)
	for _, v := range item {
		sqls = append(sqls, "(?,?,?)")
		args = append(args, v.ItemId, v.RoleValue, operator)
	}
	result, err := d.db.Exec(ctx, fmt.Sprintf(_batchAddItemRoleRelation, strings.Join(sqls, ",")), args...)
	if err != nil {
		return
	}
	return result.LastInsertId()
}

func (d *Dao) DeleteAuthItemRoleRelation(ctx context.Context, item []*auth.Grant) (id int64, err error) {
	if len(item) == 0 {
		return
	}
	var (
		sqls = make([]string, 0, len(item))
		args = make([]interface{}, 0)
	)
	for _, v := range item {
		sqls = append(sqls, "(?,?)")
		args = append(args, v.ItemId, v.RoleValue)
	}
	result, err := d.db.Exec(ctx, fmt.Sprintf(_batchDeleteItemRoleRelation, strings.Join(sqls, ",")), args...)
	if err != nil {
		return
	}
	return result.LastInsertId()
}
