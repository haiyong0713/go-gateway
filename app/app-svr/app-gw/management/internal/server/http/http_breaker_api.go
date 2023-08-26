package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"

	pb "go-gateway/app/app-svr/app-gw/management/api"
	"go-gateway/app/app-svr/app-gw/management/internal/model"
	"go-gateway/app/app-svr/app-gw/management/internal/model/prettyecode"

	"github.com/pkg/errors"
)

type httpServer struct{}

func (httpServer) listBreakerAPI(ctx *bm.Context) {
	req := &pb.ListBreakerAPIReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	reply, err := rawSvc.HTTP.ListBreakerAPI(ctx, req)
	if err != nil {
		ctx.JSON(nil, prettyecode.WithRawError(err))
		return
	}
	out := make([]*model.BreakerAPI, 0, len(reply.BreakerApiList))
	for _, pbba := range reply.BreakerApiList {
		ba := &model.BreakerAPI{}
		ba.FromProto(pbba)
		out = append(out, ba)
	}
	ctx.JSON(out, nil)
}

func parseAction(dst *pb.SetBreakerAPIReq, req *http.Request) error {
	action := req.Form.Get("action")
	if action == "" {
		return nil
	}
	null := &pb.BreakerByNull{}
	if err := json.Unmarshal([]byte(action), null); err != nil {
		return errors.WithStack(err)
	}
	ba := &pb.BreakerAction{}
	switch null.Name {
	case "null", "":
		null.Name = "null"
		ba.Action = &pb.BreakerAction_Null{Null: null}
	case "ecode":
		ecode := &pb.BreakerByEcode{}
		if err := json.Unmarshal([]byte(action), ecode); err != nil {
			return errors.WithStack(err)
		}
		ba.Action = &pb.BreakerAction_Ecode{Ecode: ecode}
	case "placeholder":
		placeholder := &pb.BreakerByPlaceholder{}
		if err := json.Unmarshal([]byte(action), placeholder); err != nil {
			return errors.WithStack(err)
		}
		checker := make(map[string]interface{})
		err := json.Unmarshal([]byte(placeholder.Data), &checker)
		if err != nil {
			return errors.WithStack(err)
		}
		ba.Action = &pb.BreakerAction_Placeholder{Placeholder: placeholder}
	case "directly_backup":
		directlyBackup := &pb.BreakerByDirectlyBackup{}
		if err := json.Unmarshal([]byte(action), directlyBackup); err != nil {
			return errors.WithStack(err)
		}
		u, err := url.Parse(directlyBackup.BackupUrl)
		if err != nil {
			return errors.WithStack(err)
		}
		if u.Scheme == "" || u.Host == "" {
			return errors.WithStack(fmt.Errorf("invalid backup_url: %v", directlyBackup.BackupUrl))
		}
		ba.Action = &pb.BreakerAction_DirectlyBackup{
			DirectlyBackup: directlyBackup,
		}
	case "retry_backup":
		retryBackup := &pb.BreakerByRetryBackup{}
		if err := json.Unmarshal([]byte(action), retryBackup); err != nil {
			return errors.WithStack(err)
		}
		u, err := url.Parse(retryBackup.BackupUrl)
		if err != nil {
			return errors.WithStack(err)
		}
		if u.Scheme == "" || u.Host == "" {
			return errors.WithStack(fmt.Errorf("invalid backup_url: %v", retryBackup.BackupUrl))
		}
		ba.Action = &pb.BreakerAction_RetryBackup{RetryBackup: retryBackup}
	}
	dst.Action = ba
	return nil
}

func parseFlowCopy(dst *pb.SetBreakerAPIReq, req *http.Request) error {
	flowCopy := req.Form.Get("flow_copy")
	if flowCopy == "" {
		return nil
	}
	null := &pb.CopyByNull{}
	if err := json.Unmarshal([]byte(flowCopy), null); err != nil {
		return errors.WithStack(err)
	}
	fc := &pb.FlowCopy{}
	switch null.Name {
	case "null", "":
		null.Name = "null"
		fc.Flow = &pb.FlowCopy_Null{Null: null}
	case "ratio":
		ratio := &pb.CopyByRatio{}
		if err := json.Unmarshal([]byte(flowCopy), ratio); err != nil {
			return errors.WithStack(err)
		}
		fc.Flow = &pb.FlowCopy_Ratio{Ratio: ratio}
	case "qps":
		qps := &pb.CopyByQPS{}
		if err := json.Unmarshal([]byte(flowCopy), qps); err != nil {
			return errors.WithStack(err)
		}
		fc.Flow = &pb.FlowCopy_Qps{Qps: qps}
	}
	dst.FlowCopy = fc
	return nil
}

func (httpServer) addBreakerAPI(ctx *bm.Context) {
	req := &pb.SetBreakerAPIReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	if err := parseAction(req, ctx.Request); err != nil {
		ctx.JSON(nil, prettyecode.WithError(ecode.RequestErr, err))
		return
	}
	if err := parseFlowCopy(req, ctx.Request); err != nil {
		ctx.JSON(nil, prettyecode.WithError(ecode.RequestErr, err))
		return
	}
	req.Username = getUsername(ctx)
	res, err := rawSvc.HTTP.SetBreakerAPI(ctx, req)
	if err != nil {
		ctx.JSON(nil, prettyecode.WithRawError(err))
		return
	}
	ctx.JSON(res, nil)
}

func (httpServer) updateBreakerAPI(ctx *bm.Context) {
	req := &pb.SetBreakerAPIReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	if err := parseAction(req, ctx.Request); err != nil {
		ctx.JSON(nil, prettyecode.WithError(ecode.RequestErr, err))
		return
	}
	if err := parseFlowCopy(req, ctx.Request); err != nil {
		ctx.JSON(nil, prettyecode.WithError(ecode.RequestErr, err))
		return
	}
	req.Username = getUsername(ctx)
	res, err := rawSvc.HTTP.UpdateBreakerAPI(ctx, req)
	if err != nil {
		ctx.JSON(nil, prettyecode.WithRawError(err))
		return
	}
	ctx.JSON(res, nil)
}

func (httpServer) enableBreakerAPI(ctx *bm.Context) {
	req := &pb.EnableBreakerAPIReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	req.Username = getUsername(ctx)
	res, err := rawSvc.HTTP.EnableBreakerAPI(ctx, req)
	if err != nil {
		ctx.JSON(nil, prettyecode.WithRawError(err))
		return
	}
	ctx.JSON(res, nil)
}

func (httpServer) disableBreakerAPI(ctx *bm.Context) {
	req := &pb.EnableBreakerAPIReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	req.Username = getUsername(ctx)
	res, err := rawSvc.HTTP.DisableBreakerAPI(ctx, req)
	if err != nil {
		ctx.JSON(nil, prettyecode.WithRawError(err))
		return
	}
	ctx.JSON(res, nil)
}

func (httpServer) deleteBreakerAPI(ctx *bm.Context) {
	req := &pb.DeleteBreakerAPIReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	req.Username = getUsername(ctx)
	res, err := rawSvc.HTTP.DeleteBreakerAPI(ctx, req)
	if err != nil {
		ctx.JSON(nil, prettyecode.WithRawError(err))
		return
	}
	ctx.JSON(res, nil)
}
