package model

// SeasonRank .
type SeasonRank struct {
	ID         int64  `json:"id" form:"id"`
	Gid        int64  `json:"gid" form:"gid" validate:"min=0"`
	GameName   string `json:"game_name" gorm:"-"`
	Sid        int64  `json:"sid" form:"sid" validate:"required"`
	SeasonName string `json:"season_name" gorm:"-"`
	Rank       int64  `json:"rank" form:"rank" validate:"min=1,max=20"`
	IsDeleted  int64  `json:"is_deleted" form:"-"`
}

// TableName .
func (s SeasonRank) TableName() string {
	return "es_season_ranks"
}
