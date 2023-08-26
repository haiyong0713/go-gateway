package common

// SerialInfosReq 合集基本信息
type SerialInfosReq struct {
	FmCommonIds []int64 // FM 普通合集
	//FmUpIds     []int64 // FM up主稿件合集
	VideoIds []int64 // 视频合集
}

type SerialInfosResp struct {
	FmCommon map[int64]*SerialInfo // FM 普通合集 key: 合集ID
	//FmUp     map[int64]*SerialInfo // FM up主稿件合集 key: 合集ID
	Video map[int64]*SerialInfo // 视频合集 key: 合集ID
}

type SerialInfo struct {
	Title string `json:"title"` // 合集标题
	Cover string `json:"cover"` // 合集封面
	Count int    `json:"count"` // 合集稿件数量
}

// SerialArcsReq 合集内部稿件（分页）
type SerialArcsReq struct {
	FmCommon []*SerialArcReq // FM 普通合集
	//FmUp     []*SerialArcReq // FM up主稿件合集
	Video []*SerialArcReq // 视频合集
}

type SerialArcReq struct {
	SerialId int64
	SerialPageReq
}

type SerialArcsResp struct {
	FmCommon map[int64]*SerialArcs // FM合集稿件aid（分页）
	//FmUp     map[int64]*SerialArcs // FM up主稿件合集aid（分页）
	Video map[int64]*SerialArcs // 视频合集稿件aid（分页）
}

type SerialArcs struct {
	Aids []int64
	SerialPageResp
}

type SerialPageReq struct {
	PageNext *SerialPageInfo // 下一页，如果无需查下一页，请传nil
	PagePre  *SerialPageInfo // 上一页，如果无需查上一页，请传nil
	Ps       int             // 全局分页大小，优先级高于PageInfo内Ps
}

type SerialPageResp struct {
	PageNext    *SerialPageInfo // 下一页，向下翻页时透传回来
	PagePre     *SerialPageInfo // 上一页，向上翻页时透传回来
	HasNext     bool            // 是否到底
	HasPrevious bool            // 是否到顶
}

type SerialPageInfo struct {
	Ps          int   `json:"ps,omitempty"`           // 分页大小
	Oid         int64 `json:"oid,omitempty"`          // 游标，ugc为aid
	WithCurrent bool  `json:"with_current,omitempty"` // 是否包含当前游标的稿件，默认不包含
}
