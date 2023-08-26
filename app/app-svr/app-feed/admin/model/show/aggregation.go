package show

const (
	NoAuditing = 0 // 未审核

	AggregationNew  = 1
	AggregationDown = 2
	AggregationUp   = 3

	AggregationActiveOnline  = 1
	AggregationActiveOffline = 2
)

// AIAggregation def.
type AIAggregation struct {
	TagID int64 `json:"tag_id"`
}

// AIAggregationItems def.
type AIAggregationItems struct {
	List []*AIAggregation `json:"list"`
}

// Aggregation def.
type Aggregation struct {
	ID          int64  `json:"id"`
	Plat        string `json:"plat"`
	HotTitle    string `json:"hot_title"`
	State       int    `json:"state"`
	Image       string `json:"image"`
	Title       string `json:"title"`
	Subtitle    string `json:"subtitle"`
	ActiveState int    `json:"active_state"`
}

// HotWordDatabus .
type HotWordDatabus struct {
	Old *Aggregation `json:"old"`
	New *Aggregation `json:"new"`
}

type AggregationMsg struct {
	SpiderType string             `json:"spider_type"` // 抓取类型
	Timestamp  int64              `json:"timestamp"`   // 发送databus的时间
	RankData   []*AggregationItem `json:"rank_data"`   // 榜单内容
}

type AggregationItem struct {
	Platform string `json:"platform"`  // 平台 weibo/douyin/zhihu/acfun/bilibili/kuaishou/xigua
	DataType int    `json:"data_type"` // 榜单类型
	// acfun -（榜单类型: 1 香蕉榜 2 热搜榜）
	// bilibili -（榜单类型: 1 热搜榜）
	// douyin - (榜单类型: 1 热点榜)
	// weibo - (榜单类型: 1 热搜榜-实时热点 2 热搜榜-实时上升热点 3 话题榜
	// zhihu - (榜单类型: 1热搜榜)
	RankURL          string `json:"rank_url"`           // 抓取页面的链接
	SourceURL        string `json:"source_url"`         // 抓取的数据源链接
	SourceCreateTime int64  `json:"source_create_time"` // 抓取的数据源创建时间
	Title            string `json:"title"`              // 榜单内容的标题
	Rank             int    `json:"rank"`               // 榜单名次
	HotFactor        int    `json:"hot_factor"`         // 热度值
	Play             int    `json:"play"`               // 阅读量 or 播放量
	Discuss          int    `json:"discuss"`            // 讨论数 or 评论数
	Answer           int    `json:"answer"`             // 答题数
	Follow           int    `json:"follow"`             // 关注数
	Banana           int    `json:"banana"`             // 香蕉数
	IsTop            int    `json:"is_top"`             // 是否置顶
	HotSearchType    string `json:"hot_search_type"`    // 上榜类型
	SpiderDate       int    `json:"spider_date"`        // 爬虫抓取时间
	Hash64           int64  `json:"hash64"`             // hash散列的值
	// 二次处理的值
	RankType    int   `json:"rank_type"`    // 排名类型
	RankValue   int   `json:"rank_value"`   // 排名值
	DatabusTime int64 `json:"databus_time"` // databus时间(更新时间)
}

type AggAI struct {
	UpCnt    int                 `json:"up_cnt"`
	CardList map[int64]*CardList `json:"card_list"`
}

// CardList AI return .
type CardList struct {
	ID         int64            `json:"id"`
	Goto       string           `json:"goto"`
	FromType   string           `json:"from_type"`
	Desc       string           `json:"desc"`
	CornerMark int8             `json:"corner_mark"`
	CoverGif   string           `json:"cover_gif"`
	Condition  []*CardCondition `json:"condition"`
	Tag        string           `json:"tag"`
	// 补充信息
	TagNames string `json:"tag_names"`
}

// CardCondition .
type CardCondition struct {
	Plat      int8   `json:"plat"`
	Condition string `json:"conditions"`
	Build     int    `json:"build"`
}

type ArcInfo struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Author    string `json:"author"`
	View      int32  `json:"view"`
	ViewSpeed int32  `json:"view_speed"`
}
