package dynamic

type SourceItem struct {
	//类型，1：ugc视频，2：pgc，3：专栏，8：播单
	Type   int   `json:"type"`
	ItemID int64 `json:"item_id"`
	FID    int64 `json:"fid"`
}

type SourceReply struct {
	Offset   int64         `json:"offset"`
	HasMore  int32         `json:"has_more"`
	ItemList []*SourceItem `json:"item_list"`
}

type ProductItem struct {
	ItemID   int64  `json:"item_id"`
	Title    string `json:"title"`
	ImageURL string `json:"image_url"`
	LinkURL  string `json:"link_url"`
}

type ProductReply struct {
	Offset   int64          `json:"offset"`
	HasMore  int32          `json:"has_more"`
	ItemList []*ProductItem `json:"item_list"`
}
