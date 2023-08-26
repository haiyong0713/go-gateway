package client

import (
	"context"

	favApi "go-main/app/community/favorite/service/api"

	"google.golang.org/grpc"
)

const (
	clientOfFavorite = "community_favorite"

	Path4FavoriteOfIsFavoreds = "IsFavoreds"
)

func FavoriteSvrIsFavoreds(ctx context.Context, req interface{}, opts ...grpc.CallOption) (interface{}, error) {
	return favClient.IsFavoreds(ctx, req.(*favApi.IsFavoredsReq), opts...)
}

func FavoriteRpcCalling(ctx context.Context, path string, udf rpcCallingFunc, req interface{}, opts ...grpc.CallOption) (interface{}, error) {
	return innerRpcCalling(ctx, clientOfFavorite, path, udf, req, opts...)
}
