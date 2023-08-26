package search

type TopGameMaterial struct {
	TopGameData *TopGameData
	InlineFn    func(i *Item)
}

type TopGameData struct {
	GameBaseId          int64      `json:"game_base_id"`
	GameName            string     `json:"game_name"`
	GameIcon            string     `json:"game_icon"`
	GameLink            string     `json:"game_link"`
	GameStatus          int64      `json:"game_status"`
	GameTags            string     `json:"game_tags"`
	GameOfficialAccount int64      `json:"game_official_account"`
	NoticeTitle         string     `json:"notice_title"` // 公告标题，默认“公告”
	Notice              string     `json:"notice"`       // 公告（游戏中心小标题）
	Grade               float64    `json:"grade"`
	TabInfo             []*TabInfo `json:"tab_info"`
	VideoCoverImage     string     `json:"video_cover_image"`
	BackgroundImage     string     `json:"background_image"`
	CoverDefaultColor   string     `json:"cover_default_color"`
	GaussianBlurValue   string     `json:"gaussian_blur_value"`
	MarkColorValue      string     `json:"mask_color_value"`
	MaskOpacity         string     `json:"mask_opacity"`
	ModuleColor         string     `json:"module_color"`
	ButtonType          int64      `json:"button_type"`
	Avid                int64      `json:"avid"`
	RoomId              int64      `json:"room_id"`
}

type TabInfo struct {
	TabName string `json:"tab_name"`
	TabUrl  string `json:"tab_url"`
	Sort    int64  `json:"sort"`
}

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

type NewGame struct {
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

type TopGameInlineInfo struct {
	InlineInfos []*TopGameConfigInline `json:"inline_infos,omitempty"`
}

type TopGameConfigInline struct {
	GameId int64 `json:"game_id"`
	CardId int64 `json:"card_id"`
	Avid   int64 `json:"avid"`
}
