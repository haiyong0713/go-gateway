package bwsonline

import (
	"context"
	"strconv"
	"time"

	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/model/bwsonline"
	"go-gateway/app/web-svr/activity/interface/model/like"

	"go-common/library/sync/errgroup.v2"

	accapi "git.bilibili.co/bapis/bapis-go/account/service"
)

func (s *Service) Main(ctx context.Context, mid int64, bid int64) (*bwsonline.Main, error) {
	egGroup := errgroup.WithContext(ctx)
	var accInfo *accapi.InfoReply
	egGroup.Go(func(ctx context.Context) (err error) {
		if accInfo, err = s.accClient.Info3(ctx, &accapi.MidReq{Mid: mid}); err != nil {
			log.Errorc(ctx, "Main s.accClient.Info3 mid:%d error:%v", mid, err)
		}
		return nil
	})
	var userCurrency map[int64]int64
	egGroup.Go(func(ctx context.Context) (err error) {
		if userCurrency, err = s.userCurrency(ctx, mid, bid); err != nil {
			log.Errorc(ctx, "Main s.userCurrency mid:%d error:%v", mid, err)
		}
		return nil
	})
	var userPiece []*bwsonline.UserPiece
	egGroup.Go(func(ctx context.Context) (err error) {
		var pieceData map[int64]int64
		if pieceData, err = s.userPiece(ctx, mid, bid); err != nil {
			log.Errorc(ctx, "Main userPiece mid:%d error:%v", mid, err)
			return nil
		}
		for i, v := range pieceData {
			switch i {
			case _pieceCopperID, _pieceSilverID, _pieceGoldID:
				userPiece = append(userPiece, &bwsonline.UserPiece{Pid: i, Num: v, Level: pieceLevels[i]})
			default:
				continue
			}
		}
		if len(userPiece) == 0 {
			userPiece = []*bwsonline.UserPiece{}
		}
		return nil
	})
	var userDress []*bwsonline.Dress
	egGroup.Go(func(ctx context.Context) (err error) {
		if userDress, err = s.userDress(ctx, mid, true); err != nil {
			log.Errorc(ctx, "Main s.userDress mid:%d error:%v", mid, err)
		}
		return nil
	})
	var isActivated int
	egGroup.Go(func(ctx context.Context) (err error) {
		var hasReserve *like.HasReserve
		if hasReserve, err = s.likeDao.ReserveOnly(ctx, s.c.BwsOnline.BuyTicketSid, mid); err != nil {
			log.Errorc(ctx, "Main s.likeDao.ReserveOnly sid:%d mid:%d error:%v", s.c.BwsOnline.BuyTicketSid, mid, err)
			return nil
		}
		if hasReserve != nil && hasReserve.ID > 0 && hasReserve.State == 1 {
			isActivated = 1
		}
		return nil
	})
	if err := egGroup.Wait(); err != nil {
		log.Errorc(ctx, "Main egGroup.Wait error:%v", err)
	}
	return &bwsonline.Main{
		Mid:         mid,
		Name:        accInfo.GetInfo().GetName(),
		Face:        accInfo.GetInfo().GetFace(),
		Energy:      userCurrency[bwsonline.CurrTypeEnergy],
		Currency:    userCurrency[bwsonline.CurrTypeCoin],
		Piece:       userPiece,
		Dress:       userDress,
		IsActivated: isActivated,
	}, nil
}

func todayDate() int64 {
	dayStr := time.Now().Format("20060102")
	res, err := strconv.ParseInt(dayStr, 10, 64)
	if err != nil {
		return 0
	}
	return res
}

func hourInt(ts int64) int64 {
	hourStr := time.Unix(ts, 0).Format("2006010215")
	res, err := strconv.ParseInt(hourStr, 10, 64)
	if err != nil {
		return 0
	}
	return res
}
