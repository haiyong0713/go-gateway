package dao

import (
	"context"

	"go-gateway/app/web-svr/player/interface/model"

	"go-common/library/net/rpc"
)

const (
	_arcTags = "RPC.ArcTags"
)

const (
	_appid = "main.community.tag"
)

// Service .
type TagRPCService struct {
	client *rpc.Client2
}

// New2 .
func NewTagRPC(c *rpc.ClientConfig) (s *TagRPCService) {
	s = &TagRPCService{}
	s.client = rpc.NewDiscoveryCli(_appid, c)
	return
}

// ArcTags .
func (s *TagRPCService) ArcTags(c context.Context, arg *model.ArgAid) (res []*model.Tag, err error) {
	err = s.client.Call(c, _arcTags, arg, &res)
	return
}
