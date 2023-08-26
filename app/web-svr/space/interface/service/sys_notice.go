package service

import (
	"context"

	"go-common/library/log"
	"go-gateway/app/web-svr/space/interface/model"
)

// sysNoticeMap system notice
func (s *Service) sysNoticeMap(c context.Context) {
	var (
		err          error
		sysNotice    map[int64]*model.SysNotice
		sysNoticeUID []*model.SysNoticeUid
	)
	tmp := make(map[int64]*model.SysNotice)
	if sysNotice, err = s.dao.SysNoticelist(c); err != nil {
		log.Error("Service.sysNoticeMap SysNoticelist error(%v)", err)
		return
	}
	if sysNoticeUID, err = s.dao.SysNoticeUIDlist(c); err != nil {
		log.Error("Service.sysNoticeMap SysNoticeUIDlist error(%v)", err)
		return
	}
	for sysID, notice := range sysNotice {
		for _, uidValue := range sysNoticeUID {
			if sysID == uidValue.SystemNoticeId {
				tmp[uidValue.Uid] = notice
			}
		}
	}
	log.Info("loadBlacklist success len(blacklist):%d", len(tmp))
	s.SysNotice = tmp
}
