package splash

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/log/infoc.v2"
	"go-common/library/net/metadata"
	"go-common/library/stat/prom"
	"go-common/library/sync/errgroup"
	xtime "go-common/library/time"
	"go-gateway/app/app-svr/app-feed/interface-ng/card-schema/util/sets"
	feedcommon "go-gateway/app/app-svr/app-feed/interface/common"
	"go-gateway/app/app-svr/app-resource/interface/conf"
	addao "go-gateway/app/app-svr/app-resource/interface/dao/ad"
	locdao "go-gateway/app/app-svr/app-resource/interface/dao/location"
	magerdao "go-gateway/app/app-svr/app-resource/interface/dao/manager"
	spdao "go-gateway/app/app-svr/app-resource/interface/dao/splash"
	"go-gateway/app/app-svr/app-resource/interface/model"
	"go-gateway/app/app-svr/app-resource/interface/model/manager"
	"go-gateway/app/app-svr/app-resource/interface/model/splash"

	advo "git.bilibili.co/bapis/bapis-go/bcg/sunspot/ad/vo"
	"github.com/pkg/errors"
	"github.com/robfig/cron"
)

// Service is splash service.
type Service struct {
	c                    *conf.Config
	dao                  *spdao.Dao
	ad                   *addao.Dao
	loc                  *locdao.Dao
	mager                *magerdao.Dao
	brandCache           *manager.SplashList
	collectionBrandCache *manager.CollectionSplashList
	// whitelist
	whiteMidCache   map[int64]int
	whiteBuvidCache map[string]int
	// cron
	cron *cron.Cron
	// infoc
	inf2 infoc.Infoc
	prom *prom.Prom
}

// New new a splash service.
func New(c *conf.Config, ic infoc.Infoc) *Service {
	s := &Service{
		c:     c,
		dao:   spdao.New(c),
		ad:    addao.New(c),
		loc:   locdao.New(c),
		mager: magerdao.New(c),
		// whitelist
		whiteMidCache:   map[int64]int{},
		whiteBuvidCache: map[string]int{},
		// cron
		cron: cron.New(),
		inf2: ic,
		prom: prom.BusinessInfoCount,
	}
	s.initCron()
	s.cron.Start()
	return s
}

func (s *Service) initCron() {
	s.loadWhiteListCache()
	s.loadBrandCache()
	s.loadCollectionSplash()
	var err error
	if err = s.cron.AddFunc(s.c.Cron.LoadWhiteListCache, s.loadWhiteListCache); err != nil {
		panic(err)
	}
	if err = s.cron.AddFunc(s.c.Cron.LoadBrandSplash, s.loadBrandCache); err != nil {
		panic(err)
	}
	if err = s.cron.AddFunc(s.c.Cron.LoadCollectionBrandSplash, s.loadCollectionSplash); err != nil {
		panic(err)
	}
}

// 仅输出会展示 topview 的 banner 资源位
// 3143/3150/4332/4336
func filterTopViewBannerResource(in int64) int64 {
	topViewSet := sets.NewInt64(3143, 3150, 4332, 4336)
	if topViewSet.Has(in) {
		return in
	}
	return 0
}

// AdList ad splash list
func (s *Service) AdList(c context.Context, plat int8, mobiApp, device, buvid, birth, adExtra string, height, width, build int, mid int64, userAgent, network, loadedCreativeList, clientKeepIds string) (res *splash.CmSplash, err error) {
	var (
		list      []*splash.List
		show      []*splash.Show
		config    *splash.CmConfig
		topview   map[int64]int64
		requestID string
	)
	if ok := model.IsOverseas(plat); ok {
		err = ecode.NotModified
		return
	}
	bannerResourceID := feedcommon.BannerResourceID(c, mobiApp, int64(build), 0, plat)
	bannerResourceID = filterTopViewBannerResource(bannerResourceID)
	targetLoadCreativeList := s.deriveLoadCreativeList(loadedCreativeList)
	g, ctx := errgroup.WithContext(c)
	g.Go(func() error {
		var e error
		if list, config, topview, e = s.ad.SplashList(ctx, mobiApp, device, buvid, birth, adExtra, height, width, build,
			mid, bannerResourceID, userAgent, network, targetLoadCreativeList, clientKeepIds); e != nil {
			log.Error("cm s.ad.SplashList error(%v)", e)
			return e
		}
		return nil
	})
	g.Go(func() error {
		var e error
		if show, requestID, e = s.ad.SplashShow(ctx, mobiApp, device, buvid, birth, adExtra, height, width, build, mid, userAgent, network); e != nil {
			log.Error("cm s.ad.SplashShow error(%v)", e)
			return e
		}
		return nil
	})
	if err = g.Wait(); err != nil {
		log.Error("cm splash errgroup.WithContext error(%v)", err)
		return
	}
	// 替换show里面的id改成topview id
	for _, v := range show {
		if t, ok := topview[v.ID]; ok {
			v.ID = t
		}
	}
	res = &splash.CmSplash{
		CmConfig:        config,
		List:            list,
		Show:            show,
		SplashRequestId: requestID,
	}
	return
}

type AdShowReply struct {
	Show            []*splash.AdShow `json:"show"`
	SplashRequestId string           `json:"splash_request_id"`
}

func (s *Service) AdRtShow(ctx context.Context, param *splash.SplashRequest) (*AdShowReply, error) {
	if ok := model.IsOverseas(param.Plat); ok {
		return nil, ecode.NotModified
	}
	bannerResourceID := feedcommon.BannerResourceID(ctx, param.MobiApp, int64(param.Build), 0, param.Plat)
	bannerResourceID = filterTopViewBannerResource(bannerResourceID)
	showResult, err := s.ad.SplashShowSearch(ctx, &advo.SspSplashRequestVo{
		MobiApp:        param.MobiApp,
		Height:         int32(param.Height),
		Width:          int32(param.Width),
		Build:          int32(param.Build),
		Birth:          param.Birth,
		Mid:            param.Mid,
		Ip:             metadata.String(ctx, metadata.RemoteIP),
		Device:         param.Device,
		Buvid:          param.Buvid,
		AdExtra:        param.AdExtra,
		BannerResource: int32(bannerResourceID),
	})
	if err != nil {
		log.Error("Failed to request SplashShowSearch: %+v", err)
		return nil, err
	}
	return &AdShowReply{
		Show:            convertToSplashShow(ctx, showResult.GetShowPeriods()),
		SplashRequestId: showResult.GetSplashRequestId(),
	}, nil
}

func convertToSplashShow(ctx context.Context, show []*advo.SplashShowPeriodVo) []*splash.AdShow {
	out := make([]*splash.AdShow, 0, len(show))
	for _, v := range show {
		out = append(out, &splash.AdShow{
			ID:            v.GetId(),
			Stime:         xtime.Time(v.GetStime()),
			Etime:         xtime.Time(v.GetEtime()),
			Adcb:          v.GetAdCb(),
			SplashContent: parseSplashContent(ctx, v.GetSplashContent()),
		})
	}
	return out
}

func parseSplashContent(ctx context.Context, content string) *splash.List {
	if content == "" {
		return nil
	}
	splashContent := &splash.List{}
	if err := json.Unmarshal([]byte(content), splashContent); err != nil {
		log.Error("Failed to Unmarshal SplashContent: %+v", errors.WithStack(err))
		return nil
	}
	splashContent.IsAdLoc = true
	splashContent.ClientIP = metadata.String(ctx, metadata.RemoteIP)
	return splashContent
}

// State splash state
func (s *Service) State(c context.Context, buvid string, mid int64, state int8) (res *splash.Config) {
	res = &splash.Config{}
	var (
		_showClose    = int8(0) // 展示效果：不展示
		_showOpen     = int8(1) // 展示效果：展示
		_stateDefalut = int8(0) // 开关效果：不控制
		// _stateClose   = int8(1) // 关闭启动动画（客户端展示开关打开）
		_stateOpen  = int8(2) // 开启启动动画（客户端展示开关关闭）
		svrState    int
		ok          bool
		abtestState = s.c.Splash.AbtestState
	)
	if svrState, ok = s.whiteMidCache[mid]; !ok {
		if svrState, ok = s.whiteBuvidCache[buvid]; !ok {
			//nolint:gomnd
			switch abtestState {
			case 1: // 第二阶段【关闭闪屏设置】开关关闭的设备、【关闭闪屏设置】开关打开的设备
				switch state {
				case _stateOpen: // 开启启动动画（客户端展示开关关闭）
					svrState = 1
				default:
					svrState = 0
				}
			case 2: // 第三阶段 升级至5.46版本的所有设备
				svrState = 2
			default: // 第一阶段 升级至5.46版本的所有设备
				svrState = 0
			}
		}
	}
	//nolint:gomnd
	switch svrState {
	case 1: // 展示效果：不展示、开关效果：不控制
		res.Show = _showClose
		res.State = _stateDefalut
	case 2: // 展示效果：不展示、开关效果：关闭
		res.Show = _showClose
		res.State = _stateOpen
	default: // 默认 展示效果：展示、开关效果：不控制
		res.Show = _showOpen
		res.State = _stateDefalut
	}
	return
}

// Close dao
func (s *Service) Close() {
	s.dao.Close()
}

func (s *Service) loadWhiteListCache() {
	log.Info("cronLog start loadWhiteListCache")
	var (
		filePath = s.c.Splash.WhiteFile
		tmpMids  = map[int64]int{}
	)
	file, err := os.Open(s.c.Splash.WhiteFile)
	if err != nil {
		log.Error("os.Open(%s) error(%v)", filePath, err)
		return
	}
	defer file.Close()

	bs, _ := ioutil.ReadAll(file)
	var list struct {
		Mids   map[string]int `json:"mids"`
		Buvids map[string]int `json:"buvids"`
	}
	if err = json.Unmarshal(bs, &list); err != nil {
		log.Error("json.Unmarshal() file(%s) error(%v)", filePath, err)
		return
	}
	for mid, state := range list.Mids {
		midInt, err := strconv.ParseInt(mid, 10, 64)
		if err != nil {
			continue
		}
		tmpMids[midInt] = state
	}
	s.whiteMidCache = tmpMids
	s.whiteBuvidCache = list.Buvids
}

func (s *Service) deriveLoadCreativeList(srcList string) string {
	lcl := strings.Split(srcList, ",")
	loadedCreativeSet := sets.NewString(lcl...)
	if len(lcl) != loadedCreativeSet.Len() {
		s.prom.Incr("loadedCreativeList有重复id")
	}
	loadedCreativeSet.Delete("")
	return strings.Join(loadedCreativeSet.List(), ",")
}
