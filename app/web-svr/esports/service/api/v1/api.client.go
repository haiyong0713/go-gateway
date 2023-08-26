// Code generated by grpclocal. DO NOT EDIT.
package v1

import (
	"context"
	"go-common/library/net/rpc/warden"
	"go-gateway/app/web-svr/activity/tools/lib/grpclocal"
	"google.golang.org/grpc"
)

type localEsportsServiceClient struct {
}

var (
	localEsportsServiceServer EsportsServiceServer
	_                         EsportsServiceClient = &localEsportsServiceClient{}
)

func InitLocalEsportsServiceServer(svc EsportsServiceServer) {
	localEsportsServiceServer = svc
}
func NewLocalClient(cfg *warden.ClientConfig, opts ...grpc.DialOption) (EsportsServiceClient, error) {
	return &localEsportsServiceClient{}, nil
}

func (s *localEsportsServiceClient) GetContestInfo(ctx context.Context, in *GetContestRequest, opts ...grpc.CallOption) (rly *ContestInfo, err error) {
	if localEsportsServiceServer == nil {
		panic("Call InitLocalEsportsServiceServer First")
	}
	grpclocal.ServerLogging(ctx, "/"+_EsportsService_serviceDesc.ServiceName+"/GetContestInfo", in.String(), func() error {
		rly, err = localEsportsServiceServer.GetContestInfo(ctx, in)
		return err
	})
	return
}

func (s *localEsportsServiceClient) GetContestsByTime(ctx context.Context, in *GetTimeContestsRequest, opts ...grpc.CallOption) (rly *GetTimeContestsResponse, err error) {
	if localEsportsServiceServer == nil {
		panic("Call InitLocalEsportsServiceServer First")
	}
	grpclocal.ServerLogging(ctx, "/"+_EsportsService_serviceDesc.ServiceName+"/GetContestsByTime", in.String(), func() error {
		rly, err = localEsportsServiceServer.GetContestsByTime(ctx, in)
		return err
	})
	return
}

func (s *localEsportsServiceClient) GetContests(ctx context.Context, in *GetContestsRequest, opts ...grpc.CallOption) (rly *ContestsResponse, err error) {
	if localEsportsServiceServer == nil {
		panic("Call InitLocalEsportsServiceServer First")
	}
	grpclocal.ServerLogging(ctx, "/"+_EsportsService_serviceDesc.ServiceName+"/GetContests", in.String(), func() error {
		rly, err = localEsportsServiceServer.GetContests(ctx, in)
		return err
	})
	return
}

func (s *localEsportsServiceClient) AddContestFav(ctx context.Context, in *FavRequest, opts ...grpc.CallOption) (rly *NoArgsResponse, err error) {
	if localEsportsServiceServer == nil {
		panic("Call InitLocalEsportsServiceServer First")
	}
	grpclocal.ServerLogging(ctx, "/"+_EsportsService_serviceDesc.ServiceName+"/AddContestFav", in.String(), func() error {
		rly, err = localEsportsServiceServer.AddContestFav(ctx, in)
		return err
	})
	return
}

func (s *localEsportsServiceClient) DelContestFav(ctx context.Context, in *FavRequest, opts ...grpc.CallOption) (rly *NoArgsResponse, err error) {
	if localEsportsServiceServer == nil {
		panic("Call InitLocalEsportsServiceServer First")
	}
	grpclocal.ServerLogging(ctx, "/"+_EsportsService_serviceDesc.ServiceName+"/DelContestFav", in.String(), func() error {
		rly, err = localEsportsServiceServer.DelContestFav(ctx, in)
		return err
	})
	return
}

func (s *localEsportsServiceClient) GetContestSubscribers(ctx context.Context, in *GetSubscribersRequest, opts ...grpc.CallOption) (rly *ContestSubscribers, err error) {
	if localEsportsServiceServer == nil {
		panic("Call InitLocalEsportsServiceServer First")
	}
	grpclocal.ServerLogging(ctx, "/"+_EsportsService_serviceDesc.ServiceName+"/GetContestSubscribers", in.String(), func() error {
		rly, err = localEsportsServiceServer.GetContestSubscribers(ctx, in)
		return err
	})
	return
}

func (s *localEsportsServiceClient) GetGames(ctx context.Context, in *GetGamesRequest, opts ...grpc.CallOption) (rly *GamesResponse, err error) {
	if localEsportsServiceServer == nil {
		panic("Call InitLocalEsportsServiceServer First")
	}
	grpclocal.ServerLogging(ctx, "/"+_EsportsService_serviceDesc.ServiceName+"/GetGames", in.String(), func() error {
		rly, err = localEsportsServiceServer.GetGames(ctx, in)
		return err
	})
	return
}

func (s *localEsportsServiceClient) GetTeams(ctx context.Context, in *GetTeamsRequest, opts ...grpc.CallOption) (rly *TeamsResponse, err error) {
	if localEsportsServiceServer == nil {
		panic("Call InitLocalEsportsServiceServer First")
	}
	grpclocal.ServerLogging(ctx, "/"+_EsportsService_serviceDesc.ServiceName+"/GetTeams", in.String(), func() error {
		rly, err = localEsportsServiceServer.GetTeams(ctx, in)
		return err
	})
	return
}

func (s *localEsportsServiceClient) SaveContest(ctx context.Context, in *SaveContestReq, opts ...grpc.CallOption) (rly *NoArgsResponse, err error) {
	if localEsportsServiceServer == nil {
		panic("Call InitLocalEsportsServiceServer First")
	}
	grpclocal.ServerLogging(ctx, "/"+_EsportsService_serviceDesc.ServiceName+"/SaveContest", in.String(), func() error {
		rly, err = localEsportsServiceServer.SaveContest(ctx, in)
		return err
	})
	return
}

func (s *localEsportsServiceClient) GetContestModel(ctx context.Context, in *GetContestModelReq, opts ...grpc.CallOption) (rly *ContestModel, err error) {
	if localEsportsServiceServer == nil {
		panic("Call InitLocalEsportsServiceServer First")
	}
	grpclocal.ServerLogging(ctx, "/"+_EsportsService_serviceDesc.ServiceName+"/GetContestModel", in.String(), func() error {
		rly, err = localEsportsServiceServer.GetContestModel(ctx, in)
		return err
	})
	return
}

func (s *localEsportsServiceClient) GetSeasonModel(ctx context.Context, in *GetSeasonModelReq, opts ...grpc.CallOption) (rly *SeasonModel, err error) {
	if localEsportsServiceServer == nil {
		panic("Call InitLocalEsportsServiceServer First")
	}
	grpclocal.ServerLogging(ctx, "/"+_EsportsService_serviceDesc.ServiceName+"/GetSeasonModel", in.String(), func() error {
		rly, err = localEsportsServiceServer.GetSeasonModel(ctx, in)
		return err
	})
	return
}

func (s *localEsportsServiceClient) GetSeasonDetail(ctx context.Context, in *GetSeasonModelReq, opts ...grpc.CallOption) (rly *SeasonDetail, err error) {
	if localEsportsServiceServer == nil {
		panic("Call InitLocalEsportsServiceServer First")
	}
	grpclocal.ServerLogging(ctx, "/"+_EsportsService_serviceDesc.ServiceName+"/GetSeasonDetail", in.String(), func() error {
		rly, err = localEsportsServiceServer.GetSeasonDetail(ctx, in)
		return err
	})
	return
}

func (s *localEsportsServiceClient) GetSeasonByTime(ctx context.Context, in *GetSeasonByTimeReq, opts ...grpc.CallOption) (rly *GetSeasonByTimeResponse, err error) {
	if localEsportsServiceServer == nil {
		panic("Call InitLocalEsportsServiceServer First")
	}
	grpclocal.ServerLogging(ctx, "/"+_EsportsService_serviceDesc.ServiceName+"/GetSeasonByTime", in.String(), func() error {
		rly, err = localEsportsServiceServer.GetSeasonByTime(ctx, in)
		return err
	})
	return
}

func (s *localEsportsServiceClient) GetTeamModel(ctx context.Context, in *GetTeamModelReq, opts ...grpc.CallOption) (rly *TeamModel, err error) {
	if localEsportsServiceServer == nil {
		panic("Call InitLocalEsportsServiceServer First")
	}
	grpclocal.ServerLogging(ctx, "/"+_EsportsService_serviceDesc.ServiceName+"/GetTeamModel", in.String(), func() error {
		rly, err = localEsportsServiceServer.GetTeamModel(ctx, in)
		return err
	})
	return
}

func (s *localEsportsServiceClient) ClearTeamCache(ctx context.Context, in *ClearTeamCacheReq, opts ...grpc.CallOption) (rly *NoArgsResponse, err error) {
	if localEsportsServiceServer == nil {
		panic("Call InitLocalEsportsServiceServer First")
	}
	grpclocal.ServerLogging(ctx, "/"+_EsportsService_serviceDesc.ServiceName+"/ClearTeamCache", in.String(), func() error {
		rly, err = localEsportsServiceServer.ClearTeamCache(ctx, in)
		return err
	})
	return
}

func (s *localEsportsServiceClient) RefreshContestStatusInfo(ctx context.Context, in *RefreshContestStatusInfoReq, opts ...grpc.CallOption) (rly *NoArgsResponse, err error) {
	if localEsportsServiceServer == nil {
		panic("Call InitLocalEsportsServiceServer First")
	}
	grpclocal.ServerLogging(ctx, "/"+_EsportsService_serviceDesc.ServiceName+"/RefreshContestStatusInfo", in.String(), func() error {
		rly, err = localEsportsServiceServer.RefreshContestStatusInfo(ctx, in)
		return err
	})
	return
}

func (s *localEsportsServiceClient) GetSeasonContests(ctx context.Context, in *GetSeasonContestsReq, opts ...grpc.CallOption) (rly *SeasonContests, err error) {
	if localEsportsServiceServer == nil {
		panic("Call InitLocalEsportsServiceServer First")
	}
	grpclocal.ServerLogging(ctx, "/"+_EsportsService_serviceDesc.ServiceName+"/GetSeasonContests", in.String(), func() error {
		rly, err = localEsportsServiceServer.GetSeasonContests(ctx, in)
		return err
	})
	return
}

func (s *localEsportsServiceClient) GetContestInfoListBySeason(ctx context.Context, in *GetContestInfoListBySeasonReq, opts ...grpc.CallOption) (rly *GetContestInfoListBySeasonResponse, err error) {
	if localEsportsServiceServer == nil {
		panic("Call InitLocalEsportsServiceServer First")
	}
	grpclocal.ServerLogging(ctx, "/"+_EsportsService_serviceDesc.ServiceName+"/GetContestInfoListBySeason", in.String(), func() error {
		rly, err = localEsportsServiceServer.GetContestInfoListBySeason(ctx, in)
		return err
	})
	return
}

func (s *localEsportsServiceClient) GetSeasonSeriesModel(ctx context.Context, in *GetSeasonSeriesReq, opts ...grpc.CallOption) (rly *GetSeasonSeriesResponse, err error) {
	if localEsportsServiceServer == nil {
		panic("Call InitLocalEsportsServiceServer First")
	}
	grpclocal.ServerLogging(ctx, "/"+_EsportsService_serviceDesc.ServiceName+"/GetSeasonSeriesModel", in.String(), func() error {
		rly, err = localEsportsServiceServer.GetSeasonSeriesModel(ctx, in)
		return err
	})
	return
}

func (s *localEsportsServiceClient) GetContestGameModel(ctx context.Context, in *GetContestGameReq, opts ...grpc.CallOption) (rly *GameModel, err error) {
	if localEsportsServiceServer == nil {
		panic("Call InitLocalEsportsServiceServer First")
	}
	grpclocal.ServerLogging(ctx, "/"+_EsportsService_serviceDesc.ServiceName+"/GetContestGameModel", in.String(), func() error {
		rly, err = localEsportsServiceServer.GetContestGameModel(ctx, in)
		return err
	})
	return
}

func (s *localEsportsServiceClient) GetContestGameDetail(ctx context.Context, in *GetContestGameReq, opts ...grpc.CallOption) (rly *GameDetail, err error) {
	if localEsportsServiceServer == nil {
		panic("Call InitLocalEsportsServiceServer First")
	}
	grpclocal.ServerLogging(ctx, "/"+_EsportsService_serviceDesc.ServiceName+"/GetContestGameDetail", in.String(), func() error {
		rly, err = localEsportsServiceServer.GetContestGameDetail(ctx, in)
		return err
	})
	return
}

func (s *localEsportsServiceClient) RefreshActiveSeasons(ctx context.Context, in *NoArgsRequest, opts ...grpc.CallOption) (rly *ActiveSeasonsResponse, err error) {
	if localEsportsServiceServer == nil {
		panic("Call InitLocalEsportsServiceServer First")
	}
	grpclocal.ServerLogging(ctx, "/"+_EsportsService_serviceDesc.ServiceName+"/RefreshActiveSeasons", in.String(), func() error {
		rly, err = localEsportsServiceServer.RefreshActiveSeasons(ctx, in)
		return err
	})
	return
}

func (s *localEsportsServiceClient) RefreshSeasonContestIdsCache(ctx context.Context, in *RefreshSeasonContestIdsReq, opts ...grpc.CallOption) (rly *RefreshSeasonContestIdsResponse, err error) {
	if localEsportsServiceServer == nil {
		panic("Call InitLocalEsportsServiceServer First")
	}
	grpclocal.ServerLogging(ctx, "/"+_EsportsService_serviceDesc.ServiceName+"/RefreshSeasonContestIdsCache", in.String(), func() error {
		rly, err = localEsportsServiceServer.RefreshSeasonContestIdsCache(ctx, in)
		return err
	})
	return
}

func (s *localEsportsServiceClient) RefreshContestCache(ctx context.Context, in *RefreshContestCacheReq, opts ...grpc.CallOption) (rly *NoArgsResponse, err error) {
	if localEsportsServiceServer == nil {
		panic("Call InitLocalEsportsServiceServer First")
	}
	grpclocal.ServerLogging(ctx, "/"+_EsportsService_serviceDesc.ServiceName+"/RefreshContestCache", in.String(), func() error {
		rly, err = localEsportsServiceServer.RefreshContestCache(ctx, in)
		return err
	})
	return
}

func (s *localEsportsServiceClient) RefreshTeamCache(ctx context.Context, in *RefreshTeamCacheReq, opts ...grpc.CallOption) (rly *NoArgsResponse, err error) {
	if localEsportsServiceServer == nil {
		panic("Call InitLocalEsportsServiceServer First")
	}
	grpclocal.ServerLogging(ctx, "/"+_EsportsService_serviceDesc.ServiceName+"/RefreshTeamCache", in.String(), func() error {
		rly, err = localEsportsServiceServer.RefreshTeamCache(ctx, in)
		return err
	})
	return
}

func (s *localEsportsServiceClient) RefreshSeriesCache(ctx context.Context, in *RefreshSeriesCacheReq, opts ...grpc.CallOption) (rly *NoArgsResponse, err error) {
	if localEsportsServiceServer == nil {
		panic("Call InitLocalEsportsServiceServer First")
	}
	grpclocal.ServerLogging(ctx, "/"+_EsportsService_serviceDesc.ServiceName+"/RefreshSeriesCache", in.String(), func() error {
		rly, err = localEsportsServiceServer.RefreshSeriesCache(ctx, in)
		return err
	})
	return
}

func (s *localEsportsServiceClient) RefreshGameCache(ctx context.Context, in *NoArgsRequest, opts ...grpc.CallOption) (rly *NoArgsResponse, err error) {
	if localEsportsServiceServer == nil {
		panic("Call InitLocalEsportsServiceServer First")
	}
	grpclocal.ServerLogging(ctx, "/"+_EsportsService_serviceDesc.ServiceName+"/RefreshGameCache", in.String(), func() error {
		rly, err = localEsportsServiceServer.RefreshGameCache(ctx, in)
		return err
	})
	return
}

func (s *localEsportsServiceClient) GetReplyWallList(ctx context.Context, in *GetReplyWallListReq, opts ...grpc.CallOption) (rly *GetReplyWallListResponse, err error) {
	if localEsportsServiceServer == nil {
		panic("Call InitLocalEsportsServiceServer First")
	}
	grpclocal.ServerLogging(ctx, "/"+_EsportsService_serviceDesc.ServiceName+"/GetReplyWallList", in.String(), func() error {
		rly, err = localEsportsServiceServer.GetReplyWallList(ctx, in)
		return err
	})
	return
}

func (s *localEsportsServiceClient) GetReplyWallModel(ctx context.Context, in *GetReplyWallModelReq, opts ...grpc.CallOption) (rly *SaveReplyWallModel, err error) {
	if localEsportsServiceServer == nil {
		panic("Call InitLocalEsportsServiceServer First")
	}
	grpclocal.ServerLogging(ctx, "/"+_EsportsService_serviceDesc.ServiceName+"/GetReplyWallModel", in.String(), func() error {
		rly, err = localEsportsServiceServer.GetReplyWallModel(ctx, in)
		return err
	})
	return
}

func (s *localEsportsServiceClient) SaveReplyWall(ctx context.Context, in *SaveReplyWallModel, opts ...grpc.CallOption) (rly *NoArgsResponse, err error) {
	if localEsportsServiceServer == nil {
		panic("Call InitLocalEsportsServiceServer First")
	}
	grpclocal.ServerLogging(ctx, "/"+_EsportsService_serviceDesc.ServiceName+"/SaveReplyWall", in.String(), func() error {
		rly, err = localEsportsServiceServer.SaveReplyWall(ctx, in)
		return err
	})
	return
}