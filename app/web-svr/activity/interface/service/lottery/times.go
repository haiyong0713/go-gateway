package lottery

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	actplat "git.bilibili.co/bapis/bapis-go/platform/interface/act-plat-v2"
	"go-common/library/cache"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/client"
	actmdl "go-gateway/app/web-svr/activity/interface/model/actplat"
	cardsmdl "go-gateway/app/web-svr/activity/interface/model/cards"
	l "go-gateway/app/web-svr/activity/interface/model/lottery_v2"

	"github.com/pkg/errors"
)

// GetRecordByOrderNo 获取中奖记录
func (s *Service) GetRecordByOrderNo(c context.Context, id int64, orderNo string) (*l.InsertRecord, error) {
	return s.lottery.RawLotteryActionByOrderNo(c, id, orderNo)
}

// AddLotteryTimes 增加抽奖次数接口
func (s *Service) AddLotteryTimes(c context.Context, sid string, mid, cid int64, actionType, num int, orderNo string, isOut bool) (err error) {
	log.Infoc(c, "do new lottery")
	var (
		ip = metadata.String(c, metadata.RemoteIP)
	)
	lottery, timesConfig, err := s.lotteryConfig(c, sid)
	if err != nil {
		return err
	}
	// 检查抽奖信息
	if err = s.checkTimesConf(c, lottery, timesConfig); err != nil {
		return
	}
	if isOut && (actionType > l.TimesFollowType && actionType < l.TimesLikeType && actionType != l.TimesFeType) {
		return
	}
	// 验证点赞投币情况
	if actionType == l.TimesLikeType || actionType == l.TimesCoinType {
		err = s.judgeLotteryCounter(c, actionType, mid, timesConfig)
		if err != nil {
			return
		}
	}
	err = s.JudgeAddTimes(c, actionType, num, lottery.ID, mid, cid, timesConfig, orderNo, ip)
	return
}

func (s *Service) AddTimesByTask(c context.Context, sid string, mid int64, activity, actionType string) (err error) {
	timeStamp := time.Now().Unix()
	activityPoints := &actmdl.ActivityPoints{
		Timestamp: timeStamp,
		Mid:       mid,
		Source:    mid,
		Activity:  activity,
		Business:  actionType,
	}

	err = s.actplat.Send(c, mid, activityPoints)
	if err != nil {
		log.Errorc(c, "act platform send error, data(%+v) err(%v) ", activityPoints, err)
		s.cache.SyncDo(c, func(ctx context.Context) {
			var (
				i    int
				err2 error
			)

			for i = 0; i < retry; i++ {
				err2 = s.actplat.Send(ctx, mid, activityPoints)
				if err2 == nil {
					return
				}
				log.Errorc(ctx, "act platform send error, retry [%d/%d] info:%v", i, retry, err2)
				time.Sleep(timeSleep)
			}

			log.Errorc(ctx, "s.actDao.Send end error data(%+v) err(%v) ", activityPoints, err2)
			content := fmt.Sprintf("timestamp:%d,mid:%d,Activity:%s,Business:%s,retry:[%d/%d], err:%v",
				timeStamp, mid, activity, actionType, i, retry, err2)
			if err := s.wechatdao.SendWeChat(ctx, s.c.Wechat.PublicKey, "[任务平台推送失败重试]", content, "zhanghao09"); err != nil {
				log.Errorc(ctx, " s.wechatdao.SendWeChat(%v)", err)
			}
		})
	}
	return
}

func (s *Service) TaskInfo(c context.Context, mid, activityId int64) (res *cardsmdl.TaskReply, err error) {
	actList, ok := s.actTaskMap[activityId]
	if !ok || actList == nil {
		return nil, errors.Wrapf(err, "no task for activity:[%d]", activityId)
	}

	eg := errgroup.WithContext(c)
	taskList := make([]*cardsmdl.TaskMember, 0)
	res = &cardsmdl.TaskReply{}
	resTaskList := make([]*cardsmdl.TaskDetail, 0)
	res.List = resTaskList

	var (
		mutex sync.Mutex
	)
	for _, task := range actList {
		taskCounter := task.Counter
		taskFinishTimes := task.FinishTimes
		activity := task.Activity
		eg.Go(func(ctx context.Context) (err error) {
			counter, err := s.actplat.GetCounter(c, mid, activity, taskCounter)
			if err != nil {
				log.Errorc(ctx, "s.getCounter err(%v)", err)
				return err
			}
			var state int
			if counter >= taskFinishTimes {
				state = cardsmdl.StateFinish
				counter = taskFinishTimes
			}
			mutex.Lock()
			taskList = append(taskList, &cardsmdl.TaskMember{
				Count:   counter,
				State:   state,
				Counter: taskCounter,
			})
			mutex.Unlock()
			return
		})
	}

	if err = eg.Wait(); err != nil {
		log.Errorc(c, "eg.Wait error(%v)", err)
		return res, err
	}

	taskMapList := make(map[string]*cardsmdl.TaskMember)
	for _, v := range taskList {
		taskMapList[v.Counter] = v
	}
	for _, task := range actList {
		var taskMember = &cardsmdl.TaskMember{}
		if member, ok := taskMapList[task.Counter]; ok {
			taskMember = member
		}
		taskDetail := &cardsmdl.TaskDetail{
			Task: &cardsmdl.SimpleTask{
				TaskName:    task.TaskName,
				LinkName:    task.LinkName,
				Desc:        task.Desc,
				Link:        task.Link,
				FinishTimes: task.FinishTimes,
			},
			Member: taskMember,
		}
		resTaskList = append(resTaskList, taskDetail)
	}
	res.List = resTaskList
	return
}

func (s *Service) ProgressRate(c context.Context, sid string) (res *l.ProcessRate, err error) {
	if s.c.AprilFoolsAct == nil || s.c.AprilFoolsAct.Sid != sid {
		return nil, errors.New(fmt.Sprintf("sid[%v] error", sid))
	}

	res = &l.ProcessRate{}
	nowTime := time.Now().Unix()
	clueMap := make(map[string]bool)
	for idx, v := range s.c.AprilFoolsAct.Clues {
		log.Infoc(c, "ProgressRate %v , %v", nowTime, v.TimePoint.Unix())
		if nowTime < v.TimePoint.Unix() {
			if idx > 0 {
				pre := s.c.AprilFoolsAct.Clues[idx-1]
				res.Rate += float64(nowTime-pre.TimePoint.Unix()) * (v.Process - pre.Process) / float64(v.TimePoint.Unix()-pre.TimePoint.Unix())
				res.Rate, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", res.Rate), 64)
			}
			break
		}
		res.Rate = v.Process
		clueMap[v.KeyEvent] = true
	}

	for _, item := range s.cluesSrcs {
		if _, ok := clueMap[item.Title]; ok {
			clue := l.Clue{
				Item:   item,
				Status: true,
			}
			if res.Clues == nil {
				res.Clues = []l.Clue{clue}
				continue
			}
			res.Clues = append(res.Clues, clue)
		}
	}

	return
}

// LotteryCount ...
func (s *Service) LotteryCount(c context.Context, sid string, mid int64, actionType int) (num *l.CountNumReply, err error) {
	if actionType != l.TimesLikeType && actionType != l.TimesCoinType {
		return &l.CountNumReply{Num: 0}, nil
	}
	lottery, timesConfig, err := s.lotteryConfig(c, sid)
	if err != nil {
		return nil, err
	}
	// 检查抽奖信息
	if err = s.checkTimesConf(c, lottery, timesConfig); err != nil {
		return
	}
	count, err := s.getLotteryCounter(c, actionType, mid, timesConfig)
	return &l.CountNumReply{Num: count}, err

}

func (s *Service) getLotteryCounter(c context.Context, actionType int, mid int64, timesConfig []*l.TimesConfig) (count int64, err error) {
	for _, v := range timesConfig {
		if v.Type == actionType {

			info := &l.TimesInfo{}
			err = json.Unmarshal([]byte(v.Info), info)
			if err != nil {
				return
			}
			count, err = s.getCounter(c, info.Counter, info.Activity, mid)
			if err != nil {
				return
			}
			if count >= info.Count {
				return info.Count, nil
			}
			return count, nil
		}
	}
	return 0, ecode.ActivityLotteryTimesTypeError
}

func (s *Service) judgeLotteryCounter(c context.Context, actionType int, mid int64, timesConfig []*l.TimesConfig) (err error) {
	for _, v := range timesConfig {
		if v.Type == actionType {
			return s.judgeCounterTimes(c, mid, v)
		}
	}
	return ecode.ActivityLotteryTimesTypeError
}

func (s *Service) judgeCounterTimes(c context.Context, mid int64, times *l.TimesConfig) (err error) {
	info := &l.TimesInfo{}
	err = json.Unmarshal([]byte(times.Info), info)
	if err != nil {
		return
	}
	count, err := s.getCounter(c, info.Counter, info.Activity, mid)
	if err != nil {
		return
	}
	if count >= info.Count {
		return
	}
	return ecode.ActivityLotteryTimesNotEnough
}

func (s *Service) getCounter(c context.Context, counter, activity string, mid int64) (int64, error) {

	resp, err := client.ActPlatClient.GetCounterRes(c, &actplat.GetCounterResReq{
		Counter:  counter,
		Activity: activity,
		Mid:      mid,
		Time:     time.Now().Unix(),
	})
	if err != nil {
		log.Error("s.actPlatClient.GetCounterRes(%v) error(%v)", mid, err)
		err = errors.Wrapf(err, "s.actPlatClient.GetCounterRes %d", mid)
		return 0, err
	}
	if resp == nil || len(resp.CounterList) != 1 {
		log.Error("s.actPlatClient.GetCounterRes(%v) error(%v)", mid, err)
		err = errors.Wrapf(err, "s.actPlatClient.GetCounterRes %d return nil", mid)
		return 0, err
	}
	count := resp.CounterList[0]
	return count.Val, nil

}

// GetIsCanAddTimes 是否可以增加抽奖次数
func (s *Service) GetIsCanAddTimes(c context.Context, sid string, id, mid int64, actionType, num int) (*l.CountStateReply, error) {
	if actionType != l.TimesLikeType && actionType != l.TimesCoinType {
		return &l.CountStateReply{State: l.TimesAddTimesStateNone}, nil
	}
	lottery, timesConfig, err := s.lotteryConfig(c, sid)
	if err != nil {
		return nil, err
	}
	// 检查抽奖信息
	if err = s.checkTimesConf(c, lottery, timesConfig); err != nil {
		return &l.CountStateReply{State: l.TimesAddTimesStateNone}, nil
	}
	state, err := s.counterCanAddTimes(c, actionType, num, id, lottery.ID, mid, timesConfig)
	return &l.CountStateReply{State: state}, err
}

// counterCanAddTimes 能否增加抽奖次数
func (s *Service) counterCanAddTimes(c context.Context, actionType, num int, id, sid, mid int64, timesConfig []*l.TimesConfig) (state int, err error) {
	err = s.judgeLotteryCounter(c, actionType, mid, timesConfig)
	if err != nil {
		if xecode.EqualError(ecode.ActivityLotteryTimesNotEnough, err) {
			return l.TimesAddTimesStateNone, nil
		}
		return l.TimesAddTimesStateNone, err
	}
	incr, _, _, err := s.canAddTimes(c, actionType, num, sid, mid, id, timesConfig)
	if err != nil {
		if xecode.EqualError(ecode.ActivityLotteryAddTimesLimit, err) {
			return l.TimesAddTimesStateAlready, nil
		}
		return l.TimesAddTimesStateNone, err
	}
	if incr > 0 {
		return l.TimesAddTimesStateWait, nil
	}
	return l.TimesAddTimesStateAlready, nil
}

func (s *Service) canAddTimes(c context.Context, actionType, num int, sid int64, mid, id int64, timesConfig []*l.TimesConfig) (incr int, cid int64, timesKey string, err error) {
	day := time.Now().Format("2006-01-02")
	var (
		lotteryMap = make(map[string]*l.TimesConfig, len(timesConfig))
		times      = make(map[string]int)
	)
	for _, v := range timesConfig {
		key := getRecordKey(v.Type, v.ID)
		if v.Type == actionType && ((id > 0 && v.ID == id) || id == 0) {
			if v.AddType == l.DailyAddType {
				lotteryMap[getRecordKeyRedis(key, day)] = v
				timesKey = getRecordKeyRedis(key, day)
				cid = v.ID
				continue
			}
			lotteryMap[getRecordKeyRedis(key, "0")] = v
			timesKey = getRecordKeyRedis(key, "0")
			cid = v.ID
		}
	}

	if actionType > l.TimesFollowType && id > 0 {
		cid = id
	}
	// 未配置或配置行为增加次数为0
	if timesKey == "" || lotteryMap[timesKey] == nil || lotteryMap[timesKey].Times == 0 {
		return
	}
	if times, err = s.getMidAddTimes(c, timesConfig, sid, mid); err != nil {
		log.Errorc(c, "judgeAddTimes: s.getMidAddTimes sid(%d) mid(%d) error(%v)", sid, mid, err)
		return
	}
	// 增加次数达到上限
	if times[timesKey] >= lotteryMap[timesKey].Most {
		err = ecode.ActivityLotteryAddTimesLimit
		return
	}
	// incr 配置的默认一次行为增加的次数
	incr = lotteryMap[timesKey].Times
	if num > 0 {
		incr = num
	}
	// 增加次数达上限
	if times[timesKey]+incr > lotteryMap[timesKey].Most {
		incr = lotteryMap[timesKey].Most - times[timesKey]
	}
	return
}

// JudgeAddTimes 增加次数判断
func (s *Service) JudgeAddTimes(c context.Context, actionType, num int, sid int64, mid, id int64, timesConfig []*l.TimesConfig, orderNo, ip string) (err error) {
	incr, cid, timesKey, err := s.canAddTimes(c, actionType, num, sid, mid, id, timesConfig)
	if err != nil {
		return err
	}
	if incr > 0 {
		var addTimesMap = make(map[string]int)
		addTimesMap[timesKey] = incr
		if _, err = s.lottery.InsertLotteryAddTimes(c, sid, mid, actionType, incr, cid, ip, orderNo); err != nil {
			if strings.Contains(err.Error(), "Duplicate entry") {
				err = nil
				return err
			}
			log.Errorc(c, "JudgeAddTimes: s.dao.InsertLotteryAddTimes sid(%d) mid(%d) type(%d) val(%d) ip(%s) orderNo(%s) error(%v)", sid, mid, actionType, incr, ip, orderNo, err)
		}

		if err = s.lottery.IncrTimes(c, sid, mid, addTimesMap, l.AddTimesKey); err != nil {
			log.Errorc(c, "JudgeAddTimes: s.dao.IncrTimes sid(%d) mid(%d) val(%d) type(%s) error(%v)", sid, mid, incr, l.AddTimesKey, err)
		}
	}

	return
}

// lotteryConfig config
func (s *Service) lotteryConfig(c context.Context, sid string) (*l.Lottery, []*l.TimesConfig, error) {
	var (
		lottery          *l.Lottery
		lotteryTimesConf []*l.TimesConfig
	)
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) (err error) {
		if lottery, err = s.base(ctx, sid); err != nil {
			err = errors.Wrapf(err, "s.dao.Lottery %s", sid)
		}
		return
	})
	eg.Go(func(ctx context.Context) (err error) {
		if lotteryTimesConf, err = s.timesConfig(ctx, sid); err != nil {
			err = errors.Wrapf(err, "s.dao.LotteryTimesConfig %s", sid)
		}
		return
	})
	if err := eg.Wait(); err != nil {
		log.Errorc(c, "Failed to fetch lottery: %+v", err)
		return nil, nil, err
	}
	return lottery, lotteryTimesConf, nil
}

// GetMyList 获取我的抽奖记录接口
func (s *Service) GetMyList(c context.Context, sid string, pn, ps int, mid int64, needAddress bool) (res *l.RecordReply, err error) {
	log.Infoc(c, "do new lottery")
	var (
		lottery    *l.Lottery
		gift       []*l.Gift
		recordList []*l.RecordDetail
		start      = int64((pn - 1) * ps)
		end        = start + int64(ps) - 1
		ok         bool
		count      int
	)
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) (err error) {
		if lottery, err = s.base(ctx, sid); err != nil {
			err = errors.Wrapf(err, "s.dao.Lottery %s", sid)
		}
		return
	})
	eg.Go(func(ctx context.Context) (err error) {
		if gift, err = s.gift(ctx, sid); err != nil {
			err = errors.Wrapf(err, "s.dao.LotteryGift %s", sid)
		}
		return
	})
	if err = eg.Wait(); err != nil {
		log.Errorc(c, "eg.Wait error(%v)", err)
		return
	}
	if lottery.ID == 0 {
		err = ecode.ActivityNotExist
		return
	}
	if recordList, err = s.lotteryRecord(c, lottery.ID, mid, start, end); err != nil {
		log.Errorc(c, "Failed to fetch lottery record:%d %d %+v", lottery.ID, mid, err)
		return
	}
	giftMap := make(map[int64]*l.Gift)
	for _, v := range gift {
		giftMap[v.ID] = v
	}
	list := make([]*l.RecordDetail, 0)
	for _, value := range recordList {
		// 未中奖
		if value.GiftID == 0 {
			tmp := &l.RecordDetail{ID: value.ID, Mid: value.Mid, Num: value.Num, GiftID: 0, GiftName: "未中奖", ImgURL: "", Type: value.Type, GiftType: 0, Ctime: value.Ctime}
			list = append(list, tmp)
			continue
		}
		tmp := &l.RecordDetail{ID: value.ID, Mid: value.Mid, Num: value.Num, GiftID: value.GiftID, GiftName: giftMap[value.GiftID].Name, GiftType: giftMap[value.GiftID].Type, ImgURL: giftMap[value.GiftID].ImgURL, Type: value.Type, Ctime: value.Ctime}
		list = append(list, tmp)
	}
	// 判断是否填写过地址
	if needAddress {
		var addrID int64
		if addrID, err = s.lotteryAddr(c, lottery.ID, mid); err != nil {
			log.Error("GetMyList s.dao.LotteryAddr id(%d) mid(%d) error(%v)", lottery.ID, mid, err)
			return
		}
		if addrID != 0 {
			ok = true
		}
	}
	count = len(recordList)
	page := &l.Page{Num: pn, Size: ps, Total: count}
	res = &l.RecordReply{List: list, Page: page, IsAddAddress: ok}
	return
}

// lotteryRecord ..
func (s *Service) lotteryRecord(c context.Context, id, mid, start, end int64) ([]*l.RecordDetail, error) {
	lotteryRecordList, err := s.lottery.CacheLotteryActionLog(c, id, mid, start, end)
	if err != nil {
		return nil, err
	}
	if len(lotteryRecordList) == 0 {
		rawList, err := s.lottery.RawLotteryUsedTimes(c, id, mid)
		if err != nil {
			return nil, err
		}
		s.cache.SyncDo(c, func(c context.Context) {
			s.lottery.AddCacheLotteryActionLog(c, id, mid, rawList)
		})
		rawLen := int64(len(rawList))
		if rawLen < end {
			end = rawLen
		}
		if start >= rawLen {
			return nil, nil
		}
		lotteryRecordList = rawList[start:end]
	}
	return lotteryRecordList, nil
}

// lotteryAddr ...
func (s *Service) lotteryAddr(c context.Context, id int64, mid int64) (res int64, err error) {
	res, err = s.lottery.CacheLotteryAddrCheck(c, id, mid)
	if err != nil {
		log.Errorc(c, "s.lottery.CacheLotteryAddrCheck(%d,%d)", id, mid)
	}
	if err == nil {
		cache.MetricHits.Inc("LotteryAddr")
		return res, nil
	}
	cache.MetricMisses.Inc("LotteryAddr")
	res, err = s.lottery.RawLotteryAddrCheck(c, id, mid)
	if err != nil {
		return res, err
	}
	if res != 0 {
		s.cache.Do(c, func(c context.Context) {
			s.lottery.AddCacheLotteryAddrCheck(c, id, mid, res)
		})
	}
	return
}

// GetMyWinList 获取我的中奖记录接口
func (s *Service) GetMyWinList(c context.Context, sid string, mid int64, pn, ps int, needAddress bool) (res *l.RecordReply, err error) {
	log.Infoc(c, "do new lottery")
	var (
		lottery      *l.Lottery
		gift         []*l.Gift
		recordDetail []*l.MidWinList
		ok           bool
	)
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) (err error) {
		if lottery, err = s.base(ctx, sid); err != nil {
			err = errors.Wrapf(err, "s.dao.Lottery %s", sid)
		}
		return
	})
	eg.Go(func(ctx context.Context) (err error) {
		if gift, err = s.gift(ctx, sid); err != nil {
			err = errors.Wrapf(err, "s.dao.LotteryGift %s", sid)
		}
		return
	})
	if err = eg.Wait(); err != nil {
		log.Errorc(c, "eg.Wait error(%v)", err)
		return
	}
	if lottery.ID == 0 {
		err = ecode.ActivityNotExist
		return
	}
	eg2 := errgroup.WithContext(c)
	eg2.Go(func(ctx context.Context) (err error) {
		if recordDetail, err = s.myRealyWinList(ctx, lottery.ID, mid, pn, ps); err != nil {
			log.Errorc(c, "Failed to fetch lottery record:%d %d %+v", lottery.ID, mid, err)
		}
		return
	})
	// 判断是否填写过地址
	if needAddress {
		eg2.Go(func(ctx context.Context) (err error) {
			var addrID int64
			if addrID, err = s.lotteryAddr(ctx, lottery.ID, mid); err != nil {
				log.Errorc(c, "Failed to get addr: %d, %d, %+v", lottery.ID, mid, err)
				err = nil
				return
			}
			if addrID != 0 {
				ok = true
			}
			return
		})
	}
	if err = eg2.Wait(); err != nil {
		log.Errorc(c, "eg.Wait error(%v)", err)
		return
	}
	giftMap := make(map[int64]*l.Gift)
	for _, v := range gift {
		giftMap[v.ID] = v
	}
	list := make([]*l.RecordDetail, 0)
	for _, value := range recordDetail {
		if value.GiftID <= 0 {
			continue
		}
		gift, ok := giftMap[value.GiftID]
		if !ok {
			continue
		}
		item := &l.RecordDetail{
			Mid:      value.Mid,
			GiftID:   value.GiftID,
			GiftName: gift.Name,
			GiftType: gift.Type,
			ImgURL:   gift.ImgURL,
			Ctime:    value.Mtime,
		}
		list = append(list, item)
	}
	res = &l.RecordReply{List: list, IsAddAddress: ok}
	return
}

// myWinList ...
func (s *Service) myWinList(c context.Context, id, mid int64) (res []*l.RecordDetail, err error) {
	addCache := true
	res, err = s.lottery.CacheLotteryActionLog(c, id, mid, 0, -1)
	if err != nil {
		addCache = false
		err = nil
	}
	if len(res) != 0 {
		cache.MetricHits.Inc("LotteryMyWinList")
		return
	}
	cache.MetricMisses.Inc("LotteryMyWinList")
	res, err = s.lottery.RawLotteryUsedTimes(c, id, mid)
	if err != nil {
		return
	}
	miss := res
	if len(res) == 0 {
		return
	}
	if !addCache {
		return
	}
	s.cache.Do(c, func(ctx context.Context) {
		s.lottery.AddCacheLotteryActionLog(ctx, id, mid, miss)
	})
	return
}

// GetUnusedTimes 获取未抽奖次数接口
func (s *Service) GetUnusedTimes(c context.Context, sid string, mid int64) (res *l.TimesReply, err error) {
	log.Infoc(c, "do new lottery")
	var (
		addTimes  = make(map[string]int, 8)
		usedTimes = make(map[string]int, 8)
	)
	lottery, timesConf, err := s.lotteryConfig(c, sid)
	if err != nil {
		return nil, err
	}
	// 检查抽奖信息
	if err = s.checkTimesConf(c, lottery, timesConf); err != nil {
		return
	}

	usedTimes, addTimes, err = s.getMidLotteryTimes(c, timesConf, lottery.ID, mid)
	if err != nil {
		log.Errorc(c, "GetUnusedTimes: s.getMidLotteryTimes id(%d) mid(%d) error(%v)", lottery.ID, mid, err)
		return
	}
	canUsedTimes, _, _, err := s.getMidLotteryTimesAndWinTimes(c, timesConf, usedTimes, addTimes, 0)
	if err != nil {
		log.Errorc(c, "GetUnusedTimes: s.getMidLotteryTimesAndWinTimes timesConf(%v) usedTimes(%v) addTimes(%v) error(%v)", timesConf, usedTimes, addTimes, err)
		return
	}
	if canUsedTimes < 0 {
		canUsedTimes = 0
	}
	res = &l.TimesReply{Times: canUsedTimes}
	return
}

func (s *Service) CheckAddTimes(c context.Context, sid string, mid, cid int64, actionType, num int) (leftTimes int, err error) {
	log.Infoc(c, "Check Add Times , sid:[%v] , mid:[%d] ,cid:[%d] ,actionType:[%d]", sid, mid, cid, actionType)
	lottery, timesConfig, err := s.lotteryConfig(c, sid)
	if err != nil {
		return
	}

	leftTimes, id, timesKey, err := s.canAddTimes(c, actionType, num, lottery.ID, mid, cid, timesConfig)
	log.Infoc(c, "Check Add Times , timesKey:[%v] , cid:[%v] leftTimes:[%v]", timesKey, id, leftTimes)
	return
}
