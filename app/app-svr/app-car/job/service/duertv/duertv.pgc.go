package duertv

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"go-common/library/log"
	"go-common/library/railgun"
	"go-gateway/app/app-svr/app-car/job/model/bangumi"
	"go-gateway/app/app-svr/app-car/job/model/duertv"
)

const (
	_archiveTable = "archive"
)

func (s *Service) initduertvBangumiAllRailGun(cronInputer *railgun.CronInputerConfig, cronProcessor *railgun.CronProcessorConfig, cfg *railgun.Config) {
	// 每小时定时跑一次
	inputer := railgun.NewCronInputer(cronInputer)
	processor := railgun.NewCronProcessor(cronProcessor, func(ctx context.Context) railgun.MsgPolicy {
		s.duertvBangumiDay()
		return railgun.MsgPolicyNormal
	})
	r := railgun.NewRailGun("小度PGC媒资数据推送", cfg, inputer, processor)
	s.duertvBangumiAllRailGun = r
}

func (s *Service) duertvBangumiAll() {
	ctx := context.Background()
	// 五种类型1：番剧，2：电影，3：纪录片，4：国漫，5：电视剧，7:综艺
	seasonTypes := []int{1, 2, 3, 4, 5, 7}
	for _, stype := range seasonTypes {
		pn := 1
		ps := 50
		var bgms []*bangumi.Content
		for {
			var (
				bgm []*bangumi.Content
				err error
			)
			// 重试2次
			for i := 0; i < 2; i++ {
				if bgm, err = s.bgm.ChannelContent(ctx, pn, ps, stype, "xiaodu"); err == nil {
					break
				}
				time.Sleep(time.Millisecond * 500)
			}
			if err != nil {
				log.Error("日志告警 获取PGC全量数据获取失败 s.bgm.ChannelContent seasonType(%d) error(%v)", stype, err)
				// 重试3次都不行直接跳出这个循环，去请求下一个类型
				break
			}
			if len(bgm) == 0 {
				break
			}
			bgms = append(bgms, bgm...)
			// +1
			pn++
		}
		// 上传
		s.bangumiMsg(bgms)
		log.Info("duertv bangumi all seasonType(%d) success", stype)
	}
	log.Info("duertv bangumi all success")
}

func (s *Service) duertvBangumiOffshelve() {
	ctx := context.Background()
	// 五种类型1：番剧，2：电影，3：纪录片，4：国漫，5：电视剧，7:综艺
	seasonTypes := []int{1, 2, 3, 4, 5, 7}
	for _, stype := range seasonTypes {
		pn := 1
		ps := 50
		var bgms []*bangumi.Offshelve
		for {
			var (
				bgm []*bangumi.Offshelve
				err error
			)
			// 重试2次
			for i := 0; i < 2; i++ {
				if bgm, err = s.bgm.ChannelContentoffshelve(ctx, pn, ps, stype); err == nil {
					break
				}
				time.Sleep(time.Millisecond * 500)
			}
			if err != nil {
				log.Error("日志告警 获取PGC全量下架数据获取失败 s.bgm.ChannelContentoffshelve seasonType(%d) error(%v)", stype, err)
				// 重试2次都不行直接跳出这个循环，去请求下一个类型
				break
			}
			if len(bgm) == 0 {
				break
			}
			bgms = append(bgms, bgm...)
			// +1
			pn++
		}
		s.bangumiOffshelve(bgms)
		log.Info("duertv bangumi offshelve seasonType(%d) success", stype)
	}
	log.Info("duertv bangumi offshelve success")
}

func (s *Service) duertvBangumiDay() {
	ctx := context.Background()
	now := time.Now()
	// 当前时间整点
	thisTime := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, now.Location())
	// 前一个小时
	yesTime := thisTime.Add(-time.Hour * 1)
	// 五种类型1：番剧，2：电影，3：纪录片，4：国漫，5：电视剧，7:综艺
	seasonTypes := []int{1, 2, 3, 4, 5, 7}
	for _, stype := range seasonTypes {
		pn := 1
		ps := 50
		var bgms []*bangumi.Content
		for {
			var (
				bgm []*bangumi.Content
				err error
			)
			// 重试2次
			for i := 0; i < 2; i++ {
				if bgm, err = s.bgm.ChannelContentChange(ctx, pn, ps, stype, "xiaodu", yesTime, thisTime); err == nil {
					break
				}
				time.Sleep(time.Millisecond * 500)
			}
			if err != nil {
				log.Error("日志告警 获取PGC每日数据获取失败 s.bgm.ChannelContent seasonType(%d) start_time(%s) end_time(%s) error(%v)", stype, yesTime.Format("2006-01-02 15:04:05"), thisTime.Format("2006-01-02 15:04:05"), err)
				// 重试3次都不行直接跳出这个循环，去请求下一个类型
				break
			}
			if len(bgm) == 0 {
				break
			}
			bgms = append(bgms, bgm...)
			// +1
			pn++
		}
		// 上传
		s.bangumiMsg(bgms)
		log.Info("duertv bangumi day seasonType(%d) start_time(%s) end_time(%s) success", stype, yesTime.Format("2006-01-02 15:04:05"), thisTime.Format("2006-01-02 15:04:05"))
	}
	log.Info("duertv bangumi day success")
}

func (s *Service) bangumiMsg(bgms []*bangumi.Content) {
	// 上传
	seasonm := map[int64]struct{}{}
	for _, bgm := range bgms {
		if len(bgm.Episodes) == 0 || bgm.Season == nil {
			continue
		}
		if _, ok := seasonm[bgm.Season.ID]; !ok {
			// 插入一条只有season的数据，用于合辑信息上报
			seasonm[bgm.Season.ID] = struct{}{}
			pubgm := &bangumi.Content{}
			*pubgm = *bgm
			pubgm.Episodes = []*bangumi.Episode{}
			if len(bgm.Episodes) > 0 {
				pubgm.Episodes = append(pubgm.Episodes, bgm.Episodes[0]) //第一张EP的图
			}
			pubgm.IsPushSeason = true
			s.duertvBgmChan <- pubgm
		}
		// 一次最多上传10条
		const _max = 10
		pn := 0
		var (
			eps   []*bangumi.Episode
			isend bool
		)
		for {
			start := pn * _max
			end := start + _max
			if end < len(bgm.Episodes) {
				eps = bgm.Episodes[start:end]
			} else if start < len(bgm.Episodes) {
				eps = bgm.Episodes[start:]
				isend = true
			} else {
				eps = bgm.Episodes
				isend = true
			}
			pubgm := &bangumi.Content{}
			*pubgm = *bgm
			pubgm.Episodes = eps
			pubgm.IsPushSeason = false
			s.duertvBgmChan <- pubgm
			// +1
			pn++
			// 如果OK表示已经只剩下最后的数据了，执行完跳出这个ep的循环
			if isend {
				break
			}
		}
	}
}

func (s *Service) bangumiOffshelve(bgms []*bangumi.Offshelve) {
	for _, bgm := range bgms {
		if len(bgm.Episodes) == 0 {
			pubgm := &bangumi.Offshelve{}
			*pubgm = *bgm
			s.duertvBgmOffshelveChan <- pubgm
			continue
		}
		// 一次最多上传10条
		const _max = 10
		pn := 0
		var (
			eps   []*bangumi.OffshelveEpInfo
			isend bool
		)
		for {
			start := pn * _max
			end := start + _max
			if end < len(bgm.Episodes) {
				eps = bgm.Episodes[start:end]
			} else if start < len(bgm.Episodes) {
				eps = bgm.Episodes[start:]
				isend = true
			} else {
				eps = bgm.Episodes
				isend = true
			}
			pubgm := &bangumi.Offshelve{}
			*pubgm = *bgm
			pubgm.Episodes = eps
			s.duertvBgmOffshelveChan <- pubgm
			// +1
			pn++
			// 如果OK表示已经只剩下最后的数据了，执行完跳出这个ep的循环
			if isend {
				break
			}
		}
	}
}

func (s *Service) duertvPushproc() {
	defer s.waiter.Done()
	for {
		bgm, ok := <-s.duertvBgmChan
		if !ok {
			log.Warn("duertv s.duertvPush.Cloesd")
			return
		}
		var pudatas []*duertv.DuertvPush
		for _, ep := range bgm.Episodes {
			pd := &duertv.DuertvPush{}
			if ok := pd.FromBangumi(ep, bgm, s.c.Duertv.Partner); !ok {
				continue
			}
			pudatas = append(pudatas, pd)
		}
		// 没有ep信息，表示当前是合辑数据
		if bgm.IsPushSeason && len(bgm.Episodes) <= 1 {
			pd := &duertv.DuertvPush{}
			if ok := pd.FromBangumiSeason(bgm, s.c.Duertv.Partner); !ok {
				continue
			}
			pudatas = append(pudatas, pd)
		}
		if len(pudatas) == 0 {
			continue
		}
		ctx := context.Background()
		// 重试2次
		if err := retry(func() error {
			return s.dt.Push(ctx, pudatas, time.Now())
		}); err != nil {
			log.Error("日志告警 小度媒资推送失败 duertv pgc s.dt.Push error(%v)", err)
			continue
		}
		log.Info("duertv push pgc success")
	}
}

func (s *Service) duertvPushOffshelveproc() {
	defer s.waiter.Done()
	for {
		bgm, ok := <-s.duertvBgmOffshelveChan
		if !ok {
			log.Warn("duertv s.duertvPushOffshelveproc.Cloesd")
			return
		}
		var pudatas []*duertv.DuertvPush
		for _, ep := range bgm.Episodes {
			pd := &duertv.DuertvPush{}
			if ok := pd.FromBangumiEPOffshelve(ep, bgm, s.c.Duertv.Partner); !ok {
				continue
			}
			pudatas = append(pudatas, pd)
		}
		// 没有ep信息，表示当前是合辑数据
		if len(bgm.Episodes) == 0 {
			pd := &duertv.DuertvPush{}
			if ok := pd.FromBangumiSeasonOffshelve(bgm, s.c.Duertv.Partner); !ok {
				continue
			}
			pudatas = append(pudatas, pd)
		}
		if len(pudatas) == 0 {
			continue
		}
		ctx := context.Background()
		// 重试2次
		if err := retry(func() error {
			return s.dt.Push(ctx, pudatas, time.Now())
		}); err != nil {
			log.Error("日志告警 小度媒资推送失败 duertv pgc s.dt.Push error(%v)", err)
			continue
		}
		log.Info("duertv push pgc offshelve success")
	}
}

func retry(callback func() error) error {
	var err error
	for i := 0; i < 2; i++ {
		if err = callback(); err == nil {
			return nil
		}
		time.Sleep(time.Millisecond * 100)
	}
	return err
}

func (s *Service) initPGCRailGun(databus *railgun.DatabusV2Config, singleConfig *railgun.SingleConfig, cfg *railgun.Config) {
	inputer := railgun.NewDatabusV2Inputer(databus)
	processor := railgun.NewSingleProcessor(singleConfig, s.seasonRailGunUnpack, s.seasonRailGunDo)
	g := railgun.NewRailGun("PGC小度下架媒资推送", cfg, inputer, processor)
	s.archiveRailGun = g
	g.Start()
}

// 获取PGC下架数据
func (s *Service) seasonRailGunUnpack(msg railgun.Message) (*railgun.SingleUnpackMsg, error) {
	message := new(bangumi.DatabusEntity)
	if err := json.Unmarshal(msg.Payload(), &message); err != nil {
		return nil, err
	}
	// 只获取变更信息
	if message.EntityChange == nil {
		return nil, nil
	}
	// 只处理消息变更和下架状态
	if message.EventType != "STATUS_UPDATED" && message.EntityChange.Value != "OFFLINE" {
		return nil, nil
	}
	seasonID, _ := strconv.ParseInt(message.EntityID, 10, 64)
	return &railgun.SingleUnpackMsg{
		Group: seasonID,
		Item:  message,
	}, nil
}

func (s *Service) seasonRailGunDo(ctx context.Context, item interface{}) railgun.MsgPolicy {
	nw := item.(*bangumi.DatabusEntity)
	if nw == nil {
		return railgun.MsgPolicyIgnore
	}
	func() {
		d := &duertv.DuertvPush{}
		if ok := d.FromBangumiMessage(nw, s.c.Duertv.Partner); !ok {
			return
		}
		pudatas := []*duertv.DuertvPush{d}
		// 重试2次
		if err := retry(func() error {
			return s.dt.Push(ctx, pudatas, time.Now())
		}); err != nil {
			log.Error("日志告警 小度媒资推送失败 duertv pgc s.dt.Push error(%v)", err)
			return
		}
	}()
	return railgun.MsgPolicyNormal
}
