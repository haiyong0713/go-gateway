package common

type VideoTabV2Item struct {
	Type      int64  `json:"type"`
	Id        int64  `json:"id"`
	Name      string `json:"name"`
	IsDefault bool   `json:"is_default"`
}

type VideoTabV2Resp struct {
	Items    []*VideoTabV2Item `json:"items"`
	Exchange int               `json:"exchange"` // 是否交换fm和视频tab 0: 不交换 1: 交换
}
