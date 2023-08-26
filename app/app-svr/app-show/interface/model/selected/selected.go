package selected

import (
	"fmt"

	"go-common/library/time"

	"go-gateway/app/app-svr/app-card/interface/model/card"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
)

const (
	_statusDiaster = 4
	_colorWhite    = 2
)

// SelectedParam def
type SelectedParam struct {
	Type    string `form:"type" validate:"required"`
	Number  int64  `form:"number" validate:"required"`
	MobiApp string `form:"mobi_app"`
	Device  string `form:"device"`
}

// SerieCore is the core fields of selected serie
type SerieCore struct {
	ID      int64     `json:"-"`
	Type    string    `json:"-"`
	Number  int64     `json:"number"`
	Subject string    `json:"subject"`
	Stime   time.Time `json:"-"`
	Etime   time.Time `json:"-"`
	Status  int       `json:"status"`
}

// SerieFilter is the unit that the user chooses in the first page. It's also the unit in Redis Zset
type SerieFilter struct {
	SerieCore
	Name string `json:"name"`
}

// BuildName def: 2018第2期 01.15 - 01.22
func (v *SerieFilter) Init() {
	v.Name = v.Stime.Time().Format("2006") +
		fmt.Sprintf("第%d期 %s - %s", v.Number, v.Stime.Time().Format("01.02"), v.Etime.Time().Format("01.02"))
	if v.Status == _statusDiaster { // 灾备数据
		v.Subject = "哔哩哔哩每周必看"
	}
}

// SerieConfig is the structure in the selected series API
type SerieConfig struct {
	SerieCore
	Label         string `json:"label"`
	Hint          string `json:"hint"`
	Color         int    `json:"color"`
	Cover         string `json:"cover"`
	ShareTitle    string `json:"share_title"`
	ShareSubtitle string `json:"share_subtitle"`
	MediaID       int64  `json:"media_id"` // 播单ID
}

// Init def:
func (v *SerieConfig) Init() {
	v.Label = fmt.Sprintf("第%d期(%s更新)",
		v.Number, v.Etime.Time().AddDate(0, 0, 1).Format("0102")) // 第8期(0102更新)，用于左上角展示
	if v.Status == _statusDiaster {
		v.Subject = "哔哩哔哩每周必看"
		v.Color = _colorWhite
		v.Hint = v.Stime.Time().Format("2006") + fmt.Sprintf("年第%d期:", v.Number)                 // 2019年第8期
		v.Cover = "http://i0.hdslb.com/bfs/archive/1eed634b8071d0d37dfdb6d68513242ae9e0897b.jpg" // 默认头图
		v.ShareTitle = "「哔哩哔哩每周必看」" + v.Stime.Time().Format("2006") +
			fmt.Sprintf("年第%d期", v.Number) // [哔哩哔哩每周必看]2019年第3期
	}
}

// IsDisaster def.
func (v *SerieConfig) IsDisaster() bool {
	return v.Status == _statusDiaster
}

// SelectedRes represents selected resources
type SelectedRes struct {
	RID        int64  `json:"rid"`
	Rtype      string `json:"rtype"`
	SerieID    int64  `json:"serie_id"`
	Position   int    `json:"position"`
	RcmdReason string `json:"rcmd_reason"`
}

// ToAIItem def.
func (c *SelectedRes) ToAIItem() (a *ai.Item) {
	a = &ai.Item{
		Goto: c.Rtype,
		ID:   c.RID,
	}
	return
}

// SerieFull is the full structure of one serie in MC
type SerieFull struct {
	Config *SerieConfig   `json:"config"`
	List   []*SelectedRes `json:"list"`
}

// SerieShow is the structure of serie to show on the h5 page
type SerieShow struct {
	Config   *SerieConfig   `json:"config"`
	Reminder string         `json:"reminder"`
	List     []card.Handler `json:"list"` // small_cover_h5
}

// FavStatus def.
type FavStatus struct {
	Status bool `json:"status"`
}
