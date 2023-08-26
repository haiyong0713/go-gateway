package service

import (
	"context"

	"go-gateway/app/web-svr/space/interface/model"
)

// TagSub subscribe tag.
func (s *Service) TagSub(c context.Context, mid, tid int64) error {
	return s.dao.TagSub(c, mid, tid)
}

// TagCancelSub cancel subscribe tag.
func (s *Service) TagCancelSub(c context.Context, mid, tid int64) error {
	return s.dao.TagCancelSub(c, mid, tid)
}

// TagSubList get tag subscribe list by mid.
func (s *Service) TagSubList(c context.Context, mid, vmid int64, pn, ps int) (rs []*model.Tag, total int, err error) {
	if mid != vmid {
		if err = s.privacyCheck(c, vmid, model.PcyTag); err != nil {
			return
		}
	}
	if rs, total, err = s.dao.TagSubList(c, vmid, pn, ps); err != nil {
		return
	}
	if len(rs) == 0 {
		rs = make([]*model.Tag, 0)
	}
	return
}
