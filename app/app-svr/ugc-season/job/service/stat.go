package service

import (
	"context"
	"encoding/json"
	xtime "go-common/library/time"
	"sync"
	"time"

	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/queue/databus"
	"go-common/library/railgun"
	"go-common/library/sync/errgroup.v2"

	arcApi "go-gateway/app/app-svr/archive/service/api"

	"go-gateway/app/app-svr/ugc-season/job/model/retry"
	"go-gateway/app/app-svr/ugc-season/job/model/stat"
	"go-gateway/app/app-svr/ugc-season/service/api"
)

const (
	_maxAids      = 50
	_maxRetryTime = 10
)

// consumerproc consumer all stats' topic and merge data into the one channel
func (s *Service) consumerSnproc(k string, d *databus.Databus) {
	defer s.waiterSeason.Done()
	var msgs = d.Messages()
	for {
		var (
			err error
			ok  bool
			msg *databus.Message
			now = time.Now().Unix()
		)
		msg, ok = <-msgs
		if !ok || s.closeSub {
			log.Info("databus(%s) consumer exit", k)
			return
		}
		_ = msg.Commit()
		var ms = &stat.Count{}
		if err = json.Unmarshal(msg.Value, ms); err != nil {
			log.Error("json.Unmarshal(%s) error(%v)", string(msg.Value), err)
			continue
		}
		if ms.Aid <= 0 || (ms.Type != "archive" && ms.Type != "archive_his") {
			log.Warn("message(%s) error", msg.Value)
			continue
		}
		if now-ms.TimeStamp > 8*60*60 {
			log.Warn("topic(%s) message(%s) too early", msg.Topic, msg.Value)
			continue
		}
		statMsg := &stat.Msg{Aid: ms.Aid, Type: k, Ts: ms.TimeStamp}
		switch k {
		case stat.TypeForView:
			statMsg.Click = ms.Count
		case stat.TypeForDm:
			statMsg.DM = ms.Count
		case stat.TypeForReply:
			statMsg.Reply = ms.Count
		case stat.TypeForFav:
			statMsg.Fav = ms.Count
		case stat.TypeForCoin:
			statMsg.Coin = ms.Count
		case stat.TypeForShare:
			statMsg.Share = ms.Count
		case stat.TypeForRank:
			statMsg.HisRank = ms.Count
		case stat.TypeForLike:
			statMsg.Like = ms.Count
			statMsg.DisLike = ms.DisLike
		default:
			log.Error("unknow type(%s) message(%s)", k, msg.Value)
			continue
		}
		s.statCh <- statMsg
		log.Info("SeasonStat TopicMsg got message(%+v)", statMsg)
	}
}

func (s *Service) statSnDealproc() {
	defer s.waiterSeason.Done()
	var c = context.Background()
	for {
		var (
			ok  bool
			msg interface{}
		)
		if msg, ok = <-s.statCh; !ok {
			log.Warn("SeasonStat statSnDealproc quit")
			return
		}
		switch ms := msg.(type) {
		case *stat.Msg: // stat msg, increase stat
			log.Info("SeasonStat statSnDealproc Aid %d, Type *model.StatMsg", ms.Aid)
			s.snMsgUpdate(c, ms)
		case *stat.SeasonResult:
			log.Info("SeasonStat statSnDealproc Sid %d, Type *model.StatMsg", ms.SeasonID)
			if ms.Action == retry.ActionDel {
				log.Warn("SeasonStat Delete SeasonID %d", ms.SeasonID)
				s.seasonResDel(c, ms.SeasonID)
			} else if ms.Action == retry.ActionUp {
				log.Warn("SeasonStat Update SeasonID %d", ms.SeasonID)
				s.seasonResUpdate(c, ms.SeasonID)
			}
		default:
			log.Warn("SeasonStat statSnDealproc UnknownType %v", msg)
		}
	}
}

// nolint:gomnd
func (s *Service) snMsgUpdate(c context.Context, ms *stat.Msg) {
	var (
		now        = time.Now().Unix() //当前时间戳
		err        error
		seasonStat *api.Stat
		aids       []int64
		lock       bool
	)
	//根据aid获取season_id
	arc, err := s.GetArc(c, ms.Aid)
	if err != nil || arc == nil {
		log.Error("s.GetArc is error %+v %+v %+v", ms, err, arc)
		return
	}
	if arc.SeasonID == 0 {
		return
	}
	//根据season_id获取合集计数
	seasonStat, err = s.GetStCache(c, arc.SeasonID) //缓存
	if err != nil || seasonStat == nil {
		log.Warn("s.GetStCache is error %+v %+v %+v", ms, err, arc.SeasonID)
		seasonStat, err = s.statDao.SnStat(c, arc.SeasonID) //数据库
		if err != nil || seasonStat == nil {
			log.Warn("s.statDao.SnStat (%d) error(%v)", arc.SeasonID, err)
			return
		}
	}
	//获取合集下的所有aid
	seasonInfo, err := s.GetSeason(c, arc.SeasonID)
	if err != nil || seasonInfo == nil {
		log.Error("s.GetSeason is error %+v %+v", arc.SeasonID, err)
		_, aids, err = s.resultDao.SnArcs(c, arc.SeasonID)
	}
	if seasonInfo != nil {
		for _, sec := range seasonInfo.GetSections() {
			for _, ep := range sec.GetEpisodes() {
				aids = append(aids, ep.Aid)
			}
		}
	}
	for i := 0; i < 3; i++ { //加锁
		lock, err = s.TryLock(c, statWatchLock(arc.Aid), "lock", 1)
		if !lock {
			log.Warn("s.TryLock fail %+v %+v %+v", arc.Aid, lock, err)
			time.Sleep(1 * time.Second)
			continue
		}
		break
	}
	if !lock {
		log.Error("s.TryLock error %+v %+v %+v", arc.Aid, lock, err)
		return
	}
	defer func() { //释放锁
		if unlock := s.UnLock(c, statWatchLock(arc.Aid), "lock"); !unlock {
			log.Error("s.UnLock is error %+v %+v", arc.Aid, unlock)
		}
	}()
	var snStat *api.Stat // season的stat数据重新计算
	if snStat, err = s.snStatReSum(c, arc.SeasonID, aids); err != nil {
		log.Error("snMsgUpdate SeasonStat snStatPick Sid %d, Aids %v, Err %v", err, aids, err)
		return
	}
	snStat.Mtime = xtime.Time(time.Now().Unix())      //更新缓存Mtime为最新时间
	if err = s.updateSnCache(c, snStat); err != nil { // update redis
		log.Error("snMsgUpdate SeasonStat UpdateSnStat redis Sid %d, Err %v", arc.SeasonID, err)
		return
	}
	if now-int64(seasonStat.Mtime) > 120 {
		if _, err = s.statDao.UpdateSnStat(c, snStat); err != nil { // update DB
			log.Error("snMsgUpdate SeasonStat UpdateSnStat DB Sid %d, Err %v", arc.SeasonID, err)
			return
		}
	}
}

func (s *Service) UpdateSeasonStatDBAndCache(ctx context.Context, sid int64) error {
	var aids []int64
	//获取合集下的所有aid
	seasonInfo, err := s.GetSeason(ctx, sid)
	if err != nil || seasonInfo == nil {
		log.Error("s.GetSeason is error %+v %+v", sid, err)
		_, aids, err = s.resultDao.SnArcs(ctx, sid)
	}
	if seasonInfo != nil {
		for _, sec := range seasonInfo.GetSections() {
			for _, ep := range sec.GetEpisodes() {
				aids = append(aids, ep.Aid)
			}
		}
	}
	if err != nil {
		log.Error("SeasonStat SnArcs Sid %d, Err %v", sid, err)
		return err
	}
	var snStat *api.Stat // season的stat数据重新加总 并且 更新：DB && Redis
	if snStat, err = s.snStatReSum(ctx, sid, aids); err != nil {
		log.Error("SeasonStat snStatPick Sid %d, Aids %v, Err %v", err, aids, err)
		return err
	}
	snStat.Mtime = xtime.Time(time.Now().Unix())        //更新缓存Mtime为最新时间
	if err = s.updateSnCache(ctx, snStat); err != nil { // update redis
		log.Error("SeasonStat UpdateSnStat redis Sid %d, Err %v", sid, err)
		return err
	}
	if _, err = s.statDao.UpdateSnStat(ctx, snStat); err != nil { // update DB
		log.Error("SeasonStat UpdateSnStat DB Sid %d, Err %v", sid, err)
		return err
	}
	return nil
}

// seasonResDel deletes the season from memory, redis & db and tell JD
func (s *Service) seasonResDel(c context.Context, sid int64) {
	err := s.UpdateSeasonStatDBAndCache(c, sid)
	if err != nil {
		log.Error("seasonResDel ID %d Err %v", sid, err)
		rt := &retry.Info{Action: retry.FailUpSeasonStat}
		rt.Data.SeasonID = sid
		_ = s.PushToRetryList(c, rt)
	}
}

func (s *Service) seasonResUpdate(c context.Context, sid int64) {
	if err := s.statUpdate(c, sid); err != nil {
		log.Error("seasonResUpdate ID %d Err %v", sid, err)
		rt := &retry.Info{Action: retry.FailUpSeasonStat}
		rt.Data.SeasonID = sid
		_ = s.PushToRetryList(c, rt)
	}
}

// statUpdate checks the new arcs, removed arcs and re-establish the season's stat
func (s *Service) statUpdate(c context.Context, sid int64) (err error) {
	err = s.UpdateSeasonStatDBAndCache(c, sid)
	if err != nil {
		log.Error("statUpdate s.UpdateSeasonStatDBAndCache is err %+v %+v", sid, err)
		return
	}
	return
}

// snStatReSum resets the season's stat in DB & Redis, and it returns the season's stat result to update in MEMORY
func (s *Service) snStatReSum(c context.Context, sid int64, aids []int64) (snStat *api.Stat, err error) {
	var (
		stats map[int64]*arcApi.Stat
	)
	if stats, err = s.getStats(c, aids); err != nil {
		log.Error("SeasonStat snStatReSum sid %d, aids %v, Err %v", sid, aids, err)
		return
	}
	snStat = &api.Stat{
		SeasonID: sid,
	}
	for _, v := range stats {
		stat.MergeArcStat(snStat, v)
	}
	return
}

func (s *Service) getStats(c context.Context, aids []int64) (stats map[int64]*arcApi.Stat, err error) {
	var (
		aidsLen = len(aids)
		mutex   = sync.Mutex{}
	)
	stats = make(map[int64]*arcApi.Stat, aidsLen)
	gp := errgroup.WithContext(c)
	for i := 0; i < aidsLen; i += _maxAids {
		var partAids []int64
		if i+_maxAids > aidsLen {
			partAids = aids[i:]
		} else {
			partAids = aids[i : i+_maxAids]
		}
		gp.Go(func(ctx context.Context) (err error) {
			var tmpRes *arcApi.StatsReply
			arg := &arcApi.StatsRequest{Aids: partAids}
			if tmpRes, err = s.getStatsWithRetry(ctx, arg); err != nil {
				log.Error("initAidToSn s.arcClient.Stats error(%+v)", err)
				return
			}
			if len(tmpRes.Stats) > 0 {
				mutex.Lock()
				for aid, stat2 := range tmpRes.Stats {
					stats[aid] = stat2
				}
				mutex.Unlock()
			}
			return err
		})
	}
	if err = gp.Wait(); err != nil {
		log.Error("initAidToSn gp.Wait() %+v", err)
		return
	}
	return
}

func (s *Service) getStatsWithRetry(ctx context.Context, arg *arcApi.StatsRequest) (tmpRes *arcApi.StatsReply, err error) {
	for i := 0; i < _maxRetryTime; i++ {
		if tmpRes, err = s.arcClient.Stats(ctx, arg); err != nil {
			log.Error("getStatsWithRetry s.arcClient.Stats arg(%+v) error(%+v)", arg, err)
			continue
		}
		return tmpRes, err
	}
	log.Error("getStatsWithRetry exceed max retry times s.arcClient.Stats arg(%+v) error(%+v)", arg, err)
	return tmpRes, err
}

func (s *Service) GetArc(ctx context.Context, aid int64) (*arcApi.Arc, error) {
	req := &arcApi.ArcRequest{Aid: aid}
	res, err := s.arcClient.Arc(ctx, req)
	if err != nil {
		log.Error("GetArc arg(%+v) error(%+v)", req, err)
		return nil, err
	}
	return res.Arc, nil
}

func (s *Service) GetSeason(ctx context.Context, sid int64) (*api.View, error) {
	req := &api.ViewRequest{
		SeasonID: sid,
	}
	res, err := s.seasonClient.View(ctx, req)
	if err != nil {
		log.Error("GetSeason arg(%+v) error(%+v)", req, err)
		return nil, err
	}
	return res.GetView(), nil
}

func (s *Service) SnUnpack(msg railgun.Message) (*railgun.SingleUnpackMsg, error) {
	m := &stat.Count{}
	if err := json.Unmarshal(msg.Payload(), m); err != nil {
		log.Error("json.Unmarshal(%+v) error(%+v)", string(msg.Payload()), err)
		return nil, err
	}
	return &railgun.SingleUnpackMsg{
		Group: m.Aid,
		Item:  m,
	}, nil
}

func (s *Service) CoinSnUpDo(ctx context.Context, item interface{}) railgun.MsgPolicy {
	m, ok := item.(*stat.Count)
	if !ok || m == nil {
		return railgun.MsgPolicyIgnore
	}
	if m.Aid <= 0 || (m.Type != "archive" && m.Type != "archive_his") {
		log.Warn("message(%+v) error", m)
		return railgun.MsgPolicyIgnore
	}
	var now = time.Now().Unix()
	if now-m.TimeStamp > 8*60*60 {
		log.Warn("topic message(%+v) too early", m)
		return railgun.MsgPolicyIgnore
	}
	statMsg := &stat.Msg{
		Aid:  m.Aid,
		Type: stat.TypeForCoin,
		Ts:   m.TimeStamp,
		Coin: m.Count,
	}
	s.statCh <- statMsg
	log.Info("CoinSnUpDo TopicMsg got message(%+v)", statMsg)
	return railgun.MsgPolicyNormal
}

func (s *Service) CoinSnRailgunHttp() func(*bm.Context) {
	return s.CoinSnSubV2.BMHandler
}
