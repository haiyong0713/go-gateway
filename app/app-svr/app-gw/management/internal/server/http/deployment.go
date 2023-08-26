package http

import (
	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	pb "go-gateway/app/app-svr/app-gw/management/api"
	"go-gateway/app/app-svr/app-gw/management/internal/model/prettyecode"
)

func resolveDeploymentType(req string) (string, error) {
	if req == "" {
		return "http", nil
	}
	if req != "http" && req != "grpc" {
		return "", ecode.Error(ecode.RequestErr, "发布单类型错误")
	}
	return req, nil
}

func deployment(ctx *bm.Context) {
	req := &pb.ListDeploymentReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	deploymentType, err := resolveDeploymentType(req.DeploymentType)
	if err != nil {
		ctx.JSON(nil, prettyecode.WithRawError(err))
		return
	}
	req.DeploymentType = deploymentType
	reply, err := rawSvc.Deploy.Deployment(ctx, req)
	if err != nil {
		ctx.JSON(nil, prettyecode.WithRawError(err))
		return
	}
	ctx.JSON(reply, nil)
}

func createDeployment(ctx *bm.Context) {
	req := &pb.CreateDeploymentReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	username, _ := ctx.Get("username")
	req.Username = username.(string)
	deploymentType, err := resolveDeploymentType(req.DeploymentType)
	if err != nil {
		ctx.JSON(nil, prettyecode.WithRawError(err))
		return
	}
	req.DeploymentType = deploymentType
	reply, err := rawSvc.Deploy.CreateDeployment(ctx, req)
	if err != nil {
		ctx.JSON(nil, prettyecode.WithRawError(err))
		return
	}
	ctx.JSON(reply, nil)
}

func compareDeployment(ctx *bm.Context) {
	req := &pb.DeploymentReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	username, _ := ctx.Get("username")
	req.Username = username.(string)
	deploymentType, err := resolveDeploymentType(req.DeploymentType)
	if err != nil {
		ctx.JSON(nil, prettyecode.WithRawError(err))
		return
	}
	req.DeploymentType = deploymentType
	reply, err := rawSvc.Deploy.CompareDeployment(ctx, req)
	if err != nil {
		ctx.JSON(nil, prettyecode.WithRawError(err))
		return
	}
	ctx.JSON(reply, nil)
}

func confirmDeployment(ctx *bm.Context) {
	req := &pb.DeploymentReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	username, _ := ctx.Get("username")
	req.Username = username.(string)
	deploymentType, err := resolveDeploymentType(req.DeploymentType)
	if err != nil {
		ctx.JSON(nil, prettyecode.WithRawError(err))
		return
	}
	req.DeploymentType = deploymentType
	reply, err := rawSvc.Deploy.ConfirmDeployment(ctx, req)
	if err != nil {
		ctx.JSON(nil, prettyecode.WithRawError(err))
		return
	}
	ctx.JSON(reply, nil)
}

func deployDeployment(ctx *bm.Context) {
	req := &pb.DeploymentReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	username, _ := ctx.Get("username")
	req.Username = username.(string)
	deploymentType, err := resolveDeploymentType(req.DeploymentType)
	if err != nil {
		ctx.JSON(nil, prettyecode.WithRawError(err))
		return
	}
	req.DeploymentType = deploymentType
	reply, err := rawSvc.Deploy.DeployDeployment(ctx, req)
	if err != nil {
		ctx.JSON(nil, prettyecode.WithRawError(err))
		return
	}
	ctx.JSON(reply, nil)
}

func deployDeploymentAll(ctx *bm.Context) {
	req := &pb.DeploymentReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	username, _ := ctx.Get("username")
	req.Username = username.(string)
	deploymentType, err := resolveDeploymentType(req.DeploymentType)
	if err != nil {
		ctx.JSON(nil, prettyecode.WithRawError(err))
		return
	}
	req.DeploymentType = deploymentType
	reply, err := rawSvc.Deploy.DeployDeploymentAll(ctx, req)
	if err != nil {
		ctx.JSON(nil, prettyecode.WithRawError(err))
		return
	}
	ctx.JSON(reply, nil)
}

func deployDeploymentProfile(ctx *bm.Context) {
	req := &pb.DeploymentReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	username, _ := ctx.Get("username")
	req.Username = username.(string)
	deploymentType, err := resolveDeploymentType(req.DeploymentType)
	if err != nil {
		ctx.JSON(nil, prettyecode.WithRawError(err))
		return
	}
	req.DeploymentType = deploymentType
	reply, err := rawSvc.Deploy.DeployDeploymentProfile(ctx, req)
	if err != nil {
		ctx.JSON(nil, prettyecode.WithRawError(err))
		return
	}
	ctx.JSON(reply, nil)
}

func finishDeployment(ctx *bm.Context) {
	req := &pb.DeploymentReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	username, _ := ctx.Get("username")
	req.Username = username.(string)
	deploymentType, err := resolveDeploymentType(req.DeploymentType)
	if err != nil {
		ctx.JSON(nil, prettyecode.WithRawError(err))
		return
	}
	req.DeploymentType = deploymentType
	reply, err := rawSvc.Deploy.FinishDeployment(ctx, req)
	if err != nil {
		ctx.JSON(nil, prettyecode.WithRawError(err))
		return
	}
	ctx.JSON(reply, nil)
}

func rollbackDeployment(ctx *bm.Context) {
	req := &pb.DeploymentReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	username, _ := ctx.Get("username")
	req.Username = username.(string)
	deploymentType, err := resolveDeploymentType(req.DeploymentType)
	if err != nil {
		ctx.JSON(nil, prettyecode.WithRawError(err))
		return
	}
	req.DeploymentType = deploymentType
	reply, err := rawSvc.Deploy.RollbackDeployment(ctx, req)
	if err != nil {
		ctx.JSON(nil, prettyecode.WithRawError(err))
		return
	}
	ctx.JSON(reply, nil)
}

func closeDeployment(ctx *bm.Context) {
	req := &pb.DeploymentReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	username, _ := ctx.Get("username")
	req.Username = username.(string)
	deploymentType, err := resolveDeploymentType(req.DeploymentType)
	if err != nil {
		ctx.JSON(nil, prettyecode.WithRawError(err))
		return
	}
	req.DeploymentType = deploymentType
	reply, err := rawSvc.Deploy.CloseDeployment(ctx, req)
	if err != nil {
		ctx.JSON(nil, prettyecode.WithRawError(err))
		return
	}
	ctx.JSON(reply, nil)
}

func cancelDeployment(ctx *bm.Context) {
	req := &pb.DeploymentReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	username, _ := ctx.Get("username")
	req.Username = username.(string)
	deploymentType, err := resolveDeploymentType(req.DeploymentType)
	if err != nil {
		ctx.JSON(nil, prettyecode.WithRawError(err))
		return
	}
	req.DeploymentType = deploymentType
	reply, err := rawSvc.Deploy.CancelDeployment(ctx, req)
	if err != nil {
		ctx.JSON(nil, prettyecode.WithRawError(err))
		return
	}
	ctx.JSON(reply, nil)
}
