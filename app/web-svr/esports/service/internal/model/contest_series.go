package model

type ContestSeriesModel struct {
	ID          int64  `json:"id"`
	ParentTitle string `json:"parent_title"`
	ChildTitle  string `json:"child_title"`
	ScoreId     string `json:"score_id"`
	SeasonId    int64  `json:"season_id"`
	StartTime   int64  `json:"start_time"`
	EndTime     int64  `json:"end_time"`
	Type        int64  `json:"type"`
	IsDeleted   int64  `json:"is_deleted"`
}

type ContestSeriesByScoreRule struct {
	ID              int64  `json:"id" form:"id"`
	Type            int64  `json:"type" form:"type"`
	ParentTitle     string `json:"parent_title" form:"parent_title" validate:"required"`
	ChildTitle      string `json:"child_title" form:"child_title"`
	StartTime       int64  `json:"start_time" form:"start_time" validate:"min=1"`
	EndTime         int64  `json:"end_time" form:"end_time" validate:"min=1"`
	ScoreID         string `json:"score_id" form:"score_id" validate:"required"`
	SeasonID        int64  `json:"season_id" form:"season_id" validate:"min=1"`
	ViewGenerated   bool   `json:"view_generated,omitempty" form:"-" gorm:"-"`
	IsDeleted       int    `json:"-" form:"-" gorm:"column:is_deleted"`
	ScoreRuleConfig string `json:"score_rule_config" gorm:"column:score_rule_config"`
}

type PUBGContestSeriesScoreRule struct {
	KillScore  int64   `json:"kill_score" form:"kill_score"`
	RankScores []int64 `json:"rank_scores" form:"rank_scores"`
}
