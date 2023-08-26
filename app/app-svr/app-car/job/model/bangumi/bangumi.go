package bangumi

import "go-gateway/app/app-svr/app-car/job/model"

type Content struct {
	OpType          int            `json:"op_type"`
	Mtime           int64          `json:"mtime"`
	Name            string         `json:"name"`
	Alias           string         `json:"alias"`
	Type            int            `json:"type"`
	DisplayAddress  string         `json:"display_address"`
	DownloadAddress string         `json:"download_address"`
	Premieredate    string         `json:"premieredate"`
	CoverImage      string         `json:"cover_image"`
	Duration        int64          `json:"duration"`
	PubTime         string         `json:"pub_time"`
	PubRealTime     int64          `json:"pub_real_time"`
	Copyright       string         `json:"copyright"`
	IsFinish        int            `json:"is_finish"`
	Episodes        []*Episode     `json:"episodes"`
	Season          *Season        `json:"season"`
	SeasonSeries    []*SeasonSerie `json:"season_series"`
	Tag             []*Tag         `json:"tag"`
	PlayCount       int64          `json:"play_count"`
	Country         string         `json:"country"`
	Actors          string         `json:"actors"`
	Akira           string         `json:"akira"`
	Intro           string         `json:"intro"`
	IsPushSeason    bool           //小度媒资推送时是否推送Season类型
}

type Episode struct {
	ID          int64  `json:"id"`
	Index       int    `json:"index"`
	PlayURL     string `json:"play_url"`
	Cover       string `json:"cover"`
	Title       string `json:"title"`
	IndexTitle  string `json:"index_title"`
	Duration    int64  `json:"duration"`
	PubRealTime string `json:"pub_real_time"`
	Status      int    `json:"status"`
}

type Season struct {
	ID            int64   `json:"id"`
	PaymentStatus int     `json:"paymentstatus"`
	Index         int     `json:"index"`
	Title         string  `json:"title"`
	Version       string  `json:"version"`
	TotalCount    int     `json:"total_count"`
	PayPrice      float64 `json:"pay_price"`
}

type SeasonSerie struct {
	ID    int64  `json:"id"`
	Index int    `json:"index"`
	Title string `json:"title"`
}

type Tag struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func (c *Content) ContentTypeString() string {
	// season类型 1：番剧，2：电影，3：纪录片，4：国漫，5：电视剧，7:综艺
	switch c.Type {
	case model.SeasonTypeBangumi:
		return "番剧"
	case model.SeasonTypeMovie:
		return "电影"
	case model.SeasonTypeDocumentary:
		return "纪录片"
	case model.SeasonTypeGc:
		return "国漫"
	case model.SeasonTypeTv:
		return "电视剧"
	case model.SeasonTypeZi:
		return "综艺"
	}
	return ""
}

// 资费信息（付费/免费/vip）
func (e *Episode) EpisodeCost() string {
	switch e.Status {
	// ep付费状态 2-免费 6-付费,大会员免费 7-付费抢先,大会员免费 8-全付费观看 9-全付费抢先  12-霹雳付费 13-仅大会员可看
	case model.EpFree:
		return "免费"
	case model.EpVipFree, model.EpVipFree2, model.EpOnlyVip:
		return "vip"
	default:
		return "付费"
	}
}

// 资费信息（付费/免费/vip）
func (s *Season) SeasonCost() string {
	switch s.PaymentStatus {
	// season付费状态 0-免费可看  1-VIP免费可看 2-VIP付费可看 3-其他
	case model.SeasonFree:
		return "免费"
	case model.SeasonVip:
		return "vip"
	default:
		return "付费"
	}
}

type Offshelve struct {
	SeasonID    int64              `json:"seasonid"`
	Type        int                `json:"type"`
	Name        string             `json:"name"`
	SeasonTitle string             `json:"season_title"`
	Episodes    []*OffshelveEpInfo `json:"episodes"`
}

type OffshelveEpInfo struct {
	Index   int    `json:"index"`
	PlayURL string `json:"play_url"`
}

func (c *Offshelve) OffshelveTypeString() string {
	// season类型 1：番剧，2：电影，3：纪录片，4：国漫，5：电视剧，7:综艺
	switch c.Type {
	case model.SeasonTypeBangumi:
		return "番剧"
	case model.SeasonTypeMovie:
		return "电影"
	case model.SeasonTypeDocumentary:
		return "纪录片"
	case model.SeasonTypeGc:
		return "国漫"
	case model.SeasonTypeTv:
		return "电视剧"
	case model.SeasonTypeZi:
		return "综艺"
	}
	return ""
}

type DatabusEntity struct {
	EntityID     string         `json:"entityId"`
	EntityType   string         `json:"entityType"`
	EventType    string         `json:"eventType"`
	Time         string         `json:"time"`
	Value        string         `json:"value"`
	PayLoad      *PayLoad       `json:"payLoad"`
	EntityChange *DatabusEntity `json:"entityChange"`
}

// ArcMsg is
type PayLoad struct {
	Aid        int64 `json:"aid"`
	Cid        int64 `json:"cid"`
	SeasonType int64 `json:"seasonType"`
}

func (c *PayLoad) DataBusTypeString() string {
	// season类型 1：番剧，2：电影，3：纪录片，4：国漫，5：电视剧，7:综艺
	switch c.SeasonType {
	case model.SeasonTypeBangumi:
		return "番剧"
	case model.SeasonTypeMovie:
		return "电影"
	case model.SeasonTypeDocumentary:
		return "纪录片"
	case model.SeasonTypeGc:
		return "国漫"
	case model.SeasonTypeTv:
		return "电视剧"
	case model.SeasonTypeZi:
		return "综艺"
	}
	return ""
}
