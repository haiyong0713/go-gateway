package service

import (
	"context"
	"fmt"
	"strconv"
	"sync/atomic"
	"time"

	"go-common/library/cache/memcache"
	"go-common/library/log"
	"go-common/library/net/netutil"
	"go-common/library/retry"
	espPb "go-gateway/app/web-svr/esports/interface/api/v1"
	"go-gateway/app/web-svr/esports/job/component"
	"go-gateway/app/web-svr/esports/job/conf"
	"go-gateway/app/web-svr/esports/job/dao"
	"go-gateway/app/web-svr/esports/job/model"
	mdlesp "go-gateway/app/web-svr/esports/job/model"
	"go-gateway/app/web-svr/esports/job/tool"
	espServicePb "go-gateway/app/web-svr/esports/service/api/v1"
)

const (
	_componentContestCardListCacheKey       = "esport:component:contestList:v2:sid:%v"
	_componentContestBattleCardListCacheKey = "esport:component:battle:card:v2:sid:%v"
	_componentGoingSeasonListCacheKey       = "esport:component:going:seasonList"
	_componentGoingBattleSeasonListCacheKey = "esport:component:going:battle:seasonList"
	_componentSeasonsCount                  = "going_seasons"
	_componentBattleSeasonsCount            = "going_battle_seasons"
	_contestStatusUpdate                    = "contest_status_update_count"
	_seasonTypeGeneral                      = 0
	_seasonTypeBattle                       = 1
)

var (
	goingSeasonsComponent       atomic.Value
	goingBattleSeasonsComponent atomic.Value
)

func componentContestCardListCacheKey(sid int64) string {
	return fmt.Sprintf(_componentContestCardListCacheKey, sid)
}

func componentContestBattleCardListCacheKey(sid int64) string {
	return fmt.Sprintf(_componentContestBattleCardListCacheKey, sid)
}

func init() {
	tmpGoingSeasons := make([]*model.Season, 0)
	goingSeasonsComponent.Store(tmpGoingSeasons)
	goingBattleSeasonsComponent.Store(tmpGoingSeasons)
}

func storeComponentGoingSeasons(ctx context.Context, seasons []*model.Season, componentCfg *conf.SeasonContestComponent) {
	if len(seasons) == 0 {
		log.Warnc(ctx, "contest component watchGoingSeasons storeComponentGoingSeasons empty")
		return
	}
	goingSeasonsComponent.Store(seasons)
	item := &memcache.Item{
		Key:        _componentGoingSeasonListCacheKey,
		Object:     seasons,
		Expiration: componentCfg.ExpiredDuration,
		Flags:      memcache.FlagJSON,
	}
	if err := retry.WithAttempts(ctx, "component_going_season_job_set_cache", 3, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
		return globalMemcache.Set(ctx, item)
	}); err != nil {
		log.Errorc(ctx, "contest component storeComponentGoingSeasons globalMemcache.Set error(%+v)", err)
	}
	tool.Metric4Component.WithLabelValues([]string{_componentSeasonsCount}...).Set(float64(len(seasons)))
}

func storeComponentGoingBattleSeasons(ctx context.Context, seasons []*model.Season, componentCfg *conf.SeasonContestComponent) {
	goingBattleSeasonsComponent.Store(seasons)
	item := &memcache.Item{
		Key:        _componentGoingBattleSeasonListCacheKey,
		Object:     seasons,
		Expiration: componentCfg.ExpiredDuration,
		Flags:      memcache.FlagJSON,
	}
	if err := retry.WithAttempts(ctx, "component_going_battle_season_job_set_cache", 3, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
		return globalMemcache.Set(ctx, item)
	}); err != nil {
		log.Errorc(ctx, "contest component storeComponentGoingBattleSeasons globalMemcache.Set error(%+v)", err)
	}
	tool.Metric4Component.WithLabelValues([]string{_componentBattleSeasonsCount}...).Set(float64(len(seasons)))
}

func loadComponentGoingSeasons() []*model.Season {
	return goingSeasonsComponent.Load().([]*model.Season)
}

func loadComponentGoingBattleSeasons() []*model.Season {
	return goingBattleSeasonsComponent.Load().([]*model.Season)
}

func WatchGoingSeasonsComponent(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 1)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			rebuildAndStoreGoingSeasons(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func rebuildAndStoreGoingSeasons(ctx context.Context) {
	cfg := conf.LoadSeasonContestComponentCfg()
	if cfg == nil {
		return
	}
	if !cfg.CanWatch {
		return
	}
	// 查询进行中的赛季
	nowTime := time.Now().Unix()
	before := nowTime + int64(time.Duration(cfg.StartTimeBefore).Seconds())
	after := nowTime - int64(time.Duration(cfg.EndTimeAfter).Seconds())
	goingSeasons, err := dao.GoingSeasons(ctx, before, after)
	if err != nil {
		log.Errorc(ctx, "contest component watchGoingSeasons error(%+v)", err)
		return
	}
	generalGoingSeasons := make([]*model.Season, 0)
	battleGoingSeasons := make([]*model.Season, 0)
	for _, season := range goingSeasons {
		switch season.SeasonType {
		case _seasonTypeGeneral:
			generalGoingSeasons = append(generalGoingSeasons, season)
		case _seasonTypeBattle:
			battleGoingSeasons = append(battleGoingSeasons, season)
		default:
			log.Errorc(ctx, "contest component watchGoingSeasons SeasonType(%+v) other", season)
		}
	}
	storeComponentGoingSeasons(ctx, generalGoingSeasons, cfg)
	storeComponentGoingBattleSeasons(ctx, battleGoingSeasons, cfg)
}

func WatchSeasonContestComponent(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 1)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			watchGoingSeasonContests(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func watchGoingSeasonContests(ctx context.Context) {
	cfg := conf.LoadSeasonContestComponentCfg()
	if cfg == nil {
		return
	}
	// 查询进行中的赛季
	goingSeasons := loadComponentGoingSeasons()
	if len(goingSeasons) == 0 {
		log.Warnc(ctx, "contest component watchGoingSeasonContests goingSeasons empty")
		return
	}
	for _, season := range goingSeasons {
		updateContestsByGoingSeason(ctx, season.ID, cfg)
	}
}

func updateContestsByGoingSeason(ctx context.Context, seasonID int64, componentCfg *conf.SeasonContestComponent) {
	if !componentCfg.CanWatch {
		return
	}
	arg := &espPb.ComponentSeasonContestListRequest{Sid: seasonID}
	componentContestListReply, err := component.EspClient.ComponentSeasonContestList(ctx, arg)
	if err != nil {
		log.Errorc(ctx, "contest component updateContestsByGoingSeason component.EspClient.ComponentSeasonContestList() seasonID(%d) error(%+v)", seasonID, err)
		return
	}
	contestListCacheKey := componentContestCardListCacheKey(seasonID)
	count := len(componentContestListReply.ComponentContestList)
	mcRes := make(map[int64][]*espPb.ContestCardComponent, count)
	for startDate, cardList := range componentContestListReply.ComponentContestList {
		mcRes[startDate] = cardList.List
	}
	item := &memcache.Item{
		Key:        contestListCacheKey,
		Object:     mcRes,
		Expiration: componentCfg.ExpiredDuration,
		Flags:      memcache.FlagJSON,
	}
	if err = retry.WithAttempts(ctx, "component_contest_list_job_set_cache", 3, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
		e := globalMemcache.Set(ctx, item)
		if e != nil {
			log.Errorc(ctx, "contest-component updateContestsByGoingSeason seasonID(%d) globalMemcache.Set error(%+v)", seasonID, e)
		}
		return e
	}); err != nil {
		log.Errorc(ctx, "contest component updateContestsByGoingSeason seasonID(%d) globalMemcache.Set error(%+v)", seasonID, err)
	}
	tool.Metric4Component.WithLabelValues([]string{strconv.FormatInt(seasonID, 10)}...).Set(float64(len(componentContestListReply.ComponentContestList)))
}

func WatchSeasonContestBattleComponent(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 1)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			watchGoingBattleSeasonContests(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func watchGoingBattleSeasonContests(ctx context.Context) {
	cfg := conf.LoadSeasonContestComponentCfg()
	if cfg == nil {
		return
	}
	// 查询进行中的赛季吃鸡类比赛
	goingSeasons := loadComponentGoingBattleSeasons()
	if len(goingSeasons) == 0 {
		log.Warnc(ctx, "contest component watchGoingBattleSeasonContests goingSeasons empty")
		return
	}
	for _, season := range goingSeasons {
		updateContestBattleByGoingSeason(ctx, season.ID, cfg)
	}
}

func updateContestBattleByGoingSeason(ctx context.Context, seasonID int64, componentCfg *conf.SeasonContestComponent) {
	if !componentCfg.CanWatch {
		return
	}
	arg := &espPb.ComponentSeasonContestBattleRequest{Sid: seasonID}
	ComponentSeasonContestBattleReply, err := component.EspClient.ComponentSeasonContestBattle(ctx, arg)
	if err != nil {
		log.Errorc(ctx, "contest component updateContestBattleByGoingSeason component.EspClient.ComponentSeasonContestList() seasonID(%d) error(%+v)", seasonID, err)
		return
	}
	contestListCacheKey := componentContestBattleCardListCacheKey(seasonID)
	count := len(ComponentSeasonContestBattleReply.ComponentContestBattle)
	mcRes := make(map[int64][]*espPb.ContestBattleCardComponent, count)
	for startDate, cardList := range ComponentSeasonContestBattleReply.ComponentContestBattle {
		mcRes[startDate] = cardList.List
	}
	item := &memcache.Item{
		Key:        contestListCacheKey,
		Object:     mcRes,
		Expiration: componentCfg.ExpiredDuration,
		Flags:      memcache.FlagJSON,
	}
	if err = retry.WithAttempts(ctx, "component_contest_battle_job_set_cache", 3, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
		e := globalMemcache.Set(ctx, item)
		if e != nil {
			log.Errorc(ctx, "contest-component updateContestBattleByGoingSeason seasonID(%d) globalMemcache.Set error(%+v)", seasonID, e)
		}
		return e
	}); err != nil {
		log.Errorc(ctx, "contest component updateContestBattleByGoingSeason seasonID(%d) globalMemcache.Set error(%+v)", seasonID, err)
	}
	tool.Metric4Component.WithLabelValues([]string{strconv.FormatInt(seasonID, 10)}...).Set(float64(len(ComponentSeasonContestBattleReply.ComponentContestBattle)))
}

func (s *Service) WatchGoingBattleSeasonsContestsTeams(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 300)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			seasons := loadComponentGoingBattleSeasons()
			for _, season := range seasons {
				log.Infoc(ctx, "[Cron][UpdateContestTeams][BySeason] Begin, seasonId:%d", season.ID)
				s.watchGoingBattleSeasonContestsTeams(ctx, season.ID)
				log.Infoc(ctx, "[Cron][UpdateContestTeams][BySeason] End, seasonId:%d", season.ID)
			}
		case <-ctx.Done():
			return
		}
	}
}

func RefreshAllContestStatusInfoLoop(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(conf.Conf.ContestStatusRefresh.RefreshDuration))
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			updateContestStatusInfo(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (s *Service) FixContestStatus(c context.Context) error {
	nowTime := time.Now().Add(time.Second)
	ctx := context.Background()
	contestList, err := fetchContestList(ctx, nowTime.Unix())
	if err != nil {
		log.Errorc(ctx, "FixContestStatus fetchFreezeContestIDs() nowTime(%d) error(%+v)", nowTime, err)
		return err
	}
	for _, contest := range contestList {
		contestStatus := contestStatusCalculate(contest.Stime, contest.Etime)
		if err = dao.UpContestStatus(ctx, contestStatus, contest.ID); err != nil {
			log.Errorc(ctx, "FixContestStatus dao.UpContestStatus() contestID(%d) error(%+v)", contest.ID, err)
		}
	}
	if err = dao.UpContestStatusDoIng(ctx); err != nil {
		log.Errorc(ctx, "FixContestStatus dao.UpContestStatusDoIng() error(%+v)", err)
	}
	return nil
}

func contestStatusCalculate(startTime int64, endTime int64) (contestStatus int64) {
	now := time.Now().Unix()
	if now < startTime {
		return _contestStatusWaiting
	}
	if now >= startTime && now < endTime {
		return _contestStatusIng
	}
	return _contestStatusOver
}

func updateContestStatusInfo(ctx context.Context) {
	if !conf.Conf.ContestStatusRefresh.RefreshSwitchDo {
		log.Warnc(ctx, "conf.Conf.ContestStatusRefresh.RefreshSwitchDo : %v", conf.Conf.ContestStatusRefresh.RefreshSwitchDo)
		return
	}
	nowTime := time.Now().Add(time.Second)
	contestIDList, err := fetchContestIDsByTime(ctx, nowTime.Unix())
	if err != nil {
		log.Errorc(ctx, "RefreshAllContestStatusInfoLoop updateContestStatusInfo fetchContestIDsByTime() nowTime(%d) error(%+v)", nowTime, err)
		return
	}
	for _, contestID := range contestIDList {
		arg := &espServicePb.RefreshContestStatusInfoReq{ContestId: contestID}
		if _, err = component.EspServiceClient.RefreshContestStatusInfo(ctx, arg); err != nil {
			log.Errorc(ctx, "ASyncResetContestStatus updateContestStatusInfo component.EspServiceClient.UpdateContestStatus() tmpContestID(%d) error(%+v)", contestID, err)
		}
	}
	tool.Metric4Component.WithLabelValues([]string{_contestStatusUpdate}...).Set(float64(len(contestIDList)))
}

func fetchContestIDsByTime(ctx context.Context, nowTime int64) (res []int64, err error) {
	if err = retry.WithAttempts(ctx, "job_update_contest_status", 3, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
		res, err = dao.ContestIDsByTime(ctx, nowTime)
		return err
	}); err != nil {
		log.Errorc(ctx, "ASyncResetContestStatus fetchContestByTime dao.ContestStatusByTime() nowTime(%d) error(%+v)", nowTime, err)
	}
	return
}

func fetchContestList(ctx context.Context, nowTime int64) (res []*mdlesp.Contest, err error) {
	if err = retry.WithAttempts(ctx, "job_update_freeze_contest_status", 3, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
		res, err = dao.GetContestListByTime(ctx, nowTime)
		return err
	}); err != nil {
		log.Errorc(ctx, "fetchContestList dao.FreezeContestIDsByTime() nowTime(%d) error(%+v)", nowTime, err)
	}
	return
}

func (s *Service) watchGoingBattleSeasonContestsTeams(ctx context.Context, seasonId int64) {
	arg := &espPb.ComponentSeasonContestBattleRequest{Sid: seasonId}
	response, err := component.EspClient.ComponentSeasonContestBattle(ctx, arg)
	if err != nil || response == nil || response.ComponentContestBattle == nil {
		log.Errorc(ctx, "[Job][Service][RebuildSeasonContestsTeams][Error] component.EspClient.ComponentSeasonContestList() seasonID(%d) error(%+v)", seasonId, err)
		return
	}

	for _, cardList := range response.ComponentContestBattle {
		contests := cardList.List
		contestIds := make([]int64, 0)
		contestInfoMap := make(map[int64]*espPb.ContestBattleCardComponent)
		for _, contest := range contests {
			contestIds = append(contestIds, contest.ID)
			contestInfoMap[contest.ID] = contest
		}
		err = s.RawContestTeamsCacheByContestIds(ctx, seasonId, contestIds, contestInfoMap)
		if err != nil {
			log.Errorc(ctx, "[Job][Service][RebuildSeasonContestsTeams][RawContestTeamsCacheByContestIds][Error] seasonID:(%d) error(%+v)", seasonId, err)
			return
		}
	}
	log.Infoc(ctx, "[Job][Service][RebuildSeasonContestsTeams][RawContestTeamsCacheByContestIds][Error] seasonID:(%d)", seasonId)
}

func (s *Service) WatchActiveSeasonInfo(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.activeSeasonRefresh(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (s *Service) WatchGamesAllInfo(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 300)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.gamesCacheRefresh(ctx)
		case <-ctx.Done():
			return
		}
	}
}
