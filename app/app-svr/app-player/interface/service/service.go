package service

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"math"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"go-common/component/metadata/device"
	"go-common/library/cache/redis"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/pipeline/fanout"

	xecode "go-gateway/app/app-svr/app-player/ecode"
	api "go-gateway/app/app-svr/app-player/interface/api/playurl"
	"go-gateway/app/app-svr/app-player/interface/conf"
	playurldao "go-gateway/app/app-svr/app-player/interface/dao/playurl"
	vipdao "go-gateway/app/app-svr/app-player/interface/dao/vip"
	"go-gateway/app/app-svr/app-player/interface/model"
	"go-gateway/app/app-svr/archive/middleware"
	playurlV2Api "go-gateway/app/app-svr/playurl/service/api/v2"

	locgrpc "git.bilibili.co/bapis/bapis-go/community/service/location"
	ott "git.bilibili.co/bapis/bapis-go/ott/service"
	"git.bilibili.co/go-tool/libbdevice/pkg/pd"
)

const (
	_androidBuild = 5340000
	_iosBuild     = 8230
	_ipadHDBuild  = 12070
	_ispCU        = "联通"
	_ispCT        = "电信"
	_ispCM        = "移动"
	SteinsUpgrade = 1
	PayArcUpgrade = 2
)

type CdnScore struct {
	WwanScoreIps []string
	WifiScoreIps []string
	LastUpTime   int64
}

// Service is space service
type Service struct {
	c           *conf.Config
	playURLDao  *playurldao.Dao
	vipDao      *vipdao.Dao
	cdnRedis    *redis.Pool
	cdnScores   map[string]CdnScore
	cdnScoresMu sync.RWMutex
	cache       *fanout.Fanout
	locGRPC     locgrpc.LocationClient
}

// New new space
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:          c,
		playURLDao: playurldao.New(c),
		vipDao:     vipdao.New(c),
		cdnRedis:   redis.NewPool(c.Redis.CdnScore),
		cdnScores:  make(map[string]CdnScore),
		cache:      fanout.New("cdn_cache"),
	}
	var err error
	if s.locGRPC, err = locgrpc.NewClient(c.LocationGRPC); err != nil {
		panic(fmt.Sprintf("locgrpc.NewClient err(%+v)", err))
	}
	return
}

// Close dao.
func (s *Service) Close() {
	s.cache.Close()
}

func (s *Service) validBuild(plat int8, build int32) (aid, cid int64) {
	if (model.IsIphone(plat) && build < _iosBuild) || (model.IsAndroid(plat) && build < _androidBuild) {
		aid = s.c.Custom.PhoneAid
		cid = s.c.Custom.PhoneCid
	} else if model.IsIPad(plat) && build <= _iosBuild {
		aid = s.c.Custom.PadAid
		cid = s.c.Custom.PadCid
	} else if model.IsIPadHD(plat) && build <= _ipadHDBuild {
		aid = s.c.Custom.PadHDAid
		cid = s.c.Custom.PadHDCid
	}
	return
}

// PlayURLV2 is
func (s *Service) PlayURLV2(c context.Context, mid int64, params *model.Param, plat int8) (playurl *model.PlayurlV2Reply, err error) {
	var reply *playurlV2Api.PlayURLReply
	upgradeAid, upgradeCid := s.validBuild(plat, params.Build)
	if reply, err = s.playURLDao.PlayURLV2(c, params, mid, upgradeAid, upgradeCid); err != nil {
		log.Error("d.playDao.PlayURLV2 error(%+v)", err)
		if ecode.EqualError(xecode.PlayURLSteinsUpgrade, err) {
			playurl = new(model.PlayurlV2Reply)
			playurl.UpgradeLimit = s.UpgradeLimit(params.MobiApp)
			err = nil
		}
		return
	}
	if reply == nil || reply.Playurl == nil {
		return
	}
	playurl = new(model.PlayurlV2Reply)
	playurl.FormatPlayURL(reply.Playurl)
	return
}

// PlayURLGRPC is
func (s *Service) PlayURLGRPC(c context.Context, mid int64, params *model.Param, plat int8) (playurl *api.PlayURLReply, err error) {
	var reply *playurlV2Api.PlayURLReply
	upgradeAid, upgradeCid := s.validBuild(plat, params.Build)
	if reply, err = s.playURLDao.PlayURLV2(c, params, mid, upgradeAid, upgradeCid); err != nil {
		log.Error("d.playURLDao.PlayURLV2 error(%+v)", err)
		if ecode.EqualError(xecode.PlayURLSteinsUpgrade, err) {
			playurl = new(api.PlayURLReply)
			uplimit := s.UpgradeLimit(params.MobiApp)
			playurl.UpgradeLimit = &api.UpgradeLimit{
				Code:    int32(uplimit.Code),
				Message: uplimit.Message,
				Image:   uplimit.Image,
				Button: &api.UpgradeButton{
					Title: uplimit.Button.Title,
					Link:  uplimit.Button.Link,
				},
			}
			err = nil
		}
		return
	}
	if reply == nil || reply.Playurl == nil {
		return
	}
	playurl = model.FormatPlayURLGRPC(reply.Playurl)
	return
}

// UpgradeLimit is
func (s *Service) UpgradeLimit(mobiApp string) (upgradeLimit *model.UpgradeLimit) {
	var cfg = s.c.Custom.SteinsBuild
	upgradeLimit = &model.UpgradeLimit{
		Message: cfg.Message,
		Image:   cfg.Image,
	}
	button := &model.UpgradeButton{
		Title: cfg.ButtonText,
	}
	if cfg.UseCustomLink { // 支持开关
		switch mobiApp {
		case "iphone":
			button.Link = cfg.LinkPink
		case "ipad":
			button.Link = cfg.LinkHD
		case "iphone_b":
			button.Link = cfg.LinkBlue
		case "android":
			button.Link = cfg.LinkAndroid
		default:
			button.Link = cfg.ButtonLink
		}
	} else {
		button.Link = cfg.ButtonLink
	}
	upgradeLimit.Button = button
	upgradeLimit.Code = xecode.PlayURLSteinsUpgrade.Code()
	return
}

// Project is
func (s *Service) Project(c context.Context, mid int64, params *model.Param, plat int8) (playurl *api.PlayURLReply, err error) {
	var reply *playurlV2Api.ProjectReply
	if reply, err = s.playURLDao.Project(c, params, mid); err != nil {
		log.Error("d.playURLDao.Project error(%+v)", err)
		return
	}
	if reply == nil || reply.Playurl == nil {
		return
	}
	playurl = model.FormatPlayURLGRPC(reply.Playurl)
	return
}

// DlNum is
func (s *Service) DlNum(c context.Context, mid int64, params *model.DlNumParam) (err error) {
	if err = s.vipDao.ReportOfflineDownloadNum(c, mid, params); err != nil {
		log.Error("d.vipDao.ReportOfflineDownloadNum error(%+v)", err)
		return
	}
	return
}

// PlayConfEdit .
func (s *Service) PlayConfEdit(c context.Context, cloudParam *model.CloudEditParam, arg *api.PlayConfEditReq) (*api.PlayConfEditReply, error) {
	err := s.playURLDao.PlayEdit(c, cloudParam, arg)
	if err != nil {
		log.Error("s.playURLDao.PlayEdit(%v %v) error(%v)", cloudParam, arg.PlayConf, err)
		return nil, err
	}
	return &api.PlayConfEditReply{}, nil
}

// PlayConf .
func (s *Service) PlayConf(c context.Context, cloudParam *model.CloudEditParam, mid int64) (*api.PlayConfReply, error) {
	rly, e := s.playURLDao.PlayConf(c, cloudParam, mid)
	if e != nil {
		log.Error("s.playURLDao.PlayConf(%s) error(%v)", cloudParam.Buvid, e)
		return nil, e
	}
	reply := &api.PlayConfReply{}
	if rly != nil && rly.PlayConf != nil {
		// 云控信息拼接 用户设置信息
		reply.PlayConf = model.FormatPlayConf(rly.PlayConf)
	}
	return reply, nil
}

// PlayView  .
func (s *Service) PlayView(c context.Context, mid int64, params *model.Param, plat int8) (*api.PlayViewReply, error) {
	upgradeAid, upgradeCid := s.validBuild(plat, params.Build)
	reply, err := s.playURLDao.PlayView(c, params, mid, upgradeAid, upgradeCid)
	if err != nil {
		res := &api.PlayViewReply{}
		//处理未付费
		if ecode.EqualError(xecode.PlayURLArcPayNotice, err) {
			res.PlayLimit = s.buildSeasonPayPlayLimit()
			return res, nil
		}
	}
	// 视频云信息
	if reply == nil {
		log.Error("d.playURLDao.PlayView reply is nil params(%+v) mid(%d)", params, mid)
		return nil, ecode.NothingFound
	}

	locInfo, err := s.locGRPC.Info2(c, &locgrpc.AddrReq{Addr: params.IP})
	if err != nil {
		log.Error("s.locGRPC.Info2 err(%+v) ip(%s)", err, params.IP)
	}

	playRly := &api.PlayViewReply{}
	if reply.PlayUrl != nil {
		// 互动视频升级提示信息
		if reply.PlayUrl.IsSteinsUpgrade == SteinsUpgrade {
			uplimit := s.UpgradeLimit(params.MobiApp)
			playRly.UpgradeLimit = &api.UpgradeLimit{
				Code:    int32(uplimit.Code),
				Message: uplimit.Message,
				Image:   uplimit.Image,
				Button: &api.UpgradeButton{
					Title: uplimit.Button.Title,
					Link:  uplimit.Button.Link,
				},
			}
		}
		// 付费视频提示升级
		if reply.PlayUrl.IsSteinsUpgrade == PayArcUpgrade {
			playRly.UpgradeLimit = s.buildSeasonPayUpgradeLimit(c)
		}
		// 视频播放地址处理
		if reply.PlayUrl.Playurl != nil {
			cdnScore := s.calCdnScore(c, reply.PlayUrl.Playurl, params.Buvid, locInfo, mid)
			var vipFree = reply.GetVipConf().GetLimitFree() == 1
			playRly.VideoInfo = model.FormatPlayInfoGRPC(reply.PlayUrl.Playurl, reply.PlayUrl.ExtInfo, params, vipFree, cdnScore, reply.GetVipConf().GetSubtitle())
		}
		// 视频音量信息处理
		if reply.Volume != nil && playRly.VideoInfo != nil {
			playRly.VideoInfo.Volume = &api.VolumeInfo{
				MeasuredI:         reply.Volume.MeasuredI,
				MeasuredLra:       reply.Volume.MeasuredLra,
				MeasuredTp:        reply.Volume.MeasuredTp,
				MeasuredThreshold: reply.Volume.MeasuredThreshold,
				TargetOffset:      reply.Volume.TargetOffset,
				TargetI:           reply.Volume.TargetI,
				TargetTp:          reply.Volume.TargetTp,
			}
		}
	}
	//对安卓特定机型做清晰度的屏蔽
	defer s.qnShieldToAndroid(c, playRly)
	// 如果是下载请求，只需要播放地址相关信息即可
	if params.Download > 0 {
		return playRly, nil
	}
	// 云控信息拼接 用户设置信息 54版本才需要（service做了版本控制）
	// 64版本后返回播放配置值的信息
	if reply.PlayConf != nil {
		playRly.PlayConf = model.FormatPlayConf(reply.PlayConf)
	}
	// 云控稿件相关设置
	if reply.PlayArc != nil {
		playRly.PlayArc = model.FormatPlayArcConf(reply.PlayArc)
	}
	// chronos
	if reply.Chronos != nil {
		playRly.Chronos = &api.Chronos{File: reply.Chronos.File, Md5: reply.Chronos.Md5}
	}
	if shake := reply.GetEvent().GetShake(); shake != nil {
		playRly.Event = &api.Event{
			Shake: &api.Shake{File: shake.File},
		}
	}
	if reply.Ab != nil && reply.Ab.Glance != nil {
		playRly.Ab = &api.AB{
			Glance: &api.Glance{Duration: reply.Ab.Glance.Duration, Times: reply.Ab.Glance.Times, CanWatch: reply.Ab.Glance.CanWatch},
			Group:  api.Group(reply.Ab.Group),
		}
	}
	return playRly, nil
}

func (s *Service) qnShieldToAndroid(ctx context.Context, in *api.PlayViewReply) {
	dev, _ := device.FromContext(ctx)
	shieldKey := fmt.Sprintf("%s_%s", dev.Model, dev.Brand)
	shieldQnListStr, ok := s.c.AndroidQnShield[shieldKey]
	if !ok {
		return
	}
	qnShieldMap, err := parseShieldQnStr(shieldQnListStr)
	if err != nil {
		log.Error("parseShieldQnStr error(%+v)", err)
		return
	}
	var (
		tempStreamList         []*api.Stream
		defaultQnShouldDegrade bool
	)
	if in.VideoInfo == nil {
		return
	}
	// streamlist中的清晰度由高到低
	for _, v := range in.VideoInfo.StreamList {
		if _, ok := qnShieldMap[int64(v.StreamInfo.Quality)]; !ok {
			tempStreamList = append(tempStreamList, v)
			if defaultQnShouldDegrade { // 需要降级
				in.VideoInfo.Quality = v.StreamInfo.Quality
			}
			continue
		}
		if in.VideoInfo.Quality == v.StreamInfo.Quality { //默认播放的清晰度被屏蔽,降一路
			defaultQnShouldDegrade = true
		}
	}
	in.VideoInfo.StreamList = tempStreamList
}

func parseShieldQnStr(in string) (map[int64]struct{}, error) {
	var (
		qnStrList   = strings.Split(in, ",")
		qnShieldMap = make(map[int64]struct{}, len(qnStrList))
	)

	for _, v := range qnStrList {
		vi, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, err
		}
		qnShieldMap[vi] = struct{}{}
	}
	return qnShieldMap, nil
}

func (s *Service) Bubble(c context.Context, param *model.BubbleParam, mid int64) (*ott.ProjectionActivityReply, error) {
	reply, err := s.playURLDao.Bubble(c, param, mid)
	if err != nil {
		log.Error("Bubble param %v mid %d err %v", param, mid, err)
		return &ott.ProjectionActivityReply{Show: false}, nil
	}
	return reply, nil
}

func (s *Service) BubbleSubmit(c context.Context, mid int64, code string) (*ott.ProjectionActivitySubmitReply, error) {
	reply, err := s.playURLDao.BubbleSubmit(c, &ott.ProjectionActivitySubmitReq{
		Mid:  mid,
		Code: code,
	})

	if err != nil {
		log.Error("BubbleSubmit error mid:%d code:%s err:%+v", mid, code, err)
		return nil, err
	}

	return reply, nil
}

func (s *Service) calCdnScore(c context.Context, p *playurlV2Api.ResponseMsg, buvid string, locInfo *locgrpc.InfoCompleteReply, mid int64) map[string]map[string]string {
	if hit := s.hitTest(mid, buvid); !hit {
		return nil
	}
	var provinceID int64
	zoneID := locInfo.GetInfo().GetZoneId()
	zoneLen := 3
	if len(zoneID) >= zoneLen {
		provinceID = zoneID[2]
	}
	// 提取需要进行处理的域名（音频不处理）
	if p.Dash == nil || provinceID == 0 {
		return nil
	}
	// 运营商改为完全匹配，忽略三大运营商之外请求
	isp := ""
	locIsp := locInfo.GetInfo().Isp
	switch locIsp {
	case _ispCM:
		isp = "cm"
	case _ispCT:
		isp = "ct"
	case _ispCU:
		isp = "cu"
	default:
		return nil
	}
	// 将需要处理的第三方域名拿出来
	tmpVideoHost := make(map[string]struct{})
	for _, v := range p.Dash.Video {
		for _, bkp := range v.BackupUrl {
			bdm := model.GetThirdDomain(bkp)
			if bdm != "" {
				tmpVideoHost[bdm] = struct{}{}
			}
		}
		dm := model.GetThirdDomain(v.BaseUrl)
		if dm != "" {
			tmpVideoHost[dm] = struct{}{}
		}
	}
	for _, v := range p.Dash.Audio {
		for _, bkp := range v.BackupUrl {
			bdm := model.GetThirdDomain(bkp)
			if bdm != "" {
				tmpVideoHost[bdm] = struct{}{}
			}
		}
		dm := model.GetThirdDomain(v.BaseUrl)
		if dm != "" {
			tmpVideoHost[dm] = struct{}{}
		}
	}
	// 如果没有第三方域名
	if len(tmpVideoHost) == 0 {
		return nil
	}
	// 获取第三方域名选中的ip
	cdnMap := make(map[string]map[string]string)
	for vh := range tmpVideoHost {
		chooseIp := s.chooseCdn(c, middleware.CdnZoneKey(vh, isp, provinceID))
		if len(chooseIp) == 0 {
			continue
		}
		cdnMap[vh] = chooseIp
	}
	return cdnMap
}

func (s *Service) hitTest(mid int64, buvid string) bool {
	for _, cmid := range s.c.Custom.CdnMids {
		if mid == cmid {
			return true
		}
	}
	return crc32.ChecksumIEEE([]byte(buvid))%1000 < s.c.Custom.CdnScoreGray
}

func (s *Service) chooseCdn(c context.Context, cdnZoneKey string) map[string]string {
	s.cdnScoresMu.RLock()
	score, ok := s.cdnScores[cdnZoneKey]
	s.cdnScoresMu.RUnlock()
	if !ok || time.Now().Unix()-score.LastUpTime > s.c.Custom.ScoreInternal { //暂定60秒更新一次可配置
		s.cache.Do(c, func(c context.Context) {
			s.setCdnCache(cdnZoneKey)
		})
	}

	res := make(map[string]string, 2)
	rand.Seed(time.Now().UnixNano())
	if len(score.WifiScoreIps) != 0 {
		ix := rand.Intn(len(score.WifiScoreIps))
		res["wifi"] = score.WifiScoreIps[ix]
	}
	if len(score.WwanScoreIps) != 0 {
		ix := rand.Intn(len(score.WwanScoreIps))
		res["wwan"] = score.WwanScoreIps[ix]
	}
	return res
}

func (s *Service) setCdnCache(cdnZoneKey string) {
	//从redis获取host对应ip及评分
	conn := s.cdnRedis.Get(context.Background())
	defer conn.Close()
	bs, err := redis.Bytes(conn.Do("GET", cdnZoneKey))
	if err != nil {
		if err == redis.ErrNil {
			s.writeCdnScore(cdnZoneKey, CdnScore{LastUpTime: time.Now().Unix()})
			return
		}
		log.Error("calCdnScore redis.ByteSlices err(%+v) cdnZoneKey(%s)", err, cdnZoneKey)
		return
	}
	var cdnScore map[string]map[string]float64
	if err := json.Unmarshal(bs, &cdnScore); err != nil {
		log.Error("calCdnScore json.Unmarshal err(%+v) cdnZoneKey(%s)", err, cdnZoneKey)
		return
	}
	s.writeCdnScore(cdnZoneKey, CdnScore{WwanScoreIps: s.getBestIpList(cdnScore["wwan"]),
		WifiScoreIps: s.getBestIpList(cdnScore["wifi"]), LastUpTime: time.Now().Unix()})
}

func (s *Service) getBestIpList(ipScores map[string]float64) []string {
	if len(ipScores) == 0 {
		return nil
	}
	var list []*model.IpScore
	for ip, score := range ipScores {
		list = append(list, &model.IpScore{
			Ip:    ip,
			Score: score,
		})
	}
	// 分数越小越靠前
	sort.SliceStable(list, func(i, j int) bool {
		return list[i].Score < list[j].Score
	})
	ix := int64(math.Floor(float64(len(list)) * s.c.Custom.ScoreRank))
	if ix < 1 {
		ix = 1
	}
	list = list[:ix]
	var bestIpList []string
	for _, v := range list {
		bestIpList = append(bestIpList, v.Ip)
	}
	return bestIpList
}

func (s *Service) writeCdnScore(cdnZoneKey string, cdnScore CdnScore) {
	log.Warn("writeCdnScore cdnZoneKey(%s) cdnScore(%+v)", cdnZoneKey, cdnScore)
	s.cdnScoresMu.Lock()
	s.cdnScores[cdnZoneKey] = cdnScore
	s.cdnScoresMu.Unlock()
}

func (s *Service) ProjPageAct(c context.Context, params *model.ProjPageParam) (*ott.ProjPageActReply, error) {
	act, err := s.playURLDao.ProjPageAct(c, params)
	if err != nil {
		log.Error("ProjPageAct error, param: %+v, err: %+v", params, err)
		return nil, ecode.RequestErr
	}
	return act, nil
}

func (s *Service) ProjActAll(c context.Context, params *model.ProjActAllParam) (*ott.ProjActivityAllReply, error) {
	act, err := s.playURLDao.ProjActAll(c, params)
	if err != nil {
		log.Error("ProjActAll error, param: %+v, err: %+v", params, err)
		return nil, err
	}
	return act, nil
}

// buildPlayLimit 返回付费合集播放限制提示
func (s *Service) buildSeasonPayPlayLimit() *api.PlayLimit {
	return &api.PlayLimit{
		Code:    api.PlayLimitCode_PLCUgcNotPayed,
		Message: s.c.Custom.UpgradeInfo.PlayLimitMessage,
		Button: &api.ButtonStyle{
			Text:      s.c.Custom.UpgradeInfo.PlayLimitButtonText,
			TextColor: "#FFFFFF",
			BgColor:   "#FFB027",
		},
	}
}

// buildSeasonPayUpgradeLimit 返回付费合集升级提示
func (s *Service) buildSeasonPayUpgradeLimit(ctx context.Context) *api.UpgradeLimit {
	versionMatch := pd.WithContext(ctx).Where(func(pdContext *pd.PDContext) {
		pdContext.IsPlatIPadHD().And().Build(">=", 34300000)
	}).OrWhere(func(pdContext *pd.PDContext) {
		pdContext.IsPlatIPad().And().Build(">=", 67500000)
	}).FinishOr(false)

	code := int32(xecode.PlayURLSteinsUpgrade.Code())

	if versionMatch {
		code = int32(xecode.PlayURLArcPayUpgrade.Code())
	}

	return &api.UpgradeLimit{
		Code:    code,
		Message: s.c.Custom.UpgradeInfo.UpgradeLimitMessage,
		Button: &api.UpgradeButton{
			Title: s.c.Custom.UpgradeInfo.UpgradeLimitButtonText,
			Link:  "bilibili://base/app-upgrade",
		},
	}
}
