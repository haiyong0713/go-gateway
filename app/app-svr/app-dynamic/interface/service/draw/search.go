package draw

import (
	"context"
	"strconv"

	model "go-gateway/app/app-svr/app-dynamic/interface/model/draw"

	"go-common/library/log"

	"go-common/library/sync/errgroup.v2"
)

func (s *Service) AllKeyword(ctx context.Context, req *model.SearchAllReq) (resp *model.SearchAllReply, err error) {
	var (
		needLbs      = false
		needProcsNum = 3 // 目前搜索部分依赖4个接口，如有添加须要调整这个值
	)
	eg := errgroup.WithContext(ctx)
	resp = &model.SearchAllReply{
		Users:     []*model.UserReply{},
		Topics:    []*model.TopicReply{},
		Locations: []*model.LocationReply{},
		Items:     []*model.ItemReply{},
	}

	if req.Lat != 0.0 && req.Lng != 0.0 {
		needLbs = true
		needProcsNum++
	}
	eg.GOMAXPROCS(needProcsNum)
	eg.Go(func(ctx context.Context) (err error) {
		users, err := s.allSearchUserPart(ctx, req)
		if err != nil {
			log.Error("search all user part error, req:(%v), err(%v)", req, err)
			err = nil // 将err置空，否则errgroup会调ctx的cancel停掉其他任务
			return
		}
		if len(users) > 0 {
			resp.Users = users
		}
		return
	})
	eg.Go(func(ctx context.Context) (err error) {
		topics, err := s.allSearchTopicPart(ctx, req)
		if err != nil {
			log.Error("search all topic part error, req:(%v), err(%v)", req, err)
			err = nil // 将err置空，否则errgroup会调ctx的cancel停掉其他任务
			return
		}
		if len(topics) > 0 {
			resp.Topics = topics
		}
		return
	})
	eg.Go(func(ctx context.Context) (err error) {
		items, err := s.allSearchItemPart(ctx, req)
		if err != nil {
			log.Error("search all items part error, req:(%v), err(%v)", req, err)
			err = nil // 将err置空，否则errgroup会调ctx的cancel停掉其他任务
			return
		}
		if len(items) > 0 {
			resp.Items = items
		}
		return
	})
	if needLbs {
		eg.Go(func(ctx context.Context) (err error) {
			locations, err := s.allSearchLbsPart(ctx, req)
			if err != nil {
				log.Error("search all locations part error, req:(%v), err(%v)", req, err)
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

func (s *Service) UsersKeyword(ctx context.Context, req *model.SearchUsersReq) (resp *model.SearchUsersReply, err error) {
	resp = new(model.SearchUsersReply)
	resp.Users = []*model.UserReply{}
	users, hasMore, err := s.dao.SearchUser(ctx, req.Uid, req.Keyword, int(req.Page), int(req.PageSize))
	if err != nil {
		return
	}
	if len(users) == 0 {
		return
	}
	for _, user := range users {
		if user != nil {
			ur := new(model.UserReply)
			ur.Uid = user.Mid
			ur.Profile = "https:" + user.Face
			ur.Name = user.Name
			resp.Users = append(resp.Users, ur)
		}
	}
	if hasMore {
		resp.HasMore = 1
	}
	return
}

func (s *Service) TopicKeyword(ctx context.Context, req *model.SearchTopicsReq) (resp *model.SearchTopicsReply, err error) {
	resp = new(model.SearchTopicsReply)
	resp.Topics = []*model.TopicReply{}
	topics, hasMore, err := s.dao.SearchTopic(ctx, req.Keyword, int(req.Page), int(req.PageSize))
	if (err != nil) || (len(topics) == 0) {
		resp.Topics = []*model.TopicReply{}
		return
	}
	for _, topic := range topics {
		if topic != nil {
			resp.Topics = append(resp.Topics, &model.TopicReply{
				TopicId:   topic.TopicId,
				TopicName: topic.TopicName,
			})
		}
	}
	if hasMore {
		resp.HasMore = 1
	}
	return
}

func (s *Service) LocationsKeyword(ctx context.Context, req *model.SearchLocationsReq) (resp *model.SearchLocationsReply, err error) {
	resp = new(model.SearchLocationsReply)
	resp.Locations = []*model.LocationReply{}
	if req.Lat == 0.0 && req.Lng == 0.0 {
		return
	}
	locations, hasMore, err := s.dao.SearchLabs(ctx, req.Keyword, req.Lat, req.Lng, int(req.Page), int(req.PageSize))
	if (err != nil) || (len(locations) == 0) {
		return
	}
	for _, loc := range locations {
		if loc != nil {
			resp.Locations = append(resp.Locations, &model.LocationReply{
				Poi: loc.Pio,
			})
		}
	}
	if hasMore {
		resp.HasMore = 1
	}
	return
}

func (s *Service) ItemsKeyword(ctx context.Context, req *model.SearchItemsReq) (resp *model.SearchItemsReply, err error) {
	resp = new(model.SearchItemsReply)
	items, hasMore, err := s.dao.SearchMallItems(ctx, req.Uid, req.Keyword, int(req.Page), int(req.PageSize))
	if (err != nil) || (len(items) == 0) {
		resp.Items = []*model.ItemReply{}
		return
	}
	for _, item := range items {
		if item == nil {
			continue
		}
		var priceEqual = 1
		if item.PriceEqual != nil {
			log.Info("priceEqual != nil")
			priceEqual = *item.PriceEqual
		}
		resp.Items = append(resp.Items, &model.ItemReply{
			Name:           item.Title,
			Url:            item.Url,
			ItemId:         item.ItemId,
			SourceType:     2,
			RequiredNumber: item.RequiredNumber,
			Price:          strconv.FormatFloat(item.Price/100, 'f', 2, 64), // 电商接口返回的price以"分"为单位
			Cover:          item.Cover,
			Brief:          item.Brief,
			PriceEqual:     priceEqual,
		})
	}
	if hasMore {
		resp.HasMore = 1
	}
	return
}
