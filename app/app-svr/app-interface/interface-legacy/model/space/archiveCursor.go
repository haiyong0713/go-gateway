package space

import (
	"go-gateway/app/app-svr/archive/service/api"
)

type ArchiveCursorParam struct {
	Vmid          int64  `form:"vmid" validate:"min=1"`
	Aid           int64  `form:"aid"`
	FromViewAid   int64  `form:"from_view_aid"`
	Sort          string `form:"sort"`
	Order         string `form:"order"`
	Ps            int    `form:"ps"`
	SLocaleP      string `form:"s_locale"`
	CLocaleP      string `form:"c_locale"`
	IncludeCursor bool   `form:"include_cursor"`
}

type ArcPlayerCursor struct {
	// 稿件信息
	*api.ArcPlayer
	// 游标信息
	CursorAttr *CursorAttr
}

type ArcPlayerCursorExtra struct {
	// 稿件总数
	Total int64
	// 是否存在下一刷
	HasMore bool
	// 第一刷是否可以展示定位符，具体显示逻辑客户端控制
	CanDisplay bool
}

type ArcCursorList struct {
	EpisodicButton *EpisodicButton `json:"episodic_button,omitempty"`
	Order          []*ArcOrder     `json:"order,omitempty"`
	Count          int64           `json:"count"`
	Item           []*ArcItem      `json:"item"`
	// 上次观看定位符
	LastWatchedLocator *LastWatchedLocator `json:"last_watched_locator,omitempty"`
	// 存在下一刷视频数据
	HasNext bool `json:"has_next"`
	// 存在上一刷视频数据
	HasPrev bool `json:"has_prev"`
}

type LastWatchedLocator struct {
	// 定位符触发门槛配置
	DisplayThreshold int `json:"display_threshold,omitempty"`
	// 定位符插入位置配置
	InsertRanking int `json:"insert_ranking,omitempty"`
	// 定位符展示文字
	Text string `json:"text,omitempty"`
	// 是否能够展示定位符
	CanDisplay bool `json:"can_display,omitempty"`
}

type CursorAttr struct {
	// 标识上次观看稿件
	IsLastWatchedArc bool `json:"is_last_watched_arc"`
	// 游标序列排序位置
	Rank int64 `json:"rank"`
}
