package view

import (
	"fmt"

	cardmdl "go-gateway/app/app-svr/app-card/interface/model"
	viewApi "go-gateway/app/app-svr/app-view/interface/api/view"
	seasonApi "go-gateway/app/app-svr/ugc-season/service/api"
	"go-gateway/pkg/idsafe/bvid"

	"go-gateway/app/app-svr/archive/service/api"
)

const (
	_signExlusive         = 1
	_signFirst            = 2
	_signExText           = "独家"
	_signFtText           = "首发"
	_signTextColor        = "#ffffff"
	_signTextBgColor      = "#fb7299"
	_signTextNightColor   = "#e5e5e5"
	_signTextBgNightColor = "#bb5b76"
	_finishedFmt          = "全部%d话"
	_notFinishedText      = "查看详情"
	// 未购买按钮
	_unPayButtonTextColor      = "#FFFFFF"
	_unPayButtonBgColor        = "#FFB027"
	_unPayButtonTextNightColor = "#FFFFFF"
	_unPayButtonBgNightColor   = "#DB8700"
	_unPayButtonText           = "立即购买"
	// 已购买按钮
	_payButtonTextColor      = "#9499A0"
	_payButtonBgColor        = "#E3E5E7"
	_payButtonTextNightColor = "#757A81"
	_payButtonBgNightColor   = "#2F3134"
	_payButtonText           = "已购买"
)

var (
	_rate = map[int]int64{15: 464, 16: 464, 32: 1028, 48: 1328, 64: 2192, 74: 3192, 80: 3192, 112: 6192, 116: 6192, 66: 1820}
)

// BuildMetas builds the meta for page
func BuildMetas(duration int64) (metas []*Meta) {
	metas = make([]*Meta, 0, 4)
	for q, r := range _rate {
		meta := &Meta{
			Quality: q,
			Size:    int64(float64(r*duration) * 1.1 / 8.0),
		}
		metas = append(metas, meta)
	}
	return
}

// UgcSeason Def.
type UgcSeason struct {
	// archive-service season fields
	Id       int64          `json:"id"`
	Title    string         `json:"title"`
	Cover    string         `json:"cover"`
	Intro    string         `json:"intro"`
	Sections []*Section     `json:"sections,omitempty"`
	Stat     seasonApi.Stat `json:"stat"`
	// custom fields
	LabelText           string `json:"label_text,omitempty"`
	LabelTextColor      string `json:"label_text_color,omitempty"`
	LabelBgColor        string `json:"label_bg_color,omitempty"`
	LabelTextNightColor string `json:"label_text_night_color,omitempty"`
	LabelBgNightColor   string `json:"label_bg_night_color,omitempty"`
	DescRight           string `json:"desc_right,omitempty"`
	EpCount             int64  `json:"ep_count"`
	// seasonType
	SeasonType          viewApi.SeasonType `json:"season_type"`
	ShowContinualButton bool               `json:"show_continual_button,omitempty"`
	Activity            *viewApi.UgcSeasonActivity
	SeasonAbility       []string `json:"season_ability,omitempty"`
	EpNum               int64    `json:"ep_num,omitempty"`
	// 是否付费合集
	SeasonPay bool `json:"season_pay,omitempty"`
	// 合集绑定商品信息
	GoodsInfo viewApi.GoodsInfo `json:"goods_info,omitempty"`
	// 按钮：立即购买/已购买
	PayButton viewApi.ButtonStyle `json:"pay_button,omitempty"`
	// 新标签文案，(付费,签约，独家)
	LabelTextNew string `json:"label_text_new,omitempty"`
}

// FromSeason def.
func (v *UgcSeason) FromSeason(season *seasonApi.View) {
	if season == nil || season.Season == nil {
		return
	}
	sn := season.Season
	v.Id = sn.ID
	v.Title = sn.Title
	v.Cover = sn.Cover
	v.Intro = sn.Intro
	v.Stat = sn.Stat
	for _, section := range season.Sections {
		sectionView := new(Section)
		sectionView.FromSection(section)
		v.Sections = append(v.Sections, sectionView)
	}
	v.EpCount = sn.EpCount
	v.EpNum = sn.EpCount
	if sn.EpNum > sn.EpCount {
		v.EpNum = sn.EpNum
	}
	if sn.SignState == _signExlusive || sn.SignState == _signFirst { // 签约状态和style
		v.LabelTextColor = _signTextColor
		v.LabelBgColor = _signTextBgColor
		v.LabelTextNightColor = _signTextNightColor
		v.LabelBgNightColor = _signTextBgNightColor
		if sn.SignState == _signExlusive {
			v.LabelText = _signExText
		} else {
			v.LabelText = _signFtText
		}
		v.LabelTextNew = v.LabelText
	}

	if sn.AttrVal(seasonApi.AttrSnFinished) == seasonApi.AttrSnYes && sn.EpCount > 0 { // 是否完结
		v.DescRight = fmt.Sprintf(_finishedFmt, sn.EpCount)
	} else {
		v.DescRight = _notFinishedText
	}
	if sn.AttrVal(seasonApi.AttrSnActType) == seasonApi.AttrSnYes {
		return
	}
	v.SeasonType = viewApi.SeasonType_Good
	if sn.AttrVal(seasonApi.AttrSnType) == seasonApi.AttrSnYes {
		v.SeasonType = viewApi.SeasonType_Base
	}
}

// Section def.
type Section struct {
	Id       int64      `json:"id"`
	Title    string     `json:"title"`
	Type     int64      `json:"type"`
	Episodes []*Episode `json:"episodes,omitempty"`
}

// FromEpisode builds app-view's section from archive-service's section
func (v *Section) FromSection(section *seasonApi.Section) {
	if section == nil {
		return
	}
	v.Id = section.ID
	v.Title = section.Title
	v.Type = section.Type
	for _, ep := range section.Episodes {
		epView := new(Episode)
		epView.FromEpisode(ep)
		v.Episodes = append(v.Episodes, epView)
	}
}

// Episode def.
type Episode struct {
	Id                 int64               `json:"id"`
	Aid                int64               `json:"aid"`
	Cid                int64               `json:"cid"`
	Title              string              `json:"title"` // ep's title
	Cover              string              `json:"cover"` // arc's cover
	CoverRightText     string              `json:"cover_right_text,omitempty"`
	*seasonApi.ArcPage                     // page info
	Stat               *seasonApi.ArcStat  `json:"stat"` // archive's stat
	Metas              []*Meta             `json:"metas"`
	BvID               string              `json:"bvid,omitempty"`
	Author             *seasonApi.Author   `json:"author"`
	AuthorDesc         string              `json:"author_desc,omitempty"`
	BadgeStyle         *viewApi.BadgeStyle `json:"badge_style,omitempty"`
	NeedPay            bool                `json:"need_pay,omitempty"`
	EpisodePay         bool                `json:"episode_pay,omitempty"`
	FreeWatch          bool                `json:"free_watch,omitempty"`
	FirstFrame         string              `json:"first_frame,omitempty"`
}

// FromEpisode builds app-view's episode from archive-service's episode
func (v *Episode) FromEpisode(ep *seasonApi.Episode) {
	if ep == nil {
		return
	}
	v.Id = ep.ID
	v.Aid = ep.Aid
	v.Cid = ep.Cid
	v.Title = ep.Title
	if ep.Arc != nil {
		v.Cover = ep.Arc.Pic
		v.Stat = ep.Arc.Stat
		v.FirstFrame = ep.Arc.FirstFrame
		v.CoverRightText = cardmdl.PubDataString(ep.Arc.PubDate.Time())
		v.Author = ep.Arc.Author
		v.AuthorDesc = ep.Arc.Author.GetName()
		if ep.Arc.AttrVal(api.AttrBitIsCooperation) == int64(api.AttrYes) {
			v.AuthorDesc = fmt.Sprintf("%s 等联合创作", v.AuthorDesc)
		}
	}
	// 如果是免费观看
	if ep.AttrVal(seasonApi.EpisodeAttrSnFreeWatch) == seasonApi.AttrSnYes {
		v.FreeWatch = true
	}
	if ep.Page != nil {
		v.ArcPage = ep.Page
		v.Metas = BuildMetas(v.ArcPage.Duration)
	}
	v.BvID, _ = bvid.AvToBv(ep.Aid)
}

func FormatOrderButton(title string, selected bool) string {
	if title != "" {
		return title
	}
	if selected {
		return "已预约"
	} else {
		return "预约"
	}
}

func FormatFavButton(title string, selected bool) string {
	if title != "" {
		return title
	}
	if selected {
		return "已订阅"
	} else {
		return "订阅"
	}
}

func (v *UgcSeason) FormGoodInfo(arc *api.GoodsInfo) {
	v.GoodsInfo.GoodsId = arc.GoodsId
	v.GoodsInfo.Category = viewApi.Category_CategorySeason
	v.GoodsInfo.GoodsPrice = arc.GoodsPrice
	if arc.PayState == api.PayState_PayStateActive {
		v.GoodsInfo.PayState = viewApi.PayState_PayStateActive
	} else {
		v.GoodsInfo.PayState = viewApi.PayState_PayStateUnknown
	}
	v.GoodsInfo.GoodsName = arc.GoodsName
	v.GoodsInfo.PriceFmt = arc.GoodsPriceFmt
}

func (v *UgcSeason) NewPayedButton() {
	// 已购买
	if v.GoodsInfo.PayState == viewApi.PayState_PayStateActive {
		v.PayButton = viewApi.ButtonStyle{
			Text:           _payButtonText,
			TextColor:      _payButtonTextColor,
			TextColorNight: _payButtonTextNightColor,
			BgColor:        _payButtonBgColor,
			BgColorNight:   _payButtonBgNightColor,
			JumpLink:       "",
		}
		return
	}
	// 未购买
	v.PayButton = viewApi.ButtonStyle{
		Text:           _unPayButtonText,
		TextColor:      _unPayButtonTextColor,
		TextColorNight: _unPayButtonTextNightColor,
		BgColor:        _unPayButtonBgColor,
		BgColorNight:   _unPayButtonBgNightColor,
		JumpLink:       "",
	}
}

func (v *UgcSeason) UpdateEpisodePayState() {
	for _, sec := range v.Sections {
		for _, e := range sec.Episodes {
			// 属于付费合集类型
			e.EpisodePay = true
			// 如果未支付且不免费，则下发角标
			if v.GoodsInfo.PayState == viewApi.PayState_PayStateUnknown && !e.FreeWatch {
				e.NeedPay = true
				// 设置付费角标
				e.BadgeStyle = &viewApi.BadgeStyle{
					Text:             "付费",
					TextColor:        "#FFFFFF",
					TextColorNight:   "#E5E5E5",
					BgColor:          "#FF6699",
					BgColorNight:     "#D44E7D",
					BorderColor:      "#FF6699",
					BorderColorNight: "#D44E7D",
					BgStyle:          1,
				}
			}
		}
	}
}
