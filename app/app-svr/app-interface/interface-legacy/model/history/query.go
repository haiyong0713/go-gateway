package history

// 查询是否存在记录请求参数
type SearchQuery struct {
	Mid   int64
	Buvid string
	// 业务名
	Businesses []string
	// 搜索内容
	Keyword string
}
