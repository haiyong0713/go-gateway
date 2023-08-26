package article

import "go-common/library/queue/databus/report"

const (
	DelFromUnknow = 0
	DelFromAdmin  = 1
	DelFromUser   = 2
	DelFromUp     = 3
	DelFromAssit  = 4
	DelFromFilter = 5
	DelFromGovern = 6
)

type ReplyDelMsg struct {
	Oid     int64 `json:"oid"`
	Type    int64 `json:"type"`
	Mid     int64 `json:"mid"`
	RpID    int64 `json:"rpid"`
	DelFrom int64 `json:"del_from"`
	Extra   struct {
		UserInfo    *report.UserInfo    `json:"user_info"`
		ManagerInfo *report.ManagerInfo `json:"manager_info"`
		Reason      string              `json:"reason"`
	} `json:"extra,omitempty"`
}
