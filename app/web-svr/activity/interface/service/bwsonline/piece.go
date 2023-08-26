package bwsonline

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/model/bwsonline"
)

const (
	_pieceCopperID = 1
	_pieceSilverID = 2
	_pieceGoldID   = 3
	_pieceFindUse  = 1
)

var pieceLevels = map[int64]int64{
	_pieceCopperID: 1,
	_pieceSilverID: 2,
	_pieceGoldID:   3,
}

func (s *Service) PieceFind(ctx context.Context, mid, bid int64) (int64, int64, error) {
	if check, err := s.likeDao.RsSetNX(ctx, fmt.Sprintf("piece_find_%d", mid), 1); err != nil || !check {
		log.Warnc(ctx, "PieceFreeFind mid:%d to fast err:%v", mid, err)
		return 0, 0, ecode.ActivityRapid
	}
	// 检查能量值
	curr, err := s.dao.UserCurrency(ctx, mid, bid)
	if err != nil {
		log.Errorc(ctx, "PieceFind UserCurrency:%d error:%v", mid, err)
		return 0, 0, err
	}
	if curr[bwsonline.CurrTypeEnergy] < 1 {
		return 0, 0, ecode.BwsOnlineEnergyLow
	}
	level, num, err := s.pieceFind(ctx, mid, 0, true, bid)
	if err != nil {
		log.Errorc(ctx, "PieceFind pieceFind mid:%d error:%v", mid, err)
		return 0, 0, err
	}
	return level, num, nil
}

func (s *Service) SendSpecialPiece(ctx context.Context, mid, id int64, token string, bid int64) error {
	piece, err := s.dao.Piece(ctx, id)
	if err != nil {
		log.Errorc(ctx, "SendSpecialPiece Piece id:%d error:%v", id, err)
		return err
	}
	if piece == nil || piece.ID != id || piece.Token != token {
		log.Warnc(ctx, "SendSpecialPiece piece:%+v not match id:%d token:%s", piece, id, token)
		return xecode.RequestErr
	}
	if _, _, err = s.pieceFind(ctx, mid, id, false, bid); err != nil {
		log.Errorc(ctx, "PieceFind pieceFind mid:%d id:%d error:%v", mid, id, err)
		return err
	}
	return nil
}

func (s *Service) PieceFreeFind(ctx context.Context, mid, fromMid, bid int64) (int64, int64, error) {
	if check, err := s.likeDao.RsSetNX(ctx, fmt.Sprintf("piece_free_find_%d", mid), 1); err != nil || !check {
		log.Warnc(ctx, "PieceFreeFind mid:%d to fast err:%v", mid, err)
		return 0, 0, ecode.ActivityRapid
	}
	today := todayDate()
	usedTimes, err := s.dao.UsedTimes(ctx, mid, today)
	if err != nil {
		log.Errorc(ctx, "PieceFreeFind UsedTimes mid:%d date:%d error:%v", mid, today, err)
		return 0, 0, err
	}
	if _, ok := usedTimes[bwsonline.UsedTimeTypeLed]; ok {
		return 0, 0, ecode.BwsOnlineTimeUsed
	}
	if _, err = s.dao.AddUsedTimes(ctx, mid, bwsonline.UsedTimeTypeLed, today); err != nil {
		log.Errorc(ctx, "PieceFreeFind AddUsedTimes mid:%d date:%d error:%v", mid, today, err)
		return 0, 0, ecode.BwsOnlineTimeUsed
	}
	level, num, err := s.pieceFind(ctx, mid, 0, false, bid)
	if err != nil {
		log.Errorc(ctx, "PieceFreeFind led pieceFind mid:%d error:%v", mid, err)
		return 0, 0, err
	}
	s.cache.Do(ctx, func(ctx context.Context) {
		fromUsedTimes, fromErr := s.dao.UsedTimes(ctx, fromMid, today)
		if fromErr != nil {
			log.Errorc(ctx, "PieceFreeFind fromMid:%d error:%v", fromMid, fromErr)
			return
		}
		if cnt, ok := fromUsedTimes[bwsonline.UsedTimeTypeShare]; !ok || cnt <= 0 {
			log.Warnc(ctx, "PieceFreeFind fromMid mid:%d time used", mid)
			return
		}
		if _, addErr := s.dao.AddUsedTimes(ctx, fromMid, bwsonline.UsedTimeTypeShare, today); addErr != nil {
			log.Errorc(ctx, "PieceFreeFind AddUsedTimes mid:%d date:%d error:%v", mid, today, err)
			return
		}
		if _, _, fromErr = s.pieceFind(ctx, fromMid, 0, false, bid); err != nil {
			log.Errorc(ctx, "PieceFreeFind share pieceFind mid:%d error:%v", fromMid, fromErr)
		}
	})
	return level, num, nil
}

func (s *Service) pieceFind(ctx context.Context, mid, id int64, needEnergy bool, bid int64) (level int64, num int64, err error) {
	num = 1
	if id == 0 {
		rnd := rand.New(rand.NewSource(time.Now().UnixNano())).Int63n(100)
		for _, v := range s.c.BwsOnline.PiecePR {
			if rnd < v.PR {
				id = v.ID
				num = v.Num
				break
			}
		}
	}
	// 先减能量
	if needEnergy {
		if err := s.upUserCurrency(ctx, mid, bwsonline.CurrTypeEnergy, 0, -_pieceFindUse, bid); err != nil {
			return 0, 0, err
		}
	}
	if _, err := s.dao.PieceAddLog(ctx, mid, id, num, bid); err != nil {
		return 0, 0, err
	}
	if _, err := s.dao.AddUserPiece(ctx, mid, id, num, bid); err != nil {
		return 0, 0, err
	}
	s.cache.Do(ctx, func(ctx context.Context) {
		s.dao.DelCacheUserPiece(ctx, mid, bid)
	})
	return pieceLevels[id], num, nil
}

func (s *Service) userPiece(ctx context.Context, mid, bid int64) (map[int64]int64, error) {
	data, err := s.dao.UserPiece(ctx, mid, bid)
	if err != nil {
		return nil, err
	}
	res := make(map[int64]int64, len(data))
	for _, v := range data {
		res[v.Pid] = v.Num
	}
	return res, nil
}
