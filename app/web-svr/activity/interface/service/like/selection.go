package like

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
	"unicode/utf8"

	accmdl "git.bilibili.co/bapis/bapis-go/account/service"
	fliapi "git.bilibili.co/bapis/bapis-go/filter/service"
	fligrpc "git.bilibili.co/bapis/bapis-go/filter/service"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	arcmdl "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/client"
	"go-gateway/app/web-svr/activity/interface/model/like"
	"go-gateway/pkg/idsafe/bvid"
)

var (
	selCategory atomic.Value
	_emptyPR    = make([]*like.ProductRole, 0)
	_emptyArc   = make([]*like.ArcBvInfo, 0)
)

const (
	_fileOnePath       = "/data/selectionTableOne.csv"
	_fileTwoPath       = "/data/selectionTableTwo_%d.csv"
	_answersCount      = 5
	_hotOrder          = 0
	_timeOrder         = 1
	_percentDecimalFmt = "%.3f"
	_riskAction        = "cartoon_vote"
	_riskActivityUID   = "cartoon_selection"
	_riskApi           = "/x/activity/selection/vote"
)

func init() {
	categorys := make([]*like.SelCategory, 0)
	selCategory.Store(categorys)
}

func (s *Service) selectionDBFmt(list []*like.SelectionQADB) (res []*like.SelectionQA) {
	var (
		questionOrders []int64
		questionMap    map[int64][]*like.SelectionAnswer
		questionName   map[int64]string
	)
	questionMap = make(map[int64][]*like.SelectionAnswer, 10)
	questionName = make(map[int64]string, 10)
	for _, n := range list {
		if n.Product != "" || n.Role != "" {
			questionMap[n.QuestionOrder] = append(questionMap[n.QuestionOrder], &like.SelectionAnswer{Product: n.Product, Role: n.Role})
		}
		if _, ok := questionName[n.QuestionOrder]; !ok {
			questionOrders = append(questionOrders, n.QuestionOrder)
			questionName[n.QuestionOrder] = n.Question
		}
	}
	for _, questionOrder := range questionOrders {
		answers, ok := questionMap[questionOrder]
		if !ok {
			answers = make([]*like.SelectionAnswer, 0)
		}
		res = append(res, &like.SelectionQA{
			Question: questionName[questionOrder],
			Answer:   answers,
		})
	}
	return
}

func (s *Service) selectionDB(ctx context.Context, mid int64) (res []*like.SelectionQA, err error) {
	var list []*like.SelectionQADB
	if list, err = s.dao.SelSelectionDBsByMid(ctx, mid); err != nil {
		log.Errorc(ctx, "selectionDB s.dao.SelSelectionDBsByMid mid(%d) error(%+v)", mid, err)
		return
	}
	res = s.selectionDBFmt(list)
	return
}

func (s *Service) selInfo(ctx context.Context, mid int64) (res []*like.SelectionQA, err error) {
	addCache := true
	res, err = s.dao.CacheSelectionInfo(ctx, mid)
	if err != nil {
		log.Errorc(ctx, "selInfo s.dao.CacheSelectionInfo mid(%d) error(%+v)", mid, res)
		addCache = false
		err = nil
	}
	defer func() {
		if len(res) == 1 && res[0] != nil && res[0].Question == "-1" {
			res = nil
		}
	}()
	if len(res) != 0 {
		return
	}
	if res, err = s.selectionDB(ctx, mid); err != nil {
		log.Errorc(ctx, "selInfo s.dao.SelSelectionQA mid(%d) error(%+v)", mid, err)
		return
	}
	miss := res
	if len(miss) == 0 {
		miss = []*like.SelectionQA{{Question: "-1"}}
	}
	if !addCache {
		return
	}
	s.cache.Do(ctx, func(c context.Context) {
		s.dao.AddCacheSelectionInfo(ctx, mid, miss)
	})
	return
}

func (s *Service) SelectionInfo(ctx context.Context, mid int64) (res *like.SelectionQAInfo, err error) {
	var (
		list  []*like.SelectionQA
		selQA interface{}
	)
	if err = s.checkSelAct(); err != nil {
		return
	}
	if err = s.checkUser(ctx, mid); err != nil {
		return
	}
	if list, err = s.selInfo(ctx, mid); err != nil {
		return
	}
	isJoin := len(list) > 0
	if isJoin {
		selQA = list
	} else {
		selQA = struct{}{}

	}
	res = &like.SelectionQAInfo{
		IsJoin:      isJoin,
		SelectionQA: selQA,
	}
	return
}

func (s *Service) checkUser(ctx context.Context, mid int64) (err error) {
	var user *accmdl.Profile
	if user, err = s.userInfo(ctx, mid); err != nil {
		err = ecode.ActivityLotteryNetWorkError
		log.Errorc(ctx, "SelectionInfo s.userInfo mid(%d) error(%+v)", mid, err)
		return
	}
	if user.JoinTime >= s.c.Selection.JoinTime {
		err = ecode.ActivitySelectionJoinErr
	}
	return
}

func (s *Service) SelectionSensitive(ctx context.Context, answers string) (res *like.SelSensitive, err error) {
	if err = s.checkSelAct(); err != nil {
		return
	}
	var filRly *fliapi.FilterReply
	res = &like.SelSensitive{}
	if filRly, err = client.FilterClient.Filter(ctx, &fligrpc.FilterReq{
		Area:    "activity",
		Keys:    []string{"act:anime_vote_2020"},
		Message: answers}); err != nil {
		log.Errorc(ctx, "SelectionSensitive s.fliClient.Filter(%s)", answers)
		return
	}
	if filRly != nil && filRly.Level >= s.c.Selection.FilterLevel {
		log.Errorc(ctx, "SelectionSensitive s.fliClient.Filter(%s) level(%d)", answers, filRly.Level)
		res.IsSensitive = true
	}
	return
}

func (s *Service) checkSelAct() error {
	nowTs := time.Now().Unix()
	if nowTs < s.c.Selection.BeginTime {
		return ecode.ActivityNotStart
	}
	if nowTs > s.c.Selection.EndTime {
		return ecode.ActivityOverEnd
	}
	return nil
}

func (s *Service) SelectionSubmit(ctx context.Context, mid int64, contests string) (err error) {
	var (
		addAffected   int64
		params, dbRes []*like.SelectionQA
		check         bool
		nxKey         = fmt.Sprintf("sel_submit_%d", mid)
		filterMsg     string
	)
	defer func() {
		if addAffected == 0 && err != nil && !xecode.EqualError(ecode.ActivityRapid, err) {
			s.dao.RsDelNX(ctx, nxKey)
		}
	}()
	if check, err = s.dao.RsSetNX(ctx, nxKey, 1); err != nil || !check {
		log.Warn("SelectionSubmit s.dao.RsSetNX mid:%d to fast err:%v", mid, err)
		err = ecode.ActivityRapid
		return
	}
	if err = s.checkUser(ctx, mid); err != nil {
		return
	}
	if err = s.checkSelAct(); err != nil {
		return err
	}
	if selInfo, e := s.selInfo(ctx, mid); e != nil {
		log.Errorc(ctx, "SelectionSubmit s.selInfo mid(%d) error(%+v)", mid, err)
		err = ecode.ActivityLotteryNetWorkError
		return
	} else if len(selInfo) > 0 {
		err = ecode.ActivityRepeatSubmit
		return
	}
	if err = json.Unmarshal([]byte(contests), &params); err != nil {
		log.Errorc(ctx, "SelectionSubmit mid(%d) contests(%s) error(%+v)", mid, contests, err)
		err = ecode.ActivityLotteryNetWorkError
		return
	}
	if len(params) == 0 {
		log.Errorc(ctx, "SelectionSubmit mid(%d) contests(%s) params count 0", mid, contests)
		err = ecode.ActivitySelectionOneErr
		return
	}
	productRoleMap := s.productRoleQuestions()
	// 去掉重复作品名
	for order, userQA := range params {
		var (
			answerMap map[string]struct{}
			answers   []*like.SelectionAnswer
		)
		int64Order := int64(order + 1)
		answerMap = make(map[string]struct{}, 5)
		if len(userQA.Answer) == 0 {
			answers = make([]*like.SelectionAnswer, 0)
		} else {
			for _, userAnswer := range userQA.Answer {
				if _, ok := productRoleMap[int64Order]; ok {
					if userAnswer.Product == "" || userAnswer.Role == "" || utf8.RuneCountInString(userAnswer.Product) > s.c.Selection.LimitWords || utf8.RuneCountInString(userAnswer.Role) > s.c.Selection.LimitWords {
						log.Errorc(ctx, "userAnswer mid(%d) Product(%s) Role(%s)", mid, userAnswer.Product, userAnswer.Role)
						err = xecode.RequestErr
						return
					}
				} else {
					if userAnswer.Product == "" || utf8.RuneCountInString(userAnswer.Product) > s.c.Selection.LimitWords {
						err = xecode.RequestErr
						log.Errorc(ctx, "userAnswer mid(%d) Product(%s) Role(%s)", mid, userAnswer.Product, userAnswer.Role)
						return
					}
				}
				productAndRole := userAnswer.Product + userAnswer.Role
				if filterMsg == "" {
					filterMsg = productAndRole
				} else {
					filterMsg += "\n" + productAndRole
				}
				if _, ok := answerMap[productAndRole]; !ok {
					answers = append(answers, &like.SelectionAnswer{
						Product: userAnswer.Product,
						Role:    userAnswer.Role,
					})
					answerMap[productAndRole] = struct{}{}
				}
			}
		}
		dbRes = append(dbRes, &like.SelectionQA{
			Question: userQA.Question,
			Answer:   answers,
		})
	}
	sensitive, e := s.SelectionSensitive(ctx, filterMsg)
	if e != nil {
		log.Errorc(ctx, "SelectionSubmit s.SelectionSensitive mid(%d) message(%s) error(%+v)", mid, filterMsg, e)
		err = ecode.ActivityLotteryNetWorkError
		return
	}
	if sensitive != nil && sensitive.IsSensitive {
		log.Errorc(ctx, "SelectionSubmit filter true mid(%d) message(%s)", mid, filterMsg)
		err = xecode.RequestErr
		return
	}
	if addAffected, err = s.dao.AddSelectionQA(ctx, mid, dbRes); err != nil {
		log.Errorc(ctx, "SelectionSubmit mid(%+v) contests(%s) error(%+v)", mid, contests, err)
		err = ecode.ActivitySelectionAddErr
		return
	}
	if addAffected > 0 {
		if e := s.dao.AddCacheSelectionInfo(ctx, mid, dbRes); e != nil {
			log.Errorc(ctx, "SelectionSubmit s.dao.AddCacheSelectionInfo mid(%d) error(%+v)", mid, e)
			s.cache.Do(ctx, func(ctx context.Context) {
				retry(func() error {
					return s.dao.DelCacheSelectionInfo(ctx, mid)
				})
			})
		}
	}
	return
}

func retry(callback func() error) error {
	var err error
	for i := 0; i < 3; i++ {
		if err = callback(); err == nil {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return err
}

func (s *Service) selAllSelection(ctx context.Context) (res []*like.SelectionQADB, err error) {
	id := int64(0)
	for {
		var list []*like.SelectionQADB
		if list, err = s.dao.SelAllSelection(ctx, id); err != nil {
			log.Error("s.dao.SelAllSelection(%d) error(%+v)", id)
			return
		}
		count := len(list)
		if count == 0 {
			log.Info("SelAllSelection mid(%d) success count(%d)", id, count)
			break
		}
		id = list[count-1].ID
		res = append(res, list...)
	}
	return
}

func getVoteKey(mid, categoryID int64) string {
	return fmt.Sprintf("sel_vote_%d_%d", mid, categoryID)
}

func getVoteDayKey(mid, categoryID int64, nowDate string) string {
	return fmt.Sprintf("sel_vote_%d_%d_%s", mid, categoryID, nowDate)
}

// SeleList
func (s *Service) SeleList(ctx context.Context, mid, categoryID int64) (res *like.CategoryPR, err error) {
	var (
		midVote int
		list    []*like.ProductRole
		nxKey   string
	)
	nowTime := time.Now()
	if nowTime.Unix() >= s.c.Selection.NewKeyTime {
		nxKey = getVoteDayKey(mid, categoryID, nowTime.Format("20060102"))
	} else {
		nxKey = getVoteKey(mid, categoryID)
	}
	res = &like.CategoryPR{List: _emptyPR}
	res.IsLogin = mid > 0
	if midVote, err = s.dao.RiGet(ctx, nxKey); err != nil {
		log.Errorc(ctx, "SeleList s.dao.RiGet mid(%d) category(%d) error(%+v)", mid, categoryID, err)
		return
	}
	res.ShowVotes = s.c.Selection.ShowVotes == 1                              // 显示票数
	res.IsChecking = nowTime.Unix() > s.c.Selection.VoteEnd && !res.ShowVotes //大天结束投票时间并且不显示票数
	res.IsStart = nowTime.Unix() >= s.c.Selection.VoteBegin
	res.IsVote = midVote > 0
	if res.IsVote || res.ShowVotes {
		if list, err = s.showVotedList(ctx, categoryID, int64(midVote), res.ShowVotes); err != nil {
			log.Errorc(ctx, "SeleList s.showVotedList mid(%d) category(%d) error(%+v)", mid, categoryID, err)
			return
		}
	} else { // 无投票显示
		if list, err = s.dao.CachePrNotVote(ctx, categoryID); err != nil {
			log.Errorc(ctx, "SeleList s.dao.CachePrNotVote  mid(%d) category(%d) error(%+v)", mid, categoryID, err)
			return
		}
	}
	if len(list) == 0 {
		list = _emptyPR
	}
	res.List = list
	return
}

func (s *Service) showVotedList(ctx context.Context, categoryID, voteProductRoleID int64, showVote bool) (list []*like.ProductRole, err error) {
	var (
		voteMap    map[int64]*like.ProductroleVote
		maxVoteNum int64
	)
	if list, voteMap, maxVoteNum, err = s.dao.CacheProductRoles(ctx, categoryID); err != nil {
		log.Errorc(ctx, "SeleList showVotedList s.dao.CacheProductRoles category(%d) error(%+v)", categoryID, err)
		return
	}
	for _, pr := range list {
		if pr.ID == voteProductRoleID {
			pr.Voted = true
		}
		if voteMtime, ok := voteMap[pr.ID]; ok {
			if maxVoteNum > 0 {
				pr.Percent = float64(voteMtime.VoteNum) / float64(maxVoteNum)
				pr.Percent, _ = strconv.ParseFloat(fmt.Sprintf(_percentDecimalFmt, pr.Percent), 64)
			}
			if showVote {
				pr.VoteNum = voteMtime.VoteNum
			}
			pr.HideVote = voteMtime.VoteNum
			pr.Mtime = voteMtime.Mtime
		}
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].HideVote != list[j].HideVote {
			return list[i].HideVote > list[j].HideVote
		}
		if list[i].Mtime != list[j].Mtime {
			return list[i].Mtime < list[j].Mtime
		}
		return list[i].ID < list[j].ID
	})
	curOrder := -1
	curVoteNum := int64(-1)
	for order, pr := range list {
		if pr.HideVote != curVoteNum {
			curVoteNum = pr.HideVote
			curOrder = order + 1
		}
		pr.OrderNum = curOrder
	}
	return
}

func getExpire() int32 {
	nowTime := time.Now()
	year, month, day := nowTime.Date()
	tm2 := time.Date(year, month, day, 23, 59, 59, 0, nowTime.Location())
	dayLongTime := tm2.Unix() + 1 - nowTime.Unix()
	tm2.Unix()
	return int32(dayLongTime)
}

func loadCategoryMap(categoryID int64) (res map[int64]*like.SelCategory) {
	res = make(map[int64]*like.SelCategory, 6)
	list := selCategory.Load().([]*like.SelCategory)
	for _, category := range list {
		res[categoryID] = category
	}
	return
}

func (s *Service) checkVoteUser(ctx context.Context, mid int64) (err error) {
	var user *accmdl.Profile
	if user, err = s.userInfo(ctx, mid); err != nil {
		err = ecode.ActivityLotteryNetWorkError
		log.Errorc(ctx, "SelectionVote s.userInfo mid(%d) error(%+v)", mid, err)
		return
	}
	if user.JoinTime >= s.c.Selection.VoteJoinTime {
		err = ecode.ActivitySelectionJoinErr
	}
	return
}

func (s *Service) checkVoteAct() error {
	nowTs := time.Now().Unix()
	if nowTs < s.c.Selection.VoteBegin {
		return ecode.ActivityNotStart
	}
	if nowTs > s.c.Selection.VoteEnd {
		return ecode.ActivityOverEnd
	}
	return nil
}

func (s *Service) SelectionRank(ctx context.Context, mid int64, params *like.ParamVote) (res int, err error) {
	var list []*like.ProductRole
	if list, err = s.showVotedList(ctx, params.CategoryID, params.ProductRoleID, true); err != nil {
		log.Errorc(ctx, "SelectionRank s.showVotedList mid(%d) category(%d) error(%+v)", mid, params.CategoryID, err)
		return
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].VoteNum != list[j].VoteNum {
			return list[i].VoteNum > list[j].VoteNum
		}
		if list[i].Mtime != list[j].Mtime {
			return list[i].Mtime < list[j].Mtime
		}
		return list[i].ID < list[j].ID
	})
	curOrder := -1
	curVoteNum := int64(-1)
	for order, pr := range list {
		if pr.VoteNum != curVoteNum {
			curVoteNum = pr.VoteNum
			curOrder = order + 1
		}
		if pr.ID == params.ProductRoleID {
			res = curOrder
			return
		}
	}
	return
}

func (s *Service) SelectionVote(ctx context.Context, mid int64, params *like.ParamVote) (err error) {
	var (
		upAffected int64
		check      bool
		nxKey      string
		list       []*like.ProductRole
		currentPR  *like.ProductRole
		voteMap    map[int64]*like.ProductroleVote
		risk       bool
		riskErr    error
		expire     int32
	)
	nowTime := time.Now()
	if nowTime.Unix() >= s.c.Selection.NewKeyTime {
		nxKey = getVoteDayKey(mid, params.CategoryID, nowTime.Format("20060102"))
		expire = 86400
	} else {
		nxKey = getVoteKey(mid, params.CategoryID)
		expire = getExpire()
	}
	if s.c.Selection.VoteSwitch == 1 {
		err = ecode.ActivityVoteCheckErr
		return
	}
	if err = s.checkVoteAct(); err != nil {
		return
	}
	loadCategoryMap := loadCategoryMap(params.CategoryID)
	if category, ok := loadCategoryMap[params.CategoryID]; !ok {
		err = xecode.RequestErr
		log.Errorc(ctx, "SelectionVote s.dao.CacheProductRoles mid(%d) category(%d) productroleID(%d) error(%+v)", mid, params.CategoryID, params.ProductRoleID, err)
		return
	} else {
		params.CategoryName = category.CategoryName
	}
	defer func() {
		if upAffected == 0 && err != nil && !xecode.EqualError(ecode.ActivityVoteRepeatErr, err) {
			s.dao.RsDelNX(ctx, nxKey)
		}
	}()
	if check, err = s.dao.SetNXValue(ctx, nxKey, params.ProductRoleID, expire); err != nil || !check {
		log.Errorc(ctx, "SelectionSubmit s.dao.SetNXValue mid:%d categoryID:%d productroleID:%d to fast err:%v", mid, params.CategoryID, params.ProductRoleID, err)
		err = ecode.ActivityVoteRepeatErr
		return
	}
	if err = s.checkVoteUser(ctx, mid); err != nil {
		return
	}
	if list, voteMap, _, err = s.dao.CacheProductRoles(ctx, params.CategoryID); err != nil {
		log.Errorc(ctx, "SelectionVote s.dao.CacheProductRoles mid(%d) category(%d) productroleID(%d) error(%+v)", mid, params.CategoryID, params.ProductRoleID, err)
		return
	}
	currentVote, ok := voteMap[params.ProductRoleID]
	if !ok {
		log.Errorc(ctx, "SelectionVote s.dao.CacheProductRoles mid(%d) category(%d) productroleID(%d) voteMap(%+v) not ok", mid, params.CategoryID, params.ProductRoleID, voteMap)
		err = xecode.RequestErr
		return
	}
	for _, pr := range list {
		if pr.ID == params.ProductRoleID {
			currentPR = pr
			break
		}
	}
	if currentPR == nil {
		log.Errorc(ctx, "SelectionVote s.dao.CacheProductRoles mid(%d) category(%d) productroleID(%d) currentPR not exists", mid, params.CategoryID, params.ProductRoleID)
		err = xecode.RequestErr
		return
	}
	if currentPR.CategoryType == 1 {
		params.ProductName = currentPR.Role
	} else {
		params.ProductName = currentPR.Product
	}
	if risk, riskErr = s.checkRisk(ctx, mid, params); riskErr != nil {
		log.Errorc(ctx, "SelectionVote s.checkRisk mid(%d) category(%d) productroleID(%d) error(%+v)", mid, params.CategoryID, params.ProductRoleID, riskErr)
	}
	if risk {
		err = ecode.ActivityVoteRiskErr
		log.Errorc(ctx, "SelectionVote s.checkRisk mid(%d) category(%d) productroleID(%d) risk is true error(%+v)", mid, params.CategoryID, params.ProductRoleID, err)
		return
	}
	if upAffected, err = s.dao.VoteTransact(ctx, mid, params.CategoryID, params.ProductRoleID, nowTime); err != nil {
		log.Errorc(ctx, "SelectionSubmit s.dao.UpPRVote mid:%d categoryID:%d productroleID:%d err:%v", mid, params.CategoryID, params.ProductRoleID, err)
		return
	}
	if e := s.SetVoteCache(ctx, mid, params.CategoryID, currentPR, currentVote, nowTime); e != nil {
		log.Errorc(ctx, "SelectionSubmit s.SetVoteCache mid:%d categoryID:%d productroleID:%d err:%v", mid, params.CategoryID, params.ProductRoleID, e)
	}
	return
}

func (s *Service) checkRisk(ctx context.Context, mid int64, params *like.ParamVote) (res bool, err error) {
	otherEventCtx := &like.VoteEventCtx{
		Action:       _riskAction,
		Mid:          mid,
		ActivityUid:  _riskActivityUID,
		ID:           params.ProductRoleID,
		Content:      params.ProductName,
		CategoryID:   params.CategoryID,
		CategoryName: params.CategoryName,
		Buvid:        params.Buvid,
		Ip:           metadata.String(ctx, metadata.RemoteIP),
		Platform:     params.Platform,
		Ctime:        time.Now().Format("2006-01-02 15:04:05"),
		Api:          _riskApi,
		Origin:       params.Origin,
		UserAgent:    params.UA,
		Build:        params.Build,
		Referer:      params.Referer,
		MobiApp:      params.MobiApp,
	}
	if res, err = s.silverDao.RuleCheck(ctx, _riskAction, otherEventCtx); err != nil {
		log.Errorc(ctx, "SelectionVote checkRisk mid(%d) otherEventCtx(%+v) error(%+v)", mid, otherEventCtx, err)
	}
	return
}

func (s *Service) SetVoteCache(ctx context.Context, mid, categoryID int64, pr *like.ProductRole, currentVote *like.ProductroleVote, nowTime time.Time) (err error) {
	currentVote.Mtime = nowTime.Unix()
	currentVote.VoteNum++
	if err = s.dao.AddCachePRVote(ctx, categoryID, pr, currentVote); err == nil {
		return
	}
	log.Errorc(ctx, "SelectionSubmit  SetVoteCache mid(%d)  categoryID(%d) productrole(%+v) error(%+v)", mid, categoryID, pr)
	s.cache.Do(ctx, func(ctx context.Context) {
		retry(func() error {
			return s.dao.AddCachePRVote(ctx, categoryID, pr, currentVote)
		})
	})
	return
}

func (s *Service) SeleAssistance(ctx context.Context, params *like.ParamAssistance) (res *like.AssistanceRes, err error) {
	var (
		aids  []int64
		arcs  *arcmdl.ArcsReply
		list  []*like.ArcBvInfo
		total int
	)
	res = &like.AssistanceRes{
		Page: &like.Page{
			Num:  params.Pn,
			Size: params.Ps,
		},
		List: _emptyArc,
	}
	if res.Role, res.Product, err = s.prName(ctx, params.CategoryID, params.ProductRoleID); err != nil {
		log.Errorc(ctx, "SeleAssistance s.prName() params(%+v) error(%+v)", params, err)
		return
	}
	if res.Product == "" && res.Role == "" {
		err = xecode.RequestErr
		log.Errorc(ctx, "SeleAssistance productrole not exists params(%+v) error(%+v)", params, err)
		return
	}
	if params.OrderType == _hotOrder {
		if aids, total, err = s.hotAids(ctx, params); err != nil {
			log.Errorc(ctx, "SeleAssistance s.hotAids params(%+v) error(%+v)", params, err)
			return
		}
	} else if params.OrderType == _timeOrder {
		if aids, total, err = s.timeAids(ctx, params); err != nil {
			log.Errorc(ctx, "SeleAssistance s.timeAids params(%+v) error(%+v)", params, err)
			return
		}
	} else {
		err = xecode.RequestErr
		return
	}
	if len(aids) == 0 || total == 0 {
		return
	}
	if arcs, err = client.ArchiveClient.Arcs(ctx, &arcmdl.ArcsRequest{Aids: aids}); err != nil {
		log.Error("s.arcClient.Archives3(%v) error(%v)", aids, err)
		return
	}
	for _, aid := range aids {
		if arc, ok := arcs.Arcs[aid]; ok {
			bvidStr, e := bvid.AvToBv(arc.Aid)
			if e != nil {
				continue
			}
			HideArcAttribute(arc)
			list = append(list, &like.ArcBvInfo{
				Arc:  arc,
				Bvid: bvidStr,
			})
		}
	}
	res.Page.Total = int64(total)
	res.List = list
	return
}

func (s *Service) prName(ctx context.Context, categoryID, productRoleID int64) (role, product string, err error) {
	var list []*like.ProductRole
	if list, err = s.dao.CachePrNotVote(ctx, categoryID); err != nil {
		log.Errorc(ctx, "SeleList s.dao.CachePrNotVote  categoryID(%+v) productRoleID(%d) error(%+v)", categoryID, productRoleID, err)
		return
	}
	for _, pr := range list {
		if productRoleID == pr.ID {
			product = pr.Product
			role = pr.Role
			break
		}
	}
	return
}

func (s *Service) hotAids(ctx context.Context, params *like.ParamAssistance) (aids []int64, total int, err error) {
	var (
		assistanceArc []*like.ProductRoleArc
		start         = (params.Pn - 1) * params.Ps
		end           = start + params.Ps - 1
	)
	if assistanceArc, total, err = s.dao.CacheHotAssistance(ctx, params.ProductRoleID, start, end); err != nil {
		log.Errorc(ctx, "hotAids SeleAssistance s.dao.CacheHotAssistance params(%+v) error(%+v)", params, err)
		return
	}
	for _, arc := range assistanceArc {
		aids = append(aids, arc.Aid)
	}
	return
}

func (s *Service) timeAids(ctx context.Context, params *like.ParamAssistance) (aids []int64, total int, err error) {
	var (
		assistanceArc, timeArc []*like.ProductRoleArc
		start                  = (params.Pn - 1) * params.Ps
		end                    = start + params.Ps - 1
	)
	if assistanceArc, err = s.dao.CacheTimeAssistance(ctx, params.ProductRoleID); err != nil {
		log.Errorc(ctx, "SeleAssistance s.dao.CacheHotAssistance params(%+v) error(%+v)", params, err)
		return
	}
	total = len(assistanceArc)
	if total == 0 {
		return
	}
	sort.Slice(assistanceArc, func(i, j int) bool {
		if assistanceArc[i].PubDate != assistanceArc[j].PubDate {
			return assistanceArc[i].PubDate > assistanceArc[j].PubDate
		}
		return assistanceArc[i].Aid > assistanceArc[j].Aid
	})
	if total == 0 || total < start {
		return
	}
	if total > end+1 {
		timeArc = assistanceArc[start : end+1]
	} else {
		timeArc = assistanceArc[start:]
	}
	for _, arc := range timeArc {
		aids = append(aids, arc.Aid)
	}
	return
}

func storeCategory(list []*like.SelCategory) {
	selCategory.Store(list)
}

func (s *Service) initSelCategory() {
	if len(s.c.Selection.VoteStage) > 0 {
		storeCategory(s.c.Selection.VoteStage)
	}
}

func (s *Service) initProductRole() {
	ctx := context.Background()
	list := selCategory.Load().([]*like.SelCategory)
	for _, category := range list {
		tmpCategory := category
		go s.SetCategoryCache(ctx, tmpCategory)
	}
}

func (s *Service) SetCategoryCache(ctx context.Context, category *like.SelCategory) {
	if category.CategoryID == 0 {
		return
	}
	productRoles, err := s.dao.SelProductRoleByCategory(ctx, category.CategoryID)
	if err != nil {
		log.Errorc(ctx, "SetCategoryCache s.dao.SelProductRoleByCategory(%d) error(%+v)", category.CategoryID, err)
		return
	}
	if len(productRoles) == 0 {
		log.Errorc(ctx, "SetCategoryCache s.dao.SelProductRoleByCategory categoryID(%d) productRoles empty", category.CategoryID)
		return
	}
	if err = s.dao.SetCacheProductRoles(ctx, category.CategoryID, productRoles); err != nil {
		log.Errorc(ctx, "SetCategoryCache s.dao.SetCacheProductRoles categoryID(%d) error(%+v)", category.CategoryID, err)
		return
	}
	s.setCachePrNotVote(ctx, category, false)
}

func (s *Service) setCachePrNotVote(ctx context.Context, category *like.SelCategory, reSet bool) {
	field := "product"
	if category.CategoryType == 1 {
		field = "role"
	}
	productRoles, err := s.dao.SelPrNotVoteByCategory(ctx, category.CategoryID, field)
	if err != nil {
		log.Errorc(ctx, "SetCategoryCache s.dao.SelPrNotVoteByCategory(%d) error(%+v)", category.CategoryID, err)
		return
	}
	if err = s.dao.SetCachePrNotVote(ctx, category.CategoryID, productRoles, reSet); err != nil {
		log.Errorc(ctx, "SetCategoryCache s.dao.SetCachePrNotVote categoryID(%d) error(%+v)", category.CategoryID, err)
	}
	return
}

// ReSetCacheProductRole 分类投票回源DB
func (s *Service) ReSetCacheProductRole(ctx context.Context, categoryID int64) {
	c := context.Background()
	productRoles, err := s.dao.SelProductRoleByCategory(c, categoryID)
	if err != nil {
		log.Errorc(ctx, "ReSetSetCacheProductRole s.dao.SelProductRoleByCategory(%d) error(%+v)", categoryID, err)
		return
	}
	if err = s.dao.AddCacheProductRoles(ctx, categoryID, productRoles); err != nil {
		log.Errorc(ctx, "ReSetSetCacheProductRole s.dao.ReSetCacheProductRoles categoryID(%d) error(%+v)", categoryID, err)
	}
}

// ReSetCachePrNotVote 所有分类未投票回源DB
func (s *Service) ReSetCachePrNotVote() {
	ctx := context.Background()
	list := selCategory.Load().([]*like.SelCategory)
	for _, category := range list {
		s.setCachePrNotVote(ctx, category, true)
	}
}

func (s *Service) ProductRoleMaxVote(ctx context.Context, categoryID int64) (maxVoteNum int64, err error) {
	if _, _, maxVoteNum, err = s.dao.CacheProductRoles(ctx, categoryID); err != nil {
		log.Errorc(ctx, "SeleList s.dao.CacheProductRoles category(%d) error(%+v)", categoryID, err)
	}
	return
}

// ExportTableOne .
func (s *Service) ExportTableOne() {
	var (
		midSel map[int64][]*like.SelectionQADB
	)
	ctx := context.Background()
	listAll, err := s.selAllSelection(ctx)
	midSel = make(map[int64][]*like.SelectionQADB, len(listAll))
	if err != nil {
		log.Errorc(ctx, "s.selAllSelection  error(%+v)", err)
		return
	}
	for _, list := range listAll {
		midSel[list.Mid] = append(midSel[list.Mid], list)
	}
	if len(midSel) > 0 {
		s.selectionsOneIntoLocalFile(midSel)
	}
}

func (s *Service) selectionsOneIntoLocalFile(midSel map[int64][]*like.SelectionQADB) {
	f, err := os.Create(_fileOnePath)
	if err != nil {
		return
	}
	defer func() {
		_ = f.Close()
	}()
	f.Write([]byte("\xEF\xBB\xBF"))
	productRoleMap := s.productRoleQuestions()
	w := csv.NewWriter(f)
	_ = w.Write([]string{
		"mid", "问题1", "作品名1", "作品名2", "作品名3", "作品名4", "作品名5",
		"问题2", "作品名1", "作品名2", "作品名3", "作品名4", "作品名5",
		"问题3", "角色名1", "作品名1", "角色名2", "作品名2", "角色名3", "作品名3", "角色名4", "作品名4", "角色名", "作品名5",
		"问题4", "角色名1", "作品名1", "角色名2", "作品名2", "角色名3", "作品名3", "角色名4", "作品名4", "角色名", "作品名5",
		"问题5", "作品名1", "作品名2", "作品名3", "作品名4", "作品名5",
		"问题6", "作品名1", "作品名2", "作品名3", "作品名4", "作品名5",
	})
	for mid, list := range midSel {
		var (
			answers             []string
			questionMap         map[int64]struct{}
			questionBeforeMap   map[int64]int
			beforeQuestionOrder int64
		)
		questionMap = make(map[int64]struct{}, 10)
		questionBeforeMap = make(map[int64]int, 10)
		answers = append(answers, strconv.FormatInt(mid, 10))
		sort.Slice(list, func(i, j int) bool {
			if list[i].QuestionOrder == list[j].QuestionOrder {
				return list[i].ID < list[j].ID
			}
			return list[i].QuestionOrder < list[j].QuestionOrder
		})
		for _, selection := range list {
			if _, ok := questionMap[selection.QuestionOrder]; !ok {
				if questionBeforeMap[beforeQuestionOrder] > 0 && questionBeforeMap[beforeQuestionOrder] < 5 {
					for i := 0; i <= _answersCount-questionBeforeMap[beforeQuestionOrder]; i++ {
						answers = append(answers, "")
						if _, ok := productRoleMap[beforeQuestionOrder]; ok {
							answers = append(answers, "")
						}
					}
				}
				answers = append(answers, selection.Question)
				beforeQuestionOrder = selection.QuestionOrder
				questionBeforeMap[beforeQuestionOrder]++
			}
			if _, ok := productRoleMap[beforeQuestionOrder]; ok {
				answers = append(answers, selection.Role)
			}
			answers = append(answers, selection.Product)
			questionMap[selection.QuestionOrder] = struct{}{}
			questionBeforeMap[beforeQuestionOrder]++
		}
		if questionBeforeMap[beforeQuestionOrder] < 5 {
			for i := 1; i < _answersCount-questionBeforeMap[beforeQuestionOrder]; i++ {
				answers = append(answers, "")
				if _, ok := productRoleMap[beforeQuestionOrder]; ok {
					answers = append(answers, "")
				}
			}
		}
		_ = w.Write(answers)
	}
	w.Flush()
}

func (s *Service) productRoleQuestions() (res map[int64]struct{}) {
	res = make(map[int64]struct{}, len(s.c.Selection.ProductRole))
	for _, order := range s.c.Selection.ProductRole {
		res[order] = struct{}{}
	}
	return res
}

// ExportTableTwo .
func (s *Service) ExportTableTwo() {
	var (
		QuestionOrderSel map[int64][]*like.SelectionQADB
	)
	ctx := context.Background()
	listAll, err := s.selAllSelection(ctx)
	QuestionOrderSel = make(map[int64][]*like.SelectionQADB, len(listAll))
	if err != nil {
		log.Errorc(ctx, "s.selAllSelection  error(%+v)", err)
		return
	}
	for _, list := range listAll {
		QuestionOrderSel[list.QuestionOrder] = append(QuestionOrderSel[list.QuestionOrder], list)
	}
	if len(QuestionOrderSel) > 0 {
		for questionOrder, list := range QuestionOrderSel {
			s.selectionsTwoIntoLocalFile(questionOrder, list)
		}
	}
}

func (s *Service) selectionsTwoIntoLocalFile(loopIndex int64, QuestionSel []*like.SelectionQADB) {
	filePath := fmt.Sprintf(_fileTwoPath, loopIndex)
	f, err := os.Create(filePath)
	if err != nil {
		return
	}
	defer func() {
		_ = f.Close()
	}()
	productRoleMap := s.productRoleQuestions()
	_, isProductRole := productRoleMap[loopIndex]
	answerCountMap := make(map[string]*like.TwoRes, 1000)
	for _, list := range QuestionSel {
		var key string
		product := strings.ToLower(list.Product)
		role := strings.ToLower(list.Role)
		if isProductRole {
			key = product + role
		} else {
			key = list.Product
		}
		if _, ok := answerCountMap[key]; !ok {
			answerCountMap[key] = &like.TwoRes{
				Product: list.Product,
				Role:    list.Role,
			}
		}
		answerCountMap[key].Count++
	}
	f.Write([]byte("\xEF\xBB\xBF"))
	w := csv.NewWriter(f)
	_ = w.Write([]string{"角色", "作品", "次数"})
	for _, two := range answerCountMap {
		_ = w.Write([]string{
			fmt.Sprintf("%v", two.Role),
			fmt.Sprintf("%v", two.Product),
			fmt.Sprintf("%v", two.Count)})
	}
	w.Flush()
}
