package app

const (
	WXNotifyAccessTokenExpired = 42001

	WXNotifyType_Text = iota
	WXNotifyType_Markdown
	WXNotifyType_Image
)

type WXNotifyAccessTokenResp struct {
	ErrCode     int    `json:"errcode"`
	ErrMsg      string `json:"errmsg"`
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}

type WXNotifyUserListRes struct {
	ErrCode  int     `json:"errcode"`
	ErrMsg   string  `json:"errmsg"`
	UserList []*User `json:"userlist"`
}

type WXNotifyMessage struct {
	Touser                 string            `json:"touser"`
	Toparty                string            `json:"toparty"`
	Totag                  string            `json:"totag"`
	MsgType                string            `json:"msgtype"`
	Agentid                string            `json:"agentid"`
	Text                   *WXNotifyText     `json:"text,omitempty"`
	Markdown               *WXNotifyMarkdown `json:"markdown,omitempty"`
	Image                  *WXNotifyImage    `json:"image,omitempty"`
	Safe                   int64             `json:"safe"`
	IsDuplicateCheck       int64             `json:"enable_duplicate_check"`
	DuplicateCheckInterval int64             `json:"duplicate_check_interval"`
}

type WXNotifyText struct {
	Content string `json:"content"`
}

type WXNotifyMarkdown struct {
	Content string `json:"content"`
}

type WXNotifyImage struct {
	MediaId string `json:"media_id"`
}

type User struct {
	UserID   string `json:"userid"`
	Name     string `json:"name"`
	Position string `json:"position"`
	Mobile   string `json:"mobile"`
	Email    string `json:"email"`
	Alias    string `json:"alias"`
	Avatar   string `json:"avatar"`
	CTime    int64  `json:"ctime"`
	MTime    int64  `json:"mtime"`
}

type WXNotifyTmpFileResp struct {
	ErrCode   int64  `json:"errcode"`
	ErrMsg    string `json:"errmsg"`
	Type      string `json:"type"`
	MediaId   string `json:"media_id"`
	CreatedAt string `json:"created_at"`
}

type WXNotifyMsgResp struct {
	ErrCode int64  `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
	MsgId   string `json:"msgid"`
}
