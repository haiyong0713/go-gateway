package service

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	cfcgrpc "git.bilibili.co/bapis/bapis-go/content-flow-control/service"
	"go-common/library/ecode"
	"go-common/library/log"
	arcmdl "go-gateway/app/app-svr/archive/service/api"
	dyapi "go-gateway/app/web-svr/dynamic/service/api/v1"
	"go-gateway/app/web-svr/web/interface/model"

	"github.com/pkg/errors"
)

const (
	_onlinePubdateLimit = 86400
	_onlineDanmuLimit   = 3
)

// OnlineArchiveCount Get Archive Count.
func (s *Service) OnlineArchiveCount(_ context.Context) (rs *model.Online) {
	rs = &model.Online{
		RegionCount: s.regionCount,
	}
	return
}

// OnlineList online archive list.
// nolint: gocognit
func (s *Service) OnlineList(ctx context.Context) ([]*model.OnlineArc, error) {
	arcs, err := func() ([]*model.OnlineArc, error) {
		var aids []int64
		for _, v := range s.onlineAids {
			if v != nil && v.Aid > 0 {
				aids = append(aids, v.Aid)
			}
		}
		if len(aids) == 0 {
			return nil, errors.New("OnlineList aids nil")
		}
		arcs, cfcInfos, arcErr := s.batchArchivesAndCfcInfos(ctx, aids)
		if arcErr != nil {
			return nil, arcErr
		}
		var list []*model.OnlineArc
		for _, v := range s.onlineAids {
			if v == nil {
				continue
			}
			var (
				cfcItem []*cfcgrpc.ForbiddenItem
				ok      bool
			)
			if cfcItem, ok = cfcInfos[v.Aid]; !ok {
				log.Warn("s.OnlineList forbidden is empty aid:%d", v.Aid)
			}
			arcForbidden := model.ItemToArcForbidden(cfcItem)
			if arc, ok := arcs[v.Aid]; ok && arc != nil && arc.IsNormal() && !arcForbidden.NoRank {
				if arc.AttrVal(arcmdl.AttrBitIsBangumi) == arcmdl.AttrNo && arc.AttrVal(arcmdl.AttrBitIsMovie) == arcmdl.AttrNo {
					if time.Now().Unix()-int64(arc.PubDate) > _onlinePubdateLimit {
						if arc.Stat.Danmaku == 0 || v.Count/int64(arc.Stat.Danmaku) > _onlineDanmuLimit {
							continue
						}
					}
				}
				model.ClearAttrAndAccess(arc)
				list = append(list, &model.OnlineArc{Arc: arc, Bvid: s.avToBv(v.Aid), OnlineCount: v.Count})
			}
			if len(list) >= s.c.WEB.OnlineCount {
				break
			}
		}
		if len(list) == 0 {
			return nil, errors.New(fmt.Sprintf("OnlineList list count:%d error", len(list)))
		}
		return list, nil
	}()
	if err != nil {
		log.Error("OnlineList error:%v", err)
		//get from remote cache
		res, err := s.dao.OnlineListBakCache(ctx)
		if err != nil {
			log.Error("日志告警 OnlineList OnlineListBakCache error:%v", err)
			return []*model.OnlineArc{}, nil
		}
		return res, nil
	}
	if err := s.cache.Do(ctx, func(ctx context.Context) {
		if err := s.dao.SetOnlineListBakCache(ctx, arcs); err != nil {
			log.Error("%+v", err)
		}
	}); err != nil {
		log.Error("%+v", err)
	}
	return arcs, nil
}

func (s *Service) OnlineTotal(_ context.Context, token string) (int64, error) {
	if token != s.c.Rule.OnlineToken {
		return 0, ecode.AccessDenied
	}
	return s.onlineTotal, nil
}

func (s *Service) loadNewCount() {
	if s.newCountRunning {
		return
	}
	s.newCountRunning = true
	defer func() {
		s.newCountRunning = false
	}()
	arg := &dyapi.RegCountReq{Rid: s.c.Rule.MainRids}
	reply, err := s.dyGRPC.RegCount(context.Background(), arg)
	if err != nil {
		log.Error("loadNewCount s.dyGRPC.RegCount(%v) error (%v)", arg, err)
		return
	}
	if len(reply.RegCountMap) == 0 {
		log.Error("loadNewCount s.arc.RanksTopCount2(%v) res len(%d) == 0", arg, len(reply.RegCountMap))
		return
	}
	s.regionCount = reply.RegCountMap
	var allCount int64
	for rid, count := range reply.RegCountMap {
		// only first need all count
		for _, val := range s.c.Rule.Rids {
			if rid == val {
				allCount += count
				break
			}
		}
	}
}

func (s *Service) loadOnlineTotal() {
	if s.onlineTotalRunning {
		return
	}
	s.onlineTotalRunning = true
	defer func() {
		s.onlineTotalRunning = false
	}()
	total, err := s.dao.OnlineTotal(context.Background())
	if err != nil || total == nil {
		log.Error("loadOnlineTotal s.dao.OnlineTotal error(%v)", err)
		return
	}
	if total.BuvidCount > 0 {
		atomic.StoreInt64(&s.onlineTotal, total.BuvidCount)
	}
}

// loadOnlineListProc  online list proc.
func (s *Service) loadOnlineList() {
	if s.onlineListRunning {
		return
	}
	s.onlineListRunning = true
	defer func() {
		s.onlineListRunning = false
	}()
	ctx := context.Background()
	var aids []*model.OnlineAid
	cacheErr := retry(func() (err error) {
		aids, err = s.dao.OnlineListCache(ctx)
		return err
	})
	if cacheErr != nil {
		log.Error("日志告警 loadOnlineList OnlineListCache error:%v", cacheErr)
		return
	}
	if cacheCnt := len(aids); cacheCnt < s.c.WEB.OnlineCount {
		log.Error("日志告警 loadOnlineList len OnlineListCache count:%d", cacheCnt)
		return
	}
	s.onlineAids = aids
}
