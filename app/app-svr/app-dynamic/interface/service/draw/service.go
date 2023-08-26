package draw

import (
	"context"

	"go-gateway/app/app-svr/app-dynamic/interface/conf"
	dao "go-gateway/app/app-svr/app-dynamic/interface/dao/draw"
	model "go-gateway/app/app-svr/app-dynamic/interface/model/draw"
)

// Service .
type Service struct {
	dao *dao.Dao
}

func New(c *conf.Config) *Service {
	return &Service{
		dao: dao.New(c),
	}
}

func (s *Service) SearchAll(ctx context.Context, req *model.SearchAllReq) (resp *model.SearchAllReply, err error) {
	if req.Keyword == "" {
		return s.AllNoKeyword(ctx, req)
	}
	return s.AllKeyword(ctx, req)
}

func (s *Service) SearchUsers(ctx context.Context, req *model.SearchUsersReq) (resp *model.SearchUsersReply, err error) {
	if req.Keyword == "" {
		return s.UserNoKeyword(ctx, req)
	}
	return s.UsersKeyword(ctx, req)
}

func (s *Service) SearchTopics(ctx context.Context, req *model.SearchTopicsReq) (resp *model.SearchTopicsReply, err error) {
	if req.Keyword == "" {
		return s.TopicsNoKeyword(ctx, req)
	}
	return s.TopicKeyword(ctx, req)
}

func (s *Service) SearchLocations(ctx context.Context, req *model.SearchLocationsReq) (resp *model.SearchLocationsReply, err error) {
	if req.Keyword == "" {
		return s.LocationsNoKeyword(ctx, req)
	}
	return s.LocationsKeyword(ctx, req)
}

func (s *Service) SearchItems(ctx context.Context, req *model.SearchItemsReq) (resp *model.SearchItemsReply, err error) {
	if req.Keyword == "" {
		return s.ItemsNoKeyword(ctx, req)
	}
	return s.ItemsKeyword(ctx, req)
}
