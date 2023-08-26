package model

type VideoList struct {
	ID        int64  `json:"id" form:"id"`
	ListName  string `json:"list_name" form:"list_name" validate:"required"`
	UgcAids   string `json:"ugc_aids" form:"ugc_aids"`
	GameID    int64  `json:"game_id" form:"game_id"`
	MatchID   int64  `json:"match_id" form:"match_id"`
	YearID    int64  `json:"year_id" form:"year_id"`
	Stime     int64  `json:"stime" form:"stime" validate:"required"`
	Etime     int64  `json:"etime" form:"etime" validate:"required"`
	IsDeleted int    `json:"is_deleted" form:"is_deleted"`
}

type CheckArchive struct {
	WrongList []string `json:"wrong_list"`
}

type ParamVideoFilter struct {
	GameID  int64 `form:"game_id"   validate:"gte=0"`
	MatchID int64 `form:"match_id"   validate:"gte=0"`
	YearID  int64 `form:"year_id"   validate:"gte=0"`
}

// TableName .
func (t VideoList) TableName() string {
	return "es_video_lists"
}
