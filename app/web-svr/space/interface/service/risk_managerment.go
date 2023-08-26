package service

import (
	"context"
	"encoding/json"
	gaiamdl "git.bilibili.co/bapis/bapis-go/silverbullet/gaia/interface"
	"go-common/library/log"
	"go-gateway/app/web-svr/space/interface/model"
	"time"
)

// RiskVerifyAndManager 极验 + 风控
func (s *Service) RiskVerifyAndManager(ctx context.Context, riskParams *model.RiskManagement) (res *model.RiskResult) {
	res = &model.RiskResult{
		IsRisk:      false,
		GaiaResType: model.GaiaResponseType_Default,
	}
	if riskParams.Token != "" {
		vtRes, vtErr := s.gaiaGRPC.VerifyToken(ctx, &gaiamdl.VerifyTokenReq{Scene: riskParams.Scene, Mid: riskParams.Mid, Token: riskParams.Token})
		if vtErr == nil {
			log.Error("VerifyToken sence:%s mid:%d token:%s", riskParams.Scene, riskParams.Mid, riskParams.Token)
		}
		if vtRes != nil && vtRes.IsValid == 0 {
			res.GaiaResType = model.GaiaResponseType_TokenInvalid
			return res
		}
	} else {
		riskRes, err := s.RiskManager(ctx, riskParams)
		if err == nil && riskRes != nil {
			if len(riskRes.Decisions) > 0 && IsContained(riskRes.Decisions[0], s.c.RiskManagement.RiskDecisions) {
				res.GaiaResType = model.GaiaResponseType_NeedFECheck
				res.GaiaData = riskRes
				return res
			}
		}
	}
	return nil
}

// 风控
func (s *Service) RiskManager(ctx context.Context, riskParams *model.RiskManagement) (*gaiamdl.RuleCheckReply, error) {
	now := time.Now()
	ctxStr, err := json.Marshal(riskParams)
	if err != nil {
		log.Error("space scene: %+s ,json Marshal %+v error:%+v", riskParams.Api, riskParams, err)
		return nil, err
	}
	checkReply, err := s.gaiaGRPC.RuleCheck(ctx, &gaiamdl.RuleCheckReq{
		Scene:    riskParams.Action,
		EventCtx: string(ctxStr),
		EventTs:  now.Unix(),
	})
	if err != nil {
		log.Error("space scene: %+s , RuleCheck mid:%d error:%v", riskParams.Api, riskParams.Mid, err)
		return nil, err
	}
	return checkReply, nil
}

func IsContained(target string, array []string) bool {
	for _, v := range array {
		if target == v {
			return true
		}
	}
	return false
}
