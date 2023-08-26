package timemachine

import (
	"context"

	"go-common/library/ecode"
	"go-common/library/log"

	"go-gateway/app/web-svr/activity/interface/model/timemachine"
)

// Timemachine2019Raw .
func (s *Service) Timemachine2019Raw(c context.Context, loginMid, mid int64) (data *timemachine.Item, err error) {
	if _, ok := s.tmMidMap[loginMid]; !ok {
		err = ecode.AccessDenied
		return
	}
	if mid == 0 {
		mid = loginMid
	}
	if data, err = s.dao.RawTimemachine(c, mid); err != nil {
		log.Error("Timemachine2018 s.dao.RawTimemachine(%d) error(%v)", mid, err)
	}
	return
}

func (s *Service) Timemachine2019Cache(c context.Context, loginMid, mid int64) (data *timemachine.Item, err error) {
	if _, ok := s.tmMidMap[loginMid]; !ok {
		err = ecode.AccessDenied
		return
	}
	if mid == 0 {
		mid = loginMid
	}
	if data, err = s.dao.CacheTimemachine(c, mid); err != nil {
		log.Error("Timemachine2018 s.dao.RawTimemachine(%d) error(%v)", mid, err)
	}
	return
}

// Timemachine2019Raw .
func (s *Service) Timemachine2019Reset(c context.Context, loginMid, mid int64) (err error) {
	if _, ok := s.tmMidMap[loginMid]; !ok {
		err = ecode.AccessDenied
		return
	}
	if mid == 0 {
		mid = loginMid
	}
	data, err := s.dao.RawTimemachine(c, mid)
	if err != nil {
		log.Error("Timemachine2019Reset s.dao.RawTimemachine(%d) error(%v)", mid, err)
		return
	}
	if err = s.dao.AddCacheTimemachine(c, mid, data); err != nil {
		log.Error("Timemachine2019Reset s.dao.AddCacheTimemachine(%d) error(%v)", mid, err)
	}
	return
}
