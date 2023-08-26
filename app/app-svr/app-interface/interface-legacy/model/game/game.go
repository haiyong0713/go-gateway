package game

// PlayGame app game struct.
type PlayGame struct {
	GameBaseID int64   `json:"game_base_id"`
	GameName   string  `json:"game_name"`
	GameIcon   string  `json:"game_icon"`
	Grade      float64 `json:"grade"`
	DetailURL  string  `json:"detail_url"`
}

// RecentGame .
type RecentGame struct {
	List       []*PlayGame `json:"list"`
	TotalCount int         `json:"total_count"`
}

type PlayGameSub struct {
	PlayGame
	GameTags   []string `json:"game_tags"`
	Notice     string   `json:"notice"`
	GiftTitle  string   `json:"gift_title"`
	GameStatus int      `json:"game_status"`
}

type RecentGameSub struct {
	List       []*PlayGameSub `json:"list"`
	TotalCount int            `json:"total_count"`
}

type Game struct {
	GameBaseID  int64     `json:"game_base_id,omitempty"`
	IsOnline    bool      `json:"is_online,omitempty"`
	GameName    string    `json:"game_name,omitempty"`
	Cover       string    `json:"cover,omitempty"`
	GameIcon    string    `json:"game_icon,omitempty"`
	GameStatus  int32     `json:"game_status,omitempty"`
	GameLink    string    `json:"game_link,omitempty"`
	GradeStatus int32     `json:"grade_status,omitempty"`
	Grade       float64   `json:"grade,omitempty"`
	BookNum     int64     `json:"book_num,omitempty"`
	GameTags    string    `json:"game_tags,omitempty"`
	DownloadNum int64     `json:"download_num,omitempty"`
	NoticeTitle string    `json:"notice_title,omitempty"`
	Notice      string    `json:"notice,omitempty"`
	GiftTitle   string    `json:"gift_title,omitempty"`
	GiftUrl     string    `json:"gift_url,omitempty"`
	GameRank    int64     `json:"game_rank,omitempty"`
	RankType    int64     `json:"rank_type,omitempty"`
	RankInfo    *RankInfo `json:"rank_info,omitempty"`
}

type RankInfo struct {
	SearchNightIconUrl   string `json:"search_night_icon_url,omitempty"`
	SearchDayIconUrl     string `json:"search_day_icon_url,omitempty"`
	SearchBkgNightColor  string `json:"search_bkg_night_color,omitempty"`
	SearchBkgDayColor    string `json:"search_bkg_day_color,omitempty"`
	SearchFontNightColor string `json:"search_font_night_color,omitempty"`
	SearchFontDayColor   string `json:"search_font_day_color,omitempty"`
	RankContent          string `json:"rank_content,omitempty"`
	RankLink             string `json:"rank_link,omitempty"`
}

type TopGameConfigButton struct {
	Content string `json:"content,omitempty"`
	Url     string `json:"url,omitempty"`
}

type TopGameButtonInfos struct {
	GameId int64                  `json:"game_id,omitempty"`
	CardId int64                  `json:"card_id,omitempty"`
	Infos  []*TopGameConfigButton `json:"infos,omitempty"`
}

type TopGameConfig struct {
	ButtonInfos []*TopGameButtonInfos `json:"button_infos,omitempty"`
}
