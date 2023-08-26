package service

import (
	"context"
	"strconv"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"

	"go-gateway/app/app-svr/archive/job/model/archive"
	"go-gateway/app/app-svr/archive/job/model/result"
	"go-gateway/app/app-svr/archive/job/model/retry"
	"go-gateway/app/app-svr/archive/service/api"
)

const (
	_arcVideosUpdate = 0
	_arcActionUpSid  = 1
	_arcActionRmSid  = 2
)

func (s *Service) isPGC(upFrom int32) bool {
	if upFrom == archive.UpFromPGC || upFrom == archive.UpFromPGCSecret {
		return true
	}
	return false
}

func (s *Service) ugcConsumer() {
	defer s.waiter.Done()
	for {
		aid, ok := <-s.ugcAidChan
		if !ok {
			log.Error("s.videoupAids chan closed")
			return
		}
		s.limiter.ugc.Wait()
		s.arcUpdate(aid, _arcVideosUpdate, 0)
	}
}

func (s *Service) pgcConsumer() {
	defer s.waiter.Done()
	for {
		aid, ok := <-s.pgcAidChan
		if !ok {
			log.Error("s.pgcAids closed")
			return
		}
		s.limiter.ogv.Wait()
		s.arcUpdate(aid, _arcVideosUpdate, 0)
	}
}

func (s *Service) nbConsumer() {
	defer s.waiter.Done()
	for {
		aid, ok := <-s.nbAidChan
		if !ok {
			log.Error("s.nbAids closed")
			return
		}
		s.limiter.other.Wait()
		log.Info("nbConsumer aid(%d)", aid)
		s.arcUpdate(aid, _arcVideosUpdate, 0)
	}
}

func (s *Service) arcUpdate(aid int64, upAction int, sid int64) {
	var (
		oldResult     *api.Arc
		newResult     *api.Arc
		c             = context.TODO()
		cids, delCids []int64
		err           error
		changed       bool
		now           = time.Now()
	)
	log.Info("sync resultDB archive(%d) upAction(%d) sid(%d) start", aid, upAction, sid)
	defer func() {
		bm.MetricServerReqDur.Observe(int64(time.Since(now)/time.Millisecond), "arcUpdate", "job")
		bm.MetricServerReqCodeTotal.Inc("arcUpdate", "job", strconv.FormatInt(int64(ecode.Cause(err).Code()), 10))
		if err != nil {
			rt := &retry.Info{Action: retry.FailResultAdd}
			rt.Data.Aid = aid
			rt.Data.ArcAction = upAction
			rt.Data.SeasonID = sid
			s.PushFail(c, rt, retry.FailList)
			log.Error("s.arcUpdate(%d) error(%+v)", aid, err)
		}
	}()
	if oldResult, _, err = s.resultDao.RawArc(c, aid); err != nil { // need retry
		log.Error("s.resultDao.Archive(%d) error(%+v)", aid, err)
		return
	}
	if oldResult == nil && (upAction == _arcActionUpSid || upAction == _arcActionRmSid) {
		log.Warn("Archive %d Not passed was joined the season, Action %d", aid, upAction)
		return
	}
	switch upAction {
	case _arcVideosUpdate:
		if changed, cids, delCids, err = s.tranResult(c, aid); err != nil || !changed {
			log.Error("aid(%d) nothing changed err(%+v)", aid, err)
			return
		}
		s.putVideoShotChan(context.Background(), cids)
		s.upVideoCache(c, aid)
		s.delVideoCache(c, aid, delCids)
	case _arcActionUpSid:
		if oldResult.SeasonID == sid {
			return
		}
		if err = s.resultDao.UpArcSID(c, sid, aid); err != nil {
			log.Error("s.result.UpArcSID sid(%d) aid(%d) error(%+v)", sid, aid, err)
			return
		}
	case _arcActionRmSid:
		if oldResult.SeasonID == 0 {
			return
		}
		if err = s.resultDao.DelArcSID(c, sid, aid); err != nil {
			log.Error("s.result.DelArcSID sid(%d) aid(%d) error(%+v)", sid, aid, err)
			return
		}
	}
	if newResult, _, err = s.resultDao.RawArc(c, aid); err != nil { // need retry
		log.Error("s.resultDao.Archive(%d) error(%+v)", aid, err)
		return
	}
	if newResult == nil { // 创作端之前允许未过审稿件进入导致脏数据
		log.Warn("newResult Empty Aid %d", aid)
		return
	}
	s.updateResultCache(newResult, oldResult, true)
	action := "update"
	if oldResult == nil {
		action = "insert"
	}
	s.sendNotify(&result.ArchiveUpInfo{Table: "archive", Action: action, Nw: newResult, Old: oldResult})
	log.Info("sync resultDB archive(%d) sync old(%+v) new(%+v) inserted", aid, oldResult, newResult)
}

func (s *Service) hadPassed(c context.Context, aid int64) (had bool) {
	id, err := s.archiveDao.RawGetFirstPassByAID(c, aid)
	if err != nil {
		log.Error("hadPassed s.arc.GetFirstPassByAID error(%+v) aid(%d)", err, aid)
		return
	}
	had = id > 0
	return
}

func (s *Service) syncCreativeType() {
	var c = context.Background()
	ts, err := s.archiveDao.RawTypes(c)
	if err != nil {
		log.Error("syncArchiveType error(%+v)", err)
		return
	}
	var tids []int64
	for _, t := range ts {
		tids = append(tids, t.ID)
		if err = s.resultDao.AddType(c, t); err != nil {
			log.Error("syncArchiveType error(%+v)", err)
			continue
		}
	}
	if len(tids) > 0 {
		if err = s.resultDao.DelTypes(c, tids); err != nil {
			log.Error("syncArchiveType error(%+v)", err)
			return
		}
	}
}
