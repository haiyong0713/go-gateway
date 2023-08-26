package rewards

type AsyncSendingAwardInfo struct {
	Mid      int64  `json:"mid" form:"mid" validate:"min=1"`
	UniqueId string `json:"unique_id" form:"unique_id" validate:"required"`
	Business string `json:"business" form:"business" validate:"required"`
	AwardId  int64  `json:"award_id" form:"award_id" validate:"min=1"`
	SendTime int64  `json:"send_time"`
}
