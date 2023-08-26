package model

type AutoSubRequest struct {
	SeasonID   int64   `form:"season_id" validate:"required,min=1"`
	TeamIDList []int64 `form:"team_ids,split" validate:"required,dive,gt=0"`
}
