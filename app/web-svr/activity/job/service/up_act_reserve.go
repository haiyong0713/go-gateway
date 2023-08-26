package service

import (
	"context"
	tunnelCommon "git.bilibili.co/bapis/bapis-go/platform/common/tunnel"
	tunnelV2Mdl "git.bilibili.co/bapis/bapis-go/platform/service/tunnel/v2"
	"github.com/pkg/errors"
	"go-common/library/log"
	"go-common/library/net/netutil"
	"go-common/library/net/trace"
	"go-common/library/retry"
	"go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/job/client"
	"go-gateway/app/web-svr/activity/job/model/like"
	mdlmail "go-gateway/app/web-svr/activity/job/model/mail"
	"strconv"
	"strings"
	"time"
)

func (s *Service) upActReserveLiveStateExpire() {
	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if _, err := client.ActivityClient.UpActReserveLiveStateExpire(context.Background(), &api.UpActReserveLiveStateExpireReq{}); err != nil {
				log.Error("s.UpActReserveLiveStateExpire error(%v)", err)
				continue
			}
		}
	}
}

func (s *Service) upActReservePushVerify() {
	curTime := time.Now()
	// 过期十四天
	fourteenRange := [2]string{curTime.AddDate(0, 0, -15).Format("2006-01-02 15:04:05"), curTime.AddDate(0, 0, -14).Format("2006-01-02 15:04:05")}
	// 过期三十天
	thirtyRange := [2]string{curTime.AddDate(0, 0, -31).Format("2006-01-02 15:04:05"), curTime.AddDate(0, 0, -30).Format("2006-01-02 15:04:05")}
	ctx := trace.SimpleServerTrace(context.Background(), "upActReservePushVerify")
	base := &mdlmail.Base{
		Host:    s.c.Mail.Host,
		Port:    s.c.Mail.Port,
		Address: s.c.Mail.Address,
		Pwd:     s.c.Mail.Pwd,
		Name:    s.c.Mail.Name,
	}

	var err error
	err = s.mailFileConfig(ctx, base, s.verifyCardBuildReceivers(), []*mdlmail.Address{}, []*mdlmail.Address{}, "稿件预约催核销私信流程开始", mdlmail.TypeTextHTML, nil)
	if err != nil {
		log.Error("s.mailFile: error(%v)", err)
	}

	err = s.RegisterPushVerifyCard(ctx, like.NotifyMessageTypePushUpVerify14, "提醒up主核销预约-14天时")
	if err != nil {
		log.Errorc(ctx, "RegisterPushVerifyCard err(%+v)", err)
		return
	}
	err = s.RegisterPushVerifyCard(ctx, like.NotifyMessageTypePushUpVerify30, "提醒up主核销预约-30天时")
	if err != nil {
		log.Errorc(ctx, "RegisterPushVerifyCard err(%+v)", err)
		return
	}

	if err = s.BuildCardAndPushDataBus(ctx, fourteenRange[0], fourteenRange[1], like.NotifyMessageTypePushUpVerify14); err != nil {
		log.Errorc(ctx, like.UpActReserveRelationPushVerifyCard+"s.BuildCardAndPushDataBus range: %v, err:%v", fourteenRange, err)
	}

	if err = s.BuildCardAndPushDataBus(ctx, thirtyRange[0], thirtyRange[1], like.NotifyMessageTypePushUpVerify30); err != nil {
		log.Errorc(ctx, like.UpActReserveRelationPushVerifyCard+"s.BuildCardAndPushDataBus range: %v, err:%v", thirtyRange, err)
	}

	err = s.mailFileConfig(ctx, base, s.verifyCardBuildReceivers(), []*mdlmail.Address{}, []*mdlmail.Address{}, "稿件预约催核销私信流程结束", mdlmail.TypeTextHTML, nil)
	if err != nil {
		log.Error("s.mailFile error(%v)", err)
	}
}

func (s *Service) verifyCardBuildReceivers() []*mdlmail.Address {
	var mailReceivers []*mdlmail.Address
	receivers := strings.Split(s.c.PushVerifyEmailReceivers.EmailReceivers, ",")
	for _, v := range receivers {
		user := &mdlmail.Address{
			Address: v,
			Name:    "",
		}
		mailReceivers = append(mailReceivers, user)
	}
	return mailReceivers
}

func (s *Service) BuildCardAndPushDataBus(ctx context.Context, stime, etime string, uniqueId int64) (err error) {
	const limit = 5000
	var offset int
	for {
		var res map[int64]*like.UpActReserveRelation
		res, err = s.dao.QueryDateInterval(ctx, stime, etime, strconv.FormatInt(int64(limit), 10), strconv.FormatInt(int64(offset), 10))
		if err != nil {
			err = errors.Wrapf(err, "s.dao.QueryDateInterval err")
			return
		}
		if len(res) == 0 {
			log.Infoc(ctx, "s.dao.QueryDateInterval res invalid")
			return
		}

		for _, v := range res {
			if err = s.dao.BuildPushCard(ctx, v.Sid, uniqueId); err != nil {
				err = errors.Wrapf(err, "s.dao.BuildPushCard err")
				return
			}
		}
		time.Sleep(30 * time.Second)
		for _, v := range res {
			if !s.isCardReady(ctx, uniqueId, v.Sid) {
				log.Infoc(ctx, "card not ready, sid(%v)", v.Sid)
				continue
			}
			reqParam := &like.LotteryReserveNotify{1001, uniqueId, int64(v.Sid), []int64{v.Mid}, 1, time.Now().Unix()}
			if err = s.dao.SendLotteryNotify2Tunnel(ctx, reqParam); err != nil {
				log.Errorc(ctx, like.UpActReserveRelationPushVerifyCard+"upReservePushPub.Send error, mid(%v) uniqueID(%d) CardUniqueId(%v) error(%+v)", v.Mid, uniqueId, v.Sid, err)
				err = nil
			}
		}
		offset += limit
		if len(res) < limit {
			break
		}
	}
	log.Infoc(ctx, "s.BuildCardAndPushDataBus pushed to databus")
	return
}

func (s *Service) isCardReady(ctx context.Context, uniqueID, cardUniqueID int64) bool {
	// 查询卡片状态
	flag := false
	cardReq := &tunnelV2Mdl.CardReq{
		BizId:        1001,
		UniqueId:     uniqueID,
		CardUniqueId: cardUniqueID,
	}
	err := retry.WithAttempts(ctx, "", like.UpActReserverelationRetry, netutil.DefaultBackoffConfig, func(ctx context.Context) (err error) {
		cardRes, err := client.TunnelClient.Card(ctx, cardReq)
		log.Infoc(ctx, "client.TunnelClient.Card req(%+v) res (%+v) err(%+v)", cardReq, cardRes, err)
		if err == nil && cardRes != nil && cardRes.State == tunnelCommon.CardStateDelivering {
			flag = true
			return
		}
		if err != nil {
			log.Errorc(ctx, "isCardReady error, uniqueID(%d) CardUniqueId(%v),err is (%v)!", uniqueID, cardUniqueID, err)
		}
		return
	})

	if err != nil {
		log.Errorc(ctx, "retry err(%v)", err)
	}
	return flag
}
