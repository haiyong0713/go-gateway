package service

import (
	"context"
	"encoding/json"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	actmdl "go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/esports/admin/model"
	xecode "go-gateway/app/web-svr/esports/ecode"
	v1 "go-gateway/app/web-svr/esports/interface/api/v1"
)

const (
	_defaultTotal = 10
	_notGuess     = 0
	_haveGuess    = 1
)

func (s *Service) AddGuess(c context.Context, p *model.ParamAdd) (err error) {
	var (
		groups, newGroup []*actmdl.GuessGroup
		tmpGroup         *actmdl.GuessGroup
		contest          *model.Contest
		count            int64
	)
	if contest, err = s.contestExit(p.Cid); err != nil {
		return
	}
	if count, err = s.s10ContestExist(p.Cid); err != nil {
		err = xecode.EsportsDrawPost
		return
	}
	if count != 0 {
		eg := errgroup.WithContext(c)
		eg.Go(func(c context.Context) (err error) {
			preData := new(model.Team)
			if err = s.dao.DB.Where("id=?", contest.HomeID).First(&preData).Error; err != nil {
				log.Error("AddGuess s.dao.DB.Where id(%d) error(%d)", contest.HomeID, err)
				return
			}
			if err = s.DrawPost(c, preData, contest); err != nil {
				log.Error("s.DrawPost contest(%+v) error(%+v)", contest, err)
			}
			return
		})
		eg.Go(func(c context.Context) (err error) {
			preData := new(model.Team)
			if err = s.dao.DB.Where("id=?", contest.AwayID).First(&preData).Error; err != nil {
				log.Error("AddGuess s.dao.DB.Where id(%d) error(%d)", contest.HomeID, err)
				return
			}
			if err = s.DrawPost(c, preData, contest); err != nil {
				log.Error("s.DrawPost contest(%+v) error(%+v)", contest, err)
			}
			return
		})
		if err = eg.Wait(); err != nil {
			return
		}
	}
	log.Warn("s.DrawPost contest(%+v) count(%d) error(%+v)", contest, count, err)
	if err = json.Unmarshal([]byte(p.Groups), &groups); err != nil {
		log.Error("json.Unmarshal error(%+v)", err)
		err = ecode.RequestErr
		return
	}
	if len(groups) == 0 {
		err = ecode.RequestErr
		return
	}
	newGroup = make([]*actmdl.GuessGroup, 0, len(groups))
	for _, g := range groups {
		if g.Id == 0 && g.Title != "" {
			count := len(g.DetailAdd)
			if count == 0 {
				tmpGroup = &actmdl.GuessGroup{Title: g.Title}
			} else {
				for _, detail := range g.DetailAdd {
					detail.TotalStake = _defaultTotal
				}
				tmpDetail := make([]*actmdl.GuessDetailAdd, 0, count)
				tmpDetail = append(tmpDetail, g.DetailAdd...)
				tmpGroup = &actmdl.GuessGroup{
					Title:        g.Title,
					DetailAdd:    tmpDetail,
					TemplateType: g.TemplateType,
				}
			}
			newGroup = append(newGroup, tmpGroup)
			tmpGroup = nil
		}
	}
	if len(newGroup) == 0 {
		err = ecode.RequestErr
		return
	}
	if _, err = s.actClient.GuessAdd(c, &actmdl.GuessAddReq{Business: int64(actmdl.GuessBusiness_esportsType), StakeType: int64(actmdl.StakeType_coinType), MaxStake: s.c.Rule.MaxGuessStake, Oid: p.Cid, Stime: contest.Stime, Etime: contest.Etime, Groups: newGroup}); err != nil {
		log.Error("s.actClient.GuessAdd  param(%+v)", p)
		return
	}
	if err = s.dao.DB.Model(&model.Contest{}).Where("id=?", p.Cid).Update(map[string]int{"guess_type": _haveGuess}).Error; err != nil {
		log.Error("AddGuess s.dao.DB.Model error(%v)", err)
		return err
	}
	go s.BatchRefreshContestDataPageCache(context.Background(), []int64{p.Cid})
	// 删除赛程组件缓存.
	s.cache.Do(c, func(c context.Context) {
		if e := s.ClearComponentContestCacheByGRPC(&v1.ClearComponentContestCacheRequest{SeasonID: contest.Sid, ContestID: contest.ID}); e != nil {
			log.Error("contest component ClearComponentContestCacheGRPC SeasonID(%+v) ContestID(%d) error(%+v)", contest.Sid, contest.ID, err)
		}
	})
	return
}

func (s *Service) DelGuess(c context.Context, p *model.ParamDel) (err error) {
	var (
		rs      *actmdl.GuessGroupReply
		contest *model.Contest
	)
	if contest, err = s.contestExit(p.Cid); err != nil {
		return
	}
	if rs, err = s.actClient.GuessGroupDel(c, &actmdl.GuessGroupDelReq{MainID: p.MainID}); err != nil {
		log.Error("s.actClient.GuessGroupDel oid(%d) mainID(%d)  error(%+v)", p.Cid, p.MainID, err)
		return
	}
	if rs.HaveGuess == 0 {
		if err = s.dao.DB.Model(&model.Contest{}).Where("id=?", p.Cid).Update(map[string]int{"guess_type": _notGuess}).Error; err != nil {
			log.Error("DelGuess s.dao.DB.Model error(%v)", err)
			return err
		}
	}
	go s.BatchRefreshContestDataPageCache(context.Background(), []int64{p.Cid})
	// 删除赛程组件缓存.
	s.cache.Do(c, func(c context.Context) {
		if e := s.ClearComponentContestCacheByGRPC(&v1.ClearComponentContestCacheRequest{SeasonID: contest.Sid, ContestID: contest.ID}); e != nil {
			log.Error("contest component ClearComponentContestCacheGRPC SeasonID(%+v) ContestID(%d) error(%+v)", contest.Sid, contest.ID, err)
		}
	})
	return
}

func (s *Service) ResultGuess(c context.Context, p *model.ParamRes) (err error) {
	if _, err = s.contestExit(p.Cid); err != nil {
		return
	}
	if _, err = s.actClient.GuessUpResult(c, &actmdl.GuessUpResultReq{MainID: p.MainID, DetailID: p.DetailID}); err != nil {
		log.Error("s.actClient.GuessUpResult  oid(%d) mainID(%d) detailID(%d) error(%+v)", p.Cid, p.MainID, p.DetailID, err)
	}
	return
}

func (s *Service) contestExit(cid int64) (rs *model.Contest, err error) {
	rs = new(model.Contest)
	if err = s.dao.DB.Where("id=?", cid).First(&rs).Error; err != nil {
		log.Error("contestExit s.dao.DB.Where id(%d) error(%d)", cid, err)
	}
	return
}

func (s *Service) s10ContestExist(cid int64) (count int64, err error) {
	rs := new(model.Contest)
	if err = s.dao.DB.Model(&rs).Where("id=?", cid).Where("sid=?", s.c.S10CoinCfg.SeasonID).Count(&count).Error; err != nil {
		log.Error("s10ContestExist s.dao.DB.Where id(%d) error(%+v)", cid, err)
	}
	return
}

func (s *Service) s10GuessExist(tid int64) (count int64, err error) {
	rs := new(model.Contest)
	if err = s.dao.DB.Model(&rs).Where("sid=?", s.c.S10CoinCfg.SeasonID).Where("away_id=? OR home_id=?", tid, tid).Count(&count).Error; err != nil {
		log.Error("s10GuessExist s.dao.DB.Where tid(%d) error(%+v)", tid, err)
	}
	return
}

func (s *Service) listContest(tid int64) (rs []*model.Contest, err error) {
	rs = make([]*model.Contest, 0)
	if err = s.dao.DB.Model(&model.Contest{}).Where("guess_type=1 AND stime >= ?", time.Now().Add(10*time.Minute).Unix()).Where("away_id=? OR home_id=?", tid, tid).Find(&rs).Error; err != nil {
		log.Error("s.listContest id(%d) error(%+v)", tid, err)
	}
	return
}

func (s *Service) listContestBySeason(sid int64) (rs []*model.Contest, err error) {
	rs = make([]*model.Contest, 0)
	if err = s.dao.DB.Model(&model.Contest{}).Where("sid=?", sid).Find(&rs).Error; err != nil {
		log.Error("s.listContestBySeason sid(%d) error(%+v)", sid, err)
	}
	return
}

func (s *Service) ListGuess(c context.Context, oid int64) (rs *actmdl.GuessListAllReply, err error) {
	if _, err = s.contestExit(oid); err != nil {
		return
	}
	if rs, err = s.actClient.GuessAllList(c, &actmdl.GuessListReq{Business: int64(actmdl.GuessBusiness_esportsType), Oid: oid}); err != nil {
		log.Error("s.actClient.GuessGroupDel  oid(%d) error(%+v)", oid, err)
	}
	return
}
