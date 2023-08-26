package service

import (
	"context"
	"time"

	"github.com/pkg/errors"

	locationGRPC "git.bilibili.co/bapis/bapis-go/community/service/location"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"

	model "go-gateway/app/app-svr/app-feed/admin/model/frontpage"
	frontpageEcode "go-gateway/app/app-svr/app-feed/ecode"
	pb "go-gateway/app/app-svr/resource/service/api/v1"
	pbv2 "go-gateway/app/app-svr/resource/service/api/v2"
)

func (s *Service) FrontPage(c context.Context, req *pb.FrontPageReq) (reply *pb.FrontPageResp, err error) {
	if reply, err = s.res.GetEffectiveFrontPage(c, req); err != nil {
		log.Error("service.FrontPage GetEffectiveFrontPage req(%+v) error(%+v)", req, err)
		return
	}
	if reply == nil {
		return nil, xecode.RequestErr
	}
	return
}

// GetFrontPageConfig 根据条件获取版头配置.v2
func (s *Service) GetFrontPageConfig(ctx context.Context, req *pbv2.GetFrontPageConfigReq) (res *pbv2.FrontPageConfig, err error) {
	var (
		ableConfig *model.Config
		ipZoneID   int64
		groupAuths map[int64]*locationGRPC.GroupAuth
	)
	defer func() {
		if err != nil || ableConfig == nil {
			var (
				configs        []*model.Config
				policyGroupIDs = make([]int64, 0)
			)
			configs = s.manager.FrontpageCacheGetOnlineConfigs(0)
			for _, config := range configs {
				_config := config
				if _config.LocPolicyGroupID > 0 {
					policyGroupIDs = append(policyGroupIDs, _config.LocPolicyGroupID)
				}
			}
			// 若存在生效配置但无区域策略限制，则直接取第一个（全球)
			if len(policyGroupIDs) == 0 && len(configs) > 0 {
				ableConfig = configs[0]
			} else {
				// 否则，若存在区域限制则代表存在生效配置，取策略详情进行下一步计算
				if len(policyGroupIDs) > 0 {
					if ipZoneID, groupAuths, err = s.ZlimitInfo2(ctx, policyGroupIDs, req.Ip); err != nil {
						log.Error("Service: GetConfig ZlimitInfo2 (%v, %s) error %v", policyGroupIDs, req.Ip, err)
					}
				}
				// 或者不存在区域策略，则会去默认兜底配置
				ableConfig = s.CacheGetFrontPageDefaultConfig(ctx, ipZoneID, groupAuths)
			}
		}
		res = &pbv2.FrontPageConfig{
			Id:               ableConfig.ID,
			Name:             ableConfig.ConfigName,
			ContractId:       ableConfig.ContractID,
			ResourceId:       ableConfig.ResourceID,
			Pic:              ableConfig.Pic,
			Litpic:           ableConfig.LitPic,
			Url:              ableConfig.URL,
			Rule:             ableConfig.Rule,
			Weight:           ableConfig.Weight,
			Agency:           ableConfig.Agency,
			Price:            float64(ableConfig.Price),
			State:            pbv2.State_Enum(ableConfig.State),
			Atype:            int32(ableConfig.Atype),
			Stime:            ableConfig.STime.Time().Unix(),
			Etime:            ableConfig.ETime.Time().Unix(),
			IsSplitLayer:     int32(ableConfig.IsSplitLayer),
			SplitLayer:       ableConfig.SplitLayer,
			LocPolicyGroupId: ableConfig.LocPolicyGroupID,
			Position:         ableConfig.Position,
			Auto:             int32(ableConfig.Auto),
			Ctime:            ableConfig.CTime.Time().Unix(),
			Cuser:            ableConfig.CUser,
			Mtime:            ableConfig.MTime.Time().Unix(),
			Muser:            ableConfig.MUser,
		}
	}()

	var (
		configs        []*model.Config
		policyGroupIDs = make([]int64, 0)
	)
	if configs = s.manager.FrontpageCacheGetOnlineConfigs(req.ResourceId); len(configs) == 0 {
		return
	}
	if len(configs) == 0 {
		return
	}
	for _, config := range configs {
		_config := config
		policyGroupIDs = append(policyGroupIDs, _config.LocPolicyGroupID)
	}
	if len(policyGroupIDs) == 0 && len(configs) > 0 {
		ableConfig = configs[0]
		return
	}
	if ipZoneID, groupAuths, err = s.ZlimitInfo2(ctx, policyGroupIDs, req.Ip); err != nil {
		log.Error("Service: GetConfig ZlimitInfo2 (%v, %s) error %v", policyGroupIDs, req.Ip, err)
		return
	}
	ableConfig = s.CalculateFrontpageConfigByPolicy(configs, ipZoneID, groupAuths)

	return
}

func (s *Service) CacheGetFrontPageDefaultConfig(ctx context.Context, ipZoneID int64, authGroups map[int64]*locationGRPC.GroupAuth) (res *model.Config) {
	var (
		err            error
		defaultConfigs []*model.Config
	)
	if ipZoneID > 0 && len(authGroups) > 0 {
		if defaultConfigs, err = s.CacheGetFrontPageOnlineDefaultConfigs(ctx); err == nil && len(defaultConfigs) > 0 {
			if res = s.CalculateFrontpageConfigByPolicy(defaultConfigs, ipZoneID, authGroups); res != nil && res.ID > 0 {
				return
			}
		}
		log.Warn("Service: CacheGetFrontPageDefaultConfig CacheGetFrontPageOnlineConfigs error %v 未找到对应默认配置，将查询兜底配置", err)
	}
	if res = s.manager.FrontpageCacheGetBaseDefaultConfig(); res != nil {
		return
	}
	log.Error("Service: CacheGetFrontPageDefaultConfig FrontpageCacheGetBaseDefaultConfig error %v 将查询DB", err)
	if res, err = s.manager.FrontpageGetBaseDefaultConfig(ctx); err == nil && res != nil {
		return
	}
	log.Error("Service: CacheGetFrontPageDefaultConfig FrontpageGetBaseDefaultConfig error %v 将读取配置文件", err)
	res = s.c.Frontpage.BaseDefaultConfig
	return
}

func (s *Service) CacheGetFrontPageOnlineDefaultConfigs(ctx context.Context) (res []*model.Config, err error) {
	if res = s.manager.FrontpageCacheGetOnlineConfigs(0); len(res) > 0 {
		return
	}
	log.Warn("Service: CacheGetFrontPageDefaultConfig FrontpageCacheGetBaseDefaultConfig error %v 将查询DB", err)
	if res, err = s.manager.FrontpageGetOnlineConfigs(ctx, 0); err == nil && res != nil {
		return
	}
	log.Error("Service: CacheGetFrontPageDefaultConfig FrontpageGetOnlineConfigs error %v", err)
	return
}

func (s *Service) CacheFlushFrontPageBaseDefaultConfig() (err error) {
	var (
		baseDefaultConfig *model.Config
	)

	if baseDefaultConfig, err = s.manager.FrontpageGetBaseDefaultConfig(context.Background()); err != nil || baseDefaultConfig == nil {
		log.Error("Service: CacheFlushFrontPageBaseDefaultConfig FrontpageGetBaseDefaultConfig error %v or got nil config", err)
		return
	}
	s.manager.FrontpageCacheSetBaseDefaultConfig(baseDefaultConfig)

	return
}

func (s *Service) CacheFlushFrontPageOnlineConfigs() (err error) {
	var (
		allMenuResourceIDs []int64
	)
	ctx, cancelFunc := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelFunc()
	if allMenuResourceIDs, err = s.res.GetAllMenuResourceIDs(ctx); err != nil {
		log.Error("Service: CacheFlushFrontPageOnlineConfigs GetAllMenuResourceIDs error %v", err)
		return
	}
	eg := errgroup.WithCancel(ctx)
	for _, rid := range allMenuResourceIDs {
		_rid := rid
		eg.Go(func(_ctx context.Context) error {
			var (
				configs []*model.Config
			)
			if configs, err = s.manager.FrontpageGetOnlineConfigs(_ctx, _rid); err != nil {
				log.Error("Service: CacheFlushFrontPageOnlineConfigs FrontpageGetOnlineConfigs %d error %v", _rid, err)
				return nil
			}
			s.manager.FrontpageCacheSetOnlineConfigs(_rid, configs)
			return nil
		})
	}

	return eg.Wait()
}

func (s *Service) startCacheFlushFrontPageProc() {
	log.Info("Service: startCacheFlushFrontPageProc Start")
	if err := s.CacheFlushFrontPageBaseDefaultConfig(); err != nil {
		panic(err)
	}
	if err := s.CacheFlushFrontPageOnlineConfigs(); err != nil {
		panic(err)
	}
	log.Info("Service: startCacheFlushFrontPageProc End")

	s.cacheFlushFrontPageTicker = time.NewTicker(10 * time.Minute)
	eg := errgroup.WithContext(context.Background())
	eg.Go(func(_ context.Context) error {
		for range s.cacheFlushFrontPageTicker.C {
			log.Info("Service: startCacheFlushFrontPageProc Start")
			if err := s.CacheFlushFrontPageBaseDefaultConfig(); err != nil {
				log.Error("Service: startCacheFlushFrontPageProc CacheFlushFrontPageBaseDefaultConfig error %v", err)
			}
			if err := s.CacheFlushFrontPageOnlineConfigs(); err != nil {
				log.Error("Service: startCacheFlushFrontPageProc CacheFlushFrontPageOnlineDefaultConfigs error %v", err)
			}
			log.Info("Service: startCacheFlushFrontPageProc End")
		}
		return nil
	})
}

func (s *Service) CalculateFrontpageConfigByPolicy(configs []*model.Config, ipZoneID int64, groupAuths map[int64]*locationGRPC.GroupAuth) (ableConfig *model.Config) {
	if len(configs) == 0 || ipZoneID == 0 || len(groupAuths) == 0 {
		return nil
	}
	for _, cfg := range configs {
		if cfg == nil || cfg.ID == 0 {
			continue
		}
		// 全球生效
		if cfg.LocPolicyGroupID == 0 && ableConfig == nil {
			ableConfig = cfg
			continue
		}
		gval, gok := groupAuths[cfg.LocPolicyGroupID]
		if !gok || gval == nil {
			// 没有找到对应的规则不下发
			continue
		}
		// 遍历当前组id下的规则
		var isSuccess bool
		for _, rule := range gval.PolicyAuths {
			if rule == nil {
				continue
			}
			rval, rok := rule.ZoneAuths[ipZoneID]
			// 在所有规则中未找到对应zoneid的限制
			if !rok || rval == nil {
				continue
			}
			// 存在对应的规则，判断是否允许下发play 的：1是禁止，2是允许
			if rval.Play == int64(locationGRPC.Status_Allow) {
				isSuccess = true
			} else { // 规则中有一项不满足，则该皮肤不下发
				isSuccess = false
				break
			}
		}
		if isSuccess {
			ableConfig = cfg
			break
		}
	}
	return
}

// ZlimitInfo2 获取
func (s *Service) ZlimitInfo2(ctx context.Context, policyGroupIDs []int64, ip string) (ipZoneID int64, groupAuths map[int64]*locationGRPC.GroupAuth, err error) {
	var (
		zlimitRly *locationGRPC.ZlimitInfoReply
		zVal      *locationGRPC.InfoComplete
		zok       bool
	)
	req := &locationGRPC.ZlimitInfoReq{
		Gids:  policyGroupIDs,
		Addrs: []string{ip},
	}
	if zlimitRly, err = s.locGRPC.ZlimitInfo2(ctx, req); err != nil {
		err = errors.Wrapf(err, "Dao: ZlimitInfo2 %s", ip)
		log.Error("Service: ZlimitInfo2 error %v", err)
		return
	}
	// IP服务异常，视作没有满足
	if zlimitRly == nil {
		err = errors.Wrapf(frontpageEcode.FrontPageLocationParseError, "%s zlimit结果为空", ip)
		log.Error("Service: ZlimitInfo2 error %v", err)
		return
	}
	// 获取ip所在的zone_id
	if zVal, zok = zlimitRly.Infos[ip]; !zok || zVal == nil || len(zVal.ZoneId) < 2 {
		// IP服务异常,未找到当前ip对应的zoneid
		err = errors.Wrapf(frontpageEcode.FrontPageLocationParseError, "%s 未找到对应IP策略", ip)
		log.Error("Service: ZlimitInfo2 error %v", err)
		return
	}
	// ZoneId 第0位全地区、第1位国家、第2位省份、第3位城市
	ipZoneID = zVal.ZoneId[1]
	return ipZoneID, zlimitRly.Policy, nil
}
