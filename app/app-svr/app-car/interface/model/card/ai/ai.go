package ai

type Item struct {
	ID       int64  `json:"id,omitempty"`
	ChildID  int64  `json:"child_id,omitempty"`
	TrackID  string `json:"trackid,omitempty"`
	Goto     string `json:"goto,omitempty"`
	Position int    `json:"position,omitempty"`
	// ext
	Entrance string `json:"-"`
	DynCtime int64  `json:"-"`
	// 原始值
	Card interface{} `json:"-"`
}
