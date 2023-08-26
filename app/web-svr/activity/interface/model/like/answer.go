package like

type ParamQuestion struct {
	CurrentRound int64 `form:"current_round"`
	Pn           int   `form:"pn" validate:"min=1" default:"1"`
	Ps           int   `form:"ps" validate:"min=1,max=100" default:"20"`
}

type AnswerQuestion struct {
	CurrentRound int64           `json:"current_round"`
	UserHp       int64           `json:"user_hp"`
	KnowRule     int64           `json:"know_rule"`
	BaseID       int64           `json:"base_id"`
	List         []*QuestionItem `json:"list"`
}

type QuestionItem struct {
	QuestionOrder int      `json:"question_order"`
	ID            int64    `json:"id"`
	Attribute     int64    `json:"attribute"`
	Question      string   `json:"question"`
	Answers       []string `json:"answers"`
	Pic           string   `json:"pic"`
}

type QuestionRank struct {
	Account struct {
		Mid  int    `json:"mid"`
		Name string `json:"name"`
		Face string `json:"face"`
		Sign string `json:"sign"`
	} `json:"account"`
	UserScore   int `json:"user_score"`
	AnswerTimes int `json:"answer_times"`
	OrderNumber int `json:"order_number"`
}

type ParamResult struct {
	CurrentRound  int64  `form:"current_round" validate:"required"`
	QuestionID    int64  `form:"question_id" validate:"required"`
	QuestionOrder int    `form:"question_order" validate:"required"`
	UserAnswer    string `form:"user_answer"`
	Buvid         string `form:"buvid"`
	Origin        string `form:"origin"`
	UA            string `form:"ua"`
	Referer       string `form:"referer"`
	IP            string `form:"ip"`
	Build         int64  `form:"build"`
	Platform      string `form:"platform"`
	Device        string `form:"device"`
	MobiApp       string `form:"mobi_app"`
	OrderID       int64  `form:"-"`
	Topic         string `form:"-"`
	TopicType     int64  `form:"-"`
}

type AnswerResult struct {
	CurrentRound int64 `json:"current_round"`
	UserHp       int64 `json:"user_hp"`
	IsRight      int64 `json:"is_right"`
	TimeOut      int64 `json:"time_out"`
	NowScore     int64 `json:"now_score"`
	NowPercent   int64 `json:"now_percent"`
	UserScore    int64 `json:"user_score"`
	QuestionOver int64 `json:"question_over"`
}

type AnswerHp struct {
	CurrentBaseID int64
	Hp            int64
	NowScore      int64
	LastScore     int64
	ShareHp       int64
	StartTime     int64
	AnswerCount   int64
	HaveAddTime   int64
}

type AnswerUserInfo struct {
	IsJoin      int64         `json:"is_join"`
	UserScore   int64         `json:"user_score"`
	UserRank    int64         `json:"user_rank"`
	UserPercent int64         `json:"user_percent"`
	AnswerTimes int64         `json:"answer_times"`
	CanPendant  int64         `json:"can_pendant"`
	HavePendant int64         `json:"have_pendant"`
	LastInfo    *UserLastInfo `json:"last_info"`
	KnowRule    int64         `json:"know_rule"`
	FinishTime  int64         `json:"finish_time"`
}

type UserLastInfo struct {
	LastScore   int64 `json:"last_score"`
	LastPercent int64 `json:"last_percent"`
}

type UserRank struct {
	OrderNumber int          `json:"order_number"`
	UserScore   int64        `json:"user_score"`
	AnswerTimes int64        `json:"answer_times"`
	Account     *AccountInfo `json:"account"`
}

type RankInfo struct {
	Mid         int64 `json:"mid"`
	OrderNumber int   `json:"order_number"`
	UserScore   int64 `json:"user_score"`
	AnswerTimes int64 `json:"answer_times"`
}

// AccountInfo
type AccountInfo struct {
	Mid  int64  `json:"mid"`
	Name string `json:"name"`
	Face string `json:"face"`
}

// PendantRule
type PendantRule struct {
	CanPendant  int64 `json:"can_pendant"`
	HavePendant int64 `json:"have_pendant"`
	KnowRule    int64 `json:"know_rule"`
}

type HourPeople struct {
	WeekPeople  int64           `json:"week_people"`
	PeopleCount map[int64]int64 `json:"people_count"`
}

type GaiaResult struct {
	MID        int64  `json:"mid"`
	Buvid      string `json:"buvid"`
	IP         string `json:"ip"`
	Platform   string `json:"platform"`
	CTime      string `json:"ctime"`
	AccessKey  string `json:"access_key"`
	Caller     string `json:"caller"`
	API        string `json:"api"`
	Origin     string `json:"origin"`
	Referer    string `json:"referer"`
	UserAgent  string `json:"user_agent"`
	Build      string `json:"build"`
	Code       int64  `json:"code"`
	OrderID    int64  `json:"order_id"`
	Topic      string `json:"topic"`
	Action     string `json:"action"`
	TopicTime  string `json:"topic_time"`
	UserAnswer string `json:"user_answer"`
	TopicType  int64  `json:"topic_type"`
	Result     int64  `json:"result"`
}
