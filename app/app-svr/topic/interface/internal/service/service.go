package service

import (
	"context"

	"go-common/library/conf/paladin.v2"
	infocV2 "go-common/library/log/infoc.v2"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/rpc/warden"

	archivegrpc "go-gateway/app/app-svr/archive/service/api"
	api "go-gateway/app/app-svr/topic/interface/api"
	"go-gateway/app/app-svr/topic/interface/internal/dao"

	accgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
	activitygrpc "git.bilibili.co/bapis/bapis-go/activity/service"
	articlegrpc "git.bilibili.co/bapis/bapis-go/article/service"
	cmtGrpc "git.bilibili.co/bapis/bapis-go/community/interface/reply"
	coingrpc "git.bilibili.co/bapis/bapis-go/community/service/coin"
	favgrpc "git.bilibili.co/bapis/bapis-go/community/service/favorite"
	thumgrpc "git.bilibili.co/bapis/bapis-go/community/service/thumbup"
	dynactivitygrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/activity"
	dyndrawgrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/draw"
	dyngrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/feed"
	dynamicrevs "git.bilibili.co/bapis/bapis-go/dynamic/service/revs"
	dyntopicapi "git.bilibili.co/bapis/bapis-go/dynamic/service/topic"
	topicextgrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/topic-ext"
	dynvotegrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/vote"
	livexroom "git.bilibili.co/bapis/bapis-go/live/xroom"
	livexroomgate "git.bilibili.co/bapis/bapis-go/live/xroom-gate"
	pgrpc "git.bilibili.co/bapis/bapis-go/manager/service/popular"
	natpagegrpc "git.bilibili.co/bapis/bapis-go/natpage/interface/service"
	esportsgrpc "git.bilibili.co/bapis/bapis-go/operational/esportsservice"
	pgcinline "git.bilibili.co/bapis/bapis-go/pgc/service/card/inline"
	pgcDynGrpc "git.bilibili.co/bapis/bapis-go/pgc/service/dynamic"
	pgcEpisodeGrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/episode"
	pgcSeasonGrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/season"
	grpcShortURL "git.bilibili.co/bapis/bapis-go/platform/interface/shorturl"
	playurlgrpc "git.bilibili.co/bapis/bapis-go/playurl/service"
	topicgrpc "git.bilibili.co/bapis/bapis-go/topic/service"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/google/wire"
	"github.com/pkg/errors"
)

var Provider = wire.NewSet(New, wire.Bind(new(api.TopicServer), new(*Service)))

const (
	_epListURL       = "/pgc/internal/dynamic/v2/ep/list"
	_decorateCardURL = "/x/internal/garb/user/card/multi"
	_emojiURL        = "/x/internal/emote/by/text"
	_dynCommonBiz    = "/common_biz/v0/common_biz/fetch_biz"
	_gameInfo        = "/game/multi_get_game_info"
	_cheeseCard      = "/pugv/internal/dynamic/attach/card"
	_goodsDetails    = "/dwp/api/openApi/v1/window/get"
)

// Service service.
type Service struct {
	ac                  *paladin.Map
	customConfig        *CustomConfig
	pubInfocv2          infocV2.Infoc
	dao                 dao.Dao
	topicGRPC           topicgrpc.TopicClient
	accGRPC             accgrpc.AccountClient
	dynGRPC             dyngrpc.FeedClient
	pgcDynGRPC          pgcDynGrpc.DynamicServiceClient
	livexroomGateGRPC   livexroomgate.XroomgateClient
	livexroomGRPC       livexroom.RoomClient
	dynDrawGRPC         dyndrawgrpc.DrawClient
	dynamicActivityGRPC dynactivitygrpc.ActPromoRPCClient
	archiveGRPC         archivegrpc.ArchiveClient
	cmtGrpc             cmtGrpc.ReplyInterfaceClient
	thumbGRPC           thumgrpc.ThumbupClient
	natPageGrpcClient   natpagegrpc.NaPageClient
	relGRPC             relationgrpc.RelationClient
	shortURLGRPC        grpcShortURL.ShortUrlClient
	articleGRPC         articlegrpc.ArticleGRPCClient
	pgcSeasonGRPC       pgcSeasonGrpc.SeasonClient
	pgcEpisodeGRPC      pgcEpisodeGrpc.EpisodeClient
	dynVoteGRPC         dynvotegrpc.VoteSvrClient
	dynTopicGRPC        dyntopicapi.TopicClient
	topicExtGRPC        topicextgrpc.TopicExtClient
	dynRevGRPC          dynamicrevs.RevsClient
	pgcInlineGRPC       pgcinline.InlineCardClient
	favGRPC             favgrpc.FavoriteClient
	coinGRPC            coingrpc.CoinClient
	actClient           activitygrpc.ActivityClient
	managerPopClient    pgrpc.PopularClient
	playurlGRPC         playurlgrpc.PlayURLClient
	esportGRPC          esportsgrpc.EsportsServiceClient
	roomGateClient      livexroomgate.XroomgateClient
	httpMgr             *bm.Client
	httpGameCo          *bm.Client
	epList              string
	decorateCards       string
	emojiURL            string
	dynCommonBiz        string // 动态通用模板信息路由
	gameMultiInfos      string // 批量获取游戏数据
	attachCheeseCard    string // 课程信息卡片
	goodsDetailsURL     string // 商品卡片
}

type CustomConfig struct {
	PubInfoc           *InfocConf
	TopicServiceConfig *TopicServiceConfig
}

type TopicServiceConfig struct {
	PubEventsIncreaseThreshold      int64
	PubEventsHiddenTimeoutThreshold int64
	VertOnlineRefreshTime           int64
}

type InfocConf struct {
	LogID string
	Infoc *infocV2.Config
}

// New new a service and return.
func New(d dao.Dao) (s *Service, cf func(), err error) {
	type Host struct {
		ApiCo  string
		VcAPI  string
		GameCo string
		CmCom  string
	}
	var cfg struct {
		TopicGRPC           *warden.ClientConfig
		AccountGRPC         *warden.ClientConfig
		DynamicGRPC         *warden.ClientConfig
		PgcDynGRPC          *warden.ClientConfig
		LivexroomGRPC       *warden.ClientConfig
		DynDrawGRPC         *warden.ClientConfig
		DynamicActivityGRPC *warden.ClientConfig
		ArchiveGRPC         *warden.ClientConfig
		CmtGrpc             *warden.ClientConfig
		ThumbGRPC           *warden.ClientConfig
		NatPageGRPC         *warden.ClientConfig
		RelGRPC             *warden.ClientConfig
		ShortURLGRPC        *warden.ClientConfig
		ArticleGRPC         *warden.ClientConfig
		PgcSeasonGRPC       *warden.ClientConfig
		PgcEpisodeGRPC      *warden.ClientConfig
		DynVoteGRPC         *warden.ClientConfig
		DynTopicClient      *warden.ClientConfig
		TopicExtGRPC        *warden.ClientConfig
		DynRevsGRPC         *warden.ClientConfig
		PgcInlineGRPC       *warden.ClientConfig
		FavGRPC             *warden.ClientConfig
		CoinGRPC            *warden.ClientConfig
		ActivityGRPC        *warden.ClientConfig
		ManagerPopGRPC      *warden.ClientConfig
		PlayUrlGRPC         *warden.ClientConfig
		EsportGRPC          *warden.ClientConfig
		LiveGRPC            *warden.ClientConfig
		HTTPClient          *bm.ClientConfig
		HTTPGameCo          *bm.ClientConfig
		Host                *Host
	}
	if err = paladin.Get("application.toml").UnmarshalTOML(&cfg); err != nil {
		err = errors.WithStack(err)
		return
	}
	s = &Service{
		ac:               &paladin.TOML{},
		customConfig:     new(CustomConfig),
		dao:              d,
		epList:           cfg.Host.ApiCo + _epListURL,
		decorateCards:    cfg.Host.ApiCo + _decorateCardURL,
		emojiURL:         cfg.Host.ApiCo + _emojiURL,
		dynCommonBiz:     cfg.Host.VcAPI + _dynCommonBiz,
		gameMultiInfos:   cfg.Host.GameCo + _gameInfo,
		attachCheeseCard: cfg.Host.ApiCo + _cheeseCard,
		goodsDetailsURL:  cfg.Host.CmCom + _goodsDetails,
	}
	if err = paladin.Watch("application.toml", s.ac); err != nil {
		return
	}
	if err = s.ac.Get("customConfig").UnmarshalTOML(&s.customConfig); err != nil {
		panic(err)
	}
	if s.pubInfocv2, err = infocV2.New(s.customConfig.PubInfoc.Infoc); err != nil {
		panic(err)
	}
	if s.topicGRPC, err = topicgrpc.NewClient(cfg.TopicGRPC); err != nil {
		err = errors.WithStack(err)
		return
	}
	if s.accGRPC, err = accgrpc.NewClient(cfg.AccountGRPC); err != nil {
		err = errors.WithStack(err)
		return
	}
	if s.dynGRPC, err = dyngrpc.NewClient(cfg.DynamicGRPC); err != nil {
		err = errors.WithStack(err)
		return
	}
	if s.pgcDynGRPC, err = pgcDynGrpc.NewClient(cfg.PgcDynGRPC); err != nil {
		err = errors.WithStack(err)
		return
	}
	if s.livexroomGateGRPC, err = livexroomgate.NewClientXroomgate(cfg.LivexroomGRPC); err != nil {
		err = errors.WithStack(err)
		return
	}
	if s.livexroomGRPC, err = livexroom.NewClient(cfg.LivexroomGRPC); err != nil {
		err = errors.WithStack(err)
		return
	}
	if s.dynDrawGRPC, err = dyndrawgrpc.NewClient(cfg.DynDrawGRPC); err != nil {
		err = errors.WithStack(err)
		return
	}
	if s.dynamicActivityGRPC, err = dynactivitygrpc.NewClient(cfg.DynamicActivityGRPC); err != nil {
		err = errors.WithStack(err)
		return
	}
	if s.archiveGRPC, err = archivegrpc.NewClient(cfg.ArchiveGRPC); err != nil {
		err = errors.WithStack(err)
		return
	}
	if s.cmtGrpc, err = cmtGrpc.NewClient(cfg.CmtGrpc); err != nil {
		err = errors.WithStack(err)
		return
	}
	if s.thumbGRPC, err = thumgrpc.NewClient(cfg.ThumbGRPC); err != nil {
		err = errors.WithStack(err)
		return
	}
	if s.natPageGrpcClient, err = natpagegrpc.NewClient(cfg.NatPageGRPC); err != nil {
		err = errors.WithStack(err)
		return
	}
	if s.relGRPC, err = relationgrpc.NewClient(cfg.RelGRPC); err != nil {
		err = errors.WithStack(err)
		return
	}
	if s.shortURLGRPC, err = grpcShortURL.NewClient(cfg.ShortURLGRPC); err != nil {
		err = errors.WithStack(err)
		return
	}
	if s.articleGRPC, err = articlegrpc.NewClient(cfg.ArticleGRPC); err != nil {
		err = errors.WithStack(err)
		return
	}
	if s.pgcEpisodeGRPC, err = pgcEpisodeGrpc.NewClient(cfg.PgcEpisodeGRPC); err != nil {
		err = errors.WithStack(err)
		return
	}
	if s.pgcSeasonGRPC, err = pgcSeasonGrpc.NewClient(cfg.PgcSeasonGRPC); err != nil {
		err = errors.WithStack(err)
		return
	}
	if s.dynVoteGRPC, err = dynvotegrpc.NewClient(cfg.DynVoteGRPC); err != nil {
		err = errors.WithStack(err)
		return
	}
	if s.dynTopicGRPC, err = dyntopicapi.NewClient(cfg.DynTopicClient); err != nil {
		err = errors.WithStack(err)
		return
	}
	if s.topicExtGRPC, err = topicextgrpc.NewClient(cfg.TopicExtGRPC); err != nil {
		err = errors.WithStack(err)
		return
	}
	if s.dynRevGRPC, err = dynamicrevs.NewClient(cfg.TopicExtGRPC); err != nil {
		err = errors.WithStack(err)
		return
	}
	if s.pgcInlineGRPC, err = pgcinline.NewClient(cfg.PgcInlineGRPC); err != nil {
		err = errors.WithStack(err)
		return
	}
	if s.favGRPC, err = favgrpc.NewClient(cfg.FavGRPC); err != nil {
		err = errors.WithStack(err)
		return
	}
	if s.coinGRPC, err = coingrpc.NewClient(cfg.CoinGRPC); err != nil {
		err = errors.WithStack(err)
		return
	}
	if s.actClient, err = activitygrpc.NewClient(cfg.ActivityGRPC); err != nil {
		err = errors.WithStack(err)
		return
	}
	if s.managerPopClient, err = pgrpc.NewClient(cfg.ManagerPopGRPC); err != nil {
		err = errors.WithStack(err)
		return
	}
	if s.playurlGRPC, err = playurlgrpc.NewClient(cfg.PlayUrlGRPC); err != nil {
		err = errors.WithStack(err)
		return
	}
	if s.esportGRPC, err = esportsgrpc.NewClient(cfg.EsportGRPC); err != nil {
		err = errors.WithStack(err)
		return
	}
	if s.roomGateClient, err = livexroomgate.NewClientXroomgate(cfg.LiveGRPC); err != nil {
		err = errors.WithStack(err)
		return
	}
	s.httpMgr = bm.NewClient(cfg.HTTPClient)
	s.httpGameCo = bm.NewClient(cfg.HTTPGameCo)
	cf = s.Close
	return
}

// Ping ping the resource.
func (s *Service) Ping(ctx context.Context, _ *empty.Empty) (*empty.Empty, error) {
	return &empty.Empty{}, s.dao.Ping(ctx)
}

// Close close the resource.
func (s *Service) Close() {
}
