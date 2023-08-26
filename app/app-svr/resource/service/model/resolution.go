package model

type LimitFreeInfo struct {
	Aid       int64  `json:"aid"`
	LimitFree int64  `json:"limit_free"`
	Subtitle  string `json:"subtitle"`
}

type LimitFreeReply struct {
	LimitFreeWithAid map[int64]*LimitFreeInfo `json:"limit_free_with_aid"`
}
