package grpc

import (
	"context"
	v1 "go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/interface/service"
)

func (s *activityService) TasksProgress(ctx context.Context, req *v1.TasksProgressReq) (*v1.TasksProgressReply, error) {
	var (
		err   error
		reply = &v1.TasksProgressReply{}
	)
	reply.Tasks, err = service.S10Svc.Tasks(ctx, req.Mid)
	return reply, err
}

func (s *activityService) TotalPoints(ctx context.Context, req *v1.TotalPointsdReq) (*v1.TotalPointsReply, error) {
	var (
		err   error
		reply = &v1.TotalPointsReply{}
	)
	_, reply.Total, err = service.S10Svc.RestPoint(ctx, req.Mid)
	return reply, err
}

func (s *activityService) HasUserPredict(ctx context.Context, req *v1.HasUserPredictReq) (*v1.HasUserPredictReply, error) {
	var (
		err   error
		reply = &v1.HasUserPredictReply{}
	)
	reply.Records, err = service.LolSvc.HasUserPredict(ctx, req.Mid, req.ContestIds)
	return reply, err
}

func (s *activityService) TaskPub(ctx context.Context, req *v1.TaskPubReq) (*v1.NoReply, error) {
	reply := new(v1.NoReply)
	err := service.S10Svc.TaskPub(ctx, req.Mid, req.Timestamp, req.Act)
	return reply, err
}
