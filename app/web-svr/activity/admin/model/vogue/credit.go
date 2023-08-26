package vogue

import (
	"go-common/library/time"
)

const (
	CategoryWithdraw    = "withdraw"
	CategoryDeposit     = "deposit"
	MethodView          = "view"
	MethodInvite        = "invite"
	MethodPrize         = "prize"
	MethodLottery       = "lottery"
	MethodDeduct        = "deduct"
	ScoreSymbolNegtive  = "-"
	ScoreSymbolPositive = "+"
	// responses are limited to 1000 history records
	ActPlatGetLimit = 1000
)

var (
	CategoryToStr = map[string]string{
		CategoryWithdraw: "支出",
		CategoryDeposit:  "入账",
	}
	MethodToStr = map[string]string{
		MethodView:    "观看视频",
		MethodInvite:  "邀请好友",
		MethodPrize:   "兑换",
		MethodLottery: "抽奖",
		MethodDeduct:  "异常扣除",
	}
)

// 积分进度信息
type CreditData struct {
	Uid int64 `json:"uid" gorm:"column:uid"`
	// 用户昵称，后期拼进来
	NickName string `json:"nickname" gorm:"-"`
	Goods    int64  `json:"goods" gorm:"column:goods"`
	// 礼物名称
	GoodsName string `json:"goods_name" gorm:"column:goods_name"`
	// 礼物要求积分
	GoodsScoreSetting int64 `json:"goods_score_setting" gorm:"column:goods_score_setting"`
	// 好友邀请所得积分
	ScoreInvite int64 `json:"score_invite" gorm:"-"`
	// 看视频积分
	ScoreView int64 `json:"score_view" gorm:"-"`
	// 已消耗积分
	ScoreCost int64 `json:"score_cost" gorm:"-"`
	// 累计积分
	ScoreTotal int64 `json:"score_total" gorm:"-"`
	// 剩余积分
	ScoreRemain int64 `json:"score_remain" gorm:"-"`
	// 是否存在异常
	Risk bool `json:"risk" gorm:"-"`
	// 是否存在异常的文案：无/黑名单/信用分低
	RiskMsg string `json:"risk_msg" gorm:"-"`
	// 提交礼物时间
	TaskStartTime string    `json:"task_start_time" gorm:"-"`
	Ctime         time.Time `json:"ctime" gorm:"column:ctime" time_format:"2006-01-02 15:04:05"`
	Mtime         time.Time `json:"mtime" gorm:"column:mtime" time_format:"2006-01-02 15:04:05"`
}

type RiskInfo struct {
	Risk    bool   `json:"risk" gorm:"-"`
	RiskMsg string `json:"risk_msg" gorm:"-"`
}

// 积分进度
type CreditSearch struct {
	Uid int64 `form:"uid" default:"-1"`
	Pn  int64 `form:"pn" default:"1" validate:"min=1"`
	Ps  int64 `form:"ps" default:"15" validate:"min=1,max=50000"`
}

type CreditExportSearch struct {
	Uid int64 `form:"uid" default:"-1"`
	Pn  int64 `form:"pn" default:"1"`
	Ps  int64 `form:"ps" default:"50000"`
}

// ListRsp
type CreditListRsp struct {
	List []*CreditData `json:"list"`
	Page *Page         `json:"page"`
}

// 用户积分详细条目
type CreditItem struct {
	// 积分变化时间
	Ctime time.Time `json:"ctime" gorm:"column:ctime"`
	// 明细类型："withdraw" - 取出积分，"deposit" - 存入积分
	Category string `json:"category" gorm:"column:category"`
	// 方式："view" - 观看视频，"invite" - 邀请好友，"prize" - 兑换，"lottery" - 抽奖，"deduct" - 异常扣除
	Method string `json:"method" gorm:"column:method"`
	// 积分变化
	Score int64 `json:"score" gorm:"column:score"`
	// 积分变化符号 +/-
	ScoreSymbol string `json:"score_symbol" default:"+"`
	// 剩余积分
	ScoreRemain int64 `json:"score_remain"`
	// 积分对应详细信息，invite时为好友信息，view时为av号
	Detail string `json:"detail"`
	// 视频信息
	Video string `json:"video"`
	// 好友信息
	Friend string `json:"friend"`
}

// 用户积分详情搜索参数
type CreditDetailSearch struct {
	Uid int64 `form:"uid" validate:"required"`
}

// 用户积分详情搜索返回
type CreditDetailListRsp struct {
	List []*CreditItem `json:"list"`
}

type CreditUserInvite struct {
	ID    int64     `json:"id" gorm:"column:id"`
	Uid   int64     `json:"uid" gorm:"column:uid"`
	Mid   int64     `json:"mid" gorm:"column:mid"`
	Score int64     `json:"score" gorm:"column:score"`
	Ctime time.Time `json:"ctime" gorm:"column:ctime" time_format:"2006-01-02 15:04:05"`
	Mtime time.Time `json:"mtime" gorm:"column:mtime" time_format:"2006-01-02 15:04:05"`
}

type CreditUserCost struct {
	ID    int64     `json:"id" gorm:"column:id"`
	Mid   int64     `json:"mid" gorm:"column:mid"`
	Cost  int64     `json:"score" gorm:"column:cost"`
	Goods int64     `json:"goods" gorm:"column:goods"`
	Ctime time.Time `json:"ctime" gorm:"column:ctime" time_format:"2006-01-02 15:04:05"`
	Mtime time.Time `json:"mtime" gorm:"column:mtime" time_format:"2006-01-02 15:04:05"`
}

// CreditViewSource view source msg
type CreditViewSource struct {
	Mid int64 `json:"mid"`
	Aid int64 `json:"aid"`
	Cid int64 `json:"cid"`
}

type CreditExportData struct {
	FilePath string `json:"file_path"`
	Ctime    int64  `json:"ctime"`
	Mtime    int64  `json:"mtime"`
}
