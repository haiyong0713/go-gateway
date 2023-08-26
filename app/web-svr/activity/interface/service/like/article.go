package like

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"go-common/library/sync/errgroup.v2"

	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/model/like"
)

const _giantExpire = 86400

// ArticleGiant get article giant data.
func (s *Service) ArticleGiant(c context.Context, mid int64) (res *like.ArticleGiant, err error) {
	var data *like.ArticleGiant
	if data, err = s.dao.ArticleGiant(c, mid); err != nil {
		log.Error("ArticleGiant s.dao.ArticleGiant(%d) error(%v)", mid, err)
		err = nil
		res = new(like.ArticleGiant)
		return
	}
	res = data
	return
}

// UpArticleLists .
func (s *Service) UpArticleLists(c context.Context, mid int64) (list []*like.ArticleList, err error) {
	if list, err = s.dao.UpArtLists(c, mid); err != nil {
		log.Error("UpArticleLists mid(%d) error(%v)", mid, err)
		list = make([]*like.ArticleList, 0)
	}
	return
}

// AddArtList .
func (s *Service) AddArtList(c context.Context, mid, listID int64) (err error) {
	var (
		join     int
		artInfos []*like.ArticleList
	)
	// 校验是否预约
	if join, err = s.LikeCheckJoin(c, mid, s.c.ArticleList.JoinSid); err != nil {
		return
	}
	if join == 0 {
		err = ecode.ActivityNotJoin
		return
	}
	// 校验文集信息
	if artInfos, err = s.dao.ArticleLists(c, []int64{listID}); err != nil {
		log.Error("s.dao.ArticleLists (%d) error(%v)", listID, err)
		err = xecode.NothingFound
		return
	}
	if len(artInfos) == 0 {
		err = xecode.NothingFound
		return
	}
	artInfo := artInfos[0]
	if artInfo.Mid != mid {
		err = ecode.ActivityLikeNotOwner
	}
	_, err = s.LikeAddText(c, &like.ParamText{Sid: s.c.ArticleList.ListSid, Wid: listID}, mid)
	return
}

func (s *Service) ArticleGiantV4List(c context.Context, mid int64) (res []*like.List, leftTimes int64, winTimes int, err error) {
	subject, err := s.dao.ActSubject(c, s.c.GiantV4.Sid)
	if err != nil {
		log.Error("ArticleGiantV4List sid:%d error(%v)", s.c.GiantV4.Sid, err)
		return
	}
	if subject == nil || subject.ID == 0 {
		err = ecode.ActivityNotExist
		return
	}
	now := time.Now()
	nowTs := now.Unix()
	if int64(subject.Stime) > nowTs {
		err = ecode.ActivityNotStart
		return
	}
	if int64(subject.Etime) < nowTs {
		err = ecode.ActivityOverEnd
		return
	}
	articles, err := s.dao.ArticleGiantV4(c, mid)
	if err != nil {
		log.Error("ArticleGiantV4List s.dao.ArticleGiantV4 mid:%d error(%v)", mid, err)
		return
	}
	var likeList []*like.List
	for _, v := range articles {
		item := func() *like.Item {
			for _, giantItem := range s.giantLikes {
				if v != nil && giantItem != nil && v.ID == giantItem.Wid {
					return giantItem
				}
			}
			return nil
		}()
		if item == nil {
			err = ecode.ActivityIDNotExists
			log.Error("ArticleGiantV4List articles(%+v) not found article:%+v", articles, v)
			s.cache.Do(c, func(ctx context.Context) {
				if resetErr := s.dao.ArticleGiantV4Reset(ctx, mid); resetErr != nil {
					log.Error("ArticleGiantV4List s.dao.ArticleGiantV4Reset mid:%d error(%v)", mid, resetErr)
					return
				}
			})
			return
		}
		likeList = append(likeList, &like.List{Item: item})
	}
	g := errgroup.WithContext(c)
	g.Go(func(ctx context.Context) error {
		return s.article(ctx, likeList, mid)
	})
	g.Go(func(ctx context.Context) error {
		var leftErr error
		leftTimes, leftErr = s.StoryKingLeftTime(ctx, s.c.GiantV4.Sid, mid)
		if leftErr != nil {
			log.Error("ArticleGiantV4List s.StoryKingLeftTime sid:%d mid:%d error(%v)", s.c.GiantV4.Sid, mid, leftErr)
		}
		return nil
	})
	g.Go(func(ctx context.Context) error {
		var winErr error
		winTimes, winErr = s.dao.RiGet(c, artWinKey(mid, s.c.GiantV4.Sid, now.Format("20060102")))
		if winErr != nil {
			log.Error("ArticleGiantV4List s.dao.RiGet sid:%d mid:%d error(%v)", s.c.GiantV4.Sid, mid, winErr)
		}
		return nil
	})
	if err = g.Wait(); err != nil {
		log.Error("ArticleGiantV4List s.articles mid:%d error(%v)", mid, err)
		return
	}
	res = likeList
	return
}

func (s *Service) GiantChoose(c context.Context, mid, lid int64) (scores map[int64]int64, err error) {
	likeItem, ok := s.giantLikes[lid]
	if !ok || likeItem == nil {
		err = ecode.ActivityIDNotExists
		return
	}
	articles, err := s.dao.ArticleGiantV4(c, mid)
	if err != nil {
		log.Error("GiantChoose s.dao.ArticleGiantV4 mid:%d error(%v)", mid, err)
		return
	}
	if _, ok := articles[likeItem.Wid]; !ok {
		err = ecode.ActivityIDNotExists
		return
	}
	var lids []int64
	for _, v := range articles {
		itemID := func() int64 {
			for _, giantItem := range s.giantLikes {
				if v != nil && giantItem != nil && v.ID == giantItem.Wid {
					return giantItem.ID
				}
			}
			return 0
		}()
		if itemID == 0 {
			log.Error("GiantChoose article(%+v) not found", v)
			err = ecode.ActivityIDNotExists
			return
		}
		lids = append(lids, itemID)
	}
	// reset giant list
	if err = s.dao.ArticleGiantV4Reset(c, mid); err != nil {
		log.Error("GiantChoose s.dao.ArticleGiantV4Reset mid:%d error(%v)", mid, err)
		return
	}
	if _, err = s.StoryKingAct(c, &like.ParamStoryKingAct{Sid: s.c.GiantV4.Sid, Lid: lid, Score: 1}, mid); err != nil {
		log.Error("GiantChoose s.LikeAct sid:%d lid:%d mid:%d error(%v)", s.c.GiantV4.Sid, lid, mid, err)
		return
	}
	if scores, err = s.dao.LikeActLidCounts(c, lids); err != nil {
		log.Error("GiantChoose s.dao.LikeActLidCounts(%v) error(%+v)", lids, err)
		return
	}
	// add lottery
	chooseMax := func() bool {
		lidScore := scores[lid]
		var maxScore int64
		for _, v := range scores {
			if v > maxScore {
				maxScore = v
			}
		}
		if lidScore == maxScore {
			return true
		}
		return false
	}()
	if chooseMax {
		now := time.Now()
		s.cache.Do(c, func(ctx context.Context) {
			s.dao.IncrWithExpire(ctx, artWinKey(mid, s.c.GiantV4.Sid, now.Format("20060102")), _giantExpire)
			lotterySid := s.c.GiantV4.LotterySid
			cid := s.c.GiantV4.Cid
			lottType := _other
			orderNo := strconv.FormatInt(mid, 10) + strconv.FormatInt(s.c.GiantV4.Sid, 10) + strconv.FormatInt(now.Unix(), 10)
			s.AddLotteryTimes(ctx, lotterySid, mid, cid, lottType, 0, orderNo, false)
		})
	}
	return
}

func (s *Service) cronGiantArticles() {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(100)
	time.Sleep(time.Second * time.Duration(rnd))
	s.loadGiantArticles()
}

func (s *Service) loadGiantArticles() {
	c := context.Background()
	likes, err := s.dao.LikeList(c, s.c.GiantV4.Sid)
	if err != nil {
		log.Error("loadGiantArticles s.dao.LikeList sid:%d error(%v)", s.c.GiantV4.Sid, err)
		return
	}
	if len(likes) == 0 {
		log.Warn("loadGiantArticles len likes == 0")
		return
	}
	tmp := make(map[int64]*like.Item, len(likes))
	for _, v := range likes {
		if v == nil {
			continue
		}
		tmp[v.ID] = v
	}
	s.giantLikes = tmp
}

func artWinKey(mid, sid int64, day string) string {
	return fmt.Sprintf("art_win_%d_%d_%s", mid, sid, day)
}
