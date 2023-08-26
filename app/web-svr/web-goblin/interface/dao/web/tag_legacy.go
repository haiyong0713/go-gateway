package web

import (
	"context"

	model "go-gateway/app/web-svr/web-goblin/interface/model/web"

	"go-common/library/net/rpc"
)

const (
	_infoByID         = "RPC.InfoByID"
	_infoByIDs        = "RPC.InfoByIDs"
	_resTags          = "RPC.ResTags"
	_channelResources = "RPC.ChannelResources"
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

// InfoByID .
func (s *TagRPCService) InfoByID(c context.Context, arg *model.ArgID) (res *model.Tag, err error) {
	res = new(model.Tag)
	err = s.client.Call(c, _infoByID, arg, res)
	return
}

// InfoByIDs .
func (s *TagRPCService) InfoByIDs(c context.Context, arg *model.ArgIDs) (res []*model.Tag, err error) {
	err = s.client.Call(c, _infoByIDs, arg, &res)
	return
}

// ResTags .
func (s *TagRPCService) ResTags(c context.Context, arg *model.ArgResTags) (res map[int64][]*model.Tag, err error) {
	err = s.client.Call(c, _resTags, arg, &res)
	return
}

// ChannelResources .
func (s *TagRPCService) ChannelResources(c context.Context, arg *model.ArgChannelResource) (res *model.ChannelResource, err error) {
	err = s.client.Call(c, _channelResources, arg, &res)
	return
}
