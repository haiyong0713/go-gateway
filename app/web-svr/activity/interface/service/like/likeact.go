package like

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"math"
	"net"
	"sort"
	"strconv"
	"strings"
	"time"

	accapi "git.bilibili.co/bapis/bapis-go/account/service"
	accRelationApi "git.bilibili.co/bapis/bapis-go/account/service/relation"
	relationapi "git.bilibili.co/bapis/bapis-go/account/service/relation"
	relmdl "git.bilibili.co/bapis/bapis-go/account/service/relation"
	spyapi "git.bilibili.co/bapis/bapis-go/account/service/spy"
	artmdl "git.bilibili.co/bapis/bapis-go/article/model"
	artapi "git.bilibili.co/bapis/bapis-go/article/service"
	thumbupapi "git.bilibili.co/bapis/bapis-go/community/service/thumbup"
	passapi "git.bilibili.co/bapis/bapis-go/passport/service/user"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/sync/errgroup.v2"
	xtime "go-common/library/time"
	arccli "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/interface/client"
	"go-gateway/app/web-svr/activity/interface/dao/like"
	l "go-gateway/app/web-svr/activity/interface/model/like"
	likemdl "go-gateway/app/web-svr/activity/interface/model/like"
	suitapi "go-main/app/account/usersuit/service/api"

	"github.com/pkg/errors"
)

const (
	_countryCodeCN       = "86"
	_starFanLimit        = 10000
	_aidFlowControlSize  = 30
	_riskUpAction        = "Up_vote"
	_riskVideoVoteAction = "activity_common_vote"
	_riskUpApi           = "/x/activity/up/act"
	_riskUpActivityUID   = "knowledge_up_selection"
)

var _emptyList = make([]*l.List, 0)

// Subject service
func (s *Service) Subject(c context.Context, sid int64) (res *l.Subject, err error) {
	var (
		mc      = true
		subErr  error
		likeErr error
	)
	if res, err = s.dao.InfoCache(c, sid); err != nil {
		err = nil
		mc = false
	} else if res != nil {
		if res, err = s.LikeArc(c, res); err != nil {
			return
		}
	}
	eg := errgroup.WithContext(c)
	var ls = make([]*l.Like, 0)
	eg.Go(func(errCtx context.Context) error {
		res, subErr = s.dao.Subject(errCtx, sid)
		return subErr
	})
	eg.Go(func(errCtx context.Context) error {
		ls, likeErr = s.dao.LikeTypeList(errCtx, sid)
		return likeErr
	})
	if err = eg.Wait(); err != nil {
		log.Error("eg.Wait error(%v)", err)
		return
	}
	if res != nil {
		res.List = ls
	}
	if mc {
		err = s.dao.SetInfoCache(c, res, sid)
		if err != nil {
			log.Error("SetInfoCache error(%v)", err)
		}
	}
	if res, err = s.LikeArc(c, res); err != nil {
		return
	}
	return
}

// LikeArc service
func (s *Service) LikeArc(c context.Context, sub *l.Subject) (res *l.Subject, err error) {
	if sub != nil {
		if sub.ID == 0 {
			res = nil
		} else {
			res = sub
			var (
				ok   bool
				arcs map[int64]*arccli.Arc
				aids []int64
			)
			for _, l := range res.List {
				aids = append(aids, l.Wid)
			}
			if arcs, err = s.archives(c, aids); err != nil || arcs == nil {
				log.Error("s.archives(arcAids:(%v), arcs), err(%v)", aids, err)
				return
			}
			for _, l := range res.List {
				if l.Archive, ok = arcs[l.Wid]; !ok {
					log.Info("s.arcs.wid:(%d),ok(%v)", l.Wid, ok)
					continue
				}
			}
		}
	}
	return
}

// OnlineVote Service
func (s *Service) OnlineVote(c context.Context, mid, vote, stage, aid int64) (res bool, err error) {
	res = true
	if vote != _yes && vote != _no {
		err = nil
		res = false
		return
	}
	var incrKey string
	midStr := strconv.FormatInt(mid, 10)
	aidStr := strconv.FormatInt(aid, 10)
	stageStr := strconv.FormatInt(stage, 10)
	midKye := midStr + ":" + aidStr + ":" + stageStr
	if res, err = s.dao.RsSetNX(c, midKye, 0); err != nil {
		log.Error("s.OnlineVote.reids(mid:(%v),,vote:(%v),stage:(%v)), err(%v)", mid, vote, stage, err)
		return
	}
	if !res {
		return
	}
	if vote == _yes {
		incrKey = aidStr + ":" + stageStr + ":yes"
	} else {
		incrKey = aidStr + ":" + stageStr + ":no"
	}
	if mid == 288239 || mid == 26366366 || mid == 20453897 {
		log.Info("288239,26366366,20453897")
		if res, err = s.dao.Incrby(c, incrKey); err != nil {
			log.Error("s.OnlineVote.Incrby(key:(%v)", incrKey)
			return
		}
	} else {
		if res, err = s.dao.Incr(c, incrKey); err != nil {
			log.Error("s.OnlineVote.Incr(key:(%v)", incrKey)
			return
		}
	}
	s.dao.CVoteLog(c, 0, aid, mid, stage, vote)
	return
}

// Ltime service
func (s *Service) Ltime(c context.Context, sid int64) (res map[string]interface{}, err error) {
	var key = "ltime:" + strconv.FormatInt(sid, 10)
	var b []byte
	if b, err = s.dao.Rb(c, key); err != nil {
		log.Error("s.dao.Rb((%v), err(%v)", key, err)
		return
	}
	if b == nil {
		res = nil
		return
	}
	if err = json.Unmarshal(b, &res); err != nil {
		log.Error("s.Ltime.Unmarshal((%v), err(%v)", b, err)
		return
	}
	if res["time"] != nil {
		if st, ok := res["time"].(float64); ok {
			res["currentTime"] = time.Now().Unix() - int64(st)
		}
	}
	return
}

func (s *Service) yellowGreenPeriod(sid int64) *likemdl.YellowGreenPeriod {
	for _, period := range s.c.YellowAndGreen.Period {
		if period.YellowSid == sid || period.GreenSid == sid {
			return period
		}
	}
	return nil
}

// LikeAct service
func (s *Service) LikeAct(c context.Context, p *l.ParamAddLikeAct, mid int64) (res *l.ActReply, err error) {
	var (
		subject   *l.SubjectItem
		likeItem  *l.Item
		memberRly *accapi.ProfileReply
		subErr    error
		likeErr   error
		actID     int64
	)
	eg := errgroup.WithContext(c)
	eg.Go(func(errCtx context.Context) error {
		subject, subErr = s.dao.ActSubject(errCtx, p.Sid)
		return subErr
	})
	eg.Go(func(errCtx context.Context) error {
		likeItem, likeErr = s.dao.Like(errCtx, p.Lid)
		return likeErr
	})
	if err = eg.Wait(); err != nil {
		log.Error("eg.Wait error(%v)", err)
		return
	}
	if subject.ID == 0 || subject.Type == l.STORYKING {
		err = ecode.ActivityHasOffLine
		return
	}
	if likeItem.ID == 0 || likeItem.Sid != p.Sid {
		err = ecode.ActivityLikeHasOffLine
		return
	}
	nowTs := time.Now().Unix()
	if int64(subject.Lstime) >= nowTs {
		err = ecode.ActivityLikeNotStart
		return
	}
	if int64(subject.Letime) <= nowTs {
		err = ecode.ActivityLikeHasEnd
		return
	}
	yellowGreenPeriod := s.yellowGreenPeriod(p.Sid)
	// 只上报风控,忽略错误
	if _, riskErr := s.checkRiskVote(c, mid, p); riskErr != nil {
		// 黄绿投票风控返回错误
		if yellowGreenPeriod != nil && xecode.EqualError(ecode.ActivityRiskRejectErr, riskErr) {
			err = ecode.ActivityVoteRejectErr
			return
		}
		log.Errorc(c, "LikeAct s.checkRiskVote mid(%d) p(%+v) error(%+v)", mid, p, riskErr)
	}
	if memberRly, err = s.accClient.Profile3(c, &accapi.MidReq{Mid: mid}); err != nil {
		log.Error(" s.acc.Profile3(c,&accmdl.ArgMid{Mid:%d}) error(%v)", mid, err)
		return
	}
	if err = s.judgeUser(c, subject, memberRly.Profile); err != nil {
		return
	}
	var (
		isStar       bool
		starFanLimit int64 = _starFanLimit
	)
	for _, v := range s.starConf {
		if v.JoinSid == p.Sid {
			isStar = true
			if v.FanLimit > 0 {
				starFanLimit = v.FanLimit
			}
			break
		}
	}
	if isStar || p.Sid == s.c.Star.JoinSid {
		var stat *relationapi.StatReply
		if stat, err = s.relClient.Stat(c, &relationapi.MidReq{Mid: mid}); err != nil {
			log.Error("s.relClient.Stat(%d) error(%v)", mid, err)
			return
		} else if stat.Follower > starFanLimit {
			err = ecode.ActivityUpFanLimit
			return
		}
	}
	var (
		likeAct map[int64]int
		lids    = []int64{p.Lid}
	)
	if likeAct, err = s.dao.LikeActs(c, p.Sid, mid, lids); err != nil {
		log.Error("s.dao.LikeActMidList(%v) error(%+v)", p, err)
		return
	}
	if _, ok := likeAct[p.Lid]; !ok {
		log.Error("s.dao.LikeActMidList() get lid value error()")
		return
	}
	//科学3活动直接返回结果
	if subject.IsDailyLike() {
		if actID, _, err = s.actDailyLike(c, p.Sid, p.Lid, p.Score, mid, nowTs, subject, likeItem, memberRly); err != nil {
			return
		}
		res = &l.ActReply{Lid: p.Lid, Score: p.Score, ActID: actID}
		return
	}
	isLikeType := s.isLikeType(subject.Type)
	if likeAct[p.Lid] == like.HasLike {
		if isLikeType == _like {
			err = ecode.ActivityLikeHasLike
		} else if isLikeType == _vote {
			err = ecode.ActivityLikeHasVote
		} else {
			err = ecode.ActivityLikeHasGrade
		}
		return
	}
	if subject.LikeLimit > 0 {
		nowScore, _ := s.dao.GetLikeLimitNum(c, p.Sid, mid)
		if nowScore >= subject.LikeLimit {
			err = ecode.ActivityOverLikeLimit
			return
		}
	}
	// yellow green fight rule
	if yellowGreenPeriod != nil {
		var yeGrCheck bool
		if yeGrCheck, err = s.dao.RsSetNX(c, s.yeGrKey(mid, nowTs, yellowGreenPeriod), s.c.Rule.YeGrExpire); err != nil {
			return
		}
		if !yeGrCheck {
			err = ecode.ActivityLikeHasLike
			return
		}
	}
	var score int64
	if isLikeType == _like || isLikeType == _vote {
		score = l.LIKESCORE
		if !subject.AttrFlag(l.FLAGMONTHSCORE) {
			if memberRly.Profile.Vip.IsValid() && !memberRly.Profile.Vip.IsAnnual() {
				score += subject.MonthScore
			}
		}
		if !subject.AttrFlag(l.FLAGYEARSCORE) {
			if memberRly.Profile.Vip.IsAnnual() {
				score += subject.YearScore
			}
		}
	} else {
		score = p.Score
	}
	if likeItem.StickTop == 0 {
		if err = s.dao.SetRedisCache(c, p.Sid, p.Lid, score, likeItem.Type); err != nil {
			log.Error("s.dao.SetRedisCache(%v) error(%+v)", p, err)
			return
		}
	} else {
		if err = s.dao.DelLikeListLikes(c, p.Sid, []*l.Item{likeItem}); err != nil {
			log.Error("s.dao.SetRedisCache(%v) error(%+v)", p, err)
			return
		}
	}
	likeActAdd := &l.Action{
		Lid:    p.Lid,
		Mid:    mid,
		Sid:    p.Sid,
		Action: score,
		IPv6:   make([]byte, 0),
	}
	if IPv6 := net.ParseIP(metadata.String(c, metadata.RemoteIP)); IPv6 != nil {
		likeActAdd.IPv6 = IPv6
	}
	if actID, err = s.dao.LikeActAdd(c, likeActAdd); err != nil {
		log.Error("s.dao.LikeActAdd(%v) error(%+v)", p, err)
		return
	}
	s.dao.AddCacheLikeActs(c, p.Sid, mid, map[int64]int{p.Lid: like.HasLike})
	res = &l.ActReply{Lid: p.Lid, Score: score, ActID: actID}
	if subject.LikeLimit > 0 {
		s.dao.SetLikeLimitNum(c, p.Sid, mid, 1)
	}
	//  yellow green fight award
	if yellowGreenPeriod != nil {
		s.cache.Do(context.Background(), func(ctx context.Context) {
			lotterySid := yellowGreenPeriod.LotterySid
			cid := yellowGreenPeriod.Cid
			lottType := _other
			orderNo := strconv.FormatInt(mid, 10) + strconv.FormatInt(yellowGreenPeriod.YellowSid, 10) + strconv.FormatInt(nowTs, 10)
			if e := s.lotterySvr.AddLotteryTimes(ctx, lotterySid, mid, cid, lottType, 0, orderNo, false); e != nil {
				log.Errorc(c, "s.lotterySvr.AddLotteryTimes(%d,%s,%d) error(%+v)", mid, lotterySid, cid, e)
			}
		})
	}
	// special add suit .
	if suitID, ok := s.c.Rule.AutoSuitIDs[strconv.FormatInt(p.Sid, 10)]; ok && suitID > 0 {
		s.cache.Do(c, func(c context.Context) {
			if _, e := s.suitClient.GrantByMids(c, &suitapi.GrantByMidsReq{Mids: []int64{mid}, Pid: suitID, Expire: s.c.Rule.PaySuitExpire}); e != nil {
				log.Error("LikeAct s.suitClient.GrantByMids mid(%d) suidID(%d) expire(%d) error(%v)", mid, suitID, s.c.Rule.PaySuitExpire, e)
				return
			}
			log.Info("special send suit success(%d,%d,%d)", mid, suitID, s.c.Rule.PaySuitExpire)
		})
	}
	return
}

func (s *Service) checkRiskVote(ctx context.Context, mid int64, params *l.ParamAddLikeAct) (res bool, err error) {
	otherEventCtx := &l.VideoVoteEventCtx{
		Action:      _riskVideoVoteAction,
		Mid:         mid,
		ActivityUid: _riskVideoVoteAction,
		TargetID:    params.Lid,
		ID:          params.Sid,
		Score:       params.Score,
		Buvid:       params.Buvid,
		Ip:          params.IP,
		Platform:    params.Platform,
		Ctime:       time.Now().Format("2006-01-02 15:04:05"),
		Api:         params.API,
		Origin:      params.Origin,
		UserAgent:   params.UA,
		Build:       params.Build,
		Referer:     params.Referer,
		MobiApp:     params.MobiApp,
	}
	if res, err = s.silverDao.RuleCheckCommon(ctx, _riskVideoVoteAction, otherEventCtx); err != nil {
		log.Errorc(ctx, "LikeAct checkRiskVote mid(%d) VideoVoteEventCtx(%+v) error(%+v)", mid, otherEventCtx, err)
	}
	return
}

// LikeActBySidCVId
func (s *Service) LikeActBySidCVId(c context.Context, p *l.ParamAddLikeActWithSidCVId, mid int64) (err error) {
	// CVId => CV12378468
	spData := strings.Split(p.CVId, "cv")
	if len(spData) != 2 {
		err = ecode.ActivityLikeWithSidCvidParamErr
		return
	}
	prefix := spData[0]
	if prefix != "" {
		err = ecode.ActivityLikeWithSidCvidParamErr
		return
	}
	postfix := spData[1]
	wid, e := strconv.ParseInt(postfix, 10, 64)
	if e != nil {
		err = ecode.ActivityLikeWithSidCvidParamErr
		return
	}
	// 通过CVId查询likes表中lid
	lid, err := s.GetLidByWid(c, wid)

	// 网络请求失败
	if err != nil {
		err = ecode.ActivityLikeWithSidCvidNetErr
		return
	}

	// 数据不存在
	if lid == 0 {
		err = ecode.ActivityLikeWithSidCvidDataErr
		return
	}

	// 调用原有方法进行点赞
	_, err = s.LikeAct(c, &l.ParamAddLikeAct{Sid: p.Sid, Lid: lid, Score: p.Score}, mid)

	return
}

func (s *Service) GetLidByWid(c context.Context, wid int64) (lid int64, err error) {
	// 读缓存
	if lid, err = s.dao.GetLidByWidFromCache(c, wid); err != nil {
		log.Errorc(c, "GetLidByWid Err %v", err)
	}
	// 读取到数据返回结果
	if lid > 0 {
		return
	}
	// 未读取到数据或者缓存读取连接异常回源数据库
	if lid, err = s.dao.GetLidByWidFromDB(c, wid); err != nil {
		log.Errorc(c, "GetLidByWid GetLidByWidFromDB Err :%v", err)
		return
	}
	// 拿到正常的数据
	if lid > 0 {
		if err = s.dao.SetLidByWidToCache(c, wid, lid); err != nil {
			if err = s.dao.SetLidByWidToCache(c, wid, lid); err != nil {
				log.Errorc(c, "GetLidByWid %v", err)
				// 终止 不再写过期时间 但是函数已经通过DB获取到准确数据 不再返回err
				err = nil
				return
			}
		}
		// 写缓存正常同时设置过期时间
		if err = s.dao.SetLidByWidToCacheExpireTime(c, wid); err != nil {
			if err = s.dao.SetLidByWidToCacheExpireTime(c, wid); err != nil {
				log.Errorc(c, "SetLidByWidToCacheExpireTime %v", err)
				// 同上 函数已经通过DB获取到准确数据 不再返回err
				err = nil
			}
		}
	}

	return
}

// BatchLikeAct .
func (s *Service) BatchLikeAct(c context.Context, mid, sid int64, lids []int64) (err error) {
	var (
		subject         *l.SubjectItem
		likeItems       map[int64]*l.Item
		memberRly       *accapi.ProfileReply
		likeActs        map[int64]int
		likeType        int64
		subErr, likeErr error
	)
	if sid != s.c.Taaf.Sid && sid != s.c.Timemachine.FlagSid {
		err = xecode.RequestErr
		return
	}
	// 时光机最多选3个
	if sid == s.c.Timemachine.FlagSid && len(lids) > 3 {
		err = xecode.RequestErr
		return
	}
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) error {
		subject, subErr = s.dao.ActSubject(ctx, sid)
		return subErr
	})
	eg.Go(func(ctx context.Context) error {
		likeItems, likeErr = s.dao.Likes(ctx, lids)
		return likeErr
	})
	if err = eg.Wait(); err != nil {
		log.Error("BatchLikeAct mid(%d) sid(%d) lids(%v) error(%v)", mid, sid, lids, err)
		return
	}
	if subject.ID == 0 || subject.Type == l.STORYKING {
		err = ecode.ActivityHasOffLine
		return
	}
	nowTs := time.Now().Unix()
	if int64(subject.Lstime) >= nowTs {
		err = ecode.ActivityLikeNotStart
		return
	}
	if int64(subject.Letime) <= nowTs {
		err = ecode.ActivityLikeHasEnd
		return
	}
	for _, lid := range lids {
		if item, ok := likeItems[lid]; !ok || item == nil || item.ID == 0 || item.Sid != sid {
			err = ecode.ActivityLikeHasOffLine
			return
		}
		likeType = likeItems[lid].Type
	}
	if memberRly, err = s.accClient.Profile3(c, &accapi.MidReq{Mid: mid}); err != nil {
		log.Error(" s.acc.Profile3(c,&accmdl.ArgMid{Mid:%d}) error(%v)", mid, err)
		return
	}
	if err = s.judgeUser(c, subject, memberRly.Profile); err != nil {
		return
	}
	// check if has liked
	if likeActs, err = s.dao.LikeActs(c, sid, mid, lids); err != nil {
		log.Error("s.dao.LikeActs sid(%d) mid(%v) error(%+v)", sid, mid, err)
		return
	}
	isLikeType := s.isLikeType(subject.Type)
	for _, lid := range lids {
		if hasLike, ok := likeActs[lid]; ok && hasLike == like.HasLike {
			if isLikeType == _like {
				err = ecode.ActivityLikeHasLike
			} else if isLikeType == _vote {
				err = ecode.ActivityLikeHasVote
			} else {
				err = ecode.ActivityLikeHasGrade
			}
			return
		}
	}
	if sid == s.c.Taaf.Sid || sid == s.c.Timemachine.FlagSid {
		taafLids, _ := s.dao.LikeActLids(c, sid, mid)
		if len(taafLids) > 0 {
			err = ecode.ActivityOverLikeLimit
			return
		}
	}
	IPv6 := net.ParseIP(metadata.String(c, metadata.RemoteIP))
	if IPv6 == nil {
		IPv6 = make([]byte, 0)
	}
	// 支持多票
	adds := make(map[int64]*l.Action)
	for _, lid := range lids {
		if _, ok := adds[lid]; ok {
			adds[lid].Action++
			continue
		}
		adds[lid] = &l.Action{
			Lid:    lid,
			Mid:    mid,
			Sid:    sid,
			Action: l.LIKESCORE,
			IPv6:   IPv6,
		}
		likeActs[lid] = like.HasLike
	}
	if err = s.dao.BatchSetLikeScoreCache(c, sid, l.LIKESCORE, likeType, lids); err != nil {
		log.Error("s.dao.SetRedisCache sid(%d) lids(%v) error(%+v)", sid, lids, err)
		return
	}
	if err = s.dao.BatchLikeActAdd(c, adds); err != nil {
		log.Error("s.dao.LikeActAdd(%v) error(%+v)", adds, err)
		return
	}
	s.cache.Do(c, func(c context.Context) {
		s.dao.AddCacheLikeActs(c, sid, mid, likeActs)
		// cache user like list lids
		if sid == s.c.Taaf.Sid || sid == s.c.Taaf.SidV2 || sid == s.c.Timemachine.FlagSid {
			items := make([]*l.LidItem, 0, len(lids))
			for _, v := range adds {
				items = append(items, &l.LidItem{Lid: v.Lid, Action: v.Action, ActTime: xtime.Time(nowTs)})
			}
			s.dao.AddCacheLikeActLids(c, sid, items, mid)
		}
	})
	return
}

// LikeActLikes get user sid all likes.
func (s *Service) LikeActLikes(c context.Context, sid, mid int64) (data []*l.LikeListItem, err error) {
	var (
		lidItems []*l.LidItem
		lids     []int64
		contents map[int64]*l.LikeContent
	)
	if sid != s.c.Taaf.Sid && sid != s.c.Taaf.SidV2 {
		err = xecode.RequestErr
		return
	}
	if lidItems, err = s.dao.LikeActLids(c, sid, mid); err != nil {
		log.Error("LikeActLikes s.dao.LikeActLids(%d,%d) error(%v)", sid, mid, err)
		err = nil
		return
	}
	if len(lidItems) == 0 {
		return
	}
	for _, v := range lidItems {
		if v != nil {
			lids = append(lids, v.Lid)
		}
	}
	if contents, err = s.dao.LikeContent(c, lids); err != nil {
		log.Error("LikeActLikes s.dao.LikeContent(%v) error(%v)", lids, err)
		err = nil
		return
	}
	for _, v := range lidItems {
		if v != nil {
			if item, ok := contents[v.Lid]; ok && item != nil {
				data = append(data, &l.LikeListItem{ActTime: v.ActTime, Action: v.Action, LikeContent: item})
			}
		}
	}
	return
}

func (s *Service) isUpWholeAct(sid int64) int64 {
	if categoryID, ok := s.c.UpWholeActive.ExtraSids[strconv.FormatInt(sid, 10)]; ok {
		return categoryID
	}
	return 0
}

// upWholeLikeCheck .
func (s *Service) upWholeLikeCheck(c context.Context, sid, lid, mid, storyMaxAct int64) (left int64, isShare bool, err error) {
	var (
		sumScore, extraTimes, shareTimes int64
		shareSid                         int
		actList                          []*l.Action
	)
	eg := errgroup.WithContext(c)
	eg.Go(func(errCtx context.Context) (e error) {
		actList, e = s.upWholeEchoTimes(errCtx, sid, mid)
		return
	})
	// 获取额外次数
	eg.Go(func(errCtx context.Context) (e error) {
		extraTimes, e = s.upWholeExtraTimes(errCtx, sid, mid)
		return
	})
	// 获取额外分享次数
	eg.Go(func(errCtx context.Context) (e error) {
		shareTimes, e = s.upWholeExtraTimes(errCtx, s.c.UpWholeActive.ParentSid, mid)
		return
	})
	if err = eg.Wait(); err != nil {
		err = errors.Wrap(err, "eg.Wait()")
		return
	}
	for _, act := range actList {
		if act.Mid <= 0 { // 因为有空缓存
			continue
		}
		sumScore++
		if act.Lid == lid { // 一个up主只能投一次
			return
		}
	}
	left = storyMaxAct - sumScore + extraTimes
	if left > 0 {
		return
	}
	// 判断是否分享
	if shareTimes > 0 {
		if shareSid, err = s.dao.RiGet(c, upShareKey(mid, s.c.UpWholeActive.ParentSid)); err != nil {
			log.Errorc(c, "upWholeLikeCheck s.dao.RsGet mid:%d sid:%d err:%v", mid, s.c.UpWholeActive.ParentSid, err)
			err = xecode.RequestErr
			return
		}
		if shareSid == 0 {
			isShare = true
			left = s.c.UpWholeActive.ExtraNum
			return
		}
		if int64(shareSid) == sid { // 使用分享投票的要把分享次数加上
			left += shareTimes
		}
	}
	if left < 0 {
		left = 0
	}
	return
}

func upShareKey(mid, sid int64) string {
	return fmt.Sprintf("k_vote_%d_%d", mid, sid)
}

func (s *Service) upWholeVotes(ctx context.Context, sid, mid int64) (res map[int64]struct{}, err error) {
	var actList []*l.Action
	if actList, err = s.upWholeEchoTimes(ctx, sid, mid); err != nil {
		log.Errorc(ctx, "upWholeUsers s.upWholeEchoTimes sid(%d) mid(%d) error(%+v)", sid, mid, err)
		return
	}
	res = make(map[int64]struct{}, len(actList))
	for _, act := range actList {
		if act.Mid > 0 {
			res[act.Lid] = struct{}{}
		}
	}
	return
}

func (s *Service) actUpWholeLike(ctx context.Context, sid, lid, score, mid int64, subject *l.SubjectItem, likeItem *l.Item) (res int64, err error) {
	var (
		leftTime int64
		isShare  bool
	)
	if leftTime, isShare, err = s.upWholeLikeCheck(ctx, sid, lid, mid, subject.LikeLimit); err != nil {
		log.Errorc(ctx, "StoryKingAct actUpWholeLike s.upWholeLikeCheck(%d,%d,%d) error(%+v)", sid, lid, mid, err)
		return
	}
	if leftTime < score {
		if leftTime > 0 {
			score = leftTime
		} else {
			err = ecode.ActivityPollAlreadyVoted
			return
		}
	}
	if err = s.dao.SetRedisCache(ctx, sid, lid, score, likeItem.Type); err != nil {
		log.Error("StoryKingAct actUpWholeLike s.dao.SetRedisCache(%d,%d,%d) error(%+v)", sid, lid, score, err)
		return
	}
	likeActAdd := &l.Action{
		Lid:         lid,
		Mid:         mid,
		Sid:         sid,
		Action:      score,
		IPv6:        make([]byte, 0),
		ExtraAction: 0,
	}
	if IPv6 := net.ParseIP(metadata.String(ctx, metadata.RemoteIP)); IPv6 != nil {
		likeActAdd.IPv6 = IPv6
	}
	if res, err = s.dao.LikeActAdd(ctx, likeActAdd); err != nil {
		log.Errorc(ctx, "StoryKingAct actUpWholeLike s.dao.LikeActAdd(%+v) error(%+v)", likeActAdd, err)
		return
	}
	eg := errgroup.WithContext(ctx)
	eg.Go(func(ctx context.Context) error {
		appendErr := s.dao.AppendUpActionCache(ctx, sid, mid, likeActAdd)
		if appendErr != nil {
			log.Errorc(ctx, "StoryKingAct actUpWholeLike s.dao.AppendUpActionCache(%d,%d) error(%+v)", sid, mid, err)
		}
		return appendErr
	})
	eg.Go(func(ctx context.Context) error {
		if isShare {
			strSid := strconv.FormatInt(sid, 10)
			shareErr := s.dao.RsSet(ctx, upShareKey(mid, s.c.UpWholeActive.ParentSid), strSid)
			if shareErr != nil {
				log.Errorc(ctx, "StoryKingAct actUpWholeLike s.dao.RsSet(%d,%d) error(%+v)", mid, s.c.UpWholeActive.ParentSid, err)
				s.cache.Do(ctx, func(ctx context.Context) {
					retry(func() error {
						return s.dao.RsSet(ctx, upShareKey(mid, s.c.UpWholeActive.ParentSid), strSid)
					})
				})
			}
		}
		return nil
	})
	if err = eg.Wait(); err != nil {
		log.Error("StoryKingAct actUpWholeLikes eg.Wait error(%v)", err)
		return
	}
	return
}

func (s *Service) actDailyLike(c context.Context, sid, lid, score, mid, nowTs int64, subject *l.SubjectItem, likeItem *l.Item, memberRly *accapi.ProfileReply) (res int64, totalScore int64, err error) {
	var (
		leftTime, maxLeft, extraAction int64
	)
	if leftTime, maxLeft, err = s.storyLikeCheck(c, sid, lid, mid, subject.DailyLikeLimit, subject.DailySingleLikeLimit); err != nil {
		log.Error(" s.storyLikeCheck(%d,%d,%d) error(%+v)", sid, lid, mid, err)
		return
	}
	if leftTime < score {
		if leftTime > 0 {
			score = leftTime
		} else {
			err = ecode.ActivityOverDailyScore
			return
		}
	}
	// vip extra action
	if !subject.AttrFlag(l.FLAGMONTHSCORE) {
		if memberRly.Profile.Vip.IsValid() && !memberRly.Profile.Vip.IsAnnual() {
			extraAction = score * subject.MonthScore
		}
	}
	if !subject.AttrFlag(l.FLAGYEARSCORE) {
		if memberRly.Profile.Vip.IsAnnual() {
			extraAction = score * subject.YearScore
		}
	}
	totalScore = score + extraAction
	if err = s.dao.SetRedisCache(c, sid, lid, totalScore, likeItem.Type); err != nil {
		log.Error("s.dao.SetRedisCache(%d,%d,%d) error(%+v)", sid, lid, totalScore, err)
		return
	}
	likeActAdd := &l.Action{
		Lid:         lid,
		Mid:         mid,
		Sid:         sid,
		Action:      int64(score),
		IPv6:        make([]byte, 0),
		ExtraAction: extraAction,
	}
	if IPv6 := net.ParseIP(metadata.String(c, metadata.RemoteIP)); IPv6 != nil {
		likeActAdd.IPv6 = IPv6
	}
	if res, err = s.dao.LikeActAdd(c, likeActAdd); err != nil {
		log.Error("s.dao.LikeActAdd(%v) error(%+v)", likeActAdd, err)
		return
	}
	s.storyLikeActSet(c, sid, lid, mid, score)
	if sid == s.c.Taaf.SidV2 {
		s.cache.Do(c, func(ctx context.Context) {
			s.dao.AppendCacheLikeActLids(ctx, sid, &l.LidItem{Lid: lid, ActTime: xtime.Time(nowTs)}, mid)
		})
	}
	// bdf 最后一次投票加抽奖机会
	isBdf := func() bool {
		if s.c.BdfOnline == nil {
			return false
		}
		for _, v := range s.c.BdfOnline.Sids {
			if v == sid {
				return true
			}
		}
		return false
	}()
	if isBdf && score >= maxLeft {
		s.cache.Do(c, func(ctx context.Context) {
			orderNo := strconv.FormatInt(mid, 10) + strconv.FormatInt(sid, 10) + strconv.FormatInt(nowTs, 10)
			if bdfErr := s.AddLotteryTimes(ctx, s.c.BdfOnline.LotterySid, mid, s.c.BdfOnline.LotteryCid, _other, 0, orderNo, false); bdfErr != nil {
				log.Error("Bdf AddLotteryTimes sid:%s mid:%d error(%v)", s.c.BdfOnline.LotterySid, mid, bdfErr)
			}
		})
	}
	return
}

// StoryKingAct .
func (s *Service) StoryKingAct(c context.Context, p *l.ParamStoryKingAct, mid int64) (res map[string]int64, err error) {
	var (
		subject           *l.SubjectItem
		likeItem          *l.Item
		memberRly         *accapi.ProfileReply
		subErr            error
		likeErr           error
		actID, totalScore int64
	)
	eg := errgroup.WithContext(c)
	eg.Go(func(errCtx context.Context) error {
		subject, subErr = s.dao.ActSubject(errCtx, p.Sid)
		return subErr
	})
	eg.Go(func(errCtx context.Context) error {
		likeItem, likeErr = s.dao.Like(errCtx, p.Lid)
		return likeErr
	})
	if err = eg.Wait(); err != nil {
		log.Error("eg.Wait error(%v)", err)
		return
	}
	if subject.ID == 0 || subject.Type != l.STORYKING {
		err = ecode.ActivityHasOffLine
		return
	}
	if likeItem.ID == 0 || likeItem.Sid != p.Sid {
		err = ecode.ActivityLikeHasOffLine
		return
	}
	if memberRly, err = s.accClient.Profile3(c, &accapi.MidReq{Mid: mid}); err != nil {
		log.Error(" s.acc.Profile3(c,&accmdl.ArgMid{Mid:%d}) error(%v)", mid, err)
		return
	}
	if err = s.judgeUser(c, subject, memberRly.Profile); err != nil {
		return
	}
	nowTs := time.Now().Unix()
	if int64(subject.Lstime) >= nowTs {
		err = ecode.ActivityLikeNotStart
		return
	}
	if int64(subject.Letime) <= nowTs {
		err = ecode.ActivityLikeHasEnd
		return
	}
	categoryID := s.isUpWholeAct(p.Sid)
	if categoryID > 0 {
		if err = s.upWholeVote(c, mid, likeItem.Wid, categoryID, p, memberRly.Profile.Name, memberRly.Profile.Level); err != nil {
			log.Errorc(c, "StoryKingAct s.upWholeVote mid(%d) category(%d) lid(%d)  error(%+v)", mid, categoryID, p.Lid, err)
			return
		}
		if actID, err = s.actUpWholeLike(c, p.Sid, p.Lid, p.Score, mid, subject, likeItem); err != nil {
			log.Errorc(c, "StoryKingAct s.actUpWholeLike mid(%d) category(%d) lid(%d) error(%+v)", mid, categoryID, p.Lid, err)
			return
		}
		totalScore = p.Score
	} else {
		if actID, totalScore, err = s.actDailyLike(c, p.Sid, p.Lid, p.Score, mid, nowTs, subject, likeItem, memberRly); err != nil {
			return
		}
	}
	if p.Token != "" {
		var tokenInfo *l.ExtendTokenDetail
		if tokenInfo, err = s.dao.LikeExtendInfo(c, p.Sid, p.Token); err != nil {
			log.Error("s.dao.LikeExtendToken(%d,%s) error(%v)", p.Sid, p.Token, err)
			return
		}
		if tokenInfo != nil && tokenInfo.Mid != 0 && tokenInfo.Max != 0 && tokenInfo.Mid != mid {
			// 加额外次数
			s.cache.Do(c, func(ctx context.Context) {
				// 根据token加额外次数
				s.addExtraTimes(ctx, tokenInfo)
			})
		}
	}
	res = make(map[string]int64, 2)
	res["act_id"] = actID
	res["score"] = totalScore
	return
}

func (s *Service) upWholeVote(c context.Context, mid, likeWid, categoryID int64, p *l.ParamStoryKingAct, userName string, level int32) (err error) {
	var (
		risk    bool
		riskErr error
	)
	if mid != likeWid {
		relReq := &relmdl.RelationReq{
			Mid: mid,
			Fid: likeWid,
		}
		relRsp, e := s.relClient.Relation(c, relReq)
		if e != nil {
			log.Errorc(c, "StoryKingAct s.relClient.Relation.mid(%v) fid(%d) error(%v)", mid, likeWid, e)
			err = ecode.ActivityLotteryNetWorkError
			return
		}
		if relRsp == nil || relRsp.Attribute == 0 || relRsp.Attribute >= 128 {
			err = ecode.ActivityUpFollowErr
			if relRsp != nil {
				log.Infoc(c, "StoryKingAct s.relClient.Relation.mid(%v) fid(%d) error(%v) relRsp(%v)", mid, likeWid, err, relRsp.Attribute)
			} else {
				log.Infoc(c, "StoryKingAct s.relClient.Relation.mid(%v) fid(%d) error(%v) relRsp is nil", mid, likeWid, err)
			}
			return
		}
	}
	if risk, riskErr = s.checkRiskCommon(c, mid, likeWid, categoryID, p, userName, _riskUpActivityUID, level); riskErr != nil {
		log.Errorc(c, "StoryKingAct s.checkRisk mid(%d) upMid(%d) category(%d) lid(%d) error(%+v)", mid, likeWid, categoryID, p.Lid, riskErr)
	}
	if risk {
		err = riskErr
		log.Errorc(c, "StoryKingAct s.checkRisk mid(%d) category(%d) lid(%d) risk is true error(%+v)", mid, categoryID, p.Lid, err)
		return
	}
	return
}

func (s *Service) checkRiskCommon(ctx context.Context, mid, upMid, categoryID int64, params *l.ParamStoryKingAct, userName, activityUid string, level int32) (res bool, err error) {
	otherEventCtx := &l.UpVoteEventCtx{
		Action:       _riskUpAction,
		Mid:          mid,
		ActivityUid:  activityUid,
		UpMid:        upMid,
		Content:      userName,
		UpCategoryID: categoryID,
		Buvid:        params.Buvid,
		Ip:           params.IP,
		Platform:     params.Platform,
		Ctime:        time.Now().Format("2006-01-02 15:04:05"),
		Api:          _riskUpApi,
		Origin:       params.Origin,
		UserAgent:    params.UA,
		Build:        params.Build,
		Referer:      params.Referer,
		MobiApp:      params.MobiApp,
		Level:        level,
	}
	if res, err = s.silverDao.RuleCheckCommon(ctx, _riskUpAction, otherEventCtx); err != nil {
		log.Errorc(ctx, "StoryKingAct checkRisk mid(%d) otherEventCtx(%+v) error(%+v)", mid, otherEventCtx, err)
	}
	return
}

// UpAddVoteTime .
func (s *Service) UpAddVoteTime(c context.Context, sid, mid int64) (res int64, err error) {
	var (
		times   int64
		subject *l.SubjectItem
	)
	if s.c.UpWholeActive.ParentSid != sid {
		err = xecode.RequestErr
		log.Errorc(c, "UpAddVoteTime s.isUpWholeAct(%d,%d) error(%v)", sid, mid, err)
		return
	}
	if subject, err = s.dao.ActSubject(c, sid); err != nil {
		log.Errorc(c, "UpAddVoteTime s.dao.ActSubject(%d,%d) error(%v)", sid, mid, err)
		return
	}
	if subject.ID == 0 || subject.Type != l.STORYKING {
		err = ecode.ActivityHasOffLine
		return
	}
	nowTs := time.Now().Unix()
	if int64(subject.Lstime) >= nowTs {
		log.Infoc(c, "UpAddVoteTime s.dao.ActSubject(%d,%d) act not start", sid, mid)
		return
	}
	if int64(subject.Letime) <= nowTs {
		log.Infoc(c, "UpAddVoteTime s.dao.ActSubject(%d,%d) act is end", sid, mid)
		return
	}
	// 查询额外增加次数
	if times, err = s.upVoteExtraTimes(c, sid, mid); err != nil {
		log.Errorc(c, "UpAddVoteTime s.upVoteExtraTimes(%d,%d) error(%v)", sid, mid, err)
		return
	}
	if times >= subject.LikeLimit {
		err = ecode.ActivityAlreadyShare
		return
	}
	if err = s.upAddTimes(c, sid, mid); err != nil {
		log.Error("UpAddVoteTime s.upAddTimes(%d,%d) error(%v)", sid, mid, err)
	}
	return
}

func (s *Service) upAddTimes(c context.Context, sid, mid int64) (err error) {
	var id int64
	if id, err = s.dao.IncrLikeExtraTimes(c, sid, mid, s.c.UpWholeActive.ExtraNum); err != nil {
		log.Error("UpAddVoteTime s.dao.IncrLikeExtraTimes(%d,%d) error(%v)", sid, mid, err)
		return
	}
	if err = s.dao.AddLikeExtraTimes(c, sid, mid, &l.ExtraTimesDetail{ID: id, Sid: sid, Mid: mid, Num: int(s.c.UpWholeActive.ExtraNum), Ctime: xtime.Time(time.Now().Unix())}); err != nil {
		log.Error("UpAddVoteTime s.dao.AddLikeExtraTimes(%d,%d) error(%v)", sid, mid, err)
	}
	return
}

// UpAddVoteTime .
func (s *Service) UpVoteAppendTimes(ctx context.Context, sid, lid, isAdd int64) (err error) {
	var (
		list  []*l.Action
		check bool
		nxKey = fmt.Sprintf("know_times_%d_%d", sid, lid)
	)
	if list, err = s.retryUpWholeUsers(ctx, sid, lid); err != nil {
		log.Errorc(ctx, "UpVoteAppendTimes s.retryUpWholeUsers(%d,%d) error(%v)", sid, lid, err)
		return
	}
	log.Infoc(ctx, "UpVoteAppendTimes s.retryUpWholeUsers(%d,%d) count(%v)", sid, lid, len(list))
	if isAdd == 0 {
		log.Infoc(ctx, "UpVoteAppendTimes s.retryUpWholeUsers(%d,%d) isAdd false", sid, lid)
		return
	}
	if check, err = s.dao.RsSetNX(ctx, nxKey, 8640000); err != nil || !check {
		log.Infoc(ctx, "UpVoteAppendTimes s.dao.RsSetNX(%d,%d) check(%v) error(%v)", sid, lid, check, err)
		err = ecode.ActivityRapid
		return
	}
	for _, act := range list {
		if err = s.upAddTimes(ctx, sid, act.Mid); err != nil {
			log.Errorc(ctx, "UpAddVoteTime s.upAddTimes(%d,%d) error(%v)", sid, act.Mid, err)
		}
		log.Infoc(ctx, "UpAddVoteTime s.upAddTimes(%d,%d) error(%v)", sid, act.Mid, err)
	}
	return
}

// UpAppendTime .
func (s *Service) UpAppendTime(ctx context.Context, sid, mid int64) (err error) {
	var (
		check bool
		nxKey = fmt.Sprintf("know_time_%d_%d", sid, mid)
	)
	if check, err = s.dao.RsSetNX(ctx, nxKey, 86400); err != nil || !check {
		log.Infoc(ctx, "UpAppendTime s.dao.RsSetNX(%d,%d) check(%v) error(%v)", sid, mid, check, err)
		err = ecode.ActivityRapid
		return
	}
	if err = s.upAddTimes(ctx, sid, mid); err != nil {
		log.Errorc(ctx, "UpAppendTime s.upAddTimes(%d,%d) error(%v)", sid, mid, err)
	}
	return
}

func (s *Service) upVoteExtraTimes(c context.Context, sid, mid int64) (res int64, err error) {
	var (
		list []*l.ExtraTimesDetail
	)
	if list, err = s.dao.CacheLikeExtraTimes(c, sid, mid, 0, -1); err != nil {
		log.Error("upVoteExtraTimes s.dao.CacheLikeExtendTimes(%d,%d) error(%v)", sid, mid, err)
		return
	}
	if len(list) == 0 {
		if list, err = s.dao.RawLikeExtraTimes(c, sid, mid); err != nil {
			log.Error("upVoteExtraTimes s.dao.RawLikeExtraTimes(%d,%d) error(%v)", sid, mid, err)
			return
		}
		miss := make([]*l.ExtraTimesDetail, 0)
		for _, v := range list {
			res = res + int64(v.Num)
		}
		if err = s.dao.AddCacheLikeExtraTimes(c, sid, mid, miss); err != nil {
			log.Error("upVoteExtraTimes s.dao.AddCacheLikeExtendTimes(%d,%d,%v) error(%v)", sid, mid, miss, err)
			return
		}
		return
	}
	for _, v := range list {
		res = res + int64(v.Num)
	}
	return
}

// StoryKingLeftTime .
func (s *Service) StoryKingLeftTime(c context.Context, sid, mid int64) (res int64, err error) {
	var (
		subject   *l.SubjectItem
		memberRly *accapi.ProfileReply
	)
	if subject, err = s.dao.ActSubject(c, sid); err != nil {
		return
	}
	if subject.ID == 0 || (subject.Type != l.STORYKING && !subject.IsDailyLike()) {
		err = ecode.ActivityHasOffLine
		return
	}
	if memberRly, err = s.accClient.Profile3(c, &accapi.MidReq{Mid: mid}); err != nil {
		log.Error(" s.acc.Profile3(c,&accmdl.ArgMid{Mid:%d}) error(%v)", mid, err)
		return
	}
	if err = s.simpleJudge(c, subject, memberRly.Profile); err != nil {
		if !subject.IsDailyLike() {
			err = nil
			res = 0
		}
		return
	}
	nowTime := time.Now().Unix()
	if int64(subject.Lstime) >= nowTime || int64(subject.Letime) <= nowTime {
		res = 0
		return
	}
	if res, err = s.storySumUsed(c, sid, mid); err != nil {
		log.Error("s.storySumUsed(%d,%d) error(%+v)", sid, mid, err)
		return
	}
	// 获取额外次数
	extendScore := int64(0)
	// 文豪活动不需要额外次数
	if sid != s.c.GiantV4.Sid {
		if extendScore, err = s.storyExtraTimes(c, sid, mid); err != nil {
			log.Error("s.storyExtraTimes(%d,%d) error(%+v)", sid, mid, err)
			return
		}
	}
	res = subject.DailyLikeLimit - res + extendScore
	if res < 0 {
		res = 0
	}
	return
}

// UpList .
func (s *Service) UpList(c context.Context, p *l.ParamList, mid int64) (res *l.ListInfo, err error) {
	if p.Sid == s.c.Taaf.Sid && p.Version == 0 {
		res = s.taafLikes
		return
	}
	if p.Sid == s.c.Timemachine.FlagSid {
		res = s.timemachineLikes
		return
	}
	switch p.Type {
	case like.EsOrderLikes, like.EsOrderCoin, like.EsOrderReply, like.EsOrderShare, like.EsOrderClick, like.EsOrderDm, like.EsOrderFav:
		res, err = s.EsList(c, p, mid)
	case like.ActOrderCtime, like.ActOrderLike, like.ActOrderRandom, like.ActOrderStochastic:
		res, err = s.StoryKingList(c, p, mid)
	default:
		err = errors.New("type error")
	}
	return
}

// LikeActUpList .
func (s *Service) LikeActUpList(ctx context.Context, sid int64, mid int64) (res *l.LikeActList, err error) {
	var (
		upList   *l.ListInfo
		upErr    error
		lidItems []*l.LidItem
		likeErr  error
		lidMap   map[int64]struct{}
	)
	p := &l.ParamList{
		Sid:     sid,
		Pn:      1,
		Ps:      s.c.Amusement.ListPs,
		Version: 100,
	}
	res = &l.LikeActList{
		LikeList: _emptyList,
		List:     _emptyList,
	}
	eg := errgroup.WithContext(ctx)
	eg.Go(func(ctx context.Context) error {
		if upList, upErr = s.StoryKingList(ctx, p, mid); upErr != nil {
			log.Errorc(ctx, "LikeActUpList s.StoryKingList sid(%d) mid(%d) error(%+v)", p, sid, mid, upErr)
		}
		return upErr
	})
	eg.Go(func(ctx context.Context) error {
		if mid > 0 {
			if lidItems, likeErr = s.dao.LikeActLids(ctx, sid, mid); likeErr != nil {
				log.Errorc(ctx, "LikeActUpList s.dao.LikeActLids sid(%d) mid(%d) error(%+v)", p, sid, mid, likeErr)
			}
			lidMap = make(map[int64]struct{}, 3)
			for _, item := range lidItems {
				if item.Action > 0 {
					lidMap[item.Lid] = struct{}{}
				}
			}
		}
		return likeErr
	})
	if err = eg.Wait(); err != nil {
		log.Errorc(ctx, "LikeActUpList errGroup sid(%d) mid(%d) error(%+v)", p, sid, mid, err)
		return
	}
	if upList == nil {
		return
	}
	res.ShowVote = upList.ShowVote
	if len(upList.List) == 0 {
		return
	}
	likeListMap := make(map[int64]*l.List, 3)
	for _, listValue := range upList.List {
		if listValue == nil {
			continue
		}
		if mid > 0 {
			if _, ok := lidMap[listValue.ID]; ok {
				likeListMap[listValue.ID] = listValue
				continue
			}
		}
		res.List = append(res.List, listValue)
	}
	if mid > 0 && len(likeListMap) > 0 {
		for _, lidItem := range lidItems {
			if lidItem.Action <= 0 {
				continue
			}
			if listValue, ok := likeListMap[lidItem.Lid]; ok {
				res.LikeList = append(res.LikeList, listValue)
			}
		}
	}
	return
}

// UpListRelation .
func (s *Service) UpListRelation(c context.Context, sid, mid int64) (reply l.ActKnowledgeRes, err error) {
	var (
		subject     *l.SubjectItem          // 活动基本信息
		infoRes     map[int64]*l.GetMIDInfo // 获取非top3 up主mid信息
		top3InfoRes map[int64]*l.GetMIDInfo // 获取top3   up主mid信息
	)
	reply.ActKnowledgeDetailList = make([]*l.GetActKnowledgeDetail, 0)
	if subject, err = s.dao.ActSubject(c, sid); err != nil {
		log.Errorc(c, "UpListRelation s.dao.ActSubject sid(%d) mid(%d) error(%+v)", sid, mid, err)
		return
	}
	if subject.ID <= 0 {
		err = ecode.ActivityHasOffLine
		return
	}

	if items, ok := s.knowledge[sid]; ok {
		var (
			upsMIDCollection []int64          // 赛道up主mid列表
			top3MIDs         []int64          // 赛道Top3 up主的mid
			followRes        []*l.FollowReply // 关注了哪些UP主详细信息
			followUpMIDs     []int64          // 关注了哪些UP的mid
		)
		voted := make(map[int64]struct{})
		upsMIDInfo2Map := make(map[int64]*l.LIDWithVote) // 转换格式 支持(O1)复杂度查询
		upsLIDInfo2Map := make(map[int64]*l.LIDWithVote) // 转换格式 支持(O1)复杂度查询
		getFollowTime := make(map[int64]int64)           // 整理被关注人的关注时间

		for index, item := range items {
			// 获取前三名up主id
			if index >= 0 && index < 3 {
				top3MIDs = append(top3MIDs, item.Wid)
			}
			upsMIDCollection = append(upsMIDCollection, item.Wid)
			upsMIDInfo2Map[item.Wid] = item
			upsLIDInfo2Map[item.ID] = item
		}

		if len(items) > 0 {
			eg := errgroup.WithContext(c)
			eg.Go(func(c context.Context) (e error) {
				// 查询关注关系
				if mid > 0 {
					followRes, err = s.GetUpsRelationData(c, mid, upsMIDCollection)
					if err != nil {
						log.Errorc(c, "UpListRelation s.GetUpsRelationData sid(%d) mid(%d) error(%+v)", sid, mid, err)
						err = ecode.ActivityGetAccRelationGRPCErr
						return
					}
				}
				return
			})
			eg.Go(func(c context.Context) (e error) {
				// 获取本次请求的mid给哪些up主投过票
				if mid > 0 {
					voted, err = s.upWholeVotes(c, sid, mid)
					if err != nil {
						log.Errorc(c, "UpListRelation s.upWholeVotes sid(%d) mid(%d) error(%+v)", sid, mid, err)
						err = ecode.ActivityGetAccRelationGRPCErr
						return
					}
				}
				return
			})
			if err = eg.Wait(); err != nil {
				err = errors.Wrap(err, "eg.Wait()")
				log.Errorc(c, "UpListRelation actKnowledge eg err :%v", err)
				return
			}

			for _, v := range followRes {
				followUpMIDs = append(followUpMIDs, v.MID)
				getFollowTime[v.MID] = v.MTime.Time().Unix()
			}

			// 如果本次请求用户也在排名里面 查询up主需要加进去
			if _, ok := upsMIDInfo2Map[mid]; ok {
				followUpMIDs = append(followUpMIDs, mid)
			}

			// 如果给这个up主投过票 但是没在关注关系里面 就证明取关了 投票后取关的话 也要展现出来
			// 根据返回的lid查询用户的mid 大于0必然是登录用户
			if len(voted) > 0 {
				for lid := range voted {
					// 通过lid获取mid
					if v, ok := upsLIDInfo2Map[lid]; ok {
						upMid := v.Wid
						in := false
						// 如果不在这里面 加进去 并且关注时间设置为0
						for _, followUpMID := range followUpMIDs {
							if followUpMID == upMid {
								in = true
							}
						}
						if in == false {
							followUpMIDs = append(followUpMIDs, upMid)
							getFollowTime[upMid] = 0
						}
					}
				}
			}

			eg = errgroup.WithContext(c)
			eg.Go(func(c context.Context) (e error) {
				// 查询关注的up主信息
				if mid > 0 {
					infoRes, err = s.GetUpsDetailInfo(c, followUpMIDs)
					if err != nil {
						log.Errorc(c, "UpListRelation s.GetUpsDetailInfo sid(%d) mid(%d) count(%d) data(%v) error(%+v)", sid, mid, len(followUpMIDs), followUpMIDs, err)
						err = ecode.ActivityGetAccRelationGRPCErr
						return
					}
				}
				return
			})
			eg.Go(func(c context.Context) (e error) {
				// 查询前三名信息
				top3InfoRes, err = s.GetUpsDetailInfo(c, top3MIDs)
				if err != nil {
					log.Errorc(c, "UpListRelation s.GetUpsDetailInfo sid(%d) mid(%d) count(%d) data(%v) error(%+v)", sid, mid, len(top3InfoRes), top3MIDs, err)
					err = ecode.ActivityGetAccRelationGRPCErr
					return
				}
				return
			})
			if err = eg.Wait(); err != nil {
				err = errors.Wrap(err, "eg.Wait()")
				log.Errorc(c, "UpListRelation actKnowledge eg err :%v", err)
				return
			}
			// 存在关注关系
			if len(infoRes) > 0 {
				// 循环关注关系，将一些数据放到最终返回结果里面
				for _, v := range infoRes {
					voteInfo, ok1 := upsMIDInfo2Map[v.Mid]  // 投票信息
					upInfo, ok2 := infoRes[v.Mid]           // up主信息
					followTime, ok3 := getFollowTime[v.Mid] // 关注关系最后的关注时间
					if ok1 == false || ok2 == false {
						log.Warnc(c, "UpListRelation Can`t Get User Info ok1:%v ok2:%v", ok1, ok2)
						continue
					}
					eachDetail := &l.GetActKnowledgeDetail{
						ID:         voteInfo.ID,
						MID:        v.Mid,
						Name:       upInfo.Name,
						Face:       upInfo.Face,
						FollowTime: 0,
						VoteNum:    voteInfo.Vote,
						OrderNum:   voteInfo.Order,
						IsUp:       0,
						IsSelect:   0,
					}
					if ok3 {
						eachDetail.FollowTime = followTime
					}
					if eachDetail.OrderNum == 1 || eachDetail.OrderNum == 2 || eachDetail.OrderNum == 3 {
						eachDetail.VoteNum = 0
					}
					if v.Mid == mid {
						eachDetail.IsUp = 1
					}
					// 该用户给此up主投票过
					if _, ok := voted[voteInfo.ID]; ok {
						eachDetail.IsSelect = 1
					}
					reply.ActKnowledgeDetailList = append(reply.ActKnowledgeDetailList, eachDetail)
				}
			}
			// 前三名数据展示
			if len(top3InfoRes) > 0 {
				for _, v := range top3InfoRes {
					voteInfo, ok1 := upsMIDInfo2Map[v.Mid] // 投票信息
					upInfo, ok2 := top3InfoRes[v.Mid]      // up主信息
					if ok1 == false || ok2 == false {
						log.Warnc(c, "UpListRelation Can`t Get User Info ok1:%v ok2:%v", ok1, ok2)
						continue
					}
					eachDetail := &l.GetActKnowledgeDetail{
						ID:         voteInfo.ID,
						MID:        v.Mid,
						Name:       upInfo.Name,
						Face:       upInfo.Face,
						FollowTime: 0,
						VoteNum:    0,
						OrderNum:   voteInfo.Order,
						IsUp:       0,
						IsSelect:   0,
					}
					if v.Mid == mid {
						eachDetail.IsUp = 1
					}
					reply.ActKnowledgeTop3List = append(reply.ActKnowledgeTop3List, eachDetail)
				}
			}
		}
	}

	if len(reply.ActKnowledgeDetailList) > 0 {
		sort.Sort(reply.ActKnowledgeDetailList)
	}

	// 获取用户是否投票过
	if mid > 0 {
		reply.Left, _, err = s.upWholeLikeCheck(c, sid, -1, mid, subject.LikeLimit)
		if err != nil {
			reply.Left = 0
		}
	}

	return
}

// EsList .
func (s *Service) EsList(c context.Context, p *l.ParamList, mid int64) (res *l.ListInfo, err error) {
	var (
		subject *l.SubjectItem
	)
	if subject, err = s.dao.ActSubject(c, p.Sid); err != nil {
		return
	}
	if subject.ID == 0 {
		err = ecode.ActivityHasOffLine
		return
	}
	if res, err = s.dao.ListFromES(c, p.Sid, p.Type, p.Ps, p.Pn, 0, p.Zone); err != nil {
		log.Error("s.dao.ListFromES(%d) error(%+v)", p.Sid, err)
		return
	}
	if res == nil || len(res.List) == 0 {
		return
	}
	if err = s.getContent(c, res.List, subject, subject.Type, mid, p.Type); err != nil {
		log.Error("s.getContent(%d) error(%v)", p.Sid, err)
	}
	return
}

// Slider .
func (s *Service) Slider(c context.Context, lids []int64, sid, mid int64) (slider *l.Slider, err error) {
	var (
		subject  *l.SubjectItem
		items    map[int64]*l.Item
		tLids    []int64
		likeAct  map[int64]int64
		likedMap map[int64]int64
		res      []*l.List
	)
	eg := errgroup.WithContext(c)
	eg.Go(func(errCtx context.Context) (e error) {
		subject, e = s.dao.ActSubject(errCtx, sid)
		return
	})
	eg.Go(func(errCtx context.Context) (e error) {
		if items, e = s.dao.Likes(errCtx, lids); e != nil {
			log.Error("s.dao.Like(%v) error(%v)", lids, e)
		}
		return
	})
	if err = eg.Wait(); err != nil {
		return
	}
	if subject.ID == 0 {
		err = ecode.ActivityHasOffLine
		return
	}
	res = make([]*l.List, 0)
	tLids = make([]int64, 0)
	for _, v := range items {
		if v.Sid == sid {
			temp := &l.List{Item: v}
			res = append(res, temp)
			tLids = append(tLids, v.ID)
		}
	}
	if len(tLids) == 0 {
		return
	}
	egTwo := errgroup.WithContext(c)
	egTwo.Go(func(errC context.Context) (e error) {
		if e = s.getContent(errC, res, subject, subject.Type, mid, like.ActOrderLike); e != nil {
			log.Error("s.getContent(%d) error(%v)", sid, e)
		}
		return
	})
	if subject.AttrFlag(l.FLAGRANKCLOSE) {
		egTwo.Go(func(errC context.Context) error {
			likeAct, _ = s.dao.LikeActLidCounts(errC, tLids)
			return nil
		})
	}
	if mid > 0 {
		egTwo.Go(func(errC context.Context) error {
			likedMap, _ = s.LikedInfos(errC, tLids, subject, mid)
			return nil
		})
	}
	if err = egTwo.Wait(); err != nil {
		return
	}
	for _, val := range res {
		if _, ok := likeAct[val.Item.ID]; ok {
			val.Like = likeAct[val.Item.ID]
		}
		if _, k := likedMap[val.Item.ID]; k {
			val.Liked = likedMap[val.Item.ID]
		}
	}
	slider = &l.Slider{List: res}
	return
}

// OneItem .
func (s *Service) OneItem(c context.Context, lid, sid, mid int64) (res *l.List, err error) {
	var (
		subject  *l.SubjectItem
		list     = &l.List{}
		likeAct  map[int64]int64
		likedMap map[int64]int64
	)
	eg := errgroup.WithContext(c)
	eg.Go(func(errCtx context.Context) (e error) {
		subject, e = s.dao.ActSubject(errCtx, sid)
		return
	})
	list = &l.List{}
	eg.Go(func(errCtx context.Context) (e error) {
		if list.Item, e = s.dao.Like(errCtx, lid); e != nil {
			log.Error("s.dao.Like(%d) error(%v)", lid, e)
		}
		return
	})
	if err = eg.Wait(); err != nil {
		return
	}
	if subject.ID == 0 {
		err = ecode.ActivityHasOffLine
		return
	}
	if list.Item.ID == 0 || sid != list.Item.Sid {
		return
	}
	if err = s.getContent(c, []*l.List{list}, subject, subject.Type, mid, like.ActOrderLike); err != nil {
		log.Error("s.getContent(%d) error(%v)", sid, err)
		return
	}
	egTwo := errgroup.WithContext(c)
	if subject.AttrFlag(l.FLAGRANKCLOSE) {
		egTwo.Go(func(errC context.Context) error {
			likeAct, _ = s.dao.LikeActLidCounts(errC, []int64{lid})
			return nil
		})
	}
	egTwo.Go(func(errC context.Context) error {
		likedMap, _ = s.LikedInfos(errC, []int64{lid}, subject, mid)
		return nil
	})
	egTwo.Wait()
	if _, ok := likeAct[lid]; ok {
		list.Like = likeAct[lid]
	}
	if _, k := likedMap[lid]; k {
		list.Liked = likedMap[lid]
	}
	res = list
	return
}

// LikeMyList .
func (s *Service) LikeMyList(c context.Context, sid, mid int64, ps, pn int) (res *l.ListInfo, err error) {
	var (
		subject *l.SubjectItem
		lids    []int64
	)
	if subject, err = s.dao.ActSubject(c, sid); err != nil {
		return
	}
	//暂时只支持图片类型活动
	if subject.ID == 0 || (subject.Type != l.PICTURELIKE && subject.Type != l.PICTURE) {
		err = ecode.ActivityHasOffLine
		return
	}
	if res, err = s.dao.MyListFromEs(c, sid, mid, "id", ps, pn, 0); err != nil {
		log.Error("s.dao.MyListFromEs(%d,%d) error(%v)", sid, mid, err)
		return
	}
	if res == nil || len(res.List) == 0 {
		return
	}
	lids = make([]int64, 0, len(res.List))
	for _, v := range res.List {
		lids = append(lids, v.ID)
	}
	if err = s.getContent(c, res.List, subject, subject.Type, mid, like.ActOrderLike); err != nil {
		log.Error("s.getContent() error(%+v)", err)
		return
	}
	if subject.AttrFlag(l.FLAGRANKCLOSE) {
		likeAct, _ := s.dao.LikeActLidCounts(c, lids)
		for _, v := range res.List {
			if _, ok := likeAct[v.ID]; ok {
				v.Like = likeAct[v.ID]
			}
		}
	}
	return
}

// ActLikes .
func (s *Service) ActLikes(c context.Context, a *l.ArgActLikes) (list *l.ActLikes, err error) {
	var (
		subject   *l.SubjectItem
		lids      []int64
		total     int64
		scores    map[int64]int64
		liked     map[int64]int64
		infos     []*l.LidLikeRes
		items     map[int64]*l.Item
		resObj    []*l.ItemObj
		hasScores = false
	)
	if subject, err = s.dao.ActSubject(c, a.Sid); err != nil {
		return
	}
	if subject.ID == 0 {
		err = ecode.ActivityHasOffLine
		return
	}
	disLike := subject.DisplayLike()
	var start, end int64
	// 兼容新老版本需求
	if a.Offset >= 0 {
		start = a.Offset
	} else {
		start = int64((a.Pn - 1) * a.Ps)
	}
	end = start + int64(a.Ps) - 1
	switch a.SortType {
	case api.ActOrderCtimeNum:
		if lids, err = s.dao.LikeCtime(c, a.Sid, a.Zone, start, end); err != nil {
			log.Error("ActLikes s.dao.LikeCtime(%d) error(%v)", a.Sid, err)
			return
		}
	case api.ActOrderStochasticNum:
		if lids, err = s.stochasticLids(c, a.Sid, a.Zone, a.Ps); err != nil {
			log.Error("ActLikes s.stochasticLids(%d) error(%v)", a.Sid, err)
			return
		}
	case api.ActOrderEsLikeNum:
		var esRly *l.EsLikesReply
		esRly, err = s.EsLikesIDs(c, a.Sid, a.Zone, start, end)
		if err != nil {
			log.Error("s.dao.EsLikesIDs(%d) error(%+v)", a.Sid, err)
			return
		}
		if esRly != nil {
			lids = esRly.Lids
			total = esRly.Count
		}
	default:
		if infos, err = s.dao.RedisCache(c, a.Sid, a.Zone, start, end); err != nil {
			log.Error("ActLikes s.dao.RedisCache(%d) error(%v)", a.Sid, err)
			return
		}
		lids = make([]int64, 0, len(infos))
		scores = make(map[int64]int64, len(infos))
		for _, v := range infos {
			lids = append(lids, v.Lid)
			if disLike {
				scores[v.Lid] = v.Score
			}
		}
		hasScores = true
	}
	lidsLen := len(lids)
	eg := errgroup.WithCancel(c)
	if lidsLen > 0 {
		eg.Go(func(ctx context.Context) (e error) {
			if items, e = s.dao.Likes(ctx, lids); e != nil {
				log.Error("s.dao.Likes(%v) error(%v)", lids, e)
			}
			return
		})
		if !hasScores && disLike {
			eg.Go(func(ctx context.Context) (e error) {
				if scores, e = s.dao.LikeActLidCounts(ctx, lids); e != nil {
					log.Error("s.dao.LikeActLidCounts(%v) error(%v)", lids, e)
					e = nil
				}
				return
			})
		}
		if a.Mid > 0 {
			eg.Go(func(ctx context.Context) (e error) {
				if liked, e = s.LikedInfos(ctx, lids, subject, a.Mid); e != nil {
					log.Error("s.LikedInfos(%v) error(%v)", lids, e)
					e = nil
				}
				return
			})
		}
	}
	if a.SortType != api.ActOrderEsLikeNum {
		eg.Go(func(ctx context.Context) (e error) {
			total, e = s.dao.LikeCount(ctx, a.Sid, a.Zone)
			if e != nil {
				log.Error("s.dao.EsTotal or LikeCount (%d) error(%v)", a.Sid, e)
				e = nil
			}
			return
		})
	}
	if err = eg.Wait(); err != nil {
		return
	}
	list = &l.ActLikes{Sub: subject}
	resObj = make([]*l.ItemObj, 0, lidsLen)
	for _, v := range lids {
		if _, ok := items[v]; ok {
			tmp := &l.ItemObj{Item: items[v]}
			if !disLike {
				tmp.Score = -1
			} else if _, k := scores[v]; k {
				tmp.Score = scores[v]
			}
			if _, o := liked[v]; o {
				tmp.HasLiked = liked[v]
			}
			resObj = append(resObj, tmp)
		}
	}
	if a.SortType == api.ActOrderStochasticNum {
		// 随机排序offset赋值0
		list.Offset = 0
	}
	list.Offset = a.Offset
	if a.Offset >= 0 {
		list.Offset = list.Offset + int64(lidsLen)
		if total > list.Offset {
			list.HasMore = 1
		}
	}
	list.Total = total
	list.List = resObj
	return
}

// EsLikesIDs .
func (s *Service) EsLikesIDs(c context.Context, sid, ltype, start, end int64) (reply *l.EsLikesReply, err error) {
	if start > 2000 {
		return
	}
	return s.dao.ActEsLikesIDs(c, sid, ltype, start, end)
}

// StoryKingList .
func (s *Service) StoryKingList(c context.Context, p *l.ParamList, mid int64) (res *l.ListInfo, err error) {
	var (
		subject  *l.SubjectItem
		likeList []*l.List
		total    int64
		lids     []int64
		liked    map[int64]int64
		showVote bool
	)
	if subject, err = s.dao.ActSubject(c, p.Sid); err != nil {
		return
	}
	if subject.ID == 0 {
		err = ecode.ActivityHasOffLine
		return
	}
	if subject.ID == s.c.GiantV4.Sid {
		subject.Type = l.ARTICLE
	}
	if p.Version == 100 {
		if s.c.Amusement.ShowVotes == 1 || subject.DisplayLike() {
			p.Type = like.ActOrderLike
			showVote = true
		} else {
			p.Type = like.ActOrderRandom
		}
	}
	switch p.Type {
	case like.ActOrderCtime:
		likeList, err = s.orderByCtime(c, p.Sid, p.Pn, p.Ps, subject, p.Zone)
	case like.ActOrderRandom:
		likeList, err = s.orderByRandom(c, p.Sid, p.Pn, p.Ps, subject, p.Zone)
	case like.ActOrderStochastic:
		likeList, err = s.orderByStochastic(c, p.Sid, p.Ps, subject, p.Zone)
	default:
		likeList, err = s.orderByLike(c, p.Sid, p.Pn, p.Ps, subject, p.Zone)
	}
	if err != nil {
		log.Error("s.orderBy(%s)(%d) error(%v)", p.Type, p.Sid, err)
		return
	}
	res = &l.ListInfo{ShowVote: showVote}
	if len(likeList) == 0 {
		return
	}
	lids = make([]int64, 0, len(likeList))
	for _, v := range likeList {
		lids = append(lids, v.ID)
	}
	eg := errgroup.WithContext(c)
	eg.Go(func(errCtx context.Context) error {
		if e := s.getContent(errCtx, likeList, subject, subject.Type, mid, p.Type); e != nil {
			log.Error("s.getContent(%d) error(%v)", p.Sid, e)
		}
		return nil
	})
	if p.Type == like.ActOrderRandom {
		eg.Go(func(errCtx context.Context) error {
			total, _ = s.dao.LikeRandomCount(errCtx, p.Sid, p.Zone)
			return nil
		})
	} else {
		eg.Go(func(errCtx context.Context) error {
			total, _ = s.dao.LikeCount(errCtx, p.Sid, p.Zone)
			return nil
		})
	}
	eg.Go(func(errCtx context.Context) error {
		liked, _ = s.LikedInfos(errCtx, lids, subject, mid)
		return nil
	})
	eg.Wait()
	for _, v := range likeList {
		if _, k := liked[v.ID]; k {
			v.Liked = liked[v.ID]
		}
	}
	if p.Sid == s.c.Rule.LimitArcSid && p.Type == like.ActOrderStochastic {
		var tmpList []*l.List
		for _, v := range likeList {
			if v.State == 0 {
				continue
			}
			tmpList = append(tmpList, v)
		}
		likeList = tmpList
		if len(likeList) > p.Ps {
			likeList = likeList[:p.Ps]
		}
		if len(likeList) < p.Ps {
			log.Warn("LimitArcSid not full")
		}
	}
	res = &l.ListInfo{List: likeList, Page: &l.Page{Size: p.Ps, Num: p.Pn, Total: total}, ShowVote: showVote}
	return
}

// LikedInfos liked[lid]score score>=0.
func (s *Service) LikedInfos(c context.Context, lids []int64, sub *l.SubjectItem, mid int64) (liked map[int64]int64, err error) {
	var (
		need    string
		likeAct map[int64]int
		upAct   map[int64]int64
	)
	if mid <= 0 {
		return
	}
	switch sub.Type {
	case l.PICTURE, l.PICTURELIKE, l.DRAWYOO, l.DRAWYOOLIKE, l.TEXT, l.TEXTLIKE, l.QUESTION, l.RESERVATION,
		l.VIDEOLIKE, l.VIDEO, l.VIDEO2, l.SMALLVIDEO, l.PHONEVIDEO, l.ONLINEVOTE:
		need = l.LIKETYPECOMMON
	case l.STORYKING:
		need = l.LIKETYPEUP
	default:
		need = l.LIKETYPENO
	}
	// 科学三分钟活动特殊处理
	if sub.IsDailyLike() {
		need = l.LIKETYPEUP
	}
	if need == l.LIKETYPECOMMON {
		if likeAct, err = s.dao.LikeActs(c, sub.ID, mid, lids); err != nil {
			log.Error("s.dao.LikeActs(%v,%d) error(%+v)", lids, mid, err)
			return
		}
		liked = make(map[int64]int64, len(likeAct))
		for k, v := range likeAct {
			if v == like.HasLike {
				liked[k] = like.HasLike
			}
		}
	} else if need == l.LIKETYPEUP {
		if upAct, err = s.dao.BatchStoryEachLikeSum(c, sub.ID, mid, lids); err != nil {
			log.Error("s.storyDailyUsed(%v,%d) error(%+v)", lids, mid, err)
			return
		}
		liked = make(map[int64]int64, len(upAct))
		for i, val := range upAct {
			if val >= sub.DailySingleLikeLimit {
				liked[i] = val
			}
		}
	}
	return
}

// getContent get likes extends 后期接入其他活动补充完善.
func (s *Service) getContent(c context.Context, list []*l.List, subject *l.SubjectItem, subType, mid int64, order string) (err error) {
	isBdf := func() bool {
		if s.c.BdfOnline == nil {
			return false
		}
		if len(list) == 0 {
			return false
		}
		for _, v := range s.c.BdfOnline.Sids {
			for _, item := range list {
				if item != nil && item.Sid == v {
					return true
				}
			}
		}
		return false
	}()
	if isBdf {
		subType = l.VIDEOLIKE
	}
	switch subType {
	case l.STORYKING:
		err = s.actContent(c, list, mid)
	case l.PICTURE, l.PICTURELIKE, l.DRAWYOO, l.DRAWYOOLIKE, l.TEXT, l.TEXTLIKE, l.QUESTION, l.RESERVATION:
		err = s.contentAccount(c, list, 0)
	case l.VIDEOLIKE, l.VIDEO, l.VIDEO2, l.SMALLVIDEO, l.PHONEVIDEO, l.ONLINEVOTE:
		// 过滤稿件
		// resList, err := s.filterList(c, list, subject)
		// if err != nil {
		// 	return err
		// }
		err = s.arcTag(c, list, order, mid)
		// list = resList
	case l.ARTICLE:
		err = s.article(c, list, mid)
	default:
		err = xecode.RequestErr
	}
	return
}

func (s *Service) filterList(c context.Context, list []*likemdl.List, subject *l.SubjectItem) (resList []*likemdl.List, err error) {
	aids := make([]int64, 0)
	resList = make([]*likemdl.List, 0)
	for _, v := range list {
		if v.Wid > 0 {
			aids = append(aids, v.Wid)
		}
	}
	if len(aids) > 0 {
		aidsMap, err := s.filterAid(c, aids, subject)
		if err != nil {
			return nil, err
		}
		for _, v := range list {
			if _, ok := aidsMap[v.Wid]; ok {
				resList = append(resList, v)
			}
		}
	}
	return resList, nil

}

func (s *Service) filterAid(c context.Context, aids []int64, subject *l.SubjectItem) (list map[int64]struct{}, err error) {
	aidFlowControlMap, err := s.archive.ArchiveFlowControl(c, aids)
	if err != nil {
		return
	}
	list = make(map[int64]struct{})
	for _, aid := range aids {
		if flowControl, ok := aidFlowControlMap[aid]; ok {
			if flowControl != nil && flowControl.ForbiddenItems != nil {
				for _, control := range flowControl.ForbiddenItems {
					if control != nil && control.Value == l.FlowControlYes {
						switch control.Key {
						case l.ArchiveNoRank:
							if subject.IsShieldRank() {
								continue
							}
						case l.ArchiveNoDynamic:
							if subject.IsShieldDynamic() {
								continue
							}
						case l.ArchiveNoRecommend:
							if subject.IsShieldRecommend() {
								continue
							}
						case l.ArchiveNoHot:
							if subject.IsShieldHot() {
								continue
							}
						case l.ArchiveNoFansDynamic:
							if subject.IsShieldFansDynamic() {
								continue
							}
						case l.ArchiveNoSearch:
							if subject.IsShieldSearch() {
								continue
							}
						case l.ArchiveNoOversea:
							if subject.IsShieldOversea() {
								continue
							}
						}
					}
				}
			}
		}
		list[aid] = struct{}{}
	}
	return
}

// article .
func (s *Service) article(c context.Context, list []*l.List, mid int64) (err error) {
	var (
		lt    = len(list)
		wids  = make([]int64, 0, lt)
		reply *artapi.ArticleMetasReply
		metas map[int64]*artmdl.Meta
		thump *thumbupapi.HasLikeReply
	)
	for _, v := range list {
		if v.Wid > 0 {
			wids = append(wids, v.Wid)
		}
	}
	eg := errgroup.WithContext(c)
	eg.Go(func(errCtx context.Context) (e error) {
		if reply, e = s.artClient.ArticleMetas(errCtx, &artapi.ArticleMetasReq{Ids: wids}); e != nil {
			log.Error(" s.artClient.ArticleMetas(%v) error(%+v)", wids, e)
			e = nil
		} else {
			metas = reply.Res
		}
		return
	})
	if mid > 0 {
		eg.Go(func(errCtx context.Context) (e error) {
			if thump, e = s.thumbupClient.HasLike(errCtx, &thumbupapi.HasLikeReq{Business: _businessLikeArt, MessageIds: wids, Mid: mid}); e != nil {
				log.Error("s.thumbup.HasLike(%d) error(%+v)", mid, e)
				e = nil
			}
			return
		})
	}
	if err = eg.Wait(); err != nil {
		log.Error("article wait error(%+v)", err)
		return
	}
	for _, v := range list {
		obj := &l.ArticleTag{}
		if v.Wid > 0 {
			if _, ok := metas[v.Wid]; ok {
				obj.Meta = metas[v.Wid]
			}
			if mid > 0 && thump != nil {
				if state, ok := thump.States[v.Wid]; ok && state != nil {
					obj.HasLike = int32(state.State)
				}
			}
		}
		v.Object = obj
	}
	return
}

// actContent get like_content and account info.
func (s *Service) contentAccount(c context.Context, list []*l.List, sid int64) (err error) {
	var (
		lt     = len(list)
		lids   = make([]int64, 0, lt)
		mids   = make([]int64, 0, lt)
		cont   map[int64]*l.LikeContent
		actRly *accapi.InfosReply
	)
	for _, v := range list {
		lids = append(lids, v.ID)
		if v.Mid > 0 {
			mids = append(mids, v.Mid)
		}
	}
	eg := errgroup.WithContext(c)
	eg.Go(func(errCtx context.Context) (e error) {
		if cont, e = s.dao.LikeContent(errCtx, lids); e != nil {
			log.Error("s.dao.LikeContent(%v) error(%+v)", lids, e)
		}
		return
	})
	if len(mids) > 0 {
		eg.Go(func(errCtx context.Context) (e error) {
			if actRly, e = s.accClient.Infos3(errCtx, &accapi.MidsReq{Mids: mids}); e != nil {
				log.Error("s.accClient.Infos3(%v) error(%+v)", mids, err)
			}
			return
		})
	}
	eg.Wait()
	for _, v := range list {
		obj := &l.ContentTag{}
		if _, ok := cont[v.ID]; ok {
			obj.Cont = cont[v.ID]
			if obj.Cont.Reply != "" {
				obj.Cont.Reply = template.HTMLEscapeString(obj.Cont.Reply)
			}
			if obj.Cont.Message != "" {
				obj.Cont.Message = template.HTMLEscapeString(obj.Cont.Message)
			}
			if v.State != 1 {
				obj.Cont.Image = ""
			}
			// bdf special
			if sid == s.c.Bdf.Sid {
				var (
					schCnt int64
					e      error
				)
				if obj.Cont.Link != "" {
					if schCnt, e = strconv.ParseInt(obj.Cont.Link, 10, 64); e != nil {
						log.Warn("bdf link(%s) error(%v)", obj.Cont.Link, e)
					}
				}
				v.Likes = v.Like*23 + schCnt*233
			}
		}
		if actRly != nil {
			if _, k := actRly.Infos[v.Mid]; k {
				actRly.Infos[v.Mid].Sign = ""
				obj.Act = actRly.Infos[v.Mid]
			}
		}
		v.Object = obj
	}
	if sid == s.c.Bdf.Sid {
		sort.Slice(list, func(i, j int) bool {
			return list[i].Likes > list[j].Likes
		})
		// group archive data
		var aids []int64
		for i, v := range list {
			// only return first 3 arc data
			if i > 2 {
				break
			}
			if piece, ok := s.bdfData[v.ID]; ok {
				aids = append(aids, piece...)
			}
		}
		if len(aids) > 0 {
			reply, err := client.ArchiveClient.Arcs(c, &arccli.ArcsRequest{Aids: aids})
			if err != nil {
				log.Error("Bdf s.arcClient.Arcs aids(%v) error(%v)", aids, err)
				err = nil
			}
			arcs := reply.GetArcs()
			for i, v := range list {
				if i > 2 {
					break
				}
				if piece, ok := s.bdfData[v.ID]; ok {
					for _, aid := range piece {
						if arc, ok := arcs[aid]; ok && arc != nil && arc.IsNormal() {
							v.List = append(v.List, l.CopyFromArc(arc))
						}
					}
				}
			}
		}
	}
	return
}

// actContent get like_content and account info.
func (s *Service) actContent(c context.Context, list []*l.List, mid int64) (err error) {
	var (
		lt              = len(list)
		lids            = make([]int64, 0, lt)
		wids            = make([]int64, 0, lt)
		cont            map[int64]*l.LikeContent
		accRly          *accapi.InfosReply
		ip              = metadata.String(c, metadata.RemoteIP)
		followersRly    *accapi.RelationsReply
		followersNumRly *accRelationApi.StatsReply
	)
	for _, v := range list {
		lids = append(lids, v.ID)
		if v.Wid > 0 {
			wids = append(wids, v.Wid)
		}
	}
	eg := errgroup.WithContext(c)
	eg.Go(func(errCtx context.Context) (e error) {
		cont, e = s.dao.LikeContent(errCtx, lids)
		return
	})
	if len(wids) > 0 {
		eg.Go(func(errCtx context.Context) (e error) {
			accRly, e = s.accClient.Infos3(errCtx, &accapi.MidsReq{Mids: wids})
			return
		})
		eg.Go(func(errCtx context.Context) (e error) {
			followersNumRly, e = client.RelationClient.Stats(errCtx, &accRelationApi.MidsReq{Mids: wids})
			return
		})
	}
	if mid > 0 {
		eg.Go(func(errCtx context.Context) (e error) {
			followersRly, e = s.accClient.Relations3(errCtx, &accapi.RelationsReq{Mid: mid, Owners: wids, RealIp: ip})
			return
		})
	}
	if err = eg.Wait(); err != nil {
		log.Error("actContent:eg.Wait() error(%v)", err)
		return
	}
	for _, v := range list {
		obj := make(map[string]interface{}, 2)
		if _, ok := cont[v.ID]; ok {
			obj["cont"] = cont[v.ID]
		}
		var t struct {
			*accapi.Info
			Following    bool  `json:"following"`
			FollowerNum  int64 `json:"follower_num"`
			FollowingNum int64 `json:"following_num"`
		}
		if accRly != nil {
			if _, k := accRly.Infos[v.Wid]; k {
				t.Info = accRly.Infos[v.Wid]
				t.Info.Birthday = 0
			}
		}
		if mid > 0 {
			if _, f := followersRly.Relations[v.Wid]; f {
				t.Following = followersRly.Relations[v.Wid].Following
			}
		}
		if followersNumRly != nil {
			if statReply, n := followersNumRly.StatReplyMap[v.Wid]; n {
				t.FollowerNum = statReply.Follower
				t.FollowingNum = statReply.Following
			}
		}
		obj["act"] = t
		v.Object = obj
	}
	return
}

// orderByCtime .
func (s *Service) orderByCtime(c context.Context, sid int64, pn, ps int, subInfo *l.SubjectItem, ltype int64) (res []*l.List, err error) {
	var (
		lids    []int64
		start   = int64((pn - 1) * ps)
		end     = start + int64(ps) - 1
		items   map[int64]*l.Item
		likeAct map[int64]int64
	)
	if lids, err = s.dao.LikeCtime(c, sid, ltype, start, end); err != nil {
		log.Error("s.dao.LikeCtime(%d,%d,%d) error(%+v)", sid, start, end, err)
		return
	}
	lt := len(lids)
	if lt == 0 {
		return
	}
	eg := errgroup.WithContext(c)
	eg.Go(func(errCtx context.Context) (e error) {
		items, e = s.dao.Likes(errCtx, lids)
		return
	})
	if subInfo.DisplayLike() {
		eg.Go(func(errCtx context.Context) (e error) {
			likeAct, e = s.dao.LikeActLidCounts(errCtx, lids)
			return
		})
	}
	if err = eg.Wait(); err != nil {
		log.Error("orderByCtime:eg.Wait() error(%+v)", err)
		return
	}
	res = make([]*l.List, 0, lt)
	for _, v := range lids {
		if _, ok := items[v]; ok && items[v].ID > 0 {
			t := &l.List{Item: items[v]}
			if subInfo.DisplayLike() {
				if _, f := likeAct[v]; f {
					t.Like = likeAct[v]
				}
			}
			res = append(res, t)
		} else {
			log.Info("s.dao.CacheLikes(%d) not found", v)
		}
	}
	return
}

// stochasticLids .
func (s *Service) stochasticLids(c context.Context, sid, ltype int64, ps int) (lids []int64, err error) {
	var (
		allLids []int64
		randMap map[int64]int64
	)
	if allLids, err = s.dao.ActStochastic(c, sid, ltype); err != nil {
		log.Error("s.dao.LikeRandom(%d) error(%+v)", sid, err)
		return
	}
	if len(allLids) == 0 {
		return
	}
	randMap = make(map[int64]int64, len(allLids))
	for _, v := range allLids {
		randMap[v] = v
	}
	lids = make([]int64, 0, ps)
	i := 0
	for _, v := range randMap {
		if i < ps {
			lids = append(lids, v)
		} else {
			break
		}
		i++
	}
	return
}

// orderByStochastic order by random
func (s *Service) orderByStochastic(c context.Context, sid int64, ps int, subInfo *l.SubjectItem, ltype int64) (res []*l.List, err error) {
	var (
		lids    []int64
		items   map[int64]*l.Item
		likeAct map[int64]int64
	)
	if sid == s.c.Rule.LimitArcSid {
		// 特殊活动多获取20个id
		ps = ps + 20
	}
	if lids, err = s.stochasticLids(c, sid, ltype, ps); err != nil {
		log.Error("s.stochasticLids(%d) error(%+v)", sid, err)
		return
	}
	lt := len(lids)
	if lt == 0 {
		return
	}
	eg := errgroup.WithContext(c)
	eg.Go(func(errCtx context.Context) (e error) {
		items, e = s.dao.Likes(errCtx, lids)
		return
	})
	if subInfo.DisplayLike() {
		eg.Go(func(errCtx context.Context) (e error) {
			likeAct, _ = s.dao.LikeActLidCounts(errCtx, lids)
			return
		})
	}
	if err = eg.Wait(); err != nil {
		log.Error("orderByRandom:eg.Wait() error(%+v)", err)
		return
	}
	res = make([]*l.List, 0, lt)
	for _, v := range lids {
		if _, ok := items[v]; ok && items[v].ID > 0 {
			t := &l.List{Item: items[v]}
			if subInfo.DisplayLike() {
				if _, f := likeAct[v]; f {
					t.Like = likeAct[v]
				}
			}
			res = append(res, t)
		} else {
			log.Info("s.dao.orderByRandom(%d) not found", v)
		}
	}
	return
}

// orderByRandom order by random
func (s *Service) orderByRandom(c context.Context, sid int64, pn, ps int, subInfo *l.SubjectItem, ltype int64) (res []*l.List, err error) {
	var (
		lids    []int64
		start   = (pn - 1) * ps
		end     = start + ps - 1
		items   map[int64]*l.Item
		likeAct map[int64]int64
	)
	if start > 500 {
		return
	}
	if lids, err = s.dao.ActRandom(c, sid, ltype, int64(start), int64(end)); err != nil {
		log.Error("s.dao.LikeRandom(%d,%d,%d) error(%+v)", sid, start, end, err)
		return
	}
	if len(lids) == 0 {
		return
	}
	eg := errgroup.WithContext(c)
	eg.Go(func(errCtx context.Context) (e error) {
		items, e = s.dao.Likes(errCtx, lids)
		return
	})
	if subInfo.DisplayLike() {
		eg.Go(func(errCtx context.Context) (e error) {
			likeAct, _ = s.dao.LikeActLidCounts(errCtx, lids)
			return
		})
	}
	if err = eg.Wait(); err != nil {
		log.Error("orderByRandom:eg.Wait() error(%+v)", err)
		return
	}
	res = make([]*l.List, 0)
	for _, v := range lids {
		if _, ok := items[v]; ok && items[v].ID > 0 {
			t := &l.List{Item: items[v]}
			if subInfo.DisplayLike() {
				if _, f := likeAct[v]; f {
					t.Like = likeAct[v]
				}
			}
			res = append(res, t)
		} else {
			log.Info("s.dao.orderByRandom(%d) not found", v)
		}
	}
	return
}

// orderByLike only fo like .
func (s *Service) orderByLike(c context.Context, sid int64, pn, ps int, subInfo *l.SubjectItem, ltype int64) (res []*l.List, err error) {
	var (
		lids  []int64
		lt    int
		items map[int64]*l.Item
		infos []*l.LidLikeRes
		start = int64((pn - 1) * ps)
		end   = start + int64(ps) - 1
		isEnt bool
	)
	for _, v := range s.c.Ent.UpSids {
		if v == sid {
			isEnt = true
			break
		}
	}
	if isEnt {
		infos, err = s.dao.EntCache(c, sid, start, end)
	} else {
		infos, err = s.dao.RedisCache(c, sid, ltype, start, end)
	}
	if err != nil {
		log.Error("s.dao.RedisCache(%d,%d,%d,%d) error(%+v)", sid, ltype, start, end, err)
		return
	}
	lt = len(infos)
	if lt == 0 {
		return
	}
	lids = make([]int64, 0, lt)
	for _, v := range infos {
		lids = append(lids, v.Lid)
	}
	if items, err = s.dao.Likes(c, lids); err != nil {
		log.Error("s.dao.CacheLikes(%v) error(%+v)", lids, err)
		return
	}
	res = make([]*l.List, 0, lt)
	for _, v := range infos {
		if item, ok := items[v.Lid]; ok && item != nil && item.ID > 0 {
			t := &l.List{Item: item}
			// 部分活隐藏点赞数
			if subInfo.DisplayLike() {
				t.Like = v.Score
			}
			res = append(res, t)
		} else {
			log.Info("s.dao.CacheLikes(%d) not found", v.Lid)
		}
	}
	return
}

// storyLikeCheck .
func (s *Service) storyLikeCheck(c context.Context, sid, lid, mid, storyMaxAct, storyEachMaxAct int64) (left, maxLeft int64, err error) {
	var (
		sumScore, lScore  int64
		lLeft, extraTimes int64
	)
	eg := errgroup.WithContext(c)
	eg.Go(func(errCtx context.Context) (e error) {
		sumScore, e = s.storySumUsed(errCtx, sid, mid)
		return
	})
	eg.Go(func(errCtx context.Context) (e error) {
		lScore, e = s.storyEachUsed(errCtx, sid, mid, lid)
		return
	})
	// 获取额外次数
	eg.Go(func(errCtx context.Context) (e error) {
		extraTimes, e = s.storyExtraTimes(errCtx, sid, mid)
		return
	})
	if err = eg.Wait(); err != nil {
		err = errors.Wrap(err, "eg.Wait()")
		return
	}
	maxLeft = storyMaxAct - sumScore + extraTimes
	lLeft = storyEachMaxAct - lScore
	left = int64(math.Min(float64(maxLeft), float64(lLeft)))
	if left <= 0 {
		left = 0
	}
	return
}

// storyLikeActSet .
func (s *Service) storyLikeActSet(c context.Context, sid, lid, mid int64, score int64) (err error) {
	eg := errgroup.WithContext(c)
	eg.Go(func(errCtx context.Context) (e error) {
		_, e = s.dao.IncrStoryLikeSum(errCtx, sid, mid, score)
		return
	})
	eg.Go(func(errCtx context.Context) (e error) {
		_, e = s.dao.IncrStoryEachLikeAct(errCtx, sid, mid, lid, score)
		return
	})
	if err = eg.Wait(); err != nil {
		log.Error("storyLikeActSet:eg.Wait() error(%+v)", err)
	}
	return
}

// storySumUsed .
func (s *Service) storySumUsed(c context.Context, sid, mid int64) (res int64, err error) {
	if res, err = s.dao.StoryLikeSum(c, sid, mid); err != nil {
		log.Error("s.dao.StoryLikeSum(%d,%d) error(%+v)", sid, mid, err)
		return
	}
	if res == -1 {
		today := time.Now().Format("2006-01-02")
		etime := fmt.Sprintf("%s 23:59:59", today)
		stime := fmt.Sprintf("%s 00:00:00", today)
		if res, err = s.dao.StoryLikeActSum(c, sid, mid, stime, etime); err != nil {
			log.Error("s.dao.StoryLikeActSum(%d,%d) error(%+v)", sid, mid, err)
			return
		}
		if err = s.dao.SetLikeSum(c, sid, mid, res); err != nil {
			log.Error("s.dao.SetLikeSum(%d,%d,%d) error(%+v)", sid, mid, res, err)
		}
	}
	return
}

func (s *Service) storyEachUsed(c context.Context, sid, mid, lid int64) (res int64, err error) {

	if res, err = s.dao.StoryEachLikeSum(c, sid, mid, lid); err != nil {
		log.Error("s.dao.StoryEachLikeSum(%d,%d) error(%+v)", sid, mid, err)
		return
	}
	if res == -1 {
		today := time.Now().Format("2006-01-02")
		etime := fmt.Sprintf("%s 23:59:59", today)
		stime := fmt.Sprintf("%s 00:00:00", today)
		if res, err = s.dao.StoryEachLikeAct(c, sid, mid, lid, stime, etime); err != nil {
			log.Error("s.dao.StoryLikeActSum(%d,%d) error(%+v)", sid, mid, err)
			return
		}
		if err = s.dao.SetEachLikeSum(c, sid, mid, lid, res); err != nil {
			log.Error("s.dao.SetEachLikeSum(%d,%d,%d) error(%+v)", sid, mid, res, err)
		}
	}
	return
}

func (s *Service) storyExtraTimes(c context.Context, sid, mid int64) (res int64, err error) {
	var (
		list []*l.ExtraTimesDetail
	)
	if list, err = s.dao.CacheLikeExtraTimes(c, sid, mid, 0, -1); err != nil {
		log.Error("storyExtraTimes s.dao.CacheLikeExtendTimes(%d,%d) error(%v)", sid, mid, err)
		return
	}
	nowT := time.Now().Format("2006-01-02")
	timeTemplate := "2006-01-02 15:04:05"
	start, _ := time.ParseInLocation(timeTemplate, nowT+" 00:00:00", time.Local)
	st := start.Unix()
	end, _ := time.ParseInLocation(timeTemplate, nowT+" 23:59:59", time.Local)
	et := end.Unix()
	if len(list) == 0 {
		if list, err = s.dao.RawLikeExtraTimes(c, sid, mid); err != nil {
			log.Error("storyExtraTimes s.dao.RawLikeExtraTimes(%d,%d) error(%v)", sid, mid, err)
			return
		}
		miss := make([]*l.ExtraTimesDetail, 0)
		for _, v := range list {
			if v != nil && v.Ctime >= xtime.Time(st) && v.Ctime <= xtime.Time(et) {
				miss = append(miss, v)
				res = res + int64(v.Num)
			}
		}
		if err = s.dao.AddCacheLikeExtraTimes(c, sid, mid, miss); err != nil {
			log.Error("storyExtraTimes s.dao.AddCacheLikeExtendTimes(%d,%d,%v) error(%v)", sid, mid, miss, err)
			return
		}
		return
	}
	for _, v := range list {
		if v != nil && v.Ctime >= xtime.Time(st) && v.Ctime <= xtime.Time(et) {
			res = res + int64(v.Num)
		}
	}
	return
}

func (s *Service) upWholeEchoTimes(c context.Context, sid, mid int64) (list []*l.Action, err error) {
	addCache := true
	if list, err = s.dao.CacheUpActionTimes(c, sid, mid, 0, -1); err != nil {
		log.Error("upWholeEchoTimes s.dao.CacheLikeExtendTimes(%d,%d) error(%v)", sid, mid, err)
		addCache = false
		err = nil
	}
	defer func() {
		if len(list) == 1 && list[0] != nil && list[0].ID == -1 {
			list = nil
		}
	}()
	if len(list) != 0 {
		return
	}
	if list, err = s.retryUpWholeAction(c, sid, mid); err != nil {
		log.Error("upWholeEchoTimes s.dao.RawLikeExtraTimes(%d,%d) error(%v)", sid, mid, err)
		return
	}
	miss := list
	if len(miss) == 0 {
		miss = []*l.Action{{ID: -1}}
	}
	if !addCache {
		return
	}
	s.cache.Do(c, func(c context.Context) {
		if e := s.dao.AddCacheUpActionTimes(c, sid, mid, miss); e != nil {
			log.Error("upWholeEchoTimes s.dao.AddCacheLikeExtendTimes(%d,%d,%v) error(%v)", sid, mid, list, e)
		}
	})
	return
}

func (s *Service) retryUpWholeAction(c context.Context, sid, mid int64) (list []*l.Action, err error) {
	for i := 0; i < _retryTime; i++ {
		if list, err = s.dao.UpWholeAction(c, sid, mid); err == nil {
			break
		}
		time.Sleep(1 * time.Millisecond)
	}
	return
}

func (s *Service) retryUpWholeUsers(c context.Context, sid, lid int64) (list []*l.Action, err error) {
	for i := 0; i < _retryTime; i++ {
		if list, err = s.dao.UpWholeUsers(c, sid, lid); err == nil {
			break
		}
		time.Sleep(1 * time.Millisecond)
	}
	return
}

func (s *Service) upWholeExtraTimes(c context.Context, sid, mid int64) (res int64, err error) {
	var (
		list []*l.ExtraTimesDetail
	)
	if list, err = s.dao.CacheLikeExtraTimes(c, sid, mid, 0, -1); err != nil {
		log.Error("storyExtraTimes s.dao.CacheLikeExtendTimes(%d,%d) error(%v)", sid, mid, err)
		return
	}
	if len(list) == 0 {
		if list, err = s.dao.RawLikeExtraTimes(c, sid, mid); err != nil {
			log.Error("storyExtraTimes s.dao.RawLikeExtraTimes(%d,%d) error(%v)", sid, mid, err)
			return
		}
		miss := make([]*l.ExtraTimesDetail, 0)
		for _, v := range list {
			if v != nil {
				miss = append(miss, v)
				res = res + int64(v.Num)
			}
		}
		if err = s.dao.AddCacheLikeExtraTimes(c, sid, mid, miss); err != nil {
			log.Error("storyExtraTimes s.dao.AddCacheLikeExtendTimes(%d,%d,%v) error(%v)", sid, mid, miss, err)
			return
		}
		return
	}
	for _, v := range list {
		if v != nil {
			res = res + int64(v.Num)
		}
	}
	return
}

// LikeActList get sid&lid likeact list .
func (s *Service) LikeActList(c context.Context, sid, mid int64, lids []int64) (res map[int64]interface{}, err error) {
	var (
		likeCounts map[int64]int64
		likeActs   map[int64]int
		likeCount  int64
		isLike     int
	)
	group := errgroup.WithContext(c)
	group.Go(func(ctx context.Context) (e error) {
		if likeCounts, e = s.dao.LikeActLidCounts(ctx, lids); e != nil {
			log.Error("s.dao.LikeActLidCounts(%v) error(%+v)", lids, e)
			return e
		}
		return nil
	})
	if mid > 0 {
		group.Go(func(ctx context.Context) (e error) {
			if likeActs, e = s.dao.LikeActs(ctx, sid, mid, lids); e != nil {
				log.Error("s.dao.LikeActMidList(%v,%d,%d) error(%+v)", lids, sid, mid, e)
				return e
			}
			return nil
		})
	}
	if err = group.Wait(); err != nil {
		log.Error("get likeactListerror(%v)", err)
		return
	}
	res = make(map[int64]interface{}, len(lids))
	for _, lid := range lids {
		if _, ok := likeCounts[lid]; ok {
			likeCount = likeCounts[lid]
		} else {
			likeCount = 0
		}
		if _, ok := likeActs[lid]; ok {
			isLike = likeActs[lid]
		} else {
			isLike = 0
		}
		res[lid] = map[string]interface{}{
			"likeCount": likeCount,
			"isLike":    isLike,
		}
	}
	return
}

// isLikeType range liketype find out real type .
func (s *Service) isLikeType(subType int64) (res string) {
	for _, ty := range l.LIKETYPE {
		if subType == ty {
			res = _like
			return
		}
	}
	if subType == l.MUSIC {
		res = _vote
	} else {
		res = _grade
	}
	return
}

// simpleJudge judge user could like or not .
func (s *Service) simpleJudge(c context.Context, subject *l.SubjectItem, member *accapi.Profile) (err error) {
	if member.Silence == _silenceForbid {
		err = ecode.ActivityMemberBlocked
		return
	}
	if subject.Flag == 0 {
		return
	}
	if (subject.Flag & l.FLAGSPY) == l.FLAGSPY {
		var reply *spyapi.InfoReply
		if reply, err = s.spyClient.Info(c, &spyapi.InfoReq{Mid: member.Mid}); err != nil {
			log.Error("s.spy.UserScore(%d) error(%v)", member.Mid, err)
			return
		}
		if reply.Ui == nil || int64(reply.Ui.Score) <= s.c.Rule.Spylike {
			err = ecode.ActivityLikeScoreLower
			return
		}
	}
	if (subject.Flag & l.FLAGUSTIME) == l.FLAGUSTIME {
		if subject.Ustime <= xtime.Time(member.JoinTime) {
			err = ecode.ActivityLikeRegisterLimit
			return
		}
	}
	if (subject.Flag & l.FLAGUETIME) == l.FLAGUETIME {
		if subject.Uetime >= xtime.Time(member.JoinTime) {
			err = ecode.ActivityLikeBeforeRegister
			return
		}
	}
	if (subject.Flag & l.FLAGPHONEBIND) == l.FLAGPHONEBIND {
		if member.TelStatus != 1 {
			err = ecode.ActivityTelValid
			return
		}
	}
	if (subject.Flag & l.FLAGLEVEL) == l.FLAGLEVEL {
		if subject.Level > int64(member.Level) {
			err = ecode.ActivityLikeLevelLimit
		}
	}
	return
}

// judgeUser judge user could like or not .
func (s *Service) judgeUser(c context.Context, subject *l.SubjectItem, member *accapi.Profile) (err error) {
	if member.Silence == _silenceForbid {
		err = ecode.ActivityMemberBlocked
		return
	}
	if subject.Flag == 0 {
		return
	}
	if (subject.Flag & l.FLAGIP) == l.FLAGIP {
		ip := metadata.String(c, metadata.RemoteIP)
		var used int
		if used, err = s.dao.CacheIPRequestCheck(c, ip); err != nil {
			log.Error("s.dao.CacheIPRequestCheck(%s) error(%+v)", ip, err)
			return
		}
		if used == 0 {
			if err = s.dao.AddCacheIPRequestCheck(c, ip, 1); err != nil {
				log.Error("s.dao.AddCacheIPRequestCheck(%s) error(%+v)", ip, err)
				return
			}
		} else {
			err = ecode.ActivityLikeIPFrequence
			return
		}
	}
	if (subject.Flag & l.FLAGSPY) == l.FLAGSPY {
		var reply *spyapi.InfoReply
		if reply, err = s.spyClient.Info(c, &spyapi.InfoReq{Mid: member.Mid}); err != nil {
			log.Error("s.spyClient.Info(%d) error(%v)", member.Mid, err)
			return
		}
		if reply.Ui == nil || int64(reply.Ui.Score) <= s.c.Rule.Spylike {
			err = ecode.ActivityLikeScoreLower
			return
		}
	}
	if (subject.Flag & l.FLAGUSTIME) == l.FLAGUSTIME {
		if subject.Ustime <= xtime.Time(member.JoinTime) {
			err = ecode.ActivityLikeRegisterLimit
			return
		}
	}
	if (subject.Flag & l.FLAGUETIME) == l.FLAGUETIME {
		if subject.Uetime >= xtime.Time(member.JoinTime) {
			err = ecode.ActivityLikeBeforeRegister
			return
		}
	}
	if (subject.Flag & l.FLAGPHONEBIND) == l.FLAGPHONEBIND {
		if member.TelStatus != 1 {
			err = ecode.ActivityTelValid
			return
		}
	}
	if (subject.Flag & l.FLAGLEVEL) == l.FLAGLEVEL {
		if subject.Level > int64(member.Level) {
			err = ecode.ActivityLikeLevelLimit
			return
		}
	}
	// taaf check realname or cn phone num
	if subject.ID == s.c.Taaf.Sid || subject.ID == s.c.Taaf.SidV2 {
		if member.Identification == 0 {
			// 绑定手机才查国家码
			if member.TelStatus == 1 {
				var detailReply *passapi.UserDetailReply
				if detailReply, err = s.passportClient.UserDetail(c, &passapi.UserDetailReq{Mid: member.Mid}); err != nil {
					log.Error("s.passportClient.UserDetail mid(%d) error(%v)", member.Mid, err)
					err = ecode.ActivityNoIdentification
					return
				}
				if detailReply.CountryCode != _countryCodeCN {
					err = ecode.ActivityNoIdentification
					return
				}
			} else {
				err = ecode.ActivityNoIdentification
				return
			}
		}
	}
	return
}

func (s *Service) loadArcType() {
	if types, err := client.ArchiveClient.Types(context.Background(), &arccli.NoArgRequest{}); err != nil {
		log.Error("s.arcRPC.Types2 error(%v)", err)
		return
	} else {
		if types != nil {
			s.arcType = types.Types
		}
	}
	log.Info("loadArcType() success")
}

func (s *Service) loadActSource() {
	if s.c.Rule.DialectSid != 0 {
		s.updateActSourceList(context.Background(), s.c.Rule.DialectSid, _typeAll)
	}
	if len(s.c.Rule.SpecialSids) > 0 {
		for _, sid := range s.c.Rule.SpecialSids {
			if sid > 0 {
				s.updateActSourceList(context.Background(), sid, _typeRegion)
			}
		}
	}
	log.Info("loadActSource() success")
}

// LikeDel .
func (s *Service) LikeDel(c context.Context, sid, lid, mid int64) (ef int64, err error) {
	var (
		subject *l.SubjectItem
		state   = -1
	)
	if subject, err = s.dao.ActSubject(c, sid); err != nil {
		return
	}
	//暂时只支持图片类型活动
	if subject.ID == 0 || (subject.Type != l.PICTURE && subject.Type != l.PICTURELIKE) {
		err = ecode.ActivityHasOffLine
		return
	}
	if ef, err = s.dao.StateModify(c, lid, mid, state); err != nil {
		log.Error("s.dao.StateModify(%d) error(%v)", lid, err)
	}
	if ef > 0 {
		log.Warn("sid:%d,%d has been del by %d", sid, lid, mid)
	}
	return
}

func (s *Service) yeGrKey(mid, nowTs int64, period *likemdl.YellowGreenPeriod) string {
	nowD := time.Now().Format("2006-01-02")
	timeTemplate := "2006-01-02 15:04:05"
	start, _ := time.ParseInLocation(timeTemplate, nowD+" 12:00:00", time.Local) //时间转Time类型
	st := start.Unix()
	if nowTs < st {
		return fmt.Sprintf("yegl_%d_%d_%d_%s", mid, period.YellowSid, period.GreenSid, time.Now().AddDate(0, 0, -1).Format("20060102"))
	}
	return fmt.Sprintf("yegl_%d_%d_%d_%s", mid, period.YellowSid, period.GreenSid, time.Now().Format("20060102"))
}

func (s *Service) addExtraTimes(c context.Context, extend *l.ExtendTokenDetail) (err error) {
	var (
		times, id int64
	)
	// 查询额外增加次数
	if times, err = s.storyExtraTimes(c, extend.Sid, extend.Mid); err != nil {
		log.Error("addExtraTimes s.storyExtendTimes(%d,%d) error(%v)", extend.Sid, extend.Mid, err)
		return
	}
	if times >= extend.Max {
		return
	}
	if id, err = s.dao.IncrLikeExtraTimes(c, extend.Sid, extend.Mid, s.c.AnnualVoting.Num); err != nil {
		log.Error("addExtraTimes s.dao.IncrLikeExtraTimes(%d,%d) error(%v)", extend.Sid, extend.Mid, err)
		return
	}
	if err = s.dao.AddLikeExtraTimes(c, extend.Sid, extend.Mid, &l.ExtraTimesDetail{ID: id, Sid: extend.Sid, Mid: extend.Mid, Num: 1, Ctime: xtime.Time(time.Now().Unix())}); err != nil {
		log.Error("addExtraTimes s.dao.AddLikeExtraTimes(%d,%d) error(%v)", extend.Sid, extend.Mid, err)
	}
	return
}

func (s *Service) AddLikeHisList(c context.Context, sid int64) (err error) {
	var (
		start    int64 = 0
		end      int64 = 150
		nowRank  []*l.LidLikeRes
		rankByte []byte
	)
	if nowRank, err = s.dao.RedisCache(c, sid, 0, start, end); err != nil {
		log.Error("CacheUpListCtime s.dao.RedisCache(%d) error(%v)", sid, err)
		return
	}
	if rankByte, err = json.Marshal(nowRank); err != nil {
		log.Error("CacheUpListCtime json.Marshal error(%v)", err)
		return
	}
	rankStr := string(rankByte)
	log.Warn("CacheUpListCtime sid(%d) ranks(%s)", sid, rankStr)
	if err = s.dao.AddCacheHisLikeScore(c, sid, rankStr); err != nil {
		log.Error("s.dao.AddCacheHisLikeScore %d error(%v)", sid, err)
	}
	return
}

func (s *Service) LikeActToken(c context.Context, sid, mid int64) (*l.ExtendTokenDetail, error) {
	if sid < 1 || mid < 1 {
		return nil, xecode.RequestErr
	}
	if sid == s.c.SpringCardAct.InviteSid {
		memberRly, err := s.accClient.Profile3(c, &accapi.MidReq{Mid: mid})
		if err != nil {
			log.Error("LikeActToken s.accRPC.Profile3(c,&accmdl.ArgMid{Mid:%d}) error(%v)", mid, err)
			return nil, err
		}
		// 一次性活动等级限制
		if memberRly == nil || memberRly.Profile.Level < 2 {
			return nil, ecode.ActivityLikeLevelLimit
		}
	}
	let, err := s.dao.LikeExtendToken(c, sid, mid)
	if err != nil {
		log.Error("LikeActToken s.dao.LikeExtendToken(%d,%d) error(%v)", sid, mid, err)
		return nil, err
	}
	if let == nil || let.ID == 0 {
		token := strconv.FormatInt(sid, 10) + strconv.FormatInt(mid, 10) + time.Now().Format("20060102150405")
		id, err := s.dao.IncrLikeExtendToken(c, sid, mid, s.c.AnnualVoting.Max, token)
		if err != nil {
			log.Error("LikeActToken s.dao.IncrLikeExtendToken(%d,%d,%s) error(%v)", sid, mid, token, err)
			return nil, err
		}
		let = &l.ExtendTokenDetail{
			ID:    id,
			Sid:   sid,
			Mid:   mid,
			Token: token,
			Max:   s.c.AnnualVoting.Max,
			Ctime: xtime.Time(time.Now().Unix()),
		}
		s.cache.Do(c, func(c context.Context) {
			s.dao.AddCacheLikeExtendToken(c, sid, let, mid)
		})
	}
	return let, nil
}

// LikeHisList .
func (s *Service) LikeHisList(c context.Context, sid int64) (res []*l.List, err error) {
	var (
		rankList []*l.LidLikeRes
		lids     []int64
		items    map[int64]*l.Item
	)
	if rankList, err = s.dao.CacheHisLikeScore(c, sid); err != nil {
		log.Error("s.dao.CacheHisLikeScore sid(%d) error(%v)", sid, err)
		return
	}
	lt := len(rankList)
	if lt == 0 {
		return
	}
	lids = make([]int64, 0, lt)
	for _, v := range rankList {
		lids = append(lids, v.Lid)
	}
	if items, err = s.dao.Likes(c, lids); err != nil {
		log.Error("s.dao.CacheLikes(%v) error(%+v)", lids, err)
		return
	}
	res = make([]*l.List, 0, lt)
	for _, v := range rankList {
		if item, ok := items[v.Lid]; ok && item != nil && item.ID > 0 {
			t := &l.List{Item: item}
			t.Like = v.Score
			res = append(res, t)
		} else {
			log.Info("s.dao.Likes(%d) not found", v.Lid)
		}
	}
	if err = s.contentAccount(c, res, 0); err != nil {
		log.Error("LikeHisList s.contentAccount error(%v)", err)
	}
	return
}

func (s *Service) InviteTimes(c context.Context, sid, mid int64) (res *l.Invite, err error) {
	var total int64
	total = s.c.AnnualVoting.Max
	res = &l.Invite{Total: total}
	if mid <= 0 {
		res.HasInvited = 0
		return
	}
	var invited int64
	if invited, err = s.storyExtraTimes(c, sid, mid); err != nil {
		log.Error("s.storyExtraTimes(%d,%d) error(%v)", sid, mid, err)
	}
	res.HasInvited = invited
	return
}

func (s *Service) GetUpsRelationData(c context.Context, mid int64, upMIDs []int64) (reply []*l.FollowReply, err error) {
	var (
		res *accRelationApi.FollowingMapReply
	)
	req := &accRelationApi.RelationsReq{
		Mid:    mid,
		Fid:    upMIDs,
		RealIp: metadata.String(c, metadata.RemoteIP),
	}
	for i := 0; i < 3; i++ {
		res, err = s.accRelation.Relations(c, req)
		if err == nil {
			break
		}
	}
	if err != nil {
		log.Errorc(c, "GetUpsRelationData GRPC s.accRelation.Relations Err mid:%v upMIDs:%v err:%v", mid, upMIDs, err)
		return
	}

	if len(res.FollowingMap) > 0 {
		for _, v := range res.FollowingMap {
			reply = append(reply, &l.FollowReply{
				MID:   v.Mid,
				MTime: v.MTime,
			})
		}
	}

	return
}

func (s *Service) GetUpsDetailInfo(c context.Context, upMIDs []int64) (reply map[int64]*l.GetMIDInfo, err error) {
	var (
		res *accapi.InfosReply
	)
	reply = make(map[int64]*l.GetMIDInfo)
	req := new(accapi.MidsReq)
	for _, v := range upMIDs {
		req.Mids = append(req.Mids, v)
	}
	for i := 0; i < 3; i++ {
		res, err = s.accClient.Infos3(c, req)
		if err == nil {
			break
		}
	}
	if err != nil {
		log.Errorc(c, "GetUpsRelationData GRPC s.accClient.Infos3 Err upMIDs:%v err:%v", upMIDs, err)
		return
	}

	for _, v := range res.Infos {
		reply[v.Mid] = &l.GetMIDInfo{
			Mid:  v.Mid,
			Name: v.Name,
			Face: v.Face,
		}
	}

	return
}
