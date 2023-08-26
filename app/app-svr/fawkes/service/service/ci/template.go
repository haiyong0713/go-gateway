package ci

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"go-common/library/ecode"

	"go-gateway/app/app-svr/fawkes/service/conf"
	appmdl "go-gateway/app/app-svr/fawkes/service/model/app"
	cimdl "go-gateway/app/app-svr/fawkes/service/model/ci"
	mailmdl "go-gateway/app/app-svr/fawkes/service/model/mail"
	"go-gateway/app/app-svr/fawkes/service/model/template"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

var packTypeDict = map[int8]string{
	0:  "unknown",
	1:  "debug",
	2:  "release",
	3:  "enter",
	4:  "publish",
	5:  "fast-debug",
	6:  "enter-debug",
	7:  "fast-release",
	8:  "cover-debug",
	9:  "testflight",
	10: "app-bundle",
	11: "Assets",
}

// 拼接邮件通知
func (s *Service) combineMailNotify(c context.Context, app *appmdl.APP, ci *cimdl.BuildPack, notifyMailType cimdl.NotifyMail) (mail *mailmdl.Mail, err error) {
	var (
		mailtoList               []string
		content, appKey          string
		toAddresses, ccAddresses []*mailmdl.Address
	)
	if ci == nil {
		log.Errorc(c, "buildPack is nil")
		return
	}
	mailTitle := fmt.Sprintf("【Fawkes】[%v] %v 分支 ", app.Name, ci.GitName)
	toAddrStr := ci.Operator + "@bilibili.com"
	toAddress := &mailmdl.Address{Address: toAddrStr}
	toAddresses = append(toAddresses, toAddress)
	var packType = packTypeDict[ci.PkgType]
	switch notifyMailType {
	case cimdl.NotifyMailCC:
		mailTitle = mailTitle + fmt.Sprintf("%v 测试包更新", packType)
		if mailtoList, err = s.fkDao.AppMailtoList(c, appKey, mailmdl.CINotifyGroupMail, mailmdl.ReceiverWithCC); err != nil {
			log.Error("combineMailNotify AppMailtoList: %v", err)
			return
		}
		for _, mailto := range mailtoList {
			ccAddrStr := mailto + "@bilibili.com"
			ccAddress := &mailmdl.Address{Address: ccAddrStr}
			ccAddresses = append(ccAddresses, ccAddress)
		}
	default:
		mailTitle = mailTitle + fmt.Sprintf("%v 包打包成功", packType)
	}
	if content, err = s.combineMailContent(c, app, ci); err != nil {
		log.Errorc(c, "combineMailContent error(%v), build_id = %v", err, ci.BuildID)
	}
	mail = &mailmdl.Mail{Subject: mailTitle, Body: content, ToAddresses: toAddresses, CcAddresses: ccAddresses, Type: mailmdl.TypeTextHTML}
	return
}

// 拼接邮件content
func (s *Service) combineMailContent(c context.Context, app *appmdl.APP, ci *cimdl.BuildPack) (content string, err error) {
	var (
		groupName, projectName string
		mailContent            *template.CIBuildInfoTemplate
	)
	if ci == nil {
		log.Errorc(c, "buildPack is nil")
		return
	}
	strInt64 := strconv.FormatInt(ci.Size/1024, 10)
	id16, _ := strconv.Atoi(strInt64)
	var readableSize = fmt.Sprintf("%.2f MB", float32(id16)/1024.0)
	saveURLDir := conf.Conf.LocalPath.LocalDomain + "/pack/" + ci.AppKey + "/" + strconv.FormatInt(ci.GitlabJobID, 10)
	if len(ci.GitPath) > 0 {
		if strings.HasPrefix(ci.GitPath, "git@") {
			pathComps := strings.Split(ci.GitPath, ":")
			projectNameComp := strings.Split(pathComps[len(pathComps)-1], ".git")[0]
			projectNameComps := strings.Split(projectNameComp, "/")
			groupName = projectNameComps[0]
			projectName = projectNameComps[1]
		} else {
			pathComps := strings.Split(ci.GitPath, "/")
			projectName = strings.Split(pathComps[len(pathComps)-1], ".git")[0]
			groupName = pathComps[len(pathComps)-2]
		}
		ci.GitlabJobURL = conf.Conf.Gitlab.Host + "/" + groupName + "/" + projectName + "/-/jobs/" + strconv.FormatInt(ci.GitlabJobID, 10)
	}
	saveDir := filepath.Join(conf.Conf.LocalPath.LocalDir, "pack", ci.AppKey, strconv.FormatInt(ci.GitlabJobID, 10))
	urlComp := strings.Split(ci.PkgURL, "/")
	qrURLDir := strings.Join(urlComp[0:len(urlComp)-1], "/")
	changeLogHTML := strings.Replace(ci.ChangeLog, "\n", "<br/>", -1)
	mailContent = &template.CIBuildInfoTemplate{AppID: ci.AppID, AppKey: ci.AppKey, AppName: app.Name, Version: ci.Version, VersionCode: ci.VersionCode, InternalVersionCode: ci.InternalVersionCode, Commit: ci.Commit, GitlabJobID: ci.GitlabJobID,
		GitlabJobURL: ci.GitlabJobURL, ReadableSize: readableSize, SaveURLDir: saveURLDir, PkgURL: ci.PkgURL, QrURLDir: qrURLDir, ChangeLogHTML: changeLogHTML}
	err = filepath.Walk(saveDir, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			log.Error("filepath.Walk error(%v) build_id = %v", err, ci.BuildID)
			return err
		}
		if f == nil {
			errMsg := "found no file"
			err = fmt.Errorf(errMsg)
			log.Error(errMsg)
			return err
		}
		if f.IsDir() {
			return nil
		}
		// ignore tmp files
		if strings.HasPrefix(f.Name(), ".") {
			return nil
		}
		// 对 install.png 的特殊处理
		if f.Name() == "install.png" {
			fileURL := strings.Replace(path, conf.Conf.LocalPath.LocalDir, conf.Conf.LocalPath.LocalDomain, -1)
			//addFileUrl
			mailContent.AddFileUrl = true
			mailContent.FileURL = fileURL
		}
		return err
	})
	if err != nil {
		log.Errorc(c, "filepath.Walk error(%v)", err)
	}
	if content, err = s.fkDao.TemplateAlter(mailContent, template.CIBuildInfoTemp_Mail); err != nil {
		log.Errorc(c, "templateAlter error(%v)", err)
	}
	return
}

// 拼接企业微信content
func (s *Service) combineWeChatContent(c context.Context, app *appmdl.APP, ci *cimdl.BuildPack) (content string, err error) {
	var (
		weChatContent          *template.CIBuildInfoTemplate
		groupName, projectName string
	)
	if ci == nil {
		log.Errorc(c, "buildPack is nil")
		return
	}
	if app == nil {
		log.Errorc(c, "app is nil, build_id = %v", ci.BuildID)
		return
	}
	var packType = packTypeDict[ci.PkgType]
	strInt64 := strconv.FormatInt(ci.Size/1024, 10)
	id16, _ := strconv.Atoi(strInt64)
	var readableSize = fmt.Sprintf("%.2f MB", float32(id16)/1024.0)
	saveURLDir := conf.Conf.LocalPath.LocalDomain + "/pack/" + ci.AppKey + "/" + strconv.FormatInt(ci.GitlabJobID, 10)
	if len(ci.GitPath) > 0 {
		if strings.HasPrefix(ci.GitPath, "git@") {
			pathComps := strings.Split(ci.GitPath, ":")
			projectNameComp := strings.Split(pathComps[len(pathComps)-1], ".git")[0]
			projectNameComps := strings.Split(projectNameComp, "/")
			groupName = projectNameComps[0]
			projectName = projectNameComps[1]
		} else {
			pathComps := strings.Split(ci.GitPath, "/")
			projectName = strings.Split(pathComps[len(pathComps)-1], ".git")[0]
			groupName = pathComps[len(pathComps)-2]
		}
		ci.GitlabJobURL = conf.Conf.Gitlab.Host + "/" + groupName + "/" + projectName + "/-/jobs/" + strconv.FormatInt(ci.GitlabJobID, 10)
	}
	weChatContent = &template.CIBuildInfoTemplate{AppName: app.Name, GitName: ci.GitName, PackType: packType, AppID: ci.AppID, AppKey: ci.AppKey,
		Version: ci.Version, VersionCode: ci.VersionCode, InternalVersionCode: ci.InternalVersionCode, Commit: ci.Commit, GitlabJobID: ci.GitlabJobID,
		GitlabJobURL: ci.GitlabJobURL, ReadableSize: readableSize, SaveURLDir: saveURLDir, PkgURL: ci.PkgURL, ChangeLog: ci.ChangeLog, Result: template.CiResultString(int(ci.Status))}
	saveDir := filepath.Join(conf.Conf.LocalPath.LocalDir, "pack", ci.AppKey, strconv.FormatInt(ci.GitlabJobID, 10))
	err = filepath.Walk(saveDir, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			log.Errorc(c, "filepath.Walk error(%v), build_id = %v", err, ci.BuildID)
			return err
		}
		if f == nil {
			errMsg := "found no file"
			err = fmt.Errorf(errMsg)
			log.Error(errMsg)
			return err
		}
		if f.IsDir() {
			return nil
		}
		// ignore tmp files
		if strings.HasPrefix(f.Name(), ".") {
			return nil
		}
		// 对 install.png 的特殊处理
		if f.Name() == "install.png" {
			fileURL := strings.Replace(path, conf.Conf.LocalPath.LocalDir, conf.Conf.LocalPath.LocalDomain, -1)
			weChatContent.AddFileUrl = true
			weChatContent.FileURL = fileURL
		}
		return err
	})
	if err != nil {
		log.Errorc(c, "filepath.Walk error(%v), bulid_id = %v", err, ci.BuildID)
	}
	if content, err = s.fkDao.TemplateAlter(weChatContent, template.CIBuildInfoTemp_WeChat); err != nil {
		log.Errorc(c, " TemplateAlter error(%v), build_id = %v", err, ci.BuildID)
	}
	return
}

func (s *Service) notifyBuildInfo(c context.Context, buildID int64) (ci *cimdl.BuildPack, app *appmdl.APP, err error) {
	if ci, err = s.fkDao.BuildPackById(c, buildID); err != nil {
		log.Errorc(c, "BuildPackById error: %v build_id = %v", err, buildID)
		return
	}
	if ci == nil {
		log.Errorc(c, "buildPack is nil, build_id = %v", buildID)
		err = ecode.Error(ecode.RequestErr, fmt.Sprintf("buildPack is nil"))
		return
	}
	if app, err = s.fkDao.AppPass(c, ci.AppKey); err != nil {
		log.Errorc(c, "notifyBuildInfo: %v build_id = %v", err, buildID)
		return
	}
	if app == nil {
		log.Errorc(c, "app is nil, build_id = %v", buildID)
		err = ecode.Error(ecode.RequestErr, fmt.Sprintf("app is nil"))
		return
	}
	return
}
