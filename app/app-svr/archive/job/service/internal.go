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
	"go-common/library/railgun"
	jobmdl "go-gateway/app/app-svr/archive/job/model/databus"
	"go-gateway/app/app-svr/archive/job/model/retry"
	achmdl "go-gateway/app/app-svr/archive/service/api"
	apimdl "go-gateway/app/app-svr/archive/service/api"
	arcmdl "go-gateway/app/app-svr/archive/service/model"

	serGRPC "git.bilibili.co/bapis/bapis-go/content-flow-control/service"
)

func (s *Service) InternalSubUnpack(msg railgun.Message) (*railgun.SingleUnpackMsg, error) {
	m := &jobmdl.InternalMessage{}
	if err := json.Unmarshal(msg.Payload(), m); err != nil {
		log.Error("json.Unmarshal(%+v) error(%+v)", string(msg.Payload()), err)
		return nil, err
	}
	return &railgun.SingleUnpackMsg{
		Group: m.Data.Oid,
		Item:  m,
	}, nil
}

func (s *Service) InternalSubProcDo(ctx context.Context, item interface{}) railgun.MsgPolicy {
	if item == nil {
		return railgun.MsgPolicyIgnore
	}
	m := item.(*jobmdl.InternalMessage)
	if m == nil {
		return railgun.MsgPolicyIgnore
	}
	switch m.Router {
	case "archive-service":
		//更新数据库和缓存redis taishan
		//更新失败进失败队列，进入重试逻辑
		s.internalUpdate(m.Data.Oid)
	default:
	}
	return railgun.MsgPolicyNormal
}

func (s *Service) internalUpdate(aid int64) {
	var (
		err error
		c   = context.Background()
		now = time.Now()
	)
	if aid <= 0 { //过滤小于等于0的数据
		return
	}
	defer func() { // multi error retry only once
		bm.MetricServerReqDur.Observe(int64(time.Since(now)/time.Millisecond), "internalUpdate", "job")
		bm.MetricServerReqCodeTotal.Inc("internalUpdate", "job", strconv.FormatInt(int64(ecode.Cause(err).Code()), 10))
		if err != nil {
			rt := &retry.Info{Action: retry.FailUpInternal}
			rt.Data.Aid = aid
			s.PushFail(c, rt, retry.FailInternalList)
			log.Error("internalUpdate aid(%d) error(%+v)", aid, err)
		}
	}()
	//test retry
	log.Info("internalUpdate start aid(%d)", aid)
	var attrInfo []*serGRPC.InfoItem
	if attrInfo, err = s.controlDao.GetInternalAttr(c, aid); err != nil {
		if err == ecode.RequestErr { //-400 错误码不需要重试
			log.Error("internalUpdate not need update aid(%d)", aid)
			err = nil
		}
		return
	}
	log.Info("internalUpdate update db aid(%d) result(%+v)", aid, attrInfo)
	//是否存在数据库中
	var row *achmdl.ArcInternal
	if row, err = s.resultDao.RawInternal(c, aid); err != nil { //查询错误retry
		return
	}
	if row == nil {
		out := &achmdl.ArcInternal{Aid: aid, Attribute: ModifyAttr(0, attrInfo)}
		err = s.resultDao.AddInternal(c, out)
	} else {
		out := &achmdl.ArcInternal{Aid: aid, Attribute: ModifyAttr(row.Attribute, attrInfo)}
		err = s.resultDao.UpInternal(c, out)
	}
	if err != nil { //更新失败
		return
	}
	log.Info("internalUpdate success aid(%d)", aid)
	//更新缓存 数据库已经处理
	s.internalCache(aid)
}

func ModifyAttr(old int64, rly []*serGRPC.InfoItem) int64 {
	var (
		attr = old
		bit  uint
	)
	for _, v := range rly {
		switch v.Key {
		case "54": //oversea_block
			bit = apimdl.InterAttrBitOverseaLock
		default:
			continue
		}
		if v.Value == 1 { // open
			attr = attr | (1 << bit)
		} else { //close
			attr = attr &^ (1 << bit)
		}
	}
	return attr
}

func (s *Service) internalCache(aid int64) {
	c := context.Background()
	select {
	case s.internalChan <- aid:
	default:
		s.Prom.Incr("internalChan Full")
		rt := &retry.Info{Action: retry.FailInternalCache}
		rt.Data.Aid = aid
		s.PushFail(c, rt, retry.FailInternalList)
		log.Error("internalChan aid(%d)  error(internalChan Full)", aid)
	}
}

func (s *Service) internalCacheProc() {
	defer s.waiter.Done()
	for {
		in, ok := <-s.internalChan
		if !ok {
			log.Error("internalCacheProc exit")
			return
		}
		s.internalCacheHandler(context.Background(), in)
	}
}

func (s *Service) internalCacheHandler(c context.Context, aid int64) {
	var (
		err error
		now = time.Now()
	)
	defer func() { // multi error retry only once
		bm.MetricServerReqDur.Observe(int64(time.Since(now)/time.Millisecond), "internalCacheHandler", "job")
		bm.MetricServerReqCodeTotal.Inc("internalCacheHandler", "job", strconv.FormatInt(int64(ecode.Cause(err).Code()), 10))
		if err != nil {
			rt := &retry.Info{Action: retry.FailInternalCache}
			rt.Data.Aid = aid
			//rt.Data.ArcAction = ftype
			s.PushFail(c, rt, retry.FailInternalList)
			log.Error("internalCacheHandler aid(%d) error(%+v)", aid, err)
		}
	}()
	log.Info("internalCacheHandler start(%d)", aid)
	//获取数据库中的数据,正常逻辑数据库中一定存在
	var row *achmdl.ArcInternal
	if row, err = s.resultDao.RawInternal(c, aid); err != nil { //查询错误重试
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

func (s *Service) setInternalCache(c context.Context, in *achmdl.ArcInternal) error {
	if in == nil {
		return nil
	}
	bs, err := in.Marshal()
	if err != nil { //错误无法重试
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
