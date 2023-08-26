package lol

import esmdl "go-gateway/app/web-svr/activity/interface/model/esports_model"

const (
	FormulaName  = "total"
	ActivityName = "S10Activity"
)

var OpTypeTipMap = map[string]string{
	"sign":      "完成活动签到",
	"share":     "完成视频/直播间分享",
	"pred":      "完成参与预测",
	"pred_succ": "赛事预测正确",
	"live_view": "完成观看直播",
	"view":      "完成观看视频",
	"videoup":   "完成视频投稿",
	"goods":     "兑换%s",
	"b_member":  "完成影视番剧分会场任务",
	"b_mall":    "完成会员购分会场任务",
	"b_manga":   "完成漫画分会场任务",
	"b_live":    "关注推荐主播",
}

type PointMsg struct {
	Timestamp int64  `json:"timestamp"`
	OpTypeTip string `json:"op_type_tip"`
	OpType    string `json:"op_type"`
	Point     int64  `json:"point"`
}

type PointList struct {
	List     []*PointMsg `json:"list"`
	SeasonID int64       `json:"season_id"`
}

type ContestDetail struct {
	ContestID  int64               `json:"contest_id"`
	Timestamp  int64               `json:"timestamp"`
	Title      string              `json:"title"`
	Home       esmdl.Team4Frontend `json:"home"`
	Away       esmdl.Team4Frontend `json:"away"`
	Status     string              `json:"status"`
	Win        string              `json:"win"`
	Predict    string              `json:"predict"`
	Coins      float64             `json:"coins"`
	Settlement string              `json:"settlement_status"`
	MyGuess    string              `json:"my_guess"`
	Stake      float64             `json:"stake"`
}

type PredictDetail struct {
	ContestID  int64  `json:"contest_id"`
	PredTeam   string `json:"pred_team"`
	PredStatus string `json:"pred_status"`
	PredCoins  int    `json:"pred_coins"`
	WinCoins   int    `json:"win_coins"`
}

type PredictMsg struct {
	Predicts    int     `json:"predicts"`
	PredictWins int     `json:"predict_wins"`
	Coins       float64 `json:"coins"`
}

type UserGuessOid struct {
	ID         int64   `json:"id"`
	MainID     int64   `json:"main_id"`
	DetailID   int64   `json:"detail_id"`
	Stake      float64 `json:"stake"`
	Income     float64 `json:"income"`
	Status     int64   `json:"status"`
	DetailName string  `json:"detail_name"`
	Oid        int64   `json:"oid"`
}

// DetailOption.
type DetailOption struct {
	MainID   int64  `json:"main_id"`
	DetailID int64  `json:"detail_id"`
	Option   string `json:"option"`
	Oid      int64  `json:"oid"`
}
