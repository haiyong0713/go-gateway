package es

const (
	ChannelHide = 0
	ChannelOK   = 2
)

type SearchChannel struct {
	CID   int64 `json:"cid"`
	State int   `json:"state"`
}

type ChannelResult struct {
	TrackID       string         `json:"trackid"`
	Pages         int            `json:"pages"`
	Total         int            `json:"total"`
	FaildNum      int            `json:"faild_num"`
	ExpStr        string         `json:"exp_str"`
	Items         []*ChannleItem `json:"items,omitempty"`
	Extend        *ChannleItem2  `json:"extend,omitempty"`
	NoSearchLabel string         `json:"no_search_label,omitempty"`
	NoMoreLabel   string         `json:"no_more_label,omitempty"`
}

type ChannleItem struct {
	ID             int64          `json:"id,omitempty"`
	Title          string         `json:"title,omitempty"`
	Cover          string         `json:"cover,omitempty"`
	URI            string         `json:"uri,omitempty"`
	Param          string         `json:"param,omitempty"`
	Goto           string         `json:"goto,omitempty"`
	IsAtten        int            `json:"is_atten"`
	Label          string         `json:"label,omitempty"`
	Label2         string         `json:"label2,omitempty"`
	TypeIcon       string         `json:"type_icon,omitempty"`
	Right          bool           `json:"-"`
	Icon           string         `json:"icon,omitempty"`
	Button         *SearchButton  `json:"button,omitempty"`
	Items          []*ChannleItem `json:"items,omitempty"`
	CoverLeftText1 string         `json:"cover_left_text_1,omitempty"`
	CoverLeftIcon1 int            `json:"cover_left_icon_1,omitempty"`
	Badge          *ChannelBadge  `json:"badge,omitempty"`
	More           *SearchButton  `json:"more,omitempty"`
	ThemeColor     string         `json:"theme_color,omitempty"`
	Alpha          int32          `json:"alpha,omitempty"`
	// 夜间模式颜色，服务端对明暗度做了调整
	ThemeColorNight string `json:"theme_color_night,omitempty"`
}

type ChannleItem2 struct {
	Label     string         `json:"label"`
	ModelType string         `json:"model_type"`
	Items     []*ChannleItem `json:"items"`
}

type SearchButton struct {
	Text string `json:"text,omitempty"`
	URI  string `json:"uri,omitempty"`
}

type ChannelBadge struct {
	Text      string `json:"text,omitempty"`
	IconBgURL string `json:"icon_bg_url,omitempty"`
}
