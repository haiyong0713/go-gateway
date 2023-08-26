package bind

import (
	"context"
	"go-common/library/cache/memcache"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	errgroup2 "go-common/library/sync/errgroup.v2"
	v1 "go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/interface/conf"
	"go-gateway/app/web-svr/activity/interface/dao/bind"
	bind2 "go-gateway/app/web-svr/activity/interface/model/bind"
	"go-gateway/app/web-svr/activity/interface/tool"
	"go-gateway/app/web-svr/activity/tools/lib/initialize"
	"strconv"
	"time"

	api "git.bilibili.co/bapis/bapis-go/account/service"
	"github.com/bluele/gcache"
)

const (
	_defaultCacheTtl      = 300
	_defaultRefreshTicker = 1
)

var localS *Service

type Service struct {
	conf                   *conf.Config
	dao                    *bind.Dao
	bindConfigCacheMapping gcache.Cache
}

func New(c *conf.Config) *Service {
	if localS != nil {
		return localS
	}
	s := &Service{
		conf:                   c,
		dao:                    bind.New(c),
		bindConfigCacheMapping: gcache.New(c.BindConf.CacheSize).LFU().Build(),
	}
	go initialize.CallC(s.storeBindConfigTicker)
	localS = s
	return s
}

func (s *Service) storeBindConfigTicker(ctx context.Context) (err error) {
	duration := time.Duration(_defaultRefreshTicker) * time.Second
	if conf.Conf.BindConf != nil && conf.Conf.BindConf.RefreshTickerSecond != 0 {
		duration = time.Duration(conf.Conf.BindConf.RefreshTickerSecond) * time.Second
	}
	ticker := time.NewTicker(duration)
	for {
		select {
		case <-ticker.C:
			s.storeBindConfig(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (s *Service) storeBindConfig(ctx context.Context) {
	configsMapping, err := s.dao.GetBindConfigsCache(ctx)
	if err != nil {
		log.Errorc(ctx, "[StoreBindConfig][Error], err:%+v", err)
		return
	}
	if len(configsMapping) == 0 {
		s.bindConfigCacheMapping = gcache.New(s.conf.BindConf.CacheSize).LFU().Build()
		return
	}
	for k, v := range configsMapping {
		_ = s.bindConfigCacheMapping.SetWithExpire(k, v, time.Second*_defaultCacheTtl)
	}
}

func (s *Service) GetBindConfig(ctx context.Context, req *v1.GetBindConfigReq) (resp *v1.GetBindConfigResp, err error) {
	log.Errorc(ctx, "%+v", req)
	resp = new(v1.GetBindConfigResp)
	if req == nil || req.ID == 0 {
		err = xecode.Errorf(xecode.RequestErr, "参数错误")
		return
	}
	configInfo, err := s.getBindConfig(ctx, req.ID, req.SkipCache)
	if err != nil {
		log.Errorc(ctx, "[GetBindConfig][getBindConfig][Error], err:%+v, req:%+v", err, req)
		return
	}
	resp.ConfigInfo = configInfo
	return
}

func (s *Service) getBindConfig(ctx context.Context, configId int64, skipCache bool) (configInfo *v1.BindConfigInfo, err error) {
	configInfo = new(v1.BindConfigInfo)
	if !skipCache {
		configInfo, err = s.getBindInfoFromCache(ctx, configId)
		if err != nil && err != memcache.ErrNotFound {
			log.Errorc(ctx, "[GetBindConfig][GetBindInfoFromCache][Error], err:%+v, id:%+v", err, configId)
			return
		}
	}
	// 回源
	configs, err := s.dao.GetAllBindConfigs(ctx)
	if err != nil {
		log.Errorc(ctx, "[GetBindConfig][GetAllBindConfigs][Error], err:%+v, id:%+v", err, configId)
		return
	}
	errS := s.dao.StoreBindConfigsCache(ctx, configs)
	if errS != nil {
		log.Errorc(ctx, "[GetBindConfig][StoreBindConfigsCache][Error], err:%+v, configId:%+v", err, configId)
		return
	}
	for _, v := range configs {
		if v.ID == configId {
			configInfo = v
			return
		}
	}
	err = xecode.Errorf(xecode.RequestErr, "配置不存在")
	return
}

func (s *Service) GetBindHandler(ctx context.Context, configId int64, mid int64) (bindParams *bind2.BindParams, err error) {
	if mid == 0 {
		err = xecode.Errorf(xecode.NoLogin, "未登陆")
		return
	}
	configInfo, err := s.getBindConfig(ctx, configId, false)
	if err != nil {
		return
	}
	if configInfo.BindAccount != bind2.BindAccountTrue {
		err = xecode.Errorf(xecode.RequestErr, "活动无需进行绑定")
		return
	}
	bindParams = new(bind2.BindParams)

	if configInfo.BindExternal != bind2.BindExternalTX {
		err = xecode.Errorf(xecode.RequestErr, "暂不支持的游戏类型")
	}
	gameConfig, err := s.GetGameConfig(ctx, configInfo.GameType)
	if err != nil {
		return
	}
	timestamp := time.Now().Unix()
	code, userInfo, err := s.getUserDetail(ctx, mid, gameConfig)
	if err != nil {
		return
	}
	sign := s.getTencentSign(ctx, code, gameConfig, configInfo, mid, timestamp)

	bindParams = &bind2.BindParams{
		Sign:       sign,
		BasePath:   gameConfig.BasePath,
		FaceUrl:    userInfo.Info.Face,
		Code:       code,
		T:          timestamp,
		LivePlatId: gameConfig.AppId,
		NickName:   userInfo.Info.Name,
		GameIdList: gameConfig.GameName,
		OriginId:   gameConfig.OriginId,
	}
	return
}

func (s *Service) getUserDetail(ctx context.Context, mid int64, gameConfig *conf.BindGame) (code string, userInfo *api.InfoReply, err error) {
	errgroup := errgroup2.WithContext(ctx)
	errgroup.Go(func(ctx context.Context) (errG error) {
		userInfo, errG = s.dao.GetUserInfo(ctx, mid)
		if errG != nil {
			log.Errorc(ctx, "[getTencentSign][GetUserInfo][Error], err:%+v", errG)
			return
		}
		return
	})
	errgroup.Go(func(ctx context.Context) (errG error) {
		code, errG = s.dao.GetCode(ctx, gameConfig.ClientId, gameConfig.Business, mid)
		if errG != nil {
			log.Errorc(ctx, "[getTencentSign][GetCode][Error], err:%+v", errG)
			return
		}
		return
	})
	err = errgroup.Wait()
	if err != nil {
		log.Errorc(ctx, "[getTencentSign][errgroup][wait], err:%+v", err)
		xecode.Errorf(xecode.ServerErr, "获取用户信息失败")
		return
	}
	return
}

func (s *Service) getBindInfoFromCache(ctx context.Context, configId int64) (configInfo *v1.BindConfigInfo, err error) {
	cache, errCache := s.bindConfigCacheMapping.Get(configId)
	if errCache != nil {
		log.Errorc(ctx, "[GetBindInfoFromCache][Error], err:%+v", errCache)
	} else {
		configInfo = cache.(*v1.BindConfigInfo)
		return
	}
	// 获取缓存
	configMapping, err := s.dao.GetBindConfigsCache(ctx)
	if err != nil {
		log.Errorc(ctx, "[GetBindInfoFromCache][GetBindConfigsCache][Error], err:%+v", err)
		return
	}
	configInfo, ok := configMapping[configId]
	if ok {
		return
	} else {
		err = xecode.Errorf(xecode.RequestErr, "配置不存在")
		return
	}
}

func (s *Service) getTencentSign(ctx context.Context, code string, gameConfig *conf.BindGame, configInfo *v1.BindConfigInfo, mid int64, timestamp int64) string {
	params := make(map[string]string)
	params["gameIdList"] = gameConfig.GameName
	params["code"] = code
	params["livePlatId"] = gameConfig.AppId
	params["t"] = strconv.FormatInt(timestamp, 10)
	return tool.TencentMd5Sign(gameConfig.SignKey, params)
}

func (s *Service) GetGameConfig(ctx context.Context, gameId int64) (gameConfig *conf.BindGame, err error) {
	if len(s.conf.BindConf.Games) == 0 {
		err = xecode.Errorf(xecode.RequestErr, "配置缺失")
		return
	}
	for _, v := range s.conf.BindConf.Games {
		if v.GameId == gameId {
			gameConfig = v
			return
		}
	}
	err = xecode.Errorf(xecode.RequestErr, "配置缺失")
	log.Errorc(ctx, "[GetGameConfig][MISS][Error], err:%+v, gameId:%+v, games:%+v", err, gameId, s.conf.BindConf.Games)
	return
}

func (s *Service) SaveBindConfig(ctx context.Context, in *v1.BindConfigInfo) (resp *v1.NoReply, err error) {
	resp = new(v1.NoReply)
	err = s.saveBindConfigParamsCheck(ctx, in)
	if err != nil {
		return
	}
	if in.ID == 0 {
		err = s.dao.InsertBindConfig(ctx, in)
	} else {
		err = s.dao.UpdateBindConfig(ctx, in)
	}
	return
}

func (s *Service) saveBindConfigParamsCheck(ctx context.Context, in *v1.BindConfigInfo) (err error) {
	if in == nil || in.GameType == 0 || in.BindExternal == 0 {
		err = xecode.Errorf(xecode.RequestErr, "必要参数为空")
		return
	}
	gameConfig, err := s.GetGameConfig(ctx, in.GameType)
	if err != nil {
		return
	}
	if gameConfig.ExternalId != in.BindExternal {
		err = xecode.Errorf(xecode.RequestErr, "绑定的第三方id传入不正确")
		return
	}
	return
}

func (s *Service) GetBindConfigList(ctx context.Context, in *v1.GetBindConfigListReq) (resp *v1.GetBindConfigListResp, err error) {
	resp = new(v1.GetBindConfigListResp)
	resp.List = make([]*v1.BindConfigInfo, 0)
	if in.ID != 0 {
		configInfo, errG := s.getBindConfig(ctx, in.ID, true)
		if errG != nil {
			err = errG
			return
		}
		resp.List = append(resp.List, configInfo)
		resp.Total = int64(len(resp.List))
		return
	} else {
		resp.List, resp.Total, err = s.dao.GetBindConfigByOffset(ctx, int(in.Pn), int(in.Ps))
		if err != nil {
			return
		}
	}
	return
}

func (s *Service) GetBindGames(ctx context.Context, in *v1.NoReply) (resp *v1.GetBindGamesResp, err error) {
	resp = new(v1.GetBindGamesResp)
	resp.Games = make([]*v1.BindGameInfo, 0)
	games := s.conf.BindConf.Games
	for _, game := range games {
		resp.Games = append(resp.Games, &v1.BindGameInfo{
			GameId:       game.GameId,
			GameName:     game.GameName,
			GameTitle:    game.GameTitle,
			ExternalName: game.ExternalName,
			ExternalId:   game.ExternalId,
		})
	}
	return
}

func (s *Service) GetBindAndGameConfig(ctx context.Context, configId int64) (configInfo *v1.BindConfigInfo, gameConfig *conf.BindGame, err error) {
	configInfo, err = s.getBindConfig(ctx, configId, false)
	if err != nil {
		log.Errorc(ctx, "[GetBindInfo][getBindConfig][Error], err:%+v", err)
		return
	}

	gameConfig, err = s.GetGameConfig(ctx, configInfo.GameType)
	if err != nil {
		return
	}
	return
}

func (s *Service) GetBindInfo(ctx context.Context, configId int64, mid int64, refresh int32) (userBindInfo *bind2.UserBindInfo, err error) {
	userBindInfo = &bind2.UserBindInfo{}
	configInfo, gameConfig, err := s.GetBindAndGameConfig(ctx, configId)
	if err != nil {
		return
	}
	userBindInfo.ConfigInfo = configInfo
	if userBindInfo.ConfigInfo.BindAccount == bind2.BindAccountTrue && userBindInfo.ConfigInfo.BindExternal == bind2.BindExternalTX {
		bindInfo, errG := s.dao.GetTencentBindInfo(ctx, gameConfig.ClientId, gameConfig.Business, gameConfig.GameName, userBindInfo.ConfigInfo.ActId, refresh, mid)
		if errG != nil {
			err = errG
			return
		}
		if userBindInfo.ConfigInfo.BindType == bind2.IsBindRole && bindInfo.BindType == bind2.IsBindTrue &&
			bindInfo.RoleInfo != nil && bindInfo.RoleInfo.RoleName != "" {
			bindInfo.BindType = bind2.IsBindRole
		}
		userBindInfo.BindInfo = bindInfo
	}
	return
}

func (s *Service) GetBindOpenId(ctx context.Context, configId int64, mid int64) (openId string, err error) {
	configInfo, err := s.getBindConfig(ctx, configId, false)
	if err != nil {
		return
	}
	gameConfig, err := s.GetGameConfig(ctx, configInfo.GameType)
	if err != nil {
		return
	}
	openId, err = s.dao.GetOpenIdByMid(ctx, gameConfig.ClientId, mid)
	return
}

func (s *Service) GetBindExternals(ctx context.Context, in *v1.NoReply) (resp *v1.GetBindExternalsResp, err error) {
	resp = new(v1.GetBindExternalsResp)
	resp.Externals = make([]*v1.BindExternal, 0)
	if len(s.conf.BindConf.ExternalConfig) == 0 {
		return
	}
	for _, v := range s.conf.BindConf.ExternalConfig {
		resp.Externals = append(resp.Externals, &v1.BindExternal{
			BindExternal: v.BindExternal,
			ExternalName: v.ExternalName,
		})
	}
	return
}

func (s *Service) RefreshBindConfigCache(ctx context.Context, in *v1.NoReply) (resp *v1.NoReply, err error) {
	resp = new(v1.NoReply)
	configs, err := s.dao.GetAllBindConfigs(ctx)
	if err != nil {
		return
	}
	err = s.dao.StoreBindConfigsCache(ctx, configs)
	if err != nil {
		return
	}
	return
}
