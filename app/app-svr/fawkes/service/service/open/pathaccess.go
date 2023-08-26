package open

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"go-common/library/ecode"

	"go-gateway/app/app-svr/fawkes/service/api/app/open"
	openmdl "go-gateway/app/app-svr/fawkes/service/model/open"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
	"go-gateway/app/app-svr/fawkes/service/tools/utils"

	"github.com/golang/protobuf/ptypes/empty"
)

type RouterList []*open.Router

func (s *Service) AddPath(ctx context.Context, req *open.AddPathReq) (resp *empty.Empty, err error) {
	resp = new(empty.Empty)
	p, err := s.fkDao.SelectProjectInfo(ctx, req.ProjectId)
	if err != nil {
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
		err = ecode.Error(ecode.Unauthorized, "请联系fawkes超管操作")
		return
	}
	paths, err := s.fkDao.SelectProjectPath(ctx, req.ProjectId)
	if err != nil {
		log.Errorc(ctx, "%v", err)
		err = ecode.Error(ecode.ServerErr, "SelectProjectPath error")
		return
	}
	if containPath(paths, req.RouterAccess) {
		err = ecode.Error(ecode.ServerErr, "存在已添加过的路径地址，请勿重复添加")
		return
	}
	if err = s.fkDao.AddProjectPath(ctx, req.ProjectId, req.RouterAccess); err != nil {
		log.Errorc(ctx, "%v", err)
		err = ecode.Error(ecode.ServerErr, "add project path error")
		return
	}
	// 同步casbin权限
	pt := genAddEventArgs(p.Token, req.RouterAccess)
	s.event.Publish(AuthAddEvent, ctx, AuthAddArg{PT: pt})
	return
}

func (s *Service) UpdatePath(ctx context.Context, req *open.UpdatePathReq) (resp *empty.Empty, err error) {
	resp = new(empty.Empty)
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
	p, err := s.fkDao.SelectProjectInfo(ctx, req.ProjectId)
	if err != nil {
		log.Errorc(ctx, "%v", err)
		err = ecode.Error(ecode.ServerErr, "select project info error")
		return
	}
	paths, err := s.fkDao.SelectProjectPath(ctx, req.ProjectId)
	if err != nil {
		log.Errorc(ctx, "%v", err)
		err = ecode.Error(ecode.ServerErr, "select project path error")
		return
	}
	idNameMap := genIdPathMap(paths)
	_, err = s.fkDao.UpdateProjectPath(ctx, req.PathUpdate)
	if err != nil {
		log.Errorc(ctx, "%v", err)
		err = ecode.Error(ecode.ServerErr, "UpdateProjectPath error")
		return
	}
	oldPt, newPt := genUpdateEventArgs(p.Token, idNameMap, req.PathUpdate)
	s.event.Publish(AuthUpdateEvent, ctx, AuthUpdateArg{OldPT: oldPt, NewPT: newPt})
	return
}

func (s *Service) DeletePath(ctx context.Context, req *open.DeletePathReq) (resp *open.DeletePathResp, err error) {
	var effected int64
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
	p, err := s.fkDao.SelectProjectInfo(ctx, req.ProjectId)
	if err != nil {
		log.Errorc(ctx, "%v", err)
		err = ecode.Error(ecode.ServerErr, "SelectProjectInfo error")
		return
	}
	paths, err := s.fkDao.SelectProjectPaths(ctx, req.PathId)
	if err != nil {
		log.Errorc(ctx, "%v", err)
		err = ecode.Error(ecode.ServerErr, "SelectProjectPaths error")
		return
	}
	if effected, err = s.fkDao.DeleteProjectPath(ctx, req.PathId); err != nil {
		log.Errorc(ctx, "%v", err)
		err = ecode.Error(ecode.ServerErr, "delete project error")
		return
	}
	pt := genDeleteEventArgs(p.Token, paths)
	s.event.Publish(AuthDeleteEvent, ctx, AuthDeleteArg{PT: pt})
	resp = &open.DeletePathResp{
		DeletedCount: effected,
	}
	return
}

func (s *Service) PathList(ctx context.Context, req *open.PathListReq) (resp *open.PathListResp, err error) {
	op := utils.GetUsername(ctx)
	info, err := s.fkDao.SelectProjectInfo(ctx, req.ProjectId)
	if err != nil {
		log.Errorc(ctx, "%v", err)
		err = ecode.Error(ecode.ServerErr, "SelectProjectInfo error")
		return
	}
	if info == nil {
		err = ecode.Error(ecode.ServerErr, fmt.Sprintf("project_id[%d] 不存在", req.ProjectId))
		return
	}
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
	if err != nil {
		log.Errorc(ctx, "%v", err)
		err = ecode.Error(ecode.ServerErr, fmt.Sprintf("select project[%d] error", req.ProjectId))
		return
	}
	paths, err := s.fkDao.SelectProjectPath(ctx, req.ProjectId)
	if err != nil {
		log.Errorc(ctx, "%v", err)
		err = ecode.Error(ecode.ServerErr, "select paths error")
		return
	}
	var routers []*open.RouterAccess
	for _, v := range paths {
		routers = append(routers, &open.RouterAccess{
			Id:          v.Id,
			Path:        v.Router,
			AppKey:      strings.Split(v.AppKey, openmdl.Comma),
			Description: v.Description,
		})
	}
	resp = &open.PathListResp{
		ProjectId:    req.ProjectId,
		ProjectName:  info.ProjectName,
		RouterAccess: routers,
	}
	return
}

func (s *Service) GetOpenApiList(ctx context.Context, req *open.GetOpenApiListReq) (resp *open.GetOpenApiListResp, err error) {
	paths := strings.Split(fmt.Sprintf("%v", s.OpenPaths), " ")
	var list RouterList
	for _, p := range paths {
		if strings.HasPrefix(p, "/x/admin/fawkes/openapi/") {
			list = append(list, &open.Router{Path: p})
		}
	}
	sort.Sort(list)
	resp = &open.GetOpenApiListResp{
		Router: list,
	}
	return
}

func genIdPathMap(paths []*openmdl.PathAccess) map[int64]*openmdl.PathAccess {
	m := make(map[int64]*openmdl.PathAccess)
	for _, v := range paths {
		m[v.Id] = v
	}
	return m
}

func genUpdateEventArgs(token string, nameMap map[int64]*openmdl.PathAccess, update []*open.PathUpdate) (oldToken []*PathToken, newToken []*PathToken) {
	var oldPt []*PathToken
	for _, v := range update {
		t := &PathToken{
			Token: token,
			Path:  nameMap[v.PathId].Router,
		}
		if len(nameMap[v.PathId].AppKey) != 0 {
			t.AppKey = strings.Split(nameMap[v.PathId].AppKey, openmdl.Comma)
		}
		oldPt = append(oldPt, t)
	}

	var newPt []*PathToken
	for _, v := range update {
		t := &PathToken{
			Token:  token,
			Path:   nameMap[v.PathId].Router,
			AppKey: v.AppKey,
		}
		newPt = append(newPt, t)
	}
	return oldPt, newPt
}

func genAddEventArgs(token string, access []*open.RouterAccess) []*PathToken {
	var pt []*PathToken
	for _, v := range access {
		t := &PathToken{
			Token:  token,
			Path:   v.Path,
			AppKey: v.AppKey,
		}
		pt = append(pt, t)
	}
	return pt
}

func genDeleteEventArgs(token string, access []*openmdl.PathAccess) []*PathToken {
	var pt []*PathToken
	for _, v := range access {
		t := &PathToken{
			Token: token,
			Path:  v.Router,
		}
		if len(v.AppKey) != 0 {
			t.AppKey = strings.Split(v.AppKey, openmdl.Comma)
		}
		pt = append(pt, t)
	}
	return pt
}

func containPath(paths []*openmdl.PathAccess, access []*open.RouterAccess) bool {
	for _, p := range paths {
		for _, v := range access {
			if p.Router == v.Path {
				return true
			}
		}
	}
	return false
}

func (r RouterList) Len() int {
	return len(r)
}

func (r RouterList) Less(i int, j int) bool {
	return strings.Compare(r[i].Path, r[j].Path) < 0
}

func (r RouterList) Swap(i int, j int) {
	r[i], r[j] = r[j], r[i]
}
