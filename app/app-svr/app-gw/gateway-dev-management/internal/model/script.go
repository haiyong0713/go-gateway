package model

import "go-common/library/time"

type NewScriptReq struct {
	UserName string `form:"user_name"`
	Type     string `form:"type"`
	App      string `form:"app"`
	Zone     string `form:"zone"`
}

type Script struct {
	ID        string    `json:"id"`
	UserName  string    `json:"userName"`
	Type      string    `json:"type"`
	Parameter string    `json:"parameter"`
	CTime     time.Time `json:"ctime"`
	MTime     time.Time `json:"mtime"`
	APP       string    `json:"app"`
}

type GetScriptReply struct {
	UserID  string    `json:"id"`
	Scripts []*Script `json:"scripts"`
}

type GetTokenReply struct {
	Errcode     int64  `json:"errcode"`
	Errmsg      string `json:"errmsg"`
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}

type SendMessageReq struct {
	Touser  string `json:"touser"`
	Toparty string `json:"toparty"`
	Totag   string `json:"totag"`
	Msgtype string `json:"msgtype"`
	Agentid int64  `json:"agentid"`
	Text    struct {
		Content string `json:"content"`
	} `json:"text"`
	Safe                   int64 `json:"safe"`
	EnableIdTrans          int64 `json:"enable_id_trans"`
	EnableDuplicateCheck   int64 `json:"enable_duplicate_check"`
	DuplicateCheckInterval int64 `json:"duplicate_check_interval"`
}

type SendCardReq struct {
	Touser   string `json:"touser"`
	Toparty  string `json:"toparty"`
	Totag    string `json:"totag"`
	Msgtype  string `json:"msgtype"`
	Agentid  int64  `json:"agentid"`
	Textcard struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Url         string `json:"url"`
		Btntxt      string `json:"btntxt"`
	} `json:"textcard"`
	EnableIdTrans          int64 `json:"enable_id_trans"`
	EnableDuplicateCheck   int64 `json:"enable_duplicate_check"`
	DuplicateCheckInterval int64 `json:"duplicate_check_interval"`
}

type GetUserList struct {
	Errcode  int64  `json:"errcode"`
	Errmsg   string `json:"errmsg"`
	Userlist []struct {
		Userid      string  `json:"userid"`
		Name        string  `json:"name"`
		Department  []int64 `json:"department"`
		EnglishName string  `json:"english_name"`
	} `json:"userlist"`
}

type RestartParam struct {
	Zone string `json:"zone"`
}

type HRCoreAuthReply struct {
	Code int `json:"code"`
	Data struct {
		ExpiredInS int    `json:"expired_in(s)"`
		Token      string `json:"token"`
	} `json:"data"`
	Message string `json:"message"`
}

type HRCoreINFOReply struct {
	Code int `json:"code"`
	Data []struct {
		WxAccount string `json:"wxAccount"`
	} `json:"data"`
	Message string `json:"message"`
}
