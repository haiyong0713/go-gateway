package service

import (
	"context"

	"go-common/library/log"
)

func (s *Service) loadFawkesVersion() {
	log.Info("start load fawkes version")
	fawkesTmp, err := s.dao.FawkesVersion(context.Background())
	if err != nil {
		log.Error("【loadFawkesVersion】load fawkes error (%v)", err)
		return
	}
	s.fawkesVersionCache = fawkesTmp
}
