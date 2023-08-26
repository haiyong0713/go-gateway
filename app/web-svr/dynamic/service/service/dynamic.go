package service

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	cfcgrpc "git.bilibili.co/bapis/bapis-go/content-flow-control/service"
	"go-common/library/log"
	"go-common/library/net/metadata"
	arcmdl "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/dynamic/service/conf"
	"go-gateway/app/web-svr/dynamic/service/dao"
	"go-gateway/app/web-svr/dynamic/service/model"

	"go-common/library/sync/errgroup.v2"
)

var (
	_maxAids    = 100
	_emptyArcs3 = make([]*arcmdl.Arc, 0)
)

const (
	_arcMuch   = 5
	_guoChuang = 168
)

// GoRegionTotal total  dynamic of regin
func (s *Service) GoRegionTotal(c context.Context) (res map[string]int) {
	res = map[string]int{}
	for k, v := range s.regionTotal {
		res[strconv.FormatInt(int64(k), 10)] = v
	}
	res["live"] = s.live
	return
}

// regionNeedCache
func regionNeedCache(regionArcs map[int32][]int64) bool {
	if len(regionArcs) <= 0 {
		return false
	}
	for k, v := range regionArcs {
		if len(v) < conf.Conf.Rule.MinRegionCount {
			log.Error("arcNeedCache key(%d) len(%v)<=0", k, v)
			return false
		}
	}
	return true
}

func (s *Service) DynamicRegion(ctx context.Context, business string, rid, pn, ps int64, isFilter bool) ([]*arcmdl.Arc, int64, error) {
	if business == "" {
		arcs, count, err := s.GoRegionArcs3(ctx, int32(rid), int(pn), int(ps), isFilter)
		return arcs, int64(count), err
	}
	key := fmt.Sprintf("%s_%d", business, rid)
	aids, ok := s.regionBusinessArcs[key]
	if !ok {
		return _emptyArcs3, 0, nil
	}
	count := int64(len(aids))
	start := (pn - 1) * ps
	end := start + ps + _arcMuch
	if start > count {
		return _emptyArcs3, 0, nil
	}
	if end > count {
		aids = aids[start:]
	} else {
		aids = aids[start:end]
	}
	arcs, err := s.normalArcs3(ctx, aids, int(ps))
	if err != nil {
		return nil, 0, err
	}
	return arcs, count, nil
}

func regionBusinessNeedCache(regionArcs map[string][]int64) bool {
	if len(regionArcs) <= 0 {
		return false
	}
	for k, v := range regionArcs {
		if len(v) < conf.Conf.Rule.MinRegionCount {
			log.Error("regionBusinessNeedCache key(%s) len(%v)<=0", k, v)
			return false
		}
	}
	return true
}

// GoRegionArcs3 get new arcs of region.
func (s *Service) GoRegionArcs3(c context.Context, rid int32, pn, ps int, isFilter bool) (arcs []*arcmdl.Arc, count int, err error) {
	var (
		ok         bool
		start, end int
		aids       []int64
	)
	if aids, ok = s.regionArcs[rid]; !ok {
		arcs = _emptyArcs3
		return
	}
	if isFilter {
		if filterAids, ok := s.regionFilterArcs[rid]; ok {
			aids = filterAids
		}
	}
	count = len(aids)
	start = (pn - 1) * ps
	end = start + ps + _arcMuch
	if start > count {
		arcs = _emptyArcs3
		return
	}
	if end > count {
		aids = aids[start:]
	} else {
		aids = aids[start:end]
	}
	if arcs, err = s.normalArcs3(c, aids, ps); err != nil {
		log.Error("archives(%v) error(%v)", aids, err)
	}
	return
}

// normalArcs3 .
func (s *Service) normalArcs3(c context.Context, aids []int64, ps int) (res []*arcmdl.Arc, err error) {
	var (
		arcs   []*arcmdl.Arc
		tmpRes map[int64]*arcmdl.Arc
	)
	archivesLog("normalArcs3", aids)
	if tmpRes, err = s.archives3(c, aids); err != nil {
		return
	}
	for _, aid := range aids {
		if arc, ok := tmpRes[aid]; ok {
			arcs = append(arcs, arc)
		} else {
			log.Error("normalArcs s.archives aid(%d) nil", aid)
		}
	}
	res = s.filterArc3(c, arcs, ps)
	return
}

// archives3 .
func (s *Service) archives3(c context.Context, aids []int64) (res map[int64]*arcmdl.Arc, err error) {
	if res, err = s.circleReqArcs(c, aids); err != nil {
		log.Error("[archives3] s.circleReqArcs() aids(%v) error(%v)", aids, err)
	}
	return
}

func (s *Service) filterArc3(c context.Context, arcs []*arcmdl.Arc, count int) (res []*arcmdl.Arc) {
	tmpPs := 1
	var aids []int64
	for _, arc := range arcs {
		if arc != nil {
			aids = append(aids, arc.Aid)
		}
	}
	infos, err := s.batchCfcInfos(c, aids)
	if err != nil {
		log.Error("日志告警 filterArc3 s.arc aids:%d error:%v", aids, err)
	}
	for _, arc := range arcs {
		var info *cfcgrpc.FlowCtlInfoV2Reply
		if arc != nil {
			info = infos[arc.Aid]
		}
		if tmpPs <= count && s.isShow3(arc, info) {
			res = append(res, arc)
			tmpPs = tmpPs + 1
		} else if tmpPs > count {
			break
		}
	}
	return
}

// isShow3
func (s *Service) isShow3(a *arcmdl.Arc, info *cfcgrpc.FlowCtlInfoV2Reply) bool {
	forbidden := model.ItemToArcForbidden(info)
	return a.IsNormal() && !forbidden.NoDynamic
}

// GoRegionTagArcs3 get new arcs of region and hot tag.
func (s *Service) GoRegionTagArcs3(c context.Context, rid int32, tagID int64, pn, ps int) (arcs []*arcmdl.Arc, count int, err error) {
	var (
		ok         bool
		start, end int
		aids       []int64
	)
	key := regionTagKey(rid, tagID)
	if aids, ok = s.regionTagArcs[key]; !ok {
		arcs = _emptyArcs3
		return
	}
	count = len(aids)
	start = (pn - 1) * ps
	end = start + ps + _arcMuch
	if start > count {
		arcs = _emptyArcs3
		return
	}
	if end > count {
		aids = aids[start:]
	} else {
		aids = aids[start:end]
	}
	if arcs, err = s.normalArcs3(c, aids, ps); err != nil {
		log.Error("archives(%v) error(%v)", aids, err)
	}
	return
}

// GoRegionsArcs3 get batch new arcs of regions.
func (s *Service) GoRegionsArcs3(c context.Context, rids []int32, count int) (mArcs map[int32][]*arcmdl.Arc, err error) {
	var (
		ok      bool
		noRids  []int32
		allAids []int64
		aids    []int64
		mAids   map[int32][]int64
		res     map[int64]*arcmdl.Arc
		ip      = metadata.String(c, metadata.RemoteIP)
	)
	mAids = make(map[int32][]int64, len(rids))
	for _, rid := range rids {
		end := count + _arcMuch
		if aids, ok = s.regionArcs[rid]; !ok || len(aids) == 0 {
			continue
		}
		if end > len(aids) {
			end = len(aids)
		}
		allAids = append(allAids, aids[:end]...)
		mAids[rid] = aids[:end]
	}
	archivesLog("RegionsArcs3", allAids)
	if res, err = s.archives3(c, allAids); err != nil {
		log.Error("archives(%v) error(%v)", allAids, err)
		return
	}
	mArcs = make(map[int32][]*arcmdl.Arc, len(rids))
	for _, rid := range rids {
		var arcs []*arcmdl.Arc
		for _, aid := range mAids[rid] {
			if arc, ok := res[aid]; ok {
				arcs = append(arcs, arc)
			} else {
				log.Error("RegionsArcs s.archives aid(%d) nil", aid)
			}
		}
		mArcs[rid] = s.filterArc3(c, arcs, count)
		if len(mArcs[rid]) < count {
			dao.PromError("一级分区数据错误", "RegionsArcs rid(%d) len(mArcs[rid])(%d) count(%d)", rid, len(mArcs[rid]), count)
			noRids = append(noRids, rid)
		}
	}
	//last back up from rankIndexArc.
	if len(noRids) > 0 {
		err = s.rankIndexArc3(c, noRids, 1, count, ip, mArcs)
	}
	return
}

// rankIndexArc3 archives3.
func (s *Service) rankIndexArc3(c context.Context, rids []int32, pn, ps int, ip string, mArcs map[int32][]*arcmdl.Arc) (err error) {
	var mutex = sync.Mutex{}
	group := errgroup.WithCancel(c)
	for _, rid := range rids {
		trid := rid
		group.Go(func(ctx context.Context) error {
			var (
				topRes, tmpRes []*arcmdl.Arc
			)
			if trid == _guoChuang {
				if topRes, _, err = s.RegAllArc(ctx, int64(trid), ps+_arcMuch, pn); err != nil {
					log.Error("[rankIndexArc3] s.RegAllArc() rid(%d) pn(%d) ps(%d) ip(%s) error(%v)", trid, pn, ps, ip, err)
					return nil
				}
			} else {
				if topRes, err = s.RecentThrdRegArc(c, trid, pn, ps+_arcMuch); err != nil {
					log.Error("[rankIndexArc3] s.RecentThrdRegArc() rid(%d) pn(%d) ps(%d) ip(%s) error(%v)", trid, pn, ps, ip, err)
					return nil
				}
			}
			tmpRes = s.filterArc3(c, topRes, ps)
			mutex.Lock()
			if len(tmpRes) == 0 {
				mArcs[trid] = _emptyArcs3
			} else {
				mArcs[trid] = tmpRes
			}
			mutex.Unlock()
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		log.Error("%+v", err)
	}
	return
}

// circleReqArcs .
func (s *Service) circleReqArcs(ctx context.Context, aids []int64) (aidMap map[int64]*arcmdl.Arc, err error) {
	var (
		aidsLen = len(aids)
		mutex   = sync.Mutex{}
	)
	aidMap = make(map[int64]*arcmdl.Arc, aidsLen)
	gp := errgroup.WithContext(ctx)
	for i := 0; i < aidsLen; i += _maxAids {
		var partAids []int64
		if i+_maxAids > aidsLen {
			partAids = aids[i:]
		} else {
			partAids = aids[i : i+_maxAids]
		}
		gp.Go(func(ctx context.Context) (err error) {
			var tmpRes *arcmdl.ArcsReply
			arg := &arcmdl.ArcsRequest{Aids: partAids}
			if tmpRes, err = s.arcClient.Arcs(ctx, arg); err != nil {
				return
			}
			if len(tmpRes.Arcs) > 0 {
				mutex.Lock()
				for aid, arc := range tmpRes.Arcs {
					aidMap[aid] = arc
				}
				mutex.Unlock()
			}
			return err
		})
	}
	err = gp.Wait()
	return
}
