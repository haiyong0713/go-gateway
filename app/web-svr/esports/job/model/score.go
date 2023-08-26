package model

import (
	"context"
	"fmt"
	"strconv"

	"go-common/library/log"
	"go-gateway/app/web-svr/esports/job/component"
	"go-gateway/app/web-svr/esports/job/sql"
)

// ParamScore score params.
type ParamScore struct {
	MatchID      int64  `json:"match_id"`
	SerieID      int64  `json:"serie_id"`
	OriginID     int64  `json:"origin_id"`
	BattleString string `json:"battle_string"`
	Pn           string `json:"pn"`
	Ps           int64  `json:"ps"`
}

type ScoreOriginTeamAnalysis struct {
	TournamentID         string `json:"tournament_id"`
	TeamID               string `json:"team_id"`
	Name                 string `json:"team_name"`
	Image                string `json:"team_image"`
	KDA                  string `json:"KDA"`
	TotalRound           string `json:"MACTH_TIMES"`
	AvgRoundDuration     string `json:"AVERAGE_TIME"`
	FirstBloodRate       string `json:"FIRSTBLOODKILL"`
	AvgKills             string `json:"AVERAGE_KILLS"`
	AvgAssists           string `json:"AVERAGE_ASSISTS"`
	AvgDeaths            string `json:"AVERAGE_DEATHS"`
	AvgDamage            string `json:"AVERAGE_CHAMPIONS"`
	AvgDamageMin         string `json:"MINUTE_OUTPUT"`
	AvgHitsMin           string `json:"MINUTE_HITS"`
	AvgEconomy           string `json:"AVERAGE_MONEY"`
	AvgEconomyMin        string `json:"MINUTE_MONEY"`
	AvgSmallDargon       string `json:"AVERAGE_SMALLDRAGON"`
	SmallDargonRate      string `json:"SMALLDRAGON_RATE"`
	AvgWardsPlacedMin    string `json:"MINUTE_WARDSPLACED"`
	AvgWardsKilledMin    string `json:"MINUTE_WARDSKILLED"`
	AvgTowersDestroyed   string `json:"AVERAGE_TOWER_SUCCESS"`
	AvgTowersBeDestroyed string `json:"AVERAGE_TOWER_FAIL"`
	AvgBigDargon         string `json:"AVERAGE_BIGDRAGON"`
	BigDargonRate        string `json:"BIGDRAGON_RATE"`
	UpdatedAt            string `json:"update_time"`
	Win                  string `json:"win"`
	Lose                 string `json:"los"`
	WinRate              string `json:"VICTORY_RATE"`
	SmallRounds          string `json:"RESULT_TIMES"`
	Score                string `json:"f_score"`
	TotalKills           string `json:"total_kills"`
	TotalDeaths          string `json:"total_deaths"`
	TotalSmallDargon     string `json:"total_SMALLDRAGON"`
	TotalBigDargon       string `json:"total_BIGDRAGON"`
	TotalAssists         string `json:"total_assists"`
}

type ScoreOriginPlayerAnalysis struct {
	TournamentID      string `json:"tournament_id"`
	PlayerID          string `json:"player_id"`
	Name              string `json:"player_name"`
	Image             string `json:"player_image"`
	TeamID            string `json:"team_id"`
	TeamName          string `json:"team_name"`
	TeamImage         string `json:"team_image"`
	Position          string `json:"position"`
	PositionId        string `json:"position_id"`
	KDA               string `json:"KDA"`
	Played            string `json:"PLAYS_TIMES"`
	ParticipateRate   string `json:"OFFERED_RATE"`
	AvgKills          string `json:"AVERAGE_KILLS"`
	AvgAssists        string `json:"AVERAGE_ASSISTS"`
	AvgDeaths         string `json:"AVERAGE_DEATHS"`
	AvgEconomyMin     string `json:"MINUTE_ECONOMIC"`
	AvgHitsMin        string `json:"MINUTE_HITS"`
	AvgDamageMin      string `json:"MINUTE_DAMAGEDEALT"`
	AvgDamageTakenMin string `json:"MINUTE_DAMAGETAKEN"`
	AvgWardsPlacedMin string `json:"MINUTE_WARDSPLACED"`
	AvgWardsKilledMin string `json:"MINUTE_WARDKILLED"`
	DamageRate        string `json:"DAMAGEDEALT_RATE"`
	DamageTakenRate   string `json:"DAMAGETAKEN_RATE"`
	UpdatedAt         string `json:"update_time"`
	Mvp               string `json:"mvp"`
	EnName            string `json:"player_chinese_name"`
	Win               string `json:"win"`
	Lose              string `json:"lose"`
	WinRate           string `json:"VICTORY_RATE"`
	CountryID         string `json:"country_id"`
	CountryImage      string `json:"country_image"`
	Score             string `json:"f_score"`
	TotalKills        string `json:"total_kills"`
	TotalDeaths       string `json:"total_deaths"`
	TotalAssists      string `json:"total_assists"`
}

type ScoreOriginHeroAnalysis struct {
	TournamentID string `json:"tournament_id"`
	HeroID       string `json:"hero_id"`
	Name         string `json:"hero_name"`
	Image        string `json:"hero_image"`
	Position     string `json:"position_name"`
	PositionId   string `json:"position_id"`
	KDA          string `json:"KDA"`
	Picked       string `json:"appear_count"`
	Baned        string `json:"prohibit_count"`
	Win          string `json:"victory_count"`
	TotalRound   string `json:"game_count"`
	UpdatedAt    string `json:"update_time"`
	Score        string `json:"f_score"`
	AvgKills     string `json:"AVERAGE_KILLS"`
	AvgAssists   string `json:"AVERAGE_ASSISTS"`
	AvgDeaths    string `json:"AVERAGE_DEATHS"`
	GameVersion  string `json:"game_ver"`
	PickedRate   string `json:"APPEAR"`
	BanedRate    string `json:"PROHIBIT"`
	WinRate      string `json:"VICTORY_RATE"`
	EnName       string `json:"hero_name_en"`
	TwName       string `json:"hero_name_tw"`
}

type ScoreTeamAnalysis struct {
	TournamentID         int64   `json:"tournamentID"`
	TeamID               int64   `json:"teamID"`
	Name                 string  `json:"name"`
	Image                string  `json:"image"`
	KDA                  float64 `json:"kda"`
	TotalRound           int64   `json:"totalRound"`
	AvgRoundDuration     string  `json:"avgRoundDuration"`
	FirstBloodRate       float64 `json:"firstBloodRate"`
	AvgKills             float64 `json:"avgKills"`
	AvgAssists           float64 `json:"avgAssists"`
	AvgDeaths            float64 `json:"avgDeaths"`
	AvgDamage            float64 `json:"avgDamage"`
	AvgDamageMin         float64 `json:"avgDamageMin"`
	AvgHitsMin           float64 `json:"avgHitsMin"`
	AvgEconomy           float64 `json:"avgEconomy"`
	AvgEconomyMin        float64 `json:"avgEconomyMin"`
	AvgSmallDargon       float64 `json:"avgSmallDargon"`
	SmallDargonRate      float64 `json:"smallDargonRate"`
	AvgWardsPlacedMin    float64 `json:"avgWardsPlacedMin"`
	AvgWardsKilledMin    float64 `json:"avgWardsKilledMin"`
	AvgTowersDestroyed   float64 `json:"avgTowersDestroyed"`
	AvgTowersBeDestroyed float64 `json:"avgTowersBeDestroyed"`
	AvgBigDargon         float64 `json:"avgBigDargon"`
	BigDargonRate        float64 `json:"bigDargonRate"`
	UpdatedAt            int64   `json:"updatedAt"`
	Win                  int64   `json:"win"`
	Lose                 int64   `json:"lose"`
	WinRate              float64 `json:"winRate"`
	SmallRounds          int64   `json:"smallRounds"`
	Score                float64 `json:"score"`
	TotalKills           int64   `json:"totalKills"`
	TotalDeaths          int64   `json:"totalDeaths"`
	TotalSmallDargon     int64   `json:"totalSmallDargon"`
	TotalBigDargon       int64   `json:"totalBigDargon"`
	TotalAssists         int64   `json:"totalAssists"`
}

type ScorePlayerAnalysis struct {
	TournamentID      int64   `json:"tournamentID"`
	PlayerID          int64   `json:"playerID"`
	Name              string  `json:"name"`
	EnName            string  `json:"enName"`
	Image             string  `json:"image"`
	TeamID            int64   `json:"team_ID"`
	TeamName          string  `json:"teamName"`
	TeamImage         string  `json:"teamImage"`
	Position          string  `json:"position"`
	PositionId        int64   `json:"positionID"`
	KDA               float64 `json:"kda"`
	Played            int64   `json:"played"`
	ParticipateRate   float64 `json:"participateRate"`
	AvgKills          float64 `json:"avgKills"`
	AvgAssists        float64 `json:"avgAssists"`
	AvgDeaths         float64 `json:"avgDeaths"`
	AvgEconomyMin     float64 `json:"avgEconomyMin"`
	AvgHitsMin        float64 `json:"avgHitsMin"`
	DamageRate        float64 `json:"damageRate"`
	DamageTakenRate   float64 `json:"damageTakenRate"`
	AvgDamageMin      float64 `json:"avgDamageMin"`
	AvgDamageTakenMin float64 `json:"avgDamageTakenMin"`
	AvgWardsPlacedMin float64 `json:"avgWardsPlacedMin"`
	AvgWardsKilledMin float64 `json:"avgWardsKilledMin"`
	UpdatedAt         int64   `json:"updatedAt"`
	Mvp               int64   `json:"mvp"`
	Win               int64   `json:"win"`
	Lose              int64   `json:"lose"`
	WinRate           float64 `json:"winRate"`
	CountryID         int64   `json:"countryID"`
	CountryImage      string  `json:"countryImage"`
	Score             float64 `json:"score"`
	TotalKills        int64   `json:"totalKills"`
	TotalDeaths       int64   `json:"totalDeaths"`
	TotalAssists      int64   `json:"totalAssists"`
}

type ScoreHeroAnalysis struct {
	TournamentID int64   `json:"tournamentID"`
	HeroID       int64   `json:"heroID"`
	Name         string  `json:"name"`
	EnName       string  `json:"enName"`
	TwName       string  `json:"twName"`
	Image        string  `json:"image"`
	Position     string  `json:"position"`
	PositionId   int64   `json:"positionID"`
	KDA          float64 `json:"kda"`
	Picked       int64   `json:"picked"`
	PickedRate   float64 `json:"pickedRate"`
	Baned        int64   `json:"baned"`
	BanedRate    float64 `json:"banedRate"`
	Win          int64   `json:"win"`
	WinRate      float64 `json:"winRate"`
	TotalRound   int64   `json:"totalRound"`
	UpdatedAt    int64   `json:"updatedAt"`
	Score        float64 `json:"score"`
	AvgKills     float64 `json:"avgKills"`
	AvgAssists   float64 `json:"avgAssists"`
	AvgDeaths    float64 `json:"avgDeaths"`
	GameVersion  string  `json:"gameVersion"`
}

const (
	sqlOfInsertUpdate4ScoreTeamAnalysis = `
INSERT INTO score_analysis_lol_team (tournament_id, team_id, team_name, image, kda
	, total_round, avg_round_duration, first_blood_rate, avg_kills, avg_assists
	, avg_deaths, avg_damage, avg_damage_min, avg_hits_min, avg_economy
	, avg_economy_min, avg_small_dargon, small_dargon_rate, avg_wards_placed_min, avg_wards_killed_min
	, avg_towers_destroyed, avg_towers_be_destroyed, avg_big_dargon, big_dargon_rate, updated_at
	, win, lose, win_rate, small_rounds, score
	, total_kills, total_deaths, total_small_dargon, total_big_dargon, total_assists)
VALUES (?, ?, ?, ?, ?
	, ?, ?, ?, ?, ?
	, ?, ?, ?, ?, ?
	, ?, ?, ?, ?, ?
	, ?, ?, ?, ?, ?
	, ?, ?, ?, ?, ?
	, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE tournament_id = ?, team_id = ?, team_name = ?, image = ?, kda = ?
, total_round = ?, avg_round_duration = ?, first_blood_rate = ?, avg_kills = ?, avg_assists = ?
, avg_deaths = ?, avg_damage = ?, avg_damage_min = ?, avg_hits_min = ?, avg_economy = ?
, avg_economy_min = ?, avg_small_dargon = ?, small_dargon_rate = ?, avg_wards_placed_min = ?, avg_wards_killed_min = ?
, avg_towers_destroyed = ?, avg_towers_be_destroyed = ?, avg_big_dargon = ?, big_dargon_rate = ?, updated_at = ?
, win = ?, lose = ?, win_rate = ?, small_rounds = ?, score = ?
, total_kills = ?, total_deaths = ?, total_small_dargon = ?, total_big_dargon = ?, total_assists = ?
`
	sqlOfInsertUpdate4ScorePlayerAnalysis = `
INSERT INTO score_analysis_lol_player (tournament_id, player_id, player_name, player_en_name, image
	, team_id, team_name, team_image, position_name, position_id
	, kda, played, participate_rate, avg_kills, avg_assists
	, avg_deaths, avg_economy_min, avg_hits_min, avg_damage_min, avg_damage_taken_min
	, damage_rate, damage_taken_rate, avg_wards_placed_min, avg_wards_killed_min, updated_at
	, mvp, win, lose, win_rate, country_id
	, country_image, score, total_kills, total_deaths, total_assists)
VALUES (?, ?, ?, ?, ?
	, ?, ?, ?, ?, ?
	, ?, ?, ?, ?, ?
	, ?, ?, ?, ?, ?
	, ?, ?, ?, ?, ?
	, ?, ?, ?, ?, ?
	, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE tournament_id = ?, player_id = ?, player_name = ?, player_en_name = ?, image = ?
, team_id = ?, team_name = ?, team_image = ?, position_name = ?, position_id = ?
, kda = ?, played = ?, participate_rate = ?, avg_kills = ?, avg_assists = ?
, avg_deaths = ?, avg_economy_min = ?, avg_hits_min = ?, avg_damage_min = ?, avg_damage_taken_min = ?
, damage_rate = ?, damage_taken_rate = ?, avg_wards_placed_min = ?, avg_wards_killed_min = ?, updated_at = ?
, mvp = ?, win = ?, lose = ?, win_rate = ?, country_id = ?
, country_image = ?, score = ?, total_kills = ?, total_deaths = ?, total_assists = ?
`
	sqlOfInsertUpdate4ScoreHeroAnalysis = `
INSERT INTO score_analysis_lol_hero (tournament_id, hero_id, hero_name, hero_en_name, hero_tw_name
	, image, position_id, position_name, kda, picked
	, picked_rate, baned, baned_rate, win_rate, win
	, total_round, updated_at, score, avg_kills, avg_assists
	, avg_deaths, game_version)
VALUES (?, ?, ?, ?, ?
	, ?, ?, ?, ?, ?
	, ?, ?, ?, ?, ?
	, ?, ?, ?, ?, ?
	, ?, ?)
ON DUPLICATE KEY UPDATE tournament_id = ?, hero_id = ?, hero_name = ?, hero_en_name = ?, hero_tw_name = ?
, image = ?, position_id = ?, position_name = ?, kda = ?, picked = ?
, picked_rate = ?, baned = ?, baned_rate = ?, win_rate = ?, win = ?
, total_round = ?, updated_at = ?, score = ?, avg_kills = ?, avg_assists = ?
, avg_deaths = ?, game_version = ?
`
)

func (team *ScoreTeamAnalysis) InsertUpdate(ctx context.Context) (err error) {
	_, err = sql.GlobalDB.Exec(
		ctx,
		sqlOfInsertUpdate4ScoreTeamAnalysis,
		team.TournamentID, team.TeamID, team.Name, team.Image, team.KDA,
		team.TotalRound, team.AvgRoundDuration, team.FirstBloodRate, team.AvgKills, team.AvgAssists,
		team.AvgDeaths, team.AvgDamage, team.AvgDamageMin, team.AvgHitsMin, team.AvgEconomy,
		team.AvgEconomyMin, team.AvgSmallDargon, team.SmallDargonRate, team.AvgWardsPlacedMin, team.AvgWardsKilledMin,
		team.AvgTowersDestroyed, team.AvgTowersBeDestroyed, team.AvgBigDargon, team.BigDargonRate, team.UpdatedAt,
		team.Win, team.Lose, team.WinRate, team.SmallRounds, team.Score,
		team.TotalKills, team.TotalDeaths, team.TotalSmallDargon, team.TotalBigDargon, team.TotalAssists,
		team.TournamentID, team.TeamID, team.Name, team.Image, team.KDA,
		team.TotalRound, team.AvgRoundDuration, team.FirstBloodRate, team.AvgKills, team.AvgAssists,
		team.AvgDeaths, team.AvgDamage, team.AvgDamageMin, team.AvgHitsMin, team.AvgEconomy,
		team.AvgEconomyMin, team.AvgSmallDargon, team.SmallDargonRate, team.AvgWardsPlacedMin, team.AvgWardsKilledMin,
		team.AvgTowersDestroyed, team.AvgTowersBeDestroyed, team.AvgBigDargon, team.BigDargonRate, team.UpdatedAt,
		team.Win, team.Lose, team.WinRate, team.SmallRounds, team.Score,
		team.TotalKills, team.TotalDeaths, team.TotalSmallDargon, team.TotalBigDargon, team.TotalAssists)

	return
}

//func (origin *ScoreOriginTeamAnalysis) Convert2TeamAnalysis(ctx context.Context) (analysis *ScoreTeamAnalysis, err error) {
//	analysis = new(ScoreTeamAnalysis)
//	{
//		filename := genFileName4Team(origin.TournamentID, origin.TeamID)
//
//		analysis.TournamentID, err = strconv.ParseInt(origin.TournamentID, 10, 64)
//		if err!=nil{
//			log.Errorc(ctx,"Convert2TeamAnalysis() error(%+v)",err)
//		}
//		analysis.TeamID, err = strconv.ParseInt(origin.TeamID, 10, 64)
//		if err!=nil{
//			log.Errorc(ctx,"Convert2TeamAnalysis() error(%+v)",err)
//		}
//		analysis.Name = origin.Name
//		analysis.Image, err = component.UploadBFSImageResourceByUrl(ctx, origin.Image, filename, component.UploadType4ImageOfTeam)
//		if err!=nil{
//			log.Errorc(ctx,"Convert2TeamAnalysis() error(%+v)",err)
//		}
//		analysis.KDA, err = strconv.ParseFloat(origin.KDA,64)
//		if err!=nil{
//			log.Errorc(ctx,"Convert2TeamAnalysis() error(%+v)",err)
//		}
//		analysis.TotalRound, err = strconv.ParseInt(origin.TotalRound, 10, 64)
//		if err!=nil{
//			log.Errorc(ctx,"Convert2TeamAnalysis() error(%+v)",err)
//		}
//		analysis.AvgRoundDuration = origin.AvgRoundDuration
//		analysis.FirstBloodRate, err = strconv.ParseFloat(origin.FirstBloodRate,64)
//		if err!=nil{
//			log.Errorc(ctx,"Convert2TeamAnalysis() error(%+v)",err)
//		}
//		analysis.AvgKills, err = strconv.ParseFloat(origin.AvgKills,64)
//		if err!=nil{
//			log.Errorc(ctx,"Convert2TeamAnalysis() error(%+v)",err)
//		}
//		analysis.AvgAssists, err = strconv.ParseFloat(origin.AvgAssists,64)
//		if err!=nil{
//			log.Errorc(ctx,"Convert2TeamAnalysis() error(%+v)",err)
//		}
//		analysis.AvgDeaths, err = strconv.ParseFloat(origin.AvgDeaths,64)
//		if err!=nil{
//			log.Errorc(ctx,"Convert2TeamAnalysis() error(%+v)",err)
//		}
//		analysis.AvgDamage, err = strconv.ParseFloat(origin.AvgDamage,64)
//		if err!=nil{
//			log.Errorc(ctx,"Convert2TeamAnalysis() error(%+v)",err)
//		}
//		analysis.AvgDamageMin, err = strconv.ParseFloat(origin.AvgDamageMin,64)
//		if err!=nil{
//			log.Errorc(ctx,"Convert2TeamAnalysis() error(%+v)",err)
//		}
//		analysis.AvgHitsMin, err = strconv.ParseFloat(origin.AvgHitsMin,64)
//		if err!=nil{
//			log.Errorc(ctx,"Convert2TeamAnalysis() error(%+v)",err)
//		}
//		analysis.AvgEconomy, err = strconv.ParseFloat(origin.AvgEconomy,64)
//		if err!=nil{
//			log.Errorc(ctx,"Convert2TeamAnalysis() error(%+v)",err)
//		}
//		analysis.AvgEconomyMin, err = strconv.ParseFloat(origin.AvgEconomyMin,64)
//		if err!=nil{
//			log.Errorc(ctx,"Convert2TeamAnalysis() error(%+v)",err)
//		}
//		analysis.AvgSmallDargon, err = strconv.ParseFloat(origin.AvgSmallDargon,64)
//		if err!=nil{
//			log.Errorc(ctx,"Convert2TeamAnalysis() error(%+v)",err)
//		}
//		analysis.SmallDargonRate, err = strconv.ParseFloat(origin.SmallDargonRate,64)
//		if err!=nil{
//			log.Errorc(ctx,"Convert2TeamAnalysis() error(%+v)",err)
//		}
//		analysis.AvgWardsPlacedMin, err = strconv.ParseFloat(origin.AvgWardsPlacedMin,64)
//		if err!=nil{
//			log.Errorc(ctx,"Convert2TeamAnalysis() error(%+v)",err)
//		}
//		analysis.AvgWardsKilledMin, err = strconv.ParseFloat(origin.AvgWardsKilledMin,64)
//		if err!=nil{
//			log.Errorc(ctx,"Convert2TeamAnalysis() error(%+v)",err)
//		}
//		analysis.AvgTowersDestroyed, err = strconv.ParseFloat(origin.AvgTowersDestroyed,64)
//		if err!=nil{
//			log.Errorc(ctx,"Convert2TeamAnalysis() error(%+v)",err)
//		}
//		analysis.AvgTowersBeDestroyed, err = strconv.ParseFloat(origin.AvgTowersBeDestroyed,64)
//		if err!=nil{
//			log.Errorc(ctx,"Convert2TeamAnalysis() error(%+v)",err)
//		}
//		analysis.AvgBigDargon, err = strconv.ParseFloat(origin.AvgBigDargon,64)
//		if err!=nil{
//			log.Errorc(ctx,"Convert2TeamAnalysis() error(%+v)",err)
//		}
//		analysis.BigDargonRate, err = strconv.ParseFloat(origin.BigDargonRate,64)
//		if err!=nil{
//			log.Errorc(ctx,"Convert2TeamAnalysis() error(%+v)",err)
//		}
//		analysis.UpdatedAt, err = strconv.ParseInt(origin.UpdatedAt, 10, 64)
//		if err!=nil{
//			log.Errorc(ctx,"Convert2TeamAnalysis() error(%+v)",err)
//		}
//		analysis.Win, err = strconv.ParseInt(origin.Win, 10, 64)
//		if err!=nil{
//			log.Errorc(ctx,"Convert2TeamAnalysis() error(%+v)",err)
//		}
//		analysis.Lose, err = strconv.ParseInt(origin.Lose, 10, 64)
//		if err!=nil{
//			log.Errorc(ctx,"Convert2TeamAnalysis() error(%+v)",err)
//		}
//		analysis.WinRate, err = strconv.ParseFloat(origin.WinRate, 64)
//		if err!=nil{
//			log.Errorc(ctx,"Convert2TeamAnalysis() error(%+v)",err)
//		}
//		analysis.SmallRounds, err = strconv.ParseInt(origin.SmallRounds, 10, 64)
//		if err!=nil{
//			log.Errorc(ctx,"Convert2TeamAnalysis() error(%+v)",err)
//		}
//		analysis.Score, err = strconv.ParseFloat(origin.Score, 64)
//		if err!=nil{
//			log.Errorc(ctx,"Convert2TeamAnalysis() error(%+v)",err)
//		}
//		analysis.TotalKills, err = strconv.ParseInt(origin.TotalKills, 10, 64)
//		if err!=nil{
//			log.Errorc(ctx,"Convert2TeamAnalysis() error(%+v)",err)
//		}
//		analysis.TotalDeaths, err = strconv.ParseInt(origin.TotalDeaths, 10, 64)
//		if err!=nil{
//			log.Errorc(ctx,"Convert2TeamAnalysis() error(%+v)",err)
//		}
//		analysis.TotalSmallDargon, err = strconv.ParseInt(origin.TotalSmallDargon, 10, 64)
//		if err!=nil{
//			log.Errorc(ctx,"Convert2TeamAnalysis() error(%+v)",err)
//		}
//		analysis.TotalBigDargon, err = strconv.ParseInt(origin.TotalBigDargon, 10, 64)
//		if err!=nil{
//			log.Errorc(ctx,"Convert2TeamAnalysis() error(%+v)",err)
//		}
//		analysis.TotalAssists, err = strconv.ParseInt(origin.TotalAssists, 10, 64)
//		if err!=nil{
//			log.Errorc(ctx,"Convert2TeamAnalysis() error(%+v)",err)
//		}
//	}
//
//	return
//}

func genFileName4Team(tournamentID, teamID string) string {
	return fmt.Sprintf("0830_%v_%v.png", tournamentID, teamID)
}

//func (origin *ScoreOriginPlayerAnalysis) Convert2PlayerAnalysis(ctx context.Context) (analysis *ScorePlayerAnalysis, err error) {
//	analysis = new(ScorePlayerAnalysis)
//	{
//		filename := fmt.Sprintf("0830_%v_%v_%v.png", origin.TournamentID, origin.TeamID, origin.PlayerID)
//		filename4Team := genFileName4Team(origin.TournamentID, origin.TeamID)
//
//		analysis.TournamentID, err = strconv.ParseInt(origin.TournamentID, 10, 64)
//		if err!=nil{
//			log.Errorc(ctx,"Convert2PlayerAnalysis() error(%+v)",err)
//		}
//		analysis.PlayerID, err = strconv.ParseInt(origin.PlayerID, 10, 64)
//		if err!=nil{
//			log.Errorc(ctx,"Convert2PlayerAnalysis() error(%+v)",err)
//		}
//		analysis.Name = origin.Name
//		analysis.EnName = origin.EnName
//		analysis.Image, err = component.UploadBFSImageResourceByUrl(ctx, origin.Image, filename, component.UploadType4ImageOfPlayer)
//		if err!=nil{
//			log.Errorc(ctx,"Convert2PlayerAnalysis() error(%+v)",err)
//		}
//		analysis.TeamID, err = strconv.ParseInt(origin.TeamID, 10, 64)
//		if err!=nil{
//			log.Errorc(ctx,"Convert2PlayerAnalysis() error(%+v)",err)
//		}
//		analysis.TeamName = origin.TeamName
//		analysis.TeamImage, err = component.UploadBFSImageResourceByUrl(ctx, origin.TeamImage, filename4Team, component.UploadType4ImageOfTeam)
//		if err!=nil{
//			log.Errorc(ctx,"Convert2PlayerAnalysis() error(%+v)",err)
//		}
//		analysis.Position = origin.Position
//		analysis.PositionId, err = strconv.ParseInt(origin.PositionId, 10, 64)
//		if err!=nil{
//			log.Errorc(ctx,"Convert2PlayerAnalysis() error(%+v)",err)
//		}
//		analysis.KDA, err = strconv.ParseFloat(origin.KDA,64)
//		if err!=nil{
//			log.Errorc(ctx,"Convert2PlayerAnalysis() error(%+v)",err)
//		}
//		analysis.Played, err = strconv.ParseInt(origin.Played, 10, 64)
//		if err!=nil{
//			log.Errorc(ctx,"Convert2PlayerAnalysis() error(%+v)",err)
//		}
//		analysis.ParticipateRate, err = strconv.ParseFloat(origin.ParticipateRate,64)
//		if err!=nil{
//			log.Errorc(ctx,"Convert2PlayerAnalysis() error(%+v)",err)
//		}
//		analysis.AvgKills, err = strconv.ParseFloat(origin.AvgKills,64)
//		if err!=nil{
//			log.Errorc(ctx,"Convert2PlayerAnalysis() error(%+v)",err)
//		}
//		analysis.AvgAssists, err = strconv.ParseFloat(origin.AvgAssists,64)
//		if err!=nil{
//			log.Errorc(ctx,"Convert2PlayerAnalysis() error(%+v)",err)
//		}
//		analysis.AvgDeaths, err = strconv.ParseFloat(origin.AvgDeaths,64)
//		if err!=nil{
//			log.Errorc(ctx,"Convert2PlayerAnalysis() error(%+v)",err)
//		}
//		analysis.AvgEconomyMin, err = strconv.ParseFloat(origin.AvgEconomyMin,64)
//		if err!=nil{
//			log.Errorc(ctx,"Convert2PlayerAnalysis() error(%+v)",err)
//		}
//		analysis.AvgHitsMin, err = strconv.ParseFloat(origin.AvgHitsMin,64)
//		if err!=nil{
//			log.Errorc(ctx,"Convert2PlayerAnalysis() error(%+v)",err)
//		}
//		analysis.DamageRate, err = strconv.ParseFloat(origin.DamageRate,64)
//		if err!=nil{
//			log.Errorc(ctx,"Convert2PlayerAnalysis() error(%+v)",err)
//		}
//		analysis.DamageTakenRate, err = strconv.ParseFloat(origin.DamageTakenRate,64)
//		if err!=nil{
//			log.Errorc(ctx,"Convert2PlayerAnalysis() error(%+v)",err)
//		}
//		analysis.AvgDamageMin, err = strconv.ParseFloat(origin.AvgDamageMin,64)
//		if err!=nil{
//			log.Errorc(ctx,"Convert2PlayerAnalysis() error(%+v)",err)
//		}
//		analysis.AvgDamageTakenMin, err = strconv.ParseFloat(origin.AvgDamageTakenMin,64)
//		if err!=nil{
//			log.Errorc(ctx,"Convert2PlayerAnalysis() error(%+v)",err)
//		}
//		analysis.AvgWardsPlacedMin, err = strconv.ParseFloat(origin.AvgWardsPlacedMin,64)
//		if err!=nil{
//			log.Errorc(ctx,"Convert2PlayerAnalysis() error(%+v)",err)
//		}
//		analysis.AvgWardsKilledMin, err = strconv.ParseFloat(origin.AvgWardsKilledMin,64)
//		if err!=nil{
//			log.Errorc(ctx,"Convert2PlayerAnalysis() error(%+v)",err)
//		}
//		analysis.UpdatedAt, err = strconv.ParseInt(origin.UpdatedAt, 10, 64)
//		if err!=nil{
//			log.Errorc(ctx,"Convert2PlayerAnalysis() error(%+v)",err)
//		}
//		analysis.Mvp, err = strconv.ParseInt(origin.Mvp, 10, 64)
//		if err!=nil{
//			log.Errorc(ctx,"Convert2PlayerAnalysis() error(%+v)",err)
//		}
//		analysis.Win, err = strconv.ParseInt(origin.Win, 10, 64)
//		if err!=nil{
//			log.Errorc(ctx,"Convert2PlayerAnalysis() error(%+v)",err)
//		}
//		analysis.Lose, err = strconv.ParseInt(origin.Lose, 10, 64)
//		if err!=nil{
//			log.Errorc(ctx,"Convert2PlayerAnalysis() error(%+v)",err)
//		}
//		analysis.WinRate, err = strconv.ParseFloat(origin.WinRate, 64)
//		if err!=nil{
//			log.Errorc(ctx,"Convert2PlayerAnalysis() error(%+v)",err)
//		}
//		analysis.CountryID, err = strconv.ParseInt(origin.CountryID, 10, 64)
//		if err!=nil{
//			log.Errorc(ctx,"Convert2PlayerAnalysis() error(%+v)",err)
//		}
//		analysis.CountryImage, err = component.UploadBFSImageResourceByUrl(ctx, origin.CountryImage, origin.CountryID, component.UploadType4ImageOfCountry)
//		if err!=nil{
//			log.Errorc(ctx,"Convert2PlayerAnalysis() error(%+v)",err)
//		}
//		analysis.Score, err = strconv.ParseFloat(origin.Score, 64)
//		if err!=nil{
//			log.Errorc(ctx,"Convert2PlayerAnalysis() error(%+v)",err)
//		}
//		analysis.TotalKills, err = strconv.ParseInt(origin.TotalKills, 10, 64)
//		if err!=nil{
//			log.Errorc(ctx,"Convert2PlayerAnalysis() error(%+v)",err)
//		}
//		analysis.TotalDeaths, err = strconv.ParseInt(origin.TotalDeaths, 10, 64)
//		if err!=nil{
//			log.Errorc(ctx,"Convert2PlayerAnalysis() error(%+v)",err)
//		}
//		analysis.TotalAssists, err = strconv.ParseInt(origin.TotalAssists, 10, 64)
//		if err!=nil{
//			log.Errorc(ctx,"Convert2PlayerAnalysis() error(%+v)",err)
//		}
//	}
//
//	return
//}

func (origin *ScoreOriginHeroAnalysis) Convert2HeroAnalysis(ctx context.Context) (analysis *ScoreHeroAnalysis, err error) {
	analysis = new(ScoreHeroAnalysis)
	{
		filename := fmt.Sprintf("0830_%v_%v.png", origin.TournamentID, origin.HeroID)

		analysis.TournamentID, err = strconv.ParseInt(origin.TournamentID, 10, 64)
		if err != nil {
			log.Errorc(ctx, "Convert2HeroAnalysis() error(%+v)", err)
		}
		analysis.HeroID, err = strconv.ParseInt(origin.HeroID, 10, 64)
		if err != nil {
			log.Errorc(ctx, "Convert2HeroAnalysis() error(%+v)", err)
		}
		analysis.Name = origin.Name
		analysis.EnName = origin.EnName
		analysis.TwName = origin.TwName
		analysis.Image, err = component.UploadBFSImageResourceByUrl(ctx, origin.Image, filename, component.UploadType4ImageOfHero)
		if err != nil {
			log.Errorc(ctx, "Convert2HeroAnalysis() error(%+v)", err)
		}
		analysis.Position = origin.Position
		analysis.PositionId, err = strconv.ParseInt(origin.PositionId, 10, 64)
		if err != nil {
			log.Errorc(ctx, "Convert2HeroAnalysis() error(%+v)", err)
		}
		analysis.KDA, err = strconv.ParseFloat(origin.KDA, 64)
		if err != nil {
			log.Errorc(ctx, "Convert2HeroAnalysis() error(%+v)", err)
		}
		analysis.Picked, err = strconv.ParseInt(origin.Picked, 10, 64)
		if err != nil {
			log.Errorc(ctx, "Convert2HeroAnalysis() error(%+v)", err)
		}
		analysis.PickedRate, err = strconv.ParseFloat(origin.PickedRate, 64)
		if err != nil {
			log.Errorc(ctx, "Convert2HeroAnalysis() error(%+v)", err)
		}
		analysis.Baned, err = strconv.ParseInt(origin.Baned, 10, 64)
		if err != nil {
			log.Errorc(ctx, "Convert2HeroAnalysis() error(%+v)", err)
		}
		analysis.BanedRate, err = strconv.ParseFloat(origin.BanedRate, 64)
		if err != nil {
			log.Errorc(ctx, "Convert2HeroAnalysis() error(%+v)", err)
		}
		analysis.Win, err = strconv.ParseInt(origin.Win, 10, 64)
		if err != nil {
			log.Errorc(ctx, "Convert2HeroAnalysis() error(%+v)", err)
		}
		analysis.WinRate, err = strconv.ParseFloat(origin.WinRate, 64)
		if err != nil {
			log.Errorc(ctx, "Convert2HeroAnalysis() error(%+v)", err)
		}
		analysis.TotalRound, err = strconv.ParseInt(origin.TotalRound, 10, 64)
		if err != nil {
			log.Errorc(ctx, "Convert2HeroAnalysis() error(%+v)", err)
		}
		analysis.UpdatedAt, err = strconv.ParseInt(origin.UpdatedAt, 10, 64)
		if err != nil {
			log.Errorc(ctx, "Convert2HeroAnalysis() error(%+v)", err)
		}
		analysis.Score, err = strconv.ParseFloat(origin.Score, 64)
		if err != nil {
			log.Errorc(ctx, "Convert2HeroAnalysis() error(%+v)", err)
		}
		analysis.AvgKills, err = strconv.ParseFloat(origin.AvgKills, 64)
		if err != nil {
			log.Errorc(ctx, "Convert2HeroAnalysis() error(%+v)", err)
		}
		analysis.AvgDeaths, err = strconv.ParseFloat(origin.AvgDeaths, 64)
		if err != nil {
			log.Errorc(ctx, "Convert2HeroAnalysis() error(%+v)", err)
		}
		analysis.AvgAssists, err = strconv.ParseFloat(origin.AvgAssists, 64)
		if err != nil {
			log.Errorc(ctx, "Convert2HeroAnalysis() error(%+v)", err)
		}
		analysis.GameVersion = origin.GameVersion
	}
	return
}

func (player *ScorePlayerAnalysis) InsertUpdate(ctx context.Context) (err error) {
	_, err = sql.GlobalDB.Exec(
		ctx,
		sqlOfInsertUpdate4ScorePlayerAnalysis,
		player.TournamentID, player.PlayerID, player.Name, player.EnName, player.Image,
		player.TeamID, player.TeamName, player.Image, player.Position, player.PositionId,
		player.KDA, player.Played, player.ParticipateRate, player.AvgKills, player.AvgAssists,
		player.AvgDeaths, player.AvgEconomyMin, player.AvgHitsMin, player.AvgDamageMin, player.AvgDamageTakenMin,
		player.DamageRate, player.DamageTakenRate, player.AvgWardsPlacedMin, player.AvgWardsKilledMin, player.UpdatedAt,
		player.Mvp, player.Win, player.Lose, player.WinRate, player.CountryID,
		player.CountryImage, player.Score, player.TotalKills, player.TotalDeaths, player.TotalAssists,
		player.TournamentID, player.PlayerID, player.Name, player.EnName, player.Image,
		player.TeamID, player.TeamName, player.Image, player.Position, player.PositionId,
		player.KDA, player.Played, player.ParticipateRate, player.AvgKills, player.AvgAssists,
		player.AvgDeaths, player.AvgEconomyMin, player.AvgHitsMin, player.AvgDamageMin, player.AvgDamageTakenMin,
		player.DamageRate, player.DamageTakenRate, player.AvgWardsPlacedMin, player.AvgWardsKilledMin, player.UpdatedAt,
		player.Mvp, player.Win, player.Lose, player.WinRate, player.CountryID,
		player.CountryImage, player.Score, player.TotalKills, player.TotalDeaths, player.TotalAssists)

	return
}

func (hero *ScoreHeroAnalysis) InsertUpdate(ctx context.Context) (err error) {
	_, err = sql.GlobalDB.Exec(
		ctx,
		sqlOfInsertUpdate4ScoreHeroAnalysis,
		hero.TournamentID, hero.HeroID, hero.Name, hero.EnName, hero.TwName,
		hero.Image, hero.PositionId, hero.Position, hero.KDA, hero.Picked,
		hero.PickedRate, hero.Baned, hero.BanedRate, hero.WinRate, hero.Win,
		hero.TotalRound, hero.UpdatedAt, hero.Score, hero.AvgKills, hero.AvgAssists,
		hero.AvgDeaths, hero.GameVersion,
		hero.TournamentID, hero.HeroID, hero.Name, hero.EnName, hero.TwName,
		hero.Image, hero.PositionId, hero.Position, hero.KDA, hero.Picked,
		hero.PickedRate, hero.Baned, hero.BanedRate, hero.WinRate, hero.Win,
		hero.TotalRound, hero.UpdatedAt, hero.Score, hero.AvgKills, hero.AvgAssists,
		hero.AvgDeaths, hero.GameVersion)

	return
}
