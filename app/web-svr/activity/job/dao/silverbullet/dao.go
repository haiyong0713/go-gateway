package silverbullet

import (
	"context"
	"encoding/json"
	"time"

	"go-common/library/log"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/job/client"

	"git.bilibili.co/bapis/bapis-go/silverbullet/gaia/interface"
	"go-gateway/app/web-svr/activity/job/conf"
)

// Dao is gaia dao
type Dao struct {
}

// New is
func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	return
}

// RuleCheckCommon .
func (d *Dao) RuleCheckCommon(ctx context.Context, scene string, otherEventCtx interface{}) (res bool, err error) {
	var (
		eventCtx []byte
	)
	if eventCtx, err = json.Marshal(otherEventCtx); err != nil {
		log.Errorc(ctx, "RuleCheck json.Marshal err(%+v) v(%+v)", err, otherEventCtx)
		return
	}
	req := &api.RuleCheckReq{
		Scene:    scene,
		EventCtx: string(eventCtx),
		EventTs:  time.Now().Unix(),
	}
	rep, err := client.GaiaClient.RuleCheck(ctx, req)
	if err != nil {
		log.Errorc(ctx, "RuleCheck d.gaiaClient.RuleCheck err(%+v)", err)
		return false, err
	}
	riskRes := rep.GetDecisions()
	if len(riskRes) > 0 {
		switch riskRes[0] {
		case "reject":
			err = ecode.ActivityRiskRejectErr
		case "overseas":
			err = ecode.ActivityRiskOverseaErr
		case "no_bind_tel":
			err = ecode.ActivityRiskTelErr
		default:
			return
		}
		return true, err
	}
	return
}
