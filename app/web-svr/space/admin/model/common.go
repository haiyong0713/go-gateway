package model

const (
	//LogBlacklist blacklist action log type id
	LogBlacklist = 1
	LogWhitelist = 2
	//LogExamine log examine
	LogExamine   = 2
	StatusOnline = 1
	Deleted      = 1
	NotDeleted   = 2
	NotDelete    = 0
	// clear msg type
	ClearTypeArc     = 1
	ClearTypeMp      = 2
	ClearTypeChName  = 3
	ClearTypeChIntro = 4
)

var ManagerLogType = map[int]int{
	ClearTypeArc:     3,
	ClearTypeMp:      4,
	ClearTypeChName:  5,
	ClearTypeChIntro: 6,
}

var ClearMsgReasons = map[int]string{
	1:  "发布赌博诈骗信息",
	2:  "发布违禁相关信息",
	3:  "发布垃圾广告信息",
	4:  "发布人身攻击信息",
	5:  "发布侵犯他人隐私信息",
	6:  "发布色情信息",
	7:  "发布低俗信息",
	8:  "发布非法网站信息",
	9:  "发布传播不实信息",
	10: "内容不适宜",
}

// Page pager
type Page struct {
	Num   int `json:"num"`
	Size  int `json:"size"`
	Total int `json:"total"`
}
