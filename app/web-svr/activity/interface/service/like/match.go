package like

import (
	"context"
	"encoding/json"
	"fmt"
	pb "go-gateway/app/web-svr/activity/interface/api"
	"strconv"
	"strings"
	"time"

	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/sync/errgroup"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/conf"
	match "go-gateway/app/web-svr/activity/interface/model/like"
	"go-main/app/account/usersuit/service/api"
	suitmdl "go-main/app/account/usersuit/service/api"

	coinmdl "git.bilibili.co/bapis/bapis-go/community/service/coin"

	errgroupV2 "go-common/library/sync/errgroup.v2"

	"github.com/pkg/errors"
)

const (
	_matchTable                 = "act_matchs"
	_objectTable                = "act_matchs_object"
	_userLogTable               = "act_match_user_log"
	_bwsPointTable              = "act_bws_points"
	_bwsAchieveTable            = "act_bws_achievements"
	_bwsPointLevelTable         = "act_bws_points_level"
	_bwsPointsAwardTable        = "act_bws_points_award"
	_bwsPointSignTable          = "act_bws_point_sign"
	_bwsOnlinePrint             = "bws_online_print"
	_bwsOnlineDress             = "bws_online_dress"
	_bwsOnlineAward             = "bws_online_award"
	_bwsOnlineAwardPackage      = "bws_online_award_package"
	_reason                     = "参与竞猜"
	_taskTable                  = "task"
	_actProtocolTable           = "act_subject_protocol"
	_actRuleTable               = "act_subject_rule"
	_quesDetailTable            = "question_detail"
	_actLotteryTable            = "act_lottery"
	_actLotteryGiftTable        = "act_lottery_gift"
	_actLotteryMemberGroupTable = "act_lottery_member_group"
	_actLotteryInfoTable        = "act_lottery_info"
	_actLotteryTimesTable       = "act_lottery_times"
	_actAwardTable              = "act_award_subject"
	_actUpTable                 = "act_up"
	_likesTable                 = "likes"
	_actSubjectTable            = "act_subject"
	_springCardsNumsTable       = "act_spring_cards_nums"
	_springArchiveNumsTable     = "act_spring_archive"
	_rankRuleTable              = "act_rank_rule"
	_rankRuleBatchTable         = "act_rank_rule_batch"
	_youthArchiveNumsTable      = "act_youth_cards_nums"
	_youthComposeUsedNumsTable  = "act_youth_compose_used"
	_upActReserveRelation       = "up_act_reserve_relation"
	_actSubjectCounterGroup     = "act_subject_counter_group"
	_actSubjectCounterNode      = "act_subject_counter_node"
	_cardsMidTable              = "act_cards_nums_"
)

var (
	_emptyMatch   = make([]*match.Match, 0)
	_emptyObjects = make([]*match.Object, 0)
	_emptyUserLog = make([]*match.UserLog, 0)
	_emptyFollow  = make([]string, 0)
)

// Match get match.
func (s *Service) Match(c context.Context, sid int64) (rs []*match.Match, err error) {
	// get from  cache.
	if rs, err = s.dao.ActMatchCache(c, sid); err != nil || len(rs) == 0 {
		if rs, err = s.dao.ActMatch(c, sid); err != nil {
			log.Error("s.dao.Match sid(%d)  error(%v)", sid, err)
			return
		}
		if len(rs) == 0 {
			rs = _emptyMatch
			return
		}
		s.cache.Do(c, func(c context.Context) {
			s.dao.SetActMatchCache(c, sid, rs)
		})
	}
	return
}

// AddGuess add match guess.
func (s *Service) AddGuess(c context.Context, mid int64, p *match.ParamAddGuess) (rs int64, err error) {
	var (
		object           *match.Object
		userGuess        []*match.UserLog
		group            *errgroup.Group
		coinErr, suitErr error
		countReply       *coinmdl.UserCoinsReply
		ip               = metadata.String(c, metadata.RemoteIP)
	)
	if p.Stake > conf.Conf.Rule.MaxGuessCoin {
		err = ecode.ActivityOverCoin
		return
	}
	//check mid coin count
	if countReply, err = s.coinClient.UserCoins(c, &coinmdl.UserCoinsReq{Mid: mid}); err != nil {
		log.Error("s.coinClient.UserCoins(%d) error(%v)", mid, err)
		return
	}
	if countReply.Count < float64(p.Stake) {
		err = ecode.ActivityNotEnoughCoin
		return
	}
	// get from  cache.
	if object, err = s.dao.ObjectCache(c, p.ObjID); err != nil || object == nil {
		if object, err = s.dao.Object(c, p.ObjID); err != nil {
			log.Error("s.dao.Match id(%d)  error(%v)", p.ObjID, err)
			return
		}
		if object == nil || object.ID == 0 {
			err = ecode.ActivityNotExist
			return
		}
		s.cache.Do(c, func(c context.Context) {
			s.dao.SetObjectCache(c, p.ObjID, object)
		})
	}
	if time.Now().Unix() < object.Stime.Time().Unix() {
		err = ecode.ActivityNotStart
		return
	} else if object.Result > 0 || time.Now().Unix() > object.Etime.Time().Unix() {
		err = ecode.ActivityOverEnd
		return
	}
	sid := object.Sid
	if userGuess, err = s.ListGuess(c, sid, mid); err != nil {
		log.Error("s.ListGuess(%d,%d) error(%v)", sid, mid, err)
		return
	}
	for _, userLog := range userGuess {
		if userLog.MOId == p.ObjID {
			err = ecode.ActivityHaveGuess
			return
		}
	}
	if rs, err = s.dao.AddGuess(c, mid, object.MatchId, p.ObjID, sid, p.Result, p.Stake); err != nil || rs == 0 {
		log.Error("s.dao.AddGuess matchID(%d) objectID(%d) sid(%d) error(%v)", object.MatchId, p.ObjID, sid, err)
		return
	}
	s.dao.DelUserLogCache(context.Background(), sid, mid)
	group, errCtx := errgroup.WithContext(c)
	if len(s.c.Rule.SuitPids) > 0 && len(userGuess)+1 == s.c.Rule.GuessCount {
		for _, v := range s.c.Rule.SuitPids {
			pid := v
			group.Go(func() error {
				mids := []int64{mid}
				if _, suitErr = s.suitClient.GrantByMids(errCtx, &suitmdl.GrantByMidsReq{Mids: mids, Pid: pid, Expire: s.c.Rule.SuitExpire}); suitErr != nil {
					log.Error("s.suitClient.GrantByMids(%d,%d,%s) error(%v)", mid, p.Stake, ip, suitErr)
				}
				return nil
			})
		}
	}
	group.Go(func() error {
		loseCoin := float64(-p.Stake)
		if _, coinErr = s.coinClient.ModifyCoins(errCtx, &coinmdl.ModifyCoinsReq{Mid: mid, Count: loseCoin, Reason: _reason, IP: ip}); coinErr != nil {
			log.Error("s.coinClient.ModifyCoin(%d,%d,%s) error(%v)", mid, p.Stake, ip, coinErr)
		}
		return nil
	})
	if s.c.Rule.MatchLotteryID > 0 {
		group.Go(func() error {
			if lotteryErr := s.dao.AddLotteryTimes(errCtx, s.c.Rule.MatchLotteryID, mid); lotteryErr != nil {
				log.Error("s.dao.AddLotteryTimes(%d,%d) error(%+v)", s.c.Rule.MatchLotteryID, mid, lotteryErr)
			}
			return nil
		})
	}
	group.Wait()
	return
}

// ListGuess get match guess list.
func (s *Service) ListGuess(c context.Context, sid, mid int64) (rs []*match.UserLog, err error) {
	// get from  cache.
	if rs, err = s.dao.UserLogCache(c, sid, mid); err != nil || len(rs) == 0 {
		if rs, err = s.dao.ListGuess(c, sid, mid); err != nil {
			log.Error("s.dao.ListGuess sid(%d) mid(%d)  error(%v)", sid, mid, err)
			return
		}
		if len(rs) == 0 {
			rs = _emptyUserLog
			return
		}
	}
	var (
		moIDs   []int64
		objects map[int64]*match.Object
	)
	for _, v := range rs {
		moIDs = append(moIDs, v.MOId)
	}
	if len(moIDs) == 0 {
		return
	}
	if objects, err = s.dao.MatchSubjects(c, moIDs); err == nil {
		for _, v := range rs {
			if obj, ok := objects[v.MOId]; ok {
				v.HomeName = obj.HomeName
				v.AwayName = obj.AwayName
				v.ObjResult = obj.Result
				v.GameStime = obj.GameStime
			}
		}
	}
	s.cache.Do(c, func(c context.Context) {
		s.dao.SetUserLogCache(c, sid, mid, rs)
	})
	return
}

// Guess user guess
func (s *Service) Guess(c context.Context, mid int64, p *match.ParamSid) (rs *match.UserGuess, err error) {
	var (
		userGuess           []*match.UserLog
		totalCont, winCount int64
	)
	if userGuess, err = s.ListGuess(c, p.Sid, mid); err != nil {
		log.Error("s.ListGuess(%d,%d) error(%v)", p.Sid, mid, err)
		return
	}
	for _, guess := range userGuess {
		if guess.ObjResult > 0 {
			if guess.Result == guess.ObjResult {
				winCount++
			}
			totalCont++
		}
	}
	rs = new(match.UserGuess)
	rs.Total = totalCont
	rs.Win = winCount
	return
}

// ClearCache del cache
func (s *Service) ClearCache(c context.Context, msg string) (err error) {
	var m struct {
		Table  string          `json:"table"`
		Action string          `json:"action"`
		New    json.RawMessage `json:"new,omitempty"`
	}
	if err = json.Unmarshal([]byte(msg), &m); err != nil {
		log.Error("ClearCache json.Unmarshal msg(%s) error(%v)", msg, err)
		return
	}
	log.Info("ClearCache json.Unmarshal msg(%s)", msg)
	switch m.Table {
	case _matchTable, _objectTable, _userLogTable:
		return s.clearMatchCache(c, msg)
	case _bwsPointTable:
		return s.clearBwsPoint(c, msg)
	case _bwsAchieveTable, _bwsPointLevelTable:
		return s.clearBwsCache(c, msg)
	case _bwsPointsAwardTable:
		return s.clearPointAward(c, msg)
	case _bwsPointSignTable:
		return s.clearPointSign(c, msg)
	case _taskTable:
		return s.clearTaskCache(c, msg)
	case _actProtocolTable:
		return s.clearActCache(c, msg)
	case _actRuleTable:
		return s.clearActRuleCache(c, msg)
	case _quesDetailTable:
		return s.clearQuesCache(c, msg)
	case _actLotteryGiftTable, _actLotteryInfoTable, _actLotteryTimesTable:
		return s.clearLotteryCache(c, msg)
	case _actLotteryTable:
		return s.clearLotteryActCache(c, msg)
	case _actUpTable:
		return s.clearUpCache(c, msg)
	case _actAwardTable:
		return s.clearActAwardCache(c, msg)
	case _likesTable:
		return s.clearLikesCache(c, msg)
	case _actLotteryMemberGroupTable:
		return s.clearLotteryMemberGroupCache(c, msg)
	case _bwsOnlineDress:
		return s.clearBwsOnlineDressCache(c, m.New)
	case _bwsOnlineAward:
		return s.clearBwsOnlineAwardCache(c, m.New)
	case _bwsOnlineAwardPackage:
		return s.clearBwsOnlineAwardPackageCache(c, m.New)
	case _bwsOnlinePrint:
		return s.clearBwsOnlinePrintCache(c, m.New)
	case _actSubjectTable:
		return s.clearActSubjectCache(c, m.New)
	case _springCardsNumsTable:
		return s.clearSpringFestivalCardsNumsCache(c, m.New)
	case _springArchiveNumsTable:
		return s.clearSpringFestivalMidArchiveNums(c, m.New)
	case _rankRuleTable:
		return s.clearRankRule(c, m.New)
	case _youthArchiveNumsTable, _youthComposeUsedNumsTable:
		return s.clearYouthCardsNumsCache(c, m.New)
	case _upActReserveRelation:
		if err = s.clearUpActReserveRelationCache(c, m.New); err != nil {
			return
		}
		if err = s.clearUpActReserveRelationReachCache(c, m.New); err != nil {
			return
		}
		if err = s.clearUpActReserveRelation4LiveCache(c, m.New); err != nil {
			return
		}
		return
	case _actSubjectCounterGroup:
		return s.clearActSubjectCounterGroupCache(c, m.New, m.Action)
	case _actSubjectCounterNode:
		return s.clearActSubjectCounterNodeCache(c, m.New)
	}
	if strings.HasPrefix(m.Table, _cardsMidTable) {

		return s.clearCardsNumsCache(c, m.New)
	}
	return
}

func (s *Service) clearActSubjectCounterGroupCache(c context.Context, msg json.RawMessage, action string) (err error) {
	var m struct {
		ID  int64 `json:"id"`
		SID int64 `json:"sid"`
	}
	if err = json.Unmarshal(msg, &m); err != nil {
		log.Errorc(c, "clearActSubjectCounterGroupCache json.Unmarshal msg(%s) error(%v)", msg, err)
		return
	}
	log.Info("clearActSubjectCounterGroupCache json.Unmarshal msg(%s)", msg)
	wg := errgroupV2.WithContext(c)
	wg.Go(func(ctx context.Context) (err error) {
		if err = s.dao.DelCacheGetReserveCounterGroupInfoByGid(c, m.ID); err != nil {
			log.Errorc(c, "clearActSubjectCounterGroupCache s.dao.DelCacheGetReserveCounterGroupInfoByGid(c, %d) error(%v)", m.ID, err)
		}
		return
	})
	if action == "insert" {
		wg.Go(func(ctx context.Context) (err error) {
			if err = s.dao.DelCacheGetReserveCounterGroupIDBySid(c, m.SID); err != nil {
				log.Errorc(c, "clearActSubjectCounterGroupCache s.dao.DelCacheGetReserveCounterGroupIDBySid(c, %d) error(%v)", m.SID, err)
			}
			return
		})
	}
	err = wg.Wait()
	return
}

func (s *Service) clearActSubjectCounterNodeCache(c context.Context, msg json.RawMessage) (err error) {
	var m struct {
		ID  int64 `json:"id"`
		GID int64 `json:"group_id"`
	}
	if err = json.Unmarshal(msg, &m); err != nil {
		log.Errorc(c, "clearActSubjectCounterNodeCache json.Unmarshal msg(%s) error(%v)", msg, err)
		return
	}
	log.Info("clearActSubjectCounterNodeCache json.Unmarshal msg(%s)", msg)
	if err = s.dao.DelCacheGetReserveCounterNodeByGid(c, m.GID); err != nil {
		log.Errorc(c, "clearActSubjectCounterNodeCache s.dao.DelCacheGetReserveCounterNodeByGid(c, %d) error(%v)", m.GID, err)
	}
	return
}

func (s *Service) clearSpringFestivalCardsNumsCache(c context.Context, msg json.RawMessage) (err error) {
	var m struct {
		MID int64 `json:"mid"`
	}
	if err = json.Unmarshal(msg, &m); err != nil {
		log.Errorc(c, "clearSpringFestivalCardsNumsCache json.Unmarshal msg(%s) error(%v)", msg, err)
		return
	}
	log.Info("clearSpringFestivalCardsNumsCache json.Unmarshal msg(%s)", msg)
	if err = s.springfestival2021Dao.DeleteMidCardDetail(c, m.MID); err != nil {
		log.Errorc(c, "clearSpringFestivalCardsNumsCache s.springfestival2021Dao.DeleteMidCardDetail(c, %d) error(%v)", m.MID, err)
	}
	return
}

func (s *Service) clearYouthCardsNumsCache(c context.Context, msg json.RawMessage) (err error) {
	var m struct {
		MID int64 `json:"mid"`
	}
	if err = json.Unmarshal(msg, &m); err != nil {
		log.Errorc(c, "clearYouthCardsNumsCache json.Unmarshal msg(%s) error(%v)", msg, err)
		return
	}
	log.Info("clearYouthCardsNumsCache json.Unmarshal msg(%s)", msg)
	if err = s.cardsDao.DeleteMidCardDetail(c, m.MID, s.c.Cards.Activity); err != nil {
		log.Errorc(c, "clearYouthCardsNumsCache s.cards.DeleteMidCardDetail(c, %d) error(%v)", m.MID, err)
	}
	return
}

func (s *Service) clearCardsNumsCache(c context.Context, msg json.RawMessage) (err error) {
	var m struct {
		MID        int64 `json:"mid"`
		ActivityID int64 `json:"activity_id"`
	}
	if err = json.Unmarshal(msg, &m); err != nil {
		log.Errorc(c, "clearYouthCardsNumsCache json.Unmarshal msg(%s) error(%v)", msg, err)
		return
	}
	log.Info("clearYouthCardsNumsCache json.Unmarshal msg(%s)", msg)
	if err = s.cardsDao.DeleteMidCardDetailNew(c, m.MID, m.ActivityID); err != nil {
		log.Errorc(c, "clearCardsNumsCache s.cards.DeleteMidCardDetailNew(c, %d,%d) error(%v)", m.MID, m.ActivityID, err)
	}
	return
}
func (s *Service) clearSpringFestivalMidArchiveNums(c context.Context, msg json.RawMessage) (err error) {
	var m struct {
		MID int64 `json:"mid"`
	}
	if err = json.Unmarshal(msg, &m); err != nil {
		log.Errorc(c, "clearSpringFestivalMidArchiveNums json.Unmarshal msg(%s) error(%v)", msg, err)
		return
	}
	log.Info("clearSpringFestivalMidArchiveNums json.Unmarshal msg(%s)", msg)
	if err = s.springfestival2021Dao.DeleteArchiveNums(c, m.MID); err != nil {
		log.Errorc(c, "clearSpringFestivalMidArchiveNums s.springfestival2021Dao.DeleteArchiveNums(c, %d) error(%v)", m.MID, err)
	}
	return
}

func (s *Service) clearRankRule(c context.Context, msg json.RawMessage) (err error) {
	var m struct {
		ID int64 `json:"id"`
	}
	if err = json.Unmarshal(msg, &m); err != nil {
		log.Errorc(c, "clearRankRuleBatch json.Unmarshal msg(%s) error(%v)", msg, err)
		return
	}
	log.Infoc(c, "clearRankRuleBatch json.Unmarshal msg(%s)", msg)
	if err = s.rankv3Dao.DeleteRankRule(c, m.ID); err != nil {
		log.Errorc(c, "clearRankRuleBatch s.rankv3Dao.DeleteRankRule(c, %d) error(%v)", m.ID, err)
	}
	return
}
func (s *Service) clearRankRuleBatch(c context.Context, msg json.RawMessage) (err error) {
	var m struct {
		RuleID int64 `json:"rule_id"`
	}
	if err = json.Unmarshal(msg, &m); err != nil {
		log.Errorc(c, "clearRankRuleBatch json.Unmarshal msg(%s) error(%v)", msg, err)
		return
	}
	log.Infoc(c, "clearRankRuleBatch json.Unmarshal msg(%s)", msg)
	if err = s.rankv3Dao.DeleteRankRule(c, m.RuleID); err != nil {
		log.Errorc(c, "clearRankRuleBatch s.rankv3Dao.DeleteRankRule(c, %d) error(%v)", m.RuleID, err)
	}
	return
}

func (s *Service) clearBwsOnlineDressCache(c context.Context, msg json.RawMessage) (err error) {
	var m struct {
		ID int64 `json:"id"`
	}
	if err = json.Unmarshal(msg, &m); err != nil {
		log.Errorc(c, "clearBwsOnlineDressCache json.Unmarshal msg(%s) error(%v)", msg, err)
		return
	}
	log.Info("clearBwsOnlineDressCache json.Unmarshal msg(%s)", msg)
	if err = s.bwsOnlineDao.DelCacheDress(c, m.ID); err != nil {
		log.Errorc(c, "clearBwsOnlineDressCache s.bwsOnlineDao.DelCacheDress(c, %d) error(%v)", m.ID, err)
	}
	return
}

func (s *Service) clearBwsOnlinePrintCache(c context.Context, msg json.RawMessage) (err error) {
	var m struct {
		ID int64 `json:"id"`
	}
	if err = json.Unmarshal(msg, &m); err != nil {
		log.Errorc(c, "clearBwsOnlinePrintCache json.Unmarshal msg(%s) error(%v)", msg, err)
		return
	}
	log.Info("clearBwsOnlinePrintCache json.Unmarshal msg(%s)", msg)
	if err = s.bwsOnlineDao.DelCachePrint(c, m.ID); err != nil {
		log.Errorc(c, "clearBwsOnlinePrintCache s.bwsOnlineDao.DelCachePrint(c, %d) error(%v)", m.ID, err)
	}
	return
}

func (s *Service) clearBwsOnlineAwardCache(c context.Context, msg json.RawMessage) (err error) {
	var m struct {
		ID int64 `json:"id"`
	}
	if err = json.Unmarshal(msg, &m); err != nil {
		log.Errorc(c, "clearBwsOnlineAwardCache json.Unmarshal msg(%s) error(%v)", msg, err)
		return
	}
	log.Info("clearBwsOnlineAwardCache json.Unmarshal msg(%s)", msg)
	if err = s.bwsOnlineDao.DelCacheAward(c, m.ID); err != nil {
		log.Errorc(c, "clearBwsOnlineAwardCache s.bwsOnlineDao.DelCacheAward(c, %d) error(%v)", m.ID, err)
	}
	return
}

func (s *Service) clearBwsOnlineAwardPackageCache(c context.Context, msg json.RawMessage) (err error) {
	var m struct {
		ID int64 `json:"id"`
	}
	if err = json.Unmarshal(msg, &m); err != nil {
		log.Errorc(c, "clearBwsOnlineAwardPackageCache json.Unmarshal msg(%s) error(%v)", msg, err)
		return
	}
	log.Info("clearBwsOnlineAwardPackageCache json.Unmarshal msg(%s)", msg)
	if err = s.bwsOnlineDao.DelCacheAwardPackage(c, m.ID); err != nil {
		log.Errorc(c, "clearBwsOnlineAwardPackageCache s.bwsOnlineDao.DelCacheAwardPackage(c, %d) error(%v)", m.ID, err)
	}
	return
}

// clearActCache .
func (s *Service) clearActCache(c context.Context, msg string) (err error) {
	var m struct {
		Table string `json:"table"`
		New   struct {
			Sid int64 `json:"sid"`
		} `json:"new,omitempty"`
	}
	if err = json.Unmarshal([]byte(msg), &m); err != nil {
		log.Error("ClearCache json.Unmarshal msg(%s) error(%v)", msg, err)
		return
	}
	log.Info("clearActCache json.Unmarshal msg(%s)", msg)
	if err = s.dao.DelCacheActSubjectProtocol(c, m.New.Sid); err != nil {
		log.Error("clearActCache s.dao.DelCacheActSubjectProtocol(%d) error(%v)", m.New.Sid, err)
	}
	return
}

// clearActRuleCache .
func (s *Service) clearActRuleCache(c context.Context, msg string) (err error) {
	var m struct {
		Table string `json:"table"`
		New   struct {
			Sid int64 `json:"sid"`
			ID  int64 `json:"id"`
		} `json:"new,omitempty"`
	}
	if err = json.Unmarshal([]byte(msg), &m); err != nil {
		log.Error("clearActRuleCache json.Unmarshal msg(%s) error(%v)", msg, err)
		return
	}
	log.Info("clearActRuleCache json.Unmarshal msg(%s)", msg)
	if err = s.dao.DelCacheSubjectRulesBySid(c, m.New.Sid); err != nil {
		log.Error("clearActRuleCache s.dao.DelCacheSubjectRulesBySid(%d) error(%v)", m.New.Sid, err)
	}
	return
}

func (s *Service) clearMatchCache(c context.Context, msg string) (err error) {
	var m struct {
		Table string `json:"table"`
		New   struct {
			ID    int64 `json:"id"`
			Sid   int64 `json:"sid"`
			MatID int64 `json:"match_id"`
			Mid   int64 `json:"mid"`
			MOId  int64 `json:"m_o_id"`
		} `json:"new,omitempty"`
	}
	if err = json.Unmarshal([]byte(msg), &m); err != nil {
		log.Error("ClearCache json.Unmarshal msg(%s) error(%v)", msg, err)
		return
	}
	log.Info("ClearCache json.Unmarshal msg(%s)", msg)
	if m.Table == _matchTable {
		if err = s.dao.DelActMatchCache(c, m.New.Sid, m.New.ID); err != nil {
			log.Error("s.dao.DelActMatchCache sid(%d) matchID(%d)  error(%v)", m.New.Sid, m.New.ID, err)
		}
	} else if m.Table == _objectTable {
		if err = s.dao.DelObjectCache(c, m.New.ID, m.New.Sid); err != nil {
			log.Error("s.dao.DelObjectCache objID(%d)  Sid(%d)  error(%v)", m.New.ID, m.New.Sid, err)
		}
	} else if m.Table == _userLogTable {
		if err = s.dao.DelUserLogCache(c, m.New.Sid, m.New.Mid); err != nil {
			log.Error("s.dao.DelUserLogCache mid(%d) error(%v)", m.New.Mid, err)
		}
	}
	return
}

func (s *Service) clearBwsCache(c context.Context, msg string) (err error) {
	var m struct {
		Table string `json:"table"`
		New   struct {
			ID  int64 `json:"id"`
			Bid int64 `json:"bid"`
		} `json:"new,omitempty"`
	}
	if err = json.Unmarshal([]byte(msg), &m); err != nil {
		log.Error("ClearCache json.Unmarshal msg(%s) error(%v)", msg, err)
		return
	}
	switch m.Table {
	case _bwsAchieveTable:
		err = s.bwsDao.DelCacheAchievements(c, m.New.Bid)
	case _bwsPointLevelTable:
		err = s.clearPointLevel(c, m.New.ID, m.New.Bid)
	}
	if err == nil {
		log.Info("ClearCache success msg(%d)", m.New.Bid)
	}
	return
}

// clearPointAward .
func (s *Service) clearPointAward(c context.Context, msg string) (err error) {
	var m struct {
		Table string `json:"table"`
		New   struct {
			ID   int64 `json:"id"`
			PlID int64 `json:"pl_id"`
		} `json:"new,omitempty"`
	}
	if err = json.Unmarshal([]byte(msg), &m); err != nil {
		log.Error("clearPointAward json.Unmarshal msg(%s) error(%v)", msg, err)
		return
	}
	eg := errgroupV2.WithContext(c)
	eg.Go(func(ctx context.Context) (e error) {
		if e = s.bwsDao.DelCachePointsAward(ctx, m.New.PlID); e != nil {
			log.Info("clearPointAward s.bwsDao.DelCachePointsAward %d error(%v)", m.New.PlID, e)
		}
		return
	})
	eg.Go(func(ctx context.Context) (e error) {
		if e = s.bwsDao.DelCacheRechargeAwards(ctx, []int64{m.New.ID}); e != nil {
			log.Info("clearPointAward s.bwsDao.DelCacheRechargeAwards(%d) error(%v)", m.New.ID, e)
		}
		return
	})
	if err = eg.Wait(); err == nil {
		log.Info("clearPointAward success msg(%d)", m.New.ID)
	}
	return
}

// clearPointLevel .
func (s *Service) clearPointLevel(c context.Context, id, bid int64) (err error) {
	eg := errgroupV2.WithContext(c)
	eg.Go(func(ctx context.Context) (e error) {
		if e = s.bwsDao.DelCachePointLevels(ctx, bid); e != nil {
			log.Info("clearPointLevel s.bwsDao.DelCachePointLevels %d error(%v)", bid, e)
		}
		return
	})
	eg.Go(func(ctx context.Context) (e error) {
		if e = s.bwsDao.DelCacheRechargeLevels(ctx, []int64{id}); e != nil {
			log.Info("clearPointLevel s.bwsDao.DelCacheRechargeLevels(%d) error(%v)", id, e)
		}
		return
	})
	err = eg.Wait()
	return
}

func (s *Service) clearPointSign(c context.Context, msg string) (err error) {
	var m struct {
		Table  string `json:"table"`
		Action string `json:"action"`
		New    struct {
			ID       int64 `json:"id"`
			Pid      int64 `json:"pid"`
			IsDelete int64 `json:"is_delete"`
		} `json:"new,omitempty"`
		Old struct {
			Pid      int64 `json:"pid"`
			IsDelete int64 `json:"is_delete"`
		} `json:"old,omitempty"`
	}
	if err = json.Unmarshal([]byte(msg), &m); err != nil {
		log.Error("clearPointSign json.Unmarshal msg(%s) error(%v)", msg, err)
		return
	}
	eg := errgroupV2.WithContext(c)
	eg.Go(func(ctx context.Context) (e error) {
		if e = s.bwsDao.DelCacheBwsSign(ctx, []int64{m.New.ID}); e != nil {
			log.Error("clearPointSign s.bwsDao.DelCacheBwsSign(%d) error(%v)", m.New.ID, e)
		}
		return
	})
	// 避免签到动作频繁更新 bid和id的对应关系
	if m.Action != match.ActUpdate || m.New.Pid != m.Old.Pid || m.New.IsDelete != m.Old.IsDelete {
		eg.Go(func(ctx context.Context) (e error) {
			if e = s.bwsDao.DelCacheSigns(ctx, m.New.Pid); e != nil {
				log.Error("clearPointSign s.bwsDao.DelCacheSigns(%d)  error(%v)", m.New.Pid, e)
			}
			return
		})
		log.Info("clearPointSign s.bwsDao.DelCachePoints(%d)", m.New.Pid)
	}
	if err = eg.Wait(); err == nil {
		log.Info("clearPointSign success id(%d)", m.New.ID)
	}
	return
}

func (s *Service) clearActAwardCache(c context.Context, msg string) (err error) {
	var m struct {
		New struct {
			ID  int64 `json:"id"`
			Sid int64 `json:"sid"`
		} `json:"new,omitempty"`
	}
	if err = json.Unmarshal([]byte(msg), &m); err != nil {
		log.Error("clearActAwardCache json.Unmarshal msg(%s) error(%v)", msg, err)
		return
	}
	log.Info("clearActAwardCache json.Unmarshal msg(%s)", msg)
	if m.New.ID <= 0 {
		return
	}
	return s.dao.DelCacheAwardSubject(c, m.New.Sid, m.New.ID)
}

// clearBwsPoint .
func (s *Service) clearBwsPoint(c context.Context, msg string) (err error) {
	var m struct {
		Table  string `json:"table"`
		Action string `json:"action"`
		New    struct {
			ID       int64 `json:"id"`
			Bid      int64 `json:"bid"`
			LockType int32 `json:"lock_type"`
			Unlocked int64 `json:"unlocked"`
			Del      int64 `json:"del"`
		} `json:"new,omitempty"`
		Old struct {
			Bid int64 `json:"bid"`
			Del int64 `json:"del"`
		} `json:"old,omitempty"`
	}
	if err = json.Unmarshal([]byte(msg), &m); err != nil {
		log.Error("clearBwsPoint json.Unmarshal msg(%s) error(%v)", msg, err)
		return
	}
	eg := errgroupV2.WithContext(c)
	eg.Go(func(ctx context.Context) (e error) {
		if e = s.bwsDao.DelCacheBwsPoints(ctx, []int64{m.New.ID}); e != nil {
			log.Error("clearBwsPoint s.bwsDao.DelCacheBwsPoints(%d) error(%v)", m.New.ID, e)
		}
		return
	})
	// 避免充能动作频繁更新 bid和id的对应关系
	if m.Action != match.ActUpdate || m.New.Bid != m.Old.Bid || m.New.Del != m.Old.Del {
		eg.Go(func(ctx context.Context) (e error) {
			if e = s.bwsDao.DelCachePoints(ctx, m.New.Bid); e != nil {
				log.Error("clearBwsPoint s.bwsDao.DelCachePoints(%d)  error(%v)", m.New.Bid, e)
			}
			return
		})
		log.Info("clearBwsPoint s.bwsDao.DelCachePoints(%d)", m.New.Bid)
	}
	if err = eg.Wait(); err == nil {
		log.Info("clearBwsPoint success id(%d)", m.New.ID)
	}
	return
}

func (s *Service) clearTaskCache(c context.Context, msg string) (err error) {
	var m struct {
		Table string `json:"table"`
		New   struct {
			ID         int64 `json:"id"`
			BusinessID int64 `json:"business_id"`
			ForeignID  int64 `json:"foreign_id"`
		} `json:"new,omitempty"`
	}
	if err = json.Unmarshal([]byte(msg), &m); err != nil {
		log.Error("ClearCache json.Unmarshal msg(%s) error(%v)", msg, err)
		return
	}
	log.Info("ClearCache json.Unmarshal msg(%s)", msg)
	switch m.Table {
	case _taskTable:
		err = s.taskDao.DelCacheTask(c, m.New.ID)
		if err != nil {
			return
		}
		return s.taskDao.DelCacheTaskIDs(c, m.New.BusinessID, m.New.ForeignID)
	}
	return
}

// AddFollow add match follow
func (s *Service) AddFollow(c context.Context, mid int64, teams []string) (err error) {
	if err = s.dao.AddFollow(c, mid, teams); err != nil {
		log.Error("s.dao.AddFollow mid(%d) teams(%v)  error(%v)", mid, teams, err)
	}
	return
}

// Follow get match follow
func (s *Service) Follow(c context.Context, mid int64) (res []string, err error) {
	if res, err = s.dao.Follow(c, mid); err != nil {
		log.Error("s.dao.Follow mid(%d)  error(%v)", mid, err)
	}
	if len(res) == 0 {
		res = _emptyFollow
	}
	return
}

// ObjectsUnStart get unstart object list.
func (s *Service) ObjectsUnStart(c context.Context, mid int64, p *match.ParamObject) (rs []*match.Object, count int, err error) {
	var (
		userGuess []*match.UserLog
		objects   []*match.Object
		start     = (p.Pn - 1) * p.Ps
		end       = start + p.Ps - 1
	)
	// get from  cache.
	if rs, count, err = s.dao.ObjectsCache(c, p.Sid, start, end); err != nil || len(rs) == 0 {
		if objects, err = s.dao.ObjectsUnStart(c, p.Sid); err != nil {
			log.Error("s.dao.ObjectsUnStart id(%d)  error(%v)", p.Sid, err)
			return
		}
		count = len(objects)
		if count == 0 || count < start {
			rs = _emptyObjects
			return
		}
		s.cache.Do(c, func(c context.Context) {
			s.dao.SetObjectsCache(c, p.Sid, objects, count)
		})
		if count > end+1 {
			rs = objects[start : end+1]
		} else {
			rs = objects[start:]
		}
	}
	if mid > 0 {
		if userGuess, err = s.ListGuess(c, p.Sid, mid); err != nil {
			log.Error("s.ListGuess(%d,%d) error(%v)", p.Sid, mid, err)
			err = nil
		}
		for _, rsObj := range rs {
			for _, guess := range userGuess {
				if rsObj.ID == guess.MOId {
					rsObj.UserResult = guess.Result
					break
				}
			}
		}
	}
	return
}

// AddSuits add more suit.
func (s *Service) AddSuits(c context.Context, mid, sid int64) (err error) {
	var guessCout int
	if sid != s.c.Rule.S9Guess.SeasonID {
		err = ecode.ActivityNoAward
		return
	}
	// checkout award time
	nowTs := time.Now().Unix()
	if nowTs < s.c.Rule.S9Guess.Stime {
		err = ecode.ActivityNotStart
		return
	}
	if nowTs > s.c.Rule.S9Guess.Etime {
		err = ecode.ActivityOverEnd
		return
	}
	if count, e := s.dao.RsNXGet(c, guessCountKey(mid, sid)); e != nil {
		log.Error("AddSuits s.dao.RsNXGet mid(%d) sid(%d) error(%v)", mid, sid, e)
	} else if guessCout, err = strconv.Atoi(count); err != nil {
		log.Error("AddSuits strconv.Atoi count(%s) mid(%d) sid(%d) error(%v)", count, mid, sid, e)
		err = ecode.ActivityTaskNoAward
		return
	}
	if guessCout < s.c.Rule.S9Guess.LimitCount {
		log.Warn("AddSuits strconv.Atoi guessCout(%d) mid(%d) sid(%d)", guessCout, mid, sid)
		err = ecode.ActivityTaskNoAward
		return
	}
	if check, e := s.dao.RsSetNX(c, suitsCheckKey(mid, sid), s.c.Rule.S9Guess.LimitExpire); e != nil || !check {
		if e != nil {
			log.Error("AwardTaskSpecials s.dao.RsSetNX  mid(%d) sid(%d) error(%v)", mid, sid, e)
		}
		err = ecode.ActivityHasAward
		return
	}
	// add mid more suits
	exps := []int64{s.c.Rule.S9Guess.SuitsExpire, s.c.Rule.S9Guess.SuitsExpire, s.c.Rule.S9Guess.SuitsExpire}
	if _, err = s.suitClient.GrantByPids(c, &api.GrantByPidsReq{Mid: mid, Pids: s.c.Rule.S9Guess.SuitIDs, Expires: exps}); err != nil {
		log.Error("s.suitClient.GrantByPids mid(%d) Pids(%v) sid(%d) error(%v)", mid, s.c.Rule.S9Guess.SuitIDs, sid, err)
		err = ecode.ActivitySuitsFail
	}
	return
}

func suitsCheckKey(mid, sid int64) string {
	return fmt.Sprintf("suits_c_k_%d_%d", mid, sid)
}

func guessCountKey(mid, sid int64) string {
	return fmt.Sprintf("guess_c_k_%d_%d", mid, sid)
}

func (s *Service) clearQuesCache(c context.Context, msg string) (err error) {
	var m struct {
		Table string `json:"table"`
		New   struct {
			ID int64 `json:"id"`
		} `json:"new,omitempty"`
	}
	if err = json.Unmarshal([]byte(msg), &m); err != nil {
		log.Error("ClearCache json.Unmarshal msg(%s) error(%v)", msg, err)
		return
	}
	log.Info("ClearCache json.Unmarshal msg(%s)", msg)
	switch m.Table {
	case _quesDetailTable:
		return s.quesDao.DelCacheDetail(c, m.New.ID)
	}
	return
}

func (s *Service) clearLotteryCache(c context.Context, msg string) (err error) {
	var m struct {
		Table string `json:"table"`
		New   struct {
			Sid string `json:"sid"`
		} `json:"new,omitempty"`
	}
	if err = json.Unmarshal([]byte(msg), &m); err != nil {
		log.Error("ClearCache json.Unmarshal msg(%s) error(%v)", msg, err)
		return
	}
	log.Info("ClearCache json.Unmarshal msg(%s)", msg)
	if m.Table == _actLotteryGiftTable {
		if err = s.lottDao.DeleteLotteryGift(c, m.New.Sid); err != nil {
			log.Error("s.lottDao.DeleteLotteryGift sid(%s)  error(%v)", m.New.Sid, err)
		}
		if err = s.lottV2Dao.DeleteLotteryGift(c, m.New.Sid); err != nil {
			log.Error("s.lottV2Dao.DeleteLotteryGift sid(%s) error(%v)", m.New.Sid, err)
		}
	} else if m.Table == _actLotteryInfoTable {
		if err = s.lottDao.DeleteLotteryInfo(c, m.New.Sid); err != nil {
			log.Error("s.lottDao.DeleteLotteryInfo sid(%s) error(%v)", m.New.Sid, err)
		}
		if err = s.lottV2Dao.DeleteLotteryInfo(c, m.New.Sid); err != nil {
			log.Error("s.lottV2Dao.DeleteLotteryInfo sid(%s) error(%v)", m.New.Sid, err)
		}
	} else if m.Table == _actLotteryTimesTable {
		if err = s.lottDao.DeleteLotteryTimesConfig(c, m.New.Sid); err != nil {
			log.Error("s.lottDao.DeleteLotteryTimesConfig sid(%s) error(%v)", m.New.Sid, err)
		}
		if err = s.lottV2Dao.DeleteLotteryTimesConfig(c, m.New.Sid); err != nil {
			log.Error("s.lottV2Dao.DeleteLotteryTimesConfig sid(%s) error(%v)", m.New.Sid, err)
		}
	}
	return
}

func (s *Service) clearLotteryActCache(c context.Context, msg string) (err error) {
	var m struct {
		Table string `json:"table"`
		New   struct {
			LotteryId string `json:"lottery_id"`
		} `json:"new,omitempty"`
	}
	if err = json.Unmarshal([]byte(msg), &m); err != nil {
		log.Error("ClearCache json.Unmarshal msg(%s) error(%v)", msg, err)
		return
	}
	log.Info("ClearCache json.Unmarshal msg(%s)", msg)
	if m.Table == _actLotteryTable {
		if err = s.lottDao.DeleteLottery(c, m.New.LotteryId); err != nil {
			log.Error("s.lottDao.DeleteLottery sid(%s) error(%v)", m.New.LotteryId, err)
		}
		if err = s.lottV2Dao.DeleteLottery(c, m.New.LotteryId); err != nil {
			log.Error("s.lottDao.DeleteLottery sid(%s) error(%v)", m.New.LotteryId, err)
		}
	}
	return
}

func (s *Service) clearLotteryMemberGroupCache(c context.Context, msg string) (err error) {
	var m struct {
		Table string `json:"table"`
		New   struct {
			Sid string `json:"sid"`
		} `json:"new,omitempty"`
	}
	if err = json.Unmarshal([]byte(msg), &m); err != nil {
		log.Error("ClearCache json.Unmarshal msg(%s) error(%v)", msg, err)
		return
	}
	log.Info("ClearCache json.Unmarshal msg(%s)", msg)
	if m.Table == _actLotteryMemberGroupTable {
		if err = s.lottV2Dao.DeleteMemberGroup(c, m.New.Sid); err != nil {
			log.Error("s.lottDao.DeleteMemberGroup sid(%s) error(%v)", m.New.Sid, err)
		}
	}
	return
}

func (s *Service) clearUpCache(c context.Context, msg string) (err error) {
	var m struct {
		Table string `json:"table"`
		New   struct {
			ID  int64 `json:"id"`
			Mid int64 `json:"mid"`
			Aid int64 `json:"aid"`
		} `json:"new,omitempty"`
	}
	if err = json.Unmarshal([]byte(msg), &m); err != nil {
		log.Error("ClearCache json.Unmarshal msg(%s) error(%v)", msg, err)
		return
	}
	log.Info("ClearCache json.Unmarshal msg(%s)", msg)
	if m.Table == _actUpTable {
		if err = s.dao.DeleteCacheActUpBySid(c, m.New.ID); err != nil {
			log.Error("s.dao.DeleteCacheActUpBySid sid(%d) error(%v)", m.New.ID, err)
		}
		if err = s.dao.DeleteCacheActUp(c, m.New.Mid); err != nil {
			log.Error("s.dao.DeleteCacheActUp mid(%d) error(%v)", m.New.Mid, err)
		}
		if err = s.dao.DeleteCacheActUpByAid(c, m.New.Aid); err != nil {
			log.Error("s.dao.DeleteCacheActUpByAid aid(%d) error(%v)", m.New.Mid, err)
		}
	}
	return
}

func (s *Service) clearLikesCache(ctx context.Context, msg string) error {
	var m struct {
		Table string `json:"table"`
		New   struct {
			Mid int64 `json:"mid"`
			Sid int64 `json:"sid"`
		} `json:"new,omitempty"`
	}
	err := json.Unmarshal([]byte(msg), &m)
	if err != nil {
		log.Errorc(ctx, "ClearCache json.Unmarshal msg(%s) error(%v)", msg, errors.WithStack(err))
		return err
	}
	log.Info("ClearCache json.Unmarshal msg(%s)", msg)
	if m.Table == _likesTable {
		if err = s.dao.DelCacheActivityArchives(ctx, m.New.Sid, m.New.Mid); err != nil {
			log.Errorc(ctx, "Failed to delete activity archive cache: sid(%d) mid(%d) error(%v)", m.New.Sid, m.New.Mid, err)
			return err
		}
	}
	return nil
}

func (s *Service) clearActSubjectCache(c context.Context, msg json.RawMessage) (err error) {
	var m struct {
		ID int64 `json:"id"`
	}
	if err = json.Unmarshal(msg, &m); err != nil {
		log.Errorc(c, "clearActSubjectCache json.Unmarshal msg(%s) error(%v)", msg, err)
		return
	}
	log.Info("clearActSubjectCache json.Unmarshal msg(%s)", msg)
	if err = s.dao.DeleteActSubjectCache(c, m.ID); err != nil {
		log.Errorc(c, "clearActSubjectCache s.dao.clearActSubjectCache(c, %d) error(%v)", m.ID, err)
	}
	if err = s.dao.DeleteActSubjectWithStateCache(c, m.ID); err != nil {
		log.Errorc(c, "clearActSubjectCache s.dao.DeleteActSubjectWithStateCache(c, %d) error(%v)", m.ID, err)
	}
	return
}

func (s *Service) clearUpActReserveRelationCache(c context.Context, msg json.RawMessage) (err error) {
	var m struct {
		Sid int64 `json:"sid"`
	}
	if err = json.Unmarshal(msg, &m); err != nil {
		log.Errorc(c, "[UpActReserveRelationDelCache]clearUpActReserveRelationCache json.Unmarshal msg(%s) error(%v)", msg, err)
		return
	}
	log.Info("clearUpActReserveRelationCache json.Unmarshal msg(%s)", msg)
	for i := 0; i < 3; i++ {
		if err = s.dao.DelUpActReserveRelationInfoCache(c, m.Sid); err == nil {
			break
		}
	}
	if err != nil {
		log.Errorc(c, "[UpActReserveRelationDelCache]clearUpActReserveRelationCache s.dao.DelUpActReserveRelationInfoCache(c, %d) error(%v)", m.Sid, err)
	}
	return
}

func (s *Service) clearUpActReserveRelationReachCache(c context.Context, msg json.RawMessage) (err error) {
	var m struct {
		Mid int64 `json:"mid"`
	}
	if err = json.Unmarshal(msg, &m); err != nil {
		log.Errorc(c, "[UpActReserveRelationDelCache]clearUpActReserveRelationReachCache json.Unmarshal msg(%s) error(%v)", msg, err)
		return
	}
	log.Info("clearUpActReserveRelationReachCache json.Unmarshal msg(%s)", msg)
	for i := 0; i < 3; i++ {
		if err = s.dao.DelUpActReserveRelationInfoReachCache(c, m.Mid); err == nil {
			break
		}
	}
	if err != nil {
		log.Errorc(c, "[UpActReserveRelationDelCache]clearUpActReserveRelationReachCache s.dao.DelUpActReserveRelationInfoReachCache(c, %d) error(%v)", m.Mid, err)
	}
	return
}

func (s *Service) clearUpActReserveRelation4LiveCache(c context.Context, msg json.RawMessage) (err error) {
	var m struct {
		Mid  int64 `json:"mid"`
		Type int64 `json:"type"`
	}
	if err = json.Unmarshal(msg, &m); err != nil {
		log.Errorc(c, "[UpActReserveRelationDelCache]clearUpActReserveRelation4LiveCache json.Unmarshal msg(%s) error(%v)", msg, err)
		return
	}
	// 非直播返回
	if m.Type != int64(pb.UpActReserveRelationType_Live) {
		return
	}
	log.Info("clearUpActReserveRelation4LiveCache json.Unmarshal msg(%s)", msg)
	for i := 0; i < 3; i++ {
		if err = s.dao.UpActReserveRelation4LiveCache(c, m.Mid); err == nil {
			break
		}
	}
	if err != nil {
		log.Errorc(c, "[UpActReserveRelationDelCache]clearUpActReserveRelation4LiveCache s.dao.UpActReserveRelation4LiveCache(c, %d) error(%v)", m.Mid, err)
	}
	return
}
