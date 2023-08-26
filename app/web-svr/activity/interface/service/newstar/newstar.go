package newstar

import (
	"context"
	"time"

	accmdl "git.bilibili.co/bapis/bapis-go/account/service"
	memberAPI "git.bilibili.co/bapis/bapis-go/account/service/member"
	"go-common/library/log"
	xtime "go-common/library/time"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/model/newstar"

	relationmdl "git.bilibili.co/bapis/bapis-go/account/service/relation"
	upmdl "git.bilibili.co/bapis/bapis-go/archive/service/up"

	"go-common/library/sync/errgroup.v2"
)

const (
	_canReplyV = 1
	_unDo      = 0
	_checking  = 1
	_finish    = 2
)

func (s *Service) creationInfo(c context.Context, ActivityUID string, mid int64) (res *newstar.Newstar, err error) {
	if res, err = s.dao.CacheCreationByMid(c, mid, ActivityUID); err != nil {
		err = nil
	}
	if res != nil {
		return
	}
	if res, err = s.dao.RawCreation(c, ActivityUID, mid); err != nil {
		return
	}
	if res != nil && res.ID > 0 {
		s.cache.Do(c, func(c context.Context) {
			s.dao.AddCacheCreationByMid(c, mid, ActivityUID, res)
		})
	}
	return
}

func (s *Service) inviteCount(c context.Context, ActivityUID string, inviterMid int64) (res int64, err error) {
	if res, err = s.dao.CacheInviteCount(c, inviterMid, ActivityUID); err != nil {
		err = nil
	}
	if res > 0 {
		return
	}
	if res, err = s.dao.InviteCount(c, ActivityUID, inviterMid); err != nil {
		return
	}
	if res > 0 {
		s.cache.Do(c, func(c context.Context) {
			s.dao.AddCacheInviteCount(c, inviterMid, ActivityUID, res)
		})
	}
	return
}

func (s *Service) JoinNewstar(c context.Context, ActivityUID string, mid, inviterMid int64) (err error) {
	var (
		vStatus     int64
		creation    *newstar.Newstar
		lastID      int64
		inviteCount int64
	)
	// 活动结束判断
	if time.Now().Unix() >= s.c.Rule.NewstarStop {
		err = ecode.ActGuessOverEnd
		return
	}
	if mid == inviterMid {
		err = ecode.ActivityStarSelfErr
		return
	}
	if inviterMid > 0 {
		if inviteCount, err = s.inviteCount(c, ActivityUID, inviterMid); err != nil {
			return
		}
		if inviteCount >= s.c.Rule.NewstarInviteMax {
			err = ecode.ActivityStarLimitErr
			return
		}
	}
	group := errgroup.WithContext(c)
	group.Go(func(ctx context.Context) error {
		if creation, err = s.creationInfo(c, ActivityUID, mid); err != nil {
			log.Error("NewstarCreation s.creationInfo (%s,%d) error(%+v)", ActivityUID, mid, err)
			return err
		}
		return nil
	})
	group.Go(func(ctx context.Context) error {
		if vStatus, err = s.isVStatus(c, mid); err != nil {
			log.Error("JoinNewstar s.isVStatus mid(%d) error(%+v)", mid, err)
			return err
		}
		return nil
	})
	if err = group.Wait(); err != nil {
		log.Error("JoinNewstar group.Wait mid(%d) error(%+v)", mid, err)
		return
	}
	if creation != nil && creation.ID > 0 {
		if creation.InviterMid > 0 {
			err = ecode.ActivityStarBeforeErr
			return
		}
		err = ecode.ActivityStarAlreadyErr
		return
	}
	lastID, err = s.dao.JoinNewstar(c, ActivityUID, vStatus, mid, inviterMid)
	if vStatus != _canReplyV {
		err = ecode.ActivityStarNotVErr
	}
	if inviterMid == 0 || lastID == 0 {
		return
	}
	s.dao.DelCacheInviteCount(c, inviterMid, ActivityUID)
	s.cache.Do(c, func(c context.Context) {
		s.dao.DelCacheCacheInvite(c, inviterMid, ActivityUID)
	})
	return
}

func (s *Service) NewstarCreation(c context.Context, ActivityUID string, mid int64) (creation *newstar.Newstar, err error) {
	var (
		user      *accmdl.Profile
		statReply *relationmdl.StatReply
	)
	group := errgroup.WithContext(c)
	group.Go(func(ctx context.Context) error {
		if creation, err = s.creationInfo(c, ActivityUID, mid); err != nil {
			log.Error("NewstarCreation s.creationInfo (%s,%d) error(%+v)", ActivityUID, mid, err)
			return err
		}
		return nil
	})
	group.Go(func(ctx context.Context) error {
		account, accErr := s.accClient.Profile3(c, &accmdl.MidReq{Mid: mid})
		if accErr != nil {
			log.Error("s.accClient.Profile3(%d) error(%v)", mid, accErr)
			return accErr
		}
		user = account.Profile
		return nil
	})
	group.Go(func(ctx context.Context) error {
		statReply, err = s.relationClient.Stat(ctx, &relationmdl.MidReq{Mid: mid})
		if err != nil || statReply == nil {
			log.Error("s.relationClient.Stat(%d) error(%v)", mid, err)
			return err
		}
		return nil
	})
	if err = group.Wait(); err != nil {
		log.Error("NewstarCreation group.Wait mid(%d) error(%+v)", mid, err)
		return
	}
	if creation == nil || creation.ID == 0 {
		return
	}
	// 未结算的取实时
	if creation.FinishTask == _unDo {
		group2 := errgroup.WithContext(c)
		group2.Go(func(ctx context.Context) error {
			if user.Identification == 0 {
				realRes, realErr := s.memberClient.RealnameApplyStatus(c, &memberAPI.MemberMidReq{Mid: mid})
				if realErr != nil {
					log.Error("NewstarCreation s.memberClient.RealnameApplyStatus(%d) error(%+v)", mid, realErr)
					creation.IsName = _unDo
					return nil
				}
				switch realRes.Status {
				case 0:
					creation.IsName = _checking
				case 1:
					creation.IsName = _finish
				default:
					creation.IsName = _unDo
				}
			} else {
				creation.IsName = _finish
			}
			return nil
		})
		group2.Go(func(ctx context.Context) error {
			if user.Official.Role == 0 {
				officialRes, officialErr := s.memberClient.OfficialDoc(c, &memberAPI.MidReq{Mid: mid})
				if officialErr != nil {
					log.Error("NewstarCreation s.memberClient.OfficialDoc(%d) error(%+v)", mid, officialErr)
					creation.IsIdentity = _unDo
					return nil
				}
				switch officialRes.State {
				case 0:
					creation.IsIdentity = _checking
				case 1:
					creation.IsIdentity = _finish
				default:
					creation.IsIdentity = _unDo
				}
			} else {
				creation.IsIdentity = _finish
			}
			return nil
		})
		if err = group2.Wait(); err != nil {
			log.Error("NewstarCreation group2.Wait mid(%d) error(%+v)", mid, err)
			return
		}
		creation.FansCount = statReply.Follower
		if user.TelStatus == 1 {
			creation.IsMobile = _finish
		}
		creation.RemainingDays = s.remainingDays(creation.Ctime)
	}
	//计算完成任务奖励
	baseMoney, inviteMoney := s.creationAward(ActivityUID, creation)
	creation.ReceiveAward = baseMoney + inviteMoney
	return
}

func (s *Service) remainingDays(ctime xtime.Time) int64 {
	startTime := time.Unix(ctime.Time().Unix(), 0)
	endTime := startTime.Add(time.Duration(s.c.Rule.NewstarDays))
	endDate := time.Date(endTime.Year(), endTime.Month(), endTime.Day(), 23, 59, 59, 0, time.Local)
	subTime := endDate.Sub(time.Now()).Hours()
	rs := int64(subTime / 24)
	if rs < 0 {
		return 0
	}
	return rs
}

func (s *Service) invitesInfo(c context.Context, ActivityUID string, mid int64) (list []*newstar.Newstar, err error) {
	if list, err = s.dao.CacheInvites(c, mid, ActivityUID); err != nil {
		err = nil
	}
	if len(list) > 0 {
		return
	}
	if list, err = s.dao.RawInvites(c, ActivityUID, mid); err != nil {
		return
	}
	if len(list) > 0 {
		s.cache.Do(c, func(c context.Context) {
			s.dao.AddCacheInvites(c, mid, ActivityUID, list)
		})
	}
	return
}

func (s *Service) NewstarInvite(c context.Context, ActivityUID string, mid int64, pn, ps int) (res *newstar.NewstarInvite, err error) {
	var (
		creation  *newstar.Newstar
		invites   []*newstar.Newstar
		userInfo  *accmdl.Info
		rsInvites []*newstar.UserInfo
	)
	group := errgroup.WithContext(c)
	group.Go(func(ctx context.Context) error {
		if creation, err = s.creationInfo(c, ActivityUID, mid); err != nil {
			log.Error("NewstarInvite s.creationInfo (%s,%d) error(%+v)", ActivityUID, mid, err)
			return err
		}
		return nil
	})
	group.Go(func(ctx context.Context) error {
		if invites, err = s.invitesInfo(c, ActivityUID, mid); err != nil {
			log.Error("NewstarInvite s.invitesInfo (%s,%d) error(%+v)", ActivityUID, mid, err)
			return err
		}
		return nil
	})
	group.Go(func(ctx context.Context) error {
		userRes, accErr := s.accClient.Info3(c, &accmdl.MidReq{Mid: mid})
		if accErr != nil {
			log.Error("s.accClient.Info3(%d) error(%v)", mid, accErr)
			return accErr
		}
		userInfo = userRes.Info
		return nil
	})
	if err = group.Wait(); err != nil {
		log.Error("NewstarInvite group.Wait mid(%d) error(%+v)", mid, err)
		return
	}
	count := len(invites)
	res = &newstar.NewstarInvite{
		VStatus:     creation.VStatus,
		Mid:         userInfo.Mid,
		Name:        userInfo.Name,
		Face:        userInfo.Face,
		InviteAward: 0,
		List:        make([]*newstar.UserInfo, 0),
		Page: &newstar.Page{
			Num:   pn,
			Size:  ps,
			Total: count,
		},
	}
	if count == 0 {
		return
	}
	inviteUser, award := s.inviteUserInfo(c, ActivityUID, invites)
	start := (pn - 1) * ps
	end := start + ps - 1
	if count > end+1 {
		rsInvites = inviteUser[start : end+1]
	} else {
		rsInvites = inviteUser[start:]
	}
	res.InviteAward = award
	s.rebuildUser(c, rsInvites)
	res.List = rsInvites
	return
}

func (s *Service) rebuildUser(c context.Context, rsInvites []*newstar.UserInfo) {
	var mids []int64
	for _, inviter := range rsInvites {
		mids = append(mids, inviter.Mid)
	}
	accReply, err := s.accClient.Infos3(c, &accmdl.MidsReq{Mids: mids})
	if err != nil {
		log.Error("s.accClient.Infos3(%+v) error(%v)", mids, err)
		return
	}
	infos := accReply.Infos
	for _, inviter := range rsInvites {
		if info, ok := infos[inviter.Mid]; ok {
			inviter.Name = info.Name
			inviter.Face = info.Face
			inviter.RemainingDays = s.remainingDays(inviter.Ctime)
		}
	}
	return
}

func (s *Service) inviteUserInfo(c context.Context, activityUID string, invites []*newstar.Newstar) (list []*newstar.UserInfo, award int64) {
	if err := s.rebuildFansRz(c, invites); err != nil {
		log.Error("inviteUserInfo s.rebuildFans error(%+v)", err)
		return
	}
	for _, invite := range invites {
		list = append(list, &newstar.UserInfo{
			ID:            invite.ID,
			Mid:           invite.Mid,
			BaseTask:      invite.IsIdentity == _finish && invite.UpArchives == 1,
			FansCount:     invite.FansCount,
			Ctime:         invite.Ctime,
			RemainingDays: s.remainingDays(invite.Ctime),
		})
		// 计算奖励.
		baseMoney, inviteMoney := s.inviteAward(activityUID, invite)
		award += baseMoney + inviteMoney
	}
	return
}

func (s *Service) rebuildFansRz(c context.Context, invites []*newstar.Newstar) error {
	var (
		mids    []int64
		err     error
		statMap map[int64]*relationmdl.StatReply
		userMap map[int64]*accmdl.ProfileWithoutPrivacy
	)
	for _, invite := range invites {
		// 未结算的取实时
		if invite.FinishTask == _unDo {
			mids = append(mids, invite.Mid)
		}
	}
	if len(mids) == 0 {
		return nil
	}
	group := errgroup.WithContext(c)
	group.Go(func(ctx context.Context) error {
		statsReply, relaErr := s.relationClient.Stats(c, &relationmdl.MidsReq{Mids: mids})
		if relaErr != nil {
			log.Error("rebuildFansRz s.relationClient.Stats(%d) error(%v)", mids, relaErr)
			return relaErr
		}
		statMap = statsReply.StatReplyMap
		return nil
	})
	group.Go(func(ctx context.Context) error {
		userRes, accErr := s.accClient.ProfilesWithoutPrivacy3(c, &accmdl.MidsReq{Mids: mids})
		if accErr != nil {
			log.Error("rebuildFansRz s.accClient.ProfilesWithoutPrivacy3(%+v) error(%v)", mids, accErr)
			return accErr
		}
		userMap = userRes.ProfilesWithoutPrivacy
		return nil
	})
	if err = group.Wait(); err != nil {
		log.Error("rebuildFansRz group.Wait mid(%+v) error(%+v)", mids, err)
		return err
	}
	for _, invite := range invites {
		if stat, ok := statMap[invite.Mid]; ok {
			invite.FansCount = stat.Follower
		}
		if user, ok := userMap[invite.Mid]; ok && user.Official.Role > 0 {
			invite.IsIdentity = _finish
		}
	}
	return nil
}

func (s *Service) creationAward(activityUID string, creation *newstar.Newstar) (baseMoney, inviteMoney int64) {
	awards, ok := s.newstarAwards[activityUID]
	if !ok || len(awards) == 0 {
		log.Error("creationAward activityUID(%s) 0", activityUID)
		return
	}
	for _, award := range awards {
		if award.AwardType == 1 && creation.IsIdentity == award.Condition && creation.UpArchives == 1 {
			baseMoney = award.FinishMoney
		}
		if award.AwardType == 2 && creation.FansCount >= award.Condition {
			inviteMoney = award.FinishMoney
		}
	}
	return
}

func (s *Service) inviteAward(activityUID string, inviter *newstar.Newstar) (baseMoney, inviteMoney int64) {
	awards, ok := s.newstarAwards[activityUID]
	if !ok || len(awards) == 0 {
		log.Error("inviteAward activityUID(%s) 0", activityUID)
		return
	}
	for _, award := range awards {
		if award.AwardType == 1 && inviter.IsIdentity == award.Condition && inviter.UpArchives == 1 {
			baseMoney = award.InviteMoney
		}
		if award.AwardType == 2 && inviter.FansCount >= award.Condition {
			inviteMoney = award.InviteMoney
		}
	}
	return
}

func (s *Service) userVInfo(c context.Context, mid int64) (user *accmdl.Profile, arcExist bool, err error) {
	group := errgroup.WithContext(c)
	group.Go(func(ctx context.Context) error {
		account, accErr := s.accClient.Profile3(c, &accmdl.MidReq{Mid: mid})
		if accErr != nil {
			log.Error("s.accClient.Profile3(%d) error(%v)", mid, accErr)
			return accErr
		}
		user = account.Profile
		return nil
	})
	group.Go(func(ctx context.Context) error {
		if upFirstReply, upcErr := s.upClient.GetUpFirstArchive(ctx, &upmdl.UpFirstArchiveReq{Mid: mid}); upcErr != nil {
			log.Error("Card s.upGRPC.UpCount(%d) error %v", mid, upcErr)
			return upcErr
		} else {
			arcExist = upFirstReply.Exist
		}
		return nil
	})
	err = group.Wait()
	return
}

func (s *Service) isVStatus(c context.Context, mid int64) (res int64, err error) {
	var (
		user     *accmdl.Profile
		arcExist bool
	)
	if user, arcExist, err = s.userVInfo(c, mid); err != nil {
		log.Error("isVStatus  s.userVInfo(%d) error(%+v)", mid, err)
		return
	}
	if user.Official.Role > s.c.Rule.NewstarRole {
		return
	}
	if !arcExist {
		res = _canReplyV
	}
	return
}
