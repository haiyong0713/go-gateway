package service

import (
	"context"
	"sort"
	"strconv"
	"time"

	accwarden "git.bilibili.co/bapis/bapis-go/account/service"
	favmdl "git.bilibili.co/bapis/bapis-go/community/model/favorite"
	favpb "git.bilibili.co/bapis/bapis-go/community/service/favorite"
	"go-common/library/log"
	"go-common/library/net/metadata"
	xtime "go-common/library/time"
	arcmdl "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/playlist/ecode"
	"go-gateway/app/web-svr/playlist/interface/model"
)

const (
	_first       = 1
	_sortDefault = 0
	_sortByMTime = 1
	_sortByView  = 2
	_defaultFid  = 0
)

var _empPlaylists = make([]*model.Playlist, 0)

// White  playlist white list.
func (s *Service) White(c context.Context, mid int64) (res map[string]bool, err error) {
	_, power := s.allowMids[mid]
	res = make(map[string]bool, 1)
	res["power"] = power
	return
}

// Add add playlist.
func (s *Service) Add(c context.Context, mid int64, public int32, name, description, cover, cookie, accessKey string) (pid int64, err error) {
	var (
		favReply *favpb.AddFolderReply
		ts       = time.Now()
	)
	if _, ok := s.allowMids[mid]; !ok {
		err = ecode.PlDenied
		return
	}
	arg := &favpb.AddFolderReq{Typ: int32(favmdl.TypePlayVideo), Mid: mid, Name: name, Description: description, Cover: cover, Public: public, Cookie: cookie, AccessKey: accessKey}
	if favReply, err = s.favClient.AddFolder(c, arg); err != nil {
		log.Error("s.favClient.AddFolder(%v) error(%v)", arg, err)
		return
	}
	if pid, err = s.dao.Add(c, mid, favReply.Fid); err != nil {
		log.Error("s.dao.Add(%d,%d) error(%v)", mid, favReply.Fid, err)
	} else if pid > 0 {
		s.cache.Do(c, func(c context.Context) {
			stat := &model.PlStat{ID: pid, Mid: mid, Fid: favReply.Fid, MTime: xtime.Time(ts.Unix())}
			s.dao.SetPlStatCache(c, mid, pid, stat)
		})
		if err = s.dao.RegReply(c, pid, mid); err != nil {
			err = nil
		}
	}
	return
}

// Del delete playlist.
func (s *Service) Del(c context.Context, mid, pid int64) (err error) {
	var (
		affected int64
		stat     *model.PlStat
	)
	if stat, err = s.plByPid(c, pid); err != nil {
		log.Error("s.plByPid(%d,%d) error(%v)", mid, pid, err)
		return
	}
	arg := &favpb.DelFolderReq{Typ: int32(favmdl.TypePlayVideo), Mid: mid, Fid: stat.Fid}
	if _, err = s.favClient.DelFolder(c, arg); err != nil {
		log.Error("s.favClient.DelFolder(%+v) error(%v)", arg, err)
		return
	}
	if affected, err = s.dao.Del(c, pid); err != nil {
		log.Error("s.dao.Del(%d) error(%v)", pid, err)
		return
	} else if affected > 0 {
		s.dao.DelPlCache(c, mid, pid)
	}
	return
}

// Update update playlist.
func (s *Service) Update(c context.Context, mid, pid int64, public int32, name, description, cover, cookie, accessKey string) (err error) {
	var (
		stat *model.PlStat
	)
	if stat, err = s.plByPid(c, pid); err != nil {
		log.Error("s.plByPid(%d) error(%v)", pid, err)
		return
	}
	arg := &favpb.UpdateFolderReq{Typ: int32(favmdl.TypePlayVideo), Fid: stat.Fid, Mid: mid, Name: name, Description: description, Cover: cover, Public: public, Cookie: cookie, AccessKey: accessKey}
	if _, err = s.favClient.UpdateFolder(c, arg); err != nil {
		log.Error("s.fav.UpdateFolder(%+v) error(%v)", arg, err)
		return
	}
	s.updatePlTime(c, mid, pid)
	return
}

func (s *Service) updatePlTime(c context.Context, mid, pid int64) (err error) {
	var (
		affected int64
		ts       = time.Now()
		stat     *model.PlStat
	)
	if affected, err = s.dao.Update(c, pid); err != nil {
		err = nil
		log.Error("s.dao.Update(%d) error(%v)", pid, err)
		return
	} else if affected > 0 {
		s.cache.Do(c, func(c context.Context) {
			if stat, err = s.plByPid(c, pid); err != nil {
				err = nil
			} else {
				stat.MTime = xtime.Time(ts.Unix())
				s.dao.SetPlStatCache(c, mid, pid, stat)
			}
		})
	}
	return
}

// Info playlist stat info.
func (s *Service) Info(c context.Context, mid, pid int64) (res *model.Playlist, err error) {
	var (
		fav        *favpb.UserFolderReply
		stat       *model.PlStat
		infoReply  *accwarden.InfoReply
		isFav      bool
		isFavReply *favpb.IsFavoredReply
		ip         = metadata.String(c, metadata.RemoteIP)
	)
	if stat, err = s.plByPid(c, pid); err != nil {
		return
	}
	if stat == nil || stat.ID == 0 {
		err = ecode.PlNotExist
		log.Error("s.plByPid(%d) error(%v)", pid, stat)
		return
	}
	arg := &favpb.UserFolderReq{Typ: int32(favmdl.TypePlayVideo), Fid: stat.Fid, Mid: stat.Mid}
	if fav, err = s.favClient.UserFolder(c, arg); err != nil || fav == nil {
		log.Error("s.favClient.UserFolder(%+v) error(%v)", arg, err)
		return
	}
	if fav.Res.State == int32(favmdl.StateIsDel) {
		err = ecode.PlNotExist
		log.Error("s.favClient.UserFolder(%d) is del error(%v)", pid, err)
		return
	}
	// author
	if infoReply, err = s.accClient.Info3(c, &accwarden.MidReq{Mid: fav.Res.Mid, RealIp: ip}); err != nil {
		log.Error("s.accClient.Info3 error(%v)", err)
		return
	}
	if mid > 0 {
		if isFavReply, err = s.favClient.IsFavored(c, &favpb.IsFavoredReq{Typ: int32(favmdl.TypePlayList), Mid: mid, Oid: pid}); err != nil || isFavReply == nil {
			log.Error("s.favClient.IsFavored(%d,%d) error(%d)", mid, pid, err)
			err = nil
			isFav = false
		} else {
			isFav = isFavReply.Faved
		}
	}
	owner := &arcmdl.Author{Mid: fav.Res.Mid, Name: infoReply.Info.Name, Face: infoReply.Info.Face}
	fav.Res.MTime = stat.MTime
	res = &model.Playlist{Pid: pid, Folder: fav.Res, Stat: &model.Stat{Pid: stat.ID, View: stat.View, Reply: stat.Reply, Fav: stat.Fav, Share: stat.Share}, Author: owner, IsFavorite: isFav}
	return
}

func (s *Service) plInfo(c context.Context, mid, pid int64, ip string) (res *model.PlStat, err error) {
	var fav *favpb.UserFolderReply
	if res, err = s.plByPid(c, pid); err != nil {
		return
	}
	if res == nil || res.ID == 0 {
		err = ecode.PlNotExist
		log.Error("s.plByPid(%d) res(%v)", pid, res)
		return
	}
	arg := &favpb.UserFolderReq{Typ: int32(favmdl.TypePlayVideo), Fid: res.Fid, Mid: res.Mid}
	if fav, err = s.favClient.UserFolder(c, arg); err != nil || fav == nil {
		log.Error("s.favClient.UserFolder((%+v) error(%v)", arg, err)
		return
	}
	if fav.Res.State == int32(favmdl.StateIsDel) {
		err = ecode.PlNotExist
		log.Error("s.favClient.UserFolder((%d) state(%d)", pid, fav.Res.State)
		return
	}
	if mid > 0 && mid != res.Mid {
		err = ecode.PlNotUser
	}
	return
}

// List playlist.
func (s *Service) List(c context.Context, mid int64, pn, ps, sortType int) (res []*model.Playlist, count int, err error) {
	var (
		start   = (pn - 1) * ps
		end     = start + ps - 1
		plStats []*model.PlStat
		ip      = metadata.String(c, metadata.RemoteIP)
	)
	if plStats, err = s.plsByMid(c, mid); err != nil {
		return
	}
	count = len(plStats)
	if count == 0 || count < start {
		res = _empPlaylists
		return
	}
	switch sortType {
	case _sortDefault, _sortByMTime:
		sort.Slice(plStats, func(i, j int) bool { return plStats[i].MTime > plStats[j].MTime })
	case _sortByView:
		sort.Slice(plStats, func(i, j int) bool { return plStats[i].View > plStats[j].View })
	}
	if count > end {
		plStats = plStats[start : end+1]
	} else {
		plStats = plStats[start:]
	}
	res, err = s.batchFav(c, mid, plStats, ip)
	return
}

// AddFavorite add playlist to favorite.
func (s *Service) AddFavorite(c context.Context, mid, pid int64) (err error) {
	if _, err = s.Info(c, 0, pid); err != nil {
		return
	}
	arg := &favpb.AddFavReq{Tp: int32(favmdl.TypePlayList), Mid: mid, Oid: pid, Fid: _defaultFid}
	if _, err = s.favClient.AddFav(c, arg); err != nil {
		log.Error("s.favClient.AddFav(%+v) error(%v)", arg, err)
	}
	return
}

// DelFavorite del playlist from favorite.
func (s *Service) DelFavorite(c context.Context, mid, pid int64) (err error) {
	arg := &favpb.DelFavReq{Tp: int32(favmdl.TypePlayList), Mid: mid, Oid: pid, Fid: _defaultFid}
	if _, err = s.favClient.DelFav(c, arg); err != nil {
		log.Error("s.favClient.DelFav(%+v) error(%v)", arg, err)
	}
	return
}

// ListFavorite playlist list.
func (s *Service) ListFavorite(c context.Context, mid, vmid int64, pn, ps, sortType int) (res []*model.Playlist, count int, err error) {
	var (
		plStats []*model.PlStat
		favRes  *favpb.FavoritesReply
		pids    []int64
		tmpFavs map[int64]*favpb.ModelFavorite
		tmpRs   []*model.Playlist
		ip      = metadata.String(c, metadata.RemoteIP)
	)
	arg := &favpb.FavoritesReq{Tp: int32(favmdl.TypePlayList), Mid: mid, Uid: vmid, Fid: _defaultFid, Pn: int32(pn), Ps: int32(ps)}
	if favRes, err = s.favClient.Favorites(c, arg); err != nil {
		log.Error("s.favClient.Favorites(%+v) error(%v)", arg, err)
		return
	}
	if favRes == nil || len(favRes.Res.List) == 0 {
		res = _empPlaylists
		return
	}
	tmpFavs = make(map[int64]*favpb.ModelFavorite)
	for _, fav := range favRes.Res.List {
		pids = append(pids, fav.Oid)
		tmpFavs[fav.Oid] = fav
	}
	if plStats, err = s.plsByPid(c, pids); err != nil {
		return
	}
	count = int(favRes.Res.Page.Count)
	tmpRs, err = s.batchFav(c, mid, plStats, ip)
	for _, v := range tmpRs {
		v.FavoriteTime = tmpFavs[v.Pid].Mtime
		res = append(res, v)
	}
	return
}

func (s *Service) batchFav(c context.Context, uid int64, plStats []*model.PlStat, ip string) (res []*model.Playlist, err error) {
	var (
		fVMids   []*favpb.FolderID
		tmpStats map[string]*model.PlStat
		favRes   *favpb.FoldersReply
		stat     *model.Stat
	)
	tmpStats = make(map[string]*model.PlStat)
	for _, v := range plStats {
		statKey := strconv.FormatInt(v.Mid, 10) + "_" + strconv.FormatInt(v.Fid, 10)
		tmpStats[statKey] = &model.PlStat{ID: v.ID, Mid: v.Mid, Fid: v.Fid, View: v.View, Reply: v.Reply, Fav: v.Fav, Share: v.Share, MTime: v.MTime}
		fVMids = append(fVMids, &favpb.FolderID{Fid: v.Fid, Mid: v.Mid})
	}
	arg := &favpb.FoldersReq{Typ: int32(favmdl.TypePlayVideo), Mid: uid, Ids: fVMids}
	if favRes, err = s.favClient.Folders(c, arg); err != nil {
		log.Error("s.favClient.Folders(%+v) error(%v)", arg, err)
		return
	}
	for _, fav := range favRes.Res {
		statKey := strconv.FormatInt(fav.Mid, 10) + "_" + strconv.FormatInt(fav.ID, 10)
		plStat := tmpStats[statKey]
		stat = &model.Stat{Pid: plStat.ID, View: plStat.View, Fav: plStat.Fav, Reply: plStat.Reply, Share: plStat.Share}
		fav.MTime = plStat.MTime
		res = append(res, &model.Playlist{Pid: plStat.ID, Folder: fav, Stat: stat})
	}
	return
}
