package model

import "encoding/xml"

type PushScheduleMessage struct {
	MsgType string `json:"msgtype"`
	Text    struct {
		Content     string   `json:"content"`
		MentionList []string `json:"mentioned_list"`
	} `json:"text"`
}

type BotVerifyReq struct {
	MsgSignature string `form:"msg_signature"`
	Timestamp    string `form:"timestamp"`
	Nonce        string `form:"nonce"`
	Echostr      string `form:"echostr"`
}

type BotCallbackReq struct {
	MsgSignature string `form:"msg_signature"`
	Timestamp    string `form:"timestamp"`
	Nonce        string `form:"nonce"`
}

type MsgContent struct {
	ToUsername   string `xml:"ToUserName"`
	FromUsername string `xml:"FromUserName"`
	CreateTime   uint32 `xml:"CreateTime"`
	MsgType      string `xml:"MsgType"`
	Event        string `xml:"Event"`
	EventKey     string `xml:"EventKey"`
	Content      string `xml:"Content"`
	Msgid        string `xml:"MsgId"`
	Agentid      uint32 `xml:"AgentId"`
}

type WXRepTextMsg struct {
	ToUserName   string
	FromUserName string
	CreateTime   int64
	MsgType      string
	Content      string
	// 若不标记XMLName, 则解析后的xml名为该结构体的名称
	XMLName xml.Name `xml:"xml"`
}

type GetKernelReply struct {
	Code int64 `json:"code"`
	Data struct {
		Result     []*GetKernelResult `json:"result"`
		ResultType string             `json:"resultType"`
	} `json:"data"`
	Message string `json:"message"`
	TTL     int64  `json:"ttl"`
}

type GetKernelResult struct {
	Metric struct {
		Host                     string `json:"__host__"`
		Name                     string `json:"__name__"`
		Cluster                  string `json:"cluster"`
		ContainerEnvAppID        string `json:"container_env_app_id"`
		ContainerEnvDeployEnv    string `json:"container_env_deploy_env"`
		ContainerEnvPodContainer string `json:"container_env_pod_container"`
		ContainerEnvPodName      string `json:"container_env_pod_name"`
		InstanceName             string `json:"instance_name"`
		Job                      string `json:"job"`
	} `json:"metric"`
	Value  []interface{} `json:"value"`
	Values interface{}   `json:"values"`
}

type GetHTTPQPSReply struct {
	Code int64 `json:"code"`
	Data struct {
		Result     []*GetHTTPQPSResult `json:"result"`
		ResultType string              `json:"resultType"`
	} `json:"data"`
	Message string `json:"message"`
	TTL     int64  `json:"ttl"`
}

type GetHTTPQPSResult struct {
	Metric struct {
		Path string `json:"path"`
	} `json:"metric"`
	Value  []interface{} `json:"value"`
	Values interface{}   `json:"values"`
}

type GetGRPCQPSReply struct {
	Code int64 `json:"code"`
	Data struct {
		Result     []*GetGRPCQPSResult `json:"result"`
		ResultType string              `json:"resultType"`
	} `json:"data"`
	Message string `json:"message"`
	TTL     int64  `json:"ttl"`
}

type GetGRPCQPSResult struct {
	Metric struct {
		Method string `json:"method"`
	} `json:"metric"`
	Value  []interface{} `json:"value"`
	Values interface{}   `json:"values"`
}

type DashboardVerifyReply struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Username  string `json:"username"`
	SessionId string `json:"session_id"`
}
