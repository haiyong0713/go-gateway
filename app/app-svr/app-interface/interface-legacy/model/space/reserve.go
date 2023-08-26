package space

// AddReserveReq def.
type AddReserveReq struct {
	Sid          int64  `form:"sid" validate:"min=1"`
	Mid          int64  `form:"mid"`
	From         string `form:"from"`
	Type         string `form:"type"`
	Oid          string `form:"oid"`
	Platform     string `form:"platform"`
	Mobiapp      string `form:"mobi_app"`
	Buvid        string `form:"buvid"`
	Spmid        string `form:"spmid"`
	ReserveTotal int64  `form:"reserve_total"`
}

type ReserveClickResp struct {
	ReserveUpdate int64           `json:"reserve_update"`
	DescUpdate    string          `json:"desc_update"`
	CalendarInfos []*CalendarInfo `json:"calendar_infos,omitempty"`
	BusinessIDs   []string        `json:"business_ids,omitempty"`
}

type CalendarInfo struct {
	Title      string `json:"title,omitempty"`
	STime      int64  `json:"s_time,omitempty"`
	ETime      int64  `json:"e_time,omitempty"`
	Comment    string `json:"comment,omitempty"`
	BusinessID string `json:"business_id,omitempty"`
}

type ReserveShareInfoReq struct {
	DynId     int64  `form:"dyn_id" validate:"min=1"`
	ShareId   string `form:"share_id"`
	ShareMode int32  `form:"share_mode"`
}
