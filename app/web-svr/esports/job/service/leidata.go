package service

import (
	"context"
	"encoding/json"
	"math"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	espClient "go-gateway/app/web-svr/esports/interface/api/v1"
	"go-gateway/app/web-svr/esports/job/component"
	"go-gateway/app/web-svr/esports/job/dao"
	mdlesp "go-gateway/app/web-svr/esports/job/model"
	"go-gateway/app/web-svr/esports/job/tool"

	"golang.org/x/net/websocket"
)

const (
	_perPage                = 100
	_firstPage              = "1"
	_decimal                = 2
	_heroesCount            = 5
	_lolType                = 1
	_dotaType               = 2
	_owType                 = 3
	_firstOwGame            = 0
	_lolGame                = "lol/games"
	_dotaGame               = "dota2/games"
	_owGame                 = "overwatch/games"
	_lolItems               = "lol/items"
	_dotaItems              = "dota2/items"
	_owMaps                 = "overwatch/maps"
	_lolChampions           = "lol/champions"
	_lolVerChampions        = "lol/versions_champions"
	_dotaHeroes             = "dota2/heroes"
	_owHeroes               = "overwatch/heroes"
	_lolSpells              = "lol/spells"
	_dotaAbilities          = "dota2/abilities"
	_lolPlayers             = "lol/players"
	_dotaPlayers            = "dota2/players"
	_owPlayers              = "overwatch/players"
	_lolTeams               = "lol/teams"
	_dotaTeams              = "dota2/teams"
	_owTeams                = "overwatch/teams"
	_lolSeriesPlayers       = "lol/series_players"
	_lolSeriesTeams         = "lol/series_teams"
	_dotaSeriesPlayers      = "dota2/series_players"
	_dotaTournamentsPlayers = "dota2/tournaments_players"
	_dotaSeriesTeams        = "dota2/series_teams"
	_dotaTournamentsTeams   = "dota2/tournaments_teams"
	_lolStats               = "stats/lol"
	_dotaStats              = "stats/dota2"
	_scoreLolGame           = "lol/games.php"
	_scoreLolSeriesPlayers  = "lol/series_players.php"
	_scoreLolPlayerStats    = "lol/players_info.php"
	_scoreLolSeriesTeams    = "lol/series_teams.php"
	_scoreLolTeamStats      = "lol/teams_info.php"
	_scorelolChampions      = "lol/champions.php"
	_scoreLiveMatchList     = "b/livedata_match_list.php"
	_scoreLiveBattleList    = "b/livedata_battle_list.php"
	_scoreLiveBattleInfo    = "b/livedata_battle_info.php"
	_scoreOfflineHero       = "b/hero_list.php"
	_scoreOfflineSkill      = "b/skill_list.php"
	_scoreOfflineDevice     = "b/device_list.php"
	_scoreOfflineTeam       = "b/tournament_team_list.php"
	_scoreLolTeamInfo       = "b/team_info.php"
	_scoreLolDataPlayer     = "b/data_player_list.php"
	_scoreLolDataHero2      = "b/data_hero2_list.php"
)

var (
	_lolRole = map[string]string{"top": "1", "mid": "2", "adc": "3", "jun": "4", "sup": "5"}
)

// BigInit big data init.
func (s *Service) BigInit(tp, sid int64) (err error) {
	if tp == 1 {
		s.fixBigData(s.c.Leidata.After.LolGameID, sid)
	} else if tp == 2 {
		s.fixBigData(s.c.Leidata.After.DotaGameID, sid)
	}
	return
}

// InfoInit info data init.
func (s *Service) InfoInit(tp string, MatchID int64) (err error) {
	if tp == "1" {
		s.infoDataCron()
	} else if tp == "game" {
		if MatchID == 0 {
			s.matchGameCron()
			return
		}
		s.FixMatchGame(MatchID)
	} else {
		if strings.Index(tp, "php") > 0 {
			s.writeScoreInfo(&mdlesp.ParamScore{Pn: _firstPage, Ps: _perPage}, _scorelolChampions)
			return
		}
		s.loadLdPages(tp, 0, false)
	}
	return
}

func (s *Service) bigDataCron() {
	go s.gameData(s.c.Leidata.After.LolGameID)
	go s.gameData(s.c.Leidata.After.DotaGameID)
}

func (s *Service) infoDataCron() {
	s.writeScoreInfo(&mdlesp.ParamScore{Pn: _firstPage, Ps: _perPage}, _scorelolChampions)
	//s.loadLdPages(_lolPlayers, 0, false)
	//s.loadLdPages(_dotaPlayers, 0, false)
	//s.loadLdPages(_owPlayers, 0, false)
	//s.loadLdPages(_lolChampions, 0, false)
	//s.loadLdPages(_lolVerChampions, 0, false)
	//s.loadLdPages(_dotaHeroes, 0, false)
	//s.loadLdPages(_owHeroes, 0, false)
	//s.loadLdPages(_lolItems, 0, false)
	//s.loadLdPages(_dotaItems, 0, false)
	//s.loadLdPages(_owMaps, 0, false)
	//s.loadLdPages(_lolSpells, 0, false)
	//s.loadLdPages(_dotaAbilities, 0, false)
	//s.loadLdPages(_lolTeams, 0, false)
	//s.loadLdPages(_dotaTeams, 0, false)
	//s.loadLdPages(_owTeams, 0, false)
}

func (s *Service) refreshLolData(ctx context.Context, leidaSid int64) {
	s.upScoreLolDataPlayer(ctx, leidaSid)
	s.upScoreLolDataHero2(ctx, leidaSid)
	s.upScoreLolTeamInfo(ctx, leidaSid)
	arg := &espClient.RefreshLolDataRequest{LeidaSid: leidaSid}
	if _, err := component.EspClient.RefreshLolData(ctx, arg); err != nil {
		log.Errorc(ctx, "refreshLolData component.EspClient.RefreshLolData() leidaSid(%d) error(%+v)", leidaSid, err)
		return
	}
	log.Errorc(ctx, "refreshLolData component.EspClient.RefreshLolData() leidaSid(%d) success", leidaSid)
}

func (s *Service) gameData(gameID int) (err error) {
	var (
		seasons []*mdlesp.Season
		ctx     = context.Background()
	)
	if seasons, err = s.dao.SeriesSeason(ctx, gameID); err != nil {
		log.Error("gameData s.dao.SeriesSeason gameID(%d) error(%+v)", gameID, err)
		return
	}
	current := time.Now()
	before, _ := time.ParseDuration("-24h")
	dayTime := current.Add(before)
	for _, season := range seasons {
		tmpS := season
		if dayTime.Unix() > tmpS.Etime {
			log.Info("gameData series id etime over gameID(%d) sid(%d) serieID(%d)", gameID, tmpS.ID, tmpS.LeidaSid)
			continue
		}
		switch gameID {
		case s.c.Leidata.After.LolGameID:
			s.refreshLolData(ctx, tmpS.LeidaSid)
			log.Info("createCron Lol start")
		case s.c.Leidata.After.DotaGameID:
			if tmpS.SerieType == 0 {
				s.loadBig(_dotaSeriesPlayers, tmpS.LeidaSid)
				s.loadBig(_dotaSeriesTeams, tmpS.LeidaSid)
			} else if tmpS.SerieType == 1 {
				s.loadBig(_dotaTournamentsPlayers, tmpS.LeidaSid)
				s.loadBig(_dotaTournamentsTeams, tmpS.LeidaSid)
			}
			log.Info("createCron Dota start")
		}
	}
	return
}

func (s *Service) fixBigData(gameID int, sid int64) (err error) {
	var (
		seasons []*mdlesp.Season
		ctx     = context.Background()
	)
	if seasons, err = s.dao.SeriesSeason(ctx, gameID); err != nil {
		log.Error("gameData s.dao.SeriesSeason gameID(%d) error(%+v)", gameID, err)
		return
	}
	for _, season := range seasons {
		tmpS := season
		if tmpS.LeidaSid != sid {
			continue
		}
		switch gameID {
		case s.c.Leidata.After.LolGameID:
			s.refreshLolData(ctx, tmpS.LeidaSid)
			log.Info("createCron Lol start")
		case s.c.Leidata.After.DotaGameID:
			if tmpS.SerieType == 0 {
				s.loadBig(_dotaSeriesPlayers, tmpS.LeidaSid)
				s.loadBig(_dotaSeriesTeams, tmpS.LeidaSid)
			} else if tmpS.SerieType == 1 {
				s.loadBig(_dotaTournamentsPlayers, tmpS.LeidaSid)
				s.loadBig(_dotaTournamentsTeams, tmpS.LeidaSid)
			}
			log.Info("createCron Dota start")
		}
	}
	return
}

func (s *Service) loadBig(tp string, serieID int64) {
	var (
		oids []int64
		c    = context.Background()
	)
	switch tp {
	case _scoreLolSeriesPlayers:
		oids = s.loadScorePages(&mdlesp.ParamScore{SerieID: serieID, Pn: _firstPage, Ps: _perPage}, tp)
	default:
		oids = s.loadLdPages(tp, serieID, true)
	}
	if len(oids) == 0 {
		return
	}
	switch tp {
	case _lolSeriesPlayers, _scoreLolSeriesPlayers:
		//s.bigLolPlayer(c, tp, serieID, oids)
		log.Info("leida big data lol players success serieID(%d)", serieID)
	case _dotaSeriesPlayers, _dotaTournamentsPlayers:
		s.bigDotaPlayer(c, tp, serieID, oids)
		log.Info("leida big data dota players success id(%d) type(%s)", serieID, tp)
	case _dotaSeriesTeams, _dotaTournamentsTeams:
		s.bigDotaTeam(c, tp, serieID, oids)
		log.Info("leida big data dota teams success id(%d) type(%s)", serieID, tp)
	}
}

func (s *Service) upScoreLolDataPlayer(ctx context.Context, serieID int64) {
	lolDataPlayer, err := s.getLolDataPlayer(ctx, serieID)
	if err != nil {
		return
	}
	s.bigLolDataPlayer(ctx, serieID, lolDataPlayer)
}

func (s *Service) upScoreLolDataHero2(ctx context.Context, serieID int64) {
	lolDataHero2, err := s.getLolDataHero2(ctx, serieID)
	if err != nil {
		return
	}
	s.bigLolDataHero2(ctx, serieID, lolDataHero2)
}

func (s *Service) upScoreLolTeamInfo(c context.Context, serieID int64) {
	oids, err := s.tournamentTeamIDs(c, serieID)
	if err != nil {
		return
	}
	s.bigLolTeamV2(c, serieID, oids)
	log.Info("score big data lol teams success serieID(%d)", serieID)
}

func (s *Service) bigLolDataPlayer(c context.Context, serieID int64, lolDataPlayer *mdlesp.LolDataPlayer) {
	var (
		lolPlayer                    *mdlesp.LolPlayer
		err                          error
		lolPlayerStats               mdlesp.LolPlayerStats
		lolPlayers                   []*mdlesp.LolPlayer
		body                         []byte
		win                          float64
		exists                       bool
		lolChampions, lolLdChampions string
		scoreStats                   struct {
			Data mdlesp.LolPlayerStats
		}
	)
	// 加载本地图片列表
	s10RankingLiveOffLineImage = s.LoadLiveOffLineImageMap()
	for _, LolDataPlayerData := range lolDataPlayer.Data.List {
		playerID := tool.StrToInt64Normal(LolDataPlayerData.PlayerID)
		if lolPlayer, err = s.dao.LolPlayerSerie(c, serieID, playerID); err != nil {
			log.Errorc(ctx, "bigLolDataPlayer bigLol s.dao.LolPlayerSerie serieID(%d) playerID(%d) error(%+v)", serieID, playerID, err)
			time.Sleep(time.Millisecond * 10)
			continue
		}
		exists = lolPlayer != nil && lolPlayer.ID > 0
		lolPlayerStats = mdlesp.LolPlayerStats{}
		if body, err = s.score(&mdlesp.ParamScore{SerieID: serieID, OriginID: playerID}, _scoreLolPlayerStats); err != nil || len(body) == 0 {
			log.Errorc(ctx, "bigLolDataPlayer bigLol s.score playerID(%d) body len(%d) error(%v)", playerID, len(body), err)
			time.Sleep(time.Millisecond * 10)
			continue
		}
		if err = json.Unmarshal(body, &scoreStats); err != nil {
			log.Errorc(ctx, "bigLolDataPlayer bigLol json.Unmarshal playerID(%d) error(%v)", playerID, err)
			time.Sleep(time.Millisecond * 10)
			continue
		}
		lolPlayerStats = scoreStats.Data
		deaths := decimal(lolPlayerStats.Stats.Averages.Deaths, _decimal)
		kills := decimal(lolPlayerStats.Stats.Averages.Kills, _decimal)
		assists := decimal(lolPlayerStats.Stats.Averages.Assists, _decimal)
		if lolPlayerStats.Stats.Totals.GamesPlayed == 0 {
			win = 0
		} else {
			tmpWin := float64(lolPlayerStats.Stats.Totals.GamesWon) / float64(lolPlayerStats.Stats.Totals.GamesPlayed)
			win = decimal(tmpWin, _decimal)
		}
		// 替换图片
		if err = s.s10RankingDataReplaceImg(ctx, &LolDataPlayerData); err != nil {
			log.Errorc(c, "bigLolDataPlayer s.s10RankingDataReplaceImg  playerID(%d) error(%+v)", playerID, err)
			continue
		}
		lolChampions, lolLdChampions = s.lolChampions(lolPlayerStats.FavoriteChampions)
		roleID := _lolRole[lolPlayerStats.Role]
		lolPlayer = &mdlesp.LolPlayer{PlayerID: playerID, TeamID: tool.StrToInt64Normal(LolDataPlayerData.TeamID), TeamAcronym: LolDataPlayerData.TeamName,
			TeamImage: LolDataPlayerData.TeamImage, LeidaSID: serieID, Name: LolDataPlayerData.PlayerName, ImageURL: LolDataPlayerData.PlayerImage, ChampionsImage: lolChampions,
			Role: roleID, KDA: tool.StrToFloatNormal(LolDataPlayerData.KDA), Kills: kills, Deaths: deaths, Assists: assists, MinionsKilled: decimal(lolPlayerStats.Stats.Averages.MinionsKilled, _decimal),
			WardsPlaced: decimal(lolPlayerStats.Stats.Averages.WardsPlaced, _decimal), GamesCount: lolPlayerStats.Stats.GamesCount,
			LeidaTeamImage: "", LeidaImage: "", LeidaChampionsImage: lolLdChampions, Win: win, MVP: tool.StrToInt64Normal(LolDataPlayerData.MVP),
			PositionID: tool.StrToInt64Normal(LolDataPlayerData.PositionID), Position: LolDataPlayerData.Position,
		}
		if exists {
			if err = s.dao.UpLolPlayer(c, lolPlayer); err != nil {
				log.Errorc(ctx, "bigLolDataPlayer bigLol s.dao.UpLolPlayer playerID(%d) serieID(%d) error(%+v)", lolPlayer.ID, lolPlayer.LeidaSID, err)
				time.Sleep(time.Millisecond * 10)
				continue
			}
		} else {
			lolPlayers = append(lolPlayers, lolPlayer)
		}
		time.Sleep(time.Millisecond * 10)
	}
	if len(lolPlayers) > 0 {
		if err = s.dao.AddLolPlayer(c, lolPlayers); err != nil {
			log.Error("bigLol s.dao.AddLolPlayer error(%+v)", err)
			return
		}
	}
}

func (s *Service) bigLolDataHero2(ctx context.Context, serieID int64, lolDataHero2 *mdlesp.LolDataHero2) {
	// 加载本地图片列表
	s10RankingLiveOffLineImage = s.LoadLiveOffLineImageMap()
	for _, LolDataHero2Data := range lolDataHero2.Data.List {
		// 替换图片
		if err := s.s10RankingDataReplaceImg(ctx, &LolDataHero2Data); err != nil {
			log.Errorc(ctx, "bigLolDataHero2 s.s10RankingDataReplaceImg() serieID(%d) error(%+v)", serieID, err)
			continue
		}
		if err := dao.InsertUpdatePlayerDataHero2(ctx, serieID, LolDataHero2Data); err != nil {
			log.Errorc(ctx, "bigLolDataHero2  dao.InsertUpdatePlayerDataHero2() serieID(%d) error(%+v)", serieID, err)
		}
	}
}

func (s *Service) getScoreBigLolTeam(ctx context.Context, tournamentID int64) (res map[int64]*mdlesp.ScoreOriginTeamAnalysis, err error) {
	values := genUrlValuesByTournamentID(tournamentID)
	resBig := struct {
		Data struct {
			List []*mdlesp.ScoreOriginTeamAnalysis `json:"list"`
		} `json:"data"`
	}{}
	if err = s.getScoreData(ctx, scorePathOfBigData4Team, values, &resBig); err != nil {
		log.Errorc(ctx, "getScoreBigLolTeam s.getScoreData() tournamentID(%d) error(%+v)", tournamentID, err)
		return
	}
	res = make(map[int64]*mdlesp.ScoreOriginTeamAnalysis, len(resBig.Data.List))
	for _, v := range resBig.Data.List {
		if v.TeamID == "" {
			continue
		}
		intTeamID, err := strconv.ParseInt(v.TeamID, 10, 64)
		if err != nil {
			log.Errorc(ctx, "getScoreBigLolTeam s.getScoreData() tournamentID(%d) teamID(%s) error(%+v)", tournamentID, v.TeamID, err)
			return nil, err
		}
		res[intTeamID] = v
	}
	return
}

func (s *Service) bigLolTeamV2(c context.Context, serieID int64, oids []int64) {
	var (
		lolTeam      *mdlesp.LolTeam
		lolTeams     []*mdlesp.LolTeam
		err          error
		exists       bool
		lolTeamStats *mdlesp.ScoreTeamInfo
		scoreImg     string
	)
	// 加载本地图片列表
	s10RankingLiveOffLineImage = s.LoadLiveOffLineImageMap()
	scoreBigTeamMap, err := s.getScoreBigLolTeam(c, serieID)
	if err != nil {
		log.Errorc(ctx, "bigLolTeamV2 s.getScoreBigLolTeam() serieID(%d) error(%+v)", serieID, err)
		return
	}
	for _, teamID := range oids {
		if lolTeam, err = s.dao.LolTeamSerie(c, serieID, teamID); err != nil {
			log.Error("bigLolTeamV2 lolSeriesTeams bilolTeam s.dao.LolTeamSerie serieID(%d) teamID(%d) error(%+v)", serieID, teamID, err)
			time.Sleep(time.Millisecond * 10)
			continue
		}
		exists = lolTeam != nil && lolTeam.ID > 0
		if lolTeamStats, err = s.scoreTeamInfo(c, serieID, teamID); err != nil || lolTeamStats == nil {
			log.Errorc(c, "bigLolTeamV2  s.scoreTeamInfo serieID(%d) teamID(%d)  error(%+v)", serieID, teamID, err)
			continue
		}
		scoreImg = lolTeamStats.Data.Team.Image
		// 替换图片
		if err = s.s10RankingDataReplaceImg(ctx, &lolTeamStats); err != nil {
			log.Errorc(c, "bigLolTeamV2: s.s10RankingDataReplaceImg  teamID(%d) error(%+v)", teamID, err)
			continue
		}
		bigLolTeamStats := rebuildBigLolTeam(teamID, serieID, lolTeamStats, scoreBigTeamMap, scoreImg)
		if bigLolTeamStats == nil {
			log.Errorc(c, "bigLolTeamV2  rebuildBigLolTeam() team big api not found teamID(%d) serieID(%d)", lolTeam.ID, lolTeam.LeidaSID)
			continue
		}
		if exists {
			if err = s.dao.UpLolTeam(c, bigLolTeamStats); err != nil {
				log.Errorc(c, "bigLolTeamV2 s.dao.UpLolTeam teamID(%d) serieID(%d) error(%+v)", lolTeam.ID, lolTeam.LeidaSID, err)
				time.Sleep(time.Millisecond * 10)
				continue
			}
		} else {
			lolTeams = append(lolTeams, bigLolTeamStats)
		}
		time.Sleep(time.Millisecond * 2)
	}
	if len(lolTeams) > 0 {
		if err = s.dao.AddLolTeam(c, lolTeams); err != nil {
			log.Errorc(c, "bigLolTeamV2 s.dao.AddLolTeam error(%+v)", err)
			return
		}
	}
}

func rebuildBigLolTeam(teamID, serieID int64, lolTeamStats *mdlesp.ScoreTeamInfo, scoreBigTeamMap map[int64]*mdlesp.ScoreOriginTeamAnalysis, scoreImg string) *mdlesp.LolTeam {
	var baronKills float64
	scoreBigTeam, ok := scoreBigTeamMap[teamID]
	if !ok {
		return nil
	}
	gameCount := strToFloat(lolTeamStats.Data.Data.Count)
	if gameCount > 0 {
		baronKills = strToFloat(lolTeamStats.Data.Data.SumDragon) / gameCount
	}
	players := getBigLolTeamPlayers(lolTeamStats)
	lolTeam := &mdlesp.LolTeam{TeamID: teamID, Acronym: lolTeamStats.Data.Team.ShortName,
		LeidaSID: serieID, Name: lolTeamStats.Data.Team.Name, ImageURL: lolTeamStats.Data.Team.Image, Win: strToFloat(scoreBigTeam.WinRate) / 100,
		KDA: strToFloat(scoreBigTeam.KDA), Kills: strToFloat(scoreBigTeam.AvgKills), Deaths: strToFloat(lolTeamStats.Data.Data.Deaths),
		Assists:    strToFloat(scoreBigTeam.AvgAssists),
		TowerKills: strToFloat(scoreBigTeam.AvgTowersDestroyed), FirstTower: strToFloat(lolTeamStats.Data.Data.FirstTowerRate) / 100,
		FirstInhibitor: 0, FirstDragon: 0, FirstBaron: 0, TotalMinionsKilled: 0, WardsPlaced: strToFloat(lolTeamStats.Data.Data.AvgWardsPlaced), GoldEarned: 0,
		FirstBlood: strToFloat(scoreBigTeam.FirstBloodRate) / 100, InhibitorKills: strToFloat(scoreBigTeam.AvgBigDargon),
		BaronKills: decimal(baronKills, _decimal), Players: string(players), GamesCount: int64(gameCount), LeidaImage: scoreImg,
		BaronRate: strToFloat(scoreBigTeam.BigDargonRate), DragonRate: strToFloat(scoreBigTeam.SmallDargonRate),
		Hits: strToFloat(scoreBigTeam.AvgHitsMin), LoseNum: strToFloat(scoreBigTeam.Lose),
		Money: strToFloat(scoreBigTeam.AvgEconomyMin), TotalDamage: strToFloat(scoreBigTeam.AvgDamageMin),
		WinNum: strToFloat(scoreBigTeam.Win), ImageThumb: lolTeamStats.Data.Team.ImageThumb, NewData: 1,
	}
	return lolTeam
}

func getBigLolTeamPlayers(lolTeamStats *mdlesp.ScoreTeamInfo) (players []byte) {
	var err error
	if len(lolTeamStats.Data.Team.Player) > 0 {
		var oldPlayer []*mdlesp.ScorePlayer
		for _, player := range lolTeamStats.Data.Team.Player {
			intPlayerID, _ := strconv.Atoi(player.PlayerID)
			oldPlayer = append(oldPlayer, &mdlesp.ScorePlayer{
				ID:       intPlayerID,
				Name:     player.Nickname,
				ImageURL: player.ImageThumb,
				Role:     player.PositionID,
			})
		}
		if players, err = json.Marshal(oldPlayer); err != nil {
			log.Error("rebuildBigLolTeam:  json.Marshal  teamID(%d) error(%+v)", lolTeamStats.Data.Team.TeamID, err)
			players = []byte("[]")
		}
	} else {
		players = []byte("[]")
	}
	return
}

func (s *Service) bigDotaPlayer(c context.Context, tp string, serieID int64, oids []int64) {
	var (
		dotaPlayer                                    *mdlesp.DotaPlayer
		dotaPlayerStats                               mdlesp.DotaPlayerStats
		dotaPlayers                                   []mdlesp.DotaPlayer
		err                                           error
		body                                          []byte
		kda, win                                      float64
		exists                                        bool
		imageUrl, teamImage, dotaHeroes, dotaLdHeroes string
	)

	for _, playerID := range oids {
		if dotaPlayer, err = s.dao.DotaPlayerSerie(c, serieID, playerID); err != nil {
			log.Error("dotaSeriesPlayers bigDotaPlayer s.dao.DotaPlayerSerie serieID(%d) playerID(%d) error(%+v)", serieID, playerID, err)
			time.Sleep(time.Millisecond * 10)
			continue
		}
		exists = dotaPlayer != nil && dotaPlayer.ID > 0
		dotaPlayerStats = mdlesp.DotaPlayerStats{}
		if body, err = s.httpGet(tp, playerID, serieID); err != nil || len(body) == 0 {
			log.Error("dotaSeriesPlayers bigDotaPlayer s.httpGet playerID(%d) body len(%d) error(%v)", playerID, len(body), err)
			time.Sleep(time.Millisecond * 10)
			continue
		}
		if err = json.Unmarshal(body, &dotaPlayerStats); err != nil {
			log.Error("dotaSeriesPlayers bigDotaPlayer  json.Unmarshal playerID(%d) error(%v)", playerID, err)
			time.Sleep(time.Millisecond * 10)
			continue
		}
		deaths := decimal(dotaPlayerStats.Stats.Averages.Deaths, _decimal)
		kills := decimal(dotaPlayerStats.Stats.Averages.Kills, _decimal)
		assists := decimal(dotaPlayerStats.Stats.Averages.Assists, _decimal)
		if deaths == 0 {
			kda = 0
		} else {
			kda = (kills + assists) / deaths
		}
		if dotaPlayerStats.Stats.Totals.GamesPlayed == 0 {
			win = 0
		} else {
			tmpWin := float64(dotaPlayerStats.Stats.Totals.GamesWon) / float64(dotaPlayerStats.Stats.Totals.GamesPlayed)
			win = decimal(tmpWin, _decimal)
		}
		if exists {
			if dotaPlayer.LeidaTeamImage == dotaPlayerStats.CurrentTeam.ImageURL {
				teamImage = dotaPlayer.TeamImage
			} else {
				teamImage = s.BfsProxy(c, dotaPlayerStats.CurrentTeam.ImageURL)
			}
			if dotaPlayer.LeidaImage == dotaPlayerStats.ImageURL {
				imageUrl = dotaPlayer.ImageURL
			} else {
				imageUrl = s.BfsProxy(c, dotaPlayerStats.ImageURL)
			}

		} else {
			teamImage = s.BfsProxy(c, dotaPlayerStats.CurrentTeam.ImageURL)
			imageUrl = s.BfsProxy(c, dotaPlayerStats.ImageURL)
		}
		dotaHeroes, dotaLdHeroes = s.dotaHeroes(exists, dotaPlayerStats.FavoriteHeroes, dotaPlayer.LeidaHeroesImage, dotaPlayer.HeroesImage)
		dotaPlayer := mdlesp.DotaPlayer{PlayerID: playerID, TeamID: dotaPlayerStats.CurrentTeam.ID, TeamAcronym: dotaPlayerStats.CurrentTeam.Acronym,
			TeamImage: teamImage, LeidaSID: serieID, Name: dotaPlayerStats.Name, ImageURL: imageUrl, HeroesImage: dotaHeroes,
			Role: dotaPlayerStats.Role, KDA: decimal(kda, 1), Kills: kills, Deaths: deaths, Assists: assists,
			WardsPlaced: decimal(dotaPlayerStats.Stats.Averages.WardsPlaced, _decimal), LastHits: decimal(dotaPlayerStats.Stats.Averages.LastHits, _decimal),
			ObserverWardsPlaced: decimal(dotaPlayerStats.Stats.Averages.ObserverWardsPlaced, _decimal), SentryWardsPlaced: decimal(dotaPlayerStats.Stats.Averages.SentryWardsPlaced, _decimal),
			XpPerMinute: decimal(dotaPlayerStats.Stats.Averages.XpPerMinute, _decimal), GoldPerMinute: decimal(dotaPlayerStats.Stats.Averages.GoldPerMinute, _decimal), GamesCount: dotaPlayerStats.Stats.GamesCount,
			LeidaTeamImage: dotaPlayerStats.CurrentTeam.ImageURL, LeidaImage: dotaPlayerStats.ImageURL, LeidaHeroesImage: dotaLdHeroes, Win: win,
		}
		if exists {
			if err = s.dao.UpDotaPlayer(c, dotaPlayer); err != nil {
				log.Error("bigDotaPlayer s.dao.UpLolPlayer playerID(%d) serieID(%d) error(%+v)", dotaPlayer.ID, dotaPlayer.LeidaSID, err)
				time.Sleep(time.Millisecond * 10)
				continue
			}
		} else {
			dotaPlayers = append(dotaPlayers, dotaPlayer)
		}
		time.Sleep(time.Millisecond * 10)
	}
	if len(dotaPlayers) > 0 {
		if err = s.dao.AddDotaPlayer(c, dotaPlayers); err != nil {
			log.Error("bigDotaPlayer s.dao.AddLolPlayer error(%+v)", err)
			return
		}
	}
}

func (s *Service) bigDotaTeam(c context.Context, tp string, serieID int64, oids []int64) {
	var (
		dotaTeam      *mdlesp.DotaTeam
		dotaTeamStats mdlesp.DotaTeamStats
		dotaTeams     []mdlesp.DotaTeam
		err           error
		body, players []byte
		kda           float64
		exists        bool
		imageUrl      string
	)
	for _, teamID := range oids {
		if dotaTeam, err = s.dao.DotaTeamSerie(c, serieID, teamID); err != nil {
			log.Error("dotaSeriesTeams bigDotaTeam s.dao.DotaTeamSerie serieID(%d) teamID(%d) error(%+v)", serieID, teamID, err)
			time.Sleep(time.Millisecond * 10)
			continue
		}
		exists = dotaTeam != nil && dotaTeam.ID > 0

		dotaTeamStats = mdlesp.DotaTeamStats{}
		if body, err = s.httpGet(tp, teamID, serieID); err != nil || len(body) == 0 {
			log.Error("dotaSeriesTeams bigDotaTeam s.httpGet teamID(%d) body len(%d) error(%v)", teamID, len(body), err)
			time.Sleep(time.Millisecond * 10)
			continue
		}
		if err = json.Unmarshal(body, &dotaTeamStats); err != nil {
			log.Error("dotaSeriesTeams bigDotaTeam json.Unmarshal teamID(%d) error(%v)", teamID, err)
			time.Sleep(time.Millisecond * 10)
			continue
		}
		deaths := decimal(dotaTeamStats.Stats.Averages.Deaths, _decimal)
		kills := decimal(dotaTeamStats.Stats.Averages.Kills, _decimal)
		assists := decimal(dotaTeamStats.Stats.Averages.Assists, _decimal)
		if deaths == 0 {
			kda = 0
		} else {
			kda = (kills + assists) / deaths
		}
		if exists {
			if dotaTeam.LeidaImage == dotaTeamStats.ImageURL {
				imageUrl = dotaTeam.ImageURL
			} else {
				imageUrl = s.BfsProxy(c, dotaTeamStats.ImageURL)
			}
		} else {
			imageUrl = s.BfsProxy(c, dotaTeamStats.ImageURL)
		}
		if len(dotaTeamStats.Players) > 0 {
			for _, player := range dotaTeamStats.Players {
				player.ImageURL = s.BfsProxy(c, player.ImageURL)
			}
			if players, err = json.Marshal(dotaTeamStats.Players); err != nil {
				log.Errorc(ctx, "dotaSeriesTeams bigDotaTeam json.Unmarshal teamID(%d) error(%v)", teamID, err)
			}
		} else {
			players = []byte("[]")
		}
		dotaTeam := mdlesp.DotaTeam{TeamID: teamID, Acronym: dotaTeamStats.Acronym,
			LeidaSID: serieID, Name: dotaTeamStats.Name, ImageURL: imageUrl, Win: dotaTeamStats.Stats.Averages.Ratios.Win,
			KDA: decimal(kda, 1), Kills: kills, Deaths: deaths, Assists: assists,
			TowerKills: decimal(dotaTeamStats.Stats.Averages.TowerKills, _decimal), LastHits: decimal(dotaTeamStats.Stats.Averages.LastHits, _decimal), ObserverUsed: decimal(dotaTeamStats.Stats.Averages.ObserverUsed, _decimal),
			SentryUsed: decimal(dotaTeamStats.Stats.Averages.SentryUsed, _decimal), XpPerMinute: decimal(dotaTeamStats.Stats.Averages.XpPerMin, _decimal), FirstBlood: decimal(dotaTeamStats.Stats.Averages.Ratios.FirstBlood, _decimal),
			Heal: decimal(dotaTeamStats.Stats.Averages.Heal, _decimal), GoldSpent: decimal(dotaTeamStats.Stats.Averages.GoldSpent, _decimal), GoldPerMin: decimal(dotaTeamStats.Stats.Averages.GoldPerMin, _decimal),
			Denies: decimal(dotaTeamStats.Stats.Averages.Denies, _decimal), DamageTaken: decimal(dotaTeamStats.Stats.Averages.DamageTaken, _decimal), CampsStacked: decimal(dotaTeamStats.Stats.Averages.CampsStacked, _decimal),
			Players: string(players), GamesCount: dotaTeamStats.Stats.GamesCount, LeidaImage: dotaTeamStats.ImageURL,
		}
		if exists {
			if err = s.dao.UpDotaTeam(c, dotaTeam); err != nil {
				log.Error("bigDotaTeam s.dao.UpDotaTeam teamID(%d) serieID(%d) error(%+v)", dotaTeam.ID, dotaTeam.LeidaSID, err)
				time.Sleep(time.Millisecond * 10)
				continue
			}
		} else {
			dotaTeams = append(dotaTeams, dotaTeam)
		}
		time.Sleep(time.Millisecond * 10)
	}
	if len(dotaTeams) > 0 {
		if err = s.dao.AddDotaTeam(c, dotaTeams); err != nil {
			log.Error("bigDotaTeam s.dao.AddDotaTeam error(%+v)", err)
			return
		}
	}
}

func (s *Service) lolChampions(FavCham []*mdlesp.FavChamps) (rsBfsCham, rsLdCham string) {
	var (
		ldChampions, champions []string
		ldCham, bfsCham        []byte
		err                    error
		c                      = context.Background()
	)
	for _, cham := range FavCham {
		if cham.Champion.ImageURL != "" {
			scoreImg := cham.Champion.ImageURL
			// 替换图片
			if err = s.s10RankingDataReplaceImg(ctx, &cham); err != nil {
				log.Errorc(c, "lolChampions s.s10RankingDataReplaceImg error(%+v)", err)
				continue
			}
			champions = append(champions, cham.Champion.ImageURL)
			ldChampions = append(ldChampions, scoreImg)
		}
		if len(champions) == _heroesCount {
			break
		}
	}
	if len(ldChampions) > 0 {
		if ldCham, err = json.Marshal(ldChampions); err != nil {
			log.Error("lolChampions not exists ldCham  json.Marshal error(%v) ", err)
		}
	} else {
		ldCham = []byte("[]")
	}
	if len(champions) > 0 {
		if bfsCham, err = json.Marshal(champions); err != nil {
			log.Error("lolChampions not exists bfsCham json.Marshal error(%v) ", err)
		}
	} else {
		bfsCham = []byte("[]")
	}
	rsLdCham = string(ldCham)
	rsBfsCham = string(bfsCham)
	return
}

func (s *Service) dotaHeroes(exists bool, FavHeros []*mdlesp.FavHeroes, dbLdHero, dbBfsHero string) (rsBfsHero, rsLdHero string) {
	var (
		ldHeroes, heroes []string
		bfsHero, ldHero  []byte
		err              error
		c                = context.Background()
	)
	if exists {
		for _, h := range FavHeros {
			if h.Hero.ImageURL != "" {
				ldHeroes = append(ldHeroes, h.Hero.ImageURL)
			}
			if len(ldHeroes) == _heroesCount {
				break
			}
		}
		if len(ldHeroes) > 0 {
			if ldHero, err = json.Marshal(ldHeroes); err != nil {
				log.Error("ldHeroes exists ldHero  json.Marshal error(%v) ", err)
				return
			}
		} else {
			ldHero = []byte("[]")
		}
		rsLdHero = string(ldHero)
		if rsLdHero == dbLdHero {
			rsBfsHero = dbBfsHero
		} else {
			for _, ldh := range ldHeroes {
				heroes = append(heroes, s.BfsProxy(c, ldh))
			}
			if len(heroes) > 0 {
				if bfsHero, err = json.Marshal(heroes); err != nil {
					log.Error("ldHeroes exists  bfsHero json.Marshal error(%v) ", err)
				}
			} else {
				bfsHero = []byte("[]")
			}
			rsBfsHero = string(bfsHero)
		}
	} else {
		for _, h := range FavHeros {
			if h.Hero.ImageURL != "" {
				heroes = append(heroes, s.BfsProxy(c, h.Hero.ImageURL))
				ldHeroes = append(ldHeroes, h.Hero.ImageURL)
			}
			if len(heroes) == _heroesCount {
				break
			}
		}
		if len(heroes) > 0 {
			if bfsHero, err = json.Marshal(heroes); err != nil {
				log.Error("ldHeroes not exists  bfsHero json.Marshal error(%v) ", err)
			}
		} else {
			bfsHero = []byte("[]")
		}
		if len(ldHeroes) > 0 {
			if ldHero, err = json.Marshal(ldHeroes); err != nil {
				log.Error("ldHeroes not exists ldHero  json.Marshal error(%v) ", err)
			}
		} else {
			ldHero = []byte("[]")
		}
		rsBfsHero = string(bfsHero)
		rsLdHero = string(ldHero)
	}
	return
}

func (s *Service) httpGet(tp string, oid, serieID int64) (rs []byte, err error) {
	var route string
	params := url.Values{}
	params.Set("key", s.c.Leidata.After.Key)
	switch tp {
	case _lolSeriesPlayers:
		params.Set("player_id", strconv.FormatInt(oid, 10))
		route = _lolStats
	case _lolSeriesTeams:
		params.Set("team_id", strconv.FormatInt(oid, 10))
		route = _lolStats
	case _dotaSeriesPlayers, _dotaTournamentsPlayers:
		params.Set("player_id", strconv.FormatInt(oid, 10))
		route = _dotaStats
	case _dotaSeriesTeams, _dotaTournamentsTeams:
		params.Set("team_id", strconv.FormatInt(oid, 10))
		route = _dotaStats
	}
	if tp == _dotaTournamentsPlayers || tp == _dotaTournamentsTeams {
		params.Set("tournament_id", strconv.FormatInt(serieID, 10))
	} else {
		params.Set("serie_id", strconv.FormatInt(serieID, 10))
	}
	tmpURL := s.c.Leidata.After.URL + "/" + route + "?" + params.Encode()
	for i := 0; i < s.c.Leidata.After.Retry; i++ {
		if rs, err = s.dao.ThirdGet(context.Background(), tmpURL); err != nil {
			time.Sleep(time.Second)
			continue
		}
		break
	}
	if err != nil {
		log.Error("json.Unmarshal url(%s) body(%s) error(%v)", tmpURL, string(rs), err)
	}
	return
}

func (s *Service) pushPoints() {
	var (
		c   = context.Background()
		res = make(map[string]interface{})
		ws  *websocket.Conn
		buf []byte
		err error
	)
	for {
	reconnect:
		url := s.c.Leidata.Socket + "?key=" + s.c.Leidata.Key
		ws, err = websocket.Dial(url, "", s.c.Leidata.Origin)
		if err != nil {
			log.Error("pushPoints websocket.Dial error(%v)", err)
			time.Sleep(time.Second)
			continue
		}
		for {
			if len(s.matchIDs.Data) == 0 {
				log.Warn("pushPoints s.matchIDs.Data empty")
				time.Sleep(time.Second)
				continue
			}
			ws.SetDeadline(time.Now().Add(time.Duration(s.c.Leidata.ConnTime)))
			if err = websocket.Message.Receive(ws, &buf); err != nil {
				log.Error("pushPoints ws.Read error(%v)", err)
				ws.Close()
				goto reconnect
			}
			if err = json.Unmarshal(buf, &res); err != nil {
				log.Error("pushPoints json.Unmarshal error(%v)", err)
				time.Sleep(time.Second)
				continue
			}
			switch res["type"] {
			case "init":
				if res["client_id"] != nil {
					s.clientID = res["client_id"].(string)
					if s.clientID != "" {
						for _, match := range s.matchIDs.Data {
							tmpM := match
							log.Warn("pushPoints init match_id(%d) clientID(%s)", tmpM.MatchID, s.clientID)
							go s.bind(tmpM)
						}
					}
				}
			case "matches_live":
				if msg, matchID, e := s.liveData(res["data"]); e != nil {
					time.Sleep(time.Second)
				} else {
					go s.dao.PushRoom(c, matchID, _matchOpt, msg)
				}
			case "events_live":
				if msg, matchID, e := s.liveData(res["data"]); e != nil {
					time.Sleep(time.Second)
				} else {
					go s.dao.PushRoom(c, matchID, _eventOpt, msg)
				}
			default:
				log.Warn("pushPoints default  received: %s.\n", string(buf))
			}
			log.Info("pushPoints websocket received (%s).\n", string(buf))
			//fmt.Printf("received: (%s).\n", string(buf))
		}
	}
}

func (s *Service) liveData(p interface{}) (msg string, matchID int64, err error) {
	var match mdlesp.DataMsg
	if p == nil {
		err = ecode.NothingFound
		return
	}
	if err = json.Unmarshal([]byte(p.(string)), &match); err != nil {
		log.Error("liveData  json.Unmarshal error(%v)", err)
		return
	}
	msg = p.(string)
	matchID = match.Match.ID
	//fmt.Printf("liveData received(%s)", msg)
	return
}

func (s *Service) loadLdPages(tp string, serieID int64, isOids bool) (res []int64) {
	var (
		err    error
		params url.Values
		count  int
	)
	params = url.Values{}
	params.Set("page", _firstPage)
	params.Set("per_page", strconv.Itoa(_perPage))
	if serieID > 0 {
		if tp == _dotaTournamentsPlayers || tp == _dotaTournamentsTeams {
			params.Set("tournament_id", strconv.FormatInt(serieID, 10))
		} else {
			params.Set("serie_id", strconv.FormatInt(serieID, 10))
		}
	}
	if tp == _lolVerChampions {
		params.Set("version_name", s.c.Leidata.Hero.Version)
	}
	if res, count, err = s.setPages(tp, params, isOids); err != nil {
		log.Error("s.setPages tp(%s) error(%+v)", tp, err)
		return
	}
	for i := 2; i <= count; i++ {
		time.Sleep(time.Second)
		params.Set("page", strconv.Itoa(i))
		params.Set("per_page", strconv.Itoa(_perPage))
		if tmp, _, e := s.setPages(tp, params, isOids); e == nil {
			res = append(res, tmp...)
		}
	}
	return
}

func (s *Service) setPages(tp string, params url.Values, isOids bool) (res []int64, count int, err error) {
	var (
		rs   json.RawMessage
		oids []*mdlesp.Oid
	)
	if rs, count, err = s.leida(params, tp); err != nil {
		log.Error("setPages tp(%v) error(%v)", tp, err)
		return
	}
	if len(rs) == 0 {
		return
	}
	if isOids {
		if err = json.Unmarshal(rs, &oids); err != nil {
			log.Error("setPages json.Unmarshal tp(%v) rs(%+v) error(%v)", tp, rs, err)
			return
		}
		for _, oid := range oids {
			res = append(res, oid.ID)
		}
	} else {
		if err = s.writeInfo(tp, rs); err != nil {
			log.Error("setPages s.writeInfo tp(%v) rs(%+v) error(%v)", tp, rs, err)
		}
	}
	return
}

func (s *Service) leida(params url.Values, route string) (rs []byte, count int, err error) {
	var body, orginBody []byte
	params.Del("route")
	params.Set("key", s.c.Leidata.After.Key)
	url := s.c.Leidata.After.URL + "/" + route + "?" + params.Encode()
	for i := 0; i < s.c.Leidata.After.Retry; i++ {
		if body, err = s.dao.ThirdGet(context.Background(), url); err != nil {
			time.Sleep(time.Second)
			continue
		}
		bodyStr := string(body[:])
		if bodyStr == "" {
			time.Sleep(time.Second)
			continue
		}
		rsPos := strings.Index(bodyStr, "[")
		if rsPos > -1 {
			orginBody = body
			body = []byte(bodyStr[rsPos:])
		} else {
			time.Sleep(time.Second)
			continue
		}
		rs = body
		totalPos := strings.Index(bodyStr, "X-Total:")
		if totalPos > 0 {
			s := string(orginBody[totalPos+9 : rsPos-4])
			if t, e := strconv.ParseFloat(s, 64); e == nil {
				count = int(math.Ceil(t / float64(_perPage)))
			}
		}
		break
	}
	if err != nil {
		log.Error("json.Unmarshal url(%s) body(%s) error(%v)", url, string(body), err)
	}
	return
}

func decimal(f float64, n int) float64 {
	n10 := math.Pow10(n)
	return math.Trunc((f+0.5/n10)*n10) / n10
}

func strToFloat(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		log.Error("strToFloat s(%s) error(%+v)", s, err)
		return 0
	}
	n10 := math.Pow10(_decimal)
	return math.Trunc((f+0.5/n10)*n10) / n10
}
