package model

import (
	gaiamdl "git.bilibili.co/bapis/bapis-go/silverbullet/gaia/interface"
	stime "time"

	"go-common/library/time"
	honorgrpc "go-gateway/app/app-svr/archive-honor/service/api"
	arcmdl "go-gateway/app/app-svr/archive/service/api"
	steinsApi "go-gateway/app/app-svr/steins-gate/service/api"
	"go-gateway/app/app-svr/ugc-season/service/api"
	chmdl "go-gateway/app/web-svr/web/interface/model/channel"
	"go-gateway/pkg/idsafe/bvid"

	cfcgrpc "git.bilibili.co/bapis/bapis-go/content-flow-control/service"

	uparcgrpc "git.bilibili.co/bapis/bapis-go/up-archive/service"

	accmdl "git.bilibili.co/bapis/bapis-go/account/service"
	ugcmdl "git.bilibili.co/bapis/bapis-go/account/service/ugcpay"
	webgrpc "git.bilibili.co/bapis/bapis-go/bilibili/web/interface/v1"
	dmmdl "git.bilibili.co/bapis/bapis-go/community/interface/dm"
	cremdl "git.bilibili.co/bapis/bapis-go/creative/open/service"
	res "git.bilibili.co/bapis/bapis-go/resource/service/v2"
)

// Special recommend type.
const (
	SpecRecmTypeCard = 1
	SpecRecmTypeArc  = 2
	SpecRecmTypeGame = 3
	StaffLabelAd     = 1
	AttrBitNoSearch  = uint(4)
	ArcMemberAccess  = 10000
)

// archive forbidden
const (
	_NoRank      = "norank"      // 排行禁止
	_NoIndex     = "noindex"     // 分区动态禁止
	_NoRecommend = "norecommend" // 推荐禁止
	_NoSearch    = "nosearch"    // 搜索禁止
	_NoHot       = "nohot"       // 热门禁止
	_NoShare     = "no_share"    // 搜索禁止
)

// View view data
type View struct {
	// archive data
	Bvid string `json:"bvid"`
	*ViewArc
	NoCache bool `json:"no_cache"`
	// video data pages
	Pages    []*arcmdl.Page         `json:"pages,omitempty"`
	Subtitle *Subtitle              `json:"subtitle"`
	Asset    *ugcmdl.AssetQueryResp `json:"asset,omitempty"`
	Label    *Label                 `json:"label,omitempty"`
	//Staff
	Staff     []*Staff           `json:"staff,omitempty"`
	UGCSeason *webgrpc.UGCSeason `json:"ugc_season,omitempty"`
	// 基础合集是否展示
	IsSeasonDisplay bool  `json:"is_season_display"`
	SteinGuideCid   int64 `json:"stein_guide_cid,omitempty"`
	// user grab
	UserGarb   *UserGarb   `json:"user_garb"`
	HonorReply *HonorReply `json:"honor_reply"`
	// 点赞icon
	LikeIcon string `json:"like_icon"`
}

// 稿件商品信息
type GoodsInfo struct {
	PayState      arcmdl.PayState `json:"pay_stat"`        // 付费稿件支付状况 1为付费合集已经付款
	GoodsId       string          `json:"goods_id"`        // 商品id
	Category      arcmdl.Category `json:"category"`        // 商品付费类型
	GoodsPrice    int64           `json:"goods_price"`     // 商品价格(分)
	FreeWatch     bool            `json:"free_watch"`      // 是否免费试看
	GoodsName     string          `json:"goods_name"`      // 商品名称
	GoodsPriceFmt string          `json:"goods_price_fmt"` // 商品价格(元)
}
type HonorReply struct {
	Honor []*Honor `json:"honor,omitempty"`
}

type Honor struct {
	*honorgrpc.Honor
	WeeklyRecommendNum int64 `json:"weekly_recommend_num"`
}

type UserGarb struct {
	URLImageAniCut string `json:"url_image_ani_cut"`
}

type ViewArc struct {
	Aid             int64               `json:"aid"`
	Videos          int64               `json:"videos"`
	TypeID          int32               `json:"tid"`
	TypeName        string              `json:"tname"`
	Copyright       int32               `json:"copyright"`
	Pic             string              `json:"pic"`
	Title           string              `json:"title"`
	PubDate         time.Time           `json:"pubdate"`
	Ctime           time.Time           `json:"ctime"`
	Desc            string              `json:"desc"`
	DescV2          []*DescV2           `json:"desc_v2"`
	State           int32               `json:"state"`
	Access          int32               `json:"access,omitempty"`
	Attribute       int32               `json:"attribute,omitempty"`
	Tag             string              `json:"-"`
	Tags            []string            `json:"tags,omitempty"`
	Duration        int64               `json:"duration"`
	MissionID       int64               `json:"mission_id,omitempty"`
	OrderID         int64               `json:"order_id,omitempty"`
	RedirectURL     string              `json:"redirect_url,omitempty"`
	Forward         int64               `json:"forward,omitempty"`
	Rights          ViewArcRights       `json:"rights"`
	Author          arcmdl.Author       `json:"owner"`
	Stat            ViewArcStat         `json:"stat"`
	ReportResult    string              `json:"report_result,omitempty"`
	Dynamic         string              `json:"dynamic"`
	FirstCid        int64               `json:"cid,omitempty"`
	Dimension       arcmdl.Dimension    `json:"dimension,omitempty"`
	StaffInfo       []*arcmdl.StaffInfo `json:"-"`
	SeasonID        int64               `json:"season_id,omitempty"`
	FestivalJumpUrl string              `json:"festival_jump_url,omitempty"`
	// 首映
	Premiere    *Premiere `json:"premiere"`
	TeenageMode int32     `json:"teenage_mode"`
	// 商品信息，如果稿件是付费时才会有
	GoodsInfo []*GoodsInfo `json:"goods_info,omitempty"`
	// 是否属于付费合集
	IsChargeableSeason bool `json:"is_chargeable_season"`
	// 是否是竖屏/故事
	IsStory bool `json:"is_story"`
}

type DescV2 struct {
	RawText string `json:"raw_text"`
	Type    int64  `json:"type"`
	BizId   int64  `json:"biz_id"`
}

type DescReply struct {
	Desc   string    `json:"desc"`
	DescV2 []*DescV2 `json:"desc_v2"`
}

type Premiere struct {
	// 首映状态
	State int32 `json:"state"`
	// 首映开始时间
	StartTime int64 `json:"start_time"`
	// 系统时间
	NowTime int64 `json:"now_time"`
	// 首映专属聊天室id
	RoomID int64 `json:"room_id"`
	// 预约id
	SID int64 `json:"sid"`
}

type PremiereInfo struct {
	// xxx人参与
	Participant int64 `json:"participant"`
	// xxx次互动
	Interaction int64 `json:"interaction"`
}

func (a *ViewArc) FmtWebArc(arc *arcmdl.Arc) {
	a.Aid = arc.Aid
	a.Videos = arc.Videos
	a.TypeID = arc.TypeID
	a.TypeName = arc.TypeName
	a.Copyright = arc.Copyright
	a.Pic = arc.Pic
	a.Title = arc.Title
	a.PubDate = arc.PubDate
	a.Ctime = arc.Ctime
	a.Desc = arc.Desc
	a.State = arc.State
	a.Tag = arc.Tag
	a.Tags = arc.Tags
	a.Duration = arc.Duration
	a.MissionID = arc.MissionID
	a.OrderID = arc.OrderID
	a.RedirectURL = arc.RedirectURL
	a.Forward = arc.Forward
	a.Rights = ViewArcRights{
		Bp:              arc.Rights.Bp,
		Elec:            arc.Rights.Elec,
		Download:        arc.Rights.Download,
		Movie:           arc.Rights.Movie,
		Pay:             arc.Rights.Pay,
		HD5:             arc.Rights.HD5,
		NoReprint:       arc.Rights.NoReprint,
		Autoplay:        arc.Rights.Autoplay,
		UGCPay:          arc.Rights.UGCPay,
		IsCooperation:   arc.Rights.IsCooperation,
		UGCPayPreview:   arc.Rights.UGCPayPreview,
		NoBackground:    arc.Rights.NoBackground,
		ArcPay:          arc.Rights.ArcPay,
		ArcPayFreeWatch: arc.Rights.ArcPayFreeWatch,
	}
	if arc.AttrValV2(arcmdl.AttrBitV2CleanMode) == arcmdl.AttrYes {
		a.Rights.CleanMode = 1
	}
	if arc.AttrVal(arcmdl.AttrBitSteinsGate) == arcmdl.AttrYes {
		a.Rights.IsSteinGate = 1
	}
	if arc.AttrValV2(arcmdl.AttrBitV2Is360) == arcmdl.AttrYes {
		a.Rights.Is360 = 1
	}
	if arc.AttrValV2(arcmdl.AttrBitTeenager) == arcmdl.AttrYes {
		a.TeenageMode = 1
	}
	a.Author = arc.Author
	a.Stat = ViewArcStat{
		Aid:     arc.Stat.Aid,
		View:    arc.Stat.View,
		Danmaku: arc.Stat.Danmaku,
		Reply:   arc.Stat.Reply,
		Fav:     arc.Stat.Fav,
		Coin:    arc.Stat.Coin,
		Share:   arc.Stat.Share,
		NowRank: arc.Stat.NowRank,
		HisRank: arc.Stat.HisRank,
		Like:    arc.Stat.Like,
		DisLike: arc.Stat.DisLike,
	}
	if arc.Access >= ArcMemberAccess {
		a.Stat.View = -1
	}
	a.ReportResult = arc.ReportResult
	a.Dynamic = arc.Dynamic
	a.FirstCid = arc.FirstCid
	a.Dimension = arc.Dimension
	a.StaffInfo = arc.StaffInfo
	a.SeasonID = arc.SeasonID
	if arc.IsNormalPremiere() {
		// 首映稿件
		a.Premiere = &Premiere{
			State:     int32(arc.GetPremiere().GetState()),
			StartTime: arc.GetPremiere().GetStartTime(),
			RoomID:    arc.GetPremiere().GetRoomId(),
			NowTime:   stime.Now().Unix(),
		}
	}
	if arc.AttrValV2(arcmdl.AttrBitV2Pay) == arcmdl.AttrYes {
		if arc.Pay != nil && arc.Pay.AttrVal(arcmdl.PaySubTypeAttrBitSeason) == arcmdl.AttrYes {
			// 付费稿件
			a.IsChargeableSeason = true
		}
		for _, gs := range arc.Pay.GoodsInfo {
			a.GoodsInfo = append(a.GoodsInfo, &GoodsInfo{
				PayState:      gs.PayState,
				GoodsId:       gs.GoodsId,
				GoodsName:     gs.GoodsName,
				GoodsPrice:    gs.GoodsPrice,
				GoodsPriceFmt: gs.GoodsPriceFmt,
				Category:      gs.Category,
			})
		}
	}
	rotate := 1
	if arc.Dimension.Width < arc.Dimension.Height || (arc.Dimension.Width > arc.Dimension.Height && arc.Dimension.Rotate == int64(rotate)) {
		// 竖屏视频
		a.IsStory = true
	}
}

type ViewArcStat struct {
	Aid        int64  `json:"aid"`
	View       int32  `json:"view"`
	Danmaku    int32  `json:"danmaku"`
	Reply      int32  `json:"reply"`
	Fav        int32  `json:"favorite"`
	Coin       int32  `json:"coin"`
	Share      int32  `json:"share"`
	NowRank    int32  `json:"now_rank"`
	HisRank    int32  `json:"his_rank"`
	Like       int32  `json:"like"`
	DisLike    int32  `json:"dislike"`
	Evaluation string `json:"evaluation"`
	ArgueMsg   string `json:"argue_msg"`
}

type ViewArcRights struct {
	Bp              int32 `json:"bp"`
	Elec            int32 `json:"elec"`
	Download        int32 `json:"download"`
	Movie           int32 `json:"movie"`
	Pay             int32 `json:"pay"`
	HD5             int32 `json:"hd5"`
	NoReprint       int32 `json:"no_reprint"`
	Autoplay        int32 `json:"autoplay"`
	UGCPay          int32 `json:"ugc_pay"`
	IsCooperation   int32 `json:"is_cooperation"`
	UGCPayPreview   int32 `json:"ugc_pay_preview"`
	NoBackground    int32 `json:"no_background"`
	CleanMode       int32 `json:"clean_mode"`
	IsSteinGate     int32 `json:"is_stein_gate"`
	Is360           int32 `json:"is_360"`
	NoShare         int32 `json:"no_share"`
	ArcPay          int32 `json:"arc_pay"`    // 是否付费稿件(attribute_v2 右移13位为付费时)
	ArcPayFreeWatch int32 `json:"free_watch"` // 是否付费稿件可免费观看, 0无法观看, 1合集内免费观看
}

// Staff .
type Staff struct {
	Mid        int64               `json:"mid"`
	Title      string              `json:"title"`
	Name       string              `json:"name"`
	Face       string              `json:"face"`
	Vip        accmdl.VipInfo      `json:"vip"`
	Official   accmdl.OfficialInfo `json:"official"`
	Follower   int64               `json:"follower"`
	LabelStyle int32               `json:"label_style"`
}

// Label .
type Label struct {
	Type int64 `json:"type"`
}

// AssetRelation .
type AssetRelation struct {
	State int `json:"state"`
}

// Stat archive stat web struct
type Stat struct {
	Aid        int64       `json:"aid"`
	Bvid       string      `json:"bvid"`
	View       interface{} `json:"view"`
	Danmaku    int32       `json:"danmaku"`
	Reply      int32       `json:"reply"`
	Fav        int32       `json:"favorite"`
	Coin       int32       `json:"coin"`
	Share      int32       `json:"share"`
	Like       int32       `json:"like"`
	NowRank    int32       `json:"now_rank"`
	HisRank    int32       `json:"his_rank"`
	NoReprint  int32       `json:"no_reprint"`
	Copyright  int32       `json:"copyright"`
	ArgueMsg   string      `json:"argue_msg"`
	Evaluation string      `json:"evaluation"`
}

// Detail detail data
type Detail struct {
	View      *View
	Card      *Card
	Tags      []*chmdl.VideoTag
	Reply     *ReplyHot
	Related   []*BvArc
	Spec      *SpecRecm
	HotShare  *HotShare  `json:"hot_share"`
	Elec      *ElecShow  `json:"elec"`
	Recommend *Recommend `json:"recommend"`
	ViewAddit *ViewAddit `json:"view_addit"`
}

type Recommend struct {
	Show  bool     `json:"show"`
	Title string   `json:"title"`
	List  []*BvArc `json:"list"`
}

type HotShare struct {
	Show bool     `json:"show"`
	List []*BvArc `json:"list"`
}

// ArchiveUserCoins .
type ArchiveUserCoins struct {
	Multiply int64 `json:"multiply"`
}

// Subtitle dm subTitle.
type Subtitle struct {
	AllowSubmit bool            `json:"allow_submit"`
	List        []*SubtitleItem `json:"list"`
}

// SubtitleItem dm subTitle.
type SubtitleItem struct {
	*dmmdl.VideoSubtitle
	Author *accmdl.Info `json:"author"`
}

// TripleRes struct
type TripleRes struct {
	Like        bool                    `json:"like"`
	Coin        bool                    `json:"coin"`
	Fav         bool                    `json:"fav"`
	Multiply    int64                   `json:"multiply"`
	UpID        int64                   `json:"-"`
	Anticheat   bool                    `json:"-"`
	IsRisk      bool                    `json:"is_risk"`
	GaiaResType GaiaResponseType        `json:"gaia_res_type"`
	GaiaData    *gaiamdl.RuleCheckReply `json:"gaia_data"`
}

type LikeRes struct {
	UpID        int64                   `json:"-"`
	IsRisk      bool                    `json:"is_risk"`
	GaiaResType GaiaResponseType        `json:"gaia_res_type"`
	GaiaData    *gaiamdl.RuleCheckReply `json:"gaia_data"`
}

type ShareRes struct {
	Shares      int64
	IsRisk      bool                    `json:"is_risk"`
	GaiaResType GaiaResponseType        `json:"gaia_res_type"`
	GaiaData    *gaiamdl.RuleCheckReply `json:"gaia_data"`
}

// SpecRecmItem .
type SpecRecmItem struct {
	ID        int64   `json:"id"`
	CardType  int     `json:"card_type"`
	CardValue string  `json:"card_value"`
	Partition []int32 `json:"partition"`
	Tag       []int64 `json:"tag"`
	Avid      []int64 `json:"avid"`
}

// SpecRecm special recommend struct.
type SpecRecm struct {
	Type    int32               `json:"type"`
	Archive *BvArc              `json:"archive,omitempty"`
	Game    *Game               `json:"game,omitempty"`
	Card    *res.WebSpecialCard `json:"card,omitempty"`
}

// BvArc .
type BvArc struct {
	*arcmdl.Arc
	Bvid       string   `json:"bvid"`
	SeasonType int64    `json:"season_type"`
	IsOGV      bool     `json:"is_ogv"`
	OGVInfo    *OGVInfo `json:"ogv_info"`
	RcmdReason string   `json:"rcmd_reason"`
}

type OGVInfo struct {
	ReleaseDateShow string `json:"release_date_show"`
}

// ArcCustomConfig is
type ArcCustomConfig struct {
	Aid       int64  `json:"aid"`
	Content   string `json:"content"`
	URL       string `json:"url"`
	Highlight string `json:"highlight"`
	Image     string `json:"image"`
	ImageBig  string `json:"image_big"`
}

type ArcRecommend struct {
	ID      int64  `json:"id"`
	Goto    string `json:"goto"`
	IsDalao int    `json:"is_dalao"`
}

type UpLikeImg struct {
	UpLikeImg *cremdl.UpLikeImgRsp `json:"up_like_img"`
}

type ArcForbidden struct {
	NoRank      bool
	NoDynamic   bool
	NoRecommend bool
	NoSearch    bool
	NoShare     bool
	NoHot       bool
	NoPushHBlog bool // 推送粉丝动态
}

type ViewAddit struct {
	NoRecommendLive     bool `json:"63"`
	NoRecommendActivity bool `json:"64"`
}

var (
	// StatAllowStates archive stat allow states
	statAllowStates = []int32{-9, -15, -30}
)

// CheckAllowState check archive stat allow state
func CheckAllowState(arc *arcmdl.Arc) bool {
	if arc.IsNormal() {
		return true
	}
	for _, allow := range statAllowStates {
		if arc.State == allow {
			return true
		}
	}
	return false
}

func CopyFromDetail(in *Detail) (out *webgrpc.ViewDetailReply) {
	if in == nil {
		return nil
	}
	out = new(webgrpc.ViewDetailReply)
	out.View = &webgrpc.View{
		Arc:           CopyFromArc(in.View.ViewArc),
		NoCache:       in.View.NoCache,
		SteinGuideCid: in.View.SteinGuideCid,
	}
	if in.View.Subtitle != nil {
		out.View.Subtitle = &webgrpc.Subtitle{
			AllowSubmit: in.View.Subtitle.AllowSubmit,
		}
		for _, v := range in.View.Subtitle.List {
			item := &webgrpc.SubtitleItem{
				Id:          v.Id,
				Lan:         v.Lan,
				LanDoc:      v.LanDoc,
				IsLock:      v.IsLock,
				AuthorMid:   v.AuthorMid,
				SubtitleUrl: v.SubtitleUrl,
			}
			if v.Author != nil {
				item.Author = &webgrpc.AccInfo{
					Mid:  v.Author.Mid,
					Name: v.Author.Name,
					Sex:  v.Author.Sex,
					Face: v.Author.Face,
					Sign: v.Author.Sign,
				}
			}
			out.View.Subtitle.List = append(out.View.Subtitle.List, item)
		}
	}
	if in.View.Asset != nil {
		out.View.Asset = &webgrpc.UGCPayAsset{
			Price:         in.View.Asset.Price,
			PlatformPrice: in.View.Asset.PlatformPrice,
		}
	}
	if in.View.Label != nil {
		out.View.Label = &webgrpc.ViewLabel{
			Type: in.View.Label.Type,
		}
	}
	for _, v := range in.View.Pages {
		out.View.Pages = append(out.View.Pages, &webgrpc.Page{
			Cid:      v.Cid,
			Page:     v.Page,
			From:     v.From,
			Part:     v.Part,
			Duration: v.Duration,
			Vid:      v.Vid,
			Desc:     v.Desc,
			Weblink:  v.WebLink,
			Dimension: webgrpc.Dimension{
				Width:  v.Dimension.Width,
				Height: v.Dimension.Height,
				Rotate: v.Dimension.Rotate,
			},
		})
	}
	for _, v := range in.View.Staff {
		if v != nil {
			out.View.Staff = append(out.View.Staff, &webgrpc.Staff{
				Mid:   v.Mid,
				Title: v.Title,
				Name:  v.Name,
				Face:  v.Face,
				Vip: &webgrpc.VipInfo{
					Type:       v.Vip.Type,
					Status:     v.Vip.Status,
					VipPayType: v.Vip.VipPayType,
					ThemeType:  v.Vip.ThemeType,
				},
				Official: &webgrpc.OfficialInfo{
					Role:  v.Official.Role,
					Title: v.Official.Title,
					Desc:  v.Official.Desc,
				},
				Follower:   v.Follower,
				LabelStyle: v.LabelStyle,
			})
		}
	}
	if in.View.UGCSeason != nil {
		out.View.UgcSeason = in.View.UGCSeason
	}
	if in.Card != nil {
		out.Card = &webgrpc.Card{
			Following:    in.Card.Following,
			ArchiveCount: in.Card.ArchiveCount,
			ArticleCount: in.Card.ArticleCount,
			Follower:     in.Card.Follower,
		}
		if in.Card.Card != nil {
			out.Card.Card = &webgrpc.AccountCard{
				Mid:      in.Card.Card.Mid,
				Name:     in.Card.Card.Name,
				Sex:      in.Card.Card.Sex,
				Rank:     in.Card.Card.Rank,
				Face:     in.Card.Card.Face,
				Spacesta: in.Card.Card.Spacesta,
				Sign:     in.Card.Card.Sign,
				LevelInfo: webgrpc.CardLevelInfo{
					Cur:     in.Card.Card.LevelInfo.Cur,
					NextExp: 0,
				},
				Pendant: webgrpc.PendantInfo{
					Pid:    in.Card.Card.Pendant.Pid,
					Name:   in.Card.Card.Pendant.Name,
					Image:  in.Card.Card.Pendant.Image,
					Expire: in.Card.Card.Pendant.Expire,
				},
				Nameplate: webgrpc.NameplateInfo{
					Nid:        in.Card.Card.Nameplate.Nid,
					Name:       in.Card.Card.Nameplate.Name,
					Image:      in.Card.Card.Nameplate.Image,
					ImageSmall: in.Card.Card.Nameplate.ImageSmall,
					Level:      in.Card.Card.Nameplate.Level,
					Condition:  in.Card.Card.Nameplate.Condition,
				},
				Official: webgrpc.OfficialInfo{
					Role:  in.Card.Card.Official.Role,
					Title: in.Card.Card.Official.Title,
					Desc:  in.Card.Card.Official.Desc,
				},
				OfficialVerify: webgrpc.OfficialVerify{
					Type: in.Card.Card.OfficialVerify.Type,
					Desc: in.Card.Card.OfficialVerify.Desc,
				},
				Vip: webgrpc.CardVip{
					Type:      in.Card.Card.Vip.Type,
					VipStatus: in.Card.Card.Vip.VipStatus,
					ThemeType: in.Card.Card.Vip.ThemeType,
				},
				Fans:      in.Card.Card.Fans,
				Friend:    in.Card.Card.Friend,
				Attention: in.Card.Card.Attention,
			}
		}
		if in.Card.Space != nil {
			out.Card.Space = &webgrpc.Space{
				SImg: in.Card.Space.SImg,
				LImg: in.Card.Space.LImg,
			}
		}
	}
	for _, v := range in.Tags {
		out.Tags = append(out.Tags, &webgrpc.Tag{
			Id:           v.ID,
			Name:         v.Name,
			Cover:        v.Cover,
			HeadCover:    v.HeadCover,
			Content:      v.Content,
			ShortContent: v.ShortContent,
			Type:         int32(v.Type),
			State:        int32(v.State),
			Ctime:        v.CTime,
			TagCount: &webgrpc.TagCount{
				View:  int64(v.Count.View),
				Use:   int64(v.Count.Use),
				Atten: int64(v.Count.Atten),
			},
			IsAtten:   int32(v.IsAtten),
			Likes:     v.Likes,
			Hates:     v.Hates,
			Attribute: int32(v.Attribute),
			Liked:     int32(v.Liked),
			Hated:     int32(v.Hated),
		})
	}
	if in.Reply != nil {
		out.Reply = new(webgrpc.HotReply)
		if in.Reply.Page != nil {
			out.Reply.Page = &webgrpc.ReplyPage{
				Acount: in.Reply.Page.Acount,
				Count:  in.Reply.Page.Count,
				Num:    in.Reply.Page.Num,
				Size_:  in.Reply.Page.Size,
			}
		}
		for _, v := range in.Reply.Replies {
			item := CopyFromReply(v)
			if item != nil {
				out.Reply.Replies = append(out.Reply.Replies, item)
			}
		}
	}
	for _, v := range in.Related {
		out.Related = append(out.Related, CopyFromBvArc(v))
	}
	return
}

func CopyFromReply(in *ReplyItem) (out *webgrpc.Reply) {
	if in == nil {
		return nil
	}
	out = &webgrpc.Reply{
		Rpid:   in.RpID,
		Oid:    in.Oid,
		Type:   int32(in.Type),
		Mid:    in.Mid,
		Root:   in.Root,
		Parent: in.Parent,
		Dialog: in.Dialog,
		Count:  int32(in.Count),
		Rcount: int32(in.RCount),
		Floor:  int32(in.Floor),
		State:  int32(in.State),
		Attr:   in.Attr,
		Ctime:  in.CTime,
		Like:   int32(in.Like),
		Hate:   int32(in.Hate),
	}
	if in.Content != nil {
		if out.Content == nil {
			out.Content = &webgrpc.ReplyContent{}
		}
		out.Content = &webgrpc.ReplyContent{
			RpId:    in.Content.RpID,
			Message: in.Content.Message,
			Ats:     in.Content.Ats,
			Topics:  in.Content.Topics,
			Ip:      in.Content.IP,
			Plat:    int32(in.Content.Plat),
			Device:  in.Content.Device,
			Version: in.Content.Version,
		}
	}
	if len(in.Replies) > 0 {
		for _, rr := range in.Replies {
			item := CopyFromReply(rr)
			if item != nil {
				out.Replies = append(out.Replies, item)
			}
		}
	}
	return
}

func CopyFromArc(in *ViewArc) (out *webgrpc.Arc) {
	if in == nil {
		return nil
	}
	out = &webgrpc.Arc{
		Aid:         in.Aid,
		Videos:      in.Videos,
		TypeId:      in.TypeID,
		TypeName:    in.TypeName,
		Copyright:   in.Copyright,
		Pic:         in.Pic,
		Title:       in.Title,
		Pubdate:     in.PubDate,
		Ctime:       in.Ctime,
		Desc:        in.Desc,
		State:       in.State,
		Access:      in.Access,
		Attribute:   in.Attribute,
		Tag:         in.Tag,
		Tags:        in.Tags,
		Duration:    in.Duration,
		MissionId:   in.MissionID,
		OrderId:     in.OrderID,
		RedirectUrl: in.RedirectURL,
		Forward:     in.Forward,
		Rights: webgrpc.Rights{
			Bp:            in.Rights.Bp,
			Elec:          in.Rights.Elec,
			Download:      in.Rights.Download,
			Movie:         in.Rights.Movie,
			Pay:           in.Rights.Pay,
			Hd5:           in.Rights.HD5,
			NoReprint:     in.Rights.NoReprint,
			Autoplay:      in.Rights.Autoplay,
			UgcPay:        in.Rights.UGCPay,
			IsCooperation: in.Rights.IsCooperation,
			UgcPayPreview: in.Rights.UGCPayPreview,
		},
		Author: webgrpc.Author{
			Mid:  in.Author.Mid,
			Name: in.Author.Name,
			Face: in.Author.Face,
		},
		Stat: webgrpc.Stat{
			Aid:     in.Stat.Aid,
			View:    in.Stat.View,
			Danmaku: in.Stat.Danmaku,
			Reply:   in.Stat.Reply,
			Fav:     in.Stat.Fav,
			Coin:    in.Stat.Coin,
			Share:   in.Stat.Share,
			NowRank: in.Stat.NowRank,
			HisRank: in.Stat.HisRank,
			Like:    in.Stat.Like,
			Dislike: in.Stat.DisLike,
		},
		ReportResult: in.ReportResult,
		Dynamic:      in.Dynamic,
		FirstCid:     in.FirstCid,
		Dimension: webgrpc.Dimension{
			Width:  in.Dimension.Width,
			Height: in.Dimension.Height,
			Rotate: in.Dimension.Rotate,
		},
		SeasonId: in.SeasonID,
	}
	return
}

func CopyFromBvArc(in *BvArc) (out *webgrpc.Arc) {
	if in == nil {
		return nil
	}
	out = &webgrpc.Arc{
		Aid:         in.Aid,
		Videos:      in.Videos,
		TypeId:      in.TypeID,
		TypeName:    in.TypeName,
		Copyright:   in.Copyright,
		Pic:         in.Pic,
		Title:       in.Title,
		Pubdate:     in.PubDate,
		Ctime:       in.Ctime,
		Desc:        in.Desc,
		State:       in.State,
		Access:      in.Access,
		Attribute:   in.Attribute,
		Tag:         in.Tag,
		Tags:        in.Tags,
		Duration:    in.Duration,
		MissionId:   in.MissionID,
		OrderId:     in.OrderID,
		RedirectUrl: in.RedirectURL,
		Forward:     in.Forward,
		Rights: webgrpc.Rights{
			Bp:            in.Rights.Bp,
			Elec:          in.Rights.Elec,
			Download:      in.Rights.Download,
			Movie:         in.Rights.Movie,
			Pay:           in.Rights.Pay,
			Hd5:           in.Rights.HD5,
			NoReprint:     in.Rights.NoReprint,
			Autoplay:      in.Rights.Autoplay,
			UgcPay:        in.Rights.UGCPay,
			IsCooperation: in.Rights.IsCooperation,
			UgcPayPreview: in.Rights.UGCPayPreview,
		},
		Author: webgrpc.Author{
			Mid:  in.Author.Mid,
			Name: in.Author.Name,
			Face: in.Author.Face,
		},
		Stat: webgrpc.Stat{
			Aid:     in.Stat.Aid,
			View:    in.Stat.View,
			Danmaku: in.Stat.Danmaku,
			Reply:   in.Stat.Reply,
			Fav:     in.Stat.Fav,
			Coin:    in.Stat.Coin,
			Share:   in.Stat.Share,
			NowRank: in.Stat.NowRank,
			HisRank: in.Stat.HisRank,
			Like:    in.Stat.Like,
			Dislike: in.Stat.DisLike,
		},
		ReportResult: in.ReportResult,
		Dynamic:      in.Dynamic,
		FirstCid:     in.FirstCid,
		Dimension: webgrpc.Dimension{
			Width:  in.Dimension.Width,
			Height: in.Dimension.Height,
			Rotate: in.Dimension.Rotate,
		},
		SeasonId: in.SeasonID,
	}
	return
}

func ArchivePage(in *steinsApi.Page) (out *arcmdl.Page) {
	out = new(arcmdl.Page)
	out.Cid = in.Cid
	out.Page = in.Page
	out.From = in.From
	out.Part = in.Part
	out.Duration = in.Duration
	out.Vid = in.Vid
	out.Desc = in.Desc
	out.WebLink = in.WebLink
	out.Dimension = arcmdl.Dimension{
		Width:  in.Dimension.Width,
		Height: in.Dimension.Height,
		Rotate: in.Dimension.Rotate,
	}
	return
}

func CopyFromUGCSeason(in *api.View) (out *webgrpc.UGCSeason) {
	if in == nil || in.Season == nil {
		return nil
	}
	out = &webgrpc.UGCSeason{
		Id:        in.Season.ID,
		Title:     in.Season.Title,
		Cover:     in.Season.Cover,
		Mid:       in.Season.Mid,
		Intro:     in.Season.Intro,
		SignState: in.Season.SignState,
		Attribute: in.Season.Attribute,
		Stat: webgrpc.SeasonStat{
			SeasonId: in.Season.Stat.SeasonID,
			View:     in.Season.Stat.View,
			Danmaku:  in.Season.Stat.Danmaku,
			Reply:    in.Season.Stat.Reply,
			Fav:      in.Season.Stat.Fav,
			Coin:     in.Season.Stat.Coin,
			Share:    in.Season.Stat.Share,
			NowRank:  in.Season.Stat.NowRank,
			HisRank:  in.Season.Stat.HisRank,
			Like:     in.Season.Stat.Like,
		},
		EpCount:     in.Season.EpCount,
		SeasonType:  in.Season.AttrVal(api.AttrSnType),
		IsPaySeason: in.Season.AttrVal(api.SeasonAttrSnPay) == api.AttrSnYes,
	}
	for _, v := range in.Sections {
		if v == nil {
			continue
		}
		item := &webgrpc.SeasonSection{
			SeasonId: v.SeasonID,
			Id:       v.ID,
			Title:    v.Title,
			Type:     v.Type,
		}
		for _, ep := range v.Episodes {
			if ep == nil || ep.Arc == nil || ep.Page == nil {
				continue
			}
			bid, _ := bvid.AvToBv(ep.Aid)
			item.Episodes = append(item.Episodes, &webgrpc.SeasonEpisode{
				SeasonId:  ep.SeasonID,
				SectionId: ep.SectionID,
				Id:        ep.ID,
				Aid:       ep.Aid,
				Bvid:      bid,
				Cid:       ep.Cid,
				Title:     ep.Title,
				Attribute: ep.Attribute,
				Arc: &webgrpc.Arc{
					Aid:      ep.Aid,
					Pic:      ep.Arc.Pic,
					Pubdate:  ep.Arc.PubDate,
					Ctime:    ep.Arc.PubDate,
					Title:    ep.Title,
					Duration: ep.Arc.Duration,
					Stat: webgrpc.Stat{
						Aid:     ep.Aid,
						View:    ep.Arc.Stat.View,
						Danmaku: ep.Arc.Stat.Danmaku,
						Reply:   ep.Arc.Stat.Reply,
						Fav:     ep.Arc.Stat.Fav,
						Coin:    ep.Arc.Stat.Coin,
						Share:   ep.Arc.Stat.Share,
						NowRank: ep.Arc.Stat.NowRank,
						HisRank: ep.Arc.Stat.HisRank,
						Like:    ep.Arc.Stat.Like,
					},
					Rights: webgrpc.Rights{
						ArcPay:    int32((ep.Arc.GetAttributeV2() >> arcmdl.AttrBitV2Pay) & int64(1)),
						FreeWatch: int32(ep.AttrVal(api.EpisodeAttrSnFreeWatch)),
					},
				},
				Page: &webgrpc.Page{
					Cid:      ep.Page.Cid,
					Page:     ep.Page.Page,
					From:     ep.Page.From,
					Part:     ep.Page.Part,
					Duration: ep.Page.Duration,
					Vid:      ep.Page.Vid,
					Desc:     ep.Page.Desc,
					Weblink:  ep.Page.WebLink,
					Dimension: webgrpc.Dimension{
						Width:  ep.Page.Dimension.Width,
						Height: ep.Page.Dimension.Height,
						Rotate: ep.Page.Dimension.Rotate,
					},
				},
			})
		}
		out.Sections = append(out.Sections, item)
	}
	return
}

func CopyFromArcToBvArc(in *arcmdl.Arc, bvid string) (out *BvArc) {
	ClearAttrAndAccess(in)
	return &BvArc{
		Arc:  in,
		Bvid: bvid,
	}
}

func CopyFromUpArcToBvArc(from *uparcgrpc.Arc, bvid string) *BvArc {
	from.Attribute = 0
	from.AttributeV2 = 0
	from.Access = 0
	to := &arcmdl.Arc{
		Aid:         from.Aid,
		Videos:      from.Videos,
		TypeID:      from.TypeID,
		TypeName:    from.TypeName,
		Copyright:   from.Copyright,
		Pic:         from.Pic,
		Title:       from.Title,
		PubDate:     from.PubDate,
		Ctime:       from.Ctime,
		Desc:        from.Desc,
		State:       from.State,
		Access:      from.Access,
		Attribute:   from.Attribute,
		Tag:         from.Tag,
		Tags:        from.Tags,
		Duration:    from.Duration,
		MissionID:   from.MissionID,
		OrderID:     from.OrderID,
		RedirectURL: from.RedirectURL,
		Forward:     from.Forward,
		Rights: arcmdl.Rights{
			Bp:            from.Rights.Bp,
			Elec:          from.Rights.Elec,
			Download:      from.Rights.Download,
			Movie:         from.Rights.Movie,
			Pay:           from.Rights.Pay,
			HD5:           from.Rights.HD5,
			NoReprint:     from.Rights.NoReprint,
			Autoplay:      from.Rights.Autoplay,
			UGCPay:        from.Rights.UGCPay,
			IsCooperation: from.Rights.IsCooperation,
			UGCPayPreview: from.Rights.UGCPayPreview,
			NoBackground:  from.Rights.NoBackground,
		},
		Author: arcmdl.Author{
			Mid:  from.Author.Mid,
			Name: from.Author.Name,
			Face: from.Author.Face,
		},
		Stat: arcmdl.Stat{
			Aid:     from.Stat.Aid,
			View:    from.Stat.View,
			Danmaku: from.Stat.Danmaku,
			Reply:   from.Stat.Reply,
			Fav:     from.Stat.Fav,
			Coin:    from.Stat.Coin,
			Share:   from.Stat.Share,
			NowRank: from.Stat.NowRank,
			HisRank: from.Stat.HisRank,
			Like:    from.Stat.Like,
			DisLike: from.Stat.DisLike,
			Follow:  from.Stat.Follow,
		},
		ReportResult: from.ReportResult,
		Dynamic:      from.Dynamic,
		FirstCid:     from.FirstCid,
		Dimension: arcmdl.Dimension{
			Width:  from.Dimension.Width,
			Height: from.Dimension.Height,
			Rotate: from.Dimension.Rotate,
		},
		SeasonID:    from.SeasonID,
		AttributeV2: from.AttributeV2,
	}
	for _, v := range from.StaffInfo {
		if v == nil {
			continue
		}
		to.StaffInfo = append(to.StaffInfo, &arcmdl.StaffInfo{
			Mid:       v.Mid,
			Title:     v.Title,
			Attribute: v.Attribute,
		})
	}
	return &BvArc{
		Arc:  to,
		Bvid: bvid,
	}
}

func ClearAttrAndAccess(in *arcmdl.Arc) {
	in.Attribute = 0
	in.AttributeV2 = 0
	in.Access = 0
}

func ItemToArcForbidden(cfcItem []*cfcgrpc.ForbiddenItem) *ArcForbidden {
	acrForbidden := &ArcForbidden{}
	if len(cfcItem) == 0 {
		return acrForbidden
	}
	for _, item := range cfcItem {
		if item == nil {
			continue
		}
		switch item.Key {
		case _NoRank:
			if item.Value == 1 {
				acrForbidden.NoRank = true
			}
		case _NoIndex:
			if item.Value == 1 {
				acrForbidden.NoDynamic = true
			}
		case _NoRecommend:
			if item.Value == 1 {
				acrForbidden.NoRecommend = true
			}
		case _NoSearch:
			if item.Value == 1 {
				acrForbidden.NoSearch = true
			}
		case _NoShare:
			if item.Value == 1 {
				acrForbidden.NoShare = true
			}
		case _NoHot:
			if item.Value == 1 {
				acrForbidden.NoHot = true
			}
		}
	}
	return acrForbidden
}

func GetTeenageMode(arc *arcmdl.Arc) int32 {
	return arc.AttrVal(arcmdl.AttrBitTeenager)
}
