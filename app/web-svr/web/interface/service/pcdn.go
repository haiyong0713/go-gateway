package service

import (
	"context"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/web-svr/web/interface/model"
	"strconv"

	bgroupgrpc "git.bilibili.co/bapis/bapis-go/platform/service/bgroup/v2"
	pcdnAccgrpc "git.bilibili.co/bapis/bapis-go/vas/pcdn/account/service"
	pcdnRewgrpc "git.bilibili.co/bapis/bapis-go/vas/pcdn/reward/service"
	pcdnVerifygrpc "git.bilibili.co/bapis/bapis-go/vas/pcdn/verify/service"
)

const (
	_errMsg             = "服务开小差了，请稍后重试~"
	_pcdnQuitFreezeCode = 201004
	_pcdnQuitFreezeMsg  = "退出计划48小时后方可再次加入～"
)

// 加入流量计划
func (s *Service) JoinPCDN(ctx context.Context, mid int64) error {
	if err := s.pcdnDao.JoinPCDN(ctx, mid); err != nil {
		log.Error("【@JoinPCDN】Join Pcdn mid: (%d), err: (%v)", mid, err)
		return ecode.Errorf(ecode.Code(ecode.Cause(err).Code()), "加入流量激励计划出错！")
	}
	return nil
}

// 开启/关闭/设置贡献档位
func (s *Service) OperatePCDN(ctx context.Context, req *model.OperatePCDNReq) error {
	if err := s.pcdnDao.OperatePCDN(ctx, req); err != nil {
		log.Error("【@StartPCDN】Start Pcdn req: (%v), err: (%v)", req, err)
		return ecode.Error(ecode.ServerErr, _errMsg)
	}
	return nil
}

// 查询用户设置
func (s *Service) UserSettings(ctx context.Context, mid int64) (res *model.PcdnUserSettingRep, err error) {
	if res, err = s.pcdnDao.UserSettings(ctx, mid); err != nil {
		log.Error("【@UserSettings】error,  mid: (%d), err: (%v)", mid, err)
		return nil, ecode.Errorf(ecode.Code(ecode.Cause(err).Code()), "获取用户设置失败～, %s", err.Error())
	}
	return res, nil
}

// 查询用户资产信息
func (s *Service) UserAccountInfo(ctx context.Context, mid int64) (res *pcdnAccgrpc.UserAccountInfoResp, err error) {
	if res, err = s.pcdnDao.UserAccountInfo(ctx, mid); err != nil {
		log.Error("【@UserAccountInfo】error,  mid: (%d), err: (%v)", mid, err)
		return nil, ecode.Errorf(ecode.Code(ecode.Cause(err).Code()), "获取用户信息失败～")
	}
	return res, nil
}

// 兑换资产
func (s *Service) Exchange(ctx context.Context, req *model.PcdnRewardExchangeReq) (err error) {
	if err = s.pcdnDao.Exchange(ctx, req); err != nil {
		log.Error("【@Exchange】error,  param: (%v), err: (%v)", *req, err)
		return ecode.Errorf(ecode.Code(ecode.Cause(err).Code()), "兑换出现了点问题哦～, %s", err.Error())
	}
	return
}

// 聚合接口
func (s *Service) PcdnV1(ctx context.Context, req *model.PcdnV1Req) (res *model.PcdnV1Rep, err error) {
	res = &model.PcdnV1Rep{}
	// 人群包
	memReq := &bgroupgrpc.MemberInReq{
		Member:    strconv.FormatInt(req.Mid, 10),
		Dimension: bgroupgrpc.Mid,
		Groups:    s.getPcdnBgroupItems(),
	}
	if bgroupRep, bgroupErr := s.bgroupGRPC.MemberIn(ctx, memReq); bgroupErr == nil && len(bgroupRep.Results) > 0 {
		for _, item := range bgroupRep.Results {
			if item.Name == s.c.Rule.PcdnGroup.Top {
				// 顶导
				res.IsInTopList = item.In
				continue
			}
			if item.Name == s.c.Rule.PcdnGroup.Pop {
				// 弹窗
				res.IsInPopList = item.In
				// 命中弹窗一定命中顶导
				if item.In {
					res.IsInTopList = true
				}
			}
		}
	}

	// 用户设置查询
	if res.IsInTopList || res.IsInPopList {
		res.UserSettings, _ = s.UserSettings(ctx, req.Mid)
	}

	// 植入fawkes的config和ff数据
	if fawekesVersion, ok := s.fawkesVersionCache[req.FawkesEnv]; ok {
		if fv, ok := fawekesVersion[req.FawkesAppKey]; ok {
			res.Fawkes = &model.Fawkes{
				ConfigVersion: fv.Config,
				FFVersion:     fv.FF,
			}
		}
	}
	return res, nil
}

// 数据校验
func (s *Service) ReportV1(c context.Context, req *pcdnVerifygrpc.ReportGlobalInfo) error {
	if err := s.pcdnDao.ReportV1(c, req); err != nil {
		log.Error("【@ReportV1】error: (%v)", err)
		return err
	}
	return nil
}

// 用户通知todo
func (s *Service) Notification(c context.Context, mid int64) (res *pcdnAccgrpc.UserNotificationResp, err error) {
	if res, err = s.pcdnDao.Notification(c, mid); err != nil {
		log.Error("【@Notification】error,  mid: (%d), err: (%v)", mid, err)
		return nil, ecode.Error(ecode.ServerErr, _errMsg)
	}
	return
}

func (s *Service) getPcdnBgroupItems() (rs []*bgroupgrpc.MemberInReq_MemberInReqSingle) {
	topItem := &bgroupgrpc.MemberInReq_MemberInReqSingle{
		Name:     s.c.Rule.PcdnGroup.Top,
		Business: s.c.Rule.PcdnGroup.Business,
	}
	popItem := &bgroupgrpc.MemberInReq_MemberInReqSingle{
		Name:     s.c.Rule.PcdnGroup.Pop,
		Business: s.c.Rule.PcdnGroup.Business,
	}
	rs = append(rs, topItem, popItem)
	return
}

func (s *Service) DigitialCollection(c context.Context, mid int64) (res *pcdnRewgrpc.DigitalRewardResp, err error) {
	if res, err = s.pcdnDao.DigitalRewardInfo(c, mid); err != nil {
		log.Error("【@DigitialCollection】error,  mid: (%d), err: (%v)", mid, err)
		return nil, ecode.Errorf(ecode.Code(ecode.Cause(err).Code()), _errMsg)
	}
	if len(res.DigitalInfo) <= 0 {
		log.Warn("【@DigitialCollection】miss DigitalInfo,  mid: (%d)", mid)
	}
	return
}

func (s *Service) Quit(c context.Context, mid int64) (err error) {
	if _, err = s.pcdnDao.Quit(c, mid); err != nil {
		log.Error("【@Quit】error,  mid: (%d), err is: (%v)", mid, err)
		code := ecode.Code(ecode.Cause(err).Code())
		msg := _errMsg
		if code == _pcdnQuitFreezeCode {
			msg = _pcdnQuitFreezeMsg
		}
		return ecode.Errorf(code, msg)
	}
	return
}

// 统一聚合页面接口
func (s *Service) PcdnPages(c context.Context, mid int64) (res *model.PcdnPagesRep) {
	res = &model.PcdnPagesRep{}
	group := errgroup.WithContext(c)
	group.Go(func(ctx context.Context) error {
		info, _ := s.Notification(ctx, mid)
		if info != nil && len(info.Result) > 0 {
			res.Notification = info.Result
		}
		return nil
	})
	group.Go(func(ctx context.Context) error {
		res.DigitalReward, _ = s.DigitialCollection(ctx, mid)
		return nil
	})
	group.Go(func(ctx context.Context) error {
		userInfo, err := s.UserAccountInfo(ctx, mid)
		if err != nil {
			return nil
		}
		if len(userInfo.Result) > 0 {
			res.AccountInfo = userInfo.Result
		}
		return nil
	})
	if err := group.Wait(); err != nil {
		log.Error("%+v", err)
	}
	return
}
