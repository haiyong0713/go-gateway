package frontpage

const DefaultTimeLayout = "2006-01-02 15:04:05"

// Pager 分页数据
type Pager struct {
	Size  int64 `json:"size"`
	Num   int64 `json:"num"`
	Total int64 `json:"total"`
}
