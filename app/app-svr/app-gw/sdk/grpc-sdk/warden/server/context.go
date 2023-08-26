package server

import (
	"context"

	"github.com/gogo/protobuf/proto"
	"google.golang.org/grpc"
)

// Context is
type Context struct {
	context.Context

	srv           interface{}
	serverStream  grpc.ServerStream
	serviceMethod string

	serviceMeta *ServiceMeta
	req         dummyMessage
}

func (ctx *Context) Req() proto.Message {
	return &ctx.req
}
