package selected

import (
	"go-gateway/app/app-svr/app-feed/admin/model/common"
)

const (
	ResourceTypeAv       = "av" //
	ResourceStatusOn     = 1    // 资源已启用
	ResourceStatusReject = 2    // 资源被禁用
	ResourceSourceAI     = 1    // 资源来源：AI
)

// ReqSelAdd def.
type ReqSelAdd struct {
	Type       string `form:"type" validate:"required"`
	Number     int64  `form:"number" validate:"required"`
	RID        string `form:"rid" validate:"required"`
	Rtype      string `form:"rtype" validate:"required"`
	RcmdReason string `form:"rcmd_reason"`
}

// ReqSelEdit def.
type ReqSelEdit struct {
	RID        string `form:"rid"   validate:"required"`
	Rtype      string `form:"rtype" validate:"required"`
	RcmdReason string `form:"rcmd_reason"`
	ID         int64  `form:"id"    validate:"required"`
	RIDInt     int64
}

// ToMap eases the update by ORM
func (v *ReqSelEdit) ToMap(origin *Resource) map[string]interface{} {
	update := make(map[string]interface{}, 3)
	update["rid"] = v.RIDInt
	update["rtype"] = v.Rtype
	update["rcmd_reason"] = v.RcmdReason
	update["status"] = 1 // 编辑过的卡片自动通过
	if origin.Source == _sourceAI {
		update["source"] = 3 // AI来源变为运营编辑
	}
	return update
}

// Resource def.
type Resource struct {
	ID         int64  `json:"id" gorm:"column:id"`
	RID        int64  `json:"rid" gorm:"column:rid"`
	Rtype      string `json:"rtype"`
	SerieID    int64  `json:"serie_id" gorm:"column:serie_id"`
	Source     int    `json:"source"`
	Creator    string `json:"creator"`
	Position   int64  `json:"position"`
	RcmdReason string `json:"rcmd_reason" gorm:"column:rcmd_reason"`
	Status     int    `json:"status"`
}

// Rejected def.
func (v *Resource) Rejected() bool {
	return v.Status == ResourceStatusReject
}

// Operator def.
type Operator struct {
	UID   int64
	Uname string
}

// FromReq def.
func (v *Resource) FromReq(req *ReqSelAdd, sid int64, creator string) {
	v.RID, _ = common.GetAvID(req.RID)
	v.Rtype = req.Rtype
	v.RcmdReason = req.RcmdReason
	v.SerieID = sid
	v.Position = 1 // 新增卡片置顶
	v.Status = 1   // 新增卡片默认通过
	v.Creator = creator
	v.Source = _sourceOPAdd
}

// TableName def.
func (v *Resource) TableName() string {
	return "selected_resource"
}

// SelSortReq def.
type SelSortReq struct {
	Type    string  `form:"type" validate:"required"`
	Number  int64   `form:"number" validate:"required"`
	CardIDs []int64 `form:"card_ids,split" validate:"required,dive,gt=0"`
}
