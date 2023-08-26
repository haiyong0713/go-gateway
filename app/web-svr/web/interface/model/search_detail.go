package model

type TipDetail struct {
	ID       int64  `json:"id"`
	Title    string `json:"title"`      // 主标题
	SubTitle string `json:"sub_title"`  // 副标题
	HasBgImg int    `json:"has_bg_img"` // 是否有背景图 1有 0无
	JumpUrl  string `json:"jump_url"`   // 跳转地址
}

type SystemNotice struct {
	Mid        int64  `json:"mid"`
	NoticeID   int64  `json:"notice_id"`
	Content    string `json:"content"`
	URL        string `json:"url"`
	NoticeType int64  `json:"notice_type"`
	Icon       string `json:"icon"`
	TextColor  string `json:"text_color"`
	BGColor    string `json:"bg_color"`
}

func (in *SystemNotice) Construct() {
	const (
		// 原样式icon
		_prInfoOldIcon = "https://i0.hdslb.com/bfs/space/7a89f7ed04b98458b23863846bd2539a90ff1153.png"
		// 缅怀提示日间icon
		_prInfoNewIcon = "https://i0.hdslb.com/bfs/space/ca6d0ed2edae23cf348db19cd2c293f2121c1b59.png"
		// 缅怀样式背景色
		_prInfoNewBgColor = "#F1F2F3"
		// 缅怀样式文字色
		_prInfoNewTextcolor = "#9499A0"
		// 原样式背景色
		_prInfoOldBgColor = "#FFF6E4"
		// 原样式文字色
		_prInfoOldTextcolor = "#FFB027"
	)
	//nolint:gomnd
	if in.NoticeType == 1 {
		in.Icon = _prInfoOldIcon
		in.BGColor = _prInfoOldBgColor
		in.TextColor = _prInfoOldTextcolor
	} else if in.NoticeType == 2 {
		in.Icon = _prInfoNewIcon
		in.BGColor = _prInfoNewBgColor
		in.TextColor = _prInfoNewTextcolor
	}
}
