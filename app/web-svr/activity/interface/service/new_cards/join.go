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

	accountapi "git.bilibili.co/bapis/bapis-go/account/service"

	"math/rand"

	"time"
)

const (
	tokenSalt       = "7Jlet2apvN"
	inviteBusiness  = "act_plat_activity_invite"
	donateBusiness  = "act_plat_activity_donate"
	followBusiness  = "act_plat_activity_follow"
	ogvBusiness     = "ogv"
	archiveBusiness = "videoup"
	liveBusiness    = "live"
	signBusiness    = "sign"
)

func (s *Service) cardsConfig(ctx context.Context, activity string) (cardsConfig *cardsmdl.Cards, err error) {
	cardsConfig, err = s.dao.CardsConfig(ctx, activity)
	if err != nil {
		log.Errorc(ctx, "s.dao.CardsConfig err(%v)", err)
		return
	}
	if cardsConfig == nil {
		err = ecode.ActivityNotExist
		return
	}
	return
}

// IsCanJoin 是否加入活动
func (s *Service) IsCanJoin(ctx context.Context, mid int64, activity string) (err error) {
	cardsConfig, err := s.cardsConfig(ctx, activity)
	if err != nil {
		log.Errorc(ctx, "s.dao.CardsConfig err(%v)", err)
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

// IsJoin 是否加入活动
func (s *Service) IsJoin(ctx context.Context, mid int64, activity string) (err error) {
	cardsConfig, err := s.cardsConfig(ctx, activity)
	if err != nil {
		log.Errorc(ctx, "s.dao.CardsConfig err(%v)", err)
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

// Bind ...
func (s *Service) Bind(ctx context.Context, mid int64, token string, activity string) (err error) {
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
		return s.IsCanJoin(ctx, mid, activity)
	})

	if err = eg.Wait(); err != nil {
		log.Errorc(ctx, "eg.Wait error(%v)", err)
		return
	}

	if inviter == mid {
		return ecode.SpringFestivalCanInviteSelfErr
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
func (s *Service) midInsertInit(ctx context.Context, mid int64, id int64, nums int64) (err error) {
	return s.dao.InitAddMidCards(ctx, id, mid, nums)
}

// Join 加入活动
func (s *Service) Join(ctx context.Context, mid int64, report *likemdl.ReserveReport, activity string) (err error) {
	cardsConfig, err := s.cardsConfig(ctx, activity)
	if err != nil {
		log.Errorc(ctx, "s.dao.CardsConfig err(%v)", err)
		return
	}
	// 通知用户 加入filter
	err = s.actDao.AddFilterSet(ctx, mid, cardsConfig.Name, s.c.Cards.Filter)
	if err != nil {
		return err
	}
	err = s.midInsertInit(ctx, mid, cardsConfig.ID, cardsConfig.CardsNum)
	if err != nil {
		log.Errorc(ctx, "s.midInsertInit (%v)", err)
		return ecode.SpringFestivalJoinErr
	}
	// 生成mid token
	token := s.createToken(ctx, mid, cardsConfig.Name, time.Now().Unix())
	_, err = s.dao.InsertSpringMidInviteToken(ctx, mid, token, activity)
	if err != nil {
		log.Errorc(ctx, "s.dao.InsertSpringMidInviteToken (%v)", err)
		return ecode.SpringFestivalJoinErr
	}
	// 加入活动
	err = s.likeSvr.AsyncReserve(ctx, cardsConfig.ReserveID, mid, 1, report)
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
		err = s.taskSvr.CardsActSend(ctx, inviter, inviteBusiness, activity, time.Now().Unix(), nil, isInternal)
		if err != nil {
			log.Errorc(ctx, "s.taskSvr.CardsActSend inviter(%d) mid(%d) invitebusiness", inviter, mid)
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

func (s *Service) accountToAccount(c context.Context, midInfo *accountapi.Info) *cardsmdl.Account {
	return &cardsmdl.Account{
		Mid:  midInfo.Mid,
		Name: midInfo.Name,
		Face: midInfo.Face,
		Sign: midInfo.Sign,
		Sex:  midInfo.Sex,
	}
}
