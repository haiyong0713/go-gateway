package model

import (
	"go-gateway/app/web-svr/esports/interface/model"
	"time"
)

var (
	ContestSpecial = int64(1)
)

// Contest .
type Contest struct {
	ID            int64  `json:"id" form:"id"`
	GameStage     string `json:"game_stage" form:"game_stage" validate:"required"`
	Stime         int64  `json:"stime" form:"stime"`
	Etime         int64  `json:"etime" form:"etime"`
	HomeID        int64  `json:"home_id" form:"home_id"`
	AwayID        int64  `json:"away_id" form:"away_id"`
	HomeScore     int64  `json:"home_score" form:"home_score"`
	AwayScore     int64  `json:"away_score" form:"away_score"`
	LiveRoom      int64  `json:"live_room" form:"live_room"`
	Aid           int64  `json:"aid" form:"aid"`
	Collection    int64  `json:"collection" form:"collection"`
	GameState     int64  `json:"game_state" form:"game_state"`
	Dic           string `json:"dic" form:"dic"`
	Status        int64  `json:"status" form:"status"`
	Sid           int64  `json:"sid" form:"sid" validate:"required"`
	Mid           int64  `json:"mid" form:"mid" validate:"required"`
	Special       int64  `json:"special" form:"special"`
	SuccessTeam   int64  `json:"success_team" form:"success_team"`
	SpecialName   string `json:"special_name" form:"special_name"`
	SpecialTips   string `json:"special_tips" form:"special_tips"`
	SpecialImage  string `json:"special_image" form:"special_image"`
	Playback      string `json:"playback" form:"playback"`
	CollectionURL string `json:"collection_url" form:"collection_url"`
	LiveURL       string `json:"live_url" form:"live_url"`
	DataType      int64  `json:"data_type" form:"data_type"`
	Data          string `json:"-" form:"data" gorm:"-"`
	Adid          int64  `json:"-" form:"adid"  gorm:"-" validate:"required"`
	MatchID       int64  `json:"match_id" form:"match_id"`
	GuessType     int64  `json:"guess_type" form:"guess_type"`
	GameStage1    string `json:"game_stage1" form:"game_stage1" validate:"required"`
	GameStage2    string `json:"game_stage2" form:"game_stage2"`
	SeriesID      int64  `json:"series_id" form:"series_id" validate:"min=0"`
	PushSwitch    int64  `json:"push_switch"`
	TeamIds       string `json:"team_ids" form:"team_ids" gorm:"-"`
	ContestStatus int64  `json:"contest_status"`
	ExternalId    int64  `json:"external_id"`
}

// ContestInfo .
type ContestInfo struct {
	*Contest
	Games       []*Game        `json:"games"`
	HomeName    string         `json:"home_name"`
	AwayName    string         `json:"away_name"`
	SuccessName string         `json:"success_name" form:"success_name"`
	Data        []*ContestData `json:"data"`
	Series      *ContestSeries `json:"series"`
}

// ContestCard
type ContestCard struct {
	*Contest
	HomeName    string `json:"home_name"`
	AwayName    string `json:"away_name"`
	SeasonName  string `json:"season_name"`
	SuccessName string `json:"success_name" form:"success_name"`
}

// TableName es_contests
func (c Contest) TableName() string {
	return "es_contests"
}

type Material struct {
	ID   int    `json:"id"`
	Type int    `json:"type"`
	Data string `json:"data"`
}

func (contest *Contest) CalculateStatus() string {
	now := time.Now().Unix()
	if now >= contest.Etime {
		return model.ContestStatusOfEnd
	} else if now >= contest.Stime {
		return model.ContestStatusOfOngoing
	}

	return model.ContestStatusOfNotStart
}
