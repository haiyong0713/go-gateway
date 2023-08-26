package service

import (
	"context"
	"strings"

	"go-gateway/app/app-svr/feature/service/api"
	degrademdl "go-gateway/app/app-svr/feature/service/model/degrade"
)

const (
	_autoLaunchDefault = 0
	_autoLaunchHit     = 1

	_logLevelDefault         = 4            // forbidden
	_logLevelForbidden       = 0            // forbidden
	_coverEnlargeDefault     = float32(1.0) // tv极速版默认折损
	_coverEnlargeForbidden   = float32(0.7) // tv极速版禁止折损
	_defalultSysVersionNone1 = 0
	_defalultSysVersionNone2 = -1
)

func (s *Service) FeatureTVSwitch(c context.Context, req *api.FeatureTVSwitchReq) (*api.FeatureTVSwitchReply, error) {
	var (
		res = new(api.FeatureTVSwitchReply)
		err error
	)
	// 机型维度的判断
	res.Id, res.IsHit = s.hitConfig(req)
	// feature维度对应的结果
	for _, feature := range req.Feature {
		tmpRe := new(api.FeatureTVSwitchItem)
		switch feature {
		case _autoLaunch:
			tmpRe.DisplayType = api.DisplayType_DisplayTypeChannel
			tmpReChannel := &api.ChannelFeatureReply{AutoLaunch: _autoLaunchDefault}
			if rules, ok := s.tvSwitchCache[feature]; ok {
				tmpRe.IsHit = true
				var hit bool
				if hit, tmpRe.HitId = s.featureTVSwitchRule(rules, req); hit {
					tmpReChannel.AutoLaunch = _autoLaunchHit
				}
			}
			tmpRe.Item = &api.FeatureTVSwitchItem_Channel{
				Channel: tmpReChannel,
			}
		case _featureCover:
			tmpRe.DisplayType = api.DisplayType_DisplayTypeDegrade
			tmpReDegrade := &api.FeatureDegradeItem{Enlarge: _coverEnlargeDefault}
			if rules, ok := s.tvSwitchCache[feature]; ok {
				tmpRe.IsHit = true
				var hit bool
				if hit, tmpRe.HitId = s.featureTVSwitchRule(rules, req); hit {
					tmpReDegrade.Enlarge = _coverEnlargeForbidden
				}
			}
			tmpRe.Item = &api.FeatureTVSwitchItem_Degrade{
				Degrade: tmpReDegrade,
			}
		case _videoShot:
			tmpRe.DisplayType = api.DisplayType_DisplayTypeDegrade
			tmpReDegrade := &api.FeatureDegradeItem{IsDegrade: false}
			if rules, ok := s.tvSwitchCache[feature]; ok {
				tmpRe.IsHit = true
				var hit bool
				if hit, tmpRe.HitId = s.featureTVSwitchRule(rules, req); hit {
					tmpReDegrade.IsDegrade = true
				}
			}
			tmpRe.Item = &api.FeatureTVSwitchItem_Degrade{
				Degrade: tmpReDegrade,
			}
		default:
			tmpRe.DisplayType = api.DisplayType_DisplayTypeDegrade
			tmpReDegrade := &api.FeatureDegradeItem{IsDegrade: false, LogLevel: _logLevelDefault} // 极速版默认黑名单
			if rules, ok := s.tvSwitchCache[feature]; ok {
				tmpRe.IsHit = true
				var hit bool
				if hit, tmpRe.HitId = s.featureTVSwitchRule(rules, req); hit {
					tmpReDegrade.IsDegrade = true
					tmpReDegrade.LogLevel = _logLevelForbidden
				}
			}
			tmpRe.Item = &api.FeatureTVSwitchItem_Degrade{
				Degrade: tmpReDegrade,
			}
		}
		if res.Switch == nil {
			res.Switch = make(map[string]*api.FeatureTVSwitchItem)
		}
		res.Switch[feature] = tmpRe
	}
	return res, err
}

func (s *Service) featureTVSwitchRule(rules []*degrademdl.TvSwitch, req *api.FeatureTVSwitchReq) (bool, int64) {
	for _, rule := range rules {
		var hitBrand, hitModel, hitChil, hitSysVerson = true, true, true, true
		// 配置非空且未命中 = false
		if rule.Chil != "" {
			var tmpHitChil bool
			for _, chil := range strings.Split(rule.Chil, ",") {
				if chil == req.Channel {
					tmpHitChil = true
					break
				}
			}
			hitChil = tmpHitChil
		}
		if rule.Brand != "" {
			var tmpHitBrand bool
			for _, brand := range strings.Split(rule.Brand, ",") {
				if brand == req.Brand {
					tmpHitBrand = true
					break
				}
			}
			hitBrand = tmpHitBrand
		}
		if rule.Model != "" {
			var tmpHitModel bool
			for _, mod := range strings.Split(rule.Model, ",") {
				if mod == req.Model {
					tmpHitModel = true
					break
				}
			}
			hitModel = tmpHitModel
		}
		if rule.SysVersion != nil {
			var tmpHitSysVerson bool
			if (req.SysVer >= rule.SysVersion.Start || rule.SysVersion.Start == _defalultSysVersionNone1 || rule.SysVersion.Start == _defalultSysVersionNone2) && (req.SysVer <= rule.SysVersion.End || rule.SysVersion.End == _defalultSysVersionNone1 || rule.SysVersion.End == _defalultSysVersionNone2) {
				tmpHitSysVerson = true
			}
			hitSysVerson = tmpHitSysVerson
		}
		if hitBrand && hitModel && hitChil && hitSysVerson {
			return true, rule.ID
		}
	}
	return false, 0
}

func (s *Service) hitConfig(req *api.FeatureTVSwitchReq) (int64, bool) {
	for _, v := range s.tvSwitchKeysCache {
		if v == nil {
			continue
		}
		if v.Brand != "" && v.Brand != req.Brand {
			continue
		}
		if v.Chil != "" && v.Chil != req.Channel {
			continue
		}
		if v.Model != "" && v.Model != req.Model {
			continue
		}
		if v.SysVersion != nil {
			if v.SysVersion.Start != -1 && v.SysVersion.Start > req.SysVer {
				continue
			}
			if v.SysVersion.End != -1 && v.SysVersion.End < req.SysVer {
				continue
			}
		}
		return v.ID, true
	}
	return 0, false
}
