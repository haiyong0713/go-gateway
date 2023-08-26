package search

type TrendingRankingReq struct {
	Buvid    string `form:"buvid"`
	Build    int64  `form:"build"`
	Limit    int64  `form:"limit" default:"20"`
	MobiApp  string `form:"build"`
	Device   string `form:"device"`
	Platform string `form:"platform"`
}

type TrendingRankingRsp struct {
	Code    int             `json:"code,omitempty"`
	TrackID string          `json:"trackid"`
	List    []*TrendingList `json:"list"`
	ExpStr  string          `json:"exp_str,omitempty"`
}

type TrendingList struct {
	Position        int                      `json:"position,omitempty"`
	Keyword         string                   `json:"keyword"`
	ShowName        string                   `json:"show_name"`
	WordType        int                      `json:"word_type,omitempty"`
	Icon            string                   `json:"icon,omitempty"`
	HotId           int64                    `json:"hot_id,omitempty"`
	URI             string                   `json:"uri,omitempty"`
	Goto            string                   `json:"goto,omitempty"`
	Param           string                   `json:"param,omitempty"`
	ResourceID      int64                    `json:"resource_id,omitempty"`
	ShowLiveIcon    bool                     `json:"show_live_icon,omitempty"`
	HeatValue       int64                    `json:"heat_value,omitempty"`
	ConfigCardItems []*RankingConfigCardItem `json:"config_card_items,omitempty"`
}

type RankingConfigCardItem struct {
	CardType          string       `json:"card_type"`
	Cover             string       `json:"cover"`
	CoverLeftShowDesc string       `json:"cover_left_show_desc"`
	CoverLeftShowImg  string       `json:"cover_left_show_img"`
	Title             string       `json:"title"`
	JumpUrl           string       `json:"jump_url"`
	Param             string       `json:"param"`
	OgvConfigs        *OgvConfigs  `json:"ogv_configs,omitempty"`
	LiveConfigs       *LiveConfigs `json:"live_configs,omitempty"`
}

type OgvConfigs struct {
	TypeName string `json:"type_name"`
}

type LiveConfigs struct {
	ShowLiveIcon bool  `json:"show_live_icon"`
	LiveStatus   int64 `json:"live_status"`
}
