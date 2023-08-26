package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go-common/library/log"
	"go-common/library/net/netutil"
	"go-common/library/retry"
	likemdl "go-gateway/app/web-svr/activity/job/model/like"
)

const (
	_articleDayUID         = "article_day"
	_riskArticleDayAction  = "activity_feb_article"
	_riskArticleHitOffline = "activity_feb_article_hit"
	_riskStatus            = 1
)

func (s *Service) checkRiskCommon(ctx context.Context, mid, wid, userAct int64) (res bool, err error) {
	otherEventCtx := &likemdl.ArticleDayEventCtx{
		Mid:         mid,
		ActivityUid: _riskArticleDayAction,
		ID:          wid,
		UserAction:  userAct,
		Sid:         s.c.Rule.ArticleDaySid,
	}
	if res, err = s.silverDao.RuleCheckCommon(ctx, _riskArticleDayAction, otherEventCtx); err != nil {
		log.Errorc(ctx, "DayClockIn checkRiskCommon mid(%d) otherEventCtx(%+v) error(%+v)", mid, otherEventCtx, err)
	}
	return
}

// 实时更新打卡数据.
func (s *Service) DayClockIn(ctx context.Context, i *likemdl.Item, opType int64) {
	var (
		publish  string
		err      error
		pubSlice []string
	)
	if i == nil || i.Mid == 0 {
		return
	}
	if time.Now().Unix() > s.c.Rule.ArticleDayEnd {
		return
	}
	// 判断sid
	if i.Sid != s.c.Rule.ArticleDaySid {
		return
	}
	// 上报风控
	if _, err = s.checkRiskCommon(ctx, i.Mid, i.Wid, opType); err != nil {
		log.Errorc(ctx, "DayClockIn s.checkRiskCommon likemdl(%+v) error(%+v)", i, err)
	}
	log.Infoc(ctx, "DayClockIn s.dao.RawArticlePublish sid(%d) likemdl(%+v) error(%+v)", i.Sid, i, err)
	if publish, err = s.dao.RawArticlePublish(ctx, i.Mid); err != nil {
		log.Errorc(ctx, "DayClockIn s.dao.RawArticlePublish likemdl(%+v) error(%+v)", i, err)
		return
	}
	article, artErr := s.arts(ctx, []int64{i.Wid}, 3)
	if artErr != nil || len(article) == 0 {
		log.Errorc(ctx, "DayClockIn count(%d) likemdl(%+v) error(%+v)", len(article), i, artErr)
		return
	}
	if isOpen(article[i.Wid].State) {
		createDate := article[i.Wid].Ctime.Time().Format("2006-01-02")
		// 判断开始投稿日期
		if article[i.Wid].Ctime.Time().Unix() < s.c.Rule.ArticleDayBegin {
			return
		}
		if publish != "" {
			pubSlice = strings.Split(publish, ",")
			for _, date := range pubSlice {
				// 已存在当天打卡
				if date == createDate {
					return
				}
			}
		}
		if publish == "" {
			publish = createDate
		} else {
			publish = publish + "," + createDate
		}
		if err = retry.WithAttempts(ctx, "upArticlePublish_clockIn_db", _retryTimes, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
			return s.dao.UpArticlePublish(ctx, publish, i.Mid)
		}); err != nil {
			log.Errorc(ctx, "DayClockIn s.dao.UpArticlePublish() mid(%d) publish(%s) error(%+v)", i.Mid, publish, err)
		}
		// 删除用户打卡缓存.
		if err = retry.WithAttempts(ctx, "upArticlePublish_clockIn_del_cache", _retryTimes, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
			return s.dao.DelCacheArticleDayByMid(ctx, i.Mid)
		}); err != nil {
			log.Errorc(ctx, "DayClockIn s.dao.DelCacheArticleDayByMid() mid(%d) publish(%s) error(%+v)", i.Mid, publish, err)
		}
	}
}

func isOpen(state int32) bool {
	switch state {
	case 0, 4, 5, 6, 7, 9, 13, 12, 14:
		return true
	default:
		return false
	}
}

// 每天3点结算.
func (s *Service) FinishArticleDay() {
	ctx := context.Background()
	var (
		likeID             int
		userArticle        = make(map[int64][]int64, 1000)
		beforePeople       int64
		yesterdayPeopleMap map[int64]int64
		midPublish         map[int64]int
		err                error
		riskMid            map[int64]struct{}
		allArticleMid      map[int64]bool
	)
	if err = retry.WithAttempts(ctx, "finishArticleDay_article_risk_db", _retryTimes, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
		riskMid, err = s.dao.SelArticleRiskMids(ctx)
		return err
	}); err != nil {
		log.Errorc(ctx, "FinishArticleDay s.dao.SelArticleRiskMids error(%+v)", err)
		return
	}
	allArticleMid = make(map[int64]bool)
	for {
		articleLikes, err := s.getArticleLike(ctx, likeID)
		if err != nil {
			log.Errorc(ctx, "FinishArticleDay s.dao.LikeItems sid(%s) offset(%d) error(%+v)", s.c.Rule.ArticleDaySid, likeID, err)
			break
		}
		if len(articleLikes) == 0 {
			log.Warn("FinishArticleDay userArticle success likeID(%d)", likeID)
			break
		}
		// 取最后一个id.
		likeID = int(articleLikes[len(articleLikes)-1].ID)
		for _, like := range articleLikes {
			if isHave, ok := allArticleMid[like.Mid]; ok {
				if !isHave && like.State == 1 {
					allArticleMid[like.Mid] = true
				}
			} else {
				allArticleMid[like.Mid] = like.State == 1
			}
			if like.State != 1 {
				continue
			}
			if _, ok := riskMid[like.Mid]; ok { // 过滤风控用户
				log.Warn("FinishArticleDay userArticle success like.Wid(%d) mid(%d)", like.Wid, like.Mid)
				continue
			}
			userArticle[like.Mid] = append(userArticle[like.Mid], like.Wid)
		}
	}
	if len(userArticle) == 0 {
		log.Warn("FinishArticleDay userArticle empty")
		return
	}
	pubDays := s.getPubDays()
	yesterdayPeopleMap = make(map[int64]int64, len(userArticle))
	midPublish = make(map[int64]int, len(userArticle))
	yesterdayDate := getYesterdayDate() //系统前一天日期格式转换
	for mid, wids := range userArticle {
		time.Sleep(100 * time.Millisecond)
		publishMap := make(map[string]string)
		articles, artErr := s.arts(ctx, wids, 3)
		if artErr != nil || len(articles) == 0 {
			log.Errorc(ctx, "FinishArticleDay  mid(%d) s.arts(%+v) error(%+v)", mid, wids, artErr)
			continue
		}
		for _, article := range articles {
			// 完成投稿.
			if isOpen(article.State) {
				// 判断开始投稿日期
				if article.Ctime.Time().Unix() < s.c.Rule.ArticleDayBegin || article.Ctime.Time().Unix() > s.c.Rule.ArticleDayEnd {
					continue
				}
				upDate := article.Ctime.Time().Format("2006-01-02")
				// 每天投稿.
				if _, ok := publishMap[upDate]; !ok {
					publishMap[upDate] = upDate
				}
				//昨日当天完成投稿人数.
				if yesterdayDate == upDate {
					if _, ok := yesterdayPeopleMap[mid]; !ok {
						yesterdayPeopleMap[mid] = mid
					}
				}
				// midPublish 奖励表不包括当天投的专栏
				if upDate == time.Now().Format("2006-01-02") {
					continue
				}
			}
		}
		var (
			dayArticles       []string
			yesterdayArticles []string
		)
		for _, date := range publishMap {
			dayArticles = append(dayArticles, date)
			if date == time.Now().Format("2006-01-02") {
				continue
			}
			yesterdayArticles = append(yesterdayArticles, date)
		}
		publishCount := len(yesterdayArticles)
		if pubDays > 0 && int64(publishCount) == pubDays {
			beforePeople++
		}
		s.updateUserInfo(ctx, mid, publishCount, dayArticles)
		// 用户有效专栏数.
		midPublish[mid] = publishCount
	}
	calcArticle := &likemdl.ArticleDay{
		YesterdayPeople: int64(len(yesterdayPeopleMap)), //昨日当天完成投专栏人数
		BeforePeople:    beforePeople,                   //截止到昨日完成全勤专栏人数
	}
	if err = retry.WithAttempts(ctx, "finishArticleDay_day_set_cache", _retryTimes, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
		return s.dao.SetArticleDay(ctx, calcArticle)
	}); err != nil {
		log.Errorc(ctx, "FinishArticleDay s.dao.SetArticleDay() error(%+v)", err)
	}
	// 更新award peoples.
	s.updateAwardPeople(ctx, midPublish)
	// 更新全部删除文章用户.
	s.updateDelAll(ctx, allArticleMid)
}

func (s *Service) updateUserInfo(ctx context.Context, mid int64, publishCount int, dayArticles []string) {
	if err := retry.WithAttempts(ctx, "finishArticleDay_publish_count_db", _retryTimes, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
		return s.dao.UpArticlePublishCount(ctx, strings.Join(dayArticles, ","), publishCount, mid)
	}); err != nil {
		log.Errorc(ctx, "FinishArticleDay s.dao.UpArticlePublishCount() mid(%d) error(%+v)", mid, err)
	}
	// 删除用户打卡缓存.
	if err := retry.WithAttempts(ctx, "finishArticleDay_article_delete_cache", _retryTimes, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
		return s.dao.DelCacheArticleDayByMid(ctx, mid)
	}); err != nil {
		log.Errorc(ctx, "FinishArticleDay s.dao.DelCacheArticleDayByMid() mid(%d) error(%+v)", mid, err)
	}
}

func (s *Service) updateDelAll(ctx context.Context, allArticleMid map[int64]bool) {
	for mid, isHave := range allArticleMid {
		if !isHave {
			s.updateUserInfo(ctx, mid, 0, []string{})
		}
	}
}

func (s *Service) getArticleLike(ctx context.Context, likeID int) (articleLikes []*likemdl.ObjItem, err error) {
	for i := 0; i < 3; i++ {
		if articleLikes, err = s.dao.LikeItems(ctx, s.c.Rule.ArticleDaySid, likeID, sidBatchNum); err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	return
}

func (s *Service) updateAwardPeople(ctx context.Context, midPublish map[int64]int) {
	var (
		ids []int64
		err error
	)
	awards := s.dao.SelArticleAward(ctx, _articleDayUID)
	for id := range awards {
		ids = append(ids, id)
	}
	yesterdayCalc := make(map[int64]int64, 5)
	for _, pubcount := range midPublish {
		for id, awardMin := range awards {
			if int64(pubcount) >= awardMin {
				yesterdayCalc[id]++
			}
		}
	}
	if err = retry.WithAttempts(ctx, "finishArticleDay_award_people_db", _retryTimes, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
		_, err = s.dao.UpAwardPeople(ctx, yesterdayCalc)
		return err
	}); err != nil {
		log.Errorc(ctx, "FinishArticleDay updateAwardPeople s.dao.UpAwardPeople() error(%+v)", err)
	}
}

func getYesterdayDate() string {
	curTime := time.Now()        //获取系统当前时间
	h := fmt.Sprintf("-%dh", 24) //减去24小时（前一天）
	dh, _ := time.ParseDuration(h)
	return curTime.Add(dh).Format("2006-01-02") //系统前一天日期格式转换
}

func (s *Service) getPubDays() int64 {
	now := time.Now()
	hours := now.Sub(time.Unix(s.c.Rule.ArticleDayBegin, 0)).Hours()
	return int64(hours / 24)
}

func (s *Service) ArticleDayRisk(ctx context.Context, mid, status int64) {
	m := &likemdl.RiskDecisionCtx{
		Mid:    mid,
		Status: status,
	}
	s.gaiaArticleRisk(m)
}

func (s *Service) gaiaArticleRisk(m *likemdl.RiskDecisionCtx) {
	var (
		ctx         = context.Background()
		err         error
		mid, status int64
	)
	if m == nil {
		log.Errorc(ctx, "gaiaRiskProc gaiaArticleRisk m is nil")
		return
	}
	if mid, err = InterfaceToInt64(m.Mid); err != nil {
		log.Errorc(ctx, "gaiaRiskProc gaiaArticleRisk Mid InterfaceToInt64 m(%+v)", m)
		return
	}
	if status, err = InterfaceToInt64(m.Status); err != nil {
		log.Errorc(ctx, "gaiaRiskProc gaiaArticleRisk Status InterfaceToInt64 m(%+v)", m)
		return
	}
	if status != _riskStatus {
		status = 0
	}
	log.Info("gaiaRiskProc gaiaArticleRisk  mid(%d)", mid)
	if mid == 0 {
		log.Errorc(ctx, "gaiaRiskProc gaiaArticleRisk risk RiskDecisionCtx empty mid(%d) status(%d)", mid, status)
		return
	}
	// 更新db
	if err = retry.WithAttempts(ctx, "gaiaArticleRisk_update_mid_db", _retryTimes, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
		return s.dao.UpArticleRisk(ctx, status, mid)
	}); err != nil {
		log.Errorc(ctx, "gaiaRiskProc gaiaArticleRisk s.dao.UpArticleRisk() mid(%d) status(%d) error(%+v)", mid, status, err)
		return
	}
	// 更新cache
	if err = retry.WithAttempts(ctx, "gaiaArticleRisk_update_mid_cache", _retryTimes, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
		return s.dao.DelCacheArticleDayByMid(ctx, mid)
	}); err != nil {
		log.Errorc(ctx, "gaiaRiskProc gaiaArticleRisk s.dao.DelCacheArticleDayByMid() mid(%d) status(%d) error(%+v)", mid, status, err)
		return
	}
}
