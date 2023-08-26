package service

import (
	"context"

	"go-common/library/log"
	actpb "go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/esports/interface/model"
)

// S9Result .
func (s *Service) S9Result(c context.Context, mid, sid int64) (res *actpb.UserGuessResultReply, err error) {
	var cids []int64
	if cids, err = s.s9Cids(c, sid); err != nil {
		return
	}
	if len(cids) == 0 {
		res = &actpb.UserGuessResultReply{}
		return
	}
	req := &actpb.UserGuessResultReq{
		Mid:       mid,
		Oids:      cids,
		Sid:       sid,
		StakeType: _guessStackType,
		Business:  _guessBusID,
	}
	return s.actClient.UserGuessResult(c, req)
}

// S9Record .
func (s *Service) S9Record(c context.Context, mid, sid, pn, ps int64) (res *actpb.UserGuessMatchsReply, err error) {
	var cids []int64
	if cids, err = s.s9Cids(c, sid); err != nil {
		return
	}
	if len(cids) == 0 {
		res = &actpb.UserGuessMatchsReply{}
		return
	}
	req := &actpb.UserGuessMatchsReq{
		Mid:      mid,
		Oids:     cids,
		Business: _guessBusID,
		Pn:       pn,
		Ps:       ps,
	}
	return s.actClient.UserGuessMatchs(c, req)
}

func (s *Service) s9Cids(c context.Context, sid int64) (rs []int64, err error) {
	p := &model.ParamContest{
		GsType: _isGuess,
		Sids:   []int64{sid},
		Pn:     1,
		Ps:     s.c.Rule.S9GuessMax,
	}
	if rs, _, err = s.dao.SearchContestQuery(c, p); err != nil {
		log.Error("s.dao.SearchContestQuery  param(%v) error(%v)", p, err)
	}
	return
}
