package gameholiday

// LikesReply 硬币数
type LikesReply struct {
	Likes int64 `json:"likes"`
	State int   `json:"state"`
}

// AddTimesReply 增加抽奖次数返回
type AddTimesReply struct {
}
