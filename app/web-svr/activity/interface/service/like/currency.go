package like

import (
	"context"
	"encoding/json"
	"sort"
	"strconv"
	"strings"
	"time"

	api "git.bilibili.co/bapis/bapis-go/cheese/service/coupon"

	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/model/currency"
	"go-gateway/app/web-svr/activity/interface/model/task"
	suitmdl "go-main/app/account/usersuit/service/api"

	"go-common/library/sync/errgroup.v2"
)

const (
	hasAward  = 1
	stepZero  = 0
	stepOne   = 1
	stepTwo   = 2
	stepThree = 3
	stepFour  = 4
	stepFive  = 5
	stepSix   = 6
	stepSeven = 7
	stepEight = 8
	numThird  = 3
	numSecond = 5
	numFirst  = 7
)

// CurCurrency .
func (s *Service) CurCurrency(c context.Context, sid, mid int64) (curReply *currency.CurCurrencyReply) {
	var (
		hasLock   int
		err       error
		addLock   bool
		lenNum    int
		curAmount int64
	)
	if sid != s.c.Image.TenSid {
		return
	}
	// 获取当天的foreign_id
	key := time.Now().Format("2006-01-02")
	curReply = &currency.CurCurrencyReply{Amount: 0, AllAmount: currency.TenUnlockAmount, Image: s.dailyImage(), H5Image: s.dailyH5Image()}
	group := errgroup.WithContext(c)
	group.Go(func(ctx context.Context) error {
		if curAmount >= currency.TenUnlockAmount {
			//判断当日是否解锁
			if hasLock, err = s.currDao.CacheUnlockState(ctx, key); err == nil && hasLock <= 0 {
				if addLock, err = s.currDao.AddCacheUnlockState(ctx, key, 1); err == nil && addLock {
					s.currDao.AddCacheLockNum(ctx, sid)
				}
			}
		}
		lenNum, _ = s.currDao.CacheLockNum(ctx, sid)
		if lenNum > 0 {
			curReply.Password = currency.TenLuckNum[:lenNum]
		}
		return nil
	})
	if mid > 0 {
		group.Go(func(ctx context.Context) error {
			var res string
			if res, err = s.dao.RsGet(ctx, couponKey(mid, s.c.Rule.TenCoupon)); err == nil && res != "" {
				curReply.HasCoupon = 1
			}
			return nil
		})
	}
	group.Wait()
	return
}

// ActUserCurrency get user currency amount.
func (s *Service) ActUserCurrency(c context.Context, mid, sid int64) (amount int64) {
	var (
		err  error
		data *currency.UserCurrency
	)
	if sid != s.c.Image.Sid && sid != s.c.Image.TenSid && sid != s.c.Eleven.ElevenSid && sid != s.c.Scholarship.Sid {
		return
	}
	if sid == s.c.Image.TenSid {
		amount = s.TenUserCurrency(c, mid, currency.BusinessAct)
	} else {
		if data, err = s.UserCurrency(c, mid, currency.BusinessAct, sid); err != nil {
			log.Error("s.UserCurrency mid(%d) sid(%d) error(%v)", mid, sid, err)
			return
		}
		amount = data.Amount
	}
	return
}

// Currency get currency .
func (s *Service) Currency(c context.Context, businessID, foreignID int64) (curr *currency.Currency, err error) {
	var currRela *currency.CurrencyRelation
	if currRela, err = s.currDao.Relation(c, businessID, foreignID); err != nil {
		log.Error("CurrencyAmount s.currDao.Relation(%d,%d) error(%v)", businessID, foreignID, err)
		return
	}
	if currRela == nil || currRela.CurrencyID == 0 {
		err = xecode.NothingFound
		return
	}
	if curr, err = s.currDao.Currency(c, currRela.CurrencyID); err != nil {
		log.Error("CurrencyAmount s.currDao.Currency(%d) error(%v)", currRela.CurrencyID, err)
		return
	}
	if curr == nil || curr.ID == 0 {
		err = xecode.NothingFound
	}
	return
}

// TenUserCurrency .
func (s *Service) TenUserCurrency(c context.Context, mid, businessID int64) (amount int64) {
	var (
		err  error
		list *currency.UserCurrency
	)
	dailySid, ok := s.dailyForeignID()
	if !ok {
		return
	}
	//获取每日任务列表积分
	if list, err = s.UserCurrency(c, mid, businessID, dailySid); err != nil {
		log.Error("s.TaskList(%d,%d,%d)", mid, businessID, dailySid)
		return
	}
	if list != nil {
		amount = list.Amount
	}
	return
}

// UserCurrency get user currency data.
func (s *Service) UserCurrency(c context.Context, mid, businessID, foreignID int64) (data *currency.UserCurrency, err error) {
	var (
		curr     *currency.Currency
		currUser *currency.CurrencyUser
	)
	if curr, err = s.Currency(c, businessID, foreignID); err != nil {
		return
	}
	data = &currency.UserCurrency{Currency: curr}
	if mid > 0 {
		if currUser, err = s.currDao.CurrencyUser(c, mid, curr.ID); err != nil {
			log.Error("UserCurrency s.currDao.CurrencyUser(%d,%d) error(%v)", mid, curr.ID, err)
			return
		}
		if currUser != nil {
			data.Amount = currUser.Amount
		}
	}
	return
}

func (s *Service) upCurrencyAmount(c context.Context, businessID, foreignID, fromMid, toMid, amount int64, remark string) (err error) {
	var (
		curr *currency.Currency
	)
	if curr, err = s.Currency(c, businessID, foreignID); err != nil {
		return
	}
	if err = s.currDao.UpUserAmount(c, curr.ID, fromMid, toMid, amount, remark); err != nil {
		log.Error("s.currDao.UpUserAmount id(%d) fromMid(%d) toMid(%d) amount(%d) remark(%s) error(%v)", curr.ID, fromMid, toMid, amount, remark, err)
	} else {
		s.cache.Do(c, func(ctx context.Context) {
			if fromMid > 0 {
				s.currDao.DelCacheCurrencyUser(ctx, fromMid, curr.ID)
			}
			if toMid > 0 {
				s.currDao.DelCacheCurrencyUser(ctx, toMid, curr.ID)
			}
		})
	}
	return
}

// MikuList get miku list.
func (s *Service) MikuList(c context.Context, mid, sid int64) (mikuReply *currency.MikuReply, err error) {
	var (
		data                  *currency.UserCurrency
		curAmount, userAmount int64
		list                  = make([]*currency.MikuState, 4)
		val                   = make([]*currency.MikuState, 4)
		res                   []*currency.MikuAward
	)
	if sid != s.c.Image.MikuSid {
		return
	}
	for i := 0; i < 4; i++ {
		list[i] = &currency.MikuState{ID: i, HasLock: 0, Award: 0}
		val[i] = &currency.MikuState{ID: i, HasLock: 0, Award: 0}
	}
	mikuReply = new(currency.MikuReply)
	curAmount = 39600667
	//获取用户当前魔法值
	if mid > 0 {
		eg := errgroup.WithContext(c)
		eg.Go(func(ctx context.Context) (e error) {
			if data, e = s.UserCurrency(ctx, mid, currency.BusinessAct, sid); e != nil {
				log.Error("MikuList s.UserCurrency mid(%d) sid(%d) error(%v)", mid, sid, e)
				return
			}
			userAmount = data.Amount
			return
		})
		eg.Go(func(ctx context.Context) (e error) {
			if res, e = s.currDao.CacheMikuAward(ctx, sid, mid); e != nil {
				log.Error("MikuList s.currDao.CacheMikuState sid(%d) mid(%d) error(%v)", sid, mid, e)
				return
			}
			if len(res) != 0 {
				for _, tmp := range res {
					val[tmp.ID].ID = tmp.ID
					val[tmp.ID].Award = tmp.Award
				}
			}
			return
		})
		if err = eg.Wait(); err != nil {
			log.Error("eg.Wait() error(%v)", err)
			return
		}
	} else {
		userAmount = 0
	}
	//redis HashMap存魔法值解锁状态
	if curAmount >= currency.MikuAmount4 && val[currency.MikuStep4].HasLock == 0 {
		list[currency.MikuStep4].ID = currency.MikuStep4
		list[currency.MikuStep4].Award = val[currency.MikuStep4].Award
		list[currency.MikuStep4].HasLock = currency.HasLock
	} else {
		list[currency.MikuStep4] = val[currency.MikuStep4]
	}
	if curAmount >= currency.MikuAmount3 && val[currency.MikuStep3].HasLock == 0 {
		list[currency.MikuStep3].ID = currency.MikuStep3
		list[currency.MikuStep3].Award = val[currency.MikuStep3].Award
		list[currency.MikuStep3].HasLock = currency.HasLock
	} else {
		list[currency.MikuStep3] = val[currency.MikuStep3]
	}
	if curAmount >= currency.MikuAmount2 && val[currency.MikuStep2].HasLock == 0 {
		list[currency.MikuStep2].ID = currency.MikuStep2
		list[currency.MikuStep2].Award = val[currency.MikuStep2].Award
		list[currency.MikuStep2].HasLock = currency.HasLock
	} else {
		list[currency.MikuStep2] = val[currency.MikuStep2]
	}
	if curAmount >= currency.MikuAmount1 && val[currency.MikuStep1].HasLock == 0 {
		list[currency.MikuStep1].ID = currency.MikuStep1
		list[currency.MikuStep1].Award = val[currency.MikuStep1].Award
		list[currency.MikuStep1].HasLock = currency.HasLock
	} else {
		list[currency.MikuStep1] = val[currency.MikuStep1]
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].ID < list[j].ID
	})
	mikuReply.List = list
	mikuReply.UserAmount = userAmount
	mikuReply.CurrAmount = curAmount
	return
}

// SpecialList get special list.
func (s *Service) SpecialList(c context.Context, mid, sid int64) (singleRes *currency.SingleAwardRes, err error) {
	var (
		list   = make([]*currency.SingleAward, 0, 9)
		res    []*currency.SingleAward
		amount int64
	)
	if sid != s.c.Scholarship.Sid {
		return
	}
	for i := 0; i < 9; i++ {
		list = append(list, &currency.SingleAward{ID: i, Award: 0})
	}
	amount = s.ActUserCurrency(c, mid, sid)
	if mid > 0 {
		if res, err = s.currDao.CacheSingleAward(c, sid, mid); err != nil {
			log.Error("SpecialList s.currDao.CacheSingleAward sid(%d) mid(%d) error(%v)", sid, mid, err)
			return
		}
		if len(res) != 0 {
			for _, tmp := range res {
				list[tmp.ID].ID = tmp.ID
				list[tmp.ID].Award = tmp.Award
			}
		}
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].ID < list[j].ID
	})
	singleRes = &currency.SingleAwardRes{
		List:       list,
		UserAmount: amount,
	}
	return
}

// SpecialAward .
func (s *Service) SpecialAward(c context.Context, sid, mid int64, awardType int) (err error) {
	var (
		suitErr                        error
		curAmount                      int64
		applyOne, applyTwo, applyThree bool
		likes, studylikes              int64
		val                            []*currency.SingleAward
		list                           []*task.TaskItem
		eg                             = errgroup.WithContext(c)
	)
	if sid != s.c.Scholarship.Sid {
		return
	}
	if awardType > 8 {
		return
	}
	eg.Go(func(ctx context.Context) (err error) {
		curAmount = s.ActUserCurrency(ctx, mid, sid)
		return
	})
	eg.Go(func(ctx context.Context) (err error) {
		if list, err = s.taskList(ctx, mid, task.BusinessAct, s.c.Scholarship.LikeSid); err != nil {
			log.Error("SpecialAward s.taskList mid(%d) sid(%d) error(%v)", mid, s.c.Scholarship.LikeSid, err)
		}
		return
	})
	eg.Wait()
	curr := s.c.Scholarship.StepCurr[awardType]
	if curr > curAmount {
		err = ecode.ActivityTaskNotFinish
		return
	}
	if val, err = s.currDao.CacheSingleAward(c, sid, mid); err != nil {
		log.Error("SpecialAward s.currDao.CacheSingleAward(%d,%d) error(%v)", sid, mid, err)
		return
	}
	singleState := make(map[int]*currency.SingleAward, 9)
	for _, v := range val {
		singleState[v.ID] = v
	}
	state, ok := singleState[awardType]
	if ok && state != nil {
		if singleState[awardType].Award == currency.HasAward {
			err = ecode.ActivityHasAward
			return
		}
	}
	for _, value := range list {
		if value.ID == s.c.Scholarship.CountLikeID {
			likes = value.UserCount
		}
		if value.ID == s.c.Scholarship.CountStudyLikeID {
			studylikes = value.UserCount
		}
	}
	if likes+studylikes >= numThird {
		applyThree = true
	}
	if likes+studylikes >= numSecond {
		applyTwo = true
	}
	if likes+studylikes >= numFirst {
		applyOne = true
	}
	switch awardType {
	case stepZero:
		//发放头像挂件
		var mids = []int64{mid}
		if _, suitErr = s.suitClient.GrantByMids(c, &suitmdl.GrantByMidsReq{Mids: mids, Pid: s.c.Scholarship.GrantPid, Expire: s.c.Scholarship.GrantExpire}); suitErr != nil {
			log.Error("s.suitClient.GrantByMids(%d,%d,%d) error(%v)", mid, s.c.Scholarship.GrantPid, s.c.Scholarship.GrantExpire, suitErr)
			return
		}
		s.cache.Do(c, func(ctx context.Context) {
			res := &currency.SingleAward{
				ID:    stepZero,
				Award: hasAward,
			}
			s.currDao.SetCacheSingleAward(ctx, sid, mid, stepZero, res)
		})
	case stepOne:
		//加抽奖机会5次
		if err = s.retryAddLotteryTimes(c, s.c.Scholarship.LotterySid, mid, 2); err != nil {
			log.Error("SpecialAward retryAddLotteryTimes sid(%d) mid(%d) awardType(%d) error(%v)", s.c.Scholarship.LotterySid, mid, awardType, err)
			return
		}
		s.cache.Do(c, func(ctx context.Context) {
			res := &currency.SingleAward{
				ID:    stepOne,
				Award: hasAward,
			}
			s.currDao.SetCacheSingleAward(ctx, sid, mid, stepOne, res)
		})
	case stepTwo:
		// 付费课程券
		if _, err = s.cheeseClient.ReceiveCoupon(c, &api.ReceiveCouponReq{Mid: mid, BatchToken: s.c.Scholarship.MallCouponId}); err != nil {
			log.Error("SpecialAward s.cheeseClient.ReceiveCoupon(%d) error(%v)", mid, err)
			return
		}
		s.cache.Do(c, func(ctx context.Context) {
			res := &currency.SingleAward{
				ID:    stepTwo,
				Award: hasAward,
			}
			s.currDao.SetCacheSingleAward(ctx, sid, mid, stepTwo, res)
		})
	case stepThree:
		//加抽奖机会10次
		s.cache.Do(c, func(ctx context.Context) {
			for i := 0; i < 2; i++ {
				if err = s.retryAddLotteryTimes(ctx, s.c.Scholarship.LotterySid, mid, 2); err != nil {
					log.Error("SpecialAward retryAddLotteryTimes sid(%d) mid(%d) i(%d) awardType(%d) error(%v)", s.c.Scholarship.LotterySid, mid, i, awardType, err)
					return
				}
			}
			res := &currency.SingleAward{
				ID:    stepThree,
				Award: hasAward,
			}
			s.currDao.SetCacheSingleAward(ctx, sid, mid, stepThree, res)
		})
	case stepFour:
		//申请三级奖学金
		if !applyThree {
			err = ecode.ActivityTaskNotFinish
			return
		}
		if err = s.DoTask(c, mid, s.c.Scholarship.ThirdPrize, false); err != nil {
			log.Error("SpecialAward s.DoTask(%d,%d) error(%v)", mid, s.c.Scholarship.ThirdPrize, err)
			return
		}
		s.cache.Do(c, func(ctx context.Context) {
			res := &currency.SingleAward{
				ID:    stepFour,
				Award: hasAward,
			}
			s.currDao.SetCacheSingleAward(ctx, sid, mid, stepFour, res)
		})
	case stepFive:
		//加抽奖机会15次
		s.cache.Do(c, func(ctx context.Context) {
			for i := 0; i < 3; i++ {
				if err = s.retryAddLotteryTimes(ctx, s.c.Scholarship.LotterySid, mid, 2); err != nil {
					log.Error("SpecialAward retryAddLotteryTimes sid(%d) mid(%d) i(%d) awardType(%d) error(%v)", s.c.Scholarship.LotterySid, mid, i, awardType, err)
					return
				}
			}
			res := &currency.SingleAward{
				ID:    stepFive,
				Award: hasAward,
			}
			s.currDao.SetCacheSingleAward(ctx, sid, mid, stepFive, res)
		})
	case stepSix:
		//申请二级奖学金
		if !applyTwo {
			err = ecode.ActivityTaskNotFinish
			return
		}
		if err = s.DoTask(c, mid, s.c.Scholarship.SecondPrize, false); err != nil {
			log.Error("SpecialAward s.DoTask(%d,%d) error(%v)", mid, s.c.Scholarship.SecondPrize, err)
			return
		}
		s.cache.Do(c, func(ctx context.Context) {
			res := &currency.SingleAward{
				ID:    stepSix,
				Award: hasAward,
			}
			s.currDao.SetCacheSingleAward(ctx, sid, mid, stepSix, res)
		})
	case stepSeven:
		//加抽奖机会20次
		s.cache.Do(c, func(ctx context.Context) {
			for i := 0; i < 4; i++ {
				if err = s.retryAddLotteryTimes(ctx, s.c.Scholarship.LotterySid, mid, 2); err != nil {
					log.Error("SpecialAward retryAddLotteryTimes sid(%d) mid(%d) i(%d) awardType(%d) error(%v)", s.c.Scholarship.LotterySid, mid, i, awardType, err)
					return
				}
			}
			res := &currency.SingleAward{
				ID:    stepSeven,
				Award: hasAward,
			}
			s.currDao.SetCacheSingleAward(ctx, sid, mid, stepSeven, res)
		})
	case stepEight:
		//申请一级奖学金
		if !applyOne {
			err = ecode.ActivityTaskNotFinish
			return
		}
		if err = s.DoTask(c, mid, s.c.Scholarship.FirstPrize, false); err != nil {
			log.Error("SpecialAward s.DoTask(%d,%d) error(%v)", mid, s.c.Scholarship.FirstPrize, err)
			return
		}
		s.cache.Do(c, func(ctx context.Context) {
			res := &currency.SingleAward{
				ID:    stepEight,
				Award: hasAward,
			}
			s.currDao.SetCacheSingleAward(ctx, sid, mid, stepEight, res)
		})
	}
	return
}

func (s *Service) retryAddLotteryTimes(c context.Context, lotteryID, mid int64, retryCnt int) (err error) {
	for i := 0; i < retryCnt; i++ {
		if err = s.dao.AddLotteryTimes(c, lotteryID, mid); err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	return
}

func (s *Service) CertificateWall(c context.Context, sid, mid int64) (res []int, err error) {
	var (
		tasks                []*task.TaskItem
		list                 []*task.TaskItem
		amount, total, likes int64
		count                int
		val                  *currency.CertificateMsg
		isFirst, isAll       bool
	)
	if mid == 0 {
		return
	}
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) (e error) {
		val = s.certificateData
		if val == nil {
			return
		}
		for _, v := range val.List {
			if v.Data.First == mid {
				isFirst = true
			}
			bs := strings.Split(v.Data.TopTen, ",")
			for _, value := range bs {
				if value == strconv.FormatInt(mid, 10) {
					count++
					break
				}
			}
		}
		if isFirst {
			res = append(res, 6)
		}
		if count >= s.c.Scholarship.CertificateLimitNum1 {
			res = append(res, 7)
		}
		if count >= s.c.Scholarship.CertificateLimitNum2 {
			res = append(res, 8)
		}
		if count >= s.c.Scholarship.CertificateLimitNum3 {
			res = append(res, 9)
		}
		return
	})
	eg.Go(func(ctx context.Context) (e error) {
		amount = s.ActUserCurrency(ctx, mid, sid)
		if amount >= s.c.Scholarship.AmountLimit {
			res = append(res, 3)
		}
		if tasks, err = s.taskList(ctx, mid, task.BusinessAct, sid); err != nil {
			log.Error("CertificateWall s.TaskList(%d,%d) error(%v)", sid, mid, err)
			return
		}
		for _, v := range tasks {
			if v.ID == s.c.Scholarship.JoinTask && v.UserFinish == task.HasFinish {
				res = append(res, 1)
				continue
			}
			if v.ID == s.c.Scholarship.SignupTask && v.UserTotalCount >= s.c.Scholarship.SignupLimit {
				res = append(res, 2)
				continue
			}
			if v.ID == s.c.Scholarship.OtherLikeTask {
				if v.UserCount == int64(len(s.c.Scholarship.OtherSid)) {
					isAll = true
				}
				continue
			}
		}
		return
	})
	eg.Go(func(ctx context.Context) (e error) {
		if list, err = s.taskList(ctx, mid, task.BusinessAct, s.c.Scholarship.LikeSid); err != nil {
			log.Error("CertificateWall s.TaskList(%d,%d) error(%v)", s.c.Scholarship.LikeSid, mid, err)
			return
		}
		for _, value := range list {
			if value.ID == s.c.Scholarship.CountLikeID {
				total = value.UserCount
			}
			if value.ID == s.c.Scholarship.CountStudyLikeID {
				likes = value.UserCount
			}
		}
		if total >= s.c.Scholarship.LikeNumLimit1 {
			res = append(res, 4)
		}
		if total >= s.c.Scholarship.LikeNumLimit2 {
			res = append(res, 5)
		}
		return
	})
	eg.Wait()
	if isAll && likes+total >= s.c.Scholarship.AllLikeNum {
		res = append(res, 10)
	}
	return
}

// certificateproc .
func (s *Service) loadCertificateData() {
	var (
		c = context.Background()
	)
	s.certificateData = &currency.CertificateMsg{}
	res, err := s.dao.SourceItem(c, s.c.Scholarship.CertificateVID)
	if err != nil {
		log.Error("loadCertificateData s.dao.SourceItem(%d) error(%v)", s.c.Scholarship.CertificateVID, err)
		return
	}
	if err = json.Unmarshal(res, s.certificateData); err != nil {
		log.Error("loadCertificateData json.Unmarshal(%s) error(%v)", res, err)
		return
	}
	log.Info("loadCertificateData success")
}
