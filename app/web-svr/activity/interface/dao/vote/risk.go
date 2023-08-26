package vote

import (
	"context"
	"encoding/json"
	"fmt"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/client"
	riskModel "go-gateway/app/web-svr/activity/interface/model/risk"
	model "go-gateway/app/web-svr/activity/interface/model/vote"
	"time"

	api "git.bilibili.co/bapis/bapis-go/silverbullet/gaia/interface"
)

// RuleCheck .
func (d *Dao) RuleCheck(ctx context.Context, riskScene string, params *riskModel.VoteNew) (res bool, err error) {
	if client.SilverbulletClient == nil {
		err = fmt.Errorf("SilverbulletClient not init")
		return
	}
	var (
		eventCtx []byte
	)
	if eventCtx, err = json.Marshal(params); err != nil {
		log.Errorc(ctx, "RuleCheck json.Marshal err(%+v) v(%+v)", err, params)
		return
	}
	var scene string
	switch riskScene {
	case model.RiskControlRuleGeneric:
		scene = riskModel.ActionVoteNew
	default:
		scene = riskScene
	}
	req := &api.RuleCheckReq{
		Scene:    scene,
		EventCtx: string(eventCtx),
		EventTs:  time.Now().Unix(),
	}
	rep, err := client.SilverbulletClient.RuleCheck(ctx, req)
	if err != nil {
		log.Errorc(ctx, "RuleCheck d.gaiaClient.RuleCheck err(%+v)", err)
		return false, err
	}
	riskRes := rep.GetDecisions()
	if len(riskRes) > 0 && riskRes[0] == "reject" {
		log.Warnc(ctx, "RuleCheck req(%+v)", req)
		res = true
	}
	return
}

func (d *Dao) getRiskType(dataSourceType string) string {
	switch dataSourceType {
	case model.DSTypeOperVideo:
		return "avid"
	case model.DSTypeOperPic:
		return "picture"
	case model.DSTypeOperUp:
		return "people"
	}
	return "other"
}
