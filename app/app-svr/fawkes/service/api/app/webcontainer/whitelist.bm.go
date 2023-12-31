// Code generated by protoc-gen-bm v0.1, DO NOT EDIT.
// source: go-gateway/app/app-svr/fawkes/service/api/app/webcontainer/whitelist.proto

/*
Package webcontainer is a generated blademaster stub package.
This code was generated with kratos/tool/protobuf/protoc-gen-bm v0.1.

It is generated from these files:

	go-gateway/app/app-svr/fawkes/service/api/app/webcontainer/whitelist.proto
*/
package webcontainer

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

var PathWhiteListAddWhiteList = "/x/admin/fawkes/app/webcontainer/whitelist/add"
var PathWhiteListDelWhiteList = "/x/admin/fawkes/app/webcontainer/whitelist/delete"
var PathWhiteListUpdateWhiteList = "/x/admin/fawkes/app/webcontainer/whitelist/update"
var PathWhiteListGetWhiteList = "/x/admin/fawkes/app/webcontainer/whitelist"
var PathWhiteListWhiteListConfig = "/x/admin/fawkes/app/webcontainer/whitelist/config"
var PathWhiteListDomainStatusSync = "/x/admin/fawkes/app/webcontainer/whitelist/domain/sync"

// WhiteListBMServer is the server API for WhiteList service.
type WhiteListBMServer interface {
	AddWhiteList(ctx context.Context, req *AddWhiteListReq) (resp *google_protobuf1.Empty, err error)

	DelWhiteList(ctx context.Context, req *DelWhiteListReq) (resp *google_protobuf1.Empty, err error)

	UpdateWhiteList(ctx context.Context, req *UpdateWhiteListReq) (resp *google_protobuf1.Empty, err error)

	GetWhiteList(ctx context.Context, req *GetWhiteListReq) (resp *GetWhiteListResp, err error)

	WhiteListConfig(ctx context.Context, req *WhiteListConfigReq) (resp *WhiteListConfigResp, err error)

	DomainStatusSync(ctx context.Context, req *google_protobuf1.Empty) (resp *google_protobuf1.Empty, err error)
}

var WhiteListSvc WhiteListBMServer

func whiteListAddWhiteList(c *bm.Context) {
	p := new(AddWhiteListReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := WhiteListSvc.AddWhiteList(c, p)
	c.JSON(resp, err)
}

func whiteListDelWhiteList(c *bm.Context) {
	p := new(DelWhiteListReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := WhiteListSvc.DelWhiteList(c, p)
	c.JSON(resp, err)
}

func whiteListUpdateWhiteList(c *bm.Context) {
	p := new(UpdateWhiteListReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := WhiteListSvc.UpdateWhiteList(c, p)
	c.JSON(resp, err)
}

func whiteListGetWhiteList(c *bm.Context) {
	p := new(GetWhiteListReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := WhiteListSvc.GetWhiteList(c, p)
	c.JSON(resp, err)
}

func whiteListWhiteListConfig(c *bm.Context) {
	p := new(WhiteListConfigReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := WhiteListSvc.WhiteListConfig(c, p)
	c.JSON(resp, err)
}

func whiteListDomainStatusSync(c *bm.Context) {
	p := new(google_protobuf1.Empty)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := WhiteListSvc.DomainStatusSync(c, p)
	c.JSON(resp, err)
}

// RegisterWhiteListBMServer Register the blademaster route
func RegisterWhiteListBMServer(e *bm.Engine, server WhiteListBMServer) {
	WhiteListSvc = server
	e.POST("/x/admin/fawkes/app/webcontainer/whitelist/add", whiteListAddWhiteList)
	e.POST("/x/admin/fawkes/app/webcontainer/whitelist/delete", whiteListDelWhiteList)
	e.POST("/x/admin/fawkes/app/webcontainer/whitelist/update", whiteListUpdateWhiteList)
	e.POST("/x/admin/fawkes/app/webcontainer/whitelist", whiteListGetWhiteList)
	e.GET("/x/admin/fawkes/app/webcontainer/whitelist/config", whiteListWhiteListConfig)
	e.POST("/x/admin/fawkes/app/webcontainer/whitelist/domain/sync", whiteListDomainStatusSync)
}
