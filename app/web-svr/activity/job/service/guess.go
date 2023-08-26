package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	xecode "go-common/library/ecode"
	"go-common/library/log"

	api "git.bilibili.co/bapis/bapis-go/community/service/coin"
	esportsPB "git.bilibili.co/bapis/bapis-go/esports/service"

	guemdl "go-gateway/app/web-svr/activity/interface/model/guess"
	"go-gateway/app/web-svr/activity/job/client"
	"go-gateway/app/web-svr/activity/job/model/guess"
	lmdl "go-gateway/app/web-svr/activity/job/model/like"
	"go-gateway/app/web-svr/activity/job/tool"

	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth/limiter"
	"github.com/prometheus/client_golang/prometheus"
)

type GuessStats struct {
	Succeed int64
	Failed  int64
}

const (
	_decimal         = 2
	_decimalIn       = 1
	_rankPsForBatch  = 2000
	_backReason      = "竞猜流盘返还"
	_winReason       = "竞猜获胜收入"
	_msgTypeTxt      = 1
	_taskPred        = "pred"
	_taskPredSucceed = "pred_succ"

	limitKey4Settlement             = "settlement"
	retryFailedFinishGuessWorkerKey = "retry_failed_guess_worker_name"

	notification4SettlementOfStart = "mainID(%v): 结算开始 %v"
	notification4SettlementOfEnd   = `mainID(%v): 结算结束, 结算耗时<font color=\"info\">%v</font>，请相关同事注意。\n
>子任务总数:<font color=\"comment\">%v</font> \n
>成功子任务数:<font color=\"info\">%v</font> \n
>失败子任务数:<font color=\"warning\">%v</font> \n
>结算成功用户数:<font color=\"info\">%v</font> \n
>结算失败用户数:<font color=\"warning\">%v</font> \n
>当前结算状态:<font color=\"warning\">%v</font> \n
>修复url:<font color=\"warning\">curl "127.0.0.1:7751/guess/finish?main_id=%v&result_id=%v&business=1&oid=%v"</font> \n
`
	notification4SettlementOfRetryGetListError = "结算失败自动重试: 获取失败列表失败 %v"
	notification4SettlementOfRetryEnd          = `结算失败自动重试: 重试结束, 重试耗时<font color=\"info\">%v</font>，请相关同事注意。\n
>重试任务总数:<font color=\"comment\">%v</font> \n
>子任务总数:<font color=\"comment\">%v</font> \n
>成功子任务数:<font color=\"info\">%v</font> \n
>失败子任务数:<font color=\"warning\">%v</font> \n
>结算成功用户数:<font color=\"info\">%v</font> \n
>结算失败用户数:<font color=\"warning\">%v</font> \n`
	settlementTipOfDoing            = "结算中"
	settlementTipOfDone             = "已结算"
	settlementTipOfSetClearedFailed = "设置已结算状态失败"

	qpsMetricKey4UserGuessCacheClear      = "user_guess_cache_clear"
	qpsMetricKey4UserGuessCacheClearOfS10 = "user_guess_cache_clear_4_S10"
)

var metric4Contest = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "webSvr_activity_job",
		Name:      "contest_pred",
		Help:      "activity contest predict record",
	},
	[]string{"main_id"})

var (
	limiter4Settlement *limiter.Limiter
)

func init() {
	limiter4Settlement = tollbooth.NewLimiter(2000, nil)
}

func (s *Service) loadCourseProc() {
	prometheus.MustRegister(
		metric4Contest,
	)
	for {
		s.loadContest()
		time.Sleep(time.Second * 3)
	}
}

// loadCourseProc load oids from mc.
func (s *Service) loadContest() {
	res, err := s.guessDao.ContestList(context.Background())
	if err != nil {
		log.Error("loadContestProc res(%+v) error(%+v)", res, err)
		return
	}
	if len(res) == 0 {
		log.Error("loadContestProc res(%+v) error(%+v)", res, err)
		return
	}
	s.contestMutex.Lock()
	for _, oid := range res {
		s.contestID[oid] = struct{}{}
	}
	s.contestMutex.Unlock()
	log.Info("s.contestID %+v", s.contestID)
	ctx := context.Background()
	// 设置竞猜缓存
	s.setGuessOption(ctx, res)
	s.getMainID(ctx, res)
}

// GetMainID get mainID from binlog.
func (s *Service) getMainID(c context.Context, oids []int64) {
	var (
		list []*guemdl.MainGuess
		err  error
	)
	if list, err = s.guessDao.MainList(c, oids); err != nil {
		return
	}
	for _, l := range list {
		if _, ok1 := s.contestID[l.Oid]; ok1 {
			if _, ok2 := s.mainID[l.ID]; !ok2 {
				s.contestMutex.Lock()
				s.mainID[l.ID] = struct{}{}
				s.mainContestID[l.ID] = l.Oid
				s.contestMutex.Unlock()
			}
		}
	}
	log.Info("s.mainID %+v", s.mainID)
	log.Info("s.mainContestID %+v", s.mainContestID)
}

func (s *Service) setGuessOption(ctx context.Context, oids []int64) {
	detailOptions, err := s.guessDao.RawMDsGuess(ctx, oids)
	count := len(detailOptions)
	if err != nil || count == 0 {
		log.Errorc(ctx, "loadContestProc loadContest setGuessOption count(%d) error(%+v)", count, err)
		return
	}
	s.retry(ctx, func() error {
		return s.guessDao.SetCacheDetailOption(ctx, detailOptions)
	})
}

func (s *Service) upUserGuess(c context.Context, newMsg json.RawMessage) {
	var (
		err      error
		newGuess = new(guess.GuessUser)
	)
	if err = json.Unmarshal(newMsg, newGuess); err != nil {
		log.Error("upUserGuess json.Unmarshal(%s) error(%+v)", newMsg, err)
		return
	}
	s.delUserGuessCache(c, newGuess.Mid)
}

func (s *Service) delUserGuessCache(c context.Context, mid int64) {
	// 删除用户春季赛竞猜记录
	s.retry(c, func() error {
		return s.guessDao.DelCacheUserGuessOid(c, mid)
	})
}

func (s *Service) pubPredictMsg(c context.Context, new json.RawMessage) {
	var (
		newMain guess.GuessUser
		err     error
	)
	if err = json.Unmarshal(new, &newMain); err != nil {
		log.Error("UserTablePubPredictMsg:json.Unmarshal(%s) error(%v)", new, err)
		return
	}
	if _, ok := s.mainID[newMain.MainID]; ok {
		metric4Contest.WithLabelValues([]string{strconv.Itoa(int(newMain.MainID))}...).Inc()
		_ = s.guessDao.AddUserListCache(c, newMain.Mid, s.mainContestID[newMain.MainID], 0)
	} else {
		metric4Contest.WithLabelValues([]string{"other"}...).Inc()
		log.Info("s.pubPredictMsg not s10 data: %d", newMain.MainID)
	}
}

// BalanceGuess  balance guess stake .
func (s *Service) BalanceGuess(c context.Context, new, old json.RawMessage) {
	var (
		newMain, oldMain guess.MainMsg
		err              error
	)
	if err = json.Unmarshal(new, &newMain); err != nil {
		log.Error("FinishGuess:json.Unmarshal(%s) error(%v)", new, err)
		return
	}
	if err = json.Unmarshal(old, &oldMain); err != nil {
		log.Error("FinishGuess:json.Unmarshal(%s) error(%v)", new, err)
		return
	}
	if oldMain.IsDeleted == 0 && newMain.IsDeleted == 1 && newMain.GuessCount > 0 { //  back stake
		go s.BackGuess(newMain.ID, newMain.Business, newMain.Oid, newMain.Title)
	} else if newMain.IsDeleted == 0 && oldMain.ResultID == 0 && newMain.ResultID > 0 { // finish stake
		go s.FinishGuess(newMain.ID, newMain.ResultID, newMain.Business, newMain.Oid, false)
	}
}

// DelGuessCache  delete cache .
func (s *Service) DelGuessCache(c context.Context, new json.RawMessage) {
	var (
		newMain guess.MainMsg
		err     error
	)
	if err = json.Unmarshal(new, &newMain); err != nil {
		log.Error("DelGuessCache:json.Unmarshal(%s) error(%v)", new, err)
		return
	}
	// del oid cache
	s.guessDao.DelGuessCache(c, newMain.Oid, newMain.Business, newMain.ID)
}

// BackGuess back guess stake.
func (s *Service) BackGuess(mainID, business, oid int64, title string) {
	var (
		i    int64
		list []*guess.GuessUser
		err  error
		c    = context.Background()
	)
	// update detail odds
	if _, err = s.DetailOdds(mainID, 0); err != nil {
		log.Warn("s.DetailOdds mainID(%d) error(%+v)", mainID, err)
	}
	for i = 0; i < 100; i++ {
		if list, err = s.guessDao.GuessFinish(context.Background(), i, mainID); err != nil {
			log.Error("backGuess s.guessDao.GuessBack mainID(%d) i(%d) error(%+v)", mainID, i, err)
			time.Sleep(200 * time.Millisecond)
			continue
		}
		if len(list) == 0 {
			continue
		}
		var ids []int64
		for _, ug := range list {
			count := float64(ug.Stake)
			if _, err = s.coinClient.ModifyCoins(context.Background(), &api.ModifyCoinsReq{Mid: ug.Mid, Count: count, Reason: _backReason}); err != nil {
				log.Error("backGuess coin error s.coinClient.ModifyCoins mid(%d) coin(%v) error(%v)", ug.Mid, count, err)
				time.Sleep(200 * time.Millisecond)
				continue
			}
			ids = append(ids, ug.ID)
			msg := fmt.Sprintf(s.c.GuessImMsg.Content, ug.Ctime, title, ug.Stake)
			if _, err = s.guessDao.SendImMsg(c, &guess.ImMsgParam{
				SenderUID: uint64(s.c.GuessImMsg.OfficialUID),
				Content:   msg,
				MsgType:   _msgTypeTxt,
				RecverIDs: []uint64{uint64(ug.Mid)},
			}); err != nil {
				log.Error("backGuess s.guessDao.SendImMsg(%d) err(%+v)", ug.Mid, err)
				time.Sleep(time.Duration(s.c.Interval.CoinInterval))
				continue
			}
			log.Info("backGuess  s.coinClient.ModifyCoins mid(%d) coin(%v)", ug.Mid, count)
			time.Sleep(time.Duration(s.c.Interval.CoinInterval))
			s.guessDao.DelUserCache(c, ug.Mid, ug.StakeType, business, ug.MainID)
		}
		if _, err = s.guessDao.UpUserStatus(c, i, ids); err != nil {
			log.Error("backGuess s.guessDao.UpUserStatus i(%d) ids(%+v) error(%+v)", i, ids, err)
		}
	}
	// del oid cache
	s.guessDao.DelGuessCache(c, oid, business, mainID)
}

func calculateAddCoins(ug *guess.GuessUser, odds float64) float64 {
	return decimal(float64(ug.Stake)*odds, _decimalIn)
}

func (s *Service) addStake(i, business, resultID int64, list []*guess.GuessUser, odds float64, oid int64) (stats GuessStats) {
	c := context.Background()
	_, courseExist := s.contestID[oid]
	for _, ug := range list {
		for {
			if limiter4Settlement.LimitReached(limitKey4Settlement) {
				time.Sleep(10 * time.Millisecond)

				continue
			}

			goto SETTLEMENT
		}

	SETTLEMENT:
		tool.Metric4CommonQps.WithLabelValues([]string{limitKey4Settlement}...).Inc()

		var incomeCount float64
		if ug.DetailID == resultID {
			count := decimal(float64(ug.Stake)*odds, _decimalIn)
			if _, err := s.coinClient.ModifyCoins(
				context.Background(),
				&api.ModifyCoinsReq{
					Mid:      ug.Mid,
					Count:    count,
					Reason:   _winReason,
					UniqueID: fmt.Sprintf("%v_%v_%v", oid, resultID, ug.Mid),
					Caller:   "esports",
					Operator: "LeeLei",
				}); err != nil {
				// if error code is 34006, it means has been modified
				code := xecode.Cause(err).Code()
				log.Error("ModifyCoinsReq mid(%d) coin(%v) error(%v) code(%v)", ug.Mid, count, err, code)
				if code != 34006 {
					stats.Failed++
					log.Error("addStake coin error s.coinClient.ModifyCoins mid(%d) coin(%v) error(%v)", ug.Mid, count, err)
					continue
				}

				err = nil
			}

			if courseExist {
				var resetS10CacheErr error
				for i := 0; i < 3; i++ {
					resetS10CacheErr = s.guessDao.AddUserListCache(context.Background(), ug.Mid, oid, count)
					if resetS10CacheErr == nil {
						break
					}
				}

				if resetS10CacheErr != nil {
					tool.Metric4CommonQps.WithLabelValues([]string{qpsMetricKey4UserGuessCacheClearOfS10}...).Inc()
					stats.Failed++

					continue
				}
			}
			log.Info("addStake s.coinRPC.ModifyCoin mid(%d) coin(%v)", ug.Mid, count)
			incomeCount = count - float64(ug.Stake)
		}

		var resetGuessStatusErr error
		for i := 0; i < 3; i++ {
			resetGuessStatusErr = s.guessDao.UpdateUserGuessRelations(context.Background(), ug.Mid, business, ug.ID, incomeCount)
			if resetGuessStatusErr == nil {
				break
			}
		}

		if resetGuessStatusErr == nil {
			stats.Succeed++
			if cacheErr := s.guessDao.DelUserCache(c, ug.Mid, ug.StakeType, business, ug.MainID); cacheErr != nil {
				tool.Metric4CommonQps.WithLabelValues([]string{qpsMetricKey4UserGuessCacheClear}...).Inc()
			}
		} else {
			log.Error(
				"addStake UpdateUserGuessRelations mid(%d) business(%v) guessID(%v) coin(%v) error(%v)",
				ug.Mid,
				business,
				ug.ID,
				incomeCount,
				resetGuessStatusErr)
			stats.Failed++
		}
	}

	return
}

func (s *Service) GuessCompensation(ctx context.Context, midList []int64, mainID, resultID, business, oid int64) (m map[string]interface{}) {
	m = make(map[string]interface{}, 0)
	{
		m["mid_list"] = midList
		m["main_id"] = mainID
		m["result_id"] = resultID
		m["business"] = business
		m["oid"] = oid
		m["error"] = ""
	}

	var (
		resultOdds                                  float64
		succeedCount, failedCount, failedBatchCount int64
		err                                         error
	)

	resultOdds, err = s.DetailOdds(mainID, resultID)
	if err != nil {
		return
	}

	midM := rebuildMidMap(midList)
	for k, v := range midM {
		records, fetchErr := s.guessDao.SingleTableGuessRecord(ctx, v, mainID)
		err = fetchErr

		if fetchErr == nil && len(records) > 0 {
			stats := s.addStake(k, business, resultID, records, resultOdds, oid)
			{
				succeedCount = succeedCount + stats.Succeed
				failedCount = failedCount + stats.Failed
			}
		}
	}

	m["succeed"] = succeedCount
	m["failed"] = failedCount
	m["batch_failed_count"] = failedBatchCount
	if err != nil {
		m["error"] = err.Error()
	}

	return
}

func tableSuffixByMid(mid int64) int64 {
	return mid % 100
}

func rebuildMidMap(midList []int64) map[int64][]int64 {
	m := make(map[int64][]int64)
	for _, v := range midList {
		k := tableSuffixByMid(v)
		if d, ok := m[k]; ok {
			d = append(d, v)
		} else {
			m[k] = []int64{v}
		}
	}

	return m
}

// FinishGuess finish stake.
func (s *Service) FinishGuess(mainID, resultID, business, oid int64, debugForRetry bool) {
	var (
		resultOdds float64
		err        error
		c          = context.Background()
		wg         = new(sync.WaitGroup)
	)
	log.Infoc(c, "FinishGuess begin (%s)", time.Now().Format("2006-01-02 15:04:05"))
	_ = s.guessDao.SetGuessAsInProcess(c, mainID, oid)
	// update detail odds
	if resultOdds, err = s.DetailOdds(mainID, resultID); err != nil {
		log.Error("s.DetailOdds mainID(%d) resultID(%d) error(%+v)", mainID, resultID, err)
		return
	}

	var (
		succeedSubTaskCount, failedSubTaskCount, succeedUserCount, failedUserCount int64
		clearedTip                                                                 = settlementTipOfDoing
	)

	now := time.Now()
	defer func() {
		notificationOfEnd := fmt.Sprintf(
			notification4SettlementOfEnd,
			mainID,
			time.Since(now),
			succeedSubTaskCount+failedSubTaskCount,
			succeedSubTaskCount,
			failedSubTaskCount,
			succeedUserCount,
			failedUserCount,
			clearedTip,
			mainID,
			resultID,
			oid)
		log.Infoc(c, "FinishGuess %v", notificationOfEnd)
		anyTaskFail := failedSubTaskCount != 0 || failedUserCount != 0
		if bs, err := tool.GenAlarmMsgDataByType(tool.AlarmMsgTypeOfMarkdown, notificationOfEnd, anyTaskFail); err == nil {
			_ = tool.SendCorpWeChatRobotAlarm(bs)
		}
	}()

	notificationOfStart := fmt.Sprintf(notification4SettlementOfStart, mainID, now.Format("2006-01-02 15:04:05"))
	if bs, err := tool.GenAlarmMsgDataByType(tool.AlarmMsgTypeOfText, notificationOfStart, false); err == nil {
		_ = tool.SendCorpWeChatRobotAlarm(bs)
	}

	for i := 0; i < 100; i++ {
		wg.Add(1)
		idx := int64(i)
		go func() {
			//debug mode: set all task to fail
			if debugForRetry {
				_ = s.guessDao.AddFinishGuessFailTask(context.Background(), guess.FinishGuessFailTask{
					MainID:     mainID,
					ResultID:   resultID,
					Business:   business,
					Oid:        oid,
					TableIndex: idx,
					Odds:       resultOdds,
				})
			} else { //standard mode: doing finish
				fsb, ssb, suc, fuc := s.FinishGuessByUserIndex(mainID, resultID, business, oid, idx, resultOdds)
				atomic.AddInt64(&failedSubTaskCount, fsb)
				atomic.AddInt64(&succeedSubTaskCount, ssb)
				atomic.AddInt64(&succeedUserCount, suc)
				atomic.AddInt64(&failedUserCount, fuc)
				//some task is fail, put it to retry queue
				if fsb != 0 || fuc != 0 {
					log.Errorc(c, "FinishGuess task failed, add it to db: [%v:%v:%v:%v:%:%v]", mainID, resultID, business, oid, idx, resultOdds)
					_ = s.guessDao.AddFinishGuessFailTask(context.Background(), guess.FinishGuessFailTask{
						MainID:     mainID,
						ResultID:   resultID,
						Business:   business,
						Oid:        oid,
						TableIndex: idx,
						Odds:       resultOdds,
					})
				}
			}
			wg.Done()
		}()
	}

	wg.Wait()
	// del oid cache
	_ = s.guessDao.DelGuessCache(c, oid, business, mainID)
	if err := s.guessDao.SetGuessAsCleared(c, mainID, oid); err != nil {
		clearedTip = settlementTipOfSetClearedFailed
	} else {
		clearedTip = settlementTipOfDone
	}

	var err2UpdateGuessVersion error
	req := new(esportsPB.UpdateSeasonGuessVersionRequest)
	{
		req.MatchId = oid
	}
	for i := 0; i < 10; i++ {
		_, err2UpdateGuessVersion = client.EsportsClient.UpdateSeasonGuessVersion(context.Background(), req)
		if err2UpdateGuessVersion == nil {
			break
		}
	}

	if err2UpdateGuessVersion != nil {
		log.Infoc(c, "UpdateGuessVersion failed(%v)", err2UpdateGuessVersion)
	}

	log.Infoc(c, "FinishGuess end (%s)", time.Now().Format("2006-01-02 15:04:05"))
}

// DetailOdds detail result odds.
func (s *Service) DetailOdds(mainID, resultID int64) (rs float64, err error) {
	var (
		details    []*guemdl.DetailGuess
		detailOdds map[int64]float64
	)
	if details, err = s.guessDao.GuessDetail(context.Background(), mainID); err != nil {
		log.Error("s.guessDao.GuessDetail mainID(%d) error(%+v)", mainID, err)
		return
	}
	if detailOdds, err = s.calcOdds(details); err != nil {
		log.Error("s.calcOdds mainID(%d) resultID(%d)  error(%v)", mainID, resultID, err)
		return
	}
	log.Info("s.DetailOdds mainID(%d) resultID(%d)  detailOdds(%+v)", mainID, resultID, detailOdds)
	rs = detailOdds[resultID]
	return
}

func (s *Service) calcOdds(details []*guemdl.DetailGuess) (detailOdds map[int64]float64, err error) {
	var (
		detailStakes map[int64]int64
	)
	detailOdds = make(map[int64]float64, len(details))
	detailStakes = make(map[int64]int64, len(details))
	for _, detail := range details {
		detailStakes[detail.ID] = detail.TotalStake
	}
	log.Info("calcOdds detailStakes(%v)", detailStakes)
	for id, stake := range detailStakes {
		if stake > 0 {
			odds := decimal(1+(s.otherStakes(id, detailStakes)/float64(stake))*s.c.Rule.GuessPercent, _decimal)
			if odds > s.c.Rule.GuessMaxOdds {
				odds = s.c.Rule.GuessMaxOdds
			}
			detailOdds[id] = odds
		}
	}
	// update  detail table odds
	if _, err = s.guessDao.UpDetailOdds(context.Background(), detailOdds); err != nil {
		log.Error("s.guessDao.UpDetailOdds  error(%v)", err)
	}
	return
}

func (s *Service) otherStakes(id int64, dStakes map[int64]int64) (rs float64) {
	var stakes int64
	for k, stake := range dStakes {
		if k != id {
			stakes = stakes + stake
		}
	}
	return float64(stakes)
}

func decimal(f float64, n int) float64 {
	n10 := math.Pow10(n)
	return math.Trunc((f+0.5/n10)*n10) / n10
}

// CalcRank calculate rank.
func (s *Service) CalcRank(business int64) {
	var (
		c                       = context.Background()
		users                   []*guess.UserLog
		err                     error
		rank, i, rate, interval int64
		count                   int
	)
	if count, err = s.guessDao.UserLogCount(c, business); err != nil {
		log.Error("CalcRank s.guessDao.UserLogCount business(%d) error(%+v)", business, err)
		return
	}
	ps := math.Ceil(float64(count) / float64(_rankPsForBatch))
	log.Info("CalcRank ps(%v)  business(%d)", ps, business)
	interval = 1
	for i = 1; i <= int64(ps); i++ {
		if users, err = s.userLogs(business, i); err != nil {
			log.Error("CalcRank s.userLogs business(%d) i(%d) error(%+v)", business, i, err)
			continue
		}
		if len(users) == 0 {
			break
		}
		mapRank := make(map[int64]int64, _rankPsForBatch)
		for _, user := range users {
			if user.SuccessRate == rate {
				interval += 1
			} else {
				rank += interval
				interval = 1
			}
			rate = user.SuccessRate
			mapRank[user.ID] = rank
			s.guessDao.DelStatCache(c, user.Mid, user.StakeType, user.Business)
		}
		log.Info("CalcRank  s.guessDao.UpUserRank business(%d) i(%d) mapRank(%+v)", business, i, mapRank)
		if _, err = s.guessDao.UpUserRank(c, mapRank); err != nil {
			log.Error("CalcRank  s.guessDao.UpUserRank business(%d) i(%d) mapRank(%+v) error(%+v)", business, i, mapRank, err)
		}
		time.Sleep(200 * time.Millisecond)
	}
}

func (s *Service) userLogs(business, page int64) (rs []*guess.UserLog, err error) {
	offset := (page - 1) * _rankPsForBatch
	for i := 0; i < _retryTimes; i++ {
		if rs, err = s.guessDao.UserRank(context.Background(), business, offset, _rankPsForBatch); err == nil {
			break
		}
		time.Sleep(time.Millisecond * 200)
	}
	if err != nil {
		log.Error("userLogs s.guessDao.UserRank business(%d) offset(%d) error(%+v)", business, offset, err)
	}
	return
}

func (s *Service) guessRank() {
	for _, business := range s.c.Rule.GuessBusiness {
		if business > 0 {
			go s.CalcRank(int64(business))
			time.Sleep(time.Second)
		}
	}
}

func (s *Service) loadGuessOidsproc() {
	var (
		c         = context.Background()
		oidList   []int64
		err       error
		mainLists []*guemdl.MainGuess
	)
	if oidList, err = s.guessDao.OidList(c, s.c.Rule.GuessSID); err != nil {
		log.Error("s.guessDao.OidList sid(%d) error(%+v)", s.c.Rule.GuessSID, err)
		return
	}
	if len(oidList) == 0 {
		return
	}
	if mainLists, err = s.guessDao.MainList(c, oidList); err != nil {
		log.Error("s.guessDao.OidList sids(%v) error(%+v)", oidList, err)
		return
	}
	count := len(mainLists)
	log.Warn("loadGuessOidsproc mainLists sids(%v) count(%+v)", s.c.Rule.GuessSID, count)
	if count == 0 {
		return
	}
	tmp := make(map[int64]struct{}, count)
	for _, main := range mainLists {
		if main != nil && main.ID > 0 {
			tmp[main.ID] = struct{}{}
		}
	}
	s.guessSesaon = tmp
	log.Info("loadGuessOidsproc success()")
}

func (s *Service) addGuessLotteryTimes(c context.Context, new json.RawMessage) {
	var (
		newUser guess.GuessUser
		err     error
	)
	if err = json.Unmarshal(new, &newUser); err != nil {
		log.Error("AddGuessLotteryTimes:json.Unmarshal(%s) error(%v)", new, err)
		return
	}
	if _, ok := s.guessSesaon[newUser.MainID]; !ok {
		return
	}
	if cfg, ok := s.lotteryAdds[s.c.Rule.GuessSID]; ok {
		countKey := guessCountKey(newUser.Mid, s.c.Rule.GuessSID)
		if _, err := s.dao.Incr(c, countKey, cfg.Expire); err != nil {
			log.Error("addGuessLotteryTimes s.dao.Incr(%s,%d) error(%v)", countKey, cfg.Expire, err)
			return
		}
	} else {
		log.Error("addGuessLotteryTimes s.lotteryAdds not ok sid(%d) mid(%d) ObjID(%d)", s.c.Rule.GuessSID, newUser.Mid, newUser.ID)
		return
	}
	s.lotteryActionch <- &lmdl.LotteryMsg{MissionID: s.c.Rule.GuessSID, Mid: newUser.Mid, ObjID: newUser.ID}
	log.Info("AddGuessLotteryTimes success GuessSID:%d Mid:%d ObjID:%d", s.c.Rule.GuessSID, newUser.Mid, newUser.ID)
}

func guessCountKey(mid, sid int64) string {
	return fmt.Sprintf("guess_c_k_%d_%d", mid, sid)
}

func (s *Service) RetryFailedFinishGuessLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	ctx := context.Background()
	for range ticker.C {
		list, err := s.guessDao.GetAllFinishGuessFailTask(ctx)
		if err != nil {
			//sent alert for get fail list error
			if bs, err := tool.GenAlarmMsgDataByType(tool.AlarmMsgTypeOfMarkdown, fmt.Sprintf(notification4SettlementOfRetryGetListError, err), true); err == nil {
				log.Errorc(ctx, "RetryFailedFinishGuess: send message via WeChat error: %v", tool.SendCorpWeChatRobotAlarm(bs))
			}
			log.Errorc(ctx, "RetryFailedFinishGuess get list error: %v", err)
			continue
		}
		if len(list) == 0 {
			continue
		}
		now := time.Now()
		log.Infoc(ctx, "RetryFailedFinishGuess start")
		var failedSubTaskCount int64
		var succeedSubTaskCount int64
		var failedUserCount int64
		var succeedUserCount int64
		//start retry for every failed task
		for _, task := range list {
			_ = s.guessDao.SetGuessAsInProcess(ctx, task.MainID, task.Oid)
			// update detail odds
			fsb, ssb, suc, fuc := s.FinishGuessByUserIndex(task.MainID, task.ResultID, task.Business, task.Oid, task.TableIndex, task.Odds)
			if fsb != 0 || fuc != 0 { //retry fail, do some log
				log.Errorc(ctx, "RetryFailedFinishGuess for id(%v) still failed after retry", task.Id)
			} else { //retry success, mark task as done
				_ = s.guessDao.MarkFinishGuessFailTaskAsDone(ctx, task.Id)
			}
			failedSubTaskCount += fsb
			succeedSubTaskCount += ssb
			failedUserCount += fuc
			succeedUserCount += suc
			_ = s.guessDao.SetGuessAsCleared(ctx, task.MainID, task.Oid)
		}
		notificationOfEnd := fmt.Sprintf(
			notification4SettlementOfRetryEnd,
			time.Since(now),
			len(list),
			succeedSubTaskCount+failedSubTaskCount,
			succeedSubTaskCount,
			failedSubTaskCount,
			succeedUserCount,
			failedUserCount)
		log.Infoc(ctx, "RetryFailedFinishGuess %v", notificationOfEnd)
		anyTaskFail := failedUserCount != 0 || failedSubTaskCount != 0
		if bs, err := tool.GenAlarmMsgDataByType(tool.AlarmMsgTypeOfMarkdown, notificationOfEnd, anyTaskFail); err == nil {
			log.Errorc(ctx, "RetryFailedFinishGuess: send message via WeChat error: %v", tool.SendCorpWeChatRobotAlarm(bs))
		}
		log.Infoc(ctx, "RetryFailedFinishGuess end")
	}
}

// FinishGuessByUserIndex: user guess log is split into 100 tables, this func is used for finish specified table
// tableIndex=1 will finish all record in table `act_guess_user_01`
func (s *Service) FinishGuessByUserIndex(mainID, resultID, business, oid, tableIndex int64, odds float64) (
	failedSubTaskCount int64,
	succeedSubTaskCount int64,
	succeedUserCount int64,
	failedUserCount int64) {

	var startID int64

	for {
		list, err := s.guessDao.GuessFinishByLimit(context.Background(), tableIndex, mainID, startID)
		log.Error("GuessFinishByLimit_%v_%v_%v startID(%v)", mainID, resultID, tableIndex, startID)
		if err != nil {
			atomic.AddInt64(&failedSubTaskCount, 1)
			log.Error("s.guessDao.GuessFinish mainID(%d) resultID(%d) i(%d) error(%+v)", mainID, resultID, tableIndex, err)
			return
		}

		atomic.AddInt64(&succeedSubTaskCount, 1)
		listLen := len(list)
		if listLen == 0 {
			return
		}

		startID = list[len(list)-1].ID
		stats := s.addStake(tableIndex, business, resultID, list, odds, oid)
		{
			atomic.AddInt64(&succeedUserCount, stats.Succeed)
			atomic.AddInt64(&failedUserCount, stats.Failed)
		}

		if listLen < 1000 {
			return
		}
	}
}
