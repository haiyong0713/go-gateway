package cards

import (
	"context"
	"encoding/json"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/client"

	riskmdl "go-gateway/app/web-svr/activity/interface/model/risk"

	silverbulletapi "git.bilibili.co/bapis/bapis-go/silverbullet/gaia/interface"
	"github.com/pkg/errors"
)

// risk 风险用户
func (s *Service) risk(c context.Context, mid int64, risk interface{}, esTime int64) (riskReply *silverbulletapi.RuleCheckReply, err error) {

	var bs []byte
	bs, _ = json.Marshal(risk)
	riskReply, err = client.SilverbulletClient.RuleCheck(c, &silverbulletapi.RuleCheckReq{
		Scene:    "common_activity_collect",
		EventCtx: string(bs),
		EventTs:  esTime,
	})
	if err != nil {
		err = errors.Wrapf(err, "client.silverbulletClient.RuleCheck %d", mid)
		log.Errorc(c, "client.silverbulletClient.RuleCheck(%d) error(%v)", mid, err)
		return nil, nil
	}
	return riskReply, nil
}

// riskCheck 是否风险用户
func (s *Service) riskCheck(c context.Context, mid int64, riskReply *silverbulletapi.RuleCheckReply) (err error) {
	if riskReply == nil || riskReply.Decisions == nil {
		log.Errorc(c, "client.silverbulletClient.RuleCheck(%d) riskReply(%v)", mid, riskReply)
		return nil
	}
	// 风险命中判断
	if riskReply == nil || riskReply.Decisions == nil || len(riskReply.Decisions) == 0 {
		return nil
	}
	riskMap := riskReply.Decisions[0]
	if riskMap == riskmdl.Reject {
		return ecode.SpringFestivalRiskMemberErr
	}
	return nil
}
