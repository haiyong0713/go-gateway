package spmode

type SearchReq struct {
	Mid         int64  `form:"mid"`
	DeviceToken string `form:"device_token"`
}

type SearchRly struct {
	List []*SearchItem `json:"list"`
}

type SearchItem struct {
	RelatedKey  string `json:"related_key"`
	Model       int64  `json:"model"`
	Mid         int64  `json:"mid"`
	DeviceToken string `json:"device_token"`
	Password    string `json:"password"`
	Mtime       string `json:"mtime"`
	State       int64  `json:"state"`
	PwdType     int64  `json:"pwd_type"`
}

type RelieveReq struct {
	RelatedKey string `form:"related_key" validate:"required"`
}

type LogReq struct {
	RelatedKey string `form:"related_key" validate:"required"`
}

type LogRly struct {
	List []*LogItem `json:"list"`
}

type LogItem struct {
	Operator string `json:"operator"`
	Ctime    string `json:"ctime"`
	Content  string `json:"content"`
}
