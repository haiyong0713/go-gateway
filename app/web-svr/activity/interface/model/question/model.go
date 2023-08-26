package question

//go:generate kratos t protoc --grpc question.proto

const (
	BusinessTypeAct = 1
)

// DetailItem .
type DetailItem struct {
	ID        int64    `json:"id"`
	BaseID    int64    `json:"base_id"`
	Name      string   `json:"name"`
	Pic       string   `json:"pic"`
	Answers   []string `json:"right_answer"`
	Attribute int64    `json:"attribute"`
}

// Answer .
type Answer struct {
	IsRight    int    `json:"is_right"`
	Finish     int    `json:"finish"`
	RightCount int    `json:"right_count"`
	Answer     string `json:"answer"`
	AnswerTime int64  `json:"answer_time"`
}

// Item .
type Item struct {
	ID      int64    `json:"id"`
	PoolID  int64    `json:"pool_id"`
	Name    string   `json:"name"`
	Pic     string   `json:"pic"`
	Answers []string `json:"answers"`
	Index   int64    `json:"index"`
}

// AnswerArg .
type AnswerArg struct {
	ID    int64 `form:"id" validate:"min=1"`
	Index int64 `form:"index" validate:"min=1"`
}

// QAReply QA接口返回
type QAReply struct {
	List   []*QAItem `json:"list"`
	PoolID int64     `json:"pool_id"`
}

// QAItem QA返回项
type QAItem struct {
	ID          int64    `json:"id"`
	Question    string   `json:"question"`
	RightAnswer []string `json:"right_answer"`
	AllAnswer   []string `json:"all_answer"`
}

// GKQuest  高考活动请求
type GKQuestReq struct {
	Sid      int64  `form:"sid"  json:"sid"   validate:"min=1"`
	Year     int    `form:"year"  json:"year" `
	Province string `form:"province"  json:"province" `
	Qtype    string `form:"qtype" json:"qtype" `
	Qid      string `form:"qid"  json:"qid"`
}

// GKRankReq  高考活动排行请求
type GKRankReq struct {
	Year       int    `form:"year" json:"year" validate:"min=1,max=2022"`
	Province   string `form:"province"  json:"province" validate:"required"`
	Course     string `form:"course"  json:"course" validate:"required"`
	Score      int    `form:"score" json:"score"`
	UsedTime   int    `form:"used_time" json:"used_time" validate:"min=1"`
	ReportTime int64  `form:"report_time" json:"report_time"`
}

type GKRankReply struct {
	ReportTime int64 `json:"report_time"`
	Total      int64 `json:"total"`
	Rank       int64 `json:"rank"`
}

// GKQAReply 高考QA接口返回
type GKQAReply struct {
	List []*GKQAItem `json:"list"`
}

type GKQAItem struct {
	Qid         int64    `json:"qid"`
	Qtype       int64    `json:"qtype"`
	Img         string   `json:"img"`
	Question    string   `json:"question"`
	RightAnswer []string `json:"right_answer"`
	AllAnswer   []string `json:"all_answer"`
}
