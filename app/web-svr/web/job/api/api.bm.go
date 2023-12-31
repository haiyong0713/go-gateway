// Code generated by protoc-gen-bm v0.1, DO NOT EDIT.
// source: api.proto

/*
Package api is a generated blademaster stub package.
This code was generated with kratos/tool/protobuf/protoc-gen-bm v0.1.

package 命名使用 {appid}.{version} 的方式, version 形如 v1, v2 ..

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

var PathWebJobPing = "/web.job.v1.WebJob/Ping"

// WebJobBMServer is the server API for WebJob service.
type WebJobBMServer interface {
	Ping(ctx context.Context, req *google_protobuf1.Empty) (resp *google_protobuf1.Empty, err error)
}

var WebJobSvc WebJobBMServer

func webJobPing(c *bm.Context) {
	p := new(google_protobuf1.Empty)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := WebJobSvc.Ping(c, p)
	c.JSON(resp, err)
}

// RegisterWebJobBMServer Register the blademaster route
func RegisterWebJobBMServer(e *bm.Engine, server WebJobBMServer) {
	WebJobSvc = server
	e.GET("/web.job.v1.WebJob/Ping", webJobPing)
}
