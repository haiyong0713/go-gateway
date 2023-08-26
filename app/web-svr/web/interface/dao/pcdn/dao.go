package pcdn

import (
	"context"
	"go-common/library/ecode"
	"go-common/library/net/metadata"
	"go-gateway/app/web-svr/web/interface/conf"
	"go-gateway/app/web-svr/web/interface/model"
	"time"

	pcdnAccgrpc "git.bilibili.co/bapis/bapis-go/vas/pcdn/account/service"
	"git.bilibili.co/bapis/bapis-go/vas/pcdn/common"
	pcdnRewgrpc "git.bilibili.co/bapis/bapis-go/vas/pcdn/reward/service"
	pcdnVerifygrpc "git.bilibili.co/bapis/bapis-go/vas/pcdn/verify/service"

	"github.com/pkg/errors"
)

// Dao web PCDN dao.
type Dao struct {
	accClient    pcdnAccgrpc.PcdnAccountServiceClient // pcdn账号服务
	rewardClient pcdnRewgrpc.PcdnRewardServiceClient  // pcdn奖励服务
	verifyClient pcdnVerifygrpc.PcdnVerifyClient      // pcdn校验服务
	c            *conf.Config
}

const (
	_freeze_duration = 2 * 24 * 60 * 60
)

// New get pcdn dao
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c: c,
	}
	var err error
	if d.accClient, err = pcdnAccgrpc.NewClientPcdnAccountService(c.PcdnAccGRPC); err != nil {
		panic(err)
	}
	if d.rewardClient, err = pcdnRewgrpc.NewClientPcdnRewardService(c.PcdnRewardGRPC); err != nil {
		panic(err)
	}
	if d.verifyClient, err = pcdnVerifygrpc.NewClientPcdnVerify(c.PcdnVerifyGRPC); err != nil {
		panic(err)
	}
	return
}

func (d *Dao) JoinPCDN(c context.Context, mid int64) error {
	if _, err := d.accClient.JoinPcdn(c, &pcdnAccgrpc.JoinPcdnReq{
		Mid: mid,
	}); err != nil {
		return err
	}
	return nil
}

func (d *Dao) OperatePCDN(c context.Context, req *model.OperatePCDNReq) error {
	if _, err := d.accClient.OperatePcdn(c, &pcdnAccgrpc.OperatePcdnReq{
		Mid:     req.Mid,
		Operate: pcdnAccgrpc.PcdnOperateType(req.Operate),
		Level:   int64(req.Level),
	}); err != nil {
		return err
	}
	return nil
}

func (d *Dao) UserSettings(c context.Context, mid int64) (*model.PcdnUserSettingRep, error) {
	res, err := d.accClient.PcdnUserStatus(c, &pcdnAccgrpc.PcdnUserStatusReq{
		Mid: mid,
	})
	if err != nil {
		return nil, err
	}
	rs := &model.PcdnUserSettingRep{
		Level:   int(res.Level),
		Started: res.Started,
		Joined:  res.Joined,
	}
	curTime := time.Now().Unix()
	freezeTime := res.Mtime + _freeze_duration
	if res.Quit && freezeTime > curTime {
		rs.LeftTime = freezeTime - curTime
	}
	return rs, nil
}

func (d *Dao) UserAccountInfo(c context.Context, mid int64) (res *pcdnAccgrpc.UserAccountInfoResp, err error) {
	res, err = d.accClient.UserAccountInfo(c, &pcdnAccgrpc.UserAccountInfoReq{
		Mid: mid,
	})
	if err != nil {
		return nil, err
	}
	if len(res.Result) <= 0 {
		return nil, errors.Wrapf(ecode.NothingFound, "啥也没找到~")
	}

	return
}

func (d *Dao) Exchange(c context.Context, req *model.PcdnRewardExchangeReq) error {
	if _, err := d.rewardClient.Exchange(c, &pcdnRewgrpc.ExchangeReq{
		Mid:  req.Mid,
		Type: common.CurrencyType(req.Type),
		Num:  req.Num,
	}); err != nil {
		return err
	}
	return nil
}

func (d *Dao) Notification(c context.Context, mid int64) (*pcdnAccgrpc.UserNotificationResp, error) {
	return d.accClient.UserNotification(c, &pcdnAccgrpc.UserNotificationReq{
		Mid: mid,
	})
}

func (d *Dao) ReportV1(c context.Context, req *pcdnVerifygrpc.ReportGlobalInfo) error {
	ip := metadata.String(c, metadata.RemoteIP)
	req.Ip = ip
	if _, err := d.verifyClient.ReportV1(c, req); err != nil {
		return err
	}
	return nil
}

func (d *Dao) DigitalRewardInfo(c context.Context, mid int64) (*pcdnRewgrpc.DigitalRewardResp, error) {
	return d.rewardClient.DigitalReward(c, &pcdnRewgrpc.DigitalRewardReq{
		Mid: mid,
	})
}

func (d *Dao) Quit(c context.Context, mid int64) (*pcdnAccgrpc.QuitPcdnResp, error) {
	return d.accClient.QuitPcdn(c, &pcdnAccgrpc.QuitPcdnReq{
		Mid: mid,
	})
}
