// Code generated by protoc-gen-bm v0.1, DO NOT EDIT.
// source: go-gateway/app/app-svr/fawkes/service/api/app/auth/auth.proto

/*
Package auth is a generated blademaster stub package.
This code was generated with kratos/tool/protobuf/protoc-gen-bm v0.1.

It is generated from these files:

	go-gateway/app/app-svr/fawkes/service/api/app/auth/auth.proto
*/
package auth

import (
	"context"

	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/binding"
)
import google_protobuf1 "github.com/golang/protobuf/ptypes/empty"

// to suppressed 'imported but not used warning'
var _ *bm.Context
var _ context.Context
var _ binding.StructValidator

var PathAuthAddAuthItemGroup = "/x/admin/fawkes/auth/group/add"
var PathAuthUpdateAuthItemGroup = "/x/admin/fawkes/auth/group/update"
var PathAuthAddAuthItem = "/x/admin/fawkes/auth/item/add"
var PathAuthUpdateAuthItem = "/x/admin/fawkes/auth/item/update"
var PathAuthActiveAuthItem = "/x/admin/fawkes/auth/item/switch"
var PathAuthDeleteAuthItem = "/x/admin/fawkes/auth/item/delete"
var PathAuthGrantRole = "/x/admin/fawkes/auth/grant"
var PathAuthListAuth = "/x/admin/fawkes/auth/list"

// AuthBMServer is the server API for Auth service.
type AuthBMServer interface {
	// 新增权限组
	AddAuthItemGroup(ctx context.Context, req *AddAuthItemGroupReq) (resp *google_protobuf1.Empty, err error)

	// 更新权限组
	UpdateAuthItemGroup(ctx context.Context, req *UpdateAuthItemGroupReq) (resp *google_protobuf1.Empty, err error)

	// 新增权限点
	AddAuthItem(ctx context.Context, req *AddAuthItemReq) (resp *google_protobuf1.Empty, err error)

	// 更新权限点
	UpdateAuthItem(ctx context.Context, req *UpdateAuthItemReq) (resp *google_protobuf1.Empty, err error)

	// 启用权限点
	ActiveAuthItem(ctx context.Context, req *ActiveAuthItemReq) (resp *google_protobuf1.Empty, err error)

	// 删除权限点
	DeleteAuthItem(ctx context.Context, req *DeleteAuthItemReq) (resp *google_protobuf1.Empty, err error)

	// 给角色授权
	GrantRole(ctx context.Context, req *GrantRoleReq) (resp *google_protobuf1.Empty, err error)

	// 拉取所有权限点
	ListAuth(ctx context.Context, req *ListAuthReq) (resp *ListAuthResp, err error)
}

var AuthSvc AuthBMServer

func authAddAuthItemGroup(c *bm.Context) {
	p := new(AddAuthItemGroupReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := AuthSvc.AddAuthItemGroup(c, p)
	c.JSON(resp, err)
}

func authUpdateAuthItemGroup(c *bm.Context) {
	p := new(UpdateAuthItemGroupReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := AuthSvc.UpdateAuthItemGroup(c, p)
	c.JSON(resp, err)
}

func authAddAuthItem(c *bm.Context) {
	p := new(AddAuthItemReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := AuthSvc.AddAuthItem(c, p)
	c.JSON(resp, err)
}

func authUpdateAuthItem(c *bm.Context) {
	p := new(UpdateAuthItemReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := AuthSvc.UpdateAuthItem(c, p)
	c.JSON(resp, err)
}

func authActiveAuthItem(c *bm.Context) {
	p := new(ActiveAuthItemReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := AuthSvc.ActiveAuthItem(c, p)
	c.JSON(resp, err)
}

func authDeleteAuthItem(c *bm.Context) {
	p := new(DeleteAuthItemReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := AuthSvc.DeleteAuthItem(c, p)
	c.JSON(resp, err)
}

func authGrantRole(c *bm.Context) {
	p := new(GrantRoleReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := AuthSvc.GrantRole(c, p)
	c.JSON(resp, err)
}

func authListAuth(c *bm.Context) {
	p := new(ListAuthReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := AuthSvc.ListAuth(c, p)
	c.JSON(resp, err)
}

// RegisterAuthBMServer Register the blademaster route
func RegisterAuthBMServer(e *bm.Engine, server AuthBMServer) {
	AuthSvc = server
	e.POST("/x/admin/fawkes/auth/group/add", authAddAuthItemGroup)
	e.POST("/x/admin/fawkes/auth/group/update", authUpdateAuthItemGroup)
	e.POST("/x/admin/fawkes/auth/item/add", authAddAuthItem)
	e.POST("/x/admin/fawkes/auth/item/update", authUpdateAuthItem)
	e.POST("/x/admin/fawkes/auth/item/switch", authActiveAuthItem)
	e.POST("/x/admin/fawkes/auth/item/delete", authDeleteAuthItem)
	e.POST("/x/admin/fawkes/auth/grant", authGrantRole)
	e.GET("/x/admin/fawkes/auth/list", authListAuth)
}