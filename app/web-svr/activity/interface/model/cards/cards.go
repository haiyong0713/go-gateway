package cards

// CardMid ...
type CardMid struct {
	ActivityID int64 `json:"activity_id"`
	MID        int64 `json:"mid"`
	CardID     int64 `json:"card_id"`
	Nums       int64 `json:"nums"`
	Used       int64 `json:"used"`
}

// CardMidRes ...
type CardMidRes struct {
	CardID int64 `json:"card_id"`
	Nums   int64 `json:"nums"`
}

// CardsReplyNew ...
type CardsReplyNew struct {
	Cards      []*CardMidRes `json:"cards"`
	CanCompose bool          `json:"can_compose"`
}

type CardsComposeMessage struct {
	MID        int64  `json:"mid"`
	Timestamp  int64  `json:"timestamp"`
	Nums       int64  `json:"nums"`
	Activity   string `json:"activity"`
	ActivityID int64  `json:"activity_id"`
}
