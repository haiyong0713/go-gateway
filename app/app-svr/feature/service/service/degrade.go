package service

import (
	"context"
	"strconv"

	"go-common/library/log"

	"go-gateway/app/app-svr/feature/service/api"
	degrademdl "go-gateway/app/app-svr/feature/service/model/degrade"
)

const (
	_decodeTypeNone = 0

	_featureCover = "cover"
	_videoShot    = "video_shot" // 进度缩略图
	_autoLaunch   = "auto_launch"
)

func (s *Service) FeatureDegrades(c context.Context, req *api.FeatureDegradesReq) (res *api.FeatureDegradesReply, err error) {
	featureDegrades := make(map[string]*api.FeatureDegradeItem)
	for _, feature := range req.Feature {
		featureDegrades[feature] = s.featureDegrade(c, feature, req)
	}
	return &api.FeatureDegradesReply{Items: featureDegrades}, nil
}

/*
《白名单》
最少命中一个就显示。
全部不命中（包括未配置）不显示

《黑名单》
最少命中一个不显示
全部未命中（包括未配置）展示
*/
func (s *Service) featureDegrade(c context.Context, feature string, req *api.FeatureDegradesReq) *api.FeatureDegradeItem {
	if feature == _featureCover {
		return s.getCoverDegrade(c, req)
	}
	if feature == _videoShot {
		return s.getVideoShotDegrade(c, req)
	}
	fea := s.displayLimit[feature]
	hit1, logLevel1 := s.hitDegrade(fea, req.Channel, "chid", req.Build, req.SysVer)
	hit2, logLevel2 := s.hitDegrade(fea, req.Brand, "brand", req.Build, req.SysVer)
	hit3, logLevel3 := s.hitDegrade(fea, req.Model, "model", req.Build, req.SysVer)
	a, ok := s.c.Degrade.FeatureQnMap[feature]
	if a == 1 {
		if hit1 || hit2 || hit3 {
			var ll int32
			if hit1 {
				ll = logLevel1
			} else if hit2 {
				ll = logLevel2
			} else if hit3 {
				ll = logLevel3
			}
			return &api.FeatureDegradeItem{IsDegrade: false, LogLevel: ll}
		}
		return &api.FeatureDegradeItem{IsDegrade: true, LogLevel: s.c.Degrade.Cfg.DefaultLogLevel}
	}
	if !ok || a == 2 {
		if hit1 || hit2 || hit3 {
			var ll int32
			if hit1 {
				ll = logLevel1
			} else if hit2 {
				ll = logLevel2
			} else if hit3 {
				ll = logLevel3
			}
			return &api.FeatureDegradeItem{IsDegrade: true, LogLevel: ll}
		}
		return &api.FeatureDegradeItem{IsDegrade: false, LogLevel: s.c.Degrade.Cfg.DefaultLogLevel}
	}
	log.Error("featureDegrade unknow direction=%d", a)
	return &api.FeatureDegradeItem{IsDegrade: false, LogLevel: s.c.Degrade.Cfg.DefaultLogLevel}
}

func (s *Service) getCoverDegrade(_ context.Context, req *api.FeatureDegradesReq) *api.FeatureDegradeItem {
	for _, l := range s.coverLimit {
		hitMemory := func() bool {
			if req.Memory >= l.Rules.StoreGte && (req.Memory <= l.Rules.StoreLte || l.Rules.StoreLte == degrademdl.RuleNone) {
				return true
			}
			return false
		}()
		hitSysver := func() bool {
			if req.SysVer >= l.Rules.SysGte && (req.SysVer <= l.Rules.SysLte || l.Rules.SysLte == degrademdl.RuleNone) {
				return true
			}
			return false
		}()
		if hitMemory && hitSysver {
			return &api.FeatureDegradeItem{Enlarge: l.Rules.Enlarge}
		}
	}
	return &api.FeatureDegradeItem{Enlarge: s.c.Degrade.Cfg.DefaultEnlarge}
}

/*
进度缩略图降级.

只要命中就不展示
未命中，或者未配置，都展示
*/
func (s *Service) getVideoShotDegrade(_ context.Context, req *api.FeatureDegradesReq) *api.FeatureDegradeItem {
	for _, l := range s.videoShotLimit {
		hitMemory := func() bool {
			if req.Memory >= l.Rules.StoreGte && (req.Memory <= l.Rules.StoreLte || l.Rules.StoreLte == degrademdl.RuleNone) {
				return true
			}
			return false
		}()
		hitSysver := func() bool {
			if req.SysVer >= l.Rules.SysGte && (req.SysVer <= l.Rules.SysLte || l.Rules.SysLte == degrademdl.RuleNone) {
				return true
			}
			return false
		}()
		if hitMemory && hitSysver {
			return &api.FeatureDegradeItem{IsDegrade: true}
		}
	}
	return &api.FeatureDegradeItem{}
}

/*
【白名单】
只要命中一个配置，就展示
3个都不命中，或者都未配置，则不展示
【黑名单】
只要命中一个就不展示
都未命中，或者都未配置，则展示
*/
func (s *Service) hitDegrade(rules map[string]map[string]*degrademdl.Range, req string, limitTy string, build int64, sysVer int64) (bool, int32) {
	if len(rules) == 0 {
		return false, degrademdl.DefLogLevel
	}
	if req == "" {
		return false, degrademdl.DefLogLevel
	}
	if _, ok := rules[limitTy]; !ok {
		return false, degrademdl.DefLogLevel
	}
	limit, ok := rules[limitTy][req]
	if !ok {
		return false, degrademdl.DefLogLevel
	}
	hit := degrademdl.InIntervals(limit.BuildRange, build) && degrademdl.InIntervals(limit.SysVerRange, sysVer)
	return hit, limit.LogLevel
}

// feature后台3期：qn限制
// 基于判断display_type是否为数字，和Direction是否为1或2，来决定一条规则是否为清晰度tab的规则
func (s *Service) feature3QnLimits(res map[string][]*degrademdl.DisplayLimitRes) {
	displayLimit := make(map[string]map[string]map[string]*degrademdl.Range)
	for k, v := range res {
		if len(v) <= 0 {
			continue
		}

		_, err := strconv.ParseInt(k, 10, 64)
		if err != nil {
			continue
		}

		limits := v[0]
		if limits.Direction != 1 && limits.Direction != 2 {
			continue
		}

		displayLimit[k] = degrademdl.ToMap(v)
	}
	s.displayQnLimit = displayLimit
}

func (s *Service) ChannelFeature(c context.Context, req *api.ChannelFeatureReq) (*api.ChannelFeatureReply, error) {
	decodeType := func() int64 {
		fea, ok := s.chanFeature[req.Channel]
		if !ok || fea.DecodeType == _decodeTypeNone {
			// cms后台无该渠道配置情况下
			for _, def := range s.c.Degrade.Cfg.ChDefaults { // 对特定渠道配置的默认DecodeType
				if def.Code == req.Channel {
					return def.DecodeType
				}
			}
			return s.c.Degrade.Cfg.DefaultDecode
		}
		return fea.DecodeType
	}()
	autoLaunch := func() int32 {
		fea, ok := s.chanFeature[req.Channel]
		if ok {
			return fea.AutoLaunch
		}
		return s.c.Degrade.Cfg.DefaultAutoLaunch
	}()
	return &api.ChannelFeatureReply{DecodeType: decodeType, AutoLaunch: autoLaunch}, nil
}
