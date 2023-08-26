package pgc

import (
	"go-gateway/app/app-svr/app-feed/admin/conf"

	inlinegrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card/inline"
	epgrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/episode"
	seasongrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/season"
)

// Dao is show dao.
type Dao struct {
	// grpc
	rpcClient    seasongrpc.SeasonClient
	epClient     epgrpc.EpisodeClient
	inlineClient inlinegrpc.InlineCardClient

	userFeed *conf.UserFeed
}

// New new a bangumi dao.
func New(c *conf.Config) (*Dao, error) {
	var ep epgrpc.EpisodeClient
	var inline inlinegrpc.InlineCardClient

	rpcClient, err := seasongrpc.NewClient(nil)
	if err != nil {
		panic(err)
	}
	if ep, err = epgrpc.NewClient(nil); err != nil {
		panic(err)
	}
	if inline, err = inlinegrpc.NewClient(nil); err != nil {
		panic(err)
	}
	return &Dao{
		rpcClient:    rpcClient,
		epClient:     ep,
		inlineClient: inline,
		userFeed:     c.UserFeed,
	}, nil
}
