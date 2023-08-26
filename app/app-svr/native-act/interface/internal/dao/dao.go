package dao

import (
	"context"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	relationGrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
	activitygrpc "git.bilibili.co/bapis/bapis-go/activity/service"
	appshowgrpc "git.bilibili.co/bapis/bapis-go/app/show/v1"
	articlegrpc "git.bilibili.co/bapis/bapis-go/article/service"
	channelgrpc "git.bilibili.co/bapis/bapis-go/community/interface/channel"
	taggrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	favoritegrpc "git.bilibili.co/bapis/bapis-go/community/service/favorite"
	commscoregrpc "git.bilibili.co/bapis/bapis-go/community/service/score"
	dynfeedgrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/feed"
	dyntopicgrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/topic"
	dynvotegrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/vote"
	hmtgrpc "git.bilibili.co/bapis/bapis-go/hmt-channel/interface"
	liveplaygrpc "git.bilibili.co/bapis/bapis-go/live/live-play/v1"
	xroomfeedgrpc "git.bilibili.co/bapis/bapis-go/live/xroom-feed"
	roomgategrpc "git.bilibili.co/bapis/bapis-go/live/xroom-gate"
	populargrpc "git.bilibili.co/bapis/bapis-go/manager/service/popular"
	pgcappgrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card/app"
	pgcfollowgrpc "git.bilibili.co/bapis/bapis-go/pgc/service/follow"
	chargrpc "git.bilibili.co/bapis/bapis-go/pgc/service/media"
	actplatv2grpc "git.bilibili.co/bapis/bapis-go/platform/interface/act-plat-v2"
	"github.com/google/wire"
	"go-common/library/conf/paladin.v2"
	"go-common/library/naming/discovery"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/resolver"
	"go-common/library/net/rpc/warden"
	"go-common/library/sync/pipeline/fanout"

	appdyngrpc "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	natpagegrpc "go-gateway/app/web-svr/native-page/interface/api"
)

var Provider = wire.NewSet(New)

// Dao dao interface
//
//go:generate kratos tool btsgen
type Dao interface {
	Dependency

	Close()
	Ping(ctx context.Context) (err error)
}

// dao dao.
type dao struct {
	cache *fanout.Fanout
	cfg   Config

	httpClient          *bm.Client
	gameHttpClient      *bm.Client
	businessHttpClient  *bm.Client
	mangaHttpClient     *bm.Client
	showHttpClient      *bm.Client
	natpageClient       natpagegrpc.NaPageClient
	archiveClient       arcgrpc.ArchiveClient
	livePlayClient      liveplaygrpc.TopicClient
	articleClient       articlegrpc.ArticleGRPCClient
	favoriteClient      favoritegrpc.FavoriteClient
	activityClient      activitygrpc.ActivityClient
	accountClient       accountgrpc.AccountClient
	dyntopicClient      dyntopicgrpc.TopicClient
	appdynClient        appdyngrpc.DynamicClient
	dynfeedClient       dynfeedgrpc.FeedClient
	tagClient           taggrpc.TagRPCClient
	actplatv2Client     actplatv2grpc.ActPlatClient
	popularClient       populargrpc.PopularClient
	appshowClient       appshowgrpc.AppShowClient
	liveXRoomFeedClient xroomfeedgrpc.DynamicClient
	characterClient     chargrpc.CharacterClient
	pgcappClient        pgcappgrpc.AppCardClient
	hmtClient           hmtgrpc.ChannelRPCClient
	relationClient      relationGrpc.RelationClient
	channelClient       channelgrpc.ChannelRPCClient
	pgcfollowClient     pgcfollowgrpc.FollowClient
	dynvoteClient       dynvotegrpc.VoteSvrClient
	roomGateClient      roomgategrpc.XroomgateClient
	commScoreClient     commscoregrpc.ScoreClient
}

type Config struct {
	Host struct {
		ApiCo    string
		GameCo   string
		ApiVcCo  string
		Business string
		Manga    string
		Show     string
	}
}

// New new a dao and return.
func New() (d Dao, cf func(), err error) {
	return newDao()
}

func newDao() (d *dao, cf func(), err error) {
	appCfg := Config{}
	if err = paladin.Get("application.toml").UnmarshalTOML(&appCfg); err != nil {
		return
	}
	d = &dao{
		cache: fanout.New("cache"),
		cfg:   appCfg,
	}
	newHTTPClient(d)
	newGRPCClient(d)
	cf = d.Close
	return
}

func newHTTPClient(d *dao) {
	cfg := struct {
		HTTPClient         *bm.ClientConfig
		HTTPGameClient     *bm.ClientConfig
		HTTPBusinessClient *bm.ClientConfig
		HTTPMangaClient    *bm.ClientConfig
		HTTPShowClient     *bm.ClientConfig
	}{}
	if err := paladin.Get("http.toml").UnmarshalTOML(&cfg); err != nil {
		panic(err)
	}
	d.httpClient = bm.NewClient(cfg.HTTPClient, bm.SetResolver(resolver.New(nil, discovery.Builder())))
	d.gameHttpClient = bm.NewClient(cfg.HTTPGameClient)
	d.businessHttpClient = bm.NewClient(cfg.HTTPBusinessClient)
	d.mangaHttpClient = bm.NewClient(cfg.HTTPMangaClient)
	d.showHttpClient = bm.NewClient(cfg.HTTPShowClient)
}

type grpcCfg struct {
	Natpage       *warden.ClientConfig
	Archive       *warden.ClientConfig
	LivePlay      *warden.ClientConfig
	Article       *warden.ClientConfig
	Favorite      *warden.ClientConfig
	Activity      *warden.ClientConfig
	Account       *warden.ClientConfig
	Dyntopic      *warden.ClientConfig
	Appdyn        *warden.ClientConfig
	Dynfeed       *warden.ClientConfig
	Tag           *warden.ClientConfig
	Actplatv2     *warden.ClientConfig
	Popular       *warden.ClientConfig
	Appshow       *warden.ClientConfig
	LiveXRoomFeed *warden.ClientConfig
	Character     *warden.ClientConfig
	Pgcapp        *warden.ClientConfig
	Hmt           *warden.ClientConfig
	Relation      *warden.ClientConfig
	Channel       *warden.ClientConfig
	Pgcfollow     *warden.ClientConfig
	Dynvote       *warden.ClientConfig
	RoomGate      *warden.ClientConfig
	Commsocre     *warden.ClientConfig
}

func newGRPCClient(d *dao) {
	var err error
	cfg := grpcCfg{}
	if err = paladin.Get("grpc.toml").UnmarshalTOML(&cfg); err != nil {
		panic(err)
	}
	if d.natpageClient, err = natpagegrpc.NewClient(cfg.Natpage); err != nil {
		panic(err)
	}
	if d.archiveClient, err = arcgrpc.NewClient(cfg.Archive); err != nil {
		panic(err)
	}
	if d.livePlayClient, err = liveplaygrpc.NewClient(cfg.LivePlay); err != nil {
		panic(err)
	}
	if d.articleClient, err = articlegrpc.NewClient(cfg.Article); err != nil {
		panic(err)
	}
	if d.favoriteClient, err = favoritegrpc.NewClient(cfg.Favorite); err != nil {
		panic(err)
	}
	if d.activityClient, err = activitygrpc.NewClient(cfg.Activity); err != nil {
		panic(err)
	}
	if d.accountClient, err = accountgrpc.NewClient(cfg.Account); err != nil {
		panic(err)
	}
	if d.dyntopicClient, err = dyntopicgrpc.NewClient(cfg.Dyntopic); err != nil {
		panic(err)
	}
	if d.appdynClient, err = appdyngrpc.NewClient(cfg.Appdyn); err != nil {
		panic(err)
	}
	if d.dynfeedClient, err = dynfeedgrpc.NewClient(cfg.Dynfeed); err != nil {
		panic(err)
	}
	if d.tagClient, err = taggrpc.NewClient(cfg.Tag); err != nil {
		panic(err)
	}
	if d.actplatv2Client, err = actplatv2grpc.NewClient(cfg.Actplatv2); err != nil {
		panic(err)
	}
	if d.popularClient, err = populargrpc.NewClient(cfg.Popular); err != nil {
		panic(err)
	}
	if d.appshowClient, err = appshowgrpc.NewClient(cfg.Appshow); err != nil {
		panic(err)
	}
	if d.liveXRoomFeedClient, err = xroomfeedgrpc.NewClient(cfg.LiveXRoomFeed); err != nil {
		panic(err)
	}
	if d.characterClient, err = chargrpc.NewClientCharacter(cfg.Character); err != nil {
		panic(err)
	}
	if d.pgcappClient, err = pgcappgrpc.NewClient(cfg.Pgcapp); err != nil {
		panic(err)
	}
	if d.hmtClient, err = hmtgrpc.NewClient(cfg.Hmt); err != nil {
		panic(err)
	}
	if d.relationClient, err = relationGrpc.NewClient(cfg.Relation); err != nil {
		panic(err)
	}
	if d.channelClient, err = channelgrpc.NewClient(cfg.Channel); err != nil {
		panic(err)
	}
	if d.pgcfollowClient, err = pgcfollowgrpc.NewClient(cfg.Pgcfollow); err != nil {
		panic(err)
	}
	if d.dynvoteClient, err = dynvotegrpc.NewClient(cfg.Dynvote); err != nil {
		panic(err)
	}
	if d.roomGateClient, err = roomgategrpc.NewClientXroomgate(cfg.RoomGate); err != nil {
		panic(err)
	}
	if d.commScoreClient, err = commscoregrpc.NewClient(cfg.Commsocre); err != nil {
		panic(err)
	}
}

type Dependency interface {
	Natpage() *natpageDao
	Archive() *archiveDao
	LivePlay() *livePlayDao
	Article() *articleDao
	Bangumi() *bangumiDao
	Game() *gameDao
	Favorite() *favoriteDao
	Activity() *activityDao
	Account() *accountDao
	Dyntopic() *dyntopicDao
	Appdyn() *appdynDao
	Dynfeed() *dynfeedDao
	Tag() *tagDao
	Actplatv2() *actplatv2Dao
	Popular() *popularDao
	Appshow() *appshowDao
	LiveXRoomFeed() *liveXRoomFeedDao
	Character() *characterDao
	Pgcapp() *pgcappDao
	Hmt() *hmtDao
	Relation() *relationDao
	Channel() *channelDao
	Pgcfollow() *pgcfollowDao
	Dynvote() *dynvoteDao
	Business() *businessDao
	RoomGate() *roomGateDao
	Comic() *comicDao
	Mallticket() *mallticketDao
	Commscore() *commscoreDao
}

func (d *dao) LiveXRoomFeed() *liveXRoomFeedDao {
	return &liveXRoomFeedDao{client: d.liveXRoomFeedClient}
}

func (d *dao) Natpage() *natpageDao {
	return &natpageDao{client: d.natpageClient}
}

func (d *dao) Archive() *archiveDao {
	return &archiveDao{client: d.archiveClient}
}

func (d *dao) LivePlay() *livePlayDao {
	return &livePlayDao{client: d.livePlayClient}
}

func (d *dao) Article() *articleDao {
	return &articleDao{client: d.articleClient}
}

func (d *dao) Bangumi() *bangumiDao {
	return &bangumiDao{host: d.cfg.Host.ApiCo, httpClient: d.httpClient}
}

func (d *dao) Game() *gameDao {
	return &gameDao{host: d.cfg.Host.GameCo, httpClient: d.gameHttpClient}
}

func (d *dao) Favorite() *favoriteDao {
	return &favoriteDao{client: d.favoriteClient}
}

func (d *dao) Activity() *activityDao {
	return &activityDao{client: d.activityClient}
}

func (d *dao) Account() *accountDao {
	return &accountDao{client: d.accountClient}
}

func (d *dao) Dyntopic() *dyntopicDao {
	return &dyntopicDao{client: d.dyntopicClient, host: d.cfg.Host.ApiVcCo, httpClient: d.httpClient}
}

func (d *dao) Appdyn() *appdynDao {
	return &appdynDao{client: d.appdynClient}
}

func (d *dao) Dynfeed() *dynfeedDao {
	return &dynfeedDao{client: d.dynfeedClient}
}

func (d *dao) Tag() *tagDao {
	return &tagDao{client: d.tagClient}
}

func (d *dao) Actplatv2() *actplatv2Dao {
	return &actplatv2Dao{client: d.actplatv2Client}
}

func (d *dao) Popular() *popularDao {
	return &popularDao{client: d.popularClient}
}

func (d *dao) Appshow() *appshowDao {
	return &appshowDao{client: d.appshowClient}
}

func (d *dao) Character() *characterDao {
	return &characterDao{client: d.characterClient}
}

func (d *dao) Pgcapp() *pgcappDao {
	return &pgcappDao{client: d.pgcappClient}
}

func (d *dao) Hmt() *hmtDao {
	return &hmtDao{client: d.hmtClient}
}

func (d *dao) Relation() *relationDao {
	return &relationDao{client: d.relationClient}
}

func (d *dao) Channel() *channelDao {
	return &channelDao{client: d.channelClient}
}

func (d *dao) Pgcfollow() *pgcfollowDao {
	return &pgcfollowDao{client: d.pgcfollowClient}
}

func (d *dao) Dynvote() *dynvoteDao {
	return &dynvoteDao{client: d.dynvoteClient}
}

func (d *dao) Business() *businessDao {
	return &businessDao{host: d.cfg.Host.Business, httpClient: d.businessHttpClient}
}

func (d *dao) RoomGate() *roomGateDao {
	return &roomGateDao{client: d.roomGateClient}
}

func (d *dao) Comic() *comicDao {
	return &comicDao{host: d.cfg.Host.Manga, httpClient: d.mangaHttpClient}
}

func (d *dao) Mallticket() *mallticketDao {
	return &mallticketDao{host: d.cfg.Host.Show, httpClient: d.showHttpClient}
}

func (d *dao) Commscore() *commscoreDao {
	return &commscoreDao{client: d.commScoreClient}
}

// Close close the resource.
func (d *dao) Close() {
	d.cache.Close()
}

// Ping ping the resource.
func (d *dao) Ping(ctx context.Context) (err error) {
	return nil
}
