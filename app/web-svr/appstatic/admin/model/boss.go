package model

// CdnDoPreload .
type CdnPreheat struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    *struct {
		TaskID int `json:"task_id"`
	} `json:"data"`
}

// CdnKsyun .
type CdnKsyun struct {
	URL      string  `json:"url"`
	Progress float64 `json:"progress"`
	Status   string  `json:"status"`
}

// CdnPreloadResult .
type CdnPreloadResult struct {
	Ksyun []*CdnKsyun `json:"ksyun"`
}
