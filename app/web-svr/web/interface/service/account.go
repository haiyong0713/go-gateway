package service

import (
	"context"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/web-svr/web/interface/model"
	mecode "go-gateway/ecode"

	accmdl "git.bilibili.co/bapis/bapis-go/account/service"
	relmdl "git.bilibili.co/bapis/bapis-go/account/service/relation"
	artmdl "git.bilibili.co/bapis/bapis-go/article/service"
	thumbupgrpc "git.bilibili.co/bapis/bapis-go/community/service/thumbup"
	uparcgrpc "git.bilibili.co/bapis/bapis-go/up-archive/service"
)

const _cardBakCacheRand = 10

// Attentions get attention list.
func (s *Service) Attentions(c context.Context, mid int64) (rs []int64, err error) {
	var (
		reply    *relmdl.FollowingsReply
		remoteIP = metadata.String(c, metadata.RemoteIP)
	)
	if reply, err = s.relationGRPC.Followings(c, &relmdl.MidReq{Mid: mid, RealIp: remoteIP}); err != nil {
		log.Error("Attentions s.relationGRPC.Followings(%d,%s) error %v", mid, remoteIP, err)
	} else if reply != nil {
		rs = make([]int64, 0)
		for _, v := range reply.FollowingList {
			rs = append(rs, v.Mid)
		}
	}
	return
}

// Card get card relation archive count data.
// nolint: gocognit
func (s *Service) Card(c context.Context, mid, loginID int64, topPhoto, article bool) (rs *model.Card, err error) {
	var (
		cardReply                                          *accmdl.CardReply
		upCountReply                                       *uparcgrpc.ArcPassedTotalReply
		card                                               *model.AccountCard
		space                                              *model.Space
		upArcCount                                         int64
		infoErr, statErr, spaceErr, relErr, upcErr, artErr error
		remoteIP                                           = metadata.String(c, metadata.RemoteIP)
	)
	relation := &accmdl.RelationReply{}
	stat := &relmdl.StatReply{}
	upArts := &artmdl.UpArtMetasReply{}
	card = new(model.AccountCard)
	card.Attentions = make([]int64, 0)
	if cardReply, infoErr = s.accGRPC.Card3(c, &accmdl.MidReq{Mid: mid}); infoErr != nil {
		if ecode.EqualError(ecode.UserNotExist, infoErr) || ecode.EqualError(mecode.MemberNotExist, infoErr) {
			err = ecode.NothingFound
			return
		}
		log.Error("Card s.accGRPC.Card3(%d,%s) error %v", mid, remoteIP, infoErr)
	} else {
		card.FromCard(cardReply.Card)
		midNFTRegionMap := s.BatchNFTRegion(c, []int64{mid})
		card.FaceNftType = midNFTRegionMap[mid]
		if !s.c.SeniorMemberSwitch.ShowSeniorMember {
			card.IsSeniorMember = 0
		}
	}
	group := errgroup.WithContext(c)
	group.Go(func(ctx context.Context) error {
		var statReply *relmdl.StatReply
		if statReply, statErr = s.relationGRPC.Stat(ctx, &relmdl.MidReq{Mid: mid, RealIp: remoteIP}); statErr != nil {
			log.Error("Card s.relationGRPC.Stat(%d) error(%v)", mid, statErr)
		} else if statReply != nil {
			stat = statReply
			card.Fans = stat.Follower
			card.Attention = stat.Following
			card.Friend = stat.Following
		}
		return nil
	})
	if topPhoto {
		group.Go(func(ctx context.Context) error {
			space, spaceErr = s.dao.TopPhoto(ctx, mid)
			return nil
		})
	}
	if loginID > 0 {
		group.Go(func(ctx context.Context) error {
			var relResp *accmdl.RelationReply
			if relResp, relErr = s.accGRPC.Relation3(ctx, &accmdl.RelationReq{Mid: loginID, Owner: mid, RealIp: remoteIP}); relErr != nil {
				log.Error("Card s.accGRPC.Relation3(%d,%d,%s) error %v", loginID, mid, remoteIP, relErr)
			} else if relResp != nil {
				relation = relResp
			}
			return nil
		})
	}
	group.Go(func(ctx context.Context) error {
		if upCountReply, upcErr = s.upArcGRPC.ArcPassedTotal(ctx, &uparcgrpc.ArcPassedTotalReq{Mid: mid}); upcErr != nil {
			log.Error("Card s.upArcGRPC.ArcPassedTotal(%d) error %v", mid, upcErr)
		} else {
			upArcCount = upCountReply.Total
		}
		return nil
	})
	if article {
		group.Go(func(ctx context.Context) error {
			if upArts, artErr = s.artGRPC.UpArtMetas(ctx, &artmdl.UpArtMetasReq{Mid: mid, Pn: _samplePn, Ps: _samplePs, Ip: remoteIP}); artErr != nil {
				log.Error("Card s.art.UpArtMetas(%d) error(%v)", mid, artErr)
			}
			if upArts == nil {
				upArts = &artmdl.UpArtMetasReply{Count: 0}
			}
			return nil
		})
	}
	var likeNum int64
	if mid > 0 {
		group.Go(func(ctx context.Context) error {
			req := &thumbupgrpc.UserLikedCountsReq{
				Mid:        mid,
				Businesses: []string{_businessLike, _articleLike, _dynamicLike, _albumLike, _clipLike, _cheeseLike},
			}
			likeCnts, err := s.thumbupGRPC.UserLikedCounts(ctx, req)
			if err != nil {
				log.Error("Fail to request thumbupGRPC.UserLikedCounts, req=%+v error=%+v", req, err)
				return nil
			}
			if likeCnts == nil {
				return nil
			}
			for _, cnt := range likeCnts.LikeCounts {
				likeNum += cnt
			}
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		log.Error("%+v", err)
	}
	addCache := true
	if infoErr != nil || (topPhoto && spaceErr != nil) || (loginID > 0 && relErr != nil) || upcErr != nil {
		if cacheRs, cacheErr := s.dao.CardBakCache(c, mid); cacheErr != nil {
			addCache = false
			log.Error("Card s.dao.CardBakCache(%d) error (%v)", mid, cacheErr)
		} else if cacheRs != nil {
			if infoErr != nil {
				card = cacheRs.Card
			}
			if statErr != nil {
				stat = &relmdl.StatReply{Follower: cacheRs.Follower}
			}
			if topPhoto && spaceErr != nil {
				space = cacheRs.Space
			}
			if loginID > 0 && relErr != nil {
				relation = &accmdl.RelationReply{Following: cacheRs.Following}
			}
			if upcErr != nil {
				upArcCount = cacheRs.ArchiveCount
			}
			if artErr != nil {
				upArts = &artmdl.UpArtMetasReply{Count: cacheRs.ArticleCount}
			}
		}
		if topPhoto && space == nil {
			space = &model.Space{SImg: s.c.DefaultTop.SImg, LImg: s.c.DefaultTop.LImg}
		}
	}
	rs = &model.Card{
		Card:         card,
		Space:        space,
		Following:    relation.Following,
		ArchiveCount: upArcCount,
		ArticleCount: upArts.Count,
		Follower:     stat.Follower,
		LikeNum:      likeNum,
	}
	if addCache {
		if err := s.cache.Do(c, func(c context.Context) {
			if s.r.Intn(_cardBakCacheRand) == 1 {
				if err := s.dao.SetCardBakCache(c, mid, rs); err != nil {
					log.Error("%+v", err)
				}
			}
		}); err != nil {
			log.Error("%+v", err)
		}
	}
	return
}

// Relation .
func (s *Service) Relation(c context.Context, mid, vmid int64) (data *model.Relation) {
	data = &model.Relation{Relation: struct{}{}, BeRelation: struct{}{}}
	ip := metadata.String(c, metadata.RemoteIP)
	if mid == vmid {
		return
	}
	group := errgroup.WithContext(c)
	group.Go(func(ctx context.Context) error {
		if relation, err := s.relationGRPC.Relation(ctx, &relmdl.RelationReq{Mid: mid, Fid: vmid, RealIp: ip}); err != nil {
			log.Error("Relation s.relation.Relation(Mid:%d,Fid:%d,%s) error %v", mid, vmid, ip, err)
		} else if relation != nil {
			data.Relation = relation
		}
		return nil
	})
	group.Go(func(ctx context.Context) error {
		if beRelation, err := s.relationGRPC.Relation(ctx, &relmdl.RelationReq{Mid: vmid, Fid: mid, RealIp: ip}); err != nil {
			log.Error("Relation s.relation.Relation(Mid:%d,Fid:%d,%s) error %v", vmid, mid, ip, err)
		} else if beRelation != nil {
			data.BeRelation = beRelation
		}
		return nil
	})
	if err := group.Wait(); err != nil {
		log.Error("%+v", err)
	}
	return
}
