package service

import (
	"context"
	"sort"
	"sync"
	"time"

	xecode "go-common/library/ecode"
	"go-common/library/log"
	xtime "go-common/library/time"
	"go-gateway/app/web-svr/space/ecode"
	"go-gateway/app/web-svr/space/interface/conf"
	"go-gateway/app/web-svr/space/interface/model"

	arcmdl "git.bilibili.co/bapis/bapis-go/archive/service"

	seriesgrpc "git.bilibili.co/bapis/bapis-go/platform/interface/series"

	"go-common/library/sync/errgroup.v2"
)

var (
	_emptyChArc        = make([]*model.BvArc, 0)
	_emptyChList       = make([]*model.Channel, 0)
	_emptyChDetailList = make([]*model.ChannelDetail, 0)
	_msgTypeTopArc     = 1
	_msgTypeMp         = 2
	_msgTypeChName     = 3
	_msgTypeChIntro    = 4
)

func (s *Service) ChannelList(ctx context.Context, mid int64, isGuest bool) ([]*model.Channel, error) {
	var (
		channel  *model.Channel
		channels []*model.Channel
	)
	g := errgroup.WithCancel(ctx)
	g.Go(func(ctx context.Context) error {
		reply, err := s.seriesGRPC.ListSeries(ctx, &seriesgrpc.ListSeriesReq{Mid: mid, State: seriesgrpc.SeriesOnline})
		if err != nil {
			log.Error("%+v", err)
			return nil
		}
		var aids []int64
		for _, val := range reply.GetSeriesList() {
			meta := val.GetMeta()
			if meta.GetCategory() != seriesgrpc.SeriesLiveReplay {
				continue
			}
			channel = &model.Channel{
				Cid:            meta.SeriesId,
				Mid:            meta.Mid,
				Name:           meta.Name,
				Intro:          meta.Description,
				Mtime:          meta.MTime,
				Count:          int(meta.Total),
				IsLivePlayback: true,
			}
			aids = val.GetRecentAids()
			break
		}
		if channel == nil {
			return nil
		}
		if len(aids) == 0 {
			return nil
		}
		aids = aids[:1]
		arcs, err := s.archives(ctx, aids)
		if err != nil {
			log.Error("%+v", err)
			return nil
		}
		for _, aid := range aids {
			arc, ok := arcs[aid]
			if !ok {
				continue
			}
			channel.Cover = arc.Pic
		}
		return nil
	})
	g.Go(func(ctx context.Context) error {
		var err error
		channels, err = s.channelList(ctx, mid)
		if err != nil {
			return err
		}
		if channels == nil {
			return nil
		}
		sort.Slice(channels, func(i, j int) bool {
			return channels[i].Mtime > channels[j].Mtime
		})
		return nil
	})
	if err := g.Wait(); err != nil {
		return nil, err
	}
	if channel == nil {
		return channels, nil
	}
	return append([]*model.Channel{channel}, channels...), nil
}

// ChannelList get channel list.
func (s *Service) channelList(c context.Context, mid int64) (channels []*model.Channel, err error) {
	var (
		channelExtra map[int64]*model.ChannelExtra
		cids         []int64
		addCache     = true
	)
	if channels, err = s.dao.ChannelListCache(c, mid); err != nil {
		addCache = false
	} else if len(channels) > 0 {
		for _, channel := range channels {
			cids = append(cids, channel.Cid)
		}
		if channelExtra, err = s.channelExtra(c, mid, cids); err != nil {
			err = nil
			return
		}
		for _, channel := range channels {
			if _, ok := channelExtra[channel.Cid]; ok {
				channel.Count = channelExtra[channel.Cid].Count
				channel.Cover = channelExtra[channel.Cid].Cover
			}
		}
		return
	}
	if channels, err = s.dao.ChannelList(c, mid); err != nil {
		log.Error("s.dao.ChannelList(%d) error(%v)", mid, err)
		return
	}
	if len(channels) == 0 {
		channels = _emptyChList
		return
	}
	for _, channel := range channels {
		cids = append(cids, channel.Cid)
	}
	if channelExtra, err = s.channelExtra(c, mid, cids); err != nil {
		err = nil
		return
	}
	for _, channel := range channels {
		if _, ok := channelExtra[channel.Cid]; ok {
			channel.Count = channelExtra[channel.Cid].Count
			channel.Cover = channelExtra[channel.Cid].Cover
		}
	}
	if addCache {
		s.cache.Do(c, func(c context.Context) {
			_ = s.dao.SetChannelListCache(c, mid, channels)
		})
	}
	return
}

// Channel get channel info.
func (s *Service) Channel(c context.Context, mid, cid int64) (channel *model.Channel, err error) {
	var (
		extra    *model.ChannelExtra
		arcReply *arcmdl.ArcReply
		addCache bool
	)
	if channel, addCache, err = s.channel(c, mid, cid); err != nil {
		log.Error("s.channel(%d,%d) error(%v)", mid, cid, err)
		return
	}
	if extra, err = s.dao.ChannelExtra(c, mid, cid); err != nil {
		log.Error("s.dao.ChannelExtra(%d,%d) error(%v)", mid, cid, err)
		err = nil
	} else if extra != nil {
		channel.Count = extra.Count
		if extra.Aid > 0 {
			if arcReply, err = s.arcClient.Arc(c, &arcmdl.ArcRequest{Aid: extra.Aid}); err != nil {
				log.Error("s.arcClient.Arc(%d) error(%v)", extra.Aid, err)
				err = nil
			}
			arc := arcReply.GetArc()
			if arc != nil && arc.IsNormal() {
				channel.Cover = arc.Pic
			}
		}
	}
	if addCache {
		s.cache.Do(c, func(c context.Context) {
			_ = s.dao.SetChannelCache(c, mid, cid, channel)
		})
	}
	return
}

func (s *Service) ChannelDetail(ctx context.Context, channelID int64) (*model.Channel, error) {
	mid, cid, err := s.channelMid(ctx, channelID)
	if err != nil {
		return nil, err
	}
	if mid == 0 || cid == 0 {
		return nil, xecode.NothingFound
	}
	return s.Channel(ctx, mid, cid)
}

func (s *Service) channel(c context.Context, mid, cid int64) (res *model.Channel, addCache bool, err error) {
	addCache = true
	if res, err = s.dao.ChannelCache(c, mid, cid); err != nil {
		addCache = false
	} else if res != nil {
		return
	}
	if res, err = s.dao.Channel(c, mid, cid); err != nil {
		log.Error("s.dao.Channel(%d,%d) error(%v)", mid, cid, err)
	} else if res == nil {
		err = xecode.NothingFound
	}
	return
}

// ChannelIndex get channel index info.
func (s *Service) ChannelIndex(ctx context.Context, mid int64, isGuest bool) ([]*model.ChannelDetail, error) {
	var (
		channel  *model.ChannelDetail
		channels []*model.Channel
	)
	group := errgroup.WithContext(ctx)
	group.Go(func(ctx context.Context) error {
		var err error
		if channels, err = s.ChannelList(ctx, mid, isGuest); err != nil {
			log.Error("%+v", err)
		}
		return nil
	})
	group.Go(func(ctx context.Context) error {
		var err error
		if channel, err = s.channelLivePlayback(ctx, mid, 0, 1, int64(conf.Conf.Rule.ChIndexCnt)); err != nil {
			log.Error("%+v", err)
		}
		return nil
	})
	if err := group.Wait(); err != nil {
		log.Error("%+v", err)
	}
	var res []*model.ChannelDetail
	if len(channels) != 0 {
		group := errgroup.WithContext(ctx)
		mutex := sync.Mutex{}
		for _, channel := range channels {
			cid := channel.Cid
			group.Go(func(ctx context.Context) error {
				detail, err := s.channelVideo(ctx, mid, cid, 1, conf.Conf.Rule.ChIndexCnt, isGuest, false)
				if err != nil {
					log.Error("s.ChannelVideos(%d,%d) error(%v)", mid, cid, err)
					return nil
				}
				if detail == nil || detail.Channel == nil {
					return nil
				}
				mutex.Lock()
				res = append(res, detail)
				mutex.Unlock()
				return nil
			})
		}
		if err := group.Wait(); err != nil {
			log.Error("%+v", err)
		}
	}
	if channel == nil && len(res) == 0 {
		return _emptyChDetailList, nil
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].Mtime > res[j].Mtime
	})
	if channel != nil {
		res = append([]*model.ChannelDetail{channel}, res...)
	}
	for _, val := range res {
		// 稿件数量大于1时才出按钮
		if len(val.Archives) > 1 && !s.forbidEpisodicButton(mid) {
			val.EpisodicButton = &model.ArcListButton{
				Text: s.c.PlayButton.Text,
			}
		}
	}
	return res, nil
}

func (s *Service) seriesUpgraded(ctx context.Context, mid int64) error {
	// 全站用户都已经升级完成, 频道数据已经迁移, 默认返回频道升级中不可修改
	return ecode.ChForbitModify
}

// AddChannel add channel.
func (s *Service) AddChannel(c context.Context, mid int64, name, intro string) (cid int64, err error) {
	var ts = time.Now()
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) error {
		_, e := s.realName(ctx, mid)
		if e != nil {
			return e
		}
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		e := s.channelCheck(ctx, mid, 0, name, true, true)
		if e != nil {
			log.Error("AddChannel channelCheck(%d,%s) error(%v)", mid, name, err)
			return e
		}
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		e := s.Filter(ctx, []string{name, intro})
		if e != nil {
			return e
		}
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		return s.seriesUpgraded(ctx, mid)
	})
	if err = eg.Wait(); err != nil {
		log.Error("%+v", err)
		return
	}
	if cid, err = s.dao.AddChannel(c, mid, name, intro, ts); err != nil {
		log.Error("s.dao.AddChannel(%d,%s,%s) error(%v)", mid, name, intro, err)
		return
	} else if cid > 0 {
		s.cache.Do(c, func(c context.Context) {
			ch := &model.Channel{Cid: cid, Mid: mid, Name: name, Intro: intro, Mtime: xtime.Time(ts.Unix())}
			_ = s.dao.SetChannelCache(c, mid, cid, ch)
		})
	}
	return
}

// EditChannel edit channel.
func (s *Service) EditChannel(c context.Context, mid, cid int64, name, intro string) (err error) {
	var (
		affected int64
		ts       = time.Now()
	)
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) error {
		_, e := s.realName(ctx, mid)
		if e != nil {
			return e
		}
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		e := s.channelCheck(ctx, mid, cid, name, true, false)
		if e != nil {
			log.Error("EditChannel.channelCheck(%d,%d,%s) error(%v)", mid, cid, name, err)
			return e
		}
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		e := s.Filter(ctx, []string{name, intro})
		if e != nil {
			return e
		}
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		return s.seriesUpgraded(ctx, mid)
	})
	if err = eg.Wait(); err != nil {
		log.Error("%+v", err)
		return
	}
	if affected, err = s.dao.EditChannel(c, mid, cid, name, intro, ts); err != nil {
		log.Error("s.dao.EditChannel(%d,%s,%s) error(%v)", mid, name, intro, err)
		return
	} else if affected > 0 {
		s.cache.Do(c, func(c context.Context) {
			ch := &model.Channel{Cid: cid, Mid: mid, Name: name, Intro: intro, Mtime: xtime.Time(ts.Unix())}
			_ = s.dao.SetChannelCache(c, mid, cid, ch)
		})
	}
	return
}

// DelChannel del channel.
func (s *Service) DelChannel(c context.Context, mid, cid int64) (err error) {
	if err := s.seriesUpgraded(c, mid); err != nil {
		log.Error("%+v", err)
		return err
	}
	var affected int64
	if affected, err = s.dao.DelChannel(c, mid, cid); err != nil {
		log.Error("s.dao.DelChannel(%d,%d) error(%v)", mid, cid, err)
		return
	} else if affected > 0 {
		if err := s.dao.DelChannelCache(c, mid, cid); err != nil {
			log.Error("%+v", err)
		}
		if err := s.dao.DelChannelArcsCache(c, mid, cid); err != nil {
			log.Error("%+v", err)
		}
	}
	return
}

func (s *Service) channelExtra(c context.Context, mid int64, cids []int64) (extra map[int64]*model.ChannelExtra, err error) {
	if len(cids) == 0 {
		return
	}
	var (
		arcsReply *arcmdl.ArcsReply
		aids      = make([]int64, 0, len(cids))
	)
	extra = make(map[int64]*model.ChannelExtra, len(cids))
	for _, cid := range cids {
		var data *model.ChannelExtra
		if data, err = s.dao.ChannelExtra(c, mid, cid); err != nil {
			log.Error("s.dao.ChannelExtra(%d,%d) error(%v)", mid, cid, err)
			continue
		} else if data != nil {
			extra[cid] = &model.ChannelExtra{Aid: data.Aid, Cid: data.Cid, Count: data.Count}
			if data.Aid > 0 {
				aids = append(aids, data.Aid)
			}
		}
	}
	if len(aids) == 0 {
		return extra, nil
	}
	if arcsReply, err = s.arcClient.Arcs(c, &arcmdl.ArcsRequest{Aids: aids}); err != nil {
		log.Error("s.arcClient.Arcs(%v) error (%v)", aids, err)
		return
	}
	for _, cid := range cids {
		if _, ok := extra[cid]; ok {
			if arc, ok := arcsReply.Arcs[extra[cid].Aid]; ok && arc != nil && arc.IsNormal() {
				extra[cid].Cover = arc.Pic
			}
		}
	}
	return
}

func (s *Service) channelCheck(c context.Context, mid, cid int64, name string, nameCheck, countCheck bool) (err error) {
	var (
		channels []*model.Channel
		dbCheck  = false
	)
	if channels, err = s.dao.ChannelListCache(c, mid); err != nil {
		err = nil
		dbCheck = true
	} else if len(channels) == 0 {
		dbCheck = true
	}
	if dbCheck {
		if channels, err = s.dao.ChannelList(c, mid); err != nil {
			log.Error("s.dao.ChannelList(%d) error(%v)", mid, err)
			return
		}
	}
	if cnt := len(channels); cnt > 0 {
		if countCheck && cnt > conf.Conf.Rule.MaxChLimit {
			err = ecode.ChMaxCount
			return
		}
		if nameCheck {
			for _, channel := range channels {
				if name == channel.Name && cid != channel.Cid {
					err = ecode.ChNameExist
					return
				}
			}
		}
	}
	return
}

func (s *Service) ClearMsgCache(c context.Context, tye int, mid int64) (err error) {
	switch tye {
	case _msgTypeTopArc:
		err = s.dao.DelCacheTopArc(c, mid)
	case _msgTypeMp:
		err = s.dao.DelCacheMasterpiece(c, mid)
	case _msgTypeChName, _msgTypeChIntro:
		err = s.dao.DelChannelsCache(c, mid)
	default:
		err = xecode.RequestErr
	}
	return
}
