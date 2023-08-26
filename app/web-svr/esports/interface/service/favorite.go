package service

import (
	"context"
	"strconv"
	"strings"
	"time"

	commonEcode "go-common/library/ecode"
	"go-common/library/log"

	"go-gateway/app/web-svr/esports/ecode"
	"go-gateway/app/web-svr/esports/interface/model"

	favpb "go-main/app/community/favorite/service/api"
	favmdl "go-main/app/community/favorite/service/model"
)

const (
	_firstPs    = 5
	_firstAppPs = 50

	delimiterOfComma = ","
	_stateOk         = 1
	_stateCancel     = 2
)

var _empStime = make([]string, 0)

func (s *Service) BatchQueryFav(ctx context.Context, mid int64, idListStr string) (m map[string]interface{}, err error) {
	idStrList, idList := convertContestIDList(idListStr)
	m = make(map[string]interface{}, 0)
	{
		m["all"] = int64(len(idStrList))
		m["total"] = len(idList)
		m["faved"] = 0
		m["all_faved"] = false
	}
	detail := make(map[int64]bool, 0)

	switch len(idList) {
	case 0:
		err = commonEcode.RequestErr
	case 1:
		arg := &favpb.IsFavoredReq{
			Typ: int32(favmdl.TypeEsports),
			Mid: mid,
			Oid: idList[0],
		}
		resp, favErr := s.favClient.IsFavored(ctx, arg)
		if favErr != nil {
			err = favErr
		} else {
			detail[idList[0]] = false
			if resp.Faved {
				m["faved"] = 1
				m["all_faved"] = true
			}

			m["detail"] = detail
		}
	default:
		arg := &favpb.IsFavoredsReq{
			Typ:  int32(favmdl.TypeEsports),
			Mid:  mid,
			Oids: idList,
		}
		resp, favErr := s.favClient.IsFavoreds(ctx, arg)
		if favErr != nil {
			err = favErr
		} else {
			var faved int64
			allFaved := true
			if len(idList) != len(resp.Faveds) {
				allFaved = false
			}

			for k, ok := range resp.Faveds {
				if ok {
					faved++
				} else {
					allFaved = false
				}

				detail[k] = ok
			}

			m["faved"] = faved
			m["all_faved"] = allFaved
			m["detail"] = detail
		}
	}

	return
}

func (s *Service) BatchAddFav(ctx context.Context, mid int64, idListStr string) (m map[string]interface{}, err error) {
	idStrList, idList := convertContestIDList(idListStr)
	effectiveCount, favCount, err := s.genBatchFavList(ctx, mid, idList)

	m = make(map[string]interface{}, 0)
	{
		m["all"] = int64(len(idStrList))
		m["total"] = effectiveCount
		m["faved"] = favCount
	}

	return
}

func convertContestIDList(idListStr string) (origin []string, rebuild []int64) {
	origin = strings.Split(idListStr, delimiterOfComma)
	rebuild = make([]int64, 0)
	for _, v := range origin {
		if d, err := strconv.ParseInt(v, 10, 64); err == nil {
			rebuild = append(rebuild, d)
		}
	}

	return
}

func genEffectiveIDList(list map[int64]*model.Contest) (idList []int64) {
	idList = make([]int64, 0)
	for k, contest := range list {
		if contest == nil || contest.ID == 0 {
			continue
		}
		if contest.LiveRoom <= 0 {
			continue
		}
		nowTime := time.Now().Unix()
		if contest.Etime > 0 && nowTime >= contest.Etime {
			continue
		}
		if contest.Stime == 0 || nowTime >= contest.Stime {
			continue
		}

		idList = append(idList, k)
	}

	return
}

func (s *Service) genBatchFavList(ctx context.Context, mid int64, idList []int64) (effectiveCount, favCount int64, err error) {
	contestM := make(map[int64]*model.Contest, 0)
	if contestM, err = s.dao.EpContests(ctx, idList); err != nil {
		return
	}

	effectiveCount = int64(len(contestM))
	favIDList := genEffectiveIDList(contestM)
	switch len(favIDList) {
	case 0:
		err = ecode.EsportsContestDataErr
	case 1:
		arg := &favpb.AddFavReq{Tp: int32(favmdl.TypeEsports), Mid: mid, Oid: favIDList[0], Fid: 0}
		_, err = s.favClient.AddFav(ctx, arg)
	default:
		arg := &favpb.MultiAddReq{
			Typ:  int32(favmdl.TypeEsports),
			Mid:  mid,
			Oids: favIDList,
			Fid:  0,
		}
		_, err = s.favClient.MultiAdd(ctx, arg)
	}

	if err == nil {
		_ = s.dao.DelFavCoCache(ctx, mid)
	}

	return
}

// AddFav add favorite contest.
func (s *Service) AddFav(c context.Context, mid, cid int64) (err error) {
	var (
		contest *model.Contest
		mapC    map[int64]*model.Contest
	)
	if mapC, err = s.dao.EpContests(c, []int64{cid}); err != nil {
		return
	}
	contest = mapC[cid]
	if contest == nil || contest.ID == 0 {
		err = ecode.EsportsContestNotExist
		return
	}
	if contest.Status == 1 {
		err = commonEcode.Errorf(commonEcode.RequestErr, "当前赛程无法订阅")
		return
	}
	if contest.LiveRoom <= 0 {
		err = ecode.EsportsContestFavNot
		return
	}
	nowTime := time.Now().Unix()
	if contest.Etime > 0 && nowTime >= contest.Etime {
		err = ecode.EsportsContestEnd
		return
	}
	if contest.Stime == 0 || nowTime >= contest.Stime {
		err = ecode.EsportsContestStart
		return
	}
	arg := &favpb.AddFavReq{Tp: int32(favmdl.TypeEsports), Mid: mid, Oid: cid, Fid: 0}
	if _, err = s.favClient.AddFav(c, arg); err != nil {
		log.Error("AddFav s.favClient.AddFav(%+v) error(%v)", arg, err)
		return
	}
	s.cache.Do(c, func(c context.Context) {
		s.dao.AsyncSendBGroupDatabus(c, mid, cid, _stateOk)
	})
	if err = s.dao.DelFavCoCache(c, mid); err != nil {
		log.Error("AddFav s.dao.DelFavCoCache mid(%d) error(%v)", mid, err)
		return
	}
	return
}

// DelFav delete favorite contest.
func (s *Service) DelFav(c context.Context, mid, cid int64) (err error) {
	var (
		contest *model.Contest
		mapC    map[int64]*model.Contest
	)
	if mapC, err = s.dao.EpContests(c, []int64{cid}); err != nil {
		return
	}
	contest = mapC[cid]
	if contest == nil || contest.ID == 0 {
		err = ecode.EsportsContestNotExist
		return
	}
	arg := &favpb.DelFavReq{Tp: int32(favmdl.TypeEsports), Mid: mid, Oid: cid, Fid: 0}
	if _, err = s.favClient.DelFav(c, arg); err != nil {
		log.Error("DelFav  s.favClient.DelFav(%+v) error(%v)", arg, err)
		return
	}
	if s.isNewBGroup(cid) {
		s.cache.Do(c, func(c context.Context) {
			s.dao.AsyncSendBGroupDatabus(c, mid, cid, _stateCancel)
		})
	} else {
		s.cache.Do(c, func(c context.Context) {
			s.dao.SendTunnelDatabus(c, mid, cid, _stateCancel)
		})
	}
	if err = s.dao.DelFavCoCache(c, mid); err != nil {
		log.Error("DelFav s.dao.DelFavCoCache mid(%d) error(%v)", mid, err)
		return
	}
	return
}

func (s *Service) isNewBGroup(contestID int64) bool {
	if s.c.TunnelBGroup.SendNew == 1 {
		return true
	}
	for _, grayID := range s.c.TunnelBGroup.NewContests {
		if grayID == contestID {
			return true
		}
	}
	return false
}

// ListFav list favorite contests.
func (s *Service) ListFav(c context.Context, mid, vmid int64, pn, ps int) (rs []*model.Contest, count int, err error) {
	var (
		isFirst    bool
		uid        int64
		favRes     *favpb.FavoritesReply
		cids       []int64
		cData      map[int64]*model.Contest
		favContest []*model.Contest
	)

	if vmid > 0 {
		uid = vmid
	} else {
		uid = mid
	}
	isFirst = pn == 1 && ps == _firstPs
	if isFirst {
		if rs, count, err = s.dao.FavCoCache(c, uid); err != nil {
			err = nil
		}
		if len(rs) > 0 {
			s.fmtContest(c, rs, mid)
			return
		}
	}
	arg := &favpb.FavoritesReq{Tp: int32(favmdl.TypeEsports), Mid: mid, Uid: vmid, Fid: 0, Pn: int32(pn), Ps: int32(ps)}
	if favRes, err = s.favClient.Favorites(c, arg); err != nil {
		log.Error("ListFav s.favClient.Favorites(%+v) error(%v)", arg, err)
		return
	}
	count = int(favRes.Res.Page.Count)
	if favRes == nil || len(favRes.Res.List) == 0 || count == 0 {
		rs = _emptContest
		return
	}
	for _, fav := range favRes.Res.List {
		cids = append(cids, fav.Oid)
	}
	if cData, err = s.dao.EpContests(c, cids); err != nil {
		log.Error("s.dao.Contest error(%v)", err)
		return
	}
	for _, fav := range favRes.Res.List {
		if contest, ok := cData[fav.Oid]; ok {
			favContest = append(favContest, contest)
		}
	}
	rs = s.ContestInfo(c, cids, favContest, mid)
	if isFirst {
		s.cache.Do(c, func(c context.Context) {
			s.dao.SetFavCoCache(c, uid, rs, count)
		})
	}
	return
}

// SeasonFav list favorite season.
func (s *Service) SeasonFav(c context.Context, mid int64, p *model.ParamSeason) (rs []*model.Season, count int, err error) {
	var (
		uid        int64
		elaContest []*model.ElaSub
		mapSeasons map[int64]*model.Season
		cids       []int64
		sids       []int64
		dbContests map[int64]*model.Contest
	)
	if p.VMID > 0 {
		uid = p.VMID
	} else {
		uid = mid
	}
	if elaContest, count, err = s.dao.SeasonFav(c, uid, p); err != nil {
		log.Error("s.dao.StimeFav error(%v)", err)
		return
	}
	for _, contest := range elaContest {
		cids = append(cids, contest.Oid)
		sids = append(sids, contest.Sid)
	}
	if len(cids) > 0 {
		if dbContests, err = s.dao.EpContests(c, cids); err != nil {
			log.Error("s.dao.EpContests error(%v)", err)
			return
		}
	} else {
		rs = _emptSeason
		return
	}
	if mapSeasons, err = s.dao.EpSeasons(c, sids); err != nil {
		log.Error("s.dao.EpSeasons error(%v)", err)
		return
	}
	ms := make(map[int64]struct{}, len(cids))
	for _, contest := range elaContest {
		if _, ok := ms[contest.Sid]; ok {
			continue
		}
		// del over contest stime.
		if contest, ok := dbContests[contest.Oid]; ok {
			if contest.Etime > 0 && time.Now().Unix() > contest.Etime {
				continue
			}
		}
		ms[contest.Sid] = struct{}{}
		if season, ok := mapSeasons[contest.Sid]; ok {
			s.ldSeasonGame.Lock()
			season.GameType = s.ldSeasonGame.Data[season.LeidaSID]
			s.ldSeasonGame.Unlock()
			rs = append(rs, season)
		}
	}
	if len(rs) == 0 {
		rs = _emptSeason
	}
	return
}

// StimeFav list favorite contests stime.
func (s *Service) StimeFav(c context.Context, mid int64, p *model.ParamSeason) (rs []string, count int, err error) {
	var (
		uid        int64
		elaContest []*model.ElaSub
		cids       []int64
		dbContests map[int64]*model.Contest
	)
	if p.VMID > 0 {
		uid = p.VMID
	} else {
		uid = mid
	}
	if elaContest, count, err = s.dao.StimeFav(c, uid, p); err != nil {
		log.Error("s.dao.StimeFav error(%v)", err)
	}
	for _, contest := range elaContest {
		cids = append(cids, contest.Oid)
	}
	if len(cids) > 0 {
		if dbContests, err = s.dao.EpContests(c, cids); err != nil {
			log.Error("s.dao.EpContests error(%v)", err)
			return
		}
	} else {
		rs = _empStime
		return
	}
	ms := make(map[string]struct{}, len(cids))
	for _, contest := range elaContest {
		tm := time.Unix(contest.Stime, 0)
		stime := tm.Format("2006-01-02")
		if _, ok := ms[stime]; ok {
			continue
		}
		ms[stime] = struct{}{}
		// del over contest stime.
		if contest, ok := dbContests[contest.Oid]; ok {
			if contest.Etime > 0 && time.Now().Unix() > contest.Etime {
				continue
			}
		}
		rs = append(rs, stime)
	}
	if len(rs) == 0 {
		rs = _empStime
	}
	return
}

// ListAppFav list favorite contests.
func (s *Service) ListAppFav(c context.Context, mid int64, p *model.ParamFav) (rs []*model.Contest, count int, err error) {
	var (
		uid        int64
		cids       []int64
		isFirst    bool
		cData      map[int64]*model.Contest
		favContest []*model.Contest
	)
	if p.VMID > 0 {
		uid = p.VMID
	} else {
		uid = mid
	}
	isFirst = p.Pn == 1 && p.Ps == _firstAppPs && p.Stime == "" && p.Etime == "" && len(p.Sids) == 0 && p.Sort == 0
	if isFirst {
		if rs, count, err = s.dao.FavCoAppCache(c, uid); err != nil {
			err = nil
		}
		if len(rs) > 0 {
			s.fmtContest(c, rs, uid)
			return
		}
	}
	if cids, count, err = s.dao.SearchFav(c, uid, p); err != nil {
		log.Error("s.dao.SearchFav error(%v)", err)
		return
	}
	if len(cids) == 0 || count == 0 {
		rs = _emptContest
		return
	}
	if cData, err = s.dao.EpContests(c, cids); err != nil {
		log.Error("s.dao.Contest error(%v)", err)
		return
	}
	for _, cid := range cids {
		if contest, ok := cData[cid]; ok {
			favContest = append(favContest, contest)
		}
	}
	rs = s.ContestInfo(c, cids, favContest, uid)
	if isFirst {
		s.cache.Do(c, func(c context.Context) {
			s.dao.SetAppFavCoCache(c, uid, rs, count)
		})
	}
	return
}

func (s *Service) isFavs(c context.Context, mid int64, cids []int64) (res map[int64]bool, err error) {
	var favRes *favpb.IsFavoredsReply
	if mid > 0 {
		if favRes, err = s.favClient.IsFavoreds(c, &favpb.IsFavoredsReq{Typ: int32(favmdl.TypeEsports), Mid: mid, Oids: cids}); err != nil {
			log.Error("s.favClient.IsFavoreds(%d,%+v) error(%+v)", mid, cids, err)
			err = nil
			return
		}
		if favRes != nil {
			res = favRes.Faveds
		}
	}
	return
}
