package service

import (
	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/app-svr/archive-push/ecode"
)

// GetMIDByOpenID 根据用户Open ID获取MID
func (s *Service) GetMIDByOpenID(openID string, appKey string) (mid int64, err error) {
	if openID == "" {
		return 0, xecode.RequestErr
	}
	res, _err := s.dao.GetMIDByOpenID(openID, appKey)
	if _err != nil {
		err = ecode.AccountPlatRequestError
		log.Error("Service: GetMIDByOpenID(%s, %s) error %v", openID, appKey, err)
		return
	}
	if res.Code != 0 {
		err = ecode.AccountPlatResponseError
		log.Error("Service: GetMIDByOpenID(%s, %s) response %+v error %v", openID, appKey, res, err)
		return
	}
	mid = res.Data.MID

	return
}

// GetOpenIDByMID 根据用户MID获取Open ID
func (s *Service) GetOpenIDByMID(mid int64, appKey string) (openID string, err error) {
	if mid == 0 {
		return "", xecode.RequestErr
	}
	res, _err := s.dao.GetOpenIDByMID(mid, appKey)
	if _err != nil {
		err = ecode.AccountPlatRequestError
		log.Error("Service: GetOpenIDByMID(%d, %s) error %v", mid, appKey, err)
		return
	}
	if res.Code != 0 {
		err = ecode.AccountPlatResponseError
		log.Error("Service: GetOpenIDByMID(%d, %s) response %+v error %v", mid, appKey, res, err)
		return
	}
	openID = res.Data.OpenID

	return
}
