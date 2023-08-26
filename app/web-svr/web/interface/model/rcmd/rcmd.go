package rcmd

import (
	"encoding/json"
	"strconv"

	"go-common/library/log"
	xtime "go-common/library/time"
	"go-gateway/app/app-svr/archive/service/api"
	arcmdl "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/web/interface/model"
	"go-gateway/pkg/idsafe/bvid"

	appchanmdl "go-gateway/app/app-svr/app-channel/interface/model"

	seasongrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/season"
)

type Group string

const (
	GroupA  = Group("a")
	GroupB  = Group("b")
	AV      = "av"         // ugc 视频卡
	Live    = "live"       // 直播卡
	Ogv     = "ogv_season" // ogv综艺卡
	Ad      = "ad"         // 商业广告卡
	FromCpm = int8(1)
)

type Owner struct {
	Mid  int64  `json:"mid"`
	Name string `json:"name"`
	Face string `json:"face"`
}

type Stat struct {
	View    int32 `json:"view"`
	Like    int32 `json:"like"`
	Danmaku int32 `json:"danmaku"`
}

type Item struct {
	ID           int64               `json:"id"`
	Bvid         string              `json:"bvid"`
	Cid          int64               `json:"cid"`
	Goto         string              `json:"goto"`
	URI          string              `json:"uri"`
	Pic          string              `json:"pic"`
	Title        string              `json:"title"`
	Duration     int64               `json:"duration"`
	PubDate      xtime.Time          `json:"pubdate"`
	Owner        *Owner              `json:"owner"`
	Stat         *Stat               `json:"stat"`
	AvFeature    json.RawMessage     `json:"av_feature"`
	IsFollowed   int64               `json:"is_followed"`
	RcmdReason   json.RawMessage     `json:"rcmd_reason"`
	ShowInfo     int64               `json:"show_info"`
	TrackId      string              `json:"track_id"`
	Pos          int                 `json:"pos"`
	RoomInfo     *model.LiveRoomInfo `json:"room_info"`     // 直播房间信息
	OgvInfo      *AIOgvInfo          `json:"ogv_info"`      // 直播房间信息
	BusinessInfo *Assignment         `json:"business_info"` // ad信息
	IsStock      int                 `json:"is_stock"`      // 库存 0,1 ,1为库存
}

type AIOgvInfo struct {
	SeasonId    int32  `json:"season_id"` // season_id
	Title       string `json:"title"`
	Subtitle    string `json:"sub_title"`
	Cover       string `json:"cover"`
	Follow      string `json:"follows,omitempty"`
	View        string `json:"plays,omitempty"`
	Danmaku     string `json:"dms,omitempty"`
	Likes       string `json:"likes,omitempty"`
	SquareCover string `json:"square_cover"`        // 方图
	IndexShow   string `json:"pub_index,omitempty"` // 更新到第几话
}

type Abtest struct {
	Group Group `json:"group"`
}

type TopRcmd struct {
	Item                  []*Item            `json:"item"`
	BusinessCard          []*Item            `json:"business_card"`
	FloorInfo             []*AIRcmdFloorInfo `json:"floor_info"`
	UserFeature           json.RawMessage    `json:"user_feature"`
	Abtest                *Abtest            `json:"abtest,omitempty"`
	PreloadExposePct      float32            `json:"preload_expose_pct"`
	PreloadFloorExposePct float32            `json:"preload_floor_expose_pct"`
}

type AITopRcmd struct {
	Trackid    string          `json:"trackid"`     // 标记一次数据请求的ID
	ID         int64           `json:"id"`          // 卡片id
	Goto       string          `json:"goto"`        // 卡片类型 av｜live
	Source     string          `json:"source"`      // 卡片触发来源
	AvFeature  json.RawMessage `json:"av_feature"`  // 稿件特征
	IsFollowed int64           `json:"is_followed"` // 是否关注(前端"已关注"标签)
	RcmdReason json.RawMessage `json:"rcmd_reason"` // 推荐解释
	ShowInfo   int64           `json:"show_info"`   // 前端展示的点赞or弹幕类型 0: //前端展示点赞(def).  1: //前端展示弹幕
	// 第三行特有字段 仅在仅在fresh_idx=1 或换一换生效
	Pos          int            `json:"pos"` // 插入位置
	BusinessInfo *BusinessInfos `json:"business_info"`
	IsStock      int            `json:"is_stock"`
}

// 第三行特有字段 仅在仅在fresh_idx=1 或换一换生效
type AIRcmdFloorInfo struct {
	Id   int    `json:"id"`   // 业务楼层的枚举值
	Name string `json:"name"` // 楼层名
	Rows int    `json:"rows"` // 行数
}

func (i *Item) FromArc(arc *arcmdl.Arc, rcmd *AITopRcmd) {
	i.ID = arc.Aid
	bvid, err := bvid.AvToBv(arc.Aid)
	if err != nil {
		log.Error("日志告警 AvToBv aid:%v,error:%+v", arc.Aid, err)
	}
	i.Bvid = bvid
	i.Cid = arc.FirstCid
	i.Goto = "av"
	i.URI = "https://www.bilibili.com/video/" + bvid
	i.Pic = arc.Pic
	i.Title = arc.Title
	i.Duration = arc.Duration
	i.PubDate = arc.PubDate
	i.Owner = &Owner{
		Mid:  arc.Author.Mid,
		Name: arc.Author.Name,
		Face: arc.Author.Face,
	}
	i.Stat = &Stat{
		View:    arc.Stat.View,
		Like:    arc.Stat.Like,
		Danmaku: arc.Stat.Danmaku,
	}
	if rcmd == nil {
		return
	}
	i.AvFeature = rcmd.AvFeature
	i.IsFollowed = rcmd.IsFollowed
	i.RcmdReason = rcmd.RcmdReason
	i.ShowInfo = rcmd.ShowInfo
	i.TrackId = rcmd.Trackid
}

func (i *Item) FromLive(info *model.LiveRoomInfo, rcmd *AITopRcmd) {
	i.ID = info.RoomId
	i.Goto = "live"
	i.URI = "https://live.bilibili.com/" + strconv.FormatInt(info.RoomId, 10)
	if info.Show != nil {
		i.Pic = info.Show.Cover
		i.Title = info.Show.Title
	}
	i.Owner = &Owner{
		Mid: info.Uid,
	}
	i.RoomInfo = info
	i.formatRcmd(rcmd)
}

func (i *Item) FromOgv(info *seasongrpc.CardInfoProto, rcmd *AITopRcmd) {
	i.ID = int64(info.SeasonId)
	i.Goto = "ogv_season"
	i.URI = info.Url
	i.Pic = info.Cover
	i.Title = info.Cover
	i.OgvInfo = &AIOgvInfo{
		SeasonId:    info.SeasonId,
		Cover:       info.Cover,
		Title:       info.Title,
		SquareCover: info.SquareCover,
		Follow:      appchanmdl.Stat64String(info.Stat.Follow, ""),
		View:        appchanmdl.Stat64String(info.Stat.View, ""),
		Danmaku:     appchanmdl.Stat64String(info.Stat.Danmaku, ""),
		Likes:       appchanmdl.Stat64String(info.Stat.Likes, ""),
		Subtitle:    info.Subtitle,
		IndexShow:   info.NewEp.IndexShow,
	}
	i.formatRcmd(rcmd)
}

func (i *Item) FromAd(rcmd *AITopRcmd) {
	i.formatRcmd(rcmd)
	if rcmd == nil || rcmd.BusinessInfo == nil || rcmd.BusinessInfo.AdInfo == nil {
		return
	}
	businessInfo := rcmd.BusinessInfo
	adInfo := businessInfo.AdInfo
	if info := adInfo.Info; info != nil {
		i.BusinessInfo = &Assignment{
			CreativeType: info.CreativeType,
			Aid:          info.CreativeContent.VideoID,
			RequestID:    businessInfo.RequestId,
			SrcID:        adInfo.SourceId,
			IsAdLoc:      true,
			IsAd:         adInfo.IsAd,
			CmMark:       adInfo.CmMark,
			CreativeID:   info.CreativeID,
			AdCb:         info.AdCb,
			ShowURL:      info.CreativeContent.ShowURL,
			ClickURL:     info.CreativeContent.ClickURL,
			Name:         info.CreativeContent.Title,
			AdDesc:       info.CreativeContent.Desc,
			Pic:          info.CreativeContent.ImageURL,
			LitPic:       info.CreativeContent.ThumbnailURL,
			URL:          info.CreativeContent.URL,
			PosNum:       int(adInfo.Index),
			Title:        info.CreativeContent.Title,
			ServerType:   FromCpm,
			IsCpm:        true,
			AdverName:    info.Extra.Card.AdverName,
			CardType:     info.CardType,
			BusinessMark: info.Extra.Card.BusinessMark,
		}
	} else {
		i.BusinessInfo = &Assignment{
			IsAdLoc:   true,
			RequestID: businessInfo.RequestId,
			IsAd:      false,
			SrcID:     adInfo.SourceId,
			ResID:     int(i.ID),
			CmMark:    adInfo.CmMark,
		}
	}
}

func (i *Item) formatRcmd(rcmd *AITopRcmd) {
	if rcmd == nil {
		return
	}
	i.AvFeature = rcmd.AvFeature
	i.IsFollowed = rcmd.IsFollowed
	i.RcmdReason = rcmd.RcmdReason
	i.ShowInfo = rcmd.ShowInfo
	i.TrackId = rcmd.Trackid
	i.Pos = rcmd.Pos
}

type Showlist struct {
	Section *Section `json:"section"`
}

type Section struct {
	Items []*SectionItem `json:"items"`
}

type SectionItem struct {
	AvFeature json.RawMessage `json:"av_feature"`
	Goto      string          `json:"goto"`
	ID        int64           `json:"id"`
	Pos       int             `json:"pos"`
	Source    string          `json:"source"`
}

type TopRcmdReq struct {
	Buvid       string `form:"-"`
	Mid         int64  `form:"-"`
	Api         string `form:"-"`
	Ip          string `form:"-"`
	FreshType   int    `form:"fresh_type"`
	Ps          int    `form:"ps" validate:"min=0,max=30"`
	FreshIdx    int    `form:"fresh_idx"`
	FreshIdx1h  int    `form:"fresh_idx_1h"`
	FeedVersion string `form:"feed_version" default:"V0"`
	YNum        int    `form:"y_num"`
	IsFeed      int    `form:"is_feed"`
	HomepageVer int    `form:"homepage_ver"` // 上报用
	FetchRow    int    `form:"fetch_row"`    // 上报用
	Brush       int    `form:"brush"`
	Sid         string `form:"s_id"`
	Country     string `form:"country"`
	Province    string `form:"province"`
	City        string `form:"city"`
	UserAgent   string `form:"ua"`
	//Session     string `form:"session"`
}

type TopFeedRcmdRep struct {
	Code                  int                `json:"code"`
	Data                  []*AITopRcmd       `json:"data"`
	BusinessCards         []*AITopRcmd       `json:"business_card"` // 仅在仅在fresh_idx=1 或换一换生效
	FloorInfos            []*AIRcmdFloorInfo `json:"floor_info"`    // 仅在仅在fresh_idx=1 或换一换生效
	UserFeature           json.RawMessage    `json:"user_feature"`
	PreloadExposePct      float32            `json:"preload_expose_pct"`
	PreloadFloorExposePct float32            `json:"preload_floor_expose_pct"`
}

type TopRcmdIds struct {
	Aids      []int64 // 视频
	LiveIds   []int64 // 直播
	SeasonIds []int32 // ogv
	AdIds     []int64 // ad
	HasData   bool    // 是否有数据
}

type TopFeedRcmdReply struct {
	DataItem              []*Item
	BusinessItem          []*Item
	FloorInfo             []*AIRcmdFloorInfo
	UserFeature           json.RawMessage
	PreloadExposePct      float32
	PreloadFloorExposePct float32
}

type BusinessInfos struct {
	RequestId  string          `json:"request_id"`
	AdsControl json.RawMessage `json:"ads_control"`
	AdInfo     *AdInfo         `json:"ad_info"`
}

type AdInfo struct {
	ResourceId int   `json:"resource_id"`
	SourceId   int64 `json:"source_id"`
	Index      int64 `json:"index"`
	IsAd       bool  `json:"is_ad"`
	CmMark     int8  `json:"cm_mark"`
	Info       *Info `json:"info"`
}

type Info struct {
	CreativeID      int64 `json:"creative_id"`
	CreativeType    int8  `json:"creative_type"`
	CreativeContent struct {
		Title        string `json:"title"`
		Desc         string `json:"description"`
		VideoID      int64  `json:"video_id"`
		UserName     string `json:"username"`
		ImageURL     string `json:"image_url"`
		ImageMD5     string `json:"image_md5"`
		LogURL       string `json:"log_url"`
		LogMD5       string `json:"log_md5"`
		URL          string `json:"url"`
		ClickURL     string `json:"click_url"`
		ShowURL      string `json:"show_url"`
		ThumbnailURL string `json:"thumbnail_url"`
	} `json:"creative_content"`
	AdCb  string `json:"ad_cb"`
	Extra struct {
		Card struct {
			AdverName    string          `json:"adver_name"`
			BusinessMark json.RawMessage `json:"ad_tag_style"`
		} `json:"card"`
	} `json:"extra"`
	CardType int64 `json:"card_type"`
}

// Assignment struct
type Assignment struct {
	ID         int        `json:"id"`
	ContractID string     `json:"contract_id"`
	ResID      int        `json:"res_id"`
	AsgID      int64      `json:"asg_id"`
	PosNum     int        `json:"pos_num"`
	Name       string     `json:"name"`
	Pic        string     `json:"pic"`
	LitPic     string     `json:"litpic"`
	URL        string     `json:"url"`
	Rule       string     `json:"-"`
	Style      int32      `json:"style"`
	IsAd       bool       `json:"is_ad,omitempty"`
	Archive    *ArchiveBV `json:"archive,omitempty"`
	Aid        int64      `json:"-"`
	Weight     int        `json:"-"`
	Atype      int8       `json:"-"`
	MTime      xtime.Time `json:"-"`
	RoomID     int64      `json:"-"`
	Agency     string     `json:"agency"`
	Label      string     `json:"label"`
	Intro      string     `json:"intro"`
	// cpm
	CreativeType  int8       `json:"creative_type"`
	RequestID     string     `json:"request_id,omitempty"`
	CreativeID    int64      `json:"creative_id,omitempty"`
	SrcID         int64      `json:"src_id,omitempty"`
	ShowURL       string     `json:"show_url,omitempty"`
	ClickURL      string     `json:"click_url,omitempty"`
	Area          int8       `json:"area"`
	IsAdLoc       bool       `json:"is_ad_loc"`
	AdCb          string     `json:"ad_cb"`
	Title         string     `json:"title"`
	ServerType    int8       `json:"server_type"`
	CmMark        int8       `json:"cm_mark"`
	IsCpm         bool       `json:"-"`
	STime         xtime.Time `json:"stime"`
	Mid           string     `json:"mid"`
	ActivityID    int64      `json:"-"`
	ActivitySTime xtime.Time `json:"-"`
	ActivityETime xtime.Time `json:"-"`
	ActivityType  int8       `json:"activity_type"`
	EpID          int32      `json:"epid"`
	//Season        *seasongrpc.SeasonCard `json:"season"`
	//Room          *LiveRoomInfo          `json:"room"`
	SubTitle     string          `json:"sub_title"`
	AdDesc       string          `json:"ad_desc"`
	AdverName    string          `json:"adver_name"`
	NullFrame    bool            `json:"null_frame"`
	PicMainColor string          `json:"pic_main_color"`
	CardType     int64           `json:"card_type"`
	BusinessMark json.RawMessage `json:"business_mark"`
	// inline播放相关配置
	Inline Inline `json:"inline"`
	//消息推送方,数据来源方:英文名加固定数字
	Operater string `json:"operater"`
}

type ArchiveBV struct {
	*api.Arc
	BVID string `json:"bvid"`
}

// inline播放配置
type Inline struct {
	// inline播放和跳转是否相同 1:不相同 2:相同
	InlineUseSame int8 `json:"inline_use_same"`
	// inline播放类型 0:默认值 1:web稿件
	InlineType int8 `json:"inline_type"`
	// inline_type 对应的value(ID)
	InlineUrl string `json:"inline_url"`
	// inline弹幕开关 1:关闭 2:开启
	InlineBarrageSwitch int8 `json:"inline_barrage_switch"`
}
