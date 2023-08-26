package model

var (
	ContestTeamDeleted           int64 = 1
	ContestTeamNotDeleted        int64 = 0
	ContestTeamRankEditStatusOn        = 1
	ContestTeamRankEditStatusOff       = 0
)

type ContestTeam struct {
	ID             int64 `json:"id" form:"id"`
	ContestId      int64 `json:"contest_id" form:"contest_id"`
	TeamId         int64 `json:"team_id" form:"team_id"`
	SurvivalRank   int64 `json:"survival_rank" form:"survival_rank"`
	KillNumber     int64 `json:"kill_number" form:"kill_number"`
	Score          int64 `json:"score" form:"score"`
	RankEditStatus int   `json:"rank_edit_status" form:"rank_edit_status"`
	IsDeleted      int   `json:"is_deleted"`
}

type TeamScores struct {
	ID         int64 `json:"id" form:"id"`
	KillNumber int64 `json:"kill_number" form:"kill_number"`
}

type ContestTeamsCheckResponse struct {
	Teams []*Team `json:"teams"`
}

type ContestTeamScoresResponse struct {
	TeamScores []*ContestTeam `json:"team_scores"`
}

// TableName es_contest_teams
func (c ContestTeam) TableName() string {
	return "es_contest_teams"
}
