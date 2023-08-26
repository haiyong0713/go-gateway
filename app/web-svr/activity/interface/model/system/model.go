package system

import xtime "go-common/library/time"

const (
	SystemActivityTypeSign     = 1 // 签到活动
	SystemActivityTypeVote     = 2 // 投票活动
	SystemActivityTypeQuestion = 3 // 提问活动

	SystemActivityVoteSelect = 1 // 单选或多选

)

type User struct {
	Avatar         string `json:"avatar"`
	DepartmentName string `json:"departmentName"`
	ID             int64  `json:"id"`
	LastName       string `json:"lastName"`
	LoginID        string `json:"loginId"`
	NickName       string `json:"nickName"`
	WorkCode       string `json:"workcode"`
	UseKind        string `json:"useKind"`
}

type SystemUser struct {
	UID            string `json:"uid"`
	NickName       string `json:"nick_name"`
	Avatar         string `json:"avatar"`
	DepartmentName string `json:"department_name"`
	LastName       string `json:"last_name"`
	UseKind        string `json:"use_kind"`
	LoginID        string `json:"login_id"`
}

type DBUserInfo struct {
	ID    int64  `json:"id"`
	UID   string `json:"uid"`
	Token string `json:"token"`
}

type GetConfigArgs struct {
	Url  string `form:"url" validate:"required"`
	From string `form:"from" validate:"required"`
}

type GetConfigRes struct {
	CORPID    string `json:"corp_id"`
	Timestamp int64  `json:"timestamp"`
	NonceStr  string `json:"nonce_str"`
	Signature string `json:"signature"`
}

type WXAuthArgs struct {
	Code string `form:"code" validate:"required"`
	From string `form:"from" validate:"required"`
}

type Activity struct {
	ID     int        `json:"id"`
	Name   string     `json:"name"`
	Type   int        `json:"type"`
	Stime  xtime.Time `json:"stime"`
	Etime  xtime.Time `json:"etime"`
	Config string     `json:"config"`
}

type ActivitySign struct {
	ID       int64  `json:"id"`
	AID      int64  `json:"aid"`
	UID      string `json:"uid"`
	Location string `json:"location"`
}

type SystemActivitySignConfig struct {
	JumpURL  string `json:"jump_url"`
	Location int    `json:"location"`
	ShowSeat int    `json:"show_seat"`
	SeatID   int    `json:"seat_id"`
	SeatText string `json:"seat_text"`
}

type ActivityInfoRes struct {
	ID     int         `json:"-"`
	Name   string      `json:"name"`
	Type   int         `json:"type"`
	Stime  xtime.Time  `json:"stime"`
	Etime  xtime.Time  `json:"etime"`
	Config interface{} `json:"config"`
	Extra  interface{} `json:"extra"`
}

type SystemActivitySeat struct {
	ID      int    `json:"id"`
	AID     int64  `json:"aid"`
	UID     string `json:"uid"`
	Content string `json:"content"`
}

type Party2021Res struct {
	User struct {
		NickName    string `json:"nick_name"`
		Avatar      string `json:"avatar"`
		SeatContent string `json:"seat_content"`
	} `json:"user"`
	Sign   int64  `json:"sign"`
	QRCode string `json:"qr_code"`
}

type WXUserDetail struct {
	Errcode        int    `json:"errcode"`
	Errmsg         string `json:"errmsg"`
	Userid         string `json:"userid"`
	Name           string `json:"name"`
	Department     []int  `json:"department"`
	Order          []int  `json:"order"`
	Position       string `json:"position"`
	Mobile         string `json:"mobile"`
	Gender         string `json:"gender"`
	Email          string `json:"email"`
	IsLeaderInDept []int  `json:"is_leader_in_dept"`
	Avatar         string `json:"avatar"`
	ThumbAvatar    string `json:"thumb_avatar"`
	Telephone      string `json:"telephone"`
	Alias          string `json:"alias"`
	Address        string `json:"address"`
	OpenUserid     string `json:"open_userid"`
	MainDepartment int    `json:"main_department"`
	Extattr        struct {
		Attrs []struct {
			Type int    `json:"type"`
			Name string `json:"name"`
			Text struct {
				Value string `json:"value"`
			} `json:"text,omitempty"`
			Web struct {
				URL   string `json:"url"`
				Title string `json:"title"`
			} `json:"web,omitempty"`
		} `json:"attrs"`
	} `json:"extattr"`
	Status           int    `json:"status"`
	QrCode           string `json:"qr_code"`
	ExternalPosition string `json:"external_position"`
	ExternalProfile  struct {
		ExternalCorpName string `json:"external_corp_name"`
		ExternalAttr     []struct {
			Type int    `json:"type"`
			Name string `json:"name"`
			Text struct {
				Value string `json:"value"`
			} `json:"text,omitempty"`
			Web struct {
				URL   string `json:"url"`
				Title string `json:"title"`
			} `json:"web,omitempty"`
			Miniprogram struct {
				Appid    string `json:"appid"`
				Pagepath string `json:"pagepath"`
				Title    string `json:"title"`
			} `json:"miniprogram,omitempty"`
		} `json:"external_attr"`
	} `json:"external_profile"`
}

type SystemActivityVoteConfig struct {
	Items []struct {
		Title   string `json:"title"`
		Type    int64  `json:"type"`
		Options struct {
			Name []struct {
				Desc string `json:"desc"`
			} `json:"name"`
			LimitNum int64 `json:"limit_num"`
			Score    int64 `json:"score"`
		} `json:"options"`
	} `json:"items"`
}

type ActivityVote struct {
	ID       int64  `json:"id"`
	AID      int64  `json:"aid"`
	UID      string `json:"uid"`
	ItemID   int64  `json:"item_id"`
	OptionID int64  `json:"option_id"`
	Score    int64  `json:"score"`
}

type VoteEachItem struct {
	Options string `json:"options"`
	Score   int64  `json:"score"`
}

type SendMessage struct {
	Touser                 string             `json:"touser"`
	MsgType                string             `json:"msgtype"`
	AgentID                string             `json:"agentid"`
	EnableDuplicateCheck   string             `json:"enable_duplicate_check"`
	DuplicateCheckInterval string             `json:"duplicate_check_interval"`
	Text                   SendMessageContent `json:"text"`
}

type SendMessageContent struct {
	Content string `json:"content"`
}

type QuestionEachItem struct {
	QID      int64  `json:"qid"`
	Question string `json:"question"`
}

type ActivitySystemQuestion struct {
	ID       int64      `json:"id" gorm:"id"`
	AID      int64      `json:"aid" gorm:"column:aid"`
	QID      int64      `json:"qid" gorm:"column:qid"`
	Question string     `json:"question" gorm:"column:question"`
	UID      string     `json:"uid" gorm:"column:uid"`
	State    int64      `json:"state" gorm:"column:state"`
	Ctime    xtime.Time `json:"ctime" gorm:"column:ctime"`
}

type ActivitySystemQuestionList struct {
	ActivitySystemQuestion
	IsSelf int64 `json:"is_self"`
}

type SystemQuestionConfig struct {
	FilterSwitch int64 `json:"filter_switch"`
}

type ActivitySystemQuestionExport struct {
	Question       string `json:"question"`
	NickName       string `json:"nick_name"`
	UserName       string `json:"user_name"`
	DepartmentName string `json:"department_name"`
	State          string `json:"state"`
	Ctime          string `json:"ctime"`
}
