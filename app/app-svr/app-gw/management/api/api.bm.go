// Code generated by protoc-gen-bm v0.1, DO NOT EDIT.
// source: go-gateway/app/app-svr/app-gw/management/api/api.proto

/*
Package api is a generated blademaster stub package.
This code was generated with kratos/tool/protobuf/protoc-gen-bm v0.1.

It is generated from these files:

	go-gateway/app/app-svr/app-gw/management/api/api.proto
*/
package api

import (
	"context"

	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/binding"
)
import google_protobuf1 "github.com/golang/protobuf/ptypes/empty"

// to suppressed 'imported but not used warning'
var _ *bm.Context
var _ context.Context
var _ binding.StructValidator

var PathManagementPing = "/appgw.management.v1.Management/Ping"
var PathManagementAuthZ = "/appgw.management.v1.Management/AuthZ"
var PathManagementGateway = "/appgw.management.v1.Management/Gateway"
var PathManagementGatewayProfile = "/appgw.management.v1.Management/GatewayProfile"
var PathManagementAddGateway = "/appgw.management.v1.Management/AddGateway"
var PathManagementUpdateGateway = "/appgw.management.v1.Management/UpdateGateway"
var PathManagementDeleteGateway = "/appgw.management.v1.Management/DeleteGateway"
var PathManagementEnableALLGatewayConfig = "/appgw.management.v1.Management/EnableALLGatewayConfig"
var PathManagementDisableALLGatewayConfig = "/appgw.management.v1.Management/DisableALLGatewayConfig"
var PathManagementEnableAllGRPCGatewayConfig = "/appgw.management.v1.Management/EnableAllGRPCGatewayConfig"
var PathManagementDisableAllGRPCGatewayConfig = "/appgw.management.v1.Management/DisableAllGRPCGatewayConfig"
var PathManagementListBreakerAPI = "/appgw.management.v1.Management/ListBreakerAPI"
var PathManagementSetBreakerAPI = "/appgw.management.v1.Management/SetBreakerAPI"
var PathManagementUpdateBreakerAPI = "/appgw.management.v1.Management/UpdateBreakerAPI"
var PathManagementEnableBreakerAPI = "/appgw.management.v1.Management/EnableBreakerAPI"
var PathManagementDisableBreakerAPI = "/appgw.management.v1.Management/DisableBreakerAPI"
var PathManagementDeleteBreakerAPI = "/appgw.management.v1.Management/DeleteBreakerAPI"
var PathManagementListDynPath = "/appgw.management.v1.Management/ListDynPath"
var PathManagementAddDynPath = "/appgw.management.v1.Management/AddDynPath"
var PathManagementUpdateDynPath = "/appgw.management.v1.Management/UpdateDynPath"
var PathManagementDeleteDynPath = "/appgw.management.v1.Management/DeleteDynPath"
var PathManagementEnableDynPath = "/appgw.management.v1.Management/EnableDynPath"
var PathManagementDisableDynPath = "/appgw.management.v1.Management/DisableDynPath"
var PathManagementListLog = "/appgw.management.v1.Management/ListLog"
var PathManagementExecuteTask = "/appgw.management.v1.Management/ExecuteTask"
var PathManagementGatewayProxy = "/appgw.management.v1.Management/GatewayProxy"
var PathManagementInitGatewayConfigs = "/appgw.management.v1.Management/InitGatewayConfigs"
var PathManagementAppPromptAPI = "/appgw.management.v1.Management/AppPromptAPI"
var PathManagementConfigPromptAPI = "/appgw.management.v1.Management/ConfigPromptAPI"
var PathManagementAppPathPromptAPI = "/appgw.management.v1.Management/AppPathPromptAPI"
var PathManagementSnapshotAction = "/appgw.management.v1.Management/SnapshotAction"
var PathManagementAddSnapshot = "/appgw.management.v1.Management/AddSnapshot"
var PathManagementSnapshotProfile = "/appgw.management.v1.Management/SnapshotProfile"
var PathManagementCreateDeployment = "/appgw.management.v1.Management/CreateDeployment"
var PathManagementCompareDeployment = "/appgw.management.v1.Management/CompareDeployment"
var PathManagementConfirmDeployment = "/appgw.management.v1.Management/ConfirmDeployment"
var PathManagementDeployDeployment = "/appgw.management.v1.Management/DeployDeployment"
var PathManagementDeployDeploymentProfile = "/appgw.management.v1.Management/DeployDeploymentProfile"
var PathManagementDeployment = "/appgw.management.v1.Management/Deployment"
var PathManagementDeployDeploymentAll = "/appgw.management.v1.Management/DeployDeploymentAll"
var PathManagementRollbackDeployment = "/appgw.management.v1.Management/RollbackDeployment"
var PathManagementFinishDeployment = "/appgw.management.v1.Management/FinishDeployment"
var PathManagementCloseDeployment = "/appgw.management.v1.Management/CloseDeployment"
var PathManagementCancelDeployment = "/appgw.management.v1.Management/CancelDeployment"
var PathManagementListLimiter = "/appgw.management.v1.Management/ListLimiter"
var PathManagementAddLimiter = "/appgw.management.v1.Management/AddLimiter"
var PathManagementUpdateLimiter = "/appgw.management.v1.Management/UpdateLimiter"
var PathManagementDeleteLimiter = "/appgw.management.v1.Management/DeleteLimiter"
var PathManagementEnableLimiter = "/appgw.management.v1.Management/EnableLimiter"
var PathManagementDisableLimiter = "/appgw.management.v1.Management/DisableLimiter"
var PathManagementSetupPlugin = "/appgw.management.v1.Management/SetupPlugin"
var PathManagementZonePromptAPI = "/appgw.management.v1.Management/ZonePromptAPI"
var PathManagementPluginList = "/appgw.management.v1.Management/PluginList"
var PathManagementGRPCAppMethodPromptAPI = "/appgw.management.v1.Management/GRPCAppMethodPromptAPI"
var PathManagementGRPCAppPackagePromptAPI = "/appgw.management.v1.Management/GRPCAppPackagePromptAPI"

// ManagementBMServer is the server API for Management service.
type ManagementBMServer interface {
	Ping(ctx context.Context, req *google_protobuf1.Empty) (resp *google_protobuf1.Empty, err error)

	AuthZ(ctx context.Context, req *AuthZReq) (resp *AuthZReply, err error)

	Gateway(ctx context.Context, req *AuthZReq) (resp *GatewayReply, err error)

	GatewayProfile(ctx context.Context, req *GatewayProfileReq) (resp *GatewayProfileReply, err error)

	AddGateway(ctx context.Context, req *SetGatewayReq) (resp *google_protobuf1.Empty, err error)

	UpdateGateway(ctx context.Context, req *SetGatewayReq) (resp *google_protobuf1.Empty, err error)

	DeleteGateway(ctx context.Context, req *DeleteGatewayReq) (resp *google_protobuf1.Empty, err error)

	EnableALLGatewayConfig(ctx context.Context, req *UpdateALLGatewayConfigReq) (resp *google_protobuf1.Empty, err error)

	DisableALLGatewayConfig(ctx context.Context, req *UpdateALLGatewayConfigReq) (resp *google_protobuf1.Empty, err error)

	EnableAllGRPCGatewayConfig(ctx context.Context, req *UpdateALLGatewayConfigReq) (resp *google_protobuf1.Empty, err error)

	DisableAllGRPCGatewayConfig(ctx context.Context, req *UpdateALLGatewayConfigReq) (resp *google_protobuf1.Empty, err error)

	ListBreakerAPI(ctx context.Context, req *ListBreakerAPIReq) (resp *ListBreakerAPIReply, err error)

	SetBreakerAPI(ctx context.Context, req *SetBreakerAPIReq) (resp *google_protobuf1.Empty, err error)

	UpdateBreakerAPI(ctx context.Context, req *SetBreakerAPIReq) (resp *google_protobuf1.Empty, err error)

	EnableBreakerAPI(ctx context.Context, req *EnableBreakerAPIReq) (resp *google_protobuf1.Empty, err error)

	DisableBreakerAPI(ctx context.Context, req *EnableBreakerAPIReq) (resp *google_protobuf1.Empty, err error)

	DeleteBreakerAPI(ctx context.Context, req *DeleteBreakerAPIReq) (resp *google_protobuf1.Empty, err error)

	ListDynPath(ctx context.Context, req *ListDynPathReq) (resp *ListDynPathReply, err error)

	AddDynPath(ctx context.Context, req *SetDynPathReq) (resp *google_protobuf1.Empty, err error)

	UpdateDynPath(ctx context.Context, req *SetDynPathReq) (resp *google_protobuf1.Empty, err error)

	DeleteDynPath(ctx context.Context, req *DeleteDynPathReq) (resp *google_protobuf1.Empty, err error)

	EnableDynPath(ctx context.Context, req *EnableDynPathReq) (resp *google_protobuf1.Empty, err error)

	DisableDynPath(ctx context.Context, req *EnableDynPathReq) (resp *google_protobuf1.Empty, err error)

	ListLog(ctx context.Context, req *ListLogReq) (resp *ListLogReply, err error)

	ExecuteTask(ctx context.Context, req *ExecuteTaskReq) (resp *ExecuteTaskReply, err error)

	GatewayProxy(ctx context.Context, req *GatewayProxyReq) (resp *GatewayProxyReply, err error)

	InitGatewayConfigs(ctx context.Context, req *InitGatewayConfigsReq) (resp *google_protobuf1.Empty, err error)

	AppPromptAPI(ctx context.Context, req *AppPromptAPIReq) (resp *AppPromptAPIReply, err error)

	ConfigPromptAPI(ctx context.Context, req *ConfigPromptAPIReq) (resp *ConfigPromptAPIReply, err error)

	AppPathPromptAPI(ctx context.Context, req *AppPathPromptAPIReq) (resp *AppPathPromptAPIReply, err error)

	SnapshotAction(ctx context.Context, req *SnapshotActionReq) (resp *SnapshotActionReply, err error)

	AddSnapshot(ctx context.Context, req *AddSnapshotReq) (resp *AddSnapshotReply, err error)

	SnapshotProfile(ctx context.Context, req *SnapshotProfileReq) (resp *SnapshotProfileReply, err error)

	CreateDeployment(ctx context.Context, req *CreateDeploymentReq) (resp *CreateDeploymentReply, err error)

	CompareDeployment(ctx context.Context, req *DeploymentReq) (resp *CompareDeploymentReply, err error)

	ConfirmDeployment(ctx context.Context, req *DeploymentReq) (resp *google_protobuf1.Empty, err error)

	DeployDeployment(ctx context.Context, req *DeploymentReq) (resp *google_protobuf1.Empty, err error)

	DeployDeploymentProfile(ctx context.Context, req *DeploymentReq) (resp *DeployDeploymentProfileReply, err error)

	Deployment(ctx context.Context, req *ListDeploymentReq) (resp *ListDeploymentReply, err error)

	DeployDeploymentAll(ctx context.Context, req *DeploymentReq) (resp *google_protobuf1.Empty, err error)

	RollbackDeployment(ctx context.Context, req *DeploymentReq) (resp *google_protobuf1.Empty, err error)

	FinishDeployment(ctx context.Context, req *DeploymentReq) (resp *google_protobuf1.Empty, err error)

	CloseDeployment(ctx context.Context, req *DeploymentReq) (resp *google_protobuf1.Empty, err error)

	CancelDeployment(ctx context.Context, req *DeploymentReq) (resp *google_protobuf1.Empty, err error)

	ListLimiter(ctx context.Context, req *ListLimiterReq) (resp *ListLimiterReply, err error)

	AddLimiter(ctx context.Context, req *AddLimiterReq) (resp *google_protobuf1.Empty, err error)

	UpdateLimiter(ctx context.Context, req *SetLimiterReq) (resp *google_protobuf1.Empty, err error)

	DeleteLimiter(ctx context.Context, req *DeleteLimiterReq) (resp *google_protobuf1.Empty, err error)

	EnableLimiter(ctx context.Context, req *EnableLimiterReq) (resp *google_protobuf1.Empty, err error)

	DisableLimiter(ctx context.Context, req *EnableLimiterReq) (resp *google_protobuf1.Empty, err error)

	SetupPlugin(ctx context.Context, req *PluginReq) (resp *google_protobuf1.Empty, err error)

	ZonePromptAPI(ctx context.Context, req *ZonePromptAPIReq) (resp *ZonePromptAPIReply, err error)

	PluginList(ctx context.Context, req *PluginListReq) (resp *PluginListReply, err error)

	GRPCAppMethodPromptAPI(ctx context.Context, req *AppPathPromptAPIReq) (resp *AppPathPromptAPIReply, err error)

	GRPCAppPackagePromptAPI(ctx context.Context, req *GRPCAppPackagePromptAPIReq) (resp *GRPCAppPackagePromptAPIReply, err error)
}

var ManagementSvc ManagementBMServer

func managementPing(c *bm.Context) {
	p := new(google_protobuf1.Empty)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := ManagementSvc.Ping(c, p)
	c.JSON(resp, err)
}

func managementAuthZ(c *bm.Context) {
	p := new(AuthZReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := ManagementSvc.AuthZ(c, p)
	c.JSON(resp, err)
}

func managementGateway(c *bm.Context) {
	p := new(AuthZReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := ManagementSvc.Gateway(c, p)
	c.JSON(resp, err)
}

func managementGatewayProfile(c *bm.Context) {
	p := new(GatewayProfileReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := ManagementSvc.GatewayProfile(c, p)
	c.JSON(resp, err)
}

func managementAddGateway(c *bm.Context) {
	p := new(SetGatewayReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := ManagementSvc.AddGateway(c, p)
	c.JSON(resp, err)
}

func managementUpdateGateway(c *bm.Context) {
	p := new(SetGatewayReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := ManagementSvc.UpdateGateway(c, p)
	c.JSON(resp, err)
}

func managementDeleteGateway(c *bm.Context) {
	p := new(DeleteGatewayReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := ManagementSvc.DeleteGateway(c, p)
	c.JSON(resp, err)
}

func managementEnableALLGatewayConfig(c *bm.Context) {
	p := new(UpdateALLGatewayConfigReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := ManagementSvc.EnableALLGatewayConfig(c, p)
	c.JSON(resp, err)
}

func managementDisableALLGatewayConfig(c *bm.Context) {
	p := new(UpdateALLGatewayConfigReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := ManagementSvc.DisableALLGatewayConfig(c, p)
	c.JSON(resp, err)
}

func managementEnableAllGRPCGatewayConfig(c *bm.Context) {
	p := new(UpdateALLGatewayConfigReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := ManagementSvc.EnableAllGRPCGatewayConfig(c, p)
	c.JSON(resp, err)
}

func managementDisableAllGRPCGatewayConfig(c *bm.Context) {
	p := new(UpdateALLGatewayConfigReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := ManagementSvc.DisableAllGRPCGatewayConfig(c, p)
	c.JSON(resp, err)
}

func managementListBreakerAPI(c *bm.Context) {
	p := new(ListBreakerAPIReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := ManagementSvc.ListBreakerAPI(c, p)
	c.JSON(resp, err)
}

func managementSetBreakerAPI(c *bm.Context) {
	p := new(SetBreakerAPIReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := ManagementSvc.SetBreakerAPI(c, p)
	c.JSON(resp, err)
}

func managementUpdateBreakerAPI(c *bm.Context) {
	p := new(SetBreakerAPIReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := ManagementSvc.UpdateBreakerAPI(c, p)
	c.JSON(resp, err)
}

func managementEnableBreakerAPI(c *bm.Context) {
	p := new(EnableBreakerAPIReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := ManagementSvc.EnableBreakerAPI(c, p)
	c.JSON(resp, err)
}

func managementDisableBreakerAPI(c *bm.Context) {
	p := new(EnableBreakerAPIReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := ManagementSvc.DisableBreakerAPI(c, p)
	c.JSON(resp, err)
}

func managementDeleteBreakerAPI(c *bm.Context) {
	p := new(DeleteBreakerAPIReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := ManagementSvc.DeleteBreakerAPI(c, p)
	c.JSON(resp, err)
}

func managementListDynPath(c *bm.Context) {
	p := new(ListDynPathReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := ManagementSvc.ListDynPath(c, p)
	c.JSON(resp, err)
}

func managementAddDynPath(c *bm.Context) {
	p := new(SetDynPathReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := ManagementSvc.AddDynPath(c, p)
	c.JSON(resp, err)
}

func managementUpdateDynPath(c *bm.Context) {
	p := new(SetDynPathReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := ManagementSvc.UpdateDynPath(c, p)
	c.JSON(resp, err)
}

func managementDeleteDynPath(c *bm.Context) {
	p := new(DeleteDynPathReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := ManagementSvc.DeleteDynPath(c, p)
	c.JSON(resp, err)
}

func managementEnableDynPath(c *bm.Context) {
	p := new(EnableDynPathReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := ManagementSvc.EnableDynPath(c, p)
	c.JSON(resp, err)
}

func managementDisableDynPath(c *bm.Context) {
	p := new(EnableDynPathReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := ManagementSvc.DisableDynPath(c, p)
	c.JSON(resp, err)
}

func managementListLog(c *bm.Context) {
	p := new(ListLogReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := ManagementSvc.ListLog(c, p)
	c.JSON(resp, err)
}

func managementExecuteTask(c *bm.Context) {
	p := new(ExecuteTaskReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := ManagementSvc.ExecuteTask(c, p)
	c.JSON(resp, err)
}

func managementGatewayProxy(c *bm.Context) {
	p := new(GatewayProxyReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := ManagementSvc.GatewayProxy(c, p)
	c.JSON(resp, err)
}

func managementInitGatewayConfigs(c *bm.Context) {
	p := new(InitGatewayConfigsReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := ManagementSvc.InitGatewayConfigs(c, p)
	c.JSON(resp, err)
}

func managementAppPromptAPI(c *bm.Context) {
	p := new(AppPromptAPIReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := ManagementSvc.AppPromptAPI(c, p)
	c.JSON(resp, err)
}

func managementConfigPromptAPI(c *bm.Context) {
	p := new(ConfigPromptAPIReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := ManagementSvc.ConfigPromptAPI(c, p)
	c.JSON(resp, err)
}

func managementAppPathPromptAPI(c *bm.Context) {
	p := new(AppPathPromptAPIReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := ManagementSvc.AppPathPromptAPI(c, p)
	c.JSON(resp, err)
}

func managementSnapshotAction(c *bm.Context) {
	p := new(SnapshotActionReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := ManagementSvc.SnapshotAction(c, p)
	c.JSON(resp, err)
}

func managementAddSnapshot(c *bm.Context) {
	p := new(AddSnapshotReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := ManagementSvc.AddSnapshot(c, p)
	c.JSON(resp, err)
}

func managementSnapshotProfile(c *bm.Context) {
	p := new(SnapshotProfileReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := ManagementSvc.SnapshotProfile(c, p)
	c.JSON(resp, err)
}

func managementCreateDeployment(c *bm.Context) {
	p := new(CreateDeploymentReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := ManagementSvc.CreateDeployment(c, p)
	c.JSON(resp, err)
}

func managementCompareDeployment(c *bm.Context) {
	p := new(DeploymentReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := ManagementSvc.CompareDeployment(c, p)
	c.JSON(resp, err)
}

func managementConfirmDeployment(c *bm.Context) {
	p := new(DeploymentReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := ManagementSvc.ConfirmDeployment(c, p)
	c.JSON(resp, err)
}

func managementDeployDeployment(c *bm.Context) {
	p := new(DeploymentReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := ManagementSvc.DeployDeployment(c, p)
	c.JSON(resp, err)
}

func managementDeployDeploymentProfile(c *bm.Context) {
	p := new(DeploymentReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := ManagementSvc.DeployDeploymentProfile(c, p)
	c.JSON(resp, err)
}

func managementDeployment(c *bm.Context) {
	p := new(ListDeploymentReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := ManagementSvc.Deployment(c, p)
	c.JSON(resp, err)
}

func managementDeployDeploymentAll(c *bm.Context) {
	p := new(DeploymentReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := ManagementSvc.DeployDeploymentAll(c, p)
	c.JSON(resp, err)
}

func managementRollbackDeployment(c *bm.Context) {
	p := new(DeploymentReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := ManagementSvc.RollbackDeployment(c, p)
	c.JSON(resp, err)
}

func managementFinishDeployment(c *bm.Context) {
	p := new(DeploymentReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := ManagementSvc.FinishDeployment(c, p)
	c.JSON(resp, err)
}

func managementCloseDeployment(c *bm.Context) {
	p := new(DeploymentReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := ManagementSvc.CloseDeployment(c, p)
	c.JSON(resp, err)
}

func managementCancelDeployment(c *bm.Context) {
	p := new(DeploymentReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := ManagementSvc.CancelDeployment(c, p)
	c.JSON(resp, err)
}

func managementListLimiter(c *bm.Context) {
	p := new(ListLimiterReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := ManagementSvc.ListLimiter(c, p)
	c.JSON(resp, err)
}

func managementAddLimiter(c *bm.Context) {
	p := new(AddLimiterReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := ManagementSvc.AddLimiter(c, p)
	c.JSON(resp, err)
}

func managementUpdateLimiter(c *bm.Context) {
	p := new(SetLimiterReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := ManagementSvc.UpdateLimiter(c, p)
	c.JSON(resp, err)
}

func managementDeleteLimiter(c *bm.Context) {
	p := new(DeleteLimiterReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := ManagementSvc.DeleteLimiter(c, p)
	c.JSON(resp, err)
}

func managementEnableLimiter(c *bm.Context) {
	p := new(EnableLimiterReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := ManagementSvc.EnableLimiter(c, p)
	c.JSON(resp, err)
}

func managementDisableLimiter(c *bm.Context) {
	p := new(EnableLimiterReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := ManagementSvc.DisableLimiter(c, p)
	c.JSON(resp, err)
}

func managementSetupPlugin(c *bm.Context) {
	p := new(PluginReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := ManagementSvc.SetupPlugin(c, p)
	c.JSON(resp, err)
}

func managementZonePromptAPI(c *bm.Context) {
	p := new(ZonePromptAPIReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := ManagementSvc.ZonePromptAPI(c, p)
	c.JSON(resp, err)
}

func managementPluginList(c *bm.Context) {
	p := new(PluginListReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := ManagementSvc.PluginList(c, p)
	c.JSON(resp, err)
}

func managementGRPCAppMethodPromptAPI(c *bm.Context) {
	p := new(AppPathPromptAPIReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := ManagementSvc.GRPCAppMethodPromptAPI(c, p)
	c.JSON(resp, err)
}

func managementGRPCAppPackagePromptAPI(c *bm.Context) {
	p := new(GRPCAppPackagePromptAPIReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := ManagementSvc.GRPCAppPackagePromptAPI(c, p)
	c.JSON(resp, err)
}

// RegisterManagementBMServer Register the blademaster route
func RegisterManagementBMServer(e *bm.Engine, server ManagementBMServer) {
	ManagementSvc = server
	e.GET("/appgw.management.v1.Management/Ping", managementPing)
	e.GET("/appgw.management.v1.Management/AuthZ", managementAuthZ)
	e.GET("/appgw.management.v1.Management/Gateway", managementGateway)
	e.GET("/appgw.management.v1.Management/GatewayProfile", managementGatewayProfile)
	e.GET("/appgw.management.v1.Management/AddGateway", managementAddGateway)
	e.GET("/appgw.management.v1.Management/UpdateGateway", managementUpdateGateway)
	e.GET("/appgw.management.v1.Management/DeleteGateway", managementDeleteGateway)
	e.GET("/appgw.management.v1.Management/EnableALLGatewayConfig", managementEnableALLGatewayConfig)
	e.GET("/appgw.management.v1.Management/DisableALLGatewayConfig", managementDisableALLGatewayConfig)
	e.GET("/appgw.management.v1.Management/EnableAllGRPCGatewayConfig", managementEnableAllGRPCGatewayConfig)
	e.GET("/appgw.management.v1.Management/DisableAllGRPCGatewayConfig", managementDisableAllGRPCGatewayConfig)
	e.GET("/appgw.management.v1.Management/ListBreakerAPI", managementListBreakerAPI)
	e.GET("/appgw.management.v1.Management/SetBreakerAPI", managementSetBreakerAPI)
	e.GET("/appgw.management.v1.Management/UpdateBreakerAPI", managementUpdateBreakerAPI)
	e.GET("/appgw.management.v1.Management/EnableBreakerAPI", managementEnableBreakerAPI)
	e.GET("/appgw.management.v1.Management/DisableBreakerAPI", managementDisableBreakerAPI)
	e.GET("/appgw.management.v1.Management/DeleteBreakerAPI", managementDeleteBreakerAPI)
	e.GET("/appgw.management.v1.Management/ListDynPath", managementListDynPath)
	e.GET("/appgw.management.v1.Management/AddDynPath", managementAddDynPath)
	e.GET("/appgw.management.v1.Management/UpdateDynPath", managementUpdateDynPath)
	e.GET("/appgw.management.v1.Management/DeleteDynPath", managementDeleteDynPath)
	e.GET("/appgw.management.v1.Management/EnableDynPath", managementEnableDynPath)
	e.GET("/appgw.management.v1.Management/DisableDynPath", managementDisableDynPath)
	e.GET("/appgw.management.v1.Management/ListLog", managementListLog)
	e.GET("/appgw.management.v1.Management/ExecuteTask", managementExecuteTask)
	e.GET("/appgw.management.v1.Management/GatewayProxy", managementGatewayProxy)
	e.GET("/appgw.management.v1.Management/InitGatewayConfigs", managementInitGatewayConfigs)
	e.GET("/appgw.management.v1.Management/AppPromptAPI", managementAppPromptAPI)
	e.GET("/appgw.management.v1.Management/ConfigPromptAPI", managementConfigPromptAPI)
	e.GET("/appgw.management.v1.Management/AppPathPromptAPI", managementAppPathPromptAPI)
	e.GET("/appgw.management.v1.Management/SnapshotAction", managementSnapshotAction)
	e.GET("/appgw.management.v1.Management/AddSnapshot", managementAddSnapshot)
	e.GET("/appgw.management.v1.Management/SnapshotProfile", managementSnapshotProfile)
	e.GET("/appgw.management.v1.Management/CreateDeployment", managementCreateDeployment)
	e.GET("/appgw.management.v1.Management/CompareDeployment", managementCompareDeployment)
	e.GET("/appgw.management.v1.Management/ConfirmDeployment", managementConfirmDeployment)
	e.GET("/appgw.management.v1.Management/DeployDeployment", managementDeployDeployment)
	e.GET("/appgw.management.v1.Management/DeployDeploymentProfile", managementDeployDeploymentProfile)
	e.GET("/appgw.management.v1.Management/Deployment", managementDeployment)
	e.GET("/appgw.management.v1.Management/DeployDeploymentAll", managementDeployDeploymentAll)
	e.GET("/appgw.management.v1.Management/RollbackDeployment", managementRollbackDeployment)
	e.GET("/appgw.management.v1.Management/FinishDeployment", managementFinishDeployment)
	e.GET("/appgw.management.v1.Management/CloseDeployment", managementCloseDeployment)
	e.GET("/appgw.management.v1.Management/CancelDeployment", managementCancelDeployment)
	e.GET("/appgw.management.v1.Management/ListLimiter", managementListLimiter)
	e.GET("/appgw.management.v1.Management/AddLimiter", managementAddLimiter)
	e.GET("/appgw.management.v1.Management/UpdateLimiter", managementUpdateLimiter)
	e.GET("/appgw.management.v1.Management/DeleteLimiter", managementDeleteLimiter)
	e.GET("/appgw.management.v1.Management/EnableLimiter", managementEnableLimiter)
	e.GET("/appgw.management.v1.Management/DisableLimiter", managementDisableLimiter)
	e.GET("/appgw.management.v1.Management/SetupPlugin", managementSetupPlugin)
	e.GET("/appgw.management.v1.Management/ZonePromptAPI", managementZonePromptAPI)
	e.GET("/appgw.management.v1.Management/PluginList", managementPluginList)
	e.GET("/appgw.management.v1.Management/GRPCAppMethodPromptAPI", managementGRPCAppMethodPromptAPI)
	e.GET("/appgw.management.v1.Management/GRPCAppPackagePromptAPI", managementGRPCAppPackagePromptAPI)
}
