package tianma

// Pager 分页数据
type Pager struct {
	Ps    int   `json:"ps"`
	Pn    int   `json:"pn"`
	Total int64 `json:"total"`
}
