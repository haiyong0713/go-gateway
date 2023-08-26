package service

import (
	"context"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/archive/service/api"
	dygrpc "go-gateway/app/web-svr/dynamic/service/api/v1"
	dymdl "go-gateway/app/web-svr/dynamic/service/model"
	"go-gateway/app/web-svr/web/interface/model"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	dyncommongrpc "git.bilibili.co/bapis/bapis-go/dynamic/common"
	seasongrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/season"
)

const (
	_bangumiOfficalMid = 928123

	_dynamicEntranceIconTypeNone = "none"
	_dynamicEntranceIconTypeLive = "live"
	_dynamicEntranceIconTypeUp   = "up"
	_dynamicEntranceIconTypeDyn  = "dyn"
	_dynamicEntranceIconTypeDot  = "dot"
	_dynamicEntranceIcon         = ""
)

func (s *Service) DynamicRegion(ctx context.Context, rid, pn, ps int64, business string, isFilter bool, platform string) (*model.DynamicBvArcs, error) {
	reply, err := s.dyGRPC.RegionArcs3(ctx, &dygrpc.RegionArcs3Req{Rid: rid, Pn: pn, Ps: ps, Business: business, IsFilter: isFilter})
	if err != nil {
		log.Error("s.dyGRPC.RegionArcs3(%d,%d,%d) error(%v)", rid, pn, ps, err)
	}
	if len(reply.GetArcs()) != 0 {
		var aids []int64
		for _, a := range reply.GetArcs() {
			if a.AttrVal(api.AttrBitIsPGC) == api.AttrYes {
				aids = append(aids, a.GetAid())
			}
		}
		eps := func() map[int64]*seasongrpc.CardInfoProto {
			if len(aids) == 0 {
				return nil
			}
			reply, err := s.seasonGRPC.CardsByAids2(ctx, &seasongrpc.SeasonAidReq{Aid2S: aids})
			if err != nil {
				log.Error("%+v", err)
				return nil
			}
			return reply.GetCards()
		}()
		res := &model.DynamicBvArcs{
			Page: &dymdl.Page{
				Num:   int(pn),
				Size:  int(ps),
				Count: int(reply.GetCount()),
			},
			Archives: s.fmtArcs3(reply.GetArcs(), eps),
		}
		if err := s.cache.Do(ctx, func(ctx context.Context) {
			if err := s.dao.SetRegionBakCache(ctx, business, rid, pn, ps, isFilter, res); err != nil {
				log.Error("%+v", err)
			}
		}); err != nil {
			log.Error("%+v", err)
		}
		if platform == "wechat" {
			s.dynamicRegionFilterBindOid(ctx, &res.Archives)
		}
		return res, nil
	}
	res, err := s.dao.RegionBakCache(ctx, business, rid, pn, ps, isFilter)
	if err != nil {
		return nil, err
	}
	if res == nil || len(res.Archives) == 0 {
		for _, val := range s.c.Rule.Rids {
			if int32(rid) == val {
				log.Error("DynamicRegion res is nil rid:%d,business:%s", rid, business)
				break
			}
		}
		return nil, ecode.NothingFound
	}
	if platform == "wechat" {
		s.dynamicRegionFilterBindOid(ctx, &res.Archives)
	}
	return res, nil
}

func (s *Service) dynamicRegionFilterBindOid(c context.Context, containOidSlice *[]*model.BvArc) {
	if len(*containOidSlice) == 0 {
		return
	}
	var oidList []int64
	for _, v := range *containOidSlice {
		if v != nil {
			oidList = append(oidList, v.Aid)
		}
	}

	bindOidList, err := s.dao.TagBind(c, oidList)
	k := 0
	for _, v := range *containOidSlice {
		if err != nil || bindOidList == nil || v == nil || !inIntSlice(bindOidList, v.Aid) {
			(*containOidSlice)[k] = v
			k++
		}
	}
	*containOidSlice = (*containOidSlice)[:k]
}

// nolint:gomnd
func (s *Service) fmtArcs3(arcs []*api.Arc, eps map[int64]*seasongrpc.CardInfoProto) []*model.BvArc {
	var res []*model.BvArc
	for _, v := range arcs {
		if v.Access >= 10000 {
			v.Stat.View = -1
		}
		isOGV := v.AttrVal(api.AttrBitIsPGC) == api.AttrYes // 注意 CopyFromArcToBvArc会清除attr，所以要在它之前判断
		a := model.CopyFromArcToBvArc(v, s.avToBv(v.Aid))
		a.IsOGV = isOGV
		ep, ok := eps[v.GetAid()]
		if ok {
			a.OGVInfo = &model.OGVInfo{
				ReleaseDateShow: ep.GetPublish().GetReleaseDateShow(),
			}
		}
		res = append(res, a)
	}
	return res
}

// DynamicRegionTag get dynamic region tag.
func (s *Service) DynamicRegionTag(c context.Context, tagID, rid, pn, ps int64) (rs *model.DynamicBvArcs, err error) {
	var data *dygrpc.RegionTagArcs3Reply
	if data, err = s.dyGRPC.RegionTagArcs3(c, &dygrpc.RegionTagArcs3Req{Rid: rid, TagId: tagID, Pn: pn, Ps: ps}); err != nil {
		log.Error("s.dyGRPC.RegionTagArcs3(%d,%d,%d,%d) error(%v)", tagID, rid, pn, ps, err)
		err = nil
	} else if data != nil && len(data.Arcs) > 0 {
		rs = &model.DynamicBvArcs{
			Page: &dymdl.Page{
				Num:   int(pn),
				Size:  int(ps),
				Count: int(data.Count),
			},
			Archives: s.fmtArcs3(data.Arcs, nil),
		}
		if err := s.cache.Do(c, func(c context.Context) {
			if err := s.dao.SetRegionTagBakCache(c, tagID, rid, pn, ps, rs); err != nil {
				log.Error("%+v", err)
			}
		}); err != nil {
			log.Error("%+v", err)
		}
		return
	}
	if rs, err = s.dao.RegionTagBakCache(c, tagID, rid, pn, ps); err != nil {
		return
	}
	if rs == nil {
		err = ecode.NothingFound
	}
	return
}

// DynamicRegionTotal get dynamic region total.
func (s *Service) DynamicRegionTotal(c context.Context) (map[string]int64, error) {
	rs, err := s.dyGRPC.RegionTotal(c, &dygrpc.NoArgRequest{})
	if err != nil {
		log.Error("s.dyGRPC.RegionTotal error(%v)", err)
		return nil, err
	}
	return rs.Res, nil
}

// DynamicRegions get dynamic regions.
// nolint:gomnd
func (s *Service) DynamicRegions(c context.Context) (rs map[int32][]*model.BvArc, err error) {
	var (
		rids   []int32
		common map[int32][]*api.Arc
		bg     *dygrpc.RegionArcs3Reply
		bgid   = int32(13)
		ip     = metadata.String(c, metadata.RemoteIP)
	)
	// get first type id
	for _, rid := range s.c.Rule.Rids {
		if rid == bgid { //bangumi ignore.
			continue
		} else if rid == 167 { //guochuang use second rid 168.
			rid = 168
		}
		rids = append(rids, rid)
	}
	rs = make(map[int32][]*model.BvArc, len(rids)+1)
	if common, err = s.dy.RegionsArcs3(c, &dymdl.ArgRegions3{RegionIDs: rids, Count: 10, RealIP: ip}); err != nil {
		log.Error("s.dy.RegionsArcs3(%v) error(%v)", rids, err)
		err = nil
	}
	for _, rid := range rids {
		for _, v := range common[rid] {
			rs[rid] = append(rs[rid], &model.BvArc{Arc: v, Bvid: s.avToBv(v.Aid)})
		}
	}
	// bangumi type id 13 find 200,condition mid == 928123.
	if bg, err = s.dyGRPC.RegionArcs3(c, &dygrpc.RegionArcs3Req{Rid: int64(bgid), Pn: 1, Ps: s.c.Rule.BangumiCount}); err != nil {
		log.Error("s.dy.RegionsArcs3 error(%v)", err)
		err = nil
	} else {
		n := 1
		count := 1
		for _, arc := range bg.Arcs {
			count++
			if arc.Author.Mid == _bangumiOfficalMid {
				rs[bgid] = append(rs[bgid], &model.BvArc{Arc: arc, Bvid: s.avToBv(arc.Aid)})
			} else {
				continue
			}
			n++
			if n > s.c.Rule.RegionsCount {
				log.Info("s.dy.RegionsArcs bangumi count(%d)", count)
				break
			}
		}
		// not enough add other.
		if n <= s.c.Rule.RegionsCount {
			for _, arc := range bg.Arcs {
				count++
				if arc.Author.Mid == _bangumiOfficalMid {
					continue
				} else {
					rs[bgid] = append(rs[bgid], &model.BvArc{Arc: arc, Bvid: s.avToBv(arc.Aid)})
				}
				n++
				if n > s.c.Rule.RegionsCount {
					log.Info("s.dy.RegionsArcs bangumi count(%d)", count)
					break
				}
			}
		}
	}
	if len(rs) > 0 {
		countCheck := true
		for rid, region := range rs {
			if len(region) < s.c.Rule.MinDyCount {
				countCheck = false
				log.Info("countCheck rid(%d) len(%d) false", rid, len(region))
				break
			}
		}
		if countCheck {
			if err := s.cache.Do(c, func(c context.Context) {
				if err := s.dao.SetRegionsBakCache(c, rs); err != nil {
					log.Error("%+v", err)
				}
			}); err != nil {
				log.Error("%+v", err)
			}
			return
		}
	}
	rs, err = s.dao.RegionsBakCache(c)
	return
}

func (s *Service) DynamicEntrance(c context.Context, req *model.DynamicEntranceParam, mid int64) (*model.DynamicEntrance, error) {
	resTmp, err := s.dao.DynamicEntrance(c, mid, req.VideoOffset, req.ArticleOffset, req.AlltypeOffset)
	if err != nil {
		log.Error("%v", err)
		return nil, err
	}
	if resTmp == nil {
		return nil, nil
	}
	var accRes *accountgrpc.CardReply
	if resTmp.GetShowUid() != 0 {
		accRes, err = s.accGRPC.Card3(c, &accountgrpc.MidReq{Mid: resTmp.GetShowUid()})
		if err != nil {
			log.Error("%+v", err)
		}
	}
	var res = new(model.DynamicEntrance)
	res.Entrance = &model.DynamicEntranceItem{
		Type: s.formDynamicEntranceIconType(resTmp.GetIconType()),
		Mid:  resTmp.GetShowUid(),
		Icon: _dynamicEntranceIcon,
	}
	if accRes != nil && accRes.GetCard() != nil {
		res.Entrance.Icon = accRes.GetCard().GetFace()
	}
	res.UpdateInfo = &model.DynamicEntranceUpdateInfo{
		Type: "count",
		Item: &model.DynamicEntranceUpdateInfoItem{
			Count: resTmp.GetVideoNum() + resTmp.GetArticleNum(),
		},
	}
	return res, nil
}

func (s *Service) DynamicCardType(c context.Context, mid int64) (*model.DynamicCardType, error) {
	resTmp, err := s.dao.DynamicAttachPermissionCheck(c, mid)
	if err != nil {
		log.Error("%v", err)
		return nil, err
	}
	if resTmp != nil && resTmp.IsAllow {
		var res = new(model.DynamicCardType)
		res.Items = append(res.Items, &model.DynamicCardTypeItem{
			Title:    "课堂",
			CardType: 13,
		})
		return res, nil
	} else {
		return nil, nil
	}
}

func (s *Service) DynamicCardAdd(c context.Context, mid int64, url string) (*model.DynamicCardAdd, error) {
	addRes, err := s.dao.DynamicAttachAdd(c, mid, url)
	if err != nil {
		log.Error("%v", err)
		return nil, err
	}
	var res = new(model.DynamicCardAdd)
	res.IsAllow = false
	res.ErrorMsg = "不允许添加"
	res.SeasonID = 0
	if addRes != nil {
		res.IsAllow = addRes.IsAllow
		res.ErrorMsg = addRes.ErrorMsg
		res.SeasonID = addRes.SeasonId
		if addRes.IsAllow && addRes.SeasonId != 0 {
			profileRes, err := s.dao.DynamicSeasonProfile(c, []int32{addRes.SeasonId})
			if err != nil {
				log.Error("%v", err)
			} else if profileRes != nil && len(profileRes.Cards) > 0 {
				res.SeasonProfile = profileRes.Cards[addRes.SeasonId]
			}
		}
	}
	return res, nil
}

func (s *Service) DynamicCardCanAddContent(c context.Context, mid int64, pn, ps int32) (*model.DynamicCardCanAddContent, error) {
	resTmp, err := s.dao.DynamicUserSeason(c, mid, pn, ps)
	if err != nil {
		log.Error("%v", err)
	}
	var res = new(model.DynamicCardCanAddContent)
	res.LinkQuery = true
	if resTmp != nil {
		res.UserSeason = resTmp
	}
	return res, nil
}

func (s *Service) formDynamicEntranceIconType(iconType dyncommongrpc.IconType) string {
	switch iconType {
	case dyncommongrpc.IconType_ICON_TYPE_NONE:
		return _dynamicEntranceIconTypeNone
	case dyncommongrpc.IconType_ICON_TYPE_LIVE:
		return _dynamicEntranceIconTypeLive
	case dyncommongrpc.IconType_ICON_TYPE_UP:
		return _dynamicEntranceIconTypeUp
	case dyncommongrpc.IconType_ICON_TYPE_DYN:
		return _dynamicEntranceIconTypeDyn
	case dyncommongrpc.IconType_ICON_TYPE_DOT:
		return _dynamicEntranceIconTypeDot
	}
	return _dynamicEntranceIconTypeNone
}
