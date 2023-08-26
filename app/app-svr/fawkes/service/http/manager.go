package http

import (
	"strconv"
	"strings"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"

	mngmdl "go-gateway/app/app-svr/fawkes/service/model/manager"
)

func treeAuth(c *bm.Context) {
	var (
		params           = c.Request.Form
		res              = map[string]interface{}{}
		appKey, userName string
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(s.MngSvr.TreeAuth(c, appKey, userName))
}

func treeAuths(c *bm.Context) {
	var (
		params   = c.Request.Form
		res      = map[string]interface{}{}
		appKeys  []string
		userName string
	)
	if appKeys = strings.Split(params.Get("app_keys"), ","); len(appKeys) == 0 {
		res["message"] = "app_keys异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(s.MngSvr.TreeAuths(c, appKeys, userName))
}

func treeList(c *bm.Context) {
	var (
		res                 = map[string]interface{}{}
		sessionID, userName string
		err                 error
	)
	dsbck, err := c.Request.Cookie("_AJSESSIONID")
	if err != nil {
		res["message"] = "sessionID异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if sessionID = dsbck.Value; sessionID == "" {
		res["message"] = "sessionID值非法"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(s.MngSvr.TreeList(c, sessionID))
}

func authUserList(c *bm.Context) {
	var (
		params                            = c.Request.Form
		res                               = map[string]interface{}{}
		appKey, role, userName, filterKey string
		pn, ps                            int
		err                               error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	if pn, err = strconv.Atoi(params.Get("pn")); err != nil || pn <= 0 {
		res["message"] = "pn异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if ps, err = strconv.Atoi(params.Get("ps")); err != nil {
		res["message"] = "ps异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if ps < 1 {
		ps = 50
	}
	role = params.Get("role")
	filterKey = params.Get("filter_key")
	c.JSON(s.MngSvr.AuthUserList(c, appKey, role, userName, filterKey, pn, ps))
}

func authUserListByRole(c *bm.Context) {
	var (
		params = c.Request.Form
		res    = map[string]interface{}{}
		appKey string
		role   int
		err    error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if role, err = strconv.Atoi(params.Get("role")); err != nil {
		res["message"] = "role异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.MngSvr.AuthUserListByRole(c, appKey, role))
}

func authUserSet(c *bm.Context) {
	var (
		params                  = c.Request.Form
		res                     = map[string]interface{}{}
		appKey, userName, uname string
		role                    int
		err                     error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	if uname = params.Get("user_name"); uname == "" {
		res["message"] = "user_name异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if role, err = strconv.Atoi(params.Get("role")); err != nil || role <= 0 {
		res["message"] = "role异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, s.MngSvr.AuthUserSet(c, appKey, userName, uname, role))
}

func authUserDel(c *bm.Context) {
	var (
		params   = c.Request.Form
		res      = map[string]interface{}{}
		id       int64
		userName string
		err      error
	)
	if id, err = strconv.ParseInt(params.Get("id"), 10, 64); err != nil || id == 0 {
		res["message"] = "id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, s.MngSvr.AuthUserDel(c, id, userName))
}

func authUser(c *bm.Context) {
	var (
		params           = c.Request.Form
		res              = map[string]interface{}{}
		appKey, userName string
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(s.MngSvr.AuthUser(c, appKey, userName))
}

func authRole(c *bm.Context) {
	c.JSON(s.MngSvr.AuthRole(c))
}

func authSupervisor(c *bm.Context) {
	var (
		userName string
	)
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(s.MngSvr.AuthSupervisor(c, userName))
}

func authSupervisorRole(c *bm.Context) {
	var (
		userName string
	)
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(s.MngSvr.AuthSupervisorRole(c, userName))
}

func authRoleApply(c *bm.Context) {
	var (
		params           = c.Request.Form
		res              = map[string]interface{}{}
		appKey, userName string
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(s.MngSvr.AuthRoleApply(c, appKey, userName))
}

func authRoleApplyList(c *bm.Context) {
	var (
		params        = c.Request.Form
		res           = map[string]interface{}{}
		appKey        string
		pn, ps, state int
		err           error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if stateStr := params.Get("state"); stateStr != "" {
		if state, err = strconv.Atoi(stateStr); err != nil {
			res["message"] = "state异常"
			c.JSONMap(res, ecode.RequestErr)
			return
		}
	} else {
		state = mngmdl.AuthRoleApplyEmptyState
	}
	if pn, err = strconv.Atoi(params.Get("pn")); err != nil || pn <= 0 {
		res["message"] = "pn异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if ps, err = strconv.Atoi(params.Get("ps")); err != nil {
		res["message"] = "ps异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if ps < 1 || ps > 50 {
		ps = 50
	}
	c.JSON(s.MngSvr.AuthRoleApplyList(c, appKey, state, pn, ps))
}

func authRoleApplyAdd(c *bm.Context) {
	var (
		params                     = c.Request.Form
		res                        = map[string]interface{}{}
		appKey, userName, operator string
		role                       int
		err                        error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if operator = params.Get("operator"); operator == "" {
		res["message"] = "operator异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	if role, err = strconv.Atoi(params.Get("role")); err != nil || role <= 0 {
		res["message"] = "role异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, s.MngSvr.AuthRoleApplyAdd(c, appKey, userName, operator, role))
}

func authRoleApplyPass(c *bm.Context) {
	var (
		params           = c.Request.Form
		res              = map[string]interface{}{}
		appKey, userName string
		id               int
		err              error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	if id, err = strconv.Atoi(params.Get("id")); err != nil {
		res["message"] = "id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, s.MngSvr.AuthRoleApplyPass(c, appKey, userName, id))
}

func authRoleApplyRefuse(c *bm.Context) {
	var (
		params           = c.Request.Form
		res              = map[string]interface{}{}
		appKey, userName string
		id               int
		err              error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	if id, err = strconv.Atoi(params.Get("id")); err != nil {
		res["message"] = "id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, s.MngSvr.AuthRoleApplyRefuse(c, appKey, userName, id))
}

func authRoleApplyIgnore(c *bm.Context) {
	var (
		params           = c.Request.Form
		res              = map[string]interface{}{}
		appKey, userName string
		id               int
		err              error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	if id, err = strconv.Atoi(params.Get("id")); err != nil {
		res["message"] = "id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, s.MngSvr.AuthRoleApplyIgnore(c, appKey, userName, id))
}

func logList(c *bm.Context) {
	var (
		params                                                        = c.Request.Form
		res                                                           = map[string]interface{}{}
		appKey, env, model, operation, target, operator, stime, etime string
		pn, ps                                                        int
		err                                                           error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "appkey异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	env = params.Get("env")
	model = params.Get("model")
	operation = params.Get("operation")
	target = params.Get("target")
	operator = params.Get("operator")
	stime = params.Get("stime")
	etime = params.Get("etime")
	if pn, err = strconv.Atoi(params.Get("pn")); err != nil || pn <= 0 {
		res["message"] = "pn异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if ps, err = strconv.Atoi(params.Get("ps")); err != nil {
		res["message"] = "ps异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.MngSvr.LogList(c, appKey, env, model, operation, target, operator, stime, etime, pn, ps))
}

func eventApplyAdd(c *bm.Context) {
	var (
		params                     = c.Request.Form
		res                        = map[string]interface{}{}
		appKey, userName, operator string
		event, targetID            int
		err                        error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if operator = params.Get("operator"); operator == "" {
		res["message"] = "operator异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if targetID, err = strconv.Atoi(params.Get("target_id")); err != nil {
		res["message"] = "target_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if event, err = strconv.Atoi(params.Get("event")); err != nil {
		res["message"] = "event异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, s.MngSvr.EventApplyAdd(c, appKey, userName, operator, event, targetID))
}

func eventApplyRecall(c *bm.Context) {
	var (
		params           = c.Request.Form
		res              = map[string]interface{}{}
		appKey, userName string
		event, targetID  int
		err              error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if event, err = strconv.Atoi(params.Get("event")); err != nil {
		res["message"] = "event异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if targetID, err = strconv.Atoi(params.Get("target_id")); err != nil {
		res["message"] = "target_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, s.MngSvr.EventApplyRecall(c, appKey, userName, event, targetID))
}

func bfsRefreshCDN(c *bm.Context) {
	var (
		params = c.Request.Form
		res    = map[string]interface{}{}
		urls   string
	)
	if urls = params.Get("urls"); urls == "" {
		res["message"] = "urls异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, s.MngSvr.BfsRefreshCDN(c, urls))
}

func authMessagePush(c *bm.Context) {
	var (
		params                                 = c.Request.Form
		res                                    = map[string]interface{}{}
		appKey, userName, title, content, link string
		msgType                                int
		err                                    error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if content = params.Get("content"); content == "" {
		res["message"] = "content异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if msgType, err = strconv.Atoi(params.Get("msg_type")); err != nil {
		res["message"] = "msg_type异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	title = params.Get("title")
	link = params.Get("link")
	c.JSON(nil, s.MngSvr.AuthMessagePush(c, appKey, userName, title, content, link, msgType))
}

func authNickNameSet(c *bm.Context) {
	var (
		params             = c.Request.Form
		res                = map[string]interface{}{}
		userName, nickName string
	)
	if userName = params.Get("username"); userName == "" {
		res["message"] = "username异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if nickName = params.Get("nickname"); nickName == "" {
		res["message"] = "nickname异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, s.MngSvr.AuthNickNameSet(c, userName, nickName))
}

func userNameList(c *bm.Context) {
	var (
		params = c.Request.Form
		name   string
		pn, ps int
		err    error
	)
	name = params.Get("name")
	if pn, err = strconv.Atoi(params.Get("pn")); err != nil || pn <= 0 {
		pn = 1
	}
	if ps, err = strconv.Atoi(params.Get("ps")); err != nil {
		ps = 100
	}
	if ps < 1 || ps > 100 {
		ps = 100
	}
	c.JSON(s.MngSvr.UserNameList(c, name, ps, pn))
}

func userNameSet(c *bm.Context) {
	var (
		params             = c.Request.Form
		res                = map[string]interface{}{}
		userName, nickName string
	)
	if userName = params.Get("username"); userName == "" {
		res["message"] = "username异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if nickName = params.Get("nickname"); nickName == "" {
		res["message"] = "nickname异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, s.MngSvr.UserNameSet(c, userName, nickName))
}

func authAdminApply(c *bm.Context) {
	var (
		params = c.Request.Form
		appKey string
		res    = map[string]interface{}{}
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, s.MngSvr.AuthAdminApply(c, appKey, userName))
}
