package model

const (
	GoodsStateNormal   = 1
	GoodsStateAddress  = 2
	GoodsStateShipping = 3

	ActPlatActivity  = "fashion_618"
	ActPlatCounter   = "view"
	ActPlatMidFilter = "mid_filter"
	ActPlatCidFilter = "cid_filter"
)

type Prize struct {
	Id      int64  `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Picture string `json:"picture"`
	Stock   int64  `json:"stock"` //剩余库存
	Score   int64  `json:"score"`
	Want    int64  `json:"want"`
}

type SelectPrize struct {
	InitScore int64 `json:"init_score"`
}

type AddRes struct {
	Score int64 `json:"score"`
}
