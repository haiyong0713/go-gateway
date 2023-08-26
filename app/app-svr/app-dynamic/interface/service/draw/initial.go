package draw

import (
	"context"
	"go-common/library/log"
	model "go-gateway/app/app-svr/app-dynamic/interface/model/draw"

	"go-common/library/sync/errgroup.v2"
)

// AllNoKeyword .
func (s *Service) AllNoKeyword(ctx context.Context, req *model.SearchAllReq) (resp *model.SearchAllReply, err error) {
	var (
		needLbs      = false
		needProcsNum = 2 // 目前依赖3个接口，如有添加须要调整这个值
	)
	resp = &model.SearchAllReply{
		Users:     []*model.UserReply{},
		Topics:    []*model.TopicReply{},
		Locations: []*model.LocationReply{},
		Items:     []*model.ItemReply{},
	}
	eg := errgroup.WithContext(ctx)
	if req.Lat != 0.0 && req.Lng != 0.0 {
		needLbs = true
		needProcsNum++
	}
	eg.GOMAXPROCS(needProcsNum)
	eg.Go(func(ctx context.Context) (err error) {
		users, err := s.allInitialUserPart(ctx, req.Uid)
		if err != nil {
			log.Error("initial all user part error, req:(%v), err(%v)", req, err)
			err = nil // 将err置空，否则errgroup会调ctx的cancel停掉其他任务
			return
		}
		if len(users) > 0 {
			resp.Users = users
		}
		return
	})
	eg.Go(func(ctx context.Context) (err error) {
		topics, err := s.allInitialTopicPart(ctx, req.Uid)
		if err != nil {
			log.Error("initial all topic part error, req:(%v), err(%v)", req, err)
			err = nil // 将err置空，否则errgroup会调ctx的cancel停掉其他任务
			return
		}
		if len(topics) > 0 {
			resp.Topics = topics
		}
		return
	})
	if needLbs {
		eg.Go(func(ctx context.Context) (err error) {
			locations, err := s.allInitialLBSPart(ctx, req.Lat, req.Lng)
			if err != nil {
				log.Error("initial all lbs part error, req:(%v), err(%v)", req, err)
				err = nil // 将err置空，否则errgroup会调ctx的cancel停掉其他任务
				return
			}
			if len(locations) > 0 {
				resp.Locations = locations
			}
			return
		})
	}
	_ = eg.Wait()
	return
}

func (s *Service) UserNoKeyword(ctx context.Context, req *model.SearchUsersReq) (resp *model.SearchUsersReply, err error) {
	resp = new(model.SearchUsersReply)
	resp.Users = []*model.UserReply{}
	followUsers, err := s.dao.GetUserLatestFollowTopK(ctx, req.Uid, 10, "")
	if err != nil {
		return
	}
	if len(followUsers) == 0 {
		return
	}
	for _, user := range followUsers {
		resp.Users = append(resp.Users, &model.UserReply{
			Profile: user.Face,
			Name:    user.Name,
			Uid:     user.Mid,
		})
	}
	return
}

func (s *Service) TopicsNoKeyword(ctx context.Context, req *model.SearchTopicsReq) (resp *model.SearchTopicsReply, err error) {
	resp = new(model.SearchTopicsReply)
	topics, err := s.dao.GetHotTopicTopK(ctx, 10)
	if (err != nil) || (len(topics) == 0) {
		resp.Topics = []*model.TopicReply{}
		return
	}
	for _, topic := range topics {
		resp.Topics = append(resp.Topics, &model.TopicReply{
			TopicId:   topic.TopicId,
			TopicName: topic.TopicName,
		})
	}
	return
}

func (s *Service) LocationsNoKeyword(ctx context.Context, req *model.SearchLocationsReq) (resp *model.SearchLocationsReply, err error) {
	resp = new(model.SearchLocationsReply)
	resp.Locations = []*model.LocationReply{}
	if req.Lat == 0.0 && req.Lng == 0.0 {
		return
	}
	locations, err := s.dao.GetNearbyLocationsTopK(ctx, 10, req.Lat, req.Lng)
	if (err != nil) || (len(locations) == 0) {
		return
	}
	for _, loc := range locations {
		resp.Locations = append(resp.Locations, &model.LocationReply{Poi: loc.Pio})
	}
	return
}

func (s *Service) ItemsNoKeyword(ctx context.Context, req *model.SearchItemsReq) (resp *model.SearchItemsReply, err error) {
	return &model.SearchItemsReply{
		Items: []*model.ItemReply{},
	}, nil
}
