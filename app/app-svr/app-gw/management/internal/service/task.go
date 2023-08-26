package service

import (
	"context"
	"fmt"
	"time"

	"go-common/library/log"
	managementjobapi "go-gateway/app/app-svr/app-gw/management-job/api"
	"go-gateway/app/app-svr/app-gw/management/api"
	"go-gateway/app/app-svr/app-gw/management/audit"
)

func (s *CommonService) asyncTriggerConfigPush(ctx context.Context, node, gateway, username, deploymentType string) {
	now := time.Now()
	key := fmt.Sprintf("{trigger-push}/%s.%s_%d", node, gateway, now.Second())
	if err := s.dao.TryLock(ctx, key, []byte{}, []byte(key), 5); err != nil {
		log.Warn("Skip to push config due to pre-condition failed: %+v", err)
		return
	}
	req := &api.ExecuteTaskReq{
		Node:     node,
		Gateway:  gateway,
		Task:     matchTask(deploymentType),
		Username: username,
	}
	if _, err := s.ExecuteTask(ctx, req); err != nil {
		log.Error("Failed to trigger push config to gateway: `%s.%s`: %+v", node, gateway, err)
		return
	}
}

func (s *CommonService) ExecuteTask(ctx context.Context, req *api.ExecuteTaskReq) (*api.ExecuteTaskReply, error) {
	param := &managementjobapi.Params{
		Node:    req.Node,
		Gateway: req.Gateway,
		Ctime:   time.Now().Unix(),
		Mtime:   time.Now().Unix(),
	}
	taskDoReq := &managementjobapi.TaskDoReq{
		Name:    req.Task,
		Params:  param,
		Sponsor: req.Username,
	}
	taskDoReply, err := s.managementjob.TaskDo(ctx, taskDoReq)
	if err != nil {
		audit.SendTriggerTaskExecuteLog(req, audit.LogLevelError, audit.LogResultFailure, fmt.Sprintf("%+v", err), param.Ctime, param.Mtime)
		return nil, err
	}
	out := &api.ExecuteTaskReply{
		TaskId: taskDoReply.TaskId,
	}
	audit.SendTriggerTaskExecuteLog(req, audit.LogLevelInfo, audit.LogResultSuccess, jsonify(req), param.Ctime, param.Mtime)
	log.Info("Succeed to submit async task with task id: %q and params: %+v", out.TaskId, req)
	return out, nil
}
