package service

import (
	"context"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"go-common/library/cache/memcache"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	actApi "go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/esports/common/helper"
	"go-gateway/app/web-svr/esports/ecode"
	pb "go-gateway/app/web-svr/esports/interface/api/v1"
	"go-gateway/app/web-svr/esports/interface/conf"
	"go-gateway/app/web-svr/esports/interface/dao/match_component"
	"go-gateway/app/web-svr/esports/interface/model"
	"go-gateway/app/web-svr/esports/interface/tool"
	v1 "go-gateway/app/web-svr/esports/service/api/v1"
)

const (
	teamRegionIDOfNull = iota
	teamRegionIDOfChina
	teamRegionIDOfChinaTaiWan
	teamRegionDisplayOfNull        = "无"
	teamRegionDisplayOfChina       = "中国赛区"
	teamRegionDisplayOfChinaTaiWan = "中国台湾赛区"
)

const (
	_contestIDsBulkSize  = 100
	_esportsBusiness     = 1
	_guessMatchFirst     = 1
	_haveSubscribe       = 1
	_haveGuess           = 1
	_findAsc             = 0
	_findDesc            = 1
	_videoListPs         = 50
	_historyAbstractSize = 10
	_futureAbstractSize  = 15
)

var (
	seasonContestTimeComponent0Map map[int64]map[int64][]*pb.ContestCardComponent
	seasonContestTimeComponent1Map map[int64]map[int64][]*pb.ContestCardComponent
	seasonContestTimeComponent2Map map[int64]map[int64][]*pb.ContestCardComponent
	seasonContestTimeComponent3Map map[int64]map[int64][]*pb.ContestCardComponent
	seasonContestTimeComponent4Map map[int64]map[int64][]*pb.ContestCardComponent
	seasonContestTimeComponent5Map map[int64]map[int64][]*pb.ContestCardComponent
	seasonContestTimeComponent6Map map[int64]map[int64][]*pb.ContestCardComponent
	seasonContestTimeComponent7Map map[int64]map[int64][]*pb.ContestCardComponent
	seasonContestTimeComponent8Map map[int64]map[int64][]*pb.ContestCardComponent
	seasonContestTimeComponent9Map map[int64]map[int64][]*pb.ContestCardComponent

	seasonContestTimeComponentTsList0Map map[int64][]int64
	seasonContestTimeComponentTsList1Map map[int64][]int64
	seasonContestTimeComponentTsList2Map map[int64][]int64
	seasonContestTimeComponentTsList3Map map[int64][]int64
	seasonContestTimeComponentTsList4Map map[int64][]int64
	seasonContestTimeComponentTsList5Map map[int64][]int64
	seasonContestTimeComponentTsList6Map map[int64][]int64
	seasonContestTimeComponentTsList7Map map[int64][]int64
	seasonContestTimeComponentTsList8Map map[int64][]int64
	seasonContestTimeComponentTsList9Map map[int64][]int64

	seasonContestAllComponent0Map map[int64][]*pb.ContestCardComponent
	seasonContestAllComponent1Map map[int64][]*pb.ContestCardComponent
	seasonContestAllComponent2Map map[int64][]*pb.ContestCardComponent
	seasonContestAllComponent3Map map[int64][]*pb.ContestCardComponent
	seasonContestAllComponent4Map map[int64][]*pb.ContestCardComponent
	seasonContestAllComponent5Map map[int64][]*pb.ContestCardComponent
	seasonContestAllComponent6Map map[int64][]*pb.ContestCardComponent
	seasonContestAllComponent7Map map[int64][]*pb.ContestCardComponent
	seasonContestAllComponent8Map map[int64][]*pb.ContestCardComponent
	seasonContestAllComponent9Map map[int64][]*pb.ContestCardComponent

	// battleground.
	seasonContestBattleAllComponentMap map[int64][]*pb.ContestBattleCardComponent
	goingVideoListComponentMap         map[int64]*model.VideoList2Component
	goingBattleSeasonsListGlobal       []*model.ComponentSeason // battle season list
	goingSeasonsContestsTeams          sync.Map

	goingSeasonsListGlobal     []*model.ComponentSeason // season list
	allTeamsOfComponent        atomic.Value             // all teams
	allSeasonsOfComponent      atomic.Value             // all seasons
	_emptyVideoComponent       = make([]*model.Video, 0)
	_emptyContestCardComponent = make([]*pb.ContestCardComponent, 0)
)

func init() {
	seasonContestTimeComponent0Map = make(map[int64]map[int64][]*pb.ContestCardComponent)
	seasonContestTimeComponent1Map = make(map[int64]map[int64][]*pb.ContestCardComponent)
	seasonContestTimeComponent2Map = make(map[int64]map[int64][]*pb.ContestCardComponent)
	seasonContestTimeComponent3Map = make(map[int64]map[int64][]*pb.ContestCardComponent)
	seasonContestTimeComponent4Map = make(map[int64]map[int64][]*pb.ContestCardComponent)
	seasonContestTimeComponent5Map = make(map[int64]map[int64][]*pb.ContestCardComponent)
	seasonContestTimeComponent6Map = make(map[int64]map[int64][]*pb.ContestCardComponent)
	seasonContestTimeComponent7Map = make(map[int64]map[int64][]*pb.ContestCardComponent)
	seasonContestTimeComponent8Map = make(map[int64]map[int64][]*pb.ContestCardComponent)
	seasonContestTimeComponent9Map = make(map[int64]map[int64][]*pb.ContestCardComponent)

	seasonContestTimeComponentTsList0Map = make(map[int64][]int64)
	seasonContestTimeComponentTsList1Map = make(map[int64][]int64)
	seasonContestTimeComponentTsList2Map = make(map[int64][]int64)
	seasonContestTimeComponentTsList3Map = make(map[int64][]int64)
	seasonContestTimeComponentTsList4Map = make(map[int64][]int64)
	seasonContestTimeComponentTsList5Map = make(map[int64][]int64)
	seasonContestTimeComponentTsList6Map = make(map[int64][]int64)
	seasonContestTimeComponentTsList7Map = make(map[int64][]int64)
	seasonContestTimeComponentTsList8Map = make(map[int64][]int64)
	seasonContestTimeComponentTsList9Map = make(map[int64][]int64)

	seasonContestAllComponent0Map = make(map[int64][]*pb.ContestCardComponent)
	seasonContestAllComponent1Map = make(map[int64][]*pb.ContestCardComponent)
	seasonContestAllComponent2Map = make(map[int64][]*pb.ContestCardComponent)
	seasonContestAllComponent3Map = make(map[int64][]*pb.ContestCardComponent)
	seasonContestAllComponent4Map = make(map[int64][]*pb.ContestCardComponent)
	seasonContestAllComponent5Map = make(map[int64][]*pb.ContestCardComponent)
	seasonContestAllComponent6Map = make(map[int64][]*pb.ContestCardComponent)
	seasonContestAllComponent7Map = make(map[int64][]*pb.ContestCardComponent)
	seasonContestAllComponent8Map = make(map[int64][]*pb.ContestCardComponent)
	seasonContestAllComponent9Map = make(map[int64][]*pb.ContestCardComponent)

	seasonContestBattleAllComponentMap = make(map[int64][]*pb.ContestBattleCardComponent)
	goingVideoListComponentMap = make(map[int64]*model.VideoList2Component)
	goingBattleSeasonsListGlobal = make([]*model.ComponentSeason, 0)
	goingSeasonsContestsTeams = sync.Map{}

	goingSeasonsListGlobal = make([]*model.ComponentSeason, 0)
	tmpSeasonTeam := make(map[int64]*model.Team2TabComponent)
	allTeamsOfComponent.Store(tmpSeasonTeam)
	tmpSeason := make(map[int64]*model.SeasonComponent)
	allSeasonsOfComponent.Store(tmpSeason)
}

func sidSharding(sid int64) int64 {
	return sid % 10
}

func watchSeasonContestBiz(ctx context.Context) {
	//go watchComponentAllTeamsMap(ctx)
	//go watchGoingSeasonList(ctx)
	//go watchGoingBattleSeasonList(ctx)
	//go watchComponentContestList(ctx)
	//go watchComponentContestBattle(ctx)
	goroutineRegister(watchComponentAllTeamsMap)
	goroutineRegister(watchComponentAllSeasonsMap)
	goroutineRegister(watchGoingSeasonList)
	goroutineRegister(watchGoingBattleSeasonList)
	goroutineRegister(watchComponentContestList)
	goroutineRegister(watchComponentContestBattle)
}

func (s *Service) watchGoingBattleSeasonsContestsTeamsTimeTicker(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 3)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.watchGoingBattleSeasonsContestsTeams(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (s *Service) watchGoingLolDataHero2TimeTicker(ctx context.Context) {
	ticker := time.NewTicker(time.Minute * 10)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.watchLolDataByGoingSeason(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (s *Service) watchGoingVideoListComponent(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 3)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.rebuildGoingVideoList(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (s *Service) rebuildGoingVideoList(ctx context.Context) {
	// 查询进行中的视频列表
	nowTime := time.Now()
	before := nowTime.Add(30 * 24 * time.Hour).Unix()
	after := nowTime.Add(-30 * 24 * time.Hour).Unix()
	goingVideoList, err := match_component.GoingVideoList(ctx, before, after)
	if err != nil {
		log.Errorc(ctx, "contest component rebuildGoingVideoList match_component.GoingVideoList error(%+v)", err)
		return
	}
	count := len(goingVideoList)
	if count == 0 {
		log.Warnc(ctx, "contest component rebuildGoingVideoList goingVideoList empty")
		return
	}
	tmpVideoListMap := make(map[int64]*model.VideoList2Component, count)
	for _, videoListInfo := range goingVideoList {
		arg := &pb.EsTopicVideoListRequest{
			GameId:  videoListInfo.GameID,
			MatchId: videoListInfo.MatchID,
			YearId:  videoListInfo.YearID,
			Pn:      _firstPage,
			Ps:      _videoListPs,
		}
		firstPageVideoList, e := s.rebuildTopicVideoList(ctx, videoListInfo.UgcAids, arg)
		if e != nil {
			log.Errorc(ctx, "contest component rebuildGoingVideoList s.rebuildTopicVideoList videoList(%+v) error(%+v)", videoListInfo, e)
			continue
		}
		tmpVideoListMap[videoListInfo.ID] = firstPageVideoList
	}
	goingVideoListComponentMap = tmpVideoListMap
}

func goroutineRegister(f func(ctx context.Context)) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Error("[Async][GoRoutine][Run][Panic][Recover]err:(%+v)", err)
				return
			}
		}()
		f(context.Background())
	}()
}

func watchGoingSeasonList(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 3)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			watchGoingSeasonsByCacheWatch(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func watchGoingBattleSeasonList(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 3)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			watchGoingBattleSeasonsByCacheWatch(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func watchComponentAllTeamsMap(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 3)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			watchComponentAllTeamsWatch(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func watchComponentAllSeasonsMap(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 3)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			watchComponentAllSeasonsWatch(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (s *Service) watchGoingBattleSeasonsContestsTeams(ctx context.Context) {
	// 吃鸡类赛事赛程战队列表实例级缓存的必要性判断
	cfg := conf.LoadSeasonContestComponentWatch()
	if cfg == nil || !cfg.CanWatch {
		return
	}
	for seasonId, contestList := range seasonContestBattleAllComponentMap {
		if contestList == nil || len(contestList) == 0 {
			continue
		}
		teamsInfoMap, err := s.GetTeamsInfoBySeasonContestsSkipLocalCache(ctx, seasonId, contestList)
		if err != nil {
			log.Errorc(ctx, "[Service][TimeTicker][watchGoingBattleSeasonsContestsTeams][Error], err(%+v)", err)
			return
		}
		for contestId, teamsList := range teamsInfoMap {
			contestTeamsScoreInfo := make([]*model.ContestTeamScoreInfo, 0)
			for _, v := range teamsList {
				contestTeamsScoreInfo = append(contestTeamsScoreInfo, v.ScoreInfo)
			}
			goingSeasonsContestsTeams.Store(contestId, contestTeamsScoreInfo)
		}
	}
}

func watchGoingSeasonsByCacheWatch(ctx context.Context) {
	tmpCfg := conf.LoadSeasonContestComponentWatch()
	if tmpCfg == nil || !tmpCfg.CanWatch {
		return
	}
	tmpGoingSeasons, mcScanErr := match_component.FetchGoingSeasonsFromCache(ctx)
	if mcScanErr != nil {
		log.Errorc(ctx, "contest component watchComponentContestListByGoingSeason match_component.FetchGoingSeasonsFromCache error(%+v)", mcScanErr)
		// todo add moni
		return
	}
	goingSeasonsListGlobal = tmpGoingSeasons
}

func watchGoingBattleSeasonsByCacheWatch(ctx context.Context) {
	tmpCfg := conf.LoadSeasonContestComponentWatch()
	if tmpCfg == nil || !tmpCfg.CanWatch {
		return
	}
	tmpGoingBattleSeasons, mcScanErr := match_component.FetchGoingBattleSeasonsFromCache(ctx)
	if mcScanErr != nil {
		log.Errorc(ctx, "contest component watchGoingBattleSeasonsByCacheWatch match_component.FetchGoingBattleSeasonsFromCache error(%+v)", mcScanErr)
		return
	}
	goingBattleSeasonsListGlobal = tmpGoingBattleSeasons
}

func watchComponentAllTeamsWatch(ctx context.Context) {
	tmpCfg := conf.LoadSeasonContestComponentWatch()
	if tmpCfg == nil || !tmpCfg.CanWatch {
		return
	}
	tmpAllTeams, teamErr := match_component.FetchAllTeams(ctx)
	if teamErr != nil {
		log.Errorc(ctx, "contest component watchComponentAllTeamsWatch match_component.FetchAllTeams error(%+v)", teamErr)
		// todo add moni
		return
	}
	allTeamsOfComponent.Store(tmpAllTeams)
}

func watchComponentAllSeasonsWatch(ctx context.Context) {
	tmpCfg := conf.LoadSeasonContestComponentWatch()
	if tmpCfg == nil || !tmpCfg.CanWatch {
		return
	}
	tmpAllSeasons, err := match_component.FetchAllSeasons(ctx)
	if err != nil {
		log.Errorc(ctx, "contest component watchComponentAllSeasonWatch match_component.FetchAllSeasons error(%+v)", err)
		return
	}
	allSeasonsOfComponent.Store(tmpAllSeasons)
}

func watchComponentContestList(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 3)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			watchComponentContestListByGoingSeason(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func watchComponentContestBattle(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 3)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			watchComponentContestBattleByGoingSeason(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func watchComponentContestListByGoingSeason(ctx context.Context) {
	tmpCfg := conf.LoadSeasonContestComponentWatch()
	if tmpCfg == nil || !tmpCfg.CanWatch {
		return
	}
	seasonsContestListMap := make(map[int64]map[int64][]*pb.ContestCardComponent, 0)
	for _, season := range goingSeasonsListGlobal {
		tmpContestCardList, mcScanErr := match_component.FetchContestCardListFromCache(ctx, season.ID)
		if mcScanErr != nil {
			// todo add moni
			log.Errorc(ctx, "contest component watchComponentContestListByGoingSeason match_component.FetchContestListFromCache() sid(%d) error(%+v)", season.ID, mcScanErr)
			continue
		}
		seasonsContestListMap[season.ID] = tmpContestCardList
	}
	rebuildComponentGoingSeasonsContestListMap(seasonsContestListMap)
}

func watchComponentContestBattleByGoingSeason(ctx context.Context) {
	tmpCfg := conf.LoadSeasonContestComponentWatch()
	if tmpCfg == nil || !tmpCfg.CanWatch {
		return
	}
	seasonsBattleContestListMap := make(map[int64]map[int64][]*pb.ContestBattleCardComponent, 0)
	for _, season := range goingBattleSeasonsListGlobal {
		tmpContestCardList, mcScanErr := match_component.FetchContestBattleCardListFromCache(ctx, season.ID)
		if mcScanErr != nil {
			log.Errorc(ctx, "contest component watchComponentContestBattleByGoingSeason match_component.FetchContestBattleCardListFromCache() sid(%d) error(%+v)", season.ID, mcScanErr)
			continue
		}
		seasonsBattleContestListMap[season.ID] = tmpContestCardList
	}
	rebuildComponentGoingBattleSeasonsContestMap(seasonsBattleContestListMap)
}

func rebuildComponentGoingBattleSeasonsContestMap(goingSeasonsContestListMap map[int64]map[int64][]*pb.ContestBattleCardComponent) {
	tmpBattleAllComponentMap := make(map[int64][]*pb.ContestBattleCardComponent)
	for sid, contestListMap := range goingSeasonsContestListMap {
		tmpComponentContestBattle4All := genComponentContestBattle4All(contestListMap)
		tmpBattleAllComponentMap[sid] = tmpComponentContestBattle4All
	}
	// component all api
	seasonContestBattleAllComponentMap = tmpBattleAllComponentMap
}

func rebuildComponentGoingSeasonsContestListMap(goingSeasonsContestListMap map[int64]map[int64][]*pb.ContestCardComponent) {
	tmpTimeComponent0Map := make(map[int64]map[int64][]*pb.ContestCardComponent)
	tmpTimeComponent1Map := make(map[int64]map[int64][]*pb.ContestCardComponent)
	tmpTimeComponent2Map := make(map[int64]map[int64][]*pb.ContestCardComponent)
	tmpTimeComponent3Map := make(map[int64]map[int64][]*pb.ContestCardComponent)
	tmpTimeComponent4Map := make(map[int64]map[int64][]*pb.ContestCardComponent)
	tmpTimeComponent5Map := make(map[int64]map[int64][]*pb.ContestCardComponent)
	tmpTimeComponent6Map := make(map[int64]map[int64][]*pb.ContestCardComponent)
	tmpTimeComponent7Map := make(map[int64]map[int64][]*pb.ContestCardComponent)
	tmpTimeComponent8Map := make(map[int64]map[int64][]*pb.ContestCardComponent)
	tmpTimeComponent9Map := make(map[int64]map[int64][]*pb.ContestCardComponent)

	tmpTimeComponentTsList0Map := make(map[int64][]int64)
	tmpTimeComponentTsList1Map := make(map[int64][]int64)
	tmpTimeComponentTsList2Map := make(map[int64][]int64)
	tmpTimeComponentTsList3Map := make(map[int64][]int64)
	tmpTimeComponentTsList4Map := make(map[int64][]int64)
	tmpTimeComponentTsList5Map := make(map[int64][]int64)
	tmpTimeComponentTsList6Map := make(map[int64][]int64)
	tmpTimeComponentTsList7Map := make(map[int64][]int64)
	tmpTimeComponentTsList8Map := make(map[int64][]int64)
	tmpTimeComponentTsList9Map := make(map[int64][]int64)

	tmpAllComponent0Map := make(map[int64][]*pb.ContestCardComponent)
	tmpAllComponent1Map := make(map[int64][]*pb.ContestCardComponent)
	tmpAllComponent2Map := make(map[int64][]*pb.ContestCardComponent)
	tmpAllComponent3Map := make(map[int64][]*pb.ContestCardComponent)
	tmpAllComponent4Map := make(map[int64][]*pb.ContestCardComponent)
	tmpAllComponent5Map := make(map[int64][]*pb.ContestCardComponent)
	tmpAllComponent6Map := make(map[int64][]*pb.ContestCardComponent)
	tmpAllComponent7Map := make(map[int64][]*pb.ContestCardComponent)
	tmpAllComponent8Map := make(map[int64][]*pb.ContestCardComponent)
	tmpAllComponent9Map := make(map[int64][]*pb.ContestCardComponent)

	for sid, contestListMap := range goingSeasonsContestListMap {
		tmpComponentContestList4All := genComponentContestList4All(contestListMap)
		tmpTimestampList := genComponentContestTimeTimestampList(contestListMap)
		switch sidSharding(sid) {
		case 0:
			tmpTimeComponent0Map[sid] = contestListMap
			tmpTimeComponentTsList0Map[sid] = tmpTimestampList
			tmpAllComponent0Map[sid] = tmpComponentContestList4All
		case 1:
			tmpTimeComponent1Map[sid] = contestListMap
			tmpTimeComponentTsList1Map[sid] = tmpTimestampList
			tmpAllComponent1Map[sid] = tmpComponentContestList4All
		case 2:
			tmpTimeComponent2Map[sid] = contestListMap
			tmpTimeComponentTsList2Map[sid] = tmpTimestampList
			tmpAllComponent2Map[sid] = tmpComponentContestList4All
		case 3:
			tmpTimeComponent3Map[sid] = contestListMap
			tmpTimeComponentTsList3Map[sid] = tmpTimestampList
			tmpAllComponent3Map[sid] = tmpComponentContestList4All
		case 4:
			tmpTimeComponent4Map[sid] = contestListMap
			tmpTimeComponentTsList4Map[sid] = tmpTimestampList
			tmpAllComponent4Map[sid] = tmpComponentContestList4All
		case 5:
			tmpTimeComponent5Map[sid] = contestListMap
			tmpTimeComponentTsList5Map[sid] = tmpTimestampList
			tmpAllComponent5Map[sid] = tmpComponentContestList4All
		case 6:
			tmpTimeComponent6Map[sid] = contestListMap
			tmpTimeComponentTsList6Map[sid] = tmpTimestampList
			tmpAllComponent6Map[sid] = tmpComponentContestList4All
		case 7:
			tmpTimeComponent7Map[sid] = contestListMap
			tmpTimeComponentTsList7Map[sid] = tmpTimestampList
			tmpAllComponent7Map[sid] = tmpComponentContestList4All
		case 8:
			tmpTimeComponent8Map[sid] = contestListMap
			tmpTimeComponentTsList8Map[sid] = tmpTimestampList
			tmpAllComponent8Map[sid] = tmpComponentContestList4All
		case 9:
			tmpTimeComponent9Map[sid] = contestListMap
			tmpTimeComponentTsList9Map[sid] = tmpTimestampList
			tmpAllComponent9Map[sid] = tmpComponentContestList4All
		}
	}
	// component time api
	seasonContestTimeComponent0Map = tmpTimeComponent0Map
	seasonContestTimeComponent1Map = tmpTimeComponent1Map
	seasonContestTimeComponent2Map = tmpTimeComponent2Map
	seasonContestTimeComponent3Map = tmpTimeComponent3Map
	seasonContestTimeComponent4Map = tmpTimeComponent4Map
	seasonContestTimeComponent5Map = tmpTimeComponent5Map
	seasonContestTimeComponent6Map = tmpTimeComponent6Map
	seasonContestTimeComponent7Map = tmpTimeComponent7Map
	seasonContestTimeComponent8Map = tmpTimeComponent8Map
	seasonContestTimeComponent9Map = tmpTimeComponent9Map

	// component time ts list
	seasonContestTimeComponentTsList0Map = tmpTimeComponentTsList0Map
	seasonContestTimeComponentTsList1Map = tmpTimeComponentTsList1Map
	seasonContestTimeComponentTsList2Map = tmpTimeComponentTsList2Map
	seasonContestTimeComponentTsList3Map = tmpTimeComponentTsList3Map
	seasonContestTimeComponentTsList4Map = tmpTimeComponentTsList4Map
	seasonContestTimeComponentTsList5Map = tmpTimeComponentTsList5Map
	seasonContestTimeComponentTsList6Map = tmpTimeComponentTsList6Map
	seasonContestTimeComponentTsList7Map = tmpTimeComponentTsList7Map
	seasonContestTimeComponentTsList8Map = tmpTimeComponentTsList8Map
	seasonContestTimeComponentTsList9Map = tmpTimeComponentTsList9Map

	// component all api
	seasonContestAllComponent0Map = tmpAllComponent0Map
	seasonContestAllComponent1Map = tmpAllComponent1Map
	seasonContestAllComponent2Map = tmpAllComponent2Map
	seasonContestAllComponent3Map = tmpAllComponent3Map
	seasonContestAllComponent4Map = tmpAllComponent4Map
	seasonContestAllComponent5Map = tmpAllComponent5Map
	seasonContestAllComponent6Map = tmpAllComponent6Map
	seasonContestAllComponent7Map = tmpAllComponent7Map
	seasonContestAllComponent8Map = tmpAllComponent8Map
	seasonContestAllComponent9Map = tmpAllComponent9Map
}

func genComponentContestTimeTimestampList(m map[int64][]*pb.ContestCardComponent) []int64 {
	timestampList := make([]int64, 0)
	if len(m) == 0 {
		return timestampList
	}
	for k := range m {
		timestampList = append(timestampList, k)
	}
	sort.SliceStable(timestampList, func(i, j int) bool {
		return timestampList[i] < timestampList[j]
	})
	return timestampList
}

func genComponentContestList4All(m map[int64][]*pb.ContestCardComponent) []*pb.ContestCardComponent {
	componentAllContests := make([]*pb.ContestCardComponent, 0)
	if len(m) == 0 {
		return componentAllContests
	}
	for _, contestCard := range m {
		componentAllContests = append(componentAllContests, contestCard...)
	}
	return componentAllContests
}

func seasonContestTimeComponentListMap(sid int64) (res map[int64][]*pb.ContestCardComponent, ok bool) {
	switch sidSharding(sid) {
	case 0:
		res, ok = seasonContestTimeComponent0Map[sid]
	case 1:
		res, ok = seasonContestTimeComponent1Map[sid]
	case 2:
		res, ok = seasonContestTimeComponent2Map[sid]
	case 3:
		res, ok = seasonContestTimeComponent3Map[sid]
	case 4:
		res, ok = seasonContestTimeComponent4Map[sid]
	case 5:
		res, ok = seasonContestTimeComponent5Map[sid]
	case 6:
		res, ok = seasonContestTimeComponent6Map[sid]
	case 7:
		res, ok = seasonContestTimeComponent7Map[sid]
	case 8:
		res, ok = seasonContestTimeComponent8Map[sid]
	case 9:
		res, ok = seasonContestTimeComponent9Map[sid]
	}
	return res, ok
}

func seasonContestTimeTimestampListMap(sid int64) (res []int64, ok bool) {
	switch sidSharding(sid) {
	case 0:
		res, ok = seasonContestTimeComponentTsList0Map[sid]
	case 1:
		res, ok = seasonContestTimeComponentTsList1Map[sid]
	case 2:
		res, ok = seasonContestTimeComponentTsList2Map[sid]
	case 3:
		res, ok = seasonContestTimeComponentTsList3Map[sid]
	case 4:
		res, ok = seasonContestTimeComponentTsList4Map[sid]
	case 5:
		res, ok = seasonContestTimeComponentTsList5Map[sid]
	case 6:
		res, ok = seasonContestTimeComponentTsList6Map[sid]
	case 7:
		res, ok = seasonContestTimeComponentTsList7Map[sid]
	case 8:
		res, ok = seasonContestTimeComponentTsList8Map[sid]
	case 9:
		res, ok = seasonContestTimeComponentTsList9Map[sid]
	}
	return res, ok
}

func seasonContestAllComponentList(sid int64) (res []*pb.ContestCardComponent, ok bool) {
	switch sidSharding(sid) {
	case 0:
		res, ok = seasonContestAllComponent0Map[sid]
	case 1:
		res, ok = seasonContestAllComponent1Map[sid]
	case 2:
		res, ok = seasonContestAllComponent2Map[sid]
	case 3:
		res, ok = seasonContestAllComponent3Map[sid]
	case 4:
		res, ok = seasonContestAllComponent4Map[sid]
	case 5:
		res, ok = seasonContestAllComponent5Map[sid]
	case 6:
		res, ok = seasonContestAllComponent6Map[sid]
	case 7:
		res, ok = seasonContestAllComponent7Map[sid]
	case 8:
		res, ok = seasonContestAllComponent8Map[sid]
	case 9:
		res, ok = seasonContestAllComponent9Map[sid]
	}
	return res, ok
}

func fetchSeasonContestsList(ctx context.Context, sid int64) (res map[int64][]*pb.ContestCardComponent, err error) {
	if res, err = match_component.FetchContestCardListFromCache(ctx, sid); err != nil && err != memcache.ErrNotFound {
		log.Errorc(ctx, "contest component fetchSeasonContestsList match_component.FetchContestListFromCache() sid(%d) error(%+v)", sid, err)
		return
	}
	if err == nil {
		return
	}
	if err == memcache.ErrNotFound {
		res, err = fetchComponentContestListBySeasonID(context.Background(), sid)
		if err != nil {
			return
		}
		if e := match_component.FetchContestCardListToCache(ctx, sid, res, int32(tool.CalculateExpiredSeconds(10))); e != nil {
			log.Errorc(ctx, "contest component fetchSeasonContestsList match_component.FetchContestListToCache() sid(%d) error(%+v)", sid, e)
		}
	}
	return
}

func (s *Service) loadComponentContestCardTimeRelationsFromService(ctx context.Context, sid int64) (contestMaps map[int64][]*pb.ContestCardComponent, timestampList []int64, err error) {
	contestMaps = make(map[int64][]*pb.ContestCardComponent)
	timestampList = make([]int64, 0)
	seasonContests, err := s.esportsServiceClient.GetContestInfoListBySeason(ctx, &v1.GetContestInfoListBySeasonReq{
		SeasonID: sid,
	})
	if err != nil {
		log.Errorc(ctx, "[Service][LoadComponent][esportsServiceClient][GetSeasonContests], err:%+v", err)
		return
	}
	if seasonContests.ComponentContestList != nil && len(seasonContests.ComponentContestList) != 0 {
		for k, v := range seasonContests.ComponentContestList {
			timestampList = append(timestampList, k)
			for _, contest := range v.Contests {
				if contest == nil {
					continue
				}
				contestCard := contestDetail2Card(contest)
				if _, ok := contestMaps[k]; ok {
					contestMaps[k] = append(contestMaps[k], contestCard)
				} else {
					contestMaps[k] = []*pb.ContestCardComponent{contestCard}
				}
			}
		}
	}
	return
}

func calculateStatus(contest *v1.ContestDetail) string {
	now := time.Now().Unix()
	if now >= contest.Etime {
		return model.ContestStatusOfEnd
	} else if now >= contest.Stime {
		return model.ContestStatusOfOngoing
	}

	return model.ContestStatusOfNotStart
}

func contestDetail2Card(detail *v1.ContestDetail) *pb.ContestCardComponent {
	if detail == nil {
		return nil
	}
	guessType := 0
	if detail.IsGuessed != v1.GuessStatusEnum_HasNoGuess {
		guessType = _guessOk
	}
	return &pb.ContestCardComponent{
		ID:            detail.ID,
		StartTime:     detail.Stime,
		EndTime:       detail.Etime,
		Title:         "",
		Status:        calculateStatus(detail),
		CollectionURL: detail.CollectionURL,
		LiveRoom:      detail.LiveRoom,
		PlayBack:      detail.Playback,
		DataType:      detail.DataType,
		MatchID:       detail.MatchID,
		SeasonID:      detail.Sid,
		GuessType:     int64(guessType),
		SeriesID:      detail.SeriesId,
		IsSub:         0,
		IsGuess:       0,
		Home:          teamInfo2Team(detail.HomeTeam),
		Away:          teamInfo2Team(detail.AwayTeam),
		Series:        nil,
		ContestStatus: 0,
		ContestFreeze: 0,
		GameState:     0,
		GuessShow:     0,
		HomeScore:     0,
		AwayScore:     0,
	}
}

func teamInfo2Team(teamInfo *v1.TeamDetail) *pb.Team4FrontendComponent {
	if teamInfo == nil {
		return nil
	}
	return &pb.Team4FrontendComponent{
		ID:       teamInfo.ID,
		Icon:     teamInfo.Logo,
		Name:     teamInfo.Title,
		Region:   genTeamRegionDisplayByRegionID(teamInfo.RegionId),
		RegionID: teamInfo.RegionId,
	}
}

func loadComponentContestCardTimeRelations(ctx context.Context, sid int64) (map[int64][]*pb.ContestCardComponent, []int64, error) {
	var err error
	timestampList := make([]int64, 0)
	componentContests, ok := seasonContestTimeComponentListMap(sid)
	if !ok { // 回源.
		componentContests, err = fetchSeasonContestsList(ctx, sid)
		if err != nil {
			return nil, nil, err
		}
	}
	if len(componentContests) == 0 {
		tmpRs := make(map[int64][]*pb.ContestCardComponent, 0)
		return tmpRs, timestampList, err
	}
	componentTimeContestMap := make(map[int64][]*pb.ContestCardComponent, len(componentContests))
	for k, v := range componentContests {
		tmpList := make([]*pb.ContestCardComponent, 0)
		for _, d := range v {
			tmpCard := new(pb.ContestCardComponent)
			*tmpCard = *d
			tmpList = append(tmpList, tmpCard)
		}
		componentTimeContestMap[k] = tmpList
	}
	// 内存中取.
	tmpContestCardTsList4All, ok := seasonContestTimeTimestampListMap(sid)
	if !ok { // 重新生成数据.
		tmpContestCardTsList4All = genComponentContestTimeTimestampList(componentTimeContestMap)
	}
	if len(tmpContestCardTsList4All) > 0 {
		for _, v := range tmpContestCardTsList4All {
			timestampList = append(timestampList, v)
		}
	}
	return componentTimeContestMap, timestampList, err
}

func (s *Service) TimeContests(ctx context.Context, mid, sid int64) (res []*model.ComponentContestCardTime, err error) {
	var (
		cardListMap map[int64][]*pb.ContestCardComponent
		tsList      []int64
	)
	res = make([]*model.ComponentContestCardTime, 0)
	year, month, day := time.Now().Date()
	dayUnix := time.Date(year, month, day, 0, 0, 0, 0, time.Local).Unix()
	beforeToday := make([]int64, 0)
	afterToday := make([]int64, 0)
	if cardListMap, tsList, err = loadComponentContestCardTimeRelations(ctx, sid); err != nil {
		return
	}
	for _, v := range tsList {
		if v < dayUnix {
			if len(beforeToday) >= 4 {
				beforeToday = append(beforeToday[1:], v) // 取离当前时间最近的4天，取后边的
			} else {
				beforeToday = append(beforeToday, v)
			}
		} else if v >= dayUnix {
			if len(afterToday) >= 4 {
				//afterToday = append(afterToday[1:], v) //大于当前时间取最前边的，不用再赋值
			} else {
				afterToday = append(afterToday, v) // 取离当前时间最近的4天，多了就不再取
			}
		}
	}
	needTsList := make([]int64, 0)
	switch len(afterToday) {
	case 0:
		needTsList = beforeToday
	case 1:
		if len(beforeToday) > 3 {
			needTsList = beforeToday[len(beforeToday)-3:]
		}
		needTsList = append(needTsList, afterToday...)
	case 2:
		if len(beforeToday) > 2 {
			needTsList = beforeToday[len(beforeToday)-2:]
		}
		needTsList = append(needTsList, afterToday...)
	default:
		if len(beforeToday) > 0 {
			needTsList = beforeToday[len(beforeToday)-1:]
		}
		needTsList = append(needTsList, afterToday...)
	}
	if len(needTsList) > 4 {
		needTsList = needTsList[:4]
	}
	list := make([]*model.ComponentContestCardTime, 0)
	for _, ts := range needTsList {
		if d, ok := cardListMap[ts]; ok {
			tmpCardList := make([]*pb.ContestCardComponent, 0)
			for _, v := range d {
				tmpCard := new(pb.ContestCardComponent)
				*tmpCard = *v //deepCopyContestCard(v)
				tmpCardList = append(tmpCardList, tmpCard)
			}
			tmpCardList = sortComponentContestList(tmpCardList, contestSortByStatus)
			moreContestCard := new(model.ComponentContestCardTime)
			{
				moreContestCard.Timestamp = ts
				moreContestCard.ContestCards = tmpCardList
			}
			list = append(list, moreContestCard)
		}
	}
	res = s.rebuildComponentContestCardTime(ctx, mid, list)
	return
}

func sortComponentContestList(list []*pb.ContestCardComponent, sortType int) []*pb.ContestCardComponent {
	sort.SliceStable(list, func(i, j int) bool {
		switch sortType {
		case contestSortByEndTime:
			return func() bool {
				if list[i].EndTime > list[j].EndTime {
					return list[i].EndTime > list[j].EndTime
				}
				return list[i].ID > list[j].ID
			}()
		case contestSortByStatus:
			switch list[i].Status {
			case model.ContestStatusOfOngoing:
				switch list[j].Status {
				case model.ContestStatusOfEnd:
					return true
				default:
					return func() bool {
						if list[i].StartTime != list[j].StartTime {
							return list[i].StartTime < list[j].StartTime
						}
						return list[i].ID < list[j].ID
					}()
				}
			case model.ContestStatusOfNotStart:
				switch list[j].Status {
				case model.ContestStatusOfOngoing:
					return false
				case model.ContestStatusOfEnd:
					return true
				default:
					return func() bool {
						if list[i].StartTime != list[j].StartTime {
							return list[i].StartTime < list[j].StartTime
						}
						return list[i].ID < list[j].ID
					}()
				}
			case model.ContestStatusOfEnd:
				switch list[j].Status {
				case model.ContestStatusOfEnd:
					return func() bool {
						if list[i].EndTime != list[j].EndTime {
							return list[i].EndTime < list[j].EndTime
						}
						return list[i].ID < list[j].ID
					}()
				default:
					return false
				}
			}
		case contestSortByLive:
			// TODO
		}
		return list[i].StartTime < list[j].StartTime
	})
	return list
}

func (s *Service) rebuildComponentContestCardTime(ctx context.Context, mid int64, list []*model.ComponentContestCardTime) []*model.ComponentContestCardTime {
	contestIDList4FavComponent := make([]int64, 0)
	contestIDList4GuessComponent := make([]int64, 0)
	for _, cards := range list {
		for _, card := range cards.ContestCards {
			contestIDList4FavComponent = append(contestIDList4FavComponent, card.ID)
			if card.GuessType == 1 {
				contestIDList4GuessComponent = append(contestIDList4GuessComponent, card.ID)
			}
		}
	}
	if mid > 0 {
		subscribeMap := s.fetchFavoriteMap(ctx, mid, contestIDList4FavComponent)
		guessMap := s.fetchComponentContestGuessMap(ctx, mid, contestIDList4GuessComponent)
		if len(subscribeMap) > 0 || len(guessMap) > 0 {
			for _, cards := range list {
				for _, card := range cards.ContestCards {
					if d, ok := subscribeMap[card.ID]; ok && d {
						card.IsSub = _haveSubscribe
					}
					if d, ok := guessMap[card.ID]; ok && d {
						card.IsGuess = _haveGuess
					}
					// GameState值依赖IsSub
					card.GameState = s.resetComponentContestGameState(card)
				}
			}
		}
	}
	return list
}

func (s *Service) resetComponentContestGameState(contest *pb.ContestCardComponent) int64 {
	if contest.ContestStatus == ContestStatusEnd {
		return _gameOver
	} else if contest.ContestStatus == ContestStatusOngoing {
		if contest.LiveRoom == 0 {
			return _gameIn
		} else {
			return _gameLive
		}
	} else if contest.LiveRoom > 0 {
		if contest.IsSub == _haveSubscribe {
			return _gameSub
		} else {
			return _gameNoSub
		}
	}
	return 0
}

func (s *Service) AllContests(ctx context.Context, mid int64, p *model.ParamAllContest) ([]*pb.ContestCardComponent, int, error) {
	cardContestList, total, err := loadComponentContestCardAllRelations(ctx, p)
	if err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return cardContestList, total, nil
	}
	list := s.rebuildComponentContestCardAll(ctx, mid, cardContestList)
	return list, total, nil
}

func deepCopyContestAll(contestAll []*pb.ContestCardComponent) []*pb.ContestCardComponent {
	tmpRes := make([]*pb.ContestCardComponent, 0)
	for _, contest := range contestAll {
		tmpContest := new(pb.ContestCardComponent)
		*tmpContest = *contest
		tmpRes = append(tmpRes, tmpContest)
	}
	return tmpRes
}

func fetchComponentContestListAll(ctx context.Context, seasonID int64) (res []*pb.ContestCardComponent, err error) {
	var (
		ok                   bool
		componentContestsAll []*pb.ContestCardComponent
	)
	componentContestsAll, ok = seasonContestAllComponentList(seasonID)
	if !ok { // 回源.
		var componentContests map[int64][]*pb.ContestCardComponent
		componentContests, err = fetchSeasonContestsList(ctx, seasonID)
		if err != nil {
			return
		}
		componentContestsAll = genComponentContestList4All(componentContests)
	}
	res = deepCopyContestAll(componentContestsAll)
	return
}

func fetchComponentContestList(ctx context.Context, p *model.ParamAllContest) (res []*pb.ContestCardComponent, err error) {
	var tmpRes []*pb.ContestCardComponent
	if tmpRes, err = fetchComponentContestListAll(ctx, p.Sid); err != nil {
		log.Errorc(ctx, "contest component fetchComponentContestList  fetchComponentContestListAll Param(%+v) error(%+v)", p, err)
		return
	}
	if p.Sort == _findAsc {
		sort.SliceStable(tmpRes, func(i, j int) bool {
			if tmpRes[i].StartTime != tmpRes[j].StartTime {
				return tmpRes[i].StartTime < tmpRes[j].StartTime
			}
			return tmpRes[i].ID < tmpRes[j].ID
		})
	} else if p.Sort == _findDesc {
		sort.SliceStable(tmpRes, func(i, j int) bool {
			if tmpRes[i].StartTime != tmpRes[j].StartTime {
				return tmpRes[i].StartTime > tmpRes[j].StartTime
			}
			return tmpRes[i].ID > tmpRes[j].ID
		})
	}
	// 根据开始时间判断.
	if p.Stime != 0 && p.Etime != 0 {
		for _, contest := range tmpRes {
			if contest.StartTime >= p.Stime && contest.StartTime <= p.Etime {
				res = append(res, contest)
			}
		}
	} else if p.Stime == 0 && p.Etime != 0 {
		for _, contest := range tmpRes {
			if contest.StartTime <= p.Etime {
				res = append(res, contest)
			}
		}
	} else if p.Stime != 0 && p.Etime == 0 {
		for _, contest := range tmpRes {
			if contest.StartTime >= p.Stime {
				res = append(res, contest)
			}
		}
	} else {
		res = tmpRes
	}
	return
}

func loadComponentContestCardAllRelations(ctx context.Context, p *model.ParamAllContest) (res []*pb.ContestCardComponent, total int, err error) {
	var (
		componentContestsList []*pb.ContestCardComponent
		start                 = (p.Pn - 1) * p.Ps
		end                   = start + p.Ps - 1
	)
	res = make([]*pb.ContestCardComponent, 0)
	if componentContestsList, err = fetchComponentContestList(ctx, p); err != nil {
		return
	}
	total = len(componentContestsList)
	if total == 0 || total < start {
		return
	}
	currentContestList := make([]*pb.ContestCardComponent, p.Ps)
	if total > end+1 {
		currentContestList = componentContestsList[start : end+1]
	} else {
		currentContestList = componentContestsList[start:]
	}
	tmpList := make([]*pb.ContestCardComponent, 0)
	for _, v := range currentContestList {
		tmpCard := new(pb.ContestCardComponent)
		*tmpCard = *v
		tmpList = append(tmpList, tmpCard)
	}
	res = tmpList
	return
}

func (s *Service) rebuildComponentContestCardAll(ctx context.Context, mid int64, list []*pb.ContestCardComponent) []*pb.ContestCardComponent {
	contestIDList4FavComponent := make([]int64, 0)
	contestIDList4GuessComponent := make([]int64, 0)
	for _, card := range list {
		contestIDList4FavComponent = append(contestIDList4FavComponent, card.ID)
		if card.GuessType == 1 {
			contestIDList4GuessComponent = append(contestIDList4GuessComponent, card.ID)
		}
	}
	if mid > 0 {
		subscribeMap := s.fetchFavoriteMap(ctx, mid, contestIDList4FavComponent)
		guessMap := s.fetchComponentContestGuessMap(ctx, mid, contestIDList4GuessComponent)
		if len(subscribeMap) > 0 || len(guessMap) > 0 {
			for _, card := range list {
				if d, ok := subscribeMap[card.ID]; ok && d {
					card.IsSub = _haveSubscribe
				}
				if d, ok := guessMap[card.ID]; ok && d {
					card.IsGuess = _haveGuess
				}
				// GameState值依赖IsSub
				card.GameState = s.resetComponentContestGameState(card)
			}
		}
	}
	return list
}

func (s *Service) fetchComponentContestGuessMap(ctx context.Context, mid int64, contestIDList []int64) map[int64]bool {
	m := make(map[int64]bool, 0)
	idsCount := len(contestIDList)
	if idsCount == 0 || mid == 0 {
		return m
	}
	for i := 0; i < idsCount; i += _contestIDsBulkSize {
		var partIDs []int64
		if i+_contestIDsBulkSize > idsCount {
			partIDs = contestIDList[i:]
		} else {
			partIDs = contestIDList[i : i+_contestIDsBulkSize]
		}
		req := new(actApi.UserGuessMatchsReq)
		{
			req.Mid = mid
			req.Business = _esportsBusiness
			req.Oids = partIDs
			req.Pn = _guessMatchFirst
			req.Ps = int64(len(partIDs))
		}
		tmpResp, tmpErr := s.actClient.UserGuessMatchs(ctx, req)
		if tmpErr != nil {
			log.Errorc(ctx, "contest component componentContestGuessMap  mid(%d) contestIDList(%+v) error(%+v)", mid, contestIDList, tmpErr)
			// todo add moni
			return m
		}
		for _, userGuess := range tmpResp.UserGroup {
			m[userGuess.Oid] = true
		}
	}
	return m
}

func fetchComponentContestListBySeasonID(ctx context.Context, sid int64) (res map[int64][]*pb.ContestCardComponent, err error) {
	componentTeamMap := allTeamsOfComponent.Load().(map[int64]*model.Team2TabComponent)
	teamCount := len(componentTeamMap)
	if componentTeamMap == nil || teamCount == 0 {
		log.Warnc(ctx, "contest component updateContestsByGoingSeason seasonID(%d) seasonTeamsOfComponent teams empty", sid)
	}
	teamMap := make(map[int64]*model.Team2TabComponent, teamCount)
	for _, team := range componentTeamMap {
		teamMap[team.ID] = team
	}
	var contests []*model.Contest2TabComponent
	if contests, err = match_component.FetchContestsBySeasonComponent(ctx, sid); err != nil {
		log.Errorc(ctx, "contest component fetchComponentContestListBySeasonID match_component.FetchContestsBySeasonComponent(%d) error(%+v)", sid, err)
		err = ecode.EsportsComponentErr
		return
	}
	// deep copy.
	tmpContestList := deepCopyContestInfo(contests)
	seriesComponentMap, _ := componentSeriesByContestList(ctx, tmpContestList)
	res = generateComponentContestList4Frontend(tmpContestList, teamMap, seriesComponentMap)
	return
}

func (s *Service) GetAllTeamsOfComponent(ctx context.Context) map[int64]*model.Team2TabComponent {
	componentTeamMap := allTeamsOfComponent.Load().(map[int64]*model.Team2TabComponent)
	teamCount := len(componentTeamMap)
	if componentTeamMap == nil || teamCount == 0 {
		log.Warnc(ctx, "contest component updateContestsByGoingSeason seasonTeamsOfComponent teams empty")
	}
	return componentTeamMap
}

func (s *Service) GetAllSeasonsOfComponent(ctx context.Context) map[int64]*model.SeasonComponent {
	componentSeasonMap := allSeasonsOfComponent.Load().(map[int64]*model.SeasonComponent)
	teamCount := len(componentSeasonMap)
	if componentSeasonMap == nil || teamCount == 0 {
		log.Warnc(ctx, "contest component GetAllSeasonsOfComponent seasons empty")
	}
	return componentSeasonMap
}

func deepCopyContestInfo(list []*model.Contest2TabComponent) []*model.Contest2TabComponent {
	var tmpContestList []*model.Contest2TabComponent
	for _, contest := range list {
		tmpContest := new(model.Contest2TabComponent)
		*tmpContest = *contest
		tmpContestList = append(tmpContestList, tmpContest)
	}
	return tmpContestList
}

// 获取赛程对应阶段信息
func ComponentOidSeriesBySeason(ctx context.Context, sid int64, contestIDs []int64) (res map[int64]*pb.ContestSeriesComponent, err error) {
	var (
		contests       []*model.Contest2TabComponent
		seriesContests []*model.Contest2TabComponent
		seriesMap      map[int64]*pb.ContestSeriesComponent
	)
	count := len(contestIDs)
	res = make(map[int64]*pb.ContestSeriesComponent, count)
	if contests, err = match_component.FetchContestsBySeasonComponent(ctx, sid); err != nil {
		log.Errorc(ctx, "contest component ComponentSeriesBySeason match_component.FetchContestsBySeasonComponent(%d) error(%+v)", sid, err)
		err = ecode.EsportsComponentErr
		return
	}
	// deep copy.
	tmpContestList := deepCopyContestInfo(contests)
	tmpContestMap := make(map[int64]*model.Contest2TabComponent, count)
	for _, tmpContest := range tmpContestList {
		tmpContestMap[tmpContest.ID] = tmpContest
	}
	for _, cid := range contestIDs {
		seriesContent, ok := tmpContestMap[cid]
		if !ok {
			continue
		}
		seriesContests = append(seriesContests, seriesContent)
	}
	if seriesMap, err = componentSeriesByContestList(ctx, seriesContests); err != nil {
		log.Errorc(ctx, "contest component ComponentOidSeriesBySeason componentSeriesByContestList() sid(%d) error(%+v)", sid, err)
		return
	}
	for _, contestComponent := range seriesContests {
		if seriesComponent, ok := seriesMap[contestComponent.SeriesID]; ok {
			res[contestComponent.ID] = seriesComponent
		}
	}
	return
}

// 获取赛程阶段信息
func componentSeriesByContestList(ctx context.Context, contestList []*model.Contest2TabComponent) (seriesMap map[int64]*pb.ContestSeriesComponent, err error) {
	seriesMap = make(map[int64]*pb.ContestSeriesComponent, 0)
	idList := make([]int64, 0)
	for _, v := range contestList {
		if v.SeriesID > 0 {
			idList = append(idList, v.SeriesID)
		}
	}
	if len(idList) == 0 {
		return
	}
	if seriesMap, err = match_component.ContestSeriesComponent(ctx, idList); err != nil {
		log.Errorc(ctx, "contest component componentSeriesByContestList ContestSeriesComponent idList(%+v) error(%+v)", idList, err)
		return
	}
	return
}

func generateComponentContestList4Frontend(contestList []*model.Contest2TabComponent, teamMap map[int64]*model.Team2TabComponent, seriesMap map[int64]*pb.ContestSeriesComponent) map[int64][]*pb.ContestCardComponent {
	componentContestCardList := make(map[int64][]*pb.ContestCardComponent, 0)
	for _, contest := range contestList {
		tmpContestComponent := new(model.Contest2TabComponent)
		*tmpContestComponent = *contest
		cardList := make([]*pb.ContestCardComponent, 0)
		dateUnix := tmpContestComponent.StimeDate
		if d, ok := componentContestCardList[dateUnix]; ok {
			cardList = d
		}
		teamLID := tmpContestComponent.HomeID
		teamRID := tmpContestComponent.AwayID
		if teamL, ok := teamMap[teamLID]; ok {
			if teamR, ok := teamMap[teamRID]; ok {
				newCard := genComponentContestCardByContestTabAndTwoTeam(tmpContestComponent, teamL, teamR, seriesMap)
				cardList = append(cardList, newCard)
				componentContestCardList[tmpContestComponent.StimeDate] = cardList
			}
		}
	}
	return componentContestCardList
}

func genComponentContestCardByContestTabAndTwoTeam(contest *model.Contest2TabComponent, teamLeft, teamRight *model.Team2TabComponent, seriesMap map[int64]*pb.ContestSeriesComponent) *pb.ContestCardComponent {
	contestCard := new(pb.ContestCardComponent)
	{
		contestCard.ID = contest.ID
		contestCard.Title = contest.GameStage
		contestCard.GameStage = contest.GameStage
		contestCard.StartTime = contest.Stime
		contestCard.EndTime = contest.Etime
		contestCard.Status = contest.CalculateStatus() // todo old 之后会去掉
		contestCard.Series = genContestComponentSeriesBySeriesID(contest.SeriesID, seriesMap)
		contestCard.SeriesID = contest.SeriesID
		contestCard.CollectionURL = contest.CollectionUrl
		contestCard.LiveRoom = contest.LiveRoom
		contestCard.PlayBack = contest.PlayBack
		contestCard.DataType = contest.DataType
		contestCard.MatchID = contest.MatchID
		contestCard.GuessType = contest.GuessType
		contestCard.SeasonID = contest.SeasonID
		contestCard.ContestFreeze = contest.Status
		contestCard.ContestStatus = contest.ContestStatus
		contestCard.HomeScore = contest.HomeScore
		contestCard.AwayScore = contest.AwayScore
		if (contest.GuessType == _guessOk) && (contest.Stime-time.Now().Unix() > secondsOf10Minutes) {
			contestCard.GuessShow = _guessOk
		}
		home := pb.Team4FrontendComponent{}
		{
			home.ID = teamLeft.ID
			home.Icon = teamLeft.Logo
			home.Name = teamLeft.Title
			home.RegionID = teamLeft.RegionID
			home.Region = genTeamRegionDisplayByRegionID(teamLeft.RegionID)
			home.Wins = contest.HomeScore
		}
		away := pb.Team4FrontendComponent{}
		{
			away.ID = teamRight.ID
			away.Icon = teamRight.Logo
			away.Name = teamRight.Title
			away.RegionID = teamRight.RegionID
			away.Region = genTeamRegionDisplayByRegionID(teamRight.RegionID)
			away.Wins = contest.AwayScore
		}
		contestCard.Home = &home
		contestCard.Away = &away
	}
	return contestCard
}

func genContestComponentSeriesBySeriesID(id int64, seriesMap map[int64]*pb.ContestSeriesComponent) (series *pb.ContestSeriesComponent) {
	series = new(pb.ContestSeriesComponent)
	if id == 0 {
		return
	}
	if tmpSeries, ok := seriesMap[id]; ok {
		series = tmpSeries
	}
	return
}

func genTeamRegionDisplayByRegionID(regionID int64) (display string) {
	display = teamRegionDisplayOfNull
	switch regionID {
	case teamRegionIDOfNull:
		display = teamRegionDisplayOfNull
	case teamRegionIDOfChina:
		display = teamRegionDisplayOfChina
	case teamRegionIDOfChinaTaiWan:
		display = teamRegionDisplayOfChinaTaiWan
	}
	return
}

func (s *Service) AllFoldContests(ctx context.Context, mid int64, p *model.ParamAllFold) (res *model.ComponentContestCardFold, err error) {
	var tmpRes []*pb.ContestCardComponent
	if tmpRes, err = fetchComponentContestListAll(ctx, p.Sid); err != nil {
		log.Errorc(ctx, "contest component AllFoldContests  fetchComponentContestListAll Param(%+v) error(%+v)", p, err)
		return
	}
	// 开始时间升序.
	sort.SliceStable(tmpRes, func(i, j int) bool {
		if tmpRes[i].StartTime != tmpRes[j].StartTime {
			return tmpRes[i].StartTime < tmpRes[j].StartTime
		}
		return tmpRes[i].ID < tmpRes[j].ID
	})

	res = s.rebuildComponentContestCardFold(ctx, mid, p, tmpRes)
	return
}

func (s *Service) rebuildComponentContestCardFold(ctx context.Context, mid int64, p *model.ParamAllFold, tmpRes []*pb.ContestCardComponent) (res *model.ComponentContestCardFold) {
	contestIDList4FavComponent := make([]int64, 0)
	contestIDList4GuessComponent := make([]int64, 0)
	frontList := make([]*pb.ContestCardComponent, 0)
	middleList := make([]*pb.ContestCardComponent, 0)
	backList := make([]*pb.ContestCardComponent, 0)
	nowTime := time.Now().Unix()
	for _, contest := range tmpRes {
		if contest.StartTime < nowTime && contest.EndTime < nowTime { // 已结束.
			if p.Front == 0 {
				continue
			}
			if len(frontList) >= p.Front {
				frontList = append(frontList[1:], contest) // 取离当前时间最近赛程，取后边的.
			} else {
				frontList = append(frontList, contest)
			}
		} else if len(middleList) == 0 && contest.StartTime <= nowTime && (contest.EndTime > nowTime || contest.EndTime == 0) { // 最开始进行中.
			middleList = append(middleList, contest)
		} else { // 进行中与未开始.
			if len(middleList) == 0 { // 没有进行中的取将要开始的.
				middleList = append(middleList, contest)
				continue
			}
			if p.Back == 0 {
				continue
			}
			if len(backList) >= p.Back {
				continue
			} else {
				backList = append(backList, contest) // 取离当前时间最近赛程.
			}
		}
	}
	for _, frontContest := range frontList {
		contestIDList4FavComponent = append(contestIDList4FavComponent, frontContest.ID)
		if frontContest.GuessType == 1 {
			contestIDList4GuessComponent = append(contestIDList4GuessComponent, frontContest.ID)
		}
	}
	for _, middleContest := range middleList {
		contestIDList4FavComponent = append(contestIDList4FavComponent, middleContest.ID)
		if middleContest.GuessType == 1 {
			contestIDList4GuessComponent = append(contestIDList4GuessComponent, middleContest.ID)
		}
	}
	for _, backContest := range backList {
		contestIDList4FavComponent = append(contestIDList4FavComponent, backContest.ID)
		if backContest.GuessType == 1 {
			contestIDList4GuessComponent = append(contestIDList4GuessComponent, backContest.ID)
		}
	}
	if mid > 0 {
		subscribeMap := s.fetchFavoriteMap(ctx, mid, contestIDList4FavComponent)
		guessMap := s.fetchComponentContestGuessMap(ctx, mid, contestIDList4GuessComponent)
		for _, card := range frontList {
			if d, ok := subscribeMap[card.ID]; ok && d {
				card.IsSub = _haveSubscribe
			}
			if d, ok := guessMap[card.ID]; ok && d {
				card.IsGuess = _haveGuess
			}
			// GameState值依赖IsSub
			card.GameState = s.resetComponentContestGameState(card)
		}
		for _, card := range middleList {
			if d, ok := subscribeMap[card.ID]; ok && d {
				card.IsSub = _haveSubscribe
			}
			if d, ok := guessMap[card.ID]; ok && d {
				card.IsGuess = _haveGuess
			}
			// GameState值依赖IsSub
			card.GameState = s.resetComponentContestGameState(card)
		}
		for _, card := range backList {
			if d, ok := subscribeMap[card.ID]; ok && d {
				card.IsSub = _haveSubscribe
			}
			if d, ok := guessMap[card.ID]; ok && d {
				card.IsGuess = _haveGuess
			}
			// GameState值依赖IsSub
			card.GameState = s.resetComponentContestGameState(card)
		}
	}
	res = &model.ComponentContestCardFold{
		FrontList:  frontList,
		MiddleList: middleList,
		BackList:   backList,
	}
	return
}

func (s *Service) AbstractContests(ctx context.Context, mid int64, p *model.ParamAbstract) (res *model.ComponentContestAbstract, err error) {
	var allContests []*pb.ContestCardComponent
	if allContests, err = fetchComponentContestListAll(ctx, p.Sid); err != nil {
		log.Errorc(ctx, "contest component AbstractContests  fetchComponentContestListAll Param(%+v) error(%+v)", p, err)
		return
	}
	res = s.rebuildComponentContestAbstract(ctx, mid, allContests)
	return
}

/*
默认定位的比赛向前展示10条赛程，向后展示15条赛程
当向前取的赛程数量x＜10条，则向后取25-x
当向后取的赛程数量y＜15条，则向前取25-y
当向前取的赛程数量x＜10条，且向后取的赛程数量y＜15条，则取全部的赛程信息
*/
func (s *Service) generateComponentContestAbstract(historyAllList, futureAllList []*pb.ContestCardComponent) (historyList, futureList []*pb.ContestCardComponent) {
	historyCount := len(historyAllList)
	futureCount := len(futureAllList)
	if historyCount >= _historyAbstractSize && futureCount >= _futureAbstractSize {
		historyList = historyAllList[0:_historyAbstractSize]
		futureList = futureAllList[0:_futureAbstractSize]
	} else if historyCount < _historyAbstractSize && futureCount >= _futureAbstractSize+(_historyAbstractSize-historyCount) {
		historyList = historyAllList[0:historyCount]
		futureList = futureAllList[0 : _futureAbstractSize+(_historyAbstractSize-historyCount)]
	} else if futureCount < _futureAbstractSize && historyCount >= _historyAbstractSize+(_futureAbstractSize-futureCount) {
		historyList = historyAllList[0 : _historyAbstractSize+(_futureAbstractSize-futureCount)]
		futureList = futureAllList[0:futureCount]
	} else {
		historyList = historyAllList[0:historyCount]
		futureList = futureAllList[0:futureCount]
	}
	if len(historyList) == 0 {
		historyList = _emptyContestCardComponent
	}
	if len(futureList) == 0 {
		futureList = _emptyContestCardComponent
	}
	return
}

func (s *Service) resetComponentContestAbstractUserStatus(ctx context.Context, mid int64, contestIDs []int64, historyList, futureList []*pb.ContestCardComponent) (res *model.ComponentContestAbstract) {
	subscribeMap := s.fetchFavoriteMap(ctx, mid, contestIDs)
	guessMap := s.fetchComponentContestGuessMap(ctx, mid, contestIDs)
	for _, card := range historyList {
		if d, ok := subscribeMap[card.ID]; ok && d {
			card.IsSub = _haveSubscribe
		}
		if d, ok := guessMap[card.ID]; ok && d {
			card.IsGuess = _haveGuess
		}
	}
	for _, card := range futureList {
		if d, ok := subscribeMap[card.ID]; ok && d {
			card.IsSub = _haveSubscribe
		}
		if d, ok := guessMap[card.ID]; ok && d {
			card.IsGuess = _haveGuess
		}
	}
	res = &model.ComponentContestAbstract{
		History: historyList,
		Future:  futureList,
	}
	return
}

func (s *Service) rebuildComponentContestAbstract(ctx context.Context, mid int64, allContests []*pb.ContestCardComponent) (res *model.ComponentContestAbstract) {
	historyAllList, futureAllList, contestIDs := genHistoryFutureContest(allContests)
	history, future := s.generateComponentContestAbstract(historyAllList, futureAllList)
	res = &model.ComponentContestAbstract{
		History: history,
		Future:  future,
	}
	if len(contestIDs) == 0 || mid == 0 {
		return
	}
	res = s.resetComponentContestAbstractUserStatus(ctx, mid, contestIDs, history, future)
	return
}

func isGoingSeason(seasonID int64) (res bool) {
	for _, season := range goingSeasonsListGlobal {
		if seasonID == season.ID {
			res = true
			break
		}
	}
	return
}

func isGoingBattleSeason(seasonID int64) (res bool) {
	for _, season := range goingBattleSeasonsListGlobal {
		if seasonID == season.ID {
			res = true
			break
		}
	}
	return
}

func (s *Service) BattleContests(ctx context.Context, mid int64, p *model.ParamContestBattle) ([]*model.ContestBattleCardComponent, int, error) {
	cardContestList, total, err := loadComponentContestBattleCardAllRelations(ctx, p)
	if err != nil {
		return nil, 0, err
	}
	list, e := s.rebuildComponentContestBattleCardAll(ctx, mid, p, cardContestList)
	if e != nil {
		return nil, 0, e
	}
	return list, total, nil
}

func (s *Service) rebuildComponentContestBattleCardAll(ctx context.Context, mid int64, p *model.ParamContestBattle, list []*pb.ContestBattleCardComponent) (res []*model.ContestBattleCardComponent, err error) {
	res = make([]*model.ContestBattleCardComponent, 0)
	contestIDList4FavComponent := make([]int64, 0)
	contestIDList4GuessComponent := make([]int64, 0)
	for _, card := range list {
		contestIDList4FavComponent = append(contestIDList4FavComponent, card.ID)
		if card.GuessType == 1 {
			contestIDList4GuessComponent = append(contestIDList4GuessComponent, card.ID)
		}
	}
	subscribeMap := make(map[int64]bool, 0)
	guessMap := make(map[int64]bool, 0)
	if mid > 0 {
		subscribeMap = s.fetchFavoriteMap(ctx, mid, contestIDList4FavComponent)
		guessMap = s.fetchComponentContestGuessMap(ctx, mid, contestIDList4GuessComponent)
	}
	contestAllTeam, err := s.GetTeamsInfoBySeasonContests(ctx, p.Sid, list)
	if err != nil {
		log.Errorc(ctx, "contest component rebuildComponentContestBattleCardAll s.GetTeamsInfoByContestIds() sid(%d) error(%+v)", p.Sid, err)
		return
	}
	for _, contest := range list {
		battleCard := rebuildOneContestBattleTeam(contest, p.TeamTop, contestAllTeam, subscribeMap, guessMap)
		res = append(res, battleCard)
	}
	return
}

func fetchComponentContestBattleBySeasonID(ctx context.Context, sid int64) (res map[int64][]*pb.ContestBattleCardComponent, err error) {
	var contests []*model.ContestBattle2DBComponent
	if contests, err = match_component.FetchContestBattleBySeasonComponent(ctx, sid); err != nil {
		log.Errorc(ctx, "contest component fetchComponentContestBattleBySeasonID match_component.FetchContestBattleBySeasonComponent(%d) error(%+v)", sid, err)
		err = ecode.EsportsComponentErr
		return
	}
	// deep copy.
	tmpContestList := deepCopyContestBattleInfo(contests)
	res = generateComponentContestBattle4Frontend(tmpContestList)
	return
}

func deepCopyContestBattleInfo(list []*model.ContestBattle2DBComponent) []*model.ContestBattle2DBComponent {
	var tmpContestList []*model.ContestBattle2DBComponent
	for _, contest := range list {
		tmpContest := new(model.ContestBattle2DBComponent)
		*tmpContest = *contest
		tmpContestList = append(tmpContestList, tmpContest)
	}
	return tmpContestList
}

func generateComponentContestBattle4Frontend(contestList []*model.ContestBattle2DBComponent) map[int64][]*pb.ContestBattleCardComponent {
	componentContestCardList := make(map[int64][]*pb.ContestBattleCardComponent, 0)
	for _, contest := range contestList {
		tmpContestComponent := new(model.ContestBattle2DBComponent)
		*tmpContestComponent = *contest
		cardList := make([]*pb.ContestBattleCardComponent, 0)
		dateUnix := tmpContestComponent.StimeDate
		if d, ok := componentContestCardList[dateUnix]; ok {
			cardList = d
		}
		newCard := genComponentContestBattleCardByContest(tmpContestComponent)
		cardList = append(cardList, newCard)
		componentContestCardList[dateUnix] = cardList
	}
	return componentContestCardList
}

func genComponentContestBattleCardByContest(contest *model.ContestBattle2DBComponent) *pb.ContestBattleCardComponent {
	contestCard := new(pb.ContestBattleCardComponent)
	{
		contestCard.ID = contest.ID
		contestCard.Title = contest.GameStage
		contestCard.StartTime = contest.Stime
		contestCard.EndTime = contest.Etime
		contestCard.Status = contest.CalculateStatus()
		contestCard.CollectionURL = contest.CollectionUrl
		contestCard.LiveRoom = contest.LiveRoom
		contestCard.PlayBack = contest.PlayBack
		contestCard.MatchID = contest.MatchID
		contestCard.SeasonID = contest.SeasonID
		contestCard.GuessType = contest.GuessType
		contestCard.ContestFreeze = contest.Status
		contestCard.ContestStatus = contest.ContestStatus
		contestCard.GameStage = contest.GameStage
		if (contest.GuessType == _guessOk) && (contest.Stime-time.Now().Unix() > secondsOf10Minutes) {
			contestCard.GuessShow = _guessOk
		}
	}
	return contestCard
}

func loadComponentContestBattleCardAllRelations(ctx context.Context, p *model.ParamContestBattle) (res []*pb.ContestBattleCardComponent, total int, err error) {
	var (
		componentContestBattleList []*pb.ContestBattleCardComponent
		start                      = (p.Pn - 1) * p.Ps
		end                        = start + p.Ps - 1
	)
	res = make([]*pb.ContestBattleCardComponent, 0)
	if componentContestBattleList, err = fetchComponentContestBattleList(ctx, p); err != nil {
		return
	}
	total = len(componentContestBattleList)
	if total == 0 || total < start {
		return
	}
	currentContestList := make([]*pb.ContestBattleCardComponent, p.Ps)
	if total > end+1 {
		currentContestList = componentContestBattleList[start : end+1]
	} else {
		currentContestList = componentContestBattleList[start:]
	}
	tmpList := make([]*pb.ContestBattleCardComponent, 0)
	for _, v := range currentContestList {
		tmpCard := new(pb.ContestBattleCardComponent)
		*tmpCard = *v
		tmpList = append(tmpList, tmpCard)
	}
	res = tmpList
	return
}

func fetchComponentContestBattleList(ctx context.Context, p *model.ParamContestBattle) (res []*pb.ContestBattleCardComponent, err error) {
	var tmpRes []*pb.ContestBattleCardComponent
	if tmpRes, err = fetchComponentContestBattleAll(ctx, p.Sid); err != nil {
		log.Errorc(ctx, "contest component fetchComponentContestBattleList  fetchComponentContestListAll Param(%+v) error(%+v)", p, err)
		return
	}
	if p.Sort == _findAsc {
		sort.SliceStable(tmpRes, func(i, j int) bool {
			if tmpRes[i].StartTime != tmpRes[j].StartTime {
				return tmpRes[i].StartTime < tmpRes[j].StartTime
			}
			return tmpRes[i].ID < tmpRes[j].ID
		})
	} else if p.Sort == _findDesc {
		sort.SliceStable(tmpRes, func(i, j int) bool {
			if tmpRes[i].StartTime != tmpRes[j].StartTime {
				return tmpRes[i].StartTime > tmpRes[j].StartTime
			}
			return tmpRes[i].ID > tmpRes[j].ID
		})
	}
	// 根据开始时间判断.
	if p.Stime != 0 && p.Etime != 0 {
		for _, contest := range tmpRes {
			if contest.StartTime >= p.Stime && contest.StartTime <= p.Etime {
				res = append(res, contest)
			}
		}
	} else if p.Stime == 0 && p.Etime != 0 {
		for _, contest := range tmpRes {
			if contest.StartTime <= p.Etime {
				res = append(res, contest)
			}
		}
	} else if p.Stime != 0 && p.Etime == 0 {
		for _, contest := range tmpRes {
			if contest.StartTime >= p.Stime {
				res = append(res, contest)
			}
		}
	} else {
		res = tmpRes
	}
	return
}

func fetchComponentContestBattleAll(ctx context.Context, seasonID int64) (res []*pb.ContestBattleCardComponent, err error) {
	var (
		ok                        bool
		componentContestBattleAll []*pb.ContestBattleCardComponent
	)
	componentContestBattleAll, ok = seasonContestBattleAllComponentMap[seasonID]
	if !ok { // 回源.
		var componentContests map[int64][]*pb.ContestBattleCardComponent
		componentContests, err = fetchSeasonContestBattle(ctx, seasonID)
		if err != nil {
			return
		}
		componentContestBattleAll = genComponentContestBattle4All(componentContests)
	}
	res = deepCopyContestBattleAll(componentContestBattleAll)
	return
}

func fetchSeasonContestBattle(ctx context.Context, sid int64) (res map[int64][]*pb.ContestBattleCardComponent, err error) {
	if res, err = match_component.FetchContestBattleCardListFromCache(ctx, sid); err != nil && err != memcache.ErrNotFound {
		log.Errorc(ctx, "contest component fetchSeasonContestBattle match_component.FetchContestBattleCardListFromCache() sid(%d) error(%+v)", sid, err)
		return
	}
	if err == nil {
		return
	}
	if err == memcache.ErrNotFound {
		res, err = fetchComponentContestBattleBySeasonID(context.Background(), sid)
		if err != nil {
			return
		}
		if e := match_component.FetchContestBattleCardListToCache(ctx, sid, res, int32(tool.CalculateExpiredSeconds(10))); e != nil {
			log.Errorc(ctx, "contest component fetchSeasonContestBattle match_component.FetchContestBattleCardListToCache() sid(%d) error(%+v)", sid, e)
		}
	}
	return
}

func deepCopyContestBattleAll(contestAll []*pb.ContestBattleCardComponent) []*pb.ContestBattleCardComponent {
	tmpRes := make([]*pb.ContestBattleCardComponent, 0)
	for _, contest := range contestAll {
		tmpContest := new(pb.ContestBattleCardComponent)
		*tmpContest = *contest
		tmpRes = append(tmpRes, tmpContest)
	}
	return tmpRes
}

func genComponentContestBattle4All(m map[int64][]*pb.ContestBattleCardComponent) []*pb.ContestBattleCardComponent {
	componentAllContestBattle := make([]*pb.ContestBattleCardComponent, 0)
	if len(m) == 0 {
		return componentAllContestBattle
	}
	for _, contestCard := range m {
		componentAllContestBattle = append(componentAllContestBattle, contestCard...)
	}
	return componentAllContestBattle
}

func (s *Service) BattleContestTeams(ctx context.Context, mid int64, p *model.ParamBattleTeams) (res interface{}, err error) {
	var (
		tmpList    []*pb.ContestBattleCardComponent
		tmpContest *pb.ContestBattleCardComponent
	)
	if tmpList, err = fetchComponentContestBattleAll(ctx, p.Sid); err != nil {
		log.Errorc(ctx, "contest component BattleContestTeams fetchComponentContestBattleAll Param(%+v) error(%+v)", p, err)
		return
	}
	for _, contest := range tmpList {
		if contest.ID == p.ContestID {
			tmpContest = contest
			break
		}
	}
	if tmpContest == nil {
		res = struct{}{}
		return
	}
	if res, err = s.rebuildComponentContestBattleTeam(ctx, mid, tmpContest, 0); err != nil {
		log.Errorc(ctx, "contest component BattleContestTeams s.rebuildComponentContestBattleTeam() param(%+v) error(%+v)", p, err)
		return
	}
	return
}

func (s *Service) rebuildComponentContestBattleTeam(ctx context.Context, mid int64, contestBattle *pb.ContestBattleCardComponent, teamCount int) (res *model.ContestBattleCardComponent, err error) {
	contests := []*pb.ContestBattleCardComponent{contestBattle}
	contestAllTeam, err := s.GetTeamsInfoBySeasonContests(ctx, contestBattle.SeasonID, contests)
	if err != nil {
		log.Errorc(ctx, "contest component rebuildComponentContestBattleTeam s.GetTeamsInfoByContestIds() sid(%d) error(%+v)", contestBattle.SeasonID, err)
		return
	}
	contestIDList4FavComponent := make([]int64, 0)
	contestIDList4GuessComponent := make([]int64, 0)
	contestIDList4FavComponent = append(contestIDList4FavComponent, contestBattle.ID)
	if contestBattle.GuessType == 1 {
		contestIDList4GuessComponent = append(contestIDList4GuessComponent, contestBattle.ID)
	}
	subscribeMap := make(map[int64]bool, 0)
	guessMap := make(map[int64]bool, 0)
	if mid > 0 {
		subscribeMap = s.fetchFavoriteMap(ctx, mid, contestIDList4FavComponent)
		guessMap = s.fetchComponentContestGuessMap(ctx, mid, contestIDList4GuessComponent)
	}
	res = rebuildOneContestBattleTeam(contestBattle, teamCount, contestAllTeam, subscribeMap, guessMap)
	return
}

func rebuildOneContestBattleTeam(contestBattle *pb.ContestBattleCardComponent, teamCount int, contestAllTeam map[int64][]*model.ContestTeamInfo, subscribeMap, guessMap map[int64]bool) (res *model.ContestBattleCardComponent) {
	tmpContest := new(pb.ContestBattleCardComponent)
	*tmpContest = *contestBattle
	if isSub, ok := subscribeMap[tmpContest.ID]; isSub && ok {
		tmpContest.IsSub = _haveSubscribe
	}
	if isGuess, ok := guessMap[tmpContest.ID]; isGuess && ok {
		tmpContest.IsGuess = _haveGuess
	}
	teamList := make([]*model.ContestBattleTeam, 0)
	contestTeams, ok := contestAllTeam[tmpContest.ID]
	if !ok {
		res = &model.ContestBattleCardComponent{
			ContestBattleCardComponent: tmpContest,
			TeamList:                   make([]*model.ContestBattleTeam, 0)}
		return
	}
	isScore := false
	for index, teamResult := range contestTeams {
		if !isScore && checkIsScore(tmpContest, teamResult.ScoreInfo.SurvivalRank) {
			isScore = true
		}
		if teamCount > 0 && index == teamCount {
			break
		}
		teamList = append(teamList, &model.ContestBattleTeam{
			Title:                teamResult.TeamInfo.Title,
			Logo:                 teamResult.TeamInfo.Logo,
			ContestTeamScoreInfo: teamResult.ScoreInfo,
		})
	}
	res = &model.ContestBattleCardComponent{
		ContestBattleCardComponent: tmpContest,
		TeamCount:                  len(contestTeams),
		IsScore:                    isScore,
		TeamList:                   teamList,
	}
	return
}

func (s *Service) TeamContests(ctx context.Context, mid int64, p *model.ParamTeamContest) ([]*pb.ContestCardComponent, int, error) {
	cardContestList, total, err := loadComponentTeamContestCardRelations(ctx, p)
	if err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return cardContestList, total, nil
	}
	list := s.rebuildComponentContestCardAll(ctx, mid, cardContestList)
	return list, total, nil
}

func loadComponentTeamContestCardRelations(ctx context.Context, p *model.ParamTeamContest) (res []*pb.ContestCardComponent, total int, err error) {
	var (
		componentContestsList []*pb.ContestCardComponent
		start                 = (p.Pn - 1) * p.Ps
		end                   = start + p.Ps - 1
	)
	res = make([]*pb.ContestCardComponent, 0)
	if componentContestsList, err = fetchComponentTeamContests(ctx, p); err != nil {
		return
	}
	total = len(componentContestsList)
	if total == 0 || total < start {
		return
	}
	currentContestList := make([]*pb.ContestCardComponent, p.Ps)
	if total > end+1 {
		currentContestList = componentContestsList[start : end+1]
	} else {
		currentContestList = componentContestsList[start:]
	}
	tmpList := make([]*pb.ContestCardComponent, 0)
	for _, v := range currentContestList {
		tmpCard := new(pb.ContestCardComponent)
		*tmpCard = *v
		tmpList = append(tmpList, tmpCard)
	}
	res = tmpList
	return
}

func fetchComponentTeamContests(ctx context.Context, p *model.ParamTeamContest) (res []*pb.ContestCardComponent, err error) {
	var tmpRes []*pb.ContestCardComponent
	if tmpRes, err = fetchComponentContestListAll(ctx, p.Sid); err != nil {
		log.Errorc(ctx, "contest component fetchComponentTeamContests  fetchComponentContestListAll Param(%+v) error(%+v)", p, err)
		return
	}
	for _, contest := range tmpRes {
		if isSeriesTeamContest(p, contest) {
			res = append(res, contest)
		}
	}
	if p.Sort == _findAsc {
		sort.SliceStable(res, func(i, j int) bool {
			if res[i].StartTime != res[j].StartTime {
				return res[i].StartTime < res[j].StartTime
			}
			return res[i].ID < res[j].ID
		})
	} else if p.Sort == _findDesc {
		sort.SliceStable(res, func(i, j int) bool {
			if res[i].StartTime != res[j].StartTime {
				return res[i].StartTime > res[j].StartTime
			}
			return res[i].ID > res[j].ID
		})
	}
	return
}

func isSeriesTeamContest(p *model.ParamTeamContest, contest *pb.ContestCardComponent) bool {
	// 不分组积分表
	if p.GroupName == "" {
		return contest.SeriesID == p.SeriesID
	}
	// 分组积分表根据阶段ID与战队判断.
	if contest.SeriesID != p.SeriesID {
		return false
	}
	mapTeam := make(map[int64]struct{}, len(p.TeamIDs))
	for _, teamID := range p.TeamIDs {
		mapTeam[teamID] = struct{}{}
	}
	if _, ok := mapTeam[contest.Home.ID]; !ok {
		return false
	}
	if _, ok := mapTeam[contest.Away.ID]; !ok {
		return false
	}
	return true
}

func (s *Service) TeamContestsV2(ctx context.Context, mid int64, p *model.ParamV2TeamContest) (res *model.ComponentSeasonContests, err error) {
	var allContests []*pb.ContestCardComponent
	if p.Prev > 0 && p.Next > 0 {
		err = xecode.RequestErr
		return
	}
	if allContests, err = fetchComponentTeamContestsV2(ctx, p); err != nil {
		log.Errorc(ctx, "TeamContestsV2 fetchComponentTeamContestsV2 Param(%+v) error(%+v)", p, err)
		return
	}
	ParamSeason := &model.ParamSeasonContests{
		Sid:  p.Sid,
		Ps:   p.Ps,
		Prev: p.Prev,
		Next: p.Next,
	}
	res = s.rebuildComponentSeasonContests(ctx, mid, ParamSeason, allContests)
	return
}

func fetchComponentTeamContestsV2(ctx context.Context, p *model.ParamV2TeamContest) (res []*pb.ContestCardComponent, err error) {
	tmpAllContest, err := fetchComponentContestListAll(ctx, p.Sid)
	if err != nil {
		log.Errorc(ctx, "contest component fetchComponentTeamContestsV2 Param(%+v) error(%+v)", p, err)
		return
	}
	paramTeam := &model.ParamTeamContest{
		Sid:       p.Sid,
		SeriesID:  p.SeriesID,
		TeamIDs:   p.TeamIDs,
		GroupName: p.GroupName,
	}
	tmpList := make([]*pb.ContestCardComponent, 0)
	for _, contest := range tmpAllContest {
		if isSeriesTeamContest(paramTeam, contest) {
			tmpCard := new(pb.ContestCardComponent)
			*tmpCard = *contest
			tmpList = append(tmpList, tmpCard)
		}
	}
	res = tmpList
	return
}

// HomeAwayContest .
func (s *Service) HomeAwayContest(ctx context.Context, param *model.ParamEsGuess) (res *model.HomeAwayContestComponent, err error) {
	var teamContests []*model.Contest
	allTeamMap := s.GetAllTeamsOfComponent(ctx)
	teamCount := len(allTeamMap)
	if allTeamMap == nil || teamCount == 0 {
		err = xecode.RequestErr
		return
	}
	teamMap := make(map[int64]*model.Team2TabComponent, teamCount)
	for _, team := range allTeamMap {
		teamMap[team.ID] = team
	}
	homeTeam := new(model.Team2TabComponent)
	awayTeam := new(model.Team2TabComponent)
	if team, ok := teamMap[param.HomeID]; !ok {
		err = xecode.RequestErr
		return
	} else {
		*homeTeam = *team
	}
	if team, ok := teamMap[param.AwayID]; !ok {
		err = xecode.RequestErr
		return
	} else {
		*awayTeam = *team
	}
	if teamContests, err = fetchHomeAwayContestList(ctx, param); err != nil {
		log.Errorc(ctx, "contest component fetchHomeAwayContestList param(%+v) error(%+v)", param, err)
		return
	}
	res = rebuildHomeAwayContest(homeTeam, awayTeam, teamContests, param.Ps)
	return
}

func fetchHomeAwayContestList(ctx context.Context, param *model.ParamEsGuess) (teamContests []*model.Contest, err error) {
	var allContest []*model.Contest
	if allContest, err = homeAwayContestList(ctx, param); err != nil {
		log.Errorc(ctx, "fetchHomeAwayContestList homeAwayContestList() param(%+v) error(%+v)", param, err)
		return
	}
	for _, contest := range allContest {
		if contest.ID != param.CID {
			tmpContest := new(model.Contest)
			*tmpContest = *contest
			teamContests = append(teamContests, tmpContest)
		}
	}
	return
}

func homeAwayContestList(ctx context.Context, param *model.ParamEsGuess) (teamContests []*model.Contest, err error) {
	if teamContests, err = match_component.FetchHomeAwayContestsFromCache(ctx, param.HomeID, param.AwayID); err != nil && err != memcache.ErrNotFound {
		log.Errorc(ctx, "contest component fetchHomeAwayContestList match_component.FetchHomeAwayContestsFromCache() param(%+v) error(%+v)", param, err)
		return
	}
	if err == nil {
		return
	}
	if err == memcache.ErrNotFound {
		if teamContests, err = match_component.RawHoweAwayContest(ctx, param); err != nil {
			log.Errorc(ctx, "contest component HomeAwayContest match_component.RawHoweAwayContest param(%+v) error(%+v)", param, err)
			return
		}
		if e := match_component.FetchHomeAwayContestsToCache(ctx, param.HomeID, param.AwayID, teamContests); e != nil {
			log.Errorc(ctx, "contest component HomeAwayContest match_component.FetchHomeAwayContestsToCache param(%+v) error(%+v)", param, err)
		}
	}
	return
}

func rebuildHomeAwayContest(home, away *model.Team2TabComponent, teamContests []*model.Contest, ps int) (res *model.HomeAwayContestComponent) {
	var homeSuccess, awaySuccess int
	res = &model.HomeAwayContestComponent{
		SuccessList: make([]*model.HomeAwaySuccessContest, 0),
	}
	successContests := make([]*model.HomeAwaySuccessContest, 0)
	for _, contest := range teamContests {
		if len(successContests) == ps {
			break
		}
		if contest.HomeScore > contest.AwayScore {
			homeSuccess++
			successContests = append(successContests, &model.HomeAwaySuccessContest{
				Team2TabComponent: home,
				ContestStime:      contest.Stime,
			})
		} else if contest.HomeScore < contest.AwayScore {
			awaySuccess++
			successContests = append(successContests, &model.HomeAwaySuccessContest{
				Team2TabComponent: away,
				ContestStime:      contest.Stime,
			})
		}
	}
	res = &model.HomeAwayContestComponent{
		HomeTeam: &model.HomeAwayTeam{
			Team2TabComponent: home,
			WinCount:          homeSuccess,
		},
		AwayTeam: &model.HomeAwayTeam{
			Team2TabComponent: away,
			WinCount:          awaySuccess,
		},
		SuccessList: successContests,
	}
	return
}

func (s *Service) SeasonTeamsComponent(ctx context.Context, seasonId int64) (res *model.SeasonTeams2Component, err error) {
	var (
		tmpTeams []*model.TeamInSeason
		teams    = make([]*model.TeamInSeason, 0)
	)
	seasonsMap := s.GetAllSeasonsOfComponent(ctx)
	season, ok := seasonsMap[seasonId]
	if !ok {
		err = xecode.RequestErr
		return
	}
	if tmpTeams, err = s.GetTeamsInSeason(ctx, seasonId); err != nil {
		log.Errorc(ctx, "SeasonTeamsComponent  s.GetTeamsInSeason() sid(%d) error(%+v)", seasonId, err)
		return
	}
	if len(tmpTeams) > 0 {
		for _, team := range tmpTeams {
			if tool.Int64InSlice(team.TeamId, conf.Conf.SeriesIgnoreTeamsIDList) {
				continue
			}
			teams = append(teams, team)
		}
	}
	res = &model.SeasonTeams2Component{
		Season: season,
		Teams:  teams,
	}
	return
}

// FetchEsTopicVideoList .
func (s *Service) FetchEsTopicVideoList(ctx context.Context, param *pb.EsTopicVideoListRequest) (res *pb.EsTopicVideoListReply, err error) {
	if res, err = match_component.FetchEsTopicVideoListFromCache(ctx, param); err != nil && err != memcache.ErrNotFound {
		log.Errorc(ctx, "contest component FetchEsTopicVideoList match_component.FetchEsTopicVideoListFromCache() param(%+v) error(%+v)", param, err)
		return
	}
	if err == nil {
		return
	}
	if err == memcache.ErrNotFound {
		res, err = s.EsTopicVideoList(ctx, param)
		if err != nil {
			log.Errorc(ctx, "contest component FetchEsTopicVideoList s.EsTopicVideoList() param(%+v) error(%+v)", param, err)
			return
		}
		if e := match_component.FetchEsTopicVideoListToCache(ctx, param, res, 300); e != nil {
			log.Errorc(ctx, "contest component FetchEsTopicVideoList match_component.FetchEsTopicVideoListToCache() param(%+v) error(%+v)", param, err)
		}
	}
	return
}

func (s *Service) FetchHotVideoListFromCache(param *model.ParamVideoList) (res *model.VideoList2Component, isHot bool) {
	var (
		start        = (param.Pn - 1) * param.Ps
		end          = start + param.Ps - 1
		ugcList      = _emptyVideoComponent
		esList       []*model.Video
		firstPageRes *model.VideoList2Component
	)
	res = &model.VideoList2Component{
		UgcList: ugcList,
		VideoList: &model.VideoList{
			List: _emptyVideoComponent,
			Page: &model.Page{
				Num:  param.Pn,
				Size: param.Ps,
			},
		},
	}
	if start+param.Ps > _videoListPs {
		return
	}
	if firstPageRes, isHot = goingVideoListComponentMap[param.ID]; !isHot {
		return
	}
	if param.Pn == _firstPage {
		ugcList = firstPageRes.UgcList
	}
	searchCout := len(firstPageRes.List)
	if searchCout == 0 || searchCout < start {
		esList = _emptyVideoComponent
	} else if searchCout > end+1 {
		esList = firstPageRes.List[start : end+1]
	} else {
		esList = firstPageRes.List[start:]
	}
	res = &model.VideoList2Component{
		UgcList: ugcList,
		VideoList: &model.VideoList{
			List: esList,
			Page: &model.Page{
				Num:   param.Pn,
				Size:  param.Ps,
				Total: firstPageRes.Page.Total,
			},
		},
	}
	return
}

func (s *Service) TopicVideoListComponent(ctx context.Context, param *model.ParamVideoList) (res *model.VideoList2Component, err error) {
	var (
		videoListInfo *model.VideoListInfo
		isHot         bool
	)
	res, isHot = s.FetchHotVideoListFromCache(param)
	if isHot {
		return
	}
	if videoListInfo, err = s.dao.VideoList(ctx, param.ID); err != nil {
		log.Errorc(ctx, "TopicVideoList s.dao.VideoList() Id(%d) error(%+v)", param.ID, err)
		return
	}
	if videoListInfo == nil {
		res = &model.VideoList2Component{
			UgcList: _emptyVideoComponent,
			VideoList: &model.VideoList{
				List: _emptyVideoComponent,
				Page: &model.Page{
					Num:  param.Pn,
					Size: param.Ps,
				},
			},
		}
		return
	}
	arg := &pb.EsTopicVideoListRequest{
		GameId:  videoListInfo.GameID,
		MatchId: videoListInfo.MatchID,
		YearId:  videoListInfo.YearID,
		Pn:      int64(param.Pn),
		Ps:      int64(param.Ps),
	}
	if res, err = s.rebuildTopicVideoList(ctx, videoListInfo.UgcAids, arg); err != nil {
		log.Errorc(ctx, "TopicVideoListComponent s.rebuildTopicVideoList() param(%+v) error(%+v)", param, err)
		return
	}
	return
}

func (s *Service) rebuildTopicVideoList(ctx context.Context, strUgcAids string, arg *pb.EsTopicVideoListRequest) (res *model.VideoList2Component, err error) {
	var (
		videoListEs           *pb.EsTopicVideoListReply
		total                 int
		ugcArcList, esArcList []*model.Video
		searchAids            []int64
	)
	if videoListEs, err = s.FetchEsTopicVideoList(ctx, arg); err != nil {
		log.Errorc(ctx, "rebuildTopicVideoList s.FetchEsTopicVideoList() arg(%+v) error(%+v)", arg, err)
		return
	}
	if videoListEs != nil && videoListEs.Page != nil {
		total = int(videoListEs.Page.GetCount())
	}
	res = &model.VideoList2Component{
		UgcList: _emptyVideoComponent,
		VideoList: &model.VideoList{
			List: _emptyVideoComponent,
			Page: &model.Page{
				Num:   int(arg.Pn),
				Size:  int(arg.Ps),
				Total: total,
			},
		},
	}
	if videoListEs != nil {
		searchAids = videoListEs.SearchAids
	}
	if ugcArcList, esArcList, err = s.formatVideoListArcInfo(ctx, strUgcAids, searchAids); err != nil {
		log.Errorc(ctx, "rebuildTopicVideoLists.formatVideoListArcInfo() error(%+v)", err)
		return
	}
	if arg.Pn == _firstPage && len(ugcArcList) > 0 {
		res.UgcList = ugcArcList
	}
	if len(esArcList) > 0 {
		res.List = esArcList
	}
	return
}

func getUgcInfo(ctx context.Context, strUgcAids string) (ugcAidMap map[int64]struct{}, ugcAids []int64, err error) {
	var (
		bvidMap    map[string]int64
		archiveIDs []string
	)
	ugcAidMap = make(map[int64]struct{})
	if strUgcAids != "" {
		archiveIDs = strings.Split(strUgcAids, ",")
		if bvidMap, err = helper.BvidsToAid(ctx, archiveIDs); err != nil {
			log.Errorc(ctx, "TopicVideoListComponent rebuildTopicVideoList UgcAids helper.BvidsToAid() archiveIDs(%+v) error(%+v)", archiveIDs, err)
			return
		}
		for _, aid := range bvidMap {
			ugcAidMap[aid] = struct{}{}
		}
		for _, strAid := range archiveIDs {
			if intAid, ok := bvidMap[strAid]; ok {
				ugcAids = append(ugcAids, intAid)
			}
		}
	}
	return
}

func (s *Service) formatVideoListArcInfo(ctx context.Context, strUgcAids string, esAids []int64) (ugcArcInfo, esArcInfo []*model.Video, err error) {
	var (
		ugcAids   []int64
		ugcAidMap map[int64]struct{}
	)
	ugcArcInfo = _emptActVideos
	esArcInfo = _emptActVideos
	if ugcAidMap, ugcAids, err = getUgcInfo(ctx, strUgcAids); err != nil {
		log.Errorc(ctx, "formatVideoListArcInfo getUgcInfo() strUgcAids(%+v) error(%+v)", strUgcAids, err)
		return
	}
	group := errgroup.WithContext(ctx)
	group.Go(func(ctx context.Context) (ugcErr error) {
		if ugcArcInfo, ugcErr = s.ArcsInfo(ctx, ugcAids); ugcErr != nil {
			log.Errorc(ctx, "formatVideoListArcInfo  UgcAids s.batchArchives() ugcAids(%+v) error(%+v)", ugcAids, ugcErr)
			return
		}
		return nil
	})
	group.Go(func(ctx context.Context) (searchErr error) {
		var searchAids []int64
		for _, aid := range esAids {
			if _, ok := ugcAidMap[aid]; !ok {
				searchAids = append(searchAids, aid)
			}
		}
		if esArcInfo, searchErr = s.ArcsInfo(ctx, searchAids); err != nil {
			log.Errorc(ctx, "formatVideoListArcInfo  s.ArcsInfo() searchAids(%+v) error(%+v)", searchAids, searchErr)
			return
		}
		return nil
	})
	if err = group.Wait(); err != nil {
		log.Errorc(ctx, "formatVideoListArcInfo errGroup (%+v)", err)
	}
	return
}

func (s *Service) SeasonContests(ctx context.Context, mid int64, p *model.ParamSeasonContests) (res *model.ComponentSeasonContests, err error) {
	var allContests []*pb.ContestCardComponent
	if p.Prev > 0 && p.Next > 0 {
		err = xecode.RequestErr
		return
	}
	if allContests, err = fetchComponentContestListAll(ctx, p.Sid); err != nil {
		log.Errorc(ctx, "contest component SeasonContests  fetchComponentContestListAll Param(%+v) error(%+v)", p, err)
		return
	}
	res = s.rebuildComponentSeasonContests(ctx, mid, p, allContests)
	return
}

func genHistoryFutureContest(allContests []*pb.ContestCardComponent) (historyAllList, futureAllList []*pb.ContestCardComponent, contestIDs []int64) {
	historyAllList = _emptyContestCardComponent
	futureAllList = _emptyContestCardComponent
	nowTime := time.Now().Unix()
	for _, contest := range allContests {
		tmpContest := new(pb.ContestCardComponent)
		*tmpContest = *contest
		contestIDs = append(contestIDs, tmpContest.ID)
		if tmpContest.StartTime < nowTime && tmpContest.EndTime < nowTime { // 已结束.
			historyAllList = append(historyAllList, tmpContest)
		} else {
			futureAllList = append(futureAllList, tmpContest) // 进行中与未结束赛程.
		}
	}
	// history 按开始时间从大到小排序
	sort.SliceStable(historyAllList, func(i, j int) bool {
		if historyAllList[i].StartTime != historyAllList[j].StartTime {
			return historyAllList[i].StartTime > historyAllList[j].StartTime
		}
		return historyAllList[i].ID < historyAllList[j].ID
	})
	// future 按开始时间从小到大排序
	sort.SliceStable(futureAllList, func(i, j int) bool {
		if futureAllList[i].StartTime != futureAllList[j].StartTime {
			return futureAllList[i].StartTime < futureAllList[j].StartTime
		}
		return futureAllList[i].ID < futureAllList[j].ID
	})
	return
}

func (s *Service) rebuildComponentSeasonContests(ctx context.Context, mid int64, p *model.ParamSeasonContests, allContests []*pb.ContestCardComponent) (res *model.ComponentSeasonContests) {
	historyAllList, futureAllList, contestIDs := genHistoryFutureContest(allContests)
	history, future, historyMore, futureMore := s.generateComponentSeasonContests(p, historyAllList, futureAllList)
	prev := -1
	next := -1
	historyCount := len(history)
	futureCount := len(future)
	if historyCount == 0 {
		history = _emptyContestCardComponent
	} else if historyCount >= p.Ps && historyMore {
		prev = int(history[historyCount-1].ID)
	}
	if futureCount == 0 {
		future = _emptyContestCardComponent
	} else if futureCount >= p.Ps && futureMore {
		next = int(future[futureCount-1].ID)
	}
	res = &model.ComponentSeasonContests{
		History: history,
		Future:  future,
		Prev:    prev,
		Next:    next,
	}
	if len(contestIDs) == 0 || mid == 0 || p.Prev < 0 || p.Next < 0 {
		return
	}
	contestAbstractUserStatus := s.resetComponentContestAbstractUserStatus(ctx, mid, contestIDs, history, future)
	res.History = contestAbstractUserStatus.History
	res.Future = contestAbstractUserStatus.Future
	return
}

func defaultComponentSeasonContests(p *model.ParamSeasonContests, historyAllList, futureAllList []*pb.ContestCardComponent) (historyList, futureList []*pb.ContestCardComponent, historyMore, futureMore bool) {
	historyCount := len(historyAllList)
	futureCount := len(futureAllList)
	if historyCount >= p.Ps {
		historyList = historyAllList[0:p.Ps]
	} else if historyCount > 0 {
		historyList = historyAllList[0:historyCount]
	}
	historyMore = historyCount > p.Ps
	if futureCount >= p.Ps {
		futureList = futureAllList[0:p.Ps]
	} else if futureCount > 0 {
		futureList = futureAllList[0:futureCount]
	}
	futureMore = futureCount > p.Ps
	return
}

func prevComponentSeasonContests(p *model.ParamSeasonContests, historyAllList []*pb.ContestCardComponent) (historyList, futureList []*pb.ContestCardComponent, historyMore, futureMore bool) {
	isStart := false
	for _, contest := range historyAllList {
		if len(historyList) == p.Ps {
			historyMore = true
			return
		}
		if contest.ID == int64(p.Prev) {
			isStart = true
			continue
		}
		if !isStart {
			continue
		}
		historyList = append(historyList, contest)
	}
	return
}

func nextComponentSeasonContests(p *model.ParamSeasonContests, futureAllList []*pb.ContestCardComponent) (historyList, futureList []*pb.ContestCardComponent, historyMore, futureMore bool) {
	isStart := false
	for _, contest := range futureAllList {
		if len(futureList) == p.Ps {
			futureMore = true
			return
		}
		if contest.ID == int64(p.Next) {
			isStart = true
			continue
		}
		if !isStart {
			continue
		}
		futureList = append(futureList, contest)
	}
	return
}

func (s *Service) generateComponentSeasonContests(p *model.ParamSeasonContests, historyAllList, futureAllList []*pb.ContestCardComponent) (historyList, futureList []*pb.ContestCardComponent, historyMore, futureMore bool) {
	// default
	if p.Prev == 0 && p.Next == 0 {
		return defaultComponentSeasonContests(p, historyAllList, futureAllList)
	}
	// prev
	if p.Prev > 0 {
		return prevComponentSeasonContests(p, historyAllList)
	}
	// next
	if p.Next > 0 {
		return nextComponentSeasonContests(p, futureAllList)
	}
	return
}

func (s *Service) ContestReplyWall(ctx context.Context, p *model.ParamWall) (res *model.ComponentContestWall, err error) {
	var tmpAllContest []*pb.ContestCardComponent
	res = &model.ComponentContestWall{}
	if tmpAllContest, err = fetchComponentContestListAll(ctx, p.Sid); err != nil {
		log.Errorc(ctx, "contest component ContestWall()  fetchComponentContestListAll() Param(%+v) error(%+v)", p, err)
		return
	}
	if len(tmpAllContest) == 0 {
		res.Contest = struct{}{}
		return
	}
	tmpContest := rebuildContestWall(p.RoomID, tmpAllContest)
	if tmpContest == nil {
		res.Contest = struct{}{}
		return
	}
	res.Contest = tmpContest
	return
}

func rebuildContestWall(roomID int64, tmpAllContest []*pb.ContestCardComponent) (res *pb.ContestCardComponent) {
	var (
		roomGoingContest      []*pb.ContestCardComponent
		roomEndContest        []*pb.ContestCardComponent
		roomNotStartContest   []*pb.ContestCardComponent
		noRoomGoingContest    []*pb.ContestCardComponent
		noRoomEndContest      []*pb.ContestCardComponent
		noRoomNotStartContest []*pb.ContestCardComponent
	)
	for _, contest := range tmpAllContest {
		tmpCard := new(pb.ContestCardComponent)
		*tmpCard = *contest
		switch roomID {
		case contest.LiveRoom:
			switch contest.ContestStatus {
			case ContestStatusOngoing:
				roomGoingContest = append(roomGoingContest, tmpCard)
			case ContestStatusEnd:
				roomEndContest = append(roomEndContest, tmpCard)
			default:
				roomNotStartContest = append(roomNotStartContest, tmpCard)
			}
		default:
			switch contest.ContestStatus {
			case ContestStatusOngoing:
				noRoomGoingContest = append(noRoomGoingContest, tmpCard)
			case ContestStatusEnd:
				noRoomEndContest = append(noRoomEndContest, tmpCard)
			default:
				noRoomNotStartContest = append(noRoomNotStartContest, tmpCard)
			}
		}
	}
	// 直播间赛程
	res = getReplyWallContest(roomGoingContest, roomEndContest, roomNotStartContest)
	if res != nil {
		return
	}
	// 兜底取非直播间赛程
	res = getReplyWallContest(noRoomGoingContest, noRoomEndContest, noRoomNotStartContest)
	return
}

func getReplyWallContest(goingContest, endContest, notStartContest []*pb.ContestCardComponent) (contest *pb.ContestCardComponent) {
	// 取进行中
	if len(goingContest) > 0 {
		sort.SliceStable(goingContest, func(i, j int) bool {
			if goingContest[i].StartTime != goingContest[j].StartTime {
				return goingContest[i].StartTime > goingContest[j].StartTime
			}
			return goingContest[i].ID > goingContest[j].ID
		})
		contest = goingContest[0]
		return
	}
	sort.SliceStable(endContest, func(i, j int) bool {
		if endContest[i].StartTime != endContest[j].StartTime {
			return endContest[i].EndTime > endContest[j].EndTime //已结束的赛程的结束时间
		}
		return endContest[i].ID > endContest[j].ID
	})
	sort.SliceStable(notStartContest, func(i, j int) bool {
		if notStartContest[i].StartTime != notStartContest[j].StartTime {
			return notStartContest[i].StartTime < notStartContest[j].StartTime // 未开始的赛程的开始时间
		}
		return notStartContest[i].ID < notStartContest[j].ID
	})
	contest = getNearestContest(endContest, notStartContest)
	return
}

func getNearestContest(endContest, notStartContest []*pb.ContestCardComponent) (res *pb.ContestCardComponent) {
	endCount := len(endContest)
	notStartCount := len(notStartContest)
	if endCount == 0 && notStartCount == 0 {
		return
	}
	nowTime := time.Now().Unix()
	// 近一场的已结束的赛程的结束时间和未开始的赛程的开始时间和当前时间比，哪个更近返回哪个
	if endCount > 0 && notStartCount > 0 {
		if nowTime-endContest[0].EndTime > notStartContest[0].StartTime-nowTime {
			res = notStartContest[0]
		} else {
			res = endContest[0]
		}
		return
	}
	if endCount > 0 {
		res = endContest[0]
	}
	if notStartCount > 0 {
		res = notStartContest[0]
	}
	return
}
