package common

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/sync/errgroup.v2"
	xecode "go-gateway/app/app-svr/app-car/ecode"
	"go-gateway/app/app-svr/app-car/interface/model"
	commonmdl "go-gateway/app/app-svr/app-car/interface/model/common"
	archivegrpc "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/pkg/idsafe/bvid"
	thumbupmdl "go-main/app/community/thumbup/service/model"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
	hisgrpc "git.bilibili.co/bapis/bapis-go/community/interface/history"
	grpcShortURL "git.bilibili.co/bapis/bapis-go/platform/interface/shorturl"
	api "git.bilibili.co/bapis/bapis-go/serial/service"
)

const (
	rcmdTypeAV = "av"
)

func (s *Service) ViewDetail(c context.Context, req *commonmdl.ViewDetailReq, mid int64, buvid, cookie, referer string) (resp *commonmdl.Item, err error) {
	switch req.Otype {
	case commonmdl.ViewTypeUGC, commonmdl.ViewTypeUgcMulti, commonmdl.ViewTypeUgcSingle, commonmdl.ViewTypeVideoSerial, commonmdl.ViewTypeVideoChannel, commonmdl.ViewTypeFmSerial, commonmdl.ViewTypeFmChannel:
		if resp, err = s.viewDetailUGC(c, req, mid, buvid); err != nil {
			log.Error("ViewDetail viewDetailUGC(%v,%v,%v) error(%+v)", req, mid, buvid, err)
		}
	case commonmdl.ViewTypeOGV:
		if resp, err = s.viewDetailOGV(c, req, mid, buvid, cookie, referer); err != nil {
			log.Error("ViewDetail viewDetailOGV(%v,%v,%v,%v,%v) error(%+v)", req, mid, buvid, cookie, referer, err)
		}
	default:
		log.Error("ViewDetail invalid otype(%v)", req.Otype)
	}
	return
}

func (s *Service) viewDetailUGC(c context.Context, req *commonmdl.ViewDetailReq, mid int64, buvid string) (res *commonmdl.Item, err error) {
	// 获取物料
	eg := errgroup.WithContext(c)
	var carContext *commonmdl.CarContext
	eg.Go(func(ctx context.Context) (err error) {
		var materialParams = new(commonmdl.Params)
		materialParams.UGCViewReq = &commonmdl.UGCViewReq{Aids: []int64{req.Oid}}
		var errTmp error
		if carContext, errTmp = s.material(ctx, materialParams, req.DeviceInfo); errTmp != nil {
			b, _ := json.Marshal(materialParams)
			log.Error("Dynamic material(%+v) error(%v)", string(b), errTmp)
			return
		}
		return nil
	})
	var (
		isLike, isFavored bool
		his               *hisgrpc.ModelHistory
	)
	if mid > 0 || buvid != "" {
		eg.Go(func(ctx context.Context) (err error) {
			if his, err = s.historyDao.Progress(ctx, req.Oid, mid, buvid); err != nil {
				if code := ecode.Cause(err).Code(); code != ecode.NothingFound.Code() {
					log.Error("viewDetailUGC s.historyDao.Progress err:%+v", err)
				}
			}
			return nil
		})
		eg.Go(func(ctx context.Context) error {
			likeState, errTmp := s.thumbupDao.HasLike(ctx, mid, "archive", buvid, req.Oid)
			if errTmp != nil {
				log.Error("viewDetailUGC() HasLike(%v, %v, %v, %v) error(%+v)", mid, "archive", buvid, req.Oid, errTmp)
			}
			if likeState == thumbupmdl.StateLike {
				isLike = true
			}
			return nil
		})
		if mid > 0 {
			eg.Go(func(ctx context.Context) error {
				isFavored = s.favDao.IsFavored(ctx, mid, req.Oid)
				return nil
			})
		}
	}
	var longDesc string
	eg.Go(func(ctx context.Context) (err error) {
		var errTmp error
		if longDesc, errTmp = s.archiveDao.Description(ctx, req.Oid); errTmp != nil {
			log.Error("viewDetailUGC() Description(%+v) error(%+v)", req.Oid, errTmp)
		}
		return nil
	})
	if err = eg.Wait(); err != nil {
		log.Error("viewDetailUGC() eg() error(%+v)", err)
		return
	}
	// 聚合基础卡
	carContext.OriginData = &commonmdl.OriginData{MaterialType: commonmdl.MaterialTypeUGCView, Oid: req.Oid}
	if res = s.formItem(carContext, req.DeviceInfo); res == nil {
		return nil, xecode.AppNotVedio
	}
	var (
		tmpView         = carContext.UGCViewResp[req.Oid]
		tmpStat         *relationgrpc.StatReply
		authorRelations map[int64]*relationgrpc.InterrelationReply
	)
	if tmpView == nil {
		return
	}
	// 获取当前UP主粉丝数
	eg2 := errgroup.WithContext(c)
	eg2.Go(func(ctx context.Context) (err error) {
		var errTmp error
		if tmpStat, err = s.relationDao.StatGRPC(ctx, tmpView.Author.Mid); err != nil {
			log.Error("viewDetailUGC() Description(%+v) error(%+v)", req.Oid, errTmp)
		}
		return nil
	})
	if mid > 0 {
		eg2.Go(func(ctx context.Context) (err error) {
			var errTmp error
			if authorRelations, err = s.relationDao.RelationsInterrelations(ctx, mid, []int64{tmpView.Author.Mid}); err != nil {
				log.Error("viewDetailUGC() Description(%+v) error(%+v)", req.Oid, errTmp)
			}
			return nil
		})
	}
	if err = eg2.Wait(); err != nil {
		log.Error("viewDetailUGC() eg2 error(%+v)", err)
		return
	}
	// 详情页专用后置逻辑
	if longDesc != "" {
		res.Desc = longDesc
		if res.View != nil && res.View.Introduction != nil {
			res.View.Introduction.Desc = longDesc
		}
	}
	// 其他状态
	res.IsFollow = isFavored
	res.IsLike = isLike
	// UP主相关
	res.Author.FansCount = tmpStat.GetFollower()
	res.Author.Relation = commonmdl.RelationChange(tmpView.Author.Mid, authorRelations)
	// 历史
	if his != nil {
		res.View.History = &commonmdl.History{
			Cid:      his.Cid,
			Progress: his.Pro,
			ViewAt:   his.Unix,
		}
	}

	// 社区风险内容过滤
	if s.HitSixLimit(c, req.Oid) {
		log.Warnc(c, "viewDetailUGC HitSixLimit aid:%d, req:%s", req.Oid, toJson(req))
		return nil, xecode.AppVideoInsecurity
	}

	return
}

func (s *Service) viewDetailOGV(c context.Context, req *commonmdl.ViewDetailReq, mid int64, _, cookie, referer string) (res *commonmdl.Item, err error) {
	eg := errgroup.WithContext(c)
	var carContext *commonmdl.CarContext
	eg.Go(func(ctx context.Context) error {
		var materialParams = new(commonmdl.Params)
		materialParams.OGVViewReq = new(commonmdl.OGVViewReq)
		materialParams.OGVViewReq.Sid = req.Oid
		materialParams.OGVViewReq.AccessKey = req.AccessKey
		materialParams.OGVViewReq.Cookie = cookie
		materialParams.OGVViewReq.Referer = referer
		var errTmp error
		if carContext, errTmp = s.material(ctx, materialParams, req.DeviceInfo); errTmp != nil {
			b, _ := json.Marshal(materialParams)
			log.Error("viewDetailOGV() material(%+v) error(%+v)", string(b), errTmp)
			return errTmp
		}
		return nil
	})
	var useProfile *accountgrpc.Profile
	if mid > 0 {
		eg.Go(func(ctx context.Context) error {
			var errTmp error
			if useProfile, errTmp = s.accountDao.Profile3(ctx, mid); errTmp != nil {
				log.Error("PayState() Profile3(%v) error(%v)", mid, errTmp)
			}
			return nil
		})
	}
	if err = eg.Wait(); err != nil {
		log.Error("viewDetailOGV error(%v)", err)
		return
	}
	// 基础卡片组装
	carContext.OriginData = new(commonmdl.OriginData)
	carContext.OriginData.MaterialType = commonmdl.MaterialTypeOGVView
	carContext.OriginData.Oid = req.Oid
	res = s.formItem(carContext, req.DeviceInfo)
	// 后置逻辑 处理付费
	if res != nil && carContext.OGVViewResp != nil {
		switch carContext.OGVViewResp.Status {
		case 7, 13: // nolint:gomnd
			if useProfile != nil && useProfile.Vip.Status != 1 { // 大会员可看
				if res.View == nil {
					res.View = new(commonmdl.View)
				}
				res.View.PayInfo = &commonmdl.PayInfo{
					Ptype:      1,
					Icon:       "http://i0.hdslb.com/bfs/feed-admin/583fd78c9d7d46b4b769001bfd3536e181ba844b.png",
					Text:       "本片是大会员专享内容",
					ButtonText: "成为大会员",
					SubText:    "试看中 · 大会员免费看本片",
				}
			}
		case 8, 9, 12: // nolint:gomnd
			if carContext.OGVViewResp.UserStatus != nil && carContext.OGVViewResp.UserStatus.Pay == 0 { // 单集未付费
				if res.View == nil {
					res.View = new(commonmdl.View)
				}
				res.View.PayInfo = &commonmdl.PayInfo{
					Ptype:      2,
					Icon:       "http://i0.hdslb.com/bfs/feed-admin/583fd78c9d7d46b4b769001bfd3536e181ba844b.png",
					Text:       "本片需付费观看",
					ButtonText: "立即购买",
					SubText:    "试看中 · 本片为付费内容",
				}
			}
		}
	}
	return
}

func (s *Service) ViewRcmd(c context.Context, req *commonmdl.ViewRcmdReq, mid int64, buvid string) (*commonmdl.ViewRcmdResp, error) {
	if req.Otype == commonmdl.ViewTypeOGV {
		return s.ViewRcmdPgc(c, req, mid, buvid)
	}
	return s.ViewRcmdUgc(c, req, mid, buvid)
}

// ViewRcmdPgc ogv相关推荐.
func (s *Service) ViewRcmdPgc(c context.Context, req *commonmdl.ViewRcmdReq, mid int64, buvid string) (*commonmdl.ViewRcmdResp, error) {
	// 获取原始物料
	tmpRelates, err := s.rcmdDao.RelatePgc(c, mid, req.Oid, int64(req.Build), req.LoginEvent, buvid)
	if err != nil {
		log.Errorc(c, "ViewRcmdPgc s.rcmdDao.RelatePgc err=%+v, sid=%d, buvid=%s.", err, req.Oid, buvid)
		return &commonmdl.ViewRcmdResp{}, nil
	}
	// 聚合id
	var aidm = make(map[int64]struct{})
	for _, tmpRelate := range tmpRelates {
		if tmpRelate == nil || tmpRelate.Goto != rcmdTypeAV {
			continue
		}
		aidm[tmpRelate.ID] = struct{}{}
	}
	aids := make([]int64, 0)
	for aid := range aidm {
		aids = append(aids, aid)
	}
	if len(aids) == 0 {
		return &commonmdl.ViewRcmdResp{}, nil
	}
	creq := &commonItemsReq{
		Mid:   mid,
		Buvid: buvid,
		Aids:  aids,
	}
	items, err := s.commonItems(c, req.DeviceInfo, creq)
	if err != nil {
		log.Errorc(c, "ViewRcmdPgc s.commonItems err=%+v, buvid=%s, aids=%+v.", err, buvid, mid)
		return nil, err
	}
	if req.Build >= build203 {
		tmp := make([]*commonmdl.Item, 0)
		for _, v := range items {
			if v == nil || v.ItemType == commonmdl.ItemTypeUGCMulti {
				continue
			}
			tmp = append(tmp, v)
		}
		items = tmp
	}
	return &commonmdl.ViewRcmdResp{
		Items: items,
	}, nil
}

// ViewRcmdUgc ugv相关推荐.
func (s *Service) ViewRcmdUgc(c context.Context, req *commonmdl.ViewRcmdReq, mid int64, buvid string) (*commonmdl.ViewRcmdResp, error) {
	// 获取原始物料
	tmpRelates, err := s.rcmdDao.Relate(c, mid, req.Oid, buvid)
	if err != nil {
		log.Error("ViewList(%+v, %v, %v) Relate(%v, %v, %v) error(%+v)", req, mid, buvid, mid, req.Oid, buvid, err)
		return nil, err
	}
	// 聚合id
	var aidm = make(map[int64]struct{})
	for _, tmpRelate := range tmpRelates {
		if tmpRelate == nil || tmpRelate.Goto != rcmdTypeAV {
			continue
		}
		aidm[tmpRelate.ID] = struct{}{}
	}
	aids := make([]int64, 0)
	for aid := range aidm {
		aids = append(aids, aid)
	}
	if len(aids) == 0 {
		log.Errorc(c, "ViewRcmd ugc res is nil aid=%d, buvid=%s.", req.Oid, buvid)
		return &commonmdl.ViewRcmdResp{}, err
	}
	creq := &commonItemsReq{
		Mid:   mid,
		Buvid: buvid,
		Aids:  aids,
	}
	items, err := s.commonItems(c, req.DeviceInfo, creq)
	if err != nil {
		log.Errorc(c, "ViewList s.commonItems err=%+v, buvid=%s, aids=%+v.", err, buvid, mid)
		return nil, err
	}
	if req.Build >= build203 {
		tmp := make([]*commonmdl.Item, 0)
		for _, v := range items {
			if v == nil || v.ItemType == commonmdl.ItemTypeUGCMulti {
				continue
			}
			tmp = append(tmp, v)
		}
		items = tmp
	}
	return &commonmdl.ViewRcmdResp{
		Items: items,
	}, nil
}

// ViewSerial 合集.
func (s *Service) ViewPlaylist(c context.Context, req *commonmdl.ViewV2SerialReq) (*commonmdl.ViewV2SerialResp, error) {
	if commonmdl.ItemType(req.Otype) == commonmdl.ItemTypeVideoSerial || commonmdl.ItemType(req.Otype) == commonmdl.ItemTypeFmSerial {
		return s.viewPlaylistSerial(c, req)
	}
	return s.viewPlaylistChannel(c, req)
}

func (s *Service) viewPlaylistSerial(c context.Context, req *commonmdl.ViewV2SerialReq) (*commonmdl.ViewV2SerialResp, error) {
	// 1、按照类型（ugc合集、fm合集）查找aid列表

	// 分页
	var (
		next *commonmdl.SerialPageInfo
		pre  *commonmdl.SerialPageInfo
	)

	if req.PageNext != "" {
		next = new(commonmdl.SerialPageInfo)
		if err := json.Unmarshal([]byte(req.PageNext), next); err != nil {
			log.Errorc(c, "viewPlaylistSerial pageNext json.Unmarshal err=%+v, next=%s, buvid=%s.", err, req.PageNext, req.Buvid)
			return nil, err
		}
	}
	if req.PagePre != "" {
		pre = new(commonmdl.SerialPageInfo)
		if err := json.Unmarshal([]byte(req.PagePre), pre); err != nil {
			log.Errorc(c, "viewPlaylistSerial pagePre json.Unmarshal err=%+v, pre=%s, buvid=%s.", err, req.PagePre, req.Buvid)
			return nil, err
		}
	}

	// 合集历史记录
	var history *commonmdl.ViewV2SerialHistory
	if req.Mid > 0 {
		progress, err := s.serialDao.SerialProgress(c, req.Mid, req.Oid, commonmdl.ItemTypeToSerialBusinessType[commonmdl.ItemType(req.Otype)], req.Buvid)
		if err != nil {
			log.Errorc(c, "viewPlaylistSerial s.serialDao.SerialProgress type=%s, mid=%d, err=%+v.", req.Otype, req.Mid, err)
		}
		if progress != nil && progress.EpisodeType == api.EpisodeType_EpisodeTypeUGC {
			history = &commonmdl.ViewV2SerialHistory{
				Aid:      progress.Episode,
				Progress: progress.ViewAt,
			}
		}
	}

	// 确定起播aid：外部aid优先，其次历史aid
	var playAid int64
	if req.Aid > 0 {
		playAid = req.Aid
	} else if history != nil && history.Aid > 0 {
		playAid = history.Aid
	}
	if playAid > 0 {
		// PageNext和PagePre为空代表是第一次请求
		if req.PageNext == "" && req.PagePre == "" {
			next = &commonmdl.SerialPageInfo{
				Ps:          ten,
				Oid:         playAid,
				WithCurrent: true,
			}
			pre = &commonmdl.SerialPageInfo{
				Ps:  ten,
				Oid: playAid,
			}
		}
	}

	// 去查找合集的aid列表吧
	var video, fmCommon []*commonmdl.SerialArcReq
	switch commonmdl.ItemType(req.Otype) {
	case commonmdl.ItemTypeVideoSerial:
		video = []*commonmdl.SerialArcReq{{
			SerialId:      req.Oid,
			SerialPageReq: commonmdl.SerialPageReq{PageNext: next, PagePre: pre, Ps: req.Ps},
		}}
	case commonmdl.ItemTypeFmSerial:
		fmCommon = []*commonmdl.SerialArcReq{{
			SerialId:      req.Oid,
			SerialPageReq: commonmdl.SerialPageReq{PageNext: next, PagePre: pre, Ps: req.Ps},
		}}
	default:
		log.Errorc(c, "viewPlaylistSerial type wrong=%v, buvid=%s.", req.Otype, req.Buvid)
		return nil, ecode.RequestErr
	}
	material, err := s.material(c,
		&commonmdl.Params{
			SerialArcsReq: &commonmdl.SerialArcsReq{Video: video, FmCommon: fmCommon},
			Mid:           req.Mid,
			Buvid:         req.Buvid,
		},
		req.DeviceInfo,
	)
	if err != nil {
		log.Errorc(c, "viewPlaylistSerial s.material1 err=%+v, buvid=%s.", err, req.Buvid)
		return nil, err
	}
	if material == nil || material.SerialArcsResp == nil {
		log.Errorc(c, "viewPlaylistSerial serial material1 is nil, buvid=%s.", req.Buvid)
		return nil, xecode.AppMediaNotData
	}

	// 2、根据aid列表查找稿件信息
	var materialRes *commonmdl.SerialArcs
	switch commonmdl.ItemType(req.Otype) {
	case commonmdl.ItemTypeVideoSerial:
		materialRes = material.SerialArcsResp.Video[req.Oid]
	case commonmdl.ItemTypeFmSerial:
		materialRes = material.SerialArcsResp.FmCommon[req.Oid]
	default:
	}
	if materialRes == nil || len(materialRes.Aids) == 0 {
		log.Errorc(c, "viewPlaylistSerial serial material2 is nil, buvid=%s.", req.Buvid)
		return nil, xecode.AppMediaNotData
	}
	var cards []*commonmdl.Item
	tmp := make([]*commonmdl.Item, 0)
	creq := &commonItemsReq{
		Mid:   req.Mid,
		Buvid: req.Buvid,
		Aids:  materialRes.Aids,
	}
	cards, err = s.commonItems(c, req.DeviceInfo, creq)
	if err != nil {
		log.Errorc(c, "viewPlaylistSerial s.commonItems err=%+v, aids=%+v, buvid=%s.", err, materialRes.Aids, req.Buvid)
		return nil, err
	}
	for _, v := range cards {
		if v == nil || v.ItemType == commonmdl.ItemTypeUGCMulti {
			continue
		}
		tmp = append(tmp, v)
	}
	if len(tmp) == 0 {
		log.Errorc(c, "viewPlaylistSerial s.commonItems res is nil, aids=%+v,buvid=%s.", materialRes.Aids, req.Buvid)
		return nil, xecode.AppMediaNotData
	}
	cards = tmp

	var next1 *commonmdl.PageInfo
	var pre1 *commonmdl.PageInfo
	if materialRes.PageNext != nil {
		next1 = &commonmdl.PageInfo{
			Ps:  materialRes.PageNext.Ps,
			Oid: materialRes.PageNext.Oid,
		}
	}
	if materialRes.PagePre != nil {
		pre1 = &commonmdl.PageInfo{
			Ps:  materialRes.PagePre.Ps,
			Oid: materialRes.PagePre.Oid,
		}
	}
	return &commonmdl.ViewV2SerialResp{
		Cards:        cards,
		PageNext:     next1,
		PagePrevious: pre1,
		HasNext:      materialRes.HasNext,
		HasPrevious:  materialRes.HasPrevious,
		History:      history,
	}, nil
}

// ViewChannel 频道
func (s *Service) viewPlaylistChannel(c context.Context, req *commonmdl.ViewV2SerialReq) (*commonmdl.ViewV2SerialResp, error) {
	// 1、查询频道aid列表
	var next *commonmdl.ChannelPageInfo
	if req.PageNext != "" {
		next = new(commonmdl.ChannelPageInfo)
		if err := json.Unmarshal([]byte(req.PageNext), next); err != nil {
			log.Errorc(c, "viewPlaylistChannel pageNext json.Unmarshal err=%+v, next=%s, buvid=%s.", err, req.PageNext, req.Buvid)
			return nil, err
		}
	}

	var video, fm []*commonmdl.ChannelArcReq
	switch commonmdl.ItemType(req.Otype) {
	case commonmdl.ItemTypeVideoChannel:
		video = []*commonmdl.ChannelArcReq{{
			ChanId:   req.Oid,
			PageNext: next,
		}}
	case commonmdl.ItemTypeFmChannel:
		fm = []*commonmdl.ChannelArcReq{{
			ChanId:   req.Oid,
			PageNext: next,
		}}
	default:
		log.Errorc(c, "viewPlaylistChannel type wrong=%v, buvid=%s.", req.Otype, req.Buvid)
		return nil, ecode.RequestErr
	}
	material, err := s.material(c,
		&commonmdl.Params{
			ChannelArcsReq: &commonmdl.ChannelArcsReq{Video: video, Fm: fm},
			Mid:            req.Mid,
			Buvid:          req.Buvid,
		},
		req.DeviceInfo,
	)
	if err != nil {
		log.Errorc(c, "viewPlaylistChannel s.material1 err=%+v, buvid=%s.", err, req.Buvid)
		return nil, err
	}
	if material == nil || material.ChannelArcsResp == nil {
		log.Errorc(c, "viewPlaylistChannel channel material1 is nil, buvid=%s.", req.Buvid)
		return nil, xecode.AppMediaNotData
	}

	var materialRes *commonmdl.ChannelArcs
	switch commonmdl.ItemType(req.Otype) {
	case commonmdl.ItemTypeVideoChannel:
		materialRes = material.ChannelArcsResp.Video[req.Oid]
	case commonmdl.ItemTypeFmChannel:
		materialRes = material.ChannelArcsResp.Fm[req.Oid]
	default:
	}
	if materialRes == nil || len(materialRes.Aids) == 0 {
		log.Errorc(c, "viewPlaylistChannel serial material2 is nil, buvid=%s.", req.Buvid)
		return nil, xecode.AppMediaNotData
	}

	// 2、根据aid列表查找稿件信息

	// 如果有指定aid，则插入到第一个
	var playAid int64
	if req.Aid > 0 {
		playAid = req.Aid
	}
	j := 0
	exist := false
	for i, v := range materialRes.Aids {
		if v == playAid {
			j = i
			exist = true
			break
		}
	}
	if exist {
		t := make([]int64, 0)
		t = append(t, playAid)
		t = append(t, materialRes.Aids[:j]...)
		t = append(t, materialRes.Aids[j+1:]...)
		materialRes.Aids = t
	} else {
		t := make([]int64, 0)
		t = append(t, playAid)
		t = append(t, materialRes.Aids...)
		materialRes.Aids = t
	}

	// 查询aids
	var cards []*commonmdl.Item
	creq := &commonItemsReq{
		Mid:   req.Mid,
		Buvid: req.Buvid,
		Aids:  materialRes.Aids,
	}
	cards, err = s.commonItems(c, req.DeviceInfo, creq)
	if err != nil {
		log.Errorc(c, "viewPlaylistChannel s.commonItems err=%+v, aids=%+v, buvid=%s.", err, materialRes.Aids, req.Buvid)
		return nil, err
	}
	var pageNext *commonmdl.PageInfo
	if materialRes.PageNext != nil {
		pageNext = &commonmdl.PageInfo{
			Ps: materialRes.PageNext.Ps,
			Pn: materialRes.PageNext.Pn,
		}
	}
	return &commonmdl.ViewV2SerialResp{
		Cards:    cards,
		PageNext: pageNext,
		HasNext:  materialRes.HasNext,
	}, nil
}

var (
	_ugcURLRex = `(?i)(http(s)?://)?(((uat-)?www.bilibili.com)|(b23.tv|bili22.cn|bili33.cn|bili23.cn|bili2233.cn))(/video)?/((av[0-9]+)|((BV)1[1-9A-NP-Za-km-z]{9}))($|/|)([/.$*?~=#!%@&-A-Za-z0-9_]*)`
	_bvRex     = `(BV|bv|Bv|bV)1[1-9A-NP-Za-km-z]{9}`
	_avRex     = `(AV|av|Av|aV)[0-9]+`
	_idRex     = `[\d]+`
	_ogvURLRex = `(?i)((http(s)?://)?((uat-)?www.bilibili.com/bangumi/(play/|media/)|(b23.tv|bili22.cn|bili33.cn|bili23.cn|bili2233.cn)/)(ss|ep)[0-9]+)($|/|)([/.$*?~=#!%@&-A-Za-z0-9_]*)`
	_ogvssRex  = `(SS|ss|Ss|sS)[0-9]+`
	_ogvepRex  = `(EP|ep|Eo|eP)[0-9]+`

	_ugcURLRgx = regexp.MustCompile(_ugcURLRex)
	_avRgx     = regexp.MustCompile(_avRex)
	_bvRgx     = regexp.MustCompile(_bvRex)
	_idRgx     = regexp.MustCompile(_idRex)
	_ogvURLRgx = regexp.MustCompile(_ogvURLRex)
	_ogvssRgx  = regexp.MustCompile(_ogvssRex)
	_ogvepRgx  = regexp.MustCompile(_ogvepRex)
)

func (s *Service) MediaParse(c context.Context, url string) (*commonmdl.TeslaMediaParseResp, error) {
	access := func() bool {
		if len(s.c.MediaParseAccessIps) == 0 {
			return true
		}
		clientIp := metadata.String(c, metadata.RemoteIP)
		for _, v := range s.c.MediaParseAccessIps {
			if v == clientIp {
				return true
			}
		}
		return false
	}()
	if !access {
		log.Errorc(c, "MediaParse ip denied. ip=%s, url=%s.", metadata.String(c, metadata.RemoteIP), url)
		return nil, ecode.AccessDenied
	}

	// 获取资源类型、id
	shortUrl, err := s.grpcClientShortURL.ShortUrl(c, &grpcShortURL.ShortUrlReq{
		ShortUrl: url,
	})
	if err != nil || shortUrl == nil || shortUrl.Detail == nil || shortUrl.Detail.OriginUrl == "" {
		log.Errorc(c, "MediaParse tesla media parse err=%+v,shortUrl=%+v", err, shortUrl)
		return nil, ecode.ServerErr
	}
	materialType, oid := parseUrl(shortUrl.Detail.OriginUrl)
	if oid <= 0 || (materialType != commonmdl.MaterialTypeOGVSeaon && materialType != commonmdl.MaterialTypeOGVEP && materialType != commonmdl.MaterialTypeUGC) {
		return nil, xecode.AppCannotPlay
	}

	// 查询物料
	p := &commonmdl.Params{}
	switch materialType {
	case commonmdl.MaterialTypeUGC:
		p.ArchiveReq = &commonmdl.ArchiveReq{
			PlayAvs: []*archivegrpc.PlayAv{{
				Aid: oid,
			}},
		}
	case commonmdl.MaterialTypeOGVSeaon:
		p.SeasonReq = &commonmdl.SeasonReq{
			Sids: []int32{int32(oid)},
		}
	case commonmdl.MaterialTypeOGVEP:
		p.EpisodeReq = &commonmdl.EpisodeReq{
			Epids: []int32{int32(oid)},
		}
	default:
		// nop
	}
	var material *commonmdl.CarContext
	material, err = s.material(c, p, model.DeviceInfo{})
	if err != nil {
		log.Errorc(c, "MediaParse s.material err=%+v", err)
		return nil, ecode.ServerErr
	}
	if material == nil {
		log.Errorc(c, "MediaParse s.material res is nil.type=%s, oid=%d.", materialType, oid)
		return nil, xecode.AppNotVedio
	}

	material.OriginData = &commonmdl.OriginData{
		MaterialType: materialType,
		Oid:          oid,
	}
	item := s.formItem(material, model.DeviceInfo{})
	if item == nil {
		log.Errorc(c, "MediaParse s.formItem res is nil.type=%s, oid=%d.", materialType, oid)
		return nil, xecode.AppNotVedio
	}
	return &commonmdl.TeslaMediaParseResp{
		Param:    fmt.Sprintf("play=%s_%d_%d", string(materialType), item.Oid, item.Cid),
		Title:    item.Title,
		Desc:     item.Desc,
		Cover:    item.Cover,
		Duration: item.Duration,
	}, nil
}

func parseUrl(descURLTmp string) (commonmdl.MaterialType, int64) {
	// archive
	ugcIndex := _ugcURLRgx.FindStringIndex(descURLTmp)
	if len(ugcIndex) > 0 {
		ugcURL := descURLTmp[ugcIndex[0]:ugcIndex[1]]
		// 拆bvid
		if bvIndex := _bvRgx.FindStringIndex(ugcURL); len(bvIndex) > 0 {
			bv := ugcURL[bvIndex[0]:bvIndex[1]]
			if aid, _ := bvid.BvToAv(bv); aid != 0 {
				return commonmdl.MaterialTypeUGC, aid
			}
			return "", 0
		}
		// 拆avid
		if avIndex := _avRgx.FindStringIndex(ugcURL); len(avIndex) > 0 {
			avid := ugcURL[avIndex[0]:avIndex[1]]
			// 拆id
			if idIndex := _idRgx.FindStringIndex(avid); len(idIndex) > 0 {
				id := avid[idIndex[0]:idIndex[1]]
				if idInt64, _ := strconv.ParseInt(id, 10, 64); idInt64 != 0 {
					return commonmdl.MaterialTypeUGC, idInt64
				}
			}
		}
		return "", 0
	}
	// ogv
	if ogvIndex := _ogvURLRgx.FindStringIndex(descURLTmp); len(ogvIndex) > 0 {
		ogvURL := descURLTmp[ogvIndex[0]:ogvIndex[1]]
		// 拆ssid
		if ssidIndex := _ogvssRgx.FindStringIndex(ogvURL); len(ssidIndex) > 0 {
			ssid := ogvURL[ssidIndex[0]:ssidIndex[1]]
			if idIndex := _idRgx.FindStringIndex(ssid); len(idIndex) > 0 {
				id := ssid[idIndex[0]:idIndex[1]]
				if idInt, _ := strconv.ParseInt(id, 10, 32); idInt != 0 {
					return commonmdl.MaterialTypeOGVSeaon, idInt
				}
			}
			return "", 0
		}
		if epidIndex := _ogvepRgx.FindStringIndex(ogvURL); len(epidIndex) > 0 {
			epid := ogvURL[epidIndex[0]:epidIndex[1]]
			if idIndex := _idRgx.FindStringIndex(epid); len(idIndex) > 0 {
				id := epid[idIndex[0]:idIndex[1]]
				if idInt, _ := strconv.ParseInt(id, 10, 32); idInt != 0 {
					return commonmdl.MaterialTypeOGVEP, idInt
				}
			}
		}
	}
	return "", 0
}
