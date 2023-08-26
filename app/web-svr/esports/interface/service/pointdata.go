package service

import (
	"context"
	"sort"
	"strings"

	xecode "go-common/library/ecode"
	"go-common/library/log"
	egV2 "go-common/library/sync/errgroup.v2"
	"go-gateway/app/web-svr/esports/interface/model"
)

var (
	_emptLolPlayer  = make([]*model.LolPlayer, 0)
	_emptLolTeam    = make([]*model.LolTeam, 0)
	_emptDotaPlayer = make([]*model.DotaPlayer, 0)
	_emptDotaTeam   = make([]*model.DotaTeam, 0)
	lolPlayerHeader = []model.Header{
		{Key: "kda", Name: "KDA"},
		{Key: "kills", Name: "场均击杀"},
		{Key: "assists", Name: "场均助攻"},
		{Key: "deaths", Name: "场均死亡"},
		{Key: "minions_killed", Name: "场均补兵"},
		{Key: "wards_placed", Name: "场均插眼"},
	}
	lolTeamHeader = []model.Header{
		{Key: "win", Name: "胜率"},
		{Key: "kda", Name: "KDA"},
		{Key: "tower_kills", Name: "场均推塔"},
		{Key: "kills", Name: "场均击杀"},
		{Key: "assists", Name: "场均助攻"},
		{Key: "deaths", Name: "场均死亡"},
		{Key: "inhibitor_kills", Name: "场均龙"},
		{Key: "wards_placed", Name: "场均插眼"},
		{Key: "first_blood", Name: "一血率"},
		{Key: "first_tower", Name: "一塔率"},
	}
	dotaPlayerHeader = []model.Header{
		{Key: "kda", Name: "KDA"},
		{Key: "kills", Name: "场均击杀"},
		{Key: "assists", Name: "场均助攻"},
		{Key: "deaths", Name: "场均死亡"},
		{Key: "last_hits", Name: "正补数"},
		{Key: "xp_per_minute", Name: "分均经验"},
		{Key: "gold_per_minute", Name: "分均金币"},
	}
	dotaTeamHeader = []model.Header{
		{Key: "win", Name: "胜率"},
		{Key: "kda", Name: "KDA"},
		{Key: "kills", Name: "场均击杀"},
		{Key: "last_hits", Name: "场均正补"},
		{Key: "gold_per_min", Name: "分均经济"},
		{Key: "damage_taken", Name: "场均伤害"},
		{Key: "tower_kills", Name: "场均推塔"},
		{Key: "observer_used", Name: "场均守卫"},
		{Key: "first_blood", Name: "一血率"},
	}
)

const (
	_sortASC            = 0
	_sortDESC           = 1
	_firstPage          = 1
	_recentCount        = 8
	_win                = "win"
	_towerKills         = "tower_kills"
	_goldEarned         = "gold_earned"
	_inhibitorKills     = "inhibitor_kills"
	_totalMinionsKilled = "total_minions_killed"
	_firstBlood         = "first_blood"
	_firstTower         = "first_tower"
	_goldPerMin         = "gold_per_min"
	_damageTaken        = "damage_taken"
	_observerUsed       = "observer_used"
	_kda                = "kda"
	_kills              = "kills"
	_deaths             = "deaths"
	_assists            = "assists"
	_minionsKilled      = "minions_killed"
	_wardsPlaced        = "wards_placed"
	_lastHits           = "last_hits"
	_xpPerMinute        = "xp_per_minute"
	_goldPerMinute      = "gold_per_minute"
	_mvp                = "mvp"
)

// Types return data page game types.
func (s *Service) Types(c context.Context) (list map[int64]string, err error) {
	list = make(map[int64]string, len(s.c.GameTypes))
	for _, tp := range s.c.GameTypes {
		list[tp.ID] = tp.Name
	}
	return
}

// Roles return lol dota roles.
func (s *Service) Roles(c context.Context, tp string) (list map[string]string, err error) {
	list = s.dao.Roles(tp)
	return
}

// LeidaSeasons return leida lol dota seasons.
func (s *Service) LeidaSeasons(c context.Context, tp, TeamID int64, pn, ps int) (list []*model.Season, count int, err error) {
	var (
		tmpList, rsList []*model.Season
		start           = (pn - 1) * ps
		end             = start + ps - 1
	)
	if tmpList, err = s.ldSeasons(c, tp); err != nil {
		log.Error("s.ldSeasons tp(%d) error(%+v)", tp, err)
		return
	}
	if TeamID > 0 {
		for _, season := range tmpList {
			if s.checkTeam(tp, season.LeidaSID, TeamID) {
				rsList = append(rsList, season)
			}
		}
	} else {
		rsList = tmpList
	}
	count = len(rsList)
	if count == 0 || count < start {
		list = _emptSeason
		return
	}
	if count > end {
		list = rsList[start : end+1]
	} else {
		list = rsList[start:]
	}
	return
}

func (s *Service) checkLolDataTeam(sid, teamID int64) bool {
	teams, ok := seasonLolDataTeamMap[sid]
	if !ok {
		return false
	}
	for _, team := range teams {
		if team.TeamID == teamID {
			return true
		}
	}
	return false
}

func (s *Service) checkTeam(tp, sid, teamID int64) bool {
	if sid == 0 {
		return false
	}
	if tp == _lolType {
		return s.checkLolDataTeam(sid, teamID)
	}
	if tp == _dotaType {
		for _, team := range s.dotaBigTeams.Data[sid] {
			if team.TeamID == teamID {
				return true
			}
		}
	}
	return false
}

func (s *Service) ldSeasons(c context.Context, tp int64) (rs []*model.Season, err error) {
	if rs, err = s.dao.GameSeasonCache(c, tp); err != nil {
		err = nil
	} else if len(rs) > 0 {
		return
	}
	gameID := s.mapGameDb[tp]
	if gameID == 0 {
		err = xecode.RequestErr
		return
	}
	if rs, err = s.dao.GameSeason(c, gameID); err != nil {
		log.Error("s.dao.GameSeason tp(%d) gameID(%d) error(%+v)", err, tp, gameID)
		return
	}
	if len(rs) > 0 {
		s.cache.Do(c, func(c context.Context) {
			s.dao.SetGameSeasonCache(c, tp, rs)
		})
	}
	return
}

// Game get game.
func (s *Service) Game(c context.Context, p *model.ParamGame) (rs map[int64]interface{}, err error) {
	var (
		count     int
		gameMap   map[int64]*model.LolGame
		owGameMap map[int64]*model.OwGame
		rsGame    interface{}
	)
	rs = make(map[int64]interface{}, len(p.GameIDs))
	if rsGame, err = s.ldGame(c, p.MatchID, p.Tp); err != nil || rsGame == nil {
		return
	}
	if owGames, ok := rsGame.([]*model.OwGame); ok {
		count = len(owGames)
		if count == 0 {
			return
		}
		owGameMap = make(map[int64]*model.OwGame, count)
		for _, game := range owGames {
			owGameMap[game.GameID] = game
		}
		for _, id := range p.GameIDs {
			if game, ok := owGameMap[id]; ok {
				rs[id] = game
			}
		}
	} else if games, ok := rsGame.([]*model.LolGame); ok {
		count = len(games)
		if count == 0 {
			return
		}
		gameMap = make(map[int64]*model.LolGame, count)
		for _, game := range games {
			gameMap[game.GameID] = game
		}
		for _, id := range p.GameIDs {
			if game, ok := gameMap[id]; ok {
				rs[id] = game
			}
		}
	}
	return
}

// Items get items.
func (s *Service) Items(c context.Context, p *model.ParamLeidas) (rs map[int64]*model.LdInfo, err error) {
	rs = make(map[int64]*model.LdInfo, len(p.IDs))
	if p.Tp == _lolType {
		for _, id := range p.IDs {
			if item, ok := s.lolItemsMap.Data[id]; ok {
				rs[id] = item
			}
		}
	} else if p.Tp == _dotaType {
		for _, id := range p.IDs {
			if item, ok := s.dotaItemsMap.Data[id]; ok {
				rs[id] = item
			}
		}
	} else if p.Tp == _owType {
		for _, id := range p.IDs {
			if item, ok := s.owMapsMap.Data[id]; ok {
				rs[id] = item
			}
		}
	}
	return
}

// Heroes lol:champions ; dota2 heroes.
func (s *Service) Heroes(c context.Context, p *model.ParamLeidas) (rs map[int64]*model.LdInfo, err error) {
	rs = make(map[int64]*model.LdInfo, len(p.IDs))
	if p.Tp == _lolType {
		for _, id := range p.IDs {
			if item, ok := s.lolChampions.Data[id]; ok {
				rs[id] = item
			}
		}
	} else if p.Tp == _dotaType {
		for _, id := range p.IDs {
			if item, ok := s.dotaHeroes.Data[id]; ok {
				rs[id] = item
			}
		}
	} else if p.Tp == _owType {
		for _, id := range p.IDs {
			if item, ok := s.owHeroes.Data[id]; ok {
				rs[id] = item
			}
		}
	}
	return
}

// Abilities lol:spells;dota2:abilities.
func (s *Service) Abilities(c context.Context, p *model.ParamLeidas) (rs interface{}, err error) {
	infos := make(map[int64]*model.LdInfo, len(p.IDs))
	if p.Tp == _lolType {
		for _, id := range p.IDs {
			if info, ok := s.lolSpells.Data[id]; ok {
				infos[id] = info
			}
		}
		rs = infos
	} else if p.Tp == _dotaType {
		for _, id := range p.IDs {
			if info, ok := s.dotaAbilities.Data[id]; ok {
				infos[id] = info
			}
		}
		rs = infos
	}
	return
}

// Players get players.
func (s *Service) Players(c context.Context, p *model.ParamLeidas) (rs map[int64]*model.LdInfo, err error) {
	rs = make(map[int64]*model.LdInfo, len(p.IDs))
	if p.Tp == _lolType {
		for _, id := range p.IDs {
			if info, ok := s.lolPlayers.Data[id]; ok {
				rs[id] = info
			}
		}
	} else if p.Tp == _dotaType {
		for _, id := range p.IDs {
			if info, ok := s.dotaPlayers.Data[id]; ok {
				rs[id] = info
			}
		}
	} else if p.Tp == _owType {
		for _, id := range p.IDs {
			if info, ok := s.owPlayers.Data[id]; ok {
				rs[id] = info
			}
		}
	}
	return
}

// Teams get teams.
func (s *Service) Teams(c context.Context, p *model.ParamLeidas) (rs map[int64]*model.LdInfo, err error) {
	rs = make(map[int64]*model.LdInfo, len(p.IDs))
	if p.Tp == _lolType {
		for _, id := range p.IDs {
			if info, ok := s.lolTeams.Data[id]; ok {
				rs[id] = info
			}
		}
	} else if p.Tp == _dotaType {
		for _, id := range p.IDs {
			if info, ok := s.dotaTeams.Data[id]; ok {
				rs[id] = info
			}
		}
	} else if p.Tp == _owType {
		for _, id := range p.IDs {
			if info, ok := s.owTeams.Data[id]; ok {
				rs[id] = info
			}
		}
	}
	return
}

// BigPlayers big players.
func (s *Service) BigPlayers(c context.Context, p *model.StatsBig) (res interface{}, header interface{}, count int, err error) {
	if p.Tp == _lolType {
		res, count, err = s.LolPlayers(c, p)
		header = lolPlayerHeader
	} else if p.Tp == _dotaType {
		res, count, err = s.DotaPlayers(c, p)
		header = dotaPlayerHeader
	} else {
		err = xecode.RequestErr
	}
	return
}

// BigTeams big teams.
func (s *Service) BigTeams(c context.Context, p *model.StatsBig) (res interface{}, header interface{}, count int, err error) {
	if p.Tp == _lolType {
		res, count, err = s.LolTeams(c, p)
		header = lolTeamHeader
	} else if p.Tp == _dotaType {
		res, count, err = s.DotaTeams(c, p)
		header = dotaTeamHeader
	} else {
		err = xecode.RequestErr
	}
	return
}

// LolPlayers lol players stats.
func (s *Service) LolPlayers(c context.Context, p *model.StatsBig) (res []*model.LolPlayer, count int, err error) {
	var (
		players = make([]*model.LolPlayer, 0)
		start   = (p.Pn - 1) * p.Ps
		end     = start + p.Ps - 1
	)
	tmpPlayers, ok := seasonLolDataPlayerMap[p.Sid]
	if !ok {
		if tmpPlayers, err = s.FetchLolDataPlayer(c, p.Sid); err != nil {
			log.Errorc(c, "LolPlayers s.FetchLolDataPlayer() sid(%d) error(%+v)", p.Sid, err)
			return
		}
	}
	if len(tmpPlayers) == 0 {
		res = _emptLolPlayer
		return
	}
	if p.Role == "" {
		players = tmpPlayers
	} else {
		for _, player := range tmpPlayers {
			if player.Role == p.Role {
				players = append(players, player)
			}
		}
	}
	count = len(players)
	if count < start {
		res = _emptLolPlayer
		return
	}
	s.lolPlayerSort(p.SortValue, p.SortType, players)
	if count > end {
		res = players[start : end+1]
	} else {
		res = players[start:]
	}
	return
}

func (s *Service) lolPlayerSort(order int, field string, players []*model.LolPlayer) {
	switch field {
	case _win:
		sort.Slice(players, func(i, j int) bool {
			if players[i].Win != players[j].Win {
				if order == _sortASC {
					return players[i].Win < players[j].Win
				} else if order == _sortDESC {
					return players[i].Win > players[j].Win
				}
			}
			return players[i].ID < players[j].ID
		})
	case _kda:
		sort.Slice(players, func(i, j int) bool {
			if players[i].KDA != players[j].KDA {
				if order == _sortASC {
					return players[i].KDA < players[j].KDA
				} else if order == _sortDESC {
					return players[i].KDA > players[j].KDA
				}
			}
			return players[i].ID < players[j].ID
		})
	case _kills:
		sort.Slice(players, func(i, j int) bool {
			if players[i].Kills != players[j].Kills {
				if order == _sortASC {
					return players[i].Kills < players[j].Kills
				} else if order == _sortDESC {
					return players[i].Kills > players[j].Kills
				}
			}
			return players[i].ID < players[j].ID
		})
	case _deaths:
		sort.Slice(players, func(i, j int) bool {
			if players[i].Deaths != players[j].Deaths {
				if order == _sortASC {
					return players[i].Deaths < players[j].Deaths
				} else if order == _sortDESC {
					return players[i].Deaths > players[j].Deaths
				}
			}
			return players[i].ID < players[j].ID
		})
	case _assists:
		sort.Slice(players, func(i, j int) bool {
			if players[i].Assists != players[j].Assists {
				if order == _sortASC {
					return players[i].Assists < players[j].Assists
				} else if order == _sortDESC {
					return players[i].Assists > players[j].Assists
				}
			}
			return players[i].ID < players[j].ID
		})
	case _minionsKilled:
		sort.Slice(players, func(i, j int) bool {
			if players[i].MinionsKilled != players[j].MinionsKilled {
				if order == _sortASC {
					return players[i].MinionsKilled < players[j].MinionsKilled
				} else if order == _sortDESC {
					return players[i].MinionsKilled > players[j].MinionsKilled
				}
			}
			return players[i].ID < players[j].ID
		})
	case _wardsPlaced:
		sort.Slice(players, func(i, j int) bool {
			if players[i].WardsPlaced != players[j].WardsPlaced {
				if order == _sortASC {
					return players[i].WardsPlaced < players[j].WardsPlaced
				} else if order == _sortDESC {
					return players[i].WardsPlaced > players[j].WardsPlaced
				}
			}
			return players[i].ID < players[j].ID
		})
	case _mvp:
		sort.Slice(players, func(i, j int) bool {
			if players[i].MVP != players[j].MVP {
				if order == _sortASC {
					return players[i].MVP < players[j].MVP
				} else if order == _sortDESC {
					return players[i].MVP > players[j].MVP
				}
			}
			return players[i].ID < players[j].ID
		})
	default:
		sort.Slice(players, func(i, j int) bool {
			if players[i].KDA != players[j].KDA {
				if order == _sortASC {
					return players[i].KDA < players[j].KDA
				} else if order == _sortDESC {
					return players[i].KDA > players[j].KDA
				}
			}
			return players[i].ID < players[j].ID
		})
	}
}

// LolTeams lol team stats.
func (s *Service) LolTeams(c context.Context, p *model.StatsBig) (res []*model.LolTeam, count int, err error) {
	var (
		start = (p.Pn - 1) * p.Ps
		end   = start + p.Ps - 1
	)
	teams, ok := seasonLolDataTeamMap[p.Sid]
	if !ok {
		if teams, err = s.FetchLolDataTeam(c, p.Sid); err != nil {
			log.Errorc(c, "LolPlayers s.FetchLolDataTeam() sid(%d) error(%+v)", p.Sid, err)
			return
		}
	}
	count = len(teams)
	if count == 0 || count < start {
		res = _emptLolTeam
		return
	}
	s.lolTeamSort(p.SortValue, p.SortType, teams)
	if count > end {
		res = teams[start : end+1]
	} else {
		res = teams[start:]
	}
	return
}

func (s *Service) lolTeamSort(order int, field string, teams []*model.LolTeam) {
	switch field {
	case _win:
		sort.Slice(teams, func(i, j int) bool {
			if teams[i].Win != teams[j].Win {
				if order == _sortASC {
					return teams[i].Win < teams[j].Win
				} else if order == _sortDESC {
					return teams[i].Win > teams[j].Win
				}
			}
			return teams[i].ID < teams[j].ID
		})
	case _kda:
		sort.Slice(teams, func(i, j int) bool {
			if teams[i].KDA != teams[j].KDA {
				if order == _sortASC {
					return teams[i].KDA < teams[j].KDA
				} else if order == _sortDESC {
					return teams[i].KDA > teams[j].KDA
				}
			}
			return teams[i].ID < teams[j].ID
		})
	case _kills:
		sort.Slice(teams, func(i, j int) bool {
			if teams[i].Kills != teams[j].Kills {
				if order == _sortASC {
					return teams[i].Kills < teams[j].Kills
				} else if order == _sortDESC {
					return teams[i].Kills > teams[j].Kills
				}
			}
			return teams[i].ID < teams[j].ID
		})
	case _deaths:
		sort.Slice(teams, func(i, j int) bool {
			if teams[i].Deaths != teams[j].Deaths {
				if order == _sortASC {
					return teams[i].Deaths < teams[j].Deaths
				} else if order == _sortDESC {
					return teams[i].Deaths > teams[j].Deaths
				}
			}
			return teams[i].ID < teams[j].ID
		})
	case _towerKills:
		sort.Slice(teams, func(i, j int) bool {
			if teams[i].TowerKills != teams[j].TowerKills {
				if order == _sortASC {
					return teams[i].TowerKills < teams[j].TowerKills
				} else if order == _sortDESC {
					return teams[i].TowerKills > teams[j].TowerKills
				}
			}
			return teams[i].ID < teams[j].ID
		})
	case _goldEarned:
		sort.Slice(teams, func(i, j int) bool {
			if teams[i].GoldEarned != teams[j].GoldEarned {
				if order == _sortASC {
					return teams[i].GoldEarned < teams[j].GoldEarned
				} else if order == _sortDESC {
					return teams[i].GoldEarned > teams[j].GoldEarned
				}
			}
			return teams[i].ID < teams[j].ID
		})
	case _inhibitorKills:
		sort.Slice(teams, func(i, j int) bool {
			if teams[i].InhibitorKills != teams[j].InhibitorKills {
				if order == _sortASC {
					return teams[i].InhibitorKills < teams[j].InhibitorKills
				} else if order == _sortDESC {
					return teams[i].InhibitorKills > teams[j].InhibitorKills
				}
			}
			return teams[i].ID < teams[j].ID
		})
	case _totalMinionsKilled:
		sort.Slice(teams, func(i, j int) bool {
			if teams[i].TotalMinionsKilled != teams[j].TotalMinionsKilled {
				if order == _sortASC {
					return teams[i].TotalMinionsKilled < teams[j].TotalMinionsKilled
				} else if order == _sortDESC {
					return teams[i].TotalMinionsKilled > teams[j].TotalMinionsKilled
				}
			}
			return teams[i].ID < teams[j].ID
		})
	case _wardsPlaced:
		sort.Slice(teams, func(i, j int) bool {
			if teams[i].WardsPlaced != teams[j].WardsPlaced {
				if order == _sortASC {
					return teams[i].WardsPlaced < teams[j].WardsPlaced
				} else if order == _sortDESC {
					return teams[i].WardsPlaced > teams[j].WardsPlaced
				}
			}
			return teams[i].ID < teams[j].ID
		})
	case _firstBlood:
		sort.Slice(teams, func(i, j int) bool {
			if teams[i].FirstBlood != teams[j].FirstBlood {
				if order == _sortASC {
					return teams[i].FirstBlood < teams[j].FirstBlood
				} else if order == _sortDESC {
					return teams[i].FirstBlood > teams[j].FirstBlood
				}
			}
			return teams[i].ID < teams[j].ID
		})
	case _firstTower:
		sort.Slice(teams, func(i, j int) bool {
			if teams[i].FirstTower != teams[j].FirstTower {
				if order == _sortASC {
					return teams[i].FirstTower < teams[j].FirstTower
				} else if order == _sortDESC {
					return teams[i].FirstTower > teams[j].FirstTower
				}
			}
			return teams[i].ID < teams[j].ID
		})
	case _assists:
		sort.Slice(teams, func(i, j int) bool {
			if teams[i].Assists != teams[j].Assists {
				if order == _sortASC {
					return teams[i].Assists < teams[j].Assists
				} else if order == _sortDESC {
					return teams[i].Assists > teams[j].Assists
				}
			}
			return teams[i].ID < teams[j].ID
		})
	default:
		sort.Slice(teams, func(i, j int) bool {
			if teams[i].Win != teams[j].Win {
				if order == _sortASC {
					return teams[i].Win < teams[j].Win
				} else if order == _sortDESC {
					return teams[i].Win > teams[j].Win
				}
			}
			return teams[i].ID < teams[j].ID
		})
	}
}

// DotaPlayers dota players stats.
func (s *Service) DotaPlayers(c context.Context, p *model.StatsBig) (res []*model.DotaPlayer, count int, err error) {
	var (
		players    = make([]*model.DotaPlayer, 0)
		tmpPlayers []*model.DotaPlayer
		start      = (p.Pn - 1) * p.Ps
		end        = start + p.Ps - 1
	)
	tmpPlayers = make([]*model.DotaPlayer, len(s.dotaBigPlayers.Data[p.Sid]))
	copy(tmpPlayers, s.dotaBigPlayers.Data[p.Sid])
	if len(tmpPlayers) == 0 {
		res = _emptDotaPlayer
		return
	}
	if p.Role == "" {
		players = tmpPlayers
	} else {
		for _, player := range tmpPlayers {
			roles := strings.Split(player.Role, "/")
			for _, role := range roles {
				if role == p.Role {
					players = append(players, player)
					break
				}
			}
		}
	}
	count = len(players)
	if count < start {
		res = _emptDotaPlayer
		return
	}
	s.dotaPlayerSort(p.SortValue, p.SortType, players)
	if count > end {
		res = players[start : end+1]
	} else {
		res = players[start:]
	}
	return
}

func (s *Service) dotaPlayerSort(order int, field string, players []*model.DotaPlayer) {
	switch field {
	case _win:
		sort.Slice(players, func(i, j int) bool {
			if players[i].Win != players[j].Win {
				if order == _sortASC {
					return players[i].Win < players[j].Win
				} else if order == _sortDESC {
					return players[i].Win > players[j].Win
				}
			}
			return players[i].ID < players[j].ID
		})
	case _kda:
		sort.Slice(players, func(i, j int) bool {
			if players[i].KDA != players[j].KDA {
				if order == _sortASC {
					return players[i].KDA < players[j].KDA
				} else if order == _sortDESC {
					return players[i].KDA > players[j].KDA
				}
			}
			return players[i].ID < players[j].ID
		})
	case _kills:
		sort.Slice(players, func(i, j int) bool {
			if players[i].Kills != players[j].Kills {
				if order == _sortASC {
					return players[i].Kills < players[j].Kills
				} else if order == _sortDESC {
					return players[i].Kills > players[j].Kills
				}
			}
			return players[i].ID < players[j].ID
		})
	case _deaths:
		sort.Slice(players, func(i, j int) bool {
			if players[i].Deaths != players[j].Deaths {
				if order == _sortASC {
					return players[i].Deaths < players[j].Deaths
				} else if order == _sortDESC {
					return players[i].Deaths > players[j].Deaths
				}
			}
			return players[i].ID < players[j].ID
		})
	case _assists:
		sort.Slice(players, func(i, j int) bool {
			if players[i].Assists != players[j].Assists {
				if order == _sortASC {
					return players[i].Assists < players[j].Assists
				} else if order == _sortDESC {
					return players[i].Assists > players[j].Assists
				}
			}
			return players[i].ID < players[j].ID
		})
	case _lastHits:
		sort.Slice(players, func(i, j int) bool {
			if players[i].LastHits != players[j].LastHits {
				if order == _sortASC {
					return players[i].LastHits < players[j].LastHits
				} else if order == _sortDESC {
					return players[i].LastHits > players[j].LastHits
				}
			}
			return players[i].ID < players[j].ID
		})
	case _xpPerMinute:
		sort.Slice(players, func(i, j int) bool {
			if players[i].XpPerMinute != players[j].XpPerMinute {
				if order == _sortASC {
					return players[i].XpPerMinute < players[j].XpPerMinute
				} else if order == _sortDESC {
					return players[i].XpPerMinute > players[j].XpPerMinute
				}
			}
			return players[i].ID < players[j].ID
		})
	case _goldPerMinute:
		sort.Slice(players, func(i, j int) bool {
			if players[i].GoldPerMinute != players[j].GoldPerMinute {
				if order == _sortASC {
					return players[i].GoldPerMinute < players[j].GoldPerMinute
				} else if order == _sortDESC {
					return players[i].GoldPerMinute > players[j].GoldPerMinute
				}
			}
			return players[i].ID < players[j].ID
		})
	default:
		sort.Slice(players, func(i, j int) bool {
			if players[i].KDA != players[j].KDA {
				if order == _sortASC {
					return players[i].KDA < players[j].KDA
				} else if order == _sortDESC {
					return players[i].KDA > players[j].KDA
				}
			}
			return players[i].ID < players[j].ID
		})
	}
}

// DotaTeams dota team stats.
func (s *Service) DotaTeams(c context.Context, p *model.StatsBig) (res []*model.DotaTeam, count int, err error) {
	var (
		teams []*model.DotaTeam
		start = (p.Pn - 1) * p.Ps
		end   = start + p.Ps - 1
	)
	count = len(s.dotaBigTeams.Data[p.Sid])
	if count == 0 || count < start {
		res = _emptDotaTeam
		return
	}
	teams = make([]*model.DotaTeam, count)
	copy(teams, s.dotaBigTeams.Data[p.Sid])
	s.dotaTeamSort(p.SortValue, p.SortType, teams)
	if count > end {
		res = teams[start : end+1]
	} else {
		res = teams[start:]
	}
	return
}

func (s *Service) dotaTeamSort(order int, field string, teams []*model.DotaTeam) {
	switch field {
	case _win:
		sort.Slice(teams, func(i, j int) bool {
			if teams[i].Win != teams[j].Win {
				if order == _sortASC {
					return teams[i].Win < teams[j].Win
				} else if order == _sortDESC {
					return teams[i].Win > teams[j].Win
				}
			}
			return teams[i].ID < teams[j].ID
		})
	case _kda:
		sort.Slice(teams, func(i, j int) bool {
			if teams[i].KDA != teams[j].KDA {
				if order == _sortASC {
					return teams[i].KDA < teams[j].KDA
				} else if order == _sortDESC {
					return teams[i].KDA > teams[j].KDA
				}
			}
			return teams[i].ID < teams[j].ID
		})
	case _kills:
		sort.Slice(teams, func(i, j int) bool {
			if teams[i].Kills != teams[j].Kills {
				if order == _sortASC {
					return teams[i].Kills < teams[j].Kills
				} else if order == _sortDESC {
					return teams[i].Kills > teams[j].Kills
				}
			}
			return teams[i].ID < teams[j].ID
		})
	case _deaths:
		sort.Slice(teams, func(i, j int) bool {
			if teams[i].Deaths != teams[j].Deaths {
				if order == _sortASC {
					return teams[i].Deaths < teams[j].Deaths
				} else if order == _sortDESC {
					return teams[i].Deaths > teams[j].Deaths
				}
			}
			return teams[i].ID < teams[j].ID
		})
	case _towerKills:
		sort.Slice(teams, func(i, j int) bool {
			if teams[i].TowerKills != teams[j].TowerKills {
				if order == _sortASC {
					return teams[i].TowerKills < teams[j].TowerKills
				} else if order == _sortDESC {
					return teams[i].TowerKills > teams[j].TowerKills
				}
			}
			return teams[i].ID < teams[j].ID
		})
	case _lastHits:
		sort.Slice(teams, func(i, j int) bool {
			if teams[i].LastHits != teams[j].LastHits {
				if order == _sortASC {
					return teams[i].LastHits < teams[j].LastHits
				} else if order == _sortDESC {
					return teams[i].LastHits > teams[j].LastHits
				}
			}
			return teams[i].ID < teams[j].ID
		})
	case _goldPerMin:
		sort.Slice(teams, func(i, j int) bool {
			if teams[i].GoldPerMin != teams[j].GoldPerMin {
				if order == _sortASC {
					return teams[i].GoldPerMin < teams[j].GoldPerMin
				} else if order == _sortDESC {
					return teams[i].GoldPerMin > teams[j].GoldPerMin
				}
			}
			return teams[i].ID < teams[j].ID
		})
	case _damageTaken:
		sort.Slice(teams, func(i, j int) bool {
			if teams[i].DamageTaken != teams[j].DamageTaken {
				if order == _sortASC {
					return teams[i].DamageTaken < teams[j].DamageTaken
				} else if order == _sortDESC {
					return teams[i].DamageTaken > teams[j].DamageTaken
				}
			}
			return teams[i].ID < teams[j].ID
		})
	case _observerUsed:
		sort.Slice(teams, func(i, j int) bool {
			if teams[i].ObserverUsed != teams[j].ObserverUsed {
				if order == _sortASC {
					return teams[i].ObserverUsed < teams[j].ObserverUsed
				} else if order == _sortDESC {
					return teams[i].ObserverUsed > teams[j].ObserverUsed
				}
			}
			return teams[i].ID < teams[j].ID
		})
	case _firstBlood:
		sort.Slice(teams, func(i, j int) bool {
			if teams[i].FirstBlood != teams[j].FirstBlood {
				if order == _sortASC {
					return teams[i].FirstBlood < teams[j].FirstBlood
				} else if order == _sortDESC {
					return teams[i].FirstBlood > teams[j].FirstBlood
				}
			}
			return teams[i].ID < teams[j].ID
		})
	default:
		sort.Slice(teams, func(i, j int) bool {
			if teams[i].Win != teams[j].Win {
				if order == _sortASC {
					return teams[i].Win < teams[j].Win
				} else if order == _sortDESC {
					return teams[i].Win > teams[j].Win
				}
			}
			return teams[i].ID < teams[j].ID
		})
	}
}

// SpecialTeams big teams.
func (s *Service) SpecialTeams(c context.Context, p *model.ParamSpecTeams) (res []*model.LdTeam, count int, err error) {
	var (
		bigTeams, empAcronym, allTeams []*model.LdTeam
		start                          = (p.Pn - 1) * p.Ps
		end                            = start + p.Ps - 1
	)
	bigTeams, empAcronym = s.bigLdTeams(c, p.Tp, p.LeidaSID)
	count = len(bigTeams) + len(empAcronym)
	if count == 0 || count < start {
		res = _emptLdTeams
		return
	}
	if len(bigTeams) > 0 {
		sort.Slice(bigTeams, func(i, j int) bool {
			if p.Sort == _sortASC {
				if bigTeams[i].Acronym != bigTeams[j].Acronym {
					return bigTeams[i].Acronym < bigTeams[j].Acronym
				}
			} else if p.Sort == _sortDESC {
				if bigTeams[i].Acronym != bigTeams[j].Acronym {
					return bigTeams[i].Acronym > bigTeams[j].Acronym
				}
			}
			return bigTeams[i].TeamID < bigTeams[j].TeamID
		})
		allTeams = append(allTeams, bigTeams...)
	}
	if len(empAcronym) > 0 {
		sort.Slice(empAcronym, func(i, j int) bool {
			if p.Sort == _sortASC {
				if empAcronym[i].Name != empAcronym[j].Name {
					return empAcronym[i].Name < empAcronym[j].Name
				}
			} else if p.Sort == _sortDESC {
				if empAcronym[i].Name != empAcronym[j].Name {
					return empAcronym[i].Name > empAcronym[j].Name
				}
			}
			return empAcronym[i].TeamID < empAcronym[j].TeamID
		})
		allTeams = append(allTeams, empAcronym...)
	}
	if count > end+1 {
		res = allTeams[start : end+1]
	} else {
		res = allTeams[start:]
	}
	return
}

func (s *Service) getLolDataTeams(c context.Context, tp, sid int64) (rs []*model.LdTeam, empAcronym []*model.LdTeam) {
	var (
		tmpMap  map[int64]struct{}
		acronym string
	)
	tmpMap = make(map[int64]struct{})
	var err error
	teams, ok := seasonLolDataTeamMap[sid]
	if !ok {
		if teams, err = s.FetchLolDataTeam(c, sid); err != nil {
			log.Errorc(c, "bigLdTeams s.FetchLolDataTeam() sid(%d) error(%+v)", sid, err)
			return
		}
	}
	for _, bigTeam := range teams {
		if _, ok := tmpMap[bigTeam.TeamID]; ok {
			continue
		}
		tmpMap[bigTeam.TeamID] = struct{}{}
		acronym = strings.TrimSpace(bigTeam.Acronym)
		if acronym == "" {
			empAcronym = append(empAcronym, &model.LdTeam{ID: bigTeam.ID, TeamID: bigTeam.TeamID, Acronym: acronym, LeidaSID: bigTeam.LeidaSID, Name: bigTeam.Name, ImageURL: bigTeam.ImageURL, GameType: tp})
		} else {
			rs = append(rs, &model.LdTeam{ID: bigTeam.ID, TeamID: bigTeam.TeamID, Acronym: acronym, LeidaSID: bigTeam.LeidaSID, Name: bigTeam.Name, ImageURL: bigTeam.ImageURL, GameType: tp})
		}
	}
	return
}

func (s *Service) getDotaDataTeams(tp, sid int64) (rs []*model.LdTeam, empAcronym []*model.LdTeam) {
	var (
		tmpMap  map[int64]struct{}
		acronym string
	)
	tmpMap = make(map[int64]struct{})
	if bigTeams, ok := s.dotaBigTeams.Data[sid]; ok {
		for _, bigTeam := range bigTeams {
			if _, ok := tmpMap[bigTeam.TeamID]; ok {
				continue
			}
			tmpMap[bigTeam.TeamID] = struct{}{}
			acronym = strings.TrimSpace(bigTeam.Acronym)
			if acronym == "" {
				empAcronym = append(empAcronym, &model.LdTeam{ID: bigTeam.ID, TeamID: bigTeam.TeamID, Acronym: acronym, LeidaSID: bigTeam.LeidaSID, Name: bigTeam.Name, ImageURL: bigTeam.ImageURL, GameType: tp})
			} else {
				rs = append(rs, &model.LdTeam{ID: bigTeam.ID, TeamID: bigTeam.TeamID, Acronym: acronym, LeidaSID: bigTeam.LeidaSID, Name: bigTeam.Name, ImageURL: bigTeam.ImageURL, GameType: tp})
			}
		}
	}
	return
}

func (s *Service) bigLdTeams(c context.Context, tp, sid int64) (rs []*model.LdTeam, empAcronym []*model.LdTeam) {
	if sid == 0 {
		return
	}
	if tp == _lolType {
		return s.getLolDataTeams(c, tp, sid)
	} else if tp == _dotaType {
		return s.getDotaDataTeams(tp, sid)
	}
	return
}

// SpecTeam specail team.
func (s *Service) SpecTeam(c context.Context, mid int64, p *model.ParamSpecial) (res model.SpecialTeam, err error) {
	var team *model.Team
	if res, err = s.dao.SpecTeamCache(c, p); err != nil || res.Team == nil {
		res.GID = s.mapGameDb[p.Tp]
		if team, err = s.dao.LdTeam(c, p.ID); err != nil {
			log.Error("s.dao.LdTeam error(%+v)", err)
			err = nil
		}
		if team != nil {
			res.Team = team
		} else {
			res.Team = struct{}{}
		}
		eg := egV2.WithContext(c)
		eg.Go(func(ctx context.Context) (err error) {
			if team != nil && team.ID > 0 && p.Recent == 1 {
				res.Recent = s.teamRecent(ctx, mid, team.ID)
			}
			if len(res.Recent) == 0 {
				res.Recent = _emptContest
			}
			return
		})
		eg.Go(func(ctx context.Context) (err error) {
			res.Stats = s.teamStats(ctx, p)
			return
		})
		eg.Wait()
		if res.Team != nil {
			s.cache.Do(c, func(c context.Context) {
				s.dao.AddSpecTeamCache(c, p, res)
			})
		}
	}
	return
}

func (s *Service) getLolTeamStats(c context.Context, p *model.ParamSpecial) *model.LolTeam {
	var err error
	teams, ok := seasonLolDataTeamMap[p.LeidaSID]
	if !ok {
		if teams, err = s.FetchLolDataTeam(c, p.LeidaSID); err != nil {
			log.Errorc(c, "teamStats s.FetchLolDataTeam() sid(%d) error(%+v)", p.LeidaSID, err)
			return nil
		}
	}
	for _, lolTeamStat := range teams {
		if lolTeamStat.TeamID == p.ID {
			return lolTeamStat
		}
	}
	return nil
}

func (s *Service) teamStats(c context.Context, p *model.ParamSpecial) (stats interface{}) {
	if p.Tp == _lolType {
		stats = s.getLolTeamStats(c, p)
	} else if p.Tp == _dotaType {
		if dotaTeamStats, ok := s.dotaBigTeams.Data[p.LeidaSID]; ok {
			for _, dotaTeamStat := range dotaTeamStats {
				if dotaTeamStat.TeamID == p.ID {
					stats = dotaTeamStat
					break
				}
			}
		}
	}
	if stats == nil {
		stats = struct{}{}
	}
	return
}

func (s *Service) teamRecent(c context.Context, mid, tid int64) (rs []*model.Contest) {
	var (
		err                    error
		tmpRs, liveRs, otherRs []*model.Contest
	)
	if tmpRs, _, err = s.ListContest(c, mid, &model.ParamContest{Tid: tid, Sort: _sortDESC, Pn: _firstPage, Ps: _recentCount}); err != nil {
		err = nil
	}
	if len(tmpRs) == 0 {
		return
	}
	for _, contest := range tmpRs {
		if contest.GameState == _gameLive {
			liveRs = append(liveRs, contest)
		} else {
			otherRs = append(otherRs, contest)
		}
	}
	if len(liveRs) > 0 {
		rs = append(rs, liveRs...)
		rs = append(rs, otherRs...)
	} else {
		rs = tmpRs
	}
	return
}

// SpecPlayer specail player.
func (s *Service) SpecPlayer(c context.Context, p *model.ParamSpecial) (res interface{}, err error) {
	res = s.playerStats(c, p)
	return
}

// PlayerRecent specail player recent.
func (s *Service) PlayerRecent(c context.Context, mid int64, p *model.ParamRecent) (res []*model.Contest, err error) {
	var (
		team   *model.Team
		season *model.Season
	)
	eg := egV2.WithContext(c)
	eg.Go(func(ctx context.Context) (err error) {
		if team, err = s.dao.LdTeam(c, p.LeidaTID); err != nil {
			log.Error("s.dao.LdTeam error(%+v)", err)
		}
		return
	})
	eg.Go(func(ctx context.Context) (err error) {
		if season, err = s.dao.LdSeason(c, p.LeidaSID); err != nil {
			log.Error("s.dao.LdSeason error(%+v)", err)
		}
		return
	})
	eg.Wait()
	if team == nil || season == nil {
		res = _emptContest
		return
	}
	if res, _, err = s.ListContest(c, mid, &model.ParamContest{Tid: team.ID, Sids: []int64{season.ID}, Sort: _sortDESC, Pn: p.Pn, Ps: p.Ps}); err != nil {
		err = nil
	}
	return
}

func (s *Service) getLolPlayerStats(c context.Context, p *model.ParamSpecial) *model.LolPlayer {
	var err error
	tmpPlayers, ok := seasonLolDataPlayerMap[p.LeidaSID]
	if !ok {
		if tmpPlayers, err = s.FetchLolDataPlayer(c, p.LeidaSID); err != nil {
			log.Errorc(c, "playerStats s.FetchLolDataPlayer() sid(%d) error(%+v)", p.LeidaSID, err)
			return nil
		}
	}
	for _, lolPlayerStat := range tmpPlayers {
		if lolPlayerStat.PlayerID == p.ID {
			return lolPlayerStat
		}
	}
	return nil
}

func (s *Service) playerStats(c context.Context, p *model.ParamSpecial) (stats interface{}) {
	if p.Tp == _lolType {
		stats = s.getLolPlayerStats(c, p)
	} else if p.Tp == _dotaType {
		if dotaPlayerStats, ok := s.dotaBigPlayers.Data[p.LeidaSID]; ok {
			for _, dotaPlayerStat := range dotaPlayerStats {
				if dotaPlayerStat.PlayerID == p.ID {
					stats = dotaPlayerStat
					break
				}
			}
		}
	}
	if stats == nil {
		stats = struct{}{}
	}
	return
}
