package like

import (
	xtime "go-common/library/time"
	"go-gateway/app/web-svr/activity/job/model"
)

const (
	Publish = iota + 1
	Agree
	Coin
	Watch
)

const (
	RuleStateOnline = iota + 1
	RuleStateFrozen
)

const (
	FLATLISTFORBIDOVERSEA = uint(23)
	FLATLISTFORBIDRCMD    = uint(24)
	FLATLISTFORBIDOTHER   = uint(25)
	// arc attr
	AttrBitNoSearch = uint(4)
)

const (
	// ShieldNoRank 禁止排行
	ShieldNoRank = uint(0)
	// ShieldNoDynamic 禁止首页动态
	ShieldNoDynamic = uint(1)
	// ShieldNoRecommend 禁止推荐
	ShieldNoRecommend = uint(2)
	// ShieldNoHot 禁止热门
	ShieldNoHot = uint(3)
	// ShieldNoFansDynamic 禁止粉丝动态
	ShieldNoFansDynamic = uint(4)
	// ShieldNoSearch 禁止搜索
	ShieldNoSearch = uint(5)
	// ShieldNoOversea 禁止海外
	ShieldNoOversea = uint(6)
	// ArchiveNoRank 稿件禁止排行
	ArchiveNoRank = "norank"
	// ArchiveNoDynamic 禁止首页动态
	ArchiveNoDynamic = "nodynamic"
	// ArchiveNoRecommend 禁止推荐
	ArchiveNoRecommend = "norecommend"
	// ArchiveNoHot 禁止热门
	ArchiveNoHot = "nohot"
	// ArchiveNoFansDynamic 禁止粉丝动态
	ArchiveNoFansDynamic = "push_blog"
	// ArchiveNoSearch 禁止搜索
	ArchiveNoSearch = "nosearch"
	// ArchiveNoOversea 禁止海外
	ArchiveNoOversea = "oversea_block"
	// FlowControlYes 禁止
	FlowControlYes = 1
	// FlowControlNo 不禁止
	FlowControlNo = 0
	// SubjectRuleAidSourceTypeFile 文件
	SubjectRuleAidSourceTypeFile = 1
	// SubjectRuleAidSourceTypeSid 视频数据源
	SubjectRuleAidSourceTypeSid = 2
	// SubjectRuleAidSourceTypeFav 收藏夹数据源
	SubjectRuleAidSourceTypeFav = 3
)

const (
	ActivityReserveAutoPushTypeOne = 1
)

const PlatformActivityBizID = 1001

const UpActReserverelationRetry = 3

const (
	VIDEO             = 1
	PICTURE           = 2
	DRAWYOO           = 3
	VIDEOLIKE         = 4
	PICTURELIKE       = 5
	DRAWYOOLIKE       = 6
	TEXT              = 7
	TEXTLIKE          = 8
	ONLINEVOTE        = 9
	QUESTION          = 10
	LOTTERY           = 11
	ARTICLE           = 12
	VIDEO2            = 13
	MUSIC             = 15
	PHONEVIDEO        = 16
	SMALLVIDEO        = 17
	RESERVATION       = 18
	MISSIONGROUP      = 19
	STORYKING         = 20
	PREDICTION        = 21
	CLOCKIN           = 22
	USERACTIONSTAT    = 23
	UPRESERVATIONARC  = 24
	UPRESERVATIONLIVE = 25
)

// Subject subject
type Subject struct {
	ID       int64      `json:"id"`
	Name     string     `json:"name"`
	Dic      string     `json:"dic"`
	Cover    string     `json:"cover"`
	Stime    xtime.Time `json:"stime"`
	Etime    xtime.Time `json:"etime"`
	Interval int64      `json:"interval"`
	Tlimit   int64      `json:"tlimit"`
	Ltime    int64      `json:"ltime"`
	List     []*Like    `json:"list"`
}

// ActSubject .
type ActSubject struct {
	ID         int64         `json:"id"`
	Oid        int64         `json:"oid"`
	Type       int           `json:"type"`
	State      int           `json:"state"`
	Stime      model.StrTime `json:"stime"`
	Etime      model.StrTime `json:"etime"`
	Ctime      model.StrTime `json:"ctime"`
	Mtime      model.StrTime `json:"mtime"`
	Name       string        `json:"name"`
	Author     string        `json:"author"`
	ActURL     string        `json:"act_url"`
	Lstime     model.StrTime `json:"lstime"`
	Letime     model.StrTime `json:"letime"`
	Cover      string        `json:"cover" `
	Dic        string        `json:"dic"`
	Flag       int64         `json:"flag"`
	Uetime     model.StrTime `json:"uetime"`
	Ustime     model.StrTime `json:"ustime"`
	Level      int           `json:"level"`
	H5Cover    string        `json:"h5_cover"`
	Rank       int64         `json:"rank"`
	LikeLimit  int           `json:"like_limit"`
	ChildSids  string        `json:"child_sids"`
	ShieldFlag int64         `json:"shield_flag"`
}

// SubjectChild subjectchild
type SubjectChild struct {
	ID           int64   `json:"id"`
	ChildIds     string  `json:"child_ids"`
	ChildIdsList []int64 `json:"_"`
}

// SubjectTotalStat .
type SubjectTotalStat struct {
	SumCoin int64 `json:"sum_coin"`
	SumFav  int64 `json:"sum_fav"`
	SumLike int64 `json:"sum_like"`
	SumView int64 `json:"sum_view"`
	Count   int   `json:"count"`
}

// VipActOrder .
type VipActOrder struct {
	ID             int64         `json:"id"`
	Mid            int64         `json:"mid"`
	OrderNo        string        `json:"order_no"`
	ProductID      string        `json:"product_id"`
	Ctime          model.StrTime `json:"ctime"`
	Mtime          model.StrTime `json:"mtime"`
	PanelType      string        `json:"panel_type"`
	Months         int           `json:"months"`
	AssociateState int           `json:"associate_state"`
}

// EsParams .
type EsParams struct {
	Sid   int64  `json:"sid"`
	State int    `json:"state"`
	Ps    int    `json:"ps"`
	Pn    int    `json:"pn"`
	Order string `json:"order"`
	Sort  string `json:"sort"`
}

// EsItem .
type EsItem struct {
	ID  int64 `json:"id"`
	Wid int64 `json:"wid"`
}

type AidsData struct {
	Aids string `json:"aids"`
}

type AwardSubject struct {
	ID           int64
	Name         string
	Etime        xtime.Time
	Sid          int64
	Type         string
	SourceID     string
	SourceExpire int64
	OtherSids    string
	TaskID       int64
}

// SubjectRule ...
type SubjectRule struct {
	ID            int64           `json:"id" form:"id"`
	Sid           int64           `json:"sid" form:"sid"`
	State         int64           `json:"state" form:"state"`
	TaskID        int64           `json:"task_id" form:"task_id"`
	Category      int64           `json:"category" form:"category"`
	TypeIds       string          `json:"type_ids" form:"type_ids"`
	Tags          string          `json:"tags" form:"tags"`
	Attribute     int64           `json:"attribute" form:"attribute"`
	RuleName      string          `json:"rule_name" form:"rule_name"`
	Sids          string          `json:"sids" form:"sids"`
	AidSource     string          `json:"aid_source" form:"aid_source"`
	AidSourceMap  []*AidSourceMap `json:"_" form:"_"`
	AidSourceType int             `json:"aid_source_type" form:"aid_source_type"`
	Coefficient   string          `json:"coefficient" form:"coefficient"`
}

// AidSourceMap ...
type AidSourceMap map[string]interface{}

// Fav ...
type Fav struct {
	FID int64 `json:"fid" form:"fid"`
	MID int64 `json:"mid" form:"mid"`
}

// Sids ...
type Sids struct {
	SIDs []int64 `json:"sids" form:"sids"`
}

// IsStartAtReserve 是否是报名之后开始统计
func (r *SubjectRule) IsStartAtReserve() bool {
	return r.Attribute&2 > 0
}

type SubRuleStat struct {
	Aid         int64 `json:"aid"`
	Mid         int64 `json:"mid"`
	StateChange int64 `json:"state_change"`
	Raw         struct {
		Rule int64 `json:"rule"`
	} `json:"raw"`
}

func (sub *ActSubject) IsForbidListSearch() bool {
	return ((sub.Flag >> FLATLISTFORBIDOTHER) & int64(1)) == 1
}

func (sub *ActSubject) IsForbidListOversea() bool {
	return ((sub.Flag >> FLATLISTFORBIDOVERSEA) & int64(1)) == 1
}

func (sub *ActSubject) IsForbidListRcmd() bool {
	return ((sub.Flag >> FLATLISTFORBIDRCMD) & int64(1)) == 1
}

func (sub *ActSubject) IsForbidListOther() bool {
	return ((sub.Flag >> FLATLISTFORBIDOTHER) & int64(1)) == 1
}

// IsShieldRank 是否禁止排行
func (sub *ActSubject) IsShieldRank() bool {
	return ((sub.ShieldFlag >> ShieldNoRank) & int64(1)) == 1
}

// IsShieldDynamic 是否禁止动态
func (sub *ActSubject) IsShieldDynamic() bool {
	return ((sub.ShieldFlag >> ShieldNoDynamic) & int64(1)) == 1
}

// IsShieldRecommend 是否禁止推荐
func (sub *ActSubject) IsShieldRecommend() bool {
	return ((sub.ShieldFlag >> ShieldNoRecommend) & int64(1)) == 1
}

// IsShieldHot 是否禁止热门
func (sub *ActSubject) IsShieldHot() bool {
	return ((sub.ShieldFlag >> ShieldNoHot) & int64(1)) == 1
}

// IsShieldFansDynamic 是否禁止粉丝动态
func (sub *ActSubject) IsShieldFansDynamic() bool {
	return ((sub.ShieldFlag >> ShieldNoFansDynamic) & int64(1)) == 1
}

// IsShieldSearch 是否禁止搜索
func (sub *ActSubject) IsShieldSearch() bool {
	return ((sub.ShieldFlag >> ShieldNoSearch) & int64(1)) == 1
}

// IsShieldOversea 是否禁止海外
func (sub *ActSubject) IsShieldOversea() bool {
	return ((sub.ShieldFlag >> ShieldNoOversea) & int64(1)) == 1
}

type SubjectStat struct {
	Sid int64 `json:"sid" form:"sid"`
	Num int64 `json:"num" form:"num"`
}

type ActivityReservePub struct {
	Sid         int64 `json:"sid"`
	Mid         int64 `json:"mid"`
	State       int64 `json:"state"`
	TimeVersion int64 `json:"time_version"`
}

type ActivityActReserve struct {
	Sid   int64 `json:"sid"`
	Mid   int64 `json:"mid"`
	State int64 `json:"state"`
}

type ActReserveField struct {
	ID          int64  `json:"id,omitempty"`
	Sid         int64  `json:"sid,omitempty"`
	Mid         int64  `json:"mid,omitempty"`
	Num         int64  `json:"num,omitempty"`
	State       int64  `json:"state,omitempty"`
	IPV6        []byte `json:"ipv6,omitempty"`
	Ctime       string `json:"ctime,omitempty"`
	MTime       string `json:"etime,omitempty"`
	Score       int64  `json:"score,omitempty"`
	AdjustScore int64  `json:"adjust_score,omitempty"`
	From        string `json:"from,omitempty"`
	Typ         string `json:"typ,omitempty"`
	Oid         string `json:"oid,omitempty"`
	Platform    string `json:"platform,omitempty"`
	Mobiapp     string `json:"mobiapp,omitempty"`
	Buvid       string `json:"buvid,omitempty"`
	Spmid       string `json:"spmid,omitempty"`
}

type UpActReserveRelation struct {
	ID                int64  `json:"id"`
	Sid               int64  `json:"sid"`
	Mid               int64  `json:"mid"`
	Oid               string `json:"oid"`
	Type              int64  `json:"type"`
	State             int64  `json:"state"`
	Audit             int64  `json:"audit"`
	AuditChannel      int64  `json:"audit_channel"`
	From              int64  `json:"from"`
	DynamicID         string `json:"dynamic_id"`
	DynamicAudit      int64  `json:"dynamic_audit"`
	LotteryType       int64  `json:"lottery_type"`
	LotteryID         string `json:"lottery_id"`
	LotteryAudit      int64  `json:"lottery_audit"`
	LivePlanStartTime string `json:"live_plan_start_time"`
}

type UpAct41 struct {
	Mid   int64 `json:"mid"`
	Sid   int64 `json:"sid"`
	Time  int64 `json:"time"`
	State int64 `json:"state"`
}

type UpActReserve41 struct {
	Sid   int64 `json:"sid"`
	Total int64 `json:"total"`
	Time  int64 `json:"time"`
}

const (
	UpActReserveReject         = -3
	UpActReserveAudit          = -2
	UpActReservePassDelayAudit = -1
	UpActReservePass           = 0
)

const UpActReserveJobErrLogUnifyPrefix = "[UpActReserveJobErrLogUnifyPrefix]"
const UpActReserveLogPrefix = UpActReserveJobErrLogUnifyPrefix + "[CreateUpActNotify2Platform]"
const UpActReserveArcCronLogPrefix = UpActReserveJobErrLogUnifyPrefix + "[ArcCronStateChangeRelationLog]"
const UpActReserveRelationLotteryUserReserveState = UpActReserveJobErrLogUnifyPrefix + "[LotteryUserReserveState]"
const UpActReserveRelationChannelAuditNotify = UpActReserveJobErrLogUnifyPrefix + "[RelationChannelAuditNotify]"
const UpActReserveRelationLotteryNotifyCard = UpActReserveJobErrLogUnifyPrefix + "[LotteryUserReserveNotifyCard]"
const UpActReserveRelationLotteryNotify = UpActReserveJobErrLogUnifyPrefix + "[LotteryUserReserveNotify]"
const UpActReserveRelationPushVerifyCard = UpActReserveJobErrLogUnifyPrefix + "[UpActReserveRelationPushVerifyCard]"

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
	UpActReserveAuditChannelDefault  = 0
	UpActReserveAuditChannelPlatform = 1 // 等待审核平台
	UpActReserveAuditChannelArchive  = 2 // 等待稿件过审
)

const (
	ActSubjectStateNormal = 1
	ActSubjectStateAudit  = 0
	ActSubjectStateCancel = -1
	ActSubjectStateEdit   = -2
	ActSubjectStateReject = -3
)

type CreateDynamicCard struct {
	UID         int64                  `json:"uid"`
	Biz         int64                  `json:"biz"`
	Category    int64                  `json:"category"`
	Type        int64                  `json:"type"`
	Title       string                 `json:"title"`
	Tags        string                 `json:"tags"`
	Pictures    []CreateDynamicCardImg `json:"pictures"`
	Description string                 `json:"description"`
	Setting     int64                  `json:"setting"`
	AtUids      string                 `json:"at_uids"`
	AtControl   string                 `json:"at_control"`
	From        string                 `json:"from"`
	Extension   string                 `json:"extension"`
	AuditLevel  int64                  `json:"audit_level"`
}

type CreateDynamicCardImg struct {
	ImgSrc    string `json:"img_src"`
	ImgWidth  string `json:"img_width"`
	ImgHeight string `json:"img_height"`
	ImgSize   string `json:"img_size"`
}

type CreateDynamicCardExtension struct {
	FlagCfg struct {
		Reserve struct {
			ReserveID int64 `json:"reserve_id"`
		} `json:"reserve"`
	} `json:"flag_cfg"`
}

type UpActReserveRelationBind struct {
	ID    int64  `json:"id"`
	Sid   int64  `json:"sid"`
	Oid   string `json:"oid"`
	OType int64  `json:"o_type"`
	Rid   string `json:"rid"`
	RType int64  `json:"r_type"`
}

type BFSFileInfo struct {
	Format   string `json:"format"`
	Height   int64  `json:"height"`
	Width    int64  `json:"width"`
	FileSize int64  `json:"file_size"`
}

const CreateDynamicBiz = 3
const CreateDynamicCategory = 3
const CreateDynamicType = 0
const CreateDynamicFrom = "draft_video.reserve.svr"
const CreateDynamicAuditLevel = 0

type CreateDynamicReply struct {
	Code int64                  `json:"code"`
	Msg  string                 `json:"msg"`
	Data CreateDynamicDataReply `json:"data"`
}

type CreateDynamicDataReply struct {
	DynamicID    int64  `json:"dynamic_id"`
	ErrMsg       string `json:"errmsg"`
	DynamicIDStr string `json:"dynamic_id_str"`
}

const (
	SensitiveLevelNormal      = 0
	SensitiveLevelTest        = 10
	SensitiveLevelPass        = 15
	SensitiveLevelAudit       = 16
	SensitiveLevelIntercept20 = 20
	SensitiveLevelIntercept30 = 30
	SensitiveLevelIntercept40 = 40
)

const DeleteDynamicFrom = "delete.reserve.svr"

type UpActReserveRelationChannelAudit struct {
	ReserveID         string `json:"reserve_id"`
	ReserveAudit      int64  `json:"reserve_audit"`
	ReserveAuditFirst bool   `json:"reserve_audit_first"`
	DynamicID         string `json:"dynamic_id"`
	DynamicAudit      int64  `json:"dynamic_audit"`
	LotteryID         string `json:"lottery_id"`
	LotteryAudit      int64  `json:"lottery_audit"`
}

const (
	UpActReserveChannelReject = -1
	UpActReserveChannelAudit  = 0
	UpActReserveChannelPass   = 1
)

const UpActReserveRelationRetry = 3

const (
	DynamicLotteryLiveBizID = 10 // 直播预约
	DynamicLotteryArcBizID  = 11 // 稿件预约
)

const ErrCodeBGroupExist = 145202
const ErrCodeTunnelCardState = 108014
