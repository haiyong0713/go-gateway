package model

type TradeCreateReq struct {
	Mid       int64  `form:"-"`
	ProductId string `form:"product_id"`
	SpmID     string `form:"spm_id"`
}
