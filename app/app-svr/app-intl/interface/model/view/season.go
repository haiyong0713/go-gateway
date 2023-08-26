package view

import (
	"fmt"

	cardmdl "go-gateway/app/app-svr/app-card/interface/model"
	seasonApi "go-gateway/app/app-svr/ugc-season/service/api"
	"go-gateway/pkg/idsafe/bvid"
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
}

// FromSeason def.
func (v *UgcSeason) FromSeason(snv *seasonApi.View) {
	if snv == nil || snv.Season == nil {
		return
	}
	season := snv.Season
	v.Id = season.ID
	v.Title = season.Title
	v.Cover = season.Cover
	v.Intro = season.Intro
	v.Stat = season.Stat
	for _, section := range snv.Sections {
		sectionView := new(Section)
		sectionView.FromSection(section)
		v.Sections = append(v.Sections, sectionView)
	}
	v.EpCount = season.EpCount
	if season.SignState == _signExlusive || season.SignState == _signFirst { // 签约状态和style
		v.LabelTextColor = _signTextColor
		v.LabelBgColor = _signTextBgColor
		v.LabelTextNightColor = _signTextNightColor
		v.LabelBgNightColor = _signTextBgNightColor
		if season.SignState == _signExlusive {
			v.LabelText = _signExText
		} else {
			v.LabelText = _signFtText
		}
	}
	if season.AttrVal(seasonApi.AttrSnFinished) == seasonApi.AttrSnYes && season.EpCount > 0 { // 是否完结
		v.DescRight = fmt.Sprintf(_finishedFmt, season.EpCount)
	} else {
		v.DescRight = _notFinishedText
	}
}

// Section def.
type Section struct {
	Id       int64      `json:"id"`
	Title    string     `json:"title"`
	Type     int64      `json:"type"`
	Episodes []*Episode `json:"episodes,omitempty"`
}

// FromEpisode builds app-intl's section from archive-service's section
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
	Id                 int64              `json:"id"`
	Aid                int64              `json:"aid"`
	Cid                int64              `json:"cid"`
	Title              string             `json:"title"` // ep's title
	Cover              string             `json:"cover"` // arc's cover
	CoverRightText     string             `json:"cover_right_text,omitempty"`
	*seasonApi.ArcPage                    // page info
	Stat               *seasonApi.ArcStat `json:"stat"` // archive's stat
	Metas              []*Meta            `json:"metas"`
	BvID               string             `json:"bvid,omitempty"`
}

// FromEpisode builds app-intl's episode from archive-service's episode
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
		v.CoverRightText = cardmdl.PubDataString(ep.Arc.PubDate.Time())
	}
	if ep.Page != nil {
		v.ArcPage = ep.Page
		v.Metas = BuildMetas(v.ArcPage.Duration)
	}
	v.BvID, _ = bvid.AvToBv(ep.Aid)
}
