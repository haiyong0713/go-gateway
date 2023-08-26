package service

import (
	"context"

	"go-gateway/app/app-svr/archive/service/api"
)

// Stat3 get archive stat.
func (s *Service) Stat3(c context.Context, aid int64) (st *api.Stat, err error) {
	st, err = s.arc.Stat3(c, aid)
	return
}

// Stats3 get archive stat.
func (s *Service) Stats3(c context.Context, aids []int64) (stm map[int64]*api.Stat, err error) {
	stm, err = s.arc.Stats3(c, aids)
	return
}
