package api

// all const
const (
	ActionDel           = "delete"
	ActionUpdate        = "update"
	StatusForNormal     = 1
	StatusForDelete     = 0
	TypePrecious        = 1 //入站必刷
	TypeWeeklySelection = 2 //每周必看
	TypeRank            = 3 //排行榜
	TypeHot             = 4 //热门
	TypeChannel         = 5 //精选频道
	RankURL             = "bilibili://rank?type=all"
	HotURL              = "bilibili://home?tab_name=%e7%83%ad%e9%97%a8"
	HotDesc             = "热门"
	FailList            = "honor_fail_list"
	//热门私信相关
	HotSenderUID  = uint64(412466388) //发送人：热门菌
	HotMsgTp      = int32(10)
	HotNotifyCode = "80_0"
)

// ValidType is
var ValidType = map[int32]struct{}{
	TypePrecious:        {},
	TypeWeeklySelection: {},
	TypeRank:            {},
	TypeHot:             {},
}

var TypeOrder = []int32{TypePrecious, TypeWeeklySelection, TypeRank, TypeHot} //优先级 1入站必刷>2每周必看>3历史最高排名>4热门

// HonorMsg message
type HonorMsg struct {
	Aid    int64  `json:"aid"`
	Action string `json:"action"`
	Type   int32  `json:"type"`
	URL    string `json:"url"`
	Desc   string `json:"desc"`
	NaUrl  string `json:"na_url"`
}

// StatMsg message
type StatMsg struct {
	Type      string `json:"type"`
	Aid       int64  `json:"id"`
	Count     int    `json:"count"`
	TimeStamp int64  `json:"timestamp"`
}

// RetryInfo retry data
type RetryInfo struct {
	Action string `json:"action"`
	Data   struct {
		Aid   int64  `json:"aid"`
		Type  int32  `json:"type"`
		URL   string `json:"url"`
		Desc  string `json:"desc"`
		NaUrl string `json:"na_url"`
	} `json:"data"`
}
