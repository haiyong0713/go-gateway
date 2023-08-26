package like

import (
	xtime "go-common/library/time"
)

const (
	AsyncReserveTypeOfInsert = 1 << iota
	AsyncReserveTypeOfUpdate
)

const (
	ActSubjectStateNormal = int64(1)
	ActSubjectStateAudit  = int64(0)
	ActSubjectStateCancel = int64(-1)
	ActSubjectStateEdit   = int64(-2)
	ActSubjectStateReject = int64(-3)
)

// B端稿件状态 因为bapis没有枚举 也没有收拢统一业务出口 暂时copy到这使用
const (
	// StateOpen 开放浏览
	StateOpen = 0
	// StateOrange 橙色通过
	StateOrange = 1
	// StateForbidWait 待审
	StateForbidWait = -1
	// StateForbidRecycle 被打回
	StateForbidRecycle = -2
	// StateForbidPolice 网警锁定
	StateForbidPolice = -3
	// StateForbidLock 被锁定
	StateForbidLock = -4
	// StateForbidFackLock 管理员锁定（可浏览）
	StateForbidFackLock = -5
	// StateForbidFixed 修复待审
	StateForbidFixed = -6
	// StateForbidLater 暂缓审核
	StateForbidLater = -7
	// StateForbidPatched 补档待审
	StateForbidPatched = -8
	// StateForbidWaitXcode 等待转码
	StateForbidWaitXcode = -9
	// StateForbidAdminDelay 延迟审核
	StateForbidAdminDelay = -10
	// StateForbidFixing 视频源待修
	StateForbidFixing = -11
	// StateForbidStorageFail 转储失败
	StateForbidStorageFail = -12
	// StateForbidOnlyComment 允许评论待审
	StateForbidOnlyComment = -13
	// StateForbidTmpRecicle 临时回收站
	StateForbidTmpRecicle = -14
	// StateForbidDispatch 分发中
	StateForbidDispatch = -15
	// StateForbidXcodeFail 转码失败
	StateForbidXcodeFail = -16
	// StateWaitEventOpen  已通过审核等待第三方通知开放
	StateWaitEventOpen = -20 // NOTE:spell body can judge to change state
	// StateForbidSubmit 创建已提交
	StateForbidSubmit = -30
	// StateForbidUserDelay 定时发布
	StateForbidUserDelay = -40
	// StateForbidUpDelete 用户删除
	StateForbidUpDelete = -100
)

const (
	UpActReserveAuditPass   = int64(1)
	UpActReserveAuditReject = int64(2)
)

const (
	UpActReserveReject         = int64(-3)
	UpActReserveAudit          = int64(-2)
	UpActReservePassDelayAudit = int64(-1)
	UpActReservePass           = int64(0)
)

const (
	UpActReserveAuditChannelDefault  = int64(0)
	UpActReserveAuditChannelPlatform = int64(1) // 等待审核平台
	UpActReserveAuditChannelArchive  = int64(2) // 等待稿件过审
)

const SpecialPeriodMustAuditFrom = int64(0)

const UpActReserveLivePrefix = "直播预约："

const (
	UpActReserveDependAudit = 1 // 审核通过
)

const (
	DynamicLotteryLiveBizID = 10 // 直播预约
	DynamicLotteryArcBizID  = 11 // 稿件预约
)

// reserve info(noasync)
type AsyncReserve struct {
	PrimaryKey int64  `json:"primary_key"`
	OpType     int    `json:"op_type"`
	Timestamp  int64  `json:"timestamp"`
	TraceID    string `json:"trace_id"`

	*ActReserve
}

// ActReserve .
type ActReserve struct {
	ID          int64         `json:"id"`
	Sid         int64         `json:"sid"`
	Mid         int64         `json:"mid"`
	Num         int32         `json:"num"`
	State       int32         `json:"state"`
	Ctime       xtime.Time    `json:"ctime"`
	Mtime       xtime.Time    `json:"mtime"`
	IPv6        []byte        `json:"ipv6"`
	Score       int64         `json:"score"`
	AdjustScore int64         `json:"adjust_score"`
	Order       int64         `json:"order"`
	Report      ReserveReport `json:"report"`
}

type ReserveReport struct {
	From     string `json:"from"`
	Typ      string `json:"typ"`
	Oid      string `json:"oid"`
	Ip       string `json:"ip"`
	Platform string `json:"platform"`
	Mobiapp  string `json:"mobiapp"`
	Buvid    string `json:"buvid"`
	Spmid    string `json:"spmid"`
}

type HTTPReserveReport struct {
	From     string `json:"from" form:"from"`
	Typ      string `json:"typ" form:"typ"`
	Oid      string `json:"oid" form:"oid"`
	Platform string `json:"platform" form:"platform"`
	Mobiapp  string `json:"mobiapp" form:"mobiapp"`
	Buvid    string `json:"buvid" form:"buvid"`
	Spmid    string `json:"spmid" form:"spmid"`
}

// ActFollowingReply .
type ActFollowingReply struct {
	IsFollowing bool       `json:"is_following"`
	Total       int64      `json:"total"`
	ReserveID   int64      `json:"reserve_id"`
	Mtime       xtime.Time `json:"mtime"`
	Ctime       xtime.Time `json:"ctime"`
	Order       int64      `json:"order"`
}

// SubStat .
type SubStat struct {
	Sid int64 `json:"sid"`
	Num int64 `json:"num"`
}

type RelationReserveInfo struct {
	State     int64                      `json:"state"`
	Total     int64                      `json:"total"`
	List      []*RelationReserveInfoItem `json:"list"`
	StartTime int64                      `json:"start_time"`
	EndTime   int64                      `json:"end_time"`
	ActStatus int64                      `json:"act_status"`
	SID       int64                      `json:"sid"`
}

type RelationReserveInfoItem struct {
	Sid       int64 `json:"sid"`
	Total     int64 `json:"total"`
	State     int64 `json:"state"`
	StartTime int64 `json:"start_time"`
	EndTime   int64 `json:"end_time"`
	ActStatus int64 `json:"act_status"`
}

type RelationReserveConfig struct {
	StartTime int64 `json:"start_time"`
	EndTime   int64 `json:"end_time"`
}

type RelationFollowConfig struct {
	StartTime int64 `json:"start_time"`
	EndTime   int64 `json:"end_time"`
}

type RelationSeasonConfig struct {
	StartTime int64 `json:"start_time"`
	EndTime   int64 `json:"end_time"`
}

type RelationMallConfig struct {
	StartTime int64 `json:"start_time"`
	EndTime   int64 `json:"end_time"`
}

type RelationTopicConfig struct {
	StartTime int64 `json:"start_time"`
	EndTime   int64 `json:"end_time"`
}

type RelationFavoriteConfig struct {
	StartTime int64 `json:"start_time"`
	EndTime   int64 `json:"end_time"`
}

type RelationFavoriteInfoItem struct {
	Type    int64  `json:"type"`
	Content string `json:"content"`
}

type CreateUpActReserveArgs struct {
	Title             string `form:"title"`
	Stime             int64  `form:"stime"`
	Etime             int64  `form:"etime"`
	Type              int64  `form:"type"`
	From              int64  `form:"from" validate:"required"`
	LivePlanStartTime int64  `form:"live_plan_start_time"`
	Oid               string `form:"oid"`
	CreateDynamic     int64  `form:"create_dynamic"`
	LotteryID         string `form:"lottery_id"`
	LotteryType       int64  `form:"lottery_type"`
}

type CreateUpActReserveItem struct {
	Name  string
	Stime xtime.Time
	Etime xtime.Time
	Type  int64
	State int64
}

type UpdateUpActReserveArgs struct {
	ID int64 `form:"id" validate:"required"`
	CreateUpActReserveArgs
}

type UpActReserveInfo struct {
	ID                int64      `json:"id"`
	Title             string     `json:"title"`
	Stime             xtime.Time `json:"stime"`
	Etime             xtime.Time `json:"etime"`
	Type              int64      `json:"type"`
	LivePlanStartTime xtime.Time `json:"live_plan_start_time"`
	LotteryType       int64      `json:"lottery_type"`
	LotteryID         string     `json:"lottery_id"`
}

type UpActReserveRelationContinuingArg struct {
	Type      int64  `form:"type"`
	From      int64  `form:"from"`
	InstantID string `form:"instant_id"`
}

type UpActReserveRelationOthersArg struct {
	Type int64 `form:"type"`
	From int64 `form:"from"`
}

type ReserveCounterGroupItem struct {
	ID          int64      `json:"id" form:"id" gorm:"column:id"`
	Sid         int64      `json:"sid" form:"sid" gorm:"column:sid" validate:"min=1"`
	GroupName   string     `json:"group_name" form:"group_name" gorm:"column:group_name" validate:"required"`
	Dim1        int64      `json:"dim1" form:"dim1" gorm:"column:dim1" validate:"min=1"`
	Dim2        int64      `json:"dim2" form:"dim2" gorm:"column:dim2" validate:"min=1"`
	Threshold   int64      `json:"threshold" form:"threshold" gorm:"column:threshold"`
	CounterInfo string     `json:"counter_info" form:"counter_info" gorm:"column:counter_info" validate:"required"`
	Author      string     `json:"author" form:"author" gorm:"column:author" validate:"required"`
	Ctime       xtime.Time `json:"ctime" time_format:"2006-01-02 15:04:05" gorm:"column:ctime"`
	Mtime       xtime.Time `json:"mtime" time_format:"2006-01-02 15:04:05" gorm:"column:mtime"`
}

type ReserveCounterNodeItem struct {
	ID       int64      `json:"id" form:"id" gorm:"column:id"`
	Sid      int64      `json:"sid" gorm:"column:sid"`
	GroupID  int64      `json:"group_id" gorm:"column:group_id"`
	NodeName string     `json:"node_name" form:"node_name" gorm:"column:node_name" validate:"required"`
	NodeVal  int64      `json:"node_val" form:"node_val" gorm:"column:node_val" validate:"required"`
	Ctime    xtime.Time `json:"ctime" time_format:"2006-01-02 15:04:05" gorm:"column:ctime"`
	Mtime    xtime.Time `json:"mtime" time_format:"2006-01-02 15:04:05" gorm:"column:mtime"`
}

type CreateUpActReserveExtra struct {
	LivePlanStartTime xtime.Time
	Oid               string
	From              int64
	Audit             int64
	AuditChannel      int64
	LotteryID         string
	LotteryType       int64
	LotteryAudit      int64
}

type UpActReserveRelationUpdateFields struct {
	Sid               int64
	Mid               int64
	SubjectState      int64
	RelationState     int64
	AuditState        int64
	AuditChannelState int64
	DynamicID         string
}

type UpActReserveRelationInfoArgs struct {
	IDs string `form:"ids"`
}
