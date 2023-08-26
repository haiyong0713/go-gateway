package ng

type ToNgDispatchReply struct {
	Response   []byte            `json:"response"`
	StatusCode int32             `json:"status_code"`
	Header     map[string]Header `json:"header"`
}

type Header struct {
	Values []string `json:"values"`
}
