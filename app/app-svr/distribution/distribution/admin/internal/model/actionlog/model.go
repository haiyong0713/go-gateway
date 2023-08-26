package actionlog

type Log struct {
	Business  int64  `form:"business"`
	Mid       int64  `form:"mid"`
	Type      int64  `form:"type" default:"-1"`
	UserName  string `form:"uname"`
	CtimeFrom string `form:"ctime_from"`
	CtimeTo   string `form:"ctime_to"`
	Ps        int64  `form:"ps" default:"20"`
	Pn        int64  `form:"pn" default:"1"`
	Sort      string `form:"sort"`
}

type LogManagerItem struct {
	Mid       int    `json:"mid"`
	Type      int    `json:"type"`
	UserName  string `json:"user_name"`
	CTime     string `json:"ctime"`
	ExtraData string `json:"extra_data"`
	Business  int64  `json:"business"`
}

// LogSearch .
type LogSearch struct {
	Mid       int    `json:"mid"`
	Type      int    `json:"type"`
	UserName  string `json:"str_0"`
	CTime     string `json:"ctime"`
	ExtraData string `json:"extra_data"`
	Business  int64  `json:"business"`
}

// ManagerPage .
type ManagerPage struct {
	CurrentPage int `json:"current_page"`
	TotalItems  int `json:"total_items"`
	PageSize    int `json:"page_size"`
}

// LogManagers .
type LogManagers struct {
	Item  []*LogManagerItem `json:"item"`
	Pager ManagerPage       `json:"pager"`
}
