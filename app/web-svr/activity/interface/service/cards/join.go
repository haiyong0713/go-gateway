package cards

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/client"
	cardsmdl "go-gateway/app/web-svr/activity/interface/model/cards"
	likemdl "go-gateway/app/web-svr/activity/interface/model/like"
	riskmdl "go-gateway/app/web-svr/activity/interface/model/risk"

	accountapi "git.bilibili.co/bapis/bapis-go/account/service"

	"math/rand"

	"time"
)

const (
	tokenSalt       = "7Jlet2apvN"
	inviteBusiness  = "invite"
	donateBusiness  = "donate"
	followBusiness  = "follow"
	ogvBusiness     = "ogv"
	archiveBusiness = "videoup"
	liveBusiness    = "live"
	signBusiness    = "sign"
)

// NewIsCanJoin 是否加入活动
func (s *Service) NewIsCanJoin(ctx context.Context, mid int64, activity string) (err error) {
	cardsConfig, err := s.dao.CardsConfig(ctx, activity)
	if err != nil {
		log.Errorc(ctx, "s.dao.CardsConfig err(%v)", err)
		return
	}
	if cardsConfig == nil {
		err = ecode.ActivityNotExist
		return
	}
	repet, err := s.likeDao.ReserveOnly(ctx, cardsConfig.ReserveID, mid)
	if err != nil {
		log.Errorc(ctx, "s.likeDao.ReserveOnly err(%v)", err)
		return err
	}
	if repet != nil && repet.ID > 0 && repet.State == 1 {
		err = ecode.SpringFestivalInviterAlreadyJoinErr
		return
	}
	return nil
}

// NewIsJoin 是否加入活动
func (s *Service) NewIsJoin(ctx context.Context, mid int64, activity string) (err error) {
	cardsConfig, err := s.dao.CardsConfig(ctx, activity)
	if err != nil {
		log.Errorc(ctx, "s.dao.CardsConfig err(%v)", err)
		return
	}
	if cardsConfig == nil {
		err = ecode.ActivityNotExist
		return
	}
	repet, err := s.likeDao.ReserveOnly(ctx, cardsConfig.ReserveID, mid)
	if err != nil {
		log.Errorc(ctx, "s.likeDao.ReserveOnly err(%v)", err)
		return err
	}
	if repet != nil && repet.ID > 0 && repet.State == 1 {
		return
	}
	return ecode.SpringFestivalNotJoinErr

}

// IsCanJoin 是否加入活动
func (s *Service) IsCanJoin(ctx context.Context, mid int64) (err error) {
	repet, err := s.likeDao.ReserveOnly(ctx, s.c.Cards.JoinReserveSid, mid)
	if err != nil {
		log.Errorc(ctx, "s.likeDao.ReserveOnly err(%v)", err)
		return err
	}
	if repet != nil && repet.ID > 0 && repet.State == 1 {
		err = ecode.SpringFestivalInviterAlreadyJoinErr
		return
	}
	return nil
}

// IsJoin 是否加入活动
func (s *Service) IsJoin(ctx context.Context, mid int64) (err error) {
	repet, err := s.likeDao.ReserveOnly(ctx, s.c.Cards.JoinReserveSid, mid)
	if err != nil {
		log.Errorc(ctx, "s.likeDao.ReserveOnly err(%v)", err)
		return err
	}
	if repet != nil && repet.ID > 0 && repet.State == 1 {
		return
	}
	return ecode.SpringFestivalNotJoinErr

}

// InviteShare 分享 返回token
func (s *Service) InviteShare(ctx context.Context, mid int64, activity string) (res *cardsmdl.InviteTokenReply, err error) {
	res = &cardsmdl.InviteTokenReply{}
	token, err := s.dao.GetInviteMidToToken(ctx, mid, activity)
	if err != nil {
		log.Errorc(ctx, " s.dao.GetInviteMidToToken err(%v)", err)
		return
	}
	if token == "" {
		return res, ecode.SpringFestivalInviterTokenErr
	}
	res.Token = token
	return
}

// Bind ...
func (s *Service) Bind(ctx context.Context, mid int64, token string, risk *riskmdl.Base, mobiApp, activity string) (err error) {
	eg := errgroup.WithContext(ctx)
	var inviter, oldInviter int64
	eg.Go(func(ctx context.Context) (err error) {
		inviter, err = s.inviteTokenToMid(ctx, token, activity)
		if err != nil {
			log.Errorc(ctx, "s.inviteTokenToMid (%s) err(%v)", token, err)
			return
		}
		if inviter <= 0 {
			return ecode.SpringFestivalInviterTokenErr
		}
		return nil
	})
	eg.Go(func(ctx context.Context) (err error) {
		// 是否已经被邀请过
		oldInviter, err = s.dao.GetMidInviter(ctx, mid, activity)
		if err != nil {
			log.Errorc(ctx, "s.dao.GetMidInviter(%d) (%s)", mid, err)
			return
		}
		if oldInviter > 0 {
			return ecode.SpringFestivalInviterAlreadyBindErr

		}
		return nil
	})
	eg.Go(func(ctx context.Context) (err error) {
		// 是否已经加入活动
		return s.IsCanJoin(ctx, mid)
	})

	if err = eg.Wait(); err != nil {
		log.Errorc(ctx, "eg.Wait error(%v)", err)
		return
	}

	if inviter == mid {
		return ecode.SpringFestivalCanInviteSelfErr
	}

	spRisk := &riskmdl.Sf21Invited{
		Base:        *risk,
		Mid:         mid,
		InvitedMid:  inviter,
		ActivityUID: s.c.Cards.ActivityUID,
		MobiApp:     mobiApp,
	}
	riskReply, err := s.risk(ctx, mid, riskmdl.ActionCardsJoin, spRisk, spRisk.EsTime)
	if err != nil {
		log.Errorc(ctx, "s.risk mid(%d) invite err(%v)", mid, err)
	} else {
		err = s.riskCheck(ctx, mid, riskReply)
		if err != nil {
			return err
		}
	}
	_, err = s.dao.InsertRelationBind(ctx, inviter, mid, token, activity)
	if err != nil {
		log.Errorc(ctx, "s.dao.InsertRelationBind inviter(%d), mid(%d) err(%v)", inviter, mid, err)
		return ecode.SpringFestivalJoinErr
	}
	return nil

}

// inviteTokenToMid 邀请token转mid
func (s *Service) inviteTokenToMid(ctx context.Context, token, activity string) (mid int64, err error) {
	mid, err = s.dao.GetInviteTokenToMid(ctx, token, activity)
	if err != nil {
		return mid, ecode.SpringFestivalGetInviter
	}
	return
}

// createToken 创建token
func (s *Service) createToken(ctx context.Context, mid int64, activity string, times int64) string {
	rand.Seed(time.Now().UnixNano())
	outerRand := rand.Intn(100000)
	p := fmt.Sprintf("%d%s%d%d%s", mid, activity, times, outerRand, tokenSalt)
	hasher := md5.New()
	hasher.Write([]byte(p))
	return hex.EncodeToString(hasher.Sum(nil))
}

// midInsertSpringNums ...
func (s *Service) midInsertSpringNums(ctx context.Context, mid int64) (err error) {
	_, isEmpty, err := s.dao.MidNumsEmptyErr(ctx, mid)
	if err != nil {
		log.Errorc(ctx, "s.dao.MidNumsEmptyErr mid(%d) err(%v)", mid, err)
		return err
	}
	if isEmpty {
		_, err = s.dao.InsertSpringNums(ctx, mid)
		if err != nil {
			log.Errorc(ctx, "s.dao.InsertSpringNums mid(%d)", mid)
			return err
		}
	}
	return nil
}

// Join 加入活动
func (s *Service) Join(ctx context.Context, mid int64, report *likemdl.ReserveReport, activity string) (err error) {
	// 通知用户 加入filter
	err = s.actDao.AddFilterSet(ctx, mid, s.c.Cards.Activity, s.c.Cards.Filter)
	if err != nil {
		return err
	}
	err = s.midInsertSpringNums(ctx, mid)
	if err != nil {
		log.Errorc(ctx, "s.midInsertSpringNums (%v)", err)
		return ecode.SpringFestivalJoinErr
	}
	// 生成mid token
	token := s.createToken(ctx, mid, s.c.Cards.Activity, time.Now().Unix())
	_, err = s.dao.InsertSpringMidInviteToken(ctx, mid, token, activity)
	if err != nil {
		log.Errorc(ctx, "s.dao.InsertSpringMidInviteToken (%v)", err)
		return ecode.SpringFestivalJoinErr
	}
	// 加入活动
	err = s.likeSvr.AsyncReserve(ctx, s.c.Cards.JoinReserveSid, mid, 1, report)
	if err != nil {
		return err
	}

	err = s.dao.DeleteInviteTokenToMid(ctx, token, activity)
	if err != nil {
		log.Errorc(ctx, "s.dao.DeleteInviteTokenToMid token(%s) err(%v)", token, err)
	}
	// 是否被邀请
	inviter, err := s.dao.GetMidInviter(ctx, mid, activity)
	if err != nil {
		log.Errorc(ctx, "s.dao.MidInviter mid(%d) err(%v)", mid, err)
	}
	if inviter > 0 {
		// 给邀请人完成任务
		err = s.actSend(ctx, inviter, mid, inviteBusiness)
		if err != nil {
			log.Errorc(ctx, "s.actSend inviter(%d) mid(%d) invitebusiness", inviter, mid)
		}
	}

	return nil

}

// ShareTokenToMid 分享token转mid
func (s *Service) ShareTokenToMid(ctx context.Context, token, activity string) (res *cardsmdl.ShareTokenToMidReply, err error) {
	res = &cardsmdl.ShareTokenToMidReply{}
	res.Account = &cardsmdl.Account{}
	inviter, err := s.inviteTokenToMid(ctx, token, activity)
	if err != nil {
		log.Errorc(ctx, "s.inviteTokenToMid (%s) err(%v)", token, err)
		return
	}
	if inviter <= 0 {
		return res, ecode.SpringFestivalInviterTokenErr
	}
	res.Account, err = s.midToAccount(ctx, inviter)
	if err != nil {
		return
	}
	return
}

// midToAccount
func (s *Service) midToAccount(ctx context.Context, mid int64) (res *cardsmdl.Account, err error) {
	res = &cardsmdl.Account{}
	infosReply, err := client.AccountClient.Info3(ctx, &accountapi.MidReq{Mid: mid})
	if err != nil || infosReply == nil || infosReply.Info == nil {
		log.Errorc(ctx, "s.AccClient.Info3: error(%v) mid(%d)", err, mid)
		return res, ecode.SpringFestivalCantGetTokenMidErr
	}
	return s.accountToAccount(ctx, infosReply.Info), nil
}
