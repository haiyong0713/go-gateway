package dynamicV2

type ButtomFeedInfo struct {
	BottomDetails []*BottomDetail `json:"bottom_details"`
}

type BottomDetail struct {
	Type    int    `json:"type"`
	Rid     int64  `json:"rid"`
	Content string `json:"content"`
	JumpURL string `json:"jump_url"`
	Status  int    `json:"status"`
}
