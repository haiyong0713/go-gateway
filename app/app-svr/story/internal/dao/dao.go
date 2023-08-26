package dao

import (
	"context"

	"go-common/library/conf/paladin.v2"
	"go-common/library/naming/discovery"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/resolver"
	"go-common/library/net/rpc/warden"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	"go-gateway/app/app-svr/app-card/interface/model/card/story"
	appResourcegrpc "go-gateway/app/app-svr/app-resource/interface/api/v1"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/app-svr/story/internal/model"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
	activitygrpc "git.bilibili.co/bapis/bapis-go/activity/service"
	channelgrpc "git.bilibili.co/bapis/bapis-go/community/interface/channel"
	coingrpc "git.bilibili.co/bapis/bapis-go/community/service/coin"
	contractgrpc "git.bilibili.co/bapis/bapis-go/community/service/contract"
	favgrpc "git.bilibili.co/bapis/bapis-go/community/service/favorite"
	locationgrpc "git.bilibili.co/bapis/bapis-go/community/service/location"
	thumbupgrpc "git.bilibili.co/bapis/bapis-go/community/service/thumbup"
	dyngrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/feed"
	liverankgrpc "git.bilibili.co/bapis/bapis-go/live/rankdb/v1"
	livegrpc "git.bilibili.co/bapis/bapis-go/live/xroom-gate"
	materialgrpc "git.bilibili.co/bapis/bapis-go/material/interface"
	pgcinline "git.bilibili.co/bapis/bapis-go/pgc/service/card/inline"
	pgcstory "git.bilibili.co/bapis/bapis-go/pgc/service/card/story"
	pgcFollowClient "git.bilibili.co/bapis/bapis-go/pgc/service/follow"
	pgcfollowgrpc "git.bilibili.co/bapis/bapis-go/pgc/service/follow"
	topicgrpc "git.bilibili.co/bapis/bapis-go/topic/service"
	uparcgrpc "git.bilibili.co/bapis/bapis-go/up-archive/service"
	vogrpc "git.bilibili.co/bapis/bapis-go/videoup/open/service"
	"github.com/google/wire"
)

var Provider = wire.NewSet(New)

// Dao dao interface
type Dao interface {
	Close()
	Ping(ctx context.Context) (err error)
	// rcmd
	StoryRcmd(c context.Context, plat int8, build, pull int, buvid string, mid, aid, adResource int64, displayID int,
		storyParam, adExtra string, zone *locationgrpc.InfoReply, mobiApp, network string, feedStatus, fromAvID int64,
		fromTrackId string, disableRcmd, requestFrom int) (sv *ai.StoryView, respCode int, err error)
	StoryRcmdBackup(ctx context.Context) ([]int64, error)
	RecommendHot(c context.Context) (rs map[int64]struct{}, err error)
	// archive
	ArcsPlayer(ctx context.Context, aids []int64, from string, need1080plus bool) (res map[int64]*arcgrpc.ArcPlayer, err error)
	ArcPassedStory(ctx context.Context, in *uparcgrpc.ArcPassedStoryReq) (*uparcgrpc.ArcPassedStoryReply, error)
	Archives(c context.Context, aids []int64, mid int64, mobiApp, device string) (am map[int64]*arcgrpc.Arc, err error)
	// activity
	StoryLiveReserveKeyExists(c context.Context, req *appResourcegrpc.CheckEntranceInfocRequest) (bool, error)
	StoryLiveReserveCard(c context.Context, arg *activitygrpc.UpActReserveRelationInfo4LiveReq) (*activitygrpc.UpActReserveRelationInfo, error)
	//channel
	ResourceChannels(c context.Context, aids []int64, mid int64) (res map[int64][]*channelgrpc.Channel, err error)
	// thumbup
	HasLike(c context.Context, buvid string, mid int64, messageIDs []int64) (res map[int64]int8, err error)
	UserLikedCounts(c context.Context, mids []int64) (upCounts map[int64]int64, err error)
	MultiLikeAnimation(ctx context.Context, aids []int64) (map[int64]*thumbupgrpc.LikeAnimation, error)
	// creative
	Arguments(ctx context.Context, aids []int64) (map[int64]*vogrpc.Argument, error)
	StoryTagList(ctx context.Context, arg []*materialgrpc.StoryReq) (map[string]*materialgrpc.StoryRes, error)
	// topic
	TopicStory(ctx context.Context, arg *topicgrpc.VideoStoryReq) (*topicgrpc.VideoStoryRsp, error)
	// fav
	IsFavVideos(c context.Context, mid int64, aids []int64) (res map[int64]int8, err error)
	IsFavEp(ctx context.Context, mid int64, epids []int64) (map[int64]int8, error)
	// coin
	ArchiveUserCoins(ctx context.Context, aids []int64, mid int64) (res map[int64]int64, err error)
	// live
	LiveRoomInfos(ctx context.Context, req *livegrpc.EntryRoomInfoReq) (map[int64]*livegrpc.EntryRoomInfoResp_EntryList, error)
	LiveHotRank(ctx context.Context, ids []int64) (map[int64]*liverankgrpc.IsInHotRankResp_HotRankData, error)
	// pgc
	InlineCards(c context.Context, epIDs []int32, mobiApp, platform, device string, build int, mid int64, needHe bool, buvid string, heInlineReq []*pgcinline.HeInlineReq) (map[int32]*pgcinline.EpisodeCard, error)
	OgvPlaylist(ctx context.Context, arg *pgcstory.StoryPlayListReq) (*pgcstory.StoryPlayListReply, error)
	StatusByMid(c context.Context, mid int64, SeasonIDs []int32) (map[int32]*pgcfollowgrpc.FollowStatusProto, error)
	// account
	Cards3GRPC(c context.Context, mids []int64) (res map[int64]*accountgrpc.Card, err error)
	CheckRegTime(ctx context.Context, req *accountgrpc.CheckRegTimeReq) bool
	Card3(ctx context.Context, mid int64) (*accountgrpc.Card, error)
	// relation
	StatsGRPC(ctx context.Context, mids []int64) (res map[int64]*relationgrpc.StatReply, err error)
	RelationsInterrelations(ctx context.Context, mid int64, fids []int64) (res map[int64]*relationgrpc.InterrelationReply, err error)
	// location
	InfoGRPC(c context.Context, ipaddr string) (info *locationgrpc.InfoReply, err error)
	// search
	ArcSpaceSearch(ctx context.Context, arg *model.ArcSearchParam) (*model.ArcSearchReply, int64, error)
	// dynamic
	DynamicGeneralStory(ctx context.Context, param *dyngrpc.GeneralStoryReq) (*dyngrpc.GeneralStoryRsp, error)
	DynamicSpaceStory(ctx context.Context, param *dyngrpc.SpaceStoryReq) (*dyngrpc.SpaceStoryRsp, error)
	DynamicInsert(ctx context.Context, param *dyngrpc.InsertedStoryReq) (*dyngrpc.InsertedStoryRsp, error)
	// ad
	StoryCart(ctx context.Context, param *model.StoryCartParam) (*model.StoryCartReply, error)
	// game
	GameGifts(ctx context.Context, param *model.StoryGameParam) (*model.StoryGameReply, error)
	// community contract
	ContractShowConfig(ctx context.Context, aids []int64, mid int64) (map[int64]*story.ContractResource, error)
}

// dao dao.
type dao struct {
	ac *paladin.Map

	httpClientCfg *HttpClient
	grpcClientCfg *GrpcClient

	client       *bm.Client
	clientAsyn   *bm.Client
	searchClient *bm.Client
	adClient     *bm.Client
	gameClient   *bm.Client

	archiveClient     arcgrpc.ArchiveClient
	actClient         activitygrpc.ActivityClient
	appResourceClient appResourcegrpc.AppResourceClient
	channelClient     channelgrpc.ChannelRPCClient
	thumbupClient     thumbupgrpc.ThumbupClient
	voClient          vogrpc.VideoUpOpenClient
	materialClient    materialgrpc.MaterialClient
	topicClient       topicgrpc.TopicClient
	favClient         favgrpc.FavoriteClient
	coinClient        coingrpc.CoinClient
	liveClient        livegrpc.XroomgateClient
	liveRankClient    liverankgrpc.HotRankClient
	pgcinlineClient   pgcinline.InlineCardClient
	pgcStoryClient    pgcstory.StoryClient
	accountClient     accountgrpc.AccountClient
	relationClient    relationgrpc.RelationClient
	locationClient    locationgrpc.LocationClient
	upArcClient       uparcgrpc.UpArchiveClient
	dynamicClient     dyngrpc.FeedClient
	pgcFollowClient   pgcFollowClient.FollowClient
	contractClient    contractgrpc.ContractClient

	rcmd        string
	storyCart   string
	hot         string
	storyBackup string
}

type HttpClient struct {
	// httpAsyn
	HTTPClientAsyn *bm.ClientConfig
	// httpData
	HTTPData *bm.ClientConfig
	// httpData
	HTTPAd     *bm.ClientConfig
	HTTPSearch *bm.ClientConfig
	HTTPGame   *bm.ClientConfig
	Host       *Host
}

type Host struct {
	Data string
	Cm   string
	AI   string
}

type GrpcClient struct {
	// pgc inline grpc
	PGCInline *warden.ClientConfig
	// AccountGRPC grpc
	AccountGRPC *warden.ClientConfig
	// ActivityGRPC grpc
	ActivityClient *warden.ClientConfig
	// RelationGRPC grpc
	RelationGRPC *warden.ClientConfig
	// grpc Archive
	ArchiveGRPC *warden.ClientConfig
	ThumbupGRPC *warden.ClientConfig
	// grpc location
	LocationGRPC *warden.ClientConfig
	// grpc FavClient
	FavClient  *warden.ClientConfig
	CoinClient *warden.ClientConfig
	// grpc live
	LiveGRPC        *warden.ClientConfig
	LiveRankGRPC    *warden.ClientConfig
	PgcFollowClient *warden.ClientConfig
	PgcStoryClient  *warden.ClientConfig
	DynamicGRPC     *warden.ClientConfig
	VideoOpenClient *warden.ClientConfig
	TopicClient     *warden.ClientConfig
	MaterialClient  *warden.ClientConfig
	// grpc Channel
	ChannelGRPC       *warden.ClientConfig
	AppResourceClient *warden.ClientConfig
	UpArcClient       *warden.ClientConfig
	//
	contractClient *warden.ClientConfig
}

// New new a dao and return.
func New() (d Dao, cf func(), err error) {
	return newDao()
}

func newDao() (d *dao, cf func(), err error) {
	d = &dao{
		ac: &paladin.TOML{},
	}
	if err = paladin.Watch("application.toml", d.ac); err != nil {
		panic(err)
	}
	if err = d.ac.Get("httpClient").UnmarshalTOML(&d.httpClientCfg); err != nil {
		panic(err)
	}
	if err = d.ac.Get("grpcClient").UnmarshalTOML(&d.grpcClientCfg); err != nil {
		panic(err)
	}
	d.newHTTPClient()
	if err = d.newGRPCClient(); err != nil {
		panic(err)
	}

	cf = d.Close
	return
}

func (d *dao) newHTTPClient() {
	d.client = bm.NewClient(d.httpClientCfg.HTTPData, bm.SetResolver(resolver.New(nil, discovery.Builder())))
	d.clientAsyn = bm.NewClient(d.httpClientCfg.HTTPClientAsyn)
	d.searchClient = bm.NewClient(d.httpClientCfg.HTTPSearch)
	d.adClient = bm.NewClient(d.httpClientCfg.HTTPAd, bm.SetResolver(resolver.New(nil, discovery.Builder())))
	d.gameClient = bm.NewClient(d.httpClientCfg.HTTPGame)
	d.rcmd = d.httpClientCfg.Host.Data + _recommand
	d.storyCart = d.httpClientCfg.Host.Cm + _storyCart
	d.hot = d.httpClientCfg.Host.AI + _hot
	d.storyBackup = d.httpClientCfg.Host.AI + _storyBackup
}

func (d *dao) newGRPCClient() (err error) {
	if d.actClient, err = activitygrpc.NewClient(d.grpcClientCfg.ActivityClient); err != nil {
		return err
	}
	if d.appResourceClient, err = appResourcegrpc.NewClient(d.grpcClientCfg.AppResourceClient); err != nil {
		return err
	}
	if d.archiveClient, err = arcgrpc.NewClient(d.grpcClientCfg.ArchiveGRPC); err != nil {
		return err
	}
	if d.channelClient, err = channelgrpc.NewClient(d.grpcClientCfg.ChannelGRPC); err != nil {
		return err
	}
	if d.thumbupClient, err = thumbupgrpc.NewClient(d.grpcClientCfg.ThumbupGRPC); err != nil {
		return err
	}
	if d.voClient, err = vogrpc.NewClient(d.grpcClientCfg.VideoOpenClient); err != nil {
		return err
	}
	if d.materialClient, err = materialgrpc.NewClient(d.grpcClientCfg.MaterialClient); err != nil {
		return err
	}
	if d.topicClient, err = topicgrpc.NewClient(d.grpcClientCfg.TopicClient); err != nil {
		return err
	}
	if d.favClient, err = favgrpc.NewClient(d.grpcClientCfg.FavClient); err != nil {
		return err
	}
	if d.coinClient, err = coingrpc.NewClient(d.grpcClientCfg.CoinClient); err != nil {
		return err
	}
	if d.liveClient, err = livegrpc.NewClientXroomgate(d.grpcClientCfg.LiveGRPC); err != nil {
		return err
	}
	if d.liveRankClient, err = liverankgrpc.NewClientHotRank(d.grpcClientCfg.LiveRankGRPC); err != nil {
		return err
	}
	if d.pgcinlineClient, err = pgcinline.NewClient(d.grpcClientCfg.PGCInline); err != nil {
		return err
	}
	if d.pgcStoryClient, err = pgcstory.NewClientStory(d.grpcClientCfg.PgcStoryClient); err != nil {
		return err
	}
	if d.accountClient, err = accountgrpc.NewClient(d.grpcClientCfg.AccountGRPC); err != nil {
		return err
	}
	if d.relationClient, err = relationgrpc.NewClient(d.grpcClientCfg.RelationGRPC); err != nil {
		return err
	}
	if d.locationClient, err = locationgrpc.NewClient(d.grpcClientCfg.LocationGRPC); err != nil {
		return err
	}
	if d.upArcClient, err = uparcgrpc.NewClient(d.grpcClientCfg.UpArcClient); err != nil {
		return err
	}
	if d.dynamicClient, err = dyngrpc.NewClient(d.grpcClientCfg.DynamicGRPC); err != nil {
		return err
	}
	if d.pgcFollowClient, err = pgcfollowgrpc.NewClient(d.grpcClientCfg.PgcFollowClient); err != nil {
		return err
	}
	if d.contractClient, err = contractgrpc.NewClient(d.grpcClientCfg.contractClient); err != nil {
		return err
	}
	return nil
}

// Close close the resource.
func (d *dao) Close() {
}

// Ping ping the resource.
func (d *dao) Ping(ctx context.Context) (err error) {
	return nil
}
