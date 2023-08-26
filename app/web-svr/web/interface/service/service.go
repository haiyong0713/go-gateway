package service

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"go-common/library/conf/env"
	"go-common/library/log"
	infocV2 "go-common/library/log/infoc.v2"
	"go-common/library/net/metadata"
	"go-common/library/railgun"
	errGroup "go-common/library/sync/errgroup.v2"
	"go-common/library/sync/pipeline/fanout"
	honorgrpc "go-gateway/app/app-svr/archive-honor/service/api"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	fkmdl "go-gateway/app/app-svr/fawkes/service/model/business"
	resgrpc "go-gateway/app/app-svr/resource/service/api/v1"
	steinsgrpc "go-gateway/app/app-svr/steins-gate/service/api"
	ugcseason "go-gateway/app/app-svr/ugc-season/service/api"
	dygrpc "go-gateway/app/web-svr/dynamic/service/api/v1"
	dyrpc "go-gateway/app/web-svr/dynamic/service/rpc/client"
	"go-gateway/app/web-svr/web/interface/conf"
	"go-gateway/app/web-svr/web/interface/dao"
	bossdao "go-gateway/app/web-svr/web/interface/dao/boss"
	"go-gateway/app/web-svr/web/interface/dao/campus"
	"go-gateway/app/web-svr/web/interface/dao/captcha"
	"go-gateway/app/web-svr/web/interface/dao/live"
	"go-gateway/app/web-svr/web/interface/dao/match"
	"go-gateway/app/web-svr/web/interface/dao/pcdn"
	"go-gateway/app/web-svr/web/interface/dao/rcmd"
	tag "go-gateway/app/web-svr/web/interface/dao/tag_legacy"
	"go-gateway/app/web-svr/web/interface/model"
	"go-gateway/app/web-svr/web/interface/model/search"
	api "go-gateway/app/web-svr/web/interface/service/player-online"
	"go-gateway/pkg/idsafe/bvid"

	accgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	cougrpc "git.bilibili.co/bapis/bapis-go/account/service/coupon"
	accmgrpc "git.bilibili.co/bapis/bapis-go/account/service/member"
	relagrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
	ugcgrpc "git.bilibili.co/bapis/bapis-go/account/service/ugcpay"
	payrank "git.bilibili.co/bapis/bapis-go/account/service/ugcpay-rank"
	actgrpc "git.bilibili.co/bapis/bapis-go/activity/service"
	upgrpc "git.bilibili.co/bapis/bapis-go/archive/service/up"
	artgrpc "git.bilibili.co/bapis/bapis-go/article/service"
	playeronlinegrpc "git.bilibili.co/bapis/bapis-go/bilibili/app/playeronline/v1"
	cheesegrpc "git.bilibili.co/bapis/bapis-go/cheese/service/season/episode"
	ansgrpc "git.bilibili.co/bapis/bapis-go/community/interface/answer"
	channelgrpc "git.bilibili.co/bapis/bapis-go/community/interface/channel"
	dmgrpc "git.bilibili.co/bapis/bapis-go/community/interface/dm"
	hisgrpc "git.bilibili.co/bapis/bapis-go/community/interface/history"
	sharegrpc "git.bilibili.co/bapis/bapis-go/community/interface/share"
	taggrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	coingrpc "git.bilibili.co/bapis/bapis-go/community/service/coin"
	favgrpc "git.bilibili.co/bapis/bapis-go/community/service/favorite"
	locgrpc "git.bilibili.co/bapis/bapis-go/community/service/location"
	tagSvrgrpc "git.bilibili.co/bapis/bapis-go/community/service/tag"
	thumbgrpc "git.bilibili.co/bapis/bapis-go/community/service/thumbup"
	cfcgrpc "git.bilibili.co/bapis/bapis-go/content-flow-control/service"
	creativegrpc "git.bilibili.co/bapis/bapis-go/creative/open/service"
	garbgrpc "git.bilibili.co/bapis/bapis-go/garb/service"
	brdgrpc "git.bilibili.co/bapis/bapis-go/infra/service/broadcast"
	watchedGRPC "git.bilibili.co/bapis/bapis-go/live/watched/v1"
	roomgrpc "git.bilibili.co/bapis/bapis-go/live/xroom"
	activegrpc "git.bilibili.co/bapis/bapis-go/manager/service/active"
	populargrpc "git.bilibili.co/bapis/bapis-go/manager/service/popular"
	pangugsgrpc "git.bilibili.co/bapis/bapis-go/pangu/platform/gallery-service"
	pgcsearch "git.bilibili.co/bapis/bapis-go/pgc/service/card/search/v1"
	epgrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/episode"
	seasongrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/season"
	shareadmin "git.bilibili.co/bapis/bapis-go/platform/admin/share"
	actplatgrpc "git.bilibili.co/bapis/bapis-go/platform/interface/act-plat-v2"
	res2grpc "git.bilibili.co/bapis/bapis-go/resource/service/v2"
	gaiagrpc "git.bilibili.co/bapis/bapis-go/silverbullet/gaia/interface"
	uparcgrpc "git.bilibili.co/bapis/bapis-go/up-archive/service"

	bgroupgrpc "git.bilibili.co/bapis/bapis-go/platform/service/bgroup/v2"
	videogrpc "git.bilibili.co/bapis/bapis-go/videoup/open/service"

	"github.com/pkg/errors"
	"github.com/robfig/cron"
)

// Service service
type Service struct {
	c         *conf.Config
	dao       *dao.Dao
	rcmdDao   *rcmd.Dao
	matchDao  *match.Dao
	liveDao   *live.Dao
	campusDao *campus.Dao // 校园
	pcdnDao   *pcdn.Dao
	// rpc
	dy  *dyrpc.Service
	tag *tag.TagRPCService
	// cache proc
	cache                *fanout.Fanout
	regionCount          map[int32]int64
	onlineAids           []*model.OnlineAid
	regionList           map[string][]*model.Region
	searchTipDetailCache map[int64]*model.TipDetail
	systemNoticeCache    map[int64]*model.SystemNotice
	fawkesVersionCache   map[string]map[string]*fkmdl.Version
	// cron
	cron               *cron.Cron
	onlineTotal        int64
	onlineTotalRunning bool
	// rand source
	r          *rand.Rand
	indexIcons []*model.IndexIcon
	indexIcon  *model.IndexIcon
	// infoc2
	infocV2Log infocV2.Infoc
	// TypeNames
	typeNames map[int32]*arcgrpc.Tp
	// Broadcast grpc client
	broadcastGRPC  brdgrpc.ZergClient
	channelGRPC    channelgrpc.ChannelRPCClient
	coinGRPC       coingrpc.CoinClient
	arcGRPC        arcgrpc.ArchiveClient
	accGRPC        accgrpc.AccountClient
	accmGRPC       accmgrpc.MemberClient
	pangugsGRPC    pangugsgrpc.GalleryServiceClient
	shareGRPC      sharegrpc.ShareClient
	ugcPayGRPC     ugcgrpc.UGCPayClient
	resgrpc        resgrpc.ResourceClient
	res2grpc       res2grpc.ResourceClient
	creativeGRPC   creativegrpc.CreativeClient
	thumbupGRPC    thumbgrpc.ThumbupClient
	upGRPC         upgrpc.UpClient
	relationGRPC   relagrpc.RelationClient
	dmGRPC         dmgrpc.DMClient
	hisGRPC        hisgrpc.HistoryClient
	pgcSearchGRPC  pgcsearch.SearchClient
	epGRPC         epgrpc.EpisodeClient
	dyGRPC         dygrpc.DynamicClient
	ansGRPC        ansgrpc.AnswerClient
	tagGRPC        taggrpc.TagRPCClient
	tagSvrGRPC     tagSvrgrpc.TagRPCClient
	artGRPC        artgrpc.ArticleGRPCClient
	couponGRPC     cougrpc.CouponClient
	steinsGRPC     steinsgrpc.SteinsGateClient
	ugcSeasonGRPC  ugcseason.UGCSeasonClient
	cheeseGRPC     cheesegrpc.EpisodeClient
	locGRPC        locgrpc.LocationClient
	favGRPC        favgrpc.FavoriteClient
	payRankGRPC    payrank.UGCPayRankClient
	videoUpGRPC    videogrpc.VideoUpOpenClient
	roomGRPC       roomgrpc.RoomClient
	gaiaGRPC       gaiagrpc.GaiaClient
	shareAdminGRPC shareadmin.ShareAdminClient
	garbGRPC       garbgrpc.GarbClient
	upArcGRPC      uparcgrpc.UpArchiveClient
	comActiveGRPC  activegrpc.CommonActiveClient
	actGRPC        actgrpc.ActivityClient
	actPlatGRPC    actplatgrpc.ActPlatClient
	popularGRPC    populargrpc.PopularClient
	seasonGRPC     seasongrpc.SeasonClient
	honorGRPC      honorgrpc.ArchiveHonorClient
	watchedGRPC    watchedGRPC.WatchClient
	cfcGRPC        cfcgrpc.FlowControlClient
	onlineGRPC     playeronlinegrpc.PlayerOnlineClient
	bgroupGRPC     bgroupgrpc.BGroupServiceClient
	// searchEggs
	searchEggs map[int64]*search.SearchEggRes
	// bnj
	bnj2019View *model.Bnj2019View
	bnjElecInfo *model.ElecShow
	bnj2019List []*model.Bnj2019Related
	// special group mids
	specialMids  map[int64]struct{}
	specRcmdCard map[int64]*res2grpc.WebSpecialCard
	// wx index aids
	wxHotAids []*model.WxArchiveCard
	// hot label aids
	hotLabelAids map[int64]struct{}
	// stein gate view guide cid
	steinGuideCid int64
	bnj20Cache    *model.Bnj20Cache
	// load switch
	newCountRunning   bool
	onlineListRunning bool
	typeNamesRunning  bool
	searchEggRunning  bool
	managerRunning    bool
	specRcmdRunning   bool
	guideCidRunning   bool
	indexIconRunning  bool
	randomIconRunning bool
	wxHotRunning      bool
	hotLabelRunning   bool
	bnj19MainRunning  bool
	bnj19ListRunning  bool
	bnj20MainRunning  bool
	bnj20LiveRunning  bool
	bnj20SpRunning    bool
	bnj20ListRunning  bool
	// rank data
	rankIndexData       map[int][]int64
	rankRegionData      map[string][]*model.NewArchive
	rankRecommendData   map[int64][]int64
	lpRankRecommendData map[string][]int64
	webTopData          []int64
	// information region card
	informationRegionCardCache map[int32][]*resgrpc.InformationRegionCard
	// param config
	paramConfigCache map[string]*model.ParamConfig
	// ranking data
	rankingV2Data map[string]*model.RankV2Cache
	// activitySeason
	activitySeasonRailGun *railgun.Railgun
	activitySeasonKeyMem  map[string]*model.ActivitySeasonMem
	activitySeasonIDMem   map[int64]*model.ActivitySeasonMem
	// series
	seriesConfigBak []*model.SeriesConfig
	seriesListBak   map[int64][]*model.SeriesList
	// pwd_appeal
	PwdAppealCaptchaDao *captcha.Dao
	PwdAppealBoss       *bossdao.Dao
}

// New new
func New(c *conf.Config) *Service {
	s := &Service{
		c:           c,
		dao:         dao.New(c),
		rcmdDao:     rcmd.New(c),
		dy:          dyrpc.New(c.DynamicRPC),
		tag:         tag.NewTagRPC(c.TagRPC),
		matchDao:    match.New(c),
		campusDao:   campus.New((c)),
		pcdnDao:     pcdn.New(c),
		cache:       fanout.New("cache"),
		cron:        cron.New(),
		regionCount: make(map[int32]int64),
		r:           rand.New(rand.NewSource(time.Now().UnixNano())),
		specialMids: map[int64]struct{}{},
		bnj20Cache:  &model.Bnj20Cache{LiveGiftCnt: 1340426}, // 拜年祭结束后固定
		// information region cache
		informationRegionCardCache: make(map[int32][]*resgrpc.InformationRegionCard),
		// param config cache
		paramConfigCache: make(map[string]*model.ParamConfig),
		PwdAppealBoss:    bossdao.NewDao(c.PwdAppeal.Boss),
		// fawkes cache
		fawkesVersionCache: make(map[string]map[string]*fkmdl.Version),
		liveDao:            live.New(c),
	}
	var err error
	if s.infocV2Log, err = infocV2.New(c.InfocV2); err != nil {
		panic(err)
	}
	if s.broadcastGRPC, err = brdgrpc.NewClient(c.BroadcastClient); err != nil {
		panic(err)
	}
	if s.coinGRPC, err = coingrpc.NewClient(c.CoinClient); err != nil {
		panic(err)
	}
	if s.arcGRPC, err = arcgrpc.NewClient(c.ArcClient); err != nil {
		panic(err)
	}
	if s.accGRPC, err = accgrpc.NewClient(c.AccClient); err != nil {
		panic(err)
	}
	if s.accmGRPC, err = accmgrpc.NewClient(c.AccmClient); err != nil {
		panic(err)
	}
	if s.pangugsGRPC, err = pangugsgrpc.NewClient(c.PanguGSClient); err != nil {
		panic(err)
	}
	if s.shareGRPC, err = sharegrpc.NewClient(c.ShareClient); err != nil {
		panic(err)
	}
	if s.ugcPayGRPC, err = ugcgrpc.NewClient(c.UGCClient); err != nil {
		panic(err)
	}
	if s.resgrpc, err = resgrpc.NewClient(c.ResClient); err != nil {
		panic(err)
	}
	if s.res2grpc, err = res2grpc.NewClient(c.Res2Client); err != nil {
		panic(err)
	}
	if s.creativeGRPC, err = creativegrpc.NewClient(c.CreativeClient); err != nil {
		panic(err)
	}
	if s.thumbupGRPC, err = thumbgrpc.NewClient(c.ThumbupClient); err != nil {
		panic(err)
	}
	if s.upGRPC, err = upgrpc.NewClient(c.UpClient); err != nil {
		panic(err)
	}
	if s.relationGRPC, err = relagrpc.NewClient(c.RelationClient); err != nil {
		panic(err)
	}
	if s.dmGRPC, err = dmgrpc.NewClient(c.DMClient); err != nil {
		panic(err)
	}
	if s.pgcSearchGRPC, err = pgcsearch.NewClient(c.SeasonClient); err != nil {
		panic(err)
	}
	if s.epGRPC, err = epgrpc.NewClient(c.EpClient); err != nil {
		panic(err)
	}
	if s.hisGRPC, err = hisgrpc.NewClient(c.HisClient); err != nil {
		panic(err)
	}
	if s.dyGRPC, err = dygrpc.NewClient(c.DynamicClient); err != nil {
		panic(err)
	}
	if s.ansGRPC, err = ansgrpc.NewClient(c.AnswerClient); err != nil {
		panic(err)
	}
	if s.tagGRPC, err = taggrpc.NewClient(c.TagClient); err != nil {
		panic(err)
	}
	if s.artGRPC, err = artgrpc.NewClient(c.ArtClient); err != nil {
		panic(err)
	}
	if s.couponGRPC, err = cougrpc.NewClient(c.CouponClient); err != nil {
		panic(err)
	}
	if s.tagSvrGRPC, err = tagSvrgrpc.NewClient(c.TagSvrClient); err != nil {
		panic(err)
	}
	if s.steinsGRPC, err = steinsgrpc.NewClient(c.SteinsClient); err != nil {
		panic(err)
	}
	if s.ugcSeasonGRPC, err = ugcseason.NewClient(c.UGCSeasonClient); err != nil {
		panic(err)
	}
	if s.cheeseGRPC, err = cheesegrpc.NewClient(c.SeasonClient); err != nil {
		panic(err)
	}
	if s.locGRPC, err = locgrpc.NewClient(c.LocClient); err != nil {
		panic(err)
	}
	if s.favGRPC, err = favgrpc.NewClient(c.FavClient); err != nil {
		panic(err)
	}
	if s.payRankGRPC, err = payrank.NewClient(c.PayRankClient); err != nil {
		panic(err)
	}
	if s.videoUpGRPC, err = videogrpc.NewClient(c.VideoUpClient); err != nil {
		panic(err)
	}
	if s.roomGRPC, err = roomgrpc.NewClient(c.RoomGRPC); err != nil {
		panic(err)
	}
	if s.gaiaGRPC, err = gaiagrpc.NewClient(c.GaiaGRPC); err != nil {
		panic(err)
	}
	if s.shareAdminGRPC, err = shareadmin.NewClient(c.ShareAdminGRPC); err != nil {
		panic(err)
	}
	if s.garbGRPC, err = garbgrpc.NewClient(c.GarbGRPC); err != nil {
		panic(err)
	}
	if s.upArcGRPC, err = uparcgrpc.NewClient(c.UpArcGRPC); err != nil {
		panic(err)
	}
	if s.comActiveGRPC, err = activegrpc.NewClient(c.CommonActiveGRPC); err != nil {
		panic(err)
	}
	if s.actGRPC, err = actgrpc.NewClient(c.ActGRPC); err != nil {
		panic(err)
	}
	if s.actPlatGRPC, err = actplatgrpc.NewClient(c.ActPlatGRPC); err != nil {
		panic(err)
	}
	if s.popularGRPC, err = populargrpc.NewClient(c.PopularGRPC); err != nil {
		panic(err)
	}
	if s.channelGRPC, err = channelgrpc.NewClient(c.ChannelGRPC); err != nil {
		panic(err)
	}
	if s.seasonGRPC, err = seasongrpc.NewClient(c.SeasonGRPC); err != nil {
		panic(err)
	}
	if s.honorGRPC, err = honorgrpc.NewClient(c.HonorGRPC); err != nil {
		panic(err)
	}
	if s.watchedGRPC, err = watchedGRPC.NewClientWatch(c.WatchedGRPC); err != nil {
		panic(err)
	}
	if s.cfcGRPC, err = cfcgrpc.NewClient(c.CfcGRPC); err != nil {
		panic(err)
	}
	if s.onlineGRPC, err = api.NewClient(c.OnlineGRPC); err != nil {
		panic(err)
	}
	if s.PwdAppealCaptchaDao, err = captcha.NewDao(c.PwdAppeal.Captcha, s.dao.RedisIndex); err != nil {
		panic(err)
	}
	if s.bgroupGRPC, err = bgroupgrpc.NewClient(c.BgroupGRPC); err != nil {
		panic(err)
	}
	s.initActivitySeasonRailGun()
	s.initCron()
	return s
}

func (s *Service) initCron() {
	checkErr(s.loadRegionList())
	checkErr(s.cron.AddFunc(s.c.Cron.RegionList, func() {
		if err := s.loadRegionList(); err != nil {
			log.Error("日志告警 更新分区数据错误,error:%+v", err)
		}
	}))
	// 优先加载分区数据
	s.loadTypeName()
	checkErr(s.cron.AddFunc(s.c.Cron.Type, s.loadTypeName))
	s.loadNewCount()
	checkErr(s.cron.AddFunc(s.c.Cron.NewCount, s.loadNewCount))
	s.loadOnlineTotal()
	checkErr(s.cron.AddFunc(s.c.Cron.OnlineTotal, s.loadOnlineTotal))
	s.loadOnlineList()
	checkErr(s.cron.AddFunc(s.c.Cron.OnlineList, s.loadOnlineList))
	s.loadIndexIcon()
	checkErr(s.cron.AddFunc(s.c.Cron.IndexIcon, s.loadIndexIcon))
	checkErr(s.cron.AddFunc(s.c.Cron.IndexIconRand, s.randomIndexIcon))
	s.loadSearchEgg()
	checkErr(s.cron.AddFunc(s.c.Cron.SearchEgg, s.loadSearchEgg))
	s.loadManager()
	checkErr(s.cron.AddFunc(s.c.Cron.Manager, s.loadManager))
	s.loadSpecRcmd()
	checkErr(s.cron.AddFunc(s.c.Cron.SpecialRcmd, s.loadSpecRcmd))
	s.loadWxHot()
	checkErr(s.cron.AddFunc(s.c.Cron.WxHot, s.loadWxHot))
	s.loadHotLabel()
	checkErr(s.cron.AddFunc(s.c.Cron.WxHot, s.loadHotLabel))
	// stein guide cid
	s.loadGuideCid()
	checkErr(s.cron.AddFunc(s.c.Cron.SteinsGuide, s.loadGuideCid))
	// bnj2019
	checkErr(s.cron.AddFunc(s.c.Cron.Bnj2019, s.loadBnj2019MainArc))
	checkErr(s.cron.AddFunc(s.c.Cron.Bnj2019, s.loadBnj2019ArcList))
	// bnj2020
	checkErr(s.cron.AddFunc(s.c.Cron.Bnj2020, s.loadBnj2020MainView))
	checkErr(s.cron.AddFunc(s.c.Cron.Bnj2020, s.loadBnj2020LiveArc))
	checkErr(s.cron.AddFunc(s.c.Cron.Bnj2020, s.loadBnj2020SpView))
	checkErr(s.cron.AddFunc(s.c.Cron.Bnj2020, s.loadBnj2020ViewList))
	s.loadWebTop()
	checkErr(s.cron.AddFunc(s.c.Cron.WebTop, s.loadWebTop))
	s.loadRankIndex()
	checkErr(s.cron.AddFunc(s.c.Cron.RankIndex, s.loadRankIndex))
	s.loadRankRecommend()
	checkErr(s.cron.AddFunc(s.c.Cron.RankRecommend, s.loadRankRecommend))
	s.loadLpRankRecommend()
	checkErr(s.cron.AddFunc(s.c.Cron.RankRecommend, s.loadLpRankRecommend))
	s.loadRankRegion()
	checkErr(s.cron.AddFunc(s.c.Cron.RankRegion, s.loadRankRegion))
	s.loadRankingV2Data()
	checkErr(s.cron.AddFunc(s.c.Cron.RankV2, s.loadRankingV2Data))
	s.loadParamConfig()
	checkErr(s.cron.AddFunc(s.c.Cron.ParamConfig, s.loadParamConfig))
	s.loadInformationRegionCard()
	checkErr(s.cron.AddFunc(s.c.Cron.InformationRegionCard, s.loadInformationRegionCard))
	s.loadPopularSeries()
	checkErr(s.cron.AddFunc(s.c.Cron.Popular, s.loadPopularSeries))
	checkErr(s.loadSearchDetailCache())
	checkErr(s.cron.AddFunc(s.c.Cron.SearchTipDetail, func() {
		if err := s.loadSearchDetailCache(); err != nil {
			log.Error("%+v", err)
		}
	}))
	if env.DeployEnv == "pre" || env.DeployEnv == "prod" {
		s.loadFawkesVersion()
		checkErr(s.cron.AddFunc(s.c.Cron.Fawkes, s.loadFawkesVersion))
	}
	s.cron.Start()
}

func (s *Service) loadSearchDetailCache() error {
	log.Info("cronLog start loadSearchDetailCache")
	tips, err := s.dao.SearchTipDetail(context.Background())
	if err != nil {
		log.Error("%+v", err)
		return err
	}
	s.searchTipDetailCache = tips
	notices, err := s.dao.SearchSystemNotice(context.Background())
	if err != nil {
		log.Error("日志告警 搜索系统提示加载失败：%+v", err)
		return err
	}
	s.systemNoticeCache = notices
	return nil
}

func (s *Service) initActivitySeasonRailGun() {
	if err := s.loadCommonActivities(); err != nil {
		panic(fmt.Sprintf("initActivitySeasonRailGun error:%+v", err))
	}
	r := railgun.NewRailGun("刷新大型活动", nil, railgun.NewCronInputer(&railgun.CronInputerConfig{Spec: s.c.Cron.ActivitySeason}), railgun.NewCronProcessor(nil, func(ctx context.Context) railgun.MsgPolicy {
		if err := s.loadCommonActivities(); err != nil {
			log.Error("%+v", err)
		}
		return railgun.MsgPolicyNormal
	}))
	s.activitySeasonRailGun = r
	r.Start()
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func (s *Service) loadTypeName() {
	if s.typeNamesRunning {
		return
	}
	s.typeNamesRunning = true
	defer func() {
		s.typeNamesRunning = false
	}()
	typesReply, err := s.arcGRPC.Types(context.Background(), &arcgrpc.NoArgRequest{})
	if err != nil || typesReply == nil || len(typesReply.Types) == 0 {
		log.Error("loadTypeName s.arcGRPC.Types error(%v) or data(%+v) nil", err, typesReply)
		return
	}
	s.typeNames = typesReply.Types
}

func (s *Service) loadSpecRcmd() {
	if s.specRcmdRunning {
		return
	}
	s.specRcmdRunning = true
	defer func() {
		s.specRcmdRunning = false
	}()
	reply, err := s.res2grpc.GetWebSpecialCard(context.Background(), &res2grpc.NoArgRequest{})
	if err != nil {
		log.Error("%+v", err)
		return
	}
	resm := map[int64]*res2grpc.WebSpecialCard{}
	for _, val := range reply.GetCard() {
		val.Person = ""
		val.Ctime = 0
		val.Mtime = 0
		resm[val.Id] = val

	}
	s.specRcmdCard = resm
}

// Ping check connection success.
func (s *Service) Ping(c context.Context) (err error) {
	return
}

// Close close resource.
func (s *Service) Close() {
	s.cron.Stop()
	s.dao.Close()
}

// nolint:gomnd
func archivesArgLog(name string, aids []int64) {
	if aidLen := len(aids); aidLen >= 50 {
		log.Info("s.arc.Archives3 func(%s) len(%d), arg(%v)", name, aidLen, aids)
	}
}

func (s *Service) avToBv(aid int64) (bvID string) {
	var err error
	if bvID, err = bvid.AvToBv(aid); err != nil {
		log.Warn("avToBv(%d) error(%v)", aid, err)
	}
	return
}

func (s *Service) B2i(b bool) int8 {
	if b {
		return 1
	}
	return 0
}

func (s *Service) BatchNFTRegion(ctx context.Context, mids []int64) map[int64]pangugsgrpc.NFTRegionType {
	midNFTRegionMap := make(map[int64]pangugsgrpc.NFTRegionType)
	if len(mids) == 0 {
		return midNFTRegionMap
	}
	nftMidMap := make(map[string]int64)
	if res, e := s.accmGRPC.NFTBatchInfo(ctx, &accmgrpc.NFTBatchInfoReq{Mids: mids, Status: "inUsing", Source: "face"}); e != nil {
		log.Error("s.accmGRPC.NFTBatchInfo(%v) error(%v)", mids, e)
	} else {
		if res == nil || len(res.NftInfos) == 0 {
			return midNFTRegionMap
		}
		var nftIds []string
		for midS, nftInfo := range res.NftInfos {
			if mid, err := strconv.ParseInt(midS, 10, 64); err == nil && mid > 0 && nftInfo != nil && nftInfo.NftId != "" {
				nftMidMap[nftInfo.NftId] = mid
				nftIds = append(nftIds, nftInfo.NftId)
			}
		}
		if res, e := s.pangugsGRPC.GetNFTRegion(ctx, &pangugsgrpc.GetNFTRegionReq{NftId: nftIds}); e != nil {
			log.Error("s.pangugsGRPC.GetNFTRegion(%v) error(%v)", nftIds, e)
		} else {
			if res == nil || len(res.Region) == 0 {
				return midNFTRegionMap
			}
			for nftId, nftRegion := range res.Region {
				mid := nftMidMap[nftId]
				if nftRegion != nil && mid > 0 {
					if nftRegion.Type == pangugsgrpc.NFTRegionType_DEFAULT {
						midNFTRegionMap[mid] = pangugsgrpc.NFTRegionType_MAINLAND
					} else {
						midNFTRegionMap[mid] = nftRegion.Type
					}
				}
			}
		}
	}
	return midNFTRegionMap
}

func filterRepeatedInt64(list []int64) []int64 {
	int64Map := make(map[int64]int64, len(list))
	var filterList []int64
	for _, id := range list {
		if _, ok := int64Map[id]; ok {
			continue
		}
		filterList = append(filterList, id)
		int64Map[id] = id
	}
	return filterList
}

func (s *Service) InfocV2(i interface{}) {
	switch v := i.(type) {
	case model.UserActInfoc:
		payload := infocV2.NewLogStream(s.c.InfocV2LogID.UserActLogID, v.Buvid, v.Build, v.Client, v.Ip, v.Uid, v.Aid, v.Mid, v.Sid, v.Refer, v.Url, v.From, v.ItemID, v.ItemType, v.Action, v.ActionID, v.Ua, v.Ts, v.Extra, v.IsRisk)
		if err := s.infocV2Log.Info(context.Background(), payload); err != nil {
			log.Error("InfocV2 s.infocV2Log.Info userAct err(%+v)", err)
		}
	case model.PopularInfoc:
		payload := infocV2.NewLogStream(s.c.InfocV2LogID.PopularLogID, v.MobiApp, v.Device, v.Build, v.Time, v.LoginEvent, v.Mid, v.Buvid, v.Feed, v.Page, v.Spmid, v.URL, v.Env, v.Trackid, v.IsRec, v.ReturnCode, v.UserFeature, v.Flush)
		if err := s.infocV2Log.Info(context.Background(), payload); err != nil {
			log.Error("InfocV2 s.infocV2Log.Info popular err(%+v)", err)
		}
	default:
		log.Warn("infocproc can't process the type")
	}
}

func (s *Service) InfocRcmd(v *model.RcmdInfoc) {
	payload := infocV2.NewLogStream(s.c.InfocV2LogID.RcmdLogID, v.API, v.IP, v.Mid, v.Buvid, v.Ptype, v.Time, v.FreshType, v.IsRec, v.Trackid, v.ReturnCode, v.UserFeature, v.Showlist, v.IsFeed, v.FreshIdx, v.FreshIdx1h, v.FeedVersion)
	if err := s.infocV2Log.Info(context.Background(), payload); err != nil {
		log.Error("日志告警 首页天马infoc上报错误 data:%+v,err:%+v", v, err)
		return
	}
	log.Warn("首页天马infoc上报成功 data:%+v", v)
}

func (s *Service) popularGetTotalLikeRes(ctx context.Context, mid int64) (*actplatgrpc.GetTotalResResp, error) {
	var (
		group1, group2, group3, group4 int64
	)
	eg := errGroup.WithContext(ctx)
	eg.Go(func(ctx context.Context) (err error) {
		if group1, err = s.popularGetSingleLikeRes(ctx, mid, s.c.PopularActivity.MessageIDsGroup1); err != nil {
			return err
		}
		return nil
	})
	eg.Go(func(ctx context.Context) (err error) {
		if group2, err = s.popularGetSingleLikeRes(ctx, mid, s.c.PopularActivity.MessageIDsGroup2); err != nil {
			return err
		}
		return nil
	})
	eg.Go(func(ctx context.Context) (err error) {
		if group3, err = s.popularGetSingleLikeRes(ctx, mid, s.c.PopularActivity.MessageIDsGroup3); err != nil {
			return err
		}
		return nil
	})
	eg.Go(func(ctx context.Context) (err error) {
		if group4, err = s.popularGetSingleLikeRes(ctx, mid, s.c.PopularActivity.MessageIDsGroup4); err != nil {
			return err
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		log.Error("popularGetTotalLikeRes() %+v", err)
		return nil, err
	}
	return &actplatgrpc.GetTotalResResp{Total: group1 + group2 + group3 + group4}, nil
}

func (s *Service) popularGetSingleLikeRes(ctx context.Context, mid int64, messageIDs []int64) (int64, error) {
	args := &thumbgrpc.HasLikeReq{
		Business:   "archive",
		MessageIds: messageIDs,
		Mid:        mid,
		IP:         metadata.String(ctx, metadata.RemoteIP),
	}
	res, err := s.thumbupGRPC.HasLike(ctx, args)
	if err != nil {
		return 0, errors.Wrapf(err, "popularGetSingleLikeRes() s.thumbupGRPC.HasLike args:%+v", args)
	}
	return resolveHasLikeTotalRes(res), nil
}

func resolveHasLikeTotalRes(res *thumbgrpc.HasLikeReply) int64 {
	var total int64
	for _, v := range res.States {
		total += int64(v.State)
	}
	return total
}

func (s *Service) popularGetTotalCoinRes(ctx context.Context, mid int64) (*actplatgrpc.GetTotalResResp, error) {
	var (
		group1, group2, group3, group4 int64
	)
	eg := errGroup.WithContext(ctx)
	eg.Go(func(ctx context.Context) (err error) {
		res, err := s.dao.IsCoins(ctx, s.c.PopularActivity.MessageIDsGroup1, mid)
		if err != nil {
			return err
		}
		group1 = resolveIsCoinsTotalRes(res)
		return nil
	})
	eg.Go(func(ctx context.Context) (err error) {
		res, err := s.dao.IsCoins(ctx, s.c.PopularActivity.MessageIDsGroup2, mid)
		if err != nil {
			return err
		}
		group2 = resolveIsCoinsTotalRes(res)
		return nil
	})
	eg.Go(func(ctx context.Context) (err error) {
		res, err := s.dao.IsCoins(ctx, s.c.PopularActivity.MessageIDsGroup3, mid)
		if err != nil {
			return err
		}
		group3 = resolveIsCoinsTotalRes(res)
		return nil
	})
	eg.Go(func(ctx context.Context) (err error) {
		res, err := s.dao.IsCoins(ctx, s.c.PopularActivity.MessageIDsGroup4, mid)
		if err != nil {
			return err
		}
		group4 = resolveIsCoinsTotalRes(res)
		return nil
	})
	if err := eg.Wait(); err != nil {
		log.Error("popularGetTotalCoinRes() %+v", err)
		return nil, err
	}
	return &actplatgrpc.GetTotalResResp{Total: group1 + group2 + group3 + group4}, nil
}

func resolveIsCoinsTotalRes(res map[int64]int64) int64 {
	var total int64
	for _, v := range res {
		total += v
	}
	return total
}
