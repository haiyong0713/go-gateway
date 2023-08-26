package pgc

import (
	bm "go-common/library/net/http/blademaster"

	"go-gateway/app/app-svr/app-dynamic/interface/conf"

	pgcAppGrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card/app"
	pgcInlineGrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card/inline"
	pgcShareGrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card/share"
	pgcDynGrpc "git.bilibili.co/bapis/bapis-go/pgc/service/dynamic"
	pgcFollowGrpc "git.bilibili.co/bapis/bapis-go/pgc/service/follow"
	pgcEpisodeGrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/episode"
	pgcSeasonGrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/season"
)

type Dao struct {
	c *conf.Config
	// http client
	client *bm.Client
	// grpc client
	pgcAppGRPC     pgcAppGrpc.AppCardClient
	pgcFollowGRPC  pgcFollowGrpc.FollowClient
	pgcDynGRPC     pgcDynGrpc.DynamicServiceClient
	pgcShareGRPC   pgcShareGrpc.ShareClient
	pgcSeasonGRPC  pgcSeasonGrpc.SeasonClient
	pgcEpisodeGRPC pgcEpisodeGrpc.EpisodeClient
	pgcInlineGRPC  pgcInlineGrpc.InlineCardClient
}

func New(c *conf.Config) *Dao {
	d := &Dao{
		c:      c,
		client: bm.NewClient(c.HTTPClient),
	}
	var err error
	if d.pgcAppGRPC, err = pgcAppGrpc.NewClient(c.PGCAppGRPC); err != nil {
		panic(err)
	}
	if d.pgcFollowGRPC, err = pgcFollowGrpc.NewClient(c.PGCFollowGRPC); err != nil {
		panic(err)
	}
	if d.pgcDynGRPC, err = pgcDynGrpc.NewClient(c.PGCDynGRPC); err != nil {
		panic(err)
	}
	if d.pgcShareGRPC, err = pgcShareGrpc.NewClient(c.PGCShareGRPC); err != nil {
		panic(err)
	}
	if d.pgcSeasonGRPC, err = pgcSeasonGrpc.NewClient(c.PGCSeasonGRPC); err != nil {
		panic(err)
	}
	if d.pgcEpisodeGRPC, err = pgcEpisodeGrpc.NewClient(c.PGCEpisodeGRPC); err != nil {
		panic(err)
	}
	if d.pgcInlineGRPC, err = pgcInlineGrpc.NewClient(c.PGCInlineGRPC); err != nil {
		panic(err)
	}
	return d
}
