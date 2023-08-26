package dao

import (
	"context"
	"fmt"
	xhttp "net/http"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/database/bfs"
	"go-common/library/database/sql"
	"go-common/library/log"
	"go-common/library/naming/discovery"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/resolver"
	"go-common/library/stat/prom"

	accgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	actApi "git.bilibili.co/bapis/bapis-go/activity/service"
	artclient "git.bilibili.co/bapis/bapis-go/article/service"
	cheesedyngrpc "git.bilibili.co/bapis/bapis-go/cheese/service/dynamic"
	cheeseseasongrpc "git.bilibili.co/bapis/bapis-go/cheese/service/season/season"
	channelgrpc "git.bilibili.co/bapis/bapis-go/community/interface/channel"
	dmgrpc "git.bilibili.co/bapis/bapis-go/community/interface/dm"
	taggrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	coingrpc "git.bilibili.co/bapis/bapis-go/community/service/coin"
	emotegrpc "git.bilibili.co/bapis/bapis-go/community/service/emote"
	favgrpc "git.bilibili.co/bapis/bapis-go/community/service/favorite"
	smsgrpc "git.bilibili.co/bapis/bapis-go/community/service/sms"
	thumbgrpc "git.bilibili.co/bapis/bapis-go/community/service/thumbup"
	dynfeedgrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/feed"
	dyntopicgrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/topic"
	votegrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/vote"
	bangumiCardgrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card/app"
	pgcShareGrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card/share"
	feedadmingrpc "git.bilibili.co/bapis/bapis-go/platform/admin/app-feed"
	topicsvc "git.bilibili.co/bapis/bapis-go/topic/service"

	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	resourcegrpc "go-gateway/app/app-svr/resource/service/api/v1"
	"go-gateway/app/web-svr/web/interface/conf"

	vastradegrpc "git.bilibili.co/bapis/bapis-go/vas/trans/trade/service"
)

const (
	_rankURL          = "/data/rank/"
	_feedbackURL      = "/x/internal/feedback/ugc/add"
	_spaceTopPhotoURL = "/api/member/getTopPhoto"
	_coinAddURL       = "/x/coin/add"
	_coinExpURL       = "/x/coin/today/exp"
	_elecShowURL      = "/api/v2/rank/query/av"
	_arcReportURL     = "/videoup/archive/report"
	_arcAppealURL     = "/x/internal/workflow/appeal/add"
	_appealTagsURL    = "/x/internal/workflow/tag/list"
	_relatedURL       = "/recsys/related"
	_helpListURI      = "/kb/getQuestionTypeListByParentIdBilibili/4"
	_helpListV2URI    = "/kb/getQuestionTypeListByParentIdForBilibili/4"
	_helpSearchURI    = "/kb/searchInerDocListBilibili/4"
	_helpSearchV2URI  = "/kb/searchInerDocListForBilibili/4"
	_onlineListURL    = "/x/internal/chat/num/top/aid"
	_searchURL        = "/main/search"
	_searchRecURL     = "/search/recommend"
	_searchDefaultURL = "/query/recommend"
	_searchUpRecURL   = "/main/recommend"
	_searchEggURI     = "/x/admin/feed/eggSearchWeb"
	_payWalletURL     = "/payplatform/cashier/wallet-int/getUserWalletInfo"
	_trending         = "/main/hotword/new"
	_tagBindURL       = "/x/internal/content-classify/content/tag/bind"
)

// Dao dao
type Dao struct {
	// config
	c *conf.Config
	// http client
	httpR        *bm.Client
	httpW        *bm.Client
	httpBigData  *bm.Client
	httpHelp     *bm.Client
	httpSearch   *bm.Client
	httpPay      *bm.Client
	httpGame     *bm.Client
	bfsClient    *xhttp.Client
	bfsClientSdk *bfs.BFS
	// redis
	redis                  *redis.Pool
	RedisIndex             *redis.Pool
	redisBak               *redis.Pool
	redisPopular           *redis.Pool
	redisNlBakExpire       int32
	redisRkExpire          int32
	redisRkBakExpire       int32
	redisDynamicBakExpire  int32
	redisArchiveBakExpire  int32
	redisTagBakExpire      int32
	redisCardBakExpire     int32
	redisRcBakExpire       int32
	redisArtBakExpire      int32
	redisHelpBakExpire     int32
	redisOlListBakExpire   int32
	redisAppealLimitExpire int32
	// bigdata url
	rankURL           string
	rankIndexURL      string
	rankRegionURL     string
	rankRecURL        string
	rankTagURL        string
	feedbackURL       string
	spaceTopPhotoURL  string
	coinAddURL        string
	coinExpURL        string
	customURL         string
	elecShowURL       string
	arcReportURL      string
	appealTagsURL     string
	arcAppealURL      string
	relatedURL        string
	arcRecommendURL   string
	onlineTotalURL    string
	helpListURL       string
	helpListV2URL     string
	helpSearchURL     string
	helpSearchV2URL   string
	onlineListURL     string
	shopURL           string
	replyHotURL       string
	searchURL         string
	searchRecURL      string
	searchDefaultURL  string
	searchUpRecURL    string
	searchEggURL      string
	walletURL         string
	abServerURL       string
	wxHotURL          string
	bnjConfURL        string
	bnj20ConfURL      string
	gameInfoURL       string
	searchGameInfoURL string
	hotLabelURL       string
	dyNumURL          string
	webTopURL         string
	gamePromoteURL    string
	hotRcmdURL        string
	wxTeenageRcmdURL  string
	searchTipDetail   string
	trending          string
	tagBindURL        string
	fawkesVersionURL  string
	cpmURL            string

	// dynamic draw
	drawDetails string
	// dynCommonBiz 动态通用模板信息路由
	dynCommonBiz string
	// cache Prom
	cacheProm *prom.Prom
	// grpc
	channelClient     channelgrpc.ChannelRPCClient
	bangumiCardClient bangumiCardgrpc.AppCardClient
	arcClient         arcgrpc.ArchiveClient
	coinClient        coingrpc.CoinClient
	favClient         favgrpc.FavoriteClient
	tagClient         taggrpc.TagRPCClient
	resourctClient    resourcegrpc.ResourceClient
	voteClient        votegrpc.VoteSvrClient
	dmClient          dmgrpc.DMClient
	accClient         accgrpc.AccountClient
	smsClient         smsgrpc.SmsClient
	feedAdminClient   feedadmingrpc.FeedAdminClient
	// grpc
	artClient         artclient.ArticleGRPCClient
	cheeseDynamicGRPC cheesedyngrpc.DynamicClient
	cheeseSeasonGRPC  cheeseseasongrpc.SeasonClient
	// dynamic grpc
	dynamicFeedGRPC dynfeedgrpc.FeedClient
	pgcShareGRPC    pgcShareGrpc.ShareClient
	thumbupGRPC     thumbgrpc.ThumbupClient
	ActivityClient  actApi.ActivityClient
	EmoteClient     emotegrpc.EmoteServiceClient
	TopicGRPC       topicsvc.TopicClient
	dynTopicClient  dyntopicgrpc.TopicClient
	tradeGRPC       vastradegrpc.VasTransTradeClient
	showDB          *sql.DB
}

// New dao new
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		// config
		c: c,
		// http read client
		httpR:       bm.NewClient(c.HTTPClient.Read, bm.SetResolver(resolver.New(nil, discovery.Builder()))),
		httpW:       bm.NewClient(c.HTTPClient.Write),
		httpBigData: bm.NewClient(c.HTTPClient.BigData),
		httpHelp:    bm.NewClient(c.HTTPClient.Help),
		httpSearch:  bm.NewClient(c.HTTPClient.Search, bm.SetResolver(resolver.New(nil, discovery.Builder()))),
		httpPay:     bm.NewClient(c.HTTPClient.Pay, bm.SetResolver(resolver.New(nil, discovery.Builder()))),
		httpGame:    bm.NewClient(c.HTTPClient.Game),
		// init bfs http client
		bfsClient:    xhttp.DefaultClient,
		bfsClientSdk: bfs.New(nil),
		// redis
		redis:                  redis.NewPool(c.Redis.LocalRedis.Config),
		RedisIndex:             redis.NewPool(c.Redis.IndexRedis.Config),
		redisBak:               redis.NewPool(c.Redis.BakRedis.Config),
		redisPopular:           redis.NewPool(c.Redis.Popular.Config),
		redisNlBakExpire:       int32(time.Duration(c.Redis.BakRedis.NewlistExpire) / time.Second),
		redisRkExpire:          int32(time.Duration(c.Redis.LocalRedis.RankingExpire) / time.Second),
		redisRkBakExpire:       int32(time.Duration(c.Redis.BakRedis.RankingExpire) / time.Second),
		redisDynamicBakExpire:  int32(time.Duration(c.Redis.BakRedis.RegionExpire) / time.Second),
		redisArchiveBakExpire:  int32(time.Duration(c.Redis.BakRedis.ArchiveExpire) / time.Second),
		redisTagBakExpire:      int32(time.Duration(c.Redis.BakRedis.TagExpire) / time.Second),
		redisCardBakExpire:     int32(time.Duration(c.Redis.BakRedis.CardExpire) / time.Second),
		redisRcBakExpire:       int32(time.Duration(c.Redis.BakRedis.RcExpire) / time.Second),
		redisArtBakExpire:      int32(time.Duration(c.Redis.BakRedis.ArtUpExpire) / time.Second),
		redisHelpBakExpire:     int32(time.Duration(c.Redis.BakRedis.HelpExpire) / time.Second),
		redisOlListBakExpire:   int32(time.Duration(c.Redis.BakRedis.OlListExpire) / time.Second),
		redisAppealLimitExpire: int32(time.Duration(c.Redis.BakRedis.AppealLimitExpire) / time.Second),
		// remote source urls
		rankURL:           c.Host.Rank + _rankURL + _rankURI,
		rankIndexURL:      c.Host.Rank + _rankURL + _rankIndexURI,
		rankRegionURL:     c.Host.Rank + _rankURL + _rankRegionURI,
		rankRecURL:        c.Host.Rank + _rankURL + _rankRecURI,
		wxHotURL:          c.Host.Rank + _rankURL + _wxHotURI,
		hotLabelURL:       c.Host.Rank + _hotLabelURI,
		rankTagURL:        c.Host.Rank + _rankTagURI,
		feedbackURL:       c.Host.API + _feedbackURL,
		spaceTopPhotoURL:  c.Host.Space + _spaceTopPhotoURL,
		coinAddURL:        c.Host.API + _coinAddURL,
		coinExpURL:        c.Host.API + _coinExpURL,
		customURL:         c.Host.Rank + _rankURL + _customURI,
		elecShowURL:       c.Host.Elec + _elecShowURL,
		arcReportURL:      c.Host.ArcAPI + _arcReportURL,
		appealTagsURL:     c.Host.API + _appealTagsURL,
		arcAppealURL:      c.Host.API + _arcAppealURL,
		relatedURL:        c.Host.Data + _relatedURL,
		arcRecommendURL:   c.Host.RcmdDiscovery + _arcRecommendURI,
		onlineTotalURL:    c.Host.API + _onlineTotalURI,
		helpListURL:       c.Host.HelpAPI + _helpListURI,
		helpSearchURL:     c.Host.HelpAPI + _helpSearchURI,
		helpListV2URL:     c.Host.HelpAPINew + _helpListV2URI,
		helpSearchV2URL:   c.Host.HelpAPINew + _helpSearchV2URI,
		onlineListURL:     c.Host.API + _onlineListURL,
		shopURL:           c.Host.Mall + _shopURI,
		replyHotURL:       c.Host.ReplyDiscovery + _hotURI,
		searchURL:         c.Host.SearchMainDiscovery + _searchURL,
		searchRecURL:      c.Host.Search + _searchRecURL,
		searchDefaultURL:  c.Host.SearchDiscovery + _searchDefaultURL,
		searchUpRecURL:    c.Host.Search + _searchUpRecURL,
		searchEggURL:      c.Host.Manager + _searchEggURI,
		walletURL:         c.Host.PayDiscovery + _payWalletURL,
		abServerURL:       c.Host.AbServer + _abServerURI,
		bnjConfURL:        c.Host.LiveAPI + _bnjConfURI,
		bnj20ConfURL:      c.Host.LiveAPI + _bnj2020Conf,
		gameInfoURL:       c.Host.Game + _gameInfoURI,
		searchGameInfoURL: c.Host.Game + _searchGameInfo,
		dyNumURL:          c.Host.VcAPI + _dynamicNumURI,
		webTopURL:         c.Host.Rank + _rankURL + _webTopURI,
		gamePromoteURL:    c.Host.Rank + _rankURL + _promoteURI,
		hotRcmdURL:        c.Host.Data + _hotRcmdURI,
		wxTeenageRcmdURL:  c.Host.TeenageRcmdDiscovery + _wxTeenagerRcmdURI,
		searchTipDetail:   c.Host.Manager + _searchTipDetail,
		trending:          c.Host.SearchDiscovery + _trending,
		cpmURL:            c.Host.AdDiscovery + _adURI, // 商业广告
		// dynamic draw
		drawDetails:  c.Host.VcAPI + _drawDetailsV2,
		dynCommonBiz: c.Host.VcAPI + _dynCommonBiz,
		// prom
		cacheProm: prom.CacheHit,
		showDB:    sql.NewMySQL(c.ShowDB),
		// tag bind
		tagBindURL: c.Host.API + _tagBindURL,
		// fawkes version url
		fawkesVersionURL: c.Host.FawkesAPI + _getFawkesVersion,
	}
	//grpc
	var err error
	if d.channelClient, err = channelgrpc.NewClient(c.ChannelGRPC); err != nil {
		panic(fmt.Sprintf("rpcClient NewClientt error (%+v)", err))
	}
	if d.bangumiCardClient, err = bangumiCardgrpc.NewClient(c.PGCRPC); err != nil {
		panic(fmt.Sprintf("appCardgrpc NewClientt error (%+v)", err))
	}
	if d.arcClient, err = arcgrpc.NewClient(c.ArchiveGRPC); err != nil {
		panic(fmt.Sprintf("rpcClient NewClientt error (%+v)", err))
	}
	if d.coinClient, err = coingrpc.NewClient(c.CoinGRPC); err != nil {
		panic(err)
	}
	if d.favClient, err = favgrpc.NewClient(c.FavoriteGRPC); err != nil {
		panic(err)
	}
	if d.tagClient, err = taggrpc.NewClient(c.TagClient); err != nil {
		panic(err)
	}
	if d.resourctClient, err = resourcegrpc.NewClient(c.ResourceClient); err != nil {
		panic(err)
	}
	if d.artClient, err = artclient.NewClient(c.ArticleGRPC); err != nil {
		panic(err)
	}
	if d.cheeseDynamicGRPC, err = cheesedyngrpc.NewClient(c.CheeseDynamicGRPC); err != nil {
		panic(err)
	}
	if d.cheeseSeasonGRPC, err = cheeseseasongrpc.NewClient(c.CheeseSeasonGRPC); err != nil {
		panic(err)
	}
	if d.dynamicFeedGRPC, err = dynfeedgrpc.NewClient(c.DynamicFeedGRPC); err != nil {
		panic(err)
	}
	if d.pgcShareGRPC, err = pgcShareGrpc.NewClient(c.PGCShareGRPC); err != nil {
		panic(err)
	}
	if d.voteClient, err = votegrpc.NewClient(c.VoteGRPC); err != nil {
		panic(err)
	}
	if d.dmClient, err = dmgrpc.NewClient(c.DMGRPC); err != nil {
		panic(err)
	}
	if d.accClient, err = accgrpc.NewClient(c.AccountGRPC); err != nil {
		panic(err)
	}
	if d.EmoteClient, err = emotegrpc.NewClient(c.AccountGRPC); err != nil {
		panic(err)
	}
	if d.ActivityClient, err = actApi.NewClient(c.ActGRPC); err != nil {
		panic(err)
	}
	if d.thumbupGRPC, err = thumbgrpc.NewClient(c.ThumbupClient); err != nil {
		panic(err)
	}
	if d.smsClient, err = smsgrpc.NewClient(c.SmsGRPC); err != nil {
		panic(err)
	}
	if d.feedAdminClient, err = feedadmingrpc.NewClient(c.FeedAdminGRPC); err != nil {
		panic(err)
	}
	if d.TopicGRPC, err = topicsvc.NewClient(c.TopicGRPC); err != nil {
		panic(err)
	}
	if d.dynTopicClient, err = dyntopicgrpc.NewClient(c.DynTopicGRPC); err != nil {
		panic(err)
	}
	if d.tradeGRPC, err = vastradegrpc.NewClientVasTransTrade(c.TradeGRPC); err != nil {
		panic(err)
	}
	return d
}

// Ping check connection success.
func (dao *Dao) Ping(c context.Context) (err error) {
	if err = dao.pingRedis(c); err != nil {
		log.Error("dao.pingRedis error(%v)", err)
		return
	}
	if err = dao.pingRedisBak(c); err != nil {
		log.Error("dao.pingRedisBak error(%v)", err)
		return
	}
	return
}

// Close close  resource.
func (dao *Dao) Close() {
	if dao.redisBak != nil {
		dao.redisBak.Close()
	}
	dao.showDB.Close()
}

func (dao *Dao) pingRedis(c context.Context) (err error) {
	return
}

func (dao *Dao) pingRedisBak(c context.Context) (err error) {
	conn := dao.redisBak.Get(c)
	_, err = conn.Do("SET", "PING", "PONG")
	conn.Close()
	return
}
