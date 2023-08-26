package model

type CheckExpressionReq struct {
	MobiApp    string `form:"mobiApp" validate:"required"`
	Device     string `form:"device" validate:"required"`
	Platform   string `form:"platform"`
	Build      int64  `form:"build" validate:"required"`
	Expression string `form:"expression" validate:"required"`
}

type CheckExpressionReply struct {
	Result string `json:"result"`
}
