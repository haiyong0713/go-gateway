package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"go-common/library/log"
	"go-common/library/xstr"
	arcmdl "go-gateway/app/app-svr/archive/service/api"
	likemdl "go-gateway/app/web-svr/activity/job/model/like"
	"go-gateway/pkg/idsafe/bvid"

	bbqtaskapi "git.bilibili.co/bapis/bapis-go/bbq/task"
	actplatapi "git.bilibili.co/bapis/bapis-go/platform/interface/act-plat-v2"
	videoupapi "git.bilibili.co/bapis/bapis-go/videoup/open/service"
)

const (
	_tpExtend    = 1
	_tpBase      = 2
	_tpArchive1  = 3
	_tpArchive2  = 4
	_tpLight     = 5
	_tpBcut      = 6
	_tpWinSN     = 10
	_finishCount = 3
	_partSize    = 100
)

// ContributionSyncCounterFilter 数据同步计数过滤
func (s *Service) ContributionSyncCounterFilter() {
	c := context.Background()
	s.contributionSyncRunning.Lock()
	defer s.contributionSyncRunning.Unlock()
	aids, err := s.contributionGetAids(c)
	if err != nil {
		log.Errorc(c, "ContributionSyncCounterFilter s.contributionGetAids error(%v)", err)
		return
	}
	if len(aids) == 0 {
		log.Errorc(c, "ContributionSyncCounterFilter s.contributionGetAids aids count 0")
		return
	}
	if err = s.syncContributionAidsToActPlat(c, aids); err != nil {
		log.Errorc(c, "ContributionSyncCounterFilter s.syncContributionAidsToActPlat error(%v)", err)
		return
	}
	log.Infoc(c, "ContributionSyncCounterFilter success() aids count(%d)", len(aids))
	return
}

func (s *Service) contributionGetAids(c context.Context) ([]int64, error) {
	res, err := s.dao.SourceItem(context.Background(), s.c.S10Contribution.DayVid)
	if err != nil {
		log.Errorc(c, "Failed to load SourceItem(%d,%v)", s.c.S10Contribution.DayVid, err)
		return nil, err
	}
	tmp := new(likemdl.ArcListData)
	if err = json.Unmarshal(res, &tmp); err != nil {
		log.Errorc(c, "Failed to json unmarshal:%+v", err)
		return nil, err
	}
	aids := []int64{}
	if tmp != nil && tmp.List != nil {
		for _, v := range tmp.List {
			for _, val := range strings.Split(v.Data.Aids, ",") {
				if strings.HasPrefix(val, "BV") {
					avid, err := bvid.BvToAv(val)
					if err != nil {
						log.Errorc(c, "Failed to switch bv to av: %s %+v", val, err)
						continue
					}
					aids = append(aids, avid)
					continue
				}
				if avid, _ := strconv.ParseInt(val, 10, 64); avid > 0 {
					aids = append(aids, avid)
				}
			}
		}
	}
	log.Infoc(c, "contributionGetAids success")
	return aids, nil
}
func (s *Service) syncContributionAidsToActPlat(c context.Context, aids []int64) error {
	var values []*actplatapi.FilterMemberInt
	expireTime := int64(600)
	for _, i := range aids {
		values = append(values, &actplatapi.FilterMemberInt{Value: i, ExpireTime: expireTime})
	}
	_, err := s.actplatClient.AddFilterMemberInt(c, &actplatapi.SetFilterMemberIntReq{
		Activity: s.c.S10Contribution.ActPlatActivity,
		Counter:  s.c.S10Contribution.ActPlatCounter,
		Filter:   "filter_aid_sources",
		Values:   values,
	})
	return err
}

func (s *Service) actArchives(ctx context.Context) (archives []*arcmdl.Arc, midArcs map[int64][]int64, err error) {
	midArcs = make(map[int64][]int64)
	likeArcs, err := s.loadLikeList(ctx, s.c.S10Contribution.Sid, _retryTimes)
	if err != nil {
		log.Error("actArchives s.loadLikeList sid(%d) error(%v)", s.c.S10Contribution.Sid, err)
		return
	}
	var aids []int64
	for _, v := range likeArcs {
		if v != nil && v.Wid > 0 {
			aids = append(aids, v.Wid)
		}
	}
	aidsLen := len(aids)
	if aidsLen == 0 {
		log.Warn("actArchives len(aids) == 0")
		return
	}
	for i := 0; i < aidsLen; i += _aidBulkSize {
		time.Sleep(10 * time.Millisecond)
		var partAids []int64
		if i+_aidBulkSize > aidsLen {
			partAids = aids[i:]
		} else {
			partAids = aids[i : i+_aidBulkSize]
		}
		partArcs, err := s.arcClient.Arcs(ctx, &arcmdl.ArcsRequest{Aids: partAids})
		if err != nil {
			log.Error("actArchives s.arcClient.Arcs partAids(%v) error(%v)", partAids, err)
			continue
		}
		for _, v := range partArcs.GetArcs() {
			if v != nil && v.IsNormal() {
				archives = append(archives, v)
				// 用户稿件数
				midArcs[v.Author.Mid] = append(midArcs[v.Author.Mid], v.Aid)
			}
		}
	}
	return
}

func (s Service) userContributionInfo(ctx context.Context) (userInfo map[int64]*likemdl.ContributionUser, midContribution map[int64]*likemdl.ContributionUser, err error) {
	var contribution *likemdl.ContributionUser
	midContribution = make(map[int64]*likemdl.ContributionUser)
	archives, midArcs, err := s.actArchives(ctx)
	if err != nil {
		log.Errorc(ctx, "userContributionInfo s.actArchives error(%+v)", err)
		return
	}
	userInfo = make(map[int64]*likemdl.ContributionUser)
	for _, arc := range archives {
		if _, ok := userInfo[arc.Author.Mid]; ok {
			// 用户投稿数
			userInfo[arc.Author.Mid].UpArchives++
			// 用户视频点赞数
			userInfo[arc.Author.Mid].Likes += arc.Stat.Like
			// 用户点播放数
			userInfo[arc.Author.Mid].Views += arc.Stat.View
			if s.c.S10Contribution.IsWinSN == 1 && arc.PubDate.Time().Unix() > s.c.S10Contribution.SnArchiveTime {
				userInfo[arc.Author.Mid].SnUpArchives++
				userInfo[arc.Author.Mid].SnLikes += arc.Stat.Like
			}
			continue
		}
		userInfo[arc.Author.Mid] = &likemdl.ContributionUser{
			UpArchives: 1,
			Likes:      arc.Stat.Like,
			Views:      arc.Stat.View,
		}
		if s.c.S10Contribution.IsWinSN == 1 && arc.PubDate.Time().Unix() > s.c.S10Contribution.SnArchiveTime {
			userInfo[arc.Author.Mid].SnUpArchives = 1
			userInfo[arc.Author.Mid].SnLikes = arc.Stat.Like
		}
	}
	for mid, info := range userInfo {
		time.Sleep(10 * time.Millisecond)
		if contribution, err = s.dao.UserContribution(ctx, mid); err != nil {
			log.Errorc(ctx, "s.loadContribution mid(%d) error(%+v)", mid, err)
			continue
		}
		if contribution != nil && contribution.Mid > 0 {
			midContribution[mid] = contribution
		}
		// 计算轻视频
		if contribution != nil {
			info.LightVideos = contribution.LightVideos
		}
		if contribution == nil || contribution.Mid == 0 || contribution.LightVideos < _finishCount {
			ProgressRly, err := s.bbqtaskClient.ActivityUserProgress(ctx, &bbqtaskapi.ActivityUserProgressReq{Mid: uint64(mid)})
			if err != nil {
				log.Errorc(ctx, "s.bbqtaskClient.ActivityUserProgress(%d) error(%+v)", mid, err)
				continue
			}
			if ProgressRly != nil {
				info.LightVideos = ProgressRly.Progress
			}
		}
		// 计算bcut
		if contribution != nil {
			info.Bcuts = contribution.Bcuts
		}
		if contribution == nil || contribution.Mid == 0 || contribution.Bcuts < _finishCount {
			info.Bcuts = s.bcutCount(ctx, midArcs[mid])
		}
	}
	return
}

func (s *Service) CalcUserContribution() {
	ctx := context.Background()
	if s.isEnd() {
		log.Infoc(ctx, "CalcUserContribution  isEnd")
		return
	}
	userInfo, _, err := s.userContributionInfo(ctx)
	if err != nil {
		log.Errorc(ctx, "CalcUserContribution s.userContributionInfo error(%+v)", err)
		return
	}
	if len(userInfo) == 0 {
		log.Errorc(ctx, "CalcUserContribution s.userContributionInfo userinfo count 0")
		return
	}
	// 更新Redis用户征稿表
	s.UpUserContributionRedis(ctx, userInfo)
	// 更新奖励表
	s.UpUserAward(ctx, userInfo)
}

func (s *Service) UpUserContributionRedis(ctx context.Context, userInfo map[int64]*likemdl.ContributionUser) {
	var midSlice []*likemdl.ContributionUser
	for mid, info := range userInfo {
		info.Mid = mid
		midSlice = append(midSlice, info)
	}
	userCount := len(userInfo)
	if userCount == 0 {
		log.Warn("UpUserContributionRedis len(userInfo) == 0")
		return
	}
	for i := 0; i < userCount; i += _partSize {
		time.Sleep(10 * time.Millisecond)
		var partUser []*likemdl.ContributionUser
		if i+_partSize > userCount {
			partUser = midSlice[i:]
		} else {
			partUser = midSlice[i : i+_partSize]
		}
		if err := s.dao.AddCacheContributionUser(ctx, partUser); err != nil {
			log.Error("UpUserContributionRedis s.dao.AddCacheContributionUser() %+v", err)
		}
	}
	log.Info("UpUserContributionRedis success")
}

func (s *Service) UpUserAward(ctx context.Context, userInfo map[int64]*likemdl.ContributionUser) {
	var (
		ids           []int64
		currentViews  int32
		extendAwardID int64
	)
	awards, err := s.dao.SelContriAward(ctx)
	if err != nil {
		log.Errorc(ctx, "UpUserAward s.dao.SelContriAward error(%+v)", err)
		return
	}
	for _, award := range awards {
		ids = append(ids, award.ID)
	}
	SplitPepleCalc := make(map[int64]int64, 5)
	for _, info := range userInfo {
		currentViews += info.Views
		for _, award := range awards {
			switch award.AwardType {
			case _tpExtend:
				extendAwardID = award.ID
			case _tpBase:
				if info.UpArchives >= award.UpArchives && int64(info.Likes) >= award.Likes {
					SplitPepleCalc[award.ID]++
				}
			case _tpArchive1, _tpArchive2:
				if int64(info.Views) >= award.Views {
					SplitPepleCalc[award.ID]++
				}
			case _tpLight:
				if int64(info.LightVideos) >= award.LightVideos {
					SplitPepleCalc[award.ID]++
				}
			case _tpBcut:
				if int64(info.Bcuts) >= award.Bcuts {
					SplitPepleCalc[award.ID]++
				}
			case _tpWinSN:
				if s.c.S10Contribution.IsWinSN == 1 && info.SnUpArchives >= award.SnUpArchives && int64(info.SnLikes) >= award.SnLikes {
					SplitPepleCalc[award.ID]++
				}
			default:
			}
		}
	}
	// 更新当前总播放数.
	s.dao.UpTotalVeiwAward(ctx, extendAwardID, int64(currentViews))
	// 更新完成任务人数.
	s.dao.UpContriAwardPeople(ctx, SplitPepleCalc)
}

func (s *Service) isEnd() bool {
	nowTime := time.Now().Unix()
	if s.c.S10Contribution.IsWinSN == 1 {
		if nowTime > s.c.S10Contribution.WinSnEnd {
			return true
		}
	} else {
		if nowTime > s.c.S10Contribution.NoWinSnEnd {
			return true
		}
	}
	return false
}

func (s *Service) UpUserContributionDB() {
	var (
		ctx          = context.Background()
		contribution *likemdl.ActContributions
	)
	if s.isEnd() {
		log.Infoc(ctx, "UpUserContributionDB  isEnd")
		return
	}
	userInfo, _, err := s.userContributionInfo(ctx)
	if err != nil {
		log.Errorc(ctx, "CalcUserContribution s.userContributionInfo error(%+v)", err)
		return
	}
	if len(userInfo) == 0 {
		log.Errorc(ctx, "CalcUserContribution s.userContributionInfo userinfo count 0")
		return
	}
	for mid, info := range userInfo {
		time.Sleep(10 * time.Millisecond)
		if contribution, err = s.loadContribution(ctx, mid, _retry); err != nil {
			log.Errorc(ctx, "CalcUserContribution s.loadContribution mid(%d) error(%+v)", mid, err)
			continue
		}
		if contribution != nil && contribution.ID > 0 {
			// update
			s.dao.UpUserContribution(ctx, contribution.Mid, info)
			continue
		}
		// insert
		s.dao.AddUserContribution(ctx, mid, info)
	}
	log.Infoc(ctx, "UpUserContributionDB success userInfo count(%d)", len(userInfo))
}

func (s *Service) UpUserBcutLikes() {
	ctx := context.Background()
	_, midArcs, err := s.actArchives(ctx)
	if err != nil {
		log.Errorc(ctx, "UpUserBcutLikes s.actArchives error(%+v)", err)
		return
	}
	for mid, aids := range midArcs {
		var (
			bcutAids  []int64
			bcutLikes int32
		)
		for _, aid := range aids {
			time.Sleep(10 * time.Millisecond)
			arg := &videoupapi.ArcMaterialsReq{
				AID: aid,
				MTp: -1,
			}
			resRly, err := s.videoupClient.ArcMaterials(ctx, arg)
			if err != nil {
				log.Errorc(ctx, "bcutCount s.videoupClient.ArcMaterials aid(%d) error(%+v)", aid, err)
				continue
			}
			if isBcut(resRly.UpFrom) {
				bcutAids = append(bcutAids, aid)
			}
		}
		if len(bcutAids) > 0 {
			partArcs, err := s.arcClient.Arcs(ctx, &arcmdl.ArcsRequest{Aids: bcutAids})
			if err != nil {
				log.Error("actArchives s.arcClient.Arcs bcutAids(%v) error(%v)", bcutAids, err)
				continue
			}
			for _, v := range partArcs.GetArcs() {
				if v != nil && v.IsNormal() {
					bcutLikes += v.Stat.Like
				}
			}
			if bcutLikes > 0 {
				s.dao.UpUserBcutLikes(ctx, mid, bcutLikes)
			}
		}
	}
}

func (s *Service) bcutCount(ctx context.Context, midAids []int64) (res int32) {
	for _, aid := range midAids {
		time.Sleep(10 * time.Millisecond)
		if res == _finishCount {
			break
		}
		arg := &videoupapi.ArcMaterialsReq{
			AID: aid,
			MTp: -1,
		}
		resRly, err := s.videoupClient.ArcMaterials(ctx, arg)
		if err != nil {
			log.Errorc(ctx, "bcutCount s.videoupClient.ArcMaterials aid(%d) error(%+v)", aid, err)
			continue
		}
		if isBcut(resRly.UpFrom) {
			res++
		}
	}
	return res
}

func isBcut(upFrom int32) bool {
	switch upFrom {
	case 19, 20:
		return true
	default:
		return false
	}
}

func (s *Service) loadContribution(c context.Context, mid int64, retryCnt int) (res *likemdl.ActContributions, err error) {
	for i := 0; i < retryCnt; i++ {
		if res, err = s.dao.RawContribution(c, mid); err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	return
}

func (s Service) ScoreTotalRank() {
	var (
		arcScore []*likemdl.ArcScore
		topArcs  []int64
		ctx      = context.Background()
	)
	archives, _, err := s.actArchives(ctx)
	if err != nil {
		log.Error("ScoreTotalRank s.actArchives error(%+v)", err)
		return
	}
	for _, arc := range archives {
		arcScore = append(arcScore, &likemdl.ArcScore{
			Aid:   arc.Aid,
			Score: contriScore(arc),
		})
	}
	sort.Slice(arcScore, func(i int, j int) bool {
		return arcScore[i].Score > arcScore[j].Score
	})
	for _, topArc := range arcScore {
		if len(topArcs) == s.c.S10Contribution.RankCount {
			break
		}
		topArcs = append(topArcs, topArc.Aid)
	}
	s.dao.AddCacheTotalRank(ctx, s.c.S10Contribution.Sid, xstr.JoinInts(topArcs))
}

// contriScore 积分计算公式
func contriScore(arc *arcmdl.Arc) int64 {
	if arc.Stat.View == 0 {
		return 0
	}
	return contriPlayScore(arc) + contriQualityScore(arc) + contriTopicScore(arc)
}

// contriPlayScore 获取播放分数
func contriPlayScore(arc *arcmdl.Arc) int64 {
	videos := float64(arc.Videos)
	views := float64(arc.Stat.View)
	pRevise, _ := strconv.ParseFloat(fmt.Sprintf("%.3f", (4/(videos+3))), 64)
	aRevise, _ := strconv.ParseFloat(fmt.Sprintf("%.3f", ((300000+views)/(2*views))), 64)
	if aRevise > 1 {
		aRevise = 1
	}
	return int64(math.Floor(views*pRevise*aRevise + 0.5))
}

// contriQualityScore 获取质量分
func contriQualityScore(arc *arcmdl.Arc) int64 {
	like := float64(arc.Stat.Like)
	coin := float64(arc.Stat.Coin)
	fav := float64(arc.Stat.Fav)
	views := float64(arc.Stat.View)
	quality := like*5 + coin*10 + fav*20
	bRevise, _ := strconv.ParseFloat(fmt.Sprintf("%.3f", ((like*5+coin*10+fav*20)/(views+like*5+coin*10+fav*20))), 64)
	return int64(math.Floor(quality*bRevise + 0.5))
}

// contriTopicScore 获取讨论分
func contriTopicScore(arc *arcmdl.Arc) int64 {
	return int64((arc.Stat.Danmaku + arc.Stat.Reply)) * 15
}
