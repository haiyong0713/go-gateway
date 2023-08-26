package datamanage

type ReqDataManageSelect struct {
	Conn        string   `json:"_conn" form:"_conn"`
	Table       string   `json:"_table" form:"_table"`
	IgnoreField []string `json:"_ignore_field" form:"_ignore_field,split"`
	Offset      int64    `json:"_offset" form:"_offset"`
	Limit       int64    `json:"_limit" form:"_limit"`
}

type ReqDataManageUpdate struct {
	Conn        string   `json:"_conn" form:"_conn"`
	Table       string   `json:"_table" form:"_table"`
	IgnoreField []string `json:"_ignore_field" form:"_ignore_field,split"`
	Trim        bool     `json:"trim" form:"trim"`
	Primary     string   `json:"primary" form:"primary"`
}
