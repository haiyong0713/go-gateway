package bws

import (
	"context"
	"strconv"
	"sync"
	"time"

	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/pipeline/fanout"
	"go-gateway/app/web-svr/activity/ecode"
	aecode "go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/conf"
	"go-gateway/app/web-svr/activity/interface/dao/bws"
	lottery "go-gateway/app/web-svr/activity/interface/dao/lottery_v2"
	bwsmdl "go-gateway/app/web-svr/activity/interface/model/bws"
	suitapi "go-main/app/account/usersuit/service/api"

	accapi "git.bilibili.co/bapis/bapis-go/account/service"

	"go-common/library/sync/errgroup.v2"

	vipapi "git.bilibili.co/bapis/bapis-go/vip/service"
	"github.com/robfig/cron"
)

const (
	_accountBlocked = 1
	_noAward        = 0
	_awardAlready   = 2
)

var (
	_emptyUserAchieves = make([]*bwsmdl.UserAchieveDetail, 0)
)

// Service struct
type Service struct {
	c          *conf.Config
	dao        *bws.Dao
	accClient  accapi.AccountClient
	suitClient suitapi.UsersuitClient
	lotDao     lottery.Dao
	vipClient  vipapi.VipClient
	// bws all task ids
	bwsAllTasks  map[int64]*bwsmdl.Task
	bwsAllAwards map[int64]*bwsmdl.Award
	// bws admin mids
	allowMids   map[int64]struct{}
	awardMids   map[int64]struct{}
	lotteryMids map[int64]struct{}
	lotteryAids map[int64]struct{}
	achieveBids map[int64]struct{}
	bws2019Bids map[int64]struct{}
	bwsWhiteMid map[int64]struct{}
	cache       *fanout.Fanout
	// cron
	cron *cron.Cron
	// cache chan
	cacheCh chan func()
	// bws ups cache
	bluemCache     map[int64][]*bwsmdl.BluetoothUp
	blueUpMidCahce map[string]*bwsmdl.BluetoothUp
	blueUpKeyCahce map[string]*bwsmdl.BluetoothUp
}

// New Service
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:     c,
		dao:   bws.New(c),
		cache: fanout.New("cache", fanout.Worker(1), fanout.Buffer(1024)),
		// cache chan
		cacheCh: make(chan func(), 1024),
		// cron
		cron:   cron.New(),
		lotDao: lottery.New(c),
	}
	var err error
	if s.accClient, err = accapi.NewClient(c.AccClient); err != nil {
		panic(err)
	}
	if s.suitClient, err = suitapi.NewClient(c.SuitClient); err != nil {
		panic(err)
	}
	if s.vipClient, err = vipapi.NewClient(c.VipClient); err != nil {
		panic(err)
	}
	s.initMids()
	s.initLotteryAids()
	s.initAchieveBidMap()
	// cron
	s.initCron()
	s.cron.Start()
	return
}

func (s *Service) initMids() {
	tmpMids := make(map[int64]struct{}, len(s.c.Bws.AdminMids))
	tmpAward := make(map[int64]struct{}, len(s.c.Bws.AdminMids)+len(s.c.Bws.AwardMids))
	tmpLottery := make(map[int64]struct{}, len(s.c.Bws.AdminMids)+len(s.c.Bws.LotteryMids))
	tmpWhite := make(map[int64]struct{})
	for _, id := range s.c.Bws.AdminMids {
		tmpMids[id] = struct{}{}
		tmpAward[id] = struct{}{}
		tmpLottery[id] = struct{}{}
	}
	for _, id := range s.c.Bws.AwardMids {
		tmpAward[id] = struct{}{}
	}
	for _, id := range s.c.Bws.LotteryMids {
		tmpLottery[id] = struct{}{}
	}
	for _, v := range s.c.Bws.WhiteMid {
		tmpWhite[v] = struct{}{}
	}
	s.allowMids = tmpMids
	s.awardMids = tmpAward
	s.lotteryMids = tmpLottery
	s.bwsWhiteMid = tmpWhite
}

func (s *Service) initLotteryAids() {
	tmp := make(map[int64]struct{}, len(s.c.Bws.LotteryAids))
	for _, id := range s.c.Bws.LotteryAids {
		tmp[id] = struct{}{}
	}
	s.lotteryAids = tmp
}

func (s *Service) initAchieveBidMap() {
	tmp := make(map[int64]struct{}, len(s.c.Bws.InitAchieveBids))
	for _, bid := range s.c.Bws.InitAchieveBids {
		tmp[bid] = struct{}{}
	}
	s.achieveBids = tmp
	tmp = make(map[int64]struct{}, len(s.c.Bws.Bws2019))
	for _, bid := range s.c.Bws.Bws2019 {
		tmp[bid] = struct{}{}
	}
	s.bws2019Bids = tmp
}

// User user info.
func (s *Service) User(c context.Context, bid, mid int64, key string) (user *bwsmdl.User, err error) {
	var (
		hp, keyID                          int64
		ac                                 *accapi.CardReply
		points, dps, games, clockins, eggs []*bwsmdl.UserPointDetail
	)
	if key == "" {
		if key, err = s.midToKey(c, bid, mid); err != nil {
			return
		}
	} else {
		if mid, keyID, err = s.keyToMid(c, bid, key); err != nil {
			return
		}
	}
	user = new(bwsmdl.User)
	if mid != 0 {
		if ac, err = s.accCard(c, mid); err != nil {
			log.Error("User s.accCard(%d) error(%v)", mid, err)
			return
		}
	}
	if ac != nil && ac.Card != nil {
		user.User = &bwsmdl.UserInfo{
			Mid:  ac.Card.Mid,
			Name: ac.Card.Name,
			Face: ac.Card.Face,
			Key:  key,
		}
	} else {
		user.User = &bwsmdl.UserInfo{
			Name: strconv.FormatInt(keyID, 10),
			Key:  key,
		}
	}
	if points, err = s.userPoints(c, bid, key); err != nil {
		log.Error("User s.userPoints(%d,%s) error(%v)", bid, key, err)
		err = nil
	}
	user.Achievements = _emptyUserAchieves

	user.Items = make(map[string][]*bwsmdl.UserPointDetail, 4)
	gidMap := make(map[int64]int64, len(points))
	for _, v := range points {
		switch v.LockType {
		case bwsmdl.DpType:
			dps = append(dps, v)
		case bwsmdl.GameType:
			if v.Points == v.Unlocked {
				if _, ok := gidMap[v.Pid]; !ok {
					games = append(games, v)
				}
				gidMap[v.Pid] = v.Pid
			}
		case bwsmdl.ClockinType:
			clockins = append(clockins, v)
		case bwsmdl.EggType:
			eggs = append(eggs, v)
		}
		hp += v.Points
	}
	user.User.Hp = hp
	emp := make([]*bwsmdl.UserPointDetail, 0)
	if len(dps) == 0 {
		user.Items[bwsmdl.Dp] = emp
	} else {
		user.Items[bwsmdl.Dp] = dps
	}
	if len(games) == 0 {
		user.Items[bwsmdl.Game] = emp
	} else {
		user.Items[bwsmdl.Game] = games
	}
	if len(clockins) == 0 {
		user.Items[bwsmdl.Clockin] = emp
	} else {
		user.Items[bwsmdl.Clockin] = clockins
	}
	if len(eggs) == 0 {
		user.Items[bwsmdl.Egg] = emp
	} else {
		user.Items[bwsmdl.Egg] = eggs
	}
	return
}

// NewUser user info.
func (s *Service) NewUser(c context.Context, bid, mid int64, key string) (user *bwsmdl.User, err error) {
	var (
		hp, keyID        int64
		ac               *accapi.CardReply
		achErr, pointErr error
		items            map[string][]*bwsmdl.UserPointDetail
		mu               sync.Mutex
		achievePoints    map[string]int64
		achieveRank      int
		categoryAchieve  *bwsmdl.CategoryAchieve
		tempLocks        []int64
	)
	if key == "" {
		if key, err = s.midToKey(c, bid, mid); err != nil {
			return
		}
	} else {
		if mid, keyID, err = s.keyToMid(c, bid, key); err != nil {
			return
		}
	}
	user = new(bwsmdl.User)
	if mid != 0 {
		if ac, err = s.accCard(c, mid); err != nil {
			log.Error("User s.accCard(%d) error(%v)", mid, err)
			return
		}
	}
	if ac != nil && ac.Card != nil {
		user.User = &bwsmdl.UserInfo{
			Mid:  ac.Card.Mid,
			Name: ac.Card.Name,
			Face: ac.Card.Face,
			Key:  key,
		}
	} else {
		user.User = &bwsmdl.UserInfo{
			Name: strconv.FormatInt(keyID, 10),
			Key:  key,
		}
	}
	group := errgroup.WithContext(c)
	group.Go(func(errCtx context.Context) error {
		if categoryAchieve, achErr = s.userAchieves(errCtx, bid, key); achErr != nil {
			log.Error("User s.userAchieves(%d,%s) error(%v)", bid, key, achErr)
		}
		return nil
	})
	group.Go(func(errCtx context.Context) error {
		// 获取用户point分
		if hp, pointErr = s.dao.UserHp(errCtx, bid, key); pointErr != nil {
			log.Error("User s.dao.UserHp(%d,%s) error(%v)", bid, key, pointErr)
		}
		return nil
	})
	group.Go(func(errCtx context.Context) (e error) {
		// 获取用户成就点数
		if achievePoints, e = s.dao.AchievesPoint(errCtx, bid, []string{key}); e != nil {
			log.Error("s.dao.AchievesPoint(%d,%s) error(%v)", bid, key, e)
			e = nil
		}
		return
	})
	achieveRank = bwsmdl.DefaultRank
	if mid > 0 {
		group.Go(func(errCtx context.Context) (e error) {
			// 获取用户成就排行
			if achieveRank, e = s.dao.GetAchieveRank(errCtx, bid, mid, 0); e != nil {
				log.Error("s.dao.AchievesPoint(%d,%s) error(%v)", bid, key, e)
				e = nil
			}
			return
		})
	}
	// 获取用户已完成任务列表
	items = make(map[string][]*bwsmdl.UserPointDetail, len(_allPointsType))
	for k := range _allPointsType {
		tempLocks = append(tempLocks, k)
	}
	// 优化初始化是数据库查询等待超时
	group.Go(func(errCtx context.Context) (e error) {
		var (
			tempPoints map[int64][]*bwsmdl.UserPointDetail
		)
		if tempPoints, e = s.BatchUserLockPoints(errCtx, bid, tempLocks, key); e != nil {
			log.Error("s.userLockPoints(%d,%d,%s) error(%v)", bid, tempLocks, key, e)
			e = nil
			return
		}
		if tempPoints == nil {
			tempPoints = make(map[int64][]*bwsmdl.UserPointDetail)
		}
		for tempk, pVal := range tempPoints {
			gidMap := make(map[int64]int64)
			deal := make([]*bwsmdl.UserPointDetail, 0)
			for _, v := range pVal {
				switch tempk {
				case bwsmdl.GameType:
					if v.Points == v.Unlocked {
						if _, ok := gidMap[v.Pid]; !ok {
							deal = append(deal, v)
						}
						gidMap[v.Pid] = v.Pid
					}
				case bwsmdl.ChargeType, bwsmdl.SignType:
					if _, ok := gidMap[v.Pid]; !ok {
						deal = append(deal, v)
					}
					gidMap[v.Pid] = v.Pid
				default:
					deal = append(deal, v)
				}
			}
			if _, isOk := _allPointsType[tempk]; isOk {
				mu.Lock()
				items[_allPointsType[tempk]] = deal
				mu.Unlock()
			}
		}
		return
	})
	group.Wait()
	if categoryAchieve != nil {
		if len(categoryAchieve.Achievements) == 0 {
			user.Achievements = _emptyUserAchieves
		} else {
			user.Achievements = categoryAchieve.Achievements
		}
		if len(categoryAchieve.UnlockAchievements) == 0 {
			user.UnlockAchievements = _emptyUserAchieves
		} else {
			user.UnlockAchievements = categoryAchieve.UnlockAchievements
		}
	}
	if _, k := achievePoints[key]; k {
		user.User.AchievePoint = achievePoints[key]
	}
	if achieveRank != bwsmdl.DefaultRank {
		// 排名从0开始
		achieveRank += 1
	}
	user.User.AchieveRank = achieveRank
	user.Items = items
	user.User.Hp = hp
	return
}

func (s *Service) accCard(c context.Context, mid int64) (ac *accapi.CardReply, err error) {
	var (
		arg = &accapi.MidReq{Mid: mid}
	)
	if ac, err = s.accClient.Card3(c, arg); err != nil || ac == nil {
		log.Error("s.accRPC.Card3(%d) error(%v)", mid, err)
		err = ecode.ActivityServerTimeout
	} else if ac.Card.Silence == _accountBlocked {
		err = xecode.UserDisabled
	}
	return
}

// accCards .
func (s *Service) accCards(c context.Context, mids []int64) (ac map[int64]*accapi.Card, err error) {
	var (
		arg       = &accapi.MidsReq{Mids: mids}
		tempReply *accapi.CardsReply
	)
	if len(mids) == 0 {
		return
	}
	if tempReply, err = s.accClient.Cards3(c, arg); err != nil {
		log.Error("s.accRPC.Cards3(%d) error(%v)", mids, err)
		err = ecode.ActivityServerTimeout
		return
	}
	ac = make(map[int64]*accapi.Card)
	for k, v := range tempReply.Cards {
		if v.Silence == _accountBlocked {
			continue
		}
		ac[k] = v
	}
	return
}

// lastCompositeAchievements 异步回源使用.
func (s *Service) lastCompositeAchievements(c context.Context, mid int64) (achi *bwsmdl.Achievement, err error) {
	var (
		user map[int64]*bwsmdl.Users
	)
	bidUkey := make(map[int64]string)
	// 获取 bid和ukey对应关系
	if user, err = s.dao.RawUsersBids(c, s.c.Bws.Bws2019, mid); err != nil {
		log.Error("s.dao.RawUsersBids(%d) error(%v)", mid, err)
		return
	}
	for k, v := range user {
		bidUkey[k] = v.Key
	}
	return s.dao.LastAchievements(c, bidUkey)
}

// userAllAchieveRankLoad 总场次个人排行榜回源 异步回源使用.
func (s *Service) userAllAchieveRankLoad(c context.Context, bid, mid int64, ukey string) {
	var (
		achi        *bwsmdl.Achievement
		e           error
		allPointMap map[int64]int64
	)
	if isBws := s.dao.IsBws2019s(bid); isBws && mid > 0 {
		if allPointMap, e = s.dao.RawCompositeAchievesPoint(c, []int64{mid}); e != nil {
			log.Error("s.dao.CompositeAchievesPoint(%d,%d) error(%v)", bid, mid, e)
			return
		}
		if _, ok := allPointMap[mid]; !ok {
			return
		}
		// 获取最新的成就 直接查表
		if achi, e = s.lastCompositeAchievements(c, mid); e != nil {
			log.Error("s.dao.LastCompositeAchievements(%d) error(%v)", mid, e)
			return
		}
		if achi == nil || achi.ID == 0 {
			return
		}
		// 获取当前场次的score
		if e = s.dao.IncrAchievesPoint(c, bid, mid, allPointMap[mid], int64(achi.Ctime), true); e != nil {
			log.Error("s.dao.IncrAchievesPoint(%d,%s,%d) error(%v)", bid, ukey, mid, e)
		}
	}
}

// userAchieveRankLoad 单场次个人排行榜回源.
func (s *Service) userAchieveRankLoad(c context.Context, bid, mid int64, ukey string) {
	var (
		achi     *bwsmdl.Achievement
		pointMap map[string]int64
		e        error
	)
	if pointMap, e = s.dao.RawAchievesPoint(c, bid, []string{ukey}); e != nil {
		log.Error("s.dao.AchievesPoint(%d,%s) error(%v)", bid, ukey, e)
		return
	}
	if _, ok := pointMap[ukey]; !ok {
		return
	}
	// 获取最新的成就 直接查表
	if achi, e = s.dao.LastAchievements(c, map[int64]string{bid: ukey}); e != nil {
		log.Error("s.dao.LastAchievements(%d,%s) error(%v)", bid, ukey, e)
		return
	}
	if achi == nil || achi.ID == 0 {
		return
	}
	// 获取当前场次的score
	if e = s.dao.IncrSingleAchievesPoint(c, bid, mid, pointMap[ukey], int64(achi.Ctime), ukey, true); e != nil {
		log.Error("s.dao.IncrSingleAchievesPoint(%d,%s,%d) error(%v)", bid, ukey, mid, e)
	}
}

// bindAchieveRank 绑定user初始化总场成就 .
func (s *Service) bindAchieveRank(c context.Context, bid, mid int64) {
	var (
		err      error
		missData map[int64]int64
	)
	if isBws := s.dao.IsBws2019s(bid); isBws && mid > 0 {
		// 计算总场成就
		if missData, err = s.dao.RawCompositeAchievesPoint(c, []int64{mid}); err != nil {
			log.Error("bindAchieveRank s.dao.RawCompositeAchievesPoint(%d) error(%v)", mid, err)
			return
		}
		// 插入缓存
		if _, k := missData[mid]; k {
			s.dao.AddCacheCompositeAchievesPoint(c, missData)
		}
	}
}

// Binding binding by mid
func (s *Service) Binding(c context.Context, loginMid int64, p *bwsmdl.ParamBinding) (addAchieves []*bwsmdl.Achievement, err error) {
	var (
		achieves           *bwsmdl.Achievements
		users              *bwsmdl.Users
		checkMid           int64
		card               *accapi.CardReply
		isOk, vipCardState bool
		bidBind            map[int64]*bwsmdl.Users
	)
	// 防刷 redis无错误且写缓存没有成功 操作过于频繁
	if isOk, err = s.dao.RequestLimit(c, p.Bid, p.Key, "Binding", 1); err == nil && !isOk {
		err = aecode.ActivityFrequence
		return
	}
	group := errgroup.WithContext(c)
	group.Go(func(ctx context.Context) error {
		var e error
		if card, e = s.accCard(ctx, loginMid); e != nil {
			log.Errorc(c, "s.accCard(%d) error(%v)", loginMid, e)
		}
		return e
	})
	_, is2019Bws := s.bws2019Bids[p.Bid]
	if is2019Bws {
		group.Go(func(ctx context.Context) error {
			arg := &vipapi.CodesUseTimeReq{Mid: loginMid, BatchCodeIds: s.c.Bws.VipCardIDs}
			if vipCard, e := s.vipClient.CodesUseTimeList(ctx, arg); e != nil {
				log.Errorc(c, "Binding s.vipClient.CodesUseTimeList mid(%d) codeID(%v) error(%v)", loginMid, s.c.Bws.VipCardIDs, err)
			} else if vipCard != nil {
				for _, useTime := range vipCard.UseTimeList {
					if useTime.UseTimes >= s.c.Bws.VipCardStime && useTime.UseTimes <= s.c.Bws.VipCardEtime {
						vipCardState = true
						break
					}
				}
			}
			return nil
		})
	}
	if err = group.Wait(); err != nil {
		return
	}
	// special bid no achieve
	if p.Bid == s.c.Bws.SpecialBid {
		if users, err = s.dao.UsersKey(c, p.Bid, p.Key); err != nil {
			err = ecode.ActivityKeyFail
			return
		}
		if users != nil && users.Mid > 0 {
			if users.Mid != loginMid {
				err = ecode.ActivityKeyBindAlready
			}
			return
		}
		if users, err = s.dao.UsersMid(c, p.Bid, loginMid); err != nil {
			err = ecode.ActivityMidFail
			return
		}
		if users != nil && users.Key != "" {
			if users.Key != p.Key {
				err = ecode.ActivityMidBindAlready
			}
			return
		}
		if _, err = s.dao.CreateUser(c, p.Bid, loginMid, p.Key); err != nil {
			err = ecode.ActivityKeyFail
		}
		return
	}
	if checkMid, _, err = s.keyToMid(c, p.Bid, p.Key); err != nil {
		return
	}
	if checkMid != 0 {
		err = ecode.ActivityKeyBindAlready
		return
	}
	if users, err = s.dao.UsersMid(c, p.Bid, loginMid); err != nil {
		err = ecode.ActivityKeyFail
		return
	}
	if users != nil && users.Key != "" {
		err = ecode.ActivityMidBindAlready
		return
	}
	if err = s.dao.Binding(c, loginMid, p); err != nil {
		log.Errorc(c, "s.dao.Binding mid(%d) key(%s)  error(%v)", loginMid, p.Key, err)
		return
	}
	// 初始化总场socre，其余排行榜会在绑定成就时自动回源
	s.bindAchieveRank(c, p.Bid, loginMid)
	// 清理缓存前置
	s.dao.DelCacheUsersKey(c, p.Bid, p.Key)
	s.dao.DelCacheUsersMid(c, p.Bid, loginMid)
	if is2019Bws {
		s.cache.Do(c, func(c context.Context) {
			infos, e := s.dao.RawGradeInfo(c, p.Key)
			if e != nil {
				log.Errorc(c, "s.dao.RawGradeInfo(%s) error(%v)", p.Key, e)
				return
			}
			for _, v := range infos {
				s.dao.AddCacheUserGrade(c, v.Pid, map[int64]*bwsmdl.UserGrade{loginMid: {Amount: v.Amount, Mtime: v.Mtime}})
			}
		})
	}
	if _, active := s.achieveBids[p.Bid]; active {
		var (
			bindNewAchieve, bindOldAchieve, bindVipAchieve, bindGuangAchieve, bind2019Achieve *bwsmdl.Achievement
		)
		if achieves, err = s.dao.Achievements(c, p.Bid); err != nil || achieves == nil {
			log.Errorc(c, "s.dao.Achievements error(%v)", err)
			err = ecode.ActivityAchieveFail
			return
		}
		if len(achieves.Achievements) == 0 {
			err = ecode.ActivityNoAchieve
			return
		}
		for _, achieve := range achieves.Achievements {
			if achieve.LockType == bwsmdl.AchieveBindType {
				switch achieve.ExtraType {
				case bwsmdl.ExtraBindNew:
					bindNewAchieve = achieve
				case bwsmdl.ExtraBindOld:
					bindOldAchieve = achieve
				case bwsmdl.ExtraBindVip:
					bindVipAchieve = achieve
				case bwsmdl.ExtraBindGuang:
					bindGuangAchieve = achieve
				case bwsmdl.ExtraBind2019:
					bind2019Achieve = achieve
				}
			} else if achieve.LockType == bwsmdl.AchieveVipCard {
				if vipCardState {
					addAchieves = append(addAchieves, achieve)
				}
			}
		}
		if bindNewAchieve != nil || bindOldAchieve != nil || bindGuangAchieve != nil {
			if bidBind, err = s.dao.RawBidUsersMid(c, []int64{s.c.Bws.Bws2018Bid, s.c.Bws.Bws2019Guang, s.c.Bws.Bws2019Shang}, loginMid); err != nil {
				log.Errorc(c, "s.dao.RawBidUsersMid(%d,%d,%d) error(%v)", s.c.Bws.Bws2018Bid, s.c.Bws.Bws2019Guang, s.c.Bws.Bws2019Shang, err)
				err = nil // 错误降级处理
			}
		}
		if bindNewAchieve != nil && bindOldAchieve != nil {
			if bval, bok := bidBind[s.c.Bws.Bws2018Bid]; bok && bval != nil && bval.Key != "" {
				addAchieves = append(addAchieves, bindOldAchieve)
			} else {
				addAchieves = append(addAchieves, bindNewAchieve)
			}
		}
		if bindVipAchieve != nil {
			if card.Card.Vip.IsValid() {
				addAchieves = append(addAchieves, bindVipAchieve)
			}
		}
		//神行千里 && 云游万里
		if bindGuangAchieve != nil || bind2019Achieve != nil {
			var isGuang, isShange bool
			if bval, bok := bidBind[s.c.Bws.Bws2019Guang]; bok && bval != nil && bval.Key != "" {
				isGuang = true
			}
			if sVal, sok := bidBind[s.c.Bws.Bws2019Shang]; sok && sVal != nil && sVal.Key != "" {
				isShange = true
			}
			// 神行千里
			if bindGuangAchieve != nil && (isGuang || isShange) {
				addAchieves = append(addAchieves, bindGuangAchieve)
			}
			// 云游万里
			if bind2019Achieve != nil && isGuang && isShange {
				addAchieves = append(addAchieves, bind2019Achieve)
			}
		}
		if len(addAchieves) > 0 {
			for _, v := range addAchieves {
				tmpAchieve := v
				s.cache.Do(c, func(c context.Context) {
					s.addAchieve(c, loginMid, tmpAchieve, p.Key)
				})
			}
		}
	}
	return
}

func (s *Service) isAdmin(mid int64) bool {
	if _, ok := s.allowMids[mid]; ok {
		return true
	}
	return false
}

func (s *Service) midToKey(c context.Context, bid, mid int64) (key string, err error) {
	var users *bwsmdl.Users
	if users, err = s.dao.UsersMid(c, bid, mid); err != nil {
		err = ecode.ActivityKeyFail
		return
	}
	if users == nil || users.Key == "" {
		err = ecode.ActivityNotBind
		return
	}
	key = users.Key
	return
}

func (s *Service) keyToMid(c context.Context, bid int64, key string) (mid, keyID int64, err error) {
	var users *bwsmdl.Users
	if users, err = s.dao.UsersKey(c, bid, key); err != nil {
		err = ecode.ActivityKeyFail
		return
	}
	if users == nil || users.Key == "" {
		err = ecode.ActivityKeyNotExists
		return
	}
	if users.Mid > 0 {
		mid = users.Mid
	}
	keyID = users.ID
	return
}

func today() string {
	return time.Now().Format("20060102")
}

func (s *Service) initCron() {
	var err error
	s.loadBluetoothUpsCache()
	if err = s.cron.AddFunc(s.c.Cron.BwsBluetoothUps, s.loadBluetoothUpsCache); err != nil {
		panic(err)
	}
	s.loadAllTaskIDs()
	if err = s.cron.AddFunc(s.c.Cron.AllTask, s.loadAllTaskIDs); err != nil {
		panic(err)
	}
	s.loadAllAwards()
	if err = s.cron.AddFunc(s.c.Cron.AllAwards, s.loadAllAwards); err != nil {
		panic(err)
	}
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
