package service

import (
	"context"
	"encoding/json"
	"math/rand"
	"strconv"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"

	"go-gateway/app/app-svr/archive-shjd/job/model"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	arcmdl "go-gateway/app/app-svr/archive/service/model"
	"go-gateway/app/app-svr/archive/service/model/videoshot"
)

// UpdateCache is
func (s *Service) UpdateCache(old *model.Archive, nw *model.Archive, action string) {
	var err error
	now := time.Now()
	defer func() {
		dt := time.Since(now)
		bm.MetricServerReqDur.Observe(int64(dt/time.Millisecond), "updateCache", "job")
		bm.MetricServerReqCodeTotal.Inc("updateCache", "job", strconv.FormatInt(int64(ecode.Cause(err).Code()), 10))
		if err == nil {
			s.sendNotify(nw.AID, &model.Notify{Table: _tableArchive, Nw: nw, Old: old, Action: action})
			return
		}
		// retry
		item := &model.RetryItem{
			Old:    old,
			Nw:     nw,
			Tp:     model.TypeForUpdateArchive,
			Action: action,
		}
		if err1 := s.PushItem(context.TODO(), item); err1 != nil {
			log.Error("s.PushItem(%+v) error(%+v)", item, err1)
			return
		}
	}()
	var oldMid int64
	if old != nil && old.Mid != nw.Mid {
		oldMid = old.Mid
	}
	if nw.State >= 0 {
		if err = s.addUpperPassed(context.Background(), nw.AID); err != nil {
			log.Error("s.addUpperPassed err(%+v) aid(%d)", err, nw.AID)
			return
		}
	} else {
		if err = s.delUpperPassedCache(context.Background(), nw.AID, nw.Mid); err != nil {
			log.Error("s.delUpperPassedCache err(%+v) aid(%d) newmid(%d)", err, nw.AID, nw.Mid)
			return
		}
	}
	if oldMid != 0 {
		if err = s.delUpperPassedCache(context.Background(), nw.AID, oldMid); err != nil {
			log.Error("s.delUpperPassedCache err(%+v) aid(%d) oldmid(%d)", err, nw.AID, oldMid)
			return
		}
	}
	arc, ip, err := s.dao.Archive(context.Background(), nw.AID)
	if err != nil || arc == nil {
		log.Error("s.dao.Archive err(%+v) or arc not exist(%d)", err, nw.AID)
		return
	}
	//获取ip地址
	s.transIpv6ToLocation(context.Background(), arc, ip)
	if err = s.setArcCache(context.Background(), arc); err != nil {
		log.Error("s.setArcCache err(%+v) aid(%d)", err, nw.AID)
		return
	}
	vs, err := s.dao.Videos(context.Background(), arc.Aid)
	if err != nil {
		log.Error("s.dao.Videos err(%+v) aid(%d)", err, nw.AID)
		return
	}
	if err = s.setVideosPageCache(context.Background(), nw.AID, vs); err != nil {
		log.Error("s.setVideosPageCache err(%+v) aid(%d)", err, nw.AID)
		return
	}
	if err = s.setSimpleArcCache(context.Background(), arc, vs); err != nil {
		log.Error("s.setSimpleArcCache err(%+v) aid(%d)", err, nw.AID)
		return
	}
	if err = s.initStatCache(context.Background(), nw.AID); err != nil {
		log.Error("s.initStatCache err(%+v) aid(%d)", err, nw.AID)
		return
	}
	log.Warn("s.UpdateCache Gray Aid %d Succ", nw.AID)
}

// UpdateVideoCache is
func (s *Service) UpdateVideoCache(c context.Context, aid, cid int64) {
	var err error
	defer func() {
		if err == nil {
			log.Info("UpdateVideoCache success aid(%d) cid(%d)", aid, cid)
			return
		}
		// retry
		item := &model.RetryItem{
			AID: aid,
			CID: cid,
			Tp:  model.TypeForUpdateVideo,
		}
		if err1 := s.PushItem(context.TODO(), item); err1 != nil {
			log.Error("UpdateVideoCache s.PushItem(%+v) error(%+v)", item, err1)
			return
		}
	}()
	p, err := s.dao.Video(c, aid, cid)
	if err != nil {
		log.Error("UpdateVideoCache Video aid(%d) cid(%d) error(%+v)", aid, cid, err)
		return
	}
	if p == nil {
		return
	}
	if err = s.setVideoCache(c, aid, cid, p); err != nil {
		return
	}
	//由于video信息可以独立更新，所以jd需要更新含所有分p的缓存
	vs, err := s.dao.Videos(c, aid)
	if err != nil {
		return
	}
	if err = s.setVideosPageCache(c, aid, vs); err != nil {
		return
	}
}

// DelVideoCache del video cache
func (s *Service) DelVideoCache(c context.Context, aid, cid int64) {
	var err error
	defer func() {
		if err == nil {
			log.Info("DelVideoCache success aid(%d) cid(%d)", aid, cid)
			return
		}
		// retry
		item := &model.RetryItem{
			AID: aid,
			CID: cid,
			Tp:  model.TypeForDelVideo,
		}
		if err1 := s.PushItem(context.TODO(), item); err1 != nil {
			log.Error("DelVideoCache s.PushItem(%+v) error(%+v)", item, err1)
			return
		}
	}()
	for k, pool := range s.arcRedises {
		if err = func() (err error) {
			conn := pool.Get(c)
			defer conn.Close()
			if _, err = conn.Do("DEL", arcmdl.VideoKey(aid, cid)); err != nil {
				log.Error("DelVideoCache k(%d) aid(%d) cid(%d) err(%+v)", k, aid, cid, err)
				return err
			}
			return nil
		}(); err != nil {
			return
		}
	}
	if err = s.delTaishan(c, [][]byte{[]byte(arcmdl.VideoKey(aid, cid))}); err != nil {
		log.Error("%+v", err)
		return
	}
}

// addVideoShotCache is
func (s *Service) addVideoShotCache(c context.Context, cid, count, hdCnt, sdCnt int64, hdImg, sdImg string) {
	var err error
	defer func() {
		if err == nil {
			log.Info("addVideoShotCache(%d) count(%d) hdCnt(%d) hdImg(%s) success", cid, count, hdCnt, hdImg)
			return
		}
		// retry
		item := &model.RetryItem{
			CID:   cid,
			Count: count,
			HdCnt: hdCnt,
			HdImg: hdImg,
			Tp:    model.TypeForVideoShot,
			SdCnt: sdCnt,
			SdImg: sdImg,
		}
		if err1 := s.PushItem(context.TODO(), item); err1 != nil {
			log.Error("ClearVideoShotCache s.PushItem(%+v) error(%+v)", item, err1)
			return
		}
	}()
	vs := &videoshot.Videoshot{Cid: cid, Count: count, HDImg: hdImg, HDCount: hdCnt, SdCount: sdCnt, SdImg: sdImg}
	vsBs, err := json.Marshal(vs)
	if err != nil {
		log.Error("json.Marshal err(%+v) vs(%+v)", err, vs)
		return
	}
	for k, pool := range s.sArcRds {
		if err = func() (err error) {
			conn := pool.Get(c)
			defer conn.Close()
			if _, err = conn.Do("SET", arcmdl.NewVideoShotKey(cid), vsBs); err != nil {
				log.Error("addVideoShotCache(%d) k(%d) err(%+v)", cid, k, err)
				return err
			}
			return nil
		}(); err != nil {
			return
		}
	}
}

// nolint:bilirailguncheck
func (s *Service) sendNotify(aid int64, msg *model.Notify) {
	for i := 0; i < 10; i++ {
		if err := s.notifyPub.Send(context.Background(), strconv.FormatInt(aid, 10), msg); err == nil {
			msgStr, _ := json.Marshal(msg)
			log.Info("sendNotify(%s) successed", msgStr)
			break
		}
		time.Sleep(time.Millisecond * 10)
	}
}

func (s *Service) delRedirectCache(ctx context.Context, aid int64) {
	err := s.delTaishan(ctx, [][]byte{[]byte(arcmdl.RedirectKey(aid))})
	if err != nil {
		log.Error("delRedirectTaishan is error %+v %+v", aid, err)
		return
	}
}

func (s *Service) internalCacheHandler(c context.Context, aid int64) {
	var (
		err error
	)
	defer func() { // multi error retry only once
		if err != nil { // retry
			item := &model.RetryItem{
				AID: aid,
				Tp:  model.TypeForInternal,
			}
			if err1 := s.PushItem(c, item); err1 != nil {
				log.Error("internalCacheHandler s.PushItem(%+v) error(%+v)", item, err1)
				return
			}
		}
	}()
	log.Info("internalCacheHandler start(%d)", aid)
	//获取数据库中的数据,正常逻辑数据库中一定存在
	var row *arcgrpc.ArcInternal
	if row, err = s.dao.RawInternal(c, aid); err != nil { //查询错误重试
		return
	}
	if row == nil { //未知错误，终止
		return
	}
	//更新redis
	if err = s.setInternalCache(c, row); err != nil {
		return
	}
	log.Info("internalCacheHandler success(%d)", aid)
}

func (s *Service) setInternalCache(c context.Context, in *arcgrpc.ArcInternal) error {
	if in == nil {
		return nil
	}
	bs, err := in.Marshal()
	if err != nil { //json错误无法重试
		log.Error("日志告警:setInternalCache Marshal aid(%d) attribute(%d) err(%+v) ", in.Aid, in.Attribute, err)
		return nil
	}
	//暂定使用原有redis集群
	// 缓存设置过期时间，默认10小时+48小时内随机数
	rand.Seed(time.Now().UnixNano())
	exp := int64(36000)
	arcExp := exp + rand.Int63n(172800)
	for k, rds := range s.sArcRds {
		if err = func() error {
			conn := rds.Get(c)
			defer conn.Close()
			_, e := conn.Do("SET", arcmdl.InternalArcKey(in.Aid), bs, "EX", arcExp)
			return e
		}(); err != nil {
			log.Error("setInternalCache conn.Do key(%s) k(%d) err(%+v)", arcmdl.InternalArcKey(in.Aid), k, err)
			return err
		}
	}
	return s.setTaishan(c, []byte(arcmdl.InternalArcKey(in.Aid)), bs)
}

func (s *Service) transIpv6ToLocation(c context.Context, arc *arcgrpc.Arc, ip string) {
	if len(ip) == 0 {
		return
	}
	res, err := s.locDao.Info2WithRetry(c, ip)
	if err == nil && res != nil {
		arc.PubLocation = res.Show
	}
}
