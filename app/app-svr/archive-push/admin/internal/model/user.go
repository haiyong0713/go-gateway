package model

type GetOpenIDByMIDResp struct {
	*BaseHTTPResp
	Data *GetOpenIDByMIDRespData `json:"data"`
}

type GetOpenIDByMIDRespData struct {
	MID    int64  `json:"mid"`
	AppKey string `json:"appkey"`
	OpenID string `json:"openid"`
}

type GetMIDByUIDResp struct {
	*BaseHTTPResp
	Data *GetMIDByUIDRespData `json:"data"`
}

type GetMIDByUIDRespData struct {
	UID      string `json:"uid"`
	MID      int64  `json:"mid"`
	AppID    int64  `json:"appid"`
	Business string `json:"business"`
}

type GetMIDByOpenIDResp struct {
	*BaseHTTPResp
	Data *GetMIDByOpenIDRespData `json:"data"`
}

type GetMIDByOpenIDRespData struct {
	MID    int64  `json:"mid"`
	AppKey string `json:"appkey"`
	OpenID string `json:"openid"`
}

type GenerateAuthorizeCodeResp struct {
	*BaseHTTPResp
	Data *GetMIDByOpenIDRespData `json:"data"`
}

type GenerateAuthorizeCodeRespData struct {
	Code string `json:"code"`
}
