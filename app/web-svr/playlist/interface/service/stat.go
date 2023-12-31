package service

import (
	"context"

	favmdl "git.bilibili.co/bapis/bapis-go/community/model/favorite"
	favpb "git.bilibili.co/bapis/bapis-go/community/service/favorite"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-gateway/app/web-svr/playlist/interface/model"
)

func (s *Service) plsByMid(c context.Context, mid int64) (res []*model.PlStat, err error) {
	if res, err = s.dao.StatsCache(c, mid); err != nil || len(res) == 0 {
		err = nil
		if res, err = s.dao.PlsByMid(c, mid); err != nil {
			log.Error("s.dao.PlsByMid(%d) error(%v)", mid, err)
			return
		}
		if len(res) > 0 {
			s.cache.Do(c, func(c context.Context) {
				s.dao.SetStatsCache(c, mid, res)
			})
		}
	}
	return
}

func (s *Service) plByPid(c context.Context, pid int64) (res *model.PlStat, err error) {
	var pls []*model.PlStat
	pids := []int64{pid}
	if pls, err = s.dao.PlsCache(c, pids); err != nil {
		err = nil
	} else if len(pls) != 0 {
		res = pls[0]
		return
	}
	if res, err = s.dao.PlByPid(c, pid); err != nil {
		log.Error("s.dao.PlByPid(%d) error(%v)", pid, err)
	}
	return
}
func (s *Service) plsByPid(c context.Context, pids []int64) (res []*model.PlStat, err error) {
	var (
		tmpRs []*model.PlStat
		rsMap map[int64]*model.PlStat
	)
	if tmpRs, err = s.dao.PlsCache(c, pids); err != nil || len(pids) != len(tmpRs) || len(tmpRs) == 0 {
		err = nil
		if tmpRs, err = s.dao.PlsByPid(c, pids); err != nil {
			log.Error("s.dao.PlsByPid(%+v) error(%v)", pids, err)
		}
		if len(tmpRs) > 0 {
			s.cache.Do(c, func(c context.Context) {
				s.dao.SetPlCache(c, tmpRs)
			})
		}
	}
	rsMap = make(map[int64]*model.PlStat)
	for _, v := range tmpRs {
		rsMap[v.ID] = v
	}
	for _, v := range pids {
		if rsMap[v] == nil {
			continue
		}
		res = append(res, rsMap[v])
	}
	return
}

// SetStat  set playlist stat cache.
func (s *Service) SetStat(c context.Context, arg *model.PlStat) (err error) {
	var fav *favpb.UserFolderReply
	argFav := &favpb.UserFolderReq{Typ: int32(favmdl.TypePlayVideo), Fid: arg.Fid, Mid: arg.Mid}
	if fav, err = s.favClient.UserFolder(c, argFav); err != nil || fav == nil {
		log.Error("SetStat s.favClient.UserFolder(%+v) error(%v)", argFav, err)
		return
	}
	log.Info("service SetStat(%v) favState(%d)", arg, fav.Res.State)
	if fav.Res.State == int32(favmdl.StateNormal) {
		if err = s.dao.SetPlStatCache(c, arg.Mid, arg.ID, arg); err != nil {
			log.Error("SetStat s.dao.SetPlStatCache(%d,%d) error(%v)", arg.Mid, arg.ID, err)
		}
	}
	return
}

// PubView pub playlist view.
func (s *Service) PubView(c context.Context, pid, aid int64) (err error) {
	var (
		pls *model.PlStat
		ip  = metadata.String(c, metadata.RemoteIP)
	)
	if pls, err = s.plInfo(c, 0, pid, ip); err != nil {
		return
	}
	//TODO aid in playlist
	err = s.cache.Do(c, func(c context.Context) {
		s.dao.PubView(c, pid, aid, pls.View)
	})
	return
}

// PubShare pub playlist share.
func (s *Service) PubShare(c context.Context, pid, aid int64) (err error) {
	var (
		pls *model.PlStat
		ip  = metadata.String(c, metadata.RemoteIP)
	)
	if pls, err = s.plInfo(c, 0, pid, ip); err != nil {
		return
	}
	//TODO aid in playlist
	s.cache.Do(c, func(c context.Context) {
		s.dao.PubShare(c, pid, aid, pls.Share)
	})
	return
}
