package app

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	xtime "go-common/library/time"
)

var (
	DanmuNotShow, _  = regexp.Compile("(不展示|没有弹幕|弹幕没有|消失|不见|无弹幕|没弹幕|无法显示|无法观看弹幕|不动|花屏|黑|不显示|看不到)")
	DanmuCaton, _    = regexp.Compile("(卡|卡顿|抖动)")
	DanmuNum, _      = regexp.Compile("(少|数量)")
	DanmuSetting, _  = regexp.Compile("(开关|屏蔽|设置|不保存|失效|智能防挡|智能)")
	DanmuSubtitle, _ = regexp.Compile("(字幕)")
	DanmuNotifRobot  = "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=bcb89514-d70a-47bb-88e2-3940290e3b23"
)

const (
	Default = iota
	InProgress
	Processed
)

type FeedbackDB struct {
	ID             int64      `gorm:"column:id;primaryKey" json:"id" form:"id"`
	AppKey         string     `gorm:"column:app_key" json:"app_key" form:"app_key"`
	VersionCode    string     `gorm:"column:version_code" json:"version_code" form:"version_code"`
	Mid            int64      `gorm:"column:mid" json:"mid" form:"mid"`
	Buvid          string     `gorm:"column:buvid" json:"buvid" form:"buvid"`
	BV             string     `gorm:"column:bv" json:"bv" form:"bv"`
	Model          string     `gorm:"column:model" json:"model" form:"model"`
	Brand          string     `gorm:"column:brand" json:"brand" form:"brand"`
	Osver          string     `gorm:"column:osver" json:"osver" form:"osver"`
	Province       string     `gorm:"column:province" json:"province" form:"province"`
	Isp            string     `gorm:"column:isp" json:"isp" form:"isp"`
	CrashTime      xtime.Time `gorm:"column:crash_time" json:"crash_time" form:"crash_time"`
	Description    string     `gorm:"column:description" json:"description" form:"description"`
	Remark         string     `gorm:"column:remark" json:"remark" form:"remark"`
	Business       string     `gorm:"column:business" json:"business" form:"business"`
	Status         int64      `gorm:"column:status" json:"status" form:"status" default:"-99"`
	Principal      string     `gorm:"column:principal" json:"principal" form:"principal"`
	CrashReason    string     `gorm:"column:crash_reason" json:"crash_reason" form:"crash_reason"`
	OverviewImgUrl string     `gorm:"column:overview_img_url" json:"overview_img_url" form:"overview_img_url"`
	Contact        string     `gorm:"column:contact" json:"contact" form:"contact"`
	Operator       string     `gorm:"column:operator" json:"operator" form:"operator"`
	Editor         string     `gorm:"column:editor" json:"editor" form:"editor"`
	Mtime          xtime.Time `gorm:"column:mtime" json:"mtime" form:"mtime"`
	Ctime          xtime.Time `gorm:"column:ctime" json:"ctime" form:"ctime"`
	MediaUrls      string     `gorm:"column:media_urls" json:"media_urls" form:"media_urls"`
	WxRobots       string     `gorm:"column:wx_robots" json:"wx_robots" form:"wx_robots"`
	WxRobotIds     string     `gorm:"column:wx_robot_ids" json:"wx_robot_ids" form:"wx_robot_ids"`
	IsBug          bool       `gorm:"column:is_bug" json:"is_bug" form:"is_bug"`
	TapdUrl        string     `gorm:"column:tapd_url" json:"tapd_url" form:"tapd_url"`
	SendTo         string     `gorm:"column:send_to" json:"send_to" form:"send_to"`
}

// FeedbackReq request struct
type FeedbackReq struct {
	ID              int64      `gorm:"column:id;primaryKey" json:"id" form:"id"`
	AppKey          string     `gorm:"column:app_key" json:"app_key" form:"app_key"`
	VersionCode     string     `gorm:"column:version_code" json:"version_code" form:"version_code"`
	Mid             int64      `gorm:"column:mid" json:"mid" form:"mid"`
	Buvid           string     `gorm:"column:buvid" json:"buvid" form:"buvid"`
	BV              string     `gorm:"column:bv" json:"bv" form:"bv"` //视频号
	Model           string     `gorm:"column:model" json:"model" form:"model"`
	Brand           string     `gorm:"column:brand" json:"brand" form:"brand"`
	Osver           string     `gorm:"column:osver" json:"osver" form:"osver"`
	Province        string     `gorm:"column:province" json:"province" form:"province"`
	Isp             string     `gorm:"column:isp" json:"isp" form:"isp"`
	CrashTime       xtime.Time `gorm:"column:crash_time" json:"crash_time" form:"crash_time"`
	Description     string     `gorm:"column:description" json:"description" form:"description"`
	Remark          string     `gorm:"column:remark" json:"remark" form:"remark"`
	Business        string     `gorm:"column:business" json:"business" form:"business"`
	Status          int64      `gorm:"column:status" json:"status" form:"status" default:"-99"`
	Principal       string     `gorm:"column:principal" json:"principal" form:"principal"`
	CrashReason     string     `gorm:"column:crash_reason" json:"crash_reason" form:"crash_reason"`
	OverviewImgUrl  string     `gorm:"column:overview_img_url" json:"overview_img_url" form:"overview_img_url"`
	Contact         string     `gorm:"column:contact" json:"contact" form:"contact"`
	RobotKey        string     `json:"robot_key" form:"robot_key"`
	Operator        string     `gorm:"column:operator" json:"operator" form:"operator"`
	Editors         []string   `json:"editors,omitempty" form:"editors"`
	SendTo          string     `json:"send_to,omitempty" form:"send_to"`
	TapdUrl         string     `gorm:"column:tapd_url" json:"tapd_url" form:"tapd_url"`
	IsBug           bool       `json:"is_bug" form:"is_bug"`
	Mtime           xtime.Time `gorm:"column:mtime" json:"mtime" form:"mtime"`
	Ctime           xtime.Time `gorm:"column:ctime" json:"ctime" form:"ctime"`
	MediaUrls       []string   `json:"media_urls" form:"media_urls"`
	WxRobots        []string   `json:"wx_robots" form:"wx_robots"`
	WxRobotIds      []int64    `json:"wx_robot_ids" form:"wx_robot_ids"`
	CrashStartTime  xtime.Time `json:"crash_start_time,omitempty" form:"crash_start_time" time_format:"2006-01-02T15:04:05"`
	CrashEndTime    xtime.Time `json:"crash_end_time,omitempty" form:"crash_end_time" time_format:"2006-01-02T15:04:05"`
	CreateStartTime xtime.Time `json:"create_start_time,omitempty" form:"create_start_time" time_format:"2006-01-02T15:04:05"`
	CreateEndTime   xtime.Time `json:"create_end_time,omitempty" form:"create_end_time" time_format:"2006-01-02T15:04:05"`
	Pn              int        `json:"pn,omitempty" form:"pn" default:"1"`
	Ps              int        `json:"ps,omitempty" form:"ps" default:"20"`
}

// FeedbackRes response struct
type FeedbackRes struct {
	ID             int64       `gorm:"column:id;primaryKey" json:"id" form:"id"`
	AppKey         string      `gorm:"column:app_key" json:"app_key" form:"app_key"`
	VersionCode    string      `gorm:"column:version_code" json:"version_code" form:"version_code"`
	Mid            int64       `gorm:"column:mid" json:"mid" form:"mid"`
	Buvid          string      `gorm:"column:buvid" json:"buvid" form:"buvid"`
	BV             string      `gorm:"column:bv" json:"bv" form:"bv"`
	Model          string      `gorm:"column:model" json:"model" form:"model"`
	Brand          string      `gorm:"column:brand" json:"brand" form:"brand"`
	Osver          string      `gorm:"column:osver" json:"osver" form:"osver"`
	Province       string      `gorm:"column:province" json:"province" form:"province"`
	Isp            string      `gorm:"column:isp" json:"isp" form:"isp"`
	CrashTime      xtime.Time  `gorm:"column:crash_time" json:"crash_time" form:"crash_time"`
	Description    string      `gorm:"column:description" json:"description" form:"description"`
	Remark         string      `gorm:"column:remark" json:"remark" form:"remark"`
	Business       string      `gorm:"column:business" json:"business" form:"business"`
	Status         int64       `gorm:"column:status" json:"status" form:"status" default:"-99"`
	Principal      *UserInfo   `gorm:"column:principal" json:"principal" form:"principal"`
	CrashReason    string      `gorm:"column:crash_reason" json:"crash_reason" form:"crash_reason"`
	OverviewImgUrl string      `gorm:"column:overview_img_url" json:"overview_img_url" form:"overview_img_url"`
	Contact        string      `gorm:"column:contact" json:"contact" form:"contact"`
	Operator       *UserInfo   `json:"operator" form:"operator"`
	Editors        []*UserInfo `json:"editors,omitempty" form:"editors"`
	Mtime          xtime.Time  `gorm:"column:mtime" json:"mtime" form:"mtime"`
	Ctime          xtime.Time  `gorm:"column:ctime" json:"ctime" form:"ctime"`
	MediaUrls      []string    `json:"media_urls" form:"media_urls"`
	WxRobotIds     []int64     `json:"wx_robot_ids" form:"wx_robot_ids"`
	SendTo         string      `json:"send_to" form:"send_to"`
	TapdUrl        string      `gorm:"column:tapd_url" json:"tapd_url" form:"tapd_url"`
	IsBug          int         `json:"is_bug" form:"is_bug"`
}

type FeedbackTapdBug struct {
	WorkspaceID     string `json:"workspace_id" form:"workspace_id"`
	Title           string `json:"title" form:"title"`
	VerisonReport   string `json:"version_report" form:"version_report"`
	Module          string `json:"module" form:"module"`
	CurrentOwner    string `json:"current_owner" form:"current_owner"`
	Reporter        string `json:"reporter" form:"reporter"`
	Description     string `json:"description" form:"description"`
	CustomFieldOne  string `json:"custom_field_one" form:"custom_field_one"`
	CustomFieldFive string `json:"custom_field_five" form:"custom_field_five"`
	CustomFieldSix  string `json:"custom_field_6" form:"custom_field_6"`
	OriginPhase     string `json:"originphase" form:"originphase"`
	ID              string `json:"id" form:"id"`                             // 否	integer	ID	支持多ID查询
	Priority        string `json:"priority" form:"priority"`                 //否	string	优先级	支持枚举查询
	Severity        string `json:"severity" form:"severity"`                 //否	string	严重程度	支持枚举查询
	Status          string `json:"status" form:"status"`                     //否	string	状态	支持不等于查询、枚举查询
	IterationID     int64  `json:"iteration_id" form:"iteration_id"`         //	否	integer	迭代	支持枚举查询
	VersionTest     string `json:"version_test" form:"version_test"`         //	否	string	验证版本
	VersionFix      string `json:"version_fix" form:"version_fix"`           //	否	string	合入版本
	VersionClose    string `json:"version_close" form:"version_close"`       // 	否	string	关闭版本
	BaselineFind    string `json:"baseline_find" form:"baseline_find"`       //	否	string	发现基线
	BaselineJoin    string `json:"baseline_join" form:"baseline_join"`       //	否	string	合入基线
	BaselineTest    string `json:"baseline_test" form:"baseline_test"`       //	否	string	验证基线
	BaselineClose   string `json:"baseline_close" form:"baseline_close"`     //	否	string	关闭基线
	Cc              string `json:"cc" form:"cc"`                             //	否	string	抄送人
	Participator    string `json:"participator" form:"participator"`         //	否	string	参与人	支持多人员查询
	Te              string `json:"te" form:"te"`                             //	否	string	测试人员	支持模糊匹配
	De              string `json:"de" form:"de"`                             //	否	string	开发人员	支持模糊匹配
	Auditer         string `json:"auditer" form:"auditer"`                   //	否	string	审核人
	Confirmer       string `json:"confirmer" form:"confirmer"`               //	否	string	验证人
	Fixer           string `json:"fixer" form:"fixer"`                       //	否	string	修复人
	Closer          string `json:"closer" form:"closer"`                     //	否	string	关闭人
	Lastmodify      string `json:"lastmodify" form:"lastmodify"`             //	否	string	最后修改人
	Created         string `json:"created" form:"created"`                   //	否	datetime	创建时间	支持时间查询
	InProgressTime  string `json:"in_progress_time" form:"in_progress_time"` //	否	datetime	接受处理时间	支持时间查询
	Resolved        string `json:"resolved" form:"resolved"`                 //	否	datetime	解决时间	支持时间查询
	VerifyTime      string `json:"verify_time" form:"verify_time"`           //	否	datetime	验证时间	支持时间查询
	Closed          string `json:"closed" form:"closed"`                     //	否	datetime	关闭时间	支持时间查询
	RejectTime      string `json:"reject_time" form:"reject_time"`           //	否	datetime	拒绝时间	支持时间查询
	Modified        string `json:"modified" form:"modified"`                 //	否	datetime	最后修改时间	支持时间查询
	Begin           string `json:"begin" form:"begin"`                       //	否	date	预计开始
	Due             string `json:"due" form:"due"`                           //	否	date	预计结束
	Deadline        string `json:"deadline" form:"deadline"`                 //	否	date	解决期限
	Os              string `json:"os" form:"os"`                             //	否	string	操作系统
	Platform        string `json:"platform" form:"platform"`                 //	否	string	软件平台
	Testmode        string `json:"testmode" form:"testmode"`                 //	否	string	测试方式
	Testphase       string `json:"testphase" form:"testphase"`               //	否	string	测试阶段
	Testtype        string `json:"testtype" form:"testtype"`                 //	否	string	测试类型
	Source          string `json:"source" form:"source"`                     //	否	string	缺陷根源
	Bugtype         string `json:"bugtype" form:"bugtype"`                   //	否	string	缺陷类型
	Frequency       string `json:"frequency" form:"frequency"`               //	否	string	重现规律	支持枚举查询
	Sourcephase     string `json:"sourcephase" form:"sourcephase"`           //	否	string	引入阶段
	Resolution      string `json:"resolution" form:"resolution"`             //	否	string	解决方法	支持枚举查询
	Limit           string `json:"limit" form:"limit"`                       //	否	integer	设置返回数量限制，默认为30
	Page            string `json:"page" form:"page"`                         //	否	integer	返回当前数量限制下第N页的数据，默认为1（第一页）
	Order           string `json:"order" form:"order"`                       //	否	string	排序规则，规则：字段名 ASC或者DESC，然后 urlencode	如按创建时间逆序：order=created%20desc
	Fields          string `json:"fields" form:"fields"`                     //	否	string	设置获取的字段，多个字段间以','逗号隔开
}

// UserInfo struct
type UserInfo struct {
	UserName string `json:"user_name,omitempty" gorm:"user_name"`
	NickName string `json:"nick_name,omitempty" gorm:"nick_name"`
}

func (r *FeedbackDB) Convert2Resp(nameMap map[string]*UserInfo) (resp *FeedbackRes) {
	var (
		mediaArr   []string
		robotIdArr []int64
	)
	if len(r.MediaUrls) != 0 {
		mediaArr = strings.Split(r.MediaUrls, ",")
	}
	if len(r.WxRobotIds) != 0 {
		ids := strings.Split(r.WxRobotIds, ",")
		for _, id := range ids {
			idNum, _ := strconv.ParseInt(id, 10, 64)
			robotIdArr = append(robotIdArr, idNum)
		}
	}
	var principal = UserInfo{UserName: r.Principal}
	var operator = UserInfo{UserName: r.Operator}
	if v, ok := nameMap[r.Principal]; ok {
		principal.NickName = v.NickName
	}
	if v, ok := nameMap[r.Operator]; ok {
		operator.NickName = v.NickName
	}

	var editors []*UserInfo
	for _, v := range strings.Split(r.Editor, ",") {
		if v == "" {
			continue
		}
		var defaultEditor = UserInfo{UserName: v}
		if val, ok := nameMap[v]; ok {
			defaultEditor.NickName = val.NickName
		}
		editors = append(editors, &defaultEditor)
	}
	resp = &FeedbackRes{
		ID:             r.ID,
		AppKey:         r.AppKey,
		VersionCode:    r.VersionCode,
		Mid:            r.Mid,
		Buvid:          r.Buvid,
		BV:             r.BV,
		Model:          r.Model,
		Brand:          r.Brand,
		Osver:          r.Osver,
		Province:       r.Province,
		Isp:            r.Isp,
		CrashTime:      r.CrashTime,
		Description:    r.Description,
		Remark:         r.Remark,
		Business:       r.Business,
		Status:         r.Status,
		Principal:      &principal,
		CrashReason:    r.CrashReason,
		OverviewImgUrl: r.OverviewImgUrl,
		Contact:        r.Contact,
		Operator:       &operator,
		Editors:        editors,
		Mtime:          r.Mtime,
		Ctime:          r.Ctime,
		MediaUrls:      mediaArr,
		WxRobotIds:     robotIdArr,
		SendTo:         r.SendTo,
		TapdUrl:        r.TapdUrl,
	}
	if r.IsBug {
		resp.IsBug = 1
	}
	return
}

func (req *FeedbackReq) Convert2DB(op interface{}) (r *FeedbackDB) {
	userName := req.Operator
	if op != nil {
		userName = fmt.Sprintf("%v", op)
	}
	var idsStr []string
	for _, id := range req.WxRobotIds {
		idsStr = append(idsStr, strconv.FormatInt(id, 10))
	}
	return &FeedbackDB{
		ID:             req.ID,
		AppKey:         req.AppKey,
		VersionCode:    req.VersionCode,
		Mid:            req.Mid,
		Buvid:          req.Buvid,
		BV:             req.BV,
		Model:          req.Model,
		Brand:          req.Brand,
		Osver:          req.Osver,
		Province:       req.Province,
		Isp:            req.Isp,
		CrashTime:      req.CrashTime,
		Description:    req.Description,
		Remark:         req.Remark,
		Business:       req.Business,
		Status:         req.Status,
		Principal:      req.Principal,
		CrashReason:    req.CrashReason,
		OverviewImgUrl: req.OverviewImgUrl,
		Contact:        req.Contact,
		Operator:       userName,
		Editor:         strings.Join(req.Editors, ","),
		MediaUrls:      strings.Join(req.MediaUrls, ","),
		WxRobots:       strings.Join(req.WxRobots, ","),
		WxRobotIds:     strings.Join(idsStr, ","),
		IsBug:          req.IsBug,
		TapdUrl:        req.TapdUrl,
		SendTo:         req.SendTo,
	}
}
