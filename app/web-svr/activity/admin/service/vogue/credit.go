package vogue

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"time"

	accApi "git.bilibili.co/bapis/bapis-go/account/service"
	"go-common/library/ecode"
	"go-common/library/log"
	xtime "go-common/library/time"
	voguemdl "go-gateway/app/web-svr/activity/admin/model/vogue"

	"go-common/library/sync/errgroup.v2"
)

// List get credits information list
func (s *Service) ListCredits(c context.Context, search *voguemdl.CreditSearch) (rsp *voguemdl.CreditListRsp, err error) {
	var (
		list                              []*voguemdl.CreditData
		count                             int64
		uidList                           []int64
		batchTaskInfos                    map[int64]*voguemdl.CreditData
		users                             map[int64]*accApi.Info
		creditInviteUsersDailySumMap      map[int64]map[xtime.Time]int64
		creditCostSum                     map[int64]int64
		riskInfoMap                       map[int64]*voguemdl.RiskInfo
		usersCreditViewDayMap             map[int64]map[xtime.Time]int64
		timeCostBatchTaskInfos            int64
		timeCostUserInfos                 int64
		timeCostCreditInviteUsersDailySum int64
		timeCostCreditCostSum             int64
		timeCostBatchRiskInfo             int64
		timeCostBatchCreditViewDayMap     int64
	)
	rsp = &voguemdl.CreditListRsp{}
	listCreditsStart := time.Now().UnixNano() / 1e6

	if uidList, count, err = s.dao.TaskUsers(c, search); err != nil {
		log.Error("[ListCredits] s.dao.TaskUsers error(%v)", err)
		return
	}

	eg := errgroup.WithContext(c)
	eg.GOMAXPROCS(6)
	eg.Go(func(ctx context.Context) (e error) {
		// 用户参与任务信息
		if batchTaskInfos, e = s.batchTaskInfos(ctx, uidList); e != nil {
			log.Error("[ListCredits] Fetch users task info error, uidList:%v, error(%v)", uidList, e)
			return
		}
		timeCostBatchTaskInfos = currentTimeMicro() - listCreditsStart
		log.Info("[ListCredits] Fetch users task info done, time cost(%d)", timeCostBatchTaskInfos)
		return
	})
	eg.Go(func(ctx context.Context) (e error) {
		// 用户信息
		if users, e = s.batchUserInfos(ctx, uidList); e != nil {
			log.Error("[ListCredits] Fetch users credit view day info error, uidList:%v, error(%v)", uidList, e)
			return
		}
		timeCostBatchCreditViewDayMap = currentTimeMicro() - listCreditsStart
		log.Info("[ListCredits] Fetch users credit view day info done, time cost(%d)", timeCostBatchCreditViewDayMap)
		return
	})
	eg.Go(func(ctx context.Context) (e error) {
		// 邀请用户积分信息
		if creditInviteUsersDailySumMap, e = s.batchCreditInviteUsersDailySum(ctx, uidList); e != nil {
			log.Error("[ListCredits] Fetch users invite scores map error, uidList:%v, error(%v)", uidList, e)
			return
		}
		timeCostCreditInviteUsersDailySum = currentTimeMicro() - listCreditsStart
		log.Info("[ListCredits] Fetch users invite scores map done, time cost(%d)", timeCostCreditInviteUsersDailySum)
		return
	})
	eg.Go(func(ctx context.Context) (e error) {
		// 积分消耗信息
		if creditCostSum, e = s.batchCreditCostSum(ctx, uidList); e != nil {
			log.Error("[ListCredits] Fetch users cost error, uidList:%v, error(%v)", uidList, e)
			return
		}
		timeCostCreditCostSum = currentTimeMicro() - listCreditsStart
		log.Info("[ListCredits] Fetch users cost done, time cost(%d)", timeCostCreditCostSum)
		return
	})
	eg.Go(func(ctx context.Context) (e error) {
		// 用户风控信息
		if riskInfoMap, e = s.batchRiskInfo(ctx, uidList); e != nil {
			log.Error("[ListCredits] Fetch users risk info error, uidList:%v, error(%v)", uidList, e)
			return
		}
		timeCostBatchRiskInfo = currentTimeMicro() - listCreditsStart
		log.Info("[ListCredits] Fetch users risk info done, time cost(%d)", timeCostBatchRiskInfo)
		return
	})
	eg.Go(func(ctx context.Context) (e error) {
		// 用户观看视频信息
		if usersCreditViewDayMap, e = s.batchCreditViewDayMap(ctx, uidList); e != nil {
			log.Error("[ListCredits] Fetch users credit view day info error, uidList:%v, error(%v)", uidList, e)
			return
		}
		timeCostBatchCreditViewDayMap = currentTimeMicro() - listCreditsStart
		log.Info("[ListCredits] Fetch users credit view day info done(%d)", timeCostBatchCreditViewDayMap)
		return
	})
	if err = eg.Wait(); err != nil {
		log.Error("[ListCredits] getScore error %v", err)
		return rsp, err
	}

	// 积分每日上限配置
	configCreditLimit, err := s.ConfigCreditLimit(c)
	if err != nil {
		log.Error("[ListCredits] Fetch configCreditLimit error(%v)", err)
		return
	}

	for _, uid := range uidList {
		var (
			nickname string
			riskInfo *voguemdl.RiskInfo
			item     *voguemdl.CreditData
		)
		item = batchTaskInfos[uid]

		if user, ok := users[item.Uid]; ok {
			nickname = user.GetName()
		}
		item.NickName = nickname
		// 用户风控信息
		riskInfo = riskInfoMap[item.Uid]
		item.Risk = riskInfo.Risk
		item.RiskMsg = riskInfo.RiskMsg
		// 初始积分
		item.ScoreTotal = int64(s.c.VogueActivity.ScoreInitialValue)
		// 合并每日 邀请用户积分 和 观看视频积分，若超过当日积分上限，取上限
		// 观看视频积分
		scoreViewDailyMap := usersCreditViewDayMap[item.Uid]
		// 邀请用户所得积分
		var inviteUserDailyMap map[xtime.Time]int64
		if _, ok := creditInviteUsersDailySumMap[item.Uid]; !ok {
			inviteUserDailyMap = make(map[xtime.Time]int64)
		} else {
			inviteUserDailyMap = creditInviteUsersDailySumMap[item.Uid]
		}
		_, creditInviteScoreSum, creditViewScoreSum := calcDailySum(scoreViewDailyMap, inviteUserDailyMap, configCreditLimit)

		// 消耗积分
		if creditCostSum, ok := creditCostSum[item.Uid]; ok {
			item.ScoreCost = creditCostSum
		}
		item.ScoreInvite = creditInviteScoreSum
		item.ScoreView = creditViewScoreSum
		item.ScoreTotal = item.ScoreTotal + creditInviteScoreSum + creditViewScoreSum

		// 剩余积分
		item.ScoreRemain = item.ScoreTotal - item.ScoreCost
		// 提交礼物时间，即task表创建时间
		item.TaskStartTime = item.Ctime.Time().Format(voguemdl.TimeFormat)
		list = append(list, item)
	}
	log.Info("[ListCredits]TimeCost: Total(%d), UserTaskInfos(%d), UserInfos(%d), CreditInviteUsersDailySum(%d), CreditCostSum(%d), BatchRiskInfo(%d), BatchCreditViewDayMap(%d)", currentTimeMicro()-listCreditsStart, timeCostBatchTaskInfos, timeCostUserInfos, timeCostCreditInviteUsersDailySum, timeCostCreditCostSum, timeCostBatchRiskInfo, timeCostBatchCreditViewDayMap)

	rsp.List = list
	rsp.Page = &voguemdl.Page{
		Size:  search.Ps,
		Num:   search.Pn,
		Total: count,
	}
	return
}

func currentTimeMicro() int64 {
	return time.Now().UnixNano() / 1e6
}

// batchTaskInfos 批量获取用户参与任务信息
func (s *Service) batchTaskInfos(c context.Context, mids []int64) (batchTaskInfos map[int64]*voguemdl.CreditData, err error) {
	var (
		offset  = 0
		uidList []int64
		users   map[int64]*voguemdl.CreditData
	)
	batchTaskInfos = map[int64]*voguemdl.CreditData{}
	for {
		if offset >= len(mids) {
			break
		}
		if offset+_batchSize > len(mids) {
			uidList = mids[offset:]
		} else {
			uidList = mids[offset : offset+_batchSize]
		}

		log.Info("userList:%v", uidList)

		if users, err = s.dao.UsersTaskInfo(c, uidList); err != nil {
			log.Error("[ListCredits] Fetch users task info error while batchTaskInfos, uidList:%v, error(%v)", uidList, err)
			return
		}
		for uid, user := range users {
			batchTaskInfos[uid] = user
		}
		offset += _batchSize
		time.Sleep(_batchSleep)
	}
	return
}

func (s *Service) batchUserInfos(c context.Context, mids []int64) (batchUsers map[int64]*accApi.Info, err error) {
	var (
		offset  = 0
		uidList []int64
		users   map[int64]*accApi.Info
	)
	batchUsers = map[int64]*accApi.Info{}
	for {
		if offset >= len(mids) {
			break
		}
		if offset+_batchSize > len(mids) {
			uidList = mids[offset:]
		} else {
			uidList = mids[offset : offset+_batchSize]
		}

		log.Info("userList:%v", uidList)

		if users, err = s.dao.UserInfos(c, uidList); err != nil {
			log.Error("[ListCredits] Fetch users error while batchUserInfos, uidList:%v, error(%v)", uidList, err)
			return
		}
		for uid, user := range users {
			batchUsers[uid] = user
		}
		offset += _batchSize
		time.Sleep(_batchSleep)
	}
	return
}

func (s *Service) batchCreditInviteUsersDailySum(c context.Context, mids []int64) (creditInviteUsersDailySumMap map[int64]map[xtime.Time]int64, err error) {
	var (
		offset  = 0
		uidList []int64
		users   map[int64]map[xtime.Time]int64
	)
	creditInviteUsersDailySumMap = make(map[int64]map[xtime.Time]int64, len(mids))
	for {
		if offset >= len(mids) {
			break
		}
		if offset+_batchSize > len(mids) {
			uidList = mids[offset:]
		} else {
			uidList = mids[offset : offset+_batchSize]
		}

		log.Info("userList:%v", uidList)

		if users, err = s.dao.CreditInviteUsersDailySum(c, uidList); err != nil {
			log.Error("[ListCredits] Fetch users invite scores map error, uidList:%v, error(%v)", uidList, err)
			return
		}
		for uid, user := range users {
			creditInviteUsersDailySumMap[uid] = user
		}
		offset += _batchSize
		time.Sleep(_batchSleep)
	}
	return
}

func (s *Service) batchCreditCostSum(c context.Context, mids []int64) (creditCostSum map[int64]int64, err error) {
	var (
		offset  = 0
		uidList []int64
		users   map[int64]int64
	)
	creditCostSum = make(map[int64]int64, len(mids))
	for {
		if offset >= len(mids) {
			break
		}
		if offset+_batchSize > len(mids) {
			uidList = mids[offset:]
		} else {
			uidList = mids[offset : offset+_batchSize]
		}

		log.Info("userList:%v", uidList)

		if users, err = s.dao.CreditCostSum(c, uidList); err != nil {
			log.Error("[ListCredits] Fetch users cost error, uidList:%v, error(%v)", uidList, err)
			return
		}
		for uid, user := range users {
			creditCostSum[uid] = user
		}
		offset += _batchSize
		time.Sleep(_batchSleep)
	}
	return
}

// batchRiskInfo 批量获取用户风控信息
func (s *Service) batchRiskInfo(c context.Context, mids []int64) (riskInfoMap map[int64]*voguemdl.RiskInfo, err error) {
	var (
		risk    bool
		riskMsg string
	)
	riskInfoMap = make(map[int64]*voguemdl.RiskInfo, len(mids))

	for idx, mid := range mids {
		if idx%_batchSize == 0 {
			time.Sleep(_batchSleep)
		}
		if risk, riskMsg, err = s.dao.RiskInfo(c, mid); err != nil {
			log.Error("batchRiskInfo error(%v), user(%d)", err, mid)
			return
		}
		riskInfoMap[mid] = &voguemdl.RiskInfo{
			Risk:    risk,
			RiskMsg: riskMsg,
		}
	}

	return
}

// batchCreditViewDayMap 批量获取用户每日观看视频积分信息
func (s *Service) batchCreditViewDayMap(c context.Context, mids []int64) (resMap map[int64]map[xtime.Time]int64, err error) {
	var (
		scoreViewDailyMap map[xtime.Time]int64
	)
	resMap = make(map[int64]map[xtime.Time]int64, len(mids))
	eg := errgroup.WithContext(c)
	eg.GOMAXPROCS(2)
	eg.Go(func(ctx context.Context) (err error) {
		for _, mid := range mids {
			if scoreViewDailyMap, err = s.dao.CreditViewDayMap(ctx, mid); err != nil {
				log.Error("batchCreditViewDayMap error(%v), user(%d)", err, mid)
				return
			}
			resMap[mid] = scoreViewDailyMap
		}
		return
	})
	if err = eg.Wait(); err != nil {
		log.Error("batchCreditViewDayMap err %v", err)
		return
	}
	return
}

// calcDailySum
func calcDailySum(scoreViewDailyMap map[xtime.Time]int64, inviteUserDailyMap map[xtime.Time]int64, configCreditLimit *voguemdl.ConfigCreditLimit) (resMap map[xtime.Time]int64, creditInviteScoreSum int64, creditViewScoreSum int64) {
	log.Info("[calcDailySum] scoreViewDailyMap:(%v), inviteUserDailyMap:(%v), configCreditLimit:(%v)", scoreViewDailyMap, inviteUserDailyMap, configCreditLimit)

	creditInviteScoreSum = 0
	creditViewScoreSum = 0
	resMap = scoreViewDailyMap
	for day, n := range inviteUserDailyMap {
		if _, ok := resMap[day]; !ok {
			resMap[day] = 0
		}
		resMap[day] += n
	}
	for day, n := range resMap {
		dayUnix := day.Time().Unix()
		limit := dailyCreditLimit(dayUnix, configCreditLimit)
		log.Info("dailyCreditLimit: %s, %d", day.Time(), limit)
		if n > limit {
			log.Info("[VogueCredit] calDailySum(%d) exceed limit(%d) on (%d)", n, limit, day)
			resMap[day] = limit
		}
		if inviteUserScore, ok := inviteUserDailyMap[day]; ok {
			creditInviteScoreSum += inviteUserScore
			creditViewScoreSum += resMap[day] - inviteUserScore
		} else {
			creditViewScoreSum += resMap[day]
		}
	}
	return
}

// Export credits
func (s *Service) ExportCredits(c context.Context) (result [][]string, err error) {
	var (
		rsp *voguemdl.CreditListRsp
	)
	params := &voguemdl.CreditSearch{Uid: -1, Pn: 1, Ps: -1}
	if rsp, err = s.ListCredits(c, params); err != nil {
		log.Errorc(c, "[ListCredits] s.dao.CreditsList error(%v)", err)
		return
	}
	for _, item := range rsp.List {
		result = append(result, []string{
			item.NickName,
			strconv.FormatInt(item.Uid, 10),
			strconv.FormatInt(item.ScoreTotal, 10),
			strconv.FormatInt(item.ScoreRemain, 10),
			strconv.FormatInt(item.ScoreView, 10),
			strconv.FormatInt(item.ScoreInvite, 10),
			item.TaskStartTime,
			item.GoodsName,
			strconv.FormatInt(item.GoodsScoreSetting, 10),
			item.RiskMsg,
		})
	}
	return
}

// ExportCreditsAsync
func (s *Service) GenerateExportCreditsTask(c context.Context) (err error) {
	err = ecode.Error(ecode.RequestErr, "活动已结束，无法导出")
	return
	//if !atomic.CompareAndSwapInt64(&s.exportState, 0, 1) {
	//	log.Error("Already has export process")
	//	err = ecode.Error(ecode.RequestErr, "已经有导出执行中，请稍等勿重复点")
	//	return
	//}
	//
	//export := func(ctx context.Context) {
	//	defer atomic.CompareAndSwapInt64(&s.exportState, 1, 0)
	//	if err := s.DoExportCredits(ctx); err != nil {
	//		log.Error("s.DoExportCreditsAsync error(%+v)", err)
	//		return
	//	}
	//	log.Info("Succeed to export credits async")
	//}
	//
	//go export(context.Background())
	//return
}

func (s *Service) ExportAsyncData(c context.Context) (exportData *voguemdl.CreditExportData, err error) {
	return s.exportData, nil
}

func (s *Service) DoExportCredits(c context.Context) (err error) {
	var (
		infoStr [][]string
		ctime   = time.Now().Unix()
	)
	log.Info("[DoExportCredits]Ready to export")

	if infoStr, err = s.ExportCredits(c); err != nil {
		log.Error("[DoExportCredits]ExportCreditsAsync s.ExportCredits err(%v)", err)
		return
	}

	fileName := fmt.Sprintf("时尚活动积分进度列表_%d.csv", time.Now().Unix())
	filePath := fmt.Sprintf("/tmp/%s", fileName)
	csvFile, err := os.Create(filePath)
	if err != nil {
		log.Error("[DoExportCredits]os.Create %s error(%+v)", filePath, err)
		return
	}
	defer csvFile.Close()
	log.Info("[DoExportCredits]csvFile: %s", csvFile.Name())
	csvFile.WriteString("\xEF\xBB\xBF")
	header := []string{"昵称", "UID", "累计积分", "剩余积分", "看视频积分", "好友邀请积分", "提交礼物时间", "提交礼物名称", "礼物要求积分", "是否存在异常"}
	wr := csv.NewWriter(csvFile)
	wr.Write(header)
	for i := 0; i < len(infoStr); i++ {
		wr.Write(infoStr[i])
	}
	wr.Flush()
	log.Info("[DoExportCredits]csvFile write done")
	log.Info("[DoExportCredits]s.exportData(%+v)", s.exportData)

	s.exportData = &voguemdl.CreditExportData{
		FilePath: filePath,
		Ctime:    ctime,
		Mtime:    time.Now().Unix(),
	}
	log.Info("[DoExportCredits]s.exportData(%+v)", s.exportData)
	return
}

func (s *Service) ListCreditsDetail(c context.Context, search *voguemdl.CreditDetailSearch) (rsp *voguemdl.CreditDetailListRsp, err error) {
	var (
		list        []*voguemdl.CreditItem
		inviteList  []*voguemdl.CreditItem
		costList    []*voguemdl.CreditItem
		viewList    []*voguemdl.CreditItem
		scoreRemain int64
	)
	rsp = &voguemdl.CreditDetailListRsp{}
	if inviteList, err = s.dao.CreditInviteList(c, search); err != nil {
		log.Error("[ListCreditsDetail] s.dao.CreditInviteList error, params:%v, err(%v)", search, err)
		return
	}
	if costList, err = s.dao.CreditCostList(c, search); err != nil {
		log.Error("[ListCreditsDetail] s.dao.CreditCostList error, params:%v, err(%v)", search, err)
		return
	}
	if viewList, err = s.dao.CreditViewList(c, search); err != nil {
		log.Error("[ListCreditsDetail] s.dao.CreditViewList error, params:%v, err(%v)", search, err)
		return
	}

	list = append(inviteList, costList...)
	list = append(list, viewList...)

	// 按照Ctime列排序
	sort.Slice(list, func(i, j int) bool {
		return list[i].Ctime < list[j].Ctime
	})

	// 积分每日上限配置
	configCreditLimit, err := s.ConfigCreditLimit(c)
	if err != nil {
		log.Error("[ListCredits] Fetch configCreditLimit error(%v)", err)
		return
	}

	var dailyScore = make(map[int64]int64)

	// 计算剩余积分，视频信息提取
	for idx, item := range list {
		var dailyScoreReachLimit = false
		if idx == 0 {
			scoreRemain = int64(s.c.VogueActivity.ScoreInitialValue)
		}
		year, month, day := item.Ctime.Time().Date()
		today := time.Date(year, month, day, 0, 0, 0, 0, time.Local).Unix()
		if _, ok := dailyScore[today]; !ok {
			dailyScore[today] = 0
			dailyScoreReachLimit = false
		}
		todayLimit := dailyCreditLimit(today, configCreditLimit)
		log.Info("dailyCreditLimit: %s, %d", time.Date(year, month, day, 0, 0, 0, 0, time.Local), todayLimit)
		if item.Category == voguemdl.CategoryDeposit {
			item.ScoreSymbol = voguemdl.ScoreSymbolPositive
			// 积分已达当日上限，后续积分均给0
			if dailyScoreReachLimit {
				item.Score = 0
			} else {
				if item.Score+dailyScore[today] > todayLimit {
					item.Score = todayLimit - dailyScore[today]
					dailyScoreReachLimit = true
				}
				dailyScore[today] += item.Score
			}
			scoreRemain += item.Score

		} else {
			scoreRemain -= item.Score
			item.ScoreSymbol = voguemdl.ScoreSymbolNegtive
		}
		item.ScoreRemain = scoreRemain
		if item.Method == voguemdl.MethodView {
			var source = &voguemdl.CreditViewSource{}
			if err = json.Unmarshal([]byte(item.Detail), source); err != nil {
				log.Error("[ListCreditsDetail] CreditViewSource json.Unmarshal error, params:%s, err(%v)", item.Detail, err)
			} else {
				item.Video = strconv.FormatInt(source.Aid, 10)
			}
		}
	}

	rsp.List = list
	return
}

func dailyCreditLimit(today int64, configCreditLimit *voguemdl.ConfigCreditLimit) (creditLimit int64) {
	// 因前台state接口的双倍时间判定bug的快速修复方案为修改配置里的起始时间+1s，因此后台此处需要相应修改-1s
	log.Info("%+v", configCreditLimit)
	log.Info("configCreditLimit double start: %+s", time.Unix(configCreditLimit.ActSecondDoubleStart, 0).Format("2006-01-02 03:04:05 PM"))
	log.Info("configCreditLimit double end: %+s", time.Unix(configCreditLimit.ActSecondDoubleEnd, 0).Format("2006-01-02 03:04:05 PM"))
	if today >= configCreditLimit.ActDoubleStart-1 && today < configCreditLimit.ActDoubleEnd {
		creditLimit = configCreditLimit.DailyLimit * 2
	} else if today >= configCreditLimit.ActSecondDoubleStart-1 && today < configCreditLimit.ActSecondDoubleEnd {
		creditLimit = configCreditLimit.DailyLimit * 2
	} else {
		creditLimit = configCreditLimit.DailyLimit
	}
	return
}

// Export credits detail
func (s *Service) ExportCreditsDetail(c context.Context, params *voguemdl.CreditDetailSearch) (result [][]string, err error) {
	var (
		rsp *voguemdl.CreditDetailListRsp
	)
	if rsp, err = s.ListCreditsDetail(c, params); err != nil {
		log.Errorc(c, "[ListCreditsDetail] s.dao.ListCreditsDetail error(%v)", err)
		return
	}
	for _, item := range rsp.List {
		result = append(result, []string{
			item.Ctime.Time().Format(voguemdl.TimeFormat),
			voguemdl.CategoryToStr[item.Category],
			voguemdl.MethodToStr[item.Method],
			item.ScoreSymbol + strconv.FormatInt(item.Score, 10),
			strconv.FormatInt(item.ScoreRemain, 10),
			item.Video,
			item.Friend,
		})
	}
	return
}
