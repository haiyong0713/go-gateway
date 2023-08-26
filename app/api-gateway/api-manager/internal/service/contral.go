package service

import (
	"context"
	"fmt"

	"go-common/library/ecode"
	"go-common/library/log"

	"go-gateway/app/api-gateway/api-manager/internal/model"
)

const _defauleDomain = "lego.bilibili.co"

func (s *Service) GroupAdd(c context.Context, req *model.ContralGroupSaveReq, username string, _ int64) (string, error) {
	// 分组名校验
	tmp, err := s.dao.GroupByName(c, req.GroupName)
	if err != nil {
		log.Error("%+v", err)
		return "分组校验查询失败", err
	}
	if tmp != nil {
		return "分组已存在", ecode.RequestErr
	}
	// 落库
	args := &model.ContralGroup{
		GroupName: req.GroupName,
		Creator:   username,
		Modifier:  username,
		Desc:      req.Desc,
	}
	if _, err = s.dao.GroupInsert(c, args); err != nil {
		log.Error("%+v", err)
		return "分组创建失败", err
	}
	return "", nil
}

func (s *Service) GroupEdit(c context.Context, req *model.ContralGroupSaveReq, username string, _ int64) (string, error) {
	// 数据合法校验
	tmp, err := s.dao.GroupByIDs(c, []int64{req.ID})
	if err != nil {
		log.Error("%+v", err)
		return "分组校验查询失败", err
	}
	if tmp == nil {
		return "分组不存在", ecode.RequestErr
	}
	// 落库
	args := &model.ContralGroup{
		ID:       req.ID,
		Modifier: username,
		Desc:     req.Desc,
	}
	if _, err = s.dao.GroupUpdate(c, args); err != nil {
		log.Error("%+v", err)
		return "分组编辑失败", err
	}
	return "", nil
}

func (s *Service) GroupList(c context.Context, req *model.ContralGroupListReq) (*model.ContralGroupListReply, error) {
	list, err := s.dao.GroupList(c, req.GroupName, req.PageNum, req.PageSize)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	count, err := s.dao.GroupCount(c)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	return &model.ContralGroupListReply{
		Page: &model.ContralListPage{
			Total:    count,
			PageNum:  req.PageNum,
			PageSize: req.PageSize,
		},
		List: list,
	}, nil
}

func (s *Service) GroupFollowAction(c context.Context, req *model.ContralGroupFollowActionPeq, username string, _ int64) (string, error) {
	// 数据合法校验
	tmp, err := s.dao.GroupByIDs(c, []int64{req.Id})
	if err != nil {
		log.Error("%+v", err)
		return "分组查询失败", err
	}
	if tmp == nil {
		return "分组不存在", ecode.RequestErr
	}
	// 落库
	switch req.State {
	case "follow":
		if _, err = s.dao.GroupFollowAdd(c, req, username); err != nil {
			log.Error("%+v", err)
			return "添加关注失败", err
		}
	case "unfollow":
		if _, err = s.dao.GroupFollowDel(c, req, username); err != nil {
			log.Error("%+v", err)
			return "取消关注失败", err
		}
	}
	return "", nil
}

func (s *Service) GroupFollowList(c context.Context, username string, uid int64) (*model.ContralGroupListReply, error) {
	// 获取关注的分组列表
	tmp, err := s.dao.GroupFollowList(c, username)
	if err != nil {
		log.Error("%v", err)
		return nil, err
	}
	var list []*model.ContralGroup
	if len(tmp) > 0 {
		list, err = s.dao.GroupByIDs(c, tmp)
		if err != nil {
			log.Error("%v", err)
			return nil, err
		}
	}
	return &model.ContralGroupListReply{
		Page: &model.ContralListPage{
			Total: int64(len(tmp)),
		},
		List: list,
	}, nil
}

func (s *Service) ApiAdd(c context.Context, req *model.ContralApiAddReq, username string, _ int64) (string, error) {
	// 合法性校验
	switch req.ApiType {
	case "http":
		if req.Domain == "" {
			req.Domain = _defauleDomain
		}
	case "grpc":
	default:
		return "接口类型错误", ecode.RequestErr
	}
	// 分组名校验
	tmp, err := s.dao.ApiByName(c, req.ApiName)
	if err != nil {
		log.Error("%+v", err)
		return "接口校验查询失败", err
	}
	if tmp != nil {
		return "接口已存在", ecode.RequestErr
	}
	// 落库
	args := &model.ContralApi{
		Gid:        req.Gid,
		ApiName:    req.ApiName,
		ApiType:    req.ApiType,
		Domain:     req.Domain,
		Router:     req.Router,
		Handler:    req.Handler,
		Req:        req.Req,
		Reply:      req.Reply,
		DSLCode:    req.DSLCode,
		DSLStruct:  req.DSLStruct,
		CustomCode: req.CustomCode,
		Creator:    username,
		Modifier:   username,
		Desc:       req.Desc,
	}
	if _, err = s.dao.ApiInsert(c, args); err != nil {
		log.Error("%+v", err)
		return "接口创建失败", err
	}
	return "", nil
}

func (s *Service) ApiEdit(c context.Context, req *model.ContralApiEditReq, username string, _ int64) (string, error) {
	// 合法性校验
	switch req.ApiType {
	case "http":
		if req.Domain == "" {
			req.Domain = _defauleDomain
		}
	case "grpc":
	default:
		return "接口类型错误", ecode.RequestErr
	}
	tmp, err := s.dao.ApiByIDs(c, []int64{req.ID})
	if err != nil {
		log.Error("%+v", err)
		return "接口校验查询失败", err
	}
	if tmp == nil {
		return "接口不存在", ecode.RequestErr
	}
	// 落库
	args := &model.ContralApi{
		ID:         req.ID,
		ApiType:    req.ApiType,
		Domain:     req.Domain,
		Router:     req.Router,
		Handler:    req.Handler,
		Req:        req.Req,
		Reply:      req.Reply,
		DSLCode:    req.DSLCode,
		DSLStruct:  req.DSLStruct,
		CustomCode: req.CustomCode,
		Modifier:   username,
		Desc:       req.Desc,
	}
	if _, err = s.dao.ApiUpdate(c, args); err != nil {
		log.Error("%+v", err)
		return "接口编辑失败", err
	}
	return "", nil
}

func (s *Service) ApiList(c context.Context, req *model.ContralApiListReq) (*model.ContralApiListReply, error) {
	list, err := s.dao.ApiList(c, req.GID, req.ApiName, req.PageNum, req.PageSize)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	count, err := s.dao.ApiCount(c, req.GID)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	return &model.ContralApiListReply{
		Page: &model.ContralListPage{
			Total:    count,
			PageNum:  req.PageNum,
			PageSize: req.PageSize,
		},
		List: list,
	}, nil
}

func (s *Service) ApiConfigAdd(c context.Context, req *model.ContralApiConfigAddReq, username string, _ int64) (string, error) {
	// 数据合法校验
	tmp, err := s.dao.ApiByIDs(c, []int64{req.ApiID})
	if err != nil {
		log.Error("%+v", err)
		return "接口查询失败", err
	}
	if tmp == nil {
		return "接口不存在", ecode.RequestErr
	}
	tmp2, err := s.dao.ApiConfigByVersion(c, req.ApiID, req.Version)
	if err != nil {
		log.Error("%+v", err)
		return "配置版本查询失败", err
	}
	if tmp2 != nil {
		return "配置版本已存在", ecode.RequestErr
	}
	// 落库
	apiConfig := &model.ContralApiConfig{
		ApiID:      req.ApiID,
		Version:    req.Version,
		ApiType:    tmp[0].ApiType,
		Domain:     tmp[0].Domain,
		Router:     tmp[0].Router,
		Handler:    tmp[0].Handler,
		Req:        tmp[0].Req,
		Reply:      tmp[0].Reply,
		DSLCode:    tmp[0].DSLCode,
		DSLStruct:  tmp[0].DSLStruct,
		CustomCode: tmp[0].CustomCode,
		Creator:    username,
		Desc:       req.Desc,
	}
	if _, err = s.dao.ApiConfigInsert(c, apiConfig); err != nil {
		log.Error("%+v", err)
		return "配置版本创建失败", err
	}
	return "", nil
}

func (s *Service) ApiConfigRollback(c context.Context, req *model.ContralApiConfigRollbackReq, username string, uid int64) (string, error) {
	// 校验
	tmp, err := s.dao.ApiConfigByID(c, req.ApiConfigID)
	if err != nil {
		log.Error("%+v", err)
		return "配置校验查询失败", err
	}
	if len(tmp) == 0 || tmp[0] == nil {
		return "未找到对应版本", err
	}
	// 落库
	apiConfig := &model.ContralApi{
		ID:         tmp[0].ApiID,
		ApiType:    tmp[0].ApiType,
		Domain:     tmp[0].Domain,
		Router:     tmp[0].Router,
		Handler:    tmp[0].Handler,
		Req:        tmp[0].Req,
		Reply:      tmp[0].Reply,
		DSLCode:    tmp[0].DSLCode,
		DSLStruct:  tmp[0].DSLStruct,
		CustomCode: tmp[0].CustomCode,
		Modifier:   username,
		Desc:       tmp[0].Desc,
	}
	if _, err = s.dao.ApiUpdate(c, apiConfig); err != nil {
		log.Error("%+v", err)
		return "配置版本回滚失败", err
	}
	return "", nil
}

func (s *Service) ApiConfigList(c context.Context, req *model.ContralApiConfigListReq) (*model.ContralApiConfigListReply, error) {
	list, err := s.dao.ApiConfigList(c, req.ApiID, req.PageNum, req.PageSize)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	count, err := s.dao.ApiConfigCount(c, req.ApiID)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	return &model.ContralApiConfigListReply{
		Page: &model.ContralListPage{
			Total:    count,
			PageNum:  req.PageNum,
			PageSize: req.PageSize,
		},
		List: list,
	}, nil
}

func (s *Service) ApiPublishCallback(c context.Context, req *model.ContralapiPublishCallbackReq, username string, _ int64) (string, error) {
	// 数据合法校验
	tmp, err := s.dao.ApiByIDs(c, []int64{req.ApiID})
	if err != nil {
		log.Error("%+v", err)
		return "接口查询失败", err
	}
	if tmp == nil {
		return "接口不存在", ecode.RequestErr
	}
	// 发布任务落库
	args := &model.ContralApiPublish{
		ApiID:     req.ApiID,
		PublishID: req.PublishID,
		Version:   req.Version,
		State:     req.State,
		Operator:  username,
	}
	if _, err = s.dao.ApiPublishSave(c, args); err != nil {
		log.Error("%+v", err)
		return "发布任务创建/更新滚失败", err
	}
	return "", nil
}

func (s *Service) ApiPublishList(c context.Context, req *model.ContralApiPublishListReq) (*model.ContralApiPublishListReply, error) {
	list, err := s.dao.ApiPublishList(c, req.ApiID, req.PageNum, req.PageSize)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	count, err := s.dao.ApiPublishCount(c, req.ApiID)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	return &model.ContralApiPublishListReply{
		Page: &model.ContralListPage{
			Total:    count,
			PageNum:  req.PageNum,
			PageSize: req.PageSize,
		},
		List: list,
	}, nil
}

func (s *Service) DynPath(c context.Context, apiName string) (*model.DynpathParam, error) {
	tmp, err := s.dao.ApiByName(c, apiName)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	if tmp == nil {
		return nil, ecode.NothingFound
	}
	if tmp[0].ApiType != "http" {
		return nil, ecode.NothingFound
	}
	return &model.DynpathParam{
		Node:    "main.api-gateway-console",
		Gateway: "api-proxy-gateway",
		Pattern: fmt.Sprintf("~ %v", tmp[0].Router),
		//ClientInfo:    "",
		Enable:        1,
		ClientTimeout: 100,
	}, nil
}
