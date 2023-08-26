package model

type SourceDetailReq struct {
	SourceId string `json:"source_id"`
	Offset   int64  `json:"offset"`
	Size     int64  `json:"size"`
}

type SourceDetailRly struct {
	Offset   int64               `json:"offset"`
	HasMore  int64               `json:"has_more"`
	ItemList []*SourceDetailItem `json:"item_list"`
}

type SourceDetailItem struct {
	Type   int64 `json:"type"`
	ItemId int64 `json:"item_id"`
	Fid    int64 `json:"fid"`
}

type ProductDetailReq struct {
	SourceId string `json:"source_id"`
	Offset   int64  `json:"offset"`
	Size     int64  `json:"size"`
}

type ProductDetailRly struct {
	Offset   int64                `json:"offset"`
	HasMore  int64                `json:"has_more"`
	ItemList []*ProductDetailItem `json:"item_list"`
}

type ProductDetailItem struct {
	ItemId   int64  `json:"item_id"`
	Title    string `json:"title"`
	ImageUrl string `json:"image_url"`
	LinkUrl  string `json:"link_url"`
}
