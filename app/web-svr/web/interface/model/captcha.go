package model

type SendCaptchaReq struct {
	Mobile int64 `json:"mobile" form:"mobile" validate:"required"`
}
