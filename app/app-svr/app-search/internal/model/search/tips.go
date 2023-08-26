package search

// SearchTips is the struct of searchTips from manager search tips API
type SearchTips struct {
	Id       int64  `json:"id"`
	Title    string `json:"title"`               // 主标题
	SubTitle string `json:"sub_title,omitempty"` // 副标题
	Status   int    `json:"status"`              // 状态：2 手动下线；1 上线中；0 未上线
	HasBgImg int    `json:"has_bg_img"`          // 是否有背景图 1有 0无
	JumpUrl  string `json:"jump_url,omitempty"`  // 跳转地址
}
