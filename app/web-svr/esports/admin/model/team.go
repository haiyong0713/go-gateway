package model

// Team .
type Team struct {
	ID         int64  `json:"id" form:"id"`
	Title      string `json:"title" form:"title" validate:"required"`
	SubTitle   string `json:"sub_title" form:"sub_title"`
	ETitle     string `json:"e_title" form:"e_title"`
	CreateTime int64  `json:"create_time" form:"create_time"`
	Area       string `json:"area" form:"area"`
	Logo       string `json:"logo" form:"logo" validate:"required"`
	UID        int64  `json:"uid" form:"uid" gorm:"column:uid"`
	Members    string `json:"members" form:"members"`
	Dic        string `json:"dic" form:"dic"`
	IsDeleted  int    `json:"is_deleted" form:"is_deleted"`
	VideoURL   string `json:"video_url" form:"video_url"`
	Profile    string `json:"profile" form:"profile"`
	LeidaTid   int    `json:"leida_tid" form:"leida_tid"`
	TeamType   int64  `json:"team_type" form:"team_type"`
	ReplyID    int64  `json:"-" gorm:"-"`
	Adid       int64  `json:"-" form:"adid" gorm:"-" validate:"required"`
	RegionID   int    `json:"region_id" form:"region_id" gorm:"column:region_id"`
	PictureUrl string `json:"picture_url" form:"picture_url"`
}

// TeamInfo .
type TeamInfo struct {
	*Team
	Games []*Game `json:"games"`
}

// TableName .
func (t Team) TableName() string {
	return "es_teams"
}
