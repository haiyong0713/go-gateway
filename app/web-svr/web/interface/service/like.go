package service

import (
	"context"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/sync/errgroup.v2"

	favmdl "git.bilibili.co/bapis/bapis-go/community/model/favorite"
	favecode "git.bilibili.co/bapis/bapis-go/community/service/favorite/ecode"

	mecode "git.bilibili.co/bapis/bapis-go/community/service/thumbup/ecode"
	arcecode "go-gateway/app/app-svr/archive/ecode"
	arcmdl "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/web/interface/model"

	coinmdl "git.bilibili.co/bapis/bapis-go/community/service/coin"
	favgrpc "git.bilibili.co/bapis/bapis-go/community/service/favorite"
	thumbmdl "git.bilibili.co/bapis/bapis-go/community/service/thumbup"
)

const (
	_businessLike     = "archive"
	_articleLike      = "article"
	_dynamicLike      = "dynamic"
	_albumLike        = "album"
	_clipLike         = "clip"
	_cheeseLike       = "cheese"
	_likeSourceLike   = "1"
	_likeTripleAction = "video_triplelike"
	_likeTripleScene  = "video_triplelike"
	_likeArcAction    = "like"
	_likeArcScene     = "thumbup_video"
	_coinAddAction    = "video_coin"
	_coinAddScene     = "video_coin"
	_coinToLikeAction = "video_cointolike"
	_coinToLikeScene  = "video_cointolike"
	_shareAddAction   = "video_share"
	_shareAddScene    = "video_share"
)

// Like archive
func (s *Service) Like(c context.Context, aid, mid int64, like int8, riskParams *model.RiskManagement) (res *model.LikeRes, err error) {
	arcReply, err := s.arcGRPC.Arc(c, &arcmdl.ArcRequest{Aid: aid})
	if err != nil {
		log.Error("s.arcGRPC.Arc(%d) error(%v)", aid, err)
		return
	}
	if arcReply == nil || arcReply.Arc == nil || !arcReply.Arc.IsNormal() {
		err = arcecode.ArchiveNotExist
		return
	}
	res = &model.LikeRes{
		IsRisk:      false,
		GaiaResType: model.GaiaResponseType_Default,
	}
	upperID := arcReply.Arc.Author.Mid
	res.UpID = upperID
	riskParams.UpMid = upperID
	riskParams.Action = _likeArcAction
	riskParams.Scene = _likeArcScene
	riskParams.LikeSource = _likeSourceLike
	riskParams.Pubtime = arcReply.Arc.PubDate.Time().Format("2006-01-02 15:04:05")
	riskParams.Title = arcReply.Arc.Title
	riskParams.PlayNum = arcReply.Arc.Stat.View
	riskResult := s.RiskVerifyAndManager(c, riskParams)
	if riskResult != nil {
		res.GaiaResType = riskResult.GaiaResType
		res.IsRisk = riskResult.IsRisk
		res.GaiaData = riskResult.GaiaData
		return res, nil
	}
	_, err = s.thumbupGRPC.Like(c, &thumbmdl.LikeReq{Business: _businessLike, Mid: mid, UpMid: upperID, MessageID: aid, Action: thumbmdl.Action(like), IP: metadata.String(c, metadata.RemoteIP), Platform: "pc"})
	return
}

// LikeTriple like & coin & fav
// nolint:gomnd
func (s *Service) LikeTriple(c context.Context, aid, mid int64, riskParams *model.RiskManagement) (res *model.TripleRes, err error) {
	var (
		arcReply *arcmdl.ArcReply
		ip       = metadata.String(c, metadata.RemoteIP)
	)
	res = &model.TripleRes{
		IsRisk:      false,
		GaiaResType: model.GaiaResponseType_Default,
	}
	maxCoin := int64(1)
	multiply := int64(1)
	if arcReply, err = s.arcGRPC.Arc(c, &arcmdl.ArcRequest{Aid: aid}); err != nil {
		log.Error("s.arcGRPC.Arc(%d) error(%v)", aid, err)
		return
	}
	a := arcReply.Arc
	if !a.IsNormal() {
		err = arcecode.ArchiveNotExist
		return
	}
	if a.Copyright == int32(arcmdl.CopyrightOriginal) {
		maxCoin = 2
		multiply = 2
	}
	res.UpID = a.Author.Mid
	riskParams.UpMid = a.Author.Mid
	riskParams.Action = _likeTripleAction
	riskParams.Scene = _likeTripleScene
	riskParams.Pubtime = arcReply.Arc.PubDate.Time().Format("2006-01-02 15:04:05")
	riskParams.Title = arcReply.Arc.Title
	riskParams.PlayNum = arcReply.Arc.Stat.View
	riskResult := s.RiskVerifyAndManager(c, riskParams)
	if riskResult != nil {
		res.GaiaResType = riskResult.GaiaResType
		res.IsRisk = riskResult.IsRisk
		res.GaiaData = riskResult.GaiaData
		return res, nil
	}
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) (err error) {
		if multiply == 2 {
			if userCoins, e := s.coinGRPC.UserCoins(c, &coinmdl.UserCoinsReq{Mid: mid}); e != nil {
				log.Error("s.coinGRPC.UserCoins error(%v)", e)
			} else if userCoins != nil {
				if userCoins.Count < 1 {
					return
				}
				if userCoins.Count < 2 {
					multiply = 1
				}
			}
		}
		cArg := &coinmdl.AddCoinReq{
			IP:       ip,
			Mid:      mid,
			Upmid:    a.Author.Mid,
			MaxCoin:  maxCoin,
			Aid:      aid,
			Business: model.CoinArcBusiness,
			Number:   multiply,
			Typeid:   a.TypeID,
			PubTime:  int64(a.PubDate),
			Platform: "pc",
		}
		if _, err = s.coinGRPC.AddCoin(c, cArg); err != nil {
			log.Error("s.coinGRPC.AddCoin error(%v)", err)
			err = nil
			if arcUserCoins, e := s.coinGRPC.ItemUserCoins(c, &coinmdl.ItemUserCoinsReq{Mid: mid, Aid: aid, Business: model.CoinArcBusiness}); e != nil {
				log.Error("s.coinGRPC.ItemUserCoins error(%v)", e)
			} else if arcUserCoins != nil && arcUserCoins.Number > 0 {
				res.Coin = true
			}

		} else {
			res.Multiply = multiply
			res.Anticheat = true
			res.Coin = true
		}
		return
	})
	eg.Go(func(ctx context.Context) (err error) {
		var favReq *favgrpc.IsFavoredReply
		if favReq, err = s.favGRPC.IsFavored(context.Background(), &favgrpc.IsFavoredReq{Typ: int32(favmdl.TypeVideo), Mid: mid, Oid: aid}); err != nil {
			log.Error("s.fav.IsFav error(%v)", err)
			err = nil
		} else if favReq.Faved {
			res.Fav = true
			return
		}
		fArg := &favgrpc.AddFavReq{Tp: int32(favmdl.TypeVideo), Mid: mid, Oid: aid, Fid: 0, Platform: "pc"}
		if _, err = s.favGRPC.AddFav(c, fArg); err != nil {
			if ecode.EqualError(favecode.FavVideoExist, err) {
				res.Fav = true
				return
			}
			log.Error("s.fav.Add error(%v)", err)
			err = nil
		} else {
			res.Fav = true
			res.Anticheat = true
		}
		return
	})
	eg.Go(func(ctx context.Context) (err error) {
		if _, err = s.thumbupGRPC.Like(c, &thumbmdl.LikeReq{Business: _businessLike, Mid: mid, UpMid: res.UpID, MessageID: aid, Action: thumbmdl.Action_ACTION_LIKE, IP: ip, Platform: "pc"}); err != nil {
			if ecode.EqualError(mecode.ThumbupDupLikeErr, err) {
				res.Like = true
				return
			}
			log.Error("s.thumbup.Like error(%v)", err)
			err = nil
		} else {
			res.Like = true
			res.Anticheat = true
		}
		return
	})
	if err := eg.Wait(); err != nil {
		log.Error("%+v", err)
	}
	return
}

// HasLike get if has like.
func (s *Service) HasLike(c context.Context, aid, mid int64) (like thumbmdl.State, err error) {
	var (
		data *thumbmdl.HasLikeReply
		ip   = metadata.String(c, metadata.RemoteIP)
	)
	if data, err = s.thumbupGRPC.HasLike(c, &thumbmdl.HasLikeReq{Business: _businessLike, MessageIds: []int64{aid}, Mid: mid, IP: ip}); err != nil {
		log.Error("s.thumbupGRPC.HasLike aid(%d) mid(%d) error(%v)", aid, mid, err)
		return
	}
	if data != nil && data.States != nil {
		if v, ok := data.States[aid]; ok {
			like = v.State
		}
	}
	return
}

//点赞动效查询
func (s *Service) GetMultiLikeAnimation(ctx context.Context, aid int64) (map[int64]*thumbmdl.LikeAnimation, error) {
	req := &thumbmdl.MultiLikeAnimationReq{
		Business:   "archive",
		MessageIds: []int64{aid},
	}
	res, err := s.thumbupGRPC.MultiLikeAnimation(ctx, req)
	if err != nil {
		return nil, err
	}
	return res.LikeAnimation, nil
}
