package search

type ComicInfo struct {
	ID              int64        `json:"id"`               // 如果传入的漫画id没有对应数据，2，3，4，5都为空值
	Title           string       `json:"title"`            // 漫画标题
	Author          []string     `json:"author"`           // 漫画作者列表
	Evaluate        string       `json:"evaluate"`         // 漫画简介
	VerticalCover   string       `json:"vertical_cover"`   // 竖版封面
	IsFinish        int8         `json:"is_finish"`        // 完结状态 1:完结 0:连载 -1:未开刊
	Total           int64        `json:"total"`            // 总话数（全x话）
	LastShortTitle  string       `json:"last_short_title"` // 最新话短标题
	LastUpdateTime  string       `json:"last_update_time"` // 最新话更新时间: 秒级时间戳, 当更新时间不存在时, 最新话更新时间为0
	Url             string       `json:"url"`              // h5 跳转url
	PCUrl           string       `json:"pc_url"`           // pc 跳转链接
	FavStatus       int8         `json:"fav_status"`       // 用户是否追漫，0 未追；1 已追
	HorizontalCover string       `json:"horizontal_cover"` // 横版封面
	Status          int8         `json:"status"`           // 状态，-1:下线 0:正常 1:删除 2:定时发布
	Introduction    string       `json:"introduction"`     // 一句话简介
	LastOrdStr      string       `json:"last_ord_str"`     // 最新话文本
	Styles          []ComicStyle `json:"styles"`           // 风格标签
	ComicType       int8         `json:"comic_type"`       // 漫画类型 0一普通漫 1-有声漫
}

type ComicStyle struct {
	Id   int64
	Name string
}
