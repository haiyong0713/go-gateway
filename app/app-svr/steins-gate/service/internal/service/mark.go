package service

import (
	"context"
	"time"

	"go-common/library/log"

	"go-common/library/sync/errgroup.v2"

	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/app-svr/steins-gate/ecode"
	"go-gateway/app/app-svr/steins-gate/service/api"
	"go-gateway/app/app-svr/steins-gate/service/internal/model"
)

func (s *Service) AddMark(c context.Context, aid, mid, mark int64) (err error) {
	var (
		arc       *arcgrpc.Arc
		graphInfo *api.GraphInfo
	)
	eg := errgroup.WithContext(c)
	eg.Go(func(c context.Context) (err error) {
		if arc, err = s.arcDao.Arc(c, aid); err != nil {
			log.Error("AddMark d.Arc(aid:%d) err:%v", aid, err)
			return
		}
		return
	})
	eg.Go(func(c context.Context) (err error) {
		if graphInfo, err = s.GraphInfo(c, aid); err != nil { // graph info picking
			log.Error("AddMark s.GraphInfo(aid:%d) err(%v)", aid, err)
			return
		}
		return
	})
	if err = eg.Wait(); err != nil {
		return
	}
	if !arc.IsSteinsGate() { // not a steinsGate arc, just return
		log.Warn("AddMark Aid %d, Not SteinsGate Arc", aid)
		err = ecode.NotSteinsGateArc
		return
	}
	if err = s.markDao.AddMark(c, aid, mid, mark); err != nil {
		log.Error("AddMark s.dao.AddMark(aid:%d,mid:%d,mark:%d) err(%v)", aid, mid, mark, err)
		return
	}
	infocMsg := &model.InfocMark{ // infoc 消息
		MID:          mid,
		AID:          aid,
		GraphVersion: graphInfo.Id,
		Mark:         mark,
		LogTime:      time.Now().Unix(),
	}
	s.infoc(infocMsg)
	return
}

func (s *Service) GetMark(c context.Context, aid, mid int64) (mark int64, err error) {
	mark, err = s.markDao.Mark(c, aid, mid)
	return

}
