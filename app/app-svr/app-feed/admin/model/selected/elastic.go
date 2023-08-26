package selected

import (
	"fmt"

	"go-gateway/app/app-svr/app-feed/admin/model/common"
)

const (
	_notCutPS     = 1000
	_sourceAI     = 1
	_sourceOPAdd  = 2
	_sourceOPEdit = 3
	_statusPass   = 1
	_statusReject = 2
)

// ReqSelES def.
type ReqSelES struct {
	Pn      int
	Ps      int
	AID     int64
	Title   string
	Author  string
	Mid     int64
	Creator string
	SerieID int64
	Status  int
}

// FromHTTP def
func (v *ReqSelES) FromHTTP(param *SelResReq, sid int64) {
	v.SerieID = sid
	v.Pn = param.Pn
	v.Ps = param.Ps
	v.AID, _ = common.GetAvID(param.RID)
	v.Title = param.Rtitle
	v.Author = param.Author
	v.Mid = param.MID
	v.Creator = param.Creator
	v.Status = param.Status
}

// OneSerie def.
func (v *ReqSelES) OneSerie() {
	if v.SerieID != 0 {
		v.Pn = 1
		v.Ps = _notCutPS
	}
}

// SelES is the result of sel resources es
type SelES struct {
	Author     string `json:"author" gorm:"-"`
	Cover      string `json:"cover" gorm:"-"`
	Creator    string `json:"creator" gorm:"-"`
	ID         int64  `json:"id" gorm:"column:id"`
	MID        int64  `json:"mid" gorm:"-"`
	Position   int64  `json:"position" gorm:"column:position"`
	RcmdReason string `json:"rcmd_reason" gorm:"column:rcmd_reason"`
	RID        int64  `json:"rid" gorm:"column:rid"`
	SerieID    int64  `json:"serie_id" gorm:"column:serie_id"`
	Source     int    `json:"source"`
	Status     int    `json:"status"`
	Title      string `json:"title" gorm:"-"`
}

// TableName def.
func (v *SelES) TableName() string {
	return "selected_resource"
}

// SelShow is the plus versin of selES
type SelShow struct {
	*SelES
	SerieTitle string `json:"serie_name"`
	SourceName string `json:"source_name"` // 翻译成中文，便于导出
	StatusName string `json:"status_name"` // 翻译成中文，便于导出
	Zone       string `json:"zone"`        // 稿件一级分区名
	IsNormal   bool   `json:"is_normal"`
	Bvid       string `json:"bvid,omitempty"`
	IsHotDown  bool   `json:"is_hot_down"`
	IsNoHot    bool   `json:"is_no_hot"`
}

// Export def.
func (v *SelShow) Export() []string {
	return []string{v.SerieTitle, fmt.Sprintf("%d", v.Position),
		fmt.Sprintf("%d", v.RID), v.Title, fmt.Sprintf("%d", v.MID), v.Author, v.Zone, v.SourceName, v.RcmdReason, v.StatusName}
}

// FromES def.
func (v *SelShow) FromES(es *SelES, stitle string, zone string, cover string) {
	v.SelES = es
	switch v.Source {
	case _sourceAI:
		v.SourceName = "AI"
	case _sourceOPAdd:
		v.SourceName = "运营新建"
	case _sourceOPEdit:
		v.SourceName = "运营编辑"
	}
	switch v.Status {
	case _statusPass:
		v.StatusName = "已通过"
	case _statusReject:
		v.StatusName = "已拒绝"
	}
	v.Zone = zone
	v.SerieTitle = stitle
	if cover != "" {
		v.Cover = cover
	}
}

// Page represents the standard page structure
type Page struct {
	Num   int `json:"num"`
	Size  int `json:"size"`
	Total int `json:"total"`
}

// SelESReply def.
type SelESReply struct {
	Result []*SelES `json:"result"`
	Page   *Page    `json:"page"`
}
