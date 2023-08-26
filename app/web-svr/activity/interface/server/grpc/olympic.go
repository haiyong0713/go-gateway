package grpc

import (
	"context"
	pb "go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/interface/service"
)

// 获取奥运会单赛程详情
func (s *activityService) GetOlympicContestDetail(ctx context.Context, in *pb.GetOlympicContestDetailReq) (resp *pb.GetOlympicContestDetailResp, err error) {
	resp = new(pb.GetOlympicContestDetailResp)
	resp, err = service.OlympicSvr.GetOlympicContestDetail(ctx, in.Id, in.SkipCache)
	return
}

// 获取奥运会query词的配置信息
func (s *activityService) GetOlympicQueryConfig(ctx context.Context, in *pb.GetOlympicQueryConfigReq) (resp *pb.GetOlympicQueryConfigResp, err error) {
	resp = new(pb.GetOlympicQueryConfigResp)
	resp, err = service.OlympicSvr.GetOlympicQueryConfigs(ctx, in.SkipCache)
	return
}
