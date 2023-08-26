package service

import (
	"context"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/time"

	"go-gateway/app/app-svr/archive/service/api"
)

// UpperPassed3 upper passed.
func (s *Service) UpperPassed3(c context.Context, mid int64, pn, ps int) (as []*api.Arc, err error) {
	var cnt int
	if cnt, err = s.UpperCount(c, mid); err != nil {
		log.Error("s.UpperCount(%d) error(%v)", mid, err)
		return
	}
	if cnt == 0 {
		err = ecode.NothingFound
		return
	}
	if pn < 1 {
		pn = 1
	}
	if ps < 1 {
		ps = 20
	}
	var (
		start = (pn - 1) * ps
		end   = start + ps - 1
		aids  []int64
	)
	if ok, _ := s.arc.ExpireUpperPassedCache(c, mid); !ok {
		var (
			alls       []int64
			ptimes     []time.Time
			copyrights []int8
		)
		if alls, ptimes, copyrights, err = s.arc.RawUpperPassed(c, mid); err != nil {
			log.Error("s.arc.UpperPassed(%d) error(%v)", mid, err)
			return
		}
		length := len(alls)
		if length == 0 || length < start {
			err = ecode.NothingFound
			return
		}
		s.addCache(func() {
			_ = s.arc.AddUpperPassedCache(context.TODO(), mid, alls, ptimes, copyrights)
		})
		if length > end+1 {
			aids = alls[start : end+1]
		} else {
			aids = alls[start:]
		}
	} else {
		if aids, err = s.arc.UpperPassedCache(c, mid, start, end); err != nil {
			log.Error("s.arc.UpperPassedCache(%d) error(%v)", mid, err)
			return
		}
	}
	if len(aids) == 0 {
		return
	}
	var am map[int64]*api.Arc
	if am, err = s.Archives3(c, aids, 0, "", ""); err != nil {
		log.Error("s.arc.Archives(%v) error(%v)", aids, err)
		return
	}
	as = make([]*api.Arc, 0, len(am))
	for _, aid := range aids {
		if a, ok := am[aid]; ok {
			as = append(as, a)
		}
	}
	return
}
