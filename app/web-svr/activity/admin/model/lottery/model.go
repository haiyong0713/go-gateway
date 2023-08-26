package lottery

import (
	"strconv"
	"strings"

	"go-common/library/log"
	xtime "go-common/library/time"
)

const (
	// LotteryDraftStateDraft 草稿
	LotteryDraftStateDraft = 2
	// LotteryDraftStateWaitReview 待审
	LotteryDraftStateWaitReview = 3
	// LotteryDraftStateWaitReviewed 已通过审核并同步完成
	LotteryDraftStateWaitReviewed = 0
	// LotteryDraftStateWaitReject 已退回
	LotteryDraftStateWaitReject = 4
	// LotteryDraftStateOffline 已下线
	LotteryDraftStateOffline = 1
	// LotteryDraftListAll 列表展示全部
	LotteryDraftListAll = 100
	// CanEdit 能编辑
	CanEdit = 1
	// CanNotEdit 不能编辑
	CanNotEdit = 2
	// AddTimesBatchLogStateInit 初始化
	AddTimesBatchLogStateInit = 1
	// AddTimesBatchLogStateFinish 完成
	AddTimesBatchLogStateFinish = 2
	// AddTimesBatchLogStateError 失败
	AddTimesBatchLogStateError = 3
	// AddTimesBatchLogStateFileError 文件导入失败
	AddTimesBatchLogStateFileError = 4
)

// LotteryInfo is act_lottery model
type LotInfo struct {
	ID         int64      `json:"id"`
	LotteryID  string     `json:"lottery_id"`
	Name       string     `json:"name"`
	IsInternal int        `json:"is_internal"`
	Type       int        `json:"type"`
	State      int        `json:"state"`
	STime      xtime.Time `json:"stime"`
	ETime      xtime.Time `json:"etime"`
	CTime      xtime.Time `json:"ctime"`
	MTime      xtime.Time `json:"mtime"`
	Author     string     `json:"author"`
	CanEdit    int        `json:"can_edit"`
}

// LotInfoDraft is act_lottery model
type LotInfoDraft struct {
	ID                int64      `json:"id"`
	LotteryID         string     `json:"lottery_id"`
	Name              string     `json:"name"`
	IsInternal        int        `json:"is_internal"`
	Type              int        `json:"type"`
	State             int        `json:"state"`
	STime             xtime.Time `json:"stime"`
	ETime             xtime.Time `json:"etime"`
	CTime             xtime.Time `json:"ctime"`
	MTime             xtime.Time `json:"mtime"`
	Author            string     `json:"author"`
	Reviewer          string     `json:"reviewer"`
	CanReviewer       string     `json:"can_reviewer"`
	RejectReason      string     `json:"reject_reason"`
	LastAuditPassTime int64      `json:"last_audit_pass_time"`
}

func GetUploadKey(sid string, aid int64) string {
	return sid + "_" + strconv.FormatInt(aid, 10)
}

func SplitUploadKey(key string) (sid string, aid int64, err error) {
	rspTmp := strings.Split(key, "_")
	sid = rspTmp[0]
	if aid, err = strconv.ParseInt(rspTmp[1], 10, 64); err != nil {
		log.Error("SplitUploadKey strconv.ParseInt() failed.")
	}
	return
}

func GetTaskKey(sid string, aid int64, t int) string {
	return sid + "|" + strconv.FormatInt(aid, 10) + "|" + strconv.Itoa(t)
}

// GiftTypeVIPParams ...
type GiftTypeVIPParams struct {
	Token  string `json:"token"`
	AppKey string `json:"app_key"`
}

// GiftTypeGrantParams ...
type GiftTypeGrantParams struct {
	Pid    int `json:"pid"`
	Expire int `json:"expire"`
}

// GiftTypeCoinParams ...
type GiftTypeCoinParams struct {
	Coin int `json:"coin"`
}

// GiftTypeVipCouponParams ...
type GiftTypeVipCouponParams struct {
	Token  string `json:"token"`
	AppKey string `json:"app_key"`
}

// GiftTypeOGVParams ...
type GiftTypeOGVParams struct {
	Token string `json:"token"`
}

// GiftTypeVipBuyParams ...
type GiftTypeVipBuyParams struct {
	Token string `json:"token"`
}

// GiftTypeAwardParams ...
type GiftTypeAwardParams struct {
	AwardID int64 `json:"award_id"`
}

// GiftTypeMoneyParams ...
type GiftTypeMoneyParams struct {
	Money      int    `json:"money"`
	CustomerID string `json:"customer_id"`
	ActivityID string `json:"activity_id"`
	TransDesc  string `json:"trans_desc"`
	StartTime  int    `json:"start_time"`
}
