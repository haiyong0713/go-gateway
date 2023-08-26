// Code generated by protoc-gen-bm v0.1, DO NOT EDIT.
// source: go-gateway/app/app-svr/app-view/interface/api/view/view.proto

/*
Package api is a generated blademaster stub package.
This code was generated with kratos/tool/protobuf/protoc-gen-bm v0.1.

It is generated from these files:

	go-gateway/app/app-svr/app-view/interface/api/view/view.proto
*/
package api

import (
	"context"

	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/binding"
)

// to suppressed 'imported but not used warning'
var _ *bm.Context
var _ context.Context
var _ binding.StructValidator

var PathViewView = "/bilibili.app.view.v1.View/View"
var PathViewViewProgress = "/bilibili.app.view.v1.View/ViewProgress"
var PathViewClickPlayerCard = "/bilibili.app.view.v1.View/ClickPlayerCard"
var PathViewShortFormVideoDownload = "/bilibili.app.view.v1.View/ShortFormVideoDownload"
var PathViewClickActivitySeason = "/bilibili.app.view.v1.View/ClickActivitySeason"
var PathViewSeason = "/bilibili.app.view.v1.View/Season"
var PathViewExposePlayerCard = "/bilibili.app.view.v1.View/ExposePlayerCard"
var PathViewAddContract = "/bilibili.app.view.v1.View/AddContract"
var PathViewFeedView = "/bilibili.app.view.v1.View/FeedView"

// ViewBMServer is the server API for View service.
// View 详情页相关接口
type ViewBMServer interface {
	// 获取详情页数据
	View(ctx context.Context, req *ViewReq) (resp *ViewReply, err error)

	// 视频播放进度相关展示
	ViewProgress(ctx context.Context, req *ViewProgressReq) (resp *ViewProgressReply, err error)

	// 点击播放器卡片事件
	ClickPlayerCard(ctx context.Context, req *ClickPlayerCardReq) (resp *NoReply, err error)

	// 短视频下载
	ShortFormVideoDownload(ctx context.Context, req *ShortFormVideoDownloadReq) (resp *ShortFormVideoDownloadReply, err error)

	// 点击大型活动页预约
	ClickActivitySeason(ctx context.Context, req *ClickActivitySeasonReq) (resp *NoReply, err error)

	// 合集详情页
	Season(ctx context.Context, req *SeasonReq) (resp *SeasonReply, err error)

	// 播放器卡片曝光
	ExposePlayerCard(ctx context.Context, req *ExposePlayerCardReq) (resp *NoReply, err error)

	// 点击签订契约
	AddContract(ctx context.Context, req *AddContractReq) (resp *NoReply, err error)

	// 推荐流模式
	FeedView(ctx context.Context, req *FeedViewReq) (resp *FeedViewReply, err error)
}

var ViewSvc ViewBMServer

func viewView(c *bm.Context) {
	p := new(ViewReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := ViewSvc.View(c, p)
	c.JSON(resp, err)
}

func viewViewProgress(c *bm.Context) {
	p := new(ViewProgressReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := ViewSvc.ViewProgress(c, p)
	c.JSON(resp, err)
}

func viewClickPlayerCard(c *bm.Context) {
	p := new(ClickPlayerCardReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := ViewSvc.ClickPlayerCard(c, p)
	c.JSON(resp, err)
}

func viewShortFormVideoDownload(c *bm.Context) {
	p := new(ShortFormVideoDownloadReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := ViewSvc.ShortFormVideoDownload(c, p)
	c.JSON(resp, err)
}

func viewClickActivitySeason(c *bm.Context) {
	p := new(ClickActivitySeasonReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := ViewSvc.ClickActivitySeason(c, p)
	c.JSON(resp, err)
}

func viewSeason(c *bm.Context) {
	p := new(SeasonReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := ViewSvc.Season(c, p)
	c.JSON(resp, err)
}

func viewExposePlayerCard(c *bm.Context) {
	p := new(ExposePlayerCardReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := ViewSvc.ExposePlayerCard(c, p)
	c.JSON(resp, err)
}

func viewAddContract(c *bm.Context) {
	p := new(AddContractReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := ViewSvc.AddContract(c, p)
	c.JSON(resp, err)
}

func viewFeedView(c *bm.Context) {
	p := new(FeedViewReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := ViewSvc.FeedView(c, p)
	c.JSON(resp, err)
}

// RegisterViewBMServer Register the blademaster route
func RegisterViewBMServer(e *bm.Engine, server ViewBMServer) {
	ViewSvc = server
	e.GET("/bilibili.app.view.v1.View/View", viewView)
	e.GET("/bilibili.app.view.v1.View/ViewProgress", viewViewProgress)
	e.GET("/bilibili.app.view.v1.View/ClickPlayerCard", viewClickPlayerCard)
	e.GET("/bilibili.app.view.v1.View/ShortFormVideoDownload", viewShortFormVideoDownload)
	e.GET("/bilibili.app.view.v1.View/ClickActivitySeason", viewClickActivitySeason)
	e.GET("/bilibili.app.view.v1.View/Season", viewSeason)
	e.GET("/bilibili.app.view.v1.View/ExposePlayerCard", viewExposePlayerCard)
	e.GET("/bilibili.app.view.v1.View/AddContract", viewAddContract)
	e.GET("/bilibili.app.view.v1.View/FeedView", viewFeedView)
}