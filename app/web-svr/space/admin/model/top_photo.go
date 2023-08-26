package model

import (
	"fmt"
	"time"

	realaGRPC "git.bilibili.co/bapis/bapis-go/account/service/relation"
)

const (
	MEMBER_UPLOAD_TOP_PHOTO_NOTDELETED  = 0                      // 未删除
	MEMBER_UPLOAD_TOP_PHOTO_DELETED     = 1                      // 已删除
	TOP_PHOTO_PASSED                    = 1                      // 审核通过
	TOP_PHOTO_NOTPASSED                 = 2                      // 审核未通过
	TOP_PHOTO_UNPASS                    = 0                      // 未审核
	ACTIONPASS                          = "UploadPhotoPassPhoto" // 审核通过
	ACTIONBACK                          = "UploadPhotoBackPhoto" // 审核驳回
	ACTIONREPASS                        = "UploadPhotoRePass"    // 驳回后再通过
	BUSINESS_TOP_PHOTO_ADMIN            = 241                    // 业务ID
	NOTIFY_SEND_MC                      = "33_1_1"               //系统通知消息码
	NOTIFY_DATA_TYPE                    = 4                      // 系统通知消息类型
	ACCOUNT_BLOCK_FOREVER               = 2                      //永久封禁
	ACCOUNT_BLOCK_TEMP                  = 1                      //限时封禁
	CREDIT_BLOCK_INFO_FOREVER           = 1                      // 记录永久封禁
	CREDIT_BLOCK_INFO_TEMP              = 0                      // 记录限时
	ACCOUNT_BLOCK_SOURCE                = 3                      // 3. 后台相关
	ACCOUNT_BLOCK_AREA                  = 9                      // 9 空间头图
	ACCOUNT_BLOCK_NOTIFY                = 1                      // 1 通知
	ACCOUNT_BLOCK_NOTNOTIFY             = 0                      // 0 不通知
	CREDIT_INFO_ORIGIN_TYPE             = 9                      // 9 空间头图
	CREDIT_PUNISH_TYPE                  = 2                      //  2:封禁
	MEMBER_UPLOAD_TOPPHOTO_FROM_IOS     = 1                      // IOS
	MEMBER_UPLOAD_TOPPHOTO_FROM_ANDROID = 2                      //ANDROID
	MEMBER_UPLOAD_TOPPHOTO_FROM_IPAD    = 3                      // IPAD
	MEMBER_UPLOAD_TOPPHOTO_FROM_WEB     = 4                      // WEB
	TOPPHOTO_PLATFORM_MOBILE            = 2                      // ios & android
	TOPPHOTO_PLATFORM_CLIENT            = 1                      //  ipad&web
	VIP_AUDIT_LOG_REASON_PASS           = 0                      // 审核记录 通过
)

var BACK_REASON_REFLECT = map[int]string{
	1:  "涉及政治",
	2:  "色情信息",
	3:  "低俗信息",
	4:  "不适宜信息",
	5:  "违禁信息",
	6:  "垃圾广告信息",
	7:  "违反运营规则信息",
	8:  "赌博诈骗信息",
	9:  "侵犯他人隐私信息",
	10: "非法网站信息",
	11: "传播不实信息",
	12: "怂恿教唆信息",
}

var ACCOUNT_BLOCK_REASON = map[int]string{
	1:  "刷屏",
	2:  "抢楼",
	4:  "发布赌博诈骗信息",
	5:  "发布违禁相关信息",
	6:  "发布垃圾广告信息",
	7:  "发布人身攻击言论",
	8:  "发布侵犯他人隐私信息",
	9:  "发布引战言论",
	10: "发布剧透信息",
	11: "恶意添加无关标签",
	12: "恶意删除他人标签",
	13: "发布色情信息",
	14: "发布低俗信息",
	15: "发布暴力血腥信息",
	16: "涉及恶意投稿行为",
	17: "发布非法网站信息",
	18: "发布传播不实信息",
	19: "发布怂恿教唆信息",
	20: "恶意刷屏",
	21: "账号违规",
	22: "恶意抄袭",
	23: "冒充自制原创",
	24: "发布青少年不良内容",
	25: "破坏网络安全",
	26: "发布虚假误导信息",
	27: "仿冒官方认证账号",
	28: "发布不适宜内容",
	29: "违反运营规则",
	30: "恶意创建话题",
	31: "发布违规抽奖",
	32: "恶意冒充他人",
}

type TopPhotoArc struct {
	Mid      int64  `json:"mid" gorm:"column:mid"`
	Aid      int64  `json:"aid" gorm:"column:aid"`
	ImageURL string `json:"image_url" gorm:"column:image_url"`
}

func (t TopPhotoArc) TableName() string {
	return fmt.Sprintf("topphoto_arc_%d", t.Mid)
}

// MemberUploadTopPhoto 头图审核ORM
type MemberUploadTopPhoto struct {
	ID         int64  `json:"id" gorm:"column:id"`
	MID        int64  `json:"mid" gorm:"column:mid"`
	ImgPath    string `json:"img_path" gorm:"column:img_path"`
	PlatFrom   int    `json:"platfrom" gorm:"column:platfrom"`
	Status     int    `json:"status" gorm:"column:status"`
	Deleted    int    `json:"deleted" gorm:"column:deleted"`
	UploadDate string `json:"upload_date" gorm:"column:upload_date"`
	ModifyTime string `json:"modify_time" gorm:"column:modify_time"`
}

func (m *MemberUploadTopPhoto) TableName() string {
	return "member_upload_topphoto"
}

type JsonMemberUploadTopPhoto struct {
	Json struct {
		ID         int64  `json:"id" gorm:"column:id"`
		MID        int64  `json:"mid" gorm:"column:mid"`
		ImgPath    string `json:"img_path" gorm:"column:img_path"`
		PlatFrom   int    `json:"platfrom" gorm:"column:platfrom"`
		Status     int    `json:"status" gorm:"column:status"`
		Deleted    int    `json:"deleted" gorm:"column:deleted"`
		UploadDate string `json:"upload_date" gorm:"column:upload_date"`
		ModifyTime string `json:"modify_time" gorm:"column:modify_time"`
	} `json:"json"`
}

func (j JsonMemberUploadTopPhoto) TransformExtraData() (topPhoto *MemberUploadTopPhoto, err error) {
	var (
		strTop = j.Json
	)

	topPhoto = &MemberUploadTopPhoto{
		ID:         strTop.ID,
		MID:        strTop.MID,
		ImgPath:    strTop.ImgPath,
		PlatFrom:   strTop.PlatFrom,
		Status:     strTop.Status,
		Deleted:    strTop.Deleted,
		UploadDate: strTop.UploadDate,
		ModifyTime: strTop.ModifyTime,
	}

	return
}

// MemberUploadTopPhotoShow 头图审核前端显示
type MemberUploadTopPhotoShow struct {
	MemberUploadTopPhoto
	BackTimes     int    `json:"back_times" form:"back_times"`
	Fans          int64  `json:"fans" form:"fans"`
	Nickname      string `json:"nickname" form:"nickname"`
	Certification int32  `json:"certification" form:"certification"` // -1 未认证, 0 个人认证,  1 企业认证
}

// MemberUploadTopPhotoSearchParams  搜索参数
type MemberUploadTopPhotoSearchParams struct {
	UploadTimeStart string  `json:"upload_time_start" form:"upload_time_start"`
	UploadTimeEnd   string  `json:"upload_time_end" form:"upload_time_end"`
	PlatFrom        int     `json:"platfrom" form:"platfrom"`
	MIDs            []int64 `json:"mids" form:"mids"`
}

type TopPhotoRes struct {
	Items []*MemberUploadTopPhotoShow `json:"items" form:"items"`
	Pager *Pager                      `json:"pager" form:"pager"`
}

type TopPhotoBackTimes struct {
	Mid       int64
	BackTimes int
}

func (t *TopPhotoBackTimes) TableName() string {
	return "member_upload_topphoto"
}

type FansRes struct {
	Code    int                            `json:"code"`
	Message string                         `json:"message"`
	Data    map[int64]*realaGRPC.StatReply `json:"data"`
}

type Pager struct {
	CurrentPage int `json:"current_page" form:"current_page"`
	PageSize    int `json:"page_size" form:"page_size"`
	TotalItems  int `json:"total_items" form:"total_items"`
}

// VipAuditLog ORM
type VipAuditLog struct {
	ID            int64  `json:"id" gorm:"column:id"`
	TID           int64  `json:"tid" gorm:"column:tid"`
	MID           int64  `json:"mid" gorm:"column:mid"`
	Ctime         string `json:"ctime" gorm:"column:ctime"`
	Operator      string `json:"operator" gorm:"column:operator"`
	Reason        string `json:"reason" gorm:"column:reason"`
	ReasonDefault string `json:"reason_default" gorm:"column:reason_default"`
}

func (v *VipAuditLog) TableName() string {
	return "vip_audit_log"
}

// VipAuditLogSearch 搜索参数
type VipAuditLogSearch struct {
	UploadTimeStart string  `json:"upload_time_start" form:"upload_time_start"`
	UploadTimeEnd   string  `json:"upload_time_end" form:"upload_time_end"`
	AuditTimeStart  string  `json:"audit_time_start" form:"audit_time_start"`
	AuditTimeEnd    string  `json:"audit_time_end" form:"audit_time_end"`
	Status          []int   `json:"status" form:"status"`
	Platfrom        []int   `json:"platfrom" form:"platfrom"`
	MIDs            []int64 `json:"mids" form:"mids"`
	Operator        string  `json:"operator" form:"operator"`
}

type VipAuditLogResRaw struct {
	MemberUploadTopPhoto
	Reason        string `json:"reason" form:"reason"`
	ReasonDefault string `json:"reason_default" form:"reason_default"`
	Ctime         string `json:"ctime" form:"ctime"`
	Operator      string `json:"operator" form:"operator"`
	Fans          int64  `json:"fans" form:"fans"`
}

type VipAuditLogRes struct {
	Items []*VipAuditLogResRaw `json:"items" form:"items"`
	Pager *Pager               `json:"pager" form:"pager"`
}

// 行为日志
type LogSearchRes struct {
	Action     string `json:"action"`
	Business   int    `json:"business"`
	CTime      string `json:"ctime"`
	Department string `json:"department"`
	ExtraData  string `json:"extra_data"`
	MilliCtime string `json:"milli_ctime"`
	Str0       string `json:"str_0,omitempty"`
	Str1       string `json:"str_1,omitempty"`
	Str2       string `json:"str_2,omitempty"`
	Str3       string `json:"str_3,omitempty"`
	Str4       string `json:"str_4,omitempty"`
	Str5       string `json:"str_5,omitempty"`
	Int0       int64  `json:"int_0,omitempty"`
	Int1       int64  `json:"int_1,omitempty"`
	Int2       int64  `json:"int_2,omitempty"`
	Int3       int64  `json:"int_3,omitempty"`
	Int4       int64  `json:"int_4,omitempty"`
	OID        int64  `json:"oid"`
	Type       int    `json:"type"`
	UID        int    `json:"uid"`
	UName      string `json:"uname"`
}

type ActionLogRes struct {
	Items []*LogSearchRes `json:"items"`
	Pager Pager           `json:"pager"`
}

type LogSearchResRaw struct {
	Code    int                  `json:"code"`
	Message string               `json:"message"`
	Data    *LogSearchResRawData `json:"data"`
}

type LogSearchResRawData struct {
	Order  string          `json:"order"`
	Sort   string          `json:"sort"`
	Result []*LogSearchRes `json:"result"`
	Debug  string          `json:"debug"`
	Page   *Page           `json:"page"`
}

// MemberTopPhoto 用户头图ORM
type MemberTopPhoto struct {
	ID          int64  `json:"id" gorm:"column:id"`                     //主键ID
	MID         int64  `json:"mid" gorm:"column:mid"`                   //用户ID
	SID         int64  `json:"sid" gorm:"column:sid"`                   //头图ID
	Expire      int64  `json:"expire" gorm:"column:expire"`             //过期时间
	IsActivated int    `json:"is_activated" gorm:"column:is_activated"` //是否正在使用
	PlatFrom    int    `json:"platfrom" gorm:"column:platfrom"`         //0.后台设置 1. ipad web上传 2. iso Android上传
	ModifyTime  string `json:"modify_time" gorm:"modify_time"`          //最后修改时间
}

func (m *MemberTopPhoto) TableName() string {
	return fmt.Sprintf("member_topphoto%d", m.MID%10)
}

// VipInfo 用户VIP信息
type VipInfo struct {
	MID            int64 `json:"mid"`
	VipType        int   `json:"vipType"`
	VipStatus      int   `json:"vipStatus"`
	VipDueDate     int   `json:"vipDueDate"`
	AccessStatus   int   `json:"accessStatus"`
	VipSurplusMsec int   `json:"vipSurplusMsec"`
	IsAutoRenew    int   `json:"isAutoRenew"`
	Label          struct {
		Path string `json:"path"`
	} `json:"label"`
}

type VipInfoSearchRes struct {
	Code int      `json:"code"`
	Msg  string   `json:"msg"`
	Data *VipInfo `json:"data"`
}

type AuditLogInitParams struct {
	UName    string        `json:"uname"`      // 审核人员内网name 多个用逗号分隔，如aa,bb
	UID      int64         `json:"uid"`        // 审核人员内网uid 多个用逗号分隔，如11,22
	Business int           `json:"business"`   // 业务id，比如稿件业务 多个用逗号分隔，如11,22
	Type     int           `json:"type"`       // 操作对象的类型，如评论 多个用逗号分隔，如11,22
	OID      int64         `json:"oid"`        // 操作对象的id，如2233 多个用逗号分隔，如11,22
	Action   string        `json:"action"`     // 操作对象的id，如aaaa 多个用逗号分隔，如aa,bb
	CTime    time.Time     `json:"ctime_from"` // 操作启始时间 如"2006-01-02 15:04:05" 查询范围限制见下
	Index    []interface{} `json:"index"`
	Content  interface{}   `json:"content"`
}

type BackPhotoParam struct {
	ID            int64  `json:"id"`
	Reason        int    `json:"reason"`
	ReasonDefault string `json:"reason_default"`
}

type AccountBlockParam struct {
	AccountBlock int    `json:"account_block"`
	ReasonType   int    `json:"reason_type"`
	BlockRemark  string `json:"block_remark"`
	BlockTime    int    `json:"block_time"`
	BlockNotify  int    `json:"block_notify"`
	Moral        int    `json:"moral"`
}

type BfsRes struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// 系统通知
type NotifySendInit struct {
	Mc       string `json:"mc"`
	Title    string `json:"title"`
	DataType int    `json:"data_type"`
	Context  string `json:"context"`
	MIDList  int64  `json:"mid_list"`
}

type NotifyRes struct {
	Mc           string  `json:"mc"`
	DataType     int     `json:"data_type"`
	TotalCount   int     `json:"total_count"`
	ErrorCount   int     `json:"error_count"`
	ErrorMidList []int64 `json:"error_mid_list"`
}

type NotifyRawRes struct {
	Code int        `json:"code"`
	Data *NotifyRes `json:"data"`
}

// 账户封锁
type AccountBlockInit struct {
	MID       int64  `json:"mid"`        // 封禁的用户id
	Source    int    `json:"source"`     // 封禁来源  3. 后台相关
	Area      int    `json:"area"`       // 违规业务  9 空间头图
	Action    int    `json:"action"`     // 封禁类型  1. 限时 2. 永久
	Duration  int    `json:"duration"`   // 封禁时长 (s秒)
	StartTime int64  `json:"start_time"` // 封禁开始时间，unix time，（10位， 精确到秒）
	OpID      int64  `json:"op_id"`      // 操作人id
	Operator  string `json:"operator"`   // 操作人
	Reason    string `json:"reason"`     // 封禁原因
	Comment   string `json:"comment"`    // 封禁理由备注
	Notify    int    `json:"notify"`     // 0 不通知 1 通知
}

type BlockInfoAdd struct {
	MID            int64  `json:"mid"`             //封禁的用户id
	BlockedDays    int    `json:"blocked_days"`    // 封禁时间 一般是 3 / 7/ 15 / 没有传(0) / 自定义
	BlockedForever int    `json:"blocked_forever"` // 是否永久封禁 0 否，1是
	BlockedRemark  string `json:"blocked_remark"`  // 封禁备注
	MoralNum       int    `json:"moral_num"`       // 	扣除节操值
	OriginType     int    `json:"origin_type"`     // 封禁来源 9 空间头图
	PunishTime     int64  `json:"punish_time"`     // 惩罚时间 timestamp
	PunishType     int    `json:"punish_type"`     // 惩罚类型 1:节操 ;2:封禁; 3:永久封禁
	ReasonType     int    `json:"reason_type"`     // 封禁理由
	OperID         int64  `json:"oper_id"`         // 操作人ID
	OperatorName   string `json:"operator_name"`   // 操作人
}

type DelMoralParam struct {
	MID        int64  `json:"mid"`
	Delta      int    `json:"delta"`
	Origin     int    `json:"origin"`
	Reason     string `json:"reason"`
	ReasonType int    `json:"reason_type"`
	Operator   string `json:"operator"`
	Remark     string `json:"remark"`
	IsNotify   int    `json:"is_notify"`
}

type ResRaw struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
