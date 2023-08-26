package dynamicV2

import (
	"context"
	"fmt"

	api "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	"go-gateway/app/app-svr/app-dynamic/interface/model"
	mdlv2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
	"go-gateway/pkg/idsafe/bvid"

	activitygrpc "git.bilibili.co/bapis/bapis-go/activity/service"
	dyncommongrpc "git.bilibili.co/bapis/bapis-go/dynamic/common"
)

// 分享组件
func (s *Service) detailShareChannel(c context.Context, dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) error {
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	if len(dynCtx.ShareChannel) == 0 {
		return nil
	}
	shareReq := s.shareReqParam(general, dynCtx.Dyn)
	if shareReq == nil {
		return nil
	}
	var shareModules []*api.ShareChannel
	for _, v := range dynCtx.ShareChannel {
		shareModule := &api.ShareChannel{
			Name:    v.GetName(),
			Image:   v.GetPicture(),
			Channel: v.GetShareChannel(),
		}
		shareModules = append(shareModules, shareModule)
	}
	// 分享组件扩展
	exp := s.detailShareExp(dynCtx, general)
	if len(exp) > 0 {
		shareModules = append(shareModules, exp...)
	}
	module := &api.Module{
		ModuleType: api.DynModuleType_module_share_info,
		ModuleItem: &api.Module_ModuleShareInfo{
			ModuleShareInfo: &api.ModuleShareInfo{
				Title:         "分享至",
				ShareOrigin:   shareReq.ShareOrigin,
				Oid:           shareReq.Oid,
				Sid:           shareReq.Sid,
				ShareChannels: shareModules,
			},
		},
	}
	dynCtx.Modules = append(dynCtx.Modules, module)
	return nil
}

// nolint:gocognit
func (s *Service) detailShareExp(dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) []*api.ShareChannel {
	const (
		_liveStart = 1
		_liveAv    = 2
	)
	var items []*api.ShareChannel
	switch dynCtx.Dyn.Type {
	case mdlv2.DynTypeDraw, mdlv2.DynTypeWord:
		// 图文按钮
		items = append(items, &api.ShareChannel{
			Name:    "生成长图",
			Image:   "http://i0.hdslb.com/bfs/feed-admin/361a2b0466d7cd76660f9fc25ca9a62dbf20a969.png",
			Channel: "LONG CHART",
		})
	}
	// 预约卡
	if !general.IsPadHD() && !general.IsAndroidHD() && !general.IsPad() && !general.IsOverseas() {
		for _, v := range dynCtx.Dyn.AttachCardInfos {
			// nolint:exhaustive
			switch v.CardType {
			case dyncommongrpc.AttachCardType_ATTACH_CARD_RESERVE:
				item := &api.ShareChannel{
					Name:    s.c.Resource.ReserveShare.Name,
					Image:   s.c.Resource.ReserveShare.Image,
					Channel: s.c.Resource.ReserveShare.Channel,
				}
				up, ok := dynCtx.ResUpActRelationInfo[v.Rid]
				if !ok {
					continue
				}
				if up.UpActVisible == activitygrpc.UpActVisible_OnlyUpVisible {
					continue
				}
				switch upActState(up.State) {
				case _upAudit, _upDelete, _upCancel, _upExpired, _upOnline, _upNotStart:
					continue
				}
				reserve := &api.ShareReserve{
					Title:      up.Title,
					QrCodeIcon: s.c.Resource.ReserveShare.QrCodeIcon,
					QrCodeText: s.c.Resource.ReserveShare.QrCodeText,
					QrCodeUrl:  fmt.Sprintf(s.c.Resource.ReserveShare.QrCodeUrl, dynCtx.Dyn.DynamicID),
				}
				var typeDesc string
				// nolint:exhaustive
				switch up.Type {
				case activitygrpc.UpActReserveRelationType_Archive:
					typeDesc = s.c.Resource.ReserveShare.DescAv
				case activitygrpc.UpActReserveRelationType_Live:
					if up.LivePlanStartTime.Time().Unix() > 0 {
						reserve.Desc = model.UpPubShareDataString(up.LivePlanStartTime.Time())
					}
					typeDesc = s.c.Resource.ReserveShare.DescLive
					if upActState(up.State) == _upOnline {
						var isok bool
						// 直播状态
						if info, ok := dynCtx.ResLiveSessionInfo[up.Oid]; ok {
							live, ok := info.SessionInfoPerLive[up.Oid]
							if ok {
								isok = true
								switch live.Status {
								case _liveStart, _liveAv:
									if live.Status == _liveAv {
										aid, _ := bvid.BvToAv(live.Bvid)
										if _, ok := dynCtx.ResArcs[aid]; !ok {
											isok = false
										}
									}
								default:
									isok = false
								}
							}
						}
						if !isok {
							// 已结束（不可回放）
							continue
						}
					}
				case activitygrpc.UpActReserveRelationType_Premiere:
					continue
				}
				if reserve.Desc != "" {
					reserve.Desc = reserve.Desc + " " + typeDesc
				} else {
					reserve.Desc = typeDesc
				}
				if userInfo, ok := dynCtx.GetUser(up.Upmid); ok {
					reserve.UserInfo = &api.AdditionUserInfo{
						Name: userInfo.Name,
						Face: userInfo.Face,
					}
				}
				item.Reserve = reserve
				items = append(items, item)
			}
		}
	}
	return items
}
