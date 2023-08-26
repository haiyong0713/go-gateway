package draw

import (
	"context"
	"strconv"

	"go-common/library/log"
	model "go-gateway/app/app-svr/app-dynamic/interface/model/draw"
)

func (s *Service) allSearchUserPart(ctx context.Context, req *model.SearchAllReq) (users []*model.UserReply, err error) {
	users = make([]*model.UserReply, 0)
	searchUsers, _, err := s.dao.SearchUser(ctx, req.Uid, req.Keyword, int(req.Page), int(req.PageSize))
	if err != nil {
		return
	}
	// 最多取2个
	cutLen := model.MinInt(2, len(searchUsers))
	searchUsers = searchUsers[:cutLen]
	for _, user := range searchUsers {
		if user != nil {
			users = append(users, &model.UserReply{
				Profile: "http:" + user.Face,
				Name:    user.Name,
				Uid:     user.Mid,
			})
		}
	}
	return
}

func (s *Service) allSearchTopicPart(ctx context.Context, req *model.SearchAllReq) (topics []*model.TopicReply, err error) {
	topics = make([]*model.TopicReply, 0)
	searchTopics, _, err := s.dao.SearchTopic(ctx, req.Keyword, int(req.Page), int(req.PageSize))
	if err != nil {
		return
	}
	// 最多取2个
	cutLen := model.MinInt(2, len(searchTopics))
	searchTopics = searchTopics[:cutLen]
	for _, topic := range searchTopics {
		if topic != nil {
			topics = append(topics, &model.TopicReply{
				TopicId:   topic.TopicId,
				TopicName: topic.TopicName,
			})
		}
	}
	return
}

func (s *Service) allSearchLbsPart(ctx context.Context, req *model.SearchAllReq) (locations []*model.LocationReply, err error) {
	locations = make([]*model.LocationReply, 0)
	searchLocations, _, err := s.dao.SearchLabs(ctx, req.Keyword, req.Lat, req.Lng, int(req.Page), int(req.PageSize))
	if err != nil {
		return
	}
	// 最多取2个
	cutLen := model.MinInt(2, len(searchLocations))
	searchLocations = searchLocations[:cutLen]
	for _, loc := range searchLocations {
		if loc != nil {
			locations = append(locations, &model.LocationReply{
				Poi: loc.Pio,
			})
		}
	}
	return
}

func (s *Service) allSearchItemPart(ctx context.Context, req *model.SearchAllReq) (items []*model.ItemReply, err error) {
	items = make([]*model.ItemReply, 0)
	searchItems, _, err := s.dao.SearchMallItems(ctx, req.Uid, req.Keyword, int(req.Page), int(req.PageSize))
	if err != nil {
		return
	}
	// 最多取2个
	cutLen := model.MinInt(2, len(searchItems))
	searchItems = searchItems[:cutLen]
	for _, item := range searchItems {
		if item == nil {
			continue
		}
		var priceEqual = 1
		if item.PriceEqual != nil {
			log.Info("priceEqual != nil")
			priceEqual = *item.PriceEqual
		}
		items = append(items, &model.ItemReply{
			Name:           item.Title,
			Url:            item.Url,
			SchemaUrl:      "",
			ItemId:         item.ItemId,
			SourceType:     2,
			Cover:          item.Cover,
			Price:          strconv.FormatFloat(item.Price/100, 'f', 2, 64), // 电商接口返回的price以"分"为单位
			RequiredNumber: item.RequiredNumber,
			Brief:          item.Brief,
			PriceEqual:     priceEqual,
		})
	}
	return
}
