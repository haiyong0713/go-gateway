package common

// ChannelInfosReq 频道基本信息
type ChannelInfosReq struct {
	Fm    []int64 // FM 频道ID
	Video []int64 // 视频 频道ID
}

type ChannelInfosResp struct {
	Fm    map[int64]*ChannelInfo // FM频道 key: 频道ID
	Video map[int64]*ChannelInfo // 视频频道 key: 频道ID
}

type ChannelInfo struct {
	Title    string // 频道标题
	Cover    string // 封面图
	SubTitle string // 副标题
	HotRate  int64  // 热度值
	Count    int64  // 稿件数量
}

// ChannelArcsReq 频道内部稿件信息（分页）
type ChannelArcsReq struct {
	Fm    []*ChannelArcReq // FM频道
	Video []*ChannelArcReq // TODO 视频频道
}

type ChannelArcReq struct {
	ChanId   int64            // 频道ID（注意区分渠道ID）
	PageNext *ChannelPageInfo // 下一页，首次请求传nil，代表从头开始
	Ps       int              // 全局分页大小，优先级高于PageInfo内Ps
}

type ChannelArcsResp struct {
	Fm    map[int64]*ChannelArcs // FM频道 key: 频道ID
	Video map[int64]*ChannelArcs // TODO 视频频道 key: 频道ID
}

type ChannelArcs struct {
	Aids     []int64
	PageNext *ChannelPageInfo
	HasNext  bool
}

type ChannelPageInfo struct {
	Ps int `json:"ps,omitempty"`
	Pn int `json:"pn,omitempty"`
}
