package model

type PushParam struct {
	AppID       int64    `json:"app_id"`
	BusinessID  int64    `json:"business_id"`
	AlertTitle  string   `json:"alert_title"` // 非必传
	AlertBody   string   `json:"alert_body"`  // 非必传
	MIDs        []int64  `json:"mids"`
	Buvids      []string `json:"buvids"`
	LinkType    int64    `json:"link_type"`
	LinkValue   string   `json:"link_value"`
	UUID        string   `json:"uuid"`
	PassThrough int      `json:"pass_through"`
}
