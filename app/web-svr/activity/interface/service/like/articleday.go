package like

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go-common/library/log"
	xtime "go-common/library/time"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/model/like"
)

const (
	_articleDayUID = "article_day"
	_twoDecimalFmt = "%.2f"
	_riskStatus    = 1
)

func (s *Service) articleDayInfo(c context.Context, mid int64) (res *like.ArticleDay, err error) {
	if res, err = s.dao.CacheArticleDay(c, mid); err != nil {
		err = nil
	}
	if res != nil {
		return
	}
	if res, err = s.dao.RawArticleDay(c, mid); err != nil {
		return
	}
	if res != nil && res.ID > 0 {
		s.cache.Do(c, func(c context.Context) {
			s.dao.AddCacheArticleDay(c, mid, res)
		})
	}
	return
}

func (s *Service) checkActArticle(mid int64) error {
	nowTime := time.Now().Unix()
	for _, whiteMid := range s.c.Rule.ActWhiteList {
		if mid == whiteMid {
			return nil
		}
	}
	if nowTime < s.c.ArticleDay.ApplyTime {
		return ecode.ActivityNotStart
	}
	// 活动结束判断
	if nowTime > s.c.ArticleDay.EndTime {
		return ecode.ActivityOverEnd
	}
	return nil
}

func (s *Service) JoinArticleDay(c context.Context, mid int64) (err error) {
	var (
		articleDay *like.ArticleDay
	)
	if err = s.checkActArticle(mid); err != nil {
		return
	}
	if articleDay, err = s.articleDayInfo(c, mid); err != nil {
		log.Errorc(c, "JoinArticleDay s.articleDayInfo mid(%d) error(%+v)", mid, err)
		return
	}
	// 重复报名.
	if articleDay != nil && articleDay.ID > 0 {
		err = ecode.ActivityArticleDayAlreadyErr
		return
	}
	if _, err = s.dao.JoinArticleDay(c, mid); err != nil {
		log.Errorc(c, "JoinArticleDay  s.dao.JoinArticleDay(%d) error(%+v)", mid, err)
	}
	return
}

func (s *Service) ArticleDayInfo(c context.Context, mid int64) (res *like.ArticleDayInfo, err error) {
	var (
		isJoin           bool
		joinTime         xtime.Time
		pubCount, status int64
		publish          []string
		articleDay       *like.ArticleDay
	)
	if articleDay, err = s.articleDayInfo(c, mid); err != nil {
		log.Errorc(c, "JoinArticleDay s.articleDayInfo mid(%d) error(%+v)", mid, err)
		return
	}
	if articleDay != nil && articleDay.ID > 0 {
		isJoin = true
		status = articleDay.Status
		joinTime = articleDay.Ctime
		pubCount = articleDay.PublishCount
		if status == _riskStatus {
			articleDay.Publish = ""
		}
		if articleDay.Publish != "" {
			publish = strings.Split(articleDay.Publish, ",")
		}
	}
	if len(publish) == 0 {
		publish = make([]string, 0)
	}
	rightInfo, e := s.dao.ArticleRightInfo(c)
	if e != nil {
		// 忽略错误.
		log.Errorc(c, "ArticleDayInfo s.dao.ArticleRightInfo error(%+v)", e)
	}
	if rightInfo == nil {
		rightInfo = &like.RightInfo{}
	}
	res = &like.ArticleDayInfo{
		IsJoin:    isJoin,
		Ctime:     joinTime,
		Status:    status,
		HaveMoney: s.publishMoney(pubCount),
		ActivityDay: &like.ActivityDay{
			ApplyTime:  s.c.ArticleDay.ApplyTime,
			BeginTime:  s.c.ArticleDay.BeginTime,
			EndTime:    s.c.ArticleDay.EndTime,
			ResultTime: s.c.ArticleDay.ResultTime,
		},
		ClockIn: publish,
		RightInfo: &like.RightInfo{
			DaysLater:       s.getActDays(),
			YesterdayPeople: rightInfo.YesterdayPeople,
			BeforePeople:    rightInfo.BeforePeople,
			BeforePublish:   s.getPubDays(),
		},
	}
	return
}

func (s *Service) publishMoney(pubCount int64) (res float64) {
	awardAll := s.loadArticleDayAwardMap()
	awardArticle, ok := awardAll[_articleDayUID]
	if !ok || pubCount == 0 {
		return
	}
	for _, award := range awardArticle {
		if pubCount >= award.ConditionMin {
			if award.SplitPeople == 0 {
				continue
			}
			res += float64(award.SplitMoney) / float64(award.SplitPeople)
		}
	}
	if res > 0 {
		res, _ = strconv.ParseFloat(fmt.Sprintf(_twoDecimalFmt, res), 64)
	}
	return
}

func (s *Service) storeArticleDayAwardMap(m map[string][]*like.ArticleDayAward) {
	s.articleDayAwardMap.Store(m)
}

func (s *Service) loadArticleDayAwardMap() map[string][]*like.ArticleDayAward {
	return s.articleDayAwardMap.Load().(map[string][]*like.ArticleDayAward)
}

func (s *Service) loadArticleDayAward() {
	ctx := context.Background()
	award, err := s.dao.RawAwards(ctx)
	if err != nil {
		log.Errorc(ctx, "loadArticleDayAward s.dao.RawAwards error(%+v)", err)
		return
	}
	if len(award) > 0 {
		s.storeArticleDayAwardMap(award)
	}
}

func (s *Service) getActDays() int64 {
	actDay := s.getPubDays()
	return actDay + 1
}

func (s *Service) getPubDays() int64 {
	now := time.Now()
	hours := now.Sub(time.Unix(s.c.ArticleDay.BeginTime, 0)).Hours()
	res := int64(hours / 24)
	return res
}
