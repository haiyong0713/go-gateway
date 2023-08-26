package service

import (
	"context"
	"encoding/json"
	"math/rand"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	"go-common/library/database/taishan"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/railgun"

	"go-gateway/app/app-svr/archive/job/model/archive"
	jobmdl "go-gateway/app/app-svr/archive/job/model/databus"
	"go-gateway/app/app-svr/archive/job/model/result"
	"go-gateway/app/app-svr/archive/job/model/retry"
	"go-gateway/app/app-svr/archive/service/api"
	arcmdl "go-gateway/app/app-svr/archive/service/model"

	"go-common/library/sync/errgroup.v2"
)

const (
	_steinsRouteForStickVideo = "stick_video"
)

func (s *Service) VideoUpUnpack(msg railgun.Message) (*railgun.SingleUnpackMsg, error) {
	m := &jobmdl.Videoup{}
	if err := json.Unmarshal(msg.Payload(), m); err != nil {
		log.Error("json.Unmarshal(%+v) error(%+v)", string(msg.Payload()), err)
		return nil, err
	}
	return &railgun.SingleUnpackMsg{
		Group: m.Aid,
		Item:  m,
	}, nil
}

func (s *Service) VideoUpDo(ctx context.Context, item interface{}) railgun.MsgPolicy {
	m := item.(*jobmdl.Videoup)
	if m == nil {
		return railgun.MsgPolicyIgnore
	}
	log.Info("videoupMessage start %+v", m)
	if m.Aid <= 0 && m.Cid <= 0 {
		log.Warn("aid(%d) <= 0  && cid(%d) <= 0", m.Aid, m.Cid)
		return railgun.MsgPolicyIgnore
	}
	if m.Timestamp != 0 {
		if gap := time.Now().Unix() - m.Timestamp; gap > s.c.Custom.DBAlertSec {
			log.Error("日志告警 视频过审消息堆积预警 当前消费可能延迟大于(%d)秒，对应消息 %+v", gap, m)
		}
	}
	s.Prom.Incr(m.Route)
	switch m.Route {
	case jobmdl.RouteAutoOpen, jobmdl.RouteDelayOpen, jobmdl.RouteDeleteArchive, jobmdl.RouteSecondRound, jobmdl.RouteFirstRoundForbid, jobmdl.RouteForceSync, jobmdl.RoutePremierePass:
		arc, _ := s.archiveDao.RawArchive(context.Background(), m.Aid)
		if arc != nil && arc.AttrVal(api.AttrBitIsPGC) == api.AttrYes {
			select {
			case s.pgcAidChan <- m.Aid:
			default:
				s.pushChanFail(m.Aid, "PGC")
			}
		}
		if m.UpFrom == archive.UpFromAnnualReport { //年报投稿
			select {
			case s.nbAidChan <- m.Aid:
			default:
				s.pushChanFail(m.Aid, "年报")
			}
		}
		select {
		case s.ugcAidChan <- m.Aid:
		default:
			s.pushChanFail(m.Aid, "UGC")
		}
	case jobmdl.RouteVideoShotChanged:
		s.putVideoShotChan(context.Background(), m.CIDs)
	case jobmdl.RouteVideoFF:
		s.putVideoFFChan(context.Background(), m.Cid)
	}
	return railgun.MsgPolicyNormal
}

func (s *Service) VideoUpService(req *jobmdl.Videoup) {
	if req.Aid <= 0 && req.Cid <= 0 {
		log.Warn("aid(%d) <= 0 && cid(%d) <= 0", req.Aid, req.Cid)
		return
	}
	if req.Timestamp != 0 {
		if gap := time.Now().Unix() - req.Timestamp; gap > s.c.Custom.DBAlertSec {
			log.Error("日志告警 视频过审消息堆积预警 当前消费可能延迟大于(%d)秒，对应消息 %+v", gap, req)
		}
	}
	s.Prom.Incr(req.Route)
	switch req.Route {
	case jobmdl.RouteAutoOpen, jobmdl.RouteDelayOpen, jobmdl.RouteDeleteArchive, jobmdl.RouteSecondRound, jobmdl.RouteFirstRoundForbid, jobmdl.RouteForceSync:
		arc, _ := s.archiveDao.RawArchive(context.Background(), req.Aid)
		if arc != nil && arc.AttrVal(api.AttrBitIsPGC) == api.AttrYes {
			select {
			case s.pgcAidChan <- req.Aid:
			default:
				s.pushChanFail(req.Aid, "PGC")
			}
			return
		}
		if req.UpFrom == archive.UpFromAnnualReport { //年报投稿
			select {
			case s.nbAidChan <- req.Aid:
			default:
				s.pushChanFail(req.Aid, "年报")
			}
			return
		}
		select {
		case s.ugcAidChan <- req.Aid:
		default:
			s.pushChanFail(req.Aid, "UGC")
		}
		return
	case jobmdl.RouteVideoShotChanged:
		s.putVideoShotChan(context.Background(), req.CIDs)
	case jobmdl.RouteVideoFF:
		s.putVideoFFChan(context.Background(), req.Cid)
	}
}

func (s *Service) pushChanFail(aid int64, desc string) {
	rt := &retry.Info{Action: retry.FailResultAdd}
	rt.Data.ArcAction = _arcVideosUpdate
	rt.Data.Aid = aid
	s.PushFail(context.TODO(), rt, retry.FailList)
	log.Error("日志告警 视频过审消息堆积预警 %s Chan is full aid(%d)", desc, aid)
}

func (s *Service) delVideoCache(c context.Context, aid int64, cids []int64) {
	if aid == 0 || len(cids) == 0 {
		return
	}
	var err error
	defer func() {
		if err != nil {
			log.Error("delVideoCache aid:%d cids:%+v error:%+v", aid, cids, err)
			rt := &retry.Info{Action: retry.FailDelVideoCache}
			rt.Data.Aid = aid
			rt.Data.Cids = cids
			s.PushFail(context.TODO(), rt, retry.FailList)
			return
		}
	}()
	args := redis.Args{}
	var taiKeys []*taishan.Record
	for _, cid := range cids {
		args = args.Add(arcmdl.VideoKey(aid, cid))
		taiKeys = append(taiKeys, &taishan.Record{Key: []byte(arcmdl.VideoKey(aid, cid))})
	}
	for k, pool := range s.arcRedises {
		if err = func() error {
			for i := 0; i < len(args); i += _maxMSET {
				conn := pool.Get(c)
				var partDel redis.Args
				if i+_maxMSET > len(args) {
					partDel = args[i:]
				} else {
					partDel = args[i : i+_maxMSET]
				}
				_, err := conn.Do("DEL", partDel...)
				conn.Close()
				if err != nil {
					log.Error("delVideoCache conn.Do(DEL) k(%d) aid(%d) cids(%+v) err(%+v)", k, aid, partDel, err)
					return err
				}
			}
			return nil
		}(); err != nil {
			return
		}
	}
	if err = s.delTaishan(c, taiKeys); err != nil {
		return
	}
	log.Info("delVideoCache success aid(%d) cids(%+v)", aid, cids)
}

func (s *Service) upVideoCache(c context.Context, aid int64) {
	var err error
	defer func() {
		if err != nil {
			log.Error("%+v", err)
			rt := &retry.Info{Action: retry.FailUpVideoCache}
			rt.Data.Aid = aid
			s.PushFail(context.TODO(), rt, retry.FailList)
			return
		}
	}()
	pages, err := s.resultDao.RawVideos(c, aid)
	if err != nil || len(pages) == 0 {
		return
	}

	kvMap := make(map[string][]byte, len(pages))
	for _, p := range pages {
		if p == nil {
			continue
		}
		var pb []byte
		if pb, err = p.Marshal(); err != nil {
			log.Error("upVideoCache Marshal error(%+v) p(%+v)", err, p)
			err = nil
			continue
		}
		kvMap[arcmdl.VideoKey(aid, p.Cid)] = pb
	}

	//写缓存
	rand.Seed(time.Now().UnixNano())
	exp := int64(36000) + rand.Int63n(172800)

	s.redisMSetWithExp(c, kvMap, exp)
	s.taishanBatchSet(c, kvMap)
}

func (s *Service) SteinsGateUnpack(msg railgun.Message) (*railgun.SingleUnpackMsg, error) {
	m := &archive.SteinsCid{}
	if err := json.Unmarshal(msg.Payload(), m); err != nil {
		log.Error("json.Unmarshal(%+v) error(%+v)", string(msg.Payload()), err)
		return nil, err
	}
	return &railgun.SingleUnpackMsg{
		Group: m.Aid,
		Item:  m,
	}, nil
}

func (s *Service) SteinsGateDo(ctx context.Context, item interface{}) railgun.MsgPolicy {
	m := item.(*archive.SteinsCid)
	if m == nil {
		return railgun.MsgPolicyIgnore
	}
	log.Info("steinsGateSub (%+v) start", m)
	switch m.Route {
	case _steinsRouteForStickVideo:
		if m.Aid <= 0 || m.Cid <= 0 {
			log.Warn("Data Illegal Aid %d, Cid %d", m.Aid, m.Cid)
			return railgun.MsgPolicyIgnore
		}
		s.steinsHandler(m)
	default:
		log.Warn("steinsGate route(%s) error", m.Route)
		return railgun.MsgPolicyIgnore
	}
	log.Info("steinsGateSub (%+v) finish", m)
	return railgun.MsgPolicyNormal
}

func (s *Service) transStein(msg *archive.SteinsCid) (oldResult *api.Arc, changed bool, err error) {
	var (
		tx                     *sql.Tx
		c                      = context.Background()
		arcNb, sortNb, stickNb int64
	)
	if oldResult, _, err = s.resultDao.RawArc(c, msg.Aid); err != nil {
		log.Error("s.Result.Archive Aid %d Err %+v", msg.Aid, err)
		return
	}
	if oldResult == nil || oldResult.AttrVal(api.AttrBitSteinsGate) != api.AttrYes {
		log.Error("steinsHandler Aid %d Not SteinsGate!", msg.Aid)
		return
	}
	if tx, err = s.resultDao.BeginTran(c); err != nil {
		log.Error("s.result.BeginTran error(%+v)", err)
		return
	}
	if sortNb, err = s.resultDao.TxSortVideos(tx, msg.Aid, msg.Cid); err != nil {
		_ = tx.Rollback()
		log.Error("s.result.TxSortVideos Aid %d Cid %d error(%+v)", msg.Aid, msg.Cid, err)
		return
	}
	if stickNb, err = s.resultDao.TxStickVideo(tx, msg.Aid, msg.Cid); err != nil {
		_ = tx.Rollback()
		log.Error("s.result.TxStickVideo Aid %d Cid %d error(%+v)", msg.Aid, msg.Cid, err)
		return
	}
	if arcNb, err = s.resultDao.TxUpArcFirstCID(tx, msg.Aid, msg.Cid); err != nil {
		_ = tx.Rollback()
		log.Error("s.result.TxUpArcFirstCID Aid %d Cid %d error(%+v)", msg.Aid, msg.Cid, err)
		return
	}
	if err = tx.Commit(); err != nil {
		_ = tx.Rollback()
		log.Error("s.result.Commit Aid %d Cid %d error(%+v)", msg.Aid, msg.Cid, err)
		return
	}
	if arcNb > 0 || stickNb > 0 || sortNb > 0 {
		changed = true
	}
	log.Warn("SteinsHandler Aid %d, Cid %d ArcNb %d, StickNb %d, SortNb %d", msg.Aid, msg.Cid, arcNb, stickNb, sortNb)
	return
}

func (s *Service) steinsHandler(msg *archive.SteinsCid) {
	var (
		changed              bool
		err                  error
		oldResult, newResult *api.Arc
		c                    = context.Background()
	)
	if oldResult, changed, err = s.transStein(msg); err != nil {
		return
	}
	if !changed {
		log.Warn("SteinsHandler Aid %d, Cid %d Not Changed", msg.Aid, msg.Cid)
		return
	}
	if newResult, _, err = s.resultDao.RawArc(c, msg.Aid); err != nil || newResult == nil {
		log.Error("s.Result.Archive Aid %d Err %+v", msg.Aid, err)
		return
	}
	s.updateResultCache(newResult, oldResult, false)                                                       // update archive cache
	s.upVideoCache(c, msg.Aid)                                                                             // update video cache
	s.sendNotify(&result.ArchiveUpInfo{Table: "archive", Action: "update", Nw: newResult, Old: oldResult}) // send archive-notify T
}

func (s *Service) putVideoShotChan(c context.Context, cids []int64) {
	if len(cids) == 0 {
		return
	}
	select {
	case s.videoShotChan <- cids:
	default:
		s.Prom.Incr("videoShotChan Full")
		rt := &retry.Info{Action: retry.FailVideoShot}
		rt.Data.Cids = cids
		s.PushFail(c, rt, retry.FailVideoshotList)
	}
}

func (s *Service) videoShotproc() {
	defer s.waiter.Done()
	for {
		cids, ok := <-s.videoShotChan
		if !ok {
			return
		}
		s.videoShotHandler(context.Background(), cids)
	}
}

// videoShotHandler is
func (s *Service) videoShotHandler(c context.Context, cids []int64) {
	if len(cids) == 0 {
		return
	}
	vs, err := s.archiveDao.RawVideoShots(c, cids)
	if err != nil {
		log.Error("videoShot cids(%+v) s.archiveDao.RawVideoShots err(%+v)", cids, err)
		rt := &retry.Info{Action: retry.FailVideoShot}
		rt.Data.Cids = cids
		s.PushFail(context.TODO(), rt, retry.FailVideoshotList)
		return
	}
	g := errgroup.Group{}
	for _, v := range vs {
		if v.Count < 1 { //有高清缩略图时一定有普通缩略图
			log.Warn("videoShot cid(%d) count is zero", v.Cid)
			continue
		}
		ccid := v.Cid
		ccnt := v.Count
		hdCnt := v.HDCount
		hdImg := v.HDImg
		sdCnt := v.SdCount
		sdImg := v.SdImg
		g.Go(func(ctx context.Context) error {
			if err := func() (err error) {
				if err = s.resultDao.CheckVideoShot(ctx, ccid, ccnt); err != nil {
					if !ecode.EqualError(ecode.NothingFound, err) {
						s.Prom.Incr("videoShot-bfs查询异常")
						return err
					}
					s.Prom.Incr("videoShot-bfs不存在")
					if err = s.resultDao.DelVideoShot(context.Background(), ccid); err != nil {
						log.Error("videoShot cid(%d) s.resultDao.DelVideoShot err(%+v)", ccid, err)
						return err
					}
					if err = s.addVideoShotCache(context.Background(), ccid, 0, 0, 0, "", ""); err != nil {
						log.Error("videoShot cid(%d) s.addVideoShotCache err(%+v)", ccid, err)
						return err
					}
					return nil
				}
				if err = s.resultDao.AddVideoShot(ctx, ccid, ccnt, hdCnt, sdCnt, hdImg, sdImg); err != nil {
					log.Error("videoShot cid(%d) s.resultDao.AddVideoShot err(%+v)", ccid, err)
					return err
				}
				if err = s.addVideoShotCache(ctx, ccid, ccnt, hdCnt, sdCnt, hdImg, sdImg); err != nil {
					log.Error("videoShot cid(%d) s.addVideoShotCache err(%+v)", ccid, err)
					return err
				}
				log.Info("videoShotHandler cid(%d) cnt(%d) hdCnt(%d) hdImg(%s) success", ccid, ccnt, hdCnt, hdImg)
				return nil
			}(); err != nil {
				rt := &retry.Info{Action: retry.FailVideoShot}
				rt.Data.Cids = []int64{ccid}
				s.PushFail(context.Background(), rt, retry.FailVideoshotList)
				return nil
			}
			return nil
		})
	}
	if err = g.Wait(); err != nil {
		log.Error("videoShotHandler cids(%+v) g.Wait() error(%+v)", cids, err)
	}
}

// nolint:gocognit
func (s *Service) rebuildVideoShot() {
	const _vsKey = "rrrrrrrvsid"
	conn := s.redis.Get(context.Background())
	lastID, err := redis.Int64(conn.Do("GET", _vsKey))
	conn.Close()
	if err != nil {
		if err != redis.ErrNil {
			log.Error("%+v", err)
			return
		}
		lastID = 0
	}
	if lastID == 0 {
		id, err := s.resultDao.MaxVideoShotID(context.Background())
		if err != nil {
			log.Error("%+v", err)
			return
		}
		lastID = id
	}
	if lastID <= 0 {
		log.Warn("rebuildVideoShot id is zero")
		return
	}
	for i := 0; i < 8; i++ {
		// nolint:biligowordcheck
		go func() {
			for {
				time.Sleep(10 * time.Millisecond)
				cids, err := func() ([]int64, error) {
					conn := s.redis.Get(context.Background())
					defer conn.Close()
					for j := 0; j < 20; j++ {
						if err := conn.Send("DECR", _vsKey); err != nil {
							return nil, err
						}
					}
					if err = conn.Flush(); err != nil {
						return nil, err
					}
					var cids []int64
					for j := 0; j < 20; j++ {
						cid, err := redis.Int64(conn.Receive())
						if err != nil {
							log.Error("rebuildVideoShot err %+v", err)
							continue
						}
						if cid < 0 {
							continue
						}
						cids = append(cids, cid)
					}
					return cids, nil
				}()
				if err != nil {
					log.Error("rebuildVideoShot err %+v", err)
					continue
				}
				if len(cids) == 0 {
					break
				}
				log.Info("rebuildVideoShot cid(%+v)", cids)
				s.videoShotHandler(context.Background(), cids)
			}
		}()
	}
}

func (s *Service) putVideoFFChan(c context.Context, cid int64) {
	if cid <= 0 {
		return
	}
	select {
	case s.videoFFChan <- cid:
	default:
		s.Prom.Incr("videoFFChan Full")
		rt := &retry.Info{Action: retry.FailVideoFF}
		rt.Data.Cid = cid
		s.PushFail(c, rt, retry.FailVideoFF)
	}
}

func (s *Service) videoFFProc() {
	defer s.waiter.Done()
	for {
		cid, ok := <-s.videoFFChan
		if !ok {
			return
		}
		s.videoFFHandler(context.Background(), cid)
	}
}

// videoFFHandler is
// up db archive / archive_video
// up redis a3p_aid / psb_aid / psb_aid_cid
// up taishan
func (s *Service) videoFFHandler(c context.Context, cid int64) {
	if cid <= 0 {
		return
	}
	var err error
	defer func() {
		if err != nil {
			rt := &retry.Info{Action: retry.FailVideoFF}
			rt.Data.Cid = cid
			s.PushFail(context.TODO(), rt, retry.FailVideoFFList)
		}
	}()
	log.Info("videoFFHandler start cid(%d)", cid)
	ff, err := s.archiveDao.RawVideoFistFrame(c, cid)
	if err != nil {
		log.Error("videoFFHandler RawVideoFistFrame cid(%d) error(%+v) or ff is empty", cid, err)
		return
	}
	firstFrame := ff.FirstFrame
	// get result.video
	aid, video, err := s.resultDao.RawVideoFistFrame(c, cid)
	if err != nil || video == nil {
		log.Error("videoFFHandler RawVideoFistFrame cid(%d) error(%+v)", cid, err)
		return
	}
	// get result.archive
	resultArc, ip, err := s.resultDao.RawArc(c, aid)
	if err != nil || resultArc == nil {
		log.Error("videoFFHandler RawArc aid(%d) cid(%d) error(%+v)", aid, cid, err)
		return
	}
	if err = s.resultDao.UpVideoFF(c, cid, firstFrame); err != nil {
		log.Error("videoFFHandler UpVideoFF aid(%d) cid(%d) ff(%s) error(%+v)", aid, cid, firstFrame, err)
		return
	}
	if err = s.setVideoCache(c, aid, cid, video); err != nil {
		log.Error("videoFFHandler setVideoCache aid(%d) cid(%d) video(%+v) error(%+v)", aid, cid, video, err)
		return
	}
	videos, err := s.resultDao.RawVideos(c, aid)
	if err != nil {
		log.Error("videoFFHandler RawVideos aid(%d) cid(%d) error(%+v)", aid, cid, err)
		return
	}
	if err = s.setVideosPageCache(c, aid, videos); err != nil {
		log.Error("videoFFHandler setVideosPageCache aid(%d) cid(%d) error(%+v)", aid, cid, err)
		return
	}
	// 如果是第1p需更新archive相关数据
	if resultArc.FirstCid == cid {
		if err = s.resultDao.UpArcFF(c, aid, firstFrame); err != nil {
			log.Error("videoFFHandler UpArcFF aid(%d) cid(%d) ff(%s) error(%+v)", aid, cid, firstFrame, err)
			return
		}
		resultArc.FirstFrame = api.FFURL(firstFrame)
		//获取ip地址
		s.transIpv6ToLocation(c, resultArc, ip)
		if err = s.setArcCache(c, resultArc); err != nil {
			log.Error("videoFFHandler setArcCache aid(%d) cid(%d) error(%+v)", aid, cid, err)
			return
		}
	}
	log.Info("videoFFHandler success cid(%d) ff(%s)", cid, firstFrame)
}
