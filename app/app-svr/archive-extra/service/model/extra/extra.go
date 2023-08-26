package extra

type ArchiveExtra struct {
	Id        int64  `json:"id"`
	Aid       int64  `json:"aid"`
	BizType   string `json:"biz_type"`
	BizValue  string `json:"biz_value"`
	IsDeleted int    `json:"is_deleted"`
}
