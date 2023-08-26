package model

import "time"

const (
	BusinessIDBatch      = 830
	BusinessIDAuthor     = 831
	BusinessIDAuthorPush = 832
)

type AuditLogSearchParams struct {
	UName    string `json:"uname"`      // 审核人员内网name 多个用逗号分隔，如aa,bb
	UID      int64  `json:"uid"`        // 审核人员内网uid 多个用逗号分隔，如11,22
	Business int    `json:"business"`   // 业务id，比如稿件业务 多个用逗号分隔，如11,22
	Type     int    `json:"type"`       // 操作对象的类型，如评论 多个用逗号分隔，如11,22
	OID      int64  `json:"oid"`        // 操作对象的id，如2233 多个用逗号分隔，如11,22
	Action   string `json:"action"`     // 操作对象的id，如aaaa 多个用逗号分隔，如aa,bb
	CTime    string `json:"ctime_from"` // 操作启始时间 如"2006-01-02 15:04:05" 查询范围限制见下
	CTimeTo  string `json:"ctime_to"`   // 操作结束时间 如"2006-01-02 15:04:05" 查询范围限制见下
	Int0     int64  `json:"int_0"`      // 预留int型索引字段0 多个用逗号分隔，如11,22
	Int0From int64  `json:"int_0_from"` // 预留int型索引字段0 gte查询，如1<=int_0 填1
	Int0To   int64  `json:"int_0_to"`   // 预留int型索引字段0 lte查询，如int_0<=2 填2
	Int1     int64  `json:"int_1"`      // 预留int型索引字段1 多个用逗号分隔，如11,22
	Int1From int64  `json:"int_1_from"` // 预留int型索引字段1 gte查询，如1<=int_1 填1
	Int1To   int64  `json:"int_1_to"`   // 预留int型索引字段1 lte查询，如int_1<=2 填2
	Int2     int64  `json:"int_2"`      // 预留int型索引字段2 多个用逗号分隔，如11, 22
	Int2From int64  `json:"int_2_from"` // 预留int型索引字段2 gte查询，如1<=int_2 填1
	Int2To   int64  `json:"int_2_to"`   // 预留int型索引字段2 lte查询，如int_2<=2 填2
	Str0     string `json:"str_0"`      // 预留str型索引字段0 多个用逗号分隔，如aa, bb
	Str1     string `json:"str_1"`      // 预留str型索引字段1 多个用逗号分隔，如aa, bb
	Str2     string `json:"str_2"`      // 预留str型索引字段2 多个用逗号分隔，如aa, bb
	Order    string `json:"order"`      // 排序
	PN       int    `json:"pn"`         // 当前页码
	PS       int    `json:"ps"`         // 单页返回数量 最大值1000
}

type AuditLogInitParams struct {
	UName    string        `json:"uname"`      // 审核人员内网name 多个用逗号分隔，如aa,bb
	UID      int64         `json:"uid"`        // 审核人员内网uid 多个用逗号分隔，如11,22
	Business int           `json:"business"`   // 业务id，比如稿件业务 多个用逗号分隔，如11,22
	Type     int           `json:"type"`       // 操作对象的类型，如评论 多个用逗号分隔，如11,22
	OID      int64         `json:"oid"`        // 操作对象的id，如2233 多个用逗号分隔，如11,22
	Action   string        `json:"action"`     // 操作对象的id，如aaaa 多个用逗号分隔，如aa,bb
	CTime    time.Time     `json:"ctime_from"` // 操作启始时间 如"2006-01-02 15:04:05" 查询范围限制见下
	Index    []interface{} `json:"index"`
	Content  interface{}   `json:"content"`
}

type AuditLogSearchResRaw struct {
	Code    int                       `json:"code"`
	Message string                    `json:"message"`
	Data    *AuditLogSearchResRawData `json:"data"`
}

type AuditLogSearchResRawData struct {
	Order  string               `json:"order"`
	Sort   string               `json:"sort"`
	Result []*AuditLogSearchRes `json:"result"`
	Page   *Page                `json:"page"`
}

type AuditLogSearchRes struct {
	Action    string `json:"action"`
	Business  int    `json:"business"`
	CTime     string `json:"ctime"`
	ExtraData string `json:"extra_data"`
	Str0      string `json:"str_0"`
	Str1      string `json:"str_1"`
	Str2      string `json:"str_2"`
	Str3      string `json:"str_3"`
	Str4      string `json:"str_4"`
	Str5      string `json:"str_5"`
	Int0      int64  `json:"int_0"`
	Int1      int64  `json:"int_1"`
	Int2      int64  `json:"int_2"`
	Int3      int64  `json:"int_3"`
	Int4      int64  `json:"int_4"`
	OID       int64  `json:"oid"`
	Type      int    `json:"type"`
	UID       int    `json:"uid"`
	UName     string `json:"uname"`
}
