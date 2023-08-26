package service

import (
	"context"
	"go-common/library/log"
	"go-gateway/app/web-svr/web/interface/model"

	relaapi "git.bilibili.co/bapis/bapis-go/account/service/relation"
	ansmdl "git.bilibili.co/bapis/bapis-go/community/interface/answer"

	accmdl "git.bilibili.co/bapis/bapis-go/account/service"
	coumdl "git.bilibili.co/bapis/bapis-go/account/service/coupon"
	pangugsgrpc "git.bilibili.co/bapis/bapis-go/pangu/platform/gallery-service"

	"go-common/library/sync/errgroup.v2"
)

const _notAnswer = 1

// Nav api service
func (s *Service) Nav(c context.Context, mid int64, cookie string) (resp *model.NavResp, err error) {
	var (
		wallet      *model.Wallet
		hasShop     bool
		shopURL     string
		answer      int32
		allowance   int64
		faceNftType pangugsgrpc.NFTRegionType
	)
	profile := new(accmdl.ProfileStatReply)
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) error {
		var e error
		if profile, e = s.accGRPC.ProfileWithStat3(ctx, &accmdl.MidReq{Mid: mid}); e != nil {
			log.Error("s.accGRPC.ProfileWithStat3(%d) error %v", mid, e)
			profile = model.DefaultProfile
			profile.Profile.Mid = mid
		}
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		var shop *model.ShopInfo
		var e error
		if shop, e = s.dao.ShopInfo(ctx, mid); e == nil && shop != nil {
			hasShop = true
			shopURL = shop.JumpURL
		} else {
			log.Warn("s.dao.ShopInfo(%v) error(%+v)", mid, e)
		}
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		var e error
		if wallet, e = s.dao.Wallet(ctx, mid); e != nil || wallet == nil {
			log.Error("s.dao.Wallet(%d) error(%v)", mid, e)
		} else {
			log.Info("account wallet mid(%d)", mid)
		}
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		var (
			e     error
			reply *coumdl.AllowanceCountReply
		)
		if reply, e = s.couponGRPC.AllowanceCount(ctx, &coumdl.MidReq{Mid: mid}); e != nil {
			log.Error("s.coupon.AllowanceCount(%d) error(%v)", mid, e)
		} else {
			allowance = reply.Count
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		log.Error("%+v", err)
	}
	midNFTRegionMap := s.BatchNFTRegion(c, []int64{mid})
	faceNftType = midNFTRegionMap[mid]
	if profile.Profile.Rank < _member {
		if statusReply, er := s.ansGRPC.Status(c, &ansmdl.StatusReq{Mid: mid}); er != nil {
			log.Error("s.ansRPC.Status(%d) error(%v)", mid, er)
			answer = _notAnswer
		} else if statusReply != nil {
			answer = statusReply.Status.Status
		}
	}
	resp = &model.NavResp{
		IsLogin:            true,
		EmailVerified:      profile.Profile.EmailStatus,
		Face:               profile.Profile.Face,
		FaceNft:            profile.Profile.FaceNftNew,
		FaceNftType:        faceNftType,
		Mid:                profile.Profile.Mid,
		MobileVerified:     profile.Profile.TelStatus,
		Coins:              profile.Coins,
		Moral:              float32(profile.Profile.Moral),
		Pendant:            profile.Profile.Pendant,
		Uname:              profile.Profile.Name,
		VipDueDate:         profile.Profile.Vip.DueDate,
		VipStatus:          profile.Profile.Vip.Status,
		VipType:            profile.Profile.Vip.Type,
		VipPayType:         profile.Profile.Vip.VipPayType,
		VipThemeType:       profile.Profile.Vip.ThemeType,
		VipLabel:           profile.Profile.Vip.Label,
		VipAvatarSubscript: profile.Profile.Vip.AvatarSubscript,
		VipNicknameColor:   profile.Profile.Vip.NicknameColor,
		Vip:                profile.Profile.Vip,
		Wallet:             wallet,
		HasShop:            hasShop,
		ShopURL:            shopURL,
		AllowanceCount:     allowance,
		Official:           profile.Profile.Official,
		OfficialVerify:     model.FromOfficial(profile.Profile.Official),
		AnswerStatus:       answer,
		IsSeniorMember:     profile.Profile.IsSeniorMember,
	}
	resp.LevelInfo.Cur = profile.LevelInfo.Cur
	resp.LevelInfo.Min = profile.LevelInfo.Min
	resp.LevelInfo.NowExp = profile.LevelInfo.NowExp
	resp.LevelInfo.NextExp = profile.LevelInfo.NextExp
	if profile.LevelInfo.NextExp == -1 {
		resp.LevelInfo.NextExp = "--"
	}
	if !s.c.SeniorMemberSwitch.ShowSeniorMember {
		resp.IsSeniorMember = 0
	}
	return
}

// NavStat get nav user state(fan,follow,space dynamic)
func (s *Service) NavStat(c context.Context, mid int64) (data *model.NavStat) {
	group := errgroup.WithContext(c)
	data = new(model.NavStat)
	group.Go(func(ctx context.Context) error {
		if reply, err := s.relationGRPC.Stat(ctx, &relaapi.MidReq{Mid: mid}); err != nil {
			log.Error("NavStat s.relationGRPC.Stat(%d) error(%v)", mid, err)
		} else if reply != nil {
			data.Following = reply.Following
			data.Follower = reply.Follower
		}
		return nil
	})
	group.Go(func(ctx context.Context) error {
		if dyCount, err := s.dao.DynamicNumV2(ctx, mid); err != nil {
			log.Error("NavStat s.dao.DynamicNum(%d) error(%v)", mid, err)
		} else {
			data.DynamicCount = dyCount
		}
		return nil
	})
	if err := group.Wait(); err != nil {
		log.Error("%+v", err)
	}
	return
}
