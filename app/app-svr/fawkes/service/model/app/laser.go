package app

// const for laser.
const (
	StatusUpFaild   = -2
	StatusSendFaild = -1
	StatusQueuing   = 1
	StatusWaitSend  = 2
	StatusUpSuccess = 3
)

// 1排队中；2待上传；3业务成功；4业务到达；-1发送失败；-2业务失败 ；-3 客户端未支持
const (
	LaserCmdStatusQueue          = 1
	LaserCmdStatusWaiting        = 2
	LaserCmdStatusSuccess        = 3
	LaserCmdStatusReceiveSuccess = 4
	LaserCmdStatusUnsupport      = -3
	LaserCmdStatusFail           = -2
	LaserCmdStatusSendFail       = -1
)

const (
	ChannelFawkes      = 0
	ChannelBusinessApi = 1
)

// Laser struct.
type Laser struct {
	ID            int64  `json:"task_id"`
	AppKey        string `json:"app_key"`
	Platform      string `json:"platform"`
	MID           int64  `json:"mid"`
	Buvid         string `json:"buvid"`
	Email         string `json:"email"`
	LogDate       string `json:"log_date"`
	URL           string `json:"url"`
	Status        int8   `json:"status"`
	Operator      string `json:"operator"`
	CTime         int64  `json:"ctime"`
	MTime         int64  `json:"mtime"`
	SilenceURL    string `json:"silence_url"`
	SilenceStatus int8   `json:"silence_status"`
	ParseStatus   int    `json:"parse_status"`
	Channel       int8   `json:"channel"`
	Description   string `json:"description"`
	MobiApp       string `json:"mobi_app"`
	RecallMobiApp string `json:"recall_mobi_app"`
	Build         string `json:"build"`
	ErrorMessage  string `json:"error_msg"`
	MsgId         string `json:"msg_id"`
	MD5           string `json:"md5"`
}

// LaserCmd struct
type LaserCmd struct {
	ID            int64  `json:"task_id"`
	AppKey        string `json:"app_key"`
	Platform      string `json:"platform"`
	MID           int64  `json:"mid"`
	Buvid         string `json:"buvid"`
	Action        string `json:"action"`
	Params        string `json:"params"`
	URL           string `json:"url"`
	Result        string `json:"result"`
	Status        int8   `json:"status"`
	Operator      string `json:"operator"`
	Description   string `json:"description"`
	MobiApp       string `json:"mobi_app"`
	RecallMobiApp string `json:"recall_mobi_app"`
	Build         string `json:"build"`
	ErrorMessage  string `json:"error_msg"`
	CTime         int64  `json:"ctime"`
	MTime         int64  `json:"mtime"`
}

// LaserCmdAction struct
type LaserCmdAction struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Platform    string `json:"platform"`
	Params      string `json:"params"`
	Operator    string `json:"operator"`
	Description string `json:"description"`
	Mtime       int64  `json:"mtime"`
	Ctime       int64  `json:"ctime"`
}

// type BroadcastConsumer struct
type BroadcastConsumer struct {
	Type      string `json:"type"`
	MID       int64  `json:"mid"`
	Buvid     string `json:"buvid"`
	MobiApp   string `json:"mobi_app"`
	Platform  string `json:"platform"`
	Build     int64  `json:"build"`
	Timestamp int64  `json:"timestamp"`
}
