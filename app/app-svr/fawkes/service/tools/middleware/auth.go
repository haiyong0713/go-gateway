package middleware

import (
	"context"
	"fmt"
	"net/url"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"

	authmdl "go-gateway/app/app-svr/fawkes/service/model/auth"
	"go-gateway/app/app-svr/fawkes/service/model/tool"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
	"go-gateway/app/app-svr/fawkes/service/tools/utils"
)

// AuthVerify fawkes权限控制 todo 加个缓存？
func AuthVerify() bm.HandlerFunc {
	return func(c *bm.Context) {
		username := utils.GetUsername(c)
		if isSupervisor(c, username) {
			return
		}
		path := c.Request.URL.Path
		authItems, err := fkDao.SelectAuthItemByUrl(c, path)
		if err != nil {
			log.Errorc(c, "%v", err)
			c.Error = err
			return
		}
		if len(authItems) == 0 {
			return
		}
		item, err := matchAuthItem(authItems, c.Request.Form)
		if err != nil {
			log.Errorc(c, "%v", err)
			c.Error = err
			return
		}
		if !item.IsActive {
			return
		}
		relation, err := fkDao.SelectAuthRelation(c, item.Id)
		if err != nil {
			log.Errorc(c, "%v", err)
			c.Error = err
			return
		}
		roles := make(map[int8]bool)
		for _, v := range relation {
			roles[v.AuthRoleValue] = true
		}
		appKey := c.Request.Form.Get("app_key")
		authRoles, err := fkDao.AuthRolesApply(c, appKey, username)
		if err != nil {
			log.Errorc(c, "%v", err)
			c.Error = err
			return
		}
		if len(authRoles) == 0 {
			// 不通过
			log.Infoc(c, "AppKey:%s username:%s没有查询到角色权限", appKey, username)
			c.JSON(nil, ecode.Error(ecode.Unauthorized, "权限不足"))
			c.Abort()
			return
		}
		// 用户多个角色中任意一个通过校验视为可以访问
		var pass bool
		userRoles := make(map[int8]bool)
		for _, v := range authRoles {
			userRoles[int8(v.Role)] = true
			if _, ok := roles[int8(v.Role)]; ok && v.State == 1 {
				pass = true
			}
		}
		log.Infoc(c, "权限【%s】,可访问的角色是%v,当前用户%s在应用%s中的身份是%v", item.Name, parseRole(roles), username, appKey, parseRole(userRoles))
		if !pass {
			// 不通过
			c.JSON(nil, ecode.Error(ecode.Unauthorized, fmt.Sprintf("%s在%s中的身份是%s,不满足访问条件。可访问的角色是%v", username, appKey, parseRole(userRoles), parseRole(roles))))
			log.Infoc(c, "不满足访问条件")
			c.Abort()
		}
		log.Infoc(c, "满足访问条件")
		c.Next()
	}
}

func matchAuthItem(items []*authmdl.Item, form url.Values) (item *authmdl.Item, err error) {
	// 从多条item中匹配一条规则
	for _, v := range items {
		if len(v.UrlParam) != 0 {
			param, err1 := url.ParseQuery(v.UrlParam)
			if err1 != nil {
				err = err1
				return
			}
			for k := range param {
				if param.Get(k) != form.Get(k) {
					break
				}
			}
			item = v
			return
		}
	}
	if len(items) > 1 {
		err = ecode.Error(ecode.ServerErr, "权限项设置存在错误")
		return
	}
	item = items[0]
	return
}

func parseRole(roleMap map[int8]bool) (roles []string) {
	for k := range roleMap {
		roles = append(roles, tool.GetRole(k))
	}
	return
}

func isSupervisor(ctx context.Context, username string) bool {
	// 超管权限
	if username != "" {
		supervisorRole, err := fkDao.AuthSupervisor(ctx, username)
		if err != nil {
			log.Errorc(ctx, "%v", err)
			return false
		}
		if len(supervisorRole) > 0 {
			return true
		}
	}
	return false
}
