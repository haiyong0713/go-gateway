// Code generated by protoc-gen-bm v0.1, DO NOT EDIT.
// source: api.proto

/*
Package api is a generated blademaster stub package.
This code was generated with kratos/tool/protobuf/protoc-gen-bm v0.1.

It is generated from these files:

	api.proto
*/
package api

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

var PathManagementJobPing = "/appgw.management.job.v1.ManagementJob/Ping"
var PathManagementJobTaskDo = "/appgw.management.job.v1.ManagementJob/TaskDo"
var PathManagementJobRawConfig = "/appgw.management.job.v1.ManagementJob/RawConfig"

// ManagementJobBMServer is the server API for ManagementJob service.
type ManagementJobBMServer interface {
	Ping(ctx context.Context, req *google_protobuf1.Empty) (resp *google_protobuf1.Empty, err error)

	TaskDo(ctx context.Context, req *TaskDoReq) (resp *TaskDoReply, err error)

	RawConfig(ctx context.Context, req *RawConfigReq) (resp *RawConfigReply, err error)
}

var ManagementJobSvc ManagementJobBMServer

func managementJobPing(c *bm.Context) {
	p := new(google_protobuf1.Empty)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := ManagementJobSvc.Ping(c, p)
	c.JSON(resp, err)
}

func managementJobTaskDo(c *bm.Context) {
	p := new(TaskDoReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := ManagementJobSvc.TaskDo(c, p)
	c.JSON(resp, err)
}

func managementJobRawConfig(c *bm.Context) {
	p := new(RawConfigReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := ManagementJobSvc.RawConfig(c, p)
	c.JSON(resp, err)
}

// RegisterManagementJobBMServer Register the blademaster route
func RegisterManagementJobBMServer(e *bm.Engine, server ManagementJobBMServer) {
	ManagementJobSvc = server
	e.GET("/appgw.management.job.v1.ManagementJob/Ping", managementJobPing)
	e.GET("/appgw.management.job.v1.ManagementJob/TaskDo", managementJobTaskDo)
	e.GET("/appgw.management.job.v1.ManagementJob/RawConfig", managementJobRawConfig)
}
