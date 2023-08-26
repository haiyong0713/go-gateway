package broadcast

type ProxyResp struct {
	ReqHeader   interface{} `json:"请求Header,omitempty"`
	ReqBody     interface{} `json:"请求数据,omitempty"`
	ReqPath     string      `json:"请求方法,omitempty"`
	ReqServer   string      `json:"请求目标,omitempty"`
	ReqServerIp string      `json:"请求目标IP,omitempty"`
	ReqCluster  string      `json:"请求集群,omitempty"`
	Response    interface{} `json:"返回数据,omitempty"`
	Err         string      `json:"错误,omitempty"`
	Des         string      `json:"额外说明,omitempty"`
}

const (
	PushBuvids = "/push.service.broadcast.v2.BroadcastAPI/PushBuvids"
	PushMids   = "/push.service.broadcast.v2.BroadcastAPI/PushMids"
	PushAll    = "/push.service.broadcast.v2.BroadcastAPI/PushAll"
)
