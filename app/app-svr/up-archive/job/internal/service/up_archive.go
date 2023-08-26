package service

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"go-common/library/log"
	"go-common/library/railgun"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/app-svr/up-archive/job/internal/model"
	"go-gateway/app/app-svr/up-archive/service/api"

	cfcgrpc "git.bilibili.co/bapis/bapis-go/content-flow-control/service"

	"github.com/pkg/errors"
)

const (
	_aidBulkSize = 30
	_addBulkSize = 200
	_aidGoMax    = 10
)

func (s *Service) initUpArcRailGun(cfg *railgun.DatabusV1Config, pcfg *railgun.SingleConfig) {
	inputer := railgun.NewDatabusV1Inputer(cfg)
	processor := railgun.NewSingleProcessor(pcfg, s.upArcRailGunUnpack, s.upArcRailGunDo)
	g := railgun.NewRailGun("初始化投稿列表", nil, inputer, processor)
	s.upArcRailGun = g
	g.Start()
}

func (s *Service) upArcRailGunUnpack(msg railgun.Message) (*railgun.SingleUnpackMsg, error) {
	log.Warn("接收初始化投稿列表消息成功,data:%s", msg.Payload())
	upArcMsg := &model.UpArcSub{}
	if err := json.Unmarshal(msg.Payload(), upArcMsg); err != nil {
		return nil, err
	}
	return &railgun.SingleUnpackMsg{
		Group: upArcMsg.Mid,
		Item:  upArcMsg,
	}, nil
}

func (s *Service) upArcRailGunDo(ctx context.Context, item interface{}) railgun.MsgPolicy {
	upArcMsg := item.(*model.UpArcSub)
	if upArcMsg == nil || upArcMsg.Mid <= 0 {
		return railgun.MsgPolicyIgnore
	}
	lock, err := s.dao.BuildArcPassedLock(ctx, upArcMsg.Mid)
	if err != nil {
		log.Error("日志告警 upArcRailGunDo BuildArcPassedLock(%d) error(%v)", upArcMsg.Mid, err)
		return railgun.MsgPolicyAttempts
	}
	if !lock {
		return railgun.MsgPolicyIgnore
	}
	s.buildArcPassed(upArcMsg.Mid)
	return railgun.MsgPolicyNormal
}

// nolint:gocognit
func (s *Service) buildArcPassed(mid int64) {
	ctx := context.Background()
	var (
		arcs  []*model.UpArc
		fItem map[int64][]*cfcgrpc.ForbiddenItem
	)
	if err := retry(func() (err error) {
		arcs, fItem, err = s.arcPassed(ctx, mid)
		return err
	}); err != nil {
		log.Error("日志告警 buildArcPassed RawArcPassed mid:%d error:%+v", mid, err)
		return
	}
	var passedArcs, withoutStaffArcs, withoutNoSpace, storyArcs []*model.UpArc
	for _, v := range arcs {
		if v != nil && v.IsAllowed(fItem[v.Aid]) {
			v.RandScoreNumber()
			passedArcs = append(passedArcs, v)
			withoutStaffArcs = append(withoutStaffArcs, v)
			if !v.IsUpNoSpace(fItem[v.Aid]) {
				withoutNoSpace = append(withoutNoSpace, v)
			}
			if v.IsStory() {
				storyArcs = append(storyArcs, v)
			}
		}
	}
	var staffAids []int64
	if err := retry(func() (err error) {
		staffAids, err = s.dao.RawStaffAids(ctx, mid) //mid为副投稿人的aid，这些要忽略空间防刷
		return err
	}); err != nil {
		log.Error("日志告警 buildArcPassed RawStaffAids mid:%d error:%+v", mid, err)
		return
	}
	if len(staffAids) > 0 {
		staffArcMap, fItem, err := s.arcs(ctx, staffAids, true)
		if err != nil {
			log.Error("日志告警 buildArcPassed arcs mid:%d error:%+v", mid, err)
			return
		}
		for _, v := range staffAids {
			if arc, ok := staffArcMap[v]; ok && arc != nil && arc.IsAllowed(fItem[v]) {
				arc.RandScoreNumber()
				passedArcs = append(passedArcs, arc)
				withoutNoSpace = append(withoutNoSpace, arc)
				if arc.IsStory() {
					storyArcs = append(storyArcs, arc)
				}
			}
		}
	}
	// 初始化缓存先删
	if err := retry(func() error {
		return s.dao.DelCacheAllArcPassed(ctx, mid)
	}); err != nil {
		log.Error("日志告警 buildArcPassed DelCacheAllArcPassed mid:%d error:%+v", mid, err)
		return
	}
	s.addCacheArc(passedArcs, func(arcs []*model.UpArc) error {
		if err := retry(func() error {
			return s.dao.AddCacheArcPassed(ctx, mid, arcs, api.Without_none)
		}); err != nil {
			log.Error("日志告警 buildArcPassed AddCacheArcPassed mid:%d without:%+d error:%v", mid, api.Without_none, err)
			return err
		}
		return nil
	})
	s.addCacheArc(withoutStaffArcs, func(arcs []*model.UpArc) error {
		if err := retry(func() error {
			return s.dao.AddCacheArcPassed(ctx, mid, arcs, api.Without_staff)
		}); err != nil {
			log.Error("日志告警 appendCacheArcsPassed AddCacheArcPassed mid:%d without:%+d error:%v", mid, api.Without_staff, err)
			return err
		}
		return nil
	})
	s.addCacheArc(withoutNoSpace, func(arcs []*model.UpArc) error {
		if err := retry(func() error {
			return s.dao.AddCacheArcPassed(ctx, mid, arcs, api.Without_no_space)
		}); err != nil {
			log.Error("日志告警 appendCacheArcsPassed AddCacheArcPassed mid:%d without:%+d error:%v", mid, api.Without_no_space, err)
			return err
		}
		return nil
	})
	s.addCacheArc(storyArcs, func(arcs []*model.UpArc) error {
		if err := retry(func() error {
			return s.dao.AddCacheArcStoryPassed(ctx, mid, arcs)
		}); err != nil {
			log.Error("日志告警 appendCacheArcsPassed AddCacheArcStoryPassed mid:%d without:%+d error:%v", mid, api.Without_no_space, err)
			return err
		}
		return nil
	})
	log.Warn("buildArcPassed success mid:%d,length:none:%d,without_staff:%d,without_no_space:%d,story:%d", mid, len(passedArcs), len(withoutStaffArcs), len(withoutNoSpace), len(storyArcs))
}

func (s *Service) addCacheArc(arcs []*model.UpArc, do func(arcs []*model.UpArc) error) {
	arcsLen := len(arcs)
	if arcsLen == 0 {
		arcs = []*model.UpArc{{Aid: -1, Score: 0}}
		arcsLen = 1
	}
	arcsSplitCount := arcsLen/_addBulkSize + 1
	for i := 0; i < arcsLen; i += _addBulkSize {
		var partArcs []*model.UpArc
		if i+_addBulkSize > arcsLen {
			partArcs = arcs[i:]
		} else {
			partArcs = arcs[i : i+_addBulkSize]
		}
		if err := retry(func() error {
			return do(partArcs)
		}); err != nil {
			log.Error("日志告警 addCacheArc error:%+v", err)
		}
		if arcsSplitCount > 1 {
			time.Sleep(1 * time.Millisecond)
		}
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

func (s *Service) arcPassed(ctx context.Context, mid int64) ([]*model.UpArc, map[int64][]*cfcgrpc.ForbiddenItem, error) {
	var arcs []*model.UpArc
	if err := retry(func() (err error) {
		arcs, err = s.dao.RawArcPassed(ctx, mid)
		return err
	}); err != nil {
		return nil, nil, errors.Wrapf(err, "RawArcPassed mid:%v", mid)
	}
	aids := make([]int64, 0, len(arcs))
	for _, arc := range arcs {
		aids = append(aids, arc.Aid)
	}
	var itemMutex sync.Mutex
	group := errgroup.WithContext(ctx)
	group.GOMAXPROCS(_aidGoMax)
	aidsLen := len(aids)
	fItem := make(map[int64][]*cfcgrpc.ForbiddenItem, aidsLen)
	for i := 0; i < aidsLen; i += _aidBulkSize {
		var partAids []int64
		if i+_aidBulkSize > aidsLen {
			partAids = aids[i:]
		} else {
			partAids = aids[i : i+_aidBulkSize]
		}
		group.Go(func(ctx context.Context) error {
			var reply map[int64][]*cfcgrpc.ForbiddenItem
			if err := retry(func() (err error) {
				reply, err = s.dao.ContentFlowControlInfos(ctx, partAids)
				return err
			}); err != nil {
				return errors.Wrapf(err, "ContentFlowControlInfos partAids:%v", partAids)
			}
			itemMutex.Lock()
			for aid, v := range reply {
				fItem[aid] = v
			}
			itemMutex.Unlock()
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		return nil, nil, err
	}
	return arcs, fItem, nil
}

func (s *Service) arcs(ctx context.Context, aids []int64, noCF bool) (map[int64]*model.UpArc, map[int64][]*cfcgrpc.ForbiddenItem, error) {
	var mutex, itemMutex sync.Mutex
	group := errgroup.WithContext(ctx)
	group.GOMAXPROCS(_aidGoMax)
	aidsLen := len(aids)
	res := make(map[int64]*model.UpArc, aidsLen)
	fItem := make(map[int64][]*cfcgrpc.ForbiddenItem, aidsLen)
	for i := 0; i < aidsLen; i += _aidBulkSize {
		var partAids []int64
		if i+_aidBulkSize > aidsLen {
			partAids = aids[i:]
		} else {
			partAids = aids[i : i+_aidBulkSize]
		}
		group.Go(func(ctx context.Context) error {
			var arcs []*model.UpArc
			if err := retry(func() (err error) {
				arcs, err = s.dao.RawArcs(ctx, partAids)
				return err
			}); err != nil {
				return errors.Wrapf(err, "RawArcs partAids:%v", partAids)
			}
			mutex.Lock()
			for _, v := range arcs {
				res[v.Aid] = v
			}
			mutex.Unlock()
			return nil
		})
		group.Go(func(ctx context.Context) error {
			if noCF {
				return nil
			}
			var reply map[int64][]*cfcgrpc.ForbiddenItem
			if err := retry(func() (err error) {
				reply, err = s.dao.ContentFlowControlInfos(ctx, partAids)
				return err
			}); err != nil {
				return errors.Wrapf(err, "ContentFlowControlInfos partAids:%v", partAids)
			}
			itemMutex.Lock()
			for aid, v := range reply {
				fItem[aid] = v
			}
			itemMutex.Unlock()
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		return nil, nil, err
	}
	return res, fItem, nil
}
