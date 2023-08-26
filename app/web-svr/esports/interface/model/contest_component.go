package model

import (
	"time"

	pb "go-gateway/app/web-svr/esports/interface/api/v1"
)

type ComponentContestCardTime struct {
	Timestamp    int64                      `json:"timestamp"`
	ContestCards []*pb.ContestCardComponent `json:"contests"`
}

// ComponentSeason  component going  season  struct.
type ComponentSeason struct {
	ID        int64  `json:"id"`
	LeidaSid  int64  `json:"leida_sid"`
	Stime     int64  `json:"stime"`
	Etime     int64  `json:"etime"`
	SerieType int64  `json:"serie_type,omitempty"`
	Title     string `json:"title"`
}

// Contest contest
type Contest2TabComponent struct {
	ID            int64  `json:"id"`
	StimeDate     int64  `jsob:"date"`
	Stime         int64  `json:"stime"`
	Etime         int64  `json:"etime"`
	CollectionUrl string `json:"collection_url"`
	LiveRoom      int64  `json:"live_room"`
	PlayBack      string `json:"play_back"`
	HomeID        int64  `json:"home_id"`
	HomeScore     int64  `json:"home_score"`
	AwayID        int64  `json:"away_id"`
	AwayScore     int64  `json:"away_score"`
	DataType      int64  `json:"data_type"`
	MatchID       int64  `json:"match_id"`
	GameStage     string `json:"stage"`
	SeriesID      int64  `json:"series_id"`
	GuessType     int64  `json:"guess_type"`
	SeasonID      int64  `json:"season_id"`
	Status        int64  `json:"status"`
	ContestStatus int64  `json:"contest_status"`
}

// HomeAwayContestComponent .
type HomeAwayContestComponent struct {
	HomeTeam    *HomeAwayTeam             `json:"home_team"`
	AwayTeam    *HomeAwayTeam             `json:"away_team"`
	SuccessList []*HomeAwaySuccessContest `json:"success_list"`
}

// HomeAwayTeam .
type HomeAwayTeam struct {
	*Team2TabComponent
	WinCount int `json:"win_count"`
}

// HomeAwaySuccessContest .
type HomeAwaySuccessContest struct {
	*Team2TabComponent
	ContestStime int64 `json:"contest_stime"`
}

func (contest *Contest2TabComponent) CalculateStatus() string {
	now := time.Now().Unix()
	if now >= contest.Etime {
		return ContestStatusOfEnd
	} else if now >= contest.Stime {
		return ContestStatusOfOngoing
	}

	return ContestStatusOfNotStart
}

// Team component team.
type Team2TabComponent struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	SubTitle    string `json:"sub_title"`
	Logo        string `json:"logo"`
	RegionID    int64  `json:"region_id"`
	Region      string `json:"region"`
	ScoreTeamID int64  `json:"score_team_id"`
}

// ComponentContestCardFold.
type ComponentContestCardFold struct {
	FrontList  []*pb.ContestCardComponent `json:"front_list"`
	MiddleList []*pb.ContestCardComponent `json:"middle_list"`
	BackList   []*pb.ContestCardComponent `json:"back_list"`
}

// ComponentContestAbstract .
type ComponentContestAbstract struct {
	History []*pb.ContestCardComponent `json:"history"`
	Future  []*pb.ContestCardComponent `json:"future"`
}

// ComponentSeasonContests .
type ComponentSeasonContests struct {
	History []*pb.ContestCardComponent `json:"history"`
	Future  []*pb.ContestCardComponent `json:"future"`
	Prev    int                        `json:"prev"`
	Next    int                        `json:"next"`
}

type ContestTeamInfoResult struct {
	List    []*ContestTeamInfo `json:"list"`
	IsScore int64              `json:"is_score"`
}

type ContestTeamInfo struct {
	TeamId    int64                 `json:"team_id"`
	TeamInfo  *Team2TabComponent    `json:"team_info"`
	ScoreInfo *ContestTeamScoreInfo `json:"score_info"`
}

type ContestTeamScoreInfo struct {
	TeamId         int64 `json:"team_id"`
	Score          int64 `json:"score"`
	KillNumber     int64 `json:"kill_number"`
	SurvivalRank   int64 `json:"survival_rank"`
	SeasonTeamRank int64 `json:"season_team_rank"`
	Rank           int64 `json:"rank"` // 最终排名，用于前端显示
}

type ContestTeamDbInfo struct {
	ContestId      int64 `json:"contest_id"`
	TeamId         int64 `json:"team_id"`
	Score          int64 `json:"score"`
	KillNumber     int64 `json:"kill_number"`
	SurvivalRank   int64 `json:"survival_rank"`
	RankEditStatus int8  `json:"rank_edit_status"`
}

// Contest contest.
type ContestBattle2DBComponent struct {
	ID            int64  `json:"id"`
	StimeDate     int64  `jsob:"date"`
	Stime         int64  `json:"stime"`
	Etime         int64  `json:"etime"`
	CollectionUrl string `json:"collection_url"`
	LiveRoom      int64  `json:"live_room"`
	PlayBack      string `json:"play_back"`
	MatchID       int64  `json:"match_id"`
	GameStage     string `json:"stage"`
	SeriesID      int64  `json:"series_id"`
	GuessType     int64  `json:"guess_type"`
	SeasonID      int64  `json:"season_id"`
	Status        int64  `json:"status"`
	ContestStatus int64  `json:"contest_status"`
}

func (contest *ContestBattle2DBComponent) CalculateStatus() string {
	now := time.Now().Unix()
	if now >= contest.Etime {
		return ContestStatusOfEnd
	} else if now >= contest.Stime {
		return ContestStatusOfOngoing
	}

	return ContestStatusOfNotStart
}

type ContestBattleCardComponent struct {
	*pb.ContestBattleCardComponent
	TeamCount int                  `json:"team_count"`
	IsScore   bool                 `json:"is_score"`
	TeamList  []*ContestBattleTeam `json:"team_list"`
}

type ContestBattleTeam struct {
	Title string `json:"title"`
	Logo  string `json:"logo"`
	*ContestTeamScoreInfo
}

// SeasonComponent .
type SeasonComponent struct {
	ID       int64  `json:"id"`
	Title    string `json:"title"`
	SubTitle string `json:"sub_title"`
	Logo     string `json:"logo"`
	LeidaSid int64  `json:"leida_sid"`
}

type SeasonTeams2Component struct {
	Season *SeasonComponent `json:"season"`
	Teams  []*TeamInSeason  `json:"teams"`
}

// VideoListInfo .
type VideoListInfo struct {
	ID      int64  `json:"id"`
	UgcAids string `json:"ugc_aids"`
	GameID  int64  `json:"game_id"`
	MatchID int64  `json:"match_id"`
	YearID  int64  `json:"year_id"`
}

// VideoList2Component .
type VideoList2Component struct {
	UgcList []*Video `json:"ugc_list"`
	*VideoList
}

// VideoList .
type VideoList struct {
	List []*Video `json:"list"`
	Page *Page    `json:"page"`
}

// LolDataHero2 .
type LolDataHero2 struct {
	ID            int64   `json:"id"`
	TournamentID  int64   `json:"tournament_id"`
	HeroID        int64   `json:"hero_id"`
	HeroName      string  `json:"hero_name"`
	HeroImage     string  `json:"hero_image"`
	AppearCount   int64   `json:"appear_count"`
	ProhibitCount int64   `json:"prohibit_count"`
	VictoryCount  int64   `json:"victory_count"`
	GameCount     int64   `json:"game_count"`
	VictoryRate   float64 `json:"victory_rate"`
}

// ComponentContestWall .
type ComponentContestWall struct {
	Contest interface{} `json:"contest"`
}
