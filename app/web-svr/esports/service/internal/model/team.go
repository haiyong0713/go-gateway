package model

type TeamModel struct {
	ID int64 `json:"id"`
	// 简称
	Title string `json:"title"`
	// 全称
	SubTitle string `json:"sub_title"`
	// 英文全称
	ETitle string `json:"e_title"`
	// 地区
	Area string `json:"area"`
	// 英文全称
	Logo string `json:"logo"`
	// 地区
	UID int64 `json:"uid"`
	// 成员
	Members string `json:"members"`
	// 备注
	Dic string `json:"dic"`
	// 战队视频url
	VideoUrl string `json:"video_url"`
	// 战队简介
	Profile string `json:"profile"`
	// 三方战队id
	LeidaTId int64 `json:"leida_tid"`
	// 评论id
	ReplyId int64 `json:"reply_id"`
	// 战队类型
	TeamType int64 `json:"team_type"`
	// 战队地区
	RegionId int64 `json:"region_id"`
	// 战队头图
	PictureUrl string `json:"picture_url"`
}
