package manager

import (
	"go-gateway/app/app-svr/fawkes/service/model"
)

const (
	Event_Publish_Config              = 1
	Event_Publish_FF                  = 2
	Event_Publish_Event_Kibana_Query  = 3
	Event_Publish_Event_Field_Publish = 4
)

// role const
const (
	RoleAdmin   = 1
	RoleDev     = 2
	RoleTest    = 3
	RoleDevops  = 4
	RoleVisitor = 5

	RoleApplyPass   = 1
	RoleApplyRefuse = 2
	RoleApplyIgnore = 3
)

// empty state
const (
	AuthRoleApplyEmptyState = -99
)

// ParamsToken struct.
type ParamsToken struct {
	User     string `json:"user_name"`
	Platform string `json:"platform_id"`
}

// ResultToken struct.
type ResultToken struct {
	Token   string `json:"token"`
	User    string `json:"user_name"`
	Secret  string `json:"secret"`
	Expired int64  `json:"expired"`
}

// ResultRole struct.
type ResultRole struct {
	User    string `json:"user_name"`
	Role    int    `json:"role"`
	OldRole int    `json:"old_role,omitempty"`
	SRE     bool   `json:"rd_sre,omitempty"`
	Root    bool   `json:"rd_root,omitempty"`
	Leader  int    `json:"leader,omitempty"`
}

// ResultTree struct.
type ResultTree struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Alias string `json:"alias"`
	Type  int8   `json:"type"`
	Path  string `json:"path"`
	Tag   *struct {
		ControlCMD     string `json:"control_cmd,omitempty"`
		DeploymentPath string `json:"deployment_path,omitempty"`
		Domain         string `json:"domain,omitempty"`
		Level          int8   `json:"level,omitempty"`
		OPS            string `json:"ops,omitempty"`
		Project        string `json:"project,omitempty"`
		RDS            string `json:"rds,omitempty"`
		SingleNode     bool   `json:"single_node,omitempty"`
		SingleNodeACK  bool   `json:"single_node_ack,omitempty"`
	} `json:"tags,omitempty"`
	Children map[string]*ResultTree `json:"children,omitempty"`
}

// UserList struct
type UserList struct {
	PageInfo *model.PageInfo `json:"page"`
	Items    []*User         `json:"items"`
}

// User struct
type User struct {
	ID       int64  `json:"id"`
	AppKey   string `json:"app_key"`
	Name     string `json:"user_name"`
	NickName string `json:"nick_name"`
	Role     int    `json:"role"`
	Operator string `json:"operator"`
	MTime    int64  `json:"mtime"`
}

// Role struct
type Role struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	EName string `json:"ename"`
	Value int    `json:"value"`
	State int    `json:"state"`
	MTime int64  `json:"mtime"`
}

// Supervisor Role struct
type SupervisorRole struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Role     int    `json:"role"`
	Operator string `json:"operator"`
	MTime    int64  `json:"mtime"`
}

// FawkesUser struct
type FawkesUser struct {
	UserName        string            `json:"username"`
	Avatar          string            `json:"avatar"`
	NickName        string            `json:"nick_name"`
	SupervisorRoles []*SupervisorRole `json:"roles"`
	FawkesRoles     []*ResultRole     `json:"f_roles"`
}

// RoleApply struct
type RoleApply struct {
	ID       int64  `json:"id"`
	AppKey   string `json:"app_key"`
	Name     string `json:"name"`
	Operator string `json:"operator"`
	Role     int    `json:"role"`
	CurRole  int    `json:"cur_role"`
	State    int    `json:"state"`
	CTime    int64  `json:"ctime"`
	MTime    int64  `json:"mtime"`
}

// RoleApplyList struct
type RoleApplyList struct {
	PageInfo *model.PageInfo `json:"page"`
	Items    []*RoleApply    `json:"items"`
}

// UserInfo struct
type UserInfo struct {
	ID       int64  `json:"id"`
	UserId   string `json:"user_id"`
	Name     string `json:"user_name"`
	NickName string `json:"nick_name"`
	Avatar   string `json:"avatar"`
	MTime    int64  `json:"mtime"`
}

// Event Apply struct
type EventApply struct {
	ID        int64  `json:"id"`
	AppKey    string `json:"app_key"`
	Event     int8   `json:"event"`
	TargetID  string `json:"target_id"`
	Applicant string `json:"applicant"`
	Operator  string `json:"operator"`
	State     int    `json:"state"`
	CTime     int64  `json:"ctime"`
	MTime     int64  `json:"mtime"`
}

type BFSCDNRefreshRes struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
