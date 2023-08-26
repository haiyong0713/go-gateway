package model

import "go-gateway/app/web-svr/activity/interface/api"

type UsesrPoint struct {
	Points  int32 `json:"points"`
	IsLogin bool  `json:"is_login"`
}

type Tasks struct {
	Banner        string              `json:"banner"`
	PointsAct     string              `json:"points_act"`
	WebBanner     string              `json:"web_banner"`
	PointsActWeb  string              `json:"points_act_web"`
	TasksProgress []*api.TaskProgress `json:"tasks"`
	User          *UsesrPoint         `json:"user"`
	SeasonID      int64               `json:"season_id"`
}

type RankingDataRet struct {
	*RankingData
	Matches           map[int64]int64      `json:"matches"`
	TeamRegion        map[string]int       `json:"team_region"`
	Previous          *RankingDataPrevious `json:"previous,omitempty"`
	Picture           string               `json:"picture,omitempty"`
	PromoteNum        uint8                `json:"promote_num,omitempty"`
	EliminateNum      uint8                `json:"eliminate_num,omitempty"`
	FinalPromoteNum   uint8                `json:"final_promote_num,omitempty"`
	FinalEliminateNum uint8                `json:"final_eliminate_num,omitempty"`
	Description       string               `json:"description"`
	RoundID           string               `json:"round_id"`
	State             int                  `json:"state"`
}

type RankingDataPrevious struct {
	*RankingData
	Picture string `json:"picture,omitempty"`
}

type RankingData struct {
	Stage       int              `json:"stage"`
	ISNullPoint bool             `json:"is_null_point"`
	PointList   []*PointInfo     `json:"point_list"`
	Tree        []*RoundTreeNode `json:"tree"`
	Mtime       int64            `json:"mtime"`
}

type S10RankingInterventionData struct {
	TournamentID      string `json:"tournament_id" form:"tournament_id"`
	CurrentRound      string `json:"current_round" form:"current_round"`
	FinalistRound     string `json:"finalist_round" form:"finalist_round"`
	PromoteNum        uint8  `json:"promote_num" form:"promote_num"`
	EliminateNum      uint8  `json:"eliminate_num" form:"eliminate_num"`
	FinalPromoteNum   uint8  `json:"final_promote_num" form:"final_promote_num"`
	FinalEliminateNum uint8  `json:"final_eliminate_num" form:"final_eliminate_num"`
	RoundInfo         []struct {
		RoundID string `json:"round_id"`
		H5Pic   string `json:"h_5_pic"`
		WebPic  string `json:"web_pic"`
	} `json:"round_info" form:"round_info"`
}

type PointInfo struct {
	Letter  string `json:"letter"`
	GroupID string `json:"group_id"`
	List    []*struct {
		TeamID        string      `json:"team_id"`
		TeamShortName string      `json:"team_short_name"`
		TeamImage     string      `json:"team_image"`
		Win           string      `json:"win"`
		Los           string      `json:"los"`
		WinLose       string      `json:"win_lose"`
		WLNum         interface{} `json:"w_l_num"`
		Percent       int         `json:"percent"`
		Sorting       int         `json:"sorting"`
	} `json:"list"`
}

type RoundTreeNode struct {
	MatchID        string           `json:"match_id"`
	TeamID         string           `json:"team_id"`
	Remark         string           `json:"remark"`
	Children       []*RoundTreeNode `json:"Children,omitempty"`
	MatchStatus    string           `json:"match_status"`
	TeamShortName  string           `json:"team_short_name"`
	TeamImage      string           `json:"team_image"`
	TeamAID        string           `json:"team_a_id"`
	TeamAShortName string           `json:"team_a_short_name"`
	TeamAImage     string           `json:"team_a_image"`
	TeamAWin       string           `json:"team_a_win"`
	TeamBID        string           `json:"team_b_id"`
	TeamBShortName string           `json:"team_b_short_name"`
	TeamBImage     string           `json:"team_b_image"`
	TeamBWin       string           `json:"team_b_win"`
	MatchTime      string           `json:"match_time"`
}

// Team team.
type Team2Tab struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	SubTitle    string `json:"sub_title"`
	Logo        string `json:"logo"`
	RegionID    int    `json:"region_id"`
	Region      string `json:"region"`
	ScoreTeamID int64  `json:"score_team_id"`
}
