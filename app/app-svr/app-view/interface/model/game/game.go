package game

type Info struct {
	GameBaseID  int64   `json:"game_base_id,omitempty"`
	IsOnline    bool    `json:"is_online,omitempty"`
	GameName    string  `json:"game_name,omitempty"`
	GameIcon    string  `json:"game_icon,omitempty"`
	GameStatus  int     `json:"game_status,omitempty"`
	GameLink    string  `json:"game_link,omitempty"`
	GradeStatus int     `json:"grade_status,omitempty"`
	Grade       float64 `json:"grade,omitempty"`
	BookNum     int64   `json:"book_num,omitempty"`
	GameTags    string  `json:"game_tags,omitempty"`
	NoticeTitle string  `json:"notice_title,omitempty"`
	Notice      string  `json:"notice,omitempty"`
	GiftTitle   string  `json:"gift_title,omitempty"`
	GiftURL     string  `json:"gift_url,omitempty"`
}

type Game struct {
	GameBaseID   int64     `json:"game_base_id,omitempty"`   //游戏唯一标识ID
	IsOnline     bool      `json:"is_online,omitempty"`      //是否上架：true 上架，false 下架
	GameName     string    `json:"game_name,omitempty"`      //游戏名称
	Cover        string    `json:"cover,omitempty"`          //卡片封面
	GameIcon     string    `json:"game_icon,omitempty"`      //游戏图标
	GameStatusV2 int32     `json:"game_status_v2,omitempty"` //游戏状态：0查看、1预约、2下载、3立即玩（小游戏）、4付费、5外链
	GameLink     string    `json:"game_link,omitempty"`      //游戏跳转链接
	GradeStatus  int32     `json:"grade_status,omitempty"`   //评分状态：0 无评分，1 评分过少，2 评分正常
	Grade        float32   `json:"grade,omitempty"`          //游戏评分（10分制）
	BookNum      int64     `json:"book_num,omitempty"`       //预约人数（双端预约人数总和，游戏状态为1、2、4时特有字段）
	GameTags     string    `json:"game_tags,omitempty"`      //标签,"/"分割
	DownloadNum  int64     `json:"download_num,omitempty"`   //下载量（游戏状态为0、3、4时特有字段）
	Notice       string    `json:"notice,omitempty"`         //公告（游戏中心小标题）
	NoticeTitle  string    `json:"notice_title,omitempty"`   //公共别名
	GiftTitle    string    `json:"gift_title,omitempty"`     //礼包名称
	GiftURL      string    `json:"gift_url,omitempty"`       //礼包链接
	GameRank     int8      `json:"game_rank,omitempty"`      //游戏榜单排名，只给前十
	RankType     int8      `json:"rank_type,omitempty"`      //游榜单类型，1热度榜，5预约榜，6新游榜
	RankInfo     *RankInfo `json:"rank_info,omitempty"`
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
