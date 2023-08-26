package search

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
