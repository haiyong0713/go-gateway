package app

// app const.
const (
	AuditLogoff = -3
	AuditRefuse = -2
	AuditDel    = -1
	AuditWait   = 0
	AuditPass   = 1
)

const (
	AppServerZone_Inland = 0
	AppServerZone_Abroad = 1
)

// APP struct.
type APP struct {
	ID              int64  `json:"id"`
	AttrID          int64  `json:"attr_id"`
	DataCenterAppID int64  `json:"datacenter_app_id"`
	AppID           string `json:"app_id"`
	AppKey          string `json:"app_key"`
	MobiApp         string `json:"mobi_app"`
	Platform        string `json:"platform"`
	GitPrjID        string `json:"git_prj_id"`
	Name            string `json:"name"`
	Icon            string `json:"icon"`
	TreePath        string `json:"tree_path"`
	Desc            string `json:"description"`
	GitPath         string `json:"git_path"`
	Owners          string `json:"owners"`
	Operator        string `json:"operator"`
	State           int8   `json:"state"`
	Version         string `json:"version,omitempty"`
	VersionCode     int64  `json:"version_code,omitempty"`
	IsFollow        int8   `json:"is_follow"`
	Reason          string `json:"refusal_reason,omitempty"`
	RobotName       string `json:"robot_name"`
	RobotWebhookUrl string `json:"robot_webhook_url"`
	ManagerPlat     int    `json:"manager_plat"`
	DebugUrl        string `json:"debug_url"`
	AppDsymName     string `json:"app_dsym_name"`
	AppSymbolsoName string `json:"app_symbolso_name"`
	ServerZone      int64  `json:"server_zone"`
	WorkflowId      string `json:"workflow_id"`
	Role            int    `json:"role,omitempty"`
	CTime           int64  `json:"ctime"`
	PTime           int64  `json:"ptime"`
	UserAdmins      string `json:"user_admins"`     // 实际管理员
	LaserWebhook    string `json:"laser_webhook"`   // 主动拉取日志webhook
	IsHost          int8   `json:"is_host"`         // 是否是宿主App
	IsHighestPeak   int8   `json:"is_highest_peak"` // 是否处于业务高峰期
}

// ResMailList def.
type ResMailList struct {
	MailList []string `json:"mail_list"`
}

// Notif struct.
type Notif struct {
	ID         int64  `json:"id"`
	AppKeys    string `json:"app_keys"`
	Platform   string `json:"platform"`
	RoutePath  string `json:"route_path"`
	Title      string `json:"title"`
	Content    string `json:"content"`
	URL        string `json:"url"`
	Closeable  int    `json:"closeable"`
	State      int    `json:"state"`
	IsGlobal   int    `json:"is_global"`
	Type       int    `json:"type"`
	Operator   string `json:"operator"`
	EffectTime int64  `json:"effect_time"`
	ExpireTime int64  `json:"expire_time"`
	Mtime      int64  `json:"mtime"`
	Ctime      uint64 `json:"ctime"`
}

// Robot struct
type Robot struct {
	ID          int64  `json:"id"`
	BotName     string `json:"bot_name"`
	WebHook     string `json:"webhook,omitempty"`
	AppKeys     string `json:"app_keys"`
	FuncModule  string `json:"func_module"`
	Owner       string `json:"owner"`
	Users       string `json:"users"`
	State       int    `json:"state"`
	IsGlobal    int    `json:"is_global"`
	IsDefault   int    `json:"is_default"`
	Description string `json:"description"`
	Operator    string `json:"operator"`
	Mtime       int64  `json:"mtime"`
	Ctime       int64  `json:"ctime"`
}
