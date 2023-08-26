package app

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"go-common/library/database/sql"
	"go-common/library/ecode"

	"go-gateway/app/app-svr/fawkes/service/conf"
	"go-gateway/app/app-svr/fawkes/service/model"
	appmdl "go-gateway/app/app-svr/fawkes/service/model/app"
	cdmdl "go-gateway/app/app-svr/fawkes/service/model/cd"
	mailmdl "go-gateway/app/app-svr/fawkes/service/model/mail"
	mngmdl "go-gateway/app/app-svr/fawkes/service/model/manager"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
	"go-gateway/app/app-svr/fawkes/service/tools/utils"
)

// HotfixPushEnv push hotfix to env
func (s *Service) HotfixPushEnv(c context.Context, request *appmdl.HfPushEnvReq) (err error) {
	var hfInfo []*appmdl.HotfixInfo
	if hfInfo, err = s.fkDao.GetHotfixInfo(c, request.AppKey, request.BuildID); err != nil || hfInfo == nil {
		log.Error("s.fkDao.GetHotfixInfo failed. %v", err)
		return
	}
	if len(hfInfo) > 1 || request.Env == hfInfo[0].Env {
		return ecode.Error(ecode.Conflict, "该环境已存在此热修复包，请勿重复推送")
	}
	preEnv := appmdl.GetPreEnv(request.Env)
	var lastInternalVersionCode int64
	if request.Env == "prod" {
		if lastInternalVersionCode, err = s.fkDao.GetLastProdHfInterVer(c, request.AppKey, hfInfo[0].OrigVersionCode); err != nil {
			log.Error("s.fkDao.GetLastProdHfInterVer failed. %v", err)
			return
		}
		if hfInfo[0].InternalVersionCode <= lastInternalVersionCode {
			log.Error("internal version code must be bigger than %v", lastInternalVersionCode)
			err = errors.New("internal version code must be bigger than latest one")
			return
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
	if _, err = s.fkDao.PushHotfix(tx, request.AppKey, preEnv, request.Env, request.BuildID); err != nil {
		log.Error("s.fkDao.PushHotfix() failed. %v", err)
		return
	}
	if _, err = s.fkDao.PushHotfixConf(tx, request.AppKey, preEnv, request.Env, request.BuildID); err != nil {
		log.Error("s.fkDao.PushHotfixConf() failed. %v", err)
	}
	return
}

// HotfixConfSet add hotfix config
func (s *Service) HotfixConfSet(c context.Context, request *appmdl.HfConfSetReq, userName string) (err error) {
	var hotfixID int64
	if hotfixID, err = s.fkDao.GetHotfixID(c, request.AppKey, request.Env, request.BuildID); err != nil {
		log.Error("s.fkDao.GetHotfixID() failed. %v", err)
		return
	}
	if hotfixID == 0 {
		return ecode.Error(ecode.NothingFound, "热修复包数据不存在，请核对后操作")
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
	if _, err = s.fkDao.TxAddOrUpdateHtConfig(tx, request.AppKey, request.Env, request.Channel, request.City, request.Device,
		request.BuildID, request.UpgradNum, request.Status); err != nil {
		log.Error("s.fkDao.TxAddOrUpdateHtConfig() failed. %v", err)
		return
	}
	// add log
	var (
		hfInfos []*appmdl.HotfixInfo
		logInfo string
	)
	if hfInfos, err = s.fkDao.GetHotfixInfo(c, request.AppKey, request.BuildID); err != nil {
		log.Error("%v", err)
		err = nil
		return
	}
	if len(hfInfos) != 0 {
		for _, h := range hfInfos {
			if h == nil {
				continue
			}
			logInfo = fmt.Sprintf("版本: %v(%v), 构建ID: %v", h.OrigVersion, h.OrigVersionCode, request.BuildID)
			break
		}
	}
	if logInfo != "" {
		if _, err = s.fkDao.AddLog(c, request.AppKey, request.Env, mngmdl.ModelHotpatch, mngmdl.OperationHotpatchConfigModify, logInfo, userName); err != nil {
			log.Error("AddLog error(%v)", err)
		}
	}
	return
}

// HotfixConfGet get hotfix config
func (s *Service) HotfixConfGet(c context.Context, request *appmdl.HfConfGetReq) (hfConf appmdl.HotfixConf, err error) {
	if hfConf, err = s.fkDao.HotfixConfGet(c, request.AppKey, request.Env, request.BuildID); err != nil {
		log.Error("s.fkDao.HotfixConfGet() failed. %v", err)
	}
	return
}

// HotfixEffect Sets whether the hotfix package takes effect
func (s *Service) HotfixEffect(c context.Context, request *appmdl.HfEffectReq) (err error) {
	var hotfixConfigID int64
	if hotfixConfigID, err = s.fkDao.GetHotfixConfigID(c, request.AppKey, request.Env, request.BuildID); err != nil {
		log.Error("s.fkDao.GetHotfixID() failed. %v", err)
		return
	}
	if hotfixConfigID == 0 {
		return ecode.Error(ecode.NothingFound, "热修配置信息数据不存在，请先进行配置操作")
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
	if _, err = s.fkDao.TxHotfixEffect(tx, request.AppKey, request.Env, request.BuildID, request.Effect); err != nil {
		log.Error("s.fkDao.TxHotfixEffect() failed. %v", err)
	}
	return
}

// HotfixList get hotfix list info
func (s *Service) HotfixList(c context.Context, request *appmdl.HfListReq) (hfList model.HfList, err error) {
	var (
		count int
		page  = model.PageInfo{}
		items []*appmdl.HfListItem
	)
	if count, err = s.fkDao.GetHotfixListCount(c, request.AppKey, request.Env); err != nil {
		log.Error("%v", err)
		return
	}
	if items, err = s.fkDao.GetHotfixList(c, request.AppKey, request.Env, request.Pn, request.Ps, request.Order, request.Sort); err != nil {
		log.Error("%v", err)
		return
	}
	for i := 0; i < len(items); i++ {
		var origin appmdl.HfOrigin
		if items[i].Commit != "" {
			items[i].ShortCommit = items[i].Commit[0:8]
		}
		if items[i].GlJobID != 0 {
			//items[i].JobURL = appmdl.JobURLPre + strconv.FormatInt(items[i].GlJobID, 10)
			items[i].JobURL = conf.Conf.Gitlab.Host + "/" + items[i].GlPrjID + "/-/jobs/" + strconv.FormatInt(items[i].GlJobID, 10)
		}
		if origin, err = s.fkDao.GetHotfixOrigin(c, request.AppKey, request.Env, items[i].OriginBuildID); err != nil {
			log.Error("%v", err)
			return
		}
		items[i].Origin = origin

		var config appmdl.HfConfig
		if config, err = s.fkDao.GetHotfixConfig(c, request.AppKey, request.Env, items[i].BuildID); err != nil {
			log.Error("%v", err)
			return
		}
		items[i].Config = config
	}
	page.Ps = request.Ps
	page.Pn = request.Pn
	page.Total = count
	hfList.PageInfo = page
	hfList.Items = items
	return
}

// HotfixBuild build hotfix patch
func (s *Service) HotfixBuild(c context.Context, request *appmdl.HfBuildReq, userName, envVars string) (id int64, err error) {
	var (
		tx                 *sql.Tx
		version            appmdl.HfOriginVersion
		appID, gitlabPrjID string
	)
	if appID, _, gitlabPrjID, err = s.fkDao.AppBasicInfo(c, request.AppKey); err != nil {
		log.Error("%v", err)
		return
	}
	if version, err = s.fkDao.GetOriginInfo(c, request.AppKey, "test", request.BuildID); err != nil {
		log.Error("s.fkDao.GetOriginInfo() error(%v)", err)
		return
	}
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
	if id, err = s.fkDao.TxAddHotfixBuild(tx, request.AppKey, appID, gitlabPrjID, request.BuildID, version.VersionCode, request.InternalVersionCode, request.GitType,
		request.GitName, "test", version.Version, userName, envVars); err != nil {
		log.Error("s.fkDao.TxAddHotfixBuild() failed. %v", err)
	}
	if _, err = s.fkDao.TxHotfixBuildIDUpdate(tx, id); err != nil { //更新build_id
		log.Error("s.fkDao.TxHotfixBuildIDUpdate() failed. %v", err)
	}
	return
}

// HotfixUpdate update hotfix t
func (s *Service) HotfixUpdate(c context.Context, request *appmdl.HfUpdateReq) (err error) {
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
	if _, err = s.fkDao.TxHotfixUpdate(tx, request.PatchBuildID, request.GlJobID, request.Commit); err != nil {
		log.Error("s.fkDao.TxHfConfEnvUpdate() failed. %v", err)
		return
	}
	return
}

// HotfixUpload upload hotfix file
func (s *Service) HotfixUpload(c context.Context, request *appmdl.HfUploadReq, file multipart.File, header *multipart.FileHeader) (err error) {
	var (
		tx                          *sql.Tx
		size, orgBuildID            int64
		fmd5, hfPath, hfURL, appKey string
		hfInfo                      *appmdl.HotfixInfo
		destFile, hfFile            *os.File
		fileInfo                    os.FileInfo
	)

	if hfInfo, err = s.fkDao.GetSingleHotfixInfo(c, request.PatchBuildID); err != nil {
		log.Error("s.fkDao.GetSingleHotfixInfo() failed. %v", err)
		return
	}
	appKey = hfInfo.AppKey
	orgBuildID = hfInfo.OrigBuildID
	destFileDir := filepath.Join(conf.Conf.LocalPath.LocalDir, "pack", appKey, strconv.FormatInt(orgBuildID, 10), "hotfix", strconv.FormatInt(request.PatchBuildID, 10))
	if err = os.MkdirAll(destFileDir, 0755); err != nil {
		log.Error("os.MkdirAll error(%v)", err)
		return
	}
	destFilePath := filepath.Join(destFileDir, header.Filename)
	if destFile, err = os.Create(destFilePath); err != nil {
		log.Error("os.Create() failed. error(%v)", err)
		return
	}
	if _, err = io.Copy(destFile, file); err != nil {
		log.Error("io.Copy error(%v)", err)
	}
	defer file.Close()
	defer destFile.Close()
	if _, err = file.Seek(0, io.SeekStart); err != nil {
		log.Error("file.Seek() failed. %v", err)
		return
	}
	if err = utils.Unzip(destFilePath, destFileDir); err != nil {
		log.Error("unzip(%s, %s) error(%v)", destFilePath, destFileDir, err)
		return
	}
	if err = os.Remove(destFilePath); err != nil {
		log.Error("os.Remove(%s) error(%v)", destFilePath, err)
		return
	}
	hfPath = filepath.Join(destFileDir, request.PatchName)
	hfURL = conf.Conf.LocalPath.LocalDomain + "/pack/" + appKey + "/" + strconv.FormatInt(orgBuildID, 10) + "/hotfix/" +
		strconv.FormatInt(request.PatchBuildID, 10) + "/" + request.PatchName

	if fileInfo, err = os.Stat(hfPath); err != nil {
		log.Error("os.Stat(%s) error(%v)", hfPath, err)
		return
	}
	size = fileInfo.Size()
	buf := new(bytes.Buffer)
	if hfFile, err = os.Open(hfPath); err != nil {
		log.Error("os.Open(%s) error(%v)", hfPath, err)
		return
	}
	if _, err = io.Copy(buf, hfFile); err != nil {
		log.Error("error(%v)", err)
		return
	}
	md5Bs := md5.Sum(buf.Bytes())
	fmd5 = hex.EncodeToString(md5Bs[:])
	var (
		url      string
		filename = path.Base(destFilePath)
		folder   = strings.Replace(strings.Replace(destFilePath, s.c.LocalPath.LocalDir, "", -1), filename, "", -1)
	)
	if url, _, _, err = s.fkDao.FilePutOss(c, folder, request.PatchName, appKey); err != nil {
		log.Error("FilePutOss failed. %v", err)
		return
	}
	if tx, err = s.fkDao.BeginTran(context.Background()); err != nil {
		log.Error("s.fkDao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
	}()
	if _, err = s.fkDao.TxHotfixUpload(tx, size, request.PatchBuildID, fmd5, hfPath, hfURL, url, appKey); err != nil {
		if err = tx.Rollback(); err != nil {
			log.Error("tx.Rollback error(%v)", err)
		}
		log.Error("%v", err)
		return
	}
	if err = tx.Commit(); err != nil {
		log.Error("tx.Commit() error(%v)", err)
		return
	}
	// 异步发送邮件和 saga 通知
	s.AddHotfixProc(func() {
		if err = s.notifyHotfixJobFinished(context.Background(), hfInfo, hfURL); err != nil {
			log.Error("notifyHotfixJobFinished error(%v)", err)
		}
	})
	return
}

// HotfixCancel cancel hotfix task
func (s *Service) HotfixCancel(c context.Context, request *appmdl.HfCancelReq) (err error) {
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
	if _, err = s.fkDao.TxHotfixCancel(tx, -2, request.AppKey, request.PatchBuildID); err != nil {
		log.Error("s.fkDao.HotfixCancel() failed. %v", err)
		return
	}
	return
}

// HotfixOrigGet is get hotfix origin package's URL information
func (s *Service) HotfixOrigGet(c context.Context, request *appmdl.HfOriginInfoReq) (resp appmdl.HfOrigURLInfo, err error) {
	if resp, err = s.fkDao.GetOriginURL(c, request.AppKey, request.PatchID); err != nil {
		if err == sql.ErrNoRows {
			err = ecode.NothingFound
			return
		}
		log.Error("s.fkDao.GetOriginURL() failed. %v", err)
	}
	return
}

// HotfixDel cancel hotfix task
func (s *Service) HotfixDel(c context.Context, request *appmdl.HfDelReq) (err error) {
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
	if _, err = s.fkDao.TxHotfixDel(tx, request.AppKey, request.PatchBuildID); err != nil {
		log.Error("s.fkDao.HotfixDel() failed. %v", err)
		return
	}
	return
}

func (s *Service) notifyHotfixJobFinished(c context.Context, hfInfo *appmdl.HotfixInfo, hotfixURL string) (err error) {
	var (
		cd                                         *cdmdl.Pack
		version                                    *model.Version
		mail                                       *mailmdl.Mail
		sagatoList                                 []string
		content, weChatContent, appID, gitlabPrjID string
		toAddresses, ccAddresses                   []*mailmdl.Address
	)
	if cd, err = s.fkDao.PackByBuild(c, hfInfo.AppKey, "prod", hfInfo.OrigBuildID); err != nil {
		log.Errorc(c, "notifyHotfixJobFinished: %v", err)
		return
	}
	if cd == nil {
		if cd, err = s.fkDao.PackByBuild(c, hfInfo.AppKey, "test", hfInfo.OrigBuildID); err != nil {
			log.Errorc(c, "notifyHotfixJobFinished: %v", err)
			return
		}
	}
	if cd == nil {
		log.Errorc(c, "appKey: %s buildId:%d 找不到数据", hfInfo.AppKey, hfInfo.OrigBuildID)
		return
	}
	if version, err = s.fkDao.PackVersionByID(c, cd.AppKey, cd.VersionID); err != nil {
		log.Error("notifyHotfixJobFinished: %v", err)
		return
	}
	if appID, _, gitlabPrjID, err = s.fkDao.AppBasicInfo(c, cd.AppKey); err != nil {
		log.Error("%v", err)
		return
	}
	orgGitlabJobURL := conf.Conf.Gitlab.Host + "/" + gitlabPrjID + "/-/jobs/" + strconv.FormatInt(cd.BuildID, 10)
	hfGitlabJobURL := conf.Conf.Gitlab.Host + "/" + gitlabPrjID + "/-/jobs/" + strconv.FormatInt(hfInfo.GlJobID, 10)
	toAddrStr := hfInfo.Operator + "@bilibili.com"
	toAddress := &mailmdl.Address{Address: toAddrStr}
	toAddresses = append(toAddresses, toAddress)
	sagatoList = append(sagatoList, hfInfo.Operator)

	mailTitle := fmt.Sprintf("【Fawkes】[%v] %v 热修包打包成功", gitlabPrjID, cd.GitName)
	content = fmt.Sprintf("<table width=\"100%%\"><tr><td width=\"70\" valign=\"top\" align=\"right\">App ID:</td><td>%v</td></tr>", appID)
	content += fmt.Sprintf("<tr><td valign=\"top\" align=\"right\">App Key:</td><td>%v</td></tr>", cd.AppKey)
	content += fmt.Sprintf("<tr><td valign=\"top\" align=\"right\">Version:</td><td>%v(%v)</td></tr>", version.Version, version.VersionCode)
	content += fmt.Sprintf("<tr><td valign=\"top\" align=\"right\">源包 Internal Ver:</td><td>%v</td></tr>", cd.InternalVersionCode)
	content += fmt.Sprintf("<tr><td valign=\"top\" align=\"right\">源包 commit:</td><td>%v</td></tr>", cd.Commit)
	content += fmt.Sprintf("<tr><td valign=\"top\" align=\"right\">源包 构建号:</td><td>%v</td></tr>", cd.BuildID)
	content += fmt.Sprintf("<tr><td valign=\"top\" align=\"right\">源包 job URL:</td><td>%v</td></tr>", orgGitlabJobURL)
	content += fmt.Sprintf("<tr><td valign=\"top\" align=\"right\">源包 URL:</td><td>%v</td></tr>", cd.PackURL)
	content += fmt.Sprintf("<tr><td valign=\"top\" align=\"right\">热修包 Internal Ver:</td><td>%v</td></tr>", hfInfo.InternalVersionCode)
	content += fmt.Sprintf("<tr><td valign=\"top\" align=\"right\">热修包 commit:</td><td>%v</td></tr>", hfInfo.Commit)
	content += fmt.Sprintf("<tr><td valign=\"top\" align=\"right\">热修包 构建号:</td><td>%v</td></tr>", hfInfo.GlJobID)
	content += fmt.Sprintf("<tr><td valign=\"top\" align=\"right\">热修包 job URL:</td><td>%v</td></tr>", hfGitlabJobURL)
	content += fmt.Sprintf("<tr><td valign=\"top\" align=\"right\">热修包 URL:</td><td>%v</td></tr>", hotfixURL)
	content += "</table>"
	mail = &mailmdl.Mail{Subject: mailTitle, Body: content, ToAddresses: toAddresses, CcAddresses: ccAddresses, Type: mailmdl.TypeTextHTML}
	if err = s.fkDao.SendMail(context.Background(), mail, nil, hfInfo.AppKey, mailmdl.HotfixFinishJobMail); err != nil {
		log.Error("notifyHotfixJobFinished: %v", err)
	}

	// wechat notification
	var (
		reqURL  string
		req     *http.Request
		data    []byte
		sagaReq *model.SagaReq
		sagaRes *model.SagaRes
	)
	weChatContent = fmt.Sprintf("【Fawkes】通知\n[%v] %v 热修包打包成功", gitlabPrjID, cd.GitName)
	weChatContent += fmt.Sprintf("App ID:%v\n", appID)
	weChatContent += fmt.Sprintf("App Key:%v\n", cd.AppKey)
	weChatContent += fmt.Sprintf("Version:%v(%v)\n", version.Version, version.VersionCode)
	weChatContent += fmt.Sprintf("源包 Internal Ver:%v\n", cd.InternalVersionCode)
	weChatContent += fmt.Sprintf("源包 commit:%v\n", cd.Commit)
	weChatContent += fmt.Sprintf("源包 构建号:%v\n", cd.BuildID)
	weChatContent += fmt.Sprintf("源包 job URL:%v\n", orgGitlabJobURL)
	weChatContent += fmt.Sprintf("源包 URL:%v\n", cd.PackURL)
	weChatContent += fmt.Sprintf("热修包 Internal Ver:%v\n", hfInfo.InternalVersionCode)
	weChatContent += fmt.Sprintf("热修包 commit:%v\n", hfInfo.Commit)
	weChatContent += fmt.Sprintf("热修包 构建号:%v\n", hfInfo.GlJobID)
	weChatContent += fmt.Sprintf("热修包 job URL:%v\n", hfGitlabJobURL)
	weChatContent += fmt.Sprintf("热修包 URL:%v\n", hotfixURL)
	sagaReq = &model.SagaReq{ToUser: sagatoList, Content: weChatContent}
	reqURL = conf.Conf.Host.Saga + "/ep/admin/saga/v2/wechat/message/send"
	if data, err = json.Marshal(sagaReq); err != nil {
		log.Error("s.SendMsg json marshal error(%v)", err)
		return
	}
	if req, err = http.NewRequest(http.MethodPost, reqURL, strings.NewReader(string(data))); err != nil {
		log.Error("s.SendMsg call http.NewRequest error(%v)", err)
		return
	}
	req.Header.Add("content-type", "application/json")
	if err = s.httpClient.Do(context.Background(), req, &sagaRes); err != nil {
		log.Error("s.SendMsg call client.Do error(%v)", err)
		return
	}
	return
}
