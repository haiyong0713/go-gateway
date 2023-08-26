package service

import (
	"context"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/web-svr/dance-taiko/interface/internal/model"
)

const (
	_ott = "ott"
	_app = "app"
)

func (s *Service) KeyFrames(c context.Context, cid int64, plat string) (*model.LoadReply, error) {
	frames, err := s.ottDao.LoadFrames(c, cid)
	if err != nil {
		log.Error("KeyFrames cid(%d) err(%v)", cid, err)
		return nil, err
	}
	if frames == nil {
		return nil, ecode.NothingFound
	}
	res := &model.LoadReply{Url: frames.Url}
	switch plat {
	case _ott:
		res.Qn = s.conf.OttCfg.Qn
	case _app:
		res.Img = s.conf.OttCfg.ImgUrl
	default:
		log.Warn("KeyFrames wrong plat(%s)", plat)
	}
	return res, nil
}
