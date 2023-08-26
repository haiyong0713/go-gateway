package resource

import (
	"go-gateway/app/app-svr/app-feed/admin/model/common"
)

// SearchLogResult is.
type SearchLogResult struct {
	Order  string      `json:"order"`
	Sort   string      `json:"sort"`
	Result []AuditLog  `json:"result"`
	Page   common.Page `json:"page"`
}

// AuditLog is.
type AuditLog struct {
	UID    int64  `json:"uid"`
	Uname  string `json:"uname"`
	OID    int64  `json:"oid"`
	Type   int8   `json:"type"`
	Action string `json:"action"`
	Str0   string `json:"str_0"`
	Str1   string `json:"str_1"`
	Str2   string `json:"str_2"`
	Int0   int    `json:"int_0"`
	Int1   int    `json:"int_1"`
	Int2   int    `json:"int_2"`
	Ctime  string `json:"ctime"`
	Extra  string `json:"extra_data"`
}
