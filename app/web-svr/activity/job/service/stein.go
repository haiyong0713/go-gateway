package service

import (
	"context"

	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/model/like"
)

func (s *Service) steinListproc() {
	var (
		oneCnt, twoCnt int64
		err            error
	)
	c := context.Background()
	if oneCnt, err = s.dao.SteinRuleCount(c, s.c.Stein.Sid, s.c.Stein.OneViewRule, s.c.Stein.OneLikeRule); err != nil {
		log.Error("steinListproc oneCnt error(%v)", err)
		return
	}
	if twoCnt, err = s.dao.SteinRuleCount(c, s.c.Stein.Sid, s.c.Stein.TwoViewRule, s.c.Stein.TwoLikeRule); err != nil {
		log.Error("steinListproc twoCnt error(%v)", err)
		return
	}
	if oneCnt == 0 && twoCnt == 0 {
		log.Warn("steinListproc zero one(%d) tow(%d)", oneCnt, twoCnt)
		return
	}
	data := &like.SteinList{AwardOne: oneCnt, AwardTwo: twoCnt}
	if err = s.dao.SetSteinCache(c, data); err != nil {
		log.Error("steinListproc cache set(%v) error(%v)", data, err)
		return
	}
	log.Info("steinListproc success data(%v)", data)
}
