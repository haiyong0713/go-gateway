package model

import (
	gaiamdl "git.bilibili.co/bapis/bapis-go/silverbullet/gaia/interface"
	"math/rand"

	"go-common/library/time"

	v1 "git.bilibili.co/bapis/bapis-go/archive/service"
	cfcgrpc "git.bilibili.co/bapis/bapis-go/content-flow-control/service"
	uparcmdl "git.bilibili.co/bapis/bapis-go/up-archive/service"
)

// archive forbidden
const (
	_NoRank      = "norank"      // 排行禁止
	_NoIndex     = "noindex"     // 分区动态禁止
	_NoRecommend = "norecommend" // 推荐禁止
)

// BvArc arc add bvid.
type BvArc struct {
	*v1.Arc
	Bvid string `json:"bvid"`
}

// UpArcStat up archives stat struct.
type UpArcStat struct {
	View  int64 `json:"view"`
	Reply int64 `json:"reply"`
	Dm    int64 `json:"dm"`
	Fans  int64 `json:"fans"`
}

// ArchiveReason archive with reason struct.
type ArchiveReason struct {
	*v1.Arc
	Bvid       string `json:"bvid"`
	Reason     string `json:"reason"`
	InterVideo bool   `json:"inter_video"`
}

// SearchRes search res data.
type SearchRes struct {
	TList map[string]*SearchTList `json:"tlist"`
	VList []*SearchVList          `json:"vlist"`
}

// SearchTList search cate list.
type SearchTList struct {
	Tid   int64  `json:"tid"`
	Count int64  `json:"count"`
	Name  string `json:"name"`
}

// SearchVList video list.
type SearchVList struct {
	Comment        int64       `json:"comment"`
	TypeID         int64       `json:"typeid"`
	Play           interface{} `json:"play"`
	Pic            string      `json:"pic"`
	SubTitle       string      `json:"subtitle"`
	Description    string      `json:"description"`
	Copyright      string      `json:"copyright"`
	Title          string      `json:"title"`
	Review         int64       `json:"review"`
	Author         string      `json:"author"`
	Mid            int64       `json:"mid"`
	Created        interface{} `json:"created"`
	Length         string      `json:"length"`
	VideoReview    int64       `json:"video_review"`
	Aid            int64       `json:"aid"`
	Bvid           string      `json:"bvid"`
	HideClick      bool        `json:"hide_click"`
	IsPay          int         `json:"is_pay"`
	IsUnionVideo   int         `json:"is_union_video"`
	IsSteinsGate   int         `json:"is_steins_gate"`
	IsLivePlayback int         `json:"is_live_playback"`
}

// UpArc up archive struct
type UpArc struct {
	Count int64      `json:"count"`
	List  []*ArcItem `json:"list"`
}

// ArcItem space archive item.
type ArcItem struct {
	Aid      int64  `json:"aid"`
	Bvid     string `json:"bvid"`
	Pic      string `json:"pic"`
	Title    string `json:"title"`
	Duration int64  `json:"duration"`
	Author   struct {
		Mid  int64  `json:"mid"`
		Name string `json:"name"`
		Face string `json:"face"`
	} `json:"author"`
	Stat struct {
		View    interface{} `json:"view"`
		Danmaku int32       `json:"danmaku"`
		Reply   int32       `json:"reply"`
		Fav     int32       `json:"favorite"`
		Coin    int32       `json:"coin"`
		Share   int32       `json:"share"`
		Like    int32       `json:"like"`
	} `json:"stat"`
	Rights  v1.Rights `json:"rights"`
	Pubdate time.Time `json:"pubdate"`
}

// FromArc from archive to space act item.
// nolint:gomnd
func (ac *ArcItem) FromArc(arc *v1.Arc) {
	ac.Aid = arc.Aid
	ac.Pic = arc.Pic
	ac.Title = arc.Title
	ac.Duration = arc.Duration
	ac.Author.Mid = arc.Author.Mid
	ac.Author.Name = arc.Author.Name
	ac.Author.Face = arc.Author.Face
	ac.Stat.View = arc.Stat.View
	if arc.Access >= 10000 {
		ac.Stat.View = "--"
	}
	ac.Stat.Danmaku = arc.Stat.Danmaku
	ac.Stat.Reply = arc.Stat.Reply
	ac.Stat.Fav = arc.Stat.Fav
	ac.Stat.Share = arc.Stat.Share
	ac.Stat.Like = arc.Stat.Like
	ac.Pubdate = arc.PubDate
	ac.Rights = arc.Rights
}

// nolint:gomnd
func (ac *ArcItem) FromUpArc(arc *uparcmdl.Arc) {
	ac.Aid = arc.Aid
	ac.Pic = arc.Pic
	ac.Title = arc.Title
	ac.Duration = arc.Duration
	ac.Author.Mid = arc.Author.Mid
	ac.Author.Name = arc.Author.Name
	ac.Author.Face = arc.Author.Face
	ac.Stat.View = arc.Stat.View
	if arc.Access >= 10000 {
		ac.Stat.View = "--"
	}
	ac.Stat.Danmaku = arc.Stat.Danmaku
	ac.Stat.Reply = arc.Stat.Reply
	ac.Stat.Fav = arc.Stat.Fav
	ac.Stat.Share = arc.Stat.Share
	ac.Stat.Like = arc.Stat.Like
	ac.Pubdate = arc.PubDate
	ac.Rights = v1.Rights{
		Bp:            arc.Rights.Bp,
		Elec:          arc.Rights.Elec,
		Download:      arc.Rights.Download,
		Movie:         arc.Rights.Movie,
		Pay:           arc.Rights.Pay,
		HD5:           arc.Rights.HD5,
		NoReprint:     arc.Rights.NoReprint,
		Autoplay:      arc.Rights.Autoplay,
		UGCPay:        arc.Rights.UGCPay,
		IsCooperation: arc.Rights.IsCooperation,
		UGCPayPreview: arc.Rights.UGCPayPreview,
		NoBackground:  arc.Rights.NoBackground,
	}
}

func RandInt(r *rand.Rand, min, max int) int {
	if min >= max || max == 0 {
		return max
	}
	return r.Intn(max-min+1) + min // 左闭右闭
}

type ArcSearchRes struct {
	List           *SearchRes              `json:"list"`
	Page           *SearchPage             `json:"page"`
	EpisodicButton *ArcListButton          `json:"episodic_button,omitempty"`
	IsRisk         bool                    `json:"is_risk"`
	GaiaResType    GaiaResponseType        `json:"gaia_res_type"`
	GaiaData       *gaiamdl.RuleCheckReply `json:"gaia_data"`
}

type SearchPage struct {
	Pn    int   `json:"pn"`
	Ps    int   `json:"ps"`
	Count int64 `json:"count"`
}

type ArcListButton struct {
	Text string `json:"text"`
	URI  string `json:"uri"`
}

func ClearAttrAndAccess(in *v1.Arc) {
	in.Attribute = 0
	in.AttributeV2 = 0
	in.Access = 0
}

type ArcForbidden struct {
	NoRank      bool
	NoDynamic   bool
	NoRecommend bool
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
		}
	}
	return acrForbidden
}
