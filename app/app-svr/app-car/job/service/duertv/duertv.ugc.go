package duertv

import (
	"context"
	"encoding/json"
	"time"

	"go-common/library/log"
	"go-common/library/railgun"
	"go-gateway/app/app-svr/app-car/job/model/archive"
	"go-gateway/app/app-svr/app-car/job/model/duertv"
	"go-gateway/app/app-svr/archive/service/api"

	"go-common/library/sync/errgroup.v2"

	chanGRPC "git.bilibili.co/bapis/bapis-go/community/interface/channel"
	flowCtrlGrpc "git.bilibili.co/bapis/bapis-go/content-flow-control/service"
	creativeAPI "git.bilibili.co/bapis/bapis-go/videoup/open/service"
)

const _overseaBlockKey = "oversea_block"

func (s *Service) initArchiveRailGun(databus *railgun.DatabusV1Config, singleConfig *railgun.SingleConfig, cfg *railgun.Config) {
	inputer := railgun.NewDatabusV1Inputer(databus)
	processor := railgun.NewSingleProcessor(singleConfig, s.archiveRailGunUnpack, s.archiveRailGunDo)
	g := railgun.NewRailGun("稿件小度媒资推送", cfg, inputer, processor)
	s.archiveRailGun = g
	g.Start()
}

func (s *Service) archiveRailGunUnpack(msg railgun.Message) (*railgun.SingleUnpackMsg, error) {
	message := new(archive.Message)
	if err := json.Unmarshal(msg.Payload(), &message); err != nil {
		return nil, err
	}
	if message.Table != _archiveTable {
		return nil, nil
	}
	const (
		_updateArc = "update"
		_insertArc = "insert"
	)
	log.Warn("duertv syncArchive action(%v) old(%+v) new(%+v)", message.Action, message.Old, message.New)
	switch message.Action {
	case _insertArc, _updateArc:
		nw := message.New
		if nw == nil {
			return nil, nil
		}
		if old := message.Old; old != nil {
			// 状态一样不做处理
			if old.State == nw.State {
				return nil, nil
			}
		}
		// 过滤互动视频
		if nw.AttrVal(api.AttrBitSteinsGate) == api.AttrYes {
			return nil, nil
		}
		// 过滤PGC视频
		if nw.AttrVal(api.AttrBitIsPGC) == api.AttrYes {
			return nil, nil
		}
		// 赋值action 用于追踪上报
		nw.Action = message.Action
		return &railgun.SingleUnpackMsg{
			Group: nw.Aid,
			Item:  nw,
		}, nil
	}
	return nil, nil
}

func (s *Service) archiveRailGunDo(ctx context.Context, item interface{}) railgun.MsgPolicy {
	nw := item.(*archive.ArcMsg)
	if nw == nil {
		return railgun.MsgPolicyIgnore
	}
	func() {
		var (
			arc            *api.Arc
			chans          []*chanGRPC.Channel
			overseaBlocked bool
		)
		group := errgroup.WithContext(ctx)
		group.Go(func(cctx context.Context) (err error) {
			arc, err = s.arc.Archive(cctx, nw.Aid)
			if err != nil {
				log.Error("s.arc.Archive aid(%d) error(%v)", nw.Aid, err)
				return err
			}
			return nil
		})
		group.Go(func(cctx context.Context) (err error) {
			chans, err = s.cl.ResourceChannels(cctx, nw.Aid)
			if err != nil {
				log.Error("s.cl.ResourceChannels aid(%d) error(%v)", nw.Aid, err)
				return nil
			}
			return nil
		})
		group.Go(func(ctx context.Context) error {
			flowCtrlData, flowErr := s.arc.FlowControlInfoV2(ctx, nw.Aid)
			if flowErr != nil {
				log.Error("%+v", flowErr)
				return nil
			}
			for _, v := range flowCtrlData {
				if v != nil && v.Key == _overseaBlockKey && v.Value == 1 {
					overseaBlocked = true
					break
				}
			}
			return nil
		})
		if err := group.Wait(); err != nil {
			log.Error("%+v", err)
			return
		}
		// 如果cid等于0直接忽略掉
		if arc.FirstCid == 0 {
			return
		}
		// 海外禁止
		if overseaBlocked {
			return
		}
		var pudatas []*duertv.DuertvPushUGC
		pd := &duertv.DuertvPushUGC{}
		if ok := pd.FromUGC(arc, s.c.Duertv.Partner, chans, s.oneRegions); !ok {
			return
		}
		pudatas = append(pudatas, pd)
		pd = &duertv.DuertvPushUGC{}
		if ok := pd.FromUGCCollection(arc, s.c.Duertv.Partner, chans, s.oneRegions); !ok {
			return
		}
		pudatas = append(pudatas, pd)
		if len(pudatas) == 0 {
			return
		}
		group = errgroup.WithContext(ctx)
		for _, pudata := range pudatas {
			pd := &duertv.DuertvPushUGC{}
			*pd = *pudata
			if nw.Action == "insert" { // 只输出新增稿件
				log.Warn("duertv syncArchive fin(%+v)", pd)
			}
			group.Go(func(cctx context.Context) (err error) {
				if err := retry(func() error {
					return s.dt.PushUGC(cctx, pd, time.Now())
				}); err != nil {
					log.Error("日志告警 小度媒资推送失败 duertv ugc s.dt.Push error(%v)", err)
					return err
				}
				return nil
			})
		}
		_ = group.Wait()
	}()
	return railgun.MsgPolicyNormal
}

func (s *Service) initduertvRankAllRailGun(cronInputer *railgun.CronInputerConfig, cronProcessor *railgun.CronProcessorConfig, cfg *railgun.Config) {
	// 每小时定时跑一次
	inputer := railgun.NewCronInputer(cronInputer)
	processor := railgun.NewCronProcessor(cronProcessor, func(ctx context.Context) railgun.MsgPolicy {
		s.duertvUGCRankAll()
		return railgun.MsgPolicyNormal
	})
	r := railgun.NewRailGun("小度UGC排行榜媒资数据推送", cfg, inputer, processor)
	s.duertvRankAllRailGun = r
}

func (s *Service) initduertvHostAllRailGun(cronInputer *railgun.CronInputerConfig, cronProcessor *railgun.CronProcessorConfig, cfg *railgun.Config) {
	// 每小时定时跑一次
	inputer := railgun.NewCronInputer(cronInputer)
	processor := railgun.NewCronProcessor(cronProcessor, func(ctx context.Context) railgun.MsgPolicy {
		s.duertvUGCHots()
		return railgun.MsgPolicyNormal
	})
	r := railgun.NewRailGun("小度UGC热门媒资数据推送", cfg, inputer, processor)
	s.duertvHostRailGun = r
}

func (s *Service) duertvUGCRankAll() {
	ctx := context.Background()
	list, err := s.rcmd.RankAppAll(ctx)
	if err != nil {
		log.Error("%v", err)
		return
	}
	s.duertvUGC(ctx, list)
}

func (s *Service) duertvUGC(ctx context.Context, list []int64) {
	var (
		arc       map[int64]*api.Arc
		chans     map[int64][]*chanGRPC.Channel
		flowCtrls map[int64]*flowCtrlGrpc.FlowCtlInfoV2Reply
	)
	group := errgroup.WithContext(ctx)
	group.Go(func(cctx context.Context) (err error) {
		arc, err = s.arc.Archives(ctx, list)
		if err != nil {
			log.Error("%v", err)
			return err
		}
		return nil
	})
	group.Go(func(cctx context.Context) (err error) {
		chans, err = s.cl.ResourceChannelsAll(cctx, list)
		if err != nil {
			log.Error("s.cl.ResourceChannelsAll aid(%d) error(%v)", list, err)
			return nil
		}
		return nil
	})
	group.Go(func(cctx context.Context) error {
		var flowErr error
		flowCtrls, flowErr = s.arc.FlowControlInfosV2(cctx, list)
		if flowErr != nil {
			log.Error("s.arc.FlowControlInfosV2 list(%v) error(%v)", list, flowErr)
		}
		return nil
	})
	if err := group.Wait(); err != nil {
		log.Error("%+v", err)
		return
	}
	s.ugcUpdateMsg(ctx, arc, list, chans, flowCtrls)
}

func (s *Service) duertvUGCHots() {
	ctx := context.Background()
	var (
		tmpaids    = make(map[int64]struct{})
		forbidAids = make(map[int64]struct{})
	)
	for i := 0; i < 11; i++ {
		list, err := s.rcmd.HotHeTongTabCard(ctx, i)
		if err != nil {
			log.Error("%v", err)
			return
		}
		flowResp, err := s.creativeClient.FlowJudges(context.Background(), &creativeAPI.FlowJudgesReq{
			Oids:     list,
			Business: 4,
			Gid:      24,
		})
		if err != nil {
			log.Error("s.creativeClient.FlowJudge error(%v)", err)
			for _, v := range list {
				tmpaids[v] = struct{}{}
			}
		} else {
			for _, oid := range flowResp.Oids {
				forbidAids[oid] = struct{}{}
			}
			for _, v := range list {
				if _, ok := forbidAids[v]; ok {
					log.Info("aid(%d) is flowJundged", v)
					continue
				}
				tmpaids[v] = struct{}{}
			}
		}
	}
	var aids []int64
	for v := range tmpaids {
		aids = append(aids, v)
	}
	s.duertvUGC(ctx, aids)
}

func (s *Service) ugcUpdateMsg(ctx context.Context, arcs map[int64]*api.Arc, aids []int64, chanss map[int64][]*chanGRPC.Channel, flowCtrls map[int64]*flowCtrlGrpc.FlowCtlInfoV2Reply) {
	for _, v := range aids {
		arc, ok := arcs[v]
		if !ok || arc.FirstCid == 0 {
			continue
		}
		chans, ok := chanss[v]
		if !ok {
			continue
		}
		flowCtrlData, ok := flowCtrls[v]
		if ok && flowCtrlData != nil {
			if overseaBlocked := func() bool {
				for _, item := range flowCtrlData.Items {
					if item != nil && item.Key == _overseaBlockKey && item.Value == 1 {
						return true
					}
				}
				return false
			}(); overseaBlocked {
				continue
			}
		}
		var pudatas []*duertv.DuertvPushUGC
		pd := &duertv.DuertvPushUGC{}
		if ok := pd.FromUGC(arc, s.c.Duertv.Partner, chans, s.oneRegions); !ok {
			return
		}
		pudatas = append(pudatas, pd)
		pd = &duertv.DuertvPushUGC{}
		if ok := pd.FromUGCCollection(arc, s.c.Duertv.Partner, chans, s.oneRegions); !ok {
			return
		}
		pudatas = append(pudatas, pd)
		if len(pudatas) == 0 {
			return
		}
		group := errgroup.WithContext(ctx)
		for _, pudata := range pudatas {
			pd := &duertv.DuertvPushUGC{}
			*pd = *pudata
			group.Go(func(cctx context.Context) (err error) {
				if err := retry(func() error {
					return s.dt.PushUGC(cctx, pd, time.Now())
				}); err != nil {
					log.Error("日志告警 小度媒资排行榜推送失败 duertv ugc s.dt.Push error(%v)", err)
					return err
				}
				return nil
			})
		}
		_ = group.Wait()
	}
}
