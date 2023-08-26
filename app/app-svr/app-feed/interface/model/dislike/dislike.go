package dislike

type DisklikeReason struct {
	ReasonID   int    `json:"reason_id,omitempty"`
	ReasonName string `json:"reason_name,omitempty"`
}

var FeedbackFromType = map[string]int64{
	"operation": 1, //运营卡
	"recommend": 2, //推荐卡
}
