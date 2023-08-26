package apm

// / 聚合表 buvid 聚合数据通用字段
type AggregateCountItem struct {
	Timestamp          int64  `json:"timestamp"`
	AppKey             string `json:"app_key"`
	VersionCode        int64  `json:"version_code"`
	Count              int64  `json:"count"`
	DistinctBuvidCount int64  `json:"distinct_buvid_count"`
}

// / 网络信息聚合表
type AggregateNetInfo struct {
	Timestamp           int64   `json:"timestamp"`
	AppKey              string  `json:"app_key"`
	Command             string  `json:"command"`
	Count               int64   `json:"count"`
	ErrorRate           float64 `json:"error_rate"`
	TotalTimeQuantile80 float64 `json:"total_time_quantile_80"`
	AvgReqSize          float64 `json:"avg_req_size"`
	AvgRecvSize         float64 `json:"avg_recv_size"`
}
