package selected

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strconv"

	"go-common/library/time"

	"git.bilibili.co/bapis/bapis-go/archive/service"
	"go-gateway/app/app-svr/app-feed/admin/model/common"
)

const (
	_rtypeAv = "av"

	SERIE_TYPE_WEEKLY_SELECTED = "weekly_selected" // 每周必看类型
)

// FindSerie def.
type FindSerie struct {
	Type   string
	Number int64
	ID     int64
}

// Serie def.
type Serie struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name" gorm:"-"`
	Type          string    `json:"type"`
	Number        int64     `json:"number"`
	Stime         time.Time `json:"stime"`
	Etime         time.Time `json:"etime"`
	Pubtime       time.Time `json:"pubtime"`
	Hint          string    `json:"hint"`
	Subject       string    `json:"subject"`
	Color         int       `json:"color"`
	Cover         string    `json:"cover"`
	ShareTitle    string    `json:"share_title"`
	ShareSubtitle string    `json:"share_subtitle"`
	PushTitle     string    `json:"push_title"`
	PushSubtitle  string    `json:"push_subtitle"`
	Status        int       `json:"status"`
	TaskStatus    int       `json:"task_status" gorm:"column:task_status"`
	MediaID       int64     `json:"media_id" gorm:"column:media_id"`
}

// SerieFilter is used to filter the series
type SerieFilter struct {
	Number int64  `json:"number"`
	Name   string `json:"name"`
}

// FromSerie builds the SerieFilter
func (v *SerieFilter) FromSerie(sr *Serie) {
	v.Number = sr.Number
	v.Name = sr.SerieName()
}

// SerieName Name def. 2018第2期 01.15 - 01.22
func (v *Serie) SerieName() string {
	return v.Stime.Time().Format("2006") +
		fmt.Sprintf("第%d期 %s - %s", v.Number, v.Stime.Time().Format("01.02"), v.Etime.Time().Format("01.02"))
}

// SelResReq def.
type SelResReq struct {
	Number  int64  `form:"number"`
	Type    string `form:"type" validate:"required"`
	RID     string `form:"rid"`
	Rtitle  string `form:"rtitle"`
	Author  string `form:"author"`
	MID     int64  `form:"mid"`
	Creator string `form:"creator"`
	Status  int    `form:"status"`
	Pn      int    `form:"pn" default:"1"`
	Ps      int    `form:"ps" default:"20"`
}

// TableName def.
func (v *Serie) TableName() string {
	return "selected_serie"
}

// SelResReply def.
type SelResReply struct {
	Result []*SelShow `json:"result"`
	Page   *Page      `json:"page,omitempty"`
}

func (v *SelResReply) GetBvid() {
	if v != nil {
		for i := 0; i < len(v.Result); i++ {
			v.Result[i].Bvid, _ = common.GetBvID(v.Result[i].RID)
		}
	}
}

// PreviewReq def.
type PreviewReq struct {
	Type   string `form:"type" validate:"required"`
	Number int64  `form:"number" validate:"required"`
}

// Name def.
func (v *PreviewReq) Name() string {
	return fmt.Sprintf("selected_%d_%s", v.Number, v.Type)
}

// OpReq def.
type OpReq struct {
	ID int64 `form:"id" validate:"required"` // selected_resource表ID
}

// SelRes def.
type SelRes struct {
	Goto            string `json:"goto"`
	Param           int64  `json:"param"`
	Cover           string `json:"cover"`
	Title           string `json:"title"`
	CoverRightText1 string `json:"cover_right_text_1"`
	RightDesc1      string `json:"right_desc_1"`
	RightDesc2      string `json:"right_desc_2"`
	RcmdReason      string `json:"rcmd_reason"`
}

// TableName def.
func (v *SelRes) TableName() string {
	return "selected_resource"
}

// FromArc builds the selected resource body
func (v *SelRes) FromArc(arc *api.Arc, reason string) {
	v.Param = arc.Aid
	v.RcmdReason = reason
	v.Goto = _rtypeAv
	v.Cover = arc.Pic
	v.Title = arc.Title
	v.CoverRightText1 = DurationString(arc.Duration)
	v.RightDesc1 = arc.Author.Name
	v.RightDesc2 = ArchiveViewString(arc.Stat.View) + " · " + PubDataString(arc.PubDate.Time())
}

// SelPreview def.
type SelPreview struct {
	Config *Serie    `json:"config"`
	List   []*SelRes `json:"list"`
}

// SerieEditReq def.
type SerieEditReq struct {
	PreviewReq
	SerieDB
}

// SerieDB def.
type SerieDB struct {
	ID int64 `gorm:"column:id"`
	//	Hint          string `form:"hint" validate:"required"`
	Subject string `form:"subject" validate:"required"`
	//	Color         int    `form:"color" validate:"required"`
	// Cover         string `form:"cover" validate:"required"`
	ShareTitle    string `form:"share_title" validate:"required" gorm:"column:share_title"`
	ShareSubtitle string `form:"share_subtitle" validate:"required" gorm:"column:share_subtitle"`
}

// SeriePush def.
type SeriePush struct {
	ID           int64  `form:"id" validate:"required" gorm:"column:id"`
	PushTitle    string `form:"push_title" validate:"required" gorm:"column:push_title"`
	PushSubtitle string `form:"push_subtitle" validate:"required" gorm:"column:push_subtitle"`
}

// IsArc def.

func (v *Resource) IsArc() bool {
	return v.Rtype == _rtypeAv
}

// PushBody returns the push title of the serie
func (v *Serie) PushBody() string {
	return fmt.Sprintf("%d年第%d期 | ", v.Stime.Time().Year(), v.Number)
}

// 播单标题
func (v *Serie) MediaListTitle() string {
	return fmt.Sprintf("「%s %d年第%d期」", v.Subject, v.Stime.Time().Year(), v.Number)
}

// UUID returns the unique string of the serie to avoid push repeatedly
func (v *Serie) UUID(mid int64) string {
	var b bytes.Buffer
	b.WriteString(strconv.FormatInt(mid, 10))
	b.WriteString(strconv.FormatInt(v.ID, 10))
	b.WriteString(v.Type)
	b.WriteString(strconv.FormatInt(v.Number, 10))
	mh := md5.Sum(b.Bytes())
	return hex.EncodeToString(mh[:])
}

// SerieFull is the full structure of one serie in MC
type SerieFull struct {
	Config *SerieConfig   `json:"config"`
	List   []*SelectedRes `json:"list"`
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

// SelectedRes represents selected resources
type SelectedRes struct {
	RID        int64  `json:"rid"`
	Rtype      string `json:"rtype"`
	SerieID    int64  `json:"serie_id"`
	Position   int    `json:"position"`
	RcmdReason string `json:"rcmd_reason"`
}

type SelResSimple struct {
	Aid        int64  `json:"aid"`
	RcmdReason string `json:"rcmd_reason"`
}

type LatestSelPreviewReply struct {
	ID     int64           `json:"id"`
	Number int64           `json:"number"`
	List   []*SelResSimple `json:"list"`
}
