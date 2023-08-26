package channel

const (
	EsStateHide = 0
	EsStateOK   = 2
)

type EsRes struct {
	Code int        `json:"code"`
	Data *EsResData `json:"data"`
}
type EsResData struct {
	Order  string           `json:"order"`
	Sort   string           `json:"sort"`
	Result []*EsResDataItem `json:"result"`
	Page   struct {
		Num   int32 `json:"num"`
		Size  int32 `json:"size"`
		Total int32 `json:"total"`
	} `json:"page"`
}

type EsResDataItem struct {
	ChannelId int64 `json:"cid"`
	State     int32 `json:"state"`
}
