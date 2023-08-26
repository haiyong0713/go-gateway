package model

type MatchModel struct {
	ID int64 `json:"id"`
	// 赛事简称
	Title string `json:"title"`
	// 赛事全称
	SubTitle string `json:"sub_title"`
	// 创建年份
	CYear string `json:"c_year"`
	// 主办方
	Sponsor string `json:"sponsor"`
	// logo
	Logo string `json:"logo"`
	// 备注
	Dic string `json:"dic"`
	// 状态
	Status int64 `json:"status"`
	// 排序
	Rank int64 `json:"rank"`
}
