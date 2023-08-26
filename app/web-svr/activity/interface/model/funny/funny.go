package funny

// 首页用户信息 两个数组 和 是否增加过抽奖机会
type PageInfoReply struct {
	Task1 int `json:"task_1"`
	Task2 int `json:"task_2"`
}

// LikesReply 点赞数
type LikesReply struct {
	Likes  int64 `json:"likes"`
	Status int   `json:"status"`
}

// AddTimesReply 增加抽奖次数返回
type AddTimesReply struct {
}
