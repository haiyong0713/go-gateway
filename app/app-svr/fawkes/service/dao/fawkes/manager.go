package fawkes

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"go-common/library/database/sql"
	xsql "go-common/library/database/sql"
	"go-common/library/ecode"

	appmdl "go-gateway/app/app-svr/fawkes/service/model/app"
	mngmdl "go-gateway/app/app-svr/fawkes/service/model/manager"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"

	"github.com/pkg/errors"
)

const (
	_token    = "/v1/token"
	_role     = "/v1/node/role"
	_auth     = "/v1/auth"
	_treeApp  = "/v1/node/apptree"
	_bfsCache = "/x/admin/bfs/cache/purge"
	// sql
	_authUserCount        = `SELECT count(*) FROM auth_user WHERE %s`
	_authUserList         = `SELECT au.id,au.app_key,au.name,IFNULL(ui.nick_name,''),au.role,au.operator,unix_timestamp(au.mtime) FROM (auth_user as au LEFT JOIN user_info as ui ON au.name=ui.user_name) INNER JOIN app_attribute as aa ON aa.app_key=au.app_key WHERE au.id > 0 AND %s ORDER BY aa.is_host,au.id DESC LIMIT ?,?`
	_authUserListDistinct = `SELECT DISTINCT name FROM auth_user WHERE id > 0 AND %s`
	_authUserListByRole   = `SELECT au.id,au.app_key,au.name,ui.nick_name,au.role,au.operator,unix_timestamp(ui.mtime) FROM auth_user AS au, user_info AS ui WHERE au.name = ui.user_name AND au.app_key=? AND au.role=? ORDER BY au.id`
	_setAuthUser          = `INSERT INTO auth_user (app_key,name,role,operator) VALUES(?,?,?,?) ON DUPLICATE KEY UPDATE role=?,operator=?`
	_delAuthUser          = `DELETE FROM auth_user WHERE id=?`
	_auth2                = `SELECT name,role FROM auth_user WHERE name=? %s`
	_authRole             = `SELECT id,name,ename,value,state,unix_timestamp(mtime) FROM auth_role ORDER BY value`
	_supervisorRole       = `SELECT id,name,role,operator,unix_timestamp(mtime) FROM auth_supervisor WHERE name=? AND role=?`
	_supervisorRoles      = `SELECT id,name,role,operator,unix_timestamp(mtime) FROM auth_supervisor WHERE role=?`
	_authRoleByVal        = `SELECT id,name,ename,value,state,unix_timestamp(mtime) FROM auth_role WHERE value=? ORDER BY value`
	_authRoleApplyCount   = `SELECT count(*) FROM auth_role_apply WHERE app_key=? %s`
	_authRoleApplyList    = `SELECT a.id,a.app_key,a.name,a.role,a.operator,a.state,unix_timestamp(a.mtime),unix_timestamp(a.ctime),IFNULL(u.role,0) as cur_role FROM auth_role_apply AS a LEFT JOIN auth_user as u ON a.name=u.name AND a.app_key=u.app_key
WHERE a.app_key=? %s ORDER BY a.id DESC LIMIT ?,?`
	_addAuthRoleApply    = `INSERT auth_role_apply (app_key,name,role,operator) VALUES(?,?,?,?)`
	_upAuthRoleApply     = `UPDATE auth_role_apply SET state=?,operator=? WHERE id=? AND app_key=?`
	_upAuthUserNickName  = `UPDATE auth_user SET nick_name=? WHERE name=?`
	_authRoleApplyByName = `SELECT a.id,a.app_key,a.name,a.role,a.operator,a.state,unix_timestamp(a.mtime),unix_timestamp(a.ctime),IFNULL(u.role,0) 
FROM auth_role_apply AS a LEFT JOIN auth_user as u ON a.name=u.name AND a.app_key=u.app_key WHERE a.app_key=? AND a.name=? ORDER BY a.id DESC`
	_authRoleApplyByID = `SELECT a.id,a.app_key,a.name,a.role,a.operator,state,unix_timestamp(a.mtime),unix_timestamp(a.ctime),IFNULL(u.role,0)  
FROM auth_role_apply AS a LEFT JOIN auth_user as u ON a.name=u.name AND a.app_key=u.app_key WHERE a.app_key=? AND a.id=? ORDER BY a.id DESC`
	_authUserDistinctByName = `SELECT DISTINCT name FROM auth_user WHERE name IN (%v)`

	_eventApplyList = `SELECT id,app_key,event,target_id,applicant,operator,state,unix_timestamp(mtime),unix_timestamp(ctime) 
FROM event_apply WHERE app_key=? AND event=? AND target_id=? AND state=?`
	_addEventApply = `INSERT event_apply (app_key,event,applicant,operator,target_id) VALUES(?,?,?,?,?)`
	_upEventApply  = `UPDATE event_apply SET state=?,operator=? WHERE app_key=? AND event=? AND target_id=? AND state=0`
	_userNameList  = `SELECT id,user_id,user_name,nick_name,unix_timestamp(mtime) FROM user_info %v ORDER BY id limit ?,?`
	_userName      = `SELECT id,user_name,nick_name,avatar,unix_timestamp(mtime) FROM user_info WHERE user_name=?`
	_userInfoList  = `SELECT id,user_id,user_name,nick_name,avatar,unix_timestamp(mtime) FROM user_info WHERE user_name IN (%v)`
	_setUserName   = `INSERT INTO user_info (user_name,nick_name) VALUES(?,?) ON DUPLICATE KEY UPDATE nick_name=?`
	_batchSetUser  = "INSERT INTO user_info (user_id,user_name,nick_name,mobile,avatar) VALUES %s ON DUPLICATE KEY UPDATE user_id=VALUES(user_id),nick_name=VALUES(nick_name),mobile=VALUES(mobile),avatar=VALUES(avatar),mtime=NOW()"
)

// TreeToken get tree token.
func (d *Dao) TreeToken(c context.Context) (token string, err error) {
	var pb []byte
	if pb, err = json.Marshal(&mngmdl.ParamsToken{User: d.c.Easyst.User, Platform: d.c.Easyst.Platform}); err != nil {
		log.Error("%v", err)
		return
	}
	req, _ := http.NewRequest("POST", d.treeToken, strings.NewReader(string(pb)))
	req.Header.Set("Content-Type", "application/json")
	var res struct {
		Code    int                 `json:"code"`
		Data    *mngmdl.ResultToken `json:"data"`
		Message string              `json:"message"`
		Status  int                 `json:"status"`
	}
	if err = d.httpClient.Do(c, req, &res); err != nil {
		log.Error("%v params(%v)", err, string(pb))
		return
	}
	if res.Code != 90000 || res.Status != 200 {
		err = errors.Wrap(ecode.Int(res.Code), d.treeToken+"?"+string(pb))
		return
	}
	token = res.Data.Token
	return
}

// TreeRole get tree role.
func (d *Dao) TreeRole(c context.Context, treePath, token string) (role []*mngmdl.ResultRole, err error) {
	var req *http.Request
	if req, err = d.httpClient.NewRequest("GET", d.treeRole+"/"+treePath, "", nil); err != nil {
		return
	}
	req.Header.Set("X-Authorization-Token", token)
	var res struct {
		Code    int                  `json:"code"`
		Data    []*mngmdl.ResultRole `json:"data"`
		Message string               `json:"message"`
		Status  int                  `json:"status"`
	}
	if err = d.httpClient.Do(c, req, &res); err != nil {
		log.Error("%v token(%v)", err, token)
		return
	}
	if res.Code != 90000 || res.Status != 200 {
		err = errors.Wrap(ecode.Int(res.Code), d.treeRole+"/"+treePath+"/"+token)
		return
	}
	role = res.Data
	return
}

// TreeAuth get tree auth.
func (d *Dao) TreeAuth(c context.Context, sessionID string) (token string, err error) {
	var req *http.Request
	if req, err = d.httpClient.NewRequest("GET", d.treeAuth, "", nil); err != nil {
		return
	}
	req.Header.Set("Cookie", fmt.Sprintf("_AJSESSIONID=%s", sessionID))
	var res struct {
		Code    int                 `json:"code"`
		Data    *mngmdl.ResultToken `json:"data"`
		Message string              `json:"message"`
		Status  int                 `json:"status"`
	}
	if err = d.httpClient.Do(c, req, &res); err != nil {
		log.Error("%v _AJSESSIONID(%v)", err, sessionID)
		return
	}
	if res.Code != 90000 || res.Status != 200 {
		err = errors.Wrap(ecode.Int(res.Code), d.treeAuth+"/"+sessionID)
		return
	}
	token = res.Data.Token
	return
}

// TreeApp get tree list.
func (d *Dao) TreeApp(c context.Context, token string) (role map[string]*mngmdl.ResultTree, err error) {
	var req *http.Request
	if req, err = d.httpClient.NewRequest("GET", d.treeApp, "", nil); err != nil {
		return
	}
	req.Header.Set("X-Authorization-Token", token)
	var res struct {
		Code    int                           `json:"code"`
		Data    map[string]*mngmdl.ResultTree `json:"data"`
		Message string                        `json:"message"`
		Status  int                           `json:"status"`
	}
	if err = d.httpClient.Do(c, req, &res); err != nil {
		log.Error("%v token(%v)", err, token)
		return
	}
	if res.Code != 90000 || res.Status != 200 {
		err = errors.Wrap(ecode.Int(res.Code), d.treeApp+"/"+token)
		return
	}
	role = res.Data
	return
}

// AuthUserCount get app user count.
func (d *Dao) AuthUserCount(c context.Context, appKeys []string, role, filterKey string) (count int, err error) {
	var (
		sqls []string
		args []interface{}
	)
	if len(appKeys) > 0 {
		var sqlTmp []string
		for _, appKey := range appKeys {
			args = append(args, appKey)
			sqlTmp = append(sqlTmp, "?")
		}
		sqls = append(sqls, fmt.Sprintf(" app_key IN (%s) ", strings.Join(sqlTmp, ",")))
	}
	if role != "" {
		sqls = append(sqls, "role = ?")
		args = append(args, role)
	}
	if filterKey != "" {
		sqls = append(sqls, "name LIKE ?")
		args = append(args, "%"+filterKey+"%")
	}
	row := d.db.QueryRow(c, fmt.Sprintf(_authUserCount, strings.Join(sqls, " AND ")), args...)
	if err = row.Scan(&count); err != nil {
		if err == xsql.ErrNoRows {
			err = nil
		} else {
			log.Error("AuthUserCount %v", err)
		}
	}
	return
}

// AuthUserList get app user list.
func (d *Dao) AuthUserList(c context.Context, appKeys []string, role, filterKey string, pn, ps int) (res []*mngmdl.User, err error) {
	var (
		sqls []string
		args []interface{}
	)
	if len(appKeys) > 0 {
		var sqlTmp []string
		for _, appKey := range appKeys {
			args = append(args, appKey)
			sqlTmp = append(sqlTmp, "?")
		}
		sqls = append(sqls, fmt.Sprintf(" au.app_key IN (%s) ", strings.Join(sqlTmp, ",")))
	}
	if role != "" {
		sqls = append(sqls, " au.role = ? ")
		args = append(args, role)
	}
	if filterKey != "" {
		sqls = append(sqls, " au.name LIKE ? ")
		args = append(args, "%"+filterKey+"%")
	}
	args = append(args, (pn-1)*ps, ps)
	rows, err := d.db.Query(c, fmt.Sprintf(_authUserList, strings.Join(sqls, " AND ")), args...)
	if err != nil {
		log.Error("AuthUserList %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &mngmdl.User{}
		if err = rows.Scan(&re.ID, &re.AppKey, &re.Name, &re.NickName, &re.Role, &re.Operator, &re.MTime); err != nil {
			log.Error("AuthUserList %v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

// AuthUserListDistinct get distinct user name.
func (d *Dao) AuthUserNamesDistinct(c context.Context, appKey, role string) (res []string, err error) {
	var (
		sqls []string
		args []interface{}
	)
	if appKey != "" {
		var placeholders []string
		appKeys := strings.Split(appKey, ",")
		for _, ak := range appKeys {
			args = append(args, ak)
			placeholders = append(placeholders, "?")
		}
		sqls = append(sqls, fmt.Sprintf("app_key in (%v)", strings.Join(placeholders, ",")))
	}
	if role != "" {
		var placeholders []string
		roles := strings.Split(role, ",")
		for _, r := range roles {
			args = append(args, r)
			placeholders = append(placeholders, "?")
		}
		sqls = append(sqls, fmt.Sprintf("role in (%v)", strings.Join(placeholders, ",")))
	}
	rows, err := d.db.Query(c, fmt.Sprintf(_authUserListDistinct, strings.Join(sqls, " AND ")), args...)
	if err != nil {
		log.Error("AuthUserNamesDistinct %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &mngmdl.User{}
		if err = rows.Scan(&re.Name); err != nil {
			log.Error("AuthUserNamesDistinct %v", err)
			return
		}
		res = append(res, re.Name)
	}
	err = rows.Err()
	return
}

// AuthUserListByRole get auth user by role
func (d *Dao) AuthUserListByRole(c context.Context, appKey string, role int) (res []*mngmdl.User, err error) {
	rows, err := d.db.Query(c, _authUserListByRole, appKey, role)
	if err != nil {
		log.Error("AuthUserListByRole %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &mngmdl.User{}
		if err = rows.Scan(&re.ID, &re.AppKey, &re.Name, &re.NickName, &re.Role, &re.Operator, &re.MTime); err != nil {
			log.Error("AuthUserListByRole %v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

// TxSetAuthUser set app user.
func (d *Dao) TxSetAuthUser(tx *xsql.Tx, appKey, userName, uname string, role int) (count int64, err error) {
	res, err := tx.Exec(_setAuthUser, appKey, uname, role, userName, role, userName)
	if err != nil {
		log.Error("TxSetAuthUser %v", err)
		return
	}
	return res.RowsAffected()
}

// BatchSetUser set user info.
func (d *Dao) BatchSetUser(ctx context.Context, users []*appmdl.User) (err error) {
	var (
		sqls = make([]string, 0, len(users))
		args = make([]interface{}, 0, len(users))
	)
	if len(users) == 0 {
		return
	}
	for _, v := range users {
		sqls = append(sqls, "(?,?,?,?,?)")
		args = append(args, v.UserID, v.Alias, v.Name, v.Mobile, v.Avatar)
	}
	if _, err = d.db.Exec(ctx, fmt.Sprintf(_batchSetUser, strings.Join(sqls, ",")), args...); err != nil {
		log.Error("SetTribePackVersionTx error: %v", err)
	}
	return
}

// TxDelAuthUser del app user.
func (d *Dao) TxDelAuthUser(tx *xsql.Tx, id int64) (count int64, err error) {
	res, err := tx.Exec(_delAuthUser, id)
	if err != nil {
		log.Error("TxDelAuthUser %v", err)
		return
	}
	return res.RowsAffected()
}

// AuthUser get auth by app.
func (d *Dao) AuthUser(c context.Context, appKey, userName string) (re *mngmdl.ResultRole, err error) {
	row := d.db.QueryRow(c, fmt.Sprintf(_auth2, "AND app_key=?"), userName, appKey)
	re = &mngmdl.ResultRole{}
	if err = row.Scan(&re.User, &re.Role); err != nil {
		if err == xsql.ErrNoRows {
			err = nil
			re = nil
		} else {
			log.Error("AuthUser %v", err)
		}
	}
	return
}

// AuthFawkesRoles get auth by all app.
func (d *Dao) AuthFawkesRoles(c context.Context, userName string) (res []*mngmdl.ResultRole, err error) {
	rows, err := d.db.Query(c, fmt.Sprintf(_auth2, ""), userName)
	if err != nil {
		log.Error("AuthUserList %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &mngmdl.ResultRole{}
		if err = rows.Scan(&re.User, &re.Role); err != nil {
			log.Error("AuthUserList %v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

// AuthRole get role.
func (d *Dao) AuthRole(c context.Context) (res []*mngmdl.Role, err error) {
	rows, err := d.db.Query(c, _authRole)
	if err != nil {
		log.Error("AuthRole %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &mngmdl.Role{}
		if err = rows.Scan(&re.ID, &re.Name, &re.EName, &re.Value, &re.State, &re.MTime); err != nil {
			log.Error("AuthRole %v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

// AuthSupervisor get supervisor.
func (d *Dao) AuthSupervisor(c context.Context, uname string) (res []*mngmdl.SupervisorRole, err error) {
	rows, err := d.db.Query(c, _supervisorRole, uname, 100)
	if err != nil {
		log.Error("AuthSupervisor %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &mngmdl.SupervisorRole{}
		if err = rows.Scan(&re.ID, &re.Name, &re.Role, &re.Operator, &re.MTime); err != nil {
			log.Error("AuthSupervisor %v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

// AuthSupervisor get supervisors.
func (d *Dao) AuthSupervisors(c context.Context) (res []*mngmdl.SupervisorRole, err error) {
	rows, err := d.db.Query(c, _supervisorRoles, 100)
	if err != nil {
		log.Error("AuthSupervisors %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &mngmdl.SupervisorRole{}
		if err = rows.Scan(&re.ID, &re.Name, &re.Role, &re.Operator, &re.MTime); err != nil {
			log.Error("AuthSupervisors %v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

// AuthRoleByVal get role by value
func (d *Dao) AuthRoleByVal(c context.Context, val int) (re *mngmdl.Role, err error) {
	row := d.db.QueryRow(c, _authRoleByVal, val)
	re = &mngmdl.Role{}
	if err = row.Scan(&re.ID, &re.Name, &re.EName, &re.Value, &re.State, &re.MTime); err != nil {
		if err == xsql.ErrNoRows {
			err = nil
			re = nil
		} else {
			log.Error("AuthRoleApplyByID %v", err)
		}
	}
	return
}

// AuthRoleApplyByID get role apply info by id
func (d *Dao) AuthRoleApplyByID(c context.Context, appKey string, id int) (re *mngmdl.RoleApply, err error) {
	row := d.db.QueryRow(c, _authRoleApplyByID, appKey, id)
	re = &mngmdl.RoleApply{}
	if err = row.Scan(&re.ID, &re.AppKey, &re.Name, &re.Role, &re.Operator, &re.State, &re.MTime, &re.CTime, &re.CurRole); err != nil {
		if err == xsql.ErrNoRows {
			err = nil
			re = nil
		} else {
			log.Error("AuthRoleApplyByID %v", err)
		}
	}
	return
}

// AuthRoleApply get user role apply info
func (d *Dao) AuthRoleApply(c context.Context, appKey, userName string) (re *mngmdl.RoleApply, err error) {
	row := d.db.QueryRow(c, _authRoleApplyByName, appKey, userName)
	re = &mngmdl.RoleApply{}
	if err = row.Scan(&re.ID, &re.AppKey, &re.Name, &re.Role, &re.Operator, &re.State, &re.MTime, &re.CTime, &re.CurRole); err != nil {
		if err == xsql.ErrNoRows {
			err = nil
			re = nil
		} else {
			log.Error("AuthRoleApply %v", err)
		}
	}
	return
}

// AuthRolesApply get user role apply info
func (d *Dao) AuthRolesApply(c context.Context, appKey, userName string) (res []*mngmdl.RoleApply, err error) {
	rows, err := d.db.Query(c, _authRoleApplyByName, appKey, userName)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &mngmdl.RoleApply{}
		if err = rows.Scan(&re.ID, &re.AppKey, &re.Name, &re.Role, &re.Operator, &re.State, &re.MTime, &re.CTime, &re.CurRole); err != nil {
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

// AuthRoleApplyCount get role apply count.
func (d *Dao) AuthRoleApplyCount(c context.Context, appKey string, state int) (count int, err error) {
	var (
		sqlAdd string
		args   []interface{}
	)
	args = append(args, appKey)
	if state != mngmdl.AuthRoleApplyEmptyState {
		sqlAdd += "AND state=?"
		args = append(args, state)
	}
	row := d.db.QueryRow(c, fmt.Sprintf(_authRoleApplyCount, sqlAdd), args...)
	if err = row.Scan(&count); err != nil {
		if err == xsql.ErrNoRows {
			err = nil
		} else {
			log.Error("AuthRoleApplyCount %v", err)
		}
	}
	return
}

// AuthRoleApplyList get role apply list.
func (d *Dao) AuthRoleApplyList(c context.Context, appKey string, state, pn, ps int) (res []*mngmdl.RoleApply, err error) {
	var (
		sqlAdd string
		args   []interface{}
	)
	args = append(args, appKey)
	if state != mngmdl.AuthRoleApplyEmptyState {
		sqlAdd += "AND a.state=?"
		args = append(args, state)
	}
	args = append(args, (pn-1)*ps, ps)
	rows, err := d.db.Query(c, fmt.Sprintf(_authRoleApplyList, sqlAdd), args...)
	if err != nil {
		log.Error("AuthRoleApplyList %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &mngmdl.RoleApply{}
		if err = rows.Scan(&re.ID, &re.AppKey, &re.Name, &re.Role, &re.Operator, &re.State, &re.MTime, &re.CTime, &re.CurRole); err != nil {
			log.Error("AuthRoleApplyList %v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

// TxAddAuthRoleApply set role apply info.
func (d *Dao) TxAddAuthRoleApply(tx *xsql.Tx, appKey, userName, operator string, role int) (count int64, err error) {
	res, err := tx.Exec(_addAuthRoleApply, appKey, userName, role, operator)
	if err != nil {
		log.Error("TxAddAuthRoleApply tx.Exec error(%v)", err)
		return
	}
	return res.RowsAffected()
}

// TxUpdateAuthRoleApply update role apply info.
func (d *Dao) TxUpdateAuthRoleApply(tx *xsql.Tx, appKey, operator string, id, state int) (r int64, err error) {
	res, err := tx.Exec(_upAuthRoleApply, state, operator, id, appKey)
	if err != nil {
		log.Error("d.TxUpdateAuthRoleApply tx.Exec error(%v)", err)
		return
	}
	r, err = res.RowsAffected()
	return
}

// EventApplyList get event apply list.
func (d *Dao) EventApplyList(c context.Context, appKey string, event, targetID, state int) (res []*mngmdl.EventApply, err error) {
	rows, err := d.db.Query(c, _eventApplyList, appKey, event, targetID, state)
	if err != nil {
		log.Error("EventApplyList %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &mngmdl.EventApply{}
		if err = rows.Scan(&re.ID, &re.AppKey, &re.Event, &re.TargetID, &re.Applicant, &re.Operator, &re.State, &re.MTime, &re.CTime); err != nil {
			log.Error("EventApplyList %v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

// TxAddEventApply set event apply info.
func (d *Dao) TxAddEventApply(tx *xsql.Tx, appKey, userName, operator string, event, targetID int) (count int64, err error) {
	res, err := tx.Exec(_addEventApply, appKey, event, userName, operator, targetID)
	if err != nil {
		log.Error("TxAddEventApply tx.Exec error(%v)", err)
		return
	}
	return res.RowsAffected()
}

// TxUpdateEventApply update event apply info.
func (d *Dao) TxUpdateEventApply(tx *xsql.Tx, appKey, operator string, event, targetID, state int) (r int64, err error) {
	res, err := tx.Exec(_upEventApply, state, operator, appKey, event, targetID)
	if err != nil {
		log.Error("d.TxUpdateEventApply tx.Exec error(%v)", err)
		return
	}
	r, err = res.RowsAffected()
	return
}

// refresh bfs files
func (d *Dao) BfsRefreshCDN(urls string) (err error) {
	var (
		req    *http.Request
		res    *mngmdl.BFSCDNRefreshRes
		params = url.Values{}
	)
	if urls != "" {
		params.Set("urls", urls)
	}
	// Request AccessToken
	if req, err = http.NewRequest(http.MethodPost, d.bfsCache+"?"+params.Encode(), nil); err != nil {
		log.Error("d.BfsRefreshCDN call http.NewRequest error(%v)", err)
		return
	}
	req.Header.Add("content-type", "application/json")
	if err = d.httpClient.Do(context.Background(), req, &res); err != nil {
		log.Error("BfsRefreshCDN error(%v)", err)
		return
	}
	if res.Code != 0 {
		err = errors.Wrap(ecode.Int(res.Code), res.Message)
	}
	return
}

func (d *Dao) AuthNickNameSet(tx *xsql.Tx, userName, nickName string) (err error) {
	_, err = tx.Exec(_upAuthUserNickName, nickName, userName)
	if err != nil {
		log.Error("AuthNickNameSet tx.Exec error(%v)", err)
		return
	}
	return
}

// UserNameList authuser get user name list.
func (d *Dao) UserNameList(c context.Context, userName string, ps, pn int) (res []*mngmdl.UserInfo, err error) {
	var (
		sqlAdd string
		args   []interface{}
	)
	if userName != "" {
		sqlAdd += "WHERE (user_name LIKE ? OR nick_name LIKE ?)"
		args = append(args, "%"+userName+"%", "%"+userName+"%")
	}
	args = append(args, (pn-1)*ps, ps)
	rows, err := d.db.Query(c, fmt.Sprintf(_userNameList, sqlAdd), args...)
	if err != nil {
		log.Error("Dao UserNameList: %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &mngmdl.UserInfo{}
		if err = rows.Scan(&re.ID, &re.UserId, &re.Name, &re.NickName, &re.MTime); err != nil {
			log.Error("UserNameList Scan: %v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

// UserName authuser get user name
func (d *Dao) UserName(c context.Context, userName string) (res *mngmdl.UserInfo, err error) {
	res = &mngmdl.UserInfo{}
	err = d.db.QueryRow(c, _userName, userName).Scan(&res.ID, &res.Name, &res.NickName, &res.Avatar, &res.MTime)
	if err != nil {
		if err == sql.ErrNoRows {
			err = nil
			res = nil
		} else {
			log.Error("Dao username: %v", err)
		}
	}
	return
}

func (d *Dao) UserInfoList(c context.Context, usersName []string) (res []*mngmdl.UserInfo, err error) {
	var (
		sqls []string
		args []interface{}
	)
	if len(usersName) < 1 {
		log.Warn("users is empty")
		return
	}
	for _, userName := range usersName {
		sqls = append(sqls, "?")
		args = append(args, userName)
	}
	rows, err := d.db.Query(c, fmt.Sprintf(_userInfoList, strings.Join(sqls, ",")), args...)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &mngmdl.UserInfo{}
		if err = rows.Scan(&re.ID, &re.UserId, &re.Name, &re.NickName, &re.NickName, &re.MTime); err != nil {
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

// UserNameSet user name set
func (d *Dao) TxUserNameSet(tx *xsql.Tx, userName, nickName string) (err error) {
	_, err = tx.Exec(_setUserName, userName, nickName, nickName)
	if err != nil {
		log.Error("UserNameSet tx.Exec error(%v)", err)
		return
	}
	return
}

func (d *Dao) AuthDistinctUserInfo(c context.Context, usernames []string) (res map[string]struct{}, err error) {
	if len(usernames) < 1 {
		log.Warnc(c, "name is nil")
		return
	}
	var (
		sqls = make([]string, 0, len(usernames))
		args = make([]interface{}, 0, len(usernames))
	)
	res = make(map[string]struct{})
	for _, user := range usernames {
		sqls = append(sqls, "?")
		args = append(args, user)
	}
	rows, err := d.db.Query(c, fmt.Sprintf(_authUserDistinctByName, strings.Join(sqls, ",")), args...)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &mngmdl.User{}
		if err = rows.Scan(&re.Name); err != nil {
			return
		}
		res[re.Name] = struct{}{}
	}
	err = rows.Err()
	return
}
