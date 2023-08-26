package dao

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go-common/library/log"

	"go-gateway/app/api-gateway/api-manager/internal/model"
)

func (d *dao) GroupCount(c context.Context) (int64, error) {
	var sql = `SELECT count(1) FROM group_main`
	row := d.db.QueryRow(c, sql)
	var count int64
	if err := row.Scan(&count); err != nil {
		log.Error("%v", err)
		return 0, err
	}
	return count, nil
}

func (d *dao) GroupByIDs(c context.Context, ids []int64) (res []*model.ContralGroup, err error) {
	var (
		sql    = `SELECT id,group_name,creator,modifier,manager,description,ctime,mtime FROM group_main WHERE id IN(%v) ORDER BY id DESC`
		sqlAdd []string
		args   []interface{}
	)
	for _, id := range ids {
		sqlAdd = append(sqlAdd, "?")
		args = append(args, id)
	}
	rows, err := d.db.Query(c, fmt.Sprintf(sql, strings.Join(sqlAdd, ",")), args...)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var re = new(model.ContralGroup)
		if err = rows.Scan(&re.ID, &re.GroupName, &re.Creator, &re.Modifier, &re.Manager, &re.Desc, &re.CTime, &re.MTime); err != nil {
			log.Error("%+v", err)
			return nil, err
		}
		res = append(res, re)
	}
	if err = rows.Err(); err != nil {
		log.Error("%v", err)
		return nil, err
	}
	return res, nil
}

func (d *dao) GroupByName(c context.Context, groupName string) (res []*model.ContralGroup, err error) {
	var sql = `SELECT id,group_name,creator,modifier,manager,description,ctime,mtime FROM group_main WHERE group_name=? ORDER BY id DESC`
	rows, err := d.db.Query(c, sql, groupName)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var re = new(model.ContralGroup)
		if err = rows.Scan(&re.ID, &re.GroupName, &re.Creator, &re.Modifier, &re.Manager, &re.Desc, &re.CTime, &re.MTime); err != nil {
			log.Error("%+v", err)
			return nil, err
		}
		res = append(res, re)
	}
	if err = rows.Err(); err != nil {
		log.Error("%v", err)
		return nil, err
	}
	return res, nil
}

func (d *dao) GroupList(c context.Context, groupName string, pageNum, pageSize int64) (res []*model.ContralGroup, err error) {
	var (
		sql    = `SELECT id,group_name,creator,modifier,manager,description,ctime,mtime FROM group_main %v ORDER BY id DESC LIMIT ?,?`
		args   []interface{}
		sqlAdd string
	)
	if groupName != "" {
		sqlAdd += `WHERE group_name Like ? `
		args = append(args, "%"+groupName+"%")
	}
	// limit 组装
	if pageNum > 0 && pageSize > 0 {
		args = append(args, (pageNum-1)*pageSize, pageSize)
	}
	rows, err := d.db.Query(c, fmt.Sprintf(sql, sqlAdd), args...)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var re = new(model.ContralGroup)
		if err = rows.Scan(&re.ID, &re.GroupName, &re.Creator, &re.Modifier, &re.Manager, &re.Desc, &re.CTime, &re.MTime); err != nil {
			log.Error("%+v", err)
			return nil, err
		}
		res = append(res, re)
	}
	if err = rows.Err(); err != nil {
		log.Error("%v", err)
		return nil, err
	}
	return res, nil
}

func (d *dao) GroupInsert(c context.Context, req *model.ContralGroup) (int64, error) {
	var sql = `INSERT INTO group_main (group_name,creator,modifier,description,ctime,mtime) VALUES (?,?,?,?,?,?)`
	row, err := d.db.Exec(c, sql, req.GroupName, req.Creator, req.Modifier, req.Desc, time.Now(), time.Now())
	if err != nil {
		log.Error("%+v", err)
		return 0, err
	}
	return row.LastInsertId()
}

func (d *dao) GroupUpdate(c context.Context, req *model.ContralGroup) (int64, error) {
	var sql = `UPDATE group_main SET modifier=?,description=?,mtime=? WHERE id=?`
	row, err := d.db.Exec(c, sql, req.Modifier, req.Desc, time.Now(), req.ID)
	if err != nil {
		log.Error("%+v", err)
		return 0, err
	}
	return row.LastInsertId()
}

func (d *dao) GroupFollowAdd(c context.Context, req *model.ContralGroupFollowActionPeq, username string) (int64, error) {
	var sql = `INSERT INTO group_follow (gid,uname,ctime,mtime) VALUES (?,?,?,?)`
	row, err := d.db.Exec(c, sql, req.Id, username, time.Now(), time.Now())
	if err != nil {
		log.Error("%+v", err)
		return 0, err
	}
	return row.LastInsertId()
}

func (d *dao) GroupFollowDel(c context.Context, req *model.ContralGroupFollowActionPeq, username string) (int64, error) {
	var sql = `DELETE FROM group_follow WHERE gid=? AND uname=?`
	row, err := d.db.Exec(c, sql, req.Id, username)
	if err != nil {
		log.Error("%+v", err)
		return 0, err
	}
	return row.LastInsertId()
}

func (d *dao) GroupFollowList(c context.Context, uname string) (res []int64, err error) {
	var sql = `SELECT gid FROM group_follow WHERE uname=?`
	rows, err := d.db.Query(c, sql, uname)
	if err != nil {
		log.Error("%v", err)
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var gid int64
		if err = rows.Scan(&gid); err != nil {
			log.Error("%v", err)
			return
		}
		res = append(res, gid)
	}
	if err = rows.Err(); err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	return res, nil
}

func (d *dao) ApiCount(c context.Context, gid int64) (int64, error) {
	var sql = `SELECT count(1) FROM api WHERE gid=?`
	row := d.db.QueryRow(c, sql, gid)
	var count int64
	if err := row.Scan(&count); err != nil {
		log.Error("%v", err)
		return 0, err
	}
	return count, nil
}

func (d *dao) ApiByIDs(c context.Context, ids []int64) ([]*model.ContralApi, error) {
	var (
		sql    = `SELECT id,gid,api_name,api_type,domain,router,handler,req,reply,dsl_code,dsl_struct,custom_code,creator,modifier,description,ctime,mtime FROM api WHERE id IN (%v) ORDER BY id DESC`
		sqlAdd []string
		args   []interface{}
	)
	for _, id := range ids {
		sqlAdd = append(sqlAdd, "?")
		args = append(args, id)
	}
	rows, err := d.db.Query(c, fmt.Sprintf(sql, strings.Join(sqlAdd, ",")), args...)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	defer rows.Close()
	var res []*model.ContralApi
	for rows.Next() {
		var re = new(model.ContralApi)
		if err = rows.Scan(&re.ID, &re.Gid, &re.ApiName, &re.ApiType, &re.Domain, &re.Router, &re.Handler, &re.Req, &re.Reply,
			&re.DSLCode, &re.DSLStruct, &re.CustomCode, &re.Creator, &re.Modifier, &re.Desc, &re.CTime, &re.MTime); err != nil {
			log.Error("%+v", err)
			return nil, err
		}
		res = append(res, re)
	}
	if rows.Err() != nil {
		log.Error("%+v", rows.Err())
		return nil, err
	}
	return res, nil
}

func (d *dao) ApiByName(c context.Context, apiName string) ([]*model.ContralApi, error) {
	var sql = `SELECT id,gid,api_name,api_type,domain,router,handler,req,reply,dsl_code,dsl_struct,custom_code,creator,modifier,description,ctime,mtime FROM api WHERE api_name=? ORDER BY id DESC`
	rows, err := d.db.Query(c, sql, apiName)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	defer rows.Close()
	var res []*model.ContralApi
	for rows.Next() {
		var re = new(model.ContralApi)
		if err = rows.Scan(&re.ID, &re.Gid, &re.ApiName, &re.ApiType, &re.Domain, &re.Router, &re.Handler, &re.Req, &re.Reply,
			&re.DSLCode, &re.DSLStruct, &re.CustomCode, &re.Creator, &re.Modifier, &re.Desc, &re.CTime, &re.MTime); err != nil {
			log.Error("%+v", err)
			return nil, err
		}
		res = append(res, re)
	}
	if rows.Err() != nil {
		log.Error("%+v", rows.Err())
		return nil, err
	}
	return res, nil
}

func (d *dao) ApiList(c context.Context, gid int64, apiName string, pageNum, pageSize int64) ([]*model.ContralApi, error) {
	var (
		sql    = `SELECT id,gid,api_name,api_type,domain,router,handler,req,reply,dsl_code,dsl_struct,custom_code,creator,modifier,description,ctime,mtime FROM api WHERE %v ORDER BY id DESC LIMIT ?,?`
		sqlAdd = `gid=?`
		args   = []interface{}{gid}
	)
	// where 条件组装
	if apiName != "" {
		sqlAdd += `AND api_name LIKE ?`
		args = append(args, "%"+apiName+"%")
	}
	// limit 组装
	if pageNum > 0 && pageSize > 0 {
		args = append(args, (pageNum-1)*pageSize, pageSize)
	}
	rows, err := d.db.Query(c, fmt.Sprintf(sql, sqlAdd), args...)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	defer rows.Close()
	var res []*model.ContralApi
	for rows.Next() {
		var re = new(model.ContralApi)
		if err = rows.Scan(&re.ID, &re.Gid, &re.ApiName, &re.ApiType, &re.Domain, &re.Router, &re.Handler, &re.Req, &re.Reply,
			&re.DSLCode, &re.DSLStruct, &re.CustomCode, &re.Creator, &re.Modifier, &re.Desc, &re.CTime, &re.MTime); err != nil {
			log.Error("%+v", err)
			return nil, err
		}
		res = append(res, re)
	}
	if rows.Err() != nil {
		log.Error("%+v", rows.Err())
		return nil, err
	}
	return res, nil
}

func (d *dao) ApiInsert(c context.Context, req *model.ContralApi) (int64, error) {
	var (
		sql = `INSERT INTO api (gid,api_name,api_type,domain,router,handler,req,reply,dsl_code,dsl_struct,custom_code,creator,modifier,description,ctime,mtime) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`
		now = time.Now()
	)
	row, err := d.db.Exec(c, sql, req.Gid, req.ApiName, req.ApiType, req.Domain, req.Router, req.Handler, req.Req, req.Reply,
		req.DSLCode, req.DSLStruct, req.CustomCode, req.Creator, req.Modifier, req.Desc, now, now)
	if err != nil {
		log.Error("%+v", err)
		return 0, err
	}
	return row.LastInsertId()
}

func (d *dao) ApiUpdate(c context.Context, req *model.ContralApi) (int64, error) {
	var (
		sql = `UPDATE api SET api_type=?,domain=?,router=?,handler=?,req=?,reply=?,dsl_code=?,dsl_struct=?,custom_code=?,modifier=?,description=?,mtime=? WHERE id=?`
		now = time.Now()
	)
	row, err := d.db.Exec(c, sql, req.ApiType, req.Domain, req.Router, req.Handler, req.Req, req.Reply, req.DSLCode,
		req.DSLStruct, req.CustomCode, req.Modifier, req.Desc, now, req.ID)
	if err != nil {
		log.Error("%+v", err)
		return 0, err
	}
	return row.LastInsertId()
}

func (d *dao) ApiConfigCount(c context.Context, id int64) (int64, error) {
	var sql = `SELECT count(1) FROM api_config WHERE api_id=?`
	row := d.db.QueryRow(c, sql, id)
	var count int64
	if err := row.Scan(&count); err != nil {
		log.Error("%v", err)
		return 0, err
	}
	return count, nil
}

func (d *dao) ApiConfigByID(c context.Context, id int64) ([]*model.ContralApiConfig, error) {
	var sql = `SELECT id,api_id,version,api_type,domain,router,handler,req,reply,dsl_code,dsl_struct,custom_code,creator,description,ctime,mtime FROM api_config WHERE id=?`
	rows, err := d.db.Query(c, sql, id)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	defer rows.Close()
	var res []*model.ContralApiConfig
	for rows.Next() {
		var re = new(model.ContralApiConfig)
		if err = rows.Scan(&re.ID, &re.ApiID, &re.Version, &re.ApiType, &re.Domain, &re.Router, &re.Handler, &re.Req,
			&re.Reply, &re.DSLCode, &re.DSLStruct, &re.CustomCode, &re.Creator, &re.Desc, &re.CTime, &re.MTime); err != nil {
			log.Error("%+v", err)
			return nil, err
		}
		res = append(res, re)
	}
	if err = rows.Err(); err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	return res, nil
}

func (d *dao) ApiConfigByVersion(c context.Context, apiID int64, version string) ([]*model.ContralApiConfig, error) {
	var sql = `SELECT id,api_id,version,api_type,domain,router,handler,req,reply,dsl_code,dsl_struct,custom_code,creator,description,ctime,mtime FROM api_config WHERE api_id=? AND version=?`
	rows, err := d.db.Query(c, sql, apiID, version)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	defer rows.Close()
	var res []*model.ContralApiConfig
	for rows.Next() {
		var re = new(model.ContralApiConfig)
		if err = rows.Scan(&re.ID, &re.ApiID, &re.Version, &re.ApiType, &re.Domain, &re.Router, &re.Handler,
			&re.Req, &re.Reply, &re.DSLCode, &re.DSLStruct, &re.CustomCode, &re.Creator, &re.Desc, &re.CTime, &re.MTime); err != nil {
			log.Error("%+v", err)
			return nil, err
		}
		res = append(res, re)
	}
	if err = rows.Err(); err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	return res, nil
}

func (d *dao) ApiConfigList(c context.Context, apiID, pageNum, pageSize int64) ([]*model.ContralApiConfig, error) {
	var (
		sql  = `SELECT id,api_id,version,api_type,domain,router,handler,req,reply,dsl_code,dsl_struct,custom_code,creator,description,ctime,mtime FROM api_config WHERE api_id=? ORDER BY id DESC LIMIT ?,?`
		args = []interface{}{apiID}
	)
	// limit 组装
	if pageNum > 0 && pageSize > 0 {
		args = append(args, (pageNum-1)*pageSize, pageSize)
	}
	rows, err := d.db.Query(c, sql, args...)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	defer rows.Close()
	var res []*model.ContralApiConfig
	for rows.Next() {
		var re = new(model.ContralApiConfig)
		if err = rows.Scan(&re.ID, &re.ApiID, &re.Version, &re.ApiType, &re.Domain, &re.Router, &re.Handler,
			&re.Req, &re.Reply, &re.DSLCode, &re.DSLStruct, &re.CustomCode, &re.Creator, &re.Desc, &re.CTime, &re.MTime); err != nil {
			log.Error("%+v", err)
			return nil, err
		}
		res = append(res, re)
	}
	if rows.Err() != nil {
		log.Error("%+v", rows.Err())
		return nil, err
	}
	return res, nil
}

func (d *dao) ApiConfigInsert(c context.Context, req *model.ContralApiConfig) (int64, error) {
	var (
		sql = `INSERT INTO api_config (api_id,version,api_type,domain,router,handler,req,reply,dsl_code,dsl_struct,custom_code,creator,description,ctime,mtime) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`
		now = time.Now()
	)
	row, err := d.db.Exec(c, sql, req.ApiID, req.Version, req.ApiType, req.Domain, req.Router, req.Handler, req.Req,
		req.Reply, req.DSLCode, req.DSLStruct, req.CustomCode, req.Creator, req.Desc, now, now)
	if err != nil {
		log.Error("%+v", err)
		return 0, err
	}
	return row.LastInsertId()
}

func (d *dao) ApiPublishCount(c context.Context, id int64) (int64, error) {
	var sql = `SELECT count(1) FROM api_publish WHERE api_id=?`
	row := d.db.QueryRow(c, sql, id)
	var count int64
	if err := row.Scan(&count); err != nil {
		log.Error("%v", err)
		return 0, err
	}
	return count, nil
}

func (d *dao) ApiPublishList(c context.Context, apiID, pageNum, pageSize int64) ([]*model.ContralApiPublish, error) {
	var (
		sql  = `SELECT id,api_id,version,state,operator,ctime,mtime FROM api_publish WHERE api_id=? ORDER BY id DESC LIMIT ?,?`
		args = []interface{}{apiID}
	)
	// limit 组装
	if pageNum > 0 && pageSize > 0 {
		args = append(args, (pageNum-1)*pageSize, pageSize)
	}
	rows, err := d.db.Query(c, sql, args...)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	defer rows.Close()
	var res []*model.ContralApiPublish
	for rows.Next() {
		var re = new(model.ContralApiPublish)
		if err = rows.Scan(&re.ID, &re.ApiID, &re.Version, &re.State, &re.Operator, &re.CTime, &re.MTime); err != nil {
			log.Error("%+v", err)
			return nil, err
		}
		res = append(res, re)
	}
	if rows.Err() != nil {
		log.Error("%+v", rows.Err())
		return nil, err
	}
	return res, nil
}

func (d *dao) ApiPublishSave(c context.Context, req *model.ContralApiPublish) (int64, error) {
	var (
		sql = `INSERT INTO api_publish (api_id,version,state,operator,ctime,mtime) VALUES (?,?,?,?,?,?) ON DUPLICATE KEY UPDATE state=?,operator=?,mtime=?`
		now = time.Now()
	)
	row, err := d.db.Exec(c, sql, req.ApiID, req.Version, req.State, req.Operator, now, now, req.State, req.Operator, now)
	if err != nil {
		log.Error("%+v", err)
		return 0, err
	}
	return row.LastInsertId()
}
