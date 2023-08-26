package service

import (
	"context"

	"go-common/library/log"
	"go-gateway/app/web-svr/web/interface/model"
)

const _firstPage = 1

var _emptyHelpList = make([]*model.HelpList, 0)

// HelpList get help menu list
func (s *Service) HelpList(c context.Context, pTypeID string, version int) (res []*model.HelpList, err error) {
	if res, err = s.dao.HlCache(c, pTypeID, version); err != nil || len(res) == 0 {
		if res, err = s.dao.HelpList(c, pTypeID, version); err != nil {
			log.Error("s.do.HelpList(%s) error(%v)", pTypeID, err)
			return
		}
		if len(res) > 0 {
			if err := s.cache.Do(c, func(c context.Context) {
				if err := s.dao.SetHlCache(c, pTypeID, version, res); err != nil {
					log.Error("%+v", err)
				}
			}); err != nil {
				log.Error("%+v", err)
			}
		}
	}
	return
}

// HelpDetail get help detail
func (s *Service) HelpDetail(c context.Context, fID, qTypeID string, keyFlag, pn, ps, version int) (resD []*model.HelpDeatil, resL []*model.HelpList, total int, err error) {
	if resD, total, err = s.dao.DetailCache(c, qTypeID, keyFlag, pn, ps, version); err != nil || len(resD) == 0 {
		if resD, total, err = s.dao.HelpDetail(c, qTypeID, keyFlag, pn, ps, version); err != nil {
			log.Error("s.do.HelpDetail(%s,%d,%d,%d) error(%v)", qTypeID, keyFlag, pn, ps, err)
		}
		if pn == _firstPage && len(resD) > 0 {
			if err := s.cache.Do(c, func(c context.Context) {
				if err := s.dao.SetDetailCache(c, qTypeID, keyFlag, pn, ps, version, total, resD); err != nil {
					log.Error("%+v", err)
				}
			}); err != nil {
				log.Error("%+v", err)
			}
		}
	}
	if fID == "" {
		resL = _emptyHelpList
	} else {
		if resL, err = s.HelpList(c, fID, version); err != nil {
			log.Error("s.HelpList(%s) error(%v)", fID, err)
		}
	}
	return
}

// HelpSearch get help search
func (s *Service) HelpSearch(c context.Context, pTypeID, keyWords string, keyFlag, pn, ps, version int) (res []*model.HelpDeatil, total int, err error) {
	if res, total, err = s.dao.HelpSearch(c, pTypeID, keyWords, keyFlag, pn, ps, version); err != nil {
		log.Error("s.do.HelpDetail(%s,%d,%d,%d) error(%v)", keyWords, keyFlag, pn, ps, err)
	}
	return
}
