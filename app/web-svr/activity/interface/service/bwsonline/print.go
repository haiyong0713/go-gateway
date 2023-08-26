package bwsonline

import (
	"context"
	"math/rand"
	"strconv"
	"time"

	"go-common/library/sync/errgroup.v2"

	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/model/bwsonline"
)

func (s *Service) PrintList(ctx context.Context, mid, bid int64) ([]*bwsonline.UserPrint, error) {
	var (
		ids          []int64
		prints       map[int64]*bwsonline.Print
		unlockPieces map[int64][]*bwsonline.UserPiece
		userPieces   map[int64]int64
	)
	eg := errgroup.WithContext(ctx)
	eg.Go(func(ctx context.Context) (err error) {
		ids, err = s.dao.PrintList(ctx, bid)
		if err != nil {
			log.Errorc(ctx, "PrintList PrintList error:%v", err)
		}
		return err
	})
	if mid > 0 {
		eg.Go(func(ctx context.Context) (err error) {
			if unlockPieces, err = s.userPrint(ctx, mid); err != nil {
				log.Errorc(ctx, "PrintList s.userPrint mid:%d error:%v", mid, err)
			}
			return nil
		})
		eg.Go(func(ctx context.Context) (err error) {
			if userPieces, err = s.userPiece(ctx, mid, bid); err != nil {
				log.Errorc(ctx, "PrintList s.userPiece mid:%d error:%v", mid, err)
			}
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	for id := range unlockPieces {
		ids = append(ids, id)
	}
	ids = filterIDs(ids)
	if len(ids) == 0 {
		return []*bwsonline.UserPrint{}, nil
	}
	for i := 0; i < len(ids); i++ {
		for j := i + 1; j < len(ids); j++ {
			if ids[i] < ids[j] {
				ids[i], ids[j] = ids[j], ids[i]
			}
		}
	}
	var err error
	if prints, err = s.dao.PrintByIDs(ctx, ids); err != nil {
		log.Errorc(ctx, "PrintList s.dao.PrintByIDs ids:%v error:%v", ids, err)
		return nil, err
	}

	var list []*bwsonline.UserPrint
	for _, v := range ids {
		item, ok := prints[v]
		if !ok || item == nil {
			continue
		}
		tmp := &bwsonline.UserPrint{Print: item, UnlockCost: []*bwsonline.UserPiece{}}
		if unlockPiece, ok := unlockPieces[v]; ok {
			tmp.Unlocked = 1
			tmp.UnlockCost = unlockPiece
		}
		if pieceNum, ok := userPieces[item.PieceId]; ok && pieceNum > 0 {
			tmp.PieceState = 1
		}
		list = append(list, tmp)
	}
	return list, nil
}

func (s *Service) PrintDetail(ctx context.Context, mid, id, bid int64) (*bwsonline.UserPrintDetail, error) {
	data, err := s.dao.Print(ctx, id)
	if err != nil {
		log.Errorc(ctx, "PrintDetail s.dao.Print id:%d error:%v", id, err)
		return nil, err
	}
	if data == nil {
		return nil, xecode.NothingFound
	}
	res := &bwsonline.UserPrintDetail{
		UserPrint: &bwsonline.UserPrint{
			Print:      data,
			UnlockCost: []*bwsonline.UserPiece{},
		},
		Awards: []*bwsonline.Award{},
	}
	eg := errgroup.WithContext(ctx)
	if data.PackageId > 0 {
		eg.Go(func(ctx context.Context) error {
			res.Awards = func() []*bwsonline.Award {
				packageData, e := s.dao.AwardPackage(ctx, data.PackageId)
				if e != nil || packageData == nil {
					log.Errorc(ctx, "PrintDetail s.dao.AwardPackage id:%d error:%v", data.PackageId, err)
					return []*bwsonline.Award{}
				}
				if len(packageData.AwardIds) > 0 {
					awards, e := s.dao.AwardByIDs(ctx, packageData.AwardIds)
					if e != nil {
						log.Errorc(ctx, "PrintDetail s.dao.AwardByIDs ids:%d error:%v", packageData.AwardIds, err)
						return []*bwsonline.Award{}
					}
					var (
						awardData []*bwsonline.Award
						dressIDs  []int64
					)
					for _, awardID := range packageData.AwardIds {
						if v, ok := awards[awardID]; ok && v != nil {
							if v.TypeId == bwsonline.AwardTypeDress {
								dressID, _ := strconv.ParseInt(v.Token, 10, 64)
								if dressID > 0 {
									dressIDs = append(dressIDs, dressID)
								}
							}
							awardData = append(awardData, &bwsonline.Award{
								ID:     v.ID,
								Title:  v.Title,
								Intro:  v.Intro,
								Image:  v.Image,
								TypeId: v.TypeId,
								Num:    v.Num,
								Token:  v.Token,
								Expire: v.Expire,
								Ctime:  v.Ctime,
								Mtime:  v.Mtime,
							})
						}
					}
					if len(dressIDs) > 0 {
						var dressMap map[int64]*bwsonline.Dress
						dressMap, err = s.dao.DressByIDs(ctx, dressIDs)
						if err != nil {
							log.Errorc(ctx, "PrintDetail s.dao.DressByIDs dressIDs:%v,err:%v", dressIDs, err)
						}
						if len(dressMap) > 0 {
							for _, v := range awardData {
								if v.TypeId == bwsonline.AwardTypeDress {
									dressID, _ := strconv.ParseInt(v.Token, 10, 64)
									if dress, ok := dressMap[dressID]; ok && dress != nil {
										v.Title = dress.Title
										v.Image = dress.Image
									}
								}
							}
						}
					}
					return awardData
				}
				return []*bwsonline.Award{}
			}()
			return nil
		})
	}
	if mid > 0 {
		eg.Go(func(ctx context.Context) error {
			if userPrintIDs, e := s.dao.UserPrint(ctx, mid); e != nil {
				log.Errorc(ctx, "PrintDetail s.dao.UserPrint mid:%d error:%v", mid, e)
			} else if _, ok := userPrintIDs[id]; ok {
				res.Unlocked = 1
			}
			return nil
		})
		eg.Go(func(ctx context.Context) error {
			userPieces, e := s.userPiece(ctx, mid, bid)
			if e != nil {
				log.Errorc(ctx, "PrintDetail s.userPiece mid:%d error:%v", mid, e)
				return nil
			}
			if num, ok := userPieces[data.PieceId]; ok && num > 0 {
				res.PieceState = 1
			}
			return nil
		})
	}
	if err = eg.Wait(); err != nil {
		return nil, err
	}
	return res, nil
}

func (s *Service) PrintUnlock(ctx context.Context, mid, id int64, counts []int64, bid int64) error {
	// 检查解锁花费碎片数
	var copperCount, silverCount, goldCount int64
	allCount := func() int64 {
		var cnt int64
		for i, v := range counts {
			switch i {
			case 0:
				copperCount = v
			case 1:
				silverCount = v
			case 2:
				goldCount = v
			default:
			}
			cnt += v
		}
		return cnt
	}()
	if allCount < s.c.BwsOnline.PrintPieceMinCount || allCount > s.c.BwsOnline.PrintPieceMaxCount {
		return ecode.BwsOnlinePrintPieceNumErr
	}
	// 检查图鉴是否存在
	detail, err := s.dao.Print(ctx, id)
	if err != nil {
		log.Errorc(ctx, "PrintUnlock PrintDetail id:%d error:%v", id, err)
		return err
	}

	if detail.Bid != bid {
		log.Errorc(ctx, "PrintUnlock PrintDetail %v", detail)
		return ecode.BwsOnlinePrintUnlockFail
	}

	// 检查碎片数量
	userPiece, err := s.userPiece(ctx, mid, bid)
	if err != nil {
		log.Errorc(ctx, "PrintUnlock userPiece mid:%d error:%v", mid, err)
		return err
	}
	var (
		decrPiece    []*bwsonline.UserPiece
		hasSpecPiece bool
	)
	for pieceID, cnt := range userPiece {
		if pieceID == 0 {
			continue
		}
		var decrCnt int64
		switch pieceID {
		case _pieceCopperID:
			if copperCount > cnt {
				return xecode.Errorf(ecode.BwsOnlinePrintPieceLow, ecode.BwsOnlinePrintPieceLow.Message(), "铜")
			}
			decrCnt = copperCount
		case _pieceSilverID:
			if silverCount > cnt {
				return xecode.Errorf(ecode.BwsOnlinePrintPieceLow, ecode.BwsOnlinePrintPieceLow.Message(), "银")
			}
			decrCnt = silverCount
		case _pieceGoldID:
			if goldCount > cnt {
				return xecode.Errorf(ecode.BwsOnlinePrintPieceLow, ecode.BwsOnlinePrintPieceLow.Message(), "金")
			}
			decrCnt = goldCount
		case detail.PieceId:
			hasSpecPiece = true
		default:
		}
		if decrCnt > 0 {
			decrPiece = append(decrPiece, &bwsonline.UserPiece{Pid: pieceID, Num: decrCnt})
		}
	}
	// 检查解锁状态
	userPrint, err := s.dao.UserPrint(ctx, mid)
	if err != nil {
		return err
	}
	if _, ok := userPrint[id]; ok {
		return ecode.BwsOnlinePrintHadUnlock
	}
	// 概率
	piecePoint := 3*copperCount + 5*silverCount + 8*goldCount
	printPoint := bwsonline.PrintValue[detail.Level]
	if printPoint == 0 {
		return xecode.RequestErr
	}
	unlockPercent := int64(float64(piecePoint) / float64(printPoint) * float64(100))
	if hasSpecPiece {
		unlockPercent += 20
	}
	percentRes := func() bool {
		if unlockPercent >= 100 {
			return true
		}
		if rnd := rand.New(rand.NewSource(time.Now().UnixNano())).Int63n(100); rnd <= unlockPercent {
			return true
		}
		return false
	}()
	// 扣碎片
	if _, err = s.dao.DecrUserPiece(ctx, mid, decrPiece, bid); err != nil {
		log.Errorc(ctx, "PrintUnlock DecrUserPiece mid:%d counts:%v error:%v", mid, counts, err)
		return ecode.BwsOnlinePrintUnlockFail
	}
	s.cache.Do(ctx, func(ctx context.Context) {
		s.dao.DelCacheUserPiece(ctx, mid, bid)
	})
	// 加使用日志
	batchID := s.md5(strconv.FormatInt(mid, 10) + "_" + strconv.FormatInt(time.Now().UnixNano(), 10))
	if _, err = s.dao.PieceAddUseLog(ctx, mid, decrPiece, batchID); err != nil {
		log.Errorc(ctx, "PrintUnlock PieceAddUseLog mid:%d counts:%v batchID:%s error:%v", mid, counts, batchID, err)
		return err
	}
	if !percentRes {
		return ecode.BwsOnlinePrintUnlockFail
	}
	// 解锁图鉴
	if _, err = s.dao.AddUserPrint(ctx, mid, id, bwsonline.UserPrintHad, batchID); err != nil {
		log.Errorc(ctx, "PrintUnlock AddUserPrint mid:%d id:%d error:%v", mid, id, err)
		return err
	}
	s.cache.Do(ctx, func(ctx context.Context) {
		s.dao.DelCacheUserPrint(ctx, mid)
	})
	if detail.PackageId > 0 {
		// 发礼包
		s.cache.Do(ctx, func(ctx context.Context) {
			data, err := s.dao.AwardPackage(ctx, detail.PackageId)
			if err != nil {
				log.Errorc(ctx, "PrintUnlock AwardPackage:%d error:%v", id, err)
				return
			}
			if data == nil || data.TypeId != bwsonline.PackageTypePrint || len(data.AwardIds) == 0 {
				log.Warnc(ctx, "PrintUnlock data(%+v) not support", data)
				return
			}
			awards, err := s.dao.AwardByIDs(ctx, data.AwardIds)
			if err != nil {
				log.Errorc(ctx, "PrintUnlock AwardByIDs awardIDs(%v) error:%v", data.AwardIds, err)
				return
			}
			var (
				addUserCurrency int64
				addDressID      = make([]int64, 0, len(awards))
				awardIDs        = make([]int64, 0, len(awards))
				addAwardIDs     = make([]int64, 0, len(awards))
			)
			for _, v := range awards {
				if v == nil {
					continue
				}
				awardIDs = append(awardIDs, v.ID)
				if v.TypeId == bwsonline.AwardTypeDress {
					id, _ = strconv.ParseInt(v.Token, 10, 64)
					addDressID = append(addDressID, id)
					continue
				}
				if v.TypeId == bwsonline.AwardTypeCurrency {
					addUserCurrency = v.Num
					continue
				}
				addAwardIDs = append(addAwardIDs, v.ID)
			}
			if _, err = s.dao.AddUserAwardPackage(ctx, mid, detail.PackageId, awardIDs); err != nil {
				log.Errorc(ctx, "PrintUnlock AddUserAwardPackage mid:%d id:%d error:%v", mid, detail.PackageId, err)
				return
			}
			if len(addDressID) > 0 {
				if _, err = s.dao.DressAdd(ctx, mid, addDressID); err != nil {
					log.Errorc(ctx, "PrintUnlock DressAdd mid:%d dressID:%v error:%v", mid, addDressID, err)
				}
			}
			if addUserCurrency > 0 {
				if err = s.upUserCurrency(ctx, mid, bwsonline.CurrTypeCoin, bwsonline.CurrAddTypeNormal, addUserCurrency, bid); err != nil {
					log.Errorc(ctx, "PrintUnlock upUserCurrency mid:%d error:%v", mid, err)
				}
			}
			if len(addAwardIDs) > 0 {
				if _, err = s.dao.AddUserAward(ctx, mid, addAwardIDs, bid); err != nil {
					log.Errorc(ctx, "PrintUnlock AddUserAward mid:%d error:%v", mid, err)
				}
			}
			s.dao.DelCacheUserPackage(ctx, mid)
			s.dao.DelCacheUserDress(ctx, mid)
			s.dao.DelCacheUserCurrency(ctx, mid, bid)
			s.dao.DelCacheUserAward(ctx, mid, bid)
		})
	}
	return nil
}

func (s *Service) userPrint(ctx context.Context, mid int64) (map[int64][]*bwsonline.UserPiece, error) {
	userPrintIDs, err := s.dao.UserPrint(ctx, mid)
	if err != nil {
		return nil, err
	}
	var batchIDs []string
	for _, v := range userPrintIDs {
		batchIDs = append(batchIDs, v)
	}
	var pieceLog map[string]map[int64]int64
	if len(batchIDs) > 0 {
		pieceLog = func() map[string]map[int64]int64 {
			data, err := s.dao.PieceUsedLog(ctx, mid, batchIDs)
			if err != nil {
				log.Errorc(ctx, "userPrint PieceUsedLog mid:%d batchIDs:%v error:%v", mid, batchIDs, err)
				return nil
			}
			return data
		}()
	}
	res := make(map[int64][]*bwsonline.UserPiece)
	for printID, v := range userPrintIDs {
		if batchLogs, ok := pieceLog[v]; ok {
			for pieceID, num := range batchLogs {
				res[printID] = append(res[printID], &bwsonline.UserPiece{
					Pid:   pieceID,
					Num:   num,
					Level: pieceLevels[pieceID],
				})
			}
		}
	}
	return res, nil
}
