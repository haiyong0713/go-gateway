package medialist

import (
	"net/url"
	"strconv"
)

// MediaListReq http://bapi.bilibili.co/project/3492/interface/api/157694
type MediaListReq struct {
	Type        int64  `json:"type"`
	BizId       int64  `json:"biz_id"`
	OType       int64  `json:"otype"`
	Oid         int64  `json:"oid"`
	Bvid        string `json:"bvid"`
	Desc        bool   `json:"desc"`
	Direction   bool   `json:"direction"` // true: 向前取播单项; false(默认): 向后取播单项
	WithCurrent bool   `json:"with_current"`
	Ps          int    `json:"ps"`
	AccessKey   string `json:"access_key"`
	MobiApp     string `json:"mobi_app"`
}

// MediaListRes http://bapi.bilibili.co/project/3492/interface/api/157694
type MediaListRes struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	TTL     int    `json:"ttl"`
	Data    struct {
		MediaList  []MediaList `json:"media_list"`
		HasMore    bool        `json:"has_more"`
		TotalCount int         `json:"total_count"`
	} `json:"data"`
}

type MediaList struct {
	ID        int64  `json:"id"`
	Offset    int    `json:"offset"`
	Index     int    `json:"index"`
	Intro     string `json:"intro"`
	Attr      int    `json:"attr"`
	Tid       int    `json:"tid"`
	CopyRight int    `json:"copy_right"`
	CntInfo   struct {
		Collect   int `json:"collect"`
		Play      int `json:"play"`
		ThumbUp   int `json:"thumb_up"`
		ThumbDown int `json:"thumb_down"`
		Share     int `json:"share"`
		Reply     int `json:"reply"`
		Danmaku   int `json:"danmaku"`
		Coin      int `json:"coin"`
	} `json:"cnt_info"`
	Cover     string `json:"cover"`
	Duration  int    `json:"duration"`
	Pubtime   int    `json:"pubtime"`
	LikeState int    `json:"like_state"`
	FavState  int    `json:"fav_state"`
	Page      int    `json:"page"`
	Pages     []struct {
		ID       int    `json:"id"`
		Title    string `json:"title"`
		Intro    string `json:"intro"`
		Duration int    `json:"duration"`
		Link     string `json:"link"`
		Page     int    `json:"page"`
		Metas    []struct {
			Quality int `json:"quality"`
			Size    int `json:"size"`
		} `json:"metas"`
		From      string `json:"from"`
		Dimension struct {
			Width  int `json:"width"`
			Height int `json:"height"`
			Rotate int `json:"rotate"`
		} `json:"dimension"`
	} `json:"pages"`
	Title string `json:"title"`
	Type  int    `json:"type"`
	Upper struct {
		Mid           int64  `json:"mid"`
		Name          string `json:"name"`
		Face          string `json:"face"`
		Followed      int    `json:"followed"`
		Fans          int    `json:"fans"`
		VipType       int    `json:"vip_type"`
		VipStatue     int    `json:"vip_statue"`
		VipDueDate    int64  `json:"vip_due_date"`
		VipPayType    int    `json:"vip_pay_type"`
		OfficialRole  int    `json:"official_role"`
		OfficialTitle string `json:"official_title"`
		OfficialDesc  string `json:"official_desc"`
	} `json:"upper"`
	Link      string `json:"link"`
	BvID      string `json:"bv_id"`
	ShortLink string `json:"short_link"`
	Rights    struct {
		Bp           int `json:"bp"`
		Elec         int `json:"elec"`
		Download     int `json:"download"`
		Movie        int `json:"movie"`
		Pay          int `json:"pay"`
		UgcPay       int `json:"ugc_pay"`
		Hd5          int `json:"hd5"`
		NoReprint    int `json:"no_reprint"`
		Autoplay     int `json:"autoplay"`
		NoBackground int `json:"no_background"`
	} `json:"rights"`
	ElecInfo interface{} `json:"elec_info"`
	Coin     struct {
		MaxNum     int `json:"max_num"`
		CoinNumber int `json:"coin_number"`
	} `json:"coin"`
	OgvInfo struct {
		Epid      int64 `json:"epid"`
		SeassonId int64 `json:"season_id"`
		Aid       int64 `json:"aid"`
		Cid       int64 `json:"cid"`
	} `json:"ogv_info"`
}

func (r *MediaListReq) ToUrlValues() url.Values {
	values := url.Values{}
	if r == nil {
		return values
	}
	values.Set("type", strconv.FormatInt(r.Type, 10))
	values.Set("biz_id", strconv.FormatInt(r.BizId, 10))
	values.Set("otype", strconv.FormatInt(r.OType, 10))
	values.Set("oid", strconv.FormatInt(r.Oid, 10))
	values.Set("bvid", r.Bvid)
	values.Set("desc", strconv.FormatBool(r.Desc))
	values.Set("direction", strconv.FormatBool(r.Direction))
	values.Set("with_current", strconv.FormatBool(r.WithCurrent))
	values.Set("ps", strconv.Itoa(r.Ps))
	values.Set("access_key", r.AccessKey)
	values.Set("mobi_app", r.MobiApp)
	return values
}
