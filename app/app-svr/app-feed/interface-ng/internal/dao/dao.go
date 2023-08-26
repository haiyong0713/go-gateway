package dao

import (
	"context"
	"strconv"
	"sync"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	"go-common/library/cache/redis"
	"go-common/library/conf/paladin"
	"go-common/library/naming/discovery"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/resolver"
	"go-common/library/net/rpc"
	"go-common/library/net/rpc/warden"
	"go-common/library/sync/pipeline/fanout"
	"go-gateway/app/app-svr/app-card/interface/model"
	largecover "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/large_cover"
	selectV2 "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/select"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	resourcegrpc "go-gateway/app/app-svr/resource/service/api/v1"
	rscrpc "go-gateway/app/app-svr/resource/service/rpc/client"

	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
	articlegrpc "git.bilibili.co/bapis/bapis-go/article/service"
	channelgrpc "git.bilibili.co/bapis/bapis-go/community/interface/channel"
	taggrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	coingrpc "git.bilibili.co/bapis/bapis-go/community/service/coin"
	favgrpc "git.bilibili.co/bapis/bapis-go/community/service/favorite"
	locationgrpc "git.bilibili.co/bapis/bapis-go/community/service/location"
	thumbupgrpc "git.bilibili.co/bapis/bapis-go/community/service/thumbup"
	pgcinline "git.bilibili.co/bapis/bapis-go/pgc/service/card/inline"
	episodegrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/episode"
	tunnelgrpc "git.bilibili.co/bapis/bapis-go/platform/service/tunnel"
	vipgrpc "git.bilibili.co/bapis/bapis-go/vip/service"
	"github.com/google/wire"
)

var Provider = wire.NewSet(New, NewRedis)

// Dao dao interface
type Dao interface {
	FanoutDependency

	Close()
	Ping(ctx context.Context) (err error)
}

// dao dao.
type dao struct {
	sync.RWMutex
	cfg Config

	redis           *redis.Redis
	fanout          *fanout.Fanout
	httpClient      *bm.Client
	resourceRPC     *rscrpc.Service
	archiveClient   arcgrpc.ArchiveClient
	thumbupClient   thumbupgrpc.ThumbupClient
	channelClient   channelgrpc.ChannelRPCClient
	articleClient   articlegrpc.ArticleGRPCClient
	bangumiClient   episodegrpc.EpisodeClient
	accountClient   accountgrpc.AccountClient
	relationClient  relationgrpc.RelationClient
	inlinePgcClient pgcinline.InlineCardClient
	resourceClient  resourcegrpc.ResourceClient
	vipClient       vipgrpc.VipClient
	tunnelClient    tunnelgrpc.TunnelClient
	locationClient  locationgrpc.LocationClient
	favouriteClient favgrpc.FavoriteClient
	coinClient      coingrpc.CoinClient
	tagGRPCClient   taggrpc.TagRPCClient
}

func (d *dao) dupConfig() Config {
	d.RLock()
	defer d.RUnlock()
	dup := d.cfg
	return dup
}

// Config is
type Config struct {
	Host struct {
		Live          string
		Audio         string
		Dynamic       string
		Bangumi       string
		Shop          string
		Data          string
		BigData       string
		Ad            string
		DataDiscovery string
	}
	Inline    *largecover.Inline
	StoryIcon map[string]*model.GotoIcon
}

// New new a dao and return.
func New(r *redis.Redis) (Dao, func(), error) {
	return newDao(r)
}

func newHTTPClient() *bm.Client {
	cfg := struct {
		HTTPClient *bm.ClientConfig
	}{}
	if err := paladin.Get("http.toml").UnmarshalTOML(&cfg); err != nil {
		panic(err)
	}
	return bm.NewClient(cfg.HTTPClient, bm.SetResolver(resolver.New(nil, discovery.Builder())))
}

func newRPCClient(dst *dao) {
	cfg := struct {
		Archive     *warden.ClientConfig
		Thumbup     *warden.ClientConfig
		Channel     *warden.ClientConfig
		Article     *warden.ClientConfig
		Bangumi     *warden.ClientConfig
		Account     *warden.ClientConfig
		Relation    *warden.ClientConfig
		Tag         *rpc.ClientConfig
		ResourceRPC *rpc.ClientConfig
		InlinePGC   *warden.ClientConfig
		Resource    *warden.ClientConfig
		Vip         *warden.ClientConfig
		Tunnel      *warden.ClientConfig
		Location    *warden.ClientConfig
		Favourite   *warden.ClientConfig
		Coin        *warden.ClientConfig
		TagGRPC     *warden.ClientConfig
	}{}
	if err := paladin.Get("grpc.toml").UnmarshalTOML(&cfg); err != nil {
		panic(err)
	}
	dst.resourceRPC = rscrpc.New(cfg.ResourceRPC)

	archiveClient, err := arcgrpc.NewClient(cfg.Archive)
	if err != nil {
		panic(err)
	}
	dst.archiveClient = archiveClient

	thumbupClient, err := thumbupgrpc.NewClient(cfg.Thumbup)
	if err != nil {
		panic(err)
	}
	dst.thumbupClient = thumbupClient

	channelClient, err := channelgrpc.NewClient(cfg.Channel)
	if err != nil {
		panic(err)
	}
	dst.channelClient = channelClient

	articleClient, err := articlegrpc.NewClient(cfg.Article)
	if err != nil {
		panic(err)
	}
	dst.articleClient = articleClient

	bangumiClient, err := episodegrpc.NewClient(cfg.Bangumi)
	if err != nil {
		panic(err)
	}
	dst.bangumiClient = bangumiClient

	accountClient, err := accountgrpc.NewClient(cfg.Account)
	if err != nil {
		panic(err)
	}
	dst.accountClient = accountClient

	relationClient, err := relationgrpc.NewClient(cfg.Relation)
	if err != nil {
		panic(err)
	}
	dst.relationClient = relationClient
	inlinePGCClient, err := pgcinline.NewClient(cfg.InlinePGC)
	if err != nil {
		panic(err)
	}
	dst.inlinePgcClient = inlinePGCClient
	resourceClient, err := resourcegrpc.NewClient(cfg.Resource)
	if err != nil {
		panic(err)
	}
	dst.resourceClient = resourceClient
	vipClient, err := vipgrpc.NewClient(cfg.Vip)
	if err != nil {
		panic(err)
	}
	dst.vipClient = vipClient
	tunnelClient, err := tunnelgrpc.NewClient(cfg.Tunnel)
	if err != nil {
		panic(err)
	}
	dst.tunnelClient = tunnelClient
	locationClient, err := locationgrpc.NewClient(cfg.Location)
	if err != nil {
		panic(err)
	}
	dst.locationClient = locationClient
	favouriteClient, err := favgrpc.NewClient(cfg.Favourite)
	if err != nil {
		panic(err)
	}
	dst.favouriteClient = favouriteClient
	coinClient, err := coingrpc.NewClient(cfg.Coin)
	if err != nil {
		panic(err)
	}
	dst.coinClient = coinClient
	tagGRPCClient, err := taggrpc.NewClient(cfg.TagGRPC)
	if err != nil {
		panic(err)
	}
	dst.tagGRPCClient = tagGRPCClient
}

func newDao(r *redis.Redis) (*dao, func(), error) {
	appConfig := struct {
		Config Config
	}{}
	if err := paladin.Get("application.toml").UnmarshalTOML(&appConfig); err != nil {
		return nil, nil, err
	}
	d := &dao{
		redis:      r,
		fanout:     fanout.New("cache"),
		httpClient: newHTTPClient(),
		cfg:        appConfig.Config,
	}
	newRPCClient(d)
	closeFn := d.Close
	return d, closeFn, nil
}

// FanoutDependency is
type FanoutDependency interface {
	Archive() *arcDao
	Tag() *tagDao
	Live() *liveDao
	Article() *articleDao
	Audio() *audioDao
	Dynamic() *dynamicDao
	Bangumi() *bangumiDao
	Channel() *channelDao
	ThumbUp() *thumbupDao
	Account() *accountDao
	Relation() *relationDao
	Resource() *resourceDao
	Inline() *largecover.Inline
	FollowMode() *selectV2.FollowMode
	StoryIcon() map[int64]*model.GotoIcon
	Shop() *shopDao
	Vip() *vipDao
	Tunnel() *tunnelDao
	Recommend() *recommendDao
	Location() *locationDao
	Ad() *adDao
	Favourite() *favouriteDao
	Coin() *coinDao
}

func (d *dao) Archive() *arcDao {
	return &arcDao{
		archive: d.archiveClient,
	}
}

func (d *dao) Live() *liveDao {
	cfg := d.dupConfig()
	return &liveDao{
		client: d.httpClient,
		cfg: liveConfig{
			Host: cfg.Host.Live,
		},
	}
}

func (d *dao) ThumbUp() *thumbupDao {
	return &thumbupDao{
		thumbup: d.thumbupClient,
	}
}

func (d *dao) Tag() *tagDao {
	return &tagDao{
		tagGRPC: d.tagGRPCClient,
	}
}

func (d *dao) Channel() *channelDao {
	return &channelDao{
		channel: d.channelClient,
	}
}

func (d *dao) Article() *articleDao {
	return &articleDao{
		article: d.articleClient,
	}
}

func (d *dao) Audio() *audioDao {
	cfg := d.dupConfig()
	return &audioDao{
		client: d.httpClient,
		cfg: audioConfig{
			Host: cfg.Host.Audio,
		},
	}
}

func (d *dao) Dynamic() *dynamicDao {
	cfg := d.dupConfig()
	return &dynamicDao{
		client: d.httpClient,
		cfg: dynamicConfig{
			Host: cfg.Host.Dynamic,
		},
	}
}

func (d *dao) Bangumi() *bangumiDao {
	cfg := d.dupConfig()
	return &bangumiDao{
		bangumi: d.bangumiClient,
		client:  d.httpClient,
		cfg: bangumiConfig{
			Host: cfg.Host.Bangumi,
		},
		pgcinlineClient: d.inlinePgcClient,
	}
}

func (d *dao) Account() *accountDao {
	return &accountDao{
		account: d.accountClient,
	}
}

func (d *dao) Relation() *relationDao {
	return &relationDao{
		relation: d.relationClient,
	}
}

func (d *dao) Resource() *resourceDao {
	return &resourceDao{
		resourceRPC:  d.resourceRPC,
		resourceGRPC: d.resourceClient,
	}
}

func (d *dao) Inline() *largecover.Inline {
	cfg := d.dupConfig()
	return cfg.Inline
}

func (d *dao) FollowMode() *selectV2.FollowMode {
	return &selectV2.FollowMode{
		Title:   "提醒：是否需要开启首页推荐的“关注模式”（内测版）？",
		Desc:    "我们收到多次你对APP的首页推荐反馈“不感兴趣”。目前首页推荐的新功能——“关注模式”正在做小规模测试，如果你选择“关注模式”，首页将只推荐你关注的UP主的视频。请问你是否愿意参与“关注模式”的功能测试？",
		Buttons: []string{"不参加", "我想参加"},
	}
}

func (d *dao) Shop() *shopDao {
	cfg := d.dupConfig()
	return &shopDao{
		client: d.httpClient,
		cfg: shopConfig{
			Host: cfg.Host.Shop,
		},
	}
}

func (d *dao) Vip() *vipDao {
	return &vipDao{
		vip: d.vipClient,
	}
}

func (d *dao) Tunnel() *tunnelDao {
	return &tunnelDao{
		tunnel: d.tunnelClient,
	}
}

func (d *dao) Favourite() *favouriteDao {
	return &favouriteDao{
		favourite: d.favouriteClient,
	}
}

func (d *dao) Recommend() *recommendDao {
	cfg := d.dupConfig()
	return &recommendDao{
		client: d.httpClient,
		cfg: recommendConfig{
			DataDiscoveryHost: cfg.Host.DataDiscovery,
			BigDataHost:       cfg.Host.BigData,
			DataHost:          cfg.Host.Data,
		},
	}
}

func (d *dao) Ad() *adDao {
	cfg := d.dupConfig()
	return &adDao{
		client: d.httpClient,
		cfg: adConfig{
			Host: cfg.Host.Ad,
		},
	}
}

func (d *dao) Location() *locationDao {
	return &locationDao{
		location: d.locationClient,
	}
}

func (d *dao) Coin() *coinDao {
	return &coinDao{
		coin: d.coinClient,
	}
}

func (d *dao) StoryIcon() map[int64]*model.GotoIcon {
	cfg := d.dupConfig()
	out := make(map[int64]*model.GotoIcon, len(cfg.StoryIcon))
	for key, value := range cfg.StoryIcon {
		iconType, err := strconv.ParseInt(key, 10, 64)
		if err != nil {
			continue
		}
		out[iconType] = value
	}
	return out
}

// Close close the resource.
func (d *dao) Close() {
	d.redis.Close()
	d.fanout.Close()
}

// Ping ping the resource.
func (d *dao) Ping(ctx context.Context) error {
	return nil
}
