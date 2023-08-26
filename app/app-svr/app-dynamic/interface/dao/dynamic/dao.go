package dynamic

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-dynamic/interface/conf"
	pplApi "go-gateway/app/app-svr/app-show/interface/api"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	thumgrpc "git.bilibili.co/bapis/bapis-go/community/service/thumbup"

	relagrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
	articlegrpc "git.bilibili.co/bapis/bapis-go/article/service"
	dyngrpc "git.bilibili.co/bapis/bapis-go/dynamic/interface/feed"
	dyncampusgrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/campus-svr"
	dynagrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/feed"
	pgcAppGrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card/app"
)

type Dao struct {
	c *conf.Config
	// http client
	client *bm.Client
	// grpc
	accountGRPC     accountgrpc.AccountClient
	relaGRPC        relagrpc.RelationClient
	thumGRPC        thumgrpc.ThumbupClient
	pgcAppGRPC      pgcAppGrpc.AppCardClient
	articleGRPC     articlegrpc.ArticleGRPCClient
	dynamicGRPC     dyngrpc.FeedClient
	dynaGRPC        dynagrpc.FeedClient
	popularGRPC     pplApi.AppShowClient
	dynCampusClient dyncampusgrpc.CampusSvrClient
	// path
	videoList     string
	videoHistory  string
	topicInfo     string
	likeIcon      string
	videoPersonal string
	dynUpdOffset  string
	bottom        string
	vdUpList      string
	dynBriefs     string
	pgcInfo       string
	pgcBatch      string
	pgcSeason     string
	decoCard      string
	emoji         string
	userLike      string
	rcmd          string
	svideo        string
	drawDetailsV2 string
}

func New(c *conf.Config) (s *Dao) {
	s = &Dao{
		c:             c,
		client:        bm.NewClient(c.HTTPClient),
		videoList:     c.Hosts.VcCo + _videoListURL,
		videoHistory:  c.Hosts.VcCo + _videoHistory,
		topicInfo:     c.Hosts.VcCo + _topicInfos,
		likeIcon:      c.Hosts.VcCo + _likeIcon,
		videoPersonal: c.Hosts.VcCo + _videoPersonal,
		dynUpdOffset:  c.Hosts.VcCo + _dynUpdOffset,
		bottom:        c.Hosts.VcCo + _getBottom,
		vdUpList:      c.Hosts.VcCo + _vdUpList,
		dynBriefs:     c.Hosts.VcCo + _dynBriefs,
		pgcInfo:       c.Hosts.ApiCo + epListURL,
		pgcBatch:      c.Hosts.ApiCo + batchInfoURL,
		pgcSeason:     c.Hosts.ApiCo + seasonInfoURL,
		decoCard:      c.Hosts.ApiCo + decoCardsURL,
		emoji:         c.Hosts.ApiCo + _emojiURL,
		userLike:      c.Hosts.ApiCo + _userLikeURL,
		rcmd:          c.Hosts.Data + _rcmd,
		svideo:        c.Hosts.VcCo + _sVideo,
		drawDetailsV2: c.Hosts.VcCo + _drawDetailsV2,
	}
	var err error
	if s.accountGRPC, err = accountgrpc.NewClient(c.AccountGRPC); err != nil {
		panic(err)
	}
	if s.relaGRPC, err = relagrpc.NewClient(c.RelaGRPC); err != nil {
		panic(err)
	}
	if s.thumGRPC, err = thumgrpc.NewClient(c.ThumGRPC); err != nil {
		panic(err)
	}
	if s.articleGRPC, err = articlegrpc.NewClient(c.ArticleGRPC); err != nil {
		panic(err)
	}
	if s.pgcAppGRPC, err = pgcAppGrpc.NewClient(c.PGCAppGRPC); err != nil {
		panic(err)
	}
	if s.dynamicGRPC, err = dyngrpc.NewClient(c.DynamicGRPC); err != nil {
		panic(err)
	}
	if s.dynaGRPC, err = dynagrpc.NewClient(c.DynaGRPC); err != nil {
		panic(err)
	}
	if s.popularGRPC, err = pplApi.NewClient(c.PopularGRPC); err != nil {
		panic(err)
	}
	if s.dynCampusClient, err = dyncampusgrpc.NewClient(c.DynamicCampusGRPC); err != nil {
		panic(err)
	}
	return
}
