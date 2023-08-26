package service

import (
	"context"
	"fmt"

	"go-common/library/conf/paladin.v2"
	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-gw/baas/internal/dao"
	"go-gateway/app/app-svr/app-gw/baas/internal/model"
	"go-gateway/app/app-svr/app-gw/baas/utils/sets"
)

type RoleService struct {
	ac   *paladin.Map
	dao  dao.Dao
	role *RoleManager
}

func newRoleService(d dao.Dao) *RoleService {
	s := &RoleService{
		ac:  &paladin.TOML{},
		dao: d,
	}
	s.role = newRoleManager(s.ac, s.authZByRole)
	if err := paladin.Watch("application.toml", s.ac); err != nil {
		panic(err)
	}
	return s
}

func newRoleManager(ac *paladin.Map, authzFunc func(context.Context, model.RoleContext) error) *RoleManager {
	return &RoleManager{
		ac:        ac,
		authzFunc: authzFunc,
	}
}

type RoleManager struct {
	ac        *paladin.Map
	authzFunc func(context.Context, model.RoleContext) error
}

func (s *RoleService) RoleAuthZ() func(*bm.Context) {
	return s.role.AuthZ
}

func (rm *RoleManager) AuthZ(ctx *bm.Context) {
	roleCtx := model.RoleContext{}
	if err := ctx.Bind(&roleCtx); err != nil {
		return
	}
	username, _ := ctx.Get("username")
	roleCtx.Username = username.(string)
	roleCtx.Cookie = ctx.Request.Header.Get("Cookie")
	if err := rm.doAuthZ(ctx, roleCtx); err != nil {
		ctx.JSON(nil, err)
		ctx.Abort()
		return
	}
}

func (rm *RoleManager) doAuthZ(ctx context.Context, roleCtx model.RoleContext) error {
	if err := rm.authzFunc(ctx, roleCtx); err != nil {
		return err
	}
	return nil
}

func (s *RoleService) authZByRole(ctx context.Context, roleCtx model.RoleContext) error {
	return s.permittedApp(ctx, roleCtx.Username, roleCtx.Cookie, roleCtx.TreeID)
}

func (s *RoleService) permittedApp(ctx context.Context, username, cookie string, treeID int64) error {
	nodes, err := s.dao.FetchRoleTree(ctx, username, cookie)
	if err != nil {
		return err
	}
	treeIDSet := makeTreeIDSet(nodes)
	if !treeIDSet.Has(treeID) {
		return ecode.Error(ecode.AccessDenied, fmt.Sprintf("denied on %s to %d", username, treeID))
	}

	return nil
}

func makeTreeIDSet(nodes []*model.Node) sets.Int64 {
	treeIDSet := sets.NewInt64()
	for _, node := range nodes {
		treeIDSet.Insert(int64(node.TreeID))
	}
	return treeIDSet
}
