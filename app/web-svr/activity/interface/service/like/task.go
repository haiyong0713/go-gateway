package like

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"time"

	accapi "git.bilibili.co/bapis/bapis-go/account/service"
	coinmdl "git.bilibili.co/bapis/bapis-go/community/service/coin"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	xtime "go-common/library/time"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/model/currency"
	"go-gateway/app/web-svr/activity/interface/model/like"
	"go-gateway/app/web-svr/activity/interface/model/task"
	suitmdl "go-main/app/account/usersuit/service/api"

	"go-common/library/sync/errgroup.v2"
)

const (
	_imageTaskRemark     = "脑洞节任务奖励"
	_imageIntlTaskRemark = "完成任务奖励"
	_imageReason         = "脑洞节抽奖"
	_imageCurrRemark     = "脑洞节抽奖消耗"
	_imageLottryRemark   = "脑洞节抽奖失败恢复"
	_imageAwardRemark    = "脑洞节奖品"
	_singleTaskRemark    = "特殊活动任务奖励"
	_elevenCurrRemark    = "双十一活动抽奖消耗"
	_elevenLottryRemark  = "双十一活动抽奖失败恢复"
	_lotteryTimesLimit   = 4
	_sourceId            = 4
	_sourceActivityId    = "scholarship"
)

// ActTaskList act task list.
func (s *Service) ActTaskList(c context.Context, mid, sid int64) (data []*task.TaskItem, err error) {
	if sid == s.c.Image.TenSid {
		data, err = s.TenTaskList(c, mid, task.BusinessAct, sid)
	} else {
		data, err = s.taskList(c, mid, task.BusinessAct, sid)
	}
	// staff special user finish state
	if sid == s.c.Staff.PicSid {
		if len(data) != 4 {
			return
		}
		// 前3个任务完成状态显示特殊处理
		if data[0].UserFinish == 0 || data[1].UserFinish == 0 {
			data[1].UserFinish = 0
			data[2].UserFinish = 0
		}
	}
	return
}

// ImageDoTask image do task
func (s *Service) ImageDoTask(c context.Context, mid, taskID int64) (err error) {
	var acts map[int64]int
	if acts, err = s.dao.LikeActs(c, s.c.Image.AppointSid, mid, []int64{s.c.Image.AppointLid}); err != nil {
		log.Error("ImageLottery s.dao.LikeActs sid(%d) mid(%d) lid(%d) error(%v)", s.c.Image.AppointSid, mid, s.c.Image.AppointLid, err)
		return
	}
	if liked, ok := acts[s.c.Image.AppointLid]; !ok || liked <= 0 {
		err = ecode.ActivityNotJoin
		return
	}
	err = s.DoTask(c, mid, taskID, true)
	return
}

// TenTaskList .
func (s *Service) TenTaskList(c context.Context, mid, businessID, foreignID int64) (list []*task.TaskItem, err error) {
	var (
		oneItem []*task.TaskItem
	)
	dailySid, ok := s.dailyForeignID()
	if !ok {
		err = ecode.ActivityNotJoin
		return
	}
	eg := errgroup.WithContext(c)
	//获取首次分享的列报表
	eg.Go(func(ctx context.Context) (e error) {
		if oneItem, e = s.taskList(ctx, mid, businessID, foreignID); e != nil {
			log.Error("s.TaskList(%d,%d,%d)", mid, businessID, foreignID)
			e = nil
		}
		return
	})
	//获取每日任务列表
	eg.Go(func(ctx context.Context) (e error) {
		if list, e = s.taskList(ctx, mid, businessID, dailySid); e != nil {
			log.Error("s.TaskList(%d,%d,%d)", mid, businessID, dailySid)
			e = nil
		}
		return
	})
	eg.Wait()
	list = append(list, oneItem...)
	sort.Slice(list, func(i, j int) bool {
		return list[i].Rank > list[j].Rank
	})
	return
}

func (s *Service) taskList(c context.Context, mid, businessID, foreignID int64) (list []*task.TaskItem, err error) {
	var (
		ids       []int64
		tasks     map[int64]*task.Task
		taskState map[string]*task.UserTask
		stateErr  error
		nowTs     = time.Now().Unix()
	)
	if ids, err = s.taskDao.TaskIDs(c, businessID, foreignID); err != nil {
		log.Error("TaskList s.taskDao.TaskIDs(%d,%d) error(%v)", businessID, foreignID, err)
		return
	}
	if len(ids) == 0 {
		err = xecode.NothingFound
		return
	}
	tasks, err = s.taskDao.Tasks(c, ids)
	if err != nil {
		log.Error("TaskList s.taskDao.Tasks(%v) error(%v)", ids, err)
		return
	}
	if mid > 0 {
		if taskState, stateErr = s.taskDao.UserTaskState(c, tasks, mid, businessID, foreignID, nowTs); stateErr != nil {
			log.Error("TaskList s.taskDao.UserTaskState(%v) error(%v)", ids, stateErr)
		}
	}
	for _, v := range tasks {
		var item *task.TaskItem
		var total int64
		if v.IsAutoReceive() {
			item = &task.TaskItem{Task: v}
		}
		if mid > 0 && stateErr == nil {
			roundList := make([]xtime.Time, 0)
			if v.IsCycle() && v.CycleDuration > 0 {
				for i := v.Round(nowTs); i >= 0; i-- {
					if val, ok := taskState[fmt.Sprintf("%d_%d", v.ID, i)]; ok {
						total++
						roundList = append(roundList, val.Ctime)
					}
				}
			}
			if state, ok := taskState[fmt.Sprintf("%d_%d", v.ID, v.Round(nowTs))]; ok {
				item = &task.TaskItem{
					Task:           v,
					Ctime:          state.Ctime,
					UserRound:      state.Round,
					UserCount:      state.Count,
					UserFinish:     state.Finish,
					UserAward:      state.Award,
					UserTotalCount: total,
					UserRoundList:  roundList,
				}
			} else if v.Round(nowTs) > 0 {
				item = &task.TaskItem{
					Task:           v,
					UserTotalCount: total,
					UserRoundList:  roundList,
				}
			}
		}
		list = append(list, item)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Rank > list[j].Rank
	})
	return
}

// LedTask led a task.
func (s *Service) LedTask(c context.Context, mid, taskID int64) (err error) {
	var (
		data      *task.Task
		userState map[string]*task.UserTask
	)
	if data, err = s.taskDao.Task(c, taskID); err != nil {
		log.Error("LedTask s.taskDao.Task(%d) error(%v)", taskID, err)
		return
	}
	// TODO pre check
	nowTs := time.Now().Unix()
	if data.Stime.Time().Unix() > nowTs {
		err = ecode.ActivityTaskNotStart
		return
	}
	if data.Etime.Time().Unix() < nowTs {
		err = ecode.ActivityTaskOverEnd
		return
	}
	taskArg := map[int64]*task.Task{
		data.ID: data,
	}
	if userState, err = s.taskDao.UserTaskState(c, taskArg, mid, data.BusinessID, data.ForeignID, nowTs); err != nil {
		log.Error("LedTask s.taskDao.UserTaskState mid(%d) arg(%+v) error(%v)", mid, data, err)
		return
	}
	userTask, ok := userState[fmt.Sprintf("%d_%d", data.ID, data.Round(nowTs))]
	if ok {
		err = ecode.ActivityTaskHasLed
		return
	}
	if err = s.taskDao.AddUserTaskState(c, mid, data.BusinessID, taskID, data.ForeignID, data.Round(nowTs), task.InitCount, task.NotFinish, task.NotAward); err != nil {
		log.Error("LedTask s.taskDao.AddUserTaskState mid(%d) arg(%+v) error(%v)", mid, data, err)
		return
	}
	s.cache.Do(c, func(ctx context.Context) {
		upTask := userTask
		s.taskDao.SetCacheUserTaskState(ctx, upTask, mid, upTask.BusinessID, upTask.ForeignID)
	})
	return
}

// DoTask add task user log.
func (s *Service) DoTask(c context.Context, mid, taskID int64, isOut bool) (err error) {
	var (
		data          *task.Task
		userState     map[string]*task.UserTask
		count, finish int64
	)
	if data, err = s.taskDao.Task(c, taskID); err != nil {
		log.Error("DoTask s.taskDao.Task(%d) error(%v)", taskID, err)
		return
	}
	//外部请求非外部外部任务
	if isOut && !data.IsOut() {
		log.Warn("DoTask warn mid(%d) taskID(%d)", mid, taskID)
		err = xecode.RequestErr
		return
	}
	nowTs := time.Now().Unix()
	// 周期任务需0点开始，1天一周期，实现中午12点任务开始
	if data.ForeignID == s.c.Stein.Sid {
		if s.c.Stein.Stime > nowTs {
			err = ecode.ActivityTaskNotStart
			return
		}
	}
	if data.ForeignID == s.c.Scholarship.Sid {
		if s.c.Scholarship.Stime > nowTs {
			err = ecode.ActivityTaskNotStart
			return
		}
	}
	if data.Stime.Time().Unix() > nowTs {
		err = ecode.ActivityTaskNotStart
		return
	}
	if data.Etime.Time().Unix() < nowTs {
		err = ecode.ActivityTaskOverEnd
		return
	}
	// 判断前置任务状态
	if data.HasRule() {
		if err = s.taskPreCheck(c, mid, taskID, nowTs); err != nil {
			return
		}
	}
	if userState, err = s.taskDao.UserTaskState(c, map[int64]*task.Task{data.ID: data}, mid, data.BusinessID, data.ForeignID, nowTs); err != nil {
		log.Error("DoTask s.taskDao.UserTaskState mid(%d) arg(%v) error(%v)", mid, data, err)
		return
	}
	userTask, ok := userState[fmt.Sprintf("%d_%d", data.ID, data.Round(nowTs))]
	if !ok {
		if !data.IsAutoReceive() {
			err = ecode.ActivityTaskNotLed
			return
		}
	} else {
		if userTask.Finish == task.HasFinish {
			err = ecode.ActivityTaskHasFinish
			return
		}
	}
	// add log
	isNewTable := data.IsNewTable()
	if isNewTable {
		if err = s.taskDao.TaskUserLogAdd(c, mid, data.BusinessID, taskID, data.ForeignID, data.Round(nowTs)); err != nil {
			log.Error("DoTask s.taskDao.TaskUserLogAdd mid(%d) arg(%v) error(%v)", mid, data, err)
			return
		}
	} else {
		if err = s.taskDao.AddUserTaskLog(c, mid, data.BusinessID, taskID, data.ForeignID, data.Round(nowTs)); err != nil {
			log.Error("DoTask s.taskDao.AddUserTaskLog mid(%d) arg(%v) error(%v)", mid, data, err)
			return
		}
	}
	award := int64(task.NotAward)
	var roundCount int64
	if userTask != nil {
		count = userTask.Count + 1
		if !data.IsNoFinish() && data.FinishCount == count {
			finish = task.HasFinish
		}
		if data.NeedDayCount() {
			roundCount = userTask.RoundCount
			if nowTs/86400 > int64(userTask.Mtime)/86400 {
				roundCount = userTask.RoundCount + 1
			}
		}
		if data.IsSpecial() {
			award = userTask.Award
		}
		if isNewTable {
			if err = s.taskDao.TaskUserStateUp(c, mid, taskID, data.Round(nowTs), count, finish, award, data.ForeignID, roundCount); err != nil {
				log.Error("DoTask s.taskDao.TaskUserStateUp mid(%d) taskID(%d) count(%d) finish(%d) error(%v)", mid, taskID, count, finish, err)
				return
			}
		} else {
			if err = s.taskDao.UpUserTaskState(c, mid, taskID, data.Round(nowTs), count, finish, award, data.ForeignID); err != nil {
				log.Error("DoTask s.taskDao.UpUserTaskState mid(%d) taskID(%d) count(%d) finish(%d) error(%v)", mid, taskID, count, finish, err)
				return
			}
		}
	} else {
		count = 1
		if !data.IsNoFinish() && data.FinishCount == count {
			finish = task.HasFinish
		}
		if data.NeedDayCount() {
			roundCount = 1
		}
		if isNewTable {
			if err = s.taskDao.TaskUserStateAdd(c, mid, data.BusinessID, taskID, data.ForeignID, data.Round(nowTs), count, finish, award, roundCount); err != nil {
				log.Error("DoTask s.taskDao.TaskUserStateAdd mid(%d) task(%+v) count(%d) finish(%d) error(%v)", mid, data, count, finish, err)
				return
			}
		} else {
			if err = s.taskDao.AddUserTaskState(c, mid, data.BusinessID, taskID, data.ForeignID, data.Round(nowTs), count, finish, award); err != nil {
				log.Error("DoTask s.taskDao.AddUserTaskState mid(%d) task(%+v) count(%d) finish(%d) error(%v)", mid, data, count, finish, err)
				return
			}
		}
	}
	s.cache.Do(c, func(ctx context.Context) {
		upTask := &task.UserTask{
			Mid:        mid,
			BusinessID: data.BusinessID,
			TaskID:     taskID,
			ForeignID:  data.ForeignID,
			Round:      data.Round(nowTs),
			Count:      count,
			Finish:     finish,
			Award:      award,
			Ctime:      xtime.Time(nowTs),
			Mtime:      xtime.Time(nowTs),
			RoundCount: roundCount,
		}
		s.taskDao.SetCacheUserTaskState(ctx, upTask, mid, upTask.BusinessID, upTask.ForeignID)
	})
	return
}

// AddAwardTask
func (s *Service) AddAwardTask(c context.Context, mid, taskID, awardCount int64) (err error) {
	var (
		data *task.Task
	)
	if data, err = s.taskDao.Task(c, taskID); err != nil {
		log.Error("AwardTask s.taskDao.Task(%d) error(%v)", taskID, err)
		return
	}
	if !data.HasAward() {
		err = ecode.ActivityTaskNoAward
		return
	}
	switch data.AwardType {
	case task.AwardTypeCurr:
		var tempForeignID = data.ForeignID
		if s.c.Image.TenTaskID == taskID {
			if tempID, ok := s.dailyForeignID(); ok {
				tempForeignID = tempID
			}
		}
		err = s.upCurrencyAmount(c, data.BusinessID, tempForeignID, 0, mid, awardCount, _imageIntlTaskRemark)
	}
	return
}

// AwardTask change task award.
func (s *Service) AwardTask(c context.Context, mid, taskID int64) (rly *task.AwardReply, err error) {
	var (
		data      *task.Task
		userState map[string]*task.UserTask
		remark    string
	)
	if data, err = s.taskDao.Task(c, taskID); err != nil {
		log.Error("AwardTask s.taskDao.Task(%d) error(%v)", taskID, err)
		return
	}
	if !data.HasAward() {
		err = ecode.ActivityTaskNoAward
		return
	}
	if data.ForeignID == s.c.Stein.Sid {
		err = xecode.RequestErr
		return
	}
	// 春节红包活动 结束后删除代码 start
	if data.ForeignID == s.c.Image.YearSid {
		nowTs := time.Now().Unix()
		if data.Stime.Time().Unix() > nowTs {
			err = ecode.ActivityTaskNotStart
			return
		}
		if data.Etime.Time().Unix() < nowTs {
			err = ecode.ActivityTaskOverEnd
			return
		}
		var awCheck bool
		// 一个任务当前活动只支持领取一次，做频繁操作限制
		awKey := fmt.Sprintf("tk_awd_%d_%d", mid, taskID)
		if awCheck, err = s.dao.RsSetNX(c, awKey, 5); err != nil {
			return
		}
		if !awCheck {
			err = ecode.ActivityHasAward
			return
		}
		// 账号等级限制
		var memberRly *accapi.ProfileReply
		if memberRly, err = s.accClient.Profile3(c, &accapi.MidReq{Mid: mid}); err != nil || memberRly == nil || memberRly.Profile == nil {
			log.Error(" s.acc.Profile3(c,&accmdl.ArgMid{Mid:%d}) error(%v)", mid, err)
			err = ecode.ActGuessDataFail
			return
		}
		// 禁封用户
		if memberRly.Profile.Silence == _silenceForbid {
			err = ecode.ActivityMemberBlocked
			return
		}
		//未绑定手机号码
		if memberRly.Profile.TelStatus != 1 {
			err = ecode.ActivityTelValid
			return
		}
		//账号等级
		if memberRly.Profile.Level < s.c.Image.YearLevel {
			err = ecode.ActivityLikeLevelLimit
			return
		}
	}
	if data.ForeignID == s.c.SpringCardAct.InviteSid || data.ForeignID == s.c.SpringCardAct.Sid {
		var awCheck bool
		// 一个任务当前活动只支持领取一次，做频繁操作限制
		awKey := fmt.Sprintf("tk_awd_%d_%d", mid, taskID)
		if awCheck, err = s.dao.RsSetNX(c, awKey, 5); err != nil {
			return
		}
		if !awCheck {
			err = ecode.ActivityHasAward
			return
		}
	}
	// 春节红包活动 结束后删除代码 end
	taskArg := map[int64]*task.Task{
		data.ID: data,
	}
	t := time.Now()
	nowTs := t.Unix()
	if userState, err = s.taskDao.UserTaskState(c, taskArg, mid, data.BusinessID, data.ForeignID, nowTs); err != nil {
		log.Error("AwardTask s.taskDao.UserTaskState mid(%d) arg(%v) error(%v)", mid, data, err)
		return
	}
	userTask, ok := userState[fmt.Sprintf("%d_%d", data.ID, data.Round(nowTs))]
	if !ok || userTask == nil {
		err = ecode.ActivityTaskNotFinish
		return
	}
	// 特殊任务可以不完成领取奖励
	if !data.IsSpecial() {
		if userTask.Finish != task.HasFinish {
			err = ecode.ActivityTaskNotFinish
			return
		}
		if userTask.Award == task.HasAward {
			err = ecode.ActivityTaskHadAward
			return
		}
	}
	if data.IsSpecial() && userTask.Count == userTask.Award {
		return
	}
	// 春节红包活动 结束后删除代码 start
	if data.ForeignID == s.c.Image.YearSid {
		if userTask.Award >= task.HasAward {
			err = ecode.ActivityTaskHadAward
			return
		}
	}
	// 春节红包活动 结束后删除代码 end
	awardState := int64(task.HasAward)
	if data.IsSpecial() {
		awardState = userTask.Count
	}
	// 春节红包活动 结束后删除代码 start
	if data.ForeignID == s.c.Image.YearSid {
		var awardCount int64
		if awardCount, err = s.specialYear(c, mid, data.ForeignID, taskID); err != nil {
			return
		}
		awardState = awardCount
	}
	// 春节红包活动 结束后删除代码 end
	if err = s.taskDao.UserTaskAward(c, mid, taskID, data.Round(nowTs), awardState, data.ForeignID); err != nil {
		log.Error("AwardTask s.taskDao.UserTaskAward mid(%d) arg(%v) error(%v)", mid, data, err)
		return
	}
	rly = &task.AwardReply{Award: awardState}
	s.cache.Do(context.Background(), func(ctx context.Context) {
		upTask := userTask
		upTask.Award = awardState
		s.taskDao.SetCacheUserTaskState(ctx, upTask, mid, upTask.BusinessID, upTask.ForeignID)
	})
	switch data.AwardType {
	case task.AwardTypeCurr:
		var tempForeignID = data.ForeignID
		if s.c.Image.TenTaskID == taskID {
			if tempID, ok := s.dailyForeignID(); ok {
				tempForeignID = tempID
				remark = _imageTaskRemark
			}
		} else {
			remark = _singleTaskRemark
		}
		if data.IsSpecial() {
			if err = s.upCurrencyAmount(c, data.BusinessID, tempForeignID, 0, mid, (userTask.Count-userTask.Award)*data.AwardCount, remark); err != nil {
				log.Error("AwardTask s.UpCurrencyAmount mid(%d) arg(%v) remark(%s) error(%v)", mid, data, remark, err)
				return
			}
		} else {
			awardCount := data.AwardCount
			if taskID == s.c.Scholarship.SignupTask {
				if t.Weekday() == time.Saturday || t.Weekday() == time.Sunday {
					awardCount = 2 * data.AwardCount
				}
			}
			if err = s.upCurrencyAmount(c, data.BusinessID, tempForeignID, 0, mid, awardCount, remark); err != nil {
				log.Error("AwardTask s.UpCurrencyAmount mid(%d) arg(%v) remark(%s) error(%v)", mid, data, remark, err)
				return
			}
		}
	case task.AwardTypePend:
		//发放头像挂件
		grantPid := data.AwardID
		var suitExpire int64
		switch data.ForeignID {
		case s.c.Staff.PicSid:
			suitExpire = s.c.Staff.SuitExpire
		case s.c.Shad.Sid:
			suitExpire = s.c.Shad.SuitExpire
		case s.c.Special.SidOne:
			suitExpire = s.c.Special.ExpireOne
		case s.c.Special.SidTwo:
			suitExpire = s.c.Special.ExpireTwo
		case s.c.Special.SidThree:
			suitExpire = s.c.Special.ExpireThree
		default:
			suitExpire = s.c.Rule.SuitExpire
			if data.AwardExpire > 0 {
				suitExpire = data.AwardExpire
			}
		}
		if _, e := s.suitClient.GrantByMids(c, &suitmdl.GrantByMidsReq{Mids: []int64{mid}, Pid: grantPid, Expire: suitExpire}); e != nil {
			log.Error("s.suitClient.GrantByMids(%d,%d,%d) error(%v)", mid, grantPid, suitExpire, e)
			err = ecode.ActivityTaskAwardFailed
			return
		}
	case task.AwardTypeCoupon:
		// 优惠券 award id 类型不支持string，配置特殊支持
		if data.ForeignID == s.c.Staff.PicSid {
			if e := s.dao.MallCoupon(c, mid, _sourceId, s.c.Staff.Coupon, _sourceActivityId); e != nil {
				log.Error("s.bnjDao.GrantCoupon(%d,%s) error(%v)", mid, s.c.Staff.Coupon, e)
				err = ecode.ActivityTaskAwardFailed
				return
			}
		}
	case task.AwardTypeLottery:
		lotterySid := ""
		cid := int64(0)
		lottType := _other
		switch data.ForeignID {
		case s.c.Stupid.Sid:
			lotterySid = s.c.Stupid.LotterySid
			cid = s.c.Stupid.Cid
		case s.c.Restart2020.Sid:
			lotterySid = s.c.Restart2020.LotterySid
			cid = s.c.Restart2020.Cid
		case s.c.MobileGame.Sid:
			lotterySid = s.c.MobileGame.LotterySid
			cid = s.c.MobileGame.Cid
		default:
			return
		}
		s.cache.Do(context.Background(), func(ctx context.Context) {
			s.AddLotteryTimes(ctx, lotterySid, mid, cid, lottType, int(data.AwardCount), strconv.FormatInt(mid, 10)+strconv.FormatInt(nowTs, 10), false)
		})
	}
	return
}

// specialYear春节红包活动 结束后删除代码 start.
func (s *Service) specialYear(c context.Context, mid, fid, currTaskID int64) (awardCount int64, err error) {
	tasks, err := s.taskList(c, mid, task.BusinessAct, fid)
	if err != nil {
		log.Error("specialYear s.taskList(%d,%d) error(%v)", mid, fid, err)
		return
	}
	ids := make(map[int64]struct{})
	account := 0
	for _, v := range tasks {
		if v.ID == currTaskID && (v.UserFinish != task.HasFinish || v.UserAward >= task.HasAward) {
			err = xecode.RequestErr
			return
		}
		// 列表数据是在发放奖励后更新的，所以计算当前的任务id
		if (v.UserFinish == task.HasFinish && v.UserAward >= task.HasAward) || v.ID == currTaskID {
			// 一个投稿活动记一次奖励&& 已经领取
			if _, ok := ids[v.AwardID]; !ok {
				ids[v.AwardID] = struct{}{}
				account++
			}
		}
	}
	// 如果是0次，肯定是有问题直接反回error
	if account == 0 {
		err = xecode.RequestErr
		return
	}
	// 获取完成次数对应的奖励
	if avl, k := s.c.Image.YearAward[strconv.Itoa(account)]; k {
		awardCount = avl
	}
	return
}

func (s *Service) AwardTaskSpecial(c context.Context, mid, sid int64) (err error) {
	if sid != s.c.Stein.Sid {
		err = xecode.RequestErr
		return
	}
	var (
		tasks                         []*task.TaskItem
		taskLike, taskShare, hasAward bool
		awardTask                     *task.TaskItem
	)
	nowTs := time.Now().Unix()
	if tasks, err = s.taskList(c, mid, task.BusinessAct, sid); err != nil {
		return
	}
	for _, v := range tasks {
		if v.Task.IsCycle() {
			if v.UserTotalCount >= s.c.Stein.LikeCount {
				taskLike = true
			}
		} else {
			if v.UserFinish == task.HasFinish {
				taskShare = true
				hasAward = v.UserAward == task.HasAward
				awardTask = v
			}
		}
	}
	if !taskLike || !taskShare {
		err = ecode.ActivityTaskNotFinish
		return
	}
	if hasAward {
		err = ecode.ActivityTaskHadAward
		return
	}
	if awardTask == nil || awardTask.ID <= 0 {
		log.Warn("AwardTaskSpecial no finish task sid(%d) mid(%d)", sid, mid)
		return
	}
	if _, err = s.suitClient.GrantByMids(c, &suitmdl.GrantByMidsReq{Mids: []int64{mid}, Pid: s.c.Stein.SuitPid, Expire: s.c.Stein.SuitExpire}); err != nil {
		log.Error("AwardTaskSpecial GrantByMids mid(%d) error(%v)", mid, err)
		return
	}
	if err = s.taskDao.UserTaskAward(c, mid, awardTask.ID, awardTask.Round(nowTs), 1, sid); err != nil {
		log.Error("AwardTaskSpecial s.taskDao.UserTaskAward mid(%d) task(%v) error(%v)", mid, awardTask, err)
		return
	}
	s.cache.Do(c, func(ctx context.Context) {
		upTask := &task.UserTask{
			Mid:        mid,
			BusinessID: task.BusinessAct,
			ForeignID:  sid,
			TaskID:     awardTask.ID,
			Round:      awardTask.Round(nowTs),
			Count:      awardTask.UserCount,
			Finish:     awardTask.UserFinish,
			Award:      task.HasAward,
		}
		s.taskDao.SetCacheUserTaskState(ctx, upTask, mid, upTask.BusinessID, upTask.ForeignID)
	})
	return
}

// ImageLottery image lottery.
func (s *Service) ImageLottery(c context.Context, mid int64) (data *like.LotteryData, err error) {
	var (
		userCurrency *currency.UserCurrency
		lottery      *like.Lottery
		acts         map[int64]int
	)
	ip := metadata.String(c, metadata.RemoteIP)
	if acts, err = s.dao.LikeActs(c, s.c.Image.AppointSid, mid, []int64{s.c.Image.AppointLid}); err != nil {
		log.Error("ImageLottery s.dao.LikeActs sid(%d) mid(%d) lid(%d) error(%v)", s.c.Image.AppointSid, mid, s.c.Image.AppointLid, err)
		return
	}
	if liked, ok := acts[s.c.Image.AppointLid]; !ok || liked <= 0 {
		err = ecode.ActivityNotJoin
		return
	}
	// currency check.
	amount := s.c.Image.LotteryAmount
	if userCurrency, err = s.UserCurrency(c, mid, currency.BusinessAct, s.c.Image.Sid); err != nil {
		return
	}
	if userCurrency.Amount < amount {
		err = ecode.ActivityCurrLackAmount
		return
	}
	if err = s.upCurrencyAmount(c, currency.BusinessAct, s.c.Image.Sid, mid, 0, amount, _imageCurrRemark); err != nil {
		return
	}
	if lottery, err = s.dao.LotteryIndex(c, s.c.Image.LotteryID, 0, 0, mid); err != nil {
		log.Error("ImageLottery need check LotteryIndex sid(%d) mid(%d) error(%v)", s.c.Image.LotteryID, mid, err)
		s.cache.Do(c, func(ctx context.Context) {
			if e := s.upCurrencyAmount(ctx, currency.BusinessAct, s.c.Image.Sid, 0, mid, amount, _imageLottryRemark); e != nil {
				log.Error("ImageLottery rollback upCurrencyAmount mid:%d amount:%d error(%v)", mid, amount, e)
			}
		})
		return
	}
	if lottery.Code != 0 {
		// no lottery default lottery
		if lottery.Code == like.NoLotteryCode {
			data = &like.LotteryData{
				Name: s.c.Image.LotteryName,
			}
			s.cache.Do(c, func(ctx context.Context) {
				if e := s.upCurrencyAmount(ctx, currency.BusinessAct, s.c.Image.Sid, 0, mid, s.c.Image.LotteryAwardAmount, _imageAwardRemark); e != nil {
					log.Error("ImageLottery need check upCurrencyAmount mid:%d amount:%d error(%v)", mid, s.c.Image.LotteryAwardAmount, e)
				}
			})
			return
		}
		log.Warn("ImageLottery LotteryIndex sid(%d) mid(%d) error code(%d)", s.c.Image.LotteryID, mid, lottery.Code)
		s.cache.Do(c, func(ctx context.Context) {
			if e := s.upCurrencyAmount(ctx, currency.BusinessAct, s.c.Image.Sid, 0, mid, amount, _imageLottryRemark); e != nil {
				log.Error("ImageLottery rollback upCurrencyAmount mid:%d amount:%d error(%v)", mid, amount, e)
			}
		})
		return
	}
	data = lottery.Data
	switch lottery.Data.Name {
	case s.c.Image.LotteryCoinOne:
		s.cache.Do(c, func(ctx context.Context) {
			if _, e := s.coinClient.ModifyCoins(ctx, &coinmdl.ModifyCoinsReq{Mid: mid, Count: s.c.Image.LotteryCoinOneAmount, Reason: _imageReason, IP: ip}); e != nil {
				log.Error("ImageLottery need check coin.ModifyCoin mid:%d count:%f error(%v)", mid, s.c.Image.LotteryCoinOneAmount, e)
			}
		})
	case s.c.Image.LotteryCoinTwo:
		s.cache.Do(c, func(ctx context.Context) {
			if _, e := s.coinClient.ModifyCoins(ctx, &coinmdl.ModifyCoinsReq{Mid: mid, Count: s.c.Image.LotteryCoinTwoAmount, Reason: _imageReason, IP: ip}); e != nil {
				log.Error("ImageLottery need check coin.ModifyCoin mid:%d count:%f error(%v)", mid, s.c.Image.LotteryCoinTwoAmount, e)
			}
		})
	case s.c.Image.LotteryCurrOne:
		s.cache.Do(c, func(ctx context.Context) {
			if e := s.upCurrencyAmount(ctx, currency.BusinessAct, s.c.Image.Sid, 0, mid, s.c.Image.LotteryCurrOneAmount, _imageAwardRemark); e != nil {
				log.Error("ImageLottery need check upCurrencyAmount mid:%d amount:%f error(%v)", mid, s.c.Image.LotteryCoinOneAmount, e)
			}
		})
	}
	return
}

// dailyForeignID .
func (s *Service) dailyForeignID() (foreignID int64, isLegal bool) {
	var (
		key = time.Now().Format("2006-01-02")
	)
	if _, ok := s.c.Rule.TenDailys[key]; ok {
		foreignID = s.c.Rule.TenDailys[key]
		isLegal = ok
	}
	return
}

// dailyImage .
func (s *Service) dailyImage() (image string) {
	var (
		key = time.Now().Format("2006-01-02")
	)
	if _, ok := s.c.Rule.TenImage[key]; ok {
		image = s.c.Rule.TenImage[key]
	}
	return
}

// dailyH5Image .
func (s *Service) dailyH5Image() (image string) {
	var (
		key = time.Now().Format("2006-01-02")
	)
	if _, ok := s.c.Rule.TenH5Image[key]; ok {
		image = s.c.Rule.TenH5Image[key]
	}
	return
}

func (s *Service) taskPreCheck(c context.Context, mid, taskID, nowTs int64) (err error) {
	var (
		taskRule     *task.TaskRule
		preData      *task.Task
		preUserState map[string]*task.UserTask
		preTaskID    int64
	)
	if taskRule, err = s.taskDao.TaskRule(c, taskID); err != nil {
		log.Error("DoTask s.taskDao.TaskRule taskid(%d) error(%v)", taskID, err)
		return
	}
	if taskRule == nil {
		return
	}
	if preTaskID, _ = strconv.ParseInt(taskRule.PreTask, 10, 64); preTaskID > 0 {
		if preData, err = s.taskDao.Task(c, preTaskID); err != nil {
			log.Error("DoTask s.taskDao.Task(%d) error(%v)", preTaskID, err)
			return
		}
		preTaskArg := map[int64]*task.Task{
			preTaskID: preData,
		}
		if preUserState, err = s.taskDao.UserTaskState(c, preTaskArg, mid, preData.BusinessID, preData.ForeignID, nowTs); err != nil {
			log.Error("DoTask s.taskDao.UserTaskState mid(%d) arg(%v) error(%v)", mid, preData, err)
			return
		}
		preUserTask, ok := preUserState[fmt.Sprintf("%d_%d", preData.ID, preData.Round(nowTs))]
		if !ok || preUserTask == nil || preUserTask.Finish != task.HasFinish {
			err = ecode.ActivityTaskPreNotCheck
			return
		}
	}
	return
}

func (s *Service) PointLottery(c context.Context, mid int64) (data *like.Lottery, err error) {
	var (
		amount int64
		limit  int
	)
	amount = s.ActUserCurrency(c, mid, s.c.Eleven.ElevenSid)
	if amount < s.c.Eleven.AmountLimit {
		err = ecode.ActivityCurrLackAmount
		return
	}
	// 次数限制
	checkKey := fmt.Sprintf("eleven_%d_%d_%s", mid, s.c.Eleven.ElevenSid, time.Now().Format("20060102"))
	if limit, err = s.dao.RiGet(c, checkKey); err != nil {
		log.Error("PointLottery s.dao.RsGet(%v) error(%v)", checkKey, err)
		err = nil
		return
	}
	if limit >= _lotteryTimesLimit {
		err = ecode.ActivityOverLotteryMax
		return
	}
	if err = s.upCurrencyAmount(c, currency.BusinessAct, s.c.Eleven.ElevenSid, mid, 0, s.c.Eleven.AmountLimit, _elevenCurrRemark); err != nil {
		return
	}
	if data, err = s.dao.LotteryIndex(c, s.c.Eleven.LotteryID, 0, 0, mid); err != nil {
		log.Error("PointLottery need check LotteryIndex sid(%d) mid(%d) error(%v)", s.c.Eleven.LotteryID, mid, err)
		s.cache.Do(c, func(ctx context.Context) {
			if e := s.upCurrencyAmount(ctx, currency.BusinessAct, s.c.Eleven.ElevenSid, 0, mid, s.c.Eleven.AmountLimit, _elevenLottryRemark); e != nil {
				log.Error("PointLottery rollback UpCurrencyAmount mid:%d amount:%d error(%v)", mid, s.c.Eleven.AmountLimit, e)
			}
		})
		return
	}
	if data.Code != 0 {
		if data.Code == like.NoLotteryCode { //未中奖
			if _, err = s.dao.Incr(c, checkKey); err != nil {
				log.Error("PointLottery s.dao.Incr(%v) error(%v)", checkKey, err)
				err = nil
			}
			s.cache.Do(c, func(ctx context.Context) { //未中奖保底5个硬币
				if _, e := s.coinClient.ModifyCoins(ctx, &coinmdl.ModifyCoinsReq{Mid: mid, Count: s.c.Eleven.CoinNum, Reason: _imageReason, IP: ""}); e != nil {
					log.Error("PointLottery s.coin.ModifyCoin mid:%d count:%f error(%v)", mid, s.c.Eleven.CoinNum, e)
				}
			})
			return
		}
		log.Warn("PointLottery LotteryIndex sid(%d) mid(%d) error code(%d)", s.c.Eleven.LotteryID, mid, data.Code)
		s.cache.Do(c, func(ctx context.Context) {
			if e := s.upCurrencyAmount(ctx, currency.BusinessAct, s.c.Eleven.ElevenSid, 0, mid, s.c.Eleven.AmountLimit, _elevenLottryRemark); e != nil {
				log.Error("PointLottery rollback UpCurrencyAmount mid:%d amount:%d error(%v)", mid, s.c.Eleven.AmountLimit, e)
			}
		})
		return
	}
	if _, err = s.dao.Incr(c, checkKey); err != nil {
		log.Error("PointLottery s.dao.Incr(%v) error(%v)", checkKey, err)
		err = nil
	}
	return
}

func (s *Service) TaskTokenDo(c context.Context, sid, mid int64, token string) (err error) {
	var memberRly *accapi.ProfileReply
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) (e error) {
		if memberRly, e = s.accClient.Profile3(ctx, &accapi.MidReq{Mid: mid}); e != nil {
			log.Error(" s.acc.Profile3(c,&accmdl.ArgMid{Mid:%d}) error(%+v)", mid, e)
		}
		return
	})
	var isNew bool
	eg.Go(func(ctx context.Context) (e error) {
		if isNew, e = s.dao.CheckTel(ctx, mid); e != nil {
			log.Error(" s.dao.CheckTel(%d) error(%+v)", mid, e)
		}
		return
	})
	if err = eg.Wait(); err != nil {
		return
	}
	if memberRly == nil || memberRly.Profile.Level >= 1 {
		err = ecode.ActivityMidBindAlready
		return
	}
	if !isNew {
		err = ecode.ActivityTelNotPassCheck
		return
	}
	var tokenMid int64
	if tokenMid, err = s.checkToken(c, sid, token); err != nil {
		return
	}
	var list []*task.TaskItem
	if list, err = s.taskList(c, tokenMid, task.BusinessAct, sid); err != nil {
		log.Error("TaskTokenDo s.taskList(%d,%d) error(%v)", tokenMid, sid, err)
		return
	}
	var taskID int64
	// 邀请好友 逐个任务完成
	for _, v := range list {
		if v.UserFinish == task.HasFinish {
			continue
		}
		taskID = v.ID
		break
	}
	if taskID == 0 {
		err = ecode.ActivityTaskHasFinish
		return
	}
	err = s.AddLotteryTimes(c, s.c.SpringCardAct.LotterySid, mid, s.c.SpringCardAct.InviteTimesID, _archive, 0, strconv.FormatInt(mid, 10)+strconv.FormatInt(time.Now().Unix(), 10), false)
	if err != nil {
		if err == ecode.ActivityLotteryAddTimesLimit {
			err = ecode.ActivityTaskHasFinish
			return
		}
		err = nil
	}
	// 完成任务
	s.cache.Do(c, func(ctx context.Context) {
		s.DoTask(ctx, tokenMid, taskID, false)
	})
	return
}

func (s *Service) TaskCheck(c context.Context, sid int64, token string) (res *task.TaskAll, err error) {
	var mid int64
	if mid, err = s.checkToken(c, sid, token); err != nil {
		return
	}
	var list []*task.TaskItem
	if list, err = s.taskList(c, mid, task.BusinessAct, sid); err != nil {
		log.Error("TaskTokenDo s.taskList(%d,%d) error(%v)", mid, sid, err)
		return
	}
	allFinish := true
	for _, v := range list {
		if v.UserFinish != task.HasFinish {
			allFinish = false
			break
		}
	}
	res = &task.TaskAll{AllFinish: allFinish}
	return
}

func (s *Service) checkToken(c context.Context, sid int64, token string) (res int64, err error) {
	if sid != s.c.SpringCardAct.InviteSid {
		err = xecode.RequestErr
		return
	}
	var tokenInfo *like.ExtendTokenDetail
	if tokenInfo, err = s.dao.LikeExtendInfo(c, sid, token); err != nil {
		log.Error("s.dao.TaskTokenDo(%d,%s) error(%v)", sid, token, err)
		return
	}
	if tokenInfo == nil || tokenInfo.Mid == 0 || tokenInfo.Max == 0 {
		err = ecode.ActivityIDNotExists
		return
	}
	res = tokenInfo.Mid
	return
}
