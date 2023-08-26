package service

import (
	"context"
	"fmt"
	"time"

	"go-common/library/cache/memcache"
	commonECode "go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/netutil"
	"go-common/library/retry"
	"go-gateway/app/web-svr/esports/ecode"
	"go-gateway/app/web-svr/esports/interface/component"
	"go-gateway/app/web-svr/esports/interface/dao"
	"go-gateway/app/web-svr/esports/interface/model"
	"go-gateway/app/web-svr/esports/interface/tool"
	favpb "go-main/app/community/favorite/service/api"
	favmdl "go-main/app/community/favorite/service/model"
)

const (
	autoSubStatusOfNotSubscribed = false
	autoSubStatusOfSubscribed    = true

	cacheKKey4AutoSub      = "autoSub:%v:%v:%v"
	cacheKey4SubSeasonTeam = "autoSub_season_teams:%v:%v"

	bizLimitKey4AutoSub     = "auto_sub"
	bizLimitKey4AutoSubFind = "auto_sub_find"

	seconds4TenHours       = 10 * 60 * 60
	ecodeCode4HasBeenFaved = 11201
)

var (
	autoSubMap map[string][]int64
)

func init() {
	autoSubMap = make(map[string][]int64)
}

func asyncAutoSubData(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			genAutoSubscribeMap(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func genAutoSubscribeMap(ctx context.Context) {
	if seasonIDList, err := dao.FetchAutoSubSeasonList(ctx); err == nil {
		if tmpM, err := dao.FetchAutoSubSeasonTeamContestIDMap(ctx, seasonIDList); err == nil {
			autoSubMap = tmpM
		}
	}
}

func FetchAutoSubscribeMap(ctx context.Context, mid int64, req *model.AutoSubRequest) (res map[int64]bool, err error) {
	res, err = AutoSubscribeFromCache(ctx, mid, req.SeasonID)
	if err == nil {
		return
	}
	if err == memcache.ErrNotFound {
		if tool.IsLimiterAllowedByUniqBizKey(bizLimitKey4AutoSubFind, bizLimitKey4AutoSubFind) {
			if res, err = dao.FetchAutoSubDetail(ctx, mid, req); err != nil {
				err = commonECode.ServiceUnavailable
				return
			}
			err = AutoSubscribeToCache(ctx, mid, req.SeasonID, res) // 设置缓存.
		} else {
			err = commonECode.ServiceUnavailable
			return
		}
	}
	return
}

func AutoSubscribeFromCache(ctx context.Context, mid, seasonID int64) (res map[int64]bool, err error) {
	res = make(map[int64]bool, 0)
	cacheKey := seasonSubTeamsCacheKey(mid, seasonID)
	err = component.GlobalMemcached4UserGuess.Get(ctx, cacheKey).Scan(&res)
	return
}

func AutoSubscribeToCache(ctx context.Context, mid, seasonID int64, subMap map[int64]bool) (err error) {
	if len(subMap) == 0 {
		subMap = make(map[int64]bool, 1)
		subMap[0] = false // 空缓存
	}
	cacheKey := seasonSubTeamsCacheKey(mid, seasonID)
	item := &memcache.Item{
		Key:        cacheKey,
		Object:     subMap,
		Expiration: int32(tool.CalculateExpiredSeconds(10)),
		Flags:      memcache.FlagJSON,
	}
	if err = retry.WithAttempts(ctx, "auto_subscribe_set_cache", 3, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
		return component.GlobalMemcached4UserGuess.Set(ctx, item)
	}); err != nil {
		log.Errorc(ctx, "AutoSubscribe AutoSubscribeToCache mid(%d) seasonID(%d) GlobalMemcached4UserGuess.Set error(%+v)", mid, seasonID, err)
		return err
	}
	return
}

func DeleteAutoSubscribeCache(ctx context.Context, mid, seasonID int64) (err error) {
	cacheKey := seasonSubTeamsCacheKey(mid, seasonID)
	if err = retry.WithAttempts(ctx, "auto_subscribe_del_cache", 3, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
		return component.GlobalMemcached4UserGuess.Delete(ctx, cacheKey)
	}); err != nil {
		log.Errorc(ctx, "AutoSubscribe DeleteAutoSubscribeCache mid(%d) seasonID(%d) GlobalMemcached4UserGuess.Delete error(%+v)", mid, seasonID, err)
		return err
	}
	return
}

func (s *Service) AutoSubscribe(ctx context.Context, mid int64, req *model.AutoSubRequest) error {
	//autoSubKeyList := make([]string, 0)
	contestIDList := make([]int64, 0)
	for _, v := range req.TeamIDList {
		autoSubKey := dao.GenAutoSubUniqKey(req.SeasonID, v)
		if d, ok := autoSubMap[autoSubKey]; ok {
			contestIDList = append(contestIDList, d...)
			//autoSubKeyList = append(autoSubKeyList, autoSubKey)
		}
	}
	//if len(autoSubKeyList) == 0 {   // 上B站看电竞会有无赛程战队
	//	return commonECode.RequestErr
	//}
	subMap, err := s.fetchAutoSubStatus(ctx, mid, req)
	if err != nil {
		log.Errorc(ctx, "AutoSubscribe s.fetchAutoSubStatus() mid(%d) req(%+v) error(%+v)", mid, req, err)
		return err
	}
	needSub := false
	for _, subValue := range subMap {
		if subValue == autoSubStatusOfNotSubscribed {
			needSub = true
			break
		}
	}
	if !needSub {
		return ecode.EsportsAutoSubed
	}
	return s.AutoSubscribeSet(ctx, mid, contestIDList, req)
}

func (s *Service) AutoSubscribeSet(ctx context.Context, mid int64, contestIDList []int64, req *model.AutoSubRequest) (err error) {
	contestIDList = tool.Unique(contestIDList) // 去重.
	if len(contestIDList) > 0 {
		arg := &favpb.MultiAddReq{
			Typ:  int32(favmdl.TypeEsports),
			Mid:  mid,
			Oids: contestIDList,
			Fid:  0,
		}
		if _, err := s.favClient.MultiAdd(ctx, arg); err != nil {
			if commonECode.Cause(err).Code() != ecodeCode4HasBeenFaved {
				return err
			}
		}
		tmpErr := s.cache.Do(ctx, func(ctx context.Context) {
			s.batchSendDatabusBGroup(ctx, mid, contestIDList)
		})
		if tmpErr != nil {
			log.Errorc(ctx, "AutoSubscribeSet s.cache.Do s.batchSendDatabusBGroup error(%+v)", tmpErr)
		}
	}

	if tool.IsLimiterAllowedByUniqBizKey(bizLimitKey4AutoSub, bizLimitKey4AutoSub) {
		if err = dao.AutoSubscribeDetail(ctx, mid, req); err != nil {
			log.Errorc(ctx, "AutoSubscribe dao.AutoSubscribeDetail() mid(%d) seasonID(%+v) error(%+v)", mid, req.SeasonID, err)
			return commonECode.ServiceUnavailable
		}
		if e := DeleteAutoSubscribeCache(ctx, mid, req.SeasonID); e != nil {
			log.Errorc(ctx, "AutoSubscribe DeleteAutoSubscribeCache() mid(%d) seasonID(%+v) error(%+v)", mid, req.SeasonID, e)
		}
		return nil
	}
	return commonECode.LimitExceed
}

func (s *Service) AutoSubscribeStatus(ctx context.Context, mid int64, req *model.AutoSubRequest) (map[int64]bool, error) {
	return s.fetchAutoSubStatus(ctx, mid, req)
}

func (s *Service) fetchAutoSubStatus(ctx context.Context, mid int64, req *model.AutoSubRequest) (m map[int64]bool, err error) {
	m = make(map[int64]bool, 0)
	subMap, tmpErr := FetchAutoSubscribeMap(ctx, mid, req)
	if tmpErr != nil {
		err = tmpErr
		log.Errorc(ctx, "AutoSubscribeStatus fetchAutoSubStatus mid(%d) req(%+v) error(%+v)", mid, req, err)
		return
	}
	for _, v := range req.TeamIDList {
		if isSub, ok := subMap[v]; ok && isSub {
			m[v] = autoSubStatusOfSubscribed
		} else {
			m[v] = autoSubStatusOfNotSubscribed
		}
	}
	return
}

// Only set expired time in nex day 0:00AM ~ 10:00AM
//func calculateExpiredSeconds() int64 {
//	now := time.Now()
//	year, month, day := now.Date()
//	nextDay := time.Date(year, month, day, 24, 0, 0, 0, now.Location()).Unix()
//	rand.Seed(time.Now().UnixNano())
//	randSeconds := rand.Int63n(seconds4TenHours)
//	return nextDay + randSeconds - now.Unix()
//}

func genAutoSubCacheKey(seasonID, teamID, mid int64) string {
	return fmt.Sprintf(cacheKKey4AutoSub, seasonID, teamID, mid)
}

func seasonSubTeamsCacheKey(mid, seasonID int64) string {
	return fmt.Sprintf(cacheKey4SubSeasonTeam, mid, seasonID)
}

func (s *Service) batchSendDatabusBGroup(c context.Context, mid int64, contestIDs []int64) {
	for _, cid := range contestIDs {
		if s.isNewBGroup(cid) {
			if err := s.dao.AsyncSendBGroupDatabus(c, mid, cid, _stateOk); err != nil {
				log.Errorc(c, "batchSendDatabusBGroup mid(%d) contestsIDs(%+v) error(%+v)", mid, contestIDs, err)
			}
		} else {
			if err := s.dao.SendTunnelDatabus(c, mid, cid, _stateOk); err != nil {
				log.Errorc(c, "SendTunnelDatabus mid(%d) contestsIDs(%+v) error(%+v)", mid, contestIDs, err)
			}
		}
	}
}
