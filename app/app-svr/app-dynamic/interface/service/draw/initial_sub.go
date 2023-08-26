package draw

import (
	"context"

	model "go-gateway/app/app-svr/app-dynamic/interface/model/draw"
)

func (s *Service) allInitialUserPart(ctx context.Context, uid uint64) (users []*model.UserReply, err error) {
	var (
		atUsers, followUsers []*model.UserSearchItem
	)
	// 从最近@和最新关注接口获取userinfo
	atUsers, err = s.dao.GetUserLatestAtUsers(ctx, uid, "")
	// nolint:gomnd
	if len(atUsers) < 2 { // 最近@用户不足2，则用最新关注补足
		followUsers, err = s.dao.GetUserLatestFollowTopK(ctx, uid, 2-len(users), "")
		if err != nil {
			return
		}
	}
	// 遍历+去重
	var (
		tc     = len(atUsers) + len(followUsers) //total count
		ac     = model.MinInt(tc, 2)             // actual count 实际取多少个userinfo
		rc     = 0                               // 当前取到的unique userinfo
		unique = make(map[uint64]struct{}, ac)   // 用于去重
	)
	if tc == 0 {
		return
	}
	retUsers := append(atUsers, followUsers...)
	for _, user := range retUsers {
		if rc >= ac {
			break
		}
		if user == nil {
			continue
		}
		if _, ok := unique[user.Mid]; !ok {
			unique[user.Mid] = struct{}{}
			users = append(users, &model.UserReply{
				Profile: user.Face,
				Name:    user.Name,
				Uid:     user.Mid,
			})
			rc++
		}
	}
	return
}

func (s *Service) allInitialTopicPart(ctx context.Context, _ uint64) (topics []*model.TopicReply, err error) {
	hotTopics, err := s.dao.GetHotTopicTopK(ctx, 2)
	if err != nil {
		return
	}
	for _, topic := range hotTopics {
		if topic != nil {
			topics = append(topics, &model.TopicReply{
				TopicId:   topic.TopicId,
				TopicName: topic.TopicName,
			})
		}
	}
	return
}

func (s *Service) allInitialLBSPart(ctx context.Context, lat, lng float64) (locations []*model.LocationReply, err error) {
	nearbylocs, err := s.dao.GetNearbyLocationsTopK(ctx, 2, lat, lng)
	if err != nil {
		return
	}
	// lbs page最小为3
	cutLen := model.MinInt(2, len(nearbylocs))
	nearbylocs = nearbylocs[:cutLen]
	for _, loc := range nearbylocs {
		if loc != nil {
			locations = append(locations, &model.LocationReply{
				Poi: loc.Pio,
			})
		}
	}
	return
}
