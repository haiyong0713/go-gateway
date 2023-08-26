package model

import (
	"fmt"
	"strconv"
)

type ScoreTeamAnalysis4Live struct {
	TeamID             int64  `json:"teamID"`
	Name               string `json:"name"`
	Image              string `json:"image"`
	TotalRound         string `json:"totalRound"`
	AvgKills           string `json:"avgKills"`
	AvgAssists         string `json:"avgAssists"`
	AvgDeaths          string `json:"avgDeaths"`
	AvgSmallDargon     string `json:"avgSmallDargon"`
	AvgTowersDestroyed string `json:"avgTowersDestroyed"`
	AvgBigDargon       string `json:"avgBigDargon"`
	WinRate            string `json:"winRate"`
}

type ScorePlayerAnalysis4Live struct {
	PlayerID        int64  `json:"playerID"`
	Name            string `json:"name"`
	EnName          string `json:"enName"`
	Image           string `json:"image"`
	TeamName        string `json:"teamName"`
	TeamImage       string `json:"teamImage"`
	Position        string `json:"position"`
	KDA             string `json:"kda"`
	Played          string `json:"played"`
	ParticipateRate string `json:"participateRate"`
	AvgKills        string `json:"avgKills"`
	AvgAssists      string `json:"avgAssists"`
	AvgDeaths       string `json:"avgDeaths"`
	AvgEconomyMin   string `json:"avgEconomyMin"`
	DamageRate      string `json:"damageRate"`
	DamageTakenRate string `json:"damageTakenRate"`
	CountryImage    string `json:"countryImage"`
}

type ScoreHeroAnalysis4Live struct {
	HeroID     int64  `json:"heroID"`
	Name       string `json:"name"`
	EnName     string `json:"enName"`
	TwName     string `json:"twName"`
	Image      string `json:"image"`
	Position   string `json:"position"`
	Picked     string `json:"picked"`
	PickedRate string `json:"pickedRate"`
	Baned      string `json:"baned"`
	BanedRate  string `json:"banedRate"`
	Win        string `json:"win"`
	WinRate    string `json:"winRate"`
	TotalRound string `json:"totalRound"`
}

type ScoreTeamAnalysis struct {
	TournamentID         int64   `json:"-"`
	TeamID               int64   `json:"teamID"`
	Name                 string  `json:"name"`
	Image                string  `json:"image"`
	KDA                  float64 `json:"-"`
	TotalRound           int64   `json:"totalRound"`
	AvgRoundDuration     string  `json:"-"`
	FirstBloodRate       float64 `json:"-"`
	AvgKills             float64 `json:"avgKills"`
	AvgAssists           float64 `json:"avgAssists"`
	AvgDeaths            float64 `json:"avgDeaths"`
	AvgDamage            float64 `json:"-"`
	AvgDamageMin         float64 `json:"-"`
	AvgHitsMin           float64 `json:"-"`
	AvgEconomy           float64 `json:"-"`
	AvgEconomyMin        float64 `json:"-"`
	AvgSmallDargon       float64 `json:"avgSmallDargon"`
	SmallDargonRate      float64 `json:"-"`
	AvgWardsPlacedMin    float64 `json:"-"`
	AvgWardsKilledMin    float64 `json:"-"`
	AvgTowersDestroyed   float64 `json:"avgTowersDestroyed"`
	AvgTowersBeDestroyed float64 `json:"-"`
	AvgBigDargon         float64 `json:"avgBigDargon"`
	BigDargonRate        float64 `json:"-"`
	UpdatedAt            int64   `json:"-"`
	Win                  int64   `json:"-"`
	Lose                 int64   `json:"-"`
	WinRate              float64 `json:"winRate"`
	SmallRounds          int64   `json:"-"`
	Score                float64 `json:"-"`
	TotalKills           int64   `json:"-"`
	TotalDeaths          int64   `json:"-"`
	TotalSmallDargon     int64   `json:"-"`
	TotalBigDargon       int64   `json:"-"`
	TotalAssists         int64   `json:"-"`
}

type ScorePlayerAnalysis struct {
	TournamentID      int64   `json:"-"`
	PlayerID          int64   `json:"playerID"`
	Name              string  `json:"name"`
	EnName            string  `json:"enName"`
	Image             string  `json:"image"`
	TeamID            int64   `json:"-"`
	TeamName          string  `json:"teamName"`
	TeamImage         string  `json:"teamImage"`
	Position          string  `json:"position"`
	PositionId        int64   `json:"-"`
	KDA               float64 `json:"kda"`
	Played            int64   `json:"played"`
	ParticipateRate   float64 `json:"participateRate"`
	AvgKills          float64 `json:"avgKills"`
	AvgAssists        float64 `json:"avgAssists"`
	AvgDeaths         float64 `json:"avgDeaths"`
	AvgEconomyMin     float64 `json:"avgEconomyMin"`
	AvgHitsMin        float64 `json:"-"`
	DamageRate        float64 `json:"damageRate"`
	DamageTakenRate   float64 `json:"damageTakenRate"`
	AvgDamageMin      float64 `json:"-"`
	AvgDamageTakenMin float64 `json:"-"`
	AvgWardsPlacedMin float64 `json:"-"`
	AvgWardsKilledMin float64 `json:"-"`
	UpdatedAt         int64   `json:"-"`
	Mvp               int64   `json:"-"`
	Win               int64   `json:"-"`
	Lose              int64   `json:"-"`
	WinRate           float64 `json:"-"`
	CountryID         int64   `json:"-"`
	CountryImage      string  `json:"countryImage"`
	Score             float64 `json:"-"`
	TotalKills        int64   `json:"-"`
	TotalDeaths       int64   `json:"-"`
	TotalAssists      int64   `json:"-"`
}

type ScoreHeroAnalysis struct {
	TournamentID int64   `json:"-"`
	HeroID       int64   `json:"heroID"`
	Name         string  `json:"name"`
	EnName       string  `json:"enName"`
	TwName       string  `json:"twName"`
	Image        string  `json:"image"`
	Position     string  `json:"position"`
	PositionId   int64   `json:"-"`
	KDA          float64 `json:"-"`
	Picked       int64   `json:"picked"`
	PickedRate   float64 `json:"pickedRate"`
	Baned        int64   `json:"baned"`
	BanedRate    float64 `json:"banedRate"`
	Win          int64   `json:"win"`
	WinRate      float64 `json:"winRate"`
	TotalRound   int64   `json:"totalRound"`
	UpdatedAt    int64   `json:"-"`
	Score        float64 `json:"-"`
	AvgKills     float64 `json:"-"`
	AvgAssists   float64 `json:"-"`
	AvgDeaths    float64 `json:"-"`
	GameVersion  string  `json:"-"`
}

func (team *ScoreTeamAnalysis) Convert2Live() *ScoreTeamAnalysis4Live {
	tmpTeam := new(ScoreTeamAnalysis4Live)
	{
		tmpTeam.TeamID = team.TeamID
		tmpTeam.Name = team.Name
		tmpTeam.Image = team.Image
		tmpTeam.TotalRound = strconv.FormatInt(team.TotalRound, 10)
		tmpTeam.AvgKills = fmt.Sprintf("%.1f", team.AvgKills)
		tmpTeam.AvgAssists = fmt.Sprintf("%.1f", team.AvgAssists)
		tmpTeam.AvgDeaths = fmt.Sprintf("%.1f", team.AvgDeaths)
		tmpTeam.AvgSmallDargon = fmt.Sprintf("%.1f", team.AvgSmallDargon)
		tmpTeam.AvgTowersDestroyed = fmt.Sprintf("%.1f", team.AvgTowersDestroyed)
		tmpTeam.AvgBigDargon = fmt.Sprintf("%.1f", team.AvgBigDargon)
		tmpTeam.WinRate = fmt.Sprintf("%.1f", team.WinRate)
	}

	return tmpTeam
}

func (player *ScorePlayerAnalysis) Convert2Live() *ScorePlayerAnalysis4Live {
	tmpPlayer := new(ScorePlayerAnalysis4Live)
	{
		tmpPlayer.PlayerID = player.PlayerID
		tmpPlayer.Name = player.Name
		tmpPlayer.EnName = player.EnName
		tmpPlayer.Image = player.Image
		tmpPlayer.TeamName = player.TeamName
		tmpPlayer.TeamImage = player.TeamImage
		tmpPlayer.Position = player.Position
		tmpPlayer.KDA = fmt.Sprintf("%.1f", player.KDA)
		tmpPlayer.Played = strconv.FormatInt(player.Played, 10)
		tmpPlayer.ParticipateRate = fmt.Sprintf("%.1f", player.ParticipateRate)
		tmpPlayer.AvgKills = fmt.Sprintf("%.1f", player.AvgKills)
		tmpPlayer.AvgAssists = fmt.Sprintf("%.1f", player.AvgAssists)
		tmpPlayer.AvgDeaths = fmt.Sprintf("%.1f", player.AvgDeaths)
		tmpPlayer.AvgEconomyMin = fmt.Sprintf("%.1f", player.AvgEconomyMin)
		tmpPlayer.DamageRate = fmt.Sprintf("%.1f", player.DamageRate)
		tmpPlayer.DamageTakenRate = fmt.Sprintf("%.1f", player.DamageTakenRate)
		tmpPlayer.CountryImage = player.CountryImage
	}

	return tmpPlayer
}

func (hero *ScoreHeroAnalysis) Convert2Live() *ScoreHeroAnalysis4Live {
	tmpHero := new(ScoreHeroAnalysis4Live)
	{
		tmpHero.HeroID = hero.HeroID
		tmpHero.Name = hero.Name
		tmpHero.EnName = hero.EnName
		tmpHero.TwName = hero.TwName
		tmpHero.Image = hero.Image
		tmpHero.Position = hero.Position
		tmpHero.Picked = strconv.FormatInt(hero.Picked, 10)
		tmpHero.PickedRate = fmt.Sprintf("%.1f", hero.PickedRate)
		tmpHero.Baned = strconv.FormatInt(hero.Baned, 10)
		tmpHero.BanedRate = fmt.Sprintf("%.1f", hero.BanedRate)
		tmpHero.Win = strconv.FormatInt(hero.Win, 10)
		tmpHero.WinRate = fmt.Sprintf("%.1f", hero.WinRate)
		tmpHero.TotalRound = strconv.FormatInt(hero.TotalRound, 10)
	}

	return tmpHero
}
