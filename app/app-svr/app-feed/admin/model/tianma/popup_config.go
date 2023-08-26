package tianma

import (
	"encoding/json"
	xecode "go-common/library/ecode"
	xtime "go-common/library/time"
)

const (
	PopupIsDeleted                  = 1        // 已删除
	PopupNotDeleted                 = 0        // 未删除
	PopupAuditStatePass             = 1        // 审核通过
	PopupAuditStateNotPass          = 0        // 审核未通过
	PopupAuditStateCanceled         = -1       // 审核被拒
	PopupAuditStateOffline          = 2        // 手动下线
	PopupTeenagePush                = 1        // 青少年模式弹出
	PopupTeenageNotPush             = -1       // 青少年模式不弹出
	PopupTypeBusiness               = 1        // 天马业务弹窗类型
	PopupAutoHideStatusHide         = 1        // 自动隐藏
	PopupAutoHideStatusNotHide      = 2        // 不自动隐藏
	PopupReTypeNone                 = -1       // 不跳转
	PopupReTypeURL                  = 1        // 跳转到URL
	PopupReTypeGame                 = 2        // 跳转到游戏小卡
	PopupReTypeVideo                = 3        // 跳转到稿件
	PopupReTypePGC                  = 4        // 跳转到PGC
	PopupReTypeLive                 = 5        // 跳转到直播
	PopupReTypeArticle              = 6        // 跳转到专栏
	PopupReTypeDaily                = 7        // 跳转到每日精选
	PopupReTypeSongList             = 8        // 跳转到歌单
	PopupReTypeSong                 = 9        // 跳转到歌曲
	PopupReTypeAlbum                = 10       // 跳转到相簿
	PopupReTypeClip                 = 11       // 跳转到小视频
	PopupCrowdTypeNone              = -1       // 不定向
	PopupCrowdTypeBGroup            = 1        // 人群包定向
	PopupCrowdTypeHive              = 2        // Hive表定向
	PopupCrowdBaseBGroupMID         = 1        // mid
	PopupCrowdBaseBGroupBuvid       = 2        // buvid
	PopupBuildsPlatformIos          = 1        // 版本限制-iOS平台
	PopupBuildsPlatformAndroid      = 2        // 版本限制-Android平台
	PopupActionLogBusinessID        = 210      // 行为日志Business ID
	PopupActionLogAddPopupConfig    = "添加弹窗配置" // 行为日志名，添加弹窗配置
	PopupActionLogUpdatePopupConfig = "更新弹窗配置" // 行为日志名，更新弹窗配置
	PopupActionLogDeletePopupConfig = "删除弹窗配置" // 行为日志名，删除弹窗配置
	PopupActionLogAuditPopupConfig  = "审核弹窗配置" // 行为日志名，审核弹窗配置
	PopupStatusOnline               = 1        // 生效中
	PopupStatusManualOffline        = 2        // 已下线
	PopupStatusReadyOnline          = 3        // 待生效
	PopupStatusAutoOffline          = 4        // 自动失效
)

// PopupConfig 天马业务弹窗配置
type PopupConfig struct {
	ID                int64      `json:"id" form:"id" gorm:"column:id"`
	ImageURL          string     `json:"img_url" form:"img_url" gorm:"column:img_url"`
	Description       string     `json:"description" form:"description" gorm:"column:description"`
	PopupType         int        `json:"popup_type" form:"popup_type" gorm:"column:popup_type"`
	TeenagePushFlag   int        `json:"teenage_push" form:"teenage_push" gorm:"column:teenage_push"`
	AutoHideStatus    int        `json:"auto_hide_status" form:"auto_hide_status" gorm:"column:auto_hide_status"`
	AutoHideCountdown int64      `json:"auto_hide_countdown" form:"auto_hide_countdown" gorm:"column:auto_hide_countdown"`
	ReType            int        `json:"redirect_type" form:"redirect_type" gorm:"column:redirect_type"`
	ReTarget          string     `json:"redirect_target" form:"redirect_target" gorm:"column:redirect_target"`
	Builds            string     `json:"builds" form:"builds" gorm:"column:builds"`
	Status            int        `json:"status" form:"-" gorm:"-"` // 配置状态。1-在线，2-已下线，3-待生效，4-已过期
	AuditState        int        `json:"audit_state" form:"audit_state" gorm:"column:audit_state"`
	CrowdType         int        `json:"crowd_type" form:"crowd_type" gorm:"column:crowd_type"`
	CrowdBase         int        `json:"crowd_base" form:"crowd_base" gorm:"column:crowd_base"`
	CrowdValue        string     `json:"crowd_value" form:"crowd_value" gorm:"column:crowd_value"`
	STime             xtime.Time `json:"stime" form:"stime" gorm:"column:stime"`
	ETime             xtime.Time `json:"etime" form:"etime" gorm:"column:etime"`
	DeletedFlag       int        `json:"-" form:"-" gorm:"column:deleted_flag"`
	CUser             string     `json:"c_user" form:"-" gorm:"column:cuser"`
	MUser             string     `json:"m_user" form:"-" gorm:"column:muser"`
	CTime             xtime.Time `json:"ctime" form:"-" gorm:"column:ctime"`
	MTime             xtime.Time `json:"mtime" form:"-" gorm:"column:mtime"`
}

// PopupConfigBuild 版本限制
type PopupConfigBuild struct {
	Plat       int    `json:"plat"`
	Build      int    `json:"build"`
	Conditions string `json:"conditions"`
}

// TableName 表名
func (*PopupConfig) TableName() string {
	return "popup_config"
}

type PopupConfigListWithPager struct {
	List  []*PopupConfig `json:"list"`
	Pager *Pager         `json:"pager"`
}

// PopupBuildsToJSON 将builds转化为JSON字符串
func PopupBuildsToJSON(builds []*PopupConfigBuild) (res string, err error) {
	//nolint:gosimple
	if builds == nil || len(builds) == 0 {
		err = xecode.NothingFound
		return
	}
	var resBytes []byte
	if resBytes, err = json.Marshal(builds); err != nil {
		return
	}
	res = string(resBytes)
	return
}

// PopupBuildsFromJSON 将JSON转化为builds结构体
func PopupBuildsFromJSON(buildsJSON string) (res []*PopupConfigBuild, err error) {
	if err = json.Unmarshal([]byte(buildsJSON), &res); err != nil {
		return
	}
	return
}
