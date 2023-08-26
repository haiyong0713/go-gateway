package http

import (
	bm "go-common/library/net/http/blademaster"

	"go-gateway/app/api-gateway/api-manager/internal/model"
)

func groupAdd(c *bm.Context) {
	var req = new(model.ContralGroupSaveReq)
	if err := c.Bind(req); err != nil {
		return
	}
	var (
		uid      int64
		username string
	)
	if uidInter, ok := c.Get("uid"); ok {
		uid = uidInter.(int64)
	}
	if usernameCtx, ok := c.Get("username"); ok {
		username = usernameCtx.(string)
	}
	errMsg, err := svc.GroupAdd(c, req, username, uid)
	if err != nil && errMsg != "" {
		c.JSONMap(map[string]interface{}{"message": errMsg}, err)
		return
	}
	c.JSON(nil, err)
}

func groupEdit(c *bm.Context) {
	var req = new(model.ContralGroupSaveReq)
	if err := c.Bind(req); err != nil {
		return
	}
	var (
		uid      int64
		username string
	)
	if uidInter, ok := c.Get("uid"); ok {
		uid = uidInter.(int64)
	}
	if usernameCtx, ok := c.Get("username"); ok {
		username = usernameCtx.(string)
	}
	errMsg, err := svc.GroupEdit(c, req, username, uid)
	if err != nil && errMsg != "" {
		c.JSONMap(map[string]interface{}{"message": errMsg}, err)
		return
	}
	c.JSON(nil, err)
}

func groupList(c *bm.Context) {
	var req = new(model.ContralGroupListReq)
	if err := c.Bind(req); err != nil {
		return
	}
	c.JSON(svc.GroupList(c, req))
}

func groupFollowAction(c *bm.Context) {
	var req = new(model.ContralGroupFollowActionPeq)
	if err := c.Bind(req); err != nil {
		return
	}
	var (
		uid      int64
		username string
	)
	if uidInter, ok := c.Get("uid"); ok {
		uid = uidInter.(int64)
	}
	if usernameCtx, ok := c.Get("username"); ok {
		username = usernameCtx.(string)
	}
	errMsg, err := svc.GroupFollowAction(c, req, username, uid)
	if err != nil && errMsg != "" {
		c.JSONMap(map[string]interface{}{"message": errMsg}, err)
		return
	}
	c.JSON(nil, err)
}

func groupFollowList(c *bm.Context) {
	var (
		uid      int64
		username string
	)
	if uidInter, ok := c.Get("uid"); ok {
		uid = uidInter.(int64)
	}
	if usernameCtx, ok := c.Get("username"); ok {
		username = usernameCtx.(string)
	}
	c.JSON(svc.GroupFollowList(c, username, uid))
}

func apiAdd(c *bm.Context) {
	var req = new(model.ContralApiAddReq)
	if err := c.Bind(req); err != nil {
		return
	}
	var (
		uid      int64
		username string
	)
	if uidInter, ok := c.Get("uid"); ok {
		uid = uidInter.(int64)
	}
	if usernameCtx, ok := c.Get("username"); ok {
		username = usernameCtx.(string)
	}
	errMsg, err := svc.ApiAdd(c, req, username, uid)
	if err != nil && errMsg != "" {
		c.JSONMap(map[string]interface{}{"message": errMsg}, err)
		return
	}
	c.JSON(nil, err)
}

func apiEdit(c *bm.Context) {
	var req = new(model.ContralApiEditReq)
	if err := c.Bind(req); err != nil {
		return
	}
	var (
		uid      int64
		username string
	)
	if uidInter, ok := c.Get("uid"); ok {
		uid = uidInter.(int64)
	}
	if usernameCtx, ok := c.Get("username"); ok {
		username = usernameCtx.(string)
	}
	errMsg, err := svc.ApiEdit(c, req, username, uid)
	if err != nil && errMsg != "" {
		c.JSONMap(map[string]interface{}{"message": errMsg}, err)
		return
	}
	c.JSON(nil, err)
}

func apiList(c *bm.Context) {
	var req = new(model.ContralApiListReq)
	if err := c.Bind(req); err != nil {
		return
	}
	c.JSON(svc.ApiList(c, req))
}

func apiConfigAdd(c *bm.Context) {
	var req = new(model.ContralApiConfigAddReq)
	if err := c.Bind(req); err != nil {
		return
	}
	var (
		uid      int64
		username string
	)
	if uidInter, ok := c.Get("uid"); ok {
		uid = uidInter.(int64)
	}
	if usernameCtx, ok := c.Get("username"); ok {
		username = usernameCtx.(string)
	}
	errMsg, err := svc.ApiConfigAdd(c, req, username, uid)
	if err != nil && errMsg != "" {
		c.JSONMap(map[string]interface{}{"message": errMsg}, err)
		return
	}
	c.JSON(nil, err)
}

func apiConfigRollback(c *bm.Context) {
	var req = new(model.ContralApiConfigRollbackReq)
	if err := c.Bind(req); err != nil {
		return
	}
	var (
		uid      int64
		username string
	)
	if uidInter, ok := c.Get("uid"); ok {
		uid = uidInter.(int64)
	}
	if usernameCtx, ok := c.Get("username"); ok {
		username = usernameCtx.(string)
	}
	errMsg, err := svc.ApiConfigRollback(c, req, username, uid)
	if err != nil && errMsg != "" {
		c.JSONMap(map[string]interface{}{"message": errMsg}, err)
		return
	}
	c.JSON(nil, err)
}

func apiConfigList(c *bm.Context) {
	var req = new(model.ContralApiConfigListReq)
	if err := c.Bind(req); err != nil {
		return
	}
	c.JSON(svc.ApiConfigList(c, req))
}

func apiPublishCallback(c *bm.Context) {
	var req = new(model.ContralapiPublishCallbackReq)
	if err := c.Bind(req); err != nil {
		return
	}
	var (
		uid      int64
		username string
	)
	if uidInter, ok := c.Get("uid"); ok {
		uid = uidInter.(int64)
	}
	if usernameCtx, ok := c.Get("username"); ok {
		username = usernameCtx.(string)
	}
	errMsg, err := svc.ApiPublishCallback(c, req, username, uid)
	if err != nil && errMsg != "" {
		c.JSONMap(map[string]interface{}{"message": errMsg}, err)
		return
	}
	c.JSON(nil, err)
}

func apiPublishList(c *bm.Context) {
	var req = new(model.ContralApiPublishListReq)
	if err := c.Bind(req); err != nil {
		return
	}
	c.JSON(svc.ApiPublishList(c, req))
}
