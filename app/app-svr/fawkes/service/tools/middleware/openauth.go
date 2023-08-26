package middleware

import (
	"context"
	"fmt"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"

	openmdl "go-gateway/app/app-svr/fawkes/service/model/open"
	toolmdl "go-gateway/app/app-svr/fawkes/service/model/tool"
	"go-gateway/app/app-svr/fawkes/service/service/casbin"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

const (
	_tokenName = "fawkes-token"
	_userName  = "fawkes-user"
)

// OpenAuth openapi的身份认证
func OpenAuth() bm.HandlerFunc {
	return func(c *bm.Context) {
		var (
			err     error
			project *openmdl.Project
		)
		openToken := c.Request.Header.Get(_tokenName)
		project, err = fkDao.SelectOpenToken(c, openToken)
		if err != nil {
			log.Errorc(c, "%v", err)
			c.JSON(nil, ecode.Error(ecode.ServerErr, "check fawkes-token occur error"))
			c.Abort()
		}
		if project == nil || !project.IsActive {
			c.JSON(nil, ecode.Error(ecode.Unauthorized, "fawkes-token check failed, make sure you have the permission."))
			c.Abort()
			return
		}
		user := fmt.Sprintf("openapi_project_%s", project.ProjectName)
		if len(c.Request.Header.Get(_userName)) != 0 {
			user = c.Request.Header.Get(_userName)
		}
		c.Set("username", user)
		c.Context = metadata.MergeContext(c.Context, metadata.MD{metadata.Caller: user})
		v := &toolmdl.ContextValues{}
		v.FromOpenAPI = true
		v.OpenAPIUser = user
		c.Context = context.WithValue(c.Context, toolmdl.ContentKey, v)
		c.Next()
	}
}

// AccessControl 权限管理
func AccessControl() bm.HandlerFunc {
	return func(c *bm.Context) {
		openToken := c.Request.Header.Get(_tokenName)
		path := c.Request.URL.Path
		appKey := c.Request.Form.Get("app_key")
		if len(appKey) == 0 {
			appKey = "*"
		}
		e := casbin.GetInstance()
		ok, err := e.Enforce(openToken, path, appKey)
		if err != nil {
			log.Errorc(c, "enforce error: %v", err)
			c.JSON(nil, ecode.Error(ecode.Unauthorized, fmt.Sprintf("You are not authorized to visit [%s]-[%s]", path, appKey)))
			c.Abort()
			return
		}
		if !ok {
			c.JSON(nil, ecode.Error(ecode.Unauthorized, fmt.Sprintf("You are not authorized to visit [%s]-[%s]", path, appKey)))
			c.Abort()
		}
		c.Next()
	}
}
