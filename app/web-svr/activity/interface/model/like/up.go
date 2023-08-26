package like

import (
	"time"

	xtime "go-common/library/time"
	"go-gateway/app/app-svr/archive/service/api"
)

const (
	UpAttrBitIsCycle = uint(0)
	UpAttrYes        = 1
	CycleDuration    = 86400
	HasFinish        = 1
	BusinessID       = 0
)

// ActUp .
type ActUp struct {
	ID          int64      `json:"id"`
	Mid         int64      `json:"mid"`
	Title       string     `json:"title"`
	Statement   string     `json:"statement"`
	Aid         int64      `json:"aid"`
	State       int32      `json:"state"`
	Offline     int32      `json:"offline"`
	Suffix      int64      `json:"suffix"`
	Attribute   int64      `json:"attribute"`
	FinishCount int64      `json:"finish_count"`
	Stime       xtime.Time `json:"stime"`
	Etime       xtime.Time `json:"etime"`
	Ctime       xtime.Time `json:"ctime"`
	Mtime       xtime.Time `json:"mtime"`
}

type ActUpReply struct {
	*ActUp
	Name  string `json:"name"`
	Image string `json:"image"`
	Face  string `json:"face"`
}

type UpCheck struct {
	Up     *ActUp `json:"up"`
	Status int    `json:"status"`
}

type ActUpPage struct {
	Dynamic interface{}   `json:"dynamic"`
	Act     *ActUpReply   `json:"act"`
	Archive *api.ArcReply `json:"archive"`
}

type UpActUserState struct {
	ID     int64      `json:"id"`
	Sid    int64      `json:"sid"`
	Mid    int64      `json:"mid"`
	Bid    int64      `json:"bid"`
	Round  int64      `json:"round"`
	Times  int64      `json:"times"`
	Finish int64      `json:"finish"`
	Result string     `json:"result"`
	Ctime  xtime.Time `json:"ctime"`
	Mtime  xtime.Time `json:"mtime"`
}

type RankUserDays struct {
	Mid  int64   `json:"mid"`
	Days float64 `json:"days"`
}

type RankUserInfo struct {
	Mid  int64  `json:"mid"`
	Days int64  `json:"days"`
	Name string `json:"name"`
	Face string `json:"face"`
}

type RankList struct {
	List     []*RankUserInfo `json:"list"`
	SelfDays int64           `json:"self_days"`
}

type AttendResult struct {
	TotalTime      int64   `json:"total_time"`
	MatchedPercent float32 `json:"matched_percent"`
}

type LIDWithVote struct {
	ID    int64 `json:"id"`
	Wid   int64 `json:"wid"`
	Vote  int64 `json:"vote"`
	Order int64 `json:"order"`
}

type GetActKnowledgeDetail struct {
	ID         int64  `json:"id"`
	MID        int64  `json:"mid"`
	Name       string `json:"name"`
	Face       string `json:"face"`
	FollowTime int64  `json:"follow_time"`
	VoteNum    int64  `json:"vote_num"`
	OrderNum   int64  `json:"order_num"`
	IsUp       int64  `json:"is_up"`
	IsSelect   int64  `json:"is_select"`
}

type ActKnowledgeRes struct {
	ActKnowledgeTop3List   `json:"top3"`
	ActKnowledgeDetailList `json:"list"`
	Left                   int64 `json:"left"`
}

type ActKnowledgeDetailList []*GetActKnowledgeDetail
type ActKnowledgeTop3List []*GetActKnowledgeDetail

func (ps ActKnowledgeDetailList) Len() int {
	return len(ps)
}

func (ps ActKnowledgeDetailList) Less(i, j int) bool {
	return ps[i].FollowTime > ps[j].FollowTime
}

func (ps ActKnowledgeDetailList) Swap(i, j int) {
	ps[i], ps[j] = ps[j], ps[i]
}

type FollowReply struct {
	MID   int64
	MTime xtime.Time
}

type GetMIDInfo struct {
	Mid  int64
	Name string
	Face string
}

func (a *ActUp) UpIsNoCycle() bool {
	return a.Attribute>>UpAttrBitIsCycle == UpAttrYes
}

func (a *ActUp) UpRound(nowT int64) (round int64) {
	if a.UpIsNoCycle() {
		return
	}
	if nowT > a.Etime.Time().Unix() {
		nowT = a.Etime.Time().Unix()
	}
	tm := a.Stime.Time().Unix()
	dataTimeStr := time.Unix(tm, 0).Format("2006-01-02")
	start, _ := time.ParseInLocation("2006-01-02 15:04:05", dataTimeStr+" 00:00:00", time.Local)
	st := start.Unix()
	return (nowT - st) / CycleDuration
}
