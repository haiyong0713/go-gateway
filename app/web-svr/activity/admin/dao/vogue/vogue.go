package vogue

import (
	"context"
	"go-common/library/ecode"

	accApi "git.bilibili.co/bapis/bapis-go/account/service"
	silver "git.bilibili.co/bapis/bapis-go/silverbullet/service/silverbullet-proxy"
	"go-common/library/log"
)

const (
	_silverbulletStrategyName = "618_room_of_requirement"
)

// Infos get user info by mids.
func (d *Dao) UserInfos(c context.Context, mids []int64) (res map[int64]*accApi.Info, err error) {
	res = make(map[int64]*accApi.Info)
	if len(mids) == 0 {
		return
	}
	var arg = &accApi.MidsReq{
		Mids: mids,
	}
	var rpcRes *accApi.InfosReply
	if rpcRes, err = d.acc.Infos3(c, arg); err != nil {
		log.Error("d.acc.Infos3(%v) error(%v)", mids, err)
		//err = creErr.CreativeAccServiceErr
		err = ecode.RequestErr
	}
	if rpcRes != nil {
		res = rpcRes.Infos
	}
	log.Info("users info: %v", res)
	return
}

// user risk info
func (d *Dao) RiskInfo(c context.Context, mid int64) (hasRisk bool, msg string, err error) {
	// 风控
	var riskinfo *silver.RiskInfoReply
	hasRisk = false
	if riskinfo, err = d.sbc.RiskInfo(c, &silver.RiskInfoReq{
		StrategyName: []string{_silverbulletStrategyName},
		Mid:          mid,
	}); err != nil {
		log.Error("d.silverBulletClient(%v)", err)
	} else {
		if risk, ok := riskinfo.GetInfos()[_silverbulletStrategyName]; ok {
			if risk.Level == 0 {
				msg = "无"
			} else if risk.Level == 5 {
				msg = "黑名单"
				hasRisk = true
			} else if risk.Level > 0 && risk.Level < 5 {
				msg = "信用分低"
				hasRisk = true
			}
			return
		}
	}
	return
}
