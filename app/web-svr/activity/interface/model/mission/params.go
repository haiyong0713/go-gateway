package mission

type CommonOriginRequest struct {
	AppKey    string `json:"app_key" form:"app_key"`
	Timestamp int64  `json:"timestamp" form:"timestamp"`
	Version   string `json:"version" form:"version"`
	RequestId string `json:"request_id" form:"request_id"`
	Sign      string `json:"sign" form:"sign"`
	Params    string `json:"params" form:"params"`
}

type TencentAwardCallBackInnerParams struct {
	UserId    string `json:"user_id"`
	TaskId    string `json:"task_id"`
	SerialNum string `json:"serial_num"`
}
