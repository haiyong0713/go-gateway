package game

type Game struct {
	GameBaseID    int64          `json:"game_base_id,omitempty"`
	IsOnline      bool           `json:"is_online,omitempty"`
	GameName      string         `json:"game_name,omitempty"`
	Cover         string         `json:"cover,omitempty"`
	GameIcon      string         `json:"game_icon,omitempty"`
	GameStatusV2  int32          `json:"game_status_v2,omitempty"`
	GameLink      string         `json:"game_link,omitempty"`
	GradeStatus   int32          `json:"grade_status,omitempty"`
	Grade         float32        `json:"grade,omitempty"`
	BookNum       int64          `json:"book_num,omitempty"`
	GameTags      string         `json:"game_tags,omitempty"`
	DownloadNum   int64          `json:"download_num,omitempty"`
	Notice        string         `json:"notice,omitempty"`
	GameRank      int8           `json:"game_rank,omitempty"`
	RankType      int8           `json:"rank_type,omitempty"`
	GameRankInfo  *GameRankInfo  `json:"rank_info,omitempty"`
	MaterialsInfo *MaterialsInfo `json:"materials_info,omitempty"`
}

type MaterialsInfo struct {
	MaterialsId   int64  `json:"matrials_id,omitempty"`
	ImageURL      string `json:"image_url,omitempty"`
	PromoteStatus int8   `json:"promote_status,omitempty"`
}

type GameRankInfo struct {
	TmDayIconURL   string `json:"tm_day_icon_url,omitempty"`
	TmNightIconURL string `json:"tm_night_icon_url,omitempty"`
	TmIconWidth    int    `json:"tm_icon_width,omitempty"`
	TmIconHeight   int    `json:"tm_icon_height,omitempty"`
}

type GameParam struct {
	GameId     int64 `json:"game_id"`
	CreativeId int64 `json:"creative_id"`
}
