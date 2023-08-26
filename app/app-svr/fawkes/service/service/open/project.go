package open

import (
	"context"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go-common/library/ecode"

	"go-gateway/app/app-svr/fawkes/service/api/app/open"
	openmdl "go-gateway/app/app-svr/fawkes/service/model/open"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
	"go-gateway/app/app-svr/fawkes/service/tools/utils"

	"github.com/golang/protobuf/ptypes/empty"
)

func (s *Service) GetProjectInfoList(ctx context.Context, req *open.GetProjectInfoListReq) (resp *open.GetProjectInfoListResp, err error) {
	var (
		relations        []*openmdl.UserProjectRelation
		projectIdFilter  []int64
		projectRespInfos []*open.ProjectInfo
	)
	op := utils.GetUsername(ctx)
	perm, err := s.isSupervisor(ctx, op)
	if err != nil {
		log.Errorc(ctx, "%v", err)
		err = ecode.Error(ecode.ServerErr, "get supervisor error")
		return
	}
	if !perm {
		// 普通用户 查询是否存在project
		if relations, err = s.fkDao.SelectProjectOwnerRelationByUser(ctx, op); err != nil {
			log.Errorc(ctx, "%v", err)
			err = ecode.Error(ecode.ServerErr, "SelectProjectOwnerRelationByUser error")
			return
		}
		if len(relations) == 0 {
			err = ecode.Error(ecode.UserDisabled, fmt.Sprintf("用户[%s]没有创建过的项目", op))
			return
		}
		for _, v := range relations {
			projectIdFilter = append(projectIdFilter, v.ProjectId)
		}
	}
	total, err := s.fkDao.CountProjectInfos(ctx, req.ProjectName, projectIdFilter)
	if err != nil {
		log.Errorc(ctx, "%v", err)
		err = ecode.Error(ecode.ServerErr, "CountProjectInfos error")
		return
	}
	projects, err := s.fkDao.SelectProjectInfos(ctx, req.ProjectName, projectIdFilter, req.Pn, req.Ps)
	if err != nil {
		log.Errorc(ctx, "%v", err)
		err = ecode.Error(ecode.ServerErr, "SelectProjectInfos error")
		return
	}
	for _, v := range projects {
		projectRespInfos = append(projectRespInfos, convert2ProjectRespMdl(v))
	}
	resp = &open.GetProjectInfoListResp{
		PageInfo:    &open.PageInfo{Total: total, Pn: req.Pn, Ps: req.Ps},
		ProjectInfo: projectRespInfos,
	}
	return
}

func (s *Service) CreateProject(ctx context.Context, req *open.CreateProjectReq) (resp *empty.Empty, err error) {
	resp = new(empty.Empty)
	var (
		token     string
		projectId int64
	)
	op := utils.GetUsername(ctx)
	perm, err := s.isSupervisor(ctx, op)
	if err != nil {
		log.Errorc(ctx, "%v", err)
		err = ecode.Error(ecode.ServerErr, "get supervisor error")
		return
	}
	if !perm {
		err = ecode.Error(ecode.Unauthorized, "请联系fawkes超管操作")
		return
	}
	owners := strings.Join(req.Owner, openmdl.Comma)
	if token, err = s.createToken(req.ProjectName); err != nil {
		log.Errorc(ctx, "%v", err)
		err = ecode.Error(ecode.ServerErr, "create token error")
		return
	}
	if projectId, err = s.fkDao.AddProject(ctx, req.ProjectName, owners, token, req.Description, op); err != nil {
		log.Errorc(ctx, "%v", err)
		err = ecode.Error(ecode.ServerErr, "add project error")
		return
	}
	if err = s.fkDao.AddUserRelations(ctx, projectId, req.Owner); err != nil {
		log.Errorc(ctx, "%v", err)
		err = ecode.Error(ecode.ServerErr, "add project user relation error")
		return
	}
	return
}

func (s *Service) GetProjectInfo(ctx context.Context, req *open.GetProjectInfoReq) (resp *open.GetProjectInfoResp, err error) {
	var (
		info      *openmdl.Project
		relations []*openmdl.UserProjectRelation
		owners    []string
	)
	if info, err = s.fkDao.SelectProjectInfo(ctx, req.ProjectId); err != nil {
		log.Errorc(ctx, "%v", err)
		err = ecode.Error(ecode.ServerErr, "select project info error")
		return
	}
	op := utils.GetUsername(ctx)
	perm, err := s.isSupervisor(ctx, op)
	if err != nil {
		log.Errorc(ctx, "%v", err)
		err = ecode.Error(ecode.ServerErr, "get supervisor error")
		return
	}
	if !contains(strings.Split(info.Owner, openmdl.Comma), op) && !perm {
		err = ecode.Error(ecode.RequestErr, fmt.Sprintf("%s不是%s的owner，没有权限", op, info.ProjectName))
		return
	}
	if relations, err = s.fkDao.SelectProjectOwnerRelationByProject(ctx, req.ProjectId); err != nil {
		log.Errorc(ctx, "%v", err)
		err = ecode.Error(ecode.ServerErr, "select project owner relations error")
		return
	}
	for _, v := range relations {
		owners = append(owners, v.UserName)
	}
	resp = &open.GetProjectInfoResp{
		ProjectInfo: &open.ProjectInfo{
			Id:          info.Id,
			ProjectName: info.ProjectName,
			Owner:       owners,
			Description: info.Description,
			Token:       info.Token,
			Applicant:   info.Applicant,
		},
	}
	return
}

func (s *Service) UpdateProject(ctx context.Context, req *open.UpdateProjectReq) (resp *empty.Empty, err error) {
	resp = new(empty.Empty)
	var (
		info *openmdl.Project
	)
	if info, err = s.fkDao.SelectProjectInfo(ctx, req.ProjectId); err != nil {
		log.Errorc(ctx, "%v", err)
		err = ecode.Error(ecode.ServerErr, "select project info error")
		return
	}
	op := utils.GetUsername(ctx)
	perm, err := s.isSupervisor(ctx, op)
	if err != nil {
		log.Errorc(ctx, "%v", err)
		err = ecode.Error(ecode.ServerErr, "get supervisor error")
		return
	}
	if !perm {
		err = ecode.Error(ecode.RequestErr, fmt.Sprintf("%s不是%s的owner，没有权限", op, info.ProjectName))
		return
	}
	if _, err = s.fkDao.UpdateProject(ctx, req.ProjectId, strings.Join(req.Owner, openmdl.Comma), req.Description); err != nil {
		log.Errorc(ctx, "%v", err)
		err = ecode.Error(ecode.ServerErr, "update project error")
		return
	}
	if err = s.fkDao.UpdateProjectOwnerRelations(ctx, req.ProjectId, req.Owner); err != nil {
		log.Errorc(ctx, "%v", err)
		err = ecode.Error(ecode.ServerErr, "update project owner relations error")
		return
	}
	return
}

func (s *Service) ActiveProject(ctx context.Context, req *open.ActiveProjectReq) (resp *empty.Empty, err error) {
	resp = new(empty.Empty)
	utils.GetUsername(ctx)
	err = s.fkDao.UpdateProjectStatus(ctx, req.IsActive, req.ProjectId)
	if err != nil {
		log.Errorc(ctx, "%v", err)
		return
	}
	return
}

func (s *Service) createToken(projectName string) (string, error) {
	claim := openmdl.TokenClaims{
		ProjectName: projectName,
		TimeStamp:   time.Now(),
	}
	data, err := json.Marshal(claim)
	if err != nil {
		return "", err
	}
	sEnc := b64.StdEncoding.EncodeToString(data)
	return sEnc, err
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func convert2ProjectRespMdl(info *openmdl.Project) *open.ProjectInfo {
	return &open.ProjectInfo{
		Id:          info.Id,
		ProjectName: info.ProjectName,
		Owner:       strings.Split(info.Owner, openmdl.Comma),
		Description: info.Description,
		Token:       info.Token,
		Applicant:   info.Applicant,
		IsActive:    info.IsActive,
	}
}
