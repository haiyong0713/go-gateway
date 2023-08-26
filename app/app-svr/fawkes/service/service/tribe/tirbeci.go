package tribe

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"go-common/library/database/sql"
	"go-common/library/ecode"

	"github.com/golang/protobuf/ptypes/empty"

	"go-gateway/app/app-svr/fawkes/service/conf"
	cimdl "go-gateway/app/app-svr/fawkes/service/model/ci"
	"go-gateway/app/app-svr/fawkes/service/model/template"

	toolmdl "go-gateway/app/app-svr/fawkes/service/model/tool"

	"go-gateway/app/app-svr/fawkes/service/api/app/tribe"
	appmdl "go-gateway/app/app-svr/fawkes/service/model/app"
	mngmdl "go-gateway/app/app-svr/fawkes/service/model/manager"
	tribemdl "go-gateway/app/app-svr/fawkes/service/model/tribe"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
	"go-gateway/app/app-svr/fawkes/service/tools/utils"
)

func (s *Service) AddTribeBuildPack(ctx context.Context, req *tribe.AddTribeBuildPackReq) (resp *empty.Empty, err error) {
	var (
		tribeInfo                   *tribemdl.Tribe
		appId, gitPath, gitlabPrjId string
		envVarMap                   = make(map[string]string)
		buildPackId                 int64
	)
	op := utils.GetUsername(ctx)
	if tribeInfo, err = s.fkDao.SelectTribeById(ctx, req.TribeId); err != nil {
		log.Errorc(ctx, "SelectTribeById tribe_id[%d] error: [%v]", req.TribeId, err)
		err = ecode.Error(ecode.NothingFound, fmt.Sprintf("组件-id[%d] 不存在", req.TribeId))
		return
	}
	if appId, gitPath, gitlabPrjId, err = s.fkDao.AppBasicInfo(ctx, tribeInfo.AppKey); err != nil {
		log.Errorc(ctx, "%v", err)
		return
	}
	if req.CiEnvVar != "" {
		if err = json.Unmarshal([]byte(req.CiEnvVar), &envVarMap); err != nil {
			log.Errorc(ctx, "AddTribeBuildPack, error: [%v]", err)
			err = ecode.Error(ecode.RequestErr, fmt.Sprintf("json.Unmarshal(%s) error(%v)", req.CiEnvVar, err))
			return
		}
	}
	if !tribeInfo.NoHost && req.DepGlJobId == 0 {
		err = ecode.Error(ecode.RequestErr, "缺少宿主")
		return
	}
	if buildPackId, err = s.fkDao.AddTribeBuildPack(ctx, tribeInfo.Id, req.DepGlJobId, req.PkgType, req.GitType, tribeInfo.AppKey, req.GitName, op, req.CiEnvVar, req.Description, req.ShouldNotify, appId, gitPath, gitlabPrjId, tribemdl.CiInWaiting); err != nil {
		log.Errorc(ctx, "AddTribeBuildPack, error: [%v]", err)
		err = ecode.Error(ecode.RequestErr, "tribe ci 创建失败")
		return
	}
	log.Infoc(ctx, "组件【%s】!tribeInfo.NoHost【%v】宿主构建号【jobId：%d】", tribeInfo.Name, !tribeInfo.NoHost, req.DepGlJobId)
	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Errorc(ctx, "recover panic, TriggerPipeline Error: %v", err)
			}
		}()
		s.TriggerPipeline(context.Background(), tribeInfo.AppKey, req.PkgType, req.DepGlJobId, buildPackId, envVarMap, req.GitType, req.GitName, tribeInfo.Name)
	}()
	return
}

func (s *Service) TriggerPipeline(ctx context.Context, appKey string, pkgType int64, depGitlabJobId int64, buildId int64, envVarMap map[string]string, gitType int64, gitName, tribeName string) {
	var variables = map[string]string{
		"APP_KEY":               appKey,
		"PKG_TYPE":              strconv.FormatInt(pkgType, 10),
		"FAWKES":                "1",
		"FAWKES_USER":           "fawkes",
		"TASK":                  tribemdl.PackBundleTask,
		"BUILD_ID":              strconv.FormatInt(buildId, 10),
		"BIZ_APK_NAME":          tribeName,
		"TRIBE_HOST_BBR_JOB_ID": strconv.FormatInt(depGitlabJobId, 10),
		"TRIBE_PRE_BBR_JOB_ID":  "",
	}
	for key, value := range envVarMap {
		variables[key] = value
	}
	log.Warnc(ctx, "trigger pipeline appKey[%s] gitType[%d] gitName[%s] variables[%v]", appKey, gitType, gitName, variables)
	var err error
	if _, err = s.gitSvr.TriggerPipeline(ctx, appKey, int(gitType), gitName, variables); err != nil {
		log.Errorc(ctx, "trigger pipeline error: %v", err)
		if _, err = s.fkDao.UpdateTribeStatus(ctx, buildId, tribemdl.CiFailed, err.Error()); err != nil {
			log.Errorc(ctx, "UpdateTribeStatus error: %v", err)
		}
	} else {
		_, _ = s.fkDao.UpdateTribeStatus(ctx, buildId, tribemdl.CiInWaiting, "")
	}
}

func (s *Service) ListTribeBuildPack(ctx context.Context, req *tribe.ListTribeBuildPackReq) (resp *tribe.ListTribeBuildPackResp, err error) {
	var tribeBuildPacks []*tribe.TribeBuildPackInfo
	appInfo, err := s.fkDao.AppPass(ctx, req.AppKey)
	if err != nil {
		err = ecode.Error(ecode.NothingFound, fmt.Sprintf("app [%s] not found", req.AppKey))
		log.Errorc(ctx, err.Error())
		return
	}
	total, err := s.fkDao.CountTribeBuildPack(ctx, req.AppKey, req.TribeId, req.GlJobId, req.DepGlJobId, req.PkgType, req.Status, req.State, req.GitName, req.Commit, req.Operator, int32(req.GitType), req.PushCd)
	if err != nil {
		log.Errorc(ctx, "%v", err)
		return
	}
	rows, err := s.fkDao.SelectTribeBuildPackByArg(ctx, req.AppKey, req.TribeId, req.GlJobId, req.DepGlJobId, req.PkgType, req.Status, req.State, int32(req.GitType), req.GitName, req.Commit, req.Operator, req.PushCd, req.OrderBy.String(), req.Sort.String(), req.Ps, req.Pn)
	if err != nil {
		log.Errorc(ctx, "%v", err)
		return
	}
	for _, v := range rows {
		dto := tribeBuildPackPO2DTO(v)
		dto.JobUrl = MakeGitPath(appInfo.GitPath, dto.GlJobId)
		tribeBuildPacks = append(tribeBuildPacks, dto)
	}
	resp = &tribe.ListTribeBuildPackResp{
		PageInfo:            &tribe.PageInfo{Total: int64(total), Pn: req.Pn, Ps: req.Ps},
		TribeBuildPackInfos: tribeBuildPacks,
	}
	return
}

func (s *Service) GetTribeBuildPackInfo(ctx context.Context, req *tribe.GetTribeBuildPackInfoReq) (resp *tribe.GetTribeBuildPackInfoResp, err error) {
	row, err := s.fkDao.SelectTribeBuildPackById(ctx, req.Id)
	if err != nil {
		log.Errorc(ctx, "SelectTribeBuildPacksByIds error: %v", err)
		return
	}
	resp = &tribe.GetTribeBuildPackInfoResp{
		TribeBuildPackInfo: tribeBuildPackPO2DTO(row),
	}
	return
}

// PushTribeBuildPackToCD ci 推送到 cd
func (s *Service) PushTribeBuildPackToCD(ctx context.Context, req *tribe.PushTribeBuildPackToCDReq) (resp *empty.Empty, err error) {
	var (
		app             *appmdl.APP
		packVersionId   int64
		buildPack       *tribemdl.BuildPack
		depAppBuildPack *cimdl.BuildPack
		cdnUrl          string
		tribeInfo       *tribemdl.Tribe
	)
	resp = new(empty.Empty)
	if buildPack, err = s.fkDao.SelectTribeBuildPackById(ctx, req.TribeBuildPackId); err != nil {
		log.Errorc(ctx, "%v", err)
		return
	}
	if buildPack == nil {
		err = ecode.Error(ecode.NothingFound, fmt.Sprintf("tribe_build_id[%d] not found", req.TribeBuildPackId))
		log.Errorc(ctx, err.Error())
		return
	}
	if depAppBuildPack, err = s.fkDao.BuildPackByJobId(ctx, buildPack.AppKey, buildPack.DepGlJobId); err != nil {
		err = ecode.Error(ecode.NothingFound, fmt.Sprintf("dep_git_job_id[%d] not found", buildPack.DepGlJobId))
		log.Errorc(ctx, err.Error())
		return
	}
	if depAppBuildPack.DidPush != cimdl.DidPush {
		err = ecode.Error(ecode.RequestErr, fmt.Sprintf("组件推送CD前，须先将宿主包[%d]推至CD", depAppBuildPack.GitlabJobID))
		log.Errorc(ctx, err.Error())
		return
	}
	if app, err = s.fkDao.AppPass(ctx, buildPack.AppKey); err != nil {
		err = ecode.Error(ecode.NothingFound, fmt.Sprintf("app_Key[%s] not found", buildPack.AppKey))
		log.Errorc(ctx, err.Error())
		return
	}
	if tribeInfo, err = s.fkDao.SelectTribeById(ctx, buildPack.TribeId); err != nil {
		err = ecode.Error(ecode.NothingFound, fmt.Sprintf("tribe_id[%d] not found", buildPack.TribeId))
		log.Errorc(ctx, err.Error())
		return
	}
	if cdnUrl, _, _, err = s.fkDao.FileUploadOss(ctx, buildPack.PkgPath, path.Join("tribe", buildPack.AppKey, tribeInfo.Name, strconv.FormatInt(buildPack.GlJobId, 10), "main.apk"), buildPack.AppKey); err != nil {
		err = ecode.Error(ecode.ServerErr, "文件上传oss失败")
		log.Errorc(ctx, "%v", err)
		return
	}
	if err = s.fkDao.Transact(ctx, func(tx *sql.Tx) error {
		var tribePackVersion *tribemdl.PackVersion
		// select version id if not exist insert
		vc, _ := strconv.ParseInt(buildPack.VersionCode, 10, 64)
		if tribePackVersion, err = s.fkDao.TxSelectTribePackVersionForUpdate(tx, buildPack.TribeId, tribemdl.TestEnv, vc); err != nil {
			log.Errorc(ctx, "error: %v", err)
			return err
		}
		if tribePackVersion == nil || tribePackVersion.Id == 0 {
			if packVersionId, err = s.fkDao.TxSetTribePackVersion(tx, buildPack.TribeId, tribemdl.TestEnv, vc, buildPack.VersionName, false); err != nil {
				log.Errorc(ctx, "error: %v", err)
				return err
			}
		} else {
			packVersionId = tribePackVersion.Id
		}
		// 同步pack表
		if _, err = s.fkDao.TxAddTribePackFromBuild(tx, buildPack, packVersionId, tribemdl.TestEnv, cdnUrl, req.Description, utils.GetUsername(ctx)); err != nil {
			log.Errorc(ctx, "error: %v", err)
			return err
		}
		if _, err = s.fkDao.TxUpdateTribeBuildPackDidPush(tx, req.TribeBuildPackId, true); err != nil {
			log.Errorc(ctx, "error: %v", err)
			return err
		}
		return err
	}); err != nil {
		log.Errorc(ctx, "error: %v", err)
		return
	}
	_, _ = s.fkDao.AddLog(ctx, app.AppKey, tribemdl.TestEnv, mngmdl.ModelCI, mngmdl.OperationCIPush, fmt.Sprintf("构建ID: %v", buildPack.GlJobId), utils.GetUsername(ctx))
	return
}

// UpdateTribeBuildPackGitInfo git pipeline开始后，更新git相关信息
func (s *Service) UpdateTribeBuildPackGitInfo(ctx context.Context, req *tribe.UpdateTribeBuildPackGitInfoReq) (resp *empty.Empty, err error) {
	resp = new(empty.Empty)
	var (
		pack                 *tribemdl.BuildPack
		tribeBuildPackStatus int8
	)
	if pack, err = s.fkDao.SelectTribeBuildPackById(ctx, req.TribeBuildPackId); err != nil || pack == nil {
		err = ecode.Error(ecode.NothingFound, fmt.Sprintf("tribe build pack [id %d] not found", req.TribeBuildPackId))
		log.Errorc(ctx, err.Error())
		return
	}
	if pack.Status == tribemdl.CiCancel {
		err = ecode.Error(ecode.OK, "本次构建已取消，无需更新")
		log.Warnc(ctx, err.Error())
		return
	}
	if req.Status.String() != tribemdl.PipelineSuccess {
		tribeBuildPackStatus = tribemdl.CiFailed
	} else {
		tribeBuildPackStatus = tribemdl.CiBuilding
	}
	if _, err = s.fkDao.UpdateTribeBuildPackGitInfo(ctx, req.TribeBuildPackId, req.GitlabJobId, req.DepGitlabJobId, req.GitPath, req.Commit, req.BuildStartTime, tribeBuildPackStatus); err != nil {
		log.Errorc(ctx, "s.fkDao.UpdateTribeBuildPack error: %v", err)
	}
	return
}

// UpdateTribeBuildPackPkgInfo 文件上传
func (s *Service) UpdateTribeBuildPackPkgInfo(ctx context.Context, req *tribe.UpdateTribeBuildPackPkgInfoReq) (resp *tribe.UpdateTribeBuildPackPkgInfoResp, err error) {
	var (
		t                                                        *tribemdl.Tribe
		buildPack                                                *tribemdl.BuildPack
		tmpPath, destFileDir, apkUrl, bbrUrl, mappingUrl, apkMD5 string
		apkFileStat                                              os.FileInfo
		metaJson                                                 map[string]interface{}
	)
	pushMavenStatus := tribemdl.PushMavenSuccess
	fileValue := ctx.Value(toolmdl.ContentKey).(*toolmdl.ContextValues)
	file := fileValue.File
	fileHeader := fileValue.FileHeader
	if req.Status.String() != tribemdl.PipelineSuccess {
		log.Infoc(ctx, "UpdateTribeBuildPackPkgInfo tribeId[%d] 创建CI失败", req.TribeBuildPackId)
		if _, err := s.fkDao.UpdateTribeStatus(ctx, req.TribeBuildPackId, tribemdl.CiFailed, "构建产物打包失败"); err != nil {
			log.Errorc(ctx, "%v", err)
		}
		return
	}
	if buildPack, err = s.fkDao.SelectTribeBuildPackById(ctx, req.TribeBuildPackId); err != nil {
		log.Errorc(ctx, "error: %v", err)
		return
	}
	if t, err = s.fkDao.SelectTribeById(ctx, buildPack.TribeId); err != nil || t == nil {
		err = ecode.Error(ecode.NothingFound, fmt.Sprintf("tribe id[%d] not found", buildPack.TribeId))
		log.Errorc(ctx, err.Error())
		return
	}
	// prefix_{{app_key}}/{{tribe_name}}/{{job_id}}/
	destFileDir = filepath.Join(conf.Conf.LocalPath.LocalDir, tribemdl.TribePath, buildPack.AppKey, t.Name, strconv.FormatInt(buildPack.GlJobId, 10))
	apkPath := filepath.Join(destFileDir, req.PkgPath)
	bbrPath := filepath.Join(destFileDir, req.BbrPath)
	mappingPath := filepath.Join(destFileDir, req.MappingPath)
	if utils.FileExists(apkPath) || utils.FileExists(bbrPath) || utils.FileExists(mappingPath) {
		err = ecode.Error(ecode.RequestErr, fmt.Sprintf("文件目录【%s】下，已存在以下全部文件或其中之一【%s %s %s]", destFileDir, req.PkgPath, req.BbrPath, req.MappingPath))
		log.Errorc(ctx, err.Error())
		return
	}
	if tmpPath, err = utils.MultipartFileCopy(file, fileHeader, destFileDir); err != nil {
		err = ecode.Error(ecode.RequestErr, err.Error())
		log.Errorc(ctx, err.Error())
		return
	}
	if req.Unzip {
		if err = utils.Unzip(tmpPath, destFileDir); err != nil {
			log.Errorc(ctx, "unzip(%s, %s) error(%v)", tmpPath, destFileDir, err)
			return
		}
		if err = os.Remove(tmpPath); err != nil {
			log.Errorc(ctx, "os.Remove(%s) error(%v)", tmpPath, err)
			return
		}
	}
	apkUrl = conf.Conf.LocalPath.LocalDomain + strings.TrimPrefix(apkPath, conf.Conf.LocalPath.LocalDir)
	bbrUrl = conf.Conf.LocalPath.LocalDomain + strings.TrimPrefix(bbrPath, conf.Conf.LocalPath.LocalDir)
	mappingUrl = conf.Conf.LocalPath.LocalDomain + strings.TrimPrefix(mappingPath, conf.Conf.LocalPath.LocalDir)
	if apkMD5, err = utils.GetFileMD5(apkPath); err != nil {
		log.Errorc(ctx, err.Error())
		return
	}
	if apkFileStat, err = os.Stat(apkPath); err != nil {
		log.Errorc(ctx, err.Error())
		return
	}
	if metaJson, err = getMetaJson(bbrPath); err != nil {
		log.Errorc(ctx, err.Error())
		return
	}
	depFeature := getDepFeature(metaJson)
	if _, err = s.fkDao.UpdateTribeBuildPackPkgInfo(ctx, req.TribeBuildPackId, apkPath, apkUrl, mappingUrl, bbrUrl, apkMD5, req.ChangeLog, depFeature, req.VersionCode, req.VersionName, req.BuildEndTime, apkFileStat.Size(), tribemdl.CiBuildSuccess); err != nil {
		log.Errorc(ctx, "s.fkDao.UpdateTribeBuildPackPkgInfo error: %v", err)
		return
	}
	if err = s.saveMaven(ctx, filepath.Join(destFileDir, req.BbrPath), buildPack.AppKey, t.Name, buildPack.GlJobId); err != nil {
		log.Errorc(ctx, "save maven error: %v", err)
		pushMavenStatus = tribemdl.PushMavenFailed
	}
	if _, err = s.fkDao.UpdateTribeMavenStatus(ctx, req.TribeBuildPackId, int64(pushMavenStatus)); err != nil {
		log.Errorc(ctx, "UpdateTribeMavenStatus error: %v", err)
		return
	}
	resp = &tribe.UpdateTribeBuildPackPkgInfoResp{
		MainApkUrl: apkUrl,
		MainBbrUrl: bbrUrl,
		MappingUrl: mappingUrl,
	}
	return
}

func getDepFeature(metaJson map[string]interface{}) (feature string) {
	var dep interface{}
	var ok bool
	dep, ok = metaJson["dependencies"]
	if !ok {
		return
	}
	if len(dep.([]interface{})) != 0 {
		m := dep.([]interface{})[0].(map[string]interface{})
		if m["name"].(string) == "host" {
			if _, ok1 := m["depFeature"]; ok1 {
				feature = m["depFeature"].(string)
			} else {
				feature = "default"
			}
		}
	}
	return
}

// 保存maven相关的文件
func (s *Service) saveMaven(ctx context.Context, src, AppKey, tribeName string, jobId int64) (err error) {
	var (
		pomContent, bundle, bundleMd5, bundleSha, pomName, pomMd5, pomSha1 string
	)
	var d = template.PomData{
		AppKey:     AppKey,
		BundleName: tribeName,
		CIJobID:    jobId,
	}
	dst := path.Join(conf.Conf.LocalPath.LocalDir, "maven", "v1", AppKey, tribeName, strconv.FormatInt(jobId, 10))
	if bundle, err = s.fkDao.TemplateAlter(d, template.BundleName); err != nil {
		log.Errorc(ctx, "%v", err)
		return
	}
	if bundleMd5, err = s.fkDao.TemplateAlter(d, template.BundleMd5); err != nil {
		log.Errorc(ctx, "%v", err)
		return
	}
	if bundleSha, err = s.fkDao.TemplateAlter(d, template.BundleSha1); err != nil {
		log.Errorc(ctx, "%v", err)
		return
	}
	if pomName, err = s.fkDao.TemplateAlter(d, template.BundlePom); err != nil {
		log.Errorc(ctx, "%v", err)
		return
	}
	if pomMd5, err = s.fkDao.TemplateAlter(d, template.BundlePomMd5); err != nil {
		log.Errorc(ctx, "%v", err)
		return
	}
	if pomSha1, err = s.fkDao.TemplateAlter(d, template.BundlePomSha1); err != nil {
		log.Errorc(ctx, "%v", err)
		return
	}
	if pomContent, err = s.fkDao.TemplateAlter(d, template.PomContent); err != nil {
		log.Errorc(ctx, "%v", err)
		return
	}
	if err = utils.FileCopy(src, path.Join(dst, bundle)); err != nil {
		log.Errorc(ctx, "%v", err)
		return
	}
	bbrMd5Content, _ := utils.GetFileMD5(src)
	bbrSha1Content, _ := utils.GetFileSHA1(src)
	_ = ioutil.WriteFile(path.Join(dst, bundleMd5), []byte(bbrMd5Content), 0600)
	_ = ioutil.WriteFile(path.Join(dst, bundleSha), []byte(bbrSha1Content), 0600)

	pomFilePath := path.Join(dst, pomName)
	pomMd5Content, _ := utils.GetFileMD5(pomFilePath)
	pomSha1Content, _ := utils.GetFileSHA1(pomFilePath)
	_ = ioutil.WriteFile(pomFilePath, []byte(pomContent), 0600)
	_ = ioutil.WriteFile(path.Join(dst, pomMd5), []byte(pomMd5Content), 0600)
	_ = ioutil.WriteFile(path.Join(dst, pomSha1), []byte(pomSha1Content), 0600)
	return
}

// CancelTribeBuildPack 取消CI
func (s *Service) CancelTribeBuildPack(ctx context.Context, req *tribe.CancelTribeBuildPackReq) (resp *empty.Empty, err error) {
	resp = new(empty.Empty)
	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Errorc(ctx, "recover panic, CancelTribeJob Error: %v", err)
			}
		}()
		_ = s.gitSvr.CancelTribeJob(context.Background(), req.TribeBuildPackId)
	}()
	if _, err = s.fkDao.UpdateTribeStatus(ctx, req.TribeBuildPackId, tribemdl.CiCancel, ""); err != nil {
		log.Errorc(ctx, "error: %v", err)
	}
	return
}

func tribeBuildPackPO2DTO(po *tribemdl.BuildPack) (tribeInfo *tribe.TribeBuildPackInfo) {
	if po == nil {
		return
	}
	var versionCodeInt int64
	if po.VersionCode != "" {
		versionCodeInt, _ = strconv.ParseInt(po.VersionCode, 10, 64)
	}
	tribeInfo = &tribe.TribeBuildPackInfo{
		Id:             po.Id,
		TribeId:        po.TribeId,
		GlJobId:        po.GlJobId,
		DepGlJobId:     po.DepGlJobId,
		AppId:          po.AppId,
		AppKey:         po.AppKey,
		GitPath:        po.GitPath,
		GitType:        int32(po.GitType),
		GitName:        po.GitName,
		Commit:         po.Commit,
		PkgType:        int32(po.PkgType),
		Operator:       po.Operator,
		Size_:          po.Size,
		Md5:            po.Md5,
		PkgPath:        po.PkgPath,
		PkgUrl:         po.PkgUrl,
		MappingUrl:     po.MappingUrl,
		BbrUrl:         po.BbrUrl,
		State:          int32(po.State),
		Status:         int32(po.Status),
		DidPush:        int32(po.DidPush),
		ChangeLog:      po.ChangeLog,
		NotifyGroup:    po.NotifyGroup == 1,
		CiEnvVars:      po.CiEnvVars,
		BuildStartTime: po.BuildStartTime.Unix(),
		BuildEndTime:   po.BuildEndTime.Unix(),
		Description:    po.Description,
		Ctime:          po.Ctime.Unix(),
		Mtime:          po.Mtime.Unix(),
		VersionCode:    versionCodeInt,
		VersionName:    po.VersionName,
		DepFeature:     po.DepFeature,
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
