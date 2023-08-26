package model

type Page struct {
	Num   int64 `json:"num"`
	Size  int64 `json:"size"`
	Total int64 `json:"total"`
}

type ListLogReply struct {
	Result []*List `json:"result"`
	Page   *Page   `json:"page"`
}

type List struct {
	Action    string `json:"action"`
	Business  int64  `json:"business"`
	Ctime     string `json:"ctime"`
	ExtraData string `json:"extra_data"`
	Str0      string `json:"str_0"`
	Str1      string `json:"str_1"`
	Str2      string `json:"str_2"`
	Str3      string `json:"str_3"`
	Str4      string `json:"str_4"`
	Str5      string `json:"str_5"`
	Oid       int64  `json:"oid"`
	Type      int64  `json:"type"`
	Uid       int64  `json:"uid"`
	Uname     string `json:"uname"`
}

type Extra struct {
	Ctime    int64  `json:"ctime"`
	Mtime    int64  `json:"mtime"`
	Level    string `json:"level"`
	Result   string `json:"result"`
	Detail   string `json:"detail"`
	Category string `json:"category"`
}
