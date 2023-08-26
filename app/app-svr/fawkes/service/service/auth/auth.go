package auth

import (
	"context"
	"fmt"

	"go-common/library/ecode"

	"github.com/golang/protobuf/ptypes/empty"

	"go-gateway/app/app-svr/fawkes/service/api/app/auth"
	authmdl "go-gateway/app/app-svr/fawkes/service/model/auth"
	mngmdl "go-gateway/app/app-svr/fawkes/service/model/manager"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
	"go-gateway/app/app-svr/fawkes/service/tools/utils"
)

func (s *Service) AddAuthItemGroup(ctx context.Context, req *auth.AddAuthItemGroupReq) (resp *empty.Empty, err error) {
	resp = new(empty.Empty)
	operator := utils.GetUsername(ctx)
	if !s.isSupervisor(ctx, operator) {
		err = ecode.Error(ecode.Unauthorized, "请联系fawkes超管操作")
		return
	}
	group, err := s.fkDao.SelectAuthGroup(ctx, req.GroupName)
	if err != nil {
		log.Errorc(ctx, "%v", err)
		return
	}
	if group != nil {
		err = ecode.Error(ecode.Conflict, "已存在同名权限组")
		return
	}
	_, err = s.fkDao.AddAuthGroup(ctx, req.GroupName, operator)
	if err != nil {
		log.Errorc(ctx, "%v", err)
		err = ecode.Error(ecode.ServerErr, "新建权限组失败")
		return
	}
	return
}

func (s *Service) UpdateAuthItemGroup(ctx context.Context, req *auth.UpdateAuthItemGroupReq) (resp *empty.Empty, err error) {
	resp = new(empty.Empty)
	operator := utils.GetUsername(ctx)
	if !s.isSupervisor(ctx, operator) {
		err = ecode.Error(ecode.Unauthorized, "请联系fawkes超管操作")
		return
	}
	group, err := s.fkDao.SelectAuthGroup(ctx, req.GroupName)
	if err != nil {
		log.Errorc(ctx, "%v", err)
		return
	}
	if group != nil {
		err = ecode.Error(ecode.Conflict, "已存在同名权限组")
		return
	}
	_, err = s.fkDao.UpdateAuthGroup(ctx, req.GroupId, req.GroupName, operator)
	if err != nil {
		log.Errorc(ctx, "%v", err)
		err = ecode.Error(ecode.ServerErr, "更新权限组失败")
		return
	}
	return
}

func (s *Service) AddAuthItem(ctx context.Context, req *auth.AddAuthItemReq) (resp *empty.Empty, err error) {
	resp = new(empty.Empty)
	operator := utils.GetUsername(ctx)
	if !s.isSupervisor(ctx, operator) {
		err = ecode.Error(ecode.Unauthorized, "请联系fawkes超管操作")
		return
	}
	item, err := s.fkDao.SelectAuthItem(ctx, req.ItemName, req.BeUrl, req.UrlParam)
	if err != nil {
		log.Errorc(ctx, "%v", err)
		return
	}
	if item != nil {
		err = ecode.Error(ecode.Conflict, fmt.Sprintf("存在重复权限项\"%s\"-\"%s\"-\"%s\"", req.ItemName, req.BeUrl, req.UrlParam))
		return
	}
	_, err = s.fkDao.AddAuthItem(ctx, req.GroupId, req.ItemName, req.FeKey, req.BeUrl, req.UrlParam, operator)
	if err != nil {
		log.Errorc(ctx, "%v", err)
		err = ecode.Error(ecode.ServerErr, "新建权限项失败")
		return
	}
	return
}

func (s *Service) ActiveAuthItem(ctx context.Context, req *auth.ActiveAuthItemReq) (resp *empty.Empty, err error) {
	resp = new(empty.Empty)
	operator := utils.GetUsername(ctx)
	if !s.isSupervisor(ctx, operator) {
		err = ecode.Error(ecode.Unauthorized, "请联系fawkes超管操作")
		return
	}
	_, err = s.fkDao.UpdateAuthItemActive(ctx, req.ItemId, req.IsActive, operator)
	if err != nil {
		log.Errorc(ctx, "%v", err)
		return
	}
	return
}

func (s *Service) UpdateAuthItem(ctx context.Context, req *auth.UpdateAuthItemReq) (resp *empty.Empty, err error) {
	resp = new(empty.Empty)
	operator := utils.GetUsername(ctx)
	if !s.isSupervisor(ctx, operator) {
		err = ecode.Error(ecode.Unauthorized, "请联系fawkes超管操作")
		return
	}
	var (
		item      *authmdl.Item
		duplicate *authmdl.Item
	)
	if item, err = s.fkDao.SelectAuthItemById(ctx, req.ItemId); err != nil {
		log.Errorc(ctx, "%v", err)
		return
	}
	if item == nil {
		err = ecode.Error(ecode.NothingFound, "权限组不存在")
		return
	}
	if duplicate, err = s.fkDao.SelectAuthItem(ctx, req.ItemName, item.BeUrl, item.UrlParam); err != nil {
		log.Errorc(ctx, "%v", err)
		return
	}
	// 除自身外有其他同名item
	if duplicate != nil && duplicate.Id != req.ItemId {
		err = ecode.Error(ecode.Conflict, fmt.Sprintf("已存在同名权限项%s-%s-%s", item.Name, item.BeUrl, item.UrlParam))
		return
	}
	_, err = s.fkDao.UpdateAuthItem(ctx, req.ItemId, req.ItemName, req.FeKey, req.BeUrl, req.UrlParam, operator)
	if err != nil {
		log.Errorc(ctx, "%v", err)
		err = ecode.Error(ecode.ServerErr, "更新权限项失败")
		return
	}
	return
}

func (s *Service) DeleteAuthItem(ctx context.Context, req *auth.DeleteAuthItemReq) (resp *empty.Empty, err error) {
	resp = new(empty.Empty)
	operator := utils.GetUsername(ctx)
	if !s.isSupervisor(ctx, operator) {
		err = ecode.Error(ecode.Unauthorized, "请联系fawkes超管操作")
		return
	}
	_, err = s.fkDao.DeleteAuthItem(ctx, req.ItemId)
	if err != nil {
		log.Errorc(ctx, "%v", err)
		err = ecode.Error(ecode.ServerErr, "删除权限项失败")
		return
	}
	return
}

func (s *Service) GrantRole(ctx context.Context, req *auth.GrantRoleReq) (resp *empty.Empty, err error) {
	resp = new(empty.Empty)
	operator := utils.GetUsername(ctx)
	if !s.isSupervisor(ctx, operator) {
		err = ecode.Error(ecode.Unauthorized, "请联系fawkes超管操作")
		return
	}
	if len(req.Item) == 0 {
		return
	}
	grantItems := make([]*auth.Grant, 0, len(req.Item))
	deleteItems := make([]*auth.Grant, 0, len(req.Item))
	for _, v := range req.Item {
		if v.IsGranted {
			grantItems = append(grantItems, v)
		} else {
			deleteItems = append(deleteItems, v)
		}
	}
	if _, err = s.fkDao.AddAuthItemRoleRelation(ctx, grantItems, operator); err != nil {
		log.Errorc(ctx, "%v", err)
		err = ecode.Error(ecode.ServerErr, "授权失败")
		return
	}
	if _, err = s.fkDao.DeleteAuthItemRoleRelation(ctx, deleteItems); err != nil {
		log.Errorc(ctx, "%v", err)
		err = ecode.Error(ecode.ServerErr, "授权失败")
		return
	}
	return
}

func (s *Service) ListAuth(ctx context.Context, req *auth.ListAuthReq) (resp *auth.ListAuthResp, err error) {
	var (
		groups    []*authmdl.Group
		items     []*authmdl.Item
		relations []*authmdl.ItemRoleRelation
	)
	// 查询所有group
	if groups, err = s.fkDao.SelectAllAuthGroup(ctx); err != nil {
		log.Errorc(ctx, "%v", err)
		err = ecode.Error(ecode.ServerErr, "查询数据失败")
		return
	}
	// 查询所有权限点
	if items, err = s.fkDao.SelectAllIAuthItem(ctx); err != nil {
		log.Errorc(ctx, "%v", err)
		err = ecode.Error(ecode.ServerErr, "查询数据失败")
		return
	}
	// 查询所有relation
	if relations, err = s.fkDao.SelectAllAuthRelation(ctx); err != nil {
		log.Errorc(ctx, "%v", err)
		err = ecode.Error(ecode.ServerErr, "查询数据失败")
		return
	}
	// 按照item_id 分组关系
	relationMap := groupRelationByItem(relations)
	// 构建权限项 并按照group_id分组
	itemGroup := groupItem(items, relationMap)

	respGroups := make([]*auth.Group, 0, len(groups))
	for _, v := range groups {
		g := groupDo2To(v)
		g.Item = itemGroup[g.GroupId]
		respGroups = append(respGroups, g)
	}
	// 权限角色
	roles, err := s.fkDao.AuthRole(ctx)
	//
	respRole := make([]*auth.Role, 0, len(roles))
	for _, r := range roles {
		respRole = append(respRole, roleDo2To(r))
	}
	resp = &auth.ListAuthResp{
		Group: respGroups,
		Role:  respRole,
	}
	return
}

// groupItem 将数据库的item构建成resp中的item结构（加上权限关系表中的role_id） 再按照group_id分组
func groupItem(items []*authmdl.Item, relationMap map[int64][]int64) map[int64][]*auth.Item {
	itemGroup := make(map[int64][]*auth.Item)
	for _, item := range items {
		to := itemDo2To(item)
		to.RoleAccess = relationMap[to.ItemId]
		//分组
		if _, ok := itemGroup[to.GroupId]; !ok {
			itemGroup[to.GroupId] = []*auth.Item{to}
		} else {
			itemGroup[to.GroupId] = append(itemGroup[to.GroupId], to)
		}
	}
	return itemGroup
}

func groupRelationByItem(relations []*authmdl.ItemRoleRelation) map[int64][]int64 {
	relationMap := make(map[int64][]int64, len(relations))
	for _, v := range relations {
		if _, ok := relationMap[v.AuthItemId]; !ok {
			relationMap[v.AuthItemId] = []int64{int64(v.AuthRoleValue)}
		} else {
			relationMap[v.AuthItemId] = append(relationMap[v.AuthItemId], int64(v.AuthRoleValue))
		}
	}
	return relationMap
}

func itemDo2To(item *authmdl.Item) *auth.Item {
	return &auth.Item{
		ItemId:   item.Id,
		ItemName: item.Name,
		GroupId:  item.AuthGroupId,
		FeKey:    item.FeKey,
		BeUrl:    item.BeUrl,
		UrlParam: item.UrlParam,
		Operator: item.Operator,
		IsActive: item.IsActive,
		Ctime:    item.Ctime.Unix(),
		Mtime:    item.Mtime.Unix(),
	}
}

func groupDo2To(group *authmdl.Group) *auth.Group {
	return &auth.Group{
		GroupId:   group.ID,
		GroupName: group.Name,
		Operator:  group.Operator,
		Ctime:     group.Ctime.Unix(),
		Mtime:     group.Mtime.Unix(),
	}
}

func roleDo2To(role *mngmdl.Role) *auth.Role {
	return &auth.Role{
		Id:    role.ID,
		Name:  role.Name,
		EName: role.EName,
		Value: int64(role.Value),
		State: int64(role.State),
	}
}
