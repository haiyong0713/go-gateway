package common

import (
	"context"
	"encoding/json"

	"go-common/library/log"

	model "go-gateway/app/app-svr/app-car/interface/model"
	commonmdl "go-gateway/app/app-svr/app-car/interface/model/common"
	archivegrpc "go-gateway/app/app-svr/archive/service/api"
)

const (
	// 我收藏的视频
	_favVideo = 2
	_favPGC   = 24

	// 追番OGV
	favTypeBangumi = 1
	favTypeCinema  = 2

	_attrBitPublic = uint32(0)
)

func (s *Service) Favorite(c context.Context, req *commonmdl.FavoriteReq, mid int64, buvid, _, _ string) (resp []*commonmdl.Favorite, err error) {
	tmpFolders, err := s.favDao.UserFolders(c, mid, 0, _favVideo)
	if err != nil {
		log.Error("Favorite UserFolders(%v, %v, %v) error(%+v)", req, mid, buvid, err)
		return
	}
	// 兼容逻辑: 收藏夹接口未返回cover 临时拉取稿件cover兼容
	var aidm = make(map[int64]struct{})
	for _, tmpFolder := range tmpFolders {
		for _, r := range tmpFolder.RecentRes {
			aidm[r.Oid] = struct{}{}
		}
	}
	var arcm map[int64]*archivegrpc.Arc
	if len(aidm) > 0 {
		var (
			aids   []int64
			errTmp error
		)
		for aid := range aidm {
			aids = append(aids, aid)
		}
		if arcm, errTmp = s.archiveDao.ArcsAll(c, aids); errTmp != nil {
			log.Error("ArcsAll UserFolders(%+v) error(%+v)", aids, err)
		}
	}
	for _, tmpFolder := range tmpFolders {
		item := &commonmdl.Favorite{
			Fid:   tmpFolder.ID,
			Mid:   tmpFolder.Mid,
			State: model.AttrVal(tmpFolder.Attr, _attrBitPublic), // 0公开 1私密
			Count: int(tmpFolder.Count),
			Name:  tmpFolder.Name,
			Cover: tmpFolder.Cover,
		}
		for _, r := range tmpFolder.RecentRes {
			if arc, ok := arcm[r.Oid]; ok {
				item.Cover = arc.Pic
				break
			}
		}
		resp = append(resp, item)
	}
	return
}

func (s *Service) FavoriteVideo(c context.Context, req *commonmdl.FavoriteVideoReq, mid int64, _ string) (resp *commonmdl.FavoriteVideoResp, err error) {
	// 翻页前置逻辑
	var pageNext *commonmdl.FavoriteVideoPageNext
	if req.PageNext != "" {
		if err = json.Unmarshal([]byte(req.PageNext), &pageNext); err != nil {
			log.Error("FavoriteVideo json.Unmarshal() error(%v)", err)
			return
		}
	}
	var (
		pn, ps = 1, 20
	)
	if pageNext != nil {
		pn = pageNext.Pn
		ps = pageNext.Ps
	}
	if req.Ps != 0 {
		ps = req.Ps
	}
	// 获取list
	tmpFavVideos, err := s.favDao.FavoritesAll(c, mid, mid, req.Fid, _favVideo, pn, ps)
	if err != nil {
		log.Error("FavoriteVideo FavoritesAll(%v,%v,%v,%v,%v,%v) error(%v)", mid, mid, req.Fid, _favVideo, pn, ps, err)
		return
	}
	// 物料ID分离
	var (
		aidm  = make(map[int64][]int64)
		epidm = make(map[int32]struct{})
	)
	for _, tmpFavVideo := range tmpFavVideos.GetList() {
		switch tmpFavVideo.Type {
		case _favVideo:
			aidm[tmpFavVideo.Oid] = []int64{}
		case _favPGC:
			epidm[int32(tmpFavVideo.Oid)] = struct{}{}
		}
	}
	var materialParams = new(commonmdl.Params)
	if len(aidm) > 0 {
		materialParams.ArchiveReq = new(commonmdl.ArchiveReq)
		for aid := range aidm {
			var playAv = &archivegrpc.PlayAv{Aid: aid}
			materialParams.ArchiveReq.PlayAvs = append(materialParams.ArchiveReq.PlayAvs, playAv)
		}
	}
	if len(epidm) > 0 {
		var epids []int32
		for epid := range epidm {
			epids = append(epids, epid)
		}
		materialParams.EpisodeReq = new(commonmdl.EpisodeReq)
		materialParams.EpisodeReq.Epids = epids
	}
	carContext, err := s.material(c, materialParams, req.DeviceInfo)
	if err != nil {
		b, _ := json.Marshal(materialParams)
		log.Error("FavoriteVideo material(%+v) error(%v)", string(b), err)
		return
	}
	// 聚合卡片
	resp = new(commonmdl.FavoriteVideoResp)
	for _, tmpFavVideo := range tmpFavVideos.GetList() {
		carContext.OriginData = new(commonmdl.OriginData)
		switch tmpFavVideo.Type {
		case _favVideo:
			carContext.OriginData.MaterialType = commonmdl.MaterialTypeUGC
			carContext.OriginData.Oid = tmpFavVideo.Oid
		case _favPGC:
			carContext.OriginData.MaterialType = commonmdl.MaterialTypeOGVEP
			carContext.OriginData.Oid = tmpFavVideo.Oid
		default:
			log.Warn("FavoriteVideo unknown fav video type %+v", tmpFavVideo)
			continue
		}
		item := s.formItem(carContext, req.DeviceInfo)
		if item != nil {
			resp.Items = append(resp.Items, item)
		}
	}
	// 翻页后置逻辑
	resp.PageNext = &commonmdl.FavoriteVideoPageNext{
		Pn: pn + 1,
		Ps: ps,
	}
	return
}

func (s *Service) FavoriteBangumi(c context.Context, req *commonmdl.FavoriteBangumiReq, mid int64, buvid string) (resp *commonmdl.FavoriteBangumiResp, err error) {
	// 翻页前置逻辑
	var (
		pageNext *commonmdl.FavoriteOGVPageNext
		pn, ps   = 1, 20
	)
	if req.PageNext != "" {
		if err = json.Unmarshal([]byte(req.PageNext), &pageNext); err != nil {
			log.Error("FavoriteBangumi json.Unmarshal() error(%v)", err)
			return
		}
	}
	if pageNext != nil {
		pn = pageNext.Pn
		ps = pageNext.Ps
	}
	if req.Ps != 0 {
		ps = req.Ps
	}
	// OGV收藏公共方法
	resp = new(commonmdl.FavoriteBangumiResp)
	resp.Items = s.FavOGV(c, req.DeviceInfo, mid, buvid, pn, ps, favTypeBangumi)
	// 分页后置逻辑
	resp.PageNext = &commonmdl.FavoriteOGVPageNext{
		Pn: pn + 1,
		Ps: ps,
	}
	return
}

func (s *Service) FavoriteCinema(c context.Context, req *commonmdl.FavoriteCinemaReq, mid int64, buvid string) (resp *commonmdl.FavoriteCinemaResp, err error) {
	// 翻页前置逻辑
	var (
		pageNext *commonmdl.FavoriteOGVPageNext
		pn, ps   = 1, 20
	)
	if req.PageNext != "" {
		if err = json.Unmarshal([]byte(req.PageNext), &pageNext); err != nil {
			log.Error("FavoriteBangumi json.Unmarshal() error(%v)", err)
			return
		}
	}
	if pageNext != nil {
		pn = pageNext.Pn
		ps = pageNext.Ps
	}
	if req.Ps != 0 {
		ps = req.Ps
	}
	// OGV收藏公共方法
	resp = new(commonmdl.FavoriteCinemaResp)
	resp.Items = s.FavOGV(c, req.DeviceInfo, mid, buvid, pn, ps, favTypeCinema)
	// 分页后置逻辑
	resp.PageNext = &commonmdl.FavoriteOGVPageNext{
		Pn: pn + 1,
		Ps: ps,
	}
	return
}

func (s *Service) FavOGV(c context.Context, deviceInfo model.DeviceInfo, mid int64, buvid string, pn, ps, favType int) (res []*commonmdl.Item) {
	tmpFavs, err := s.bangumiDao.MyFollows(c, mid, favType, pn, ps)
	if err != nil {
		log.Error("%+v", err)
		return
	}
	// 获取物料
	var sidm = make(map[int32]struct{})
	for _, tmpFav := range tmpFavs {
		if tmpFav.SeasonId != 0 {
			sidm[tmpFav.SeasonId] = struct{}{}
		}
	}
	var materialParams = new(commonmdl.Params)
	if len(sidm) > 0 {
		materialParams.SeasonReq = new(commonmdl.SeasonReq)
		var sids []int32
		for sid := range sidm {
			sids = append(sids, sid)
		}
		materialParams.SeasonReq.Sids = sids
	}
	carContext, err := s.material(c, materialParams, deviceInfo)
	if err != nil {
		b, _ := json.Marshal(materialParams)
		log.Error("FavOGV material(%v, %+v) error(%v)", favType, string(b), err)
		return
	}
	for _, tmpFav := range tmpFavs {
		carContext.OriginData = &commonmdl.OriginData{
			MaterialType: commonmdl.MaterialTypeOGVSeaon,
			Oid:          int64(tmpFav.SeasonId),
		}
		item := s.formItem(carContext, deviceInfo)
		if item != nil {
			if tmpFav.Progress != nil {
				item.IndexShow = model.HighLightString(tmpFav.Progress.IndexShow) // 看过标亮
				if tmpFav.Progress.Time == 0 {
					item.IndexShow = tmpFav.Progress.IndexShow
				}
			}
			res = append(res, item)
		}
	}
	return
}

func (s *Service) FavoriteToView(c context.Context, req *commonmdl.FavoriteToViewReq, mid int64, buvid string) (resp *commonmdl.FavoriteToViewResp, err error) {
	// 翻页前置逻辑
	var pageNext *commonmdl.FavoriteToViewPageNext
	if req.PageNext != "" {
		if err = json.Unmarshal([]byte(req.PageNext), &pageNext); err != nil {
			log.Error("FavoriteToView json.Unmarshal() error(%v)", err)
			return
		}
	}
	var (
		pn, ps = 1, 20
	)
	if pageNext != nil {
		pn = pageNext.Pn
		ps = pageNext.Ps
	}
	if req.Ps != 0 {
		ps = req.Ps
	}
	// 获取物料
	tmpToViews, err := s.favDao.UserToViews(c, mid, pn, ps)
	if err != nil {
		log.Error("FavoriteToView json.Unmarshal() error(%v)", err)
		return
	}
	var aidm = make(map[int64]struct{})
	for _, tmpToView := range tmpToViews {
		if tmpToView.Aid != 0 {
			aidm[tmpToView.Aid] = struct{}{}
		}
	}
	var materialParams = new(commonmdl.Params)
	if len(aidm) > 0 {
		materialParams.ArchiveReq = new(commonmdl.ArchiveReq)
		for aid := range aidm {
			var playAv = &archivegrpc.PlayAv{Aid: aid}
			materialParams.ArchiveReq.PlayAvs = append(materialParams.ArchiveReq.PlayAvs, playAv)
		}
	}
	resp = new(commonmdl.FavoriteToViewResp)
	carContext, err := s.material(c, materialParams, req.DeviceInfo)
	if err != nil {
		b, _ := json.Marshal(materialParams)
		log.Error("FavoriteToView material(%+v) error(%v)", string(b), err)
		return resp, nil
	}
	for _, tmpToView := range tmpToViews {
		carContext.OriginData = new(commonmdl.OriginData)
		carContext.OriginData.MaterialType = commonmdl.MaterialTypeUGC
		carContext.OriginData.Oid = tmpToView.Aid
		item := s.formItem(carContext, req.DeviceInfo)
		if item != nil {
			resp.Items = append(resp.Items, item)
		}
	}
	// 翻页后置逻辑
	resp.PageNext = &commonmdl.FavoriteToViewPageNext{
		Pn: pn + 1,
		Ps: ps,
	}
	return
}
