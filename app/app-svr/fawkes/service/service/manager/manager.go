package manager

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go-common/library/database/sql"

	"go-gateway/app/app-svr/fawkes/service/model"
	appmdl "go-gateway/app/app-svr/fawkes/service/model/app"
	mngmdl "go-gateway/app/app-svr/fawkes/service/model/manager"
	wf "go-gateway/app/app-svr/fawkes/service/model/workflow"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

// TreeAuth get tree auth.
func (s *Service) TreeAuth(c context.Context, appKey, userName string) (re *mngmdl.ResultRole, err error) {
	for _, wu := range s.c.Whitelist {
		if wu == userName {
			re = &mngmdl.ResultRole{
				User:   userName,
				Role:   1,
				Leader: 1,
			}
			return
		}
	}
	var app *appmdl.APP
	if app, err = s.fkDao.AppPass(c, appKey); err != nil {
		log.Error("%v", err)
		return
	}
	if app == nil {
		return
	}
	var (
		token    string
		treePath = app.TreePath
	)
	if token, err = s.fkDao.TreeToken(c); err != nil {
		log.Error("%v", err)
		return
	}
	if token == "" {
		log.Warn("tree_path(%v) uaername(%v) get token is empty", treePath, userName)
		return
	}
	var roles []*mngmdl.ResultRole
	if roles, err = s.fkDao.TreeRole(c, treePath, token); err != nil {
		log.Error("%v", err)
		return
	}
	for _, role := range roles {
		if role.User == userName {
			re = &mngmdl.ResultRole{}
			*re = *role
		}
	}
	return
}

// TreeAuths get tree auth by appkeys.
func (s *Service) TreeAuths(c context.Context, appKeys []string, userName string) (res map[string]*mngmdl.ResultRole, err error) {
	for _, wu := range s.c.Whitelist {
		if wu == userName {
			res = make(map[string]*mngmdl.ResultRole)
			for _, appKey := range appKeys {
				re := &mngmdl.ResultRole{
					User:   userName,
					Role:   1,
					Leader: 1,
				}
				res[appKey] = re
			}
			return
		}
	}
	var apps []*appmdl.APP
	if apps, err = s.fkDao.AppsPass(c, appKeys, userName, 0); err != nil {
		log.Error("%v", err)
		return
	}
	for _, app := range apps {
		if app == nil {
			return
		}
		var (
			token    string
			treePath = app.TreePath
		)
		if token, err = s.fkDao.TreeToken(c); err != nil {
			log.Error("%v", err)
			return
		}
		if token == "" {
			log.Warn("tree_path(%v) uaername(%v) get token is empty", treePath, userName)
			return
		}
		var roles []*mngmdl.ResultRole
		if roles, err = s.fkDao.TreeRole(c, treePath, token); err != nil {
			log.Error("%v", err)
			return
		}
		res = make(map[string]*mngmdl.ResultRole)
		for _, role := range roles {
			if role.User == userName {
				re := &mngmdl.ResultRole{}
				*re = *role
				res[app.AppKey] = re
			}
		}
	}
	return
}

// TreeList get tree list by user.
func (s *Service) TreeList(c context.Context, sessionID string) (res map[string]*mngmdl.ResultTree, err error) {
	var token string
	if token, err = s.fkDao.TreeAuth(c, sessionID); err != nil {
		log.Error("%v", err)
		return
	}
	if res, err = s.fkDao.TreeApp(c, token); err != nil {
		log.Error("%v", err)
	}
	return
}

// AuthUserList get user list.
func (s *Service) AuthUserList(c context.Context, appKey, role, userName, filterKey string, pn, ps int) (res *mngmdl.UserList, err error) {
	var hostAppKey string
	if hostAppKey, err = s.AppHost(c, appKey); err != nil {
		log.Errorc(c, "AppHost error %v", err)
		return
	}
	var (
		count   int
		appKeys []string
	)
	if appKey != "" {
		appKeys = append(appKeys, appKey)
	}
	if hostAppKey != "" && hostAppKey != appKey {
		appKeys = append(appKeys, hostAppKey)
	}
	if count, err = s.fkDao.AuthUserCount(c, appKeys, role, filterKey); err != nil {
		log.Error("%v", err)
		return
	}
	if count == 0 {
		return
	}
	var users []*mngmdl.User
	if users, err = s.fkDao.AuthUserList(c, appKeys, role, filterKey, pn, ps); err != nil {
		log.Error("%v", err)
		return
	}
	res = &mngmdl.UserList{
		PageInfo: &model.PageInfo{
			Total: count,
			Pn:    pn,
			Ps:    ps,
		},
		Items: users,
	}
	return
}

// AuthUserListByRole get user by role
func (s *Service) AuthUserListByRole(c context.Context, appKey string, role int) (res []*mngmdl.User, err error) {
	var (
		users    []*mngmdl.User
		employee []*mngmdl.User
	)
	if users, err = s.fkDao.AuthUserListByRole(c, appKey, role); err != nil {
		log.Error("%v", err)
		return
	}
	expireTime := time.Now().AddDate(0, -1, 0).Unix()
	for _, user := range users {
		if expireTime < user.MTime {
			employee = append(employee, user)
		}
	}
	res = employee
	return
}

// AuthUserSet set user.
func (s *Service) AuthUserSet(c context.Context, appKey, userName, uname string, role int) (err error) {
	var tx *sql.Tx
	if tx, err = s.fkDao.BeginTran(c); err != nil {
		log.Error("s.fkDao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	if _, err = s.fkDao.TxSetAuthUser(tx, appKey, userName, uname, role); err != nil {
		log.Error("%v", err)
	}
	return
}

// AuthUserDel del user.
func (s *Service) AuthUserDel(c context.Context, id int64, userName string) (err error) {
	var tx *sql.Tx
	if tx, err = s.fkDao.BeginTran(c); err != nil {
		log.Error("s.fkDao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	if _, err = s.fkDao.TxDelAuthUser(tx, id); err != nil {
		log.Error("%v", err)
	}
	return
}

// AuthUser get auth.
func (s *Service) AuthUser(c context.Context, appKey, userName string) (res *mngmdl.ResultRole, err error) {
	if res, err = s.fkDao.AuthUser(c, appKey, userName); err != nil {
		log.Errorc(c, "AuthUser error %v", err)
		return
	}
	// 降级处理，应用组权限共享，查找宿主app的权限信息
	if res == nil {
		var hostAppKey string
		if hostAppKey, err = s.AppHost(c, appKey); err != nil {
			log.Errorc(c, "appHost error %v", err)
			return
		}
		if hostAppKey == "" {
			return
		}
		if res, err = s.fkDao.AuthUser(c, hostAppKey, userName); err != nil {
			log.Errorc(c, "AuthUser error %v", err)
		}
	}
	return
}

// AppHost 查找宿主app
func (s *Service) AppHost(c context.Context, appKey string) (hostAppKey string, err error) {
	var app *appmdl.APP
	if app, err = s.fkDao.AppPass(c, appKey); err != nil {
		log.Errorc(c, "AppPass error %v", err)
		return
	}
	if app == nil {
		return
	}
	if hostAppKey, err = s.fkDao.AppHost(c, app.DataCenterAppID, app.Platform); err != nil {
		log.Errorc(c, "AppHost error %v", err)
		return
	}
	return
}

// AuthRole get role.
func (s *Service) AuthRole(c context.Context) (res []*mngmdl.Role, err error) {
	if res, err = s.fkDao.AuthRole(c); err != nil {
		log.Error("%v", err)
	}
	return
}

// AuthSupervisor get role.
func (s *Service) AuthSupervisor(c context.Context, userName string) (res []*mngmdl.SupervisorRole, err error) {
	if res, err = s.fkDao.AuthSupervisor(c, userName); err != nil {
		log.Error("%v", err)
	}
	return
}

// AuthSupervisorRole get role.
func (s *Service) AuthSupervisorRole(c context.Context, userName string) (res *mngmdl.FawkesUser, err error) {
	var (
		supervisorRoles []*mngmdl.SupervisorRole
		fawkesRoles     []*mngmdl.ResultRole
		user            *mngmdl.UserInfo
	)
	res = &mngmdl.FawkesUser{UserName: userName}
	if supervisorRoles, err = s.AuthSupervisor(c, userName); err != nil {
		log.Error("%v", err)
		return
	}
	res.SupervisorRoles = supervisorRoles
	if fawkesRoles, err = s.fkDao.AuthFawkesRoles(c, userName); err != nil {
		log.Error("%v", err)
		return
	}
	res.FawkesRoles = fawkesRoles
	if user, err = s.fkDao.UserName(c, userName); err != nil {
		log.Error("%v", err)
		return
	}
	if user == nil {
		return
	}
	res.NickName = user.NickName
	res.Avatar = user.Avatar
	return
}

// AuthAdminRole verify that the user is the administrator of the App
func (s *Service) AuthAdminRole(c context.Context, appKey, userName string) (hasAdminAuth bool) {
	var (
		role           *mngmdl.ResultRole
		supervisorRole []*mngmdl.SupervisorRole
		err            error
	)
	// 超管权限
	if supervisorRole, err = s.AuthSupervisor(c, userName); err != nil {
		hasAdminAuth = false
		return
	}
	if len(supervisorRole) > 0 {
		hasAdminAuth = true
		return
	}
	// App 管理员权限
	if role, err = s.AuthUser(c, appKey, userName); err != nil || role == nil {
		hasAdminAuth = false
		return
	}
	if role.Role == mngmdl.RoleAdmin {
		hasAdminAuth = true
	}
	return
}

// AuthRoleApply get user role apply.
func (s *Service) AuthRoleApply(c context.Context, appKey, userName string) (res *mngmdl.RoleApply, err error) {
	if res, err = s.fkDao.AuthRoleApply(c, appKey, userName); err != nil {
		log.Error("%v", err)
	}
	return
}

// AuthRoleApplyList get role apply list.
func (s *Service) AuthRoleApplyList(c context.Context, appKey string, state, pn, ps int) (res *mngmdl.RoleApplyList, err error) {
	var count int
	if count, err = s.fkDao.AuthRoleApplyCount(c, appKey, state); err != nil {
		log.Error("%v", err)
		return
	}
	if count == 0 {
		return
	}
	var applies []*mngmdl.RoleApply
	if applies, err = s.fkDao.AuthRoleApplyList(c, appKey, state, pn, ps); err != nil {
		log.Error("%v", err)
		return
	}
	res = &mngmdl.RoleApplyList{
		PageInfo: &model.PageInfo{
			Total: count,
			Pn:    pn,
			Ps:    ps,
		},
		Items: applies,
	}
	return
}

// AuthRoleApplyAdd add role apply.
func (s *Service) AuthRoleApplyAdd(c context.Context, appKey, userName, operator string, role int) (err error) {
	var (
		tx       *sql.Tx
		roleInfo *mngmdl.Role
		appInfo  *appmdl.APP
	)
	if tx, err = s.fkDao.BeginTran(c); err != nil {
		log.Error("s.fkDao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	if _, err = s.fkDao.TxAddAuthRoleApply(tx, appKey, userName, operator, role); err != nil {
		log.Error("%v", err)
		return
	}
	if roleInfo, err = s.fkDao.AuthRoleByVal(c, role); err != nil {
		log.Error("%v", err)
		return
	}
	if appInfo, err = s.fkDao.AppPass(c, appKey); err != nil {
		log.Error("%v", err)
		return
	}
	// 通知申请人
	_ = s.fkDao.WechatMessageNotify(fmt.Sprintf("\"%s(%s)\"的管理员 \"%s\" 已收到了您的【%s】权限申请, 请耐心等待审核结果。",
		appInfo.Name, appInfo.AppKey, operator, roleInfo.Name), userName, s.c.Comet.FawkesAppID)
	// 通知管理员
	_ = s.fkDao.WechatCardMessageNotify(
		"用户权限申请提醒",
		fmt.Sprintf("%s 提交了一个权限申请\n应用：%s(%s)\n权限：%s\n审核员：%s", userName, appInfo.Name, appInfo.AppKey, roleInfo.Name, operator),
		fmt.Sprintf("http://fawkes.bilibili.co/#/role/role-applylist?app_key=%v", appInfo.AppKey),
		"",
		operator,
		s.c.Comet.FawkesAppID)
	return
}

// AuthRoleApplyPass pass role apply.
func (s *Service) AuthRoleApplyPass(c context.Context, appKey, userName string, id int) (err error) {
	var (
		tx       *sql.Tx
		appInfo  *appmdl.APP
		roleInfo *mngmdl.Role
	)
	if tx, err = s.fkDao.BeginTran(c); err != nil {
		log.Error("s.fkDao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	var applyInfo *mngmdl.RoleApply
	if applyInfo, err = s.fkDao.AuthRoleApplyByID(c, appKey, id); err != nil {
		log.Error("%v", err)
		return
	}
	if appInfo, err = s.fkDao.AppPass(c, appKey); err != nil {
		log.Error("%v", err)
		return
	}
	if roleInfo, err = s.fkDao.AuthRoleByVal(c, applyInfo.Role); err != nil {
		log.Error("%v", err)
		return
	}
	if _, err = s.fkDao.TxSetAuthUser(tx, appKey, userName, applyInfo.Name, applyInfo.Role); err != nil {
		log.Error("%v", err)
		return
	}
	if _, err = s.fkDao.TxUpdateAuthRoleApply(tx, appKey, userName, id, mngmdl.RoleApplyPass); err != nil {
		log.Error("%v", err)
		return
	}
	_ = s.fkDao.WechatMessageNotify(fmt.Sprintf("\"%s(%s)\"的管理员 \"%s\" 通过了您的权限申请。 -【%s】",
		appInfo.Name, appInfo.AppKey, userName, roleInfo.Name), applyInfo.Name, s.c.Comet.FawkesAppID)
	return
}

// AuthRoleApplyRefuse refuse role apply.
func (s *Service) AuthRoleApplyRefuse(c context.Context, appKey, userName string, id int) (err error) {
	var (
		tx       *sql.Tx
		appInfo  *appmdl.APP
		roleInfo *mngmdl.Role
	)
	if tx, err = s.fkDao.BeginTran(c); err != nil {
		log.Error("s.fkDao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	var applyInfo *mngmdl.RoleApply
	if applyInfo, err = s.fkDao.AuthRoleApplyByID(c, appKey, id); err != nil {
		log.Error("%v", err)
		return
	}
	if appInfo, err = s.fkDao.AppPass(c, appKey); err != nil {
		log.Error("%v", err)
		return
	}
	if roleInfo, err = s.fkDao.AuthRoleByVal(c, applyInfo.Role); err != nil {
		log.Error("%v", err)
		return
	}
	if _, err = s.fkDao.TxUpdateAuthRoleApply(tx, appKey, userName, id, mngmdl.RoleApplyRefuse); err != nil {
		log.Error("%v", err)
		return
	}
	_ = s.fkDao.WechatMessageNotify(fmt.Sprintf("\"%s(%s)\"的管理员 \"%s\" 拒绝了您的权限申请。 -【%s】",
		appInfo.Name, appInfo.AppKey, userName, roleInfo.Name), applyInfo.Name, s.c.Comet.FawkesAppID)
	return
}

func (s *Service) AuthRoleApplyIgnore(c context.Context, appKey, userName string, id int) (err error) {
	err = s.fkDao.Transact(c, func(tx *sql.Tx) error {
		if _, err = s.fkDao.TxUpdateAuthRoleApply(tx, appKey, userName, id, mngmdl.RoleApplyIgnore); err != nil {
			log.Errorc(c, "TxUpdateAuthRoleApply error %v", err)
			return err
		}
		return err
	})
	return
}

func (s *Service) LogList(c context.Context, appKey, env, md, operation, target, operator, stime, etime string,
	pn, ps int) (res *mngmdl.LogResult, err error) {
	var total int
	if total, err = s.fkDao.LogCount(c, appKey, env, md, operation, target, operator, stime, etime); err != nil {
		log.Error("%v", err)
		return
	}
	if total == 0 {
		return
	}
	var logs []*mngmdl.Log
	if logs, err = s.fkDao.Log(c, appKey, env, md, operation, target, operator, stime, etime, ps, pn); err != nil {
		log.Error("%v", err)
		return
	}
	res = &mngmdl.LogResult{
		PageInfo: &model.PageInfo{
			Total: total,
			Pn:    pn,
			Ps:    ps,
		},
		Items: logs,
	}
	return
}

func (s *Service) EventApplyAdd(c context.Context, appKey, userName, operator string, event, targetID int) (err error) {
	var (
		tx      *sql.Tx
		appInfo *appmdl.APP
	)
	if tx, err = s.fkDao.BeginTran(c); err != nil {
		log.Error("s.fkDao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	if _, err = s.fkDao.TxAddEventApply(tx, appKey, userName, operator, event, targetID); err != nil {
		log.Error("%v", err)
		return
	}
	if appInfo, err = s.fkDao.AppPass(c, appKey); err != nil {
		log.Error("%v", err)
		return
	}
	// 发布申请 （ 若无人发送. 则降级推送给管理员 ）
	usernames := []string{}
	admins, _ := s.fkDao.AuthUserListByRole(c, appKey, mngmdl.RoleAdmin)
	for _, admin := range admins {
		usernames = append(usernames, admin.Name)
	}
	switch event {
	case mngmdl.Event_Publish_Config:
		// to User
		_ = s.fkDao.WechatMessageNotify(fmt.Sprintf("管理员\"%s\" 已收到了应用\"%s(%s)\"上您的Config发布申请，"+
			"请耐心等待", operator, appInfo.Name, appInfo.AppKey), userName, s.c.Comet.FawkesAppID)
		// to Admin
		_ = s.fkDao.WechatCardMessageNotify(
			"Config发布申请提醒",
			fmt.Sprintf("%s 提交了一个Config发布申请\n应用：%s(%s)\n审核员：%s", userName, appInfo.Name, appInfo.AppKey, operator),
			fmt.Sprintf("http://fawkes.bilibili.co/#/config/list?env=prod&app_key=%v", appInfo.AppKey),
			"",
			strings.Join(usernames, "|"),
			s.c.Comet.FawkesAppID)
	case mngmdl.Event_Publish_FF:
		// to User
		_ = s.fkDao.WechatMessageNotify(fmt.Sprintf("管理员\"%s\" 已收到了应用\"%s(%s)\"上您的Config发布申请，"+
			"请耐心等待", operator, appInfo.Name, appInfo.AppKey), userName, s.c.Comet.FawkesAppID)
		// to Admin
		_ = s.fkDao.WechatCardMessageNotify(
			"FeatureFlag发布申请提醒",
			fmt.Sprintf("%s 提交了一个FeatureFlag发布申请\n应用：%s(%s)\n审核员：%s", userName, appInfo.Name, appInfo.AppKey, operator),
			fmt.Sprintf("http://fawkes.bilibili.co/#/featureflag/list?env=prod&app_key=%v", appInfo.AppKey),
			"",
			strings.Join(usernames, "|"),
			s.c.Comet.FawkesAppID)
	case mngmdl.Event_Publish_Event_Kibana_Query:
		event, _ := s.fkDao.ApmEvent(c, int64(targetID))
		if event == nil {
			return
		}
		// to User
		_ = s.fkDao.WechatMessageNotify("管理员已收到了您的Kibana配置申请，请耐心等待。 如有加急，请联系Fawkes小姐姐~", userName, s.c.Comet.FawkesAppID)
		// to Admin
		_ = s.fkDao.WechatCardMessageNotify(
			"Kibana配置申请",
			fmt.Sprintf("事件名：%s\n业务组：%s\n申请人：%s", event.Name, event.BusName, userName),
			"http://fawkes.bilibili.co/#/apm-manager/apm-event",
			"",
			strings.Join(s.c.AlarmReceiver.EventMonitorReceiver, "|"),
			s.c.Comet.FawkesAppID)
	case mngmdl.Event_Publish_Event_Field_Publish:
		event, _ := s.fkDao.ApmEvent(c, int64(targetID))
		if event == nil {
			return
		}
		// to User
		_ = s.fkDao.WechatMessageNotify("管理员已收到了您的技术埋点字段申请，请耐心等待。 如有加急，请联系Fawkes小姐姐~", userName, s.c.Comet.FawkesAppID)
		// to Admin
		_ = s.fkDao.WechatCardMessageNotify(
			"技术埋点字段审核通知",
			fmt.Sprintf("事件名：%s\n业务组：%s\n申请人：%s", event.Name, event.BusName, userName),
			"http://fawkes.bilibili.co/#/apm-manager/apm-event",
			"",
			strings.Join(s.c.AlarmReceiver.EventMonitorReceiver, "|"),
			s.c.Comet.FawkesAppID)
	}
	return
}

func (s *Service) EventApplyRecall(c context.Context, appKey, userName string, event, targetID int) (err error) {
	var (
		tx         *sql.Tx
		appInfo    *appmdl.APP
		applyInfos []*mngmdl.EventApply
	)
	if tx, err = s.fkDao.BeginTran(c); err != nil {
		log.Error("s.fkDao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	if applyInfos, err = s.fkDao.EventApplyList(c, appKey, event, targetID, 0); err != nil {
		log.Error("%v", err)
		return
	}
	if len(applyInfos) == 0 {
		return
	}
	if appInfo, err = s.fkDao.AppPass(c, appKey); err != nil {
		log.Error("%v", err)
		return
	}
	if _, err = s.fkDao.TxUpdateEventApply(tx, appKey, userName, event, targetID, 1); err != nil {
		log.Error("%v", err)
		return
	}
	appliers := []string{}
	for _, applyInfo := range applyInfos {
		appliers = append(appliers, applyInfo.Applicant)
	}
	// 数据去重 - 防止对于单个用户多次推送消息
	appliers = deleteRepeat(appliers)
	// 管理员信息
	admins, _ := s.fkDao.AuthUserListByRole(c, appKey, mngmdl.RoleAdmin)
	usernames := []string{}
	for _, admin := range admins {
		usernames = append(usernames, admin.Name)
	}
	switch event {
	case mngmdl.Event_Publish_Config:
		_ = s.fkDao.WechatMessageNotify(fmt.Sprintf("管理员\"%s\" 已发布了您申请应用\"%s(%s)\"上的Config配置信息。", userName, appInfo.Name, appInfo.AppKey), strings.Join(appliers, "|"), s.c.Comet.FawkesAppID)
		_ = s.fkDao.WechatMessageNotify(fmt.Sprintf("管理员\"%s\" 已发布了应用\"%s(%s)\"上的Config配置信息。", userName, appInfo.Name, appInfo.AppKey), strings.Join(usernames, "|"), s.c.Comet.FawkesAppID)
	case mngmdl.Event_Publish_FF:
		_ = s.fkDao.WechatMessageNotify(fmt.Sprintf("管理员\"%s\" 已发布了您申请应用\"%s(%s)\"上的FeatureFlag配置信息。", userName, appInfo.Name, appInfo.AppKey), strings.Join(appliers, "|"), s.c.Comet.FawkesAppID)
		_ = s.fkDao.WechatMessageNotify(fmt.Sprintf("管理员\"%s\" 已发布了应用\"%s(%s)\"上的FeatureFlag配置信息。", userName, appInfo.Name, appInfo.AppKey), strings.Join(usernames, "|"), s.c.Comet.FawkesAppID)
	case mngmdl.Event_Publish_Event_Kibana_Query:
		event, _ := s.fkDao.ApmEvent(c, int64(targetID))
		if event == nil {
			return
		}
		_ = s.fkDao.WechatMessageNotify(fmt.Sprintf("管理员\"%s\" 已完成了您申请的Kibana查询配置。 -- %s", userName, event.Name), strings.Join(appliers, "|"), s.c.Comet.FawkesAppID)
		_ = s.fkDao.WechatMessageNotify(fmt.Sprintf("管理员\"%s\" 已完成了Kibana查询配置。-- %s", userName, event.Name), strings.Join(s.c.AlarmReceiver.EventMonitorReceiver, "|"), s.c.Comet.FawkesAppID)
	case mngmdl.Event_Publish_Event_Field_Publish:
		event, _ := s.fkDao.ApmEvent(c, int64(targetID))
		if event == nil {
			return
		}
		_ = s.fkDao.WechatMessageNotify(fmt.Sprintf("管理员\"%s\" 已完成了您申请的技术埋点字段申请。 -- %s", userName, event.Name), strings.Join(appliers, "|"), s.c.Comet.FawkesAppID)
		_ = s.fkDao.WechatMessageNotify(fmt.Sprintf("管理员\"%s\" 已完成了技术埋点字段审核。-- %s", userName, event.Name), strings.Join(s.c.AlarmReceiver.EventMonitorReceiver, "|"), s.c.Comet.FawkesAppID)
	}
	return
}

func deleteRepeat(list []string) []string {
	mapdata := make(map[string]interface{})
	if len(list) <= 0 {
		return nil
	}
	for _, v := range list {
		mapdata[v] = "true"
	}
	var datas []string
	for k := range mapdata {
		if k == "" {
			continue
		}
		datas = append(datas, k)
	}
	return datas
}

func (s *Service) BfsRefreshCDN(c context.Context, urls string) (err error) {
	if err = s.fkDao.BfsRefreshCDN(urls); err != nil {
		log.Error("%v", err)
	}
	return
}

func (s *Service) AuthMessagePush(c context.Context, appKey, userName, title, content, link string, msg_type int) (err error) {
	// content 必须包含操作人username信息. 否则为异常信息
	if isContain := strings.Contains(content, userName); !isContain {
		log.Error("AuthMessagePush without operator name: %v", userName)
		return
	}
	admins, _ := s.fkDao.AuthUserListByRole(c, appKey, mngmdl.RoleAdmin)
	usernames := []string{}
	for _, admin := range admins {
		usernames = append(usernames, admin.Name)
	}
	switch msg_type {
	case appmdl.NOTIDY_WECHART_MESSAGE_TYPE_TEXT:
		_ = s.fkDao.WechatMessageNotify(content, strings.Join(usernames, "|"), s.c.Comet.FawkesAppID)
	case appmdl.NOTIDY_WECHART_MESSAGE_TYPE_CARD:
		_ = s.fkDao.WechatCardMessageNotify(title, content, link, "", strings.Join(usernames, "|"), s.c.Comet.FawkesAppID)
	}
	return
}

func (s *Service) AuthNickNameSet(c context.Context, userName, nickName string) (err error) {
	var tx *sql.Tx
	if tx, err = s.fkDao.BeginTran(c); err != nil {
		log.Error("s.fkDao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	err = s.fkDao.AuthNickNameSet(tx, userName, nickName)
	return
}

func (s *Service) UserNameList(c context.Context, userName string, ps, pn int) (res []*mngmdl.UserInfo, err error) {
	if res, err = s.fkDao.UserNameList(c, userName, ps, pn); err != nil {
		log.Error("UserNameList: %v", err)
	}
	return
}

func (s *Service) UserNameSet(c context.Context, userName, nickName string) (err error) {
	var tx *sql.Tx
	if tx, err = s.fkDao.BeginTran(c); err != nil {
		log.Error("s.fkDao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	err = s.fkDao.TxUserNameSet(tx, userName, nickName)
	return
}

func (s *Service) AuthAdminApply(c context.Context, appKey, operator string) (err error) {
	var (
		admins      []*mngmdl.User
		supervisors []*mngmdl.SupervisorRole
	)
	if admins, err = s.AuthUserListByRole(c, appKey, mngmdl.RoleAdmin); err != nil {
		log.Errorc(c, "AuthUserListByRole error %v", err)
		return
	}
	if supervisors, err = s.fkDao.AuthSupervisors(c); err != nil {
		log.Errorc(c, "AuthSupervisors error %v", err)
		return
	}
	var (
		adminsName      []string
		supervisorsName []string
	)
	for _, admin := range admins {
		adminsName = append(adminsName, admin.Name)
	}
	for _, supervisor := range supervisors {
		supervisorsName = append(supervisorsName, supervisor.Name)
	}
	// 工单通知
	params := map[string]interface{}{
		"dealer_specified_by_user": map[string]interface{}{
			"app_admin":         adminsName,
			"fawkes_supervisor": supervisorsName,
		},
	}
	workflow := &wf.Workflow{
		Title:    fmt.Sprintf("%s应用管理员 - 权限申请", appKey),
		Name:     "fawkes_app_admin_role_apply",
		Operator: operator,
		Params:   params,
	}
	if _, err = s.fkDao.CreateWorkflow(c, workflow, s.c.Comet); err != nil {
		log.Errorc(c, "workflow创建失败, error %v", err)
	}
	return
}
