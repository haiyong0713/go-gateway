package bender

type TopResourceResp struct {
	Code    int           `json:"code"`
	Message string        `json:"message"`
	Data    *ResourceData `json:"data"`
}

type ResourceData struct {
	Resources []*Item `json:"resources"`
}

type Item struct {
	Key     string `json:"key"`
	Url     string `json:"url"`
	Md5     string `json:"md5"`
	Size    int64  `json:"size"`
	Popular int    `json:"popular"`
}
