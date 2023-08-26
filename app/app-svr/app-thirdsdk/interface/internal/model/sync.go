package model

type UserBindParam struct {
	Platform   string `form:"platform" validate:"required"`
	BOpenID    string `form:"bOpenId"`
	OOpenID    string `form:"oOpenId"`
	Action     string `form:"action"`
	ActionTime string `form:"actionTime"`
	ActionMsg  string `form:"actionMsg"`
}

type ArcStatusParam struct {
	Platform   string `form:"platform" validate:"required"`
	Bvid       string `form:"bvid"`
	Ovid       string `form:"ovid"`
	Status     string `form:"status"`
	StatusTime string `form:"statusTime"`
	StatusMsg  string `form:"statusMsg"`
}
