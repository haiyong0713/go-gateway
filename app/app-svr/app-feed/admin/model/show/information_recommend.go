package show

import (
	"time"

	xtime "go-common/library/time"
	"go-gateway/app/app-svr/app-feed/admin/model/common"
)

const (
	StatusDeleted     = "deleted"
	StatusToAudit     = "to_audit"
	StatusAuditPass   = "audit_pass"
	StatusAuditReject = "audit_reject"
	StatusOnline      = "online"
	StatusOffline     = "offline"
	StatusUnknown     = "unknown"

	AuditStatusToAudit = 1
	AuditStatusPass    = 2
	AuditStatusReject  = 3

	OnlineStatus  = 0
	OfflineStatus = 1

	OpPass   = "pass"
	OpReject = "reject"

	// 视频卡
	CardTypeAv = 1
	// 动态卡
	CardTypeDynamic = 2
	// 专栏卡
	CardTypeArticle = 3
)

var (
	CardPreviewType = map[int]string{
		CardTypeAv:      common.CardAv,
		CardTypeDynamic: common.CardDynamic,
		CardTypeArticle: common.CardArticle,
	}
)

// RecommendCard
type RecommendCard struct {
	ID            int64      `json:"id" gorm:"column:id"`
	CardType      int        `json:"card_type" gorm:"column:card_type"`
	CardID        string     `json:"card_id" gorm:"column:card_id"`
	AvID          int64      `json:"av_id"`
	BvID          string     `json:"bv_id,omitempty"`
	CardTitle     string     `json:"card_title"`
	CardPos       int        `json:"card_pos" gorm:"column:card_pos"`
	PosIndex      int        `json:"pos_index" gorm:"column:pos_index"`
	Stime         xtime.Time `json:"stime" gorm:"column:stime" time_format:"2006-01-02 15:04:05"`
	Etime         xtime.Time `json:"etime" gorm:"column:etime" time_format:"2006-01-02 15:04:05"`
	Uid           int64      `json:"uid" gorm:"column:uid"`
	Uname         string     `json:"uname" gorm:"column:uname"`
	IsCover       int        `json:"is_cover" gorm:"column:is_cover"`
	ApplyReason   string     `json:"apply_reason" gorm:"column:apply_reason"`
	CoverImg      string     `json:"cover_img" gorm:"cover_img"`
	Status        string     `json:"status"`
	AuditStatus   int        `json:"audit_status" gorm:"column:audit_status"`
	OfflineStatus int        `json:"offline_status" gorm:"column:offline_status"`
	IsDeleted     int        `gorm:"column:is_deleted"`
	Ctime         xtime.Time `json:"ctime" gorm:"column:ctime" time_format:"2006-01-02 15:04:05"`
	Mtime         xtime.Time `json:"mtime" gorm:"column:mtime" time_format:"2006-01-02 15:04:05"`
}

// tag string to tags list
func (c *RecommendCard) StatusVal() (status string) {
	status = StatusUnknown
	now := time.Now().Unix()
	if c.IsDeleted == 1 {
		status = StatusDeleted
		return
	}
	// 已失效 = 被人工下线，或者结束时间已过
	if c.OfflineStatus == OfflineStatus || c.Etime.Time().Unix() < now {
		status = StatusOffline
		return
	}
	// 已拒绝
	if c.AuditStatus == AuditStatusReject {
		status = StatusAuditReject
		return
	}
	// 已通过 = 生效时间未到，已通过，未下线
	if c.AuditStatus == AuditStatusPass && c.OfflineStatus == 0 && c.Stime.Time().Unix() > now {
		status = StatusAuditPass
		return
	}
	// 待审核
	if c.AuditStatus == AuditStatusToAudit {
		status = StatusToAudit
		return
	}
	// 已生效
	if c.OfflineStatus == OnlineStatus && c.AuditStatus == AuditStatusPass && c.Stime.Time().Unix() <= now && c.Etime.Time().Unix() >= now {
		status = StatusOnline
		return
	}
	return
}

// RecommendCardList
type RecommendCardList struct {
	List []*RecommendCard `json:"list"`
	Page common.Page      `json:"page"`
}

// TableName .
func (RecommendCard) TableName() string {
	return "information_recommend_card"
}

// RecommendCardSearchReq
type RecommendCardListReq struct {
	CardType int `form:"card_type"`
	// 卡片id
	CardID string `form:"card_id"`
	AvID   int64
	// 展示开始时间
	Stime xtime.Time `form:"stime"`
	// 展示结束时间
	Etime xtime.Time `form:"etime"`
	// 创建人名
	Uname string `form:"uname"`
	// 卡片状态："all"-全部，"to_audit"-待审核，"audit_pass"-已通过，"audit_reject"-已拒绝，"online"-已生效，"offline"-已失效
	Status string `form:"status"`
	// 分页大小
	Ps int `form:"ps" default:"20"`
	// 第几个分页
	Pn int `form:"pn" default:"1"`
}

// RecommendCardAddReq
type RecommendCardAddReq struct {
	CardType    int        `form:"card_type" validate:"required,gte=1,lte=3"`
	CardID      string     `form:"card_id" gorm:"column:card_id" validate:"required"`
	AvID        int64      `gorm:"-"`
	CardPos     int        `form:"card_pos" default:"1" validate:"required,gte=1,lte=6"`
	PosIndex    int        `form:"pos_index" default:"1" validate:"required,gte=1,lte=100"`
	Stime       xtime.Time `form:"stime" time_format:"2006-01-02 15:04:05" validate:"required"`
	Etime       xtime.Time `form:"etime" time_format:"2006-01-02 15:04:05" validate:"required"`
	Uid         int64      `form:"uid"`
	Uname       string     `form:"uname"`
	IsCover     int        `form:"is_cover" default:"0"`
	ApplyReason string     `form:"apply_reason" validate:"required,max=16"`
	CoverImg    string     `form:"cover_img"`
}

// RecommendCardModifyReq
type RecommendCardModifyReq struct {
	RecommendCardAddReq
	ID            int64 `form:"id" gorm:"column:id" validate:"required,gte=1"`
	AuditStatus   int   `gorm:"column:audit_status"`
	OfflineStatus int   `gorm:"column:offline_status"`
}

type RecommendCardIntervalCheckReq struct {
	ID       int64
	CardPos  int        `form:"card_pos" default:"1" validate:"required,gte=1,lte=6"`
	PosIndex int        `form:"pos_index" default:"1" validate:"required,gte=1,lte=100"`
	Stime    xtime.Time `form:"stime" time_format:"2006-01-02 15:04:05" validate:"required"`
	Etime    xtime.Time `form:"etime" time_format:"2006-01-02 15:04:05" validate:"required"`
}

// RecommendCardOpReq
type RecommendCardOpReq struct {
	ID       int64 `form:"id" gorm:"column:id" validate:"required,gte=1"`
	CardType int
	CardID   string
	CardPos  int
	PosIndex int
	Stime    xtime.Time
	Etime    xtime.Time
	Uid      int64  `form:"uid"`
	Uname    string `form:"uname"`
	Op       string
}

// tag string to tags list
func (req *RecommendCardListReq) AvIDVal() (avid int64, err error) {
	avid = 0
	if req.CardID == "" {
		return
	}
	if avid, err = common.GetAvID(req.CardID); err != nil {
		return
	}
	return
}
