package like

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	xtime "go-common/library/time"
	"go-gateway/app/web-svr/activity/ecode"
	pb "go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/interface/model/question"
	"go-main/app/account/usersuit/service/api"
)

const (
	_firstIndex = 1
)

// QuestionStart get new question zone.
func (s *Service) QuestionStart(c context.Context, sid, mid int64) (data *question.Item, err error) {
	var (
		base          *question.Base
		lastLog       *question.UserAnswerLog
		quesData      *question.DetailItem
		ids           map[int64]int64
		lastID, limit int64
	)
	if base, err = s.checkQuestionBase(sid); err != nil {
		return
	}
	now := time.Now()
	day := now.Format("2006-01-02")
	nowTs := now.Unix()
	if nowTs > base.Etime.Time().Unix() {
		err = ecode.ActivityOverEnd
		return
	}
	if limit, err = s.quesDao.QuesLimit(c, mid, base.ID, day); err != nil {
		log.Errorc(c, "s.quesDao.QuesLimit mid(%d) baseID(%d) error(%d)", mid, base.ID, err)
		err = nil
	} else if limit > s.c.Rule.QuestionLimit {
		err = ecode.ActivityQuestionLimit
		return
	}
	if lastLog, err = s.quesDao.LastQuesLog(c, mid, base.ID); err != nil {
		log.Errorc(c, "QuestionStart s.quesDao.LastQuesLog mid(%d) baseID(%d) error(%v)", mid, base.ID, err)
		return
	}
	if lastLog != nil && nowTs-lastLog.Ctime.Time().Unix() < s.c.Rule.QuestionCD {
		err = ecode.ActivityQuestionCD
		return
	}
	// pre 10 second.
	var poolID int64
	if ids, poolID, err = s.quesDao.PoolQuestionIDsWithDefault(c, base.ID, nowTs, int(base.Count)); err != nil {
		log.Errorc(c, "QuestionStart s.quesDao.CacheLastQuesLog mid(%d) baseID(%d) poolID(%d) error(%v)", mid, base.ID, poolID, err)
		return
	}
	if len(ids) < int(base.Count) {
		log.Errorc(c, "QuestionStart s.quesDao.CacheLastQuesLog mid(%d) baseID(%d) len(%v)", mid, base.ID, ids)
		err = ecode.ActivityQuestionNo
		return
	}
	quesID, ok := ids[_firstIndex]
	if !ok {
		err = ecode.ActivityQuestionNo
		return
	}
	if quesData, err = s.questionDetail(c, quesID, base); err != nil {
		return
	}
	if lastID, err = s.quesDao.AddUserLog(c, mid, base.ID, quesID, poolID, _firstIndex, now); err != nil {
		return
	}
	// 写用户维度答题明细
	s.quesDao.AddUserRecords(c, mid, base.ID, poolID, base.Count, 0, 0, 0)
	data = &question.Item{
		ID:      quesID,
		PoolID:  poolID,
		Name:    quesData.Name,
		Answers: quesData.Answers,
		Pic:     quesData.Pic,
		Index:   _firstIndex,
	}
	s.cache.Do(c, func(ctx context.Context) {
		log := &question.UserAnswerLog{
			ID:           lastID,
			Mid:          mid,
			BaseID:       base.ID,
			DetailID:     quesID,
			PoolID:       poolID,
			Index:        _firstIndex,
			QuestionTime: xtime.Time(nowTs),
			Ctime:        xtime.Time(nowTs),
		}
		if e := s.quesDao.AddCacheLastQuesLog(ctx, mid, log, base.ID); e != nil {
			return
		}
		// add limit time
		if e := s.quesDao.IncrQuesLimit(ctx, mid, base.ID, day); e != nil {
			return
		}
	})
	return
}

// Next get next question.
func (s *Service) Next(c context.Context, sid, poolID, mid int64) (data *question.Item, err error) {
	var (
		base               *question.Base
		lastLog            *question.UserAnswerLog
		nextQuesID, lastID int64
		nextQues           *question.DetailItem
	)
	now := time.Now()
	if base, err = s.checkQuestionBase(sid); err != nil {
		return
	}
	if now.Unix() > base.Etime.Time().Unix() {
		err = ecode.ActivityOverEnd
		return
	}
	if lastLog, err = s.quesDao.LastQuesLog(c, mid, base.ID); err != nil {
		log.Errorc(c, "Next s.quesDao.LastQuesLog mid(%d) baseID(%d) error(%v)", mid, base.ID, err)
		return
	}
	if lastLog == nil {
		err = ecode.ActivityQuestionNotStart
		return
	}
	if lastLog.Index >= base.Count {
		err = ecode.ActivityQuestionFinish
		return
	}
	if lastLog.PoolID != poolID {
		log.Errorc(c, "Next s.quesDao.LastQuesLog mid(%d) baseID(%d) lastlog(%v)", mid, base.ID, lastLog)
		err = xecode.RequestErr
		return
	}
	nextIndex := lastLog.Index + 1
	if nextQuesID, err = s.quesDao.PoolIndexQuestionID(c, base.ID, lastLog.PoolID, nextIndex); err != nil {
		return
	}
	if nextQues, err = s.questionDetail(c, nextQuesID, base); err != nil {
		return
	}
	if lastID, err = s.quesDao.AddUserLog(c, mid, base.ID, nextQues.ID, lastLog.PoolID, nextIndex, now); err != nil {
		return
	}
	log := &question.UserAnswerLog{
		ID:           lastID,
		Mid:          mid,
		BaseID:       base.ID,
		DetailID:     nextQues.ID,
		PoolID:       lastLog.PoolID,
		Index:        nextIndex,
		QuestionTime: xtime.Time(now.Unix()),
		Ctime:        xtime.Time(now.Unix()),
	}
	if err = s.quesDao.AddCacheLastQuesLog(c, mid, log, base.ID); err != nil {
		return
	}
	data = &question.Item{
		ID:      nextQues.ID,
		PoolID:  lastLog.PoolID,
		Name:    nextQues.Name,
		Pic:     nextQues.Pic,
		Answers: nextQues.Answers,
		Index:   nextIndex,
	}
	return
}

// Answer answer a question and give next question or finish score.
func (s *Service) Answer(c context.Context, sid, poolID, mid, quesID, index int64, answer string) (data *question.Answer, err error) {
	var (
		base              *question.Base
		lastLog           *question.UserAnswerLog
		quesData          *question.Detail
		userLogs          []*question.UserAnswerLog
		isRight, isFinish int
	)
	if base, err = s.checkQuestionBase(sid); err != nil {
		return
	}
	if lastLog, err = s.quesDao.LastQuesLog(c, mid, base.ID); err != nil {
		log.Errorc(c, "Answer s.quesDao.LastQuesLog mid(%d) baseID(%d) error(%v)", mid, base.ID, err)
		return
	}
	if lastLog == nil {
		err = ecode.ActivityQuestionNotStart
		return
	}
	if lastLog.Index != index || lastLog.DetailID != quesID || lastLog.PoolID != poolID {
		log.Errorc(c, "Answer s.quesDao.LastQuesLog mid(%d) quesID(%d) index(%d) baseID(%d) lastlog(%v)", mid, quesID, index, base.ID, lastLog)
		err = xecode.RequestErr
		return
	}
	if quesData, err = s.quesDao.Detail(c, quesID); err != nil {
		return
	}
	if lastLog.AnswerTime > 0 {
		isRight = int(lastLog.IsRight)
	} else {
		// TODO not only one right answer
		if quesData.RightAnswer == answer {
			isRight = 1
		}
		// up log
		if err = s.quesDao.UpUserLog(c, isRight, time.Now(), lastLog.ID, lastLog.BaseID); err != nil {
			return
		}
		if err = s.quesDao.AddCacheLastQuesLog(c, mid, &question.UserAnswerLog{
			ID:           lastLog.ID,
			Mid:          lastLog.Mid,
			BaseID:       lastLog.BaseID,
			DetailID:     lastLog.DetailID,
			PoolID:       lastLog.PoolID,
			Index:        lastLog.Index,
			QuestionTime: lastLog.QuestionTime,
			Ctime:        lastLog.Ctime,
			Mtime:        xtime.Time(time.Now().Unix()),
			IsRight:      int64(isRight),
			AnswerTime:   xtime.Time(time.Now().Unix()),
		}, base.ID); err != nil {
			return
		}
	}
	if index == base.Count {
		isFinish = 1
	}
	data = &question.Answer{
		IsRight:    isRight,
		Finish:     isFinish,
		Answer:     quesData.RightAnswer,
		AnswerTime: lastLog.AnswerTime.Time().Unix(),
	}
	if isFinish == 1 {
		// right count
		var rightCount int
		if userLogs, err = s.quesDao.RawUserLogs(c, mid, base.ID, lastLog.PoolID); err != nil {
			log.Errorc(c, "Answer s.quesDao.RawUserLogs(mid:%d,baseID:%d,poolID:%d) error(%v)", mid, base.ID, lastLog.PoolID, err)
		}
		for _, v := range userLogs {
			if v.IsRight == 1 {
				rightCount++
			}
		}
		data.RightCount = rightCount
		// 更新用户维度答题明细
		s.quesDao.UpUserRecords(c, mid, base.ID, lastLog.PoolID, int64(len(userLogs)), int64(rightCount), 1)
		// send user suit
		if sid == s.c.Rule.S9QuesSid && rightCount >= s.c.Rule.S9Right {
			s.cache.Do(c, func(c context.Context) {
				limitKey := fmt.Sprintf("s9_suit_%d_%d", mid, sid)
				if hasCheck, e := s.dao.RsSetNX(c, limitKey, s.c.Rule.S9CacheExpire); e != nil || !hasCheck {
					log.Warn("Answer s.dao.RsSetNX(%s) hasCheck(%v) error(%v)", limitKey, hasCheck, e)
				} else {
					arg := &api.GrantByMidsReq{Mids: []int64{mid}, Pid: s.c.Rule.S9SuitID, Expire: s.c.Rule.S9SuitExpire}
					if _, suitErr := s.suitClient.GrantByMids(c, arg); suitErr != nil {
						log.Warn("Answer s.suitClient.GrantByMids(%v) error(%v)", arg, e)
						// del cache
						if e := s.dao.RsDelNX(c, limitKey); e != nil {
							log.Errorc(c, "Answer s.dao.RsDelNX(%s) error(%v)", limitKey, e)
						}
					}
				}
			})
		}
	}
	return
}

func (s *Service) loadQuestionBaseData() {
	nowTime := xtime.Time(time.Now().Unix())
	ctx := context.Background()
	if base, err := s.quesDao.RawBases(ctx, nowTime); err != nil {
		log.Errorc(ctx, "questionBaseproc s.quesDao.RawBases(%d,%v,%d) error(%v)", question.BusinessTypeAct, s.c.Rule.QuestionSid, nowTime, err)
		return
	} else {
		tmp := make(map[string]*question.Base, len(base))
		for _, v := range base {
			tmp[baseKey(v.BusinessID, v.ForeignID)] = v
		}
		s.questionBase = tmp
	}
	log.Infoc(ctx, "loadQuestionBaseData() success")
}

func (s *Service) questionDetail(c context.Context, id int64, base *question.Base) (data *question.DetailItem, err error) {
	var detail *question.Detail
	if detail, err = s.quesDao.Detail(c, id); err != nil {
		return
	}
	answers := append(strings.Split(detail.RightAnswer, base.Separator), strings.Split(detail.WrongAnswer, base.Separator)...)
	s.r.Shuffle(len(answers), func(i, j int) {
		answers[i], answers[j] = answers[j], answers[i]
	})
	data = &question.DetailItem{
		ID:        detail.ID,
		BaseID:    detail.BaseID,
		Name:      detail.Name,
		Pic:       detail.Pic,
		Answers:   answers,
		Attribute: detail.Attribute,
	}
	return
}

func (s *Service) checkQuestionBase(sid int64) (data *question.Base, err error) {
	base, ok := s.questionBase[baseKey(question.BusinessTypeAct, sid)]
	if !ok || base == nil {
		err = xecode.NothingFound
		return
	}
	data = base
	return
}

func baseKey(businessID, foreignID int64) string {
	return fmt.Sprintf("%d_%d", businessID, foreignID)
}

func (s *Service) QuestionAnswerDetail(c context.Context, sid int64, extraCode int64) (detail map[int64]*question.Detail, poolID int64, err error) {
	var (
		base   *question.Base
		idsMap map[int64]int64
	)
	if base, err = s.checkQuestionBase(sid); err != nil {
		return
	}
	if _, ok := s.internalQuestionSids[sid]; !ok && base.DistributeType <= 1 {
		err = xecode.AccessDenied
		log.Warnc(c, "QuestionAnswer forbid internal sid:%d", sid)
		return
	}
	now := time.Now()
	nowTs := now.Unix()
	if nowTs > base.Etime.Time().Unix() {
		err = ecode.ActivityOverEnd
		return
	}
	// pre 10 second.
	var redisKey = base.ID
	if extraCode > 0 {
		redisKey = redisKey + extraCode
	}
	if idsMap, poolID, err = s.quesDao.PoolQuestionIDsWithDefault(c, redisKey, nowTs, int(base.Count)); err != nil {
		log.Errorc(c, "QuestionAnswer s.quesDao.PoolQuestionIDs baseID(%d) poolID(%d) error(%v)", base.ID, poolID, err)
		return
	}
	if len(idsMap) < int(base.Count) {
		log.Errorc(c, "QuestionAnswer s.quesDao.PoolQuestionIDs baseID(%d) len(%v)", base.ID, idsMap)
		err = ecode.ActivityQuestionNo
		return
	}
	ids := make([]int64, 0)
	for _, v := range idsMap {
		ids = append(ids, v)
	}
	detail, err = s.quesDao.Details(c, ids)
	if err != nil {
		log.Errorc(c, "QuestionAnswer s.quesDao.Details baseID(%d) len(%v)", base.ID, idsMap)
	}
	return
}

// QuestionAnswer question and answer
func (s *Service) QuestionAnswer(c context.Context, sid int64) (data *question.QAReply, err error) {
	var base *question.Base
	if base, err = s.checkQuestionBase(sid); err != nil {
		return
	}
	var detail map[int64]*question.Detail
	var poolID int64
	if detail, poolID, err = s.QuestionAnswerDetail(c, sid, 0); err != nil {
		return
	}
	return s.questionAnswerDetail(c, poolID, detail, base)
}

func (s *Service) QuestionAnswerAll(c context.Context, sid, poolID, mid int64, answer map[int64]string) (ret *pb.QuestionAnswerAllReply, err error) {
	var base *question.Base
	if base, err = s.checkQuestionBase(sid); err != nil {
		return
	}
	now := time.Now()
	if now.After(base.Etime.Time()) {
		err = ecode.ActivityOverEnd
		return
	}
	if now.Before(base.Stime.Time()) {
		err = ecode.ActivityNotStart
		return
	}
	if len(answer) != int(base.Count) {
		log.Errorc(c, "QuestionAnswerAll len(answer) != int(base.Count) mid(%d) baseID(%d)", mid, base.ID)
		err = xecode.RequestErr
		return
	}
	var record *question.UserAnswerRecord
	if record, err = s.quesDao.RawUserRecord(c, mid, base.ID, poolID); err != nil {
		log.Errorc(c, "QuestionAnswerAll s.quesDao.RawUserRecord baseID(%d) poolID(%d) mid(%d) error(%v)", base.ID, poolID, mid, err)
		return
	}

	var idsMap map[int64]int64
	if idsMap, err = s.quesDao.PoolQuestionIDs(c, base.ID, poolID, int(base.Count)); err != nil {
		log.Errorc(c, "QuestionAnswerAll s.quesDao.PoolQuestionIDs baseID(%d) poolID(%d) error(%v)", base.ID, poolID, err)
		return
	}
	ids := make([]int64, 0)
	for _, v := range idsMap {
		ids = append(ids, v)
	}
	var detail map[int64]*question.Detail
	detail, err = s.quesDao.Details(c, ids)
	if err != nil {
		log.Errorc(c, "QuestionAnswerAll s.quesDao.Details baseID(%d) len(%v)", base.ID, idsMap)
	}

	ret = new(pb.QuestionAnswerAllReply)
	ret.AnswerCount = int64(len(answer))
	ret.Answer = make(map[int64]string)
	for qid, ans := range answer {
		q := detail[qid]
		if q == nil {
			log.Errorc(c, "QuestionAnswerAll question err qid[%d]", qid)
			err = xecode.RequestErr
		}
		ret.Answer[qid] = q.RightAnswer
		if q.RightAnswer == ans {
			ret.RightCount++
		}
	}
	if record != nil && record.State == 1 {
		return ret, nil
	}
	if record != nil && record.State == 0 {
		err = s.quesDao.UpUserRecords(c, mid, base.ID, poolID, ret.AnswerCount, ret.RightCount, 1)
	} else {
		_, err = s.quesDao.AddUserRecords(c, mid, base.ID, poolID, base.Count, ret.AnswerCount, ret.RightCount, 1)
	}
	return
}

func (s *Service) questionAnswerDetail(c context.Context, poolID int64, detail map[int64]*question.Detail, base *question.Base) (*question.QAReply, error) {
	list := &question.QAReply{
		List:   []*question.QAItem{},
		PoolID: poolID,
	}
	for _, v := range detail {
		answers := append(strings.Split(v.RightAnswer, base.Separator), strings.Split(v.WrongAnswer, base.Separator)...)
		s.r.Shuffle(len(answers), func(i, j int) {
			answers[i], answers[j] = answers[j], answers[i]
		})
		data := &question.QAItem{
			ID:          v.ID,
			Question:    v.Name,
			RightAnswer: strings.Split(v.RightAnswer, base.Separator),
			AllAnswer:   answers,
		}
		list.List = append(list.List, data)
	}
	return list, nil
}

func (s *Service) gaokaoQuestionDetail(ctx context.Context, detail map[int64]*question.Detail, splitTag string) (*question.GKQAReply, error) {
	list := &question.GKQAReply{
		List: []*question.GKQAItem{},
	}
	for k, v := range detail {
		answers := append(strings.Split(v.RightAnswer, splitTag), strings.Split(v.WrongAnswer, splitTag)...)
		s.r.Shuffle(len(answers), func(i, j int) {
			answers[i], answers[j] = answers[j], answers[i]
		})

		data := &question.GKQAItem{
			Qid:         k,
			Qtype:       v.Attribute,
			Img:         v.Pic,
			Question:    v.Name,
			RightAnswer: strings.Split(v.RightAnswer, ","),
			AllAnswer:   answers,
		}
		list.List = append(list.List, data)
	}

	if len(list.List) > 0 {
		//排序，实现比较方法即可
		sort.Slice(list.List, func(i, j int) bool {
			return list.List[i].Qtype < list.List[j].Qtype
		})
		switchMap := make(map[int64]bool)
		tmpList := []*question.GKQAItem{}
		var tmpItem *question.GKQAItem
		for _, v := range list.List {
			if _, ok := switchMap[v.Qtype]; ok {
				tmpItem = v
				continue
			}
			switchMap[v.Qtype] = true
			tmpList = append(tmpList, v)
		}
		list.List = append(tmpList, tmpItem)
	}

	return list, nil
}

func (s *Service) GKQuestion(ctx context.Context, req *question.GKQuestReq) (data *question.GKQAReply, err error) {

	if req.Qid != "" {
		var qids []int64
		qidArr := strings.Split(req.Qid, ",")
		for _, idStr := range qidArr {
			var id int64
			if id, err = strconv.ParseInt(idStr, 10, 64); err != nil {
				return
			}
			qids = append(qids, id)
		}
		if len(qids) > 0 {
			if len(qids) > 100 {
				// 超过限制
				return nil, ecode.ActivityOverMissionLimit
			}
			var detail map[int64]*question.Detail
			if detail, err = s.quesDao.Details(ctx, qids); err != nil {
				log.Errorc(ctx, "QuestionAnswer s.quesDao.Details qids(%v)", qids)
				return
			}
			return s.gaokaoQuestionDetail(ctx, detail, s.c.GaoKaoActConf.SpitTag)
		}
	}

	if req.Year <= 0 || req.Province == "" || req.Qtype == "" {
		return nil, ecode.SystemActivityParamsErr
	}
	// 随机选题
	var (
		qCode int64
		ok    bool
	)
	if qCode, ok = s.c.GaoKaoActConf.QtypeMap[req.Qtype]; !ok {
		log.Warnc(ctx, "GaoKaoActConf QuestionList can not  find:%+v", *req)
		return nil, ecode.SystemActivityParamsErr
	}
	var detail map[int64]*question.Detail
	if detail, _, err = s.QuestionAnswerDetail(ctx, int64(req.Year), qCode); err != nil {
		return
	}
	return s.gaokaoQuestionDetail(ctx, detail, s.c.GaoKaoActConf.SpitTag)
}

func (s *Service) GKRank(ctx context.Context, req *question.GKRankReq) (reply *question.GKRankReply, err error) {
	var (
		rankScore   int64
		rank, total int64
	)
	if rankScore, err = getRankScore(req, s.c.GaoKaoActConf.ActEndTime, s.c.GaoKaoActConf.MaxExamTime); err != nil {
		return
	}
	eg := errgroup.WithContext(ctx)

	eg.Go(func(ctx context.Context) (err1 error) {
		if s.c.GaoKaoActConf.EnableCache {
			rank, err1 = s.quesDao.GetRankByScore(ctx, req.Score, req.Province, req.Course)
			log.Infoc(ctx, "EnableCache GaoKaoActConf  GetRankByScore :%v , err:%+v", rank, err1)
		}

		if rank <= 0 || err1 != nil {
			rank, err1 = s.quesDao.SelectUserRank(ctx, req.Province, req.Course, rankScore)
			if rank > 0 && err1 == nil {
				if err3 := s.quesDao.CacheRankByScore(ctx, req.Score, req.Province, req.Course, rank); err3 != nil {
					log.Warnc(ctx, "CacheRankByScore s.dao.ReportGaia error(%+v)", err3)
				}
			}
			log.Infoc(ctx, "GKRank rankScore:%v ,rank:%v ", rankScore, rank)
		}
		if rank <= 0 {
			rank = 1
		}
		return
	})

	eg.Go(func(ctx context.Context) (err2 error) {
		if s.c.GaoKaoActConf.EnableCache {
			total, err2 = s.quesDao.GetTotalGaokaoCount(ctx)
			log.Infoc(ctx, "EnableCache GaoKaoActConf  GetTotalGaokaoCount :%v , err:%+v", total, err2)
		}

		if total <= 0 || err2 != nil {
			total, err2 = s.quesDao.SelectTotalCount(ctx)
			if total > 0 && err2 == nil {
				if err3 := s.quesDao.CacheTotalGaokaoCount(ctx, total); err3 != nil {
					log.Warnc(ctx, "CacheTotalGaokaoCount s.dao.ReportGaia error(%+v)", err3)
				}
			}
			log.Infoc(ctx, "GKRank rankScore total:%v ", total)
		}
		return
	})

	if err = eg.Wait(); err != nil {
		return
	}
	if rank > total {
		rank = total
	}
	reply = &question.GKRankReply{
		ReportTime: req.ReportTime,
		Rank:       rank,
		Total:      total,
	}
	return
}

func (s *Service) GKReportScore(ctx context.Context, mid int64, req *question.GKRankReq) (interface{}, error) {

	var (
		rankScore int64
		err       error
	)
	if req.Score > s.c.GaoKaoActConf.MaxScore {
		req.Score = s.c.GaoKaoActConf.MaxScore
	}
	if req.UsedTime > s.c.GaoKaoActConf.MaxExamTime {
		req.UsedTime = s.c.GaoKaoActConf.MaxExamTime
	}
	if rankScore, err = getRankScore(req, s.c.GaoKaoActConf.ActEndTime, s.c.GaoKaoActConf.MaxExamTime); err != nil {
		return nil, err
	}

	var isFilter bool
	isFilter, err = s.quesDao.FilterRepeatedReport(ctx, mid, req.Year, req.UsedTime, req.Score, req.Province, req.Course)
	if isFilter == false || err != nil {
		log.Infoc(ctx, "FilterRepeatedReport isFilter:%v , err:%+v", isFilter, err)
		return struct {
			ReportTime int64 `json:"report_time"`
			LastID     int64 `json:"last_id"`
		}{
			ReportTime: req.ReportTime,
		}, nil
	}
	var lastID int64
	lastID, err = s.quesDao.AddUserScore(ctx, mid, req.Year, req.Province, req.Course, req.Score, req.UsedTime, rankScore)
	return struct {
		ReportTime int64 `json:"report_time"`
		LastID     int64 `json:"last_id"`
	}{
		ReportTime: req.ReportTime,
		LastID:     lastID,
	}, err
}

func getRankScore(req *question.GKRankReq, endTime int64, maxExamTime int) (int64, error) {
	if req.ReportTime == 0 {
		req.ReportTime = time.Now().Unix()
	}
	if endTime < req.ReportTime {
		return 0, ecode.ActivityOverEnd
	}
	var rankScore int64
	rankScore = int64(req.Score)*10000 + int64(maxExamTime-req.UsedTime)
	rankScore = rankScore*10000000000 + (endTime - req.ReportTime)
	return rankScore, nil
}

func (s *Service) MyQuestionRecords(ctx context.Context, mid int64, sids, state []int64, pn, ps int64) (interface{}, error) {
	baseIDs := make([]int64, 0, len(sids))
	info := make(map[int64]*question.Base)
	for _, sid := range sids {
		if base, err := s.checkQuestionBase(sid); err != nil {
			return nil, err
		} else {
			baseIDs = append(baseIDs, base.ID)
			info[base.ID] = base
		}
	}
	records, err := s.quesDao.RawUserRecords(ctx, mid, baseIDs, state, (pn-1)*ps, ps+1)
	if err != nil {
		return nil, err
	}
	hasNext := 1
	if len(records) <= int(ps) {
		hasNext = 0
	} else {
		records = records[0:ps]
	}
	rets := make([]map[string]interface{}, 0, len(records))
	for _, record := range records {
		rets = append(rets, map[string]interface{}{
			"base": map[string]interface{}{
				"id":    info[record.BaseID].ForeignID,
				"name":  info[record.BaseID].Name,
				"count": info[record.BaseID].Count,
			},
			"record": record,
		})
	}
	return map[string]interface{}{
		"list":     rets,
		"has_next": hasNext,
		"next": map[string]interface{}{
			"pn": pn + 1,
			"ps": ps,
		},
	}, nil
}
