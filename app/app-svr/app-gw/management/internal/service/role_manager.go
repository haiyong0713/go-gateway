package service

import (
	"context"
	"fmt"

	"go-common/library/conf/paladin.v2"
	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-gw/management/internal/model"
	"go-gateway/app/app-svr/app-gw/management/internal/model/sets"
)

type RoleManager struct {
	ac        *paladin.Map
	authzFunc func(context.Context, model.RoleContext) error
}

func newRoleManager(ac *paladin.Map, authzFunc func(context.Context, model.RoleContext) error) *RoleManager {
	return &RoleManager{
		ac:        ac,
		authzFunc: authzFunc,
	}
}

func (rm *RoleManager) AdminNames() []string {
	out := struct {
		AdminNames []string
	}{}
	if err := rm.ac.Get("RoleManager").UnmarshalTOML(&out); err != nil {
		log.Error("Failed to unmarshal role manager: %+v", err)
		return nil
	}
	return out.AdminNames
}

func (rm *RoleManager) judgeRole(roleCtx *model.RoleContext) {
	admins := sets.NewString(rm.AdminNames()...)
	if admins.Has(roleCtx.Username) {
		roleCtx.Role = model.RoleAdmin
		return
	}
	roleCtx.Role = model.RoleUser
}

func (rm *RoleManager) doAuthZ(ctx context.Context, roleCtx model.RoleContext) error {
	rm.judgeRole(&roleCtx)
	if err := rm.authzFunc(ctx, roleCtx); err != nil {
		return err
	}
	return nil
}

func (s *CommonService) authZByRole(ctx context.Context, roleCtx model.RoleContext) error {
	if roleCtx.Role == model.RoleAdmin {
		return nil
	}
	return s.permittedApp(ctx, roleCtx.Username, roleCtx.Cookie, roleCtx.Node, roleCtx.TargetGateway())
}

func (rm *RoleManager) AuthZ(ctx *bm.Context) {
	roleCtx := model.RoleContext{}
	if err := ctx.Bind(&roleCtx); err != nil {
		return
	}
	if roleCtx.TargetGateway() == "" {
		ctx.JSON(nil, ecode.Error(ecode.RequestErr, "empty `app_name` and `gateway`"))
		ctx.Abort()
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

func (s *CommonService) RoleAuthZ() func(*bm.Context) {
	return s.role.AuthZ
}

func (s *CommonService) AuthByAppKey() bm.HandlerFunc {
	return func(ctx *bm.Context) {
		params := ctx.Request.Form
		appkey := params.Get("appkey")
		var appKeys map[string]string
		if err := s.ac.Get("appKeys").UnmarshalTOML(&appKeys); err != nil {
			log.Error("Failed to unmarshal appKeys: %+v", err)
			ctx.JSON(nil, ecode.ServerErr)
			ctx.Abort()
			return
		}
		caller, ok := appKeys[appkey]
		if !ok {
			ctx.JSON(nil, ecode.Unauthorized)
			ctx.Abort()
			return
		}
		ctx.Set("username", fmt.Sprintf("internal:%s", caller))
	}
}
