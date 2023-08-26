package abtest

import (
	resource "go-gateway/app/app-svr/resource/service/model"
)

const (
	GrpupDefault = "default"
)

type List struct {
	ID   int64  `json:"group_id,omitempty"`
	Name string `json:"group_name,omitempty"`
}

func (l *List) ListChange(r *resource.AbTest) {
	l.ID = r.ID
	l.Name = r.Name
}

type AbTestListParam struct {
	Platform string `form:"platform"`
	MobiApp  string `form:"mobi_app"`
	Device   string `form:"device"`
	Build    int64  `form:"build"`
	Keys     string `form:"keys"`
}

type AbTestListReply struct {
	List map[string]*AbTestItem `json:"list"`
}

type AbTestItem struct {
	Result string `json:"result"`
}

type TinyAbReply struct {
	PopupStyle    *ABTest            `json:"popup_style,omitempty"`
	UpgradeInform *UpgradeInform     `json:"upgrade_inform,omitempty"`
	ABResult      map[string]*ABTest `json:"ab_result"`
}

type UpgradeInform struct {
	Text   string  `json:"text"`
	Title  string  `json:"title"`
	Timing int64   `json:"timing"`
	ABTest *ABTest `json:"abtest"`
}

type ABTest struct {
	//0-没命中 1-命中
	Exp int64
	//实验组id
	GroupID int64
}
