package bnj

type page struct {
	Num   int `json:"num"`
	Size  int `json:"size"`
	Total int `json:"total"`
}

type Score2CouponRule struct {
	ID     int64 `json:"id"`
	Score  int64 `json:"score"`
	Coupon int64 `json:"coupon"`
}

// Result .
type Result struct {
	Page   *page            `json:"page"`
	Result []*UserActionLog `json:"result"`
}

// UserActionLog .
type UserActionLog struct {
	Oid   int64  `json:"oid"`
	IP    string `json:"ip"`
	Mid   int64  `json:"mid"`
	Int0  int64  `json:"int_0"`
	Ctime string `json:"ctime"`
	Extra string `json:"extra_data"`
}

// PendantCheck pendant upgrade check
type PendantCheck struct {
	SubCheck  bool `json:"sub_check"`
	LiveCheck bool `json:"live_check"`
	LikeCheck bool `json:"like_check"`
}
