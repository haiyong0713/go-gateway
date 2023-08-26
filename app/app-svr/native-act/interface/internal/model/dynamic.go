package model

const (
	// DynType
	DynTypeVideo   = 8
	DynTypeArticle = 64
	// 动态分组类型、顺序
	DynGroupEmpty = 0
	DynGroupTop   = 1
	DynGroupHot   = 2
	DynGroupFeed  = 3
	// Dyn SortBy
	DynSortHot    = 0 //热度
	DynSortTime   = 1 //时间
	DynSortCompre = 2 //综合
	// Dyn From
	DynFromNative = "activity_page"
)

var (
	DynGroupNames = map[int64]string{
		DynGroupTop:  "置顶",
		DynGroupHot:  "热门",
		DynGroupFeed: "最新",
	}
)

type BriefDynsReq struct {
	TopicID int64  `json:"topic_id"`
	From    string `json:"from"`
	Offset  string `json:"offset"`
	Types   string `json:"types"`
	Ps      int64  `json:"ps"`
	Mid     int64  `json:"mid"`
	SortBy  int64  `json:"sort_by"`
}

type BriefDynsRly struct {
	HasMore  int64       `json:"has_more"`
	Offset   string      `json:"offset"`
	Dynamics []*BriefDyn `json:"dynamics"`
}

type BriefDyn struct {
	Rid  int64 `json:"rid"`
	Type int64 `json:"type"`
}

type ActiveUsersReq struct {
	TopicID int64 `json:"topic_id"`
	NoLimit int64 `json:"no_limit"`
}

type ActiveUsersRly struct {
	ViewCount    int64 `json:"view_count"`
	DiscussCount int64 `json:"discuss_count"`
}
