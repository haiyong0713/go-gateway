package service

import (
	"context"
	"html/template"
	"sync"

	favmdl "git.bilibili.co/bapis/bapis-go/community/model/favorite"
	favpb "git.bilibili.co/bapis/bapis-go/community/service/favorite"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/sync/errgroup"
	arcmdl "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/playlist/ecode"
	"go-gateway/app/web-svr/playlist/interface/conf"
	"go-gateway/app/web-svr/playlist/interface/model"
)

const (
	_vUpload     = "vupload"
	_aidBulkSize = 50
)

var _empAids = make([]int64, 0)

// Videos get playlist video list by pid.
func (s *Service) Videos(c context.Context, pid int64, pn, ps int) (res *model.ArcList, err error) {
	var (
		aids     []int64
		arcSorts []*model.ArcSort
		arcs     map[int64]*arcmdl.ViewReply
		stat     *model.PlStat
		ip       = metadata.String(c, metadata.RemoteIP)
	)
	res = &model.ArcList{List: make([]*model.PlView, 0)}
	if stat, err = s.plByPid(c, pid); err != nil || stat == nil {
		return
	}
	if _, err = s.favClient.UserFolder(c, &favpb.UserFolderReq{Typ: int32(favmdl.TypePlayVideo), Mid: stat.Mid, Fid: stat.Fid}); err != nil {
		log.Error("s.favClient.UserFolder(%d,%d) error(%v)", stat.Mid, stat.Fid, err)
		return
	}
	start := (pn - 1) * ps
	end := start + ps - 1
	if arcSorts, err = s.videos(c, pid, start, end); err != nil {
		return
	}
	//TODO check aids from fav
	for _, v := range arcSorts {
		aids = append(aids, v.Aid)
	}
	if arcs, err = s.views(c, aids, ip); err != nil {
		log.Error("s.arc.Views3(%v) error(%v)", aids, err)
		return
	}
	for _, v := range arcSorts {
		if arc, ok := arcs[v.Aid]; ok {
			view := &model.PlView{View: &model.View{Arc: arc.Arc, Pages: arc.Pages}, PlayDesc: template.HTMLEscapeString(v.Desc)}
			res.List = append(res.List, view)
		}
	}
	return
}

// ToView get playlist view page data.
func (s *Service) ToView(c context.Context, mid, pid int64) (res *model.ToView, err error) {
	var (
		aids     []int64
		arcSorts []*model.ArcSort
		views    map[int64]*arcmdl.ViewReply
		info     *model.Playlist
		ip       = metadata.String(c, metadata.RemoteIP)
	)
	if info, err = s.Info(c, mid, pid); err != nil {
		return
	}
	res = &model.ToView{Playlist: info}
	if arcSorts, err = s.videos(c, pid, 0, int(info.Count)-1); err != nil {
		return
	}
	//TODO check aids from fav
	for _, v := range arcSorts {
		aids = append(aids, v.Aid)
	}
	if views, err = s.views(c, aids, ip); err != nil {
		log.Error("s.views(%v) error(%v)", aids, err)
		return
	}
	res.List = make([]*model.View, 0)
	for _, v := range arcSorts {
		if arc, ok := views[v.Aid]; ok {
			view := &model.View{Arc: arc.Arc, Pages: arc.Pages}
			res.List = append(res.List, view)
		}
	}
	return
}

// CheckVideo add  video to playlist.
func (s *Service) CheckVideo(c context.Context, mid, pid int64, aids []int64) (videos model.Videos, err error) {
	var (
		stat *model.PlStat
		fav  *favpb.UserFolderReply
		ip   = metadata.String(c, metadata.RemoteIP)
	)
	if stat, err = s.plByPid(c, pid); err != nil {
		return
	}
	if stat == nil {
		err = ecode.PlNotExist
		log.Error("s.plByPid(%d) error(%v)", pid, err)
		return
	}
	argFolder := &favpb.UserFolderReq{Typ: int32(favmdl.TypePlayVideo), Fid: stat.Fid, Mid: stat.Mid, Vmid: _defaultFid}
	if fav, err = s.favClient.UserFolder(c, argFolder); err != nil || fav == nil {
		log.Error("s.favClient.UserFolder(%+v) error(%v)", argFolder, err)
		return
	}
	if videos, _, _, err = s.filterArc(c, mid, pid, aids, ip); err != nil {
		log.Error("s.filterArc(%v) error(%v)", aids, err)
	}
	return
}

// AddVideo add  video to playlist.
func (s *Service) AddVideo(c context.Context, mid, pid int64, aids []int64) (videos model.Videos, err error) {
	var (
		lastID, sort, fid int64
		arcSorts          []*model.ArcSort
		ip                = metadata.String(c, metadata.RemoteIP)
	)
	if videos, sort, fid, err = s.filterArc(c, mid, pid, aids, ip); err != nil {
		log.Error("s.filterArc(%v) error(%v)", aids, err)
		return
	}
	if len(videos.RightAids) == 0 {
		return
	}
	for _, aid := range videos.RightAids {
		sort += conf.Conf.Rule.SortStep
		arcSorts = append(arcSorts, &model.ArcSort{Aid: aid, Sort: sort, Desc: ""})
	}
	arg := &favpb.MultiAddReq{Typ: int32(favmdl.TypePlayVideo), Mid: mid, Oids: videos.RightAids, Fid: fid}
	if _, err = s.favClient.MultiAdd(c, arg); err != nil {
		log.Error("s.favClient.MultiAdd(%+v) error(%v)", arg, err)
		return
	}
	if lastID, err = s.dao.BatchAddArc(c, pid, arcSorts); err != nil || lastID == 0 {
		log.Error("s.dao.BatchAddArc(%d,%+v) error(%v)", pid, arcSorts, err)
		return
	}
	if lastID > 0 {
		s.cache.Do(c, func(c context.Context) {
			s.dao.SetArcsCache(c, pid, arcSorts)
		})
	}
	s.updatePlTime(c, mid, pid)
	return
}

// DelVideo del video from playlist.
func (s *Service) DelVideo(c context.Context, mid, pid int64, aids []int64) (err error) {
	var (
		affected int64
		stat     *model.PlStat
	)
	if stat, err = s.plByPid(c, pid); err != nil {
		return
	}
	if stat == nil {
		err = ecode.PlNotExist
		log.Error("s.plByPid(%d) error(%v)", pid, err)
		return
	}
	arg := &favpb.MultiDelReq{Typ: int32(favmdl.TypePlayVideo), Mid: mid, Oids: aids, Fid: stat.Fid}
	if _, err = s.favClient.MultiDel(c, arg); err != nil {
		log.Error("s.favClient.MultiDel(%+v) error(%v)", arg, err)
		return
	}
	if affected, err = s.dao.BatchDelArc(c, pid, aids); err != nil || affected == 0 {
		log.Error("s.dao.BatchDelArc(%d,%v) error(%v)", mid, aids, err)
		return
	}
	if affected > 0 {
		s.cache.Do(c, func(c context.Context) {
			s.dao.DelArcsCache(c, pid, aids)
		})
	}
	s.updatePlTime(c, mid, pid)
	return
}

// SortVideo sort playlist video.
func (s *Service) SortVideo(c context.Context, mid, pid, aid, sort int64) (err error) {
	var (
		info                                         *favpb.FavoritesReply
		aidSort, preSort, afSort, orderNum, affected int64
		desc                                         string
		arcs                                         []*model.ArcSort
		plStat                                       *model.PlStat
		top, bottom, reset                           bool
		isPlaylist                                   *favpb.IsFavoredReply
	)
	if plStat, err = s.plByPid(c, pid); err != nil {
		return
	}
	if plStat == nil {
		err = ecode.PlNotExist
		log.Error("s.plByPid(%d) error(%v)", pid, err)
		return
	}
	if plStat.ID == 0 {
		err = ecode.PlNotExist
		return
	} else if mid != plStat.Mid {
		err = ecode.PlNotUser
		return
	}
	if info, err = s.favClient.Favorites(c, &favpb.FavoritesReq{Tp: int32(favmdl.TypePlayVideo), Mid: mid, Fid: plStat.Fid, Pn: 1, Ps: 1}); err != nil || info == nil {
		log.Error("s.favClient.Favorites(%d,%d) error(%v)", mid, plStat.Fid, err)
		return
	} else if sort > int64(info.Res.Page.Count) {
		err = ecode.PlSortOverflow
		return
	}
	if isPlaylist, err = s.favClient.IsFavoredByFid(c, &favpb.IsFavoredByFidReq{Type: int32(favmdl.TypePlayVideo), Mid: mid, Fid: plStat.Fid, Oid: aid}); err != nil || isPlaylist == nil {
		log.Error("s.favClient.IsFavoredByFid(%d,%d,%d) error(%v)", mid, plStat.Fid, aid, err)
		return
	} else if !isPlaylist.Faved {
		err = ecode.PlVideoAlreadyDel
		return
	}
	if arcs, err = s.videos(c, pid, 0, int(info.Res.Page.Count)-1); err != nil {
		return
	}
	if sort == _first {
		top = true
	} else if sort == int64(info.Res.Page.Count) {
		bottom = true
	}
	for k, v := range arcs {
		if k == 0 && top {
			afSort = v.Sort
		}
		if k == len(arcs)-1 && bottom {
			preSort = v.Sort
		}
		if aid == v.Aid {
			if sort == int64(k+1) {
				return
			}
			aidSort = v.Sort
			desc = v.Desc
		}
		if sort == int64(k+1) {
			if !top && !bottom {
				if aidSort > sort {
					preSort = arcs[k].Sort
					afSort = arcs[k+1].Sort
				} else {
					preSort = arcs[k-1].Sort
					afSort = arcs[k].Sort
				}
			}
		}
	}
	if top {
		orderNum = afSort / 2
	} else if bottom {
		orderNum = preSort + conf.Conf.Rule.SortStep
	} else {
		orderNum = preSort + (afSort-preSort)/2
	}
	if orderNum == preSort || orderNum == afSort || orderNum <= conf.Conf.Rule.MinSort || orderNum > s.maxSort {
		reset = true
		if affected, err = s.resetArcSort(c, pid); err != nil {
			log.Error("s.dao.UpdateArcSort(%d,%d) error(%v)", pid, aid, err)
			return
		}
	} else {
		if affected, err = s.dao.UpdateArcSort(c, pid, aid, orderNum); err != nil {
			log.Error("s.dao.UpdateArcSort(%d,%d) error(%v)", pid, aid, err)
			return
		}
	}
	if affected > 0 {
		s.cache.Do(c, func(c context.Context) {
			if reset {
				if err = s.dao.DelCache(c, pid); err != nil {
					log.Error("s.dao.DelCache() pid(%d), error(%v)", pid, err)
					return
				}
				s.videos(c, pid, 0, int(info.Res.Page.Count)-1)
			} else {
				s.dao.AddArcCache(c, pid, &model.ArcSort{Aid: aid, Sort: orderNum, Desc: desc})
			}
		})
	}
	return
}

// EditVideoDesc edit playlist video desc.
func (s *Service) EditVideoDesc(c context.Context, mid, pid, aid int64, desc string) (err error) {
	var (
		affected   int64
		plStat     *model.PlStat
		isPlaylist *favpb.IsFavoredReply
	)
	if plStat, err = s.plByPid(c, pid); err != nil {
		return
	}
	if plStat == nil {
		err = ecode.PlNotExist
		log.Error("s.plByPid(%d) error(%v)", pid, err)
		return
	}
	if plStat.ID == 0 {
		err = ecode.PlNotExist
		return
	} else if mid != plStat.Mid {
		err = ecode.PlNotUser
		return
	}
	if isPlaylist, err = s.favClient.IsFavoredByFid(c, &favpb.IsFavoredByFidReq{Type: int32(favmdl.TypePlayVideo), Mid: mid, Fid: plStat.Fid, Oid: aid}); err != nil || isPlaylist == nil {
		log.Error("s.fav.IsFavedByFid(%d,%d,%d) error(%v)", mid, plStat.Fid, aid, err)
		return
	} else if !isPlaylist.Faved {
		err = ecode.PlVideoAlreadyDel
		return
	}
	if affected, err = s.dao.UpdateArcDesc(c, pid, aid, desc); err != nil {
		log.Error("s.dao.UpdateArcDesc(%d,%d,%s) error(%v)", pid, aid, desc, err)
		return
	}
	if affected > 0 {
		s.cache.Do(c, func(c context.Context) {
			s.dao.SetArcDescCache(c, pid, aid, desc)
		})
	}
	s.updatePlTime(c, mid, pid)
	return
}

// SearchVideos search add videos.
func (s *Service) SearchVideos(c context.Context, pn, ps int, query string) (res []*model.SearchArc, count int, err error) {
	if res, count, err = s.dao.SearchVideo(c, pn, ps, query); err != nil {
		log.Error("s.dao.SearchVideo(%s) error(%v)", query, err)
	}
	if len(res) == 0 {
		res = make([]*model.SearchArc, 0)
	}
	return
}

func (s *Service) videos(c context.Context, pid int64, start, end int) (res []*model.ArcSort, err error) {
	var (
		arcs []*model.ArcSort
	)
	if arcs, err = s.dao.ArcsCache(c, pid, start, end); err != nil || len(arcs) == 0 {
		err = nil
		if arcs, err = s.dao.Videos(c, pid); err != nil {
			log.Error("s.dao.Videos(%d) error(%v)", pid, err)
			return
		}
		length := len(arcs)
		if length > 0 {
			s.cache.Do(c, func(c context.Context) {
				s.dao.SetArcsCache(c, pid, arcs)
			})
		}
		if length < start {
			res = []*model.ArcSort{}
			return
		}
		if length > end+1 {
			res = arcs[start : end+1]
		} else {
			res = arcs[start:]
		}
		return
	}
	res = arcs
	return
}

func (s *Service) resetArcSort(c context.Context, pid int64) (affected int64, err error) {
	var (
		arcs, afArcs []*model.ArcSort
	)
	if arcs, err = s.dao.Videos(c, pid); err != nil {
		log.Error("s.dao.Videos(%d) error(%v)", pid, err)
		return
	}
	sort := conf.Conf.Rule.BeginSort
	for _, v := range arcs {
		sort += s.c.Rule.SortStep
		afArcs = append(afArcs, &model.ArcSort{Aid: v.Aid, Desc: v.Desc, Sort: sort})
	}
	affected, err = s.dao.BatchUpdateArcSort(c, pid, afArcs)
	return
}

func (s *Service) filterArc(c context.Context, mid, pid int64, aids []int64, ip string) (res model.Videos, sort, fid int64, err error) {
	var (
		mutex                         = sync.Mutex{}
		aidsLen                       = len(aids)
		rightAids, rsRight, wrongAids []int64
		group, errCtx                 = errgroup.WithContext(c)
		tmpArc                        []*model.ArcSort
		exists                        map[int64]bool
		stat                          *model.Playlist
	)
	sort = conf.Conf.Rule.BeginSort
	if stat, err = s.Info(c, 0, pid); err != nil {
		return
	}
	if mid != stat.Mid {
		err = ecode.PlNotUser
		return
	}
	fid = stat.ID
	exists = make(map[int64]bool, stat.Count)
	if stat.Count > 0 {
		if stat.Count > int32(conf.Conf.Rule.MaxVideoCnt) {
			err = ecode.PlVideoOverflow
			return
		}
		if tmpArc, err = s.videos(c, pid, 0, int(stat.Count)-1); err != nil {
			return
		}
		for _, v := range tmpArc {
			exists[v.Aid] = true
		}
		if tmpLen := len(tmpArc); tmpLen < int(stat.Count) {
			sort = tmpArc[tmpLen-1].Sort
		} else {
			sort = tmpArc[stat.Count-1].Sort
		}
	}
	tmpRight := make(map[int64]struct{})
	for i := 0; i < aidsLen; i += _aidBulkSize {
		var partAids []int64
		if i+_aidBulkSize > aidsLen {
			partAids = aids[i:]
		} else {
			partAids = aids[i : i+_aidBulkSize]
		}
		group.Go(func() (err error) {
			var arcs *arcmdl.ViewsReply
			arg := &arcmdl.ViewsRequest{Aids: partAids}
			if arcs, err = s.arcClient.Views(errCtx, arg); err != nil {
				log.Error("s.arcClient.Views(%v) error(%v)", partAids, err)
				return
			}
			mutex.Lock()
			for _, aid := range partAids {
				if arcReply, ok := arcs.Views[aid]; !ok || arcs.Views[aid] == nil {
					wrongAids = append(wrongAids, aid)
				} else if !arcReply.Arc.IsNormal() ||
					exists[aid] ||
					arcReply.Arc.Rights.UGCPay == 1 ||
					arcReply.Arc.AttrVal(arcmdl.AttrBitIsBangumi) == arcmdl.AttrYes ||
					arcReply.Arc.AttrVal(arcmdl.AttrBitIsMovie) == arcmdl.AttrYes ||
					(len(arcReply.Pages) > 0 && arcReply.Pages[0].From != _vUpload) {
					wrongAids = append(wrongAids, aid)
				} else {
					rightAids = append(rightAids, aid)
					tmpRight[aid] = struct{}{}
				}
			}
			mutex.Unlock()
			return
		})
	}
	err = group.Wait()
	if rightAids == nil {
		rightAids = _empAids
		rsRight = _empAids
	} else if wrongAids == nil {
		wrongAids = _empAids
	}
	if int(stat.Count)+len(rightAids) > conf.Conf.Rule.MaxVideoCnt {
		err = ecode.PlVideoOverflow
		return
	}
	for _, aid := range aids {
		if _, ok := tmpRight[aid]; ok {
			rsRight = append(rsRight, aid)
		}
	}
	res = model.Videos{RightAids: rsRight, WrongAids: wrongAids}
	return
}

func (s *Service) views(c context.Context, aids []int64, ip string) (views map[int64]*arcmdl.ViewReply, err error) {
	var (
		mutex         = sync.Mutex{}
		aidsLen       = len(aids)
		group, errCtx = errgroup.WithContext(c)
	)
	views = make(map[int64]*arcmdl.ViewReply, aidsLen)
	for i := 0; i < aidsLen; i += _aidBulkSize {
		var partAids []int64
		if i+_aidBulkSize > aidsLen {
			partAids = aids[i:]
		} else {
			partAids = aids[i : i+_aidBulkSize]
		}
		group.Go(func() (err error) {
			var arcs *arcmdl.ViewsReply
			arg := &arcmdl.ViewsRequest{Aids: partAids}
			if arcs, err = s.arcClient.Views(errCtx, arg); err != nil {
				log.Error("s.arcClient.Views(%v) error(%v)", partAids, err)
				return
			}
			mutex.Lock()
			for _, v := range arcs.Views {
				views[v.Arc.Aid] = v
			}
			mutex.Unlock()
			return
		})
	}
	err = group.Wait()
	return
}
