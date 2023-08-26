package service

import (
	"context"
	"encoding/json"
	"time"

	"go-common/library/log"
	"go-common/library/railgun"

	jobmdl "go-gateway/app/app-svr/archive/job/model/databus"
	"go-gateway/app/app-svr/archive/service/model"
	"go-gateway/app/app-svr/archive/service/model/videoshot"
)

func (s *Service) CacheSubProcUnpack(msg railgun.Message) (*railgun.SingleUnpackMsg, error) {
	m := &jobmdl.Rebuild{}
	if err := json.Unmarshal(msg.Payload(), m); err != nil {
		log.Error("json.Unmarshal(%+v) error(%+v)", string(msg.Payload()), err)
		return nil, err
	}
	return &railgun.SingleUnpackMsg{
		Group: m.Aid,
		Item:  m,
	}, nil
}

func (s *Service) CacheSubProcDo(ctx context.Context, item interface{}) railgun.MsgPolicy {
	if item == nil {
		return railgun.MsgPolicyIgnore
	}
	m := item.(*jobmdl.Rebuild)
	if m == nil {
		return railgun.MsgPolicyIgnore
	}
	for i := 0; i < 10; i++ {
		err := func() error {
			arc, ip, err := s.resultDao.RawArc(ctx, m.Aid)
			if err != nil || arc == nil {
				log.Error("RawArc err(%+v) or aid not exist(%d)", err, m.Aid)
				return err
			}
			//获取ip地址
			s.transIpv6ToLocation(ctx, arc, ip)
			if err = s.setArcCache(ctx, arc); err != nil {
				return err
			}
			return nil
		}()
		if err == nil {
			log.Info("aid(%d) rebuild cache ok", m.Aid)
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	return railgun.MsgPolicyNormal
}

// addVideoShotCache is
func (s *Service) addVideoShotCache(c context.Context, cid, count, hdCnt, sdCnt int64, hdImg, sdImg string) (err error) {
	vs := &videoshot.Videoshot{Cid: cid, Count: count, HDImg: hdImg, HDCount: hdCnt,
		SdCount: sdCnt, SdImg: sdImg}
	vsBs, err := json.Marshal(vs)
	if err != nil {
		log.Error("json.Marshal err(%+v) vs(%+v)", err, vs)
		return
	}
	for k, pool := range s.sArcRds {
		if err = func() (err error) {
			conn := pool.Get(c)
			defer conn.Close()
			if _, err = conn.Do("SET", model.NewVideoShotKey(cid), vsBs); err != nil {
				log.Error("addVideoShotCache(%d) k(%d) err(%+v)", cid, k, err)
				return err
			}
			return nil
		}(); err != nil {
			return err
		}
	}
	return
}
