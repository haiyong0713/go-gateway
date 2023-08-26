package dao

import (
	"context"
	"go-common/library/net/rpc"
	"go-gateway/app/web-svr/web/interface/model"
)

const (
	_tagTop = "RPC.TagTop"
)

const (
	_appid = "main.community.tag"
)

type TagRPCService struct {
	client *rpc.Client2
}

func NewTagRPC(c *rpc.ClientConfig) (s *TagRPCService) {
	s = &TagRPCService{}
	s.client = rpc.NewDiscoveryCli(_appid, c)
	return
}

func (s *TagRPCService) TagTop(c context.Context, arg *model.ReqTagTop) (res *model.TagTop, err error) {
	err = s.client.Call(c, _tagTop, arg, &res)
	return
}
