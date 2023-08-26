package model

type LpArc struct {
	Archives []*BvArc `json:"archives"`
	Page     *LpPage  `json:"page"`
}

type LpPage struct {
	Count int32 `json:"count"`
	Num   int32 `json:"num"`
	Size  int32 `json:"size"`
}

type LpNewlistReq struct {
	Rid      int64  `form:"rid" validate:"min=0"`
	Pn       int32  `form:"pn" validate:"min=1"`
	Ps       int32  `form:"ps" validate:"min=1,max=50"`
	Business string `form:"business" validate:"required"`
}
