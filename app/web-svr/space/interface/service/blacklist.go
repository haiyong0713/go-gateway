package service

import (
	"context"

	"go-common/library/log"
)

// blacklist space blacklist
func (s *Service) blacklist(c context.Context) {
	var (
		blTmp []int64
		err   error
	)
	if blTmp, err = s.dao.Blacklist(c); err != nil {
		log.Error("Service.blacklist error(%v)", err)
		return
	}
	blacklist := make(map[int64]struct{}, len(blTmp))
	for _, mid := range blTmp {
		blacklist[mid] = struct{}{}
	}
	log.Info("loadBlacklist success len(blacklist):%d", len(blacklist))
	s.BlacklistValue = blacklist
}
