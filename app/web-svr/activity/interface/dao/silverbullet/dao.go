package silverbullet

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go-common/library/log"
	"go-gateway/app/web-svr/activity/ecode"

	"git.bilibili.co/bapis/bapis-go/silverbullet/gaia/interface"
	"go-gateway/app/web-svr/activity/interface/conf"
)

// Dao is gaia dao
type Dao struct {
	// grpc
	gaiaClient api.GaiaClient
}

// New is
func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.gaiaClient, err = api.NewClient(c.GaiaClient); err != nil {
		panic(fmt.Sprintf("Gaia Client not found err(%v)", err))
	}
	return
}

// RuleCheck .
func (d *Dao) RuleCheck(ctx context.Context, scene string, otherEventCtx interface{}) (res bool, err error) {
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
	rep, err := d.gaiaClient.RuleCheck(ctx, req)
	if err != nil {
		log.Errorc(ctx, "RuleCheck d.gaiaClient.RuleCheck err(%+v)", err)
		return false, err
	}
	riskRes := rep.GetDecisions()
	if len(riskRes) > 0 && riskRes[0] == "reject" {
		log.Warnc(ctx, "RuleCheck req(%+v)", req)
		return true, nil
	}
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
	rep, err := d.gaiaClient.RuleCheck(ctx, req)
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
