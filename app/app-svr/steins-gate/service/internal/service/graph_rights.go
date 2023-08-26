package service

import (
	"context"
	"go-gateway/app/app-svr/steins-gate/service/api"
	"go-gateway/app/app-svr/steins-gate/service/internal/model"
)

const (
	_interventionBuildIphone   = 8910
	_interventionBuildIpad     = 8900
	_interventioNBuildAndroid  = 5490000
	_interventioNBuildAndroidI = 3000000
	_interventionBuildIphoneI  = 64400200

	_featureDisabledBuildIphone   = 9040
	_featureDisabledBuildIphoneB  = 9050
	_featureDisabledBuildIpad     = 12300
	_featureDisabledBuildAndroid  = 5520000
	_featureDisabledBuildAndroidI = 3000000
	_featureDisabledBuildIphoneI  = 64400200
)

func hasFeatureDisabled(graph *api.GraphInfo) bool {
	if graph.NoTutorial > 0 {
		return true
	}
	if graph.NoBacktracking > 0 {
		return true
	}
	if graph.NoEvaluation > 0 {
		return true
	}
	if graph.GuestOverwriteRegionalVars <= 0 {
		return true
	}
	return false
}

// GraphRights 检查对应版本是否可以播放
func (s *Service) GraphRights(c context.Context, req *api.GraphRightsReq) (allowPlay bool, err error) {
	var graph *api.GraphInfo
	if graph, err = s.GraphInfo(c, req.Aid); err != nil {
		return
	}
	if hasFeatureDisabled(graph) {
		if featureDisabledVersionMatch(req) {
			s.promBusiness.Incr("GraphRights_FeatureDisabled_ValidBuild")
			return true, nil
		}
		s.promBusiness.Incr("GraphRights_FeatureDisabled_InvalidBuild")
		return false, nil
	}
	if !model.BuildRestrict(graph) { // 无限制返回true
		allowPlay = true
		s.promBusiness.Incr("GraphRights_NoIntervention")
		return
	}
	if (req.MobiApp == "iphone" && req.Device == "phone" && req.Build > _interventionBuildIphone) ||
		(req.MobiApp == "iphone" && req.Device == "pad" && req.Build > _interventionBuildIpad) ||
		(req.MobiApp == "android" && req.Build >= _interventioNBuildAndroid) ||
		(req.MobiApp == "android_i" && req.Build >= _interventioNBuildAndroidI) ||
		(req.MobiApp == "iphone_i" && req.Build >= _interventionBuildIphoneI) ||
		(req.MobiApp == "android_hd") {
		allowPlay = true
		s.promBusiness.Incr("GraphRights_Intervention_ValidBuild")
		return
	}
	s.promBusiness.Incr("GraphRights_Intervention_InvalidBuild")
	return // 返回默认值false
}

func featureDisabledVersionMatch(req *api.GraphRightsReq) bool {
	return (req.MobiApp == "iphone" && req.Build > _featureDisabledBuildIphone) || // 包括 ipad 粉版
		(req.MobiApp == "iphone_b" && req.Build > _featureDisabledBuildIphoneB) ||
		(req.MobiApp == "ipad" && req.Build > _featureDisabledBuildIpad) ||
		(req.MobiApp == "android" && req.Build >= _featureDisabledBuildAndroid) ||
		(req.MobiApp == "android_i" && req.Build >= _featureDisabledBuildAndroidI) ||
		(req.MobiApp == "iphone_i" && req.Build >= _featureDisabledBuildIphoneI) ||
		(req.MobiApp == "android_hd")
}
