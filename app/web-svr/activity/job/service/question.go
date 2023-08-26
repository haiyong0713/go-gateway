package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"image"
	_ "image/png"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"
)

import (
	accountapi "git.bilibili.co/bapis/bapis-go/account/service"
	gaialibapi "git.bilibili.co/bapis/bapis-go/silverbullet/gaia/service/lib"
	"go-common/library/log"
	xtime "go-common/library/time"
	"go-common/library/xstr"
	"go-gateway/app/web-svr/activity/interface/model/question"
	"go-gateway/app/web-svr/activity/job/component/boss"
	quesmdl "go-gateway/app/web-svr/activity/job/model/question"
	"go-gateway/app/web-svr/activity/job/tool"
)

const (
	_prevStime  = 3600
	_moreExpire = 3600
	_maxBigPn   = 15000
)

func (s *Service) poolCreateproc() {
	if len(s.questionBase) == 0 {
		return
	}
	ts := time.Now().Unix()
	var (
		answerBaseIDs map[int64]struct{}
	)
	s10BaseCount := len(s.c.S10Answer.BaseIDs)
	answerBaseIDs = make(map[int64]struct{}, s10BaseCount)
	if s10BaseCount > 0 {
		for _, baseID := range s.c.S10Answer.BaseIDs {
			answerBaseIDs[baseID] = struct{}{}
		}
	}
	for _, v := range s.questionBase {
		var poolIDs []int64
		if v == nil {
			continue
		}
		expire := int32((v.OneTs * (v.Count + 1)) + _moreExpire)
		// s10 答题活动.
		if _, ok := answerBaseIDs[v.ID]; ok {
			poolDetails := s.questionRandPool(v.Details, s.c.S10Answer.MaxQuestion)
			for _, detail := range poolDetails {
				poolIDs = append(poolIDs, detail.ID)
			}
			strPoolIDs := xstr.JoinInts(poolIDs)
			if err := s.question.SetAllDetails(context.Background(), v.ID, ts, strPoolIDs, expire); err != nil {
				log.Error("poolCreateproc SetAllDetails s.question.SetPoolIDsCache baseID(%d) poolID(%d) ids(%v) expire(%d) error(%v)", v.ID, ts, poolIDs, expire, err)
				continue
			}
		} else {
			if s.c.GaoKaoAnswer != nil {
				var filter []*question.Detail
				for _, fk := range s.c.GaoKaoAnswer.ForeignId {
					if fk == v.ForeignID {
						for _, randItem := range s.c.GaoKaoAnswer.RandMethod {
							filter = s.fiterPoolIDsByAttr(randItem.AttrNum, v.Details, int(v.Count))
							var redisId = randItem.Code + v.ID
							poolIDs = []int64{}
							for _, detail := range filter {
								poolIDs = append(poolIDs, detail.ID)
							}
							poolIDStr, _ := json.Marshal(poolIDs)
							log.Info("poolCreateproc flush gaokao question , redis_key:%v , base_ids:%s ", redisId, poolIDStr)
							if len(poolIDs) > 0 {
								if err := s.question.SetPoolIDsCache(context.Background(), redisId, ts, poolIDs, expire); err != nil {
									log.Error("poolCreateproc s.question.SetPoolIDsCache gaokao baseID(%d) poolID(%d) ids(%v) expire(%d) error(%v)", redisId, ts, poolIDs, expire, err)
								}
							}
						}
						break
					}
				}
				//continue
			}
			log.Info("poolCreateproc common base_id:%v", v.ID)
			poolDetails := s.questionRandPool(v.Details, int(v.Count))
			for _, detail := range poolDetails {
				poolIDs = append(poolIDs, detail.ID)
			}
			if err := s.question.SetPoolIDsCache(context.Background(), v.ID, ts, poolIDs, expire); err != nil {
				log.Error("poolCreateproc s.question.SetPoolIDsCache baseID(%d) poolID(%d) ids(%v) expire(%d) error(%v)", v.ID, ts, poolIDs, expire, err)
				continue
			}
		}
	}
	log.Info("poolCreateproc success()")
}

func (s *Service) fiterPoolIDsByAttr(AttrNum map[string]int, details []*question.Detail, count int) (filterDetail []*question.Detail) {

	AttrMap := make(map[int64][]*question.Detail)
	for _, v := range details {
		var (
			list []*question.Detail
			ok   bool
		)

		if list, ok = AttrMap[v.Attribute]; !ok {
			list = []*question.Detail{}
		}
		AttrMap[v.Attribute] = append(list, v)
	}

	for attribute, idlist := range AttrMap {
		if num, ok := AttrNum[strconv.FormatInt(attribute, 10)]; ok && num > 0 {
			filterDetail = append(filterDetail, s.questionRandPool(idlist, num)...)
		}
	}

	leftRandNum := count - len(filterDetail)
	if leftRandNum > 0 {
		AttrCheck := make(map[int64]bool)
		for _, v := range filterDetail {
			AttrCheck[v.ID] = true
		}

		var left []*question.Detail
		for _, v := range details {
			if _, ok := AttrCheck[v.ID]; ok {
				continue
			}
			if _, ok := AttrNum[strconv.FormatInt(v.Attribute, 10)]; ok {
				left = append(left, v)
			}
		}
		filterDetail = append(filterDetail, s.questionRandPool(left, leftRandNum)...)
	}
	return
}

func (s *Service) questionBaseproc() {
	nowTs := time.Now().Unix()
	startTime := xtime.Time(nowTs - _prevStime)
	endTime := xtime.Time(nowTs)
	if bases, err := s.question.RawBases(context.Background(), startTime, endTime); err != nil {
		log.Error("s.question.RawBases() error(%v)", err)
	} else {
		tmp := make(map[string]*quesmdl.NewBaseItem, len(bases))
		for _, v := range bases {
			if v == nil || v.ID <= 0 {
				continue
			}
			if items, err := s.question.RawDetailList(context.Background(), v.ID, quesmdl.StateOnline); err != nil {
				log.Error("questionDetailproc s.question.RawDetailList(%d) error(%v)", v.ID, err)
				continue
			} else {
				tmp[fmt.Sprintf("%d_%d", v.BusinessID, v.ForeignID)] = &quesmdl.NewBaseItem{Base: v, Details: items}
			}
		}
		s.questionBase = tmp
	}
	log.Info("questionBaseproc success()")
}

func (s *Service) questionRandPool(details []*question.Detail, count int) (res []*question.Detail) {
	for i, v := range s.questionRand.Perm(len(details)) {
		if i == count {
			break
		}
		res = append(res, details[v])
	}
	return
}

func (s *Service) AnswerHour() {
	var (
		ctx            = context.Background()
		userSliceRank  []*quesmdl.UserRank
		infoReply      *accountapi.InfoReply
		thisWeekPeople int64
		gaiaReply      *gaialibapi.ExistKeyInListReply
	)
	strWeek := s.getUserInfoKey()
	if strWeek == "" {
		log.Errorc(ctx, "AnswerHour s.getUserInfoKey() empty")
		return
	}
	t := time.Now()
	zeroTime := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	nowDate := t.Format("20060102")
	userSlice, err := s.allUsers(ctx, strWeek)
	if err != nil {
		log.Errorc(ctx, "AnswerCalc s.allUsers error(%+v)", err)
		return
	}
	if len(userSlice) == 0 {
		log.Errorc(ctx, "AnswerCalc s.allUsers empty", err)
		return
	}
	// 排序
	sort.Slice(userSlice, func(i, j int) bool {
		if userSlice[i].UserScore == userSlice[j].UserScore {
			if userSlice[i].AnswerTimes == userSlice[j].AnswerTimes {
				if userSlice[i].FinishTime == userSlice[j].FinishTime {
					return userSlice[i].ID < userSlice[j].ID
				}
				return userSlice[i].FinishTime < userSlice[j].FinishTime
			}
			return userSlice[i].AnswerTimes < userSlice[j].AnswerTimes
		}
		return userSlice[i].UserScore > userSlice[j].UserScore
	})
	peopleCountMap := make(map[int64]int64, len(userSlice))
	var topCount int
	for _, userInfo := range userSlice {
		// 计算答对题数信息
		thisWeekPeople++
		peopleCountMap[userInfo.UserScore]++
		// 风控用户不进排行榜，并且分数清0
		if gaiaReply, err = s.checkRisk(ctx, strWeek, userInfo); err != nil {
			log.Errorc(ctx, "AnswerHour s.checkRisk mid(%d) error(%+v)", userInfo.Mid, err)
		} else if gaiaReply != nil && gaiaReply.Exist {
			log.Errorc(ctx, "AnswerHour  ExistKeyInList  Exist is true mid(%d)", userInfo.Mid)
			continue
		}
		// 判断用户答题次数不进榜
		if userInfo.AnswerTimes <= s.c.S10Answer.TimesNoRank && userInfo.UserScore >= s.c.S10Answer.ScoreMaxNoRank {
			continue
		}
		// 前100 score 大于用户
		if topCount < 100 && userInfo.UserScore > 0 {
			topCount++
			if infoReply, err = s.accClient.Info3(ctx, &accountapi.MidReq{Mid: userInfo.Mid}); err != nil || infoReply == nil {
				log.Error("AnswerHour s.accClient.Info3: error(%v)", err)
				continue
			}
			userSliceRank = append(userSliceRank, &quesmdl.UserRank{
				OrderNumber: topCount,
				UserScore:   userInfo.UserScore,
				AnswerTimes: userInfo.AnswerTimes,
				Account: &quesmdl.AccountInfo{
					Mid:  userInfo.Mid,
					Name: infoReply.Info.Name,
					Face: infoReply.Info.Face,
				},
			})
		}
	}
	hourPeople := &quesmdl.HourPeople{
		WeekPeople:  thisWeekPeople,
		PeopleCount: peopleCountMap,
	}
	// 每小时 用户信息用户答题数
	for i := 0; i < 3; i++ {
		if err = s.question.AddCacheHourPeople(ctx, hourPeople); err == nil {
			log.Info("AnswerCalc hourpeople  s.question.AddCacheHourPeople ok")
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if err != nil { // 忽略错误
		log.Errorc(ctx, "AnswerCalc s.question.AddCacheHourPeople error(%+v)", err)
	}
	// 每小时top 100 用户信息，排行榜直接使用
	for i := 0; i < 3; i++ {
		if err = s.question.AddCacheUserTop(ctx, userSliceRank); err == nil {
			log.Info("AnswerCalc hourtop s.question.AddCacheUserTop ok")
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if err != nil {
		log.Errorc(ctx, "AnswerCalc s.question.AddCacheUserTop error(%+v)", err)
		return
	}
	nowTime := time.Now().Unix()
	if nowTime >= zeroTime.Unix() && nowTime < zeroTime.Unix()+3000 { //0点到1点前时间段
		// 每周三0点不需要更新数据库，AnswerWeek会更新数据库.
		for _, finishWeek := range s.c.S10Answer.FinishTime {
			if finishWeek == nowDate {
				return
			}
		}
		// 更新数据库 每天0点更新数据库
		s.updateUserDB(ctx, userSlice)
	}
}

func (s *Service) AnswerHttpWeek(strWeek string) {
	s.answerWeekCalc(strWeek, false)
}

func (s *Service) AnswerWeek() {
	var (
		ctx                 = context.Background()
		finishWeek, strWeek string
	)
	t := time.Now()
	nowDate := t.Format("20060102")
	for finishWeek, strWeek = range s.c.S10Answer.FinishTime {
		if finishWeek == nowDate {
			break
		}
	}
	if strWeek == "" {
		log.Errorc(ctx, "AnswerWeek  strWeek empty")
		return
	}
	s.answerWeekCalc(strWeek, true)
}

func (s *Service) answerWeekCalc(strWeek string, isClear bool) {
	var (
		ctx       = context.Background()
		err       error
		userSlice []*quesmdl.AnswerUserInfo
		roundTop  []int64
		gaiaReply *gaialibapi.ExistKeyInListReply
	)
	userSlice, err = s.allUsers(ctx, strWeek)
	if err != nil {
		log.Errorc(ctx, "answerWeekCalc s.allUsers error(%+v)", err)
		return
	}
	if len(userSlice) == 0 {
		log.Errorc(ctx, "answerWeekCalc s.allUsers empty", err)
		return
	}
	// 排序
	sort.Slice(userSlice, func(i, j int) bool {
		if userSlice[i].UserScore == userSlice[j].UserScore {
			if userSlice[i].AnswerTimes == userSlice[j].AnswerTimes {
				if userSlice[i].FinishTime == userSlice[j].FinishTime {
					return userSlice[i].ID < userSlice[j].ID
				}
				return userSlice[i].FinishTime < userSlice[j].FinishTime
			}
			return userSlice[i].AnswerTimes < userSlice[j].AnswerTimes
		}
		return userSlice[i].UserScore > userSlice[j].UserScore
	})
	var topCount int
	for _, userInfo := range userSlice {
		// 风控用户不进排行榜，并且分数清0
		if gaiaReply, err = s.checkRisk(ctx, strWeek, userInfo); err != nil {
			log.Errorc(ctx, "answerWeekCalc s.checkRisk mid(%d) error(%+v)", userInfo.Mid, err)
		} else if gaiaReply != nil && gaiaReply.Exist {
			log.Errorc(ctx, "answerWeekCalc  ExistKeyInList  Exist is true mid(%d)", userInfo.Mid)
			continue
		}
		// 前200 score 大于用户
		if topCount < 200 && userInfo.UserScore > 0 {
			topCount++
			roundTop = append(roundTop, userInfo.Mid)
			continue
		}
		break
	}
	// 每周三要删除排行榜
	if isClear {
		for i := 0; i < 3; i++ {
			if err = s.question.AddCacheUserTop(ctx, []*quesmdl.UserRank{}); err == nil {
				log.Info("answerWeekCalc hourtop s.question.AddCacheUserTop ok")
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
		if err != nil { // 删除小时榜，可以忽略错误
			log.Errorc(ctx, "answerWeekCalc s.question.AddCacheUserTop error(%+v)", err)
		}
	}
	// 每周三输出top 200
	for i := 0; i < 3; i++ {
		if len(roundTop) > 0 {
			strRoundTop := xstr.JoinInts(roundTop)
			if err = s.question.AddCacheRoundTop(ctx, strRoundTop); err == nil { // 输出一个接口，也可以从数据平台，每天一张表中获取
				log.Info("answerWeekCalc weektop s.question.AddCacheRoundTop ok")
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
	if err != nil { // 可以忽略错误，下边执行更新数据库
		log.Errorc(ctx, "answerWeekCalc s.question.AddCacheUserTop error(%+v)", err)
	}
	// 更新数据库
	s.updateUserDB(ctx, userSlice)
}

func (s *Service) updateUserDB(ctx context.Context, userSlice []*quesmdl.AnswerUserInfo) {
	var err error
	if len(userSlice) == 0 {
		log.Error("updateUserDB userSlice empty")
		return
	}
	// 更新数据库 每天0点更新数据库
	if err = s.question.UpUserRankZero(ctx); err != nil {
		log.Errorc(ctx, "updateUserDB s.dao.UpUserRankZero error(%+v)", err)
		return
	}
	for _, userInfo := range userSlice {
		if err = s.question.UpdateUserData(ctx, userInfo); err != nil {
			log.Errorc(ctx, "updateUserDB s.dao.UpdateUserData userInfo(%+v) error(%+v)", userInfo, err)
			continue
		}
		time.Sleep(10 * time.Millisecond)
	}
	log.Info("answerWeekCalc updateUserDB Answer Day UpdateUserDB success")
}

func (s *Service) allUsers(ctx context.Context, strWeek string) (res []*quesmdl.AnswerUserInfo, err error) {
	var pn int64
	for i := 0; i < _maxBigPn; i++ {
		vUsers, err := s.question.AnswerUsers(ctx, pn)
		if err != nil {
			log.Error("allUsers i(%d) error(%+v)", i, err)
			time.Sleep(time.Second)
			continue
		}
		if len(vUsers) == 0 {
			log.Info("allUsers success i(%d)", i)
			break
		}
		for _, vUser := range vUsers {
			pn = vUser.ID
			userInfo, e := s.question.CacheUserInfo(ctx, vUser.Mid, strWeek)
			if e != nil {
				log.Errorc(ctx, "allUsers s.question.CacheUserInfo mid(%d) error(%+v)", vUser.Mid, e)
				continue
			} else if userInfo == nil {
				log.Errorc(ctx, "allUsers s.question.CacheUserInfo mid(%d) userInfo is nil", vUser.Mid)
				continue
			}
			userInfo.ID = vUser.ID
			userInfo.Mid = vUser.Mid
			res = append(res, userInfo)
		}
	}
	log.Info("allUsers join people count(%d)", len(res))
	return
}

func (s *Service) getUserInfoKey() string {
	nowTime := time.Now().Unix()
	for _, currentKey := range s.c.S10Answer.UserInfoKey {
		currentTime, _ := time.ParseInLocation("20060102", currentKey, time.Local)
		if nowTime >= currentTime.Unix() {
			return currentKey
		}
	}
	return ""
}

func (s *Service) checkRisk(ctx context.Context, strWeek string, userInfo *quesmdl.AnswerUserInfo) (gaiaReply *gaialibapi.ExistKeyInListReply, err error) {
	arg := &gaialibapi.ExistKeyInListReq{
		Key:     strconv.FormatInt(userInfo.Mid, 10),
		ListStr: "s10_answer_black",
	}
	for i := 0; i < _retry; i++ {
		if gaiaReply, err = s.gaialibClient.ExistKeyInList(ctx, arg); err == nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if err != nil {
		log.Errorc(ctx, "checkRisk s.gaialibClient.ExistKeyInList arg(%+v) error(%+v)", arg, err)
		return
	}
	if gaiaReply != nil && gaiaReply.Exist {
		userInfo.UserScore = 0
		if e := s.question.AddCacheUserInfo(ctx, userInfo.Mid, strWeek, userInfo); e != nil {
			log.Errorc(ctx, "checkRisk s.question.AddCacheUserInfo  mid(%d)", userInfo.Mid, e)
		}
	}
	return
}

type Auth struct {
	User     string `json:"user"`
	Password string `json:"password"`
}

type Latex2pngReq struct {
	Auth       Auth   `json:"auth"`
	Latex      string `json:"latex"`
	Resolution int64  `json:"resolution"`
	Color      string `json:"color"`
}

func (s *Service) UpdateQuestionRecord() {
	nowTs := time.Now().Unix()
	startTime := xtime.Time(nowTs - _prevStime)
	endTime := xtime.Time(nowTs)

	ctx := context.Background()
	bases, err := s.question.RawBases(ctx, startTime, endTime)
	if err != nil {
		log.Errorc(ctx, "UpdateQuestionRecord s.question.RawBases() error(%v)", err)
	}
	gaoKaoAct := make(map[int64]bool)
	for _, v := range s.c.GaoKaoAnswer.ForeignId {
		gaoKaoAct[v] = true
	}
	log.Infoc(ctx, "UpdateQuestionRecord  start:%+v", gaoKaoAct)
	for _, v := range bases {
		if _, ok := gaoKaoAct[v.ForeignID]; ok {
			items, err := s.question.RawDetailList(context.Background(), v.ID, quesmdl.State4Process)
			if err != nil {
				log.Errorc(ctx, "UpdateQuestionRecord  s.question.RawDetailList(%d) error(%v)", v.ID, err)
				continue
			}
			for _, detail := range items {
				if detail.Name, err = s.TransLatexInfo(ctx, detail.Name); err != nil {
					log.Errorc(ctx, "UpdateQuestionRecord transLatexInfo name(%s) error(%+v)", detail.Name, err)
					continue
				}
				if detail.RightAnswer, err = s.TransLatexInfo(ctx, detail.RightAnswer); err != nil {
					log.Errorc(ctx, "UpdateQuestionRecord transLatexInfo RightAnswer:(%s) error(%+v)", detail.RightAnswer, err)
					continue
				}
				if detail.WrongAnswer, err = s.TransLatexInfo(ctx, detail.WrongAnswer); err != nil {
					log.Errorc(ctx, "UpdateQuestionRecord transLatexInfo WrongAnswer:(%s) error(%+v)", detail.WrongAnswer, err)
					continue
				}
				if err = s.question.UpdateQuestionDetail(ctx, detail, quesmdl.StateInit); err != nil {
					log.Errorc(ctx, "UpdateQuestionRecord (%d) error(%v)", detail.ID, err)
				}
			}
		}
	}
}

func (s *Service) TransLatexInfo(ctx context.Context, content string) (newContent string, err error) {

	newContent = content
	spitTag := "$$$"
	if s.c.GaoKaoAnswer.SpitTag != "" {
		spitTag = s.c.GaoKaoAnswer.SpitTag
	}
	if content != "" && strings.Contains(content, spitTag) {
		contentList := strings.Split(content, spitTag)
		if len(contentList)%2 != 0 {
			for index, txt := range contentList {
				if index%2 != 0 {
					var png string
					if png, err = s.LaTeX2PNG(ctx, txt); err != nil {
						log.Errorc(ctx, "LaTeX2PNG fail :%+v ", err)
						return
					}
					var bosUrl string
					if bosUrl, err = s.UploadPng2Bos(ctx, png); err != nil {
						log.Errorc(ctx, "UploadPng2Bos fail :%+v ", err)
						return
					}
					contentList[index] = fmt.Sprintf("[%s]", bosUrl)
				}
			}
			newContent = strings.Join(contentList, "")
		}
	}
	return
}

func (s *Service) LaTeX2PNG(ctx context.Context, latex string) (png string, err error) {
	var msgBytes []byte
	urlHost := "http://www.latex2png.com/api/convert"

	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	if err = encoder.Encode(Latex2pngReq{
		Auth: Auth{
			User:     "guest",
			Password: "guest",
		},
		Latex:      latex,
		Resolution: 600,
		Color:      "000000",
	}); err != nil {
		return
	}
	msgBytes = buffer.Bytes()

	shellStr := fmt.Sprintf("curl --location --request POST '%s' --data '%s'", urlHost, msgBytes)
	out, err := exec.Command("bash", "-c", shellStr).Output()
	res := &struct {
		ResultCode    int    `json:"result_code"`
		ResultMessage string `json:"result_message"`
		Url           string `json:"url"`
	}{}

	if err = json.Unmarshal(out, &res); err != nil {
		log.Errorc(ctx, "LaTeX2PNG d.client.Do msgBytes:%s error(%+v)", msgBytes, err)
		return
	}
	log.Infoc(ctx, "LaTeX2PNG d.client.Do params:%s replay:%+v", msgBytes, *res)
	if res.ResultCode != 0 {
		return "", errors.Wrapf(err, "result_code:%v , result_message:%v", res.ResultCode, res.ResultMessage)
	}
	return "http://www.latex2png.com" + res.Url, nil
}

func (s *Service) UploadPng2Bos(ctx context.Context, pngUrl string) (bosUrl string, err error) {

	localFileName := fmt.Sprintf("/tmp/activity_gaokao_latex_%v.png", time.Now().UnixNano())
	if _, err = exec.Command("bash", "-c", fmt.Sprintf("curl -o  %s  %s", localFileName, pngUrl)).Output(); err != nil {
		return
	}
	if tool.IsFileExists(localFileName) {
		var filename string
		if filename, err = getFileName(localFileName); err != nil {
			return
		}

		var reader *os.File
		if reader, err = os.Open(localFileName); err != nil {
			log.Warnc(ctx, "Impossible to open the file:%v , %+v", localFileName, err)
			return
		}

		defer reader.Close()

		return boss.Client.UploadObject(ctx, boss.Bucket, fmt.Sprintf("actgaokao/latex/%v.png", filename), reader)
	}
	err = errors.New(fmt.Sprintf("download png:%v  fail", pngUrl))
	return
}

func getFileName(localPath string) (filename string, err error) {
	var reader *os.File
	reader, err = os.Open(localPath)
	if err != nil {
		return
	}
	defer reader.Close()
	var im image.Config

	if im, _, err = image.DecodeConfig(reader); err != nil {
		err = errors.Wrapf(err, "can not decode :%s", localPath)
		return
	}
	filename = fmt.Sprintf("gk_act_%d_%d_%d", time.Now().UnixNano(), im.Width, im.Height)
	return
}
