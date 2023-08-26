package like

import xtime "go-common/library/time"

type SelectionQAInfo struct {
	IsJoin      bool        `json:"is_join"`
	SelectionQA interface{} `json:"contests"`
}

type SelectionQA struct {
	Question string             `json:"question"`
	Answer   []*SelectionAnswer `json:"answer"`
}

type SelectionAnswer struct {
	Product string `json:"product"`
	Role    string `json:"role"`
}

type SelectionQADB struct {
	ID            int64  `json:"id"`
	Mid           int64  `json:"mid"`
	QuestionOrder int64  `json:"question_order"`
	Question      string `json:"question"`
	Product       string `json:"product"`
	Role          string `json:"role"`
}

type SelSensitive struct {
	IsSensitive bool `json:"is_sensitive"`
}

type TwoRes struct {
	Count   int    `json:"count"`
	Product string `json:"product"`
	Role    string `json:"role"`
}

type SelCategory struct {
	CategoryID   int64
	CategoryName string
	CategoryType int64
}

type ProductRoleDB struct {
	ID           int64      `json:"id"`
	CategoryID   int64      `json:"category_id"`
	CategoryType int64      `json:"category_type"`
	Role         string     `json:"role"`
	Product      string     `json:"product"`
	Tags         string     `json:"tags"`
	TagsType     int64      `json:"tags_type"`
	VoteNum      int64      `json:"vote_num"`
	Ctime        xtime.Time `json:"ctime"`
	Mtime        xtime.Time `json:"mtime"`
}

type ProductRole struct {
	ID           int64   `json:"id"`
	CategoryID   int64   `json:"category_id"`
	CategoryType int64   `json:"category_type"`
	Role         string  `json:"role"`
	Product      string  `json:"product"`
	Voted        bool    `json:"voted"`
	Percent      float64 `json:"percent"`
	VoteNum      int64   `json:"vote_num"`
	Mtime        int64   `json:"mtime"`
	OrderNum     int     `json:"order_num"`
	HideVote     int64   `json:"-"`
}

type CategoryPR struct {
	ShowVotes  bool `json:"show_votes"`
	IsVote     bool `json:"is_vote"`
	IsLogin    bool `json:"is_login"`
	IsStart    bool `json:"is_start"`
	IsChecking bool `json:"is_checking"`
	List       []*ProductRole
}

type ProductroleVote struct {
	VoteNum int64
	Mtime   int64
}

type ProductRoleArc struct {
	Aid     int64 `json:"aid"`
	PubDate int64 `json:"pub_date"`
}

type ParamVote struct {
	CategoryID    int64  `form:"category_id" validate:"required"`
	ProductRoleID int64  `form:"productrole_id" validate:"required"`
	Buvid         string `form:"buvid"`
	Origin        string `form:"origin"`
	UA            string `form:"ua"`
	Referer       string `form:"referer"`
	IP            string `form:"ip"`
	Build         string `form:"build"`
	Platform      string `form:"platform"`
	Device        string `form:"device"`
	MobiApp       string `form:"mobi_app"`
	CategoryName  string `form:"-"`
	ProductName   string `form:"-"`
}

type VoteEventCtx struct {
	Action       string `json:"action"`
	Mid          int64  `json:"mid"`
	ActivityUid  string `json:"activity_uid"`
	ID           int64  `json:"id"`
	Content      string `json:"content"`
	CategoryID   int64  `json:"category_id"`
	CategoryName string `json:"category_name"`
	Buvid        string `json:"buvid"`
	Ip           string `json:"ip"`
	Platform     string `json:"platform"`
	Ctime        string `json:"ctime"`
	Api          string `json:"api"`
	Origin       string `json:"origin"`
	UserAgent    string `json:"user_agent"`
	Build        string `json:"build"`
	MobiApp      string `json:"mobi_app"`
	Referer      string `json:"referer"`
}

type ParamAssistance struct {
	CategoryID    int64 `form:"category_id" validate:"required"`
	ProductRoleID int64 `form:"productrole_id" validate:"required"`
	OrderType     int64 `form:"order_type"`
	Pn            int   `form:"pn" validate:"min=1" default:"1"`
	Ps            int   `form:"ps" validate:"min=1,max=100" default:"50"`
}

type AssistanceRes struct {
	Product string       `json:"product"`
	Role    string       `json:"role"`
	List    []*ArcBvInfo `json:"list"`
	Page    *Page        `json:"page"`
}
