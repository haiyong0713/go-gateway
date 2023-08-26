package app

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"go-common/library/database/sql"
	ecode "go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"

	"go-gateway/app/app-svr/fawkes/service/conf"
	"go-gateway/app/app-svr/fawkes/service/model"
	appmdl "go-gateway/app/app-svr/fawkes/service/model/app"
	cdmdl "go-gateway/app/app-svr/fawkes/service/model/cd"
	cimdl "go-gateway/app/app-svr/fawkes/service/model/ci"
	mailmdl "go-gateway/app/app-svr/fawkes/service/model/mail"
	mngmdl "go-gateway/app/app-svr/fawkes/service/model/manager"
	wf "go-gateway/app/app-svr/fawkes/service/model/workflow"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
	"go-gateway/app/app-svr/fawkes/service/tools/utils"
)

// AppInfo get app info.
func (s *Service) AppInfo(c context.Context, appKey string, ID int64) (app *appmdl.APP, err error) {
	if app, err = s.fkDao.AppInfo(c, appKey, ID); err != nil {
		log.Errorc(c, "%v", err)
		return
	}
	if app == nil {
		err = ecode.Error(ecode.RequestErr, fmt.Sprintf("appkey %s 不存在", appKey))
		return
	}
	// 废弃app表的owner. app.owner = 应用管理员
	var usernames []string
	if usernames, err = s.fkDao.AuthUserNamesDistinct(c, app.AppKey, fmt.Sprintf("%d", mngmdl.RoleAdmin)); err != nil {
		log.Errorc(c, "%v", err)
		return
	}
	if len(usernames) > 0 {
		app.UserAdmins = strings.Join(usernames, ",")
	}
	return
}

// AppEdit update app edit info.
func (s *Service) AppEdit(c context.Context, id, datacenterAppID, serverZone, isHost int64, appID, mobiApp, platform, gitPath, icon, owners, name, desc, treePath, dsymName, symbolsoName, projectID, userName, laserWebhook string, file []byte) (err error) {
	if err = s.MatchApp(c, "", mobiApp, platform, id, datacenterAppID); err != nil {
		return
	}
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
	if len(file) > 0 {
		if icon, err = s.Upload(c, model.BFSBucket, "", "", file); err != nil {
			log.Error("AppEdit - s.Upload error")
			return
		}
	}
	if _, err = s.fkDao.TxUpApp(tx, id, appID, mobiApp, platform, gitPath, name, icon, desc, treePath, projectID); err != nil {
		log.Error("%v", err)
		return
	}
	if isHost != 0 {
		var hostAppKey string
		if hostAppKey, err = s.fkDao.AppHost(c, datacenterAppID, platform); err != nil {
			log.Errorc(c, "AppHost error %v", err)
			return
		}
		if hostAppKey != "" {
			log.Errorc(c, "已存在宿主应用：%v, 申请应用名称: %v", hostAppKey, name)
			isHost = 0
		}
	}
	if _, err = s.fkDao.TxAppAttributeUpdate(tx, id, datacenterAppID, serverZone, isHost, owners, dsymName, symbolsoName, userName, laserWebhook); err != nil {
		log.Error("%v", err)
		return
	}
	return
}

// AppEdit update app edit info.
func (s *Service) AppUpdateIsHighestPeak(c context.Context, appKey string, isHighestPeak int64) (err error) {
	_ = s.fkDao.Transact(c, func(tx *sql.Tx) error {
		if _, err = s.fkDao.TxAppUpdateIsHighestPeak(tx, appKey, isHighestPeak); err != nil {
			log.Error("%v", err)
			return err
		}
		return nil
	})
	return
}

// AppFollowList get follow app list.
func (s *Service) AppFollowList(c context.Context, username string) (apps []*appmdl.APP, err error) {
	var appKeys []string
	if appKeys, err = s.fkDao.AppFollow(c, username); err != nil {
		log.Error("%v", err)
		return
	}
	if len(appKeys) > 0 {
		if apps, err = s.fkDao.AppsPass(c, appKeys, username, 0); err != nil {
			log.Error("%v", err)
		}
	}
	return
}

// AppList get app passed.
func (s *Service) AppList(c context.Context, username string, datacenterAppId int64) (apps []*appmdl.APP, err error) {
	var (
		akm     map[string]string
		appKeys []string
	)
	if appKeys, err = s.fkDao.AppFollow(c, username); err != nil {
		log.Error("%v", err)
		return
	}
	akm = make(map[string]string)
	for _, appKey := range appKeys {
		akm[appKey] = appKey
	}
	if apps, err = s.fkDao.AppsPass(c, []string{}, username, datacenterAppId); err != nil || len(apps) == 0 {
		log.Error("AppList %v or apps is 0", err)
		return
	}
	var aks []string
	for _, app := range apps {
		aks = append(aks, app.AppKey)
	}
	var versions map[string][]*model.Version
	if len(aks) > 0 {
		if versions, err = s.fkDao.PackVersionByAppKeys(c, "prod", aks); err != nil {
			log.Error("%v", err)
			return
		}
	}
	for _, app := range apps {
		if _, ok := akm[app.AppKey]; ok {
			app.IsFollow = 1
		}
		if vs, ok := versions[app.AppKey]; ok {
			for _, v := range vs {
				if app.VersionCode < v.VersionCode {
					app.VersionCode = v.VersionCode
				}
			}
		}
	}
	return
}

// AppFollowAdd add app follow.
func (s *Service) AppFollowAdd(c context.Context, appKey, username string) (err error) {
	var apps []*appmdl.APP
	if apps, err = s.fkDao.AppsPass(c, []string{appKey}, username, 0); err != nil {
		log.Error("%v", err)
	}
	if len(apps) == 0 {
		log.Error("AppFollowAdd %v", fmt.Errorf("appkey(%v) is invalid", appKey))
		return
	}
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
	if _, err = s.fkDao.TxInAppFollow(tx, appKey, username); err != nil {
		log.Error("%v", err)
	}
	return
}

// AppFollowDeL del app follow.
func (s *Service) AppFollowDeL(c context.Context, appKey, username string) (err error) {
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
	if _, err = s.fkDao.TxDelAppFollow(tx, appKey, username); err != nil {
		log.Error("%v", err)
	}
	return
}

// AppAdd add app.
func (s *Service) AppAdd(c context.Context, datacenterAppId, isHost int64, appID, appKey, mobiApp, platform, name, treePath, desc, gitPath, owners, userName string, file []byte) (err error) {
	if err = s.MatchApp(c, appKey, mobiApp, platform, 0, datacenterAppId); err != nil {
		return
	}
	var icon string
	if icon, err = s.Upload(c, model.BFSBucket, "", "", file); err != nil {
		log.Error("%v", err)
		return
	}
	var gitPrjID string
	if len(gitPath) > 0 {
		if strings.HasPrefix(gitPath, "git@") {
			pathComps := strings.Split(gitPath, ":")
			gitPrjID = strings.Split(pathComps[len(pathComps)-1], ".git")[0]
		} else {
			pathComps := strings.Split(gitPath, "/")
			projectName := strings.Split(pathComps[len(pathComps)-1], ".git")[0]
			groupName := pathComps[len(pathComps)-2]
			gitPrjID = groupName + "/" + projectName
		}
	}
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
	var id int64
	if id, err = s.fkDao.TxAppAdd(tx, appID, appKey, mobiApp, platform, name, treePath, icon, desc, gitPath, gitPrjID); err != nil {
		log.Error("%v", err)
		return
	}
	// 工单通知
	var (
		workflowId     string
		isWfDatacenter bool
	)
	if platform == "android" || platform == "ios" {
		isWfDatacenter = true
	}
	params := map[string]interface{}{
		"fawkes_app_name":           name,
		"fawkes_app_key":            appKey,
		"mobi_app":                  mobiApp,
		"require_datacenter_prodid": isWfDatacenter,
		"datacenter_app_id":         strconv.FormatInt(datacenterAppId, 10),
	}
	workflow := &wf.Workflow{
		Title:    fmt.Sprintf("%s - 申请Fawkes应用准入", name),
		Name:     "fawkes_app_apply",
		Operator: userName,
		Params:   params,
	}
	if workflowId, err = s.fkDao.CreateWorkflow(c, workflow, s.c.Comet); err != nil {
		log.Error("workflow创建失败, error %v", err)
	}
	if isHost != 0 {
		var hostAppKey string
		if hostAppKey, err = s.fkDao.AppHost(c, datacenterAppId, platform); err != nil {
			log.Errorc(c, "AppHostInfo error %v", err)
			return
		}
		if hostAppKey != "" {
			log.Errorc(c, "已存在宿主应用：%v，申请应用名称：%v", hostAppKey, name)
			isHost = 0
		}
	}
	if _, err = s.fkDao.TxAppAttributeAdd(tx, id, datacenterAppId, isHost, workflowId, appKey, owners, userName); err != nil {
		log.Error("%v", err)
	}
	// 通知管理员有用户申请APP
	supervisors, _ := s.fkDao.AuthSupervisors(c)
	usernames := make([]string, 0, len(supervisors))
	if len(supervisors) > 0 {
		for _, supervisor := range supervisors {
			usernames = append(usernames, supervisor.Name)
		}
		_ = s.fkDao.WechatCardMessageNotify(
			"应用申请通知",
			fmt.Sprintf("%s 提交了一个应用申请\n应用：%s(%s)\n", userName, name, appKey),
			"http://fawkes.bilibili.co/#/app-audit/list",
			"",
			strings.Join(usernames, "|"),
			s.c.Comet.FawkesAppID)
	}
	return
}

// MatchApp 为了符合数据平台一个mobiApp只能对应一个datacenterAppId的要求，并且兼容Fawkes平台已有操作规则，应用在注册和修改时，需要满足 1. appKey不能为空 2. mobiApp与datacenterAppId + platform保持强一致(兼容粉版android 32位和64位)
func (s *Service) MatchApp(c context.Context, appKey, mobiApp, platform string, id, datacenterAppId int64) (err error) {
	// datacenterAppId为空，跳过匹配
	if datacenterAppId == 0 {
		return
	}
	// mobiApp白名单，跳过匹配
	for _, app := range s.c.MobiAppWhiteList {
		if app == mobiApp {
			return
		}
	}
	if appKey != "" {
		var count int64
		if count, err = s.fkDao.AppCount(c, appKey); err != nil {
			log.Errorc(c, "[ExistApp]-appKey error %v", err)
			return
		}
		if count > 0 {
			err = ecode.Error(ecode.RequestErr, "appKey已经存在")
			return
		}
	}
	var (
		mobiApps []*appmdl.APP
		dcApps   []*appmdl.APP
	)
	if mobiApps, err = s.fkDao.AppListByMobiApp(c, mobiApp); err != nil {
		log.Errorc(c, "[ExistApp]-mobiApp error %v", err)
		return
	}
	if dcApps, err = s.fkDao.AppListByDatacenterAppId(c, datacenterAppId); err != nil {
		log.Errorc(c, "[ExistApp]-appId error %v", err)
		return
	}
	var (
		isExistAppId, isExistMobileApp bool
		existAppId                     int64
		existMobiApp                   string
	)
	for _, mobiApp := range mobiApps {
		if mobiApp.ID == id || mobiApp.DataCenterAppID == 0 {
			continue
		}
		if mobiApp.DataCenterAppID != datacenterAppId {
			isExistAppId = true
			existAppId = mobiApp.DataCenterAppID
			break
		}
	}
	for _, dcApp := range dcApps {
		if dcApp.ID == id {
			continue
		}
		if dcApp.MobiApp != mobiApp && dcApp.Platform == platform {
			isExistMobileApp = true
			existMobiApp = dcApp.MobiApp
			break
		}
	}
	if isExistAppId {
		return ecode.Errorf(ecode.RequestErr, "mobiApp在platform:%v,datacenterAppId:%v已存在", platform, existAppId)
	}
	if isExistMobileApp {
		return ecode.Errorf(ecode.RequestErr, "datacenterAppId在mobiApp:%v已存在", existMobiApp)
	}
	return
}

// AppKeys make appkey.
func (s *Service) AppKeys(c context.Context) (key string, err error) {
	for {
		kinds := [][]int{{10, 48}, {26, 97}}
		keyb := make([]byte, 4)
		rand.Seed(time.Now().UnixNano())
		for i := 0; i < 4; i++ {
			ikind := rand.Intn(2)
			scope, base := kinds[ikind][0], kinds[ikind][1]
			keyb[i] = uint8(base + rand.Intn(scope))
		}
		var count int64
		if count, err = s.fkDao.AppCount(c, string(keyb)); err != nil {
			log.Error("%v", err)
			return
		}
		if count > 0 {
			continue
		}
		key = string(keyb)
		break
	}
	return
}

// AppAuditList get app audit list.
func (s *Service) AppAuditList(c context.Context) (auditApps []*appmdl.APP, err error) {
	if auditApps, err = s.fkDao.AppsAudit(c); err != nil {
		log.Error("%v", err)
	}
	return
}

// AppAudit audit app.
func (s *Service) AppAudit(c context.Context, appKey, reason, userName string, status, isActive int, id int64) (err error) {
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
	var count int64
	if count, err = s.fkDao.TxUpAppAudit(tx, appKey, reason, status, id); err != nil {
		log.Error("%v", err)
		return
	}
	if count == 0 {
		return
	}
	// 应用注销激活. 无需初始化基础配置
	if status == appmdl.AuditPass && isActive != 1 {
		// init config default.
		if _, err = s.fkDao.TxSetConfigVersion(tx, appKey, "test", "default", 0, "system"); err != nil {
			log.Error("%v", err)
			return
		}
		if _, err = s.fkDao.TxSetConfigVersion(tx, appKey, "prod", "default", 0, "system"); err != nil {
			log.Error("%v", err)
			return
		}
		var app *appmdl.APP
		if app, err = s.fkDao.AppByID(c, id); err != nil {
			log.Error("%v", err)
			return
		}
		if app == nil {
			return
		}
		// 创建应用同步静态渠道列表
		if app.Platform == "android" {
			var (
				chLists []*appmdl.Channel
				sqls    []string
				args    []interface{}
			)
			if chLists, err = s.fkDao.ChannelAllList(c); err != nil {
				log.Error("%v", err)
				return
			}
			for _, ch := range chLists {
				sqls = append(sqls, "(?,?,?)")
				args = append(args, appKey, ch.ID, userName)
			}
			if len(args) > 0 {
				if _, err = s.fkDao.AppChannelAdds(tx, sqls, args); err != nil {
					log.Error("%v", err)
				}
			}
		}
		// 将owners 设为管理员
		if app.Owners != "" {
			for _, owner := range strings.Split(app.Owners, ",") {
				_, _ = s.fkDao.TxSetAuthUser(tx, appKey, userName, owner, mngmdl.RoleAdmin)
			}
		}
	}
	return
}

// AppMailtoList get the mailto list of the app
func (s *Service) AppMailtoList(c context.Context, appKey, funcModule string, receiverType int64) (res *appmdl.ResMailList, err error) {
	var (
		mailUnameList []string
	)
	if mailUnameList, err = s.fkDao.AppMailtoList(c, appKey, funcModule, receiverType); err != nil {
		log.Errorc(c, "%v", err)
		return
	}
	res = &appmdl.ResMailList{MailList: mailUnameList}
	return
}

// UpdateAppMailtoList update the mailto list of the app
func (s *Service) UpdateAppMailtoList(c context.Context, appKey, unameListStr, funcModule string, receiverType int64) (err error) {
	var (
		oldUnameList, newUnameList []string
		oldFindFlgs                map[string]bool
	)
	oldFindFlgs = make(map[string]bool)
	if oldUnameList, err = s.fkDao.AppMailtoList(c, appKey, funcModule, receiverType); err != nil {
		log.Errorc(c, "%v", err)
		return
	}
	newUnameList = strings.Split(unameListStr, ",")
	var tx *sql.Tx
	if tx, err = s.fkDao.BeginTran(context.Background()); err != nil {
		log.Error("s.dao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
	}()
	for _, newUname := range newUnameList {
		newFindFlg := false
		for _, oldUname := range oldUnameList {
			if newUname == oldUname {
				newFindFlg = true
				oldFindFlgs[oldUname] = true
				break
			}
		}
		if !newFindFlg {
			if _, err = s.fkDao.TxAppMailtoAdd(tx, appKey, funcModule, newUname, receiverType); err != nil {
				if err = tx.Rollback(); err != nil {
					log.Error("tx.Rollback error(%v)", err)
				}
				log.Errorc(c, "%v", err)
				return
			}
		}
	}

	for _, oldUname := range oldUnameList {
		if _, ok := oldFindFlgs[oldUname]; !ok {
			if _, err = s.fkDao.TxAppMailtoDel(tx, appKey, funcModule, oldUname, receiverType); err != nil {
				if err = tx.Rollback(); err != nil {
					log.Error("tx.Rollback error(%v)", err)
				}
				log.Error("%v", err)
				return
			}
		}
	}
	if err = tx.Commit(); err != nil {
		log.Error("tx.Commit error(%v)", err)
	}
	return
}

func (s *Service) RobotNotify(c context.Context, webhookUrl string, msgType, msgBody string, botId int64) (err error) {
	var notify interface{}
	if webhookUrl == "" && botId == 0 {
		return
	}
	switch msgType {
	case "text":
		var textMessage *appmdl.Text
		if err = json.Unmarshal([]byte(msgBody), &textMessage); err != nil {
			log.Errorc(c, "%v", err)
			return
		}
		notify = textMessage
	case "image":
		var imageMessage *appmdl.Image
		if err = json.Unmarshal([]byte(msgBody), &imageMessage); err != nil {
			log.Errorc(c, "%v", err)
			return
		}
		notify = imageMessage
	case "markdown":
		var markdownMessage *appmdl.Markdown
		if err = json.Unmarshal([]byte(msgBody), &markdownMessage); err != nil {
			log.Errorc(c, "%v", err)
			return
		}
		notify = markdownMessage
	case "news":
		var newsMessage *appmdl.News
		if err = json.Unmarshal([]byte(msgBody), &newsMessage); err != nil {
			log.Errorc(c, "%v", err)
			return
		}
		notify = newsMessage
	}
	if botId != 0 {
		var robot *appmdl.Robot
		if robot, err = s.fkDao.AppRobotInfoById(c, botId); err != nil {
			log.Errorc(c, "%v", err)
			return
		}
		if robot != nil {
			webhookUrl = robot.WebHook
		}
	}
	if err = s.fkDao.RobotNotify(webhookUrl, notify); err != nil {
		log.Errorc(c, "%v", err)
	}
	return
}

func (s *Service) AppWXAppNotify(c context.Context, appKeys, roles, content, userName, assignedUsers string, isTest int) (err error) {
	var notifyUserNames []string
	if notifyUserNames, err = s.WechatAppNotifyUserFilter(c, appKeys, roles, userName, assignedUsers, isTest); err != nil {
		log.Errorc(c, "WechatAppNotifyUserFilter error %v", err)
		return
	}
	err = s.fkDao.WechatMessageNotify(content, strings.Join(notifyUserNames, "|"), s.c.Comet.FawkesAppID)
	log.Error("AppWXAppNotify message send length: %v", len(notifyUserNames))
	return
}

func (s *Service) AppWXAppPictureNotify(c context.Context, appKeys, roles, userName, assignedUsers string, pic multipart.File, picHeader *multipart.FileHeader, isTest int) (err error) {
	var notifyUserNames []string
	if notifyUserNames, err = s.WechatAppNotifyUserFilter(c, appKeys, roles, userName, assignedUsers, isTest); err != nil {
		log.Errorc(c, "WechatAppNotifyUserFilter error %v", err)
		return
	}
	var (
		users   []*mngmdl.UserInfo
		userIds []string
	)
	if users, err = s.fkDao.UserInfoList(c, notifyUserNames); err != nil {
		log.Errorc(c, "UserInfoList error %v", err)
		return
	}
	for _, user := range users {
		userIds = append(userIds, user.UserId)
	}
	if len(userIds) < 1 {
		log.Warnc(c, "userId is nil,users %v", notifyUserNames)
		return
	}
	var tmpPic *appmdl.WXNotifyTmpFileResp
	if tmpPic, err = s.WechatAppTmpFileUpload(c, "image", pic, picHeader); err != nil {
		log.Errorc(c, "AppWechatTmpFileUpload error %v", err)
		return
	}
	if _, err = s.AppWechatNotify(c, strings.Join(userIds, "|"), appmdl.WXNotifyType_Image, tmpPic.MediaId); err != nil {
		log.Errorc(c, "AppWechatNotify error %v", err)
	}
	return
}

func (s *Service) WechatAppNotifyUserFilter(c context.Context, appKeys, roles, userName, assignedUsers string, isTest int) (notifyUserNames []string, err error) {
	var (
		usernames    []string
		distinctName map[string]struct{}
	)
	if isTest != 0 {
		usernames = append(usernames, userName)
	} else {
		if assignedUsers == "" {
			if usernames, err = s.fkDao.AuthUserNamesDistinct(c, appKeys, roles); err != nil {
				log.Error("%v", err)
				return
			}
		} else {
			assignedUserList := strings.Split(assignedUsers, ",")
			usernames = append(usernames, assignedUserList...)
		}
	}
	if distinctName, err = s.fkDao.AuthDistinctUserInfo(c, usernames); err != nil {
		log.Error("AuthDistinctUserInfo error %v")
	}
	for _, name := range usernames {
		if _, ok := distinctName[name]; ok {
			notifyUserNames = append(notifyUserNames, name)
		}
	}
	return
}

// WechatAppTmpFileUpload TODO 封装成wechat client
func (s *Service) WechatAppTmpFileUpload(c context.Context, fileType string, file multipart.File, fileHeader *multipart.FileHeader) (resp *appmdl.WXNotifyTmpFileResp, err error) {
	var token string
	if token, err = s.fkDao.GetWechatToken(c, s.c.WXNotify.CorpSecret); err != nil {
		log.Errorc(c, "GetCacheWechatToken error %v", err)
		return
	}
	if resp, err = s.fkDao.WechatTmpFileUpload(context.Background(), fileType, file, fileHeader, token); err != nil {
		log.Errorc(c, "WechatTmpFileUpload error %v", err)
		return
	}
	// token过期，再次获取
	if resp.ErrCode == appmdl.WXNotifyAccessTokenExpired {
		if token, err = s.fkDao.GetWechatToken(c, s.c.WXNotify.CorpSecret); err != nil {
			log.Errorc(c, "GetWechatToken error %v", err)
			return
		}
		if resp, err = s.fkDao.WechatTmpFileUpload(context.Background(), fileType, file, fileHeader, token); err != nil {
			log.Errorc(c, "WechatTmpFileUpload error %v", err)
			return
		}
	}
	return
}

// AppWechatNotify TODO 封装成wechat client
func (s *Service) AppWechatNotify(c context.Context, users string, notifyType int64, content string) (resp *appmdl.WXNotifyMsgResp, err error) {
	var (
		token   string
		message *appmdl.WXNotifyMessage
	)
	if token, err = s.fkDao.GetWechatToken(c, s.c.WXNotify.CorpSecret); err != nil {
		log.Errorc(c, "GetCacheWechatToken error %v", err)
		return
	}
	switch notifyType {
	case appmdl.WXNotifyType_Text:
		message = &appmdl.WXNotifyMessage{
			MsgType: "text",
			Text: &appmdl.WXNotifyText{
				Content: content,
			},
		}
	case appmdl.WXNotifyType_Markdown:
		message = &appmdl.WXNotifyMessage{
			MsgType: "markdown",
			Markdown: &appmdl.WXNotifyMarkdown{
				Content: content,
			},
		}
	case appmdl.WXNotifyType_Image:
		message = &appmdl.WXNotifyMessage{
			MsgType: "image",
			Image: &appmdl.WXNotifyImage{
				MediaId: content,
			},
		}
	default:
		log.Error("AppWechatNotify unknown type")
		return
	}
	message.Touser = users
	message.Agentid = s.c.WXNotify.AgentID
	if resp, err = s.fkDao.WechatAppNotify(context.Background(), message, token); err != nil {
		log.Errorc(c, "WechatAppNotify error %v", err)
		return
	}
	// token过期，再次获取
	if resp.ErrCode == appmdl.WXNotifyAccessTokenExpired {
		if token, err = s.fkDao.GetWechatToken(c, s.c.WXNotify.CorpSecret); err != nil {
			log.Errorc(c, "GetWechatToken error %v", err)
			return
		}
		if resp, err = s.fkDao.WechatAppNotify(context.Background(), message, token); err != nil {
			log.Errorc(c, "WechatAppNotify error %v", err)
			return
		}
	}
	return
}

func (s *Service) AppRobotSet(c context.Context, appKey, rebotName, rebotWebhookUrl, userName string) (err error) {
	var (
		appRobot *appmdl.Robot
	)
	if appRobot, err = s.fkDao.AppRobotInfoByWebhook(c, rebotWebhookUrl); err != nil {
		log.Errorc(c, "AppRobotInfoByWebhook error %v", err)
		return
	}
	err = s.fkDao.Transact(c, func(tx *sql.Tx) error {
		// 判断是否有消息机器人
		if appRobot != nil {
			keys := strings.Split(appRobot.AppKeys, ",")
			keys = append(keys, appKey)
			appKeys := strings.Join(removeDuplication(keys), ",")
			if err = s.fkDao.TxAppRobotUpdate(tx, rebotName, rebotWebhookUrl, appKeys, appmdl.MessageBot, "", "", userName, appmdl.RobotOpen, appmdl.RobotNotGlobal, appmdl.RobotDefault, appRobot.ID); err != nil {
				log.Errorc(c, "TxAppRobotUpdate error %v", err)
			}
		} else {
			if err = s.fkDao.TxAppRobotAdd(tx, rebotName, rebotWebhookUrl, appKey, appmdl.MessageBot, "", "", userName, appmdl.RobotOpen, appmdl.RobotNotGlobal, appmdl.RobotDefault); err != nil {
				log.Errorc(c, "TxAppRobotAdd error %v", err)
			}
		}
		return err
	})
	return
}

// AppRobotList get robot list
func (s *Service) AppRobotList(c context.Context, appKey, funcModule, botName, userName string, state int) (res []*appmdl.Robot, err error) {
	var (
		robots       []*appmdl.Robot
		isSupervisor bool
		isAdmin      bool
	)
	if isSupervisor, isAdmin, err = s.JudgeRole(c, appKey, userName); err != nil {
		log.Errorc(c, "JudgeRole error %v", err)
		return
	}
	if robots, err = s.fkDao.AppRobotList(c, appKey, funcModule, botName, state); err != nil {
		log.Errorc(c, "AppRobotList error %v", err)
		return
	}
	for _, robot := range robots {
		var (
			users  []string
			isUser bool
		)
		users = strings.Split(robot.Users, ",")
		if robot.Users == "" {
			isUser = true
		} else {
			for _, user := range users {
				if userName == user {
					isUser = true
					break
				}
			}
		}
		if !isSupervisor && !isAdmin {
			robot.WebHook = ""
		}
		if isSupervisor || isUser || isAdmin {
			res = append(res, robot)
		}
	}
	return
}

// AppRobotAdd add app robot
func (s *Service) AppRobotAdd(c context.Context, botName, webhook, appKeys, funcModule, users, description, userName string, state, isGlobal, isDefault int) (data *appmdl.RobotAddStatus, err error) {
	var existRobot *appmdl.Robot

	if existRobot, err = s.fkDao.AppRobotInfoByWebhook(c, webhook); err != nil {
		log.Errorc(c, "AppRobotInfoByWebhook error %v", err)
		return
	}
	err = s.fkDao.Transact(c, func(tx *sql.Tx) error {
		if existRobot != nil {
			if existRobot.AppKeys != "" {
				appKeys = fmt.Sprintf("%s,%s", existRobot.AppKeys, appKeys)
			}
			keys := strings.Split(strings.TrimSuffix(appKeys, ","), ",")
			rdKeys := removeDuplication(keys)
			if err = s.fkDao.TxAppRobotUpdateAppKey(tx, strings.Join(rdKeys, ","), existRobot.ID); err != nil {
				log.Errorc(c, "TxAppRobotUpdate: %v", err)
				return err
			}
			data = &appmdl.RobotAddStatus{
				Msg:  fmt.Sprintf("存在相同webhook的机器人:【%s】，已自动加入该机器人", existRobot.BotName),
				Code: 1,
			}
		} else {
			if err = s.fkDao.TxAppRobotAdd(tx, botName, webhook, appKeys, funcModule, users, description, userName, state, isGlobal, isDefault); err != nil {
				log.Errorc(c, "TxAppRobotAdd: %v", err)
				return err
			}
			data = &appmdl.RobotAddStatus{
				Msg:  fmt.Sprintf("添加机器人:【%s】成功", botName),
				Code: 0,
			}
		}
		return err
	})
	return
}

// AppRobotUpdate Update app robot
func (s *Service) AppRobotUpdate(c context.Context, botName, webhook, appKeys, funcModule, users, description, userName string, state, isGlobal, isDefault int, id int64) (err error) {
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
	if err = s.fkDao.TxAppRobotUpdate(tx, botName, webhook, appKeys, funcModule, users, description, userName, state, isGlobal, isDefault, id); err != nil {
		log.Error("TxAppRobotUpdate: %v", err)
	}
	return
}

// AppRobotDel delete app robot
func (s *Service) AppRobotDel(c context.Context, id int64) (err error) {
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
	if err = s.fkDao.TxAppRobotDel(tx, id); err != nil {
		log.Error("TxAppRobotDel: %v", err)
	}
	return
}

// UploadFile upload a file
func (s *Service) UploadFile(c context.Context, dir string, file multipart.File, header *multipart.FileHeader) (res *appmdl.RobotUploadRes, err error) {
	var (
		destFile *os.File
	)
	destFileDir := filepath.Join(conf.Conf.LocalPath.LocalDir, dir)
	// 若文件夹不存在. 则新建一个文件夹
	if _, err = os.Stat(destFileDir); err != nil {
		if os.IsNotExist(err) {
			_ = os.MkdirAll(destFileDir, 0755)
		}
	}
	destFilePath := filepath.Join(destFileDir, header.Filename)
	// 文件存在则先删除
	if _, err = os.Stat(destFilePath); err == nil {
		if err = os.Remove(destFilePath); err != nil {
			log.Error("os.Remove(%s) error(%v)", destFilePath, err)
			return
		}
	}
	if destFile, err = os.Create(destFilePath); err != nil {
		log.Error("%v", err)
		return
	}
	if _, err = io.Copy(destFile, file); err != nil {
		log.Error("%v", err)
		return
	}
	defer file.Close()
	defer destFile.Close()
	res = &appmdl.RobotUploadRes{
		FileURL: conf.Conf.LocalPath.LocalDomain + "/" + dir + "/" + header.Filename,
	}
	return
}

// AppNotificationList service
func (s *Service) AppNotificationList(c context.Context, appKey, platform string, state int64) (res []*appmdl.Notif, err error) {
	if res, err = s.fkDao.AppNotificationList(c, appKey, platform, state); err != nil {
		log.Error("AppNotificationList: %v", err)
	}
	return
}

func (s *Service) AppNotificationUpdate(c context.Context, id int64, appKeys, platform, routePath, title, content, url string, state, isGlobal, showType, closeable int64, effectTime, expireTime string, username string) (err error) {
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
	if _, err = s.fkDao.TxAppNotificationUpdate(tx, id, appKeys, platform, routePath, title, content, url, state, isGlobal, showType, closeable, effectTime, expireTime, username); err != nil {
		log.Error("%v", err)
	}
	return
}

func (s *Service) AppNotificationAdd(c context.Context, appKeys, platform, routePath, title, content, url string, state, isGlobal, showType, closeable int64, effectTime, expireTime string, username string) (err error) {
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
	if _, err = s.fkDao.TxAppNotificationAdd(tx, appKeys, platform, routePath, title, content, url, state, isGlobal, showType, closeable, effectTime, expireTime, username); err != nil {
		log.Error("%v", err)
	}
	return
}

func (s *Service) AppMailConfigAdd(c context.Context, appKey, funcModule, host, address, pwd, name, operator string, port int) (err error) {
	if err = s.fkDao.AppMailConfigAdd(c, appKey, funcModule, host, address, pwd, name, operator, port); err != nil {
		log.Errorc(c, "%v", err)
	}
	return
}

func (s *Service) AppMailConfigDel(c context.Context, id int64) (err error) {
	if err = s.fkDao.AppMailConfigDel(c, id); err != nil {
		log.Errorc(c, "%v", err)
	}
	return
}

func (s *Service) AppMailConfigUpdate(c context.Context, appKey, funcModule, host, address, pwd, name, operator string, port int, id int64) (err error) {
	if err = s.fkDao.AppMailConfigUpdate(c, appKey, funcModule, host, address, pwd, name, operator, port, id); err != nil {
		log.Errorc(c, "%v", err)
	}
	return
}

func (s *Service) AppMailConfigList(c context.Context, appKey, funcModule, username string) (res []*mailmdl.SenderConfig, err error) {
	var (
		senders      []*mailmdl.SenderConfig
		isSupervisor bool
		isAdmin      bool
	)
	if isSupervisor, isAdmin, err = s.JudgeRole(c, appKey, username); err != nil {
		log.Errorc(c, "JudgeRole error %v", err)
		return
	}
	if senders, err = s.fkDao.AppMailConfigList(c, appKey, funcModule); err != nil {
		log.Errorc(c, "%v", err)
	}
	if !isSupervisor && !isAdmin {
		for _, sender := range senders {
			sender.Host = ""
			sender.Port = 0
			sender.Pwd = ""
		}
	}
	res = senders
	return
}

func (s *Service) AppMailList(c context.Context, appKey, funcModule string) (res []*mailmdl.AppMailWithModule, err error) {
	if res, err = s.fkDao.AppMailList(c, appKey, funcModule); err != nil {
		log.Errorc(c, "%v", err)
	}
	return
}

func (s *Service) JudgeRole(c context.Context, appKey, username string) (isSupervisor, isAdmin bool, err error) {
	var (
		supervisors []*mngmdl.SupervisorRole
		role        *mngmdl.ResultRole
	)
	if supervisors, err = s.fkDao.AuthSupervisor(c, username); err != nil {
		log.Errorc(c, "AuthSupervisor error %v", err)
		return
	}
	if role, err = s.fkDao.AuthUser(c, appKey, username); err != nil {
		log.Errorc(c, "AuthUser error %v", err)
		return
	}
	if role == nil || role.Role == mngmdl.RoleAdmin {
		isAdmin = true
	}
	for _, supervisor := range supervisors {
		if supervisor.Name == username {
			isSupervisor = true
			break
		}
	}
	return
}

func (s *Service) AppTriggerPipeline(c *bm.Context, appKey string, buildId int64, envVars, sender string) (res interface{}, err error) {
	var pack *cdmdl.Pack
	if pack, err = s.fkDao.PackByBuild(c, appKey, "test", buildId); err != nil {
		log.Errorc(c, "PackByBuild err %v", err)
		return
	}
	if pack == nil {
		err = ecode.Error(ecode.RequestErr, "构建包为空")
		return
	}
	var variables = map[string]string{
		"APP_KEY":     appKey,
		"FAWKES":      "1",
		"BUILD_ID":    strconv.FormatInt(buildId, 10),
		"FAWKES_USER": utils.GetUsername(c),
		"TASK":        "publish",
		"ARCHIVE_URL": strings.TrimSuffix(pack.PackURL, "version.txt"),
	}
	var envVarMap = make(map[string]string)
	if envVars != "" {
		if err = json.Unmarshal([]byte(envVars), &envVarMap); err != nil {
			err = ecode.Error(ecode.RequestErr, "env_var error")
			return
		}
	}
	for key, value := range envVarMap {
		variables[key] = value
	}
	if err = s.fkDao.Transact(c, func(tx *sql.Tx) error {
		if _, err = s.fkDao.TxUpdatePackSender(tx, sender, appKey, "test", buildId); err != nil {
			log.Errorc(c, "TxUpdatePackSender error %v", err)
			return err
		}
		return err
	}); err != nil {
		log.Errorc(c, "Transact error %v", err)
		return
	}
	if _, err = s.gitSvr.TriggerPipeline(c, appKey, cimdl.GitTypeCommit, pack.Commit, variables); err != nil {
		return
	}
	return
}

func removeDuplication(arr []string) []string {
	set := make(map[string]struct{}, len(arr))
	j := 0
	for _, v := range arr {
		_, ok := set[v]
		if ok {
			continue
		}
		set[v] = struct{}{}
		arr[j] = v
		j++
	}
	return arr[:j]
}
