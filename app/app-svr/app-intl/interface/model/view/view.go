package view

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	dmApi "git.bilibili.co/bapis/bapis-go/community/interface/dm"
	xtime "go-common/library/time"
	"go-gateway/app/app-svr/app-intl/interface/model"
	"go-gateway/app/app-svr/app-intl/interface/model/bangumi"
	"go-gateway/app/app-svr/app-intl/interface/model/tag"
	"go-gateway/app/app-svr/archive/service/api"
	resmdel "go-gateway/app/app-svr/resource/service/model"
	steinsApi "go-gateway/app/app-svr/steins-gate/service/api"

	v1 "git.bilibili.co/bapis/bapis-go/pgc/service/season/season"
)

// View struct
type View struct {
	*ViewStatic
	// owner_ext
	OwnerExt OwnerExt `json:"owner_ext"`
	// now user
	ReqUser *ReqUser `json:"req_user,omitempty"`
	// tag info
	Tag []*tag.Tag `json:"tag,omitempty"`
	// tag 类型对应icon
	TIcon map[string]*tag.TIcon `json:"t_icon,omitempty"`
	// movie
	Movie *bangumi.Movie `json:"movie,omitempty"`
	// bangumi
	Season *bangumi.Season `json:"season,omitempty"`
	// bp
	Bp json.RawMessage `json:"bp,omitempty"`
	// history
	History *History `json:"history,omitempty"`
	// audio
	Audio *Audio `json:"audio,omitempty"`
	// contribute data
	Contributions []*Contribution `json:"contributions,omitempty"`
	// relate data
	Relates     []*Relate `json:"relates,omitempty"`
	ReturnCode  string    `json:"-"`
	UserFeature string    `json:"-"`
	IsRec       int8      `json:"-"`
	// dislike reason
	Dislikes []*Dislike `json:"dislike_reasons,omitempty"`
	// 同步粉版
	DislikeV2 *Dislike2 `json:"dislike_reasons_v2,omitempty"`
	// dm
	DMSeg int `json:"dm_seg,omitempty"`
	// player_icon
	PlayerIcon *resmdel.PlayerIcon `json:"player_icon,omitempty"`
	// vip_active
	VIPActive string `json:"vip_active,omitempty"`
	// cm config
	CMConfig *CMConfig `json:"cm_config,omitempty"`
	// ugc season info
	UgcSeason   *UgcSeason   `json:"ugc_season,omitempty"`
	Interaction *Interaction `json:"interaction,omitempty"`
	Staff       []*Staff     `json:"staff,omitempty"`
	ActivityURL string       `json:"activity_url,omitempty"`
	Label       *Label       `json:"label,omitempty"`
	// config
	Config *Config `json:"config,omitempty"`
	// AI experiments
	PlayParam int             `json:"play_param"` // 1=play automatically the relates, 0=not
	PvFeature json.RawMessage `json:"-"`
	BvID      string          `json:"bvid,omitempty"`
	ForbidRec int64           `json:"-"`
}

// ViewStatic struct
type ViewStatic struct {
	*api.Arc
	Pages []*Page `json:"pages,omitempty"`
}

// ReqUser struct
type ReqUser struct {
	Attention int  `json:"attention"`
	Favorite  int8 `json:"favorite"`
	Like      int8 `json:"like"`
	Dislike   int8 `json:"dislike"`
	Coin      int8 `json:"coin"`
}

// Page struct
type Page struct {
	*api.Page
	Metas  []*Meta            `json:"metas"`
	DMLink string             `json:"dmlink"`
	Audio  *Audio             `json:"audio,omitempty"`
	DM     *dmApi.SubjectInfo `json:"dm,omitempty"`
}

// Meta struct
type Meta struct {
	Quality int    `json:"quality"`
	Format  string `json:"format"`
	Size    int64  `json:"size"`
}

// History struct
type History struct {
	Cid      int64 `json:"cid"`
	Progress int64 `json:"progress"`
}

// CMConfig struct
type CMConfig struct {
	AdsControl  json.RawMessage `json:"ads_control,omitempty"`
	MonitorInfo json.RawMessage `json:"monitor_info,omitempty"`
}

// Dislike struct
type Dislike struct {
	ID   int    `json:"reason_id"`
	Name string `json:"reason_name"`
}

// OwnerExt struct
type OwnerExt struct {
	OfficialVerify struct {
		Type int    `json:"type"`
		Desc string `json:"desc"`
	} `json:"official_verify,omitempty"`
	Vip struct {
		Type          int    `json:"vipType"`
		DueDate       int64  `json:"vipDueDate"`
		DueRemark     string `json:"dueRemark"`
		AccessStatus  int    `json:"accessStatus"`
		VipStatus     int    `json:"vipStatus"`
		VipStatusWarn string `json:"vipStatusWarn"`
	} `json:"vip"`
	Assists  []int64 `json:"assists"`
	Fans     int     `json:"fans"`
	Archives int     `json:"archives"`
}

// Relate struct
type Relate struct {
	Aid         int64       `json:"aid,omitempty"`
	Pic         string      `json:"pic,omitempty"`
	Title       string      `json:"title,omitempty"`
	Author      *api.Author `json:"owner,omitempty"`
	Stat        api.Stat    `json:"stat,omitempty"`
	Duration    int64       `json:"duration,omitempty"`
	Goto        string      `json:"goto,omitempty"`
	Param       string      `json:"param,omitempty"`
	URI         string      `json:"uri,omitempty"`
	Rating      float64     `json:"rating,omitempty"`
	Reserve     string      `json:"reserve,omitempty"`
	From        string      `json:"from,omitempty"`
	Desc        string      `json:"desc,omitempty"`
	RcmdReason  string      `json:"rcmd_reason,omitempty"`
	Badge       string      `json:"badge,omitempty"`
	Cid         int64       `json:"cid,omitempty"`
	SeasonType  int32       `json:"season_type,omitempty"`
	RatingCount int32       `json:"rating_count,omitempty"`
	// cm ad
	AdIndex      int             `json:"ad_index,omitempty"`
	CmMark       int             `json:"cm_mark,omitempty"`
	SrcID        int64           `json:"src_id,omitempty"`
	RequestID    string          `json:"request_id,omitempty"`
	CreativeID   int64           `json:"creative_id,omitempty"`
	CreativeType int64           `json:"creative_type,omitempty"`
	Type         int             `json:"type,omitempty"`
	Cover        string          `json:"cover,omitempty"`
	ButtonTitle  string          `json:"button_title,omitempty"`
	View         int             `json:"view,omitempty"`
	Danmaku      int             `json:"danmaku,omitempty"`
	IsAd         bool            `json:"is_ad,omitempty"`
	IsAdLoc      bool            `json:"is_ad_loc,omitempty"`
	AdCb         string          `json:"ad_cb,omitempty"`
	ShowURL      string          `json:"show_url,omitempty"`
	ClickURL     string          `json:"click_url,omitempty"`
	ClientIP     string          `json:"client_ip,omitempty"`
	Extra        json.RawMessage `json:"extra,omitempty"`
	Button       *Button         `json:"button,omitempty"`
	CardIndex    int             `json:"card_index,omitempty"`
	Source       string          `json:"-"`
	AvFeature    json.RawMessage `json:"-"`
}

// Button struct
type Button struct {
	Title string `json:"title,omitempty"`
	URI   string `json:"uri,omitempty"`
}

// Contribution struct
type Contribution struct {
	Aid    int64      `json:"aid,omitempty"`
	Pic    string     `json:"pic,omitempty"`
	Title  string     `json:"title,omitempty"`
	Author api.Author `json:"owner,omitempty"`
	Stat   api.Stat   `json:"stat,omitempty"`
	CTime  xtime.Time `json:"ctime,omitempty"`
}

// Audio struct
type Audio struct {
	Title    string `json:"title"`
	Cover    string `json:"cover_url"`
	SongID   int    `json:"song_id"`
	Play     int    `json:"play_count"`
	Reply    int    `json:"reply_count"`
	UpperID  int    `json:"upper_id"`
	Entrance string `json:"entrance"`
	SongAttr int    `json:"song_attr"`
}

// VipPlayURL playurl token struct.
type VipPlayURL struct {
	From  string `json:"from"`
	Ts    int64  `json:"ts"`
	Aid   int64  `json:"aid"`
	Cid   int64  `json:"cid"`
	Mid   int64  `json:"mid"`
	VIP   int    `json:"vip"`
	SVIP  int    `json:"svip"`
	Owner int    `json:"owner"`
	Fcs   string `json:"fcs"`
}

// NewRelateRec struct
type NewRelateRec struct {
	TrackID    string          `json:"trackid"`
	Oid        int64           `json:"id"`
	Source     string          `json:"source"`
	AvFeature  json.RawMessage `json:"av_feature"`
	Goto       string          `json:"goto"`
	Title      string          `json:"title"`
	IsDalao    int8            `json:"is_dalao"`
	RcmdReason *RcmdReason     `json:"rcmd_reason,omitempty"`
}

type RcmdReason struct {
	Content    string `json:"content,omitempty"`
	Style      int    `json:"style,omitempty"`
	CornerMark int8   `json:"corner_mark,omitempty"`
}

// RelateRes is
type RelateRes struct {
	Code              int             `json:"code"`
	Data              []*NewRelateRec `json:"data"`
	UserFeature       string          `json:"user_feature"`
	DalaoExp          int             `json:"dalao_exp"`
	PlayParam         int             `json:"play_param"`
	PvFeature         json.RawMessage `json:"pv_feature"`
	AutoplayCountdown int             `json:"autoplay_countdown"`
	ReturnPage        int             `json:"return_page_exp"`
	AutoplayToast     string          `json:"autoplay_toast"`
	GamecardStyleExp  int             `json:"gamecard_style_exp"`
}

// FromAv func
func (r *Relate) FromAv(a *api.Arc, from string, ap *api.PlayerInfo, trackid string) {
	if a == nil {
		return
	}
	r.Aid = a.Aid
	r.Title = a.Title
	r.Pic = a.Pic
	r.Author = &a.Author
	r.Stat = a.Stat
	r.Duration = a.Duration
	r.Cid = a.FirstCid
	r.Goto = model.GotoAv
	r.Param = strconv.FormatInt(a.Aid, 10)
	r.URI = model.FillURI(r.Goto, r.Param, model.AvPlayHandlerGRPC(a, ap, trackid))
	r.From = from
}

// FromOperate func
func (r *Relate) FromOperate(i *NewRelateRec, a *api.Arc, from, trackid string) {
	switch i.Goto {
	case model.GotoAv:
		r.FromAv(a, from, nil, trackid)
	}
	if r.Title == "" {
		r.Title = i.Title
	}
	if i.RcmdReason != nil && i.RcmdReason.Content != "" {
		r.RcmdReason = i.RcmdReason.Content
	}
}

type Interaction struct {
	HistoryNode  *Node  `json:"history_node,omitempty"`
	GraphVersion int64  `json:"graph_version"`
	Msg          string `json:"msg,omitempty"`
	Evaluation   string `json:"evaluation,omitempty"`
	Mark         int64  `json:"mark,omitempty"`
}

type Node struct {
	NodeID int64  `json:"node_id"`
	Title  string `json:"title"`
	CID    int64  `json:"cid"`
}

// Staff from cooperation
type Staff struct {
	Mid            int64  `json:"mid,omitempty"`
	Title          string `json:"title,omitempty"`
	Face           string `json:"face,omitempty"`
	Name           string `json:"name,omitempty"`
	OfficialVerify struct {
		Type int    `json:"type"`
		Desc string `json:"desc"`
	} `json:"official_verify"`
	Vip struct {
		Type          int    `json:"vipType"`
		DueDate       int64  `json:"vipDueDate"`
		DueRemark     string `json:"dueRemark"`
		AccessStatus  int    `json:"accessStatus"`
		VipStatus     int    `json:"vipStatus"`
		VipStatusWarn string `json:"vipStatusWarn"`
		ThemeType     int    `json:"themeType"`
	} `json:"vip"`
	Attention  int   `json:"attention"`
	LabelStyle int32 `json:"label_style,omitempty"`
}

// Label
type Label struct {
	Type int8   `json:"type,omitempty"`
	URI  string `json:"uri,omitempty"`
}

type Config struct {
	RelatesTitle      string `json:"relates_title,omitempty"`
	AutoplayCountdown int    `json:"autoplay_countdown,omitempty"`
	AutoplayDesc      string `json:"autoplay_desc,omitempty"`
	PageRefresh       int    `json:"page_refresh,omitempty"`
}

func ArchivePage(in *steinsApi.Page) (out *api.Page) {
	out = new(api.Page)
	out.Cid = in.Cid
	out.Page = in.Page
	out.From = in.From
	out.Part = in.Part
	out.Duration = in.Duration
	out.Vid = in.Vid
	out.Desc = in.Desc
	out.WebLink = in.WebLink
	out.Dimension = api.Dimension{
		Width:  in.Dimension.Width,
		Height: in.Dimension.Height,
		Rotate: in.Dimension.Rotate,
	}
	return
}

// FromBangumi func
func (r *Relate) FromBangumi(ban *v1.CardInfoProto, aid int64) {
	r.Title = ban.Title
	r.Pic = ban.NewEp.Cover
	r.Stat = api.Stat{
		Danmaku: int32(ban.Stat.Danmaku),
		View:    int32(ban.Stat.View),
		Fav:     int32(ban.Stat.Follow),
	}
	r.Goto = model.GotoBangumi
	r.Param = strconv.FormatInt(int64(ban.SeasonId), 10)
	r.URI = model.FillURI(r.Goto, r.Param, nil)
	if aid != 0 && r.URI != "" {
		if strings.Contains(r.URI, "?") {
			r.URI += fmt.Sprintf("&from_av=%d", aid)
		} else {
			r.URI += fmt.Sprintf("?from_av=%d", aid)
		}
	}
	r.SeasonType = ban.SeasonType
	r.Badge = ban.SeasonTypeName
	r.Desc = ban.NewEp.IndexShow
	if ban.Rating != nil {
		r.Rating = float64(ban.Rating.Score)
		r.RatingCount = ban.Rating.Count
	}
}

// DislikeReasons .
func (v *View) DislikeReasons() {
	const (
		_noSeason = 1
		_region   = 2
		_channel  = 3
		_upper    = 4
		_tagMAX   = 2
	)
	var (
		taginfo *tag.Tag
	)
	v.DislikeV2 = &Dislike2{
		Title: "选择不想看的原因，减少相似内容推荐",
	}
	if v.Author.Name != "" {
		v.DislikeV2.Reasons = append(v.DislikeV2.Reasons, &DislikeReasons{Id: _upper, Name: "UP主:" + v.Author.Name, Mid: v.Author.Mid})
	}
	if len(v.Tag) > 0 {
		for i, t := range v.Tag {
			v.DislikeV2.Reasons = append(v.DislikeV2.Reasons, &DislikeReasons{Id: _channel, Name: "频道:" + t.Name, TagId: t.TagID})
			if i == 0 {
				taginfo = t
			}
			if i+1 >= _tagMAX {
				break
			}
		}
	}
	dislike := &DislikeReasons{Id: _noSeason, Name: "我不想看这个内容", Mid: v.Author.Mid, Rid: v.TypeID}
	if taginfo != nil {
		dislike.TagId = taginfo.TagID
	}
	v.DislikeV2.Reasons = append(v.DislikeV2.Reasons, dislike)
}

type Dislike2 struct {
	Title    string            `json:"title,omitempty"`
	Subtitle string            `json:"subtitle,omitempty"`
	Reasons  []*DislikeReasons `json:"reasons,omitempty"`
}

type DislikeReasons struct {
	Id    int64  `json:"id,omitempty"`
	Mid   int64  `json:"mid,omitempty"`
	Rid   int32  `json:"rid,omitempty"`
	TagId int64  `json:"tag_id,omitempty"`
	Name  string `json:"name,omitempty"`
}
