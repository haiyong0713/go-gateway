package appstore

import (
	"time"
)

const (
	StateUserIsReceived  int64 = 1
	StateUserNotReceived int64 = 2

	StateAppstoreEnd int64 = 2
)

const (
	MatchkindFingerprint      int64 = 1
	MatchkindLocalFingerprint int64 = 2
	MatchkindLocalBuvid       int64 = 3
)

const (
	RiskName                  = "appstore_main_activity"
	PassportNotFoundUserByTel = -626
	VipBatchNotEnoughErr      = 69006 // 资源池数量不足
	SilenceForbid             = 1
)

// AppStoreStateArg .
type AppStoreStateArg struct {
	ModelName        string `form:"model_name" validate:"required"`
	Fingerprint      string `form:"fingerprint"`
	LocalFingerprint string `form:"local_fingerprint"`
	Buvid            string `form:"buvid"`
	MID              int64
}

// AppStoreReceiveArg .
type AppStoreReceiveArg struct {
	ModelName        string `form:"model_name" validate:"required"`
	Fingerprint      string `form:"fingerprint"`
	LocalFingerprint string `form:"local_fingerprint"`
	Buvid            string `form:"buvid"`
	Build            string `form:"build" validate:"required"`
	MID              int64
}

// ActivityAppstore .
type ActivityAppstore struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	ModelName  string    `json:"model_name"`
	BatchToken string    `json:"batch_token"`
	Appkey     string    `json:"appkey"`
	Operator   string    `json:"operator"`
	Remark     string    `json:"remark"`
	Ctime      time.Time `json:"ctime"`
	Mtime      time.Time `json:"mtime"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
	State      int64     `json:"state"`
}

// ActivityAppstoreReceived .
type ActivityAppstoreReceived struct {
	ID               int64     `json:"id"`
	Mid              int64     `json:"mid"`
	TelHash          string    `json:"tel_hash"`
	BatchToken       string    `json:"batch_token"`
	Fingerprint      string    `json:"fingerprint"`
	LocalFingerprint string    `json:"local_fingerprint"`
	Buvid            string    `json:"buvid"`
	MatchLabel       string    `json:"match_label"`
	MatchKind        int64     `json:"match_kind"`
	Build            string    `json:"build"`
	OrderNo          string    `json:"order_no"`
	State            int64     `json:"state"`
	UserIP           []byte    `user_ip`
	Ctime            time.Time `json:"ctime"`
	Mtime            time.Time `json:"mtime"`
}
