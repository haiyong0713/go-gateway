package common

import (
	"context"
	"encoding/json"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-car/interface/dao/fm"
	"go-gateway/app/app-svr/app-car/interface/model"
	"go-gateway/app/app-svr/app-car/interface/model/common"
	"go-gateway/app/app-svr/app-car/interface/model/fm_v2"

	api "git.bilibili.co/bapis/bapis-go/community/service/thumbup"
	"github.com/pkg/errors"
)

func (s *Service) FmListRefactor(c context.Context, param *fm_v2.FmListParam) (*fm_v2.FmListResp, error) {
	// Step1: 提取FM播单的id列表
	handlerResp, err := FmListParamStrategy(c, param)
	if err != nil {
		log.Error("FmListRefactor FmListParamStrategy error:%+v, param:%+v", err, param)
		return nil, err
	}
	// Step2: 依据id列表获取原始稿件信息 + 分P信息
	matResp, err := s.material(c, handlerResp.OidParam, param.DeviceInfo)
	if err != nil {
		log.Error("FmListRefactor s.material error:%+v, params:%+v ,device:%+v", err, handlerResp.OidParam, param.DeviceInfo)
		return nil, err
	}
	// Step3: 原始信息转为Item
	fmItems := s.generateFmItems(c, handlerResp.OidList, matResp, param)
	resp := &fm_v2.FmListResp{
		FmItems:      fmItems,
		PageNext:     handlerResp.PageNext,
		PagePrevious: handlerResp.PagePre,
		HasNext:      handlerResp.HasNext,
		HasPrevious:  handlerResp.HasPrevious,
	}
	return resp, nil
}

func (s *Service) generateFmItems(c context.Context, oids []int64, materials *common.CarContext, param *fm_v2.FmListParam) []*common.FmItem {
	fmItems := make([]*common.FmItem, 0)
	for _, id := range oids {
		originData := &common.OriginData{
			MaterialType: common.MaterialTypeUGC,
			Oid:          id,
		}
		materials.OriginData = originData
		item := s.formItem(materials, param.DeviceInfo)
		if item == nil {
			log.Warn("generateFmItems no item, oid:%d", id)
			continue
		}
		// 填充PlayList
		if len(materials.UGCViewResp) > 0 {
			view := materials.UGCViewResp[id]
			if view != nil {
				playList := make([]*common.Playlist, 0)
				for _, page := range view.Pages {
					playList = append(playList, &common.Playlist{
						Title: page.Part,
						Aid:   id,
						Cid:   page.Cid,
					})
				}
				item.Playlist = playList
			} else {
				log.Warn("generateFmItems no view, oid:%d", id)
			}
		}
		fmItems = append(fmItems, &common.FmItem{Item: item})
	}
	return s.postGenerateFmItems(c, fmItems, param)
}

// postGenerateFmItems FM播单列表后置处理
func (s *Service) postGenerateFmItems(c context.Context, items []*common.FmItem, param *fm_v2.FmListParam) []*common.FmItem {
	if len(items) == 0 {
		return items
	}
	// 点赞
	aids := make([]int64, 0)
	for _, v := range items {
		aids = append(aids, v.Oid)
	}
	likeBatch, err := s.thumbupDao.HasLikeBatch(c, param.Mid, _thumbupBiz, param.Buvid, aids)
	if err != nil {
		log.Error("postGenerateFmItems s.thumbupDao.HasLikeBatch error:%+v, aids:%+v, param:%+v", err, aids, param)
	} else {
		for _, v := range items {
			v.IsLike = likeBatch[v.Oid] == api.State_STATE_LIKE
		}
	}
	// mini播控栏标题（一期均使用所属合集标题）
	miniTitle := s.getMiniTitle(c, param)
	for _, v := range items {
		v.MiniTitle = miniTitle
	}
	return items
}

// getMiniTitle mini播控栏标题
func (s *Service) getMiniTitle(c context.Context, param *fm_v2.FmListParam) string {
	if param.FmType == fm_v2.AudioRelate {
		return _searchMiniTitle
	}
	if param.ServerExtra != "" && param.ServerExtra != "null" {
		extra := &fm_v2.ServerExtra{}
		err := json.Unmarshal([]byte(param.ServerExtra), extra)
		if err != nil {
			log.Error("postGenerateFmItems json unmarshal err:%+v, param:%+v", err, param)
			return ""
		}
		return extra.FmTitle
	} else {
		// 降级走接口查询
		log.Warn("postGenerateFmItems no server_extra, param:%+v", param)
		req := &fm_v2.HandleTabItemsReq{
			DeviceInfo: param.DeviceInfo,
			FmType:     param.FmType,
			FmId:       param.FmId,
		}
		itemsResp, err := TabItemsStrategy(c, req)
		if err != nil {
			log.Error("postGenerateFmItems TabItemsStrategy err:%+v, req:%+v", err, req)
			return ""
		}
		if len(itemsResp.TabItems) > 0 {
			return itemsResp.TabItems[0].Title
		}
	}
	return ""
}

// extractPageReq 解析请求中的分页参数
func extractPageReq(fmType fm_v2.FmType, next string, pre string, ps int, manualRefresh int, dev model.DeviceInfo) (*fm_v2.PageReq, error) {
	var (
		pageNext  *fm_v2.PageInfo
		pagePre   *fm_v2.PageInfo
		nextEmpty bool
		preEmpty  bool
		pageSize  int
		err       error
	)
	pageNext, nextEmpty, err = validatePageInfo(fmType, next)
	if err != nil {
		return nil, errors.Wrap(err, "validate page next error")
	}
	pagePre, preEmpty, err = validatePageInfo(fmType, pre)
	if err != nil {
		return nil, errors.Wrap(err, "validate page pre error")
	}

	if !nextEmpty {
		pageSize = getPageSize(pageNext, ps)
	} else if !preEmpty {
		pageSize = getPageSize(pagePre, ps)
	} else if ps > 0 {
		pageSize = ps
	} else {
		pageSize = _defaultPs
		if fmType == fm_v2.AudioHome {
			pageSize = fm.GetHomePs(dev)
		}
	}
	return &fm_v2.PageReq{
		PageNext:      pageNext,
		PagePre:       pagePre,
		NextEmpty:     nextEmpty,
		PreEmpty:      preEmpty,
		PageSize:      pageSize,
		ManualRefresh: manualRefresh,
	}, nil
}

func validatePageInfo(fmType fm_v2.FmType, input string) (info *fm_v2.PageInfo, empty bool, err error) {
	if input == "" || input == "null" {
		return new(fm_v2.PageInfo), true, nil
	}
	info = new(fm_v2.PageInfo)
	if err = json.Unmarshal([]byte(input), info); err != nil {
		return nil, false, errors.Wrap(ecode.RequestErr, err.Error())
	}
	if fmType == fm_v2.AudioHistory && (info.Max <= 0 || info.ViewAt <= 0) ||
		fmType == fm_v2.AudioVertical && info.Pn <= 0 ||
		fmType == fm_v2.AudioUp && info.Oid <= 0 ||
		fmType == fm_v2.AudioSeason && info.Oid <= 0 ||
		fmType == fm_v2.AudioSeasonUp && info.Oid <= 0 ||
		fmType == fm_v2.AudioHome && info.Pn <= 0 ||
		fmType == fm_v2.AudioHomeV2 && info.Pn <= 0 {
		return nil, false, errors.Wrapf(ecode.RequestErr, "page info incomplete:%+v, fmType:%s", info, fmType)
	}
	return info, false, nil
}

func getPageSize(info *fm_v2.PageInfo, globalPs int) int {
	if globalPs > 0 {
		return globalPs
	}
	if info != nil && info.Ps > 0 {
		return info.Ps
	}
	return _defaultPs
}
