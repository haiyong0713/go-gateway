package app

// RobotReq robot request struct
type RobotReq struct {
	MsgType  string    `json:"msgtype"`
	Text     *Text     `json:"text,omitempty"`
	Markdown *Markdown `json:"markdown,omitempty"`
	Image    *Image    `json:"image,omitempty"`
	News     *News     `json:"news,omitempty"`
}

// RobotRes robot
type RobotRes struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

type RobotUploadRes struct {
	FileURL string `json:"url"`
}

// Text struct
type Text struct {
	Content             string   `json:"content"`
	MentionedList       []string `json:"mentioned_list,omitempty"`
	MentionedMobileList []string `json:"mentioned_mobile_list,omitempty"`
}

// Markdown struct
type Markdown struct {
	Content string `json:"content"`
}

// Image struct
type Image struct {
	Base64 string `json:"base64"`
	Md5    string `json:"md5"`
}

// News struct
type News struct {
	Articles []*Article `json:"articles"`
}

// Article struct
type Article struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url"`
	PicURL      string `json:"picurl"`
}

type RobotAddStatus struct {
	Msg  string `json:"msg"`
	Code int64  `json:"code"`
}

const (
	CINotifyGroupBot       = "CI_NOTIFY_GROUP"
	CDReleaseNotifyBot     = "CD_RELEASE_NOTIFY"
	HotfixFinishJobBot     = "HOTFIX_FINISH_JOB"
	MessageBot             = "MESSAGE"
	FeedBackBot            = "FEEDBACK"
	FeedBackDefaultPushBot = "FEEDBACK_DEFAULT_PUSH"

	RobotOpen  = 1
	RobotClose = 0

	RobotGlobal    = 1
	RobotNotGlobal = 0

	RobotDefault    = 1
	RobotNotDefault = 0
)
