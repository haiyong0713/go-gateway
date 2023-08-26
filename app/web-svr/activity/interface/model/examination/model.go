package examination

type UpInfo struct {
	Account *Account `json:"account"`
	Live    *Live    `json:"live"`
	Reserve *Reserve `json:"reserve"`
}

// Live
type Live struct {
	LiveStatus int64  `json:"live_status"`
	Title      string `json:"title"`
}

type Reserve struct {
	IsFollowing bool  `json:"is_following"`
	SID         int64 `json:"sid"`
}

// Account ...
type Account struct {
	MID  int64  `json:"mid"`
	Name string `json:"name"`
	Sex  string `json:"sex"`
	Face string `json:"face"`
	Sign string `json:"sign"`
}

// UpRes
type UpRes struct {
	TimeStamp int64     `json:"timestamp"`
	UpOther   []*UpInfo `json:"up_other"`
	UpToday   []*UpInfo `json:"up_today"`
}

// Ups ...
type Ups struct {
	MID int64 `json:"uid" validate:"required"`
	SID int64 `json:"sid" validate:"required"`
}

// UpReq
type UpReq struct {
	UpToday []*Ups `json:"up_today"`
	UpOther []*Ups `json:"up_other"`
}
