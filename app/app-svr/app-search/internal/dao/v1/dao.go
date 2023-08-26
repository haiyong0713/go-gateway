package v1

import (
	"context"
	"fmt"
	"time"

	"go-common/component/metadata/device"
	"go-common/component/metadata/network"
	"go-common/component/tinker"
	"go-common/library/conf/paladin.v2"
	infocv2 "go-common/library/log/infoc.v2"
	"go-common/library/naming/discovery"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/resolver"
	"go-common/library/net/rpc"
	"go-common/library/net/rpc/warden"

	appdynamicgrpc "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	searchadm "go-gateway/app/app-svr/app-feed/admin/model/search"
	"go-gateway/app/app-svr/app-search/configs"
	"go-gateway/app/app-svr/app-search/internal/model/search"
	arcmiddle "go-gateway/app/app-svr/archive/middleware/v1"
	arcapi "go-gateway/app/app-svr/archive/service/api"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	resmdl "go-gateway/app/app-svr/resource/service/model"
	resrpc "go-gateway/app/app-svr/resource/service/rpc/client"
	siriext "go-gateway/app/app-svr/siri-ext/service/api"
	ugcSeasonGrpc "go-gateway/app/app-svr/ugc-season/service/api"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	memberAPI "git.bilibili.co/bapis/bapis-go/account/service/member"
	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
	managersearch "git.bilibili.co/bapis/bapis-go/ai/search/mgr/interface"
	article "git.bilibili.co/bapis/bapis-go/article/model"
	artclient "git.bilibili.co/bapis/bapis-go/article/service"
	baikegrpc "git.bilibili.co/bapis/bapis-go/community/interface/baike"
	channelgrpc "git.bilibili.co/bapis/bapis-go/community/interface/channel"
	hisApi "git.bilibili.co/bapis/bapis-go/community/interface/history"
	coingrpc "git.bilibili.co/bapis/bapis-go/community/service/coin"
	favgrpc "git.bilibili.co/bapis/bapis-go/community/service/favorite"
	locationgrpc "git.bilibili.co/bapis/bapis-go/community/service/location"
	thumbupgrpc "git.bilibili.co/bapis/bapis-go/community/service/thumbup"
	dyngrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/feed"
	esportGRPC "git.bilibili.co/bapis/bapis-go/esports/service"
	livexfans "git.bilibili.co/bapis/bapis-go/live/xfansmedal"
	livexroom "git.bilibili.co/bapis/bapis-go/live/xroom"
	livexroomgate "git.bilibili.co/bapis/bapis-go/live/xroom-gate"
	gameEntryClient "git.bilibili.co/bapis/bapis-go/manager/operation/game-entry"
	esportsservice "git.bilibili.co/bapis/bapis-go/operational/esportsservice"
	gallerygrpc "git.bilibili.co/bapis/bapis-go/pangu/platform/gallery-service"
	mediagrpc "git.bilibili.co/bapis/bapis-go/pgc/servant/media"
	pgcinline "git.bilibili.co/bapis/bapis-go/pgc/service/card/inline"
	pgcsearch "git.bilibili.co/bapis/bapis-go/pgc/service/card/search/v1"
	reviewgrpc "git.bilibili.co/bapis/bapis-go/pgc/service/review"
	seasongrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/season"
	pgcstat "git.bilibili.co/bapis/bapis-go/pgc/service/stat/v1"
	"github.com/google/wire"
)

var Provider = wire.NewSet(New)

const (
	_recommand     = "/recommand"
	_bangumiCard   = "/pgc/internal/season/search/card"
	_dynamicDetail = "/dynamic_detail/v0/Dynamic/details"
	_dynamicTopics = "/topic_svr/v1/topic_svr/dyn_topics"
	_searchChannel = "/x/admin/search"
	// game
	_topGameButton = "/x/admin/manager/search/game/buttonInfo"
	_gameInfo      = "/game/multi_get_game_info"
	_topGame       = "/game/multi_get_game_info/for_intensify_card"
	_topGameInline = "/x/admin/manager/search/game/inlineInfo"
	// live
	_appMRoom    = "/xlive/internal/app-interface/v1/index/RoomsForAppIndex"
	_visibleInfo = "/rc/v1/Glory/get_visible"
	_usersInfo   = "/user/v3/User/getMultiple"
	// search
	_main         = "/main/search"
	_suggest      = "/main/suggest"
	_hot          = "/main/hotword"
	_trending     = "/main/hotword/new"
	_defaultWords = "/widget/getSearchDefaultWords"
	_rcmd         = "/query/recommend"
	_rcmdNoResult = "/search/recommend"
	_suggest3     = "/main/suggest/new"
	_pre          = "/search/frontpage"
	_upper        = "/main/recommend"
	_space        = "/space/search/v2"
	_searchTips   = "/x/admin/feed/open/search/tips"
	_mInfo        = "/x/internal/tag/minfo"
)

// Dao dao interface
type Dao interface {
	Close()
	Ping(ctx context.Context) (err error)
	GetConfig() *configs.Config
	// resource
	SpecialCards(c context.Context) (map[int64]*searchadm.SpreadConfig, error)
	ALLSearchSystemNotice(ctx context.Context) (map[int64]*search.SystemNotice, error)
	Banner(c context.Context, mobiApp, device, network, channel, buvid, adExtra, resIDStr string, build int, plat int8, mid int64) (res map[int][]*resmdl.Banner, err error)
	// search
	DynamicSearch(ctx context.Context, mid, vmid int64, keyword string, pn, ps int64) (dynamicIDs []int64, searchWords []string, total int64, err error)
	DynamicDetail(ctx context.Context, mid int64, dynamicIDs []int64, searchWords []string, playerArgs *arcmiddle.PlayerArgs, dev device.Device, ip string, net network.Network) (map[int64]*appdynamicgrpc.DynamicItem, error)
	Search(c context.Context, mid int64, mobiApp, device, platform, buvid, keyword, duration, order, filtered, fromSource, recommend, parent, adExtra, extraWord, tidList, durationList, qvid string, plat int8, seasonNum, movieNum, upUserNum, uvLimit, userNum, userVideoLimit, biliUserNum, biliUserVideoLimit, rid, highlight, build, pn, ps, isQuery, teenagersMode, lessonsMode int,
		old, isOgvExpNewUser bool, now time.Time, newPGC, flow, isNewOrder bool, autoPlayCard int64) (res *search.Search, code int, err error)
	Season(c context.Context, mid int64, keyword, mobiApp, device, platform, buvid, filtered string, plat int8, build, pn, ps int, now time.Time) (st *search.TypeSearch, code int, err error)
	Upper(c context.Context, mid int64, keyword, mobiApp, device, platform, buvid, filtered, order, qvid string, biliUserVL, highlight, build, userType, orderSort, pn, ps int, old bool, now time.Time, notices map[int64]*search.SystemNotice) (st *search.TypeSearch, code int, err error)
	MovieByType(c context.Context, mid int64, keyword, mobiApp, device, platform, buvid, filtered string, plat int8, build, pn, ps int, now time.Time) (st *search.TypeSearch, code int, err error)
	LiveByType(c context.Context, mid int64, keyword, mobiApp, device, platform, buvid, filtered, order, sType, qvid string, plat int8, build, pn, ps int, now time.Time) (st *search.TypeSearch, code int, err error)
	Live(c context.Context, mid int64, keyword, mobiApp, platform, buvid, device, order, sType, qvid string, build, pn, ps int) (st *search.TypeSearch, err error)
	LiveAll(c context.Context, mid int64, keyword, mobiApp, platform, buvid, device, order, sType string, build, pn, ps int) (st *search.TypeSearchLiveAll, err error)
	ArticleByType(c context.Context, mid int64, keyword, mobiApp, device, platform, buvid, filtered, order, sType, qvid string, plat int8, categoryID, build, highlight, pn, ps int, now time.Time) (st *search.TypeSearch, code int, err error)
	HotSearch(c context.Context, buvid string, mid int64, build, limit, zoneId int, mobiApp, device, platform string, now time.Time) (res *search.Hot, err error)
	Trending(c context.Context, buvid string, mid int64, build, limit, zoneId int, mobiApp, device, platform string, now time.Time, isRanking bool) (res *search.Hot, err error)
	Suggest(c context.Context, mid int64, buvid, term string, build int, mobiApp, device string, now time.Time) (res *search.Suggest, err error)
	Suggest2(c context.Context, mid int64, platform, buvid, term string, build int, mobiApp string, now time.Time) (res *search.Suggest2, err error)
	Suggest3(c context.Context, mid int64, platform, buvid, term, device string, build, highlight int, mobiApp string, now time.Time) (res *search.Suggest3, err error)
	Season2(c context.Context, mid int64, keyword, mobiApp, device, platform, buvid, qvid string, highlight, build, pn, ps int, fnver, fnval, qn, fourk int64) (st *search.TypeSearch, code int, err error)
	MovieByType2(c context.Context, mid int64, keyword, mobiApp, device, platform, buvid, qvid string, highlight, build, pn, ps int, fnver, fnval, qn, fourk int64) (st *search.TypeSearch, code int, err error)
	User(c context.Context, mid int64, keyword, mobiApp, device, platform, buvid, filtered, order, fromSource string, highlight, build, userType, orderSort, pn, ps int, now time.Time) (user []*search.User, err error)
	Recommend(c context.Context, mid int64, build, from, show, disableRcmd int, buvid, platform, mobiApp, device string) (res *search.RecommendResult, err error)
	DefaultWords(c context.Context, mid int64, build, from int, buvid, platform, mobiApp, device string, loginEvent int64, extParam *search.DefaultWordsExtParam) (res *search.DefaultWords, err error)
	RecommendNoResult(c context.Context, platform, mobiApp, device, buvid, keyword string, build, pn, ps int, mid int64) (res *search.NoResultRcndResult, err error)
	Channel(c context.Context, mid int64, keyword, mobiApp, platform, buvid, device, order, sType string, build, pn, ps, highlight int) (st *search.TypeSearch, code int, err error)
	RecommendPre(c context.Context, platform, mobiApp, device, buvid string, build, ps int, mid int64) (res *search.RecommendPreResult, err error)
	Video(c context.Context, mid int64, keyword, mobiApp, device, platform, buvid, order string, highlight, build, pn, ps int) (st *search.TypeSearch, code int, err error)
	Follow(c context.Context, platform, mobiApp, device, buvid string, build int, mid, vmid int64) (ups []*search.Upper, trackID string, err error)
	Converge(c context.Context, mid, cid int64, trackID, platform, mobiApp, device, buvid, order, sort string, plat int8, build, pn, ps int) (st *search.ResultConverge, err error)
	Space(c context.Context, mobiApp, platform, device, keyword, group, order, fromSource, buvid string, plat int8, build, rid, isTitle, highlight, pn, ps int, vmid, mid, attrNot int64, now time.Time) (res *search.Space, err error)
	ChannelNew(c context.Context, mid int64, keyword, mobiApp, platform, buvid, device string, build, pn, ps, highlight int) (st *search.ChannelResult, tids []int64, err error)
	SearchTips(c context.Context) (map[int64]*search.SearchTips, error)
	GetEsportConfigs(ctx context.Context, req *managersearch.GetEsportConfigsReq) (*managersearch.GetEsportConfigsResp, error)
	CheckNewDeviceAndUser(ctx context.Context, mid int64, buvid, periods string) bool
	// channel
	Details(c context.Context, tids []int64) (res map[int64]*channelgrpc.ChannelCard, err error)
	SearchChannel(c context.Context, mid int64, channelIDs []int64) (res map[int64]*channelgrpc.SearchChannel, err error)
	SearchChannelsInfo(c context.Context, mid int64, channelIDs []int64) (res map[int64]*channelgrpc.SearchChannelCard, err error)
	RelativeChannel(c context.Context, mid int64, channelIDs []int64) (res []*channelgrpc.RelativeChannel, err error)
	ChannelList(c context.Context, mid int64, ctype int32, offset string) (res *channelgrpc.ChannelListReply, err error)
	SearchChannelInHome(c context.Context, channelIDs []int64) (res *channelgrpc.SearchChannelInHomeReply, err error)
	ChannelInfos(ctx context.Context, channelIDs []int64) (map[int64]*channelgrpc.Channel, error)
	ChannelFav(ctx context.Context, mid, ps int64, offset string) (*channelgrpc.SubChannelReply, error)
	ChannelDetail(ctx context.Context, arg *channelgrpc.ChannelDetailReq) (*channelgrpc.ChannelDetailReply, error)
	ChannelFeed(ctx context.Context, arg *baikegrpc.ChannelFeedReq) (*baikegrpc.ChannelFeedReply, error)
	GetMediaBizInfoByMediaBizId(ctx context.Context, mediaId int64) (*mediagrpc.MediaBizInfoGetReply, error)
	GetMediaReviewInfo(ctx context.Context, mediaId int64) (*reviewgrpc.ReviewInfoReply, error)
	GetMediaAllowReview(ctx context.Context, mediaId int32) (*reviewgrpc.AllowReviewReply, error)
	// account
	Relations3(c context.Context, owners []int64, mid int64) (follows map[int64]bool)
	ProfilesWithoutPrivacy3(c context.Context, mids []int64) (map[int64]*accountgrpc.ProfileWithoutPrivacy, error)
	Cards3(c context.Context, mids []int64) (res map[int64]*accountgrpc.Card, err error)
	Infos3(c context.Context, mids []int64) (res map[int64]*accountgrpc.Info, err error)
	CheckRegTime(ctx context.Context, req *accountgrpc.CheckRegTimeReq) bool
	// ai
	AiRecommend(c context.Context) (rs map[int64]struct{}, err error)
	GetAiRecommendTags(ctx context.Context, style, numNot1st, nonPersonality int64, gt, id1st string, isHant bool) (*search.RecommendTagsRsp, error)
	// archive
	ArcsPlayer(c context.Context, playAvs []*arcapi.PlayAv, autoplayAreaValidate bool) (map[int64]*arcapi.ArcPlayer, error)
	Arcs(c context.Context, aids []int64, mobiApp, device string, mid int64) (map[int64]*arcapi.ArcPlayer, error)
	Archives(c context.Context, aids []int64, mobiApp, device string, mid int64) (map[int64]*arcapi.Arc, error)
	NFTBatchInfo(ctx context.Context, in *memberAPI.NFTBatchInfoReq) (*memberAPI.NFTBatchInfoReply, error)
	// bangumi
	SeasonsStatGRPC(ctx context.Context, seasonIds []int32) (result map[int32]*pgcstat.SeasonStatProto, err error)
	SeasonCards(ctx context.Context, seasonIds []int32) (res map[int32]*seasongrpc.CardInfoProto, err error)
	SearchEpsGrpc(ctx context.Context, req *search.EpisodesNewReq) (reply *pgcsearch.SearchEpReply, err error)
	BangumiCard(c context.Context, mid int64, sids []int64) (s map[string]*search.Card, err error)
	SearchPGCCards(ctx context.Context, seps []*pgcsearch.SeasonEpReq, query, mobiApp, device_, platform string, mid int64, fnver, fnval, qn, fourk, build int64, isWithPlayURL bool) (result map[int32]*pgcsearch.SearchCardProto, medias map[int32]*pgcsearch.SearchMediaProto, err error)
	InlineCards(c context.Context, epIDs []int32, mobiApp, platform, device string, build int, mid int64) (map[int32]*pgcinline.EpisodeCard, error)
	SugOGV(c context.Context, ssids []int32) (res map[int32]*pgcsearch.SearchCardProto, err error)
	// es
	EsSearchChannel(c context.Context, mid int64, keyword string, pn, ps, state int) (st *search.ChannelResult, tids []int64, err error)
	// game
	CloudGameEntry(ctx context.Context, req *gameEntryClient.MultiShowReq) (*gameEntryClient.MultiShowResp, error)
	MultiGameInfos(ctx context.Context, mid int64, ids []int64, build, sdkType int) (map[int64]*search.NewGame, error)
	TopGame(ctx context.Context, mid int64, topGameIDs []int64, sdkType int) ([]*search.TopGameData, error)
	FetchTopGameConfigs(ctx context.Context, gameIds []int64) (*search.TopGameConfig, error)
	FetchTopGameInlineConfigs(ctx context.Context, cardIds []int64) (*search.TopGameInlineInfo, error)
	// live
	QueryMedalStatus(c context.Context, mid int64) (status int64, err error)
	LiveGetMultiple(ctx context.Context, roomIDs []int64) (map[int64]*livexroom.Infos, error)
	EntryRoomInfo(ctx context.Context, req *livexroomgate.EntryRoomInfoReq) (map[int64]*livexroomgate.EntryRoomInfoResp_EntryList, error)
	GetMultipleWithPlayUrl(c context.Context, roomIDs []int64, param *search.LiveParam) (map[int64]*livexroom.Infos, map[int64]*livexroom.LivePlayUrlData, error)
	GetMultiple(ctx context.Context, roomIDs []int64) (map[int64]*livexroom.Infos, error)
	AppMRoom(c context.Context, roomids []int64, platform string) (map[int64]*search.Room, error)
	UserInfo(c context.Context, uids []int64) (userInfo map[int64]map[string]*search.Exp, err error)
	LiveGlory(c context.Context, uid int64) (glory []*search.LiveGlory, err error)
	// relation
	Interrelations(ctx context.Context, mid int64, owners []int64) (res map[int64]*relationgrpc.InterrelationReply, err error)
	// tag
	TagInfos(c context.Context, tags []int64, mid int64) (tagMyInfo []*search.Tag, err error)
	// thumb up
	HasLike(ctx context.Context, buvid string, mid int64, messageIDs []int64) (map[int64]thumbupgrpc.State, error)
	// ugcseason
	SeasonView(ctx context.Context, req *ugcSeasonGrpc.ViewRequest) (*ugcSeasonGrpc.ViewReply, error)
	// siri ext
	ResolveCommand(ctx context.Context, req *siriext.ResolveCommandReq) (*siriext.ResolveCommandReply, error)
	// gallery
	GetNFTRegionBatch(ctx context.Context, nftID []string) (*gallerygrpc.GetNFTRegionReply, error)
	// fav
	IsFavVideos(ctx context.Context, mid int64, aids []int64) (map[int64]int8, error)
	// coin
	ArchiveUserCoins(ctx context.Context, aids []int64, mid int64) (map[int64]int64, error)
	// bplus
	DynamicDetails(c context.Context, ids []int64, from string) (details map[int64]*search.Detail, err error)
	DynamicTopics(c context.Context, dynamicIDs []int64, platform, mobiApp string, build int) (cs map[int64]*search.DynamicTopics, err error)
	// match
	Matchs(c context.Context, mid int64, matchIDs []int64) (res map[int64]*esportGRPC.Contest, err error)
	GetSportsEventMatches(ctx context.Context, req *esportsservice.GetSportsEventMatchesReq) (res *esportsservice.GetSportsEventMatchesResponse, err error)
	// article
	Articles(c context.Context, aids []int64) (arts map[int64]*article.Meta, err error)
	// location
	LocationInfo(c context.Context, ipaddr string) (info *locationgrpc.InfoReply, err error)
	// history
	GetHistoryFrequent(ctx context.Context, req *hisApi.HistoryFrequentReq) (*hisApi.HistoryFrequentReply, error)
	// comic
	GetComicInfos(ctx context.Context, ids []int64) (map[int64]*search.ComicInfo, error)
}

// dao dao.
type dao struct {
	ac *paladin.Map
	c  *configs.Config

	infocv2 infocv2.Infoc

	httpClientCfg *HttpClient
	grpcClientCfg *GrpcClient

	client          *bm.Client
	searchClient    *bm.Client
	gameClient      *bm.Client
	liveClient      *bm.Client
	bangumiClient   *bm.Client
	feedAdminClient *bm.Client

	accountClient        accountgrpc.AccountClient
	memberClient         memberAPI.MemberClient
	archiveClient        arcgrpc.ArchiveClient
	artClient            artclient.ArticleGRPCClient
	pgcstatClient        pgcstat.StatServiceClient
	pgcsearchClient      pgcsearch.SearchClient
	seasonRpcClient      seasongrpc.SeasonClient
	channelClient        channelgrpc.ChannelRPCClient
	baikeClient          baikegrpc.BaikeClient
	ogvMediaClient       mediagrpc.MediaClient
	ogvReviewClient      reviewgrpc.ReviewClient
	galleryClient        gallerygrpc.GalleryServiceClient
	cloudGameEntryClient gameEntryClient.OperationItemGameEntryV1Client
	liveRpcClient        livexfans.AnchorClient
	roomRPCClient        livexroom.RoomClient
	roomGateClient       livexroomgate.XroomgateClient
	locationClient       locationgrpc.LocationClient
	esportClient         esportGRPC.EsportsClient
	sportClient          esportsservice.EsportsServiceClient
	relationClient       relationgrpc.RelationClient
	resRPC               *resrpc.Service
	appDynamicClient     appdynamicgrpc.DynamicClient
	managersearch        managersearch.SearchMgrInterfaceClient
	siriExtClient        siriext.SiriExtClient
	ugcSeasonClient      ugcSeasonGrpc.UGCSeasonClient
	hisClient            hisApi.HistoryClient

	thumbupClient   thumbupgrpc.ThumbupClient
	favClient       favgrpc.FavoriteClient
	coinClient      coingrpc.CoinClient
	pgcinlineClient pgcinline.InlineCardClient

	dynamicClient dyngrpc.FeedClient

	rcmd              string
	rcmdAi            string
	rcmdTag           string
	bangumiCard       string
	dynamicDetail     string
	dynamicTopics     string
	searchChannel     string
	topGame           string
	gameMultiInfos    string
	topGameConfig     string
	topGameInlineInfo string
	// live
	appMRoom    string
	visibleInfo string
	userInfo    string
	// search
	main         string
	suggest      string
	hot          string
	trending     string
	defaultWords string
	rcmdNoResult string
	suggest3     string
	pre          string
	upper        string
	space        string
	searchTips   string
	// tag
	mInfo string
	// comic
	comicInfos string
}

type HttpClient struct {
	// httpData
	HTTPData      *bm.ClientConfig
	HTTPSearch    *bm.ClientConfig
	HTTPGame      *bm.ClientConfig
	HTTPLive      *bm.ClientConfig
	HTTPBangumi   *bm.ClientConfig
	HTTPFeedAdmin *bm.ClientConfig

	Host          *Host
	HostDiscovery *HostDiscovery
}

type Host struct {
	Data      string
	APICo     string
	VC        string
	Manager   string
	GameCo    string
	APILiveCo string
	WWW       string
	FeedAdmin string
	ComicCo   string
}

type HostDiscovery struct {
	Data       string
	Search     string
	FeedAdmin  string
	SearchMain string
}

type GrpcClient struct {
	// pgc inline grpc
	PGCInline *warden.ClientConfig
	// AccountGRPC grpc
	AccountGRPC *warden.ClientConfig
	MemClient   *warden.ClientConfig
	// ArticleGRPC grpc
	ArticleGRPC *warden.ClientConfig
	// ActivityGRPC grpc
	ActivityClient *warden.ClientConfig
	// PGCRPC grpc
	PGCRPC *warden.ClientConfig
	// gallery
	GalleryGRPC *warden.ClientConfig
	// game
	GameEntryGRPC *warden.ClientConfig

	// RelationGRPC grpc
	RelationGRPC *warden.ClientConfig
	// grpc Archive
	ArchiveGRPC *warden.ClientConfig
	ThumbupGRPC *warden.ClientConfig
	// grpc location
	LocationGRPC *warden.ClientConfig
	// esport
	ESportsGRPC *warden.ClientConfig
	// rpc client
	ResourceRPC *rpc.ClientConfig
	// dynamic client
	AppDynamicGRPC *warden.ClientConfig
	// manager client
	ManagerSearchGRPC *warden.ClientConfig
	// siri
	SiriExtGRPC *warden.ClientConfig
	// ugc season
	UGCSeasonGRPC *warden.ClientConfig
	// history
	HistoryGRPC *warden.ClientConfig
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
	OgvMediaGRPC      *warden.ClientConfig
	AppResourceClient *warden.ClientConfig
}

// New new a dao and return.
func New() (d Dao, cf func(), err error) {
	return newDao()
}

func newDao() (d *dao, cf func(), err error) {
	d = &dao{
		ac: &paladin.TOML{},
		c:  &configs.Config{},
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
	if err = d.ac.Get("config").UnmarshalTOML(&d.c); err != nil {
		panic(err)
	}
	// init infocv2
	d.infocv2, _ = infocv2.New(nil)
	// init tinker
	abt := tinker.Init(d.infocv2, nil)
	defer abt.Close()
	d.newHTTPClient()
	if err = d.newGRPCClient(); err != nil {
		panic(err)
	}

	cf = d.Close
	return
}

func (d *dao) newHTTPClient() {
	d.client = bm.NewClient(d.httpClientCfg.HTTPData, bm.SetResolver(resolver.New(nil, discovery.Builder())))
	d.searchClient = bm.NewClient(d.httpClientCfg.HTTPSearch, bm.SetResolver(resolver.New(nil, discovery.Builder())))
	d.gameClient = bm.NewClient(d.httpClientCfg.HTTPGame)
	d.liveClient = bm.NewClient(d.httpClientCfg.HTTPLive)
	d.bangumiClient = bm.NewClient(d.httpClientCfg.HTTPBangumi)
	d.feedAdminClient = bm.NewClient(d.httpClientCfg.HTTPFeedAdmin, bm.SetResolver(resolver.New(nil, discovery.Builder())))

	d.rcmd = d.httpClientCfg.HostDiscovery.Search + _rcmd
	d.rcmdAi = d.httpClientCfg.HostDiscovery.Data + _recommand
	d.rcmdTag = d.httpClientCfg.Host.Data + _recommand
	d.bangumiCard = d.httpClientCfg.Host.APICo + _bangumiCard
	d.dynamicDetail = d.httpClientCfg.Host.VC + _dynamicDetail
	d.dynamicTopics = d.httpClientCfg.Host.VC + _dynamicTopics
	d.searchChannel = d.httpClientCfg.Host.APICo + _searchChannel
	d.topGameConfig = d.httpClientCfg.Host.Manager + _topGameButton
	d.topGameInlineInfo = d.httpClientCfg.Host.Manager + _topGameInline
	d.gameMultiInfos = d.httpClientCfg.Host.GameCo + _gameInfo
	d.topGame = d.httpClientCfg.Host.GameCo + _topGame
	d.appMRoom = d.httpClientCfg.Host.APILiveCo + _appMRoom
	d.visibleInfo = d.httpClientCfg.Host.APILiveCo + _visibleInfo
	d.userInfo = d.httpClientCfg.Host.APILiveCo + _usersInfo
	d.main = d.httpClientCfg.HostDiscovery.SearchMain + _main
	d.suggest = d.httpClientCfg.HostDiscovery.Search + _suggest
	d.hot = d.httpClientCfg.HostDiscovery.Search + _hot
	d.trending = d.httpClientCfg.HostDiscovery.Search + _trending
	d.defaultWords = d.httpClientCfg.Host.WWW + _defaultWords
	d.rcmdNoResult = d.httpClientCfg.HostDiscovery.Search + _rcmdNoResult
	d.suggest3 = d.httpClientCfg.HostDiscovery.Search + _suggest3
	d.pre = d.httpClientCfg.HostDiscovery.Search + _pre
	d.upper = d.httpClientCfg.HostDiscovery.Search + _upper
	d.space = d.httpClientCfg.HostDiscovery.Search + _space
	d.searchTips = d.httpClientCfg.Host.Manager + _searchTips
	d.mInfo = d.httpClientCfg.Host.APICo + _mInfo
	d.comicInfos = d.httpClientCfg.Host.ComicCo + _comicInfos
}

func (d *dao) newGRPCClient() (err error) {
	if d.artClient, err = artclient.NewClient(d.grpcClientCfg.ArticleGRPC); err != nil {
		panic(err)
	}
	if d.pgcsearchClient, d.pgcinlineClient, err = newPgcClient(d.grpcClientCfg.PGCRPC); err != nil {
		panic(fmt.Sprintf("pgcsearch pgcinline newClient error (%+v)", err))
	}
	if d.pgcstatClient, err = newStatClient(d.grpcClientCfg.PGCRPC); err != nil {
		panic(fmt.Sprintf("pgcstat newStatClient error (%+v)", err))
	}
	if d.seasonRpcClient, err = seasongrpc.NewClient(d.grpcClientCfg.PGCRPC); err != nil {
		panic(fmt.Sprintf("seasongrpc NewClientt error (%+v)", err))
	}
	if d.archiveClient, err = arcgrpc.NewClient(d.grpcClientCfg.ArchiveGRPC); err != nil {
		return err
	}
	if d.channelClient, err = channelgrpc.NewClient(d.grpcClientCfg.ChannelGRPC); err != nil {
		return err
	}
	if d.baikeClient, err = baikegrpc.NewClientBaike(d.grpcClientCfg.ChannelGRPC); err != nil {
		panic(fmt.Sprintf("rpcClient NewClientt error (%+v)", err))
	}
	if d.ogvMediaClient, err = mediagrpc.NewClientMedia(d.grpcClientCfg.OgvMediaGRPC); err != nil {
		panic(fmt.Sprintf("rpcClient NewClientMedia error (%+v)", err))
	}
	if d.ogvReviewClient, err = reviewgrpc.NewClientReview(d.grpcClientCfg.OgvMediaGRPC); err != nil {
		panic(fmt.Sprintf("rpcClient NewClientReview error (%+v)", err))
	}
	if d.galleryClient, err = gallerygrpc.NewClient(d.grpcClientCfg.GalleryGRPC); err != nil {
		panic(err)
	}
	if d.cloudGameEntryClient, err = gameEntryClient.NewClientOperationItemGameEntryV1(d.grpcClientCfg.GameEntryGRPC); err != nil {
		panic(err)
	}
	if d.liveRpcClient, err = newLiveClient(d.grpcClientCfg.LiveGRPC); err != nil {
		panic(fmt.Sprintf("livexfans newClient error (%+v)", err))
	}
	if d.roomRPCClient, err = newLiveRoomClient(d.grpcClientCfg.LiveGRPC); err != nil {
		panic(fmt.Sprintf("livexroom newLiveRoomClient error (%+v)", err))
	}
	if d.roomGateClient, err = livexroomgate.NewClientXroomgate(d.grpcClientCfg.LiveGRPC); err != nil {
		panic(fmt.Sprintf("livexroomgate NewClientXroomgate error (%+v)", err))
	}
	if d.thumbupClient, err = thumbupgrpc.NewClient(d.grpcClientCfg.ThumbupGRPC); err != nil {
		return err
	}
	if d.favClient, err = favgrpc.NewClient(d.grpcClientCfg.FavClient); err != nil {
		return err
	}
	if d.coinClient, err = coingrpc.NewClient(d.grpcClientCfg.CoinClient); err != nil {
		return err
	}
	if d.pgcinlineClient, err = pgcinline.NewClient(d.grpcClientCfg.PGCInline); err != nil {
		return err
	}
	if d.accountClient, err = accountgrpc.NewClient(d.grpcClientCfg.AccountGRPC); err != nil {
		return err
	}
	if d.memberClient, err = memberAPI.NewClient(d.grpcClientCfg.MemClient); err != nil {
		panic(err)
	}
	if d.relationClient, err = relationgrpc.NewClient(d.grpcClientCfg.RelationGRPC); err != nil {
		return err
	}
	if d.locationClient, err = locationgrpc.NewClient(d.grpcClientCfg.LocationGRPC); err != nil {
		return err
	}
	if d.esportClient, err = esportGRPC.NewClient(d.grpcClientCfg.ESportsGRPC); err != nil {
		panic(err)
	}
	if d.sportClient, err = esportsservice.NewClient(d.grpcClientCfg.ESportsGRPC); err != nil {
		panic(err)
	}
	d.resRPC = resrpc.New(d.grpcClientCfg.ResourceRPC)
	if d.appDynamicClient, err = appdynamicgrpc.NewClient(d.grpcClientCfg.AppDynamicGRPC); err != nil {
		panic(err)
	}
	if d.managersearch, err = managersearch.NewClientSearchMgrInterface(d.grpcClientCfg.ManagerSearchGRPC); err != nil {
		panic(err)
	}
	if d.siriExtClient, err = siriext.NewClient(d.grpcClientCfg.SiriExtGRPC); err != nil {
		panic(err)
	}
	if d.ugcSeasonClient, err = ugcSeasonGrpc.NewClient(d.grpcClientCfg.UGCSeasonGRPC); err != nil {
		panic(fmt.Sprintf("rpcClient NewClientt error (%+v)", err))
	}
	if d.hisClient, err = hisApi.NewClient(d.grpcClientCfg.HistoryGRPC); err != nil {
		panic(fmt.Sprintf("hisApi.NewClient error (%+v)", err))
	}
	if d.dynamicClient, err = dyngrpc.NewClient(d.grpcClientCfg.DynamicGRPC); err != nil {
		return err
	}
	return nil
}

// Close close the resource.
func (d *dao) Close() {
	_ = d.infocv2.Close()
}

func (d *dao) GetConfig() *configs.Config {
	return d.c
}

// Ping ping the resource.
func (d *dao) Ping(ctx context.Context) (err error) {
	return nil
}
