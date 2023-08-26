package service

import (
	"context"

	"go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/app-svr/playurl/service/model/archive"
)

// 收藏此视频的用户、Up主、联合投稿人 返回 true
func (s *Service) validateForOnlyFav(c context.Context, mid int64, arc *archive.Info) (bool, error) {
	if mid == 0 || arc == nil {
		return false, nil
	}
	if arc.Mid == mid {
		return true, nil
	}
	if arc.AttrVal(api.AttrBitIsCooperation) == api.AttrYes {
		staffs, err := s.arcDao.Creators(c, arc.Aid)
		if err != nil {
			return false, err
		}
		for _, staff := range staffs {
			if mid == staff {
				return true, nil
			}
		}
	}
	is, err := s.arcDao.IsFav(c, arc.Aid, mid)
	if err != nil {
		return false, err
	}
	return is, nil
}
