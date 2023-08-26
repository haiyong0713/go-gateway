package service

import (
	"context"
	"sync"
	"time"

	xecode "go-common/library/ecode"
	"go-common/library/log"
	xtime "go-common/library/time"
	"go-gateway/app/web-svr/space/ecode"
	"go-gateway/app/web-svr/space/interface/conf"
	"go-gateway/app/web-svr/space/interface/model"

	arcapi "git.bilibili.co/bapis/bapis-go/archive/service"
	arcmdl "git.bilibili.co/bapis/bapis-go/archive/service"
	seriesgrpc "git.bilibili.co/bapis/bapis-go/platform/interface/series"

	"go-common/library/sync/errgroup.v2"
)

const _aidBulkSize = 50

// AddChannelArc add channel archive.
func (s *Service) AddChannelArc(ctx context.Context, mid, cid int64, aids []int64) (fakeAids []int64, err error) {
	var (
		lastID          int64
		orderNum        int
		chAids, addAids []int64
		arcs            map[int64]*arcmdl.Arc
		videos          []*model.ChannelArc
		videoMap        map[int64]int64
		remainVideos    []*model.ChannelArcSort
		ts              = time.Now()
	)
	fakeAids = make([]int64, 0)
	eg := errgroup.WithContext(ctx)
	eg.Go(func(ctx context.Context) error {
		if _, _, err := s.channel(ctx, mid, cid); err != nil {
			log.Error("s.dao.Channel(%d,%d) error(%v)", mid, cid, err)
			return err
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
	if videos, err = s.dao.ChannelVideos(ctx, mid, cid, false); err != nil {
		log.Error("s.dao.channelVideos(%d,%d) error(%v)", mid, cid, err)
		return
	} else if orderNum = len(videos); orderNum > 0 {
		if len(aids)+orderNum > conf.Conf.Rule.MaxChArcLimit {
			err = ecode.ChMaxArcCount
			return
		}
		videoMap = make(map[int64]int64)
		for _, video := range videos {
			chAids = append(chAids, video.Aid)
			videoMap[video.Aid] = video.Aid
		}
	}
	for _, aid := range aids {
		if _, ok := videoMap[aid]; ok {
			fakeAids = append(fakeAids, aid)
		} else {
			addAids = append(addAids, aid)
		}
	}
	if len(addAids) == 0 {
		err = ecode.ChAidsExist
		return
	}
	if err = s.arcsCheck(ctx, mid, chAids); err != nil {
		return
	}
	if arcs, err = s.archives(ctx, addAids); err != nil {
		log.Error("s.arc.Archive3(%v) error(%v)", addAids, err)
		return
	}
	for _, arc := range arcs {
		if arc.AttrVal(arcapi.AttrBitIsPUGVPay) == arcapi.AttrYes {
			err = ecode.SpacePayUGV
			return
		}
	}
	aidsLen := len(addAids)
	for _, aid := range addAids {
		arc, ok := arcs[aid]
		if !ok || !arc.IsNormal() {
			fakeAids = append(fakeAids, aid)
			continue
		}
		if arc.Author.Mid != mid {
			if aidsLen == 1 {
				return nil, ecode.ChArcStaff
			}
			fakeAids = append(fakeAids, aid)
			continue
		}
		orderNum++
		remainVideos = append(remainVideos, &model.ChannelArcSort{Aid: aid, OrderNum: orderNum})
	}
	if len(remainVideos) == 0 {
		err = ecode.ChAidsExist
		return
	}
	if lastID, err = s.dao.AddChannelArc(ctx, mid, cid, ts, remainVideos); err != nil {
		log.Error("s.dao.AddChannelArc(mid:%d,cid:%d) error(%v)", mid, cid, err)
		return
	} else if lastID > 0 {
		var arcs []*model.ChannelArc
		for _, v := range remainVideos {
			arc := &model.ChannelArc{ID: lastID, Mid: mid, Cid: cid, Aid: v.Aid, OrderNum: v.OrderNum, Mtime: xtime.Time(ts.Unix())}
			arcs = append(arcs, arc)
		}
		_ = s.dao.AddChannelArcCache(context.Background(), mid, cid, arcs)
	}
	return
}

func (s *Service) arcsCheck(c context.Context, mid int64, aids []int64) (err error) {
	var arcs map[int64]*arcmdl.Arc
	if arcs, err = s.archives(c, aids); err != nil {
		log.Error("s.archives error(%v)", err)
		return
	}
	for _, aid := range aids {
		if arc, ok := arcs[aid]; !ok || !arc.IsNormal() || arc.Author.Mid != mid {
			err = ecode.ChFakeAid
			return
		}
	}
	return
}

// DelChannelArc delete channel archive.
func (s *Service) DelChannelArc(ctx context.Context, mid, cid, aid int64) error {
	var orderNum int
	eg := errgroup.WithContext(ctx)
	eg.Go(func(ctx context.Context) error {
		videos, err := s.dao.ChannelVideos(ctx, mid, cid, false)
		if err != nil {
			log.Error("s.dao.Channel(%d,%d) error(%v)", mid, cid, err)
			return err
		}
		if len(videos) == 0 {
			return ecode.ChNoArcs
		}
		check := false
		for _, video := range videos {
			if aid == video.Aid {
				check = true
				orderNum = video.OrderNum
			}
		}
		if !check {
			return ecode.ChNoArc
		}
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		return s.seriesUpgraded(ctx, mid)
	})
	if err := eg.Wait(); err != nil {
		log.Error("%+v", err)
		return err
	}
	affected, err := s.dao.DelChannelArc(ctx, mid, cid, aid, orderNum)
	if err != nil {
		log.Error("s.dao.DelChannelArc(%d,%d) error(%v)", mid, aid, err)
		return err
	}
	if affected > 0 {
		if err := s.dao.DelChannelArcCache(ctx, mid, cid, aid); err != nil {
			log.Error("%+v", err)
		}
		if err := s.setChannelArcSortCache(ctx, mid, cid); err != nil {
			log.Error("%+v", err)
		}
	}
	return nil
}

// SortChannelArc sort channel archive.
func (s *Service) SortChannelArc(ctx context.Context, mid, cid, aid int64, orderNum int) error {
	var (
		videos                                 []*model.ChannelArc
		bfSortBegin, bfSortEnd, chSort, afSort []*model.ChannelArcSort
		affected                               int64
		aidIndex, aidOn                        int
		aidCheck                               bool
		ts                                     = time.Now()
	)
	eg := errgroup.WithContext(ctx)
	eg.Go(func(ctx context.Context) error {
		var err error
		if videos, err = s.dao.ChannelVideos(ctx, mid, cid, false); err != nil {
			log.Error("s.dao.ChannelVideos(%d,%d) error(%v)", mid, cid, err)
			return err
		}
		if len(videos) == 0 {
			return ecode.ChNoArcs
		}
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		return s.seriesUpgraded(ctx, mid)
	})
	if err := eg.Wait(); err != nil {
		log.Error("%+v", err)
		return err
	}
	if len(videos) != 0 {
		videoLen := len(videos)
		if orderNum > videoLen {
			return xecode.RequestErr
		}
		for index, video := range videos {
			if aid == video.Aid {
				aidCheck = true
				aidIndex = index
				aidOn = video.OrderNum
				break
			}
		}
		if !aidCheck {
			return xecode.RequestErr
		}
		if orderNum > aidOn {
			chSort = append(chSort, &model.ChannelArcSort{Aid: aid, OrderNum: orderNum})
			for i, v := range videos {
				if i < videoLen-orderNum {
					bfSortBegin = append(bfSortBegin, &model.ChannelArcSort{Aid: v.Aid, OrderNum: v.OrderNum})
				} else if i >= videoLen-orderNum && i < aidIndex {
					chSort = append(chSort, &model.ChannelArcSort{Aid: v.Aid, OrderNum: v.OrderNum - 1})
				} else if i > aidIndex {
					bfSortEnd = append(bfSortEnd, &model.ChannelArcSort{Aid: v.Aid, OrderNum: v.OrderNum})
				}
			}
		} else if orderNum < aidOn {
			for i, v := range videos {
				if i < aidIndex {
					bfSortBegin = append(bfSortBegin, &model.ChannelArcSort{Aid: v.Aid, OrderNum: v.OrderNum})
				} else if i > aidIndex && i <= videoLen-orderNum {
					chSort = append(chSort, &model.ChannelArcSort{Aid: v.Aid, OrderNum: v.OrderNum + 1})
				} else if i > videoLen-orderNum {
					bfSortEnd = append(bfSortEnd, &model.ChannelArcSort{Aid: v.Aid, OrderNum: v.OrderNum})
				}
			}
			chSort = append(chSort, &model.ChannelArcSort{Aid: aid, OrderNum: orderNum})
		} else {
			return nil
		}
		afSort = append(afSort, bfSortBegin...)
		afSort = append(afSort, chSort...)
		afSort = append(afSort, bfSortEnd...)
	}
	affected, err := s.dao.EditChannelArc(ctx, mid, cid, ts, chSort)
	if err != nil {
		log.Error("s.dao.s.dao.EditChannelArc(%d,%d,%d,%d) error(%v)", mid, cid, aid, orderNum, err)
		return err
	}
	if affected > 0 {
		_ = s.dao.SetChannelArcSortCache(ctx, mid, cid, afSort)
	}
	return nil
}

func (s *Service) ChannelVideos(ctx context.Context, mid, cid int64, pn, ps int, isGuest, order bool, ctype int64) (*model.ChannelDetail, *model.ArcListButton, error) {
	var (
		res *model.ChannelDetail
		err error
	)
	if ctype == 1 {
		res, err = s.channelLivePlayback(ctx, mid, cid, int64(pn), int64(ps))
	} else {
		res, err = s.channelVideo(ctx, mid, cid, pn, ps, isGuest, order)
	}
	if err != nil {
		return nil, nil, err
	}
	// 稿件数量大于1时才出按钮
	var button *model.ArcListButton
	if len(res.Archives) > 1 && !s.forbidEpisodicButton(mid) {
		button = &model.ArcListButton{
			Text: s.c.PlayButton.Text,
		}
	}
	return res, button, nil
}

// nolint:gomnd
func (s *Service) channelLivePlayback(ctx context.Context, mid, seriesID, pn, ps int64) (*model.ChannelDetail, error) {
	reply, err := s.seriesGRPC.ListSeries(ctx, &seriesgrpc.ListSeriesReq{Mid: mid, State: seriesgrpc.SeriesOnline})
	if err != nil {
		return nil, err
	}
	var channel *model.Channel
	for _, val := range reply.GetSeriesList() {
		meta := val.GetMeta()
		if seriesID > 0 && meta.GetSeriesId() != seriesID {
			continue
		}
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
		break
	}
	if channel == nil {
		return nil, xecode.NothingFound
	}
	arcsReply, err := s.seriesGRPC.ListArchives(ctx, &seriesgrpc.ListArchivesReq{Mid: mid, SeriesId: channel.Cid, OnlyNormal: true, Pn: pn, Ps: ps})
	if err != nil {
		return nil, err
	}
	aids := arcsReply.GetAids()
	if len(aids) == 0 {
		return &model.ChannelDetail{Channel: channel}, nil
	}
	archives, err := s.archives(ctx, aids)
	if err != nil {
		return nil, err
	}
	var arcs []*model.ChannelArcs
	for _, aid := range aids {
		if arc, ok := archives[aid]; ok {
			if !arc.IsNormal() {
				continue
			}
			if arc.Access >= 10000 {
				arc.Stat.View = -1
			}
			var isLivePlayback bool
			for _, val := range s.c.LivePlayback.UpFrom {
				if arc.UpFromV2 == val {
					isLivePlayback = true
					break
				}
			}
			model.ClearAttrAndAccess(arc)
			arc := &model.ChannelArcs{BvArc: &model.BvArc{Arc: arc, Bvid: s.avToBv(arc.Aid)}, InterVideo: arc.AttrVal(arcapi.AttrBitSteinsGate) == arcapi.AttrYes, IsLivePlayback: isLivePlayback}
			arcs = append(arcs, arc)
		}
	}
	return &model.ChannelDetail{
		Channel:  channel,
		Archives: arcs,
	}, nil
}

// CshannelVideos get channel and channel video info.
func (s *Service) channelVideo(c context.Context, mid, cid int64, pn, ps int, isGuest, order bool) (res *model.ChannelDetail, err error) {
	var (
		channel  *model.Channel
		start    = (pn - 1) * ps
		end      = start + ps - 1
		arcs     []*model.BvArc
		chanArcs []*model.ChannelArcs
	)
	if channel, err = s.Channel(c, mid, cid); err != nil {
		return
	}
	res = &model.ChannelDetail{Channel: channel}
	if arcs, err = s.channelArc(c, mid, cid, start, end, isGuest, order); err != nil {
		return
	}
	chanArcs = make([]*model.ChannelArcs, 0)
	for _, v := range arcs {
		if v == nil {
			continue
		}
		if v.Arc != nil && !v.Arc.IsNormal() {
			v.Arc.Stat = arcmdl.Stat{}
		}
		var isLivePlayback bool
		for _, val := range s.c.LivePlayback.UpFrom {
			if v.Arc.UpFromV2 == val {
				isLivePlayback = true
				break
			}
		}
		tmp := &model.ChannelArcs{BvArc: v, InterVideo: v.AttrVal(arcapi.AttrBitSteinsGate) == arcapi.AttrYes, IsLivePlayback: isLivePlayback}
		chanArcs = append(chanArcs, tmp)
	}
	res.Archives = chanArcs
	return
}

func (s *Service) ChannelAids(ctx context.Context, channelID int64, sort string) ([]int64, error) {
	var order bool
	if sort == "asc" {
		order = true
	}
	mid, cid, err := s.channelMid(ctx, channelID)
	if err != nil {
		return nil, err
	}
	if mid == 0 || cid == 0 {
		return nil, xecode.NothingFound
	}
	res, err := s.channelVideos(ctx, mid, cid, 0, s.c.Rule.MaxChArcLimit, order)
	if err != nil {
		return nil, err
	}
	if len(res) == 0 {
		return nil, xecode.NothingFound
	}
	var aids []int64
	for _, val := range res {
		if val == nil {
			continue
		}
		aids = append(aids, val.Aid)
	}
	return aids, nil
}

func (s *Service) channelMid(ctx context.Context, channelID int64) (mid int64, cid int64, err error) {
	cached := true
	if mid, cid, err = s.dao.ChannelMidCache(ctx, channelID); err != nil {
		log.Error("%+v", err)
		cached = false
	}
	if mid != 0 && cid != 0 {
		return mid, cid, nil
	}
	if mid, cid, err = s.dao.ChannelMid(ctx, channelID); err != nil {
		return 0, 0, err
	}
	if cached {
		s.cache.Do(ctx, func(ctx context.Context) {
			if err := s.dao.SetChannelMidCache(ctx, channelID, mid, cid); err != nil {
				log.Error("%+v", err)
			}
		})
	}
	return mid, cid, nil
}

func (s *Service) channelVideos(c context.Context, mid, cid int64, start, end int, order bool) (res []*model.ChannelArc, err error) {
	var (
		videos   []*model.ChannelArc
		addCache = true
	)
	if res, err = s.dao.ChannelArcsCache(c, mid, cid, start, end, order); err != nil {
		addCache = false
	} else if len(res) > 0 {
		return
	}
	if videos, err = s.dao.ChannelVideos(c, mid, cid, order); err != nil {
		log.Error("s.dao.ChannelVideos(%d,%d) error(%v)", mid, cid, err)
		return
	} else if len(videos) > 0 {
		if addCache {
			s.cache.Do(c, func(c context.Context) {
				if err := s.dao.SetChannelArcsCache(c, mid, cid, videos); err != nil {
					log.Error("%+v", err)
				}
				if err := s.setChannelArcSortCache(c, mid, cid); err != nil {
					log.Error("%+v", err)
				}
			})
		}
		length := len(videos)
		if length < start {
			res = make([]*model.ChannelArc, 0)
			return
		}
		if length > end {
			res = videos[start : end+1]
		} else {
			res = videos[start:]
		}
	}
	return
}

// CheckChannelVideo check useless channel video.
func (s *Service) CheckChannelVideo(c context.Context, mid, cid int64) (err error) {
	var (
		videos []*model.ChannelArc
		aids   []int64
	)
	if videos, err = s.dao.ChannelVideos(c, mid, cid, false); err != nil {
		log.Error("s.dao.channelVideos(%d,%d) error(%v)", mid, cid, err)
		return
	}
	for _, v := range videos {
		aids = append(aids, v.Aid)
	}
	err = s.arcsCheck(c, mid, aids)
	return
}

// nolint:gomnd
func (s *Service) channelArc(c context.Context, mid, cid int64, start, end int, _, order bool) (res []*model.BvArc, err error) {
	var (
		videoAids []*model.ChannelArc
		archives  map[int64]*arcmdl.Arc
		aids      []int64
	)
	if videoAids, err = s.channelVideos(c, mid, cid, start, end, order); err != nil {
		log.Error("s.dao.ChannelVideos(%d,%d) error(%v)", mid, cid, err)
		return
	} else if len(videoAids) == 0 {
		res = _emptyChArc
		return
	}
	for _, video := range videoAids {
		aids = append(aids, video.Aid)
	}
	if archives, err = s.archives(c, aids); err != nil {
		log.Error("s.arc.Archives3(%v) error(%v)", aids, err)
		return
	}
	for _, video := range videoAids {
		if arc, ok := archives[video.Aid]; ok {
			if arc.IsNormal() {
				if arc.Access >= 10000 {
					arc.Stat.View = -1
				}
				model.ClearAttrAndAccess(arc)
				res = append(res, &model.BvArc{Arc: arc, Bvid: s.avToBv(arc.Aid)})
			} else {
				res = append(res, &model.BvArc{
					Arc: &arcmdl.Arc{
						Aid:     video.Aid,
						Stat:    arc.Stat,
						PubDate: arc.PubDate,
						State:   arc.State,
					},
					Bvid: s.avToBv(arc.Aid),
				})
			}
		}
	}
	return
}

func (s *Service) setChannelArcSortCache(c context.Context, mid, cid int64) (err error) {
	var (
		videos []*model.ChannelArc
		sorts  []*model.ChannelArcSort
	)
	if videos, err = s.dao.ChannelVideos(c, mid, cid, false); err != nil {
		log.Error("s.dao.ChannelVideos(%d,%d) error(%v)", mid, cid, err)
		return
	} else if len(videos) == 0 {
		return
	}
	for _, v := range videos {
		sort := &model.ChannelArcSort{Aid: v.Aid, OrderNum: v.OrderNum}
		sorts = append(sorts, sort)
	}
	return s.dao.SetChannelArcSortCache(c, mid, cid, sorts)
}

func (s *Service) archives(c context.Context, aids []int64) (map[int64]*arcmdl.Arc, error) {
	mutex := sync.Mutex{}
	aidsLen := len(aids)
	group := errgroup.WithContext(c)
	archives := make(map[int64]*arcmdl.Arc, aidsLen)
	for i := 0; i < aidsLen; i += _aidBulkSize {
		var partAids []int64
		if i+_aidBulkSize > aidsLen {
			partAids = aids[i:]
		} else {
			partAids = aids[i : i+_aidBulkSize]
		}
		group.Go(func(ctx context.Context) (err error) {
			var arcs *arcmdl.ArcsReply
			arg := &arcmdl.ArcsRequest{Aids: partAids}
			if arcs, err = s.arcClient.Arcs(ctx, arg); err != nil {
				log.Error("s.arcClient.Arcs(%v) error(%v)", partAids, err)
				return
			}
			mutex.Lock()
			for _, v := range arcs.Arcs {
				archives[v.Aid] = v
			}
			mutex.Unlock()
			return
		})
	}
	if err := group.Wait(); err != nil {
		return nil, err
	}
	return archives, nil
}
