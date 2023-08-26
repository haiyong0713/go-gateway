package usermodel

import (
	"fmt"

	xtime "go-common/library/time"
	api "go-gateway/app/app-svr/app-interface/interface-legacy/api/teenagers"
)

type Model int

var (
	TeenagersModel = Model(0)
	LessonsModel   = Model(1)
)

type Status int

var (
	CloseStatus  = Status(0) // 0:模式关闭
	OpenStatus   = Status(1) // 1:模式开启
	NotSetStatus = Status(2) // 2:未向服务端同步过模式（此状态青少年模式独有，版本升级时，客户端同步本地状态到服务端）
)

const (
	// teenager_users.operation
	OperationDefault                 = 0  //老数据的默认值
	OperationOpenSelf                = 1  //开启-主动
	OperationOpenForce               = 2  //开启-强制
	OperationOpenRealname            = 3  //开启-实名认证（废弃）
	OperationOpenDevSync             = 4  //开启-同步客户端状态
	OperationOpenFyParentControl     = 5  //开启-亲子平台家长控制
	OperationOpenForcedSync          = 6  //开启-网关强制更新导致的同步客户端状态（废弃）
	OperationOpenTeenForcedSync      = 7  //开启-强拉导致的同步客户端状态
	OperationOpenParentForcedSync    = 8  //开启-家长控制导致的同步客户端状态
	OperationOpenParentControlReopen = 9  //开启-青少年模式已开启，家长绑定亲子关系
	OperationQuitSelf                = 11 //退出-主动
	OperationQuitGuardian            = 12 //退出-监护人授权
	OperationQuitManager             = 13 //退出-后台
	OperationQuitForce               = 14 //退出-强制
	OperationQuitDevSync             = 15 //退出-同步客户端状态
	OperationQuitFyMgrUnbind         = 16 //退出-亲子平台后台解绑
	OperationQuitFyParentUnbind      = 17 //退出-亲子平台家长解绑
	OperationQuitFyParentControl     = 18 //退出-亲子平台家长控制
	OperationQuitFyChildUnbind       = 19 //退出-亲子平台孩子申诉解绑
	OperationQuitForcedSync          = 20 //退出-网关强制更新导致的同步客户端状态（废弃）
	OperationQuitAppeal              = 21 //退出-身份验证密码申诉
	OperationQuitTeenForcedSync      = 22 //退出-强退导致的同步客户端状态
	OperationQuitParentForcedSync    = 23 //退出-家长控制导致的同步客户端状态
	// pwd digit
	PwdDigit = 4 //密码位数
	// update from
	PwdFromGuardian         = "guardian"           //监护人认证页
	PwdFromPwd              = "password"           //密码页
	PwdFromDevSync          = "device_sync"        //客户端同步
	PwdFromForcedSync       = "forced_sync"        //网关强制更新导致的同步客户端状态
	PwdFromAppeal           = "pwd_appeal"         //身份验证密码申诉
	PwdFromTeenForcedSync   = "teen_forced_sync"   //强拉/强退控制导致的同步客户端状态
	PwdFromParentForcedSync = "parent_forced_sync" //家长控制导致的同步客户端状态
	// teenager_users.pwd_type
	PwdTypeSelf   = 0 //主动设置
	PwdTypeRandom = 1 //随机生成
	// related_key
	RelatedKeyTypeUser   = "usr"
	RelatedKeyTypeDevice = "dev"
	// manual_force
	ManualForceQuit = 0
	ManualForceOpen = 1
	// mf_operator
	OperatorSystem = "system"
)

var (
	From2OpOfQuit = map[string]int{
		PwdFromGuardian:         OperationQuitGuardian,
		PwdFromPwd:              OperationQuitSelf,
		PwdFromDevSync:          OperationQuitDevSync,
		PwdFromForcedSync:       OperationQuitForcedSync,
		PwdFromAppeal:           OperationQuitAppeal,
		PwdFromTeenForcedSync:   OperationQuitTeenForcedSync,
		PwdFromParentForcedSync: OperationQuitParentForcedSync,
	}
	From2OpOfOpen = map[string]int{
		PwdFromPwd:              OperationOpenSelf,
		PwdFromDevSync:          OperationOpenDevSync,
		PwdFromForcedSync:       OperationOpenForcedSync,
		PwdFromTeenForcedSync:   OperationOpenTeenForcedSync,
		PwdFromParentForcedSync: OperationOpenParentForcedSync,
	}
	Op2ContentOfLog = map[int]string{
		OperationOpenSelf:                "主动开启",
		OperationOpenForce:               "强制开启",
		OperationOpenRealname:            "实名认证开启",
		OperationOpenDevSync:             "同步客户端状态开启",
		OperationOpenFyParentControl:     "亲子平台家长控制开启",
		OperationOpenForcedSync:          "网关强制更新导致的同步客户端状态开启",
		OperationOpenTeenForcedSync:      "强拉导致的同步客户端状态开启",
		OperationOpenParentForcedSync:    "家长控制导致的同步客户端状态开启",
		OperationOpenParentControlReopen: "青少年模式已开启，家长绑定亲子关系",
		OperationQuitSelf:                "主动退出",
		OperationQuitGuardian:            "监护人授权退出",
		OperationQuitManager:             "后台解除退出",
		OperationQuitForce:               "强制退出",
		OperationQuitDevSync:             "同步客户端状态退出",
		OperationQuitFyParentUnbind:      "亲子平台家长解绑退出",
		OperationQuitFyParentControl:     "亲子平台家长控制退出",
		OperationQuitFyChildUnbind:       "亲子平台孩子申诉解绑退出",
		OperationQuitForcedSync:          "网关强制更新导致的同步客户端状态退出",
		OperationQuitAppeal:              "身份验证退出",
		OperationQuitTeenForcedSync:      "强退导致的同步客户端状态退出",
		OperationQuitParentForcedSync:    "家长控制导致的同步客户端状态退出",
	}
	isParentBindReOpen = map[int]struct{}{
		OperationOpenSelf:           {},
		OperationOpenForce:          {},
		OperationOpenRealname:       {},
		OperationOpenDevSync:        {},
		OperationOpenForcedSync:     {},
		OperationOpenTeenForcedSync: {},
	}

	isForcedOperations = map[int]struct{}{
		OperationOpenForce:          {},
		OperationQuitForce:          {},
		OperationQuitManager:        {},
		OperationOpenTeenForcedSync: {},
		OperationQuitTeenForcedSync: {},
	}
	isParentControlOperations = map[int]struct{}{
		OperationQuitFyParentControl:     {},
		OperationOpenFyParentControl:     {},
		OperationOpenParentControlReopen: {},
		OperationQuitFyChildUnbind:       {},
		OperationQuitFyMgrUnbind:         {},
		OperationOpenParentForcedSync:    {},
		OperationQuitParentForcedSync:    {},
	}
)

type User struct {
	ID          int64      `json:"id"`
	Mid         int64      `json:"mid"`
	MobiApp     string     `json:"mobi_app"`
	DeviceToken string     `json:"device_token"`
	Password    string     `json:"password"`
	State       int        `json:"state"`
	Model       Model      `json:"model"`
	Operation   int        `json:"operation"`
	QuitTime    xtime.Time `json:"quit_time"`
	PwdType     int        `json:"pwd_type"`
	ManualForce int64      `json:"manual_force"`
	MfOperator  string     `json:"mf_operator"`
	MfTime      xtime.Time `json:"mf_time"`
	// 设备同步操作状态
	DevOperation int `json:"-"`
}

type ModelStatus struct {
	TeenagersStatus Status `json:"teenagers_status"`
	Wsxcde          string `json:"wsxcde,omitempty"`
}

type UserModel struct {
	Mid             int64   `json:"mid"`
	Mode            string  `json:"mode,omitempty"`
	Wsxcde          string  `json:"wsxcde,omitempty"`
	Status          Status  `json:"status"`
	Policy          *Policy `json:"policy,omitempty"`
	IsForced        bool    `json:"is_forced"`
	MustTeen        bool    `json:"must_teen"`
	MustRealname    bool    `json:"must_realname"`
	IsParentControl bool    `json:"is_parent_control"`
	DailyDialogAb   int64   `json:"daily_dialog_ab,omitempty"`
}

type Policy struct {
	Interval     int64 `json:"interval"`
	UseLocalTime bool  `json:"use_local_time"` // 是否使用客户端本地时间
}

type AntiAddiction struct {
	DeviceToken string `form:"device_token"`
	UseTime     int64  `form:"time"`
	MID         int64  `form:"_"`
	Day         int64  `form:"-"`
}

type SpecialModeLog struct {
	RelatedKey  string `json:"related_key"`
	OperatorUid int64  `json:"operator_uid"`
	Operator    string `json:"operator"`
	Content     string `json:"content"`
}

type ManualForceLog struct {
	Mid      int64  `json:"mid"`
	Operator string `json:"operator"`
	Content  string `json:"content"`
	Remark   string `json:"remark"`
}

type UpdateTeenagerReq struct {
	MobiApp         string `form:"mobi_app"`
	DeviceToken     string `form:"device_token"`
	Pwd             string `form:"pwd"`
	Wsxcde          string `form:"wsxcde"`
	TeenagersStatus int    `form:"teenagers_status" validate:"min=0,max=1"`
	Sync            bool   `form:"sync"`
	From            string `form:"from"`
	DeviceModel     string `form:"device_model"`
}

type ModifyPwdReq struct {
	Mid         int64
	MobiApp     string
	DeviceToken string
	DeviceModel string
	OldPwd      string
	NewPwd      string
}

type VerifyPwdReq struct {
	Mid         int64
	MobiApp     string
	DeviceToken string
	Pwd         string
	PwdFrom     api.PwdFrom
	IsDynamic   bool
	CloseDevice bool
}

type UpdateStatusReq struct {
	Mid         int64
	MobiApp     string
	DeviceToken string
	DeviceModel string
	Switch      bool
	Pwd         string
	PwdFrom     api.PwdFrom
}

type ModeStatusReq struct {
	Mid         int64
	MobiApp     string
	DeviceToken string
	DeviceModel string
	IP          string
}

type FacialRecognitionVerifyReq struct {
	Mid         int64
	MobiApp     string
	DeviceToken string
	From        api.FacialRecognitionVerifyFrom
}

func LogContent(operation, status int) string {
	if content, ok := Op2ContentOfLog[operation]; ok {
		return content
	}
	if operation == OperationDefault {
		info := "退出"
		if status == int(OpenStatus) {
			info = "开启"
		}
		return fmt.Sprintf("老版本%s", info)
	}
	return fmt.Sprintf("unknown operation=%d", operation)
}

func LogOperator(operation int) string {
	switch operation {
	case OperationOpenSelf, OperationQuitSelf, OperationDefault:
		return "user"
	case OperationQuitFyParentUnbind, OperationOpenFyParentControl, OperationOpenParentControlReopen, OperationQuitFyParentControl, OperationQuitFyChildUnbind:
		return "family"
	default:
		return "system"
	}
}

func IsForced(operation int) bool {
	_, ok := isForcedOperations[operation]
	return ok
}

func IsParentReopen(operation int) bool {
	_, ok := isParentBindReOpen[operation]
	return ok
}

func IsParentControl(operation int) bool {
	_, ok := isParentControlOperations[operation]
	return ok
}

func LogRelatedKey(typ string, id int64) string {
	return fmt.Sprintf("%s_%d", typ, id)
}
