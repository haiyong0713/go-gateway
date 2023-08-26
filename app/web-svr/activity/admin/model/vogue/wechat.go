package vogue

type WeChatResp struct {
	Status int    `json:"status"`
	Msg    string `json:"msg"`
}

type WeChatBlockStatusResp struct {
	Blocked bool `json:"blocked"`
}

type WeChatCheckReq struct {
	Refresh int `form:"refresh" default:"0"`
}
