package feed

import (
	"context"

	"go-common/library/log"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
)

// ArcsWithPlayurl archives witch player
func (s *Service) ArcsPlayer(c context.Context, aids []*arcgrpc.PlayAv) (res map[int64]*arcgrpc.ArcPlayer, err error) {
	if res, err = s.arc.ArcsPlayer(c, aids); err != nil {
		log.Error("%+v", err)
	}
	return
}
