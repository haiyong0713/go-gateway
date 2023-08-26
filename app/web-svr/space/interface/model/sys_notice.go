package model

type SysNotice struct {
	ID      int64  `form:"id" json:"id"`
	Content string `form:"content" json:"content"`
	Url     string `form:"url" json:"url"`
	// 公告配置类型，1-其他类型，2-去世公告
	NoticeType int `form:"notice_type" json:"notice_type"`
	// 提示条
	Icon string `json:"icon,omitempty"`
	// 文字色
	TextColor string `json:"text_color,omitempty"`
	// 背景色
	BgColor string `json:"bg_color,omitempty"`
}

type SysNoticeUid struct {
	SystemNoticeId int64 `form:"system_notice_id" json:"system_notice_id"`
	Uid            int64 `form:"uid" json:"uid"`
}
