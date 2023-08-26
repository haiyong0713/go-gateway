package newyear2021

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go-gateway/app/web-svr/activity/interface/api"
)

type ExamProducer struct {
	Topic     string   `toml:"topic"`
	Addresses []string `toml:"addresses"`
}

type ExamPub struct {
	MID     int64 `json:"MID"`
	ItemID  int64 `json:"ItemID"`
	OptID   int64 `json:"OptID"`
	LogTime int64 `json:"LogTime"`
}

type RiskManagementReportInfoOfExam struct {
	MID       int64  `json:"mid"`
	Buvid     string `json:"buvid"`
	IP        string `json:"ip"`
	Platform  string `json:"platform"`
	CTime     string `json:"ctime"`
	AccessKey string `json:"access_key"`
	Caller    string `json:"caller"`
	API       string `json:"api"`
	Origin    string `json:"origin"`
	Referer   string `json:"referer"`
	UserAgent string `json:"user_agent"`
	Build     string `json:"build"`

	// 第几次答题
	OrderID int64 `json:"order_id"`
	// 答题时间
	TopicTime int64 `json:"topic_time"`
	// 用户选项
	UserAnswer int64 `json:"user_answer"`
	// 回答是否正确
	Result int64 `json:"result"`
}

type RiskManagementReportInfoOfGame struct {
	MID       int64  `json:"mid"`
	Buvid     string `json:"buvid"`
	IP        string `json:"ip"`
	Platform  string `json:"platform"`
	CTime     string `json:"ctime"`
	AccessKey string `json:"access_key"`
	Caller    string `json:"caller"`
	API       string `json:"api"`
	Origin    string `json:"origin"`
	Referer   string `json:"referer"`
	UserAgent string `json:"user_agent"`
	Build     string `json:"build"`

	Count    int64 `json:"playcount_end"`
	GameType int64 `json:"game_type"`
	EndTime  int64 `json:"endtime"`
	Coupon   int64 `json:"point"`
	// 游戏预设时长
	Duration int64 `json:"gametime"`
}

type AllARBlackList struct {
	Ios      *ARBlackList `json:"ios"`
	Android  *ARBlackList `json:"android"`
	Redirect *AppRedirect `json:"redirect"`
}

type AppRedirect struct {
	UnSupportAppH5   string `json:"un_support_app_h5"`
	UnSupportBuildH5 string `json:"un_support_build_h5"`
	GameH5           string `json:"game_h5"`
	ARScheme         string `json:"ar_scheme"`
	TaskGame         string `json:"task_game"`
}

type ARBlackList struct {
	ModelMap        map[string]int64 `json:"model_map"`
	VersionMap      map[string]int64 `json:"version_map"`
	SupportVersion  float64          `json:"support_version"`
	HighScore       int64            `json:"high_score"`
	MiddleScore     int64            `json:"middle_score"`
	MemoryRuleList  []*MemoryRule    `json:"memory_rule_list"`
	VersionRuleList []*VersionRule   `json:"version_rule_list"`
	SupportBuild    int64            `json:"support_build"`
}

type VersionRule struct {
	BizType int64   `json:"biz_type"`
	Score   int64   `json:"score"`
	Version float64 `json:"version"`
	Level   string  `json:"level"`
}

type MemoryRule struct {
	Threshold int64  `json:"threshold"`
	Level     string `json:"level"`
}

type UserAppInfo struct {
	Os        string  `json:"os"`
	Model     string  `json:"model"`
	OsVersion float64 `json:"os_version"`
	Build     int64   `json:"build"`
	MobiApp   string  `json:"mobi_app"`

	OsVersionOfOrigin string `json:"os_version"`
}

type ARDeviceReportInfo struct {
	Os        string `json:"Os"`
	Model     string `json:"Model"`
	OsVersion string `json:"OsVersion"`
	AppBuild  int64  `json:"AppBuild"`
	MobiApp   string `json:"MobiApp"`
	Score     int64  `json:"Score"`
	SceneID   string `json:"SceneID"`
}

type BnjExamResponse struct {
	Bank  []*BnjExamItem `json:"bank"`
	Sleep int64          `json:"sleep"`
}

type BnjExamItem struct {
	ID        int64            `json:"id" toml:"id"`
	IDStr     string           `json:"-" toml:"id_str"`
	StartTime int64            `json:"start_time" toml:"start_time"`
	EndTime   int64            `json:"end_time" toml:"end_time"`
	Title     string           `json:"title" toml:"title"`
	Options   []*BnjExamOption `json:"options" toml:"options"`
	Answer    int64            `json:"answer" toml:"answer"`
	UserOpt   int64            `json:"user_option"  toml:"user_option"`
	Status    int64            `json:"status"`
}

type BnjExamOption struct {
	ID    int64  `json:"id" toml:"id"`
	Title string `json:"title" toml:"title"`
	Count int64  `json:"count"`
}

type UserRewardInLiveRoom struct {
	MID         int64                      `json:"mid"`
	SceneID     int64                      `json:"scene_id"`
	ReceiveUnix int64                      `json:"related_id"`
	No          int64                      `json:"no"`
	Duration    int64                      `json:"duration"`
	Reward      *api.RewardsSendAwardReply `json:"reward"`
}

type LiveRewardDetail struct {
	SceneID int64 `json:"scene_id"`
	Quota   int64 `json:"quota"`
}

type PublicizeAggregation struct {
	AR         *ARInPublicize         `json:"AR"`
	Reserve    *ReserveInPublicize    `json:"reserve"`
	PCResource map[string]interface{} `json:"pc_resource"`
}

type ARInPublicize struct {
	Coupon int64 `json:"coupon"`
	DrawUV int64 `json:"draw_uv"`
}

type BnjStrategy struct {
	LiveStartTime  int64  `json:"live_start_time"`
	LiveEndTime    int64  `json:"live_end_time"`
	DrawLevelLimit int32  `json:"draw_level_limit"`
	DWLogID4Draw   string `json:"dw_log_id_4_draw"`
	BackupPub      bool   `json:"backup_pub"`
}

type ReserveInPublicize struct {
	IsLogin             int64 `json:"is_login"`
	Total               int64 `json:"total"`
	Reserved            int64 `json:"reserved"`
	ActivityID          int64 `json:"activity_id"`
	ActivityComponentID int64 `json:"activity_component_id"`
}

type ARConfig struct {
	AR             ARSetting `json:"AR" toml:"ar"`
	Notice         ARNotice  `json:"notice" toml:"notice"`
	Publish        ARPublish `json:"publish" toml:"publish"`
	Timer          int64     `json:"timer" toml:"timer"`
	ConfirmMessage string    `json:"confirm_message"`
}

type ARNotice struct {
	Show   int64  `json:"show" toml:"show"`
	Notice string `json:"notice" toml:"notice"`
}

type ARPublish struct {
	VideoArea string `json:"video_area" toml:"video_area"`
	VideoTag  string `json:"video_tag" toml:"video_tag"`
	Text      string `json:"text" toml:"text"`
}

type ARSetting struct {
	// 单局游戏时长
	GameDuration int64 `json:"duration_of_game" toml:"duration_of_game"`
	// 嘟嘴强化次数, 超过次数后不识别
	MaxPoutingTimes int64 `json:"pouting_max_times" toml:"pouting_max_times"`
	// 比心大招次数, 超过次数后不识别
	MaxHadHeardTimes int64 `json:"hand_heard_max_times" toml:"hand_heard_max_times"`
	// 怪物1血量
	Monster1Blood int64 `json:"total_blood_of_monster1" toml:"total_blood_of_monster1"`
	// 怪物2血量
	Monster2Blood int64 `json:"total_blood_of_monster2" toml:"total_blood_of_monster2"`
	// 怪物3血量
	Monster3Blood int64 `json:"total_blood_of_monster3" toml:"total_blood_of_monster3"`
	// 怪物4血量
	Monster4Blood int64 `json:"total_blood_of_monster4" toml:"total_blood_of_monster4"`
	// 飞机子弹发射速率
	AircraftBulletRate float64 `json:"firing_rate_of_aircraft_bullets" toml:"firing_rate_of_aircraft_bullets"`
	// 怪物刷新速率
	MonsterRefreshRate float64 `json:"refresh_rate_of_monsters" toml:"refresh_rate_of_monsters"`
	// 单个子弹伤害
	BulletDamage int64 `json:"damage_of_bullet" toml:"damage_of_bullet"`
	// 炸弹伤害值
	BombDamage int64 `json:"damage_of_bomb" toml:"damage_of_bomb"`
	// 怪物1消灭得分
	Monster1KillScore int64 `json:"score_of_kill_monster1" toml:"score_of_kill_monster1"`
	// 怪物2消灭得分
	Monster2KillScore int64 `json:"score_of_kill_monster2" toml:"score_of_kill_monster2"`
	// 怪物3消灭得分
	Monster3KillScore int64 `json:"score_of_kill_monster3" toml:"score_of_kill_monster3"`
	// 怪物4消灭得分
	Monster4KillScore int64 `json:"score_of_kill_monster4" toml:"score_of_kill_monster4"`
	// 怪物持续存在时长
	MonsterDisplayDuration int64 `json:"display_duration_of_monster" toml:"display_duration_of_monster"`
	//单局游戏后x秒内指定怪物的刷新频率
	MonsterRefreshRateAfter int64 `json:"time_to_change_monster_refresh_rate" toml:"time_to_change_monster_refresh_rate"`
	// 怪物1刷新速率
	Monster1RefreshRate float64 `json:"refresh_rate_of_monster1" toml:"refresh_rate_of_monster1"`
	// 怪物2刷新速率
	Monster2RefreshRate float64 `json:"refresh_rate_of_monster2" toml:"refresh_rate_of_monster2"`
	// 怪物3刷新速率
	Monster3RefreshRate float64 `json:"refresh_rate_of_monster3" toml:"refresh_rate_of_monster3"`
	// 怪物4刷新速率
	Monster4RefreshRate float64 `json:"refresh_rate_of_monster4" toml:"refresh_rate_of_monster4"`
	// 单局游戏后x秒内指定怪物1的刷新频率
	Monster1ChangeRefreshRate float64 `json:"time_to_change_monster_refresh_rate1" toml:"time_to_change_monster_refresh_rate1"`
	// 单局游戏后x秒内指定怪物2的刷新频率
	Monster2ChangeRefreshRate float64 `json:"time_to_change_monster_refresh_rate2" toml:"time_to_change_monster_refresh_rate2"`
	// 单局游戏后x秒内指定怪物3的刷新频率
	Monster3ChangeRefreshRate float64 `json:"time_to_change_monster_refresh_rate3" toml:"time_to_change_monster_refresh_rate3"`
	// 单局游戏后x秒内指定怪物4的刷新频率
	Monster4ChangeRefreshRate float64 `json:"time_to_change_monster_refresh_rate4" toml:"time_to_change_monster_refresh_rate4"`
	// 动漫脸开关
	OpenAnimationFace int64 `json:"open_animation_face" toml:"open_animation_face"`
	// 每日游戏次数
	DayGameTimes int64 `json:"day_times" toml:"day_times"`
}

type Profile struct {
	Score  int64       `json:"score"`
	Coupon *UserCoupon `json:"coupon"`
	Exp    int64       `json:"exp"`
}

type ARConfirm struct {
	Confirm int64  `json:"confirm"`
	Message string `json:"message"`
}

type GamePreCommitResp struct {
	GameCommitResp
	RequestID string `json:"request_id"`
}

type GameCommitResp struct {
	Reward Score2Coupon `json:"reward"`
	Quota  int64        `json:"quota"`
}

type GameScore struct {
	Score     int64  `json:"score" form:"score" validate:"min=0"`
	RequestID string `json:"request_id" form:"request_id" validate:"required"`
	GameType  int64  `json:"game_type" form:"game_type" validate:"min=1,max=2" default:"1"`
}

type Score2Coupon struct {
	Score  int64 `json:"score"`
	Coupon int64 `json:"coupon"`
}

type ARGameLog struct {
	Score  int64  `json:"score"`
	Coupon int64  `json:"coupon"`
	MID    int64  `json:"mid"`
	Date   string `json:"date"`
	Index  int64  `json:"index"`
}

type UserCoupon struct {
	ND   int64 `json:"nd"`
	Live int64 `json:"live"`
}

type CouponLog struct {
	ID        int64
	MID       int64
	Code      int64
	Comment   string
	Num       int64
	CreatedAt time.Time
}

const (
	AwardTypeExp = "exp"
	//拜年祭奖券
	AwardTypeLottery = "lottery"
	//漫画折扣卷(9折)
	AwardTypeComics90PercentCoupon = "comics_coupon_90"
	//角色形象
	AwardTypeRoleImage = "role_image"
	//漫画卡
	AwardTypeComicsCard = "comics_card"
	//会员购满减卷
	AwardTypeVipMallCoupon = "vip_mall_coupon"
	//魔晶
	AwardTypeMojing = "mojing"
	//头像挂件
	AwardTypePendant = "pendant"
	//限时活动动态卡片(一个月)
	AwardTypeActivityCard = "activity_card"
	//直播弹幕
	AwardTypeLiveDanmaku = "live_danmaku"
	//漫画折扣卷(7折)
	AwardTypeComics70PercentCoupon = "comics_coupon_70"
	//漫画折扣卷(5折)
	AwardTypeComics50PercentCoupon = "comics_coupon_50"
	//限时活动点赞动效
	AwardTypeLikeEfficacy = "like_efficacy"
)

type LevelTaskStatus struct {
	FinishCount int64              `json:"finish_count"`
	Tasks       []*LevelTaskResult `json:"tasks"`
}

type AwardInfo struct {
	AwardName string            `json:"award_name"`
	Type      string            `json:"type"`
	Icon      string            `json:"icon"`
	ExtraInfo map[string]string `json:"extra_info"`
}

type LevelTaskResult struct {
	Id            int64            `json:"id"`
	Name          string           `json:"name"`
	RequiredCount int64            `json:"require_count"`
	PcUrl         string           `json:"pc_url"`
	H5Url         string           `json:"h5_url"`
	IsFinish      bool             `json:"completed"`
	IsReceived    bool             `json:"received"`
	Award         *AwardInfo       `json:"award_info"`
	HiddenTask    *LevelTaskResult `json:"hidden_task"`
}

type TaskResult struct {
	Id            int64      `json:"id"`
	Name          string     `json:"name"`
	FinishCount   int64      `json:"finish_count"`
	RequiredCount int64      `json:"require_count"`
	PcUrl         string     `json:"pc_url"`
	H5Url         string     `json:"h5_url"`
	IsFinish      bool       `json:"completed"`
	IsReceived    bool       `json:"received"`
	Award         *AwardInfo `json:"award_info"`
}

type ExtraResult struct {
	Id         int64      `json:"id"`
	Name       string     `json:"name"`
	IsFinish   bool       `json:"completed"`
	IsReceived bool       `json:"received"`
	PcUrl      string     `json:"pc_url"`
	H5Url      string     `json:"h5_url"`
	Award      *AwardInfo `json:"award_info"`
}
type PersonalTaskResult struct {
	DailyTasks map[string] /*taskId*/ *TaskResult `json:"daily_tasks"`
}

type User struct {
	Mid     int64
	Lottery int64
	Exp     int64
}

const (
	CouponTypeCodeOfNiuDan = 1
	CouponTypeCodeOfLive   = 2
	CouponCommentOfARGame  = "AR打年兽获取"

	ExamStatusOfNotBegin      = 1
	ExamStatusOfDoing         = 2
	ExamStatusOfEnd           = 3
	ExamStatusOfEnd300Seconds = 4

	// 格式：得分xxx及以下， 版本在xx以上的
	VersionRuleBizType4First = 1

	regExpString4UserAgent = `os/([a-zA-Z]+) model/([a-zA-Z0-9 ,-_]+)mobi_app/([a-zA-Z0-9\./_]+) build/([1-9]\d*) .*osVer/([0-9\.]*)`

	Os4Android       = "android"
	Os4AndroidOfBlue = "android_b"
	Os4Ios           = "ios"
	Os4IPad          = "ipad"

	MobileApp4Android = "android"
	MobileApp4IPhone  = "iphone"

	ARDeviceReportSceneOfUnSupportApp      = "un_support_app"
	ARDeviceReportSceneOfUnSupportBuild    = "un_support_build"
	ARDeviceReportSceneOfBlacklist         = "blacklist"
	ARDeviceReportSceneOfScoreLow          = "score_rule_low"
	ARDeviceReportSceneOfMemoryLow         = "memory_rule_low"
	ARDeviceReportSceneOfVersionRuleLow    = "version_rule_low"
	ARDeviceReportSceneOfVersionRuleMiddle = "version_rule_middle"
	ARDeviceReportSceneOfVersionRuleHigh   = "version_rule_high"
	ARDeviceReportSceneOfUnknownMiddle     = "unknown_middle"
)

var (
	regExp4UserAgent *regexp.Regexp
)

func init() {
	regExp4UserAgent = regexp.MustCompile(regExpString4UserAgent)
}

func NewScore2Coupon(mid, score, coupon, index int64, dateStr string) (info *ARGameLog) {
	info = new(ARGameLog)
	{
		info.MID = mid
		info.Score = score
		info.Coupon = coupon
		info.Date = dateStr
		info.Index = index
	}

	return
}

func (old *BnjExamOption) DeepCopy() (newOne *BnjExamOption) {
	newOne = new(BnjExamOption)
	{
		newOne.ID = old.ID
		newOne.Title = old.Title
		newOne.Count = old.Count
	}

	return
}

func DeepCopyExamBank(old []*BnjExamItem) (newOne []*BnjExamItem) {
	newOne = make([]*BnjExamItem, 0)
	if old != nil {
		for _, v := range old {
			newOne = append(newOne, v.DeepCopy())
		}
	}

	return
}

func (item *BnjExamItem) Rebuild(now int64) {
	item.IDStr = strconv.FormatInt(item.ID, 10)
	if now == 0 {
		now = time.Now().Unix()
	}

	if item.Options == nil {
		options := make([]*BnjExamOption, 0)
		item.Options = options
	}

	if now < item.StartTime {
		item.Answer = 0
		item.Title = ""
		item.Status = ExamStatusOfNotBegin

		for _, v := range item.Options {
			v.Count = 0
			v.Title = ""
		}
	} else if now < item.EndTime {
		item.Answer = 0
		item.Status = ExamStatusOfDoing
	} else {
		item.Status = ExamStatusOfEnd
		if now-item.EndTime >= 300 {
			item.Status = ExamStatusOfEnd300Seconds
		}
	}
}

func (old *BnjExamItem) DeepCopy() (newOne *BnjExamItem) {
	newOne = new(BnjExamItem)
	{
		options := make([]*BnjExamOption, 0)
		{
			if old.Options != nil {
				for _, v := range old.Options {
					options = append(options, v.DeepCopy())
				}
			}
		}

		newOne.ID = old.ID
		newOne.IDStr = old.IDStr
		newOne.StartTime = old.StartTime
		newOne.EndTime = old.EndTime
		newOne.Title = old.Title
		newOne.Status = old.Status
		newOne.UserOpt = old.UserOpt
		newOne.Answer = old.Answer
		newOne.Options = options
	}

	return
}

func (producer *ExamProducer) DeepCopy() (newOne *ExamProducer) {
	newOne = new(ExamProducer)
	{
		newOne.Topic = producer.Topic
		newOne.Addresses = make([]string, 0)
	}

	if producer.Addresses != nil && len(producer.Addresses) > 0 {
		for _, v := range producer.Addresses {
			newOne.Addresses = append(newOne.Addresses, v)
		}
	}

	return
}

func ParseUserAgent2UserAppInfo(ua string) (info *UserAppInfo, err error) {
	info = new(UserAppInfo)
	result := regExp4UserAgent.FindAllStringSubmatch(ua, -1)
	if len(result) >= 1 && len(result[0]) == 6 {
		info.Os = strings.ToLower(result[0][1])
		info.Model = strings.ToLower(strings.ReplaceAll(result[0][2], " ", ""))
		info.Build, err = strconv.ParseInt(result[0][4], 10, 64)
		if err != nil {
			return
		}

		info.OsVersion, err = parseOsVersionWithOneDecimal(result[0][5])
		info.OsVersionOfOrigin = result[0][5]
		info.MobiApp = result[0][3]
	} else {
		err = errors.New("unMatched user_agent")
	}

	return
}

func parseOsVersionWithOneDecimal(osVer string) (version float64, err error) {
	list := strings.Split(osVer, ".")
	switch len(list) {
	case 1:
		version, err = strconv.ParseFloat(list[0], 64)
	default:
		versionStr := fmt.Sprintf("%v.%v", list[0], list[1])
		version, err = strconv.ParseFloat(versionStr, 64)
	}

	return
}

func (info *UserAppInfo) Copy2ARDeviceReportInfo(score int64) (report *ARDeviceReportInfo) {
	report = new(ARDeviceReportInfo)
	{
		report.Os = info.Os
		report.Model = info.Model
		report.OsVersion = info.OsVersionOfOrigin
		report.AppBuild = info.Build
		report.MobiApp = info.MobiApp
		report.Score = score
	}

	return
}
