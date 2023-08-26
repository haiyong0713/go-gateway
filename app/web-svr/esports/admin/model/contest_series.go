package model

import "go-gateway/app/web-svr/esports/admin/component"

const (
	Identity4NotDeleted = 0
	Identity4Deleted    = 1

	tableOfContestSeries = "contest_series"
)

type ContestSeries struct {
	ID            int64  `json:"id" form:"id"`
	Type          int64  `json:"type" form:"type"`
	ParentTitle   string `json:"parent_title" form:"parent_title" validate:"required"`
	ChildTitle    string `json:"child_title" form:"child_title"`
	StartTime     int64  `json:"start_time" form:"start_time" validate:"min=1"`
	EndTime       int64  `json:"end_time" form:"end_time" validate:"min=1"`
	ScoreID       string `json:"score_id" form:"score_id" validate:"required"`
	SeasonID      int64  `json:"season_id" form:"season_id" validate:"min=1"`
	ViewGenerated bool   `json:"view_generated,omitempty" form:"-" gorm:"-"`
	IsDeleted     int    `json:"-" form:"-" gorm:"column:is_deleted"`
}

type PUBGContestSeriesExtraRules struct {
	ScoreRules PUBGContestSeriesScoreRule `json:"score_rules"`
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

func (series ContestSeries) TableName() string {
	return tableOfContestSeries
}

func (series ContestSeriesByScoreRule) TableName() string {
	return tableOfContestSeries
}

func FindContestSeriesByID(id int64) (*ContestSeries, error) {
	series := new(ContestSeries)
	err := component.GlobalDB.Where("id = ? AND is_deleted = ?", id, Identity4NotDeleted).Find(series).Error

	return series, err
}

func (series *ContestSeries) Insert() error {
	series.ID = 0
	return component.GlobalDB.Create(series).Error
}

func (series *ContestSeries) Update() error {
	return component.GlobalDB.Save(series).Error
}

func (series *ContestSeries) Delete() error {
	return component.GlobalDB.Model(series).Updates(map[string]interface{}{"is_deleted": Identity4Deleted}).Error
}

func (series *ContestSeries) MarkAsNotDeleted() error {
	return component.GlobalDB.Model(series).Updates(map[string]interface{}{"is_deleted": Identity4NotDeleted}).Error
}

func ContestSeriesCount(seasonID int64) (count int64) {
	component.GlobalDB.Table(tableOfContestSeries).
		Where("season_id = ? and is_deleted = ?", seasonID, Identity4NotDeleted).Count(&count)

	return
}

func ContestSeriesList(seasonID, limit, offset int64) (list []*ContestSeries, err error) {
	list = make([]*ContestSeries, 0)
	err = component.GlobalDB.Where("season_id = ? and is_deleted = ?", seasonID, Identity4NotDeleted).Order("id asc").Offset(offset).Limit(limit).Find(&list).Error

	return
}
