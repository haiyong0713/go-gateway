package cd

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path"
	"sort"
	"strconv"
	"strings"

	"go-common/library/database/sql"

	"go-gateway/app/app-svr/fawkes/service/conf"
	"go-gateway/app/app-svr/fawkes/service/model"
	cdmdl "go-gateway/app/app-svr/fawkes/service/model/cd"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

// PackVers version code sort
type PackVers []*cdmdl.Pack

func (v PackVers) Len() int { return len(v) }
func (v PackVers) Less(i, j int) bool {
	var iv, jv int64
	if v[i] != nil {
		iv = v[i].VersionCode
	}
	if v[j] != nil {
		jv = v[j].VersionCode
	}
	return iv > jv
}
func (v PackVers) Swap(i, j int) { v[i], v[j] = v[j], v[i] }

// PatchList get diff list.
func (s *Service) PatchList(c context.Context, appKey string, buildID int64, pn, ps int) (res *cdmdl.PatchResult, err error) {
	var (
		patchs []*cdmdl.Patch
		total  int
	)
	if total, err = s.fkDao.PatchListCount(c, appKey, buildID); err != nil {
		log.Error("%v", err)
		return
	}
	if total == 0 {
		log.Warn("PatchList total is 0 ")
		return
	}
	if patchs, err = s.fkDao.PatchList(c, appKey, buildID, pn, ps); err != nil {
		log.Error("%v", err)
		return
	}
	info, err := s.fkDao.AppPass(c, appKey)
	if err != nil {
		return
	}
	for _, patch := range patchs {
		patch.GlJobURL = s.MakeGitPath(info.GitPath, patch.GlJobID)
	}

	if len(patchs) < 1 {
		log.Warn("PatchList patchs is 0 ")
		return
	}
	if total < len(patchs) {
		total = len(patchs)
	}
	page := &model.PageInfo{
		Total: total,
		Pn:    pn,
		Ps:    ps,
	}
	res = &cdmdl.PatchResult{
		PageInfo: page,
		Items:    patchs,
	}
	return
}

// PatchGenerate generate patch.
func (s *Service) PatchGenerate(c context.Context, appKey, fromPackPath, toPackPath, packURL string, fromVersionID, toVersionID,
	fromBuildID, toBuildID int64) (err error) {
	folder := path.Join("pack", appKey, strconv.FormatInt(toBuildID, 10), "patch")
	patchName := strconv.FormatInt(fromBuildID, 10) + "-to-" + strconv.FormatInt(toBuildID, 10) + ".patch"
	outPath := path.Join(s.c.LocalPath.LocalDir, folder, patchName)

	// if patch file exist
	patchDir := path.Join(s.c.LocalPath.LocalDir, folder)
	if _, err = os.Stat(patchDir); err != nil {
		if err = os.MkdirAll(patchDir, 0755); err != nil {
			log.Error("os.MkdirAll error:%v", err)
			return
		}
	}
	if _, err = os.Stat(outPath); err == nil {
		log.Error("%v Patch file already exist!", outPath)
		return
	}
	inetPath := s.c.LocalPath.LocalDomain + "/" + path.Join(folder, patchName)

	var (
		out    bytes.Buffer
		errOut bytes.Buffer
	)
	// nolint:gosec
	cmd := exec.Command(s.c.LocalPath.PatcherPath, "diff", "--from", fromPackPath, "--to", toPackPath, "-o", outPath)
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	log.Info("PatchGenerate: %v %v %v %v %v %v %v %v", s.c.LocalPath.PatcherPath, "diff", "--from", fromPackPath, "--to", toPackPath, "-o", outPath)
	if err = cmd.Run(); err != nil {
		log.Error("Command Run stdout=(%s) stderr=(%s) error(%v)", out.String(), errOut.String(), err)
		return
	}
	log.Info("Finished generating patch: %v", outPath)

	var (
		cdnPath, fmd5 string
		size          int64
	)
	if cdnPath, fmd5, size, err = s.fkDao.FilePutOss(c, folder, patchName, appKey); err != nil {
		log.Error("FilePutOss: %v", err)
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
	if _, err = s.fkDao.TxAddPatch(tx, appKey, toBuildID, toBuildID, toVersionID, fromBuildID, fromVersionID, size, 3,
		fmd5, outPath, inetPath, cdnPath, packURL); err != nil {
		log.Error("%v", err)
	}
	return
}

// GenerateAllPatchesTest generate all patches test
func (s *Service) GenerateAllPatchesTest(appKey, toPackPath string, toVersionID, toVersionCode, toBuildID int64) (err error) {
	var packs []*cdmdl.Pack
	// if packs, err = s.fkDao.LastPack(context.Background(), appKey, toVersionCode); err != nil {
	// 	log.Error("%v", err)
	// 	return
	// }
	if packs, err = s.getLastPack(appKey, toVersionCode); err != nil {
		log.Error("%v", err)
		return
	}
	var tx *sql.Tx
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
	//var patchSrcParams []*cdmdl.PatchPipeline
	for _, pack := range packs {
		//var patchID int64
		pTmp := &cdmdl.Pack{}
		*pTmp = *pack
		if _, err = s.fkDao.TxAddPatch(tx, appKey, toBuildID, toBuildID, toVersionID, pTmp.BuildID, pTmp.VersionID, 0, 1,
			"", "", "", "", pTmp.PackURL); err != nil {
			log.Error("%v", err)
		}
		//patchSrc := &cdmdl.PatchPipeline{}
		//patchSrc.ID = patchID
		//patchSrc.PackURL = pTmp.PackURL
		//patchSrcParams = append(patchSrcParams, patchSrc)
	}
	// 分组发送
	// var batchPatchParams []*cdmdl.PatchPipeline
	// for _, patchParam := range patchSrcParams {
	// 	var batchPatchParams []*cdmdl.PatchPipeline
	// 	batchPatchParams = append(batchPatchParams, patchParam)
	// 	if err = s.patchPipelineTriger(appKey, toPackPath, batchPatchParams); err != nil {
	// 		log.Error("patchPipelineTriger %v", err)
	// 		return
	// 	}
	// }
	return
}

//nolint:unused
func (s *Service) patchPipelineTriger(appKey, toPackPath string, params []*cdmdl.PatchPipeline) (err error) {
	srcBuild := &cdmdl.PatchPipelineData{}
	srcBuild.Data = params
	srcBuildJSON, err := json.Marshal(srcBuild)
	if err != nil {
		log.Error("srcBuildJSON Marshal %v", err)
		return
	}
	var variables = map[string]string{
		"APP_KEY":        appKey,
		"TASK":           "PATCH",
		"DST_BUILD_ID":   toPackPath,
		"SRC_BUILD_JSON": string(srcBuildJSON),
	}
	_, err = s.gitSvr.TriggerPipeline(context.Background(), appKey, 0, cdmdl.PatchGitName, variables)
	return
}

// MakeGitPath to make jobid to gitPath
func (s *Service) MakeGitPath(gitPath string, glJobID int64) (jobURL string) {
	var (
		groupName, projectName string
	)
	if len(gitPath) > 0 {
		if strings.HasPrefix(gitPath, "git@") {
			// git@git.bilibili.co:studio/android/bilibiliStudio.git
			pathComps := strings.Split(gitPath, ":")
			projectNameComp := strings.Split(pathComps[len(pathComps)-1], ".git")[0]
			projectNameComps := strings.Split(projectNameComp, "/")
			groupName = projectNameComps[0]
			i := 0
			projectName = strings.Join(append(projectNameComps[:i], projectNameComps[i+1:]...), "/")
		} else {
			pathComps := strings.Split(gitPath, "/")
			projectName = strings.Split(pathComps[len(pathComps)-1], ".git")[0]
			groupName = pathComps[len(pathComps)-2]
		}
		return conf.Conf.Gitlab.Host + "/" + groupName + "/" + projectName + "/-/jobs/" + strconv.FormatInt(glJobID, 10)
	}
	return
}

func (s *Service) getLastPack(appKey string, toVersionCode int64) (lastPacks []*cdmdl.Pack, err error) {
	var (
		packs          []*cdmdl.Pack
		notSteadyPacks []*cdmdl.Pack
	)
	// patch 规则改为 5个稳定版本 + 15个最新版本  共计s.c.PatchLimit=20 个 2021/10/12
	if packs, err = s.fkDao.LastPack(context.Background(), appKey, toVersionCode, 1, s.c.PatchSteadyLimit); err != nil {
		log.Error("%v", err)
		return
	}
	packsCount := len(packs)
	if packsCount < s.c.PatchLimit {
		if notSteadyPacks, err = s.fkDao.LastPack(context.Background(), appKey, toVersionCode, 0, s.c.PatchLimit-packsCount); err != nil {
			log.Error("%v", err)
			return
		}
	}
	lastPacks = append(packs, notSteadyPacks...)
	sort.Sort(PackVers(lastPacks))
	return
}
