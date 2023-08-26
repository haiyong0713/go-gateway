package feed

import (
	"context"

	"go-common/library/log"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
)

// ArcsPlayer .
func (s *Service) ArcsPlayer(c context.Context, aids []int64) (res map[int64]*arcgrpc.ArcPlayer, err error) {
	if res, err = s.arc.ArcsPlayer(c, aids, ""); err != nil {
		log.Error("s.arc.ArcsPlayer, error(%+v)", err)
	}
	return
}

// storyArcPlayer .
func (s *Service) storyArcsPlayer(c context.Context, aids []int64) (res map[int64]*arcgrpc.ArcPlayer, err error) {
	if res, err = s.arc.ArcsPlayer(c, aids, "story"); err != nil {
		log.Error("%+v", err)
	}
	return
}
