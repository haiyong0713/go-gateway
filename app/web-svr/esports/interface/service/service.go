package service

import (
	"context"
	"regexp"
	"sync/atomic"
	"time"

	"go-common/library/log"
	"go-common/library/sync/pipeline/fanout"
	arcclient "go-gateway/app/app-svr/archive/service/api"
	actclient "go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/esports/interface/conf"
	"go-gateway/app/web-svr/esports/interface/dao"
	"go-gateway/app/web-svr/esports/interface/model"
	esportsServiceClient "go-gateway/app/web-svr/esports/service/api/v1"
	favclient "go-main/app/community/favorite/service/api"

	accClient "git.bilibili.co/bapis/bapis-go/account/service"
	coinclient "git.bilibili.co/bapis/bapis-go/community/service/coin"
	livexroom "git.bilibili.co/bapis/bapis-go/live/xroom"

	"github.com/robfig/cron"
)

const (
	_lolType       = 1
	_dotaType      = 2
	_owType        = 3
	_lolItems      = "lol/items"
	_dotaItems     = "dota2/items"
	_owMaps        = "overwatch/maps"
	_lolChampions  = "lol/champions"
	_dotaHeroes    = "dota2/heroes"
	_owHeroes      = "overwatch/heroes"
	_lolSpells     = "lol/spells"
	_dotaAbilities = "dota2/abilities"
	_lolPlayers    = "lol/players"
	_dotaPlayers   = "dota2/players"
	_owPlayers     = "overwatch/players"
	_lolTeams      = "lol/teams"
	_dotaTeams     = "dota2/teams"
	_owTeams       = "overwatch/teams"
)

// Service service struct.
type Service struct {
	c   *conf.Config
	dao *dao.Dao
	// cache proc
	cache                                *fanout.Fanout
	arcClient                            arcclient.ArchiveClient
	favClient                            favclient.FavoriteClient
	lolItemsMap, dotaItemsMap, owMapsMap *model.SyncItem
	lolChampions, dotaHeroes, owHeroes   *model.SyncInfo
	lolSpells, dotaAbilities             *model.SyncInfo
	lolPlayers, dotaPlayers, owPlayers   *model.SyncInfo
	lolTeams, dotaTeams, owTeams         *model.SyncInfo
	//lolBigPlayers                        *model.SyncLolPlayers
	//lolBigTeams                          *model.SyncLolTeams
	dotaBigPlayers *model.SyncDotaPlayers
	dotaBigTeams   *model.SyncDotaTeams
	ldSeasonGame   *model.SyncSeasonGame
	s9Contests     []*model.Contest
	// cron
	cron                 *cron.Cron
	mapGameDb            map[int64]int64
	actClient            actclient.ActivityClient
	accClient            accClient.AccountClient
	coinclient           coinclient.CoinClient
	liveClient           livexroom.RoomClient
	esportsServiceClient esportsServiceClient.EsportsServiceClient
	regBv, regAv         *regexp.Regexp
	liveMatchID          string
	liveBattleListMap    *atomic.Value
	liveBattleInfoMap    *atomic.Value
	teamsInSeasonMap     map[int64] /*seasonId*/ []*model.TeamInSeason
	matchSeasonMap       map[int64][]*model.MatchSeason
	seasonMap            map[int64]*model.MatchSeason
	reserveMap           map[int64]int64
}

// New new service.
func New(c *conf.Config) *Service {
	// this ctx will control all async job
	//srvCtx := context.Background()
	s := &Service{
		c:     c,
		dao:   dao.New(c),
		cache: fanout.New("cache"),
		lolItemsMap: &model.SyncItem{
			Data: make(map[int64]*model.LdInfo),
		},
		dotaItemsMap: &model.SyncItem{
			Data: make(map[int64]*model.LdInfo),
		},
		owMapsMap: &model.SyncItem{
			Data: make(map[int64]*model.LdInfo),
		},
		lolChampions: &model.SyncInfo{
			Data: make(map[int64]*model.LdInfo),
		},
		dotaHeroes: &model.SyncInfo{
			Data: make(map[int64]*model.LdInfo),
		},
		owHeroes: &model.SyncInfo{
			Data: make(map[int64]*model.LdInfo),
		},
		lolSpells: &model.SyncInfo{
			Data: make(map[int64]*model.LdInfo),
		},
		dotaAbilities: &model.SyncInfo{
			Data: make(map[int64]*model.LdInfo),
		},
		lolPlayers: &model.SyncInfo{
			Data: make(map[int64]*model.LdInfo),
		},
		dotaPlayers: &model.SyncInfo{
			Data: make(map[int64]*model.LdInfo),
		},
		owPlayers: &model.SyncInfo{
			Data: make(map[int64]*model.LdInfo),
		},
		lolTeams: &model.SyncInfo{
			Data: make(map[int64]*model.LdInfo),
		},
		dotaTeams: &model.SyncInfo{
			Data: make(map[int64]*model.LdInfo),
		},
		owTeams: &model.SyncInfo{
			Data: make(map[int64]*model.LdInfo),
		},
		//lolBigPlayers: &model.SyncLolPlayers{
		//	Data: make(map[int64][]*model.LolPlayer),
		//},
		//lolBigTeams: &model.SyncLolTeams{
		//	Data: make(map[int64][]*model.LolTeam),
		//},
		dotaBigPlayers: &model.SyncDotaPlayers{
			Data: make(map[int64][]*model.DotaPlayer),
		},
		dotaBigTeams: &model.SyncDotaTeams{
			Data: make(map[int64][]*model.DotaTeam),
		},
		ldSeasonGame: &model.SyncSeasonGame{
			Data: make(map[int64]int64),
		},
		mapGameDb:         make(map[int64]int64),
		cron:              cron.New(),
		regBv:             regexp.MustCompile(regBVID),
		regAv:             regexp.MustCompile(regAVID),
		liveBattleListMap: new(atomic.Value),
		liveBattleInfoMap: new(atomic.Value),
		teamsInSeasonMap:  make(map[int64][]*model.TeamInSeason, 0),
		matchSeasonMap:    make(map[int64][]*model.MatchSeason, 0),
		seasonMap:         make(map[int64]*model.MatchSeason, 0),
		reserveMap:        make(map[int64]int64, 0),
	}
	var err error
	mList := make(map[string]*model.BattleList, 0)
	s.liveBattleListMap.Store(mList)
	mInfo := make(map[string]*model.BattleInfo, 0)
	s.liveBattleInfoMap.Store(mInfo)
	if s.arcClient, err = arcclient.NewClient(c.ArcClient); err != nil {
		panic(err)
	}
	s.favClient, err = favclient.New(c.FavClient)
	if err != nil {
		panic(err)
	}
	s.actClient, err = actclient.NewClient(c.ActClient)
	if err != nil {
		panic(err)
	}
	s.accClient, err = accClient.NewClient(c.AccClient)
	if err != nil {
		panic(err)
	}
	s.coinclient, err = coinclient.NewClient(c.CoinClient)
	if err != nil {
		panic(err)
	}
	s.liveClient, err = livexroom.NewClient(c.LiveClient)
	if err != nil {
		panic(err)
	}
	s.esportsServiceClient, err = esportsServiceClient.NewClient(c.EsportsServiceClient)
	if err != nil {
		panic(err)
	}
	s.initGameDb()
	genAutoSubscribeMap(context.Background())
	//watchAvCIDMapBySeasonWatch(context.Background())
	//watchS10ContestListBySeasonWatch(context.Background())
	//watchS10ContestSeriesBySeasonWatch(context.Background())
	//watchS10ScoreAnalysisBySeasonWatch(context.Background())
	//watchS10PosterListBySeasonWatch(context.Background())
	//watchSeasonBiz()

	// contest component.
	watchComponentAllTeamsWatch(context.Background())
	watchComponentAllSeasonsWatch(context.Background())
	watchGoingSeasonsByCacheWatch(context.Background())
	watchComponentContestListByGoingSeason(context.Background())
	watchGoingBattleSeasonsByCacheWatch(context.Background())
	watchComponentContestBattleByGoingSeason(context.Background())
	watchSeasonContestBiz(context.Background())
	s.rebuildReserveMap(context.Background())
	s.watchGoingBattleSeasonsContestsTeams(context.Background())
	s.watchLolDataByGoingSeason(context.Background())

	//if deployID := os.Getenv("DEPLOYMENT_ID"); deployID != "" {
	//	if d, err := s.dao.IncrActivityPodIndex(deployID); err == nil {
	//		if d == 1 {
	//			s.startAsyncJobs(srvCtx)
	//		}
	//	} else {
	//		s.startAsyncJobs(srvCtx)
	//	}
	//} else {
	//	s.startAsyncJobs(srvCtx)
	//}

	go s.createCron()
	go s.S10RankingDataWatch()
	go s.loadScoreLive()
	go asyncAutoSubData(context.Background())
	go s.ASyncUpdateMemoryCache(c.HotContestIDList)
	go ASyncUpdateMaxContestID(context.Background())
	go s.AsyncUpdateOngoingSeasonTeamInMemoryCache()
	go StoreHotData2Memory(context.Background())
	go s.AsyncUpdateMatchSeasonInMemoryCache(context.Background())
	go s.AsyncUpdateGoingSeasonsInfoMemoryCache(context.Background())
	goroutineRegister(s.watchGoingBattleSeasonsContestsTeamsTimeTicker)
	goroutineRegister(s.watchGoingVideoListComponent)
	goroutineRegister(s.watchGoingLolDataHero2TimeTicker)

	return s
}

func (s *Service) startAsyncJobs(ctx context.Context) {
	go s.ResetBfsBackup(ctx)
}

func (s *Service) initGameDb() {
	s.mapGameDb = make(map[int64]int64, len(s.c.GameTypes))
	for _, tp := range s.c.GameTypes {
		s.mapGameDb[tp.ID] = tp.DbGameID
	}
}

// Ping ping service.
func (s *Service) Ping(c context.Context) (err error) {
	if err = s.dao.Ping(c); err != nil {
		log.Error("s.dao.Ping error(%v)", err)
	}
	return
}

func (s *Service) createCron() {
	go s.lolPlayersCron()
	go s.dotaPlayersCron()
	go s.owPlayersCron()
	go s.infoCron()
	go s.bigDataCron()
	go s.buildKnockTree()
	go s.s9ContestCron()
	s.cron.AddFunc(s.c.Leidata.LolPlayersCron, s.lolPlayersCron)
	s.cron.AddFunc(s.c.Leidata.DotaPlayersCron, s.dotaPlayersCron)
	s.cron.AddFunc(s.c.Leidata.OwPlayersCron, s.owPlayersCron)
	s.cron.AddFunc(s.c.Leidata.InfoCron, s.infoCron)
	s.cron.AddFunc(s.c.Leidata.BigDataCron, s.bigDataCron)
	s.cron.AddFunc(s.c.Interval.KnockTreeCron, s.buildKnockTree)
	s.cron.AddFunc(s.c.Interval.S9ContestCron, s.s9ContestCron)
	s.cron.Start()
}

func (s *Service) lolPlayersCron() {
	go s.loadLdInfo(_lolPlayers)
	log.Info("createCron lolPlayersCron start")
}

func (s *Service) dotaPlayersCron() {
	go s.loadLdInfo(_dotaPlayers)
	log.Info("createCron dotaPlayersCron start")
}

func (s *Service) owPlayersCron() {
	go s.loadLdInfo(_owPlayers)
	log.Info("createCron owPlayersCron start")
}

func (s *Service) infoCron() {
	go s.loadLdInfo(_lolItems)
	go s.loadLdInfo(_dotaItems)
	go s.loadLdInfo(_owMaps)
	go s.loadLdInfo(_lolSpells)
	go s.loadLdInfo(_dotaAbilities)
	go s.loadLdInfo(_lolChampions)
	go s.loadLdInfo(_dotaHeroes)
	go s.loadLdInfo(_owHeroes)
	go s.loadLdInfo(_lolTeams)
	go s.loadLdInfo(_dotaTeams)
	go s.loadLdInfo(_owTeams)
	log.Info("createCron infoCron start")
}

func (s *Service) bigDataCron() {
	var (
		err                     error
		c                       = context.Background()
		lolSeasons, dotaSeasons []*model.Season
	)
	lolGameID := s.mapGameDb[_lolType]
	if lolGameID > 0 {
		if lolSeasons, err = s.dao.GameSeason(context.Background(), lolGameID); err != nil {
			log.Error("bigDataCron s.ldSeasons LOL error(%+v)", err)
		} else {
			for _, season := range lolSeasons {
				tmp := season
				if tmp.LeidaSID == 0 {
					continue
				}
				//go s.writeLolPlayers(c, tmp.LeidaSID)
				//go s.writeLolTeams(c, tmp.LeidaSID)
				s.ldSeasonGame.Lock()
				s.ldSeasonGame.Data[tmp.LeidaSID] = _lolType
				s.ldSeasonGame.Unlock()
			}
		}
		log.Info("bigDataCron lol start")
	}
	dotaGameID := s.mapGameDb[_dotaType]
	if dotaGameID > 0 {
		if dotaSeasons, err = s.dao.GameSeason(context.Background(), dotaGameID); err != nil {
			log.Error("bigDataCron s.ldSeasons DOTA error(%+v)", err)
		} else {
			for _, season := range dotaSeasons {
				tmp := season
				if tmp.LeidaSID == 0 {
					continue
				}
				go s.writeDotaPlayers(c, tmp.LeidaSID)
				go s.writeDotaTeams(c, tmp.LeidaSID)
				s.ldSeasonGame.Lock()
				s.ldSeasonGame.Data[tmp.LeidaSID] = _dotaType
				s.ldSeasonGame.Unlock()
			}
		}
		log.Info("bigDataCron dota start")
	}
}

//func (s *Service) writeLolPlayers(c context.Context, sid int64) {
//	var (
//		err        error
//		lolPlayers []*model.LolPlayer
//	)
//	if lolPlayers, err = s.dao.LolPlayers(c, sid); err != nil {
//		log.Error("writeLolPlayers s.dao.LolPlayers leidaSID(%d) error(%+v)", sid, err)
//		return
//	}
//	s.lolBigPlayers.Lock()
//	s.lolBigPlayers.Data[sid] = lolPlayers
//	s.lolBigPlayers.Unlock()
//}

//func (s *Service) writeLolTeams(c context.Context, sid int64) {
//	var (
//		err      error
//		lolTeams []*model.LolTeam
//	)
//	if lolTeams, err = s.dao.LolTeams(c, sid); err != nil {
//		log.Error("writeLolTeams s.dao.LolTeams leidaSID(%d) error(%+v)", sid, err)
//		return
//	}
//	s.lolBigTeams.Lock()
//	s.lolBigTeams.Data[sid] = lolTeams
//	s.lolBigTeams.Unlock()
//}

func (s *Service) writeDotaPlayers(c context.Context, sid int64) {
	var (
		err         error
		dotaPlayers []*model.DotaPlayer
	)
	if dotaPlayers, err = s.dao.DotaPlayers(c, sid); err != nil {
		log.Error("writeDotaPlayers s.dao.DotaPlayers leidaSID(%d) error(%+v)", sid, err)
		return
	}
	s.dotaBigPlayers.Lock()
	s.dotaBigPlayers.Data[sid] = dotaPlayers
	s.dotaBigPlayers.Unlock()
}

func (s *Service) writeDotaTeams(c context.Context, sid int64) {
	var (
		err       error
		dotaTeams []*model.DotaTeam
	)
	if dotaTeams, err = s.dao.DotaTeams(c, sid); err != nil {
		log.Error("writeDotaTeams s.dao.DotaTeams leidaSID(%d) error(%+v)", sid, err)
		return
	}
	s.dotaBigTeams.Lock()
	s.dotaBigTeams.Data[sid] = dotaTeams
	s.dotaBigTeams.Unlock()
}

func (s *Service) loadLdInfo(tp string) {
	var (
		err error
		rs  []*model.LdInfo
		c   = context.Background()
	)
	switch tp {
	case _lolItems:
		if rs, err = s.dao.LolItems(c); err != nil {
			log.Error("s.dao.LolItem error(%+v)", err)
			return
		}
		for _, item := range rs {
			s.lolItemsMap.Lock()
			s.lolItemsMap.Data[item.ID] = item
			s.lolItemsMap.Unlock()
		}
	case _dotaItems:
		if rs, err = s.dao.DotaItems(c); err != nil {
			log.Error("s.dao.DotaItems error(%+v)", err)
			return
		}
		for _, item := range rs {
			s.dotaItemsMap.Lock()
			s.dotaItemsMap.Data[item.ID] = item
			s.dotaItemsMap.Unlock()
		}
	case _owMaps:
		if rs, err = s.dao.OwMaps(c); err != nil {
			log.Error("s.dao.OwMaps error(%+v)", err)
			return
		}
		for _, info := range rs {
			s.owMapsMap.Lock()
			s.owMapsMap.Data[info.ID] = info
			s.owMapsMap.Unlock()
		}
	case _lolSpells:
		if rs, err = s.dao.LolSpells(c); err != nil {
			log.Error("s.dao.LolSpells error(%+v)", err)
			return
		}
		for _, info := range rs {
			s.lolSpells.Lock()
			s.lolSpells.Data[info.ID] = info
			s.lolSpells.Unlock()
		}
	case _dotaAbilities:
		if rs, err = s.dao.DotaAbility(c); err != nil {
			log.Error("s.dao.DotaAbility error(%+v)", err)
			return
		}
		for _, info := range rs {
			s.dotaAbilities.Lock()
			s.dotaAbilities.Data[info.ID] = info
			s.dotaAbilities.Unlock()
		}
	case _lolPlayers:
		if rs, err = s.dao.LolMatchPlayer(c); err != nil {
			log.Error("s.dao.LolMatchPlayer error(%+v)", err)
			return
		}
		for _, info := range rs {
			s.lolPlayers.Lock()
			s.lolPlayers.Data[info.ID] = info
			s.lolPlayers.Unlock()
		}
	case _dotaPlayers:
		if rs, err = s.dao.DotaMatchPlayer(c); err != nil {
			log.Error("s.dao.DotaMatchPlayer error(%+v)", err)
			return
		}
		for _, info := range rs {
			s.dotaPlayers.Lock()
			s.dotaPlayers.Data[info.ID] = info
			s.dotaPlayers.Unlock()
		}
	case _owPlayers:
		if rs, err = s.dao.OwMatchPlayer(c); err != nil {
			log.Error("s.dao.OwMatchPlayer error(%+v)", err)
			return
		}
		for _, info := range rs {
			s.owPlayers.Lock()
			s.owPlayers.Data[info.ID] = info
			s.owPlayers.Unlock()
		}
	case _lolChampions:
		if rs, err = s.dao.LolCham(c); err != nil {
			log.Error("s.dao.LolCham error(%+v)", err)
			return
		}
		for _, champion := range rs {
			s.lolChampions.Lock()
			s.lolChampions.Data[champion.ID] = champion
			s.lolChampions.Unlock()
		}
	case _dotaHeroes:
		if rs, err = s.dao.DotaHero(c); err != nil {
			log.Error("s.dao.DotaHero error(%+v)", err)
			return
		}
		for _, hero := range rs {
			s.dotaHeroes.Lock()
			s.dotaHeroes.Data[hero.ID] = hero
			s.dotaHeroes.Unlock()
		}
	case _owHeroes:
		if rs, err = s.dao.OwHero(c); err != nil {
			log.Error("s.dao.OwMap error(%+v)", err)
			return
		}
		for _, hero := range rs {
			s.owHeroes.Lock()
			s.owHeroes.Data[hero.ID] = hero
			s.owHeroes.Unlock()
		}

	case _lolTeams:
		if rs, err = s.dao.LolMatchTeam(c); err != nil {
			log.Error("s.dao.LolMatchPlayer error(%+v)", err)
			return
		}
		for _, info := range rs {
			s.lolTeams.Lock()
			s.lolTeams.Data[info.ID] = info
			s.lolTeams.Unlock()
		}
	case _dotaTeams:
		if rs, err = s.dao.DotaMatchTeam(c); err != nil {
			log.Error("s.dao.DotaMatchPlayer error(%+v)", err)
			return
		}
		for _, info := range rs {
			s.dotaTeams.Lock()
			s.dotaTeams.Data[info.ID] = info
			s.dotaTeams.Unlock()
		}
	case _owTeams:
		if rs, err = s.dao.OwMatchTeam(c); err != nil {
			log.Error("s.dao.OwMatchPlayer error(%+v)", err)
			return
		}
		for _, info := range rs {
			s.owTeams.Lock()
			s.owTeams.Data[info.ID] = info
			s.owTeams.Unlock()
		}
	}
}

func (s *Service) s9ContestCron() {
	var (
		dbContests []*model.Contest
		c          = context.Background()
		err        error
	)
	time.Sleep(time.Duration(conf.Conf.Rule.S9Sleep))
	stime := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day()-2, 0, 0, 0, 0, time.Local)
	etime := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day()+2, 23, 59, 59, 0, time.Local)
	if dbContests, err = s.dao.S9Contests(c, s.c.Rule.S9SwitchSID, stime.Unix(), etime.Unix()); err != nil {
		log.Error("s9ContestCron s.dao.S9Contests S9SwitchSID(%d) stime(%d) etime(%d) error(%v)", s.c.Rule.S9SwitchSID, stime.Unix(), etime.Unix(), err)
		return
	}
	s.s9Contests = dbContests
	log.Warn("s9ContestCron sid(%d) contests count(%d)", s.c.Rule.S9SwitchSID, len(s.s9Contests))
}

func (s *Service) loadScoreLive() {
	ctx := context.Background()
	ticker := time.NewTicker(time.Second * 1)
	defer func() {
		ticker.Stop()
	}()
	for {
		select {
		case <-ticker.C:
			matchOne, e := s.dao.MatchOne(ctx)
			if e != nil || len(matchOne) == 0 {
				s.liveMatchID = ""
				mList := make(map[string]*model.BattleList, 0)
				s.liveBattleListMap.Store(mList)
				log.Errorc(ctx, "loadScoreLive s.dao.MatchOne count(%d) error(%+v)", len(matchOne), e)
				continue
			}
			s.liveMatchID = matchOne[0].MatchID // 第一个接口
			tmpBattleList := make(map[string]*model.BattleList, len(matchOne))
			for _, liveMatch := range matchOne {
				bl, e := s.dao.CacheBattleList(ctx, s.liveMatchID)
				if e != nil {
					log.Errorc(ctx, "loadScoreLive s.dao.CacheBattleList matchID(%s) error(%+v)", s.liveMatchID, e)
					continue
				}
				tmpBattleList[liveMatch.MatchID] = bl
			}
			s.StoreLiveBattleListMap(tmpBattleList) // 第二个接口
			tmpBattleInfo := make(map[string]*model.BattleInfo, len(tmpBattleList))
			for _, battleInfos := range tmpBattleList {
				if battleInfos == nil {
					continue
				}
				for _, btInfo := range battleInfos.List {
					bi, e := s.dao.CacheBattleInfo(ctx, btInfo.BattleString)
					if e != nil {
						log.Errorc(ctx, "ScoreBattleInfo s.dao.CacheBattleInfo battleString(%s) error(%+v)", btInfo.BattleString, e)
						continue
					}
					tmpBattleInfo[btInfo.BattleString] = bi
				}
			}
			s.StoreLiveBattleInfoMap(tmpBattleInfo) // 第三个接口
			log.Infoc(ctx, "loadScoreLive success")
		}
	}
}
