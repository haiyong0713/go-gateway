package like

import (
	"context"
	"encoding/json"
	"fmt"
	"go-gateway/app/web-svr/activity/interface/client"
	"strconv"
	"strings"
	"time"

	garbapi "git.bilibili.co/bapis/bapis-go/garb/service"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/model/like"
	"go-gateway/app/web-svr/activity/interface/model/question"
)

var (
	_emptyQuestionList = make([]*like.QuestionItem, 0)
	_emptyRank         = make([]*like.UserRank, 0)
)

const (
	_userDefaultHP = 3
	_userHPOver    = -1
	_userRight     = 1
	_resultScene   = "s10_answer"
	_riskCaller    = "activity.service"
)

func (s *Service) checkAct() error {
	nowTs := time.Now().Unix()
	if nowTs > s.c.S10Answer.EndTime {
		return ecode.ActivityOverEnd
	}
	return nil
}

func (s *Service) AnswerUserInfo(ctx context.Context, mid int64) (res *like.AnswerUserInfo, err error) {
	if res, err = s.userAnswer(ctx, mid); err != nil {
		log.Errorc(ctx, "UserInfo s.dao.CacheUserInfo mid(%d) error(%+v)", mid, err)
		return
	}
	var pendantRule *like.PendantRule
	if pendantRule, err = s.dao.CachePendantRule(ctx, mid); err != nil {
		log.Errorc(ctx, "AnswerResult s.dao.CachePendantRule mid(%d) error(%+v)", mid, err)
		return
	}
	// 查询用户当前排行
	res.UserRank = 0
	for _, rank := range s.answerRank {
		if rank.Account.Mid == mid {
			res.UserRank = int64(rank.OrderNumber)
		}
	}
	if pendantRule != nil {
		res.CanPendant = pendantRule.CanPendant
		res.HavePendant = pendantRule.HavePendant
		res.KnowRule = pendantRule.KnowRule
	}
	return
}

func (s *Service) userAnswer(ctx context.Context, mid int64) (res *like.AnswerUserInfo, err error) {
	var strWeek string
	if strWeek, err = s.getUserInfoKey(); err != nil {
		return nil, err
	}
	if res, err = s.dao.CacheUserInfo(ctx, mid, strWeek); err != nil {
		log.Errorc(ctx, "UserInfo s.dao.CacheUserInfo mid(%d) error(%+v)", mid, err)
		return
	}
	if res == nil {
		res = &like.AnswerUserInfo{LastInfo: &like.UserLastInfo{}}
	}
	return
}

func (s *Service) AnswerQuestion(ctx context.Context, mid int64, params *like.ParamQuestion) (res *like.AnswerQuestion, err error) {
	var (
		currentBaseID    int64
		idsSlice, resIDs []int64
		questionList     []*like.QuestionItem
		start            = (params.Pn - 1) * params.Ps
		end              = start + params.Ps - 1
		knowRule         int64
		check            bool
		userInfo         *like.AnswerUserInfo
		strWeek          string
		nxKey            = fmt.Sprintf("question_list_%d", mid)
	)
	defer func() {
		if err != nil && !xecode.EqualError(ecode.ActivityRapid, err) {
			s.dao.RsDelNX(ctx, nxKey)
		}
	}()
	if check, err = s.dao.RsSetNX(ctx, nxKey, s.c.S10Answer.QuestionInterval); err != nil || !check {
		log.Warn("Question s.dao.RsSetNX mid:%d to fast err:%v", mid, err)
		err = ecode.ActivityRapid
		return
	}
	if knowRule, err = s.dao.CacheKnowRule(ctx, mid); err != nil {
		log.Errorc(ctx, "Question s.dao.CacheKnowRule mid(%d) error(%+v)", mid, err)
		return
	}
	res = &like.AnswerQuestion{
		KnowRule: knowRule,
	}
	if err = s.checkAct(); err != nil {
		return
	}
	nowTime := time.Now().Unix()
	for _, answerRound := range s.c.S10Answer.AnswerRound {
		if nowTime > answerRound.RoundDate {
			currentBaseID = answerRound.BaseID
			strWeek = time.Unix(answerRound.RoundDate, 0).Format("20060102")
			break
		}
	}
	currentDetails, ok := s.answerQuestionDetails[currentBaseID]
	if !ok || len(currentDetails) == 0 || strWeek == "" {
		err = ecode.ActivityQuestionNotStart
		return
	}
	// 设置baseID
	res.BaseID = currentBaseID
	if params.CurrentRound == 0 {
		nowTs := time.Now().Unix()
		// 获取前一秒的题库
		currentRound := nowTs - 1
		// 获取当前轮
		res.CurrentRound = currentRound
		res.UserHp = _userDefaultHP
		if err = s.dao.AddCacheUserHp(ctx, mid, currentRound, &like.AnswerHp{
			CurrentBaseID: currentBaseID,
			Hp:            _userDefaultHP,
			StartTime:     nowTs,
		}); err != nil {
			err = ecode.ActivityQuestionNo
			return
		}
		// 获取用户信息，判断用户是否加入表
		if userInfo, err = s.userAnswer(ctx, mid); err != nil {
			log.Errorc(ctx, "Question UserInfo s.dao.CacheUserInfo mid(%d) error(%+v)", mid, err)
			return
		} else {
			if userInfo == nil {
				userInfo = &like.AnswerUserInfo{}
			}
			if userInfo.IsJoin == 0 {
				if _, err = s.dao.AddUserQuestion(ctx, mid); err != nil {
					log.Errorc(ctx, "Question UserInfo s.dao.AddUserQuestion mid(%d) error(%+v)", mid, err)
					return
				} else {
					userInfo.IsJoin = 1
				}
			}
			// 答题次数，更新用户信息缓存
			userInfo.AnswerTimes++
			if err = s.dao.AddCacheUserInfo(ctx, mid, strWeek, userInfo); err != nil {
				log.Errorc(ctx, "Question UserInfo s.dao.AddUserQuestion mid(%d) error(%+v)", mid, err)
				return
			}
		}
	} else {
		res.CurrentRound = params.CurrentRound
		currentHP, e := s.dao.CacheUserHp(ctx, mid, res.CurrentRound)
		if e != nil {
			log.Errorc(ctx, "AnswerQuestion s.dao.CacheUserHp mid(%d) CurrentRound(%d) error(%+v)", mid, res.CurrentRound, e)
			err = ecode.ActivityQuestionNo
			return
		} else {
			if currentHP == nil || currentHP.Hp < 0 {
				err = ecode.ActivityAnswerHpOver
				log.Errorc(ctx, "AnswerQuestion mid(%d) CurrentRound(%d) Hp Over", mid, res.CurrentRound)
				return
			}
			if currentHP.Hp < _userHPOver {
				currentHP.Hp = _userHPOver
			}
			res.UserHp = currentHP.Hp
		}
	}
	// 获取所有题目
	if idsSlice, err = s.dao.PoolQuestionPage(ctx, currentBaseID, res.CurrentRound); err != nil {
		err = ecode.ActivityQuestionNotStart
		log.Errorc(ctx, "Question s.dao.PoolQuestionPage baseID(%d) poolID(%d) error(%v)", currentBaseID, res.CurrentRound, err)
		return
	}
	if len(idsSlice) == 0 {
		err = ecode.ActivityQuestionNotStart
		return
	}
	count := len(idsSlice)
	if count == 0 || count < start {
		res.List = _emptyQuestionList
		return
	}
	if count > end+1 {
		resIDs = idsSlice[start : end+1]
	} else {
		resIDs = idsSlice[start:]
	}
	for index, detailID := range resIDs {
		if detail, ok := currentDetails[detailID]; ok {
			answers := append(strings.Split(detail.RightAnswer, ","), strings.Split(detail.WrongAnswer, ",")...)
			s.r.Shuffle(len(answers), func(i, j int) {
				answers[i], answers[j] = answers[j], answers[i]
			})
			questionList = append(questionList, &like.QuestionItem{
				QuestionOrder: index + start + 1,
				ID:            detail.ID,
				Attribute:     detail.Attribute,
				Question:      detail.Name,
				Answers:       answers,
				Pic:           detail.Pic,
			})
		}
	}
	if len(questionList) == 0 {
		res.List = _emptyQuestionList
		return
	}
	res.List = questionList
	return
}

func (s *Service) AnswerResult(ctx context.Context, mid int64, params *like.ParamResult) (res *like.AnswerResult, err error) {
	var check bool
	defer func() {
		s.reportAnswer(ctx, mid, params, res)
	}()
	if check, err = s.dao.RsSetNX(ctx, fmt.Sprintf("result_%d", mid), 1); err != nil || !check {
		log.Warn("AnswerResult s.dao.RsSetNX mid:%d to fast err:%v", mid, err)
		err = ecode.ActivityRapid
		return
	}
	// 检查一轮一道题只能答一次
	if check, err = s.dao.RsSetNX(ctx, fmt.Sprintf("repeat_%d_%d_%d", mid, params.QuestionID, params.CurrentRound), 3600); err != nil || !check {
		log.Warn("AnswerResult s.dao.RsSetNX repeat_%d_%d_%d to repeat err:%v", mid, params.QuestionID, params.CurrentRound, err)
		err = ecode.ActivityAnswerRepeat
		return
	}
	res = &like.AnswerResult{CurrentRound: params.CurrentRound}
	currentHP, e := s.dao.CacheUserHp(ctx, mid, params.CurrentRound)
	if e != nil {
		err = ecode.ActivityAnswerHpOver
		log.Errorc(ctx, "s.dao.CacheUserHp mid(%d) poolID(%d) error(%+v)", mid, params.CurrentRound, err)
		return
	} else {
		if currentHP == nil || currentHP.Hp < 0 {
			err = ecode.ActivityAnswerHpOver
			log.Errorc(ctx, "AnswerQuestion mid(%d) CurrentRound(%d) Hp Over", mid, params.CurrentRound)
			return
		}
		if currentHP.Hp < _userHPOver {
			currentHP.Hp = _userHPOver
		}
		res.UserHp = currentHP.Hp
	}
	currentDetails, ok := s.answerQuestionDetails[currentHP.CurrentBaseID]
	questionMax := len(currentDetails)
	if !ok || questionMax == 0 {
		err = ecode.ActivityQuestionNotStart
		return
	}
	// 答题超时计算
	currentHP.AnswerCount++
	nowTime := time.Now().Unix()
	if params.QuestionOrder > questionMax {
		currentHP.Hp--
	} else if nowTime > currentHP.StartTime+(currentHP.AnswerCount*20) {
		currentHP.Hp--
		res.TimeOut = 1
	} else if currentAnswer, ok := currentDetails[params.QuestionID]; ok && currentAnswer.RightAnswer == params.UserAnswer {
		res.IsRight = _userRight
		currentHP.NowScore++
		if currentAnswer.Pic == "" {
			params.TopicType = 1
		} else {
			params.TopicType = 2
		}
		params.Topic = currentAnswer.Name
	} else {
		currentHP.Hp--
		if currentAnswer.Pic == "" {
			params.TopicType = 1
		} else {
			params.TopicType = 2
		}
		params.Topic = currentAnswer.Name
	}
	res.UserHp = currentHP.Hp
	res.NowScore = currentHP.NowScore
	if params.QuestionOrder == questionMax {
		res.QuestionOver = 1
	}
	if res.UserHp <= 0 || res.QuestionOver == 1 {
		// 更新user info 最高与上一次得分
		userInfo, e := s.userAnswer(ctx, mid)
		if e != nil {
			log.Errorc(ctx, "AnswerResult s.dao.CacheUserInfo mid(%d) error(%+v)", mid, e)
			return
		}
		nowPercent := s.calcParent(res.NowScore)
		if userInfo != nil {
			userInfo.LastInfo.LastScore = res.NowScore
			userInfo.LastInfo.LastPercent = nowPercent
		}
		res.NowPercent = nowPercent
		if userInfo != nil && userInfo.UserScore < res.NowScore {
			res.UserScore = res.NowScore
			userInfo.UserScore = res.NowScore
			userInfo.UserPercent = nowPercent
			userInfo.FinishTime = nowTime
		} else {
			if userInfo != nil {
				res.UserScore = userInfo.UserScore
			} else {
				res.UserScore = res.NowScore
			}
		}
		if res.NowScore >= s.c.S10Answer.CanPendantCount && userInfo != nil && userInfo.CanPendant == 0 && userInfo.HavePendant == 0 {
			// 更新一下挂件
			var pendantRule *like.PendantRule
			if pendantRule, err = s.dao.CachePendantRule(ctx, mid); err != nil {
				log.Errorc(ctx, "AnswerResult s.dao.CachePendantRule mid(%d) error(%+v)", mid, err)
				return
			}
			if pendantRule != nil && pendantRule.CanPendant == 0 && pendantRule.HavePendant == 0 {
				pendantRule.CanPendant = 1
				if err = s.dao.AddCachePendantRule(ctx, mid, pendantRule); err != nil {
					log.Errorc(ctx, "AnswerResult s.dao.CachePendantRule mid(%d) error(%+v)", mid, err)
					return
				}
			}
		}
		// 答题次数 在获取题目时就算一次答题
		//if currentHP.HaveAddTime == 0 {
		//	userInfo.AnswerTimes++
		//	// 保存当前HP
		//	currentHP.HaveAddTime = 1 // 一轮题，只加一次答题次数
		//}
		params.OrderID = userInfo.AnswerTimes
		var strWeek string
		if strWeek, err = s.getUserInfoKey(); err != nil {
			log.Errorc(ctx, "AnswerResult  s.getUserInfoKey() empty")
			return nil, err
		}
		eg := errgroup.WithContext(ctx)
		eg.Go(func(ctx context.Context) (e error) {
			if e = s.dao.AddCacheUserInfo(ctx, mid, strWeek, userInfo); e != nil {
				log.Errorc(ctx, "AnswerResult s.dao.AddCacheUserInfo mid(%d) error(%+v)", mid, e)
			}
			return
		})
		eg.Go(func(ctx context.Context) (e error) {
			if _, e = s.dao.UpQuestionAllTimes(ctx, mid); e != nil {
				log.Errorc(ctx, "AnswerResult s.dao.UpQuestionAllTimes mid(%d) error(%+v)", mid, e)
			}
			return
		})
		if e := eg.Wait(); e != nil {
			log.Errorc(ctx, "AnswerResult errgroup mid(%d) error(%+v)", mid, e)
			return
		}
		log.Info("AnswerResult  user round success mid(%d) nowScore(%d)", mid, res.NowScore)
	}
	// 设置HP
	if err = s.dao.AddCacheUserHp(ctx, mid, params.CurrentRound, currentHP); err != nil {
		log.Errorc(ctx, "AnswerResult s.dao.AddCacheUserHp mid(%d) round(%d) error(%d)", mid, params.CurrentRound, err)
		return
	}
	return
}

func (s *Service) AnswerRank(ctx context.Context) (res []*like.UserRank, err error) {
	res = s.answerRank
	if len(res) == 0 {
		res = _emptyRank
	}
	return
}

func (s *Service) AnswerPendant(ctx context.Context, mid int64) (err error) {
	var check bool
	if check, err = s.dao.RsSetNX(ctx, fmt.Sprintf("pendant_%d", mid), 1); err != nil || !check {
		log.Warn("AnswerPendant s.dao.RsSetNX mid:%d to fast err:%v", mid, err)
		err = ecode.ActivityRapid
		return
	}
	var pendantRule *like.PendantRule
	if pendantRule, err = s.dao.CachePendantRule(ctx, mid); err != nil {
		log.Errorc(ctx, "AnswerPendant s.dao.CachePendantRule mid(%d) error(%+v)", mid, err)
		return
	}
	if pendantRule == nil {
		pendantRule = &like.PendantRule{}
	}
	if pendantRule.HavePendant == 1 {
		return
	}
	if pendantRule.CanPendant == 0 {
		err = ecode.ActivityNotPendant
		return
	}
	eg := errgroup.WithContext(ctx)
	eg.Go(func(ctx context.Context) (e error) {
		mids := []int64{mid}
		pendentID := s.c.S10Answer.PendantID
		_, err = client.GarbClient.GrantByBiz(ctx, &garbapi.GrantByBizReq{Mids: mids, Ids: []int64{pendentID}, AddSecond: s.c.S10Answer.PendantExpire})
		return
	})
	eg.Go(func(ctx context.Context) (e error) {
		if _, e = s.dao.UpPendant(ctx, mid); e != nil {
			log.Errorc(ctx, "AnswerPendant s.dao.UpPendant mid(%d) error(%+v)", mid, e)
		}
		return
	})
	if err = eg.Wait(); err != nil {
		log.Error("AnswerPendant errgroup mid(%d) error(%+v)", mid, err)
		err = ecode.ActivitySuitsFail
		return
	}
	pendantRule.CanPendant = 0
	pendantRule.HavePendant = 1
	if err = s.dao.AddCachePendantRule(ctx, mid, pendantRule); err != nil {
		log.Errorc(ctx, "AnswerPendant s.dao.CachePendantRule mid(%d) error(%+v)", mid, err)
	}
	return
}

func (s *Service) KnowRule(ctx context.Context, mid int64) (err error) {
	var check bool
	if check, err = s.dao.RsSetNX(ctx, fmt.Sprintf("rule_%d", mid), 1); err != nil || !check {
		log.Warn("KnowRule s.dao.RsSetNX mid:%d to fast err:%v", mid, err)
		err = ecode.ActivityRapid
		return
	}
	var pendantRule *like.PendantRule
	if pendantRule, err = s.dao.CachePendantRule(ctx, mid); err != nil {
		log.Errorc(ctx, "KnowRule s.dao.CacheUserInfo mid(%d) error(%+v)", mid, err)
		return
	}
	if pendantRule == nil {
		pendantRule = &like.PendantRule{}
	}
	if pendantRule.KnowRule == 1 {
		return
	}
	eg := errgroup.WithContext(ctx)
	eg.Go(func(ctx context.Context) (e error) {
		if _, e = s.dao.UpKnowRule(ctx, mid); e != nil {
			log.Errorc(ctx, "KnowRule s.dao.UpKnowRule mid(%d) error(%+v)", mid, e)
		}
		return
	})
	eg.Go(func(ctx context.Context) (e error) {
		pendantRule.KnowRule = 1
		if e = s.dao.AddCachePendantRule(ctx, mid, pendantRule); e != nil {
			log.Errorc(ctx, "KnowRule s.dao.AddCacheUserInfo mid(%d) error(%+v)", mid, e)
		}
		return
	})
	err = eg.Wait()
	return
}

func (s *Service) ShareAddHP(ctx context.Context, mid, currentRound int64) (err error) {
	var check bool
	if check, err = s.dao.RsSetNX(ctx, fmt.Sprintf("share_%d", mid), 1); err != nil || !check {
		log.Warn("ShareAddHP s.dao.RsSetNX mid:%d to fast err:%v", mid, err)
		err = ecode.ActivityRapid
		return
	}
	currentHP, e := s.dao.CacheUserHp(ctx, mid, currentRound)
	if e != nil {
		err = xecode.RequestErr
		log.Errorc(ctx, "ShareAddHP s.dao.CacheUserHp mid(%d) CurrentRound(%d) error(%+v)", mid, currentRound, err)
		return
	} else {
		if currentHP == nil {
			err = xecode.RequestErr
			log.Errorc(ctx, "ShareAddHP mid(%d) CurrentRound(%d) HP(%+v) nil", mid, currentRound, currentHP.Hp)
			return
		}
		if currentHP.ShareHp == 1 {
			err = ecode.ActivityAlreadyShare
			return
		}
		currentHP.ShareHp = 1
		currentHP.Hp++
	}
	if err = s.dao.AddCacheUserHp(ctx, mid, currentRound, currentHP); err != nil {
		log.Errorc(ctx, "ShareAddHP s.dao.AddCacheUserHp mid(%d) round(%d) error(%d)", mid, currentRound, err)
		return
	}
	return
}

func (s *Service) loadQuestionDetail() {
	var (
		ctx = context.Background()
	)
	// 活动结束不加载
	if time.Now().Unix() > s.c.S10Answer.EndTime {
		return
	}
	tmp := make(map[int64]map[int64]*question.Detail, len(s.questionBase))
	if len(s.questionBase) == 0 {
		log.Infoc(ctx, "loadQuestionDetail questionBase count 0")
		return
	}
	roundCount := len(s.c.S10Answer.AnswerRound)
	if roundCount == 0 {
		return
	}
	baseIDMap := make(map[int64]struct{}, roundCount)
	for _, answerRound := range s.c.S10Answer.AnswerRound {
		baseIDMap[answerRound.BaseID] = struct{}{}
	}
	for _, questionBase := range s.questionBase {
		if _, ok := baseIDMap[questionBase.ID]; ok {
			details, err := s.retryDetails(ctx, questionBase.ID)
			if err != nil {
				log.Error("loadQuestionDetail s.dao.QuestionDetails baseID(%d) error(%d)", questionBase.ID, err)
				continue
			}
			if len(details) > 0 {
				tmp[questionBase.ID] = details
			}
		}
	}
	s.answerQuestionDetails = tmp
}

func (s *Service) retryDetails(ctx context.Context, baseID int64) (list map[int64]*question.Detail, err error) {
	for i := 0; i < 3; i++ {
		if list, err = s.dao.QuestionDetails(ctx, baseID); err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	return
}

func (s *Service) getUserInfoKey() (res string, err error) {
	nowTime := time.Now().Unix()
	for _, answerRound := range s.c.S10Answer.AnswerRound {
		if nowTime > answerRound.RoundDate {
			res = time.Unix(answerRound.RoundDate, 0).Format("20060102")
			break
		}
	}
	if res == "" {
		err = ecode.ActivityQuestionNotStart
	}
	return
}

func (s *Service) WeekTop(ctx context.Context) (res []int64, err error) {
	return s.dao.CacheWeekTop(ctx)
}

func (s *Service) calcParent(rightCount int64) (res int64) {
	var lessCount int64
	if s.hourPeopleInfo != nil {
		for userScore, peopleCount := range s.hourPeopleInfo.PeopleCount {
			if userScore < rightCount {
				lessCount += peopleCount
			}
		}
		if s.hourPeopleInfo.WeekPeople > 0 {
			res = int64((float64(lessCount) / float64(s.hourPeopleInfo.WeekPeople)) * 100)
		}
	} else {
		// 无数据使用规则
		for _, percent := range s.c.S10Answer.AnswerPercent {
			if rightCount >= percent.RightBegin && rightCount <= percent.RightEnd {
				diffRight := float64(percent.RightEnd - percent.RightBegin)
				diffPercent := percent.PercentEnd - percent.PercentBegin
				res = int64((float64(rightCount)/diffRight*(diffPercent*0.01) + (percent.PercentBegin * 0.01)) * 100)
			}
		}
	}
	if res < 1 {
		res = 1
	} else if res > 99 {
		res = 99
	}
	return
}

func (s *Service) loadAnswerHourPeople() {
	if s.hourPeopleInfo == nil {
		s.hourPeopleInfo = new(like.HourPeople)
	}
	res, err := s.dao.CacheHourPeople(context.Background())
	if err != nil {
		log.Error("loadAnswerHourPeople error(%+v)", err)
		return
	}
	if res != nil {
		s.hourPeopleInfo = res
	}
	log.Info("load Answer hour people success")
}

func (s *Service) loadAnswerUserRank() {
	res, err := s.dao.CacheUserTop(context.Background())
	if err != nil {
		log.Error("loadAnswerUserRank s.dao.loadAnswerUserRank error(%+v)", err)
		return
	}
	if res != nil {
		s.answerRank = res
	}
	log.Info("load Answer user rank success")
}

func (s *Service) reportAnswer(ctx context.Context, mid int64, params *like.ParamResult, res *like.AnswerResult) {
	var isRight int64
	if res != nil {
		isRight = res.IsRight
	}
	riskData := &like.GaiaResult{
		MID:        mid,
		Buvid:      params.Buvid,
		IP:         params.IP,
		Platform:   params.Platform,
		CTime:      time.Now().Format("2006-01-02 15:04:05"),
		AccessKey:  "",
		Caller:     _riskCaller,
		API:        "/x/activity/answer/result",
		Origin:     params.Origin,
		Referer:    params.Referer,
		UserAgent:  params.UA,
		Build:      strconv.FormatInt(params.Build, 10),
		Code:       0,
		OrderID:    params.OrderID,
		Topic:      params.Topic,
		Action:     "answer",
		TopicTime:  time.Unix(params.CurrentRound, 0).Format("2006-01-02 15:04:05"),
		UserAnswer: params.UserAnswer,
		TopicType:  params.TopicType,
		Result:     isRight,
	}
	if eventCtx, err1 := json.Marshal(riskData); err1 == nil {
		s.cache.Do(ctx, func(c context.Context) {
			if err := s.dao.ReportGaia(ctx, _resultScene, string(eventCtx)); err != nil {
				log.Errorc(ctx, "reportAnswer s.dao.ReportGaia mid(%d) error(%+v)", mid, err)
			}
		})
	}
}
