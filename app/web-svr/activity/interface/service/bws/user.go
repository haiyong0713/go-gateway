package bws

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"go-common/library/database/sql"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/trace"
	xtime "go-common/library/time"
	"go-gateway/app/web-svr/activity/ecode"
	bwsmdl "go-gateway/app/web-svr/activity/interface/model/bws"
	"sort"
	"strconv"
	"time"

	"go-gateway/app/web-svr/activity/interface/client"

	accountapi "git.bilibili.co/bapis/bapis-go/account/service"
	tvbwapi "git.bilibili.co/bapis/bapis-go/bw/game/common"

	"go-common/library/sync/errgroup.v2"
)

const (
	// MaxScore ...
	MaxScore = 10000
	// MaxStar
	MaxStar = 10
)

func (s *Service) IsWhiteMid(ctx context.Context, mid int64) bool {
	if _, ok := s.bwsWhiteMid[mid]; ok {
		return true
	}
	return false
}

// KeyToMid ...
func (s *Service) KeyToMid(ctx context.Context, bid int64, key string) (mid int64, err error) {
	mid, _, err = s.keyToMid(ctx, bid, key)
	return
}

// User2020 用户情况
func (s *Service) AdminUser2020(ctx context.Context, bid, loginMid int64, key string) (*bwsmdl.User2020, error) {
	return s.User2020(ctx, bid, loginMid, key)
}

// User2020 用户情况
func (s *Service) User2020(ctx context.Context, bid, loginMid int64, key string) (*bwsmdl.User2020, error) {
	mid, userToken, err := func() (int64, string, error) {
		if key != "" {
			mid, _, err := s.keyToMid(ctx, bid, key)
			if err != nil {
				return 0, "", err
			}
			if mid <= 0 {
				return 0, "", ecode.ActivityNotBind
			}
			return mid, key, nil
		}
		userToken, err := s.midToKey(ctx, bid, loginMid)
		if err != nil {
			return 0, "", err
		}
		return loginMid, userToken, nil
	}()
	if err != nil {
		return nil, err
	}
	accInfo, err := s.accClient.Info3(ctx, &accountapi.MidReq{Mid: mid})
	if err != nil {
		log.Errorc(ctx, "User2020 Info3:%d error:%v", mid, err)
		return nil, err
	}
	user := &bwsmdl.User2020{
		User: &bwsmdl.UserInfo2020{
			Mid:  mid,
			Name: accInfo.GetInfo().GetName(),
			Key:  userToken,
			Face: accInfo.GetInfo().GetFace(),
		},
	}
	eg := errgroup.WithContext(ctx)
	// user task
	var userTask []*bwsmdl.UserTask
	// eg.Go(func(ctx context.Context) (err error) {
	// 	userTask, err = s.userTask(ctx, bid, userToken)
	// 	if err != nil {
	// 		log.Errorc(ctx, "User2020 userTask userToken:%s error:%+v", userToken, err)
	// 		err = nil
	// 	}
	// 	return nil
	// })
	// lottery log
	var lotteryLog, onlineAward []*bwsmdl.LotteryLog
	eg.Go(func(ctx context.Context) (err error) {
		lotteryLog, onlineAward, err = s.userLotteryLog(ctx, userToken)
		if err != nil {
			log.Errorc(ctx, "User2020 userLotteryLog token:%s error:%+v", userToken, err)
		}
		return nil
	})
	// lottery times
	eg.Go(func(ctx context.Context) (err error) {
		_, dayStr := todayDate()
		user.User.Star, user.User.LotteryTimes, err = s.UserLotteryTimes(ctx, bid, mid, dayStr)
		if err != nil {
			log.Errorc(ctx, "User2020 UserLotteryTimes token:%s error:%+v", userToken, err)
		}
		return nil
	})
	if err = eg.Wait(); err != nil {
		log.Errorc(ctx, "User2020 eg.Wait mid:%d error(%+v)", mid, err)
		return user, nil
	}
	if len(userTask) == 0 {
		userTask = []*bwsmdl.UserTask{}
	}
	user.Tasks = userTask
	if len(lotteryLog) == 0 {
		lotteryLog = []*bwsmdl.LotteryLog{}
	}
	user.LotteryLog = lotteryLog
	if len(onlineAward) == 0 {
		onlineAward = []*bwsmdl.LotteryLog{}
	}
	user.OnlineAward = onlineAward
	return user, nil
}

// IsOwner 是否owner
func (s *Service) IsOwner(ctx context.Context, owner int64, pid int64, bid int64) (err error) {
	isAdmin := s.isAdmin(owner)
	// 获取任务信息
	pointReply, err := s.dao.BwsPoints(ctx, []int64{pid})
	if err != nil {
		return err
	}
	point, ok := pointReply[pid]
	if !ok || pointReply[pid].Bid != bid {
		return ecode.ActivityIDNotExists
	}
	if point.Ower == owner || isAdmin {
		return nil
	}
	return ecode.ActivityNotOwner
}

// Unlock2020 打卡
func (s *Service) Unlock2020(ctx context.Context, owner int64, isInternal bool, arg *bwsmdl.ParamUnlock20) error {
	log.Infoc(ctx, "Unlock2020 mid(%d) key(%s)", arg.Mid, arg.Key)
	isAdmin := s.isAdmin(owner)
	if !isInternal && !isAdmin {
		log.Errorc(ctx, "Unlock2020 isInternal(%v) owner(%d) isAdmin(%v)", isInternal, owner, isAdmin)
		return xecode.RequestErr
	}
	userToken := arg.Key
	var err error
	if userToken == "" {
		userToken, err = s.midToKey(ctx, arg.Bid, arg.Mid)
		if err != nil {
			return err
		}
	}
	if arg.Mid == 0 {
		arg.Mid, _, err = s.keyToMid(ctx, arg.Bid, userToken)
		if err != nil {
			return err
		}
	}
	if isOk, err := s.dao.RequestLimit(ctx, arg.Bid, userToken, fmt.Sprintf("Unlock2020_%d", arg.Pid), 1); err == nil && !isOk {
		return ecode.ActivityFrequence
	}
	// 获取任务信息
	pointReply, err := s.dao.BwsPoints(ctx, []int64{arg.Pid})
	if err != nil {
		return err
	}
	point, ok := pointReply[arg.Pid]
	if !ok || pointReply[arg.Pid].Bid != arg.Bid {
		return ecode.ActivityIDNotExists
	}
	if point.LockType != bwsmdl.ClockinType {
		return ecode.ActivityIDNotExists
	}
	if point.Ower != owner && !isAdmin {
		return ecode.ActivityNotOwner
	}
	dayInt, dayStr := todayDate()
	// 获取用户情况
	_, err = s.getUserDetail(ctx, arg.Bid, arg.Mid, userToken, dayStr, false)
	if err != nil {
		log.Errorc(ctx, "s.getUserDetail(%v)", err)
		return err
	}
	// 领取heart
	pidString := strconv.FormatInt(arg.Pid, 10)
	reasonStruct := &bwsmdl.LogReason{
		Reason: bwsmdl.ReasonLockHeart,
		Params: pidString,
	}
	orderNo := fmt.Sprintf("%d_%d_%d_%d", arg.Mid, arg.Bid, arg.Pid, dayInt)
	err = s.UpdateUserDetail(ctx, arg.Mid, arg.Bid, 0, 0, dayStr, point.Unlocked, nil, false, reasonStruct, orderNo, userToken)
	if err != nil {
		log.Errorc(ctx, "s.UpdateUserDetail (%v)", err)
		if xecode.EqualError(ecode.ActivityBwsDuplicateErr, err) {
			return ecode.ActivityHasUnlock
		}
		return ecode.ActivityUnlockFail
	}
	usPtID, err := s.dao.AddUserPoint(ctx, arg.Bid, point.ID, point.LockType, point.Unlocked, userToken)
	if err != nil {
		return ecode.ActivityUnlockFail
	}
	if err = s.dao.AppendUserLockPointsDayCache(ctx, arg.Bid, point.LockType, userToken, dayStr, &bwsmdl.UserPoint{ID: usPtID, Pid: point.ID, Points: point.Unlocked, Ctime: xtime.Time(time.Now().Unix())}); err != nil {
		s.cache.Do(ctx, func(ctx context.Context) {
			s.dao.DelCacheUserLockPointsDay(ctx, arg.Bid, point.LockType, userToken, dayStr)
		})
	}
	if err = s.doPointTask(ctx, userToken, bwsmdl.TaskCateOther, point.ID, dayInt, 1); err != nil {
		log.Errorc(ctx, "Unlock2020 doPointTask userToken:%s pid:%d day:%d error:%v", userToken, point.ID, dayInt, err)
	}
	s.cache.Do(ctx, func(ctx context.Context) {
		err = s.dao.DelCacheUserPoints(ctx, arg.Bid, userToken)
		if err != nil {
			log.Errorc(ctx, "s.dao.DelCacheUserPoints err(%v) token(%s)", err, userToken)
		}
	})
	return nil
}

func (s *Service) getUserDetail(ctx context.Context, bid, mid int64, key string, day string, isVip bool) (*bwsmdl.UserDetail, error) {
	// 获取用户情况
	userDetail, err := s.dao.UserDetail(ctx, bid, mid, day)
	if err != nil {
		log.Errorc(ctx, "s.dao.UserDetail(%d,%d,%v)", bid, mid, day)
		return nil, ecode.ActivityBwsMidErr
	}
	// 如果用户信息为空，插入
	if userDetail == nil || userDetail.Mid == 0 {

		defaultHeart := s.c.Bws.DefaultHeart
		if isVip {
			defaultHeart = s.c.Bws.VipHeart
		}
		userDetailID, err := s.dao.CreateUserDetail(ctx, bid, mid, day, defaultHeart)
		if err != nil {
			log.Errorc(ctx, "s.dao.CreateUserDetail(%d,%d,%v,%d)", bid, mid, day, defaultHeart)
			return nil, ecode.ActivityBwsMidErr
		}
		// 赠送快拍券
		state := bwsmdl.AwardStateInit
		if _, err = s.dao.AddUserAward(ctx, key, s.c.Bws.StockAwardID2, state); err != nil {
			log.Errorc(ctx, "Lottery2020 AddUserAward usarderToken:%s award:%d error:%v", key, s.c.Bws.StockAwardID2, err)
			return nil, err
		}

		userDetail = &bwsmdl.UserDetail{Id: userDetailID, Mid: mid, Bid: bid, BwsDate: day, Heart: defaultHeart}
		s.cache.Do(ctx, func(ctx context.Context) {
			s.dao.DelCacheUserDetail(ctx, bid, mid, day)
			retry(func() error {
				return s.dao.DelCacheUserAward(ctx, key)
			})
		})

	}

	return userDetail, nil
}

// UserLotteryTimes ...
func (s *Service) UserLotteryTimes(ctx context.Context, bid, mid int64, day string) (int64, int64, error) {
	userDetail, err := s.dao.UserDetail(ctx, bid, mid, day)
	if err != nil {
		log.Errorc(ctx, "s.dao.UserDetail(%d,%d,%v)", bid, mid, day)
		return 0, 0, ecode.ActivityBwsMidErr
	}
	return userDetail.Star, userDetail.Star - userDetail.LotteryUsed, nil
}

// getRankUpdateLimitTime 获得排行榜更新的时间limit
func (s *Service) getRankUpdateLimitTime(ctx context.Context) int64 {
	t1 := time.Now().Year()  //年
	t2 := time.Now().Month() //月
	t3 := time.Now().Day()   //日
	return time.Date(t1, t2, t3, s.c.Bws.RankStopHour, 0, 0, 0, time.Local).Unix()
}

// catchUpsCanGetStar 抓到的ups，可以获取的星星
func (s *Service) catchUpsCanGetStar(ctx context.Context, ups int64) int64 {

	if ups >= s.c.Bws.BwsUpsCatchNeedStar[bwsmdl.UpsStarOneMin] && ups <= s.c.Bws.BwsUpsCatchNeedStar[bwsmdl.UpsStarOneMax] {
		return 1
	}
	if ups >= s.c.Bws.BwsUpsCatchNeedStar[bwsmdl.UpsStarTwoMin] && ups <= s.c.Bws.BwsUpsCatchNeedStar[bwsmdl.UpsStarTwoMax] {
		return 2
	}
	if ups >= s.c.Bws.BwsUpsCatchNeedStar[bwsmdl.UpsStarThreeMin] {
		return 3
	}
	return 0
}

// UserRankList 用户排行榜列表
func (s *Service) UserRankList(ctx context.Context, bid int64, date string) (res *bwsmdl.RankRes, err error) {
	updateTime := s.getRankUpdateLimitTime(ctx)
	now := time.Now().Unix()
	list, err := s.dao.CacheUserRank(ctx, bid, date, s.c.Bws.RankTop50-1)
	dataList := make([]*bwsmdl.Account, 0)
	res = new(bwsmdl.RankRes)
	res.List = dataList
	_, dayStr := todayDate()

	if err != nil {
		return res, err
	}
	mids := make([]int64, 0)
	for _, v := range list {
		mids = append(mids, v.Mid)
	}

	eg := errgroup.WithContext(ctx)
	// user task
	var userDetails map[int64]*bwsmdl.UserDetail
	var accountDetail = &accountapi.InfosReply{}
	var userScoreDetail = make(map[int64]*bwsmdl.MidScore)
	if len(mids) == 0 {
		return res, nil
	}
	eg.Go(func(ctx context.Context) (err error) {

		userDetails, err = s.dao.UserDetails(ctx, mids, bid, date)
		if err != nil {
			log.Errorc(ctx, "s.dao.UserDetails (%v)", err)
			return err
		}
		return
	})

	eg.Go(func(ctx context.Context) (err error) {
		if len(mids) > 0 {
			accountDetail, err = s.accClient.Infos3(ctx, &accountapi.MidsReq{Mids: mids})
			if err != nil {
				log.Errorc(ctx, "s.accClient.Infos3 (%v)", err)
				return err
			}
		}
		return

	})
	eg.Go(func(ctx context.Context) (err error) {
		userScoreDetail, err = s.dao.GetRankCache(ctx, bid, date)
		if err != nil {
			log.Errorc(ctx, "s.dao.GetRankCache err(%v)", err)
		}
		return
	})
	if err = eg.Wait(); err != nil {
		log.Error("UserRankList eg.Wait  error(%+v)", err)
		return res, err
	}
	if accountDetail == nil && accountDetail.Infos == nil {
		return res, nil
	}
	midInfo := make(map[int64]*bwsmdl.MidScore)
	for _, v := range mids {
		var account *accountapi.Info
		var user *bwsmdl.UserDetail
		var ok1 bool
		var ok2 bool
		if account, ok1 = accountDetail.Infos[v]; !ok1 {
			continue
		}
		if user, ok2 = userDetails[v]; !ok2 {
			continue
		}
		var star, lastTime int64
		if now < updateTime && dayStr == date {
			star = user.StarInRank
			lastTime = user.StarLastTime
		} else {
			if userScoreDetail == nil {
				continue
			}
			if _, ok := userScoreDetail[v]; !ok {
				continue
			}
			score := userScoreDetail[v]
			star = score.Star
			lastTime = score.LastStarTime
		}
		data := &bwsmdl.Account{
			Mid:          account.Mid,
			Name:         account.Name,
			Face:         account.Face,
			Sex:          account.Sex,
			LastStarTime: lastTime,
			Star:         star,
		}
		midInfo[v] = &bwsmdl.MidScore{Star: star, LastStarTime: lastTime}
		dataList = append(dataList, data)
	}
	res.List = dataList
	if now < updateTime && dayStr == date {
		s.cache.Do(ctx, func(ctx context.Context) {
			retry(func() error {
				return s.dao.AddRankCache(ctx, bid, midInfo, date)
			})
		})
	}

	return res, nil

}

// UpdateUserDetail 更新用户数据
func (s *Service) UpdateUserDetail(ctx context.Context, mid, bid, lotteryUsed, ups int64, date string, heart int64, starMap map[int64]int64, isSuccess bool, reasonStruct *bwsmdl.LogReason, orderNo, token string) (err error) {
	var reason string
	if reasonStruct != nil {
		var bs []byte
		if bs, err = json.Marshal(&reasonStruct); err != nil {
			log.Error("reasonStruct json.Marshal() error(%v)", err)
			return
		}
		reason = string(bs)
	}
	if heart == 0 && (starMap == nil || len(starMap) == 0) && lotteryUsed == 0 && ups == 0 {
		return
	}
	var (
		tx *sql.Tx
	)
	if tx, err = s.dao.BeginTran(ctx); err != nil {
		log.Errorc(ctx, "s.dao.BeginTran() failed. error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Errorc(ctx, "%v", r)
			return
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Errorc(ctx, "tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Errorc(ctx, "tx.Commit() error(%v)", err)
		}
	}()
	// select for update
	userDetail, err := s.dao.RawUserDetailForUpdate(ctx, tx, bid, mid, date)
	if err != nil {
		return err
	}
	var playTimes, playSuccessTimes int64
	if heart != 0 {
		if userDetail.Heart+heart < 0 {
			return ecode.ActivityBwsHeartErr
		}
		// 插入log日志
		_, err := s.dao.CreateHeartLog(ctx, tx, bid, mid, heart, reason, orderNo, token)
		if err != nil {
			log.Errorc(ctx, " s.dao.CreateHeartLog(%d,%d,%d,%s,%s)", bid, mid, heart, reason, orderNo)
			return err
		}
		userDetail.Heart += heart
	}
	var star int64
	now := time.Now().Unix()
	//   如果新增捕获up主
	if ups > 0 {
		// 本次是否会更新星星 原始ups
		oldUpsStar := s.catchUpsCanGetStar(ctx, userDetail.Ups)
		userDetail.Ups += ups
		newUpsStar := s.catchUpsCanGetStar(ctx, userDetail.Ups)
		if newUpsStar > oldUpsStar {
			if starMap == nil {
				starMap = make(map[int64]int64)
			}
			starMap[bwsmdl.StarMapUps] = newUpsStar
		}
		log.Infoc(ctx, "ups oldUpsStar(%d) newUpsStar(%d) starMap(%v)", oldUpsStar, newUpsStar, starMap)
	}
	if starMap != nil && len(starMap) > 0 {
		playTimes++
		if isSuccess {
			playSuccessTimes++
		}
		oldStarMap := make(map[int64]int64)
		if err = json.Unmarshal([]byte(userDetail.StarDetail), &oldStarMap); err != nil {
			log.Error("oldStarMap json.Unmarshal(%s) error(%v)", userDetail.StarDetail, err)
			return err
		}

		for k, v := range starMap {
			var extraStar int64
			var highStar = v
			if oldStar, ok := oldStarMap[k]; ok {
				if oldStar >= MaxStar {
					extraStar = 0
					highStar = MaxStar
				} else {
					if oldStar+v >= MaxStar {
						extraStar = MaxStar - oldStar
						highStar = MaxStar
					} else {
						extraStar = v
						highStar = v + oldStar
					}
				}
			} else {
				extraStar = v
			}
			star += extraStar
			oldStarMap[k] = highStar
		}
		starInRank := s.getStarInRank(oldStarMap)
		// star 有更新
		if star > 0 {

			// 插入log日志
			_, err := s.dao.CreateStarLog(ctx, tx, bid, mid, star, reason, orderNo, token)
			if err != nil {
				log.Errorc(ctx, " s.dao.CreateHeartLog(%d,%d,%d,%s,%s)", bid, mid, heart, reason, orderNo)
				return err
			}
			userDetail.Star += star
			bs, err := json.Marshal(oldStarMap)
			if err != nil {
				log.Errorc(ctx, " json.Marshal(%v)", oldStarMap)
				return err
			}
			userDetail.StarDetail = string(bs)
			if starInRank != userDetail.StarInRank {
				userDetail.StarLastTime = now
				userDetail.StarInRank = starInRank
				// 获取top1000的最低分
				score, err := s.dao.CacheUserRankMinScore(ctx, bid, date, s.c.Bws.RankTop1000-1)
				if err != nil {
					log.Errorc(ctx, "s.dao.CacheUserRankMinScore err(%v)", err)
				}
				// 时间小于当天4点
				updateTime := s.getRankUpdateLimitTime(ctx)
				log.Infoc(ctx, "UpdateStarRank now(%d),updateTime(%d)", now, updateTime)
				if now < updateTime {
					midScore := s.starToScore(userDetail.StarInRank, now)
					log.Infoc(ctx, "UpdateStarRank lastScore(%f) midScore(%f)  midScore < score(%v)", score, midScore, midScore < score)
					//判断是否要更新榜单
					if score == 0 || midScore < score {
						log.Infoc(ctx, "s.dao.AddCacheInsertUserScore midScore(%f) score(%f) ", midScore, score)
						err = s.dao.AddCacheInsertUserScore(ctx, bid, &bwsmdl.UserRank{Mid: mid, Score: midScore}, date)
						if err != nil {
							log.Errorc(ctx, "s.dao.AddCacheInsertUserScore err(%v)", err)
							return err
						}
					}
				}
			}
		}
	}
	// 如果使用了抽奖次数
	if lotteryUsed > 0 {
		if userDetail.LotteryUsed+lotteryUsed > userDetail.Star {
			return ecode.ActivityNoTimes
		}
		// 本次活动无任务抽奖，taskID传mid
		_, err := s.dao.UseLotteryTimes(ctx, token, userDetail.Mid)
		if err != nil {
			log.Errorc(ctx, " s.dao.UseLotteryTimes(%s,%d)", token, userDetail.Mid)
			return err
		}
		userDetail.LotteryUsed += lotteryUsed
	}

	err = s.dao.UpdateUserDetail(ctx, tx, userDetail.Id, userDetail.Star, userDetail.Heart, userDetail.StarLastTime, userDetail.LotteryUsed, userDetail.Ups, userDetail.StarDetail, userDetail.StarInRank, playTimes, playSuccessTimes)
	if err != nil {
		log.Errorc(ctx, "s.dao.UpdateUserDetail(%d,%d,%d,%d,%d,%v,%d) err(%v)", userDetail.Id, userDetail.Star, userDetail.Heart, userDetail.StarLastTime, userDetail.LotteryUsed, userDetail.StarDetail, userDetail.StarInRank, err)
		return err
	}
	// 清除用户缓存
	s.cache.Do(ctx, func(ctx context.Context) {
		retry(func() error {
			return s.dao.DelCacheUserDetail(ctx, bid, mid, date)
		})
	})
	return nil
}

func (s *Service) getStarInRank(oldStarMap map[int64]int64) (star int64) {
	allStar := make([]int64, 0)

	for _, v := range oldStarMap {
		allStar = append(allStar, v)
	}
	sort.Slice(allStar, func(i, j int) bool {
		return allStar[i] > allStar[j]
	})
	count := 5
	for i, v := range allStar {
		if i < count {
			star += v
		}
	}
	return
}

// starToScore
func (s *Service) starToScore(star, lastTime int64) float64 {
	strLastTime := fmt.Sprintf("0.%d", lastTime)
	floatTime, err := strconv.ParseFloat(strLastTime, 64)
	if err != nil {
		return float64(MaxScore) - float64(star)
	}
	return float64(MaxScore) - float64(star) + floatTime
}

// CreateUserToken 创建用户token
func (s *Service) CreateUserToken(ctx context.Context, loginMid, pid, bid int64) (string, error) {

	pointReply, err := s.dao.BwsPoints(ctx, []int64{pid})
	if err != nil {
		log.Error("s.dao.BwsPoints error(%v)", err)
		return "", ecode.ActivityPointFail
	}
	point, ok := pointReply[pid]

	if !ok || point == nil || point.Bid != bid {
		return "", ecode.ActivityIDNotExists
	}
	if !s.isAdmin(loginMid) && point.Ower != loginMid {
		return "", ecode.ActivityNotOwner
	}
	userToken := createBwsKey(ctx, bid, pid, time.Now().UnixNano())
	if _, err = s.dao.CreateUser(ctx, bid, 0, userToken); err != nil {
		log.Error("CreateUserToken s.dao.CreateUser userToken(%s) error(%v)", userToken, err)
		return "", err
	}
	if err = s.dao.DelCacheUsersKey(ctx, bid, userToken); err != nil {
		log.Error("CreateUserToken s.dao.CreateUser userToken(%s) error(%v)", userToken, err)
	}
	return userToken, nil
}

func createBwsKey(ctx context.Context, bid, pid, ts int64) string {
	hasher := md5.New()
	var traceId string
	if trace, ok := trace.FromContext(ctx); ok {
		traceId = trace.TraceID()
	}
	key := fmt.Sprintf("%d_%d_%d_%s", bid, pid, ts, traceId)
	hasher.Write([]byte(key))
	return hex.EncodeToString(hasher.Sum(nil))[0:15]
}

func todayDate() (int64, string) {
	dayStr := time.Now().Format("20060102")
	res, err := strconv.ParseInt(dayStr, 10, 64)
	if err != nil {
		return 0, ""
	}
	return res, dayStr
}

// UserPlayable 用户是否可以玩游戏
func (s *Service) UserPlayable(ctx context.Context, mid, bid, gameID int64) (userDetail *bwsmdl.UserDetail, userToken string, needHeart int64, err error) {
	_, day := todayDate()
	log.Infoc(ctx, "UserPlayable  mid(%d) bid(%d) gameID(%d)", mid, bid, gameID)
	userToken, err = s.midToKey(ctx, bid, mid)
	if err != nil {
		log.Errorc(ctx, "UserPlayable s.midToKey (%d,%d)", bid, mid)
		return nil, "", 0, err
	}
	// 获取用户情况
	userDetail, err = s.getUserDetail(ctx, bid, mid, userToken, day, false)
	if err != nil {
		log.Errorc(ctx, "UserPlayable s.dao.UserDetail(%d,%d,%v)", bid, mid, day)
		return nil, userToken, 0, ecode.ActivityBwsMidErr
	}
	pointReply, err := s.dao.BwsPoints(ctx, []int64{gameID})
	if err != nil {
		log.Error("s.dao.BwsPoints error(%v)", err)
		return nil, "", 0, ecode.ActivityPointFail
	}
	point, ok := pointReply[gameID]
	if !ok || point == nil || point.Bid != bid {
		return nil, "", 0, ecode.ActivityIDNotExists
	}
	needHeart = point.Unlocked
	if userDetail.Heart < needHeart {
		log.Errorc(ctx, "UserPlayable userDetail.Heart < needHeart (%d,%d)", userDetail.Heart, needHeart)
		return userDetail, userToken, needHeart, ecode.ActivityBwsHeartErr
	}
	return userDetail, userToken, needHeart, nil

}

// UserPlayGame 用户玩游戏
func (s *Service) UserPlayGame(ctx context.Context, mid, bid, gameID, star int64, isSuccess bool) error {
	_, day := todayDate()
	now := time.Now().Unix()
	_, userToken, needHeart, err := s.UserPlayable(ctx, mid, bid, gameID)
	if err != nil {
		return err
	}
	gameString := strconv.FormatInt(gameID, 10)
	starMap := make(map[int64]int64)
	starMap[gameID] = star
	reasonStruct := &bwsmdl.LogReason{
		Reason: bwsmdl.ReasonPlayGame,
		Params: gameString,
	}
	orderNo := fmt.Sprintf("%d_%d_%d_%d", mid, bid, gameID, now)
	err = s.UpdateUserDetail(ctx, mid, bid, 0, 0, day, -needHeart, starMap, isSuccess, reasonStruct, orderNo, userToken)
	return err
}

func (s *Service) IsTest(ctx context.Context) bool {
	if s.c.Bws.IsTest == 1 {
		return true
	}
	return false
}

func (s *Service) IsVip(ctx context.Context) bool {
	if s.c.Bws.IsVip == 1 {
		return true
	}
	return false
}

// GetVipMidDate
func (s *Service) GetVipMidDate(ctx context.Context) (mid int64, date string) {
	return s.c.Bws.VipMid, s.c.Bws.VipDate
}

// GetNormalMidDate
func (s *Service) GetNormalMidDate(ctx context.Context) (mid int64, date string) {
	return s.c.Bws.NormalMid, s.c.Bws.NormalDate
}

// AdminAddStar admin+星
func (s *Service) AdminAddStar(ctx context.Context, adminMid, bid, gameID, star int64, key string, isSuccess bool) error {
	mid, _, err := s.keyToMid(ctx, bid, key)
	if err != nil {
		return err
	}
	pointReply, err := s.dao.BwsPoints(ctx, []int64{gameID})
	if err != nil {
		log.Error("s.dao.BwsPoints error(%v)", err)
		return ecode.ActivityPointFail
	}
	point, ok := pointReply[gameID]
	if !ok || point == nil || point.Bid != bid {
		return ecode.ActivityIDNotExists
	}
	if !s.isAdmin(adminMid) && point.Ower != adminMid {
		return ecode.ActivityNotOwner
	}
	_, day := todayDate()
	_, err = s.getUserDetail(ctx, bid, mid, key, day, false)
	if err != nil {
		log.Errorc(ctx, "UserPlayable s.dao.UserDetail(%d,%d,%v)", bid, mid, day)
		return ecode.ActivityBwsMidErr
	}

	return s.UserPlayGame(ctx, mid, bid, gameID, star, isSuccess)
}

// AdminAddHeart admin+体力
func (s *Service) AdminAddHeart(ctx context.Context, adminMid, bid, heart int64, key string, timestamp int64) error {
	mid, _, err := s.keyToMid(ctx, bid, key)
	if err != nil {
		return err
	}
	_, day := todayDate()

	if !s.isAdmin(adminMid) {
		return ecode.ActivityNotOwner
	}
	_, err = s.getUserDetail(ctx, bid, mid, key, day, false)
	if err != nil {
		log.Errorc(ctx, "UserPlayable s.dao.UserDetail(%d,%d,%v)", bid, mid, day)
		return ecode.ActivityBwsMidErr
	}
	orderNo := fmt.Sprintf("%d_%d_%d", mid, bid, timestamp)
	reasonStruct := &bwsmdl.LogReason{
		Reason: bwsmdl.ReasonPlayGame,
		Params: fmt.Sprintf("admin:%d", adminMid),
	}
	err = s.UpdateUserDetail(ctx, mid, bid, 0, 0, day, heart, nil, false, reasonStruct, orderNo, key)
	if err != nil {
		log.Errorc(ctx, "s.UpdateUserDetail(%d) err(%v) ", mid, err)
		return err
	}
	return err
}

// GetUserToken 获取用户token
func (s *Service) GetUserToken(ctx context.Context, bid, mid int64) (key string, err error) {
	key, err = s.midToKey(ctx, bid, mid)
	if err != nil && err != ecode.ActivityNotBind {
		log.Errorc(ctx, "VipAddHeart s.midToKey (%d,%d)", bid, mid)
		return
	}
	if xecode.EqualError(ecode.ActivityNotBind, err) {
		key = createBwsKey(ctx, bid, mid, time.Now().UnixNano())
		if _, err = s.dao.CreateUser(ctx, bid, mid, key); err != nil {
			log.Errorc(ctx, "s.dao.Binding mid(%d) key(%s)  error(%v)", mid, key, err)
			return
		}
		err = s.dao.DelCacheUsersMid(ctx, bid, mid)
		if err != nil {
			log.Errorc(ctx, "s.dao.DelCacheUsersMid (%v)", err)
		}
	}
	return
}

// InternalAddHeart 内部增加体力
func (s *Service) InternalAddHeart(ctx context.Context, bid, mid, heart int64, date string, orderNo string) (err error) {
	key, err := s.GetUserToken(ctx, bid, mid)
	if err != nil {
		log.Errorc(ctx, "InternalAddHeart s.GetUserToken (%d,%d)", bid, mid)
		return err
	}
	_, err = s.getUserDetail(ctx, bid, mid, key, date, false)
	if err != nil {
		log.Errorc(ctx, "UserPlayable s.dao.UserDetail(%d,%d,%v)", bid, mid, date)
		return ecode.ActivityBwsMidErr
	}
	reasonStruct := &bwsmdl.LogReason{
		Reason: bwsmdl.ReasonInternalAddHeart,
		Params: fmt.Sprintf("mid:%d", mid),
	}
	err = s.UpdateUserDetail(ctx, mid, bid, 0, 0, date, heart, nil, false, reasonStruct, orderNo, key)
	if err != nil {
		log.Errorc(ctx, "s.UpdateUserDetail(%d) err(%v) ", mid, err)
		return err
	}
	return err

}

// VipAddHeart vip+体力
func (s *Service) VipAddHeart(ctx context.Context, mid, bid int64, vipKey, date string) (err error) {
	key, err := s.GetUserToken(ctx, bid, mid)
	if err != nil {
		log.Errorc(ctx, "InternalAddHeart s.GetUserToken (%d,%d)", bid, mid)
		return err
	}
	_, err = s.getUserDetail(ctx, bid, mid, vipKey, date, true)
	if err != nil {
		log.Errorc(ctx, "UserPlayable s.dao.UserDetail(%d,%d,%v)", bid, mid, date)
		return ecode.ActivityBwsMidErr
	}
	eg := errgroup.WithContext(ctx)
	var (
		vipKeyInfo *bwsmdl.VipUsersToken
		vipMidInfo *bwsmdl.VipUsersToken
	)
	eg.Go(func(ctx context.Context) (err error) {
		// 验证vip key 是否存在且未绑定
		vipKeyInfo, err = s.dao.UsersVipKey(ctx, bid, vipKey)
		if err != nil {
			return err
		}
		return
	})
	eg.Go(func(ctx context.Context) (err error) {
		// 验证用户是否绑定过
		vipMidInfo, err = s.dao.UsersVipMidDate(ctx, bid, mid, date)
		if err != nil {
			return err
		}
		return
	})
	if err = eg.Wait(); err != nil {
		log.Error("UserDetail eg.Wait mid:%d error(%+v)", mid, err)
		return err
	}
	// 未找到key
	if vipKeyInfo == nil {
		return ecode.ActivityBwsVipKeyErr
	}
	if vipKeyInfo.Mid != 0 {
		return ecode.ActivityBwsVipKeyAlreadyBindErr
	}
	if vipMidInfo.VipKey != "" {
		return ecode.ActivityBwsVipMidAlreadyBindErr
	}

	orderNo := fmt.Sprintf("%d_%d_%s", mid, bid, date)
	reasonStruct := &bwsmdl.LogReason{
		Reason: bwsmdl.ReasonVipAddHeart,
		Params: fmt.Sprintf("key:%s", vipKey),
	}
	err = s.UpdateUserDetail(ctx, mid, bid, 0, 0, date, s.c.Bws.VipHeart, nil, false, reasonStruct, orderNo, key)
	if err != nil {
		log.Errorc(ctx, "s.UpdateUserDetail (%d) err(%v) ", mid, err)
		return err
	}
	err = s.dao.UseVipKey(ctx, mid, vipKey, date, bid)
	if err != nil {
		log.Errorc(ctx, "s.dao.UseVipKey mid(%d) err(%v)", mid, err)
		return err
	}
	s.cache.Do(ctx, func(c context.Context) {
		retry(func() error {
			return s.dao.DelCacheUsersVipKey(c, bid, vipKey)
		})
		retry(func() error {
			return s.dao.DelCacheUsersVipMidDate(c, bid, mid, date)
		})

	})
	return nil

}

// MidToKey
func (s *Service) MidToKey(ctx context.Context, bid int64, mid int64) (key string, err error) {
	key, err = s.midToKey(ctx, bid, mid)
	if err != nil {
		log.Errorc(ctx, "UserDetail s.midToKey (%d,%d)", bid, mid)
		return
	}
	return
}

// AdminUserDetail ...
func (s *Service) AdminUserDetail(ctx context.Context, mid int64, bid int64, day string) (userDetailReply *bwsmdl.UserDetailReply, err error) {
	var accInfo = &accountapi.InfoReply{}
	accInfo, err = s.accClient.Info3(ctx, &accountapi.MidReq{Mid: mid})
	if err != nil {
		log.Errorc(ctx, "User2020 Info3:%d error:%v", mid, err)
		return
	}
	user := &bwsmdl.UserInfo2020{
		Mid:  mid,
		Name: accInfo.GetInfo().GetName(),
		Face: accInfo.GetInfo().GetFace(),
	}
	userDetailReply = &bwsmdl.UserDetailReply{}
	userDetailReply.User = user
	return
}

// UserDetail 用户详情
func (s *Service) UserDetail(ctx context.Context, mid int64, bid int64, day string, isVip bool) (userDetailReply *bwsmdl.UserDetailReply, err error) {
	var key string
	key, err = s.midToKey(ctx, bid, mid)
	if err != nil {
		log.Errorc(ctx, "UserDetail s.midToKey (%d,%d)", bid, mid)
		return nil, err
	}
	eg := errgroup.WithContext(ctx)
	var userDetail = &bwsmdl.UserDetail{Mid: mid, Bid: bid, BwsDate: day, State: 1, Heart: s.c.Bws.DefaultHeart}
	var rank int64
	userDetailReply = &bwsmdl.UserDetailReply{UserDetail: userDetail}
	userDetailReply.StarGameDetail = make(map[int64]*bwsmdl.GameStarRank)
	userDetailReply.StarGame = make(map[int64]int64)
	var accInfo = &accountapi.InfoReply{}
	var point = make(map[string][]*bwsmdl.SinglePoints)
	var lotteryLog []*bwsmdl.LotteryLog
	var userGameRank *tvbwapi.MyScoreResp
	eg.Go(func(ctx context.Context) (err error) {
		// 获取用户情况
		userDetail, err = s.getUserDetail(ctx, bid, mid, key, day, isVip)
		if err != nil {
			log.Errorc(ctx, "UserPlayable s.dao.UserDetail(%d,%d,%v)", bid, mid, day)
			return ecode.ActivityBwsMidErr
		}
		return
	})
	eg.Go(func(ctx context.Context) (err error) {
		// 获取用户情况
		rank, err = s.dao.UserRank(ctx, bid, mid, day)
		if err != nil {
			return err
		}
		return
	})
	eg.Go(func(ctx context.Context) (err error) {
		accInfo, err = s.accClient.Info3(ctx, &accountapi.MidReq{Mid: mid})
		if err != nil {
			log.Errorc(ctx, "User2020 Info3:%d error:%v", mid, err)
			return err
		}

		return
	})
	eg.Go(func(ctx context.Context) (err error) {
		point, err = s.Points(ctx, &bwsmdl.ParamPoints{Bid: bid, Tp: bwsmdl.AchieveGameType})
		if err != nil {
			log.Errorc(ctx, "User2020 Points: bid %d error:%v", bid, err)
			return err
		}
		return
	})
	eg.Go(func(ctx context.Context) (err error) {
		lotteryLog, _, err = s.userLotteryLog(ctx, key)
		if err != nil {
			log.Errorc(ctx, "User2020 userLotteryLog token:%s error:%+v", key, err)
		}
		return nil
	})
	eg.Go(func(ctx context.Context) (err error) {
		userGameRank, err = client.TvBwClient.MyScore(ctx, &tvbwapi.MyScoreReq{
			Mid: mid,
		})
		if err != nil {
			log.Errorc(ctx, "User2020 userLotteryLog token:%s error:%+v", key, err)
		}
		return nil
	})
	if err = eg.Wait(); err != nil {
		log.Error("UserDetail eg.Wait mid:%d error(%+v)", mid, err)
		return userDetailReply, err
	}
	user := &bwsmdl.UserInfo2020{
		Mid:  mid,
		Name: accInfo.GetInfo().GetName(),
		Key:  key,
		Face: accInfo.GetInfo().GetFace(),
	}
	if userDetail != nil && userDetail.Mid != 0 {
		userDetailReply = &bwsmdl.UserDetailReply{UserDetail: userDetail}
	}
	userDetailReply.StarGameDetail = make(map[int64]*bwsmdl.GameStarRank)
	userDetailReply.StarGame = make(map[int64]int64)
	if userDetailReply != nil {
		if rank <= s.c.Bws.RankTop1000 {
			userDetailReply.Rank = rank
		}
		userDetailReply.LotteryRemain = userDetailReply.Star - userDetailReply.LotteryUsed
	}
	if userDetail.StarDetail != "" {
		if err = json.Unmarshal([]byte(userDetail.StarDetail), &userDetailReply.StarGame); err != nil {
			log.Errorc(ctx, "UserDetail StarDetail json.Unmarshal(%s) error(%v)", userDetail.StarDetail, err)
			return userDetailReply, err
		}
	}
	userGameRankDetail := make(map[int64]*tvbwapi.MyScoreItem)
	if userGameRank != nil && len(userGameRank.List) > 0 {
		for _, v := range userGameRank.List {
			userGameRankDetail[v.Gid] = v
		}
	}
	if len(point) > 0 {
		if p, ok := point[bwsmdl.Game]; ok {
			for _, k := range p {
				s := &bwsmdl.GameStarRank{}
				if star, ok := userDetailReply.StarGame[k.ID]; ok {
					s.Star = star
				}
				if rank, ok := userGameRankDetail[k.ID]; ok {
					s.Rank = rank.Rank
				}
				userDetailReply.StarGameDetail[k.ID] = s
			}
		}
	}
	userDetailReply.LotteryLog = lotteryLog
	userDetailReply.User = user
	if userGameRank != nil {
		userDetailReply.RankEntryNum = userGameRank.RankEntryNum
		userDetailReply.RankFirstNum = userGameRank.RankFirstNum
	}
	return userDetailReply, err
}

// AddUserRankInternal 干预用户加入榜单
func (s *Service) AddUserRankInternal(ctx context.Context, mid int64, bid int64, day string) (err error) {
	userDetail, err := s.dao.UserDetail(ctx, bid, mid, day)
	if err != nil {
		return err
	}
	midScore := s.starToScore(userDetail.Star, userDetail.StarLastTime)
	err = s.dao.AddCacheInsertUserScore(ctx, bid, &bwsmdl.UserRank{Mid: mid, Score: midScore}, day)
	if err != nil {
		log.Errorc(ctx, "s.dao.AddCacheInsertUserScore err(%v)", err)
		return err
	}
	return nil
}

// DelUserRank 删除用户排行
func (s *Service) DelUserRank(ctx context.Context, mid int64, bid int64, day string) (err error) {
	return s.dao.DelRankMid(ctx, bid, day, mid)
}
