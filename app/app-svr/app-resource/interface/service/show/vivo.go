package show

import (
	"context"

	"go-gateway/app/app-svr/app-resource/interface/model/show"
)

func (s *Service) VIVOPopularBadge(ctx context.Context) (*show.VIVOPopularBadgeReply, error) {
	return &show.VIVOPopularBadgeReply{
		HotUpdateInterval:     3,
		KeywordUpdateInterval: 8,
		Slogan:                "你感兴趣的视频都在B站",
	}, nil
}
