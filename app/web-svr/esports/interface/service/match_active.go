package service

import (
	"context"
	"time"

	"go-common/library/log"
	"go-common/library/sync/errgroup"
	"go-common/library/xstr"
	arcmdl "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/esports/ecode"
	"go-gateway/app/web-svr/esports/interface/model"
	"go-gateway/pkg/idsafe/bvid"
)

const _isLive = 1

var (
	_emptActDetail = make([]*model.ActiveDetail, 0)
	_emptActLives  = make([]*model.ActiveLives, 0)
	_emptActModule = make([]*model.Module, 0)
	_emptActVideos = make([]*model.Video, 0)
	_emptTreeList  = make([][]*model.TreeList, 0)
)

// ArcsInfo archive info
func (s *Service) ArcsInfo(c context.Context, aids []int64) (arc []*model.Video, err error) {
	var (
		arcsReply *arcmdl.ArcsReply
		res       map[int64]*arcmdl.Arc
	)
	arc = _emptActVideos
	if len(aids) == 0 {
		return
	}
	if arcsReply, err = s.arcClient.Arcs(c, &arcmdl.ArcsRequest{Aids: aids}); err != nil {
		log.Error("s.arcClient.Arcs(%v) error(%v)", aids, err)
		return
	}
	if arcsReply == nil {
		return
	}
	res = make(map[int64]*arcmdl.Arc)
	for k, v := range arcsReply.Arcs {
		if v != nil && v.IsNormal() {
			res[k] = v
		}
	}
	for _, aid := range aids {
		if v, ok := res[aid]; ok {
			tmp := &model.Video{
				Arc: v,
			}
			if tmp.Bvid, err = bvid.AvToBv(v.Aid); err != nil {
				log.Error("ArcsInfo AvToBv(%v)error (%v)", v.Aid, err)
				err = nil
				continue
			}
			arc = append(arc, tmp)
		}
	}
	return
}

// ActModules matchs active videos
func (s *Service) ActModules(c context.Context, mmid int64) (res []*model.Video, err error) {
	var (
		mModule *model.Module
		aids    []int64
	)
	if res, err = s.dao.GetActModuleCache(c, mmid); err != nil || res == nil {
		if mModule, err = s.dao.Module(c, mmid); err != nil {
			return
		}
		if mModule == nil {
			err = ecode.EsportsActVideoNotExist
			return
		}
		if aids, err = xstr.SplitInts(mModule.Oids); err != nil {
			return
		}
		if res, err = s.ArcsInfo(c, aids); err != nil {
			return
		}
		if res == nil {
			res = _emptActVideos
			return
		}
		s.cache.Do(c, func(c context.Context) {
			s.dao.AddActModuleCache(c, mmid, res)
		})
	}
	return
}

// MatchAct match act
func (s *Service) MatchAct(c context.Context, aid int64) (act *model.Active, err error) {
	if act, err = s.dao.GetMActCache(c, aid); err != nil || act == nil {
		if act, err = s.dao.Active(c, aid); err != nil {
			log.Error("s.dao.Active error(%v)", err)
			return
		}
		if act != nil {
			s.cache.Do(c, func(c context.Context) {
				s.dao.AddMActCache(c, aid, act)
			})
		}
	}
	return
}

// ActPage matchs active page info
func (s *Service) ActPage(c context.Context, aid, tp int64) (res *model.ActivePage, err error) {
	var (
		act                                         *model.Active
		modules                                     []*model.Module
		aids                                        []int64
		actDetail                                   []*model.ActiveDetail
		actLives                                    []*model.ActiveLives
		videos                                      []*model.Video
		moduleErr, detailError, actError, liveError error
		liveInfo                                    *model.ActiveLive
	)
	if tp == _isLive {
		if liveInfo, err = s.activeByLive(c, aid); err != nil || liveInfo == nil || liveInfo.MaID == 0 {
			return
		}
		aid = liveInfo.MaID
	}
	if res, err = s.dao.GetActPageCache(c, aid); err != nil || res == nil {
		res = &model.ActivePage{}
		group, errCtx := errgroup.WithContext(c)
		group.Go(func() error {
			if act, actError = s.MatchAct(errCtx, aid); actError != nil {
				log.Error("s.dao.Active error(%v)", moduleErr)
			}
			return actError
		})
		group.Go(func() error {
			if modules, moduleErr = s.dao.Modules(errCtx, aid); moduleErr != nil {
				log.Error("s.dao.Modules error(%v)", moduleErr)
				return nil
			}
			if len(modules) > 0 && modules[0].Oids != "" {
				if aids, moduleErr = xstr.SplitInts(modules[0].Oids); moduleErr != nil {
					log.Error("s.ActPage.SplitInts oids error(%v) error(%v)", modules[0].Oids, moduleErr)
					return nil
				}
				if videos, moduleErr = s.ArcsInfo(c, aids); moduleErr != nil {
					log.Error("s.ActPage.ArcsInfo aids(%v) error(%v)", aids, moduleErr)
				}
			}
			return nil
		})
		group.Go(func() error {
			if actDetail, detailError = s.dao.ActDetail(errCtx, aid); detailError != nil {
				log.Error("s.dao.ActDetail error(%v)", detailError)
			}
			return nil
		})
		group.Go(func() error {
			if actLives, liveError = s.dao.ActLives(errCtx, aid); liveError != nil {
				log.Error("s.dao.ActLives error(%v)", liveError)
			}
			return nil
		})
		err = group.Wait()
		if err != nil {
			return
		}
		if act == nil {
			err = ecode.EsportsActNotExist
			return
		}
		res.Active = act
		if len(actDetail) == 0 {
			res.ActiveDetail = _emptActDetail
		} else {
			res.ActiveDetail = actDetail
		}
		if len(actLives) == 0 {
			res.ActiveLives = _emptActLives
		} else {
			res.ActiveLives = actLives
		}
		if len(modules) == 0 {
			res.Modules = _emptActModule
		} else {
			res.Modules = modules
		}
		if len(videos) == 0 {
			res.Videos = _emptActVideos
		} else {
			res.Videos = videos
		}
		s.cache.Do(c, func(c context.Context) {
			s.dao.AddActPageCache(c, aid, res)
		})
	}
	return
}

// ContestCommon act match same deal logic
func (s *Service) ContestCommon(c context.Context, mid int64, p *model.ParamContest) (rs []*model.Contest, total int, err error) {
	var (
		tmpRs      []*model.Contest
		cids       []int64
		dbContests map[int64]*model.Contest
	)
	if cids, total, err = s.dao.SearchContestQuery(c, p); err != nil {
		log.Error("s.dao.SearchContest error(%v)", err)
	}
	if total == 0 || len(cids) == 0 {
		rs = _emptContest
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
	rs = s.ContestInfo(c, cids, tmpRs, mid)
	return
}

// ActTop act match top data
func (s *Service) ActTop(c context.Context, mid int64, param *model.ParamActTop) (res []*model.Contest, total int, err error) {
	var (
		act          *model.Active
		sids         []int64
		sTime, eTime time.Time
		liveInfo     *model.ActiveLive
	)
	isFirst := param.Pn == 1 && param.Sort == 0 && param.Stime == "" && param.Etime == "" && param.Tp == 0
	if isFirst {
		if res, total, err = s.dao.GetActTopCache(c, param.Aid, int64(param.Ps)); err != nil {
			err = nil
		} else if len(res) > 0 {
			s.fmtContest(c, res, mid)
			return
		}
	}
	aid := param.Aid
	if param.Tp == _isLive {
		if liveInfo, err = s.activeByLive(c, aid); err != nil || liveInfo == nil || liveInfo.MaID == 0 {
			return
		}
		aid = liveInfo.MaID
	}
	if act, err = s.MatchAct(c, aid); err != nil {
		log.Error("s.MatchAct error(%v)", err)
		return
	}
	if act == nil {
		err = ecode.EsportsActNotExist
		return
	}
	if act.Sids == "" && act.Sid > 0 {
		sids = []int64{act.Sid}
	} else {
		if sids, err = xstr.SplitInts(act.Sids); err != nil {
			log.Error("xstr.SplitInts error(%v)", err)
			return
		}
	}
	if len(sids) == 0 {
		err = ecode.EsportsActNotExist
		return
	}
	if param.Stime != "" {
		sTime, _ = time.ParseInLocation("2006-01-02", param.Stime, time.Local)
		param.Stime = time.Unix(sTime.Unix(), 0).Format("2006-01-02") + " 00:00:00"
	}
	if param.Etime != "" {
		eTime, _ = time.ParseInLocation("2006-01-02", param.Etime, time.Local)
		param.Etime = time.Unix(eTime.Unix(), 0).Format("2006-01-02") + " 23:59:59"
	}
	p := &model.ParamContest{
		Mid:   act.Mid,
		Sids:  sids,
		Sort:  param.Sort,
		Stime: param.Stime,
		Etime: param.Etime,
		Pn:    param.Pn,
		Ps:    param.Ps,
	}
	if res, total, err = s.ContestCommon(c, mid, p); err != nil {
		return
	}
	if len(res) == 0 {
		res = _emptContest
		return
	}
	if isFirst {
		s.cache.Do(c, func(c context.Context) {
			s.dao.AddActTopCache(c, param.Aid, int64(param.Ps), res, total)
		})
	}
	return
}

// ActPoints act match point match
func (s *Service) ActPoints(c context.Context, mid int64, param *model.ParamActPoint) (res []*model.Contest, total int, err error) {
	var (
		act      *model.Active
		detail   *model.ActiveDetail
		liveInfo *model.ActiveLive
		sids     []int64
	)
	aid := param.Aid
	if param.Tp == _isLive {
		if liveInfo, err = s.activeByLive(c, aid); err != nil || liveInfo == nil || liveInfo.MaID == 0 {
			return
		}
		aid = liveInfo.MaID
	}
	isFirst := param.Pn == 1 && param.Sort == 0
	if isFirst {
		if res, total, err = s.dao.GetActPointsCache(c, aid, param.MdID, int64(param.Ps)); err != nil {
			err = nil
		} else if len(res) > 0 {
			s.fmtContest(c, res, mid)
			return
		}
	}
	if act, err = s.MatchAct(c, aid); err != nil {
		log.Error("s.dao.Active error(%v)", err)
		return
	}
	if act == nil {
		err = ecode.EsportsActNotExist
		return
	}
	if act.Sids == "" && act.Sid > 0 {
		sids = []int64{act.Sid}
	} else {
		if sids, err = xstr.SplitInts(act.Sids); err != nil {
			log.Error("xstr.SplitInts error(%v)", err)
			return
		}
	}
	p := &model.ParamContest{
		Mid:  act.Mid,
		Sids: sids,
		Sort: param.Sort,
		Pn:   param.Pn,
		Ps:   param.Ps,
	}
	detail, err = s.dao.PActDetail(c, param.MdID)
	if err != nil {
		log.Error("s.dao.PActDetail error(%v)", err)
		return
	}
	if detail != nil {
		if detail.STime != 0 {
			p.Stime = time.Unix(detail.STime, 0).Format("2006-01-02 15:04:05")
		}
		if detail.ETime != 0 {
			p.Etime = time.Unix(detail.ETime, 0).Format("2006-01-02 15:04:05")
		}
	} else {
		err = ecode.EsportsActPointNotExist
		return
	}
	if res, total, err = s.ContestCommon(c, mid, p); err != nil {
		return
	}
	if len(res) == 0 {
		res = _emptContest
		return
	}
	if isFirst {
		s.cache.Do(c, func(c context.Context) {
			s.dao.AddActPointsCache(c, aid, param.MdID, int64(param.Ps), res, total)
		})
	}
	return
}

// contestTeam get contest team
func (s *Service) contestTeam(contests map[int64]*model.Contest, teams map[int64]*model.Filter) (list []*model.ContestInfo) {
	for _, v := range contests {
		contest := &model.ContestInfo{Contest: v}
		if v.HomeID > 0 {
			if team, ok := teams[v.HomeID]; ok {
				contest.HomeName = team.Title
				contest.HomeTeam = team
			} else {
				contest.HomeName = ""
				contest.HomeTeam = struct{}{}
			}
		} else {
			contest.HomeName = ""
			contest.HomeTeam = struct{}{}
		}
		if v.AwayID > 0 {
			if team, ok := teams[v.AwayID]; ok {
				contest.AwayName = team.Title
				contest.AwayTeam = team
			} else {
				contest.AwayName = ""
				contest.AwayTeam = struct{}{}
			}
		} else {
			contest.AwayName = ""
			contest.AwayTeam = struct{}{}
		}
		if v.SuccessTeam > 0 {
			if team, ok := teams[v.SuccessTeam]; ok {
				contest.SuccessName = team.Title
			} else {
				contest.SuccessName = ""
			}
		} else {
			contest.SuccessName = ""
		}
		contest.Season = struct{}{}
		contest.SuccessTeaminfo = struct{}{}
		list = append(list, contest)
	}
	return
}

// TeamMap get team and map team id to team
func (s *Service) TeamMap(ctx context.Context) (res map[int64]*model.Filter, err error) {
	var (
		teams []*model.Filter
	)
	if teams, err = s.dao.Teams(ctx); err != nil {
		log.Error("TeamMap s.dao.Teams error(%v)", err)
		return
	}
	res = make(map[int64]*model.Filter)
	for _, v := range teams {
		res[v.ID] = v
	}
	return
}

func (s *Service) buildKnockTree() {
	var (
		trees       []*model.Tree
		sTree       []*model.TreeList
		mids        []int64
		contests    map[int64]*model.Contest
		cInfos      []*model.ContestInfo
		mapContests map[int64]*model.ContestInfo
		treeList    [][]*model.TreeList
		err         error
		details     []*model.ActiveDetail
		teamMap     map[int64]*model.Filter
		c           = context.Background()
	)
	if teamMap, err = s.TeamMap(c); err != nil {
		log.Error("buildKnockTree TeamMap error(%v)", err)
		return
	}
	if details, err = s.dao.KDetails(c); err != nil {
		log.Error("buildKnockTree s.dao.KDetails error(%v)", err)
		return
	}
	//build tree
	for _, detail := range details {
		if detail.Online == _downline {
			//edit status; must add cache time
			if s.dao.AddActKnockCacheTime(c, detail.ID); err != nil {
				log.Error("buildKnockTree s.dao.AddActKnockCacheTime error(%v)", err)
				continue
			}
		} else {
			treeList = _emptTreeList
			sTree = make([]*model.TreeList, 0)
			if trees, err = s.dao.Trees(c, detail.ID); err != nil {
				log.Error("s.dao.Trees error(%v)", err)
				return
			}
			for _, tree := range trees {
				mids = append(mids, tree.Mid)
			}
			if len(mids) == 0 {
				continue
			} else {
				count := len(mids)
				if contests, err = s.dao.EpContests(c, mids); err != nil {
					log.Error("s.dao.RawEpContests error(%v)", err)
					continue
				}
				cInfos = s.contestTeam(contests, teamMap)
				mapContests = make(map[int64]*model.ContestInfo, count)
				for _, info := range cInfos {
					mapContests[info.ID] = info
				}
				for _, tree := range trees {
					if tree.Pid == 0 {
						if len(sTree) > 0 {
							treeList = append(treeList, sTree)
						}
						sTree = nil
						if cInfo, ok := mapContests[tree.Mid]; ok {
							sTree = append(sTree, &model.TreeList{Tree: tree, ContestInfo: cInfo})
						} else {
							sTree = append(sTree, &model.TreeList{Tree: tree})
						}
					} else {
						if cInfo, ok := mapContests[tree.Mid]; ok {
							sTree = append(sTree, &model.TreeList{Tree: tree, ContestInfo: cInfo})
						} else {
							sTree = append(sTree, &model.TreeList{Tree: tree})
						}
					}
				}
				if len(sTree) > 0 {
					treeList = append(treeList, sTree)
				}
				if len(treeList) > 0 {
					go s.dao.AddActKnockoutCache(context.Background(), detail.ID, treeList)
				}
			}
		}
	}
}

// ActKnockout knockout tree
func (s *Service) ActKnockout(c context.Context, madID int64) (res [][]*model.TreeList, err error) {
	if res, err = s.dao.GetActKnockoutCache(c, madID); err != nil {
		return
	}
	if len(res) == 0 {
		res = _emptTreeList
		err = ecode.EsportsActKnockNotExist
	}
	return
}

func (s *Service) activeByLive(c context.Context, liveID int64) (rs *model.ActiveLive, err error) {
	if rs, err = s.dao.GetLiveCache(c, liveID); err != nil || rs == nil {
		if rs, err = s.dao.LiveInfo(c, liveID); err != nil || rs == nil || rs.MaID == 0 {
			err = ecode.EsportsActNotExist
			return
		}
		s.cache.Do(c, func(c context.Context) {
			s.dao.AddLiveCache(c, liveID, rs)
		})
	}
	return
}
