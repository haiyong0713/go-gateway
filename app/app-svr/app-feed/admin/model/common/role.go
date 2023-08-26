package common

import (
	"go-gateway/app/app-svr/app-feed/admin/model/manager"
)

const (
	AuthWebRcmdAdmin = "WEB_RCMD_ADMIN"
	AuthAppRcmdAdmin = "POS_REC_PLUS_CHECK"
	//普通用户
	RoleOrdinary = 2
	//业务负责人
	RoleAdmin1 = 3
	//业务二级负责人
	RoleAdmin2 = 4
	//角色
	RoleType = 1
)

// Role
type Role struct {
	Role      int                      `json:"role"`
	Group     []*manager.PosRecUserMgt `json:"group"`
	RoleGroup []*manager.PosRecUserMgt `json:"role_group"`
}
