package online

import (
	"context"
	"fmt"

	"go-common/library/ecode"
	"go-common/library/log"

	v1 "go-gateway/app/app-svr/player-online/api"

	"github.com/pkg/errors"
)

const (
	threeDays = 3600 * 24 * 3
)

func (s *Service) ReportWatchGRPC(c context.Context, req *v1.ReportWatchReq) (*v1.NoReply, error) {
	res := &v1.NoReply{}
	if req.Aid <= 0 || req.Biz == "" {
		return nil, errors.New("bad params!")
	}
	if ok, _ := s.redisDao.ExistPremiereUserWatch(c, req.Aid, req.Buvid); ok {
		log.Error("ReportWatchGRPC req(%+v) ok(%t)", req, ok)
		return res, nil
	}
	if err := s.redisDao.IncreasePremiereCountCache(c, req.Aid, req.Buvid); err != nil {
		return nil, err
	}
	if err := s.redisDao.SetPremiereUserWatch(c, req.Aid, req.Buvid, threeDays); err != nil {
		return nil, err
	}
	log.Warn("ReportWatchGRPC req(%+v) success", req)
	return res, nil
}

// PremiereInfoGRPC warden server list
func (s *Service) PremiereInfoGRPC(c context.Context, req *v1.PremiereInfoReq) (reply *v1.PremiereInfoReply, err error) {
	if req.Aid <= 0 {
		return nil, errors.New("bad params!")
	}
	var (
		participant int64
		interaction int64
	)

	//获取xx人参与
	if participant, err = s.redisDao.GetPremiereCountCache(c, req.Aid); err != nil {
		return nil, ecode.NothingFound
	}
	//获取xx次互动
	if interaction, err = s.redisDao.GetRoomStatisticsCache(c, req.Aid); err != nil {
		interaction, err = s.getRoomStatistics(c, req.Aid)
		if err != nil {
			return nil, ecode.NothingFound
		}
	}
	return &v1.PremiereInfoReply{
		PremiereOverText: fmt.Sprintf("期间共%d人参与，发生%d次互动", participant, interaction),
		Participant:      participant,
		Interaction:      interaction,
	}, nil
}

func (s *Service) getRoomStatistics(c context.Context, aid int64) (int64, error) {
	//获取roomId
	a, err := s.arcDao.SimpleArc(c, aid)
	if err != nil {
		return 0, err
	}
	if a == nil || a.Premiere == nil || a.Premiere.RoomId == 0 {
		return 0, nil
	}
	res, err := s.pgcDao.GetUGCPremiereRoomStatistics(c, a.Premiere.RoomId)
	if err != nil {
		return 0, err
	}
	_ = s.redisDao.SetRoomStatisticsCache(c, aid, 300, int64(res.InteractCount))
	return int64(res.InteractCount), nil
}
