package audit

import (
	"time"

	"go-common/library/conf/env"
	"go-common/library/log"
	"go-common/library/queue/databus/report"
	pb "go-gateway/app/app-svr/app-gw/management/api"

	xuuid "github.com/google/uuid"
)

const (
	LogAuditBusiness             = 590
	LogTypeApplicationProxyTOML  = 1
	LogTypeProxyConfig           = 2
	LogTypeAsyncTask             = 3
	LogTypeDeployment            = 4
	LogTypeLimiter               = 5
	LogMngSponsor                = "management-job"
	LogLevelInfo                 = "INFO"
	LogLevelWarn                 = "WARN"
	LogLevelError                = "ERROR"
	LogActionPush                = "push"
	LogActionAdd                 = "add"
	LogActionUpdate              = "update"
	LogActionDel                 = "delete"
	LogActionEnable              = "enable"
	LogActionDisable             = "disable"
	LogActionTrigger             = "trigger"
	LogActionCreate              = "create"
	LogActionFinish              = "finish"
	LogActionClose               = "close"
	LogResultSuccess             = "success"
	LogResultFailure             = "failure"
	LogCategoryTask              = "task"
	LogResultNone                = "none"
	LogProxyConfigIdentifier     = "proxy-config.toml"
	LogGRPCProxyConfigIdentifier = "grpc-proxy-config.toml"
)

func uuid() string {
	return xuuid.New().String()
}

func formatTime(t int64) int64 {
	if t <= 0 {
		return time.Now().Unix()
	}
	return t
}

type ReportParam struct {
	GatewayGroup string `json:"gateway_group"`
	GatewayName  string `json:"gateway_name"`
	Object       int    `json:"object"`
	Ctime        int64  `json:"ctime"`
	Mtime        int64  `json:"mtime"`
	Level        string `json:"level"`
	Sponsor      string `json:"sponsor"`
	Action       string `json:"action"`
	Result       string `json:"result"`
	Detail       string `json:"detail"`
	Identifier   string `json:"identifier"`
	Env          string `json:"env"`
	Zone         string `json:"zone"`
}

func SendBreakApiLog(req *pb.SetBreakerAPIReq, action, level, result, detail string, ctime, mtime int64) {
	logData := &report.ManagerInfo{
		Uname:    req.Username,
		UID:      0,
		Business: LogAuditBusiness,
		Type:     LogTypeProxyConfig,
		Oid:      0,
		Action:   action,
		Ctime:    time.Now(),
		Index:    []interface{}{req.Node, req.Gateway, req.Api, uuid(), env.DeployEnv, env.Zone, "", "", "", 0, 0, 0, 0, 0, 0, 0, 0, 0},
		Content: map[string]interface{}{
			"ctime":    formatTime(ctime),
			"mtime":    formatTime(mtime),
			"level":    level,
			"result":   result,
			"detail":   detail,
			"category": LogCategoryTask,
		},
	}
	if err := report.Manager(logData); err != nil {
		log.Error("Failed to send task log(%+v)", err)
	}
}

func SendEnableBreakApiLog(req *pb.EnableBreakerAPIReq, action, level, result, detail string, ctime, mtime int64) {
	logData := &report.ManagerInfo{
		Uname:    req.Username,
		UID:      0,
		Business: LogAuditBusiness,
		Type:     LogTypeProxyConfig,
		Oid:      0,
		Action:   action,
		Ctime:    time.Now(),
		Index:    []interface{}{req.Node, req.Gateway, req.Api, uuid(), env.DeployEnv, env.Zone, "", "", "", 0, 0, 0, 0, 0, 0, 0, 0, 0},
		Content: map[string]interface{}{
			"ctime":    formatTime(ctime),
			"mtime":    formatTime(mtime),
			"level":    level,
			"result":   result,
			"detail":   detail,
			"category": LogCategoryTask,
		},
	}
	if err := report.Manager(logData); err != nil {
		log.Error("Failed to send task log(%+v)", err)
	}
}

func SendDeleteBreakerAPILog(req *pb.DeleteBreakerAPIReq, action, level, result, detail string, ctime, mtime int64) {
	logData := &report.ManagerInfo{
		Uname:    req.Username,
		UID:      0,
		Business: LogAuditBusiness,
		Type:     LogTypeProxyConfig,
		Oid:      0,
		Action:   action,
		Ctime:    time.Now(),
		Index:    []interface{}{req.Node, req.Gateway, req.Api, uuid(), env.DeployEnv, env.Zone, "", "", "", 0, 0, 0, 0, 0, 0, 0, 0, 0},
		Content: map[string]interface{}{
			"ctime":    formatTime(ctime),
			"mtime":    formatTime(mtime),
			"level":    level,
			"result":   result,
			"detail":   detail,
			"category": LogCategoryTask,
		},
	}
	if err := report.Manager(logData); err != nil {
		log.Error("Failed to send task log(%+v)", err)
	}
}

func SendSetDynPathLog(req *pb.SetDynPathReq, action, level, result, detail string, ctime, mtime int64) {
	logData := &report.ManagerInfo{
		Uname:    req.Username,
		UID:      0,
		Business: LogAuditBusiness,
		Type:     LogTypeProxyConfig,
		Oid:      0,
		Action:   action,
		Ctime:    time.Now(),
		Index:    []interface{}{req.Node, req.Gateway, req.Pattern, uuid(), env.DeployEnv, env.Zone, "", "", "", 0, 0, 0, 0, 0, 0, 0, 0, 0},
		Content: map[string]interface{}{
			"ctime":    formatTime(ctime),
			"mtime":    formatTime(mtime),
			"level":    level,
			"result":   result,
			"detail":   detail,
			"category": LogCategoryTask,
		},
	}
	if err := report.Manager(logData); err != nil {
		log.Error("Failed to send task log(%+v)", err)
	}
}

func SendEnableDynPathLog(req *pb.EnableDynPathReq, action, level, result, detail string, ctime, mtime int64) {
	logData := &report.ManagerInfo{
		Uname:    req.Username,
		UID:      0,
		Business: LogAuditBusiness,
		Type:     LogTypeProxyConfig,
		Oid:      0,
		Action:   action,
		Ctime:    time.Now(),
		Index:    []interface{}{req.Node, req.Gateway, req.Pattern, uuid(), env.DeployEnv, env.Zone, "", "", "", 0, 0, 0, 0, 0, 0, 0, 0, 0},
		Content: map[string]interface{}{
			"ctime":    formatTime(ctime),
			"mtime":    formatTime(mtime),
			"level":    level,
			"result":   result,
			"detail":   detail,
			"category": LogCategoryTask,
		},
	}
	if err := report.Manager(logData); err != nil {
		log.Error("Failed to send task log(%+v)", err)
	}
}

func SendDeleteDynPathLog(req *pb.DeleteDynPathReq, action, level, result, detail string, ctime, mtime int64) {
	logData := &report.ManagerInfo{
		Uname:    req.Username,
		UID:      0,
		Business: LogAuditBusiness,
		Type:     LogTypeProxyConfig,
		Oid:      0,
		Action:   action,
		Ctime:    time.Now(),
		Index:    []interface{}{req.Node, req.Gateway, req.Pattern, uuid(), env.DeployEnv, env.Zone, "", "", "", 0, 0, 0, 0, 0, 0, 0, 0, 0},
		Content: map[string]interface{}{
			"ctime":    formatTime(ctime),
			"mtime":    formatTime(mtime),
			"level":    level,
			"result":   result,
			"detail":   detail,
			"category": LogCategoryTask,
		},
	}
	if err := report.Manager(logData); err != nil {
		log.Error("Failed to send task log(%+v)", err)
	}
}

func SendSetGatewayLog(req *pb.SetGatewayReq, action, level, result, detail string, ctime, mtime int64) {
	logData := &report.ManagerInfo{
		Uname:    req.Username,
		UID:      0,
		Business: LogAuditBusiness,
		Type:     LogTypeProxyConfig,
		Oid:      0,
		Action:   action,
		Ctime:    time.Now(),
		Index:    []interface{}{req.Node, req.AppName, req.AppName, uuid(), env.DeployEnv, env.Zone, "", "", "", 0, 0, 0, 0, 0, 0, 0, 0, 0},
		Content: map[string]interface{}{
			"ctime":    formatTime(ctime),
			"mtime":    formatTime(mtime),
			"level":    level,
			"result":   result,
			"detail":   detail,
			"category": LogCategoryTask,
		},
	}
	if err := report.Manager(logData); err != nil {
		log.Error("Failed to send task log(%+v)", err)
	}
}

func SendDeleteGatewayLog(req *pb.DeleteGatewayReq, action, level, result, detail string, ctime, mtime int64) {
	logData := &report.ManagerInfo{
		Uname:    req.Username,
		UID:      0,
		Business: LogAuditBusiness,
		Type:     LogTypeProxyConfig,
		Oid:      0,
		Action:   action,
		Ctime:    time.Now(),
		Index:    []interface{}{req.Node, req.AppName, req.AppName, uuid(), env.DeployEnv, env.Zone, "", "", "", 0, 0, 0, 0, 0, 0, 0, 0, 0},
		Content: map[string]interface{}{
			"ctime":    formatTime(ctime),
			"mtime":    formatTime(mtime),
			"level":    level,
			"result":   result,
			"detail":   detail,
			"category": LogCategoryTask,
		},
	}
	if err := report.Manager(logData); err != nil {
		log.Error("Failed to send task log(%+v)", err)
	}
}

func SendUpdateALLGatewayConfigLog(req *pb.UpdateALLGatewayConfigReq, action, level, result, detail string, ctime, mtime int64) {
	logData := &report.ManagerInfo{
		Uname:    req.Username,
		UID:      0,
		Business: LogAuditBusiness,
		Type:     LogTypeProxyConfig,
		Oid:      0,
		Action:   action,
		Ctime:    time.Now(),
		Index:    []interface{}{req.Node, req.AppName, req.AppName, uuid(), env.DeployEnv, env.Zone, "", "", "", 0, 0, 0, 0, 0, 0, 0, 0, 0},
		Content: map[string]interface{}{
			"ctime":    formatTime(ctime),
			"mtime":    formatTime(mtime),
			"level":    level,
			"result":   result,
			"detail":   detail,
			"category": LogCategoryTask,
		},
	}
	if err := report.Manager(logData); err != nil {
		log.Error("Failed to send task log(%+v)", err)
	}
}

func SendTaskLog(rp *ReportParam) {
	logData := &report.ManagerInfo{
		Uname:    rp.Sponsor,
		UID:      0,
		Business: LogAuditBusiness,
		Type:     rp.Object,
		Oid:      0,
		Action:   rp.Action,
		Ctime:    time.Now(),
		Index:    []interface{}{rp.GatewayGroup, rp.GatewayName, rp.Identifier, uuid(), env.DeployEnv, env.Zone, "", "", "", 0, 0, 0, 0, 0, 0, 0, 0, 0},
		Content: map[string]interface{}{
			"ctime":    formatTime(rp.Ctime),
			"mtime":    formatTime(rp.Mtime),
			"level":    rp.Level,
			"result":   rp.Result,
			"detail":   rp.Detail,
			"category": LogCategoryTask,
		},
	}
	if err := report.Manager(logData); err != nil {
		log.Error("Failed to send task log(%+v)", err)
	}
}

func SendTriggerTaskExecuteLog(req *pb.ExecuteTaskReq, level, result, detail string, ctime, mtime int64) {
	logData := &report.ManagerInfo{
		Uname:    req.Username,
		UID:      0,
		Business: LogAuditBusiness,
		Type:     LogTypeAsyncTask,
		Oid:      0,
		Action:   LogActionTrigger,
		Ctime:    time.Now(),
		Index:    []interface{}{req.Node, req.Gateway, req.Gateway, uuid(), env.DeployEnv, env.Zone, "", "", "", 0, 0, 0, 0, 0, 0, 0, 0, 0},
		Content: map[string]interface{}{
			"ctime":    formatTime(ctime),
			"mtime":    formatTime(mtime),
			"level":    level,
			"result":   result,
			"detail":   detail,
			"category": LogCategoryTask,
		},
	}
	if err := report.Manager(logData); err != nil {
		log.Error("Failed to send task log(%+v)", err)
	}
}

func SendDeploymentLog(req *pb.DeploymentMeta, action, level, result, detail string) {
	logData := &report.ManagerInfo{
		Uname:    req.Sponsor,
		UID:      0,
		Business: LogAuditBusiness,
		Type:     LogTypeDeployment,
		Oid:      0,
		Action:   action,
		Ctime:    time.Now(),
		Index:    []interface{}{req.Node, req.Gateway, req.DeploymentId, uuid(), env.DeployEnv, env.Zone, "", "", "", 0, 0, 0, 0, 0, 0, 0, 0, 0},
		Content: map[string]interface{}{
			"ctime":    formatTime(req.CreatedAt),
			"mtime":    formatTime(req.UpdatedAt),
			"level":    level,
			"result":   result,
			"detail":   detail,
			"category": LogCategoryTask,
		},
	}
	if err := report.Manager(logData); err != nil {
		log.Error("Failed to send task log(%+v)", err)
	}
}

func SendAddLimiterLog(req *pb.AddLimiterReq, action, level, result, detail string, ctime, mtime int64) {
	logData := &report.ManagerInfo{
		Uname:    req.Username,
		UID:      0,
		Business: LogAuditBusiness,
		Type:     LogTypeLimiter,
		Oid:      0,
		Action:   action,
		Ctime:    time.Now(),
		Index:    []interface{}{req.Node, req.Gateway, req.Api, uuid(), env.DeployEnv, env.Zone, "", "", "", 0, 0, 0, 0, 0, 0, 0, 0, 0},
		Content: map[string]interface{}{
			"ctime":    formatTime(ctime),
			"mtime":    formatTime(mtime),
			"level":    level,
			"result":   result,
			"detail":   detail,
			"category": LogCategoryTask,
		},
	}
	if err := report.Manager(logData); err != nil {
		log.Error("Failed to send task log(%+v)", err)
	}
}

func SendSetLimiterLog(req *pb.SetLimiterReq, action, level, result, detail string, ctime, mtime int64) {
	logData := &report.ManagerInfo{
		Uname:    req.Username,
		UID:      0,
		Business: LogAuditBusiness,
		Type:     LogTypeLimiter,
		Oid:      0,
		Action:   action,
		Ctime:    time.Now(),
		Index:    []interface{}{req.Node, req.Gateway, req.Id, uuid(), env.DeployEnv, env.Zone, "", "", "", 0, 0, 0, 0, 0, 0, 0, 0, 0},
		Content: map[string]interface{}{
			"ctime":    formatTime(ctime),
			"mtime":    formatTime(mtime),
			"level":    level,
			"result":   result,
			"detail":   detail,
			"category": LogCategoryTask,
		},
	}
	if err := report.Manager(logData); err != nil {
		log.Error("Failed to send task log(%+v)", err)
	}
}

func SendEnableLimiterLog(req *pb.EnableLimiterReq, action, level, result, detail string, ctime, mtime int64) {
	logData := &report.ManagerInfo{
		Uname:    req.Username,
		UID:      0,
		Business: LogAuditBusiness,
		Type:     LogTypeLimiter,
		Oid:      0,
		Action:   action,
		Ctime:    time.Now(),
		Index:    []interface{}{req.Node, req.Gateway, req.Api, uuid(), env.DeployEnv, env.Zone, "", "", "", 0, 0, 0, 0, 0, 0, 0, 0, 0},
		Content: map[string]interface{}{
			"ctime":    formatTime(ctime),
			"mtime":    formatTime(mtime),
			"level":    level,
			"result":   result,
			"detail":   detail,
			"category": LogCategoryTask,
		},
	}
	if err := report.Manager(logData); err != nil {
		log.Error("Failed to send task log(%+v)", err)
	}
}

func SendDeleteLimiterLog(req *pb.DeleteLimiterReq, action, level, result, detail string, ctime, mtime int64) {
	logData := &report.ManagerInfo{
		Uname:    req.Username,
		UID:      0,
		Business: LogAuditBusiness,
		Type:     LogTypeLimiter,
		Oid:      0,
		Action:   action,
		Ctime:    time.Now(),
		Index:    []interface{}{req.Node, req.Gateway, req.Id, uuid(), env.DeployEnv, env.Zone, "", "", "", 0, 0, 0, 0, 0, 0, 0, 0, 0},
		Content: map[string]interface{}{
			"ctime":    formatTime(ctime),
			"mtime":    formatTime(mtime),
			"level":    level,
			"result":   result,
			"detail":   detail,
			"category": LogCategoryTask,
		},
	}
	if err := report.Manager(logData); err != nil {
		log.Error("Failed to send task log(%+v)", err)
	}
}
