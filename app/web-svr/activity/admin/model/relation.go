package model

const (
	ActRelationSubjectStatusNormal  = 0
	ActRelationSubjectStatusOffline = -1
)

type ActRelationListArgs struct {
	Page     int    `form:"page" default:"1" validate:"min=1"`
	PageSize int    `form:"page_size" default:"15" validate:"min=1,max=15"`
	Keyword  string `form:"keyword"`
}

type ActRelationSubject struct {
	ID             int64  `json:"id" gorm:"id"`
	Name           string `json:"name" gorm:"name" form:"name" validate:"required"`
	Description    string `json:"description" gorm:"description" form:"description"`
	NativeIDs      string `json:"native_ids" gorm:"native_ids" form:"native_ids"`
	H5IDs          string `json:"h5_ids" gorm:"h5_ids" form:"h5_ids"`
	WebIDs         string `json:"web_ids" gorm:"web_ids" form:"web_ids"`
	LotteryIDs     string `json:"lottery_ids" gorm:"lottery_ids" form:"lottery_ids"`
	ReserveIDs     string `json:"reserve_ids" gorm:"reserve_ids" form:"reserve_ids"`
	VideoSourceIDs string `json:"video_source_ids" gorm:"video_source_ids" form:"video_source_ids"`
	FollowIDs      string `json:"follow_ids" gorm:"follow_ids" form:"follow_ids"`
	SeasonIDs      string `json:"season_ids" gorm:"season_ids" form:"season_ids"`
	MallIDs        string `json:"mall_ids" gorm:"mall_ids" form:"mall_ids"`
	TopicIDs       string `json:"topic_ids" gorm:"topic_ids" form:"topic_ids"`
	FavoriteInfo   string `json:"favorite_info" gorm:"favorite_info" form:"favorite_info"`
	ReserveConfig  string `json:"reserve_config" gorm:"reserve_config" form:"reserve_config"`
	FollowConfig   string `json:"follow_config" gorm:"follow_config" form:"follow_config"`
	SeasonConfig   string `json:"season_config" gorm:"season_config" form:"season_config"`
	FavoriteConfig string `json:"favorite_config" gorm:"favorite_config" form:"favorite_config"`
	MallConfig     string `json:"mall_config" gorm:"mall_config" form:"mall_config"`
	TopicConfig    string `json:"topic_config" gorm:"topic_config" form:"topic_config"`
	State          int64  `json:"-" gorm:"state"`
}

type ActRelationListRes struct {
	List     []*ActRelationSubject `json:"list"`
	Page     int                   `json:"page"`
	PageSize int                   `json:"page_size"`
	Count    int64                 `json:"count"`
}

type ActRelationSubjectAdd struct {
	ID             int64  `json:"id"`
	Name           string `json:"name" form:"name" validate:"required"`
	Description    string `json:"description" form:"description"`
	NativeIDs      string `json:"native_ids" form:"native_ids"`
	H5IDs          string `json:"h5_ids" form:"h5_ids"`
	WebIDs         string `json:"web_ids" form:"web_ids"`
	LotteryIDs     string `json:"lottery_ids" form:"lottery_ids"`
	ReserveIDs     string `json:"reserve_ids" form:"reserve_ids"`
	VideoSourceIDs string `json:"video_source_ids" form:"video_source_ids"`
}

func (ActRelationSubject) TableName() string {
	return "act_relation_subject"
}

type ActRelationConfigRule struct {
	StartTime int `json:"start_time"`
	EndTime   int `json:"end_time"`
}

type ActRelationFavorite struct {
	Type    int64  `json:"type"`
	Content string `json:"content"`
}
