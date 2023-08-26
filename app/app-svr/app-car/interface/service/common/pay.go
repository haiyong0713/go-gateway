package common

import (
	"context"
	"fmt"
	"strconv"

	"go-gateway/app/app-svr/app-car/interface/model"
	bangumimdl "go-gateway/app/app-svr/app-car/interface/model/bangumi"
	commonmdl "go-gateway/app/app-svr/app-car/interface/model/common"

	"go-common/library/log"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
)

const (
	_car   = 1
	_sound = 2
)

func (s *Service) PayInfo(c context.Context, req *commonmdl.PayInfoReq, mid int64, buvid string) (res *commonmdl.PayInfoResp) {
	switch req.Ptype {
	case 1:
		var url string
		switch req.DeviceType {
		case _car:
			url = fmt.Sprintf("bilibili://user_center/vip/buy/152?appSubId=%v", req.Channel)
		case _sound:
			url = fmt.Sprintf("bilibili://user_center/vip/buy/153?appSubId=%v", req.Channel)
		default:
			return
		}
		res = &commonmdl.PayInfoResp{
			Title: "购买大会员",
			Desc:  "需确认扫码手机端登录用户账号与当前账号一致",
			Url:   url,
		}
	case 2: // nolint:gomnd
		res = &commonmdl.PayInfoResp{
			Title: "购买内容",
			Desc:  "需确认扫码手机端登录用户账号与当前账号一致",
			Url:   model.FillURI(model.GotoWebPGC, 0, 0, strconv.FormatInt(req.Epid, 10), nil),
		}
	default:
		log.Warn("PayInfo(%+v) invalid type", req)
	}
	return
}

func (s *Service) PayState(c context.Context, req *commonmdl.PayStateReq, mid int64, buvid, cookie, referer string) (res *commonmdl.PayStateResp, err error) {
	switch req.Ptype {
	case 1:
		var useProfile *accountgrpc.Profile
		if useProfile, err = s.accountDao.Profile3(c, mid); err != nil {
			log.Error("PayState() Profile3(%v) error(%v)", mid, err)
			return
		}
		res = new(commonmdl.PayStateResp)
		if useProfile != nil && useProfile.Vip.Status == 1 {
			res.IsSuccess = true
		}
	case 2: // nolint:gomnd
		var seasonView *bangumimdl.View
		if seasonView, err = s.bangumiDao.View(c, mid, req.SeasonId, req.AccessKey, cookie, req.MobiApp, req.Platform, buvid, referer, req.Build); err != nil {
			log.Error("PayState() View(%v) error(%v)", mid, err)
			return
		}
		res = new(commonmdl.PayStateResp)
		if seasonView != nil && seasonView.UserStatus != nil && seasonView.UserStatus.Pay == 1 {
			res.IsSuccess = true
		}
	default:
		log.Warn("PayState(%+v) invalid type", req)
	}
	return
}
