package service

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go-common/component/metadata/device"
	"go-common/component/tinker"
	"go-common/library/cache/redis"
	"go-common/library/conf/env"
	"go-common/library/ecode"
	"go-common/library/exp/ab"
	"go-common/library/log"
	"go-common/library/log/infoc.v2"
	"go-common/library/net/metadata"
	"go-common/library/net/rpc/warden"
	"go-common/library/stat/prom"
	"go-common/library/sync/errgroup.v2"
	"go-common/library/sync/pipeline/fanout"

	xecode "go-gateway/app/app-svr/app-player/ecode"
	arcApi "go-gateway/app/app-svr/archive/service/api"
	feature "go-gateway/app/app-svr/feature/service/sdk"
	v1 "go-gateway/app/app-svr/playurl/service/api"
	v2 "go-gateway/app/app-svr/playurl/service/api/v2"
	"go-gateway/app/app-svr/playurl/service/conf"
	arcdao "go-gateway/app/app-svr/playurl/service/dao/archive"
	cachedao "go-gateway/app/app-svr/playurl/service/dao/cache"
	configdao "go-gateway/app/app-svr/playurl/service/dao/config"
	rightdao "go-gateway/app/app-svr/playurl/service/dao/copyright"
	dmdao "go-gateway/app/app-svr/playurl/service/dao/dm"
	ottdao "go-gateway/app/app-svr/playurl/service/dao/ott"
	pgcdao "go-gateway/app/app-svr/playurl/service/dao/pgc"
	pudao "go-gateway/app/app-svr/playurl/service/dao/playurl"
	resdao "go-gateway/app/app-svr/playurl/service/dao/resource"
	seadao "go-gateway/app/app-svr/playurl/service/dao/season"
	"go-gateway/app/app-svr/playurl/service/dao/taishan"
	ugcpaydao "go-gateway/app/app-svr/playurl/service/dao/ugcpay"
	"go-gateway/app/app-svr/playurl/service/dao/ugcpayrank"
	vipdao "go-gateway/app/app-svr/playurl/service/dao/vip"
	"go-gateway/app/app-svr/playurl/service/model"
	arcmdl "go-gateway/app/app-svr/playurl/service/model/archive"
	taimdl "go-gateway/app/app-svr/playurl/service/model/taishan"
	seasonApi "go-gateway/app/app-svr/ugc-season/service/api"

	accountrpc "git.bilibili.co/bapis/bapis-go/account/service"
	ugcpaymdl "git.bilibili.co/bapis/bapis-go/account/service/ugcpay"
	"git.bilibili.co/bapis/bapis-go/bilibili/app/distribution"
	steampunkgrpc "git.bilibili.co/bapis/bapis-go/pcdn/steampunk"
	bcgrpc "git.bilibili.co/bapis/bapis-go/push/online/broadcast"
	hlsgrpc "git.bilibili.co/bapis/bapis-go/video/vod/playurlhls"
	tvproj "git.bilibili.co/bapis/bapis-go/video/vod/playurltvproj"
	vod "git.bilibili.co/bapis/bapis-go/video/vod/playurlugc"
	volume "git.bilibili.co/bapis/bapis-go/video/vod/playurlvolume"
	vipProfilerpc "git.bilibili.co/bapis/bapis-go/vip/profile/service"
	vipInforpc "git.bilibili.co/bapis/bapis-go/vip/service/vipinfo"
	"git.bilibili.co/go-tool/libbdevice/pkg/pd"

	"github.com/robfig/cron"
	"google.golang.org/grpc"
)

const (
	_qn480               = 32
	_qn720               = 64
	_qn1080H             = 112
	_qnHDR               = 125
	_qnDobly             = 126
	_relationPaid        = "paid"
	_playURLV3           = "/v3/playurl"
	_platformHtml5       = "html5"
	_platformHtml5New    = "html5_new"
	_platformIos         = "ios"
	_sourceProject       = "project"
	_dolbyinfoType       = "ugc"
	_dolbyScene          = "transfer"
	_newDeviceABTestName = "is_hide_playericon"
	_missValue           = "miss"
	_24h                 = 24 * 60 * 60
	//禁用投屏的错误码
	_castCodeVideoDisabled       = 190001
	_castCodeRightDisabled       = 190002
	_castCodeNoVideoPlayResource = 190004
	//禁用后台播放的错误码
	_backgroundCodeRightDisabled      = 190006
	_backgroundCodeVideoDisabled      = 190007
	_backgroundCodeSteinsGateDisabled = 190008
	_backgroundCodeCastConfDisabled   = 190011
	//灰度控制值
	MaxGray                        = 1000
	phoneExpTime                   = "2022-02-14"
	padExpTime                     = "2022-05-26"
	_payArcVersionControlAndroid   = 6750000
	_payArcVersionControlIos       = 67500000
	_payArcVersionControlAndroidHd = 1220000
	_payArcVersionControlIpadHd    = 34400000
	_payArcVersionControlPad       = 67800000
	_premiere                      = 1
)

var (
	tagABTestFlag = ab.String(_newDeviceABTestName, "PlayConf", _missValue)
)

// Service struct
type Service struct {
	c             *conf.Config
	pudao         *pudao.Dao
	arcDao        *arcdao.Dao
	seaDao        *seadao.Dao
	ugcpayDao     *ugcpaydao.Dao
	ugcpayRankDao *ugcpayrank.Dao
	resDao        *resdao.Dao
	pgcDao        *pgcdao.Dao
	ottDao        *ottdao.Dao
	vipDao        *vipdao.Dao
	taishanDao    *taishan.Dao
	dmDao         *dmdao.Dao
	cacheDao      *cachedao.Dao
	confDao       *configdao.Dao
	rightDao      *rightdao.Dao
	// paster
	pasterCache       map[int64]int64
	steinsWhite       map[int64]int64
	authorisedCallers map[string]struct{}
	// cron
	cron        *cron.Cron
	chronosConf []*arcmdl.PlayerReply
	//elec
	allowTypeIds map[int32]struct{}
	//music
	blackMusicMids map[int64]struct{}
	blackMusicAids map[int64]struct{}
	//vip free
	vipFreeAids map[int64]*arcmdl.VipFree
	// chan
	cloudInfoc infoc.Infoc
	// cache chan
	cache *fanout.Fanout
	//account
	accountClient accountrpc.AccountClient
	tinker        *tinker.ABTest
	// feature平台
	FeatureSvc *feature.Feature
	broadcast  bcgrpc.BroadcastVideoClient
	//prom
	promInfo *prom.Prom
	//online black list
	onlineBlackList  map[int64]struct{}
	vipProfileClient vipProfilerpc.VasProfileClient
	distribution     distribution.DistributionClient
	steampunkClient  steampunkgrpc.PcdnClient
}

type verifyArcArg struct {
	Aid          int64
	Cid          int64
	Mid          int64
	Qn           int64
	UpgradeAid   int64
	UpgradeCid   int64
	Platform     string
	Device       string
	MobiApp      string
	VerifySteins int32 //互动视频check参数
	Build        int32
	Source       string //来源 投屏：project 其他：""
	Buvid        string
	VerifyVip    int32 //vip 管控参数
}

type verifyReply struct {
	VipControl *vipInforpc.ControlResult
	Arc        *arcmdl.Info
}

// New init
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:             c,
		pudao:         pudao.New(c),
		arcDao:        arcdao.New(c),
		seaDao:        seadao.New(c),
		ugcpayDao:     ugcpaydao.New(c),
		resDao:        resdao.New(c),
		pgcDao:        pgcdao.New(c),
		ottDao:        ottdao.New(c),
		vipDao:        vipdao.New(c),
		taishanDao:    taishan.New(c),
		dmDao:         dmdao.New(c),
		ugcpayRankDao: ugcpayrank.New(c),
		cacheDao:      cachedao.New(c),
		confDao:       configdao.New(c),
		rightDao:      rightdao.New(c),
		// paster
		pasterCache:       make(map[int64]int64),
		steinsWhite:       make(map[int64]int64),
		authorisedCallers: make(map[string]struct{}),
		chronosConf:       make([]*arcmdl.PlayerReply, 0), //初始化
		//vip free
		vipFreeAids: make(map[int64]*arcmdl.VipFree),
		// cron
		cron:           cron.New(),
		allowTypeIds:   make(map[int32]struct{}),
		blackMusicMids: make(map[int64]struct{}),
		blackMusicAids: make(map[int64]struct{}),
		cache:          fanout.New("cache"),
		// feature平台
		FeatureSvc:      feature.New(nil),
		promInfo:        prom.BusinessInfoCount,
		onlineBlackList: make(map[int64]struct{}),
	}
	for _, id := range c.Custom.ElecShowTypeIDs {
		s.allowTypeIds[id] = struct{}{}
	}
	for _, mid := range c.Custom.MusicMids {
		s.blackMusicMids[mid] = struct{}{}
	}
	for _, aid := range c.Custom.MusicAids {
		s.blackMusicAids[aid] = struct{}{}
	}
	var err error
	if s.cloudInfoc, err = infoc.New(c.InfocConf.CloudInfoc); err != nil {
		panic(err)
	}
	if s.accountClient, err = accountrpc.NewClient(c.AccountClient); err != nil {
		panic(err)
	}
	if s.broadcast, err = bcgrpc.NewClient(c.Broadcast); err != nil {
		panic(fmt.Sprintf("env:%s no BroadcastVideoAPIClient grpc newClient error(%v)", env.DeployEnv, err))
	}
	if s.vipProfileClient, err = vipProfilerpc.NewClient(c.VipProfileClient); err != nil {
		panic(err)
	}
	if s.distribution, err = func(cfg *warden.ClientConfig, opts ...grpc.DialOption) (distribution.DistributionClient, error) {
		client := warden.NewClient(cfg, opts...)
		cc, err := client.Dial(context.Background(), fmt.Sprintf("discovery://default/%s", "app.distribution"))
		if err != nil {
			return nil, err
		}
		return distribution.NewDistributionClient(cc), nil
	}(c.DistributionClient); err != nil {
		panic(err)
	}
	if s.steampunkClient, err = steampunkgrpc.NewClientPcdn(c.SteampunkClient); err != nil {
		panic(err)
	}
	s.tinker = tinker.Init(s.cloudInfoc, nil)
	s.initCron()
	s.cron.Start()
	return
}

func (s *Service) initCron() {
	var err error
	s.loadChronos()
	if err = s.cron.AddFunc(s.c.Cron.LoadChronos, s.loadChronos); err != nil {
		panic(err)
	}
	s.loadPasterCID()
	if err = s.cron.AddFunc(s.c.Cron.LoadPasterCID, s.loadPasterCID); err != nil {
		panic(err)
	}
	s.loadSteinsWhite()
	if err = s.cron.AddFunc(s.c.Cron.LoadSteinsWhite, s.loadSteinsWhite); err != nil {
		panic(err)
	}
	s.loadCustomConfig()
	if err = s.cron.AddFunc(s.c.Cron.LoadCustomConfig, s.loadCustomConfig); err != nil {
		panic(err)
	}
	s.loadOnlineBlackList()
	if err = s.cron.AddFunc(s.c.Cron.LoadManagerConfig, s.loadOnlineBlackList); err != nil {
		panic(err)
	}
	s.loadVipFreeList()
	if err = s.cron.AddFunc(s.c.Cron.LoadVipConfig, s.loadVipFreeList); err != nil {
		panic(err)
	}
}

func (s *Service) loadVipFreeList() {
	res, err := s.resDao.FetchVipList(context.Background())
	if err != nil {
		log.Error("s.resDao.FetchVipList error(%+v)", err)
		return
	}
	s.vipFreeAids = res
	log.Info("loadVipFreeList(%v)", s.vipFreeAids)
}

func (s *Service) loadOnlineBlackList() {
	res, err := s.resDao.FetchAllOnlineBlackList(context.Background())
	if err != nil {
		log.Error("s.resDao.FetchAllOnlineBlackList error(%+v)", err)
		return
	}
	s.onlineBlackList = res
	log.Info("loadOnlineBlackList(%v)", s.onlineBlackList)
}

func (s *Service) loadChronos() {
	rly, e := s.arcDao.PlayerRules(context.Background())
	if e != nil {
		log.Error("loadChronos s.arcDao.PlayerRule error(%v)", e)
		return
	}
	tmp := make([]*arcmdl.PlayerReply, 0)
	for _, v := range rly {
		if v == nil {
			continue
		}
		t := arcmdl.FormatPlayRule(v)
		if t != nil { // t==nil 不满足check条件
			tmp = append(tmp, t)
		}
	}
	s.chronosConf = tmp
	// 查询问题时需要知道当时内存的配置信息
	tmpStr, _ := json.Marshal(tmp)
	log.Info("loadChronos success %s", tmpStr)
}

func (s *Service) loadPasterCID() {
	tmpPaster, err := s.resDao.PasterCID(context.Background())
	if err != nil {
		log.Error("%v", err)
		return
	}
	s.pasterCache = tmpPaster
}

func (s *Service) loadSteinsWhite() {
	arcInfo, err := s.arcDao.GetSimpleArc(context.Background(), s.c.Custom.SteinsWhiteAid, 0, "", "", "")
	if err != nil || arcInfo == nil {
		log.Error("loadSteinsWhite err(%v) or arcInfo = nil", err)
		return
	}
	var tmpWhite = make(map[int64]int64)
	for _, cid := range arcInfo.Cids {
		tmpWhite[cid] = cid
	}
	s.steinsWhite = tmpWhite
}

func (s *Service) loadCustomConfig() {
	tmpCallers := make(map[string]struct{})
	for _, v := range s.c.Custom.SteinsCallers {
		tmpCallers[v] = struct{}{}
	}
	s.authorisedCallers = tmpCallers
}

// PlayURL is
func (s *Service) PlayURL(ctx context.Context, req *v1.PlayURLReq) (reply *v1.PlayURLReply, err error) {
	var (
		code            int
		isSp, isPreview bool
	)
	reply = new(v1.PlayURLReply)
	_, pasterOK := s.pasterCache[req.Cid] //贴片视频白名单
	_, steinsOK := s.steinsWhite[req.Cid] //互动引导视频白名单
	if !pasterOK && !steinsOK {
		vfyArg := &verifyArcArg{
			Aid:          req.Aid,
			Cid:          req.Cid,
			Mid:          req.Mid,
			Qn:           req.Qn,
			Platform:     req.Platform,
			Device:       req.Device,
			MobiApp:      req.MobiApp,
			VerifySteins: req.VerifySteins,
			Build:        req.Build,
		}
		if isPreview, isSp, _, _, err = s.verifyArchive(ctx, vfyArg); err != nil {
			return
		}
		req.Aid = vfyArg.Aid
		req.Cid = vfyArg.Cid
		req.Qn = vfyArg.Qn
	}
	reqURL := s.c.Host.Playurl + _playURLV3
	reply, code, err = s.pudao.Playurl(ctx, req, isSp, isPreview, reqURL)
	if err != nil {
		log.Error("s.pudao.Playurl err(%+v)", err)
		reqURL = s.c.Host.PlayurlBk + _playURLV3
		reply, code, err = s.pudao.Playurl(ctx, req, isSp, isPreview, reqURL)
		if err != nil {
			log.Error("s.pudao.PlayurlBK err(%+v)", err)
			return
		}
	}
	if code != ecode.OK.Code() {
		log.Error("playurl aid(%d) cid(%d) code(%d)", req.Aid, req.Cid, code)
		err = ecode.NothingFound
		reply = nil
	}
	return
}

// PlayConfEdit .
// editSouce 编辑来源 0=用户 1=系统下发
func (s *Service) PlayConfEdit(ctx context.Context, req *v2.PlayConfEditReq, editSource string) (*v2.PlayConfEditReply, error) {
	defer func() {
		dev, _ := device.FromContext(ctx)
		infoMap := make(map[v2.ConfType]*v2.PlayConfState)
		for _, v := range req.PlayConf {
			infoMap[v.ConfType] = v
		}
		infoStr, _ := json.Marshal(infoMap)
		s.infocSave(arcmdl.CloudInfo{
			Ctime:     time.Now().Format("2006-01-02 15:04:05"),
			Buvid:     req.Buvid,
			Platform:  req.Platform,
			FMode:     req.FMode,
			Ver:       strconv.FormatInt(int64(req.Build), 10),
			Function:  string(infoStr),
			Brand:     req.Brand,
			Model:     req.Model,
			EditSouce: editSource,
			FpLocal:   dev.FpLocal,
		})
	}()
	confValueToSave, fieldValueToSave := separateEditedValue(req.PlayConf)
	if err := func() error {
		if len(confValueToSave) == 0 { //没有confValue需要修改
			return nil
		}
		anys, err := convertConfValueToAnys(confValueToSave)
		if err != nil {
			return err
		}
		if _, err = s.distribution.SetUserPreference(ctx, &distribution.SetUserPreferenceReq{
			Preference: anys,
		}); err != nil {
			return err
		}
		return nil
	}(); err != nil {
		return nil, err
	}
	if len(fieldValueToSave) == 0 { //没有fieldValue需要修改
		return &v2.PlayConfEditReply{}, nil
	}
	// 获取cache内的数据
	playc, e := s.taishanDao.PlayConfGet(ctx, req.Buvid)
	if e != nil {
		log.Error("PlayConfEdit error(%v)", e)
		return nil, e
	}
	var (
		playC map[int64]*taimdl.PlayConf
	)
	if playc != nil && playc.PlayConfs != nil { //taishan中没有值，不可能是新设备命中实验
		playC = playc.PlayConfs
		s.newDeviceEditor(ctx, req.Buvid)
	}
	if playC == nil {
		playC = make(map[int64]*taimdl.PlayConf)
	}
	for _, v := range req.PlayConf {
		if v == nil {
			continue
		}
		playC[int64(v.ConfType)] = taimdl.ConvertConf(v)
	}
	if e := s.taishanDao.PlayConfSet(ctx, &taimdl.PlayConfs{PlayConfs: playC}, req.Buvid); e != nil {
		log.Error("s.taishanDao.PlayConfSet(%s) error(%v)", req.Buvid, e)
		return nil, e
	}
	return &v2.PlayConfEditReply{}, nil
}

func (s *Service) newDeviceEditor(ctx context.Context, buvid string) {
	hasChanged, err := s.cacheDao.GetConfABTest(ctx, buvid)
	if err == nil && !hasChanged {
		if err = s.cacheDao.SetConfABTest(ctx, buvid, 1); err != nil {
			log.Error("PlayConfEdit s.cacheDao.SetConfABTest error(%+v)", err)
		}
	}
}

// PlayConf .
func (s *Service) PlayConf(c context.Context, req *v2.PlayConfReq) (*v2.PlayConfReply, error) {
	var (
		reply             = &v2.PlayConfReply{}
		hit, renew        bool
		distributionReply *distribution.GetUserPreferenceReply
	)
	var playConfMap map[int64]*taimdl.PlayConf
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) error {
		playc, e := s.taishanDao.PlayConfGet(ctx, req.Buvid)
		if e == nil && playc != nil {
			playConfMap = playc.PlayConfs
		}
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		if s.newDeviceSwitchOn(req.Platform, req.Build) {
			hit = s.hitExpFirstTime(ctx, req)
			renew = s.hitNewDeviceRenew(ctx, req)
		}
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		res, err := s.distribution.GetUserPreference(ctx, &distribution.GetUserPreferenceReq{
			TypeUrl: []string{_distributionPlayConf, _distributionCloudPlayConf},
		})
		if err != nil {
			log.Error("PlayConf s.distribution.GetUserPreference error(%+v), buvid(%s)", err, req.Buvid)
			return nil
		}
		distributionReply = res
		return nil
	})
	if err := eg.Wait(); err != nil {
		log.Error("eg.Wait(%+v)", err)
		return nil, err
	}
	reply.PlayConf = s.chooseCloudConf(c, hit, renew, playConfMap, distributionReply)
	return reply, nil
}

func (s *Service) newDeviceSwitchOn(platform string, build int32) bool {
	return s.c.Custom.NewDeviceSwitchon &&
		((platform == "android" && build > s.c.BuildLimit.NewDeviceAndBuild) || (platform == "ios" && build > s.c.BuildLimit.NewDeviceIOSBuild))
}

func (s *Service) chooseCloudConf(ctx context.Context, hit, renew bool, tp map[int64]*taimdl.PlayConf, distributionReply *distribution.GetUserPreferenceReply) *v2.PlayAbilityConf {
	//nolint:ineffassign
	var (
		playConf  = &v2.PlayAbilityConf{}
		testValue *v2.CloudConf
	)
	playConf = s.fromAbilityConf(ctx, tp, distributionReply)
	if !hit && !renew {
		return playConf
	}
	if hit {
		testValue = &v2.CloudConf{Show: false, FieldValue: &v2.FieldValue{Value: &v2.FieldValue_Switch{Switch: false}}}
	}
	if renew {
		testValue = &v2.CloudConf{Show: true, FieldValue: &v2.FieldValue{Value: &v2.FieldValue_Switch{Switch: true}}}
	}
	playConf.DislikeConf = testValue    //踩
	playConf.CoinConf = testValue       //投币
	playConf.ElecConf = testValue       //充电
	playConf.ScreenShotConf = testValue //截图/gif
	return playConf
}

func (s *Service) hitNewDeviceRenew(c context.Context, req *v2.PlayConfReq) bool {
	if req.Mid <= 0 {
		return false
	}
	//登录用户，需要判断是否是新设备做了实验之后登录
	hasChanged, err := s.cacheDao.GetConfABTest(c, req.Buvid)
	if err != nil || hasChanged {
		return false
	}
	// 是新设备且没有编辑过，展示之前隐藏的云控信息,并设置redis为编辑过,下次进来就不会重新全部展示
	if err := s.cache.Do(c, func(c context.Context) {
		if _, err := s.PlayConfEdit(c, &v2.PlayConfEditReq{PlayConf: taimdl.GrayDefault(true), Buvid: req.Buvid, Platform: req.Platform, Build: req.Build, Brand: req.Brand, Model: req.Model, FMode: req.FMode}, "1"); err != nil {
			log.Error("PlayConf s.PlayConfEdit error(%+v)", err)
			return
		}
		if err = s.cacheDao.SetConfABTest(c, req.Buvid, 1); err != nil {
			log.Error("PlayConfEdit s.cacheDao.SetConfABTest")
		}
	}); err != nil {
		log.Error("PlayConf s.cache.Do error(%+v)", err)
	}
	return true
}

func (s *Service) hitExpFirstTime(c context.Context, req *v2.PlayConfReq) bool {
	if req.Mid > 0 {
		return false
	}
	reply, err := s.accountClient.CheckRegTime(c, &accountrpc.CheckRegTimeReq{Buvid: req.Buvid, Periods: "0-48"})
	if err != nil {
		log.Error("PlayConf s.accountClient.CheckRegTime error(%+v)", err)
		return false
	}
	if !reply.GetHit() {
		return false
	}
	hit := s.newDeviceABTestRun(c, req.Buvid, tagABTestFlag)
	if hit != "1" {
		return false
	}
	// 成功说明是新设备且第一次命中实验
	set, err := s.cacheDao.SetNXConfABTest(c, req.Buvid, 0)
	if err != nil {
		if err == redis.ErrNil {
			return false
		}
		log.Error("s.cacheDao.SetNXConfABTest error(%+v)", err)
		return false
	}
	if !set {
		return false
	}
	// 异步初始化taishan缓存内的值
	if err := s.cache.Do(c, func(c context.Context) {
		if _, err := s.PlayConfEdit(c, &v2.PlayConfEditReq{PlayConf: taimdl.GrayDefault(false), Buvid: req.Buvid, Platform: req.Platform, Build: req.Build, Brand: req.Brand, Model: req.Model, FMode: req.FMode}, "1"); err != nil {
			log.Error("PlayConf s.PlayConfEdit error(%+v)", err)
			return
		}
	}); err != nil {
		log.Error("PlayConf s.cache.Do error(%+v)", err)
	}
	return true
}

// newUserABTest
func (s *Service) newDeviceABTestRun(ctx context.Context, buvid string, flag *ab.StringFlag) string {
	var (
		exp     string
		groupID int64
	)
	t, ok := ab.FromContext(ctx)
	if !ok {
		return _missValue
	}
	t.Add(ab.KVString("buvid", buvid))
	exp = flag.Value(t)
	if exp == _missValue {
		return _missValue
	}
	for _, state := range t.Snapshot() {
		if state.Type == ab.ExpHit {
			groupID = state.Value
			break
		}
	}
	s.infocSave(model.LitePlayerInfoc{
		Buvid:    buvid,
		GroupID:  groupID,
		JoinTime: time.Now().Unix(),
	})
	return exp
}

func (s *Service) conformNewDeviceWithPeriods(in *accountrpc.CheckRegTimeReply, buvid string, periodsWithExpTime map[string]string) (bool, bool) {
	var (
		isNewDevice, padIsNewDevice bool
	)
	if _, ok := s.c.NewDeviceWhiteList[buvid]; ok {
		return true, true
	}
	//不在白名单中看接口返回
	if !in.Hit { //不是从实验开始之后的新设备
		return false, false
	}
	for _, v := range in.HitRules {
		if periodsWithExpTime[phoneExpTime] == v {
			isNewDevice = true
			continue
		}
		if periodsWithExpTime[padExpTime] == v {
			padIsNewDevice = true
			continue
		}
		log.Warn("Failed to match periods (%s)", v)
	}
	return isNewDevice, padIsNewDevice
}

//nolint:gomnd
func steamHash(str, salt string) int64 {
	// md5
	b := []byte(str)
	s := []byte(salt)
	h := md5.New()
	h.Write(b) // 先写盐值
	h.Write(s)
	src := h.Sum(nil)

	var dst = make([]byte, hex.EncodedLen(len(src)))
	hex.Encode(dst, src)

	// to upper
	dst = bytes.ToUpper(dst)

	// javahash
	var n int32 = 0
	for i := 0; i < len(dst); i++ {
		n = n*31 + int32(dst[i])
	}

	// mod 1000
	if n >= 0 {
		return int64(n % 1000)
	} else {
		return int64(n%1000 + 1000)
	}
}

// PlayView .
// nolint:gocognit
func (s *Service) PlayView(ctx context.Context, req *v2.PlayViewReq) (reply *v2.PlayViewReply, err error) {
	var (
		playRly *v2.PlayURLReply
		vRly    *verifyReply
		arc     *arcmdl.Info
		vol     *volume.VolumeItem
		vipConf *v2.VipConf
	)
	reply = &v2.PlayViewReply{PlayUrl: &v2.PlayUrlInfo{}}
	// 暂时关闭版本控制，等待云控客户端接入新编辑接口在开启
	needAbility := true
	// ios和安卓粉版不需要在view接口下发云控信息,减少无效调用
	//其余版本暂无法确定
	// feature NeedAbility
	if (req.MobiApp == "iphone" && req.Build >= 10020) || (req.MobiApp == "android" && req.Build >= 6020000) {
		needAbility = false
	}
	eg := errgroup.WithContext(ctx)
	var bpcdnInfo map[string]string
	if steamHash(req.Buvid, "2233") < s.c.Custom.PCDNGrey {
		eg.Go(func(ctx context.Context) error {
			res, err := s.steampunkClient.GetUrlsByCid(ctx, &steampunkgrpc.PlayRequest{
				Cid:      uint64(req.Cid),
				Mid:      uint64(req.Mid),
				Qn:       uint32(req.Qn),
				Platform: req.Platform,
				Uip:      metadata.String(ctx, metadata.RemoteIP),
			})
			if err != nil {
				log.Error("s.steampunkClient.GetUrlsByCid error %+v, cid:%d, mid:%d", err, req.Cid, req.Mid)
				return nil
			}
			if res.Code != 0 {
				log.Error("s.steampunkClient.GetUrlsByCid code is %d message %s cid:%d, mid:%d", res.Code, res.Message, req.Cid, req.Mid)
				return nil
			}
			bpcdnInfo = res.Data
			return nil
		})
	}
	var distributionReply *distribution.GetUserPreferenceReply
	eg.Go(func(c context.Context) error {
		if !canOutputPlayConf(c) {
			return nil
		}
		res, disErr := s.distribution.GetUserPreference(c, &distribution.GetUserPreferenceReq{
			TypeUrl: []string{_distributionPlayConf, _distributionCloudPlayConf},
		})
		if disErr != nil {
			log.Error("PlayView s.distribution.GetUserPreference error(%+v), buvid(%s)", disErr, req.Buvid)
			return nil
		}
		distributionReply = res
		return nil
	})
	eg.Go(func(c context.Context) (e error) {
		disableDolby := false
		if req.TeenagersMode == 1 || req.LessonsMode == 1 { //青少年模式和课堂模式不支持杜比
			disableDolby = true
		}
		var gMsg *model.GlanceMsg
		if playRly, vRly, gMsg, vipConf, e = s.PlayURLV3(c, req, true, disableDolby); e != nil {
			log.Error("PlayView s.PlayURLV3(%d,%d) error(%v)", req.Aid, req.Cid, e)
			if ecode.EqualError(xecode.PlayURLSteinsUpgrade, e) {
				reply.PlayUrl.IsSteinsUpgrade = 1
				e = nil
			}
			if ecode.EqualError(xecode.PlayURLArcPayUpgrade, e) {
				reply.PlayUrl.IsSteinsUpgrade = 2
				//兼容675安卓升级面板
				reply.PlayUrl.Playurl = s.buggyAndroidDegreeForPayArc(c, req)
				e = nil
			}
			// 其他错误抛出，直接返回
			return
		}
		if playRly != nil {
			reply.PlayUrl.Playurl = playRly.Playurl
		}
		reply.PlayUrl.ExtInfo = &v2.ExtInfo{}
		if vRly != nil && vRly.VipControl != nil {
			reply.PlayUrl.ExtInfo.VipControl = &v2.VipControl{Control: vRly.VipControl.Control, Msg: vRly.VipControl.Msg}
		}
		//根据视频云返回值再次判断是否可试看
		reply.Ab = s.buildPlayViewAB(gMsg, reply.PlayUrl.Playurl.AcceptQuality)
		reply.VipConf = vipConf
		return
	})
	// 获取是否有字幕,默认有字幕
	hasFont := true
	eg.Go(func(c context.Context) error {
		dmrly, e := s.dmDao.SubtitleExist(c, req.Cid)
		if e != nil {
			log.Error("s.dmDao.SubtitleExist(%d) error(%v)", req.Cid, e)
			return nil
		}
		if dmrly != nil {
			hasFont = dmrly.Exist
		}
		return nil
	})
	// 获取视频是否有振动
	var shakeURL string
	eg.Go(func(c context.Context) (err error) {
		if shakeURL, err = s.confDao.ShakeConfig(c, req.Aid, req.Cid); err != nil {
			log.Error("s.confDao.ShakeConfig aid(%d) cid(%d) error(%v)", req.Aid, req.Cid, err)
			return nil
		}
		return nil
	})
	// 获取缓存信息
	var playConf map[int64]*taimdl.PlayConf
	if needAbility {
		eg.Go(func(c context.Context) error {
			playc, e := s.taishanDao.PlayConfGet(c, req.Buvid)
			if e == nil && playc != nil {
				playConf = playc.PlayConfs
			}
			return nil
		})
	}
	//获取版权中台对云控：后台播放，小窗，投屏的控制能里
	var rightConf *model.CopyRightRestriction
	eg.Go(func(c context.Context) error { //获取失败降级处理
		var e error
		if rightConf, e = s.rightDao.PlayRestriction(c, req.Aid); e != nil { //获取错误降级处理
			log.Error("s.rightDao.PlayRestriction(%d) error(%v)", req.Aid, e)
		}
		if rightConf == nil {
			rightConf = &model.CopyRightRestriction{Aid: req.Aid} //兜底不禁止后台播放,小窗,投屏
		}
		return nil
	})
	// 获取音量均衡信息, 获取失败降级处理
	if !s.c.Custom.PlayurlVolumeSwitch && req.VoiceBalance == 1 {
		eg.Go(func(c context.Context) error {
			var e error
			if vol, e = s.pudao.PlayurlVolume(c, uint64(req.Cid), uint64(req.Mid)); e != nil {
				log.Error("PlayView s.pudao.PlayurlVolume(%d,%d) error(%v)", req.Aid, req.Cid, e)
			}
			return nil
		})
	}
	// 播放地址获取失败，直接返回结果
	if err = eg.Wait(); err != nil {
		log.Error("PlayView wait aid:%d,cid:%d error(%v)", req.Aid, req.Cid, err)
		return
	}
	joinPCDNToPlayUrlInfo(bpcdnInfo, reply.PlayUrl)
	if vol != nil && vol.MeasuredI > -9 {
		reply.Volume = &v2.VolumeInfo{
			MeasuredI:         vol.MeasuredI,
			MeasuredLra:       vol.MeasuredLra,
			MeasuredTp:        vol.MeasuredTp,
			MeasuredThreshold: vol.MeasuredThreshold,
			TargetOffset:      vol.TargetOffset,
			TargetI:           vol.TargetI,
			TargetTp:          vol.TargetTp,
		}
	}
	// 如果是下载请求，只需要播放地址相关信息即可
	if req.Download > 0 {
		return
	}
	// chronos配置信息获取,纯数据校验
	reply.Chronos = s.checkChronos(arcmdl.FromPlayViewReq(req))
	//PlayURLV3存在不调用稿件信息的逻辑,重新获取，减少调用稿件信息次数
	if vRly != nil && vRly.Arc != nil {
		arc = vRly.Arc
	} else {
		s.promInfo.Incr("PlayView:vRly.Arc为空获取稿件")
		if arc, err = s.arcDao.GetSimpleArc(ctx, req.Aid, req.Mid, req.MobiApp, req.Device, req.Platform); err != nil {
			log.Error("PlayView SimpleArcService arg(%+v) err(%+v)", req.Aid, err)
			err = nil
		}
	}
	//不是青少年模式 && 不是课堂模式
	notTeenAndLesson := req.TeenagersMode != 1 && req.LessonsMode != 1
	//电视投屏 支持投屏&不是互动视频&不是pgc&&版权中台没有限制投屏
	hasCast, disabledCode, disabledReason := func() (bool, int64, string) {
		if arc != nil && (arc.IsSteinsGate() || arc.IsPGC()) {
			return false, _castCodeVideoDisabled, s.c.CastDisabledMsg[strconv.FormatInt(_castCodeVideoDisabled, 10)]
		}
		if rightConf == nil || rightConf.BanMiracast {
			return false, _castCodeRightDisabled, s.c.CastDisabledMsg[strconv.FormatInt(_castCodeRightDisabled, 10)]
		}
		if reply.PlayUrl.Playurl != nil && !reply.PlayUrl.Playurl.VideoProject {
			return false, _castCodeNoVideoPlayResource, s.c.CastDisabledMsg[strconv.FormatInt(_castCodeNoVideoPlayResource, 10)]
		}
		return true, 0, ""
	}()
	// 拼接云控信息
	if needAbility {
		reply.PlayConf = s.fromAbilityConf(ctx, playConf, nil)
		//投屏需要做版本控制，兼容54版本
		if !hasCast && reply.PlayConf != nil {
			reply.PlayConf.CastConf = nil
		}
	}
	reply.PlayArc = &v2.PlayArcConf{}
	// 没有稿件信息，不处理ArcConf
	if arc == nil {
		return
	}

	userPaid, isArcPay, inSeasonFreeWatch := checkArcPayPlay(arc)
	//非付费视频、或者用户已付费、或者免费试看
	normalPlay := !isArcPay || userPaid || inSeasonFreeWatch

	// 是否展示充电,默认false和view接口保持一致
	egV2 := errgroup.WithContext(ctx)
	var hasElc bool
	egV2.Go(func(c context.Context) error {
		hasElc = s.initElec(c, arc, req, normalPlay)
		return nil
	})

	var season *seasonApi.Season
	if arc.SeasonID > 0 {
		egV2.Go(func(c context.Context) (e error) {
			if season, e = s.seaDao.Season(c, arc.SeasonID); e != nil {
				log.Error("s.seaDao.Season(%d) error(%v)", arc.SeasonID, e)
				e = nil
			}
			return
		})
	}
	var (
		isNewDevice    bool
		padIsNewDevice bool
	)
	egV2.Go(func(ctx context.Context) error {
		durations := batchGetDurationBetweenExpAndNow(phoneExpTime, padExpTime)
		periods, periodsWithExpTime := buildPeriods(durations)
		accReply, e := s.accountClient.CheckRegTime(ctx, &accountrpc.CheckRegTimeReq{Buvid: req.Buvid, Periods: periods})
		if e != nil {
			log.Error("b.CheckNewDevice error(%+v), buvid(%s)", e, req.Buvid)
			return nil
		}
		isNewDevice, padIsNewDevice = s.conformNewDeviceWithPeriods(accReply, req.Buvid, periodsWithExpTime)
		return nil
	})
	if errV2 := egV2.Wait(); errV2 != nil { //降级处理，不影响主流程
		log.Error("PlayView egV2.Wait aid:%d,cid:%d error(%v)", req.Aid, req.Cid, errV2)
	}
	expOpts := []ExpOption{
		WithBackgroundExp(isNewDevice),
		WithBackgroundExpForPad(padIsNewDevice),
	}
	expCfg := expConfig{}
	expCfg.Apply(expOpts...)
	ctx = WithContext(ctx, expCfg)
	reply.PlayArc.ElecConf = &v2.ArcConf{IsSupport: hasElc}
	// 后台播放 不是青少年模式&&不是Sony类视频(可用后台播放)&&不是互动视频&&版权中台没有禁止
	hasBack, backgroundDisabledCode, backgroundDisabledReason := func() (bool, int64, string) {
		if rightConf == nil || rightConf.BanBackend {
			return false, _backgroundCodeRightDisabled, s.c.BackgroundDisabledMsg[strconv.FormatInt(_backgroundCodeRightDisabled, 10)]
		}
		if arc.IsSteinsGate() {
			return false, _backgroundCodeSteinsGateDisabled, s.c.BackgroundDisabledMsg[strconv.FormatInt(_backgroundCodeSteinsGateDisabled, 10)]
		}
		if arc.IsNoBackground() {
			return false, _backgroundCodeVideoDisabled, s.c.BackgroundDisabledMsg[strconv.FormatInt(_backgroundCodeVideoDisabled, 10)]
		}
		return true, 0, ""
	}()
	reply.PlayArc.BackgroundPlayConf = &v2.ArcConf{
		IsSupport:      hasBack && notTeenAndLesson,
		UnsupportScene: []int64{_premiere},
	}
	if canOutputNoBackgroundReason(req.MobiApp, req.Device, int64(req.Build)) {
		reply.PlayArc.BackgroundPlayConf = &v2.ArcConf{
			IsSupport: notTeenAndLesson,
			Disabled:  !hasBack,
			ExtraContent: &v2.ExtraContent{
				DisabledReason: backgroundDisabledReason,
				DisabledCode:   backgroundDisabledCode,
			},
			UnsupportScene: []int64{_premiere},
		}
	}
	//镜像反转
	reply.PlayArc.FlipConf = &v2.ArcConf{IsSupport: true}
	//投屏
	if !normalPlay {
		hasCast = false
	}
	reply.PlayArc.CastConf = &v2.ArcConf{
		IsSupport:      hasCast,
		UnsupportScene: []int64{_premiere},
	}
	if (req.MobiApp == "android" && req.Build >= 6480000) || (req.MobiApp == "iphone" && req.Device == "phone" && req.Build >= 64700000) {
		reply.PlayArc.CastConf = &v2.ArcConf{
			IsSupport: normalPlay,
			Disabled:  !hasCast,
			ExtraContent: &v2.ExtraContent{
				DisabledReason: disabledReason,
				DisabledCode:   disabledCode,
			},
			UnsupportScene: []int64{_premiere},
		}
	}
	// 反馈
	reply.PlayArc.FeedbackConf = &v2.ArcConf{IsSupport: true, UnsupportScene: []int64{_premiere}}
	// 字幕
	reply.PlayArc.SubtitleConf = &v2.ArcConf{IsSupport: hasFont}
	//播放速度
	reply.PlayArc.PlaybackRateConf = &v2.ArcConf{IsSupport: true, UnsupportScene: []int64{_premiere}}
	//定时停止播放
	reply.PlayArc.TimeUpConf = &v2.ArcConf{IsSupport: true, UnsupportScene: []int64{_premiere}}
	//基础合集判断
	var isBaseSeason bool
	if season != nil && season.AttrVal(seasonApi.AttrSnType) == seasonApi.AttrSnYes {
		isBaseSeason = true
	}
	// 播放方式 不是基础合集&&不是互动视频
	hasPlayType := false
	if !arc.IsSteinsGate() && !isBaseSeason {
		hasPlayType = true
	}
	//59版本支持合集连播下发
	if isBaseSeason && ((req.MobiApp == "iphone" && req.Build >= 65900100) || (req.MobiApp == "android" && req.Build >= 6590100) || (req.MobiApp == "android_hd" && req.Build >= 1150000)) {
		hasPlayType = true
	}
	reply.PlayArc.PlaybackModeConf = &v2.ArcConf{IsSupport: hasPlayType, UnsupportScene: []int64{_premiere}}
	//小窗 互动视频不显示 && ios下特定音乐相关稿件不显示 && 版权中台没有限制
	hasPlayBack := false
	//特定音乐相关稿件
	isMusicConf := s.isSpecificMusicConf(arc.Mid, arc.Aid)
	if !arc.IsSteinsGate() && !(req.Platform == _platformIos && isMusicConf) && (rightConf != nil && !rightConf.BanPip) && normalPlay {
		hasPlayBack = true
	}
	//小窗
	reply.PlayArc.SmallWindowConf = &v2.ArcConf{IsSupport: hasPlayBack, UnsupportScene: []int64{_premiere}}
	// 画面尺寸
	reply.PlayArc.ScaleModeConf = &v2.ArcConf{IsSupport: true}
	// 顶
	reply.PlayArc.LikeConf = &v2.ArcConf{IsSupport: notTeenAndLesson}
	// 踩
	reply.PlayArc.DislikeConf = &v2.ArcConf{IsSupport: notTeenAndLesson}
	//投币
	reply.PlayArc.CoinConf = &v2.ArcConf{IsSupport: notTeenAndLesson}
	//不是青少年模式&&不是仅收藏可见的&&不是稿件自见
	hasShare := false
	if notTeenAndLesson && arc.AttrValV2(arcApi.AttrBitV2OnlyFavView) != arcApi.AttrYes && arc.AttrValV2(arcApi.AttrBitV2OnlySely) != arcApi.AttrYes {
		hasShare = true
	}
	//分享
	reply.PlayArc.ShareConf = &v2.ArcConf{IsSupport: hasShare}
	//截图/gif
	reply.PlayArc.ScreenShotConf = &v2.ArcConf{IsSupport: hasShare && (!isArcPay || inSeasonFreeWatch)}
	//锁屏
	reply.PlayArc.LockScreenConf = &v2.ArcConf{IsSupport: notTeenAndLesson}
	// 相关推荐 ios不支持||青少年模式
	hasRecommend := false
	if req.MobiApp != "iphone" && req.MobiApp != "ipad" && notTeenAndLesson {
		hasRecommend = true
	}
	reply.PlayArc.RecommendConf = &v2.ArcConf{IsSupport: hasRecommend}
	//倍速
	reply.PlayArc.PlaybackSpeedConf = &v2.ArcConf{IsSupport: true}
	//清晰度
	reply.PlayArc.DefinitionConf = &v2.ArcConf{IsSupport: true}
	//下一集|选集 不是互动视频&&（仅多P视频or(剧集&非pad)or番剧）内容的播放器才有
	hasNext := false
	if !arc.IsSteinsGate() && (len(arc.Cids) > 1 || (arc.SeasonID > 0 && req.Device != "pad") || arc.IsPGC()) {
		hasNext = true
	}
	reply.PlayArc.SelectionsConf = &v2.ArcConf{IsSupport: hasNext}
	reply.PlayArc.NextConf = &v2.ArcConf{IsSupport: hasNext}
	//编辑弹幕（面板和设置都展示）
	reply.PlayArc.EditDmConf = &v2.ArcConf{IsSupport: notTeenAndLesson}
	//弹幕设置-面板上
	reply.PlayArc.OuterDmConf = &v2.ArcConf{IsSupport: notTeenAndLesson}
	//弹幕设置-三点内
	reply.PlayArc.InnerDmConf = &v2.ArcConf{IsSupport: notTeenAndLesson}
	//视频震动
	hasShake := false
	if shakeURL != "" {
		hasShake = true
		reply.Event = &v2.Event{
			Shake: &v2.Shake{File: shakeURL},
		}
	}
	reply.PlayArc.ShakeConf = &v2.ArcConf{IsSupport: hasShake}
	//全景
	reply.PlayArc.PanoramaConf = &v2.ArcConf{IsSupport: arc.Is360()}
	reply.PlayArc.ColorFilterConf = &v2.ArcConf{
		IsSupport: !isIOSShieldColorFilter(req.Device, req.MobiApp, int64(req.Build), reply.PlayUrl.GetPlayurl().GetAcceptQuality()),
	}
	//杜比
	supportDolby := false
	dolby := reply.PlayUrl.GetPlayurl().GetDash().GetDolby()
	if notTeenAndLesson {
		if dolby != nil {
			supportDolby = true
		}
		s.dolbyInfoc(dolby, reply.PlayUrl.GetPlayurl().GetAcceptQuality(), req)
	}

	reply.PlayArc.DolbyConf = &v2.ArcConf{IsSupport: supportDolby}
	//无损
	supportLoss := false
	lossless := reply.GetPlayUrl().GetPlayurl().GetDash().GetLossLessItem().GetIsLosslessAudio()
	if lossless {
		supportLoss = true
	}
	reply.PlayArc.LossLessConf = &v2.ArcConf{IsSupport: supportLoss}
	//屏幕录制
	var supportRecord bool
	if (!isArcPay || inSeasonFreeWatch) && isSupportScreenRecording(arc) {
		supportRecord = true
	}
	reply.PlayArc.ScreenRecordingConf = &v2.ArcConf{
		IsSupport: supportRecord,
	}
	if canOutputPlayConf(ctx) && !arc.IsPGC() {
		reply.PlayConf = s.fromAbilityConf(ctx, nil, distributionReply)
	}
	//首映稿件+ 首映前 || 首映中 则不返回：后台播放、电视投屏、播放反馈、小窗播放、定时提醒
	if (req.MobiApp == "android" && req.Build < 6850001) || (req.MobiApp == "iphone" && req.Device == "phone" && req.Build < 68500001) {
		if s.PremiereValidate(arc) {
			s.PremierePlayConf(reply)
		}
	}
	return
}

func joinPCDNToPlayUrlInfo(bpcdnInfo map[string]string, playUrlInfo *v2.PlayUrlInfo) {
	if len(bpcdnInfo) == 0 {
		return
	}
	if playUrlInfo == nil || playUrlInfo.Playurl == nil || playUrlInfo.Playurl.Dash == nil || len(playUrlInfo.Playurl.Dash.Video) == 0 {
		return
	}
	for _, v := range playUrlInfo.Playurl.Dash.Video {
		bpcdn, ok := bpcdnInfo[fmt.Sprintf("%d_%d", v.Id, v.Codecid)]
		if !ok {
			continue
		}
		params := url.Values{}
		params.Add("bpcdn", bpcdn)
		paramStr := params.Encode()
		// 重新encode的时候空格变成了+号问题修复
		if strings.IndexByte(paramStr, '+') > -1 {
			paramStr = strings.Replace(paramStr, "+", "%20", -1)
		}
		v.BaseUrl = fmt.Sprintf("%s&%s", v.BaseUrl, paramStr)
	}
}

func (s *Service) dolbyInfoc(dolbyItem *v2.DolbyItem, qnList []uint32, req *v2.PlayViewReq) {
	var (
		dolbyType      int
		hasDolbyAudio  bool
		hasDolbyQn     bool
		onlyDolbyAudio = 3
		onlyDolbyQn    = 2
		wholeDolby     = 1
	)
	if dolbyItem != nil && len(dolbyItem.Audio) != 0 {
		hasDolbyAudio = true
		dolbyType = onlyDolbyAudio
	}
	for _, v := range qnList {
		if v == _qnDobly {
			hasDolbyQn = true
			dolbyType = onlyDolbyQn
		}
	}
	if hasDolbyAudio && hasDolbyQn {
		dolbyType = wholeDolby
	}
	if !hasDolbyAudio && !hasDolbyQn {
		return
	}
	//上报下发杜比音频流
	s.infocSave(model.DolbyInfo{
		Build:     req.Build,
		Buvid:     req.Buvid,
		Mid:       req.Mid,
		Ctime:     time.Now().Format("2006-01-02 15:04:05"),
		MobiApp:   req.MobiApp,
		Platform:  req.Platform,
		Aid:       req.Aid,
		Cid:       req.Cid,
		Type:      _dolbyinfoType,
		Scene:     _dolbyScene,
		DolbyType: int64(dolbyType),
	})
}

//首映稿件 + 首映前 || 首映中
func (s *Service) PremiereValidate(arc *arcmdl.Info) bool {
	if arc.AttrValV2(arcApi.AttrBitV2Premiere) == arcApi.AttrYes && arc.Premiere != nil &&
		(arc.Premiere.State == arcApi.PremiereState_premiere_before || arc.Premiere.State == arcApi.PremiereState_premiere_in) {
		return true
	}
	return false
}

func (s *Service) PremierePlayConf(playConf *v2.PlayViewReply) {
	if playConf.PlayArc != nil {
		//后台播放
		playConf.PlayArc.BackgroundPlayConf = &v2.ArcConf{IsSupport: false, Disabled: true}
		//电视投屏
		playConf.PlayArc.CastConf = &v2.ArcConf{IsSupport: false, Disabled: true, ExtraContent: &v2.ExtraContent{
			DisabledReason: "首映禁用",
			DisabledCode:   _backgroundCodeCastConfDisabled,
		}}
		//播放反馈
		playConf.PlayArc.FeedbackConf = &v2.ArcConf{IsSupport: false, Disabled: true}
		//小窗播放
		playConf.PlayArc.SmallWindowConf = &v2.ArcConf{IsSupport: false, Disabled: true}
		//定时提醒
		playConf.PlayArc.TimeUpConf = &v2.ArcConf{IsSupport: false, Disabled: true}
		//播放速度
		playConf.PlayArc.PlaybackRateConf = &v2.ArcConf{IsSupport: false, Disabled: true}
		//播放方式
		playConf.PlayArc.PlaybackModeConf = &v2.ArcConf{IsSupport: false, Disabled: true}
	}
	//试看
	if playConf.Ab != nil {
		playConf.Ab.Glance = nil
	}
}

// iphone粉6.53版本在hdr清晰度或杜比清晰度开启滤镜会黑屏
func isIOSShieldColorFilter(device, mobiApp string, build int64, acpQns []uint32) bool {
	var hasHdr bool
	for _, qn := range acpQns {
		if qn >= _qnHDR {
			hasHdr = true
			break
		}
	}
	return pd.WithDevice(
		pd.NewCommonDevice(mobiApp, device, "", build),
	).Where(func(pdContext *pd.PDContext) {
		pdContext.IsPlatIPhone().And().BuildIn(65300100)
	}).MustFinish() && hasHdr
}

func isSupportScreenRecording(arcInfo *arcmdl.Info) bool {
	//ugc稿件 && 非全景类稿件
	return !arcInfo.IsPGC() && !arcInfo.Is360()
}
func canOutputNoBackgroundReason(mobiApp, rawDevice string, build int64) bool {
	return pd.WithDevice(
		pd.NewCommonDevice(mobiApp, rawDevice, "", build),
	).Where(func(pdContext *pd.PDContext) {
		pdContext.IsPlatIPhone().And().Build(">", 64800000)
	}).OrWhere(func(pdContext *pd.PDContext) {
		pdContext.IsPlatAndroid().And().Build(">", 6480000)
	}).MustFinish()
}

func (s *Service) buildPlayViewAB(msg *model.GlanceMsg, qnList []uint32) *v2.AB {
	if msg == nil || !msg.CanGlance {
		return nil
	}
	var arcHasVipQn bool
	for _, qn := range qnList {
		if qn >= _qn1080H && qn != _qnDobly { //有非杜比的大会员清晰度
			arcHasVipQn = true
		}
	}
	if !arcHasVipQn {
		return nil
	}
	return &v2.AB{
		Glance: &v2.Glance{
			CanWatch: true,
			Duration: msg.FetchGlanceTime(s.c.GlanceConf.Duration, s.c.GlanceConf.Ratio),
			Times:    s.c.GlanceConf.Times,
		},
		Group: msg.Group,
	}
}

// 特定音乐相关稿件 isSpecificMusicConf .
func (s *Service) isSpecificMusicConf(mid, aid int64) bool {
	if _, ok := s.blackMusicMids[mid]; ok {
		return true
	}
	if _, ok := s.blackMusicAids[aid]; ok {
		return true
	}
	return false
}

// fromAbilityConf .
// nolint:gocognit
func (s *Service) fromAbilityConf(ctx context.Context, playConf map[int64]*taimdl.PlayConf, distributionReply *distribution.GetUserPreferenceReply) *v2.PlayAbilityConf {
	reply := &v2.PlayAbilityConf{}
	//全景
	reply.PanoramaConf = taimdl.DefaultFormatConf()
	if bVal, ok := playConf[int64(v2.ConfType_PANORAMA)]; ok && bVal != nil {
		reply.PanoramaConf = taimdl.FormatConf(bVal)
	}
	// 后台播放
	reply.BackgroundPlayConf = taimdl.DefaultFormatConf()
	if bVal, ok := playConf[int64(v2.ConfType_BACKGROUNDPLAY)]; ok && bVal != nil {
		reply.BackgroundPlayConf = taimdl.FormatConf(bVal)
	}
	// 镜像反转
	reply.FlipConf = taimdl.DefaultFormatConf()
	if bVal, ok := playConf[int64(v2.ConfType_FLIPCONF)]; ok && bVal != nil {
		reply.FlipConf = taimdl.FormatConf(bVal)
	}
	// 电视投屏
	reply.CastConf = taimdl.DefaultFormatConf()
	if bVal, ok := playConf[int64(v2.ConfType_CASTCONF)]; ok && bVal != nil {
		reply.CastConf = taimdl.FormatConf(bVal)
	}
	// 反馈
	reply.FeedbackConf = taimdl.DefaultFormatConf()
	if bVal, ok := playConf[int64(v2.ConfType_FEEDBACK)]; ok && bVal != nil {
		reply.FeedbackConf = taimdl.FormatConf(bVal)
	}
	// 字幕
	reply.SubtitleConf = taimdl.DefaultFormatConf()
	if bVal, ok := playConf[int64(v2.ConfType_SUBTITLE)]; ok && bVal != nil {
		reply.SubtitleConf = taimdl.FormatConf(bVal)
	}
	//播放速度
	reply.PlaybackRateConf = taimdl.DefaultFormatConf()
	if bVal, ok := playConf[int64(v2.ConfType_PLAYBACKRATE)]; ok && bVal != nil {
		reply.PlaybackRateConf = taimdl.FormatConf(bVal)
	}
	//定时停止播放
	reply.TimeUpConf = taimdl.DefaultFormatConf()
	if bVal, ok := playConf[int64(v2.ConfType_TIMEUP)]; ok && bVal != nil {
		reply.TimeUpConf = taimdl.FormatConf(bVal)
	}
	//播放方式
	reply.PlaybackModeConf = taimdl.DefaultFormatConf()
	if bVal, ok := playConf[int64(v2.ConfType_PLAYBACKMODE)]; ok && bVal != nil {
		reply.PlaybackModeConf = taimdl.FormatConf(bVal)
	}
	// 画面尺寸
	reply.ScaleModeConf = taimdl.DefaultFormatConf()
	if bVal, ok := playConf[int64(v2.ConfType_SCALEMODE)]; ok && bVal != nil {
		reply.ScaleModeConf = taimdl.FormatConf(bVal)
	}
	// 顶
	reply.LikeConf = taimdl.DefaultFormatConf()
	if bVal, ok := playConf[int64(v2.ConfType_LIKE)]; ok && bVal != nil {
		reply.LikeConf = taimdl.FormatConf(bVal)
	}
	// 踩
	reply.DislikeConf = taimdl.DefaultFormatConf()
	if bVal, ok := playConf[int64(v2.ConfType_DISLIKE)]; ok && bVal != nil {
		reply.DislikeConf = taimdl.FormatConf(bVal)
	}
	//投币
	reply.CoinConf = taimdl.DefaultFormatConf()
	if bVal, ok := playConf[int64(v2.ConfType_COIN)]; ok && bVal != nil {
		reply.CoinConf = taimdl.FormatConf(bVal)
	}
	//充电
	reply.ElecConf = taimdl.DefaultFormatConf()
	if bVal, ok := playConf[int64(v2.ConfType_ELEC)]; ok && bVal != nil {
		reply.ElecConf = taimdl.FormatConf(bVal)
	}
	//分享
	reply.ShareConf = taimdl.DefaultFormatConf()
	if bVal, ok := playConf[int64(v2.ConfType_SHARE)]; ok && bVal != nil {
		reply.ShareConf = taimdl.FormatConf(bVal)
	}
	//截图/gif
	reply.ScreenShotConf = taimdl.DefaultFormatConf()
	if bVal, ok := playConf[int64(v2.ConfType_SCREENSHOT)]; ok && bVal != nil {
		reply.ScreenShotConf = taimdl.FormatConf(bVal)
	}
	//锁屏
	reply.LockScreenConf = taimdl.DefaultFormatConf()
	if bVal, ok := playConf[int64(v2.ConfType_LOCKSCREEN)]; ok && bVal != nil {
		reply.LockScreenConf = taimdl.FormatConf(bVal)
	}
	//相关推荐
	reply.RecommendConf = taimdl.DefaultFormatConf()
	if bVal, ok := playConf[int64(v2.ConfType_RECOMMEND)]; ok && bVal != nil {
		reply.RecommendConf = taimdl.FormatConf(bVal)
	}
	//倍速
	reply.PlaybackSpeedConf = taimdl.DefaultFormatConf()
	if bVal, ok := playConf[int64(v2.ConfType_PLAYBACKSPEED)]; ok && bVal != nil {
		reply.PlaybackSpeedConf = taimdl.FormatConf(bVal)
	}
	//清晰度
	reply.DefinitionConf = taimdl.DefaultFormatConf()
	if bVal, ok := playConf[int64(v2.ConfType_DEFINITION)]; ok && bVal != nil {
		reply.DefinitionConf = taimdl.FormatConf(bVal)
	}
	//选集
	reply.SelectionsConf = taimdl.DefaultFormatConf()
	if bVal, ok := playConf[int64(v2.ConfType_SELECTIONS)]; ok && bVal != nil {
		reply.SelectionsConf = taimdl.FormatConf(bVal)
	}
	//下一集
	reply.NextConf = taimdl.DefaultFormatConf()
	if bVal, ok := playConf[int64(v2.ConfType_NEXT)]; ok && bVal != nil {
		reply.NextConf = taimdl.FormatConf(bVal)
	}
	//编辑弹幕
	reply.EditDmConf = taimdl.DefaultFormatConf()
	if bVal, ok := playConf[int64(v2.ConfType_EDITDM)]; ok && bVal != nil {
		reply.EditDmConf = taimdl.FormatConf(bVal)
	}
	//弹幕设置-三点内
	reply.InnerDmConf = taimdl.DefaultFormatConf()
	if bVal, ok := playConf[int64(v2.ConfType_INNERDM)]; ok && bVal != nil {
		reply.InnerDmConf = taimdl.FormatConf(bVal)
	}
	//弹幕设置-面板上
	reply.OuterDmConf = taimdl.DefaultFormatConf()
	if bVal, ok := playConf[int64(v2.ConfType_OUTERDM)]; ok && bVal != nil {
		reply.OuterDmConf = taimdl.FormatConf(bVal)
	}
	//小窗
	reply.SmallWindowConf = taimdl.DefaultFormatConf()
	if bVal, ok := playConf[int64(v2.ConfType_SMALLWINDOW)]; ok && bVal != nil {
		reply.SmallWindowConf = taimdl.FormatConf(bVal)
	}
	//震动事件
	reply.ShakeConf = taimdl.DefaultFormatConf()
	if bVal, ok := playConf[int64(v2.ConfType_SHAKE)]; ok && bVal != nil {
		reply.ShakeConf = taimdl.FormatConf(bVal)
	}
	//杜比
	reply.DolbyConf = taimdl.DefaultFormatConf()
	if bVal, ok := playConf[int64(v2.ConfType_DOLBY)]; ok && bVal != nil {
		reply.DolbyConf = taimdl.FormatConf(bVal)
	}
	//无损
	reply.LossLessConf = taimdl.DefaultFormatConf()
	if bVal, ok := playConf[int64(v2.ConfType_LOSSLESS)]; ok && bVal != nil {
		reply.LossLessConf = taimdl.FormatConf(bVal)
	}
	reply.ColorFilterConf = taimdl.DefaultFormatConf()
	if bVal, ok := playConf[int64(v2.ConfType_COLORFILTER)]; ok && bVal != nil {
		reply.ColorFilterConf = taimdl.FormatConf(bVal)
	}

	if distributionReply == nil {
		return reply
	}
	dpc, dcpc, err := translateDistributionReply(distributionReply)
	if err != nil {
		log.Error("%+v", err)
		return reply
	}
	reply.ColorFilterConf.ConfValue = abilityConfIntValSetter(ctx, &ColorFilterWithExp{EnumValue: EnumValue{Value: dpc.ColorFilter.GetValue()}})
	reply.SubtitleConf.ConfValue = abilityConfBoolValSetter(ctx, &SubtitleWithExp{BoolValue: BoolValue{Value: dpc.EnableSubtitle.GetValue()}})
	reply.DolbyConf.ConfValue = abilityConfBoolValSetter(ctx, &DolbyWithExp{BoolValue: BoolValue{Value: dcpc.EnableDolby.GetValue()}})
	reply.BackgroundPlayConf.ConfValue = abilityConfBoolValSetter(ctx, &BackgroundWithExp{BoolValue: BoolValue{Value: dcpc.EnableBackground.GetValue()}, LastModified: dcpc.EnableBackground.GetLastModified()})
	reply.PanoramaConf.ConfValue = abilityConfBoolValSetter(ctx, &PanoramaWithExp{BoolValue: BoolValue{Value: dcpc.EnablePanorama.GetValue()}})
	reply.ShakeConf.ConfValue = abilityConfBoolValSetter(ctx, &ShakeWithExp{BoolValue: BoolValue{Value: dcpc.EnableShake.GetValue()}})
	reply.LossLessConf.ConfValue = abilityConfBoolValSetter(ctx, &LossLessWithExp{BoolValue: BoolValue{Value: dcpc.EnableLossLess.GetValue()}})
	return reply
}

// PlayURLV2 .
func (s *Service) PlayURLV2(ctx context.Context, req *v2.PlayURLReq) (*v2.PlayURLReply, error) {
	params := &v2.PlayViewReq{
		Aid:          req.Aid,
		Cid:          req.Cid,
		Qn:           req.Qn,
		Platform:     req.Platform,
		Fnver:        req.Fnver,
		Fnval:        req.Fnval,
		Mid:          req.Mid,
		BackupNum:    req.BackupNum,
		Download:     req.Download,
		ForceHost:    req.ForceHost,
		Fourk:        req.Fourk,
		UpgradeAid:   req.UpgradeAid,
		UpgradeCid:   req.UpgradeCid,
		Device:       req.Device,
		MobiApp:      req.MobiApp,
		VerifySteins: req.VerifySteins,
		H5Hq:         req.H5Hq,
		Build:        req.Build,
		Buvid:        req.Buvid,
		VerifyVip:    req.VerifyVip,
		NetType:      req.NetType,
		TfType:       req.TfType,
		IsDazhongcar: req.IsDazhongcar,
	}
	var (
		eg      = errgroup.WithContext(ctx)
		playurl *v2.PlayURLReply
	)
	eg.Go(func(ctx context.Context) error {
		reply, _, _, _, err := s.PlayURLV3(ctx, params, false, false)
		if err != nil {
			return err
		}
		playurl = reply
		return nil
	})
	var vol *v2.VolumeInfo
	if req.VoiceBalance == 1 {
		eg.Go(func(ctx context.Context) error {
			reply, err := s.pudao.PlayurlVolume(ctx, uint64(req.Cid), uint64(req.Mid))
			if err != nil {
				log.Error("s.pudao.PlayurlVolume error(%+v) cid(%d) mid(%d)", err, req.Cid, req.Mid)
				return nil
			}
			vol = &v2.VolumeInfo{
				MeasuredI:         reply.MeasuredI,
				MeasuredLra:       reply.MeasuredLra,
				MeasuredTp:        reply.MeasuredTp,
				MeasuredThreshold: reply.MeasuredThreshold,
				TargetOffset:      reply.TargetOffset,
				TargetI:           reply.TargetI,
				TargetTp:          reply.TargetTp,
			}
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	playurl.Volume = vol
	return playurl, nil
}

// PlayURLV2 .
func (s *Service) PlayURLV3(ctx context.Context, req *v2.PlayViewReq, isView, disableDolby bool) (reply *v2.PlayURLReply, vRly *verifyReply, msg *model.GlanceMsg, vipConf *v2.VipConf, err error) {
	var (
		isPreview, isSp, isFreeSp bool
		response                  *v2.ResponseMsg
		code                      int
	)
	reply = new(v2.PlayURLReply)
	_, posterOK := s.pasterCache[req.Cid] //贴片视频白名单
	_, steinsOK := s.steinsWhite[req.Cid] //互动引导视频白名单
	params := &vod.RequestMsg{
		Cid:          uint64(req.Cid),
		Qn:           uint32(req.Qn),
		Uip:          metadata.String(ctx, metadata.RemoteIP),
		Platform:     req.Platform,
		Fnver:        uint32(req.Fnver),
		Fnval:        uint32(req.Fnval),
		Mid:          uint64(req.Mid),
		BackupNum:    req.BackupNum,
		Download:     req.Download,
		ForceHost:    uint32(req.ForceHost),
		Fourk:        req.Fourk,
		FlvProj:      s.getFlvProject(req.Buvid, req.Device, 0),
		ReqSource:    vod.RequestSource(req.BusinessSource),
		IsDazhongcar: req.IsDazhongcar,
	}
	if disableDolby { //fnval=256标志支持杜比
		params.Fnval = params.Fnval &^ (1 << v1.FnvalNeedDolby)
	}
	// 增加mid灰度
	if req.Mid%100 < s.c.Custom.TFGray {
		params.NetType = vod.NetworkType(req.NetType)
		params.TfType = vod.TFType(req.TfType)
	}
	if !posterOK && !steinsOK {
		vfyArg := &verifyArcArg{
			Aid:          req.Aid,
			Cid:          req.Cid,
			Mid:          req.Mid,
			Qn:           req.Qn,
			Platform:     req.Platform,
			Device:       req.Device,
			MobiApp:      req.MobiApp,
			VerifySteins: req.VerifySteins,
			Build:        req.Build,
			Buvid:        req.Buvid,
			VerifyVip:    req.VerifyVip,
		}
		if isPreview, isSp, _, vRly, err = s.verifyArchive(ctx, vfyArg); err != nil {
			return
		}
		// set changed params
		params.Cid = uint64(vfyArg.Cid)
		params.Qn = uint32(vfyArg.Qn)
		params.Preview = isPreview
		params.IsSp = isSp
		// vip限免视频默认 可看不可下载
		limitFree, subTitle := s.isVipConf(req.Aid)
		//首p视频才下发副标题
		if !s.isFirstCid(vfyArg.Cid, vRly.Arc) {
			subTitle = ""
		}
		vipConf = &v2.VipConf{LimitFree: limitFree, Subtitle: subTitle}
		isFreeSp = limitFree == 1
		if isFreeSp && req.Download == 0 {
			params.IsSp = true
		}
		//非大会员试看大会员清晰度逻辑
		msg = &model.GlanceMsg{Mid: vfyArg.Mid, Duration: vRly.Arc.Duration, IsSp: isSp}
		func() {
			if !msg.SupportGlance() || !feature.GetBuildLimit(ctx, "service.glance", nil) {
				return
			}
			//如果647之后版本命中人群包则没有试看逻辑
			if s.hitCrowed(ctx, msg.Mid, req.Buvid, req.MobiApp, req.Device, req.Build) {
				return
			}
			msg.CanGlance = true
			msg.Group = v2.Group_B
			params.IsSp = true
		}()
	}
	if req.MobiApp != "" { //客户端用mobi_app,pc h5之类用platform
		params.Platform = req.MobiApp
		// 将ipad粉的platform设为ipad，并增加灰度控制
		if pd.WithContext(ctx).Where(func(pd *pd.PDContext) {
			pd.IsPlatIPad().And().Build(">=", 66000100)
		}).FinishOr(false) && s.ClarityGrayControl(req.Mid, req.Buvid) {
			params.Platform = "ipad"
		}
	}
	s.setStoryQn(params, req.Buvid)
	response, code, err = s.pudao.PlayurlV2(ctx, params, req.H5Hq, isView, isSp, isFreeSp)
	if err != nil {
		log.Error("PlayURLV2 s.pudao.PlayurlV2(%+v) error(%v)", params, err)
		return
	}
	if code != ecode.OK.Code() {
		log.Error("PlayURLV2 aid(%d) cid(%d) code(%d)", req.Aid, req.Cid, code)
		err = ecode.NothingFound
		reply = nil
		return
	}
	reply.Playurl = response
	return
}

func (s *Service) hitCrowed(ctx context.Context, mid int64, buvid string, mobiApp, rawDevice string, build int32) bool {
	//低版本已全量试看，不需要人群包逻辑
	if pd.WithDevice(
		pd.NewCommonDevice(mobiApp, rawDevice, "", int64(build)),
	).Where(func(pdContext *pd.PDContext) {
		pdContext.IsPlatIPhone().And().Build("<", 64700000)
	}).OrWhere(func(pdContext *pd.PDContext) {
		pdContext.IsPlatAndroid().And().Build("<", 6470000)
	}).MustFinish() {
		return false
	}
	//接入人群包
	vipProfileReq := &vipProfilerpc.VerifyCrowdCondsReq{
		Mid:   mid,
		Buvid: buvid,
		Conds: []*vipProfilerpc.ModelCond{
			{
				Cond: "crowdpack=='cp_crm_play_clarity_ctr_model_v1'",
			},
		}}
	vpReply, err := s.vipProfileClient.VerifyCrowdConds(ctx, vipProfileReq)
	//默认值为命中人群包，不给试看逻辑
	if err != nil {
		log.Error("s.vipProfileClient.VerifyCrowdConds error(%+v), mid(%d)", err, mid)
		return true
	}
	if len(vpReply.Conds) == 0 || vpReply.Conds[0] == nil {
		log.Error("vpReply nothing found mid(%d)", mid)
		return true
	}
	return vpReply.Conds[0].IsHit
}

// nolint:gomnd
func (s *Service) setStoryQn(param *vod.RequestMsg, buvid string) {
	if param.ReqSource != vod.RequestSource_STORY {
		return
	}
	// mid白名单
	for _, gmid := range s.c.Custom.StoryQnGroup1Mids {
		if param.Mid == uint64(gmid) {
			param.Qn = v1.Qn1080
			return
		}
	}
	for _, gmid := range s.c.Custom.StoryQnGroup2Mids {
		if param.Mid == uint64(gmid) {
			if param.NetType == vod.NetworkType_WIFI {
				param.Qn = v1.Qn1080
			} else {
				param.Qn = v1.QnFlv720
			}
			return
		}
	}
	storyMod := crc32.ChecksumIEEE([]byte(buvid)) % 100
	if storyMod < s.c.Custom.StoryQnGroup1 {
		// 实验组1：wifi和非wifi下均1080p
		param.Qn = v1.Qn1080
		return
	}
	if storyMod >= s.c.Custom.StoryQnGroup1 && storyMod < s.c.Custom.StoryQnGroup2 {
		// 实验组2：wifi下1080p，非wifi下720p
		if param.NetType == vod.NetworkType_WIFI {
			param.Qn = v1.Qn1080
		} else {
			param.Qn = v1.QnFlv720
		}
		return
	}
	param.Qn = v1.QnFlv720
}

// SteinsPreview is interactive archive preview for up
func (s *Service) SteinsPreview(ctx context.Context, req *v1.SteinsPreviewReq) (reply *v1.PlayURLInfo, err error) {
	if _, ok := s.authorisedCallers[metadata.String(ctx, metadata.Caller)]; !ok {
		err = ecode.AccessDenied
		log.Warn("SteinsPreview Not Authorised Caller %s", metadata.String(ctx, metadata.Caller))
		return
	}
	var (
		code     int
		response *v2.ResponseMsg
	)
	reply = new(v1.PlayURLInfo)
	params := &vod.RequestMsg{
		Cid:       uint64(req.Cid),
		Qn:        uint32(req.Qn),
		Uip:       metadata.String(ctx, metadata.RemoteIP),
		Platform:  req.Platform,
		Fnver:     uint32(req.Fnver),
		Fnval:     uint32(req.Fnval),
		Mid:       uint64(req.Mid),
		ForceHost: uint32(req.ForceHost),
		IsSp:      true,
		NetType:   vod.NetworkType(req.NetType),
		TfType:    vod.TFType(req.TfType),
	}
	response, code, err = s.pudao.PlayurlV2(ctx, params, false, false, false, false)
	if err != nil {
		log.Error("SteinsPreview s.pudao.PlayurlV2(%+v) error(%v)", params, err)
		return
	}
	if code != ecode.OK.Code() {
		log.Error("SteinsPreview aid(%d) cid(%d) code(%d)", req.Aid, req.Cid, code)
		err = ecode.NothingFound
		reply = nil
		return
	}
	reply.FromPlayurlV2(response)
	return
}

// nolint:gocognit
func (s *Service) verifyArchive(ctx context.Context, arg *verifyArcArg) (isPreview, isSp, isVip bool, vRly *verifyReply, err error) {
	var (
		arc       *arcmdl.Info
		allowPlay bool
	)
	if arc, err = s.arcDao.GetSimpleArc(ctx, arg.Aid, arg.Mid, arg.MobiApp, arg.Device, arg.Platform); err != nil {
		log.Error("verifyArchive SimpleArcService arg(%+v) err(%+v)", arg, err)
		return
	}
	vRly = &verifyReply{Arc: arc}
	if arc.State == arcApi.StateForbidUserDelay { //非首映稿件定时发布状态下允许up主观看 || 首映稿件返回播放地址
		if arc.AttrValV2(arcApi.AttrBitV2Premiere) != arcApi.AttrYes {
			if arg.Mid != arc.Mid {
				err = ecode.NothingFound
				log.Warn("verifyArchive aid(%d) state=-40 mid(%d) up(%d)", arg.Aid, arg.Mid, arc.Mid)
				return
			}
		}
	} else if !arc.IsNormal() {
		err = ecode.NothingFound
		log.Warn("verifyArchive aid(%d) can not play arc(%+v)", arg.Aid, arc)
		return
	}
	if !arc.HasCid(arg.Cid) {
		err = ecode.NothingFound
		log.Warn("verifyArchive aid(%d) has no cid(%d) arc(%+v)", arg.Aid, arg.Cid, arc)
		return
	}
	if arc.AttrVal(arcApi.AttrBitIsPUGVPay) == arcApi.AttrYes {
		err = ecode.NothingFound
		log.Warn("verifyArchive aid(%d) is PUGV", arg.Aid)
		return
	}
	//付费稿件，用户未付费或者不是免费试看，则不下发播放地址
	if arc.AttrValV2(arcApi.AttrBitV2Pay) == arcApi.AttrYes {
		//不支持提示升级，则取默认视频
		if arg.UpgradeAid != 0 {
			arg.Aid = arg.UpgradeAid
			arg.Cid = arg.UpgradeCid
			log.Warn("verifyArchive aid(%d) is 付费稿件低版本提示升级 arc(%+v) arg(%+v)", arg.Aid, arc, arg)
			s.promInfo.Incr("付费稿件:旧版本默认视频")
		} else {
			paid, isArcPlay, inSeasonFreeWatch := checkArcPayPlay(arc)
			//非免费观看视频需要判断版本和付费情况
			if !inSeasonFreeWatch {
				if !checkArcPayVersion(arg.MobiApp, arg.Build, arg.Platform, arg.Device) {
					//支持提示升级的版本，直接返回
					err = xecode.PlayURLArcPayUpgrade
					log.Warn("verifyArchive aid(%d) is 付费稿件低版本提示升级 arc(%+v) arg(%+v)", arg.Aid, arc, arg)
					s.promInfo.Incr("付费稿件:低版本提示升级")
					return
				}
				if isArcPlay && !paid {
					err = xecode.PlayURLArcPayNotice
					log.Warn("verifyArchive aid(%d) is 付费稿件未付费不可观看 arc(%+v) arg(%+v)", arg.Aid, arc, arg)
					s.promInfo.Incr("付费稿件:未付费不可观看")
					return
				}
			}
			s.promInfo.Incr("付费稿件:正常观看")
		}
	}
	// 自见稿件
	if arc.AttrValV2(arcApi.AttrBitV2OnlySely) == arcApi.AttrYes {
		if arg.Mid != arc.Mid {
			err = ecode.NothingFound
			return
		}
	}
	if arc.AttrValV2(arcApi.AttrBitV2OnlyFavView) == arcApi.AttrYes {
		ok, err := s.validateForOnlyFav(ctx, arg.Mid, arc)
		if err != nil {
			log.Error("verifyArchive validateForOnlyFav mid(%d) aid(%d) error(%+v)", arg.Mid, arg.Aid, err)
			return false, false, false, nil, err
		}
		if !ok {
			return false, false, false, nil, ecode.NothingFound
		}
	}
	// platform html only need simple check
	if arg.Platform == _platformHtml5 || arg.Platform == _platformHtml5New {
		if arc.IsPGC() || arc.AttrVal(arcApi.AttrBitUGCPay) == arcApi.AttrYes {
			log.Warn("verifyArchive html5 aid(%d) not allowed", arg.Aid)
			err = ecode.NothingFound
		}
		return
	}
	// 投屏接口不支持互动视频 不支持所有pgc内容
	if arg.Source == _sourceProject && (arc.IsSteinsGate() || arc.IsPGC()) {
		log.Warn("verifyArchive aid(%d) is steins-gate or pgc can not project", arc.Aid)
		err = ecode.NothingFound
		return
	}
	if arg.VerifySteins == 1 && arc.IsSteinsGate() {
		if allowPlay, err = s.arcDao.SteinsGraphRights(ctx, arg.MobiApp, arg.Device, arg.Build, arg.Aid); err != nil {
			log.Error("verifyArchive aid(%d) can not play because SteinsGate !! Err %v", arg.Aid, err)
			err = ecode.NothingFound
			return
		}
		if !allowPlay { // 版本号不符合要求时报错
			err = xecode.PlayURLSteinsUpgrade
			log.Warn("verifyArchive aid(%d) can not play because SteinsGate !!", arg.Aid)
			return
		}
	}
	// TODO 历史老坑 纪录片之类的会请求UGC的playurl，先打日志记一记
	if arc.IsPGC() {
		log.Warn("verifyArchive aid(%d) pay(%d) cid(%d) is pgc!!!!", arg.Aid, arc.AttrVal(arcApi.AttrBitBadgepay), arg.Cid)
		var PGCCanPlay bool
		PGCCanPlay, err = s.pgcDao.PGCCanPlay(ctx, arg.Mid, arg.Cid, arg.Platform, arg.Device, arg.MobiApp)
		if err != nil || !PGCCanPlay {
			log.Error("pgc can not play arg(%+v) or err(%+v)", arg, err)
			err = ecode.NothingFound
			return
		}
	}
	if arc.AttrVal(arcApi.AttrBitUGCPay) == 1 {
		var relation *ugcpaymdl.AssetRelationResp
		if arg.Mid > 0 && arc.Mid != arg.Mid {
			if relation, err = s.ugcpayDao.AssetRelation(ctx, arg.Aid, arg.Mid); err != nil {
				log.Error("verifyArchive AssetRelation err(%+v)", err)
				return
			}
		}
		// 老版本出引导升级付费视频，新版本有预览播预览，没有提示未付费错误
		if arg.Mid <= 0 || (relation != nil && relation.State != _relationPaid) {
			if relation != nil {
				log.Warn("verifyArchive not pay aid(%d) mid(%d) state(%s)", arg.Aid, arg.Mid, relation.State)
			}
			if arg.UpgradeAid == 0 {
				isPreview = true
				if arc.AttrVal(arcApi.AttrBitUGCPayPreview) == arcApi.AttrNo || arg.Cid != arc.Cids[0] {
					err = xecode.PlayURLNotPay
					return
				}
			} else {
				arg.Aid = arg.UpgradeAid
				arg.Cid = arg.UpgradeCid
			}
		}
	}
	if arg.Mid > 0 {
		var vipCtl *vipInforpc.ControlResult
		isVip, vipCtl = s.vipDao.Info(ctx, arg.Mid, arg.Buvid, arg.VerifyVip == 1)
		if arg.Mid == arc.Mid {
			isSp = true
		} else {
			isSp = isVip
			vRly.VipControl = vipCtl
		}
	} else {
		s.setNologinQn(arg)
	}
	return
}

func (s *Service) setNologinQn(arg *verifyArcArg) {
	//车载不限制未登录清晰度
	if arg.MobiApp == "android_bilithings" {
		return
	}
	if arg.Platform == "pc" { //pc端未登录实验，由pc控制请求qn，最高清晰度调整为720p
		if arg.Qn > _qn720 {
			arg.Qn = _qn720
		}
	} else if arg.Qn > _qn480 { //其他平台未登录限制当前最高清晰度 480p
		arg.Qn = _qn480
	}
}

// Ping Service
func (s *Service) Ping(c context.Context) (err error) {
	return
}

// Close Service
func (s *Service) Close() {
	s.cron.Stop()
	s.arcDao.Close()
	if s.tinker != nil {
		s.tinker.Close()
	}
}

// hls 公共参数获取逻辑 .
func (s *Service) hls(ctx context.Context, req *v2.HlsCommonReq) (*hlsgrpc.M3U8RequestMsg, bool, error) {
	var (
		isPreview, isSp, isVip bool
		err                    error
		platform               string
	)
	vfyArg := &verifyArcArg{
		Aid:      req.Aid,
		Cid:      req.Cid,
		Mid:      req.Mid,
		Qn:       req.Qn,
		Platform: req.Platform,
		Device:   req.Device,
		MobiApp:  req.MobiApp,
		Buvid:    req.Buvid,
		Build:    req.Build,
	}
	switch req.RequestType {
	case v2.RequestType_AIRPLAY: //投屏
		vfyArg.Source = _sourceProject
		if isPreview, isSp, isVip, _, err = s.verifyArchive(ctx, vfyArg); err != nil {
			return nil, false, err
		}
		if !s.verifyProject(ctx, req.Aid, req.Cid, req.DeviceType) {
			return nil, false, xecode.ProjectInvalidOtt
		}
		platform = v1.ProjectPlatform
	case v2.RequestType_PIP: //画中画
		vfyArg.VerifyVip = req.VerifyVip //vip管控逻辑，只针对粉板播放，不需要处理投屏
		if isPreview, isSp, isVip, _, err = s.verifyArchive(ctx, vfyArg); err != nil {
			return nil, false, err
		}
		// vip限免视频默认 可看不可下载
		if s.isVipFree(req.Aid) {
			isSp = true
		}
		// 视频云的platform 对应客户端的mobilapp,暂时只有app接入，web接入需特殊处理
		platform = req.MobiApp
		if req.Platform == _platformHtml5 || req.Platform == _platformHtml5New {
			platform = req.Platform
		}
	default:
		return nil, false, ecode.RequestErr
	}
	rly := &hlsgrpc.M3U8RequestMsg{
		Cid:       uint64(vfyArg.Cid),
		Qn:        uint32(vfyArg.Qn),
		Uip:       metadata.String(ctx, metadata.RemoteIP),
		Platform:  platform,
		Fnver:     uint32(req.Fnver),
		Fnval:     uint32(req.Fnval),
		Mid:       uint64(req.Mid),
		BackupNum: req.BackupNum,
		ForceHost: uint32(req.ForceHost),
		Preview:   isPreview,
		IsSp:      isSp,
		NetType:   hlsgrpc.NetworkType(req.NetType),
		TfType:    hlsgrpc.TFType(req.TfType),
		Type:      hlsgrpc.RequstType(req.RequestType),
		Business:  hlsgrpc.Business(req.Business),
	}
	return rly, isVip, nil
}

// HlsScheduler is
func (s *Service) HlsScheduler(ctx context.Context, req *v2.HlsCommonReq) (*v2.HlsSchedulerReply, error) {
	param, _, err := s.hls(ctx, req)
	if err != nil {
		return nil, err
	}
	rly, err := s.pudao.HlsScheduler(ctx, param)
	if err != nil {
		return nil, err
	}
	return &v2.HlsSchedulerReply{Playurl: rly}, nil
}

// MasterScheduler is
func (s *Service) MasterScheduler(ctx context.Context, req *v2.HlsCommonReq) (*v2.MasterSchedulerReply, error) {
	param, isVip, err := s.hls(ctx, req)
	if err != nil {
		return nil, err
	}
	dc := &model.DolbyConf{
		IsVip:         isVip,
		Dolby:         req.Dolby,
		TeenagersMode: req.TeenagersMode,
		LessonsMode:   req.LessonsMode,
		MobiApp:       req.MobiApp,
		Device:        req.Device,
	}
	rly, err := s.pudao.MasterScheduler(ctx, param, dc)
	if err != nil {
		return nil, err
	}
	return &v2.MasterSchedulerReply{Info: rly}, nil
}

func (s *Service) M3U8Scheduler(ctx context.Context, req *v2.HlsCommonReq) (*v2.M3U8SchedulerReply, error) {
	param, _, err := s.hls(ctx, req)
	if err != nil {
		return nil, err
	}
	if req.QnCategory == v2.QnCategory_Audio { //为音频qn,音频qn不参与降路逻辑
		param.Qn = uint32(req.Qn)
	}
	rly, err := s.pudao.M3U8Scheduler(ctx, param)
	if err != nil {
		return nil, err
	}
	return &v2.M3U8SchedulerReply{Info: rly}, nil
}

// Project is
func (s *Service) Project(ctx context.Context, req *v2.ProjectReq) (reply *v2.ProjectReply, err error) {
	var (
		isPreview, isSp bool
		response        *v2.ResponseMsg
		code            int
	)
	reply = new(v2.ProjectReply)
	params := &tvproj.RequestMsg{
		Cid:       uint64(req.Cid),
		Qn:        uint32(req.Qn),
		Uip:       metadata.String(ctx, metadata.RemoteIP),
		Platform:  req.Platform,
		Fnver:     uint32(req.Fnver),
		Fnval:     uint32(req.Fnval),
		Mid:       uint64(req.Mid),
		BackupNum: req.BackupNum,
		Download:  req.Download,
		ForceHost: uint32(req.ForceHost),
		Fourk:     req.Fourk,
		Business:  changeProjectBus(req.Business),
		FlvProj:   s.getFlvProject(req.Buvid, req.Device, req.Protocol),
	}
	vfyArg := &verifyArcArg{
		Aid:      req.Aid,
		Cid:      req.Cid,
		Mid:      req.Mid,
		Qn:       req.Qn,
		Platform: req.Platform,
		Device:   req.Device,
		MobiApp:  req.MobiApp,
		Source:   _sourceProject,
	}
	if isPreview, isSp, _, _, err = s.verifyArchive(ctx, vfyArg); err != nil {
		return
	}
	// set changed params
	params.Cid = uint64(vfyArg.Cid)
	params.Qn = uint32(vfyArg.Qn)
	params.Preview = isPreview
	params.IsSp = isSp
	params.Platform = v1.ProjectPlatform
	if !s.verifyProject(ctx, req.Aid, req.Cid, req.DeviceType) {
		err = xecode.ProjectInvalidOtt
		return
	}
	response, code, err = s.pudao.Project(ctx, params)
	if err != nil {
		log.Error("PlayURLV2 s.pudao.Project(%+v) error(%v)", params, err)
		return
	}
	if code != ecode.OK.Code() {
		log.Error("PlayURLV2 aid(%d) cid(%d) code(%d) arg(%+v)", req.Aid, req.Cid, code, params)
		err = ecode.NothingFound
		reply = nil
		return
	}
	reply.Playurl = response
	return
}

func (s *Service) verifyProject(c context.Context, aid, cid int64, deviceTp int32) bool {
	if s.c.Custom.OTTPlayVerify == 1 && deviceTp == 1 { //OTT设备才校验是否OTT过审
		ottCanPlay, err := s.ottDao.VideoAuthUgc(c, aid, cid)
		log.Warn("verifyProject aid(%d) cid(%d) deviceTp(%d) canplay(%t) err(%v)", aid, cid, deviceTp, ottCanPlay, err)
		return ottCanPlay
	}
	return true
}

func changeProjectBus(business v2.Business) tvproj.Business {
	switch business {
	case v2.Business_UGC:
		return tvproj.Business_UGC
	case v2.Business_PGC:
		return tvproj.Business_PGC
	case v2.Business_PUGV:
		return tvproj.Business_PUGV
	default:
		return tvproj.Business_UGC
	}
}

func (s *Service) getFlvProject(buvid, device string, protocol int32) bool {
	//由于ipad画中画使用了投屏地址且不支持flv，所以从实验中过滤
	if device != "pad" && protocol != v2.ProtocolAirPlay && crc32.ChecksumIEEE([]byte(buvid+"_project_flv"))%100 < s.c.Custom.FlvProjectGray {
		return true
	}
	return false
}

// initElec --与view接口逻辑保持一致
func (s *Service) initElec(c context.Context, arc *arcmdl.Info, req *v2.PlayViewReq, normalPlay bool) bool {
	// 青少年和国际版不支持充电
	if !normalPlay || req.TeenagersMode != 0 || req.LessonsMode != 0 || req.MobiApp == "iphone_i" || req.MobiApp == "android_i" {
		return false
	}
	if _, ok := s.allowTypeIds[arc.TypeID]; !ok || int8(arc.Copyright) != arcApi.CopyrightOriginal {
		return false
	}
	show, e := s.ugcpayRankDao.ArchiveElecStatus(c, arc.Mid, arc.Aid)
	if e != nil {
		return false
	}
	return show
}

// infoc
func (s *Service) infocSave(i interface{}) {
	switch v := i.(type) {
	case arcmdl.CloudInfo:
		payload := infoc.NewLogStream(s.c.InfocConf.CloudLogID, v.Ctime, v.Buvid, v.Platform, v.FMode, v.Ver, v.Function, v.Brand, v.Model, v.EditSouce, v.FpLocal)
		_ = s.cloudInfoc.Info(context.Background(), payload)
	case model.DolbyInfo:
		payload := infoc.NewLogStream(s.c.InfocConf.DolbyLogID, v.Buvid, v.Mid, v.Ctime, v.MobiApp, v.Platform, v.Build, v.Aid, v.Cid, _dolbyinfoType, _dolbyScene, "", "", "", "", v.DolbyType)
		_ = s.cloudInfoc.Info(context.Background(), payload)
	case model.LitePlayerInfoc:
		payload := infoc.NewLogStream(s.c.InfocConf.LiteLogID, 0, v.Buvid, v.GroupID, v.JoinTime)
		_ = s.cloudInfoc.Info(context.Background(), payload)
	default:
		log.Warn("infocproc can't process the type")
	}
}

// 是否有大会员限免&&副标题
func (s *Service) isVipConf(aid int64) (int32, string) {
	if val, ok := s.vipFreeAids[aid]; ok && val != nil {
		return val.LimitFree, val.Subtitle
	}
	return 0, ""
}

func (s *Service) isVipFree(aid int64) bool {
	if val, ok := s.vipFreeAids[aid]; ok && val != nil && val.LimitFree == 1 {
		return true
	}
	return false
}

func (s *Service) isFirstCid(cid int64, arcs *arcmdl.Info) bool {
	if arcs == nil {
		return false
	}
	for i, v := range arcs.Cids {
		if cid == v && i == 0 {
			return true
		}
	}
	return false
}

func (s *Service) PlayOnlineGRPC(c context.Context, arg *v2.PlayOnlineReq) (*v2.PlayOnlineReply, error) {
	res := new(v2.PlayOnlineReply)
	if s.HitBlackList(arg.Aid) {
		res.IsHide = true
		return res, nil
	}
	// 获取真实的在线人数
	business := "ugc"
	if arg.Business == v2.OnlineBusiness_OnlineOGV {
		business = "ogv"
	}
	onlineRes, err := s.broadcast.Online(c, &bcgrpc.OnlineReq{
		Business: business,
		Aid:      arg.Aid,
		Cid:      arg.Cid,
	})
	if err != nil {
		log.Error("s.bcClient.Online error(%+v)", err)
		return nil, err
	}
	onlineCount := splitOnlineRes(onlineRes)
	res.Count = map[string]int64{
		model.Total: onlineCount.Total,
		model.Web:   onlineCount.Web,
	}
	// 获取受管控时候的在线人数信息
	AppPeakMap, peakTime, hitOnlineControl := s.cacheDao.FetchOnlineInfo(c, arg.Aid)
	if !hitOnlineControl {
		return res, nil
	}
	totalFd := s.FetchDisplayedCount(c, arg.Aid, arg.Cid, AppPeakMap[arg.Cid], peakTime, onlineCount.App, res.Count[model.Web])
	if totalFd != 0 {
		res.Count[model.Total] = totalFd
	}
	return res, nil
}

func (s *Service) HitBlackList(aid int64) bool {
	if _, ok := s.onlineBlackList[aid]; ok {
		return true
	}
	return false
}

func splitOnlineRes(onlineRes *bcgrpc.OnlineReply) *model.OnlineCount {
	out := &model.OnlineCount{
		Total: onlineRes.Total,
	}
	for room, count := range onlineRes.GetRooms() {
		if strings.HasPrefix(room, model.Video) {
			out.Web = count
			continue
		}
		out.App += count
	}
	return out
}

func (s *Service) FetchDisplayedCount(ctx context.Context, aid, cid, peak, peakTime, appReal, webReal int64) int64 {
	//展示平滑在线人数,web端为真实值，其他端平滑
	var (
		now     = time.Now().Unix()
		totalFd int64
	)
	appFd := fakeDecline(peak, peakTime, appReal, now)
	log.Info("hit fakeDecline aid(%d),cid(%d),real(%d),fake(%d)", aid, cid, appReal, appFd)
	s.promInfo.State(fmt.Sprintf("onlineSmooth_real_%d_%d", aid, cid), appReal)
	s.promInfo.State(fmt.Sprintf("onlineSmooth_fake_%d_%d", aid, cid), appFd)
	if appFd != 0 {
		totalFd = appFd + webReal
	}
	return totalFd
}

func fakeDecline(peak int64, peakTime int64, realCount int64, now int64) int64 {
	if peak == 0 || peakTime == 0 {
		return 0
	}
	t := now - peakTime
	return peak - (peak-realCount)*t/_24h
}

func canOutputPlayConf(ctx context.Context) bool {
	return pd.WithContext(ctx).Where(func(pdContext *pd.PDContext) {
		pdContext.IsPlatAndroid().And().Build(">=", 6540000)
	}).OrWhere(func(pdContext *pd.PDContext) {
		pdContext.IsPlatIPhone().And().Build(">=", 66200000)
	}).MustFinish() || padCanOutputPlayConf(ctx)
}

func padCanOutputPlayConf(ctx context.Context) bool {
	return pd.WithContext(ctx).Where(func(pdContext *pd.PDContext) {
		pdContext.IsPlatIPad().And().Build(">=", 67500000)
	}).OrWhere(func(pdContext *pd.PDContext) {
		pdContext.IsPlatIPadHD().And().Build(">=", 34300000)
	}).MustFinish()
}

// ClarityGrayControl 白名单和灰度控制
func (s *Service) ClarityGrayControl(mid int64, buvid string) bool {
	// 白名单
	_, ok := s.c.IpadClarityGrayControl.Mid[strconv.FormatInt(mid, 10)]
	// 灰度控制
	group := crc32.ChecksumIEEE([]byte(buvid)) % MaxGray
	return ok || group < uint32(s.c.IpadClarityGrayControl.Gray)
}

// checkArcPayPlay 检查付费稿件是否可以播放，非付费稿件paid=true
func checkArcPayPlay(info *arcmdl.Info) (paid bool, isArcPay bool, inSeasonFreeWatch bool) {
	//不是付费稿件，默认可以播放
	if info.AttrValV2(arcApi.AttrBitV2Pay) == arcApi.AttrNo {
		return false, false, false
	}
	isArcPay = true
	//付费稿件，如果是合集付费，已付费或者免费观看，可播放
	if info.Pay != nil && info.Pay.AttrVal(arcApi.PaySubTypeAttrBitSeason) == arcApi.AttrYes {
		for _, gs := range info.Pay.GoodsInfo {
			if gs.Category == arcApi.Category_CategorySeason {
				inSeasonFreeWatch = gs.FreeWatch
				if gs.PayState == arcApi.PayState_PayStateActive {
					paid = true
				}
				break
			}
		}
	}
	//其余情况默认不可播放
	return paid, isArcPay, inSeasonFreeWatch
}

func checkArcPayVersion(mobiApp string, build int32, platform string, device string) bool {
	if platform == "pc" || platform == "html5" || platform == "html5_new" {
		return true
	}
	return (mobiApp == "android" && build >= _payArcVersionControlAndroid) ||
		(mobiApp == "iphone" && build >= _payArcVersionControlIos) ||
		(mobiApp == "iphone" && device == "phone" && build >= _payArcVersionControlPad) ||
		(mobiApp == "ipad" && build >= _payArcVersionControlIpadHd) ||
		(mobiApp == "android_hd" && build >= _payArcVersionControlAndroidHd) ||
		(mobiApp == "android_i" && build >= _payArcVersionControlAndroid) ||
		(mobiApp == "iphone_i" && build >= _payArcVersionControlIos)
}

func (s *Service) buggyAndroidDegreeForPayArc(c context.Context, req *v2.PlayViewReq) *v2.ResponseMsg {
	versionMatch := req.MobiApp == "android" && req.Build < _payArcVersionControlAndroid
	if !versionMatch {
		return nil
	}
	s.promInfo.Incr("付费稿件:android低版本获取降级视频")
	req.Aid = s.c.Custom.PayArcDegreeAid
	req.Cid = s.c.Custom.PayArcDegreeCid

	playRly, _, _, _, err := s.PlayURLV3(c, req, true, true)
	if err != nil {
		log.Error("buggyAndroidDegreeForPayArc PlayView s.PlayURLV3(%d,%d) error(%v)", req.Aid, req.Cid, err)
		return nil
	}
	if playRly != nil {
		return playRly.Playurl
	}
	return nil
}
