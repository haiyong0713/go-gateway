package bws

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"math"
	"sort"
	"strconv"
	"sync"
	"time"

	"go-common/library/log"
	xtime "go-common/library/time"
	xecode "go-gateway/app/web-svr/activity/ecode"
	bwsmdl "go-gateway/app/web-svr/activity/interface/model/bws"

	accapi "git.bilibili.co/bapis/bapis-go/account/service"

	"go-common/library/sync/errgroup.v2"
)

const _specialKeyFmt = "%d_%d_VMVFh6"

var (
	// 所有的活动类型与string对应关系
	_allPointsType = map[int64]string{
		bwsmdl.DpType:             bwsmdl.Dp,
		bwsmdl.GameType:           bwsmdl.Game,
		bwsmdl.ClockinType:        bwsmdl.Clockin,
		bwsmdl.EggType:            bwsmdl.Egg,
		bwsmdl.HideClockinType:    bwsmdl.HideClockin,
		bwsmdl.ChargeType:         bwsmdl.Recharge,
		bwsmdl.SignType:           bwsmdl.Sign,
		bwsmdl.SpecialClockinType: bwsmdl.SpecialClockin,
	}
	_signOpen    = 0
	_signClose   = -1
	_signNotOpen = -2
)

// SignInfo [ctime,stime].
func (s *Service) SignInfo(c context.Context, pid int64) (rs *bwsmdl.SignInfoReply, err error) {
	var (
		signIDs  []int64
		signInfo map[int64]*bwsmdl.PointSign
		open     *bwsmdl.PointSign
		sTime    = int64(2145888000) //取一个比较大的数
		eTime    = int64(0)
		t        = time.Now()
		dt       = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local).Unix()
		tomorrow = dt + 86400
	)
	// 获取pid下所有的签到任务
	if signIDs, err = s.dao.Signs(c, pid); err != nil {
		log.Error("s.dao.Signs(%d) error(%v)", pid, err)
		return
	}
	//获取当前时间段对应的任务
	if signInfo, err = s.dao.BwsSign(c, signIDs); err != nil {
		log.Error("s.dao.BwsSign(%d) error(%v)", pid, err)
		return
	}
	// 不支持跨天进行的活动 默认未开始
	rs = &bwsmdl.SignInfoReply{State: int32(_signNotOpen)}
	for _, v := range signInfo {
		if v.Stime >= dt && v.Stime <= tomorrow {
			if sTime > v.Stime {
				sTime = v.Stime
			}
			if eTime < v.Etime {
				eTime = v.Etime
			}
			if v.Stime <= t.Unix() && v.Etime >= t.Unix() {
				open = v
				break
			}
		}
	}
	if open != nil {
		// 获取当前任务是否已经完成
		rs.State = open.State
		rs.Stime = open.Stime
		rs.Etime = open.Etime
		rs.SignPoints = open.SignPoints
		leftPoints := open.Points - open.ProvidePoints
		if leftPoints > 0 {
			rs.SurplusPoints = leftPoints
		}
		rs.ID = open.ID
		rs.Points = open.Points
	} else {
		if t.Unix() < sTime {
			rs.State = int32(_signNotOpen)
		} else if t.Unix() > eTime {
			rs.State = int32(_signClose)
		}
	}
	return
}

// Points points list
func (s *Service) Points(c context.Context, p *bwsmdl.ParamPoints) (rs map[string][]*bwsmdl.SinglePoints, err error) {
	var (
		points   *bwsmdl.Points
		allReply map[string][]*bwsmdl.SinglePoints
		levels   map[int64][]*bwsmdl.RechargeAward
	)
	if points, err = s.dao.PointsByBid(c, p.Bid); err != nil || points == nil || len(points.Points) == 0 {
		log.Error("s.dao.Points error(%v)", err)
		err = xecode.ActivityPointFail
		return
	}
	allReply = make(map[string][]*bwsmdl.SinglePoints, len(_allPointsType))
	for _, point := range points.Points {
		if _, ok := _allPointsType[point.LockType]; !ok {
			continue
		}
		tpPoint := &bwsmdl.SinglePoints{Point: point}
		allReply[_allPointsType[point.LockType]] = append(allReply[_allPointsType[point.LockType]], tpPoint)
	}
	rs = make(map[string][]*bwsmdl.SinglePoints)
	switch p.Tp {
	case bwsmdl.AllType:
		rs = allReply
	default:
		if _, ok := _allPointsType[p.Tp]; ok {
			if _, k := allReply[_allPointsType[p.Tp]]; k {
				rs[_allPointsType[p.Tp]] = allReply[_allPointsType[p.Tp]]
			}
		}
	}
	eg := errgroup.WithContext(c)
	// 是否含有签到信息
	if _, ok := rs[bwsmdl.Sign]; ok {
		for _, val := range rs[bwsmdl.Sign] {
			temp := val
			eg.Go(func(ctx context.Context) (e error) {
				if temp.Sign, e = s.SignInfo(ctx, temp.ID); e != nil {
					log.Error("s.SignInfo(%d) error(%v)", temp.ID, e)
					e = nil
				}
				return
			})
		}
	}
	// 是否有充能信息
	if _, k := rs[bwsmdl.Recharge]; k && len(rs[bwsmdl.Recharge]) > 0 {
		eg.Go(func(ctx context.Context) (e error) {
			if levels, e = s.pointsLevelAward(ctx, p.Bid); e != nil {
				log.Error("s.pointsLevelAward(%d) error(%v)", p.Bid, e)
				e = nil
				return
			}
			for _, val := range rs[bwsmdl.Recharge] {
				if _, ok := levels[val.ID]; !ok {
					continue
				}
				for _, lve := range levels[val.ID] {
					// 判断是否解锁奖品
					if lve.Points <= val.Unlocked {
						lve.Unlock = 1
					} else {
						lve.Unlock = 0
					}
				}
				val.Recharge = append(val.Recharge, levels[val.ID]...)
			}
			return
		})
	}
	eg.Wait()
	return
}

// Point point
func (s *Service) Point(c context.Context, p *bwsmdl.ParamID) (rs *bwsmdl.PointReply, err error) {
	var (
		points map[int64]*bwsmdl.Point
		levels map[int64][]*bwsmdl.RechargeAward
	)
	if p.ID <= 0 {
		return
	}
	if points, err = s.dao.BwsPoints(c, []int64{p.ID}); err != nil {
		log.Errorc(c, "s.dao.BwsPoints error(%v)", err)
		err = xecode.ActivityPointFail
		return
	}
	if _, ok := points[p.ID]; !ok || points[p.ID].Bid != p.Bid {
		err = xecode.ActivityIDNotExists
		return
	}
	rs = &bwsmdl.PointReply{Point: points[p.ID]}
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) error {
		rs.UnlockTotal, _ = s.UnlockTotal(ctx, p.ID)
		return nil
	})
	if points[p.ID].LockType == bwsmdl.SignType {
		eg.Go(func(ctx context.Context) (e error) {
			if rs.Sign, err = s.SignInfo(ctx, p.ID); err != nil {
				log.Errorc(c, "s.SignInfo(%d) error(%v)", p.ID, err)
				err = nil
			}
			return
		})
	} else if points[p.ID].LockType == bwsmdl.ChargeType {
		eg.Go(func(ctx context.Context) (e error) {
			if levels, e = s.pointsLevelAward(ctx, p.Bid); e != nil {
				log.Errorc(c, "s.pointsLevelAward(%d) error(%v)", p.Bid, e)
				e = nil
				return
			}
			if _, ok := levels[p.ID]; !ok {
				return
			}
			for _, lve := range levels[p.ID] {
				// 判断是否解锁奖品
				if lve.Points <= points[p.ID].Unlocked {
					lve.Unlock = 1
				} else {
					lve.Unlock = 0
				}
			}
			rs.Recharge = append(rs.Recharge, levels[p.ID]...)
			return
		})
	}
	eg.Wait()
	return
}

// UnlockTotal .
func (s *Service) UnlockTotal(c context.Context, pid int64) (total int64, err error) {
	if total, err = s.dao.CacheUnlock(c, pid); err != nil {
		return
	}
	if total == -1 {
		if total, err = s.dao.RawCountUnlock(c, pid); err != nil {
			return
		}
		s.dao.AddUnlock(c, pid, total)
	}
	return
}

// NewUnlock .
func (s *Service) NewUnlock(c context.Context, owner int64, arg *bwsmdl.ParamUnlock) (addAchieves []*bwsmdl.Achievement, err error) {
	var (
		pointReply               map[int64]*bwsmdl.Point
		point                    *bwsmdl.Point
		userPoints               []*bwsmdl.UserPointDetail
		hp, lockPoint, incrPoint int64
		signReply                *bwsmdl.SignInfoReply
		signID, userPointID      int64
		isOk                     bool
	)
	if arg.Key == "" {
		if arg.Key, err = s.midToKey(c, arg.Bid, arg.Mid); err != nil {
			return
		}
	} else {
		// 校验key是否有效,获取绑定的mid
		if arg.Mid, _, err = s.keyToMid(c, arg.Bid, arg.Key); err != nil {
			return
		}
	}
	// 防刷 redis 无错误且写缓存没有成功 操作过于频繁
	if isOk, err = s.dao.RequestLimit(c, arg.Bid, arg.Key, fmt.Sprintf("newUnlock_%d", arg.Pid), 1); err == nil && !isOk {
		err = xecode.ActivityFrequence
		return
	}
	// 获取任务信息
	if pointReply, err = s.dao.BwsPoints(c, []int64{arg.Pid}); err != nil {
		return
	}
	if _, ok := pointReply[arg.Pid]; !ok || pointReply[arg.Pid].Bid != arg.Bid {
		err = xecode.ActivityIDNotExists
		return
	}
	point = pointReply[arg.Pid]
	if point.Ower != owner && !s.isAdmin(owner) {
		err = xecode.ActivityNotOwner
		return
	}
	// 游戏类型
	if point.LockType == bwsmdl.GameType {
		if arg.GameResult != bwsmdl.GameResWin && arg.GameResult != bwsmdl.GameResFail {
			err = xecode.ActivityGameResult
			return
		}
	}
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) (e error) {
		// 获取locktype下已经完成的log
		userPoints, e = s.userLockPoints(ctx, arg.Bid, point.LockType, arg.Key)
		return
	})
	eg.Go(func(ctx context.Context) (e error) {
		// 获取用户当前的point分
		if hp, e = s.dao.UserHp(ctx, arg.Bid, arg.Key); e != nil {
			log.Errorc(c, " s.dao.UserHp(%d,%s) error(%v)", arg.Bid, arg.Key, e)
		}
		return
	})
	if point.LockType == bwsmdl.SignType {
		eg.Go(func(ctx context.Context) (e error) {
			// 获取当前签到cd
			if signReply, e = s.SignInfo(ctx, point.ID); e != nil {
				log.Errorc(c, "s.SignInfo(%d) error(%v)", point.ID, e)
			}
			return
		})
	}
	if err = eg.Wait(); err != nil {
		return
	}
	switch point.LockType {
	case bwsmdl.GameType:
		lockPoint = point.Unlocked
		if arg.GameResult == bwsmdl.GameResFail {
			lockPoint = point.LoseUnlocked
		}
	case bwsmdl.ChargeType:
		if hp <= 0 {
			err = xecode.ActivityLackHp
			return
		}
		lockPoint = -1
		if arg.Recharge == bwsmdl.RechargeHalf {
			lockPoint = 0 - int64(math.Ceil(float64(hp)/2))
		} else if arg.Recharge == bwsmdl.RechargeAll {
			lockPoint = 0 - hp
		}
	case bwsmdl.SignType:
		// 判断是否到达签到时间点
		if signReply == nil || signReply.State != int32(_signOpen) {
			err = xecode.ActivitySignNotOpen
			return
		}
		// 剩余point 小于单次获取的能量，无法获取
		if signReply.SurplusPoints < signReply.SignPoints {
			err = xecode.ActivitySignNotEnough
			return
		}
		signID = signReply.ID
		lockPoint = signReply.SignPoints
	default:
		// 位置不要变更
		lockPoint = point.Unlocked
	}
	if hp+lockPoint < 0 {
		err = xecode.ActivityLackHp
		return
	}
	for _, v := range userPoints {
		if (point.LockType != bwsmdl.GameType && point.LockType != bwsmdl.ChargeType && point.LockType != bwsmdl.SignType) && v.Pid == point.ID {
			err = xecode.ActivityHasUnlock
			return
		}
		// 签到点重复判断
		if point.LockType == bwsmdl.SignType && v.Pid == point.ID {
			if int64(v.Ctime) >= signReply.Stime && int64(v.Ctime) < signReply.Etime {
				err = xecode.ActivityHasUnlock
				return
			}
		}
	}
	egTwo := errgroup.WithContext(c)
	egTwo.Go(func(ctx context.Context) (e error) {
		// 1.增加用户point分 2.解锁locktype log  3.充能类型时,新增充能点分数 4签到点能量减少
		userPointID, incrPoint, e = s.addUserLockPoint(ctx, arg.Bid, arg.Pid, point.LockType, lockPoint, signID, arg.Key)
		return
	})
	egTwo.Go(func(ctx context.Context) (e error) {
		// 当前任务参与人数加1
		if e = s.dao.IncrUnlock(ctx, arg.Pid, 1); e != nil {
			log.Errorc(c, "s.dao.IncrUnlock(%d) error(%v)", arg.Pid, e)
			e = nil
		}
		return
	})
	if err = egTwo.Wait(); err != nil {
		return
	}
	userPoints = append(userPoints, &bwsmdl.UserPointDetail{
		UserPoint:    &bwsmdl.UserPoint{ID: userPointID, Pid: arg.Pid, Points: lockPoint, Ctime: xtime.Time(time.Now().Unix())},
		Unlocked:     point.Unlocked,
		LoseUnlocked: point.LoseUnlocked,
		LockType:     point.LockType,
	})
	if addAchieves, err = s.unLockedAchieves(c, arg, point, userPoints, incrPoint); err != nil {
		return
	}
	if len(addAchieves) > 0 {
		for _, v := range addAchieves {
			tpval := v
			s.cache.Do(c, func(c context.Context) {
				s.addAchieve(c, arg.Mid, tpval, arg.Key)
			})
		}
	}
	return
}

// IsNewBws .
func (s *Service) IsNewBws(c context.Context, bid int64) bool {
	return bid >= s.c.Bws.NewBid
}

// Unlock unlock point.
func (s *Service) Unlock(c context.Context, owner int64, arg *bwsmdl.ParamUnlock) (err error) {
	var (
		pointReply    map[int64]*bwsmdl.Point
		point         *bwsmdl.Point
		userPoints    []*bwsmdl.UserPointDetail
		userAchieves  *bwsmdl.CategoryAchieve
		achieves      *bwsmdl.Achievements
		unLockCnt, hp int64
		addAchieve    *bwsmdl.Achievement
		lockAchieves  []*bwsmdl.Achievement
	)
	if arg.Key == "" {
		//special bid create key
		if arg.Bid == s.c.Bws.SpecialBid {
			var user *bwsmdl.Users
			if user, err = s.dao.UsersMid(c, arg.Bid, arg.Mid); err != nil {
				err = xecode.ActivityMidFail
				return
			}
			if user != nil && user.Key != "" {
				// has bind
				arg.Key = user.Key
			} else {
				// first unlock
				arg.Key = specialBwsKey(arg.Bid, arg.Mid)
				if _, err = s.Binding(c, arg.Mid, &bwsmdl.ParamBinding{Bid: arg.Bid, Key: arg.Key}); err != nil {
					return
				}
			}
		} else {
			if arg.Key, err = s.midToKey(c, arg.Bid, arg.Mid); err != nil {
				return
			}
		}
	}
	if _, ok := pointReply[arg.Pid]; !ok || pointReply[arg.Pid].Bid != arg.Bid {
		err = xecode.ActivityIDNotExists
		return
	}
	point = pointReply[arg.Pid]
	if point.Ower != owner && !s.isAdmin(owner) {
		err = xecode.ActivityNotOwner
		return
	}
	if point.LockType == bwsmdl.GameType {
		if arg.GameResult != bwsmdl.GameResWin && arg.GameResult != bwsmdl.GameResFail {
			err = xecode.ActivityGameResult
			return
		}
	}
	if userPoints, err = s.userPoints(c, arg.Bid, arg.Key); err != nil {
		return
	}
	userPidMap := make(map[int64]int64, len(userPoints))
	for _, v := range userPoints {
		if point.LockType != bwsmdl.GameType && v.Pid == point.ID {
			err = xecode.ActivityHasUnlock
			return
		}
		if _, ok := userPidMap[v.Pid]; !ok && v.LockType == point.LockType {
			if v.LockType == bwsmdl.GameType {
				if v.Points == v.Unlocked {
					unLockCnt++
					userPidMap[v.Pid] = v.Pid
				}
			} else {
				unLockCnt++
				userPidMap[v.Pid] = v.Pid
			}
		}
		hp += v.Points
	}
	lockPoint := point.Unlocked
	if point.LockType == bwsmdl.GameType && arg.GameResult == bwsmdl.GameResFail {
		lockPoint = point.LoseUnlocked
	}
	if hp+lockPoint < 0 {
		err = xecode.ActivityLackHp
		return
	}
	if err = s.addUserPoint(c, arg.Bid, arg.Pid, lockPoint, arg.Key); err != nil {
		return
	}
	if err = s.dao.IncrUnlock(c, arg.Pid, 1); err != nil {
		return
	}
	if _, active := s.achieveBids[arg.Bid]; active {
		if userAchieves, err = s.userAchieves(c, arg.Bid, arg.Key); err != nil {
			return
		}
		if achieves, err = s.dao.Achievements(c, arg.Bid); err != nil || achieves == nil || len(achieves.Achievements) == 0 {
			log.Errorc(c, "s.dao.Achievements error(%v)", err)
			err = xecode.ActivityAchieveFail
			return
		}
		for _, v := range achieves.Achievements {
			if point.LockType == v.LockType {
				lockAchieves = append(lockAchieves, v)
			}
		}
		if len(lockAchieves) > 0 {
			sort.Slice(lockAchieves, func(i, j int) bool { return lockAchieves[i].Unlock > lockAchieves[j].Unlock })
			if point.LockType == bwsmdl.GameType {
				if arg.GameResult == bwsmdl.GameResWin {
					unLockCnt++
				}
			} else {
				unLockCnt++
			}
			for _, ach := range lockAchieves {
				if unLockCnt >= ach.Unlock {
					addAchieve = ach
					break
				}
			}
		}
		if addAchieve != nil {
			for _, v := range userAchieves.Achievements {
				if v.Aid == addAchieve.ID {
					return
				}
			}
			s.addAchieve(c, arg.Mid, addAchieve, arg.Key)
		}
	}
	return
}

// BatchUserLockPoints .
func (s *Service) BatchUserLockPoints(c context.Context, bid int64, lockType []int64, key string) (res map[int64][]*bwsmdl.UserPointDetail, err error) {
	var (
		usPoints map[int64][]*bwsmdl.UserPoint
		points   map[int64]*bwsmdl.Point
		ids      []int64
	)
	// 特定类别下已完成任务
	if len(lockType) == 0 {
		return
	}
	if usPoints, err = s.dao.BatchUserLockPoints(c, bid, lockType, key); err != nil {
		err = xecode.ActivityUserPointFail
		return
	}
	if len(usPoints) == 0 {
		return
	}
	ids = make([]int64, 0)
	for _, v := range usPoints {
		for _, pVal := range v {
			ids = append(ids, pVal.Pid)
		}
	}
	if points, err = s.dao.BwsPoints(c, ids); err != nil || len(points) == 0 {
		log.Errorc(c, "s.dao.Points error(%v)", err)
		err = xecode.ActivityPointFail
		return
	}
	res = make(map[int64][]*bwsmdl.UserPointDetail)
	for lockKey, pVal := range usPoints {
		if len(pVal) == 0 {
			continue
		}
		for _, v := range pVal {
			detail := &bwsmdl.UserPointDetail{UserPoint: v}
			if point, ok := points[v.Pid]; ok {
				detail.Name = point.Name
				detail.Icon = point.Icon
				detail.Fid = point.Fid
				detail.Image = point.Image
				detail.Unlocked = point.Unlocked
				detail.LockType = point.LockType
				detail.Dic = point.Dic
				detail.Rule = point.Rule
				detail.Bid = point.Bid
			}
			res[lockKey] = append(res[lockKey], detail)
		}
	}
	return
}

// userLockPoints 获取特定任务下已完成列表.
func (s *Service) userLockPoints(c context.Context, bid, lockType int64, key string) (res []*bwsmdl.UserPointDetail, err error) {
	var (
		usPoints []*bwsmdl.UserPoint
		points   map[int64]*bwsmdl.Point
		ids      []int64
	)
	// 特定类别下已完成任务
	if usPoints, err = s.dao.UserLockPoints(c, bid, lockType, key); err != nil {
		err = xecode.ActivityUserPointFail
		return
	}
	if len(usPoints) == 0 {
		return
	}
	ids = make([]int64, 0, len(usPoints))
	for _, v := range usPoints {
		ids = append(ids, v.Pid)
	}
	if points, err = s.dao.BwsPoints(c, ids); err != nil || len(points) == 0 {
		log.Errorc(c, "s.dao.Points error(%v)", err)
		err = xecode.ActivityPointFail
		return
	}
	for _, v := range usPoints {
		detail := &bwsmdl.UserPointDetail{UserPoint: v}
		if point, ok := points[v.Pid]; ok {
			detail.Name = point.Name
			detail.Icon = point.Icon
			detail.Fid = point.Fid
			detail.Image = point.Image
			detail.Unlocked = point.Unlocked
			detail.LockType = point.LockType
			detail.Dic = point.Dic
			detail.Rule = point.Rule
			detail.Bid = point.Bid
		}
		res = append(res, detail)
	}
	return
}

func (s *Service) userLockPointsDay(ctx context.Context, bid, lockType int64, key, day string) (res []*bwsmdl.UserPointDetail, err error) {
	var (
		usPoints []*bwsmdl.UserPoint
		points   map[int64]*bwsmdl.Point
		ids      []int64
	)
	// 特定类别下已完成任务
	if usPoints, err = s.dao.UserLockPointsDay(ctx, bid, lockType, key, day); err != nil {
		err = xecode.ActivityUserPointFail
		return
	}
	if len(usPoints) == 0 {
		return
	}
	ids = make([]int64, 0, len(usPoints))
	for _, v := range usPoints {
		ids = append(ids, v.Pid)
	}
	if points, err = s.dao.BwsPoints(ctx, ids); err != nil || len(points) == 0 {
		log.Errorc(ctx, "s.dao.Points error(%v)", err)
		err = xecode.ActivityPointFail
		return
	}
	for _, v := range usPoints {
		detail := &bwsmdl.UserPointDetail{UserPoint: v}
		if point, ok := points[v.Pid]; ok {
			detail.Name = point.Name
			detail.Icon = point.Icon
			detail.Fid = point.Fid
			detail.Image = point.Image
			detail.Unlocked = point.Unlocked
			detail.LockType = point.LockType
			detail.Dic = point.Dic
			detail.Rule = point.Rule
			detail.Bid = point.Bid
		}
		res = append(res, detail)
	}
	return
}

// UserPoints 用户打卡情况
func (s *Service) UserPoints(c context.Context, bid int64, mid int64, pointType int64, day string) (res []*bwsmdl.UserPointDetail, err error) {
	res = make([]*bwsmdl.UserPointDetail, 0)
	userToken, err := s.midToKey(c, bid, mid)
	if err != nil {
		return nil, err
	}

	point, err := s.pointsUser(c, bid, userToken, day)
	if err != nil {
		log.Errorc(c, "s.userPoints err(%v)", err)
		return
	}
	if len(point) > 0 {
		for _, v := range point {
			if pointType == 0 || v.LockType == pointType {
				if pointType == bwsmdl.ChargeType {
					if mid == v.Owner || s.isAdmin(mid) {
						res = append(res, v)
					}
					continue
				}
				if pointType != bwsmdl.ChargeType {
					res = append(res, v)
				}
			}
		}
	}
	return
}

// UserPoints 用户打卡情况
func (s *Service) UserPointAdmin(c context.Context, bid int64, mid int64, pointType int64) (res []*bwsmdl.Point, err error) {
	res = make([]*bwsmdl.Point, 0)
	var (
		points *bwsmdl.Points
	)
	if points, err = s.dao.PointsByBid(c, bid); err != nil || points == nil || len(points.Points) == 0 {
		log.Errorc(c, "s.dao.Points error(%v)", err)
		err = xecode.ActivityPointFail
		return
	}
	if len(points.Points) > 0 {
		for _, v := range points.Points {
			if pointType == 0 || v.LockType == pointType {
				if mid == v.Ower || s.isAdmin(mid) {
					res = append(res, v)
				}
				continue
			}
		}
	}
	return
}

func (s *Service) pointsUser(c context.Context, bid int64, key string, day string) (res []*bwsmdl.UserPointDetail, err error) {
	var (
		usPoints []*bwsmdl.UserPoint
		points   *bwsmdl.Points
	)
	res = make([]*bwsmdl.UserPointDetail, 0)
	if points, err = s.dao.PointsByBid(c, bid); err != nil || points == nil || len(points.Points) == 0 {
		log.Errorc(c, "s.dao.Points error(%v)", err)
		err = xecode.ActivityPointFail
		return
	}
	if usPoints, err = s.dao.UserPoints(c, bid, key); err != nil {
		err = xecode.ActivityUserPointFail
		return
	}

	pointsMap := make(map[int64]*bwsmdl.UserPoint)
	for _, v := range usPoints {
		pointsMap[v.Pid] = v
	}
	for _, v := range points.Points {
		detail := &bwsmdl.UserPointDetail{
			Name:     v.Name,
			ID:       v.ID,
			Icon:     v.Icon,
			Fid:      v.Fid,
			Image:    v.Image,
			Unlocked: v.Unlocked,
			LockType: v.LockType,
			Dic:      v.Dic,
			Rule:     v.Rule,
			Bid:      v.Bid,
			Owner:    v.Ower,
		}
		if point, ok := pointsMap[v.ID]; ok {
			ctime := time.Unix(int64(point.Ctime), 0)
			dayStr := ctime.Format("20060102")
			if dayStr == day {
				point.IsPoint = true
				detail.UserPoint = point
			}
		}
		res = append(res, detail)
	}
	return
}

func (s *Service) userPoints(c context.Context, bid int64, key string) (res []*bwsmdl.UserPointDetail, err error) {
	var (
		usPoints []*bwsmdl.UserPoint
		points   *bwsmdl.Points
	)
	if usPoints, err = s.dao.UserPoints(c, bid, key); err != nil {
		err = xecode.ActivityUserPointFail
		return
	}
	if len(usPoints) == 0 {
		return
	}
	if points, err = s.dao.PointsByBid(c, bid); err != nil || points == nil || len(points.Points) == 0 {
		log.Errorc(c, "s.dao.Points error(%v)", err)
		err = xecode.ActivityPointFail
		return
	}
	pointsMap := make(map[int64]*bwsmdl.Point, len(points.Points))
	for _, v := range points.Points {
		pointsMap[v.ID] = v
	}
	for _, v := range usPoints {
		detail := &bwsmdl.UserPointDetail{UserPoint: v}
		if point, ok := pointsMap[v.Pid]; ok {
			detail.Name = point.Name
			detail.Icon = point.Icon
			detail.Fid = point.Fid
			detail.Image = point.Image
			detail.Unlocked = point.Unlocked
			detail.LockType = point.LockType
			detail.Dic = point.Dic
			detail.Rule = point.Rule
			detail.Bid = point.Bid
		}
		res = append(res, detail)
	}
	return
}

// addUserLockPoint .
func (s *Service) addUserLockPoint(c context.Context, bid, pid, lockType, points, signID int64, key string) (usPtID, incrPoint int64, err error) {
	if usPtID, err = s.dao.AddUserPoint(c, bid, pid, lockType, points, key); err != nil {
		err = xecode.ActivityUnlockFail
		return
	}
	// 更新hp分数
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) (e error) {
		if e = s.dao.IncrUserHp(ctx, bid, points, key); e != nil {
			log.Errorc(c, "s.dao.IncrUserHp(%d,%s,%d) err error(%v)", bid, key, points, e)
		}
		return
	})
	if points > 0 {
		eg.Go(func(ctx context.Context) (e error) {
			if incrPoint, e = s.dao.IncrUserPoints(ctx, bid, points, key); e != nil {
				log.Errorc(c, "s.dao.IncrUserPoints(%d,%s,%d) err error(%v)", bid, key, points, e)
				e = nil
			}
			return
		})
	}
	//更新缓存
	eg.Go(func(ctx context.Context) (e error) {
		e = s.dao.AppendUserLockPointsCache(ctx, bid, lockType, key, &bwsmdl.UserPoint{ID: usPtID, Pid: pid, Points: points, Ctime: xtime.Time(time.Now().Unix())})
		return
	})
	//更新充值点的分数
	if lockType == bwsmdl.ChargeType {
		eg.Go(func(ctx context.Context) (e error) {
			return s.dao.RechargePoint(ctx, pid, bwsmdl.ChargeType, points)
		})
	}
	// 签到点已使用point增加
	if lockType == bwsmdl.SignType {
		eg.Go(func(ctx context.Context) (e error) {
			return s.dao.IncrSignPoint(ctx, signID)
		})
	}
	err = eg.Wait()
	return
}

func (s *Service) addUserPoint(c context.Context, bid, pid, points int64, key string) (err error) {
	var usPtID int64
	if usPtID, err = s.dao.AddUserPoint(c, bid, pid, 0, points, key); err != nil {
		err = xecode.ActivityUnlockFail
		return
	}
	err = s.dao.AppendUserPointsCache(c, bid, key, &bwsmdl.UserPoint{ID: usPtID, Pid: pid, Points: points, Ctime: xtime.Time(time.Now().Unix())})
	return
}

func specialBwsKey(bid, mid int64) string {
	hasher := md5.New()
	key := fmt.Sprintf(_specialKeyFmt, bid, mid)
	hasher.Write([]byte(key))
	return hex.EncodeToString(hasher.Sum(nil))
}

// AchieveRank .
func (s *Service) AchieveRank(c context.Context, bid, loginMid int64, ps, ty int) (rs *bwsmdl.AchieveRank, err error) {
	var (
		mids          []int64
		pointList     map[string]int64
		compositeList map[int64]int64
		users         map[int64]*bwsmdl.Users
		ukeys         []string
		keyToMid      map[int64]string
		cards         map[int64]*accapi.Card
		achieveRank   int
		loginMids     []int64
	)
	if mids, err = s.dao.CacheAchievesRank(c, bid, ps, ty); err != nil {
		log.Errorc(c, "s.dao.AchievesRank(%d,%d,%d) error(%v)", bid, ps, ty, err)
		return
	}
	loginMids = mids
	if loginMid > 0 {
		loginMids = append(loginMids, loginMid)
	}
	eg := errgroup.WithContext(c)
	if ty != bwsmdl.CompositeRankType { // 单场次排行才需要获取key信息
		eg.Go(func(ctx context.Context) (e error) {
			if users, e = s.dao.UsersMids(ctx, bid, loginMids); e != nil {
				log.Error(" s.dao.UsersKeys(%d,%v) error(%v)", bid, mids, e)
			}
			return
		})
	}
	eg.Go(func(ctx context.Context) (e error) {
		if cards, e = s.accCards(ctx, mids); e != nil {
			log.Error("s.accCards(%v) error(%v)", mids, e)
		}
		return
	})
	achieveRank = bwsmdl.DefaultRank
	if loginMid > 0 {
		eg.Go(func(errCtx context.Context) (e error) {
			// 获取用户成就排行
			if achieveRank, e = s.dao.GetAchieveRank(errCtx, bid, loginMid, ty); e != nil {
				log.Errorc(c, "s.dao.AchievesPoint(%d,%d) error(%v)", bid, loginMid, e)
				e = nil
			}
			return
		})
	}
	if err = eg.Wait(); err != nil {
		return
	}
	if ty == bwsmdl.CompositeRankType {
		if compositeList, err = s.dao.CompositeAchievesPoint(c, loginMids); err != nil {
			log.Errorc(c, "s.dao.CompositeAchievesPoint(%v) error(%v)", mids, err)
			err = nil
		}
	} else {
		keyToMid = make(map[int64]string)
		for _, val := range users {
			if val.Key == "" {
				continue
			}
			ukeys = append(ukeys, val.Key)
			keyToMid[val.Mid] = val.Key
		}
		if pointList, err = s.dao.AchievesPoint(c, bid, ukeys); err != nil {
			log.Errorc(c, "s.dao.AchievesPoint(%d,%v) error(%v)", bid, ukeys, err)
			err = nil
		}
	}
	rs = &bwsmdl.AchieveRank{}
	rankJ := 0
	for i, v := range mids {
		if i >= ps {
			continue
		}
		if ty != bwsmdl.CompositeRankType {
			if _, ok := keyToMid[v]; !ok {
				continue
			}
		}
		if _, k := cards[v]; !k {
			continue
		}
		rankJ++
		temp := &bwsmdl.UserInfo{
			Mid:  cards[v].Mid,
			Name: cards[v].Name,
			Face: cards[v].Face,
			//Key:         keyToMid[v],
			AchieveRank: rankJ,
		}
		if ty == bwsmdl.CompositeRankType {
			if _, ook := compositeList[v]; ook {
				temp.AchievePoint = compositeList[v]
			}
		} else {
			temp.Key = keyToMid[v]
			if _, ook := pointList[keyToMid[v]]; ook {
				temp.AchievePoint = pointList[keyToMid[v]]
			}
		}
		rs.List = append(rs.List, temp)
	}
	if achieveRank != bwsmdl.DefaultRank {
		achieveRank += 1
	}
	rs.SelfRank = achieveRank
	if loginMid > 0 {
		if ty == bwsmdl.CompositeRankType {
			if _, mok := compositeList[loginMid]; mok {
				rs.SelfPoint = compositeList[loginMid]
			}
		} else {
			if _, lok := keyToMid[loginMid]; lok {
				if _, mok := pointList[keyToMid[loginMid]]; mok {
					rs.SelfPoint = pointList[keyToMid[loginMid]]
				}
			}
		}
	}
	return
}

// RechargeAward 获取解锁奖品和未解锁奖品.
func (s *Service) RechargeAward(c context.Context, p *bwsmdl.ParamRechargeAward) (list *bwsmdl.RechargeAwardReply, err error) {
	var (
		levels map[int64][]*bwsmdl.RechargeAward
		points map[int64]*bwsmdl.Point
		pids   []int64
	)
	if levels, err = s.pointsLevelAward(c, p.Bid); err != nil {
		log.Error("s.pointsLevelAward(%d) error(%v)", p.Bid, err)
		return
	}
	for k := range levels {
		pids = append(pids, k)
	}
	if len(pids) == 0 {
		return
	}
	if points, err = s.dao.BwsPoints(c, pids); err != nil {
		log.Errorc(c, "s.dao.BwsPoints(%v) error(%v)", pids, err)
		return
	}
	list = &bwsmdl.RechargeAwardReply{}
	unlocks := make(map[int64]*bwsmdl.Unlocks)
	for pid, val := range levels {
		if _, ok := points[pid]; !ok {
			continue
		}
		tp := &bwsmdl.Unlocks{}
		for _, aw := range val {
			//放入unlocked
			tempUnlocked := make([]*bwsmdl.RechargeReply, 0)
			for _, adVal := range aw.Awards {
				tempUnlocked = append(tempUnlocked, &bwsmdl.RechargeReply{
					Level:  aw.Level,
					Name:   adVal.Name,
					Icon:   adVal.Icon,
					Amount: adVal.Amount,
					ID:     adVal.ID,
				})
			}
			// 已解锁
			if aw.Points <= points[pid].Unlocked {
				tp.Unlock = append(tp.Unlock, tempUnlocked...)
			} else {
				tp.NotUnlock = append(tp.NotUnlock, tempUnlocked...)
			}
		}
		unlocks[pid] = tp
	}
	for k, val := range unlocks {
		if _, ok := points[k]; !ok {
			continue
		}
		val.Point = points[k]
		list.Recharge = append(list.Recharge, val)
	}
	// 排序输出，否者每次输出顺序不定
	sort.Slice(list.Recharge, func(i int, j int) bool {
		return list.Recharge[i].ID < list.Recharge[j].ID
	})
	return
}

// pointsLevelAward 拼接bid下所有的充能奖品信息.
func (s *Service) pointsLevelAward(c context.Context, bid int64) (levels map[int64][]*bwsmdl.RechargeAward, err error) {
	var (
		plIDs   []int64
		plInfos map[int64]*bwsmdl.PointsLevel
		awards  map[int64][]*bwsmdl.PointsAward
		muxu    sync.Mutex
	)
	// 获取pid下所有的level ids
	if plIDs, err = s.dao.PointLevels(c, bid); err != nil {
		log.Errorc(c, "d.PointLevels(%d) error(%v)", bid, err)
		return
	}
	if len(plIDs) == 0 {
		return
	}
	eg := errgroup.WithCancel(c)
	eg.Go(func(ctx context.Context) (e error) {
		// 获取等级详细信息
		if plInfos, e = s.dao.RechargeLevels(ctx, plIDs); e != nil {
			log.Errorc(c, "d.RechargeLevels(%v) error(%v)", plIDs, e)
		}
		return
	})
	awards = make(map[int64][]*bwsmdl.PointsAward)
	for _, v := range plIDs {
		plid := v
		eg.Go(func(ctx context.Context) (e error) {
			var tmp []*bwsmdl.PointsAward
			// 获取等级下奖品信息
			if tmp, e = s.pointAwardByPlID(ctx, plid); e != nil {
				log.Errorc(c, "s.pointAwardByPlID(%d) error(%v)", plid, e)
				e = nil
				return
			}
			if len(tmp) == 0 {
				return
			}
			muxu.Lock()
			awards[plid] = tmp
			muxu.Unlock()
			return
		})
	}
	if err = eg.Wait(); err != nil {
		return
	}
	levels = make(map[int64][]*bwsmdl.RechargeAward)
	for _, v := range plInfos {
		if _, ok := levels[v.Pid]; !ok {
			levels[v.Pid] = make([]*bwsmdl.RechargeAward, 0)
		}
		temp := &bwsmdl.RechargeAward{PointsLevel: v}
		if _, k := awards[v.ID]; k {
			temp.Awards = awards[v.ID]
		}
		levels[v.Pid] = append(levels[v.Pid], temp)
	}
	for _, val := range levels {
		sort.Slice(val, func(i, j int) bool {
			return val[i].Level < val[j].Level
		})
	}
	return
}

// pointAwardByPlID 获取各个等级下面的奖品信息.
func (s *Service) pointAwardByPlID(c context.Context, plID int64) (awards []*bwsmdl.PointsAward, err error) {
	var (
		awardsIDs []int64
		list      map[int64]*bwsmdl.PointsAward
	)
	if awardsIDs, err = s.dao.PointsAward(c, plID); err != nil {
		log.Error("s.dao.PointsAward(%d) error(%v)", plID, err)
		return
	}
	if len(awardsIDs) == 0 {
		return
	}
	if list, err = s.dao.RechargeAwards(c, awardsIDs); err != nil {
		log.Errorc(c, " s.dao.RechargeAwards(%d) error(%v)", plID, err)
		return
	}
	awards = make([]*bwsmdl.PointsAward, 0, len(list))
	for _, val := range list {
		awards = append(awards, val)
	}
	return
}

// Fields .
func (s *Service) Fields(c context.Context, bid int64) (rs *bwsmdl.FieldsReply, err error) {
	var (
		list *bwsmdl.ActFields
	)
	// 缓存1min
	if list, err = s.dao.ActFields(c, bid); err != nil || list == nil {
		log.Errorc(c, "s.dao.ActFields(%d) error(%v)", bid, err)
		return
	}
	rs = &bwsmdl.FieldsReply{}
	rs.Fields = make(map[int64]*bwsmdl.ActField)
	for _, v := range list.ActField {
		rs.Fields[v.ID] = v
	}
	return
}

func (s *Service) unLockedAchieves(c context.Context, arg *bwsmdl.ParamUnlock, point *bwsmdl.Point, userPoints []*bwsmdl.UserPointDetail, incrPoint int64) (addAchieves []*bwsmdl.Achievement, err error) {
	var (
		achieves     []*bwsmdl.Achievement
		typeAchieves []*bwsmdl.Achievement
		userAchieves []*bwsmdl.UserAchieveDetail
	)
	if _, active := s.achieveBids[point.Bid]; !active {
		return
	}
	group := errgroup.WithContext(c)
	group.Go(func(ctx context.Context) (e error) {
		var cateAchieve *bwsmdl.CategoryAchieve
		if cateAchieve, e = s.userAchieves(ctx, arg.Bid, arg.Key); e != nil {
			e = xecode.ActivityUserAchieveFail
			return
		}
		userAchieves = cateAchieve.Achievements
		return
	})
	group.Go(func(ctx context.Context) (e error) {
		var (
			achieve *bwsmdl.Achievements
		)
		if achieve, e = s.dao.Achievements(ctx, arg.Bid); e != nil || achieve == nil || len(achieve.Achievements) == 0 {
			log.Errorc(c, "s.dao.Achievements error(%v)", e)
			e = xecode.ActivityAchieveFail
			return
		}
		achieves = achieve.Achievements
		return
	})
	if err = group.Wait(); err != nil {
		return
	}
	if len(achieves) == 0 {
		return
	}
	userAchieveMap := make(map[int64]struct{}, len(userAchieves))
	for _, v := range userAchieves {
		userAchieveMap[v.Aid] = struct{}{}
	}
	// find type achievement
	achieveType, ok := bwsmdl.PointAchieve[point.LockType]
	for _, v := range achieves {
		if (ok && achieveType == v.LockType) || v.LockType == bwsmdl.AchieveIncrPointType || v.LockType == bwsmdl.AchieveOther {
			typeAchieves = append(typeAchieves, v)
		}
	}
	if len(typeAchieves) > 0 {
		var (
			typeUnLockCnt, pointUnLockCnt, gamePointWin, gameContinueWin, chargeHp int64
			currGameSuee                                                           bool
		)
		sort.Slice(typeAchieves, func(i, j int) bool { return typeAchieves[i].Unlock > typeAchieves[j].Unlock })
		userPointMap := make(map[int64]struct{})
		for _, v := range userPoints {
			if v.Pid == point.ID {
				pointUnLockCnt++
			}
			userPointMap[v.Pid] = struct{}{}
		}
		typeUnLockCnt = int64(len(userPointMap))
		switch point.LockType {
		case bwsmdl.ChargeType:
			for _, v := range userPoints {
				if v.Pid == point.ID {
					chargeHp += v.UserPoint.Points
				}
			}
		case bwsmdl.GameType:
			// 按时间倒序排游戏游玩历史
			sort.Slice(userPoints, func(i, j int) bool { return userPoints[i].ID > userPoints[j].ID })
			if len(userPoints) > 0 {
				if userPoints[0].Points == point.Unlocked {
					currGameSuee = true
				}
				gamePointMap := make(map[int64]struct{})
				for i, v := range userPoints {
					if i == 0 {
						gamePointMap[v.Pid] = struct{}{}
					} else {
						if v.UserPoint.Points == v.Unlocked {
							if _, ok := gamePointMap[v.Pid]; !ok {
								gameContinueWin++
							}
						} else {
							break
						}
						gamePointMap[v.Pid] = struct{}{}
					}
				}
				// 当前游戏成功才判断历史总游戏成功数
				if currGameSuee {
					gamePointMap = make(map[int64]struct{})
					for _, v := range userPoints {
						if v.UserPoint.Points == v.Unlocked {
							if _, ok := gamePointMap[v.Pid]; !ok {
								gamePointWin++
								gamePointMap[v.Pid] = struct{}{}
							}
						}
					}
				}
			}
		default:
			typeUnLockCnt = int64(len(userPointMap))
		}
		for _, ach := range typeAchieves {
			switch ach.LockType {
			case bwsmdl.AchieveChargeType:
				switch ach.ExtraType {
				case bwsmdl.ExtraChargeCnt:
					if typeUnLockCnt >= ach.Unlock {
						if _, ok := userAchieveMap[ach.ID]; !ok {
							addAchieves = append(addAchieves, ach)
						}
					}
				case bwsmdl.ExtraChargeHp:
					chargeHp = int64(math.Abs(float64(chargeHp)))
					if chargeHp >= ach.Unlock {
						if _, ok := userAchieveMap[ach.ID]; !ok {
							addAchieves = append(addAchieves, ach)
						}
					}
				}
			case bwsmdl.AchieveSignType:
				if pointUnLockCnt >= ach.Unlock {
					if _, ok := userAchieveMap[ach.ID]; !ok {
						addAchieves = append(addAchieves, ach)
					}
				}
			case bwsmdl.AchieveGameType:
				switch ach.ExtraType {
				case bwsmdl.ExtraGameFirstFail:
					if !currGameSuee {
						if _, ok := userAchieveMap[ach.ID]; !ok {
							addAchieves = append(addAchieves, ach)
						}
					}
				case bwsmdl.ExtraGameFirstSuee:
					if currGameSuee {
						if _, ok := userAchieveMap[ach.ID]; !ok {
							addAchieves = append(addAchieves, ach)
						}
					}
				case bwsmdl.ExtraGameContinueSuee:
					if currGameSuee && gameContinueWin+1 >= ach.Unlock {
						if _, ok := userAchieveMap[ach.ID]; !ok {
							addAchieves = append(addAchieves, ach)
						}
					}
				case bwsmdl.ExtraGameContinueFail:
					if !currGameSuee && gameContinueWin+1 >= ach.Unlock {
						if _, ok := userAchieveMap[ach.ID]; !ok {
							addAchieves = append(addAchieves, ach)
						}
					}
				case bwsmdl.ExtraGameContinuePlay:
					if pointUnLockCnt >= ach.Unlock {
						if _, ok := userAchieveMap[ach.ID]; !ok {
							addAchieves = append(addAchieves, ach)
						}
					}
				case bwsmdl.ExtraGameSuee:
					if gamePointWin >= ach.Unlock {
						if _, ok := userAchieveMap[ach.ID]; !ok {
							addAchieves = append(addAchieves, ach)
						}
					}
				}
			case bwsmdl.AchieveIncrPointType:
				if incrPoint >= ach.Unlock {
					if _, ok := userAchieveMap[ach.ID]; !ok {
						addAchieves = append(addAchieves, ach)
					}
				}
			case bwsmdl.AchieveOther:
				// 轻视频和特殊打卡点
				if pVal, pok := s.c.Bws.SpecAcheives[strconv.FormatInt(point.ID, 10)]; pok && ach.ID == pVal {
					if _, ok := userAchieveMap[ach.ID]; !ok {
						addAchieves = append(addAchieves, ach)
					}
				}
			default:
				if typeUnLockCnt >= ach.Unlock {
					if _, ok := userAchieveMap[ach.ID]; !ok {
						addAchieves = append(addAchieves, ach)
					}
				}
			}
		}
	}
	return
}
