package ci

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/md5"
	// nolint:gosec
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"go-common/library/database/sql"
	"go-common/library/ecode"
	xlog "go-common/library/log"

	"go-common/library/log/infoc.v2"
	"go-common/library/sync/errgroup.v2"

	"go-gateway/app/app-svr/fawkes/service/conf"
	"go-gateway/app/app-svr/fawkes/service/model"
	appmdl "go-gateway/app/app-svr/fawkes/service/model/app"
	cdmdl "go-gateway/app/app-svr/fawkes/service/model/cd"
	cimdl "go-gateway/app/app-svr/fawkes/service/model/ci"
	gitmdl "go-gateway/app/app-svr/fawkes/service/model/gitlab"
	mailmdl "go-gateway/app/app-svr/fawkes/service/model/mail"
	"go-gateway/app/app-svr/fawkes/service/model/template"
	tribemdl "go-gateway/app/app-svr/fawkes/service/model/tribe"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
	"go-gateway/app/app-svr/fawkes/service/tools/utils"
)

var validPkgType = map[int8]bool{1: true, 3: true, 5: true, 6: true, 8: true}

const _NasLogPrefix = "[NAS CLEAN]"

// PackInfo info of ci pack
func (s *Service) PackInfo(c context.Context, appKey string, buildID, glJobID int64) (res *cimdl.BuildPack, err error) {
	if res, err = s.fkDao.CIByBuild(c, appKey, buildID, glJobID); err != nil {
		log.Error("%v", err)
		return
	}
	if res != nil {
		res.SubRepos, _ = s.fkDao.BuildPackSubRepos(c, appKey, res.BuildID)
		res.GitlabJobURL = s.cdSvr.MakeGitPath(res.GitPath, res.GitlabJobID)
	}
	return
}

// BuildPackList get ci list
func (s *Service) BuildPackList(c context.Context, appKey string, pn, ps, pkgType, status, gitType int, gitKeyword,
	operator, order, sort string, gitlabJobID, ID int64, didPushCD string, hasBbrUrl bool) (builds *cimdl.ResultBuildPacks, err error) {
	var (
		groupName, projectName string
		total                  int
	)
	if total, err = s.fkDao.BuildPacksCount(c, appKey, pkgType, status, gitType, gitKeyword, operator, gitlabJobID, ID, didPushCD, hasBbrUrl); err != nil {
		log.Error("%v", err)
		return
	}
	if total < 1 {
		return
	}
	pageInfo := &cimdl.Page{Total: total, PageNum: pn, PageSize: ps}
	builds = &cimdl.ResultBuildPacks{PageInfo: pageInfo}

	if builds.Items, err = s.fkDao.BuildPacks(c, appKey, pn, ps, pkgType, status, gitType, gitKeyword, operator, order, sort, gitlabJobID, ID, didPushCD, hasBbrUrl); err != nil {
		log.Error("%v", err)
		return
	}
	for _, item := range builds.Items {
		//nolint:gomnd
		if len(item.Commit) > 8 {
			item.ShortCommit = string([]byte(item.Commit)[:8])
		}
		if len(item.GitPath) > 0 {
			if strings.HasPrefix(item.GitPath, "git@") {
				// git@git.bilibili.co:studio/android/bilibiliStudio.git
				pathComps := strings.Split(item.GitPath, ":")
				projectNameComp := strings.Split(pathComps[len(pathComps)-1], ".git")[0]
				projectNameComps := strings.Split(projectNameComp, "/")
				groupName = projectNameComps[0]
				i := 0
				projectName = strings.Join(append(projectNameComps[:i], projectNameComps[i+1:]...), "/")
			} else {
				pathComps := strings.Split(item.GitPath, "/")
				projectName = strings.Split(pathComps[len(pathComps)-1], ".git")[0]
				groupName = pathComps[len(pathComps)-2]
			}
			item.GitlabJobURL = conf.Conf.Gitlab.Host + "/" + groupName + "/" + projectName + "/-/jobs/" + strconv.FormatInt(item.GitlabJobID, 10)
		}
	}
	return
}

// RecordBuildPack record a ci build
func (s *Service) RecordBuildPack(c context.Context, appKey string, gitlabJobID int64, pkgType, gitType int, gitName,
	commit, version string, versionCode, internalVersionCode int64, operator string) (buildID int64, err error) {
	var (
		tx                          *sql.Tx
		appID, gitPath, gitlabPrjID string
	)
	if appID, gitPath, gitlabPrjID, err = s.fkDao.AppBasicInfo(c, appKey); err != nil {
		log.Error("%v", err)
		return
	}
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
	if buildID, err = s.fkDao.TxInsertBuildPack(tx, appKey, appID, gitPath, gitlabPrjID, gitlabJobID, pkgType,
		gitType, gitName, commit, version, versionCode, internalVersionCode, operator); err != nil {
		//nolint:errcheck
		tx.Rollback()
		log.Error("%v", err)
		return
	}
	if err = tx.Commit(); err != nil {
		log.Error("tx.Commit error(%v)", err)
	}
	return
}

// CreateBuildPack create a ci build
func (s *Service) CreateBuildPack(c context.Context, appKey, send string, pkgType, gitType int, gitName, operator, ciEnvVars, description, webhookURL string,
	shouldNotify bool, depGitlabJobId int64) (buildID int64, err error) {
	var (
		notifyGroup                 int8
		appID, gitPath, gitlabPrjID string
	)
	if appID, gitPath, gitlabPrjID, err = s.fkDao.AppBasicInfo(c, appKey); err != nil {
		log.Error("%v", err)
		return
	}
	if shouldNotify {
		notifyGroup = 1
	}
	err = s.fkDao.Transact(c, func(tx *sql.Tx) error {
		if buildID, err = s.fkDao.TxInsertBuildPackCreate(tx, appKey, appID, gitPath, gitlabPrjID, send, pkgType, gitType,
			gitName, operator, ciEnvVars, description, webhookURL, notifyGroup, depGitlabJobId); err != nil {
			log.Errorc(c, "TxInsertBuildPackCreate error %v", err)
			return err
		}
		return err
	})
	return
}

// CreateBuildPackCommon create a ci build common method
func (s *Service) CreateBuildPackCommon(c context.Context, appKey, send string, pkgType, gitType int, gitName, operator, ciEnvVar, description, webhookURL string, shouldNotify bool, dependGitlabJobID int64, tribeIds []int64) (buildID, resignBuildID int64, err error) {
	var (
		resignPkgType int
		updateErr     error
	)
	var variables = map[string]string{
		"TASK":                  "pack",
		"APP_KEY":               appKey,
		"PKG_TYPE":              strconv.Itoa(pkgType),
		"FAWKES":                "1",
		"FAWKES_USER":           operator,
		"TRIBE_HOST_BBR_JOB_ID": "",
	}
	if dependGitlabJobID != 0 {
		variables["TRIBE_PRE_BBR_JOB_ID"] = strconv.FormatInt(dependGitlabJobID, 10)
	}
	var envVarMap = make(map[string]string)
	if ciEnvVar != "" {
		if err = json.Unmarshal([]byte(ciEnvVar), &envVarMap); err != nil {
			log.Errorc(c, "json.Unmarshal(%s) error(%v)", ciEnvVar, err)
			return
		}
		for key, value := range envVarMap {
			variables[key] = value
		}
	}
	if buildID, err = s.CreateBuildPack(c, appKey, send, pkgType, gitType, gitName, operator, ciEnvVar, description, webhookURL, shouldNotify, dependGitlabJobID); err != nil {
		log.Errorc(c, "CreateBuildPack error(%v)", err)
		return
	}
	if len(tribeIds) != 0 {
		// 同时打tribe包
		var m map[string]int64
		if m, err = s.addTribeBuildRecord(c, tribeIds, pkgType, gitType, appKey, gitName, operator, ciEnvVar, description, shouldNotify); err != nil {
			log.Errorc(c, "%v", err)
			err = ecode.Error(ecode.ServerErr, "tribe打包失败")
			return
		}
		mapJson, err := json.Marshal(m)
		if err != nil {
			log.Errorc(c, "tribe json.Marshal error(%v)", err)
		}
		variables["TRIGGER_TRIBE_MAP"] = string(mapJson)
	}
	variables["BUILD_ID"] = strconv.FormatInt(buildID, 10)
	if envVarMap["RESIGN_TASK"] == "resign" {
		// 需重签
		if v, ok := variables["IPA_PATH"]; !ok || v == "" {
			// 新包
			resignPkgType, err = strconv.Atoi(envVarMap["RESIGN_PKG_TYPE"])
			if err != nil {
				log.Errorc(c, "resignPkgType strconv.Atoi error(%v)", err)
			}
			if resignBuildID, err = s.CreateBuildPack(c, appKey, send, resignPkgType, gitType, gitName, operator, ciEnvVar, fmt.Sprintf("【重签 FROM ID: %v】:%v", buildID, description), webhookURL, shouldNotify, dependGitlabJobID); err != nil {
				log.Errorc(c, "CreateBuildPack error(%v)", err)
				return
			}
			variables["RESIGN_BUILD_ID"] = strconv.FormatInt(resignBuildID, 10)
		} else {
			// 老包重签
			variables["RESIGN_BUILD_ID"] = strconv.FormatInt(buildID, 10)
			delete(variables, "TASK")
			// 旧包重签 因为原来的分支可能被删除 所以pipeline统一走master 或者 分支， 但是数据库存入的还是原来的分支
			customResignGitName := variables["CUSTOM_RESIGN_GIT_NAME"]
			if customResignGitName != "" {
				gitName = customResignGitName
			} else {
				gitName = "master"
			}
		}
	}
	_, err = s.gitSvr.TriggerPipeline(utils.CopyTrx(c), appKey, gitType, gitName, variables)
	if err != nil {
		log.Errorc(c, "TriggerPipeline error(%v)", err)
		updateErr = s.gitSvr.UpdateBuildPackStatus(c, buildID, cimdl.CIFailed)
		if resignBuildID != 0 {
			updateErr = s.gitSvr.UpdateBuildPackStatus(c, resignBuildID, cimdl.CIFailed)
		}
		if updateErr != nil {
			log.Errorc(c, "s.gitSvr.UpdateBuildPackStatus error(%v)", updateErr)
		}
	}
	return
}

func (s *Service) addTribeBuildRecord(ctx context.Context, tribeIds []int64, pkgType int, gitType int, appKey string, gitName string, operator string, envVar string, description string, notify bool) (tribeNameJobIdMap map[string]int64, err error) {
	var (
		appId       string
		gitPath     string
		gitlabPrjId string
	)
	tribeNameJobIdMap = make(map[string]int64)
	tribes, err := s.fkDao.SelectTribeByIds(ctx, tribeIds)
	if err != nil {
		return
	}
	if len(tribes) == 0 {
		log.Warnc(ctx, "appCiTribe-%s", fmt.Sprintf("tribeIds[%v] 没有查询到", tribeIds))
		return
	}
	if len(tribes) != len(tribeIds) {
		log.Errorc(ctx, "appCiTribe-%s", fmt.Sprintf("tribeIds[%v] 查询到%d条数据，数量不匹配", tribeIds, len(tribes)))
		return
	}
	if appId, gitPath, gitlabPrjId, err = s.fkDao.AppBasicInfo(ctx, appKey); err != nil {
		return
	}
	for _, v := range tribes {
		var packId int64
		if packId, err = s.fkDao.AddTribeBuildPack(ctx, v.Id, 0, int64(pkgType), int64(gitType), appKey, gitName, operator, envVar, description, notify, appId, gitPath, gitlabPrjId, tribemdl.CiInWaiting); err != nil {
			return
		}
		tribeNameJobIdMap[v.Name] = packId
	}
	return
}

// UpdateBuildPackBase update build pack base info.
func (s *Service) UpdateBuildPackBase(c context.Context, buildID, gitlabJobID int64, commit, version string,
	versionCode, internalVersionCode int64) (err error) {
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
	if _, err = s.fkDao.TxUpdateBuildPackBaseInfo(tx, buildID, gitlabJobID, commit, version, versionCode, internalVersionCode); err != nil {
		//nolint:errcheck
		tx.Rollback()
		log.Error("%v", err)
		return
	}
	if err = tx.Commit(); err != nil {
		log.Error("tx.Commit error(%v)", err)
	}
	return
}

// UploadBuildPack upload build pack
func (s *Service) UploadBuildPack(c context.Context, buildID int64, file multipart.File, header *multipart.FileHeader,
	unzip bool, pkgName, mappingName, bbrName, rName, rMappingName, changeLog string, subrepoCommits []*cimdl.BuildPackSubRepo) (err error) {
	var (
		size                                                                   int64
		fmd5, pkgPath, pkgURL, mappingURL, rURL, rMappingURL, bbrUrl, features string
		destFile, pkgFile                                                      *os.File
		fileInfo                                                               os.FileInfo
		metaJson                                                               map[string]interface{}
		compatibleVersions                                                     []*cimdl.Feature
		ci                                                                     *cimdl.BuildPack
	)
	if ci, err = s.fkDao.BuildPackById(c, buildID); err != nil {
		log.Errorc(c, "BuildPackById error: %v build_id = %v", err, buildID)
		return
	}
	if ci == nil {
		log.Errorc(c, "buildPack is nil, build_id = %v", buildID)
		return
	}
	destFileDir := filepath.Join(conf.Conf.LocalPath.LocalDir, "pack", ci.AppKey, strconv.FormatInt(ci.GitlabJobID, 10))
	pkgPath = filepath.Join(destFileDir, pkgName)
	// 文件存在则直接返回成功
	if _, err = os.Stat(pkgPath); err == nil {
		log.Errorc(c, "文件已存在, buildID = %v", buildID)
		return
	}
	if err = os.MkdirAll(destFileDir, 0755); err != nil {
		log.Errorc(c, "os.MkdirAll error(%v)", err)
		return
	}
	destFilePath := filepath.Join(destFileDir, header.Filename)

	if destFile, err = os.Create(destFilePath); err != nil {
		log.Errorc(c, "os.Create error(%v)", err)
		return
	}
	if _, err = io.Copy(destFile, file); err != nil {
		log.Errorc(c, "io.Copy error(%v)", err)
		return
	}
	defer file.Close()
	defer destFile.Close()
	if _, err = file.Seek(0, io.SeekStart); err != nil {
		log.Errorc(c, "file.Seek error(%v)", err)
		return
	}
	if unzip {
		if err = utils.Unzip(destFilePath, destFileDir); err != nil {
			log.Errorc(c, "unzip(%s, %s) error(%v)", destFilePath, destFileDir, err)
			return
		}
		if err = os.Remove(destFilePath); err != nil {
			log.Errorc(c, "os.Remove(%s) error(%v)", destFilePath, err)
			return
		}
	}
	saveDir := conf.Conf.LocalPath.LocalDomain + "/pack/" + ci.AppKey + "/" + strconv.FormatInt(ci.GitlabJobID, 10)
	pkgURL = saveDir + "/" + pkgName
	if mappingName != "" {
		mappingURL = saveDir + "/" + mappingName
	} else {
		mappingURL = ""
	}
	if rName != "" {
		rURL = saveDir + "/" + rName
	} else {
		rURL = ""
	}
	if rMappingName != "" {
		rMappingURL = saveDir + "/" + rMappingName
	} else {
		rMappingURL = ""
	}
	if bbrName != "" {
		bbrUrl = saveDir + "/" + bbrName
	} else {
		bbrUrl = ""
	}
	if fileInfo, err = os.Stat(pkgPath); err != nil {
		log.Errorc(c, "os.Stat(%s) error(%v)", pkgPath, err)
		return
	}
	size = fileInfo.Size()
	buf := new(bytes.Buffer)
	if pkgFile, err = os.Open(pkgPath); err != nil {
		log.Errorc(c, "os.Open(%s) error(%v)", pkgPath, err)
		return
	}
	if _, err = io.Copy(buf, pkgFile); err != nil {
		log.Errorc(c, "io.Copy error(%v)", err)
		return
	}
	md5Bs := md5.Sum(buf.Bytes())
	fmd5 = hex.EncodeToString(md5Bs[:])
	if len(bbrName) != 0 {
		bbrPath := conf.Conf.LocalPath.LocalDir + "/pack/" + ci.AppKey + "/" + strconv.FormatInt(ci.GitlabJobID, 10) + "/" + bbrName
		metaJson, err = getMetaJson(bbrPath)
		if err != nil {
			log.Errorc(c, "getMetaJson error(%v)", err)
			return
		}
		compatibleVersions = getCompatibleVersions(metaJson)
		features = getFeatures(metaJson)
	}
	err = s.fkDao.Transact(c, func(tx *sql.Tx) error {
		if _, err = s.fkDao.TxUpdateBuildPackInfo(tx, buildID, size, fmd5, pkgPath, pkgURL, mappingURL, rURL, rMappingURL, bbrUrl, changeLog, features, len(compatibleVersions) > 0); err != nil {
			log.Errorc(c, "TxUpdateBuildPackInfo error(%v)", err)
			return err
		}
		if len(subrepoCommits) > 0 {
			log.Warnc(c, "subrepoCommits = %v", subrepoCommits)
			_, err = s.fkDao.TXAddCISubRepoCommits(tx, buildID, ci.GitlabJobID, ci.AppKey, subrepoCommits)
		}
		return err
	})
	// pkgDir 与 destFileDir 不同，因为 pkgName 可能含多层目录，pkgDir 只取包文件当前目录
	pkgDir := filepath.Dir(pkgPath)
	// 异步生成二维码，并发送邮件
	s.AddCiProc(func() {
		_ = s.generateQRCode(context.Background(), pkgURL, pkgDir)
		_ = s.UploadBuildPackNotify(c, buildID, ci.Operator, ci.Send)
	})
	log.Infoc(c, "UploadBuildPack done buildID=%v", buildID)
	if bbrName != "" {
		dstPath := path.Join(conf.Conf.LocalPath.LocalDir, "maven", "v1", ci.AppKey, "host", strconv.FormatInt(ci.GitlabJobID, 10))
		if err = s.saveMaven(path.Join(destFileDir, bbrName), dstPath, ci.AppKey, "host", ci.GitlabJobID); err != nil {
			log.Errorc(c, "save maven error(%v)", err)
		}
	}
	if len(compatibleVersions) != 0 {
		// 记录兼容关系
		if err = s.fkDao.BatchAddTribeHostRelation(c, ci.AppKey, ci.GitlabJobID, compatibleVersions); err != nil {
			log.Errorc(c, "add tribe host relation error %v", err)
		}
	}
	return
}

func (s *Service) UploadBuildPackNotify(c context.Context, buildId int64, operator, send string) (err error) {
	notify := &cimdl.NotifyCI{
		BuildId:      buildId,
		NotifyMail:   cimdl.NotifyMailRecipient,
		IsNotifyBot:  true,
		IsNotifyUser: true,
		IsNotifyHook: true,
	}
	if send == "" {
		send = cimdl.SendJsonDefault
	}
	if err = json.Unmarshal([]byte(send), &notify.Receiver); err != nil {
		log.Errorc(c, "json.Unmarshal error %v, build_id = %v", err, buildId)
		return
	}
	users := strings.Split(notify.Receiver.Users, ",")
	users = append(users, operator)
	notify.Receiver.Users = strings.Join(users, ",")
	_ = s.NotifyCIJob(context.Background(), notify)
	return
}

func getFeatures(metaJson map[string]interface{}) string {
	var features []string
	if _, ok := metaJson[cimdl.MetaKeyCompatible]; ok {
		features = append(features, "default")
	}
	if _, ok := metaJson[cimdl.MetaKeyFeatures]; ok {
		// 新版metaJson
		fi := metaJson[cimdl.MetaKeyFeatures].([]interface{})
		for _, v := range fi {
			vm := v.(map[string]interface{})
			if _, ok1 := vm[cimdl.MetaKeyFeatureName]; ok1 {
				features = append(features, vm[cimdl.MetaKeyFeatureName].(string))
			}
		}
	}
	return strings.Join(features, cimdl.Comma)
}

// UploadBuildFile upload build file
func (s *Service) UploadBuildFile(c context.Context, appKey string, jobID int64, file multipart.File, header *multipart.FileHeader,
	unzip bool) (err error) {
	var (
		destFile *os.File
	)
	destFileDir := filepath.Join(conf.Conf.LocalPath.LocalDir, "pack", appKey, strconv.FormatInt(jobID, 10))
	// 若文件夹不存在. 则新建一个文件夹
	if _, err = os.Stat(destFileDir); err != nil {
		if os.IsNotExist(err) {
			_ = os.MkdirAll(destFileDir, 0755)
		}
	}
	destFilePath := filepath.Join(destFileDir, header.Filename)
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
	if _, err = file.Seek(0, io.SeekStart); err != nil {
		log.Error("%v", err)
		return
	}
	if unzip {
		if err = utils.Unzip(destFilePath, destFileDir); err != nil {
			log.Error("unzip(%s, %s) error(%v)", destFilePath, destFileDir, err)
			return
		}
		if err = os.Remove(destFilePath); err != nil {
			log.Error("os.Remove(%s) error(%v)", destFilePath, err)
			return
		}
	}
	return
}

func (s *Service) UploadMobileEPBusiness(c context.Context, appKey, business, dirname, fmd5 string, file multipart.File, header *multipart.FileHeader, unzip bool) (err error) {
	var (
		destFile     *os.File
		destFilePath string
	)
	destFileDir := filepath.Join(conf.Conf.LocalPath.LocalDir, "mobile-ep", business, appKey, dirname)
	// 若文件夹不存在. 则新建一个文件夹
	if _, err = os.Stat(destFileDir); err != nil {
		if os.IsNotExist(err) {
			_ = os.MkdirAll(destFileDir, 0755)
		}
	}
	// 需要解压的文件使用md5来命名zip包 防止同名文件操作异常
	if unzip {
		destFilePath = filepath.Join(destFileDir, fmd5+path.Ext(header.Filename))
	} else {
		destFilePath = filepath.Join(destFileDir, header.Filename)
	}
	if destFile, err = os.Create(destFilePath); err != nil {
		log.Errorc(c, "%v", err)
		return
	}
	if _, err = io.Copy(destFile, file); err != nil {
		log.Errorc(c, "%v", err)
		return
	}
	defer file.Close()
	defer destFile.Close()
	if _, err = file.Seek(0, io.SeekStart); err != nil {
		log.Errorc(c, "%v", err)
		return
	}
	if unzip {
		if err = utils.Unzip(destFilePath, destFileDir); err != nil {
			log.Errorc(c, "unzip(%s, %s) error(%v)", destFilePath, destFileDir, err)
			return
		}
		if err = os.Remove(destFilePath); err != nil {
			log.Errorc(c, "os.Remove(%s) error(%v)", destFilePath, err)
			return
		}
	}
	return
}

func (s *Service) generateQRCode(c context.Context, url, outPath string) (err error) {
	var (
		out    bytes.Buffer
		errOut bytes.Buffer
	)
	qrPath := outPath + "/install.png"
	if _, err = os.Stat(qrPath); err == nil {
		log.Warnc(c, "file exist")
		return
	}
	cmd := exec.Command("qrencode", "-o", qrPath, url)
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	if err = cmd.Run(); err != nil {
		log.Errorc(c, "Command Run stdout=(%s) stderr=(%s) error(%v)", out.String(), errOut.String(), err)
		return
	}
	return
}

// CancelBuildPack cancel a ci build
func (s *Service) CancelBuildPack(c context.Context, buildID int64) (err error) {
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
	if _, err = s.fkDao.TxUpdateBuildPackStatus(tx, buildID, -2); err != nil {
		//nolint:errcheck
		tx.Rollback()
		log.Error("%v", err)
		return
	}
	if err = tx.Commit(); err != nil {
		log.Error("tx.Commit error(%v)", err)
	}
	return
}

// DeleteBuildPack delete a ci build
func (s *Service) DeleteBuildPack(c context.Context, buildID int64) (err error) {
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
	if _, err = s.fkDao.TxDelBuild(tx, buildID); err != nil {
		//nolint:errcheck
		tx.Rollback()
		log.Error("%v", err)
		return
	}
	if err = tx.Commit(); err != nil {
		log.Error("tx.Commit error(%v)", err)
	}
	return
}

// NotifyCIJob ci notify job
func (s *Service) NotifyCIJob(c context.Context, notify *cimdl.NotifyCI) (err error) {
	if notify == nil {
		log.Errorc(c, "通知为空")
		return
	}
	buildId := notify.BuildId
	if notify.Receiver == nil {
		log.Errorc(c, "接收人为空, build_id = %v", buildId)
		return
	}
	var (
		ci  *cimdl.BuildPack
		app *appmdl.APP
	)
	if ci, err = s.fkDao.BuildPackById(c, buildId); err != nil {
		log.Errorc(c, "CI构建包查询错误 error(%v), build_id = %v", err, buildId)
		return
	}
	if ci == nil {
		log.Errorc(c, "CI构建包为空, build_id = %v", buildId)
		return
	}
	if app, err = s.fkDao.AppPass(c, ci.AppKey); err != nil {
		log.Errorc(c, "App查询错误 error(%v), build_id = %v", err, buildId)
		return
	}
	if app == nil {
		log.Errorc(c, "App查询为空, build_id = %v", buildId)
		return
	}
	log.Infoc(c, "开始推送消息 - build_id = %v ", buildId)
	eg := errgroup.WithContext(c)
	// mail notification
	if notify.NotifyMail != cimdl.NotifyMailDisable {
		eg.Go(func(ctx context.Context) error {
			log.Infoc(ctx, "邮件发送开始 - build_id = %v", buildId)
			var mail *mailmdl.Mail
			if mail, err = s.combineMailNotify(c, app, ci, notify.NotifyMail); err != nil {
				log.Errorc(ctx, "邮件信息拼接失败 error(%v), build_id = %v", err, buildId)
				return err
			}
			if err = s.fkDao.SendMail(context.Background(), mail, nil, app.AppKey, mailmdl.CINotifyGroupMail); err != nil {
				log.Errorc(ctx, "邮件推送失败 error(%v), build_id = %v", err, buildId)
				return err
			}
			log.Infoc(ctx, "邮件发送成功 - build_id = %v", buildId)
			return err
		})
	}
	var weChatContent string
	// ep wechat notification
	if notify.IsNotifyUser && notify.Receiver.Users != "" {
		eg.Go(func(ctx context.Context) error {
			log.Infoc(ctx, "企微EP消息推送开始 - build_id = %v", buildId)
			if weChatContent, err = s.combineWeChatContent(c, app, ci); err != nil {
				log.Errorc(ctx, "企微EP消息拼接失败 error(%v), build_id = %v", err, buildId)
				return err
			}
			if err = s.fkDao.WechatEPNotify(weChatContent, notify.Receiver.Users); err != nil {
				log.Errorc(ctx, "企微EP消息推送失败 error(%v), build_id = %v", err, buildId)
				return err
			}
			log.Infoc(ctx, "企微EP消息推送成功 - build_id = %v", buildId)
			return err
		})
	}
	// bot notification
	if notify.IsNotifyBot && notify.Receiver.Bots != "" {
		eg.Go(func(ctx context.Context) error {
			log.Infoc(ctx, "企微机器人消息推送开始 - build_id = %v", buildId)
			if weChatContent, err = s.combineWeChatContent(c, app, ci); err != nil {
				log.Errorc(ctx, "企微机器人消息拼接失败 error(%v), build_id = %v", err, buildId)
				return err
			}
			if err = s.NotifyBot(ctx, weChatContent, notify.Receiver.Bots); err != nil {
				log.Errorc(ctx, "企微机器人消息推送失败 error(%v), build_id = %v", err, buildId)
				return err
			}
			log.Infoc(ctx, "企微机器人消息推送成功 - build_id = %v", buildId)
			return err
		})
	}
	// hook notification
	if notify.IsNotifyHook && notify.Receiver.Webhook != nil {
		eg.Go(func(ctx context.Context) error {
			log.Infoc(c, "hook消息推送开始 - build_id = %v", buildId)
			params := &cimdl.HookParam{
				AppKey:      app.AppKey,
				AppName:     app.Name,
				BuildID:     ci.BuildID,
				GitlabJobID: ci.GitlabJobID,
				CTime:       ci.CTime,
				PackURL:     ci.PkgURL,
			}
			if err = s.fkDao.Hook(c, params, notify.Receiver.Webhook); err != nil {
				log.Errorc(c, "hook消息推送失败 error(%v), build_id = %v", err, buildId)
				return err
			}
			log.Infoc(c, "hook消息推送成功 - build_id = %v", buildId)
			return err
		})
	}
	if err = eg.Wait(); err != nil {
		log.Errorc(c, "eg.Wait error(%v), build_id = %v", err, buildId)
	}
	log.Infoc(c, "结束推送消息 - build_id = %v ", buildId)
	return
}

// SendMail send mail
func (s *Service) SendMail(c context.Context, appKey, funcModule string, m *mailmdl.Mail, attach *mailmdl.Attach) (err error) {
	if err = s.fkDao.SendMail(c, m, attach, appKey, funcModule); err != nil {
		log.Errorc(c, "%v", err)
	}
	return
}

// UpdateTestStatus Update Test Status
func (s *Service) UpdateTestStatus(c context.Context, pack *cimdl.BuildPack) (err error) {
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
	if _, err = s.fkDao.TxUpdateTestStatus(tx, pack); err != nil {
		//nolint:errcheck
		tx.Rollback()
		log.Error("%v", err)
		return
	}
	if err = tx.Commit(); err != nil {
		log.Error("tx.Commit error(%v)", err)
	}
	return
}

// CrontabList list of crontab jobs
func (s *Service) CrontabList(c context.Context, appKey string, pn, ps int) (res *cimdl.ContabResult, err error) {
	var (
		total    int
		crontabs []*cimdl.Contab
	)
	g := errgroup.WithContext(c)
	g.Go(func(ctx context.Context) (err error) {
		if total, err = s.fkDao.CiCrontabCount(ctx, appKey); err != nil {
			log.Error("%v", err)
		}
		return
	})
	g.Go(func(ctx context.Context) (err error) {
		if crontabs, err = s.fkDao.CiCrontabList(ctx, appKey, pn, ps); err != nil {
			log.Error("%v", err)
		}
		return
	})
	if err = g.Wait(); err != nil {
		log.Error("%v", err)
		return
	}
	if total == 0 || len(crontabs) == 0 {
		return
	}
	res = &cimdl.ContabResult{
		PageInfo: &model.PageInfo{
			Total: total,
			Pn:    pn,
			Ps:    ps,
		},
		Items: crontabs,
	}
	return
}

// CrontabAdd add a crontab job
func (s *Service) CrontabAdd(c context.Context, appKey, stime, tick string, gitType int, gitName string, pkgType int, send, envVars, userName string) (err error) {
	err = s.fkDao.Transact(c, func(tx *sql.Tx) error {
		if _, err = s.fkDao.TxAddCiCrontab(tx, appKey, stime, tick, gitType, gitName, pkgType, send, envVars, userName); err != nil {
			log.Errorc(c, "%v", err)
		}
		return err
	})
	return
}

// CrontabStatus change crontab status
func (s *Service) CrontabStatus(c context.Context, id int64, status int) (err error) {
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
	if _, err = s.fkDao.TxUpStatusCiCrontab(tx, id, status); err != nil {
		log.Error("%v", err)
		return
	}
	if status == cimdl.CronStop {
		log.Info("cron_ci stop cronID(%v)", id)
		s.crontabCIProc.Store(id, false)
	}
	return
}

func (s *Service) cronNotify(c context.Context, buildID int64) (err error) {
	var (
		ci     *cimdl.BuildPack
		app    *appmdl.APP
		cron   *cimdl.Contab
		appKey string
	)
	if ci, err = s.fkDao.BuildPackById(c, buildID); err != nil {
		log.Errorc(c, "BuildPackById error(%v), build_id = %v", err, buildID)
		return
	}
	if ci == nil {
		log.Errorc(c, " buildPack is nil, build_id = %v", buildID)
		return
	}
	if app, err = s.fkDao.AppPass(c, appKey); err != nil {
		log.Errorc(c, "AppPass error(%v), build_id = %v", err, buildID)
		return
	}
	if cron, err = s.fkDao.CronInfo(c, appKey, buildID); err != nil {
		log.Errorc(c, "CronInfo error(%v), build_id = %v", err, buildID)
		return
	}
	if cron == nil {
		log.Warnc(c, "cron is nil, build_id = %v", buildID)
		return
	}
	var param *model.Broadcast
	if err = json.Unmarshal([]byte(cron.Send), &param); err != nil {
		log.Errorc(c, "json.Unmarshal error(%v), build_id = %v", err, buildID)
		return
	}
	param.Param = &cimdl.HookParam{
		AppKey:      appKey,
		AppName:     app.Name,
		BuildID:     buildID,
		GitlabJobID: ci.GitlabJobID,
		CTime:       ci.CTime,
		PackURL:     ci.PkgURL,
	}
	log.Warnc(c, "broadcast start build_id = %v", buildID)
	if err = s.Broadcast(context.Background(), "ci", param); err != nil {
		log.Errorc(c, "Broadcast error(%v), build_id = %v", err, buildID)
		return
	}
	log.Warnc(c, "broadcast success build_id = %v", buildID)
	return
}

// CrontabDel delete a crontab job
func (s *Service) CrontabDel(c context.Context, id int64) (err error) {
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
	if _, err = s.fkDao.TxDelCiCrontab(tx, id); err != nil {
		log.Error("%v", err)
		return
	}
	log.Info("cron_ci del cronID(%v)", id)
	s.crontabCIProc.Store(id, false)
	return
}

func (s *Service) PackReportInfo(c context.Context, appKey string, jobID int64) (jsonString string, err error) {
	packUrl := fmt.Sprintf("%s/%s", s.c.LocalPath.LocalDomain, filepath.Join("/pack", appKey, strconv.FormatInt(jobID, 10), "pack_report.json"))
	if jsonString, err = s.fkDao.MacrossFileInfo(c, packUrl); err != nil {
		log.Error("MacrossFileInfo error %v", err)
	}
	return
}

func (s *Service) GetMonkeyList(c context.Context, appKey string, buildId int64, pn, ps int) (res []*cimdl.EPMonkey, err error) {
	if res, err = s.fkDao.GetMonkeyList(c, appKey, buildId, pn, ps); err != nil {
		log.Error("%v", err)
	}
	return
}

func (s *Service) AddMonkey(c context.Context, appKey, osver, schemeUrl, messageTo, userName string, buildId int64, execDuration int) (err error) {
	var (
		tx           *sql.Tx
		pack         *cimdl.BuildPack
		monkeyTestId int64
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
	if pack, err = s.fkDao.CIByBuild(c, appKey, buildId, 0); err != nil {
		log.Error("%v", err)
		return
	}
	if monkeyTestId, err = s.fkDao.TxAddMonkey(tx, appKey, osver, schemeUrl, messageTo, userName, buildId, 0, execDuration); err != nil {
		log.Error("%v", err)
		return
	}
	if err = s.fkDao.JenkinsJobMonkey(c, appKey, pack.PkgURL, pack.MappingURL, osver, pack.AppID, schemeUrl, userName, execDuration, monkeyTestId); err != nil {
		log.Error("%v", err)
	}
	return
}

func (s *Service) UpdateMonkeyStatus(c context.Context, appKey, logUrl string, emulators map[string][]string, id int64, status int) (err error) {
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
	var playUrl string
	for _, params := range emulators {
		if len(params) > 0 {
			playUrl = params[0]
		}
	}
	if err = s.fkDao.UpdateMonkeyStatus(tx, appKey, logUrl, playUrl, id, status); err != nil {
		log.Error("%v", err)
	}
	return
}

func (s *Service) CiEnvList(c context.Context, envKey, appKey, platform string) (res []*cimdl.BuildEnvs, err error) {
	if res, err = s.fkDao.CiEnvList(c, envKey, appKey, platform); err != nil {
		log.Error("CiEnvList Err: %v", err)
	}
	return
}

func (s *Service) AddCiEnv(c context.Context, envKey, envValues, userName string, envType int) (err error) {
	var (
		envValuesList []*cimdl.EnvValue
		tx            *sql.Tx
	)
	if err = json.Unmarshal([]byte(envValues), &envValuesList); err != nil {
		log.Error("AddCiEnv json.Unmarshal(%s) error(%v)", string(envValues), err)
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
	for _, env := range envValuesList {
		if err = s.fkDao.TxAddCiEnv(tx, envKey, env.EnvVal, env.Description, env.Platform, env.AppKeys, userName, envType, env.IsDefault, env.IsGlobal, env.PushCDAble); err != nil {
			log.Error("TxAddCiEnv: %v", err)
			return
		}
	}
	return
}

func (s *Service) UpdateCiEnv(c context.Context, envKey, envValues, userName string, envType int) (err error) {
	var (
		envValuesList []*cimdl.EnvValue
		tx            *sql.Tx
	)
	if err = json.Unmarshal([]byte(envValues), &envValuesList); err != nil {
		log.Error("UpdateCiEnv json.Unmarshal(%s) error(%v)", string(envValues), err)
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
	for _, env := range envValuesList {
		if env.ID == 0 {
			if err = s.fkDao.TxAddCiEnv(tx, envKey, env.EnvVal, env.Description, env.Platform, env.AppKeys, userName, envType, env.IsDefault, env.IsGlobal, env.PushCDAble); err != nil {
				log.Error("TxAddCiEnv: %v", err)
				return
			}
		} else {
			if err = s.fkDao.TxUpdateCiEnv(tx, envKey, env.EnvVal, env.Description, env.Platform, env.AppKeys, userName, envType, env.IsDefault, env.IsGlobal, env.PushCDAble, env.ID); err != nil {
				log.Error("TxUpdateCiEnv: %v", err)
				return
			}
		}
	}
	return
}

func (s *Service) DeleteCiEnv(c context.Context, id int64, envKey string) (err error) {
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
	if err = s.fkDao.TxDeleteCiEnv(tx, id, envKey); err != nil {
		log.Error("TxDeleteCiEnv: %v", err)
	}
	return
}

func (s *Service) DeleteCiEnvByAppKey(c context.Context, envKey, appKeys, userName string) (err error) {
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
	if err = s.fkDao.TxDeleteCiEnvByAppKey(tx, envKey, appKeys, userName); err != nil {
		log.Error("TxDeleteCiEnvByAppKey: %v", err)
	}
	return
}

func (s *Service) RecordCIJob(c context.Context, params *cimdl.JobRecordParam) (err error) {
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
	info, _ := s.fkDao.AppPass(c, params.AppKey)
	jobURL := s.cdSvr.MakeGitPath(info.GitPath, params.JobID)
	if err = s.fkDao.TxRecordCIJob(tx, params, jobURL); err != nil {
		log.Error("TxRecordCIJob: %v", err)
	}
	return
}

func (s *Service) CIJobInfo(c context.Context, typeName string) (res []string, err error) {
	if res, err = s.fkDao.CIJobInfo(c, typeName); err != nil {
		log.Error("CIJobInfo Err: %v", err)
	}
	return
}

func (s *Service) RecordCICompile(c context.Context, params *cimdl.CICompileRecordParam) (err error) {
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
	if err = s.fkDao.TxRecordCICompile(tx, params); err != nil {
		log.Error("RecordCICompile: %v", err)
	}
	return
}

// DeleteCINas 删除CI文件并且更新过期状态
func (s *Service) DeleteCINas(c context.Context, req *cimdl.CISpecifyTimeDeleteReq) (res *cimdl.CISpecifyTimeDeleteResp, err error) {
	var (
		fileDeleted []int64
		bp          *cimdl.BuildPack
	)
	now := time.Now()
	res = &cimdl.CISpecifyTimeDeleteResp{}
	for _, deleteKey := range req.DeleteKeys {
		if deleteKey.AppKey == "" || deleteKey.BuildId == 0 {
			continue
		}
		if bp, err = s.fkDao.BuildPack(c, deleteKey.AppKey, deleteKey.BuildId); err != nil {
			log.Error("%s BuildPack error(%v)", _NasLogPrefix, err)
			continue
		}
		if now.AddDate(0, -6, 0).Before(time.Unix(bp.CTime, 0)) {
			log.Error("%s buildId: %d CTime: %s 在当前时间六个月之内创建的包不可删除。", _NasLogPrefix, deleteKey.BuildId, time.Unix(bp.CTime, 0))
			continue
		}
		if _, ok := validPkgType[bp.PkgType]; !ok {
			log.Error("%s DeleteFile appKey[%s], buildID[%d] pkgType error(pkgType=%v)", _NasLogPrefix, deleteKey.AppKey, deleteKey.BuildId, bp.PkgType)
			continue
		}
		filePath := filepath.Join(conf.Conf.LocalPath.LocalDir, "pack", deleteKey.AppKey, strconv.FormatInt(bp.GitlabJobID, 10))
		if !fileExists(filePath) {
			log.Error("%s appKey[%s], buildID[%d] pkgType[%d], filePath: %s doesn't exist.", _NasLogPrefix, deleteKey.AppKey, deleteKey.BuildId, bp.PkgType, filePath)
			continue
		}
		res.NeedDelete = append(res.NeedDelete, deleteKey.BuildId)
		if err = os.RemoveAll(filePath); err != nil {
			res.BuildIdFail = append(res.BuildIdFail, deleteKey.BuildId)
			log.Error("%s appKey[%s], buildID[%d] pkgType[%d], delete file: %s FAILED! err: %+v", _NasLogPrefix, deleteKey.AppKey, deleteKey.BuildId, bp.PkgType, filePath, err)
		} else {
			log.Info("%s appKey[%s], buildID[%d] pkgType[%d], delete file: %s SUCCESS!", _NasLogPrefix, deleteKey.AppKey, deleteKey.BuildId, bp.PkgType, filePath)
			fileDeleted = append(fileDeleted, deleteKey.BuildId)
		}
	}
	if len(fileDeleted) == 0 {
		if len(res.NeedDelete) != 0 {
			log.Error("%s 以下buildId: %v需要删除但是删除失败", _NasLogPrefix, res.NeedDelete)
		}
		return
	}
	if res.AffectedRows, err = s.fkDao.UpdateCIExpiredStatus(c, fileDeleted); err != nil {
		log.Error("%s UpdateCIExpiredStatus File Fail(%v)", _NasLogPrefix, err)
	}
	log.Info("%s statistical result: %+v", _NasLogPrefix, &res)
	return
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return os.IsExist(err)
	}
	return true
}

// BuildPackSubRepoList get ci subrepo list
func (s *Service) BuildPackSubRepoList(c context.Context, appKey, commit string) (res []*cimdl.BuildPackSubRepo, err error) {
	if res, err = s.fkDao.BuildPackSubReposByCommit(c, appKey, commit); err != nil {
		log.Error("%v", err)
	}
	return
}

func (s *Service) GetAppBuildPackVersionInfo(ctx context.Context, req *cimdl.GetAppBuildPackVersionInfoReq) (resp *cimdl.GetAppBuildPackVersionInfoResp, err error) {
	if len(req.GitlabJobId) == 0 {
		err = ecode.Error(ecode.RequestErr, "req.GitlabJobId is empty")
		return
	}
	info, err := s.fkDao.SelectBuildPackVersionInfo(ctx, req.AppKey, req.State, req.GitlabJobId)
	if err != nil {
		err = ecode.Error(ecode.ServerErr, "select build version err")
		log.Errorc(ctx, "GetAppBuildPackVersionInfo %v", err)
		return
	}
	var cdGlJobId []int64
	for _, v := range info {
		if v.DidPush == 1 {
			cdGlJobId = append(cdGlJobId, v.GitlabJobID)
		}
	}
	packs, err := s.fkDao.SelectPackByGlJobId(ctx, req.AppKey, cdGlJobId)
	if err != nil {
		err = ecode.Error(ecode.ServerErr, "数据查询错误")
		log.Errorc(ctx, "GetAppBuildPackVersionInfo %v", err)
		return
	}
	var cdMap = make(map[int64][]*cdmdl.Pack)
	for _, v := range packs {
		cdMap[v.BuildID] = append(cdMap[v.BuildID], v)
	}
	var versions []*cimdl.VersionInfo
	for _, v := range info {
		cdPack := cdMap[v.GitlabJobID]
		var env []string
		var sstate int8
		for _, v := range cdPack {
			env = append(env, v.Env)
			if v.Env == "prod" {
				sstate = v.SteadyState
			}
		}
		versions = append(versions, &cimdl.VersionInfo{
			GitlabJobId: v.GitlabJobID,
			VersionCode: v.VersionCode,
			Version:     v.Version,
			IsPushCD:    v.DidPush,
			Env:         env,
			SteadyState: sstate,
		})
	}
	resp = &cimdl.GetAppBuildPackVersionInfoResp{
		VersionInfo: versions,
	}
	return
}

// saveMaven 保存maven文件 src-main.bbr地址 存储的目标文件夹
func (s *Service) saveMaven(src, dst, AppKey, bundleName string, jobId int64) (err error) {
	var (
		pomContent, bundle, bundleMd5, bundleSha, pomName, pomMd5, pomSha1 string
	)
	var d = template.PomData{
		AppKey:     AppKey,
		BundleName: bundleName,
		CIJobID:    jobId,
	}
	if bundle, err = s.fkDao.TemplateAlter(d, template.BundleName); err != nil {
		log.Error("%v", err)
		return
	}
	if bundleMd5, err = s.fkDao.TemplateAlter(d, template.BundleMd5); err != nil {
		log.Error("%v", err)
		return
	}
	if bundleSha, err = s.fkDao.TemplateAlter(d, template.BundleSha1); err != nil {
		log.Error("%v", err)
		return
	}
	if pomName, err = s.fkDao.TemplateAlter(d, template.BundlePom); err != nil {
		log.Error("%v", err)
		return
	}
	if pomMd5, err = s.fkDao.TemplateAlter(d, template.BundlePomMd5); err != nil {
		log.Error("%v", err)
		return
	}
	if pomSha1, err = s.fkDao.TemplateAlter(d, template.BundlePomSha1); err != nil {
		log.Error("%v", err)
		return
	}
	if pomContent, err = s.fkDao.TemplateAlter(d, template.PomContent); err != nil {
		log.Error("%v", err)
		return
	}
	if err = copyFileContents(src, path.Join(dst, bundle)); err != nil {
		log.Error("%v", err)
		return
	}
	_ = ioutil.WriteFile(path.Join(dst, bundleMd5), []byte(getFileMd5(src)), 0600)
	_ = ioutil.WriteFile(path.Join(dst, bundleSha), []byte(getFileSha1(src)), 0600)

	_ = ioutil.WriteFile(path.Join(dst, pomName), []byte(pomContent), 0600)
	_ = ioutil.WriteFile(path.Join(dst, pomMd5), []byte(getFileMd5(path.Join(dst, pomName))), 0600)
	_ = ioutil.WriteFile(path.Join(dst, pomSha1), []byte(getFileSha1(path.Join(dst, pomName))), 0600)
	return
}

// ParseBBR get api from main.bbr
func (s *Service) ParseBBR(c context.Context, appKey string, buildID int64, feature string, exFilter []string) (content string, err error) {
	var bbr map[string]map[string][]*cimdl.BBRItem
	if bbr, err = s.getAppAPI(c, appKey, buildID, feature); err != nil {
		log.Errorc(c, "%v", err)
		err = ecode.Error(ecode.RequestErr, "获取API信息失败")
		return
	}
	for _, v := range exFilter {
		delete(bbr, v)
	}
	if content, err = s.fkDao.TemplateAlter(bbr, template.APIInfo); err != nil {
		log.Errorc(c, "%v", err)
		err = ecode.Error(ecode.RequestErr, "格式化返回结果失败")
		return
	}
	return
}

func (s *Service) getAppAPI(c context.Context, appKey string, buildId int64, feature string) (bbr map[string]map[string][]*cimdl.BBRItem, err error) {
	pack, err := s.fkDao.BuildPack(c, appKey, buildId)
	if err != nil {
		log.Errorc(c, "%v", err)
		return
	}
	if len(pack.BbrURL) == 0 {
		log.Warnc(c, "buildID %d bbr url is empty", buildId)
		err = ecode.Error(ecode.NothingFound, "找不到资源文件")
		return
	}
	bbrPath := strings.Replace(pack.BbrURL, conf.Conf.LocalPath.LocalDomain, conf.Conf.LocalPath.LocalDir, 1)
	out, err := execTribeAPI(c, bbrPath, feature)
	if err != nil {
		log.Errorc(c, "%v", err)
		return
	}
	if err = json.Unmarshal(out.Bytes(), &bbr); err != nil {
		log.Errorc(c, "%v", err)
		return
	}
	return
}

func getFileSha1(path string) string {
	pFile, err := os.Open(path)
	if err != nil {
		log.Error("打开文件失败，path=%v, err=%v", path, err)
		return ""
	}
	defer pFile.Close()
	sha1h := sha1.New()
	_, _ = io.Copy(sha1h, pFile)
	return hex.EncodeToString(sha1h.Sum(nil))
}

func getFileMd5(path string) string {
	pFile, err := os.Open(path)
	if err != nil {
		log.Error("打开文件失败，path=%v, err=%v", path, err)
		return ""
	}
	defer pFile.Close()
	md5h := md5.New()
	_, _ = io.Copy(md5h, pFile)
	return hex.EncodeToString(md5h.Sum(nil))
}

func copyFileContents(src, dst string) (err error) {
	// 文件存在则直接返回成功
	if _, err = os.Stat(dst); err == nil {
		log.Error("文件%s已存在", dst)
		return
	}
	if err = os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		log.Error("os.MkdirAll error(%v)", err)
		return
	}
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return
}

func execTribeAPI(c context.Context, mainBbrPath string, feature string) (out bytes.Buffer, err error) {
	var (
		errOut bytes.Buffer
	)
	start := time.Now()
	defer func(startT time.Time) {
		log.Infoc(c, "start parse at %v, end at %v", startT, time.Now())
	}(start)
	if len(feature) == 0 {
		feature = "default"
	}
	cmd := exec.Command(conf.Conf.ExePath.Tribe.TribeAPI, "-bbr", mainBbrPath, "-feature", feature) //nolint:gosec
	cmd.Env = os.Environ()
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	log.Infoc(c, "%v args:%v", cmd.Path, cmd.Args)
	if err = cmd.Run(); err != nil {
		log.Errorc(c, "Command Run stdout=(%s) stderr=(%s) error(%v)", out.String(), errOut.String(), err)
		return
	}
	return
}

// getCompatibleVersions 获取兼容信息 如果没有兼容的版本 则兼容id设置为0
func getCompatibleVersions(metaJson map[string]interface{}) (features []*cimdl.Feature) {
	if _, ok := metaJson[cimdl.MetaKeyCompatible]; ok {
		// 老版metaJson
		c := metaJson[cimdl.MetaKeyCompatible].([]interface{})
		var vid int64
		if len(c) != 0 {
			vid = int64(c[len(c)-1].(float64))
			f := cimdl.Feature{
				Name:              "default",
				CompatibleVersion: vid,
			}
			features = append(features, &f)
		}
	}
	if _, ok := metaJson[cimdl.MetaKeyFeatures]; ok {
		// 新版metaJson
		fi := metaJson[cimdl.MetaKeyFeatures].([]interface{})
		for _, v := range fi {
			vm := v.(map[string]interface{})
			if _, ok1 := vm[cimdl.MetaKeyFeatureName]; !ok1 {
				return
			}
			vc := vm[cimdl.MetaKeyCompatible].([]interface{})
			var vid int64
			if len(vc) != 0 {
				vid = int64(vc[len(vc)-1].(float64))
				f := cimdl.Feature{
					Name:              vm[cimdl.MetaKeyFeatureName].(string),
					CompatibleVersion: vid,
				}
				features = append(features, &f)
			}
		}
	}
	return
}

func getMetaJson(mainbbr string) (meta map[string]interface{}, err error) {
	var zf *zip.ReadCloser
	if zf, err = zip.OpenReader(mainbbr); err != nil {
		return
	}
	for _, file := range zf.File {
		if file.Name == "meta.json" {
			var fc io.ReadCloser
			if fc, err = file.Open(); err != nil {
				return
			}
			var content []byte
			if content, err = ioutil.ReadAll(fc); err != nil {
				return
			}
			if err = json.Unmarshal(content, &meta); err != nil {
				return
			}
			break
		}
	}
	return
}

func (s *Service) jobStatusChangeAction(arg *gitmdl.GitJobStatusChangeInfo) {
	log.Info("jobStatusChangeAction %v", arg)
	if arg.BusinessType != gitmdl.CI {
		return
	}
	id := arg.Id             // cimdl.BuildPack.id
	gitJobId := arg.GitJobId // cimdl.BuildPack.gl_job_id
	switch arg.CurrentStatus {
	case gitmdl.Running:
		// 开始执行
	case gitmdl.Canceled:
		// 取消
		err := s.CanceledAction(context.Background(), id, gitJobId)
		if err != nil {
			log.Error("jobStatusChangeAction error %v", err)
			return
		}
	case gitmdl.Failed:
		// 失败
		err := s.FailedAction(context.Background(), id, gitJobId)
		if err != nil {
			log.Error("jobStatusChangeAction error %v", err)
			return
		}
	default:

	}
}

func (s *Service) CanceledAction(c context.Context, buildID, _ int64) (err error) {
	err = s.CiStatusWechatNotify(c, buildID)
	return
}

func (s *Service) FailedAction(c context.Context, buildID, _ int64) (err error) {
	err = s.CiStatusWechatNotify(c, buildID)
	return
}

func (s *Service) CiStatusWechatNotify(c context.Context, buildID int64) (err error) {
	var (
		ci *cimdl.BuildPack
	)
	if ci, err = s.fkDao.BuildPackById(c, buildID); err != nil {
		return
	}
	notify := &cimdl.NotifyCI{
		BuildId:      buildID,
		IsNotifyUser: true,
		Receiver:     &cimdl.NotifyCIReceiver{Users: ci.Operator},
	}
	if err = s.NotifyCIJob(c, notify); err != nil {
		return
	}
	return
}

// NotifyGroup ci notify, including mail,bot.
func (s *Service) NotifyGroup(c context.Context, buildId int64, notifyCCGroup bool, bots string) (err error) {
	var ci *cimdl.BuildPack
	if ci, err = s.fkDao.BuildPackById(c, buildId); err != nil {
		log.Errorc(c, "BuildPackById error %v", err)
		return
	}
	if ci == nil {
		log.Errorc(c, "build pack is nil")
		return
	}
	receiver := &cimdl.NotifyCIReceiver{
		Bots:  bots,
		Users: ci.Operator,
	}
	var mailReceiveType cimdl.NotifyMail
	if notifyCCGroup {
		mailReceiveType = cimdl.NotifyMailCC
	} else {
		mailReceiveType = cimdl.NotifyMailRecipient
	}
	notify := &cimdl.NotifyCI{
		BuildId:      buildId,
		Receiver:     receiver,
		NotifyMail:   mailReceiveType,
		IsNotifyBot:  true,
		IsNotifyUser: true,
	}
	if err = s.NotifyCIJob(c, notify); err != nil {
		log.Errorc(c, "NotifyCIJobFinished error(%v)", err)
	}
	return
}

func (s *Service) CITrack(c context.Context, msg *cimdl.TrackMessage) (err error) {
	extendedFieldMap := make(map[string]interface{})
	err = json.Unmarshal([]byte(msg.ExtendFields), &extendedFieldMap)
	if err != nil {
		log.Errorc(c, "json.Unmarshal error %v", err)
		return
	}
	extendedFieldByte, err := json.Marshal(extendedFieldMap)
	if err != nil {
		log.Errorc(c, "json.Marshal %+v  err %v", extendedFieldMap, err)
		return
	}
	payload := infoc.NewLogStreamVWithSeparator(cimdl.TrackLogID, cimdl.TrackSeparator,
		xlog.Int64(msg.BuildId),
		xlog.String(msg.HostName),
		xlog.String(msg.Arch),
		xlog.String(msg.Platform),
		xlog.String(msg.OSName),
		xlog.String(msg.OSVersion),
		xlog.String(msg.Operator),
		xlog.String(msg.HardwareInfo),
		xlog.String(string(extendedFieldByte)),
		xlog.String(msg.Operator),
	)
	if err = s.infoc.Info(context.Background(), payload); err != nil {
		log.Errorc(c, "infocLog.Info error %v", err)
	}
	return
}
