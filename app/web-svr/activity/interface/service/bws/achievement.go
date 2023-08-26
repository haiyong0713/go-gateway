package bws

import (
	"context"
	"time"

	accapi "git.bilibili.co/bapis/bapis-go/account/service"
	"go-common/library/log"
	xtime "go-common/library/time"
	"go-gateway/app/web-svr/activity/ecode"
	xecode "go-gateway/app/web-svr/activity/ecode"
	bwsmdl "go-gateway/app/web-svr/activity/interface/model/bws"
	suitmdl "go-main/app/account/usersuit/service/api"

	"go-common/library/sync/errgroup.v2"
)

// Award  achievement award
func (s *Service) Award(c context.Context, loginMid int64, p *bwsmdl.ParamAward) (err error) {
	var (
		userAchieves []*bwsmdl.UserAchieve
		userAward    int64 = -1
	)
	if _, ok := s.awardMids[loginMid]; !ok {
		err = ecode.ActivityNotAwardAdmin
		return
	}
	if p.Key == "" {
		if p.Key, err = s.midToKey(c, p.Bid, p.Mid); err != nil {
			return
		}
	}
	if userAchieves, err = s.dao.UserAchieves(c, p.Bid, p.Key); err != nil {
		err = ecode.ActivityAchieveFail
		return
	}
	if len(userAchieves) == 0 {
		err = ecode.ActivityNoAchieve
		return
	}
	for _, v := range userAchieves {
		if v.Aid == p.Aid {
			userAward = v.Award
			break
		}
	}
	if userAward == -1 {
		err = ecode.ActivityNoAchieve
		return
	} else if userAward == _noAward {
		err = ecode.ActivityNoAward
		return
	} else if userAward == _awardAlready {
		err = ecode.ActivityAwardAlready
		return
	}
	if err = s.dao.Award(c, p.Key, p.Aid); err != nil {
		log.Error("s.dao.Award key(%s)  error(%v)", p.Key, err)
	}
	s.dao.DelCacheUserAchieves(c, p.Bid, p.Key)
	return
}

// Achievements achievements list
func (s *Service) Achievements(c context.Context, p *bwsmdl.ParamID) (rs *bwsmdl.Achievements, err error) {
	var mapCnt map[int64]int64
	if rs, err = s.dao.Achievements(c, p.Bid); err != nil || rs == nil || len(rs.Achievements) == 0 {
		log.Error("s.dao.Achievements error(%v)", err)
		err = ecode.ActivityAchieveFail
		return
	}
	if mapCnt, err = s.countAchieves(c, p.Bid, p.Day); err != nil || len(mapCnt) == 0 {
		err = nil
		return
	}
	for _, achieve := range rs.Achievements {
		achieve.UserCount = mapCnt[achieve.ID]
	}
	return
}

func (s *Service) countAchieves(c context.Context, bid int64, day string) (rs map[int64]int64, err error) {
	var countAchieves []*bwsmdl.CountAchieves
	if day == "" {
		day = today()
	}
	if countAchieves, err = s.dao.AchieveCounts(c, bid, day); err != nil {
		log.Error("s.dao.RawCountAchieves error(%v)", err)
		return
	}
	rs = make(map[int64]int64, len(countAchieves))
	for _, countAchieve := range countAchieves {
		rs[countAchieve.Aid] = countAchieve.Count
	}
	return
}

// Achievement Achievement
func (s *Service) Achievement(c context.Context, p *bwsmdl.ParamID) (rs *bwsmdl.Achievement, err error) {
	var (
		achieves *bwsmdl.Achievements
	)
	if achieves, err = s.dao.Achievements(c, p.Bid); err != nil || achieves == nil || len(achieves.Achievements) == 0 {
		log.Error("s.dao.Achievements error(%v)", err)
		err = ecode.ActivityAchieveFail
		return
	}
	for _, Achievement := range achieves.Achievements {
		if Achievement.ID == p.ID {
			rs = Achievement
			break
		}
	}
	if rs == nil {
		err = ecode.ActivityIDNotExists
	}
	return
}

// userAchieves .
func (s *Service) userAchieves(c context.Context, bid int64, key string) (res *bwsmdl.CategoryAchieve, err error) {
	var (
		usAchieves []*bwsmdl.UserAchieve
		achieves   *bwsmdl.Achievements
	)
	eg := errgroup.WithCancel(c)
	eg.Go(func(ctx context.Context) (e error) {
		if usAchieves, e = s.dao.UserAchieves(ctx, bid, key); e != nil {
			e = ecode.ActivityUserAchieveFail
		}
		return
	})
	eg.Go(func(ctx context.Context) (e error) {
		if achieves, e = s.dao.Achievements(ctx, bid); e != nil || achieves == nil || len(achieves.Achievements) == 0 {
			err = ecode.ActivityAchieveFail
		}
		return
	})
	if err = eg.Wait(); err != nil {
		return
	}
	res = &bwsmdl.CategoryAchieve{}
	achievesMap := make(map[int64]*bwsmdl.Achievement, len(achieves.Achievements))
	for _, v := range achieves.Achievements {
		achievesMap[v.ID] = v
	}
	userAchieveMap := make(map[int64]*bwsmdl.UserAchieve, len(usAchieves))
	for _, val := range usAchieves {
		userAchieveMap[val.Aid] = val
	}
	for _, achieve := range achieves.Achievements {
		detail := &bwsmdl.UserAchieveDetail{}
		detail.Name = achieve.Name
		detail.Icon = achieve.Icon
		detail.Dic = achieve.Dic
		detail.LockType = achieve.LockType
		detail.Unlock = achieve.Unlock
		detail.Bid = achieve.Bid
		detail.IconBig = achieve.IconBig
		detail.IconActive = achieve.IconActive
		detail.IconActiveBig = achieve.IconActiveBig
		detail.SuitID = achieve.SuitID
		detail.AchievePoint = achieve.AchievePoint
		detail.Level = achieve.Level
		if _, ok := userAchieveMap[achieve.ID]; ok {
			detail.UserAchieve = userAchieveMap[achieve.ID]
			res.Achievements = append(res.Achievements, detail)
		} else {
			res.UnlockAchievements = append(res.UnlockAchievements, detail)
		}
	}
	return
}

// AddAchieve special achieve add.
func (s *Service) AddAchieve(c context.Context, bid, mid int64) (err error) {
	var (
		achieves   *bwsmdl.Achievements
		key        string
		addAchieve *bwsmdl.Achievement
	)
	if key, err = s.midToKey(c, bid, mid); err != nil {
		log.Error("AddAchieve s.midToKey bid(%d) mid(%d) error(%v)", bid, mid, err)
		return
	}
	if achieves, err = s.dao.Achievements(c, bid); err != nil {
		log.Error("AddAchieve s.dao.Achievements(%d) error(%v)", bid, err)
		return
	}
	for _, v := range achieves.Achievements {
		if v.ID == s.c.Bws.VipCardAid {
			addAchieve = v
			break
		}
	}
	if addAchieve != nil {
		if err = s.addAchieve(c, mid, addAchieve, key); err != nil {
			log.Error("AddAchieve s.addAchieve(%d,%v,%s) error(%v)", mid, addAchieve, key, err)
		}
	} else {
		log.Error("AddAchieve no card achieve bid(%d) mid(%d)", bid, mid)
	}
	return
}

// addAchieve .
func (s *Service) addAchieve(c context.Context, mid int64, achieve *bwsmdl.Achievement, key string) (err error) {
	var uaID int64
	if uaID, err = s.dao.AddUserAchieve(c, achieve.Bid, achieve.ID, achieve.Award, key); err != nil {
		err = ecode.ActivityAddAchieveFail
		return
	}
	if err = s.dao.AppendUserAchievesCache(c, achieve.Bid, key, &bwsmdl.UserAchieve{ID: uaID, Aid: achieve.ID, Award: achieve.Award, Ctime: xtime.Time(time.Now().Unix())}); err != nil {
		// 删除缓存，自动回源
		s.cache.Do(c, func(c context.Context) {
			s.dao.DelCacheUserAchieves(c, achieve.Bid, key)
		})
		err = nil
	}
	// 单场成就排行
	if err = s.dao.IncrSingleAchievesPoint(c, achieve.Bid, mid, achieve.AchievePoint, time.Now().Unix(), key, false); err != nil {
		log.Error("s.dao.IncrSingleAchievesPoint bid(%d) mid(%d) point(%d) %s error(%v)", achieve.Bid, mid, achieve.AchievePoint, key, err)
		//  删除缓存，自动回源
		s.cache.Do(c, func(c context.Context) {
			s.userAchieveRankLoad(c, achieve.Bid, mid, key)
		})
		err = nil
	}
	// 总场成就排行缓存
	if err = s.dao.IncrAchievesPoint(c, achieve.Bid, mid, achieve.AchievePoint, time.Now().Unix(), false); err != nil {
		log.Error("s.dao.IncrAchievesPoint bid(%d) mid(%d) point(%d) error(%v)", achieve.Bid, mid, achieve.AchievePoint, err)
		//  删除缓存，自动回源
		s.cache.Do(c, func(c context.Context) {
			s.userAllAchieveRankLoad(c, achieve.Bid, mid, key)
		})
		err = nil
	}
	s.cache.Do(c, func(c context.Context) {
		s.dao.IncrCacheAchieveCounts(c, achieve.Bid, achieve.ID, today())
		var (
			keyID int64
			e     error
		)
		if mid == 0 {
			if mid, keyID, e = s.keyToMid(c, achieve.Bid, key); e != nil || mid == 0 {
				log.Warn("Lottery keyID(%d) key(%s) error(%v)", keyID, key, e)
			}
		}
		if mid > 0 {
			if achieve.SuitID > 0 {
				arg := &suitmdl.GrantByMidsReq{Mids: []int64{mid}, Pid: achieve.SuitID, Expire: s.c.Bws.SuitExpire}
				if _, e := s.suitClient.GrantByMids(c, arg); e != nil {
					log.Error("addAchieve s.suit.suitClient(%d,%d) error(%v)", mid, achieve.SuitID, e)
				}
				log.Warn("Suit mid(%d) suitID(%d)", mid, achieve.SuitID)
			}
			if _, ok := s.lotteryAids[achieve.ID]; ok {
				s.dao.AddLotteryMidCache(c, achieve.ID, mid)
			}
		}
	})
	return
}

func (s *Service) GradeEnter(c context.Context, bid, pid, mid, enterMid, amount int64, key string) error {
	if key == "" && enterMid == 0 {
		return xecode.ActivityKeyNotExists
	}
	// 校验owner、admin
	pointReply, err := s.dao.BwsPoints(c, []int64{pid})
	if err != nil {
		return err
	}
	point, ok := pointReply[pid]
	if !ok || pointReply[pid].Bid != bid {
		return xecode.ActivityIDNotExists
	}
	if point.Ower != mid && !s.isAdmin(mid) {
		return xecode.ActivityNotOwner
	}
	if enterMid != 0 {
		key, err = s.midToKey(c, bid, enterMid)
		if err != nil {
			return xecode.ActivityNotBind
		}
	}
	userMid := int64(0)
	if key != "" {
		userMid, _, err = s.keyToMid(c, bid, key)
		if err != nil {
			log.Error("keyToMid error(%v)", err)
			return xecode.ActivityKeyNotExists
		}
	}
	gradeMap := make(map[string]int64)
	// 获取用户考分
	if gradeMap, err = s.dao.AchievesGrade(c, pid, []string{key}); err != nil {
		log.Error("s.dao.AchievesGrade(%d %v) error(%v)", pid, key, err)
		return xecode.ActivityUserPointFail
	}
	score, ok := gradeMap[key]
	if !ok || score < amount {
		// db
		_, err := s.dao.AddUserGrade(c, pid, amount, key)
		if err != nil {
			log.Error("s.dao.AddUserGrade(%d,%d,%s) error(%v)", pid, amount, key, err)
			return err
		}
		gradeMap[key] = amount
		// cache
		if err = s.dao.AddCacheAchievesGrade(c, pid, gradeMap); err != nil {
			log.Error("s.dao.AddCacheAchievesGrade(%d,%v) error(%v)", pid, gradeMap, err)
			return err
		}
		if userMid > 0 {
			if err = s.dao.AddCacheUserGrade(c, pid, map[int64]*bwsmdl.UserGrade{userMid: {Amount: amount, Mtime: xtime.Time(time.Now().Unix())}}); err != nil {
				log.Error("s.dao.AddCacheUserGrade(%d,%d,%d) error(%v)", pid, userMid, amount, err)
				return err
			}
		}
	}
	return nil
}

func (s *Service) GradeShow(c context.Context, bid, pid, mid int64) (res *bwsmdl.AchieveRank, err error) {
	rank := 0
	score := float64(0)
	eg := errgroup.WithCancel(c)
	if mid > 0 {
		eg.Go(func(ctx context.Context) (err error) {
			if rank, score, err = s.dao.CacheUserGrade(ctx, pid, mid); err != nil {
				log.Error("GradeShow s.dao.CacheUserGrade(%d,%d) error(%v)", pid, mid, err)
				return
			}
			if rank > 4999 {
				rank = -1
			}
			return
		})
	}
	list := make([]*bwsmdl.RankUserGrade, 0)
	cards := make(map[int64]*accapi.Card)
	eg.Go(func(ctx context.Context) (err error) {
		if list, err = s.dao.CacheUsersRank(ctx, pid, 0, 110); err != nil {
			log.Error("GradeShow s.dao.CacheUsersRank error(%v)", err)
			return
		}
		mids := make([]int64, 0)
		for _, v := range list {
			if v.Mid <= 0 {
				continue
			}
			mids = append(mids, v.Mid)
		}
		if cards, err = s.accCards(c, mids); err != nil {
			log.Error("GradeShow s.accCards(%v) error(%v)", mids, err)
		}
		return
	})
	if err = eg.Wait(); err != nil {
		return
	}
	if rank != bwsmdl.DefaultRank && score > 0 {
		rank += 1
	}
	res = &bwsmdl.AchieveRank{SelfRank: rank, SelfPoint: int64(score)}
	i := 1
	for _, v := range list {
		if i > 100 {
			break
		}
		if _, k := cards[v.Mid]; !k {
			continue
		}
		tmp := &bwsmdl.UserInfo{
			Mid:          v.Mid,
			Name:         cards[v.Mid].Name,
			Face:         cards[v.Mid].Face,
			AchieveRank:  i,
			AchievePoint: int64(v.Amount),
		}
		res.List = append(res.List, tmp)
		i++
	}
	return
}

func (s *Service) GradeFix(c context.Context, bid, pid, mid int64) error {
	// 校验admin
	if !s.isAdmin(mid) {
		return xecode.ActivityNotAdmin
	}
	// 查db
	list, err := s.dao.RawUsersGrade(c, pid)
	if err != nil {
		log.Error("s.dao.RawUsersGrade(%d) error(%v)", pid, err)
		return xecode.ActivityAchieveFail
	}
	keys := make([]string, 0)
	sortedMap := make(map[int64]*bwsmdl.UserGrade)
	keyMap := make(map[string]int64)
	for _, value := range list {
		keys = append(keys, value.Key)   //转mid用
		keyMap[value.Key] = value.Amount //kv缓存用
	}
	// key转mid
	usersKey, err := s.dao.UsersKeys(c, bid, keys)
	if err != nil {
		log.Error("s.dao.UsersKeys(%d) error(%v)", bid, err)
		return err
	}
	keyToMid := make(map[string]int64)
	for _, value := range usersKey {
		if value == nil || value.Mid == 0 {
			continue
		}
		keyToMid[value.Key] = value.Mid
	}
	if err = s.dao.AddCacheAchievesGrade(c, pid, keyMap); err != nil {
		log.Error("s.dao.AddCacheAchievesGrade(%d) error(%v)", pid, err)
		return err
	}
	if err = s.dao.DelUserGrade(c, pid); err != nil {
		log.Error("s.dao.DelUserGrade error(%v)", err)
		return err
	}
	for _, v := range list {
		if keyToMid[v.Key] <= 0 {
			continue
		}
		sortedMap[keyToMid[v.Key]] = v
	}
	if err = s.dao.AddCacheUserGrade(c, pid, sortedMap); err != nil {
		log.Error("s.dao.AddCacheUserGrade(%d) error(%v)", pid, err)
		return err
	}
	return nil
}
