package service

import (
	"context"
	"sort"
	"time"

	"go-gateway/app/web-svr/esports/interface/component"
	"go-gateway/app/web-svr/esports/interface/conf"
	"go-gateway/app/web-svr/esports/interface/model"
)

const (
	analysisType4Team = iota + 1
	analysisType4Player
	analysisType4Hero
)

const (
	sortDesc = iota + 1
	sortAsc
)

const (
	sortKey4TeamOfTotalRound = iota + 1
	sortKey4TeamOfWinRate
	sortKey4TeamOfAvgKills
	sortKey4TeamOfAvgAssists
	sortKey4TeamOfAvgDeaths
	sortKey4TeamOfAvgTowerDestroyed
	sortKey4TeamOfAvgSmallDargon
	sortKey4TeamOfAvgBigDargon
)

const (
	sortKey4PlayerOfTotalRound = iota + 1
	sortKey4PlayerOfKDA
	sortKey4PlayerOfAvgKills
	sortKey4PlayerOfAvgAssists
	sortKey4PlayerOfAvgDeaths
	sortKey4PlayerOfEconomyMin
	sortKey4PlayerOfDamageRate
	sortKey4PlayerOfDamageTakenRate
	sortKey4PlayerOfParticipateRate
)

const (
	sortKey4HeroOfTotalRound = iota + 1
	sortKey4HeroOfWinRate
	sortKey4HeroOfBaned
	sortKey4HeroOfBanedRate
	sortKey4HeroOfPicked
	sortKey4HeroOfPickedRate
)

var (
	scoreAnalysis4Team4Live   []*model.ScoreTeamAnalysis4Live
	scoreAnalysis4Player4Live []*model.ScorePlayerAnalysis4Live
	scoreAnalysis4Hero4Live   []*model.ScoreHeroAnalysis4Live

	highWinTeam      *model.ScoreTeamAnalysis
	maxKillsTeam     *model.ScoreTeamAnalysis
	maxBigDargonTeam *model.ScoreTeamAnalysis

	maxKillsPlayer   *model.ScorePlayerAnalysis
	maxAssistsPlayer *model.ScorePlayerAnalysis
	minDeathsPlayer  *model.ScorePlayerAnalysis

	highPickedHero *model.ScoreHeroAnalysis
	highBanedHero  *model.ScoreHeroAnalysis
	highWinHero    *model.ScoreHeroAnalysis
)

func init() {
	scoreAnalysis4Team4Live = make([]*model.ScoreTeamAnalysis4Live, 0)
	scoreAnalysis4Player4Live = make([]*model.ScorePlayerAnalysis4Live, 0)
	scoreAnalysis4Hero4Live = make([]*model.ScoreHeroAnalysis4Live, 0)

	highWinTeam = new(model.ScoreTeamAnalysis)
	maxKillsTeam = new(model.ScoreTeamAnalysis)
	maxBigDargonTeam = new(model.ScoreTeamAnalysis)

	maxKillsPlayer = new(model.ScorePlayerAnalysis)
	maxAssistsPlayer = new(model.ScorePlayerAnalysis)
	minDeathsPlayer = new(model.ScorePlayerAnalysis)

	highPickedHero = new(model.ScoreHeroAnalysis)
	highBanedHero = new(model.ScoreHeroAnalysis)
	highWinHero = new(model.ScoreHeroAnalysis)
}

func watchS10ScoreAnalysis(ctx context.Context) {
	ticker := time.NewTicker(time.Minute * 5)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			watchS10ScoreAnalysisBySeasonWatch(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func loadMaxBigDargonTeam() map[string]interface{} {
	m := make(map[string]interface{}, 0)
	{
		m["has"] = false
		m["teamName"] = ""
		m["teamImage"] = ""
		m["bigDargon"] = -0.1
	}

	if maxBigDargonTeam.TeamID > 0 {
		m["has"] = true
		m["teamName"] = maxBigDargonTeam.Name
		m["teamImage"] = maxBigDargonTeam.Image
		m["bigDargon"] = maxBigDargonTeam.AvgBigDargon
	}

	return m
}

func loadMaxKillsTeam() map[string]interface{} {
	m := make(map[string]interface{}, 0)
	{
		m["has"] = false
		m["teamName"] = ""
		m["teamImage"] = ""
		m["avgKills"] = -0.1
	}

	if maxKillsTeam.TeamID > 0 {
		m["has"] = true
		m["teamName"] = maxKillsTeam.Name
		m["teamImage"] = maxKillsTeam.Image
		m["avgKills"] = maxKillsTeam.AvgKills
	}

	return m
}

func loadHighWinTeam() map[string]interface{} {
	m := make(map[string]interface{}, 0)
	{
		m["has"] = false
		m["teamName"] = ""
		m["teamImage"] = ""
		m["winRate"] = -0.1
	}

	if highWinTeam.TeamID > 0 {
		m["has"] = true
		m["teamName"] = highWinTeam.Name
		m["teamImage"] = highWinTeam.Image
		m["winRate"] = highWinTeam.WinRate
	}

	return m
}

func loadMinDeathsPlayer() map[string]interface{} {
	m := make(map[string]interface{}, 0)
	{
		m["has"] = false
		m["teamName"] = ""
		m["teamImage"] = ""
		m["playerName"] = ""
		m["playerImage"] = ""
		m["avgDeaths"] = -0.1
	}

	if minDeathsPlayer.PlayerID > 0 {
		m["has"] = true
		m["teamName"] = minDeathsPlayer.TeamName
		m["teamImage"] = minDeathsPlayer.TeamImage
		m["playerName"] = minDeathsPlayer.Name
		m["playerImage"] = minDeathsPlayer.Image
		m["avgDeaths"] = minDeathsPlayer.AvgDeaths
	}

	return m
}

func loadMaxKillsPlayer() map[string]interface{} {
	m := make(map[string]interface{}, 0)
	{
		m["has"] = false
		m["teamName"] = ""
		m["teamImage"] = ""
		m["playerName"] = ""
		m["playerImage"] = ""
		m["avgKills"] = -0.1
	}

	if maxKillsPlayer.PlayerID > 0 {
		m["has"] = true
		m["teamName"] = maxKillsPlayer.TeamName
		m["teamImage"] = maxKillsPlayer.TeamImage
		m["playerName"] = maxKillsPlayer.Name
		m["playerImage"] = maxKillsPlayer.Image
		m["avgKills"] = maxKillsPlayer.AvgKills
	}

	return m
}

func loadMaxAssistsPlayer() map[string]interface{} {
	m := make(map[string]interface{}, 0)
	{
		m["has"] = false
		m["teamName"] = ""
		m["teamImage"] = ""
		m["playerName"] = ""
		m["playerImage"] = ""
		m["avgAssists"] = -0.1
	}

	if maxAssistsPlayer.PlayerID > 0 {
		m["has"] = true
		m["teamName"] = maxAssistsPlayer.TeamName
		m["teamImage"] = maxAssistsPlayer.TeamImage
		m["playerName"] = maxAssistsPlayer.Name
		m["playerImage"] = maxAssistsPlayer.Image
		m["avgAssists"] = maxAssistsPlayer.AvgAssists
	}

	return m
}

func loadHighPickedHero() map[string]interface{} {
	m := make(map[string]interface{}, 0)
	{
		m["has"] = false
		m["name"] = ""
		m["image"] = ""
		m["tip"] = "PICK率最多"
		m["rate"] = -0.1
	}

	if highPickedHero.HeroID > 0 {
		m["has"] = true
		m["name"] = highPickedHero.Name
		m["image"] = highPickedHero.Image
		m["tip"] = "PICK率最多"
		m["rate"] = highPickedHero.PickedRate
	}

	return m
}

func loadHighBannedHero() map[string]interface{} {
	m := make(map[string]interface{}, 0)
	{
		m["has"] = false
		m["name"] = ""
		m["image"] = ""
		m["tip"] = "BAN率最多"
		m["rate"] = -0.1
	}

	if highBanedHero.HeroID > 0 {
		m["has"] = true
		m["name"] = highBanedHero.Name
		m["image"] = highBanedHero.Image
		m["tip"] = "BAN率最多"
		m["rate"] = highBanedHero.BanedRate
	}

	return m
}

func loadHighWinHero() map[string]interface{} {
	m := make(map[string]interface{}, 0)
	{
		m["has"] = false
		m["name"] = ""
		m["image"] = ""
		m["tip"] = "胜率最高"
		m["rate"] = -0.1
	}

	if highWinHero.HeroID > 0 {
		m["has"] = true
		m["name"] = highWinHero.Name
		m["image"] = highWinHero.Image
		m["tip"] = "胜率最高"
		m["rate"] = highWinHero.WinRate
	}

	return m
}

func watchS10ScoreAnalysisBySeasonWatch(ctx context.Context) {
	d := conf.LoadSeasonContestWatch()
	if d == nil || !d.CanWatch() {
		return
	}

	teamAnalysis := make([]*model.ScoreTeamAnalysis, 0)
	mcScanErr := component.GlobalMemcached.Get(ctx, d.CacheKey4TeamAnalysis).Scan(&teamAnalysis)
	if mcScanErr == nil && len(teamAnalysis) > 0 {
		tmpList := make([]*model.ScoreTeamAnalysis4Live, 0)
		for _, v := range teamAnalysis {
			tmpList = append(tmpList, v.Convert2Live())
		}
		scoreAnalysis4Team4Live = tmpList

		sort.SliceStable(teamAnalysis, func(i, j int) bool {
			if teamAnalysis[i].WinRate == teamAnalysis[j].WinRate {
				if teamAnalysis[i].TotalRound == teamAnalysis[j].TotalRound {
					return teamAnalysis[i].TeamID < teamAnalysis[j].TeamID
				}

				return teamAnalysis[i].TotalRound > teamAnalysis[j].TotalRound
			}

			return teamAnalysis[i].WinRate > teamAnalysis[j].WinRate
		})
		highWinTeam = teamAnalysis[0]

		sort.SliceStable(teamAnalysis, func(i, j int) bool {
			if teamAnalysis[i].AvgKills == teamAnalysis[j].AvgKills {
				if teamAnalysis[i].TotalRound == teamAnalysis[j].TotalRound {
					return teamAnalysis[i].TeamID < teamAnalysis[j].TeamID
				}

				return teamAnalysis[i].TotalRound > teamAnalysis[j].TotalRound
			}

			return teamAnalysis[i].AvgKills > teamAnalysis[j].AvgKills
		})
		maxKillsTeam = teamAnalysis[0]

		sort.SliceStable(teamAnalysis, func(i, j int) bool {
			if teamAnalysis[i].AvgBigDargon == teamAnalysis[j].AvgBigDargon {
				if teamAnalysis[i].TotalRound == teamAnalysis[j].TotalRound {
					return teamAnalysis[i].TeamID < teamAnalysis[j].TeamID
				}

				return teamAnalysis[i].TotalRound > teamAnalysis[j].TotalRound
			}

			return teamAnalysis[i].AvgBigDargon > teamAnalysis[j].AvgBigDargon
		})
		maxBigDargonTeam = teamAnalysis[0]
	}

	playerAnalysis := make([]*model.ScorePlayerAnalysis, 0)
	mcScanErr = component.GlobalMemcached.Get(ctx, d.CacheKey4PlayerAnalysis).Scan(&playerAnalysis)
	if mcScanErr == nil && len(playerAnalysis) > 0 {
		tmpList := make([]*model.ScorePlayerAnalysis4Live, 0)
		for _, v := range playerAnalysis {
			tmpList = append(tmpList, v.Convert2Live())
		}
		scoreAnalysis4Player4Live = tmpList

		sort.SliceStable(playerAnalysis, func(i, j int) bool {
			if playerAnalysis[i].AvgKills == playerAnalysis[j].AvgKills {
				if playerAnalysis[i].Played == playerAnalysis[j].Played {
					return playerAnalysis[i].PlayerID < playerAnalysis[j].PlayerID
				}

				return playerAnalysis[i].Played > playerAnalysis[j].Played
			}

			return playerAnalysis[i].AvgKills > playerAnalysis[j].AvgKills
		})
		maxKillsPlayer = playerAnalysis[0]

		sort.SliceStable(playerAnalysis, func(i, j int) bool {
			if playerAnalysis[i].AvgAssists == playerAnalysis[j].AvgAssists {
				if playerAnalysis[i].Played == playerAnalysis[j].Played {
					return playerAnalysis[i].PlayerID < playerAnalysis[j].PlayerID
				}

				return playerAnalysis[i].Played > playerAnalysis[j].Played
			}

			return playerAnalysis[i].AvgAssists > playerAnalysis[j].AvgAssists
		})
		maxAssistsPlayer = playerAnalysis[0]

		sort.SliceStable(playerAnalysis, func(i, j int) bool {
			if playerAnalysis[i].AvgDeaths == playerAnalysis[j].AvgDeaths {
				if playerAnalysis[i].Played == playerAnalysis[j].Played {
					return playerAnalysis[i].PlayerID < playerAnalysis[j].PlayerID
				}

				return playerAnalysis[i].Played > playerAnalysis[j].Played
			}

			return playerAnalysis[i].AvgDeaths < playerAnalysis[j].AvgDeaths
		})
		minDeathsPlayer = playerAnalysis[0]
	}

	heroAnalysis := make([]*model.ScoreHeroAnalysis, 0)
	mcScanErr = component.GlobalMemcached.Get(ctx, d.CacheKey4HeroAnalysis).Scan(&heroAnalysis)
	if mcScanErr == nil && len(heroAnalysis) > 0 {
		tmpList := make([]*model.ScoreHeroAnalysis4Live, 0)
		for _, v := range heroAnalysis {
			tmpList = append(tmpList, v.Convert2Live())
		}
		scoreAnalysis4Hero4Live = tmpList

		sort.SliceStable(heroAnalysis, func(i, j int) bool {
			if heroAnalysis[i].PickedRate == heroAnalysis[j].PickedRate {
				if heroAnalysis[i].Picked == heroAnalysis[j].Picked {
					return heroAnalysis[i].HeroID < heroAnalysis[j].HeroID
				}

				return heroAnalysis[i].Picked > heroAnalysis[j].Picked
			}

			return heroAnalysis[i].PickedRate > heroAnalysis[j].PickedRate
		})
		highPickedHero = heroAnalysis[0]

		sort.SliceStable(heroAnalysis, func(i, j int) bool {
			if heroAnalysis[i].BanedRate == heroAnalysis[j].BanedRate {
				if heroAnalysis[i].Picked == heroAnalysis[j].Picked {
					return heroAnalysis[i].HeroID < heroAnalysis[j].HeroID
				}

				return heroAnalysis[i].Picked > heroAnalysis[j].Picked
			}

			return heroAnalysis[i].BanedRate > heroAnalysis[j].BanedRate
		})
		highBanedHero = heroAnalysis[0]

		sort.SliceStable(heroAnalysis, func(i, j int) bool {
			if heroAnalysis[i].WinRate == heroAnalysis[j].WinRate {
				if heroAnalysis[i].Picked == heroAnalysis[j].Picked {
					return heroAnalysis[i].HeroID < heroAnalysis[j].HeroID
				}

				return heroAnalysis[i].Picked > heroAnalysis[j].Picked
			}

			return heroAnalysis[i].WinRate > heroAnalysis[j].WinRate
		})
		highWinHero = heroAnalysis[0]
	}
}

func (s *Service) S10ScoreAnalysis(ctx context.Context, req *model.ScoreAnalysisRequest) map[string]interface{} {
	res := make(map[string]interface{}, 0)
	switch req.AnalysisType {
	case analysisType4Team:
		res = genTeamAnalysis(req)
	case analysisType4Player:
		res = genPlayerAnalysis(req)
	case analysisType4Hero:
		res = genHeroAnalysis(req)
	}

	return res
}

func genTeamAnalysis(req *model.ScoreAnalysisRequest) map[string]interface{} {
	m := make(map[string]interface{}, 0)
	{
		m["highWinRate"] = loadHighWinTeam()
		m["maxKills"] = loadMaxKillsTeam()
		m["maxBigDargon"] = loadMaxBigDargonTeam()
		m["list"] = scoreAnalysis4Team4Live
	}

	return m
}

func genPlayerAnalysis(req *model.ScoreAnalysisRequest) map[string]interface{} {
	m := make(map[string]interface{}, 0)
	{
		m["maxAssists"] = loadMaxAssistsPlayer()
		m["maxKills"] = loadMaxKillsPlayer()
		m["minDeaths"] = loadMinDeathsPlayer()
		m["list"] = scoreAnalysis4Player4Live
	}

	return m
}

func genHeroAnalysis(req *model.ScoreAnalysisRequest) map[string]interface{} {
	m := make(map[string]interface{}, 0)
	{
		m["highPick"] = loadHighPickedHero()
		m["highBan"] = loadHighBannedHero()
		m["highWinRate"] = loadHighWinHero()
		m["list"] = scoreAnalysis4Hero4Live
	}

	return m
}
