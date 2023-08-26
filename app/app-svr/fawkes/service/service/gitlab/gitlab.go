package gitlab

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"

	"go-common/library/database/sql"
	"go-common/library/ecode"

	"go-common/library/sync/errgroup.v2"

	"go-gateway/app/app-svr/fawkes/service/conf"
	"go-gateway/app/app-svr/fawkes/service/model"
	appmdl "go-gateway/app/app-svr/fawkes/service/model/app"
	"go-gateway/app/app-svr/fawkes/service/model/bizapk"
	cimdl "go-gateway/app/app-svr/fawkes/service/model/ci"
	gitmdl "go-gateway/app/app-svr/fawkes/service/model/gitlab"
	sagamdl "go-gateway/app/app-svr/fawkes/service/model/saga"
	"go-gateway/app/app-svr/fawkes/service/model/tribe"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"

	ggl "github.com/xanzy/go-gitlab"
)

// BranchTagList git branch or tag from gitlab
func (s *Service) BranchTagList(c context.Context, appKey string, branchType int, keyword string) (branches []*gitmdl.BranchInfo, err error) {
	if len(keyword) == 0 {
		if branches, err = s.BranchTagAll(c, appKey, branchType); err != nil {
			log.Errorc(c, "%v", err)
		}
		return
	}
	var (
		reqURL, gitlabProjectID string
		req                     *http.Request
	)
	if gitlabProjectID, err = s.fkDao.GitlabProjectID(c, appKey); err != nil {
		log.Error("s.fkDao.GitlabProjectID error(%v)", err)
		return
	}
	params := url.Values{}
	params.Set("private_token", conf.Conf.Gitlab.Token)
	// branch
	if branchType == 0 {
		reqURL = conf.Conf.Gitlab.Host + conf.Conf.Gitlab.API + "/projects/" + url.QueryEscape(gitlabProjectID) + "/repository/branches"
		params.Set("search", keyword)
		if req, err = http.NewRequest(http.MethodGet, reqURL, strings.NewReader(params.Encode())); err != nil {
			log.Error("http.NewRequest(%s) error(%v)", reqURL, err)
			return
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		res := make([]*gitmdl.Branch, 0)
		if err = s.httpClient.Do(c, req, &res); err != nil {
			log.Error("d.client.Do(%s) error(%v)", reqURL, err)
			return
		}
		for _, item := range res {
			branchInfo := &gitmdl.BranchInfo{Name: item.Name, Commit: item.Commit.ID}
			branches = append(branches, branchInfo)
		}
	} else { // tag
		reqURL = conf.Conf.Gitlab.Host + conf.Conf.Gitlab.API + "/projects/" + url.QueryEscape(gitlabProjectID) + "/repository/tags"
		if req, err = http.NewRequest(http.MethodGet, reqURL, strings.NewReader(params.Encode())); err != nil {
			log.Error("http.NewRequest(%s) error(%v)", reqURL, err)
			return
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		res := make([]*gitmdl.Tag, 0)
		if err = s.httpClient.Do(c, req, &res); err != nil {
			log.Error("d.client.Do(%s) error(%v)", reqURL, err)
			return
		}
		for _, item := range res {
			if keyword != "" && !strings.Contains(item.Name, keyword) {
				continue
			}
			branchInfo := &gitmdl.BranchInfo{Name: item.Name, Commit: item.Commit.ID}
			branches = append(branches, branchInfo)
		}
	}
	return
}

func (s *Service) BranchTagAll(c context.Context, appKey string, branchType int) (branches []*gitmdl.BranchInfo, err error) {
	var (
		gitlabProjectID string
	)
	if gitlabProjectID, err = s.fkDao.GitlabProjectID(c, appKey); err != nil {
		log.Error("s.fkDao.GitlabProjectID error(%v)", err)
		return
	}
	// branch
	if branchType == 0 {
		res, err := s.GetAllBranches(gitlabProjectID)
		if err != nil {
			log.Errorc(c, "GetAllBranches error %v", err)
			return nil, err
		}
		for _, item := range res {
			branchInfo := &gitmdl.BranchInfo{Name: item.Name, Commit: item.Commit.ID, CommitInfo: item.Commit}
			branches = append(branches, branchInfo)
		}
	} else { // tag
		res, err := s.GetAllTags(gitlabProjectID)
		if err != nil {
			log.Errorc(c, "GetAllBranches error %v", err)
			return nil, err
		}
		for _, item := range res {
			branchInfo := &gitmdl.BranchInfo{Name: item.Name, Commit: item.Commit.ID, CommitInfo: item.Commit}
			branches = append(branches, branchInfo)
		}
	}
	return
}

func (s *Service) GetAllBranches(gitlabProjectID interface{}) (branch []*ggl.Branch, err error) {
	_, page, err := s.gitlabClient.Branches.ListBranches(gitlabProjectID, &ggl.ListBranchesOptions{PerPage: 1000})
	if err != nil {
		return
	}
	pageSize := page.ItemsPerPage
	tmpResMu := sync.Mutex{}
	var eg errgroup.Group
	for i := 0; i < page.TotalPages; i++ {
		pageIndex := int(i) + 1
		eg.Go(func(ctx context.Context) (err error) {
			listBranches, _, err := s.gitlabClient.Branches.ListBranches(gitlabProjectID, &ggl.ListBranchesOptions{
				Page:    pageIndex,
				PerPage: pageSize,
			})
			if err != nil {
				return
			}
			tmpResMu.Lock()
			defer tmpResMu.Unlock()
			branch = append(branch, listBranches...)
			return
		})
	}
	err = eg.Wait()
	if err != nil {
		return
	}
	return
}

func (s *Service) GetAllTags(gitlabProjectID interface{}) (tag []*ggl.Tag, err error) {
	_, page, err := s.gitlabClient.Tags.ListTags(gitlabProjectID, &ggl.ListTagsOptions{
		ListOptions: ggl.ListOptions{PerPage: 1000},
	})
	if err != nil {
		return
	}
	pageSize := page.ItemsPerPage
	tmpResMu := sync.Mutex{}
	var eg errgroup.Group
	for i := 0; i < page.TotalPages; i++ {
		pageIndex := int(i) + 1
		eg.Go(func(ctx context.Context) (err error) {
			listTags, _, err := s.gitlabClient.Tags.ListTags(gitlabProjectID, &ggl.ListTagsOptions{
				ListOptions: ggl.ListOptions{Page: pageIndex, PerPage: pageSize},
			})
			if err != nil {
				return
			}
			tmpResMu.Lock()
			defer tmpResMu.Unlock()
			tag = append(tag, listTags...)
			return
		})
	}
	err = eg.Wait()
	if err != nil {
		return
	}
	return
}

// TriggerPipeline gitlab trigger gitType = 2时 gitName需要给commit号
func (s *Service) TriggerPipeline(c context.Context, appKey string, gitType int, gitName string, variables map[string]string) (pipeline *ggl.Pipeline, err error) {
	var (
		gitlabProjectID, triggerToken, branchName string
	)
	if gitlabProjectID, err = s.fkDao.GitlabProjectID(context.Background(), appKey); err != nil {
		log.Errorc(c, "s.fkDao.GitlabProjectID error(%v)", err)
		return
	}
	if triggerToken, err = s.getTriggerToken(c, gitlabProjectID); err != nil {
		return
	}
	// tag & branch
	if gitType != cimdl.GitTypeCommit {
		branchName = gitName
	} else { // commit
		// create a branch with a commit
		branchName = "b-" + time.Now().Format("20060102150405")
		var copt = &ggl.CreateBranchOptions{
			Branch: &branchName,
			Ref:    &gitName,
		}
		if _, _, err = s.gitlabClient.Branches.CreateBranch(gitlabProjectID, copt); err != nil {
			log.Errorc(c, "CreateBranch(%v, %v, %v) error(%v)", gitlabProjectID, branchName, gitName, err)
			return
		}
	}
	var (
		ropt = &ggl.RunPipelineTriggerOptions{
			Ref:       ggl.String(branchName),
			Token:     &triggerToken,
			Variables: variables,
		}
	)
	var resp *ggl.Response
	if pipeline, resp, err = s.gitlabClient.PipelineTriggers.RunPipelineTrigger(gitlabProjectID, ropt); err != nil {
		log.Errorc(c, "RunPipelineTrigger(%v, %v) error(%v) resp:%v", gitlabProjectID, branchName, err, resp)
	}
	log.Warnc(c, "RunPipelineTrigger trigger pipeline appKey[%s] gitType[%d] gitName[%s] variables[%v] resp[%v] pipeline[%v]", appKey, gitType, gitName, variables, resp, pipeline)
	return
}

func (s *Service) getTriggerToken(c context.Context, gitlabProjectID string) (token string, err error) {
	var (
		triggers []*ggl.PipelineTrigger
		lopt     = &ggl.ListPipelineTriggersOptions{}
	)
	if token, err = s.fkDao.GetFawkesToken(c, gitlabProjectID); err != nil {
		return
	}
	// 缓存空
	if token == "" {
		if triggers, _, err = s.gitlabClient.PipelineTriggers.ListPipelineTriggers(gitlabProjectID, lopt); err != nil {
			log.Errorc(c, "ListPipelineTriggers (projectID: %v) %v", gitlabProjectID, err)
			return
		}
		for _, trigger := range triggers {
			lowerDes := strings.ToLower(trigger.Description)
			if strings.Contains(lowerDes, "fawkes") {
				token = trigger.Token
				break
			}
		}
	}
	// 源数据空
	if token == "" {
		err = ecode.Error(ecode.NothingFound, "Can't find fawkes trigger token")
		log.Errorc(c, "Can't find fawkes trigger token. ")
		return
	}
	log.Infoc(c, "update %s token %s", gitlabProjectID, token)
	if err = s.fkDao.SetFawkesToken(c, gitlabProjectID, token); err != nil {
		return
	}
	return
}

// CancelJob cancel a gitlab job
func (s *Service) CancelJob(c context.Context, buildID int64) (err error) {
	var jobInfo *gitmdl.BuildPackJobInfo
	if jobInfo, err = s.fkDao.GitlabJobInfo(c, buildID); err != nil {
		log.Error("GitlabJobInfo(%v) error(%v)", buildID, err)
		return
	}
	if jobInfo == nil {
		return
	}
	if _, _, err = s.gitlabClient.Jobs.CancelJob(jobInfo.GitlabProjectID, jobInfo.GitlabJobID); err != nil {
		log.Error("CancelJob(%v, %v) error(%v)", jobInfo.GitlabProjectID, jobInfo.GitlabJobID, err)
	}
	return
}

// CancelHotfixJob cancel a hotfix job
func (s *Service) CancelHotfixJob(c context.Context, buildID int64) (err error) {
	var jobInfo *gitmdl.BuildPackJobInfo
	if jobInfo, err = s.fkDao.HotfixJobInfo(c, buildID); err != nil {
		log.Error("HotfixJobInfo(%v) error(%v)", buildID, err)
		return
	}
	if jobInfo == nil {
		return
	}
	if _, _, err = s.gitlabClient.Jobs.CancelJob(jobInfo.GitlabProjectID, jobInfo.GitlabJobID); err != nil {
		log.Error("CancelJob(%v, %v) error(%v)", jobInfo.GitlabProjectID, jobInfo.GitlabJobID, err)
	}
	return
}

// CancelBizApkJob cancel a bizapk job
func (s *Service) CancelBizApkJob(c context.Context, buildID int64) (err error) {
	var jobInfo *gitmdl.BuildPackJobInfo
	if jobInfo, err = s.fkDao.BizApkJobInfo(c, buildID); err != nil {
		log.Error("CancelBizApkJob(%v) error(%v)", buildID, err)
		return
	}
	if jobInfo == nil {
		return
	}
	if jobInfo.GitlabJobID != 0 {
		if _, _, err = s.gitlabClient.Jobs.CancelJob(jobInfo.GitlabProjectID, jobInfo.GitlabJobID); err != nil {
			log.Error("CancelJob(%v, %v) error(%v)", jobInfo.GitlabProjectID, jobInfo.GitlabJobID, err)
		}
	} else {
		if _, _, err = s.gitlabClient.Pipelines.CancelPipelineBuild(jobInfo.GitlabProjectID, jobInfo.GitlabPipelineID); err != nil {
			log.Error("CancelPipelineBuild(%v, %v) error(%v)", jobInfo.GitlabProjectID, jobInfo.GitlabPipelineID, err)
		}
	}
	return
}

// RefreshStatusProc refresh job status
func (s *Service) RefreshStatusProc() (err error) {
	var (
		builds []*cimdl.BuildPack
		job    *ggl.Job
	)
	if builds, err = s.fkDao.BuildPacksShouldRefresh(context.Background()); err != nil {
		log.Error("%v", err)
		return
	}
	for _, item := range builds {
		if job, _, err = s.gitlabClient.Jobs.GetJob(item.GitlabProjectID, int(item.GitlabJobID)); err != nil {
			continue
		}
		switch job.Status {
		case "running":
			if item.Status != cimdl.CIBuilding {
				_ = s.UpdateBuildPackStatus(context.Background(), item.BuildID, 2)
			}
		case "success":
			_ = s.UpdateBuildPackStatus(context.Background(), item.BuildID, 3)
		case "failed":
			_ = s.UpdateBuildPackStatus(context.Background(), item.BuildID, -1)
		case "canceled":
			_ = s.UpdateBuildPackStatus(context.Background(), item.BuildID, -2)
		}
		if item.Status != int8(gitmdl.GitJobStatus(job.Status).Val()) {
			s.event.Publish(GitJobStatusChangeEvent, &gitmdl.GitJobStatusChangeInfo{
				OriginStatus:  gitmdl.Convert2PipelineStatus(int(item.Status)),
				CurrentStatus: gitmdl.GitJobStatus(job.Status),
				BusinessType:  gitmdl.CI,
				GitJobId:      int64(job.ID),
				Id:            item.BuildID,
			})
			log.Info("RefreshStatusProc status change %v", gitmdl.GitJobStatusChangeInfo{
				OriginStatus:  gitmdl.Convert2PipelineStatus(int(item.Status)),
				CurrentStatus: gitmdl.GitJobStatus(job.Status),
				BusinessType:  gitmdl.CI,
				GitJobId:      int64(job.ID),
				Id:            item.BuildID,
			})
		}
	}
	return
}

// RefreshHotfixStatusProc refresh hotfix job status
func (s *Service) RefreshHotfixStatusProc() (err error) {
	var (
		jobs []*appmdl.HotfixJobInfo
		job  *ggl.Job
	)
	if jobs, err = s.fkDao.GetHfJobRefresh(context.Background()); err != nil {
		log.Error("%v", err)
		return
	}
	for _, item := range jobs {
		if job, _, err = s.gitlabClient.Jobs.GetJob(item.GitlabProjectID, int(item.GitlabJobID)); err != nil {
			continue
		}
		switch job.Status {
		case "running":
			if item.Status != cimdl.CIBuilding {
				_ = s.updateHotfixStatus(context.Background(), item.BuildPatchID, 2)
			}
		case "success":
			_ = s.updateHotfixStatus(context.Background(), item.BuildPatchID, 3)
		case "failed":
			_ = s.updateHotfixStatus(context.Background(), item.BuildPatchID, -1)
		case "canceled":
			_ = s.updateHotfixStatus(context.Background(), item.BuildPatchID, -2)
		}
		if item.Status != (gitmdl.GitJobStatus(job.Status).Val()) {
			s.event.Publish(GitJobStatusChangeEvent, &gitmdl.GitJobStatusChangeInfo{
				OriginStatus:  gitmdl.Convert2PipelineStatus(int(item.Status)),
				CurrentStatus: gitmdl.GitJobStatus(job.Status),
				BusinessType:  gitmdl.Hotfix,
				GitJobId:      int64(job.ID),
				Id:            item.BuildPatchID,
			})
		}
	}
	return
}

// RefreshBizApkStatusProc refresh business apk job status
func (s *Service) RefreshBizApkStatusProc() (err error) {
	var (
		jobs      []*bizapk.JobInfo
		job       *ggl.Job
		pipeline  *ggl.Pipeline
		jobStatus string
	)
	if jobs, err = s.fkDao.GetBizApkJobRefresh(context.Background()); err != nil {
		log.Error("%v", err)
		return
	}
	for _, item := range jobs {
		if item.GitlabJobID != 0 {
			if job, _, err = s.gitlabClient.Jobs.GetJob(item.GitlabProjectID, int(item.GitlabJobID)); err != nil {
				continue
			}
			jobStatus = job.Status
		} else if item.GitlabPipelineID != 0 {
			if pipeline, _, err = s.gitlabClient.Pipelines.GetPipeline(item.GitlabProjectID, int(item.GitlabPipelineID)); err != nil {
				continue
			}
			jobStatus = pipeline.Status
		} else {
			continue
		}
		switch jobStatus {
		case "running":
			if item.Status != cimdl.CIBuilding {
				_ = s.updateBizApkStatus(context.Background(), item.ID, 2)
			}
		case "success":
			_ = s.updateBizApkStatus(context.Background(), item.ID, 3)
		case "failed":
			_ = s.updateBizApkStatus(context.Background(), item.ID, -1)
		case "canceled":
			_ = s.updateBizApkStatus(context.Background(), item.ID, -2)
		}
		if item.Status != (gitmdl.GitJobStatus(job.Status).Val()) {
			s.event.Publish(GitJobStatusChangeEvent, &gitmdl.GitJobStatusChangeInfo{
				OriginStatus:  gitmdl.Convert2PipelineStatus(item.Status),
				CurrentStatus: gitmdl.GitJobStatus(job.Status),
				BusinessType:  gitmdl.Biz,
				GitJobId:      int64(job.ID),
				Id:            item.ID,
			})
		}
	}
	return
}

// RefreshTribeStatusProc refresh tribe2.0 job status
func (s *Service) RefreshTribeStatusProc() (err error) {
	var (
		jobs []*tribe.JobInfo
		job  *ggl.Job
	)
	if jobs, err = s.fkDao.GetTribeJobRefresh(context.Background()); err != nil {
		log.Error("%v", err)
		return
	}
	for _, item := range jobs {
		if job, _, err = s.gitlabClient.Jobs.GetJob(item.GitlabProjectID, int(item.GitlabJobID)); err != nil {
			continue
		}
		switch job.Status {
		case "running":
			if item.Status != cimdl.CIBuilding {
				_, _ = s.fkDao.UpdateTribeStatus(context.Background(), item.ID, tribe.CiBuilding, "")
			}
		case "success":
			_, _ = s.fkDao.UpdateTribeStatus(context.Background(), item.ID, tribe.CiBuildSuccess, "")
		case "failed":
			_, _ = s.fkDao.UpdateTribeStatus(context.Background(), item.ID, tribe.CiFailed, "")
		case "canceled":
			_, _ = s.fkDao.UpdateTribeStatus(context.Background(), item.ID, tribe.CiCancel, "")
		}
		if item.Status != (gitmdl.GitJobStatus(job.Status).Val()) {
			s.event.Publish(GitJobStatusChangeEvent, &gitmdl.GitJobStatusChangeInfo{
				OriginStatus:  gitmdl.Convert2PipelineStatus(item.Status),
				CurrentStatus: gitmdl.GitJobStatus(job.Status),
				BusinessType:  gitmdl.Tribe,
				GitJobId:      int64(job.ID),
				Id:            item.ID,
			})
		}
	}
	return
}

func (s *Service) UpdateBuildPackStatus(c context.Context, buildID int64, status int) (err error) {
	var tx *sql.Tx
	if tx, err = s.fkDao.BeginTran(c); err != nil {
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
	if _, err = s.fkDao.TxUpdateBuildPackStatus(tx, buildID, status); err != nil {
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

func (s *Service) updateHotfixStatus(c context.Context, buildPatchID int64, status int) (err error) {
	var tx *sql.Tx
	if tx, err = s.fkDao.BeginTran(c); err != nil {
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
	if _, err = s.fkDao.TxHotfixUpdateStatus(tx, buildPatchID, status); err != nil {
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

func (s *Service) updateBizApkStatus(c context.Context, bizapkBuildID int64, status int) (err error) {
	var tx *sql.Tx
	if tx, err = s.fkDao.BeginTran(c); err != nil {
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
	if err = s.fkDao.TxUpdateBizApkBuildStatus(tx, status, bizapkBuildID); err != nil {
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

// SubRepoMRHook handle the merge request of sub repos
func (s *Service) SubRepoMRHook(c context.Context, appKey string, hookMR *sagamdl.HookMR) (err error) {
	var (
		mainPrjID, newTitle string
		mainMRIID           int
		strArr              []string
		mainMR              *ggl.MergeRequest
	)
	if hookMR.ObjectAttributes == nil {
		marshal, _ := json.Marshal(hookMR)
		log.Errorc(c, "hookMR attributes is nil ObjectAttributes:%s", string(marshal))
		return
	}
	// 子仓 merge 后不做任何操作
	if hookMR.ObjectAttributes.State == "merged" {
		return
	}
	// 获取主仓 MR
	strArr = strings.Split(hookMR.ObjectAttributes.Description, "related main repo:")
	//nolint:gomnd
	if len(strArr) < 2 {
		return
	}
	tmpDesc := strings.TrimSpace(strArr[1])
	strArr = strings.Split(tmpDesc, "\n")
	for _, line := range strArr {
		line = strings.TrimSpace(line)
		line = strings.Replace(line, "https://", "http://", 1)
		if len(line) > 0 && strings.Contains(line, "merge_requests") {
			mainPrjID = strings.TrimPrefix(line, s.c.Gitlab.Host)
			break
		}
	}
	if len(mainPrjID) < 1 {
		err = errors.New("invalid subrepo merge request description")
		log.Error("invalid subrepo merge request description: %v", hookMR.ObjectAttributes.Description)
		return
	}
	strArr = strings.Split(mainPrjID, "merge_requests")
	//nolint:gomnd
	if len(strArr) < 2 {
		err = errors.New("invalid subrepo merge request description")
		log.Error("invalid subrepo merge request description: %v", hookMR.ObjectAttributes.Description)
		return
	}
	mainPrjID = strArr[0]
	if strings.Contains(mainPrjID, "-/") {
		mainPrjID = strings.TrimSuffix(mainPrjID, "-/")
	}
	mainPrjID = strings.TrimLeft(strings.TrimRight(mainPrjID, "/"), "/")
	if mainMRIID, err = strconv.Atoi(strings.TrimLeft(strArr[1], "/")); err != nil {
		log.Error("%v", err)
		return
	}
	if mainMR, _, err = s.gitlabClient.MergeRequests.GetMergeRequest(mainPrjID, mainMRIID); err != nil {
		log.Error("can't find main repo's merge request: %v", err)
		return
	}
	// 子仓大仓状态检查和同步
	if hookMR.ObjectAttributes.SourceBranch != mainMR.SourceBranch {
		err = errors.New("the source branch of sub repo and main repo are not the same")
		log.Error("the source branch of sub repo and main repo are not the same")
		return
	}
	if hookMR.ObjectAttributes.TargetBranch != mainMR.TargetBranch {
		err = errors.New("the target branch of sub repo and main repo are not the same")
		log.Error("the target branch of sub repo and main repo are not the same")
		return
	}
	var opt = &ggl.UpdateMergeRequestOptions{}
	var shouldUpdate = false
	if strings.Contains(mainMR.Description, hookMR.ObjectAttributes.URL) {
		// 如果大仓 MR 关联子仓 MR，但是子仓关闭了 MR，则从大仓中移除关联的子仓 MR
		if hookMR.ObjectAttributes.State == "closed" {
			shouldUpdate = true
			newDesc := mainMR.Description
			strArr = strings.Split(newDesc, "\r\n\r\n"+hookMR.ObjectAttributes.URL)
			newDesc = strings.Join(strArr, "")
			opt.Description = ggl.String(newDesc)
		}
	} else {
		// 如果大仓 MR 没有关联子仓 MR，但是子仓开启了 MR，则大仓关联上子仓的 MR
		if hookMR.ObjectAttributes.State == "opened" && len(mainMR.Description) > 0 {
			shouldUpdate = true
			strArr = strings.Split(mainMR.Description, "related sub repos:")
			newDesc := strings.Join(strArr, "related sub repos:\r\n\r\n"+hookMR.ObjectAttributes.URL)
			opt.Description = ggl.String(newDesc)
		}
	}
	// 根据子仓 Draft 状态，检查大仓 Draft 状态
	if hookMR.ObjectAttributes.WorkInProgress || hookMR.ObjectAttributes.MergeStatus != "can_be_merged" {
		// 子仓是 Draft，主仓不是 Draft，则给主仓加上 Draft
		if !mainMR.WorkInProgress {
			shouldUpdate = true
			newTitle = "Draft:" + mainMR.Title
			opt.Title = ggl.String(newTitle)
		}
	} else {
		// 子仓不是 Draft 状态，且可以 merge
		// 则检查大仓所有的子仓是否都不是 Draft 状态，且可以 merge，且 source branch 和 target branch 与大仓一致，且都是 opened 状态
		if newTitle, err = s.mainRepoMRCheckDraft(mainMR); err != nil {
			log.Error("%v", err)
			return
		}
		if len(newTitle) > 0 && newTitle != mainMR.Title {
			shouldUpdate = true
			opt.Title = ggl.String(newTitle)
		}
	}
	if shouldUpdate {
		if _, _, err = s.gitlabClient.MergeRequests.UpdateMergeRequest(mainPrjID, mainMRIID, opt); err != nil {
			log.Error("%v", err)
			return
		}
	}
	return
}

func (s *Service) mainRepoMRCheckDraft(mainMR *ggl.MergeRequest) (newTitle string, err error) {
	var (
		strArr      []string
		subMR       *ggl.MergeRequest
		subPrjID    string
		subMRIID    int
		removeDraft bool
	)
	removeDraft = true
	strArr = strings.Split(mainMR.Description, "related sub repos:")
	//nolint:gomnd
	if len(strArr) < 2 {
		log.Warn("invalid mainrepo merge request description")
		return
	}
	tmpDesc := strings.TrimSpace(strArr[1])
	strArr = strings.Split(tmpDesc, "\n")
	for _, line := range strArr {
		line = strings.TrimSpace(line)
		line = strings.Replace(line, "https://", "http://", 1)
		if len(line) > 0 && strings.Contains(line, "merge_requests") {
			subPrjID = strings.TrimPrefix(line, s.c.Gitlab.Host)
			if len(subPrjID) < 1 {
				err = errors.New("invalid mainrepo merge request description")
				log.Error("invalid mainrepo merge request description")
				return
			}
			strArrTmp := strings.Split(subPrjID, "merge_requests")
			//nolint:gomnd
			if len(strArrTmp) < 2 {
				err = errors.New("invalid mainrepo merge request description")
				log.Error("invalid mainrepo merge request description")
				return
			}
			subPrjID = strArrTmp[0]
			if strings.Contains(subPrjID, "-/") {
				subPrjID = strings.TrimSuffix(subPrjID, "-/")
			}
			subPrjID = strings.TrimLeft(strings.TrimRight(subPrjID, "/"), "/")
			if subMRIID, err = strconv.Atoi(strings.TrimLeft(strArrTmp[1], "/")); err != nil {
				log.Error("%v", err)
				return
			}
			// 获取对应的 sub repo 的 MR
			if subMR, _, err = s.gitlabClient.MergeRequests.GetMergeRequest(subPrjID, subMRIID); err != nil {
				log.Error("can't find sub repo's merge request: %v", err)
				return
			}
			if subMR.WorkInProgress || subMR.MergeStatus != "can_be_merged" || subMR.State != "opened" || subMR.SourceBranch != mainMR.SourceBranch || subMR.TargetBranch != mainMR.TargetBranch {
				removeDraft = false
				break
			}
		}
	}
	// 检查条件满足则将大仓的 Draft 状态移除，不满足则应保持 Draft 状态
	if removeDraft {
		if strings.Contains(mainMR.Title, "Draft:") {
			newTitle = strings.TrimPrefix(mainMR.Title, "Draft:")
		}
	} else {
		if !strings.Contains(mainMR.Title, "Draft:") {
			newTitle = "Draft:" + mainMR.Title
		}
	}
	return
}

// MainRepoMRHook handle the merge request of main repos
// nolint:gocognit
func (s *Service) MainRepoMRHook(c context.Context, appKey string, hookMR *sagamdl.HookMR) (err error) {
	var (
		subPrjID, newTitle string
		subMRIID           int
		strArr             []string
	)
	s.event.Publish(GitMergeEvent, MergeArgs{AppKey: appKey, HookMr: hookMR})
	if hookMR.ObjectAttributes.State == "merged" {
		// merge 动作以外都不进行任何操作
		if hookMR.ObjectAttributes.Action != "merge" {
			return
		}
		if strings.HasPrefix(hookMR.ObjectAttributes.TargetBranch, "release") {
			_ = s.NotifyMasterMergeRequest(*(*int)(unsafe.Pointer(&hookMR.ObjectAttributes.TargetProjectID)), hookMR.ObjectAttributes.Target.Name, hookMR.ObjectAttributes.SourceBranch, hookMR.ObjectAttributes.TargetBranch, *(*int)(unsafe.Pointer(&hookMR.ObjectAttributes.IID)))
		}
		// 无论如何只要是代码合入，最后都要运行 pipeline 进行检查
		defer func() {
			var variables = map[string]string{
				"APP_KEY":     appKey,
				"FAWKES":      "1",
				"FAWKES_USER": hookMR.User.UserName,
				"TASK":        "build",
			}
			_, _ = s.TriggerPipeline(context.Background(), appKey, 0, hookMR.ObjectAttributes.TargetBranch, variables)
		}()
		// 主仓 MR 合入，关联的子仓 MR 也一起合入，并执行 CI
		strArr = strings.Split(hookMR.ObjectAttributes.Description, "related sub repos:")
		//nolint:gomnd
		if len(strArr) < 2 {
			return
		}
		tmpDesc := strings.TrimSpace(strArr[1])
		strArr = strings.Split(tmpDesc, "\n")
		for _, line := range strArr {
			line = strings.TrimSpace(line)
			line = strings.Replace(line, "https://", "http://", 1)
			if len(line) > 0 && strings.Contains(line, "merge_requests") {
				subPrjID = strings.TrimPrefix(line, s.c.Gitlab.Host)
				if len(subPrjID) < 1 {
					err = errors.New("invalid mainrepo merge request description")
					log.Error("invalid mainrepo merge request description")
					return
				}
				strArrTmp := strings.Split(subPrjID, "merge_requests")
				//nolint:gomnd
				if len(strArrTmp) < 2 {
					err = errors.New("invalid mainrepo merge request description")
					log.Error("invalid mainrepo merge request description")
					return
				}
				subPrjID = strArrTmp[0]
				if strings.Contains(subPrjID, "-/") {
					subPrjID = strings.TrimSuffix(subPrjID, "-/")
				}
				subPrjID = strings.TrimLeft(strings.TrimRight(subPrjID, "/"), "/")
				if subMRIID, err = strconv.Atoi(strings.TrimLeft(strArrTmp[1], "/")); err != nil {
					log.Error("%v", err)
					return
				}
				var opt = &ggl.AcceptMergeRequestOptions{}
				opt.ShouldRemoveSourceBranch = ggl.Bool(true)
				if _, _, err = s.gitlabClient.MergeRequests.AcceptMergeRequest(subPrjID, subMRIID, opt); err != nil {
					log.Error("%v", err)
					return
				}
			}
		}
	} else if hookMR.ObjectAttributes.State == "closed" {
		// close 动作以外都不进行任何操作
		if hookMR.ObjectAttributes.Action != "close" {
			return
		}
		// 主仓 MR 关闭，关联的子仓 MR 也一起关闭
		strArr = strings.Split(hookMR.ObjectAttributes.Description, "related sub repos:")
		//nolint:gomnd
		if len(strArr) < 2 {
			err = errors.New("invalid mainrepo merge request description")
			log.Error("invalid mainrepo merge request description")
			return
		}
		tmpDesc := strings.TrimSpace(strArr[1])
		strArr = strings.Split(tmpDesc, "\n")
		for _, line := range strArr {
			line = strings.TrimSpace(line)
			line = strings.Replace(line, "https://", "http://", 1)
			if len(line) > 0 && strings.Contains(line, "merge_requests") {
				subPrjID = strings.TrimPrefix(line, s.c.Gitlab.Host)
				if len(subPrjID) < 1 {
					err = errors.New("invalid mainrepo merge request description")
					log.Error("invalid mainrepo merge request description")
					return
				}
				strArrTmp := strings.Split(subPrjID, "merge_requests")
				//nolint:gomnd
				if len(strArrTmp) < 2 {
					err = errors.New("invalid mainrepo merge request description")
					log.Error("invalid mainrepo merge request description")
					return
				}
				subPrjID = strArrTmp[0]
				if strings.Contains(subPrjID, "-/") {
					subPrjID = strings.TrimSuffix(subPrjID, "-/")
				}
				subPrjID = strings.TrimLeft(strings.TrimRight(subPrjID, "/"), "/")
				if subMRIID, err = strconv.Atoi(strings.TrimLeft(strArrTmp[1], "/")); err != nil {
					log.Error("%v", err)
					return
				}
				var opt = &ggl.UpdateMergeRequestOptions{}
				opt.StateEvent = ggl.String("close")
				if _, _, err = s.gitlabClient.MergeRequests.UpdateMergeRequest(subPrjID, subMRIID, opt); err != nil {
					log.Error("%v", err)
					return
				}
			}
		}
	} else if hookMR.ObjectAttributes.State == "opened" {
		var mainMR = &ggl.MergeRequest{
			ID:           *(*int)(unsafe.Pointer(&hookMR.ObjectAttributes.ID)),
			IID:          *(*int)(unsafe.Pointer(&hookMR.ObjectAttributes.IID)),
			Description:  hookMR.ObjectAttributes.Description,
			Title:        hookMR.ObjectAttributes.Title,
			TargetBranch: hookMR.ObjectAttributes.TargetBranch,
			SourceBranch: hookMR.ObjectAttributes.SourceBranch,
			State:        hookMR.ObjectAttributes.State,
			ProjectID:    hookMR.Project.ID,
		}
		// 如果是首次创建 MR，则启一个 pipeline
		if hookMR.ObjectAttributes.Action == "open" || hookMR.ObjectAttributes.Action == "reopen" {
			var assignee string
			if hookMR.Assignee == nil {
				assignee = hookMR.User.UserName
			} else {
				assignee = hookMR.Assignee.UserName
			}
			_ = s.TriggerBuild(context.Background(), appKey, hookMR.ObjectAttributes.SourceBranch, assignee, hookMR.Project.GitSSHURL, hookMR.ObjectAttributes.LastCommit.ID, hookMR.ObjectAttributes.TargetBranch, hookMR.ObjectAttributes.SourceBranch)
			return
		}
		if newTitle, err = s.mainRepoMRCheckDraft(mainMR); err != nil {
			log.Error("%v", err)
			return
		}
		if len(newTitle) > 0 && newTitle != mainMR.Title {
			var opt = &ggl.UpdateMergeRequestOptions{}
			opt.Title = ggl.String(newTitle)
			if _, _, err = s.gitlabClient.MergeRequests.UpdateMergeRequest(mainMR.ProjectID, mainMR.IID, opt); err != nil {
				log.Error("%v", err)
				return
			}
		}
	}
	return
}

// MainRepoCommentHook hook comment for main repo
func (s *Service) MainRepoCommentHook(c context.Context, appKey string, hookComment *sagamdl.HookComment) (err error) {
	s.event.Publish(GitCommentEvent, CommentArgs{AppKey: appKey, HookComment: hookComment})
	if hookComment.ObjectAttributes.NoteableType == sagamdl.HookCommentTypeMR {
		if hookComment.MergeRequest.State != "opened" || hookComment.MergeRequest.WorkInProgress || hookComment.MergeRequest.MergeStatus != "can_be_merged" {
			return
		}
		// 解决主仓没有修改时，saga 因为找不准子仓触发的 pipeline 而无法合并的问题
		if hookComment.User.UserName == "saga" && strings.Contains(hookComment.MergeRequest.LastCommit.Message, "[skip ci]") {
			var (
				gitlabProjectID  string
				opt              = &ggl.AcceptMergeRequestOptions{}
				mrIID, commentID int
				comment          *ggl.Note
			)
			if gitlabProjectID, err = s.fkDao.GitlabProjectID(c, appKey); err != nil {
				log.Error("s.fkDao.GitlabProjectID error(%v)", err)
				return
			}
			mrIID = *(*int)(unsafe.Pointer(&hookComment.MergeRequest.IID))
			commentID = *(*int)(unsafe.Pointer(&hookComment.ObjectAttributes.ID))
			// 获取最新的 comment
			if comment, _, err = s.gitlabClient.Notes.GetMergeRequestNote(gitlabProjectID, mrIID, commentID); err != nil {
				log.Error("%v", err)
				return
			}
			if strings.Contains(comment.Body, "pipeline还未成功") {
				opt.ShouldRemoveSourceBranch = ggl.Bool(true)
				if _, _, err = s.gitlabClient.MergeRequests.AcceptMergeRequest(gitlabProjectID, mrIID, opt); err != nil {
					log.Error("%v", err)
					return
				}
			}
		}
	}
	return
}

// MainRepoReleaseBranchHook release branch create webhook to lock bapis commit
func (s *Service) MainRepoReleaseBranchHook(c context.Context, appKey string, hookPush *sagamdl.HookPush) (err error) {
	if !strings.HasPrefix(hookPush.Ref, "refs/heads/release") {
		return
	}
	if hookPush.Before != "0000000000000000000000000000000000000000" {
		return
	}
	// 只有新建的 release 分支才会锁 bapis 的 commit
	var variables = map[string]string{
		"APP_KEY": appKey,
		"FAWKES":  "1",
		"TASK":    "lock_bapis",
	}
	if _, err = s.TriggerPipeline(c, appKey, 0, hookPush.Ref, variables); err != nil {
		log.Error("TriggerBuild error: %v", err)
		return
	}
	return
}

// TriggerBuild trigger main repo pipeline for building
func (s *Service) TriggerBuild(c context.Context, appKey, branchName, userName, gitURL, commit, targetBranch, sourceBranch string) (err error) {
	var (
		gitlabProjectID       string
		pipeline              *ggl.Pipeline
		mergeRequests, subMRs []*ggl.MergeRequest
		triggerError          error
	)
	if gitlabProjectID, err = s.fkDao.GitlabProjectID(c, appKey); err != nil {
		log.Error("s.fkDao.GitlabProjectID error(%v)", err)
		return
	}
	var opt = &ggl.ListProjectMergeRequestsOptions{
		SourceBranch: ggl.String(branchName),
		State:        ggl.String("opened"),
	}
	var subRepoName = ""
	strArr := strings.Split(gitURL, ":")
	//nolint:gomnd
	if len(strArr) == 2 {
		subRepoName = strArr[1]
		subRepoName = strings.Split(subRepoName, ".")[0]
	}
	if mergeRequests, _, err = s.gitlabClient.MergeRequests.ListProjectMergeRequests(gitlabProjectID, opt); err != nil {
		log.Error("%v", err)
		return
	}
	// iOS 的特殊处理
	if appKey == "iphone" {
		if len(mergeRequests) == 0 {
			return
		}
		if subRepoName != "ios/loktar" {
			// 如果是子仓触发需要检查子仓是否有 MR
			if subMRs, _, err = s.gitlabClient.MergeRequests.ListProjectMergeRequests(subRepoName, opt); err != nil {
				log.Error("%v", err)
				return
			}
			if len(subMRs) == 0 {
				return
			}
		}
	}
	var variables = map[string]string{
		"APP_KEY":         appKey,
		"FAWKES":          "1",
		"FAWKES_USER":     userName,
		"TASK":            "build",
		"SUB_REPO":        gitURL,
		"SUB_REPO_COMMIT": commit,
		"TARGET_BRANCH":   targetBranch,
		"SOURCE_BRANCH":   sourceBranch,
	}
	if pipeline, triggerError = s.TriggerPipeline(c, appKey, 0, branchName, variables); triggerError != nil {
		log.Error("TriggerBuild error: %v", err)
		return
	}
	for _, mergeRequest := range mergeRequests {
		// MR Comments 追加 pipeline 地址
		var opt = &ggl.CreateMergeRequestNoteOptions{}
		opt.Body = ggl.String("Triggered pipeline from repo " + subRepoName + ":\n\n" + conf.Conf.Gitlab.Host + "/" + gitlabProjectID + "/pipelines/" + strconv.Itoa(pipeline.ID))
		_, _, _ = s.gitlabClient.Notes.CreateMergeRequestNote(gitlabProjectID, mergeRequest.IID, opt)
	}
	return
}

// PipelineStatus get pipeline status
func (s *Service) PipelineStatus(c context.Context, appKey string, pipelineID int) (pipeline *ggl.Pipeline, err error) {
	var (
		gitlabProjectID string
	)
	if gitlabProjectID, err = s.fkDao.GitlabProjectID(c, appKey); err != nil {
		log.Error("s.fkDao.GitlabProjectID error(%v)", err)
		return
	}
	if pipeline, _, err = s.gitlabClient.Pipelines.GetPipeline(gitlabProjectID, pipelineID); err != nil {
		log.Error("get pipeline error(%v)", err)
	}
	return
}

// CheckoutBranch checkout same name branches from main repo & sub repos
func (s *Service) CheckoutBranch(c context.Context, repoID int, srcBranch, tgtBranch string) (err error) {
	var (
		babelFile []byte
		opt       = &ggl.GetRawFileOptions{
			Ref: ggl.String("master"),
		}
		bchOpt = &ggl.CreateBranchOptions{
			Branch: ggl.String(tgtBranch),
			Ref:    ggl.String(srcBranch),
		}
	)
	if babelFile, _, err = s.gitlabClient.RepositoryFiles.GetRawFile(repoID, "babelfile", opt); err != nil {
		log.Error("get raw file error(%v)", err)
		return
	}
	babelFileContent := string(babelFile)
	babelFileContent = strings.TrimSpace(babelFileContent)
	strArr := strings.Split(babelFileContent, "\n")
	for _, line := range strArr {
		line = strings.Split(line, "#")[0]
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		lineArr := strings.SplitN(line, ":", 2)
		if len(lineArr) <= 1 {
			err = fmt.Errorf("invalid babelfile format: %s", line)
			log.Error(fmt.Sprintf("invalid babelfile format: %s", line))
			return
		}
		subRepoGitInfo := strings.TrimSpace(lineArr[1])
		lineArr = strings.Split(subRepoGitInfo, " ")
		// 如果右侧还有空格，说明是指定了 branch/tag/commit，这种情况不切同名分支
		if len(lineArr) > 1 {
			continue
		}
		subRepoID := strings.TrimSuffix(subRepoGitInfo, ".git\"")
		lineArr = strings.SplitN(subRepoID, ":", 2)
		if len(lineArr) <= 1 {
			err = fmt.Errorf("invalid git ssh url: %s", subRepoGitInfo)
			log.Error(fmt.Sprintf("invalid git ssh url: %s", subRepoGitInfo))
			return
		}
		subRepoID = lineArr[1]
		if _, _, err = s.gitlabClient.Branches.CreateBranch(subRepoID, bchOpt); err != nil {
			log.Error("create branch in %v error(%v)", subRepoID, err)
			return
		}
	}
	// 主仓库最后切
	if _, _, err = s.gitlabClient.Branches.CreateBranch(repoID, bchOpt); err != nil {
		log.Error("create branch in %v error(%v)", repoID, err)
	}
	return
}

// NotifyMasterMergeRequest notify master merge request
func (s *Service) NotifyMasterMergeRequest(projectID int, appName, sourceBranch, TargetBranch string, mergeRequestIID int) (err error) {
	var (
		reqURL     string
		req        *http.Request
		data       []byte
		sagaReq    *model.SagaReq
		sagaRes    *model.SagaRes
		sagatoList []string
		notes      []*ggl.Note
		opt        = &ggl.ListMergeRequestNotesOptions{
			Page:    1,
			PerPage: 100,
		}
	)
	if notes, _, err = s.gitlabClient.Notes.ListMergeRequestNotes(projectID, mergeRequestIID, opt); err != nil {
		log.Error("gitlab ListMergeRequestNotes error(%v)", err)
		return
	}
	for _, note := range notes {
		if note.Body == "+mr" || note.Body == "+merge" {
			sagatoList = append(sagatoList, note.Author.Username)
			break
		}
	}
	if len(sagatoList) == 0 {
		return
	}
	weChatContent := fmt.Sprintf("【Fawkes】小贴士提醒您：您在 %v 的分支 %v 已经合入发布分支 %v，记得也要提 MR 到 master 主干哦！（如已提过请忽略）", appName, sourceBranch, TargetBranch)
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

// CancelTribeJob cancel a gitlab job
func (s *Service) CancelTribeJob(c context.Context, buildID int64) (err error) {
	var buildPack *tribe.BuildPack
	if buildPack, err = s.fkDao.SelectTribeBuildPackById(c, buildID); err != nil {
		log.Error("SelectTribeBuildPacksById(%v) error(%v)", buildID, err)
		return
	}
	if buildPack == nil {
		return
	}
	if _, _, err = s.gitlabClient.Jobs.CancelJob(buildPack.GlPrjId, int(buildPack.GlJobId)); err != nil {
		log.Error("CancelJob(%v, %v) error(%v)", buildPack.GlPrjId, buildPack.GlJobId, err)
	}
	return
}

func (s *Service) getProjectID(gitlabProject string) (projID int, err error) {
	var proj *ggl.Project
	if proj, _, err = s.gitlabClient.Projects.GetProject(gitlabProject); err != nil || proj == nil {
		if err != nil {
			log.Error("GetProject(%v) error(%v)", gitlabProject, err)
		} else {
			err = fmt.Errorf("cannot find proj: %s", gitlabProject)
			log.Error("Cannot find proj: %s", gitlabProject)
		}
		return 0, err
	}
	return proj.ID, err
}

func (s *Service) findMergeRequest(projectID int, sourceBranch, targetBranch string) (mergeRequest *ggl.MergeRequest, err error) {
	var (
		mergeRequests       []*ggl.MergeRequest
		listMergeRequestOpt = &ggl.ListProjectMergeRequestsOptions{
			State:        ggl.String("opened"),
			SourceBranch: ggl.String(sourceBranch),
			TargetBranch: ggl.String(targetBranch),
			View:         ggl.String("simple"),
			Sort:         ggl.String("desc"),
		}
	)
	if mergeRequests, _, err = s.gitlabClient.MergeRequests.ListProjectMergeRequests(projectID, listMergeRequestOpt); err != nil {
		log.Error("ListMergeRequests(%v %v %v) error(%v)", projectID, sourceBranch, targetBranch, err)
		return
	}
	if len(mergeRequests) > 0 {
		return mergeRequests[0], err
	}
	return nil, nil
}

func (s *Service) createMergeRequest(projectID, assigneeID int, sourceBranch, targetBranch, title, description string) (mergeRequest *ggl.MergeRequest, err error) {
	var (
		createMergeRequestOpt = &ggl.CreateMergeRequestOptions{
			Title:              ggl.String(title),
			Description:        ggl.String(description),
			SourceBranch:       ggl.String(sourceBranch),
			TargetBranch:       ggl.String(targetBranch),
			RemoveSourceBranch: ggl.Bool(true),
			AssigneeID:         ggl.Int(assigneeID),
		}
	)
	if mergeRequest, _, err = s.gitlabClient.MergeRequests.CreateMergeRequest(projectID, createMergeRequestOpt); err != nil {
		log.Error("CreateMergeRequest(%v %v %v) error(%v)", projectID, sourceBranch, targetBranch, err)
		return
	}
	return
}

func (s *Service) updateMergeRequestDesc(projectID, mergeRequestIID int, description string) (err error) {
	var (
		updateMergeRequestOpt = &ggl.UpdateMergeRequestOptions{
			Description: ggl.String(description),
		}
	)
	if _, _, err = s.gitlabClient.MergeRequests.UpdateMergeRequest(projectID, mergeRequestIID, updateMergeRequestOpt); err != nil {
		log.Error("UpdateMergeRequest(%v %v %v) error(%v)", projectID, mergeRequestIID, description, err)
		return
	}
	return
}

// CreateLinkedMR create related merge requests from main repo & sub repos
func (s *Service) CreateRelatedMR(c context.Context, gitlabProject, assigneeName, subReposStr, sourceBranch, targetBranch string) (resp *gitmdl.CreateRelatedMRResp, err error) {
	var (
		users            []*ggl.User
		mainProjID       int
		subProjID        int
		assigneeID       int
		mainDesc         string
		mainMergeRequest *ggl.MergeRequest
		subMergeRequest  *ggl.MergeRequest
		listUsersOpt     = &ggl.ListUsersOptions{
			Username: ggl.String(assigneeName),
		}
	)
	resp = &gitmdl.CreateRelatedMRResp{
		Message: "Success",
	}
	if users, _, err = s.gitlabClient.Users.ListUsers(listUsersOpt); err != nil {
		log.Error("ListUsers(%v) error(%v)", assigneeName, err)
		resp.Message = fmt.Sprintf("Cannot find user named %s.", assigneeName)
		return
	}
	if len(users) == 0 {
		err = fmt.Errorf("no user named: %s", assigneeName)
		log.Error("No User named: %v", assigneeName)
		resp.Message = fmt.Sprintf("Cannot find user named %s.", assigneeName)
		return
	}
	assigneeID = users[0].ID
	if mainProjID, err = s.getProjectID(gitlabProject); err != nil {
		log.Error("getProjectID(%v) error(%v)", gitlabProject, err)
		resp.Message = fmt.Sprintf("Cannot find repo named %s.", gitlabProject)
		return
	}
	if mainMergeRequest, err = s.findMergeRequest(mainProjID, sourceBranch, targetBranch); err != nil {
		log.Error("findMergeRequest(%v, %v, %v) error(%v)", gitlabProject, sourceBranch, targetBranch, err)
		resp.Message = fmt.Sprintf("Find merge request error in %s.", gitlabProject)
		return
	}
	var tmpMainDesc = "========================= ↓↓↓请勿删除以下文字↓↓↓ =========================\n\nrelated sub repos:\n\n========================= ↑↑↑请勿删除以上文字↑↑↑ =========================\n\n"
	if mainMergeRequest == nil {
		if mainMergeRequest, err = s.createMergeRequest(mainProjID, assigneeID, sourceBranch, targetBranch, "Draft:"+sourceBranch, ""); err != nil {
			log.Error("createMergeRequest(%v, %v, %v) error(%v)", gitlabProject, sourceBranch, targetBranch, err)
			resp.Message = fmt.Sprintf("Cannot create merge request in %s.", gitlabProject)
			return
		}
	}
	resp.Mainrepo = &gitmdl.MRInfo{
		Repo:  gitlabProject,
		MrURL: mainMergeRequest.WebURL,
	}
	// 无子仓直接结束
	if len(subReposStr) == 0 {
		return
	}
	mainDesc = mainMergeRequest.Description
	// 如果 MR 不包含关键字则加上
	var mainMRKeyword = "related sub repos:"
	if !strings.Contains(mainDesc, mainMRKeyword) {
		mainDesc = mainDesc + "\n\n" + tmpMainDesc
	}
	resp.Subrepos = make([]*gitmdl.MRInfo, 0)
	// 通过关键字拆分 description
	var mainDescArr = strings.Split(mainDesc, mainMRKeyword)
	mainDesc = mainDescArr[0] + mainMRKeyword
	for _, subrepo := range strings.Split(subReposStr, ",") {
		if subProjID, err = s.getProjectID(subrepo); err != nil {
			log.Error("getProjectID(%v) error(%v)", subrepo, err)
			resp.Message = fmt.Sprintf("Cannot find repo named %s.", subrepo)
			return
		}
		if subMergeRequest, err = s.findMergeRequest(subProjID, sourceBranch, targetBranch); err != nil {
			log.Error("findMergeRequest(%v, %v, %v) error(%v)", subrepo, sourceBranch, targetBranch, err)
			resp.Message = fmt.Sprintf("Find merge request error in %s.", subrepo)
			return
		}
		if subMergeRequest == nil {
			var tmpSubDesc = "========================= ↓↓↓请勿删除以下文字↓↓↓ =========================\n\nrelated main repo:\n\n" + mainMergeRequest.WebURL + "\n\n========================= ↑↑↑请勿删除以上文字↑↑↑ =========================\n\n"
			if subMergeRequest, err = s.createMergeRequest(subProjID, assigneeID, sourceBranch, targetBranch, "Draft:"+sourceBranch, tmpSubDesc); err != nil {
				log.Error("createMergeRequest(%v, %v, %v) error(%v)", subrepo, sourceBranch, targetBranch, err)
				resp.Message = fmt.Sprintf("Cannot create merge request in %s.", subrepo)
				return
			}
		}
		subMRInfo := &gitmdl.MRInfo{
			Repo:  subrepo,
			MrURL: subMergeRequest.WebURL,
		}
		resp.Subrepos = append(resp.Subrepos, subMRInfo)
		if !strings.Contains(mainDescArr[1], subMergeRequest.WebURL) {
			mainDesc = mainDesc + "\n\n" + subMergeRequest.WebURL
		}
	}
	mainDesc = mainDesc + mainDescArr[1]
	if err = s.updateMergeRequestDesc(mainProjID, mainMergeRequest.IID, mainDesc); err != nil {
		log.Error("updateMergeRequestDesc(%v, %v) error(%v)", gitlabProject, mainMergeRequest.IID, err)
		resp.Message = fmt.Sprintf("Update merge request error in %s.", mainMergeRequest.WebURL)
		return
	}
	return
}

// RemoteCommitFromBranch get remote commit from
func (s *Service) RemoteCommitFormBranch(c context.Context, gitlabProject, branchName string) (resp *gitmdl.BranchCommitResp, err error) {
	var (
		branch *ggl.Branch
	)
	resp = &gitmdl.BranchCommitResp{
		Message: "Success",
	}
	if branch, _, err = s.gitlabClient.Branches.GetBranch(gitlabProject, branchName); err != nil {
		log.Error("GetBranch(%v, %v) error(%v)", gitlabProject, branchName, err)
		resp.Message = fmt.Sprintf("No branch named \"%s\" in remote, please push your local branch to remote.", branchName)
		return
	}
	fmt.Println(branch.Commit.ID)
	resp.CommitSHA = branch.Commit.ID
	return
}
