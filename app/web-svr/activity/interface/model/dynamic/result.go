package dynamic

// LikedReply .
type LikedReply struct {
	Score int64  `json:"score,omitempty"`
	Toast string `json:"toast,omitempty"`
}

// VideoActReply .
type VideoActReply struct {
	List map[string]*Item `json:"list"`
}

// ActReply .
type ActReply struct {
	Items   []*Item `json:"items"`
	Offset  int64   `json:"offset"`
	HasMore int32   `json:"has_more"`
}

// EsLikesReply .
type EsLikesReply struct {
	Lid   int64 `json:"lid"`
	Score int64 `json:"score"`
}
