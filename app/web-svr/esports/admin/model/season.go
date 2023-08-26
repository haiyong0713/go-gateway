package model

var (
	SeasonTypeNormal = 0
	SeasonTypeEscape = 1
)

// Season .
type Season struct {
	ID             int64  `json:"id" form:"id"`
	Mid            int64  `json:"mid" form:"mid" validate:"required"`
	Title          string `json:"title" form:"title" validate:"required"`
	SubTitle       string `json:"sub_title" form:"sub_title"`
	Stime          int64  `json:"stime" form:"stime"`
	Etime          int64  `json:"etime" form:"etime"`
	Sponsor        string `json:"sponsor" form:"sponsor"`
	Logo           string `json:"logo" form:"logo" validate:"required"`
	Dic            string `json:"dic" form:"dic"`
	Status         int    `json:"status"  form:"is_deleted"`
	IsApp          int    `json:"is_app" form:"is_app"`
	Rank           int    `json:"rank" form:"rank" validate:"min=0,max=99"`
	URL            string `json:"url" form:"url"`
	DataFocus      string `json:"data_focus" form:"data_focus"`
	FocusURL       string `json:"focus_url" form:"focus_url"`
	ForbidIndex    int    `json:"forbid_index" form:"forbid_index"`
	LeidaSid       int    `json:"leida_sid" form:"leida_sid"`
	SerieType      int    `json:"serie_type" form:"serie_type"`
	SearchImage    string `json:"search_image" form:"search_image"`
	Platforms      string `json:"platforms" form:"platforms" gorm:"-"`
	SyncPlatform   int64  `json:"sync_platform" form:"sync_platform"`
	SeasonType     int    `json:"season_type" form:"season_type"`
	MessageSenduid int64  `json:"message_senduid" form:"message_senduid"`
}

// SeasonInfo .
type SeasonInfo struct {
	*Season
	SeasonRank int64   `json:"season_rank"`
	RankID     int64   `json:"rank_id"`
	Games      []*Game `json:"games"`
}

// TableName .
func (s Season) TableName() string {
	return "es_seasons"
}

// PlatformVal get platform val by bit.
func (s *Season) PlatformVal(bit uint) bool {
	rs := s.SyncPlatform >> bit & int64(1)
	return rs > 0
}

type TeamInSeason struct {
	//Season id
	Sid int64 `gorm:"primary_key;auto_increment:false" json:"sid" form:"sid" validate:"required"`
	//Team id
	Tid int64 `gorm:"primary_key;auto_increment:false" json:"tid" form:"tid" validate:"required"`
	//Rank priority
	Rank int64 `json:"rank" form:"rank" validate:"min=0,max=99"`
}

type TeamInSeasonParam struct {
	Sid  int64 `form:"sid" validate:"min=1"`
	Tid  int64 `form:"tid" validate:"min=1"`
	Rank int64 `form:"rank" validate:"min=0,max=99"`
}

type TeamInSeasonResponse struct {
	*Team
	Rank int64 `json:"rank" form:"rank" validate:"min=0,max=99"`
}

func NewTeamInSeason(seasonId, teamId, rank int64) *TeamInSeason {
	return &TeamInSeason{
		Sid:  seasonId,
		Tid:  teamId,
		Rank: rank,
	}
}

// TeamInSeason TableName .
func (s TeamInSeason) TableName() string {
	return "es_team_in_seasons"
}
