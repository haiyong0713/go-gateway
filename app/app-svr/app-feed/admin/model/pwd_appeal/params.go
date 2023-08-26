package pwd_appeal

type ListReq struct {
	Mid         int64  `json:"mid" form:"mid"`
	DeviceToken string `json:"device_token" form:"device_token"`
	State       int64  `json:"state" form:"state"`
	Mode        int64  `json:"mode" form:"mode"`
	BeginTime   string `json:"begin_time" form:"begin_time"`
	EndTime     string `json:"end_time" form:"end_time"`
	Pn          int64  `json:"pn" form:"pn" default:"1"`
	Ps          int64  `json:"ps" form:"ps" default:"10" validate:"max=20"`
}

type ListRly struct {
	List []*PwdAppeal `json:"list"`
	Page *Page        `json:"page"`
}

type Page struct {
	Num   int64 `json:"num"`
	Size  int64 `json:"size"`
	Total int64 `json:"total"`
}

type PhotoReq struct {
	UploadKey string `json:"upload_key" form:"upload_key" validate:"required"`
}

type PassReq struct {
	ID  int64  `json:"id" form:"id" validate:"required"`
	Pwd string `json:"pwd" form:"pwd" validate:"required"`
}

type RejectReq struct {
	ID     int64  `json:"id" form:"id" validate:"required"`
	Reason string `json:"reason" form:"reason" validate:"required"`
}

type ExportReq struct {
	ListReq
}
