package grpc

import (
	"context"
	v1 "go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/interface/service"
)

// 获取绑定配置
func (s *activityService) GetBindConfig(ctx context.Context, in *v1.GetBindConfigReq) (resp *v1.GetBindConfigResp, err error) {
	return service.ExternalBindSvr.GetBindConfig(ctx, in)
}

// 更新绑定配置
func (s *activityService) SaveBindConfig(ctx context.Context, in *v1.BindConfigInfo) (resp *v1.NoReply, err error) {
	return service.ExternalBindSvr.SaveBindConfig(ctx, in)
}

// 获取绑定配置列表
func (s *activityService) GetBindConfigList(ctx context.Context, in *v1.GetBindConfigListReq) (resp *v1.GetBindConfigListResp, err error) {
	return service.ExternalBindSvr.GetBindConfigList(ctx, in)
}

// 绑定配置时获取游戏映射
func (s *activityService) GetBindGames(ctx context.Context, in *v1.NoReply) (resp *v1.GetBindGamesResp, err error) {
	return service.ExternalBindSvr.GetBindGames(ctx, in)
}

func (s *activityService) GetBindExternals(ctx context.Context, in *v1.NoReply) (resp *v1.GetBindExternalsResp, err error) {
	return service.ExternalBindSvr.GetBindExternals(ctx, in)
}

func (s *activityService) RefreshBindConfigCache(ctx context.Context, in *v1.NoReply) (resp *v1.NoReply, err error) {
	return service.ExternalBindSvr.RefreshBindConfigCache(ctx, in)
}
