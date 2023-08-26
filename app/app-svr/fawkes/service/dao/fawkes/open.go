package fawkes

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"go-common/library/database/xsql"

	"go-gateway/app/app-svr/fawkes/service/api/app/open"
	openmdl "go-gateway/app/app-svr/fawkes/service/model/open"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

const (
	//open_project
	_addProject         = "INSERT INTO open_project (project_name, owner, token, description, applicant) VALUES (?,?,?,?,?)"
	_selectProjectInfo  = "SELECT id,project_name,owner,token,description,applicant,ctime,mtime FROM open_project WHERE id=?"
	_updateProject      = "UPDATE open_project SET owner=?,description=? WHERE id=?"
	_selectProjectInfos = "SELECT id,project_name,owner,token,description,applicant,is_active,ctime,mtime FROM open_project WHERE %s limit ?,?"
	_countProjectInfos  = "SELECT COUNT(*) FROM open_project WHERE %s"
	_activeProject      = "UPDATE open_project SET is_active=? WHERE id=?"
	_selectToken        = "SELECT token,project_name,is_active FROM open_project WHERE token=?"

	//open_path_access
	_addProjectPath         = "INSERT INTO open_path_access (project_id,router,allowed_app_key,description) VALUES %s"
	_deleteProjectPath      = "DELETE FROM open_path_access WHERE id IN (%s)"
	_selectProjectPaths     = "SELECT id,project_id,router,allowed_app_key,ctime,mtime FROM open_path_access WHERE id IN (%s)"
	_selectProjectPath      = "SELECT id,project_id,router,allowed_app_key,description,ctime,mtime FROM open_path_access WHERE project_id=?"
	_batchUpdateProjectPath = "UPDATE open_path_access SET `allowed_app_key` = CASE %s END ,`description` = CASE %s END WHERE id IN(%s)"

	//open_user_project_relation
	_addUserRelations          = "INSERT INTO open_user_project_relation (user_name, project_id) VALUES %s"
	_deleteAllRelations        = "DELETE FROM open_user_project_relation WHERE project_id=?"
	_selectProjectOwners       = "SELECT user_name,project_id FROM open_user_project_relation WHERE project_id=?"
	_selectProjectOwnersByUser = "SELECT user_name,project_id FROM open_user_project_relation WHERE user_name=?"
)

func (d *Dao) AddProject(ctx context.Context, name string, owner string, token string, description string, op string) (id int64, err error) {
	row, err := d.db.Exec(ctx, _addProject, name, owner, token, description, op)
	if err != nil {
		return
	}
	return row.LastInsertId()
}

func (d *Dao) AddUserRelations(ctx context.Context, projectId int64, owners []string) (err error) {
	var (
		sqls = make([]string, 0, len(owners))
		args = make([]interface{}, 0)
	)
	for _, v := range owners {
		sqls = append(sqls, "(?,?)")
		args = append(args, v, projectId)
	}
	if _, err = d.db.Exec(ctx, fmt.Sprintf(_addUserRelations, strings.Join(sqls, ",")), args...); err != nil {
		log.Error("AddUserRelations error: %v", err)
	}
	return err
}

func (d *Dao) UpdateProjectOwnerRelations(ctx context.Context, projectId int64, owners []string) (err error) {
	if _, err = d.db.Exec(ctx, _deleteAllRelations, projectId); err != nil {
		return
	}
	err = d.AddUserRelations(ctx, projectId, owners)
	if err != nil {
		return
	}
	return
}

func (d *Dao) SelectProjectOwnerRelationByProject(ctx context.Context, id int64) (relations []*openmdl.UserProjectRelation, err error) {
	rows, err := d.db.Query(ctx, _selectProjectOwners, id)
	if err != nil {
		return
	}
	defer rows.Close()
	if err = xsql.ScanSlice(rows, &relations); err != nil {
		log.Error("ScanSlice Error: %v", err)
		return
	}
	err = rows.Err()
	return
}

func (d *Dao) SelectProjectOwnerRelationByUser(ctx context.Context, user string) (relations []*openmdl.UserProjectRelation, err error) {
	rows, err := d.db.Query(ctx, _selectProjectOwnersByUser, user)
	if err != nil {
		return
	}
	defer rows.Close()
	if err = xsql.ScanSlice(rows, &relations); err != nil {
		log.Error("ScanSlice Error: %v", err)
		return
	}
	err = rows.Err()
	return
}

func (d *Dao) SelectProjectInfo(ctx context.Context, id int64) (project *openmdl.Project, err error) {
	rows, err := d.db.Query(ctx, _selectProjectInfo, id)
	if err != nil {
		return
	}
	defer rows.Close()
	var list []*openmdl.Project
	if err = xsql.ScanSlice(rows, &list); err != nil {
		log.Error("ScanSlice Error: %v", err)
		return
	}
	if len(list) != 1 {
		return
	}
	err = rows.Err()
	project = list[0]
	return
}

func (d *Dao) UpdateProject(ctx context.Context, projectId int64, owners string, description string) (id int64, err error) {
	var r sql.Result
	if r, err = d.db.Exec(ctx, _updateProject, owners, description, projectId); err != nil {
		log.Error("AddUserRelations error: %v", err)
		return
	}
	return r.LastInsertId()
}

func (d *Dao) AddProjectPath(ctx context.Context, projectId int64, access []*open.RouterAccess) (err error) {
	var (
		sqls = make([]string, 0, len(access))
		args = make([]interface{}, 0)
	)
	for _, v := range access {
		sqls = append(sqls, "(?,?,?,?)")
		args = append(args, projectId, v.Path, strings.Join(v.AppKey, openmdl.Comma), v.Description)
	}
	if _, err = d.db.Exec(ctx, fmt.Sprintf(_addProjectPath, strings.Join(sqls, openmdl.Comma)), args...); err != nil {
		log.Error("AddProjectPath error: %v", err)
		return
	}
	return err
}

func (d *Dao) DeleteProjectPath(ctx context.Context, ids []int64) (effected int64, err error) {
	if len(ids) == 0 {
		return
	}
	var (
		sqls = make([]string, 0, len(ids))
		args = make([]interface{}, 0)
	)
	for _, v := range ids {
		sqls = append(sqls, "?")
		args = append(args, v)
	}
	var r sql.Result
	if r, err = d.db.Exec(ctx, fmt.Sprintf(_deleteProjectPath, strings.Join(sqls, ",")), args...); err != nil {
		log.Error("AddProjectPath error: %v", err)
		return
	}
	return r.RowsAffected()
}

func (d *Dao) SelectProjectPaths(ctx context.Context, ids []int64) (paths []*openmdl.PathAccess, err error) {
	if len(ids) == 0 {
		return
	}
	var (
		sqls = make([]string, 0, len(ids))
		args = make([]interface{}, 0)
	)
	for _, v := range ids {
		sqls = append(sqls, "?")
		args = append(args, v)
	}
	rows, err := d.db.Query(ctx, fmt.Sprintf(_selectProjectPaths, strings.Join(sqls, ",")), args...)
	if err != nil {
		return
	}
	defer rows.Close()
	if err = xsql.ScanSlice(rows, &paths); err != nil {
		log.Error("ScanSlice Error: %v", err)
		return
	}
	err = rows.Err()
	return
}

func (d *Dao) SelectProjectPath(ctx context.Context, projectId int64) (paths []*openmdl.PathAccess, err error) {
	rows, err := d.db.Query(ctx, _selectProjectPath, projectId)
	if err != nil {
		return
	}
	defer rows.Close()
	if err = xsql.ScanSlice(rows, &paths); err != nil {
		log.Error("ScanSlice Error: %v", err)
		return
	}
	err = rows.Err()
	return
}

func (d *Dao) UpdateProjectPath(ctx context.Context, update []*open.PathUpdate) (effected int64, err error) {
	if len(update) == 0 {
		return
	}
	var (
		appCaseSqls  = make([]string, 0, len(update))
		descCaseSqls = make([]string, 0, len(update))
		args         = make([]interface{}, 0)
		ids          []string
	)
	for _, v := range update {
		appCaseSqls = append(appCaseSqls, "WHEN id=? THEN ?")
		args = append(args, v.PathId, strings.Join(v.AppKey, openmdl.Comma))
		ids = append(ids, strconv.FormatInt(v.PathId, 10))
		descCaseSqls = append(descCaseSqls, "WHEN id=? THEN ?")
		args = append(args, v.PathId, v.Description)
	}
	var r sql.Result
	if r, err = d.db.Exec(ctx, fmt.Sprintf(_batchUpdateProjectPath, strings.Join(appCaseSqls, " "), strings.Join(descCaseSqls, " "), strings.Join(ids, ",")), args...); err != nil {
		log.Error("UpdateProjectPath error: %v", err)
		return
	}
	return r.RowsAffected()
}

func (d *Dao) SelectProjectInfos(ctx context.Context, projectName string, filter []int64, pn int64, ps int64) (projects []*openmdl.Project, err error) {
	var (
		sqls = make([]string, 0)
		args = make([]interface{}, 0)
		idIn = make([]string, 0)
	)
	sqls = append(sqls, "project_name LIKE ?")
	args = append(args, "%"+projectName+"%")
	if len(filter) != 0 {
		for _, v := range filter {
			idIn = append(idIn, "?")
			args = append(args, strconv.FormatInt(v, 10))
		}
		sqls = append(sqls, "id IN ("+strings.Join(idIn, openmdl.Comma)+")")
	}
	args = append(args, (pn-1)*ps, ps)
	rows, err := d.db.Query(ctx, fmt.Sprintf(_selectProjectInfos, strings.Join(sqls, " AND ")), args...)
	if err != nil {
		return
	}
	defer rows.Close()
	if err = xsql.ScanSlice(rows, &projects); err != nil {
		log.Error("ScanSlice Error: %v", err)
		return
	}
	err = rows.Err()
	return
}

func (d *Dao) CountProjectInfos(ctx context.Context, projectName string, filter []int64) (count int64, err error) {
	var (
		sqls = make([]string, 0)
		args = make([]interface{}, 0)
		idIn = make([]string, 0)
	)
	sqls = append(sqls, "project_name LIKE ?")
	args = append(args, "%"+projectName+"%")
	if len(filter) != 0 {
		for _, v := range filter {
			idIn = append(idIn, "?")
			args = append(args, strconv.FormatInt(v, 10))
		}
		sqls = append(sqls, "id IN ("+strings.Join(idIn, openmdl.Comma)+")")
	}
	row := d.db.QueryRow(ctx, fmt.Sprintf(_countProjectInfos, strings.Join(sqls, " AND ")), args...)
	if err != nil {
		return
	}
	tmp := int64(0)
	if err = row.Scan(&tmp); err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		log.Errorc(ctx, "CountProjectInfos error:%+v", err)
		return -1, err
	}
	return tmp, nil
}

func (d *Dao) UpdateProjectStatus(ctx context.Context, isActive bool, projectId int64) (err error) {
	if _, err = d.db.Exec(ctx, _activeProject, isActive, projectId); err != nil {
		return
	}
	return
}

func (d *Dao) SelectOpenToken(ctx context.Context, token string) (project *openmdl.Project, err error) {
	rows, err := d.db.Query(ctx, _selectToken, token)
	if err != nil {
		return
	}
	defer rows.Close()
	var list []*openmdl.Project
	if err = xsql.ScanSlice(rows, &list); err != nil {
		log.Errorc(ctx, "ScanSlice Error: %v", err)
		return
	}
	err = rows.Err()
	if len(list) == 1 {
		return list[0], err
	}
	return nil, err
}
