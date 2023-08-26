package model

import (
	accClient "git.bilibili.co/bapis/bapis-go/account/service"

	actpb "go-gateway/app/web-svr/activity/interface/api"
)

type UserSeasonGuessReq struct {
	PageStructure

	MID      int64
	SeasonID int64 `form:"season_id" validate:"required,min=1"`
}

type UserSeasonGuessResp struct {
	PageStructure

	Data []*MatchGuess `json:"data"`
}

type SeasonGuessSummary struct {
	Total int64   `json:"total"`
	Wins  int64   `json:"wins"`
	Coins float32 `json:"coins"`
}

type MatchGuess struct {
	*actpb.GuessUserGroup

	GameStage string                `json:"game_stage"`
	HomeTeam  *Team4SimplifyEdition `json:"home_team"`
	AwayTeam  *Team4SimplifyEdition `json:"away_team"`
}

type UserGuessOptions struct {
	ID    int64                 `json:"id"`
	Title string                `json:"title"`
	IsBet bool                  `json:"is_bet"`
	Team  *Team4SimplifyEdition `json:"team"`
}

type Team4SimplifyEdition struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	Logo  string `json:"logo"`
}

type GuessParams4V2 struct {
	MID      int64
	SeasonID int64 `form:"season_id" validate:"required,min=1"`
}

// AddGuessParam .
type AddGuessParam struct {
	MID      int64
	OID      int64 `form:"oid" validate:"required"`
	MainID   int64 `form:"main_id" validate:"required"`
	DetailID int64 `form:"detail_id" validate:"required"`
	Count    int64 `form:"count" validate:"required,min=1"`
	IsFav    int64 `form:"is_fav"`
}

// GuessCollection .
type GuessCollection struct {
	Game   []*Filter `json:"game"`
	Season []*Season `json:"season"`
}

// MoreShow .
type MoreShow struct {
	Home []*Contest `json:"home"`
	Away []*Contest `json:"away"`
}

// GuessTeamStats .
type GuessTeamStats struct {
	HomeStats interface{} `json:"home_stats"`
	AwayStats interface{} `json:"away_stats"`
}

// GuessDetail .
type GuessDetail struct {
	Contest *Contest           `json:"contest"`
	Guess   []*actpb.GuessList `json:"guess"`
	Stats   *GuessTeamStats    `json:"team_stats"`
	Detail  []*ContestsData    `json:"detail"`
}

// GuessCollQues .
type GuessCollQues struct {
	Contest   *Contest           `json:"contest"`
	Questions []*actpb.GuessList `json:"questions"`
}

// GuessCollRecoParam .
type GuessCollRecoParam struct {
	Type int64 `form:"type" validate:"min=1,max=2"`
	Mid  int64 `form:"mid"`
	Pn   int64 `form:"pn"`
	Ps   int64 `form:"ps"`
}

// PageGuess .
type PageGuess struct {
	Pn    int `json:"pn"`
	Ps    int `json:"ps"`
	Total int `json:"total"`
}

// GuessCollReco .
type GuessCollReco struct {
	Contest     *Contest                `json:"contest"`
	Guess       []*actpb.GuessUserGroup `json:"guess"`
	ContestRank int64                   `json:"contest_rank"`
	ContestID   int64                   `json:"contest_id"`
}

// GuessMatchReco .
type GuessMatchReco struct {
	Guess []*actpb.GuessUserGroup `json:"guess"`
}

// GuessCollRecoRes .
type GuessCollRecoRes struct {
	GuessCollReco []*GuessCollReco `json:"record"`
	Page          *actpb.PageInfo  `json:"page"`
}

// UserProfile .
type UserProfile struct {
	Profile *accClient.Profile `json:"profile"`
	Coin    float64            `json:"coin"`
}

// GuessCollUser .
type GuessCollUser struct {
	GuessData *actpb.UserGuessDataReply `json:"guess"`
}
