package service

import (
	"context"
	"time"

	"go-common/library/log"
	"go-gateway/app/web-svr/esports/ecode"
	"go-gateway/app/web-svr/esports/interface/model"
)

const _interval = 3000

// LiveMatchs live matchs for live
func (s *Service) LiveMatchs(c context.Context, mid int64, cids []int64) (rs map[int64]*model.Contest, err error) {
	var (
		cData   []*model.Contest
		cidsTmp []int64
	)
	if len(cids) == 0 {
		return rs, nil
	}
	if rs, err = s.dao.RawEpContests(c, cids); err != nil {
		return
	}
	for _, v := range rs {
		cData = append(cData, v)
		cidsTmp = append(cidsTmp, v.ID)
	}
	rsTmp := s.ContestInfo(c, cidsTmp, cData, mid)
	for _, v := range rsTmp {
		rs[v.ID] = v
	}
	return
}

// LiveMatchsAct live matchs act for live
func (s *Service) LiveMatchsAct(c context.Context, param *model.MatchLive) (rs *model.LivePager, err error) {
	var (
		total      int
		tmpRs      []*model.Contest
		dbContests map[int64]*model.Contest
		cids       []int64
	)
	p := &model.ParamContest{
		Ps: param.Ps,
		Pn: param.Pn,
	}
	if param.GID > 0 {
		p.Gid = param.GID
	}
	if param.MID > 0 {
		p.Mid = param.MID
	}
	if param.SID > 0 {
		p.Sids = []int64{param.SID}
	}
	if param.STime != 0 {
		p.Stime = time.Unix(param.STime, 0).Format("2006-01-02 15:04:05")
	}
	rs = &model.LivePager{
		Page: model.Page{
			Num:   param.Pn,
			Size:  param.Ps,
			Total: 0,
		},
	}
	if cids, total, err = s.dao.SearchContestQuery(c, p); err != nil {
		log.Error("s.dao.LiveMatchsAct.SearchContestQuery error(%v)", err)
		return
	}
	if total == 0 || len(cids) == 0 {
		rs.Item = _emptContest
		return
	}
	if len(cids) > 0 {
		if dbContests, err = s.dao.EpContests(c, cids); err != nil {
			log.Error("s.dao.Contest error(%v)", err)
			return
		}
	}
	for _, cid := range cids {
		if contest, ok := dbContests[cid]; ok {
			tmpRs = append(tmpRs, contest)
		}
	}
	rs.Item = s.ContestInfo(c, cids, tmpRs, param.MID)
	rs.Page.Total = total
	return rs, nil
}

func (s *Service) ScoreBattleList(c context.Context, matchID string) (res interface{}, err error) {
	res = struct{}{}
	if matchID == "" {
		matchID = s.liveMatchID
	}
	if matchID == "" {
		log.Error("ScoreBattleList matchID empty")
		err = ecode.EsportsLiveNoList
		return
	}
	battleList := s.LoadLiveBattleListMap()
	bl, ok := battleList[matchID]
	if !ok {
		log.Error("ScoreBattleList memory matchID(%s) not ok", matchID)
		// 直接从redis中取
		if bl, err = s.dao.CacheBattleList(c, matchID); err != nil {
			log.Error("ScoreBattleList s.dao.CacheBattleList matchID(%s) error(%+v)", matchID, err)
			err = ecode.EsportsLiveNoList
			return
		}
	}
	if bl == nil {
		err = ecode.EsportsLiveNoList
		return
	}
	bl.Interval = s.c.Score.LiveInterval
	if bl.Interval == 0 {
		bl.Interval = _interval
	}
	res = bl
	return
}

func (s *Service) ScoreBattleInfo(c context.Context, battleString string) (res interface{}, err error) {
	res = struct{}{}
	battleInfo := s.LoadLiveBattleInfoMap()
	bi, ok := battleInfo[battleString]
	if !ok {
		if bi, err = s.dao.CacheBattleInfo(c, battleString); err != nil {
			log.Error("ScoreBattleInfo s.dao.CacheBattleInfo battleString(%s) error(%+v)", battleString, err)
			err = ecode.EsportsLiveNoInfo
			return
		}
	}
	if bi == nil {
		err = ecode.EsportsLiveNoInfo
		return
	}
	res = bi
	return
}

func (s *Service) StoreLiveBattleListMap(m map[string]*model.BattleList) {
	s.liveBattleListMap.Store(m)
}

func (s *Service) LoadLiveBattleListMap() map[string]*model.BattleList {
	return s.liveBattleListMap.Load().(map[string]*model.BattleList)
}

func (s *Service) StoreLiveBattleInfoMap(m map[string]*model.BattleInfo) {
	s.liveBattleInfoMap.Store(m)
}

func (s *Service) LoadLiveBattleInfoMap() map[string]*model.BattleInfo {
	return s.liveBattleInfoMap.Load().(map[string]*model.BattleInfo)
}
